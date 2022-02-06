package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/zueve/go-shortener/internal/services"
)

type Server struct {
	service       services.Service
	srv           *http.Server
	serverAddress string
	serviceURL    string
	pingTimeout   time.Duration
}

type ServerOption func(*Server) error

func WithAddress(address string) ServerOption {
	return func(h *Server) error {
		h.serverAddress = address
		return nil
	}
}

func WithURL(url string) ServerOption {
	return func(h *Server) error {
		h.serviceURL = url
		return nil
	}
}

func New(service services.Service, opts ...ServerOption) (Server, error) {
	const (
		defaultServerAddress = ":8080"
		defaultServiceURL    = "http://localhost:8080"
		defaultPingTimeout   = 500 * time.Millisecond
	)

	s := Server{
		srv:           nil,
		service:       service,
		serverAddress: defaultServerAddress,
		serviceURL:    defaultServiceURL,
		pingTimeout:   1 * time.Second,
	}

	for _, opt := range opts {
		if err := opt(&s); err != nil {
			return Server{}, err
		}
	}

	r := chi.NewRouter()
	r.Use(ungzipHandle)
	r.Use(gzipHandle)
	r.Use(setCookieHandler)
	r.Post("/", s.createRedirect)
	r.Post("/api/shorten/batch", s.createRedirectByBatch)
	r.Post("/api/shorten", s.createRedirectJSON)
	r.Get("/{keyID}", s.redirect)
	r.Get("/user/urls", s.GetAllUserURLs)
	r.Get("/ping", s.PingStorage)

	srv := http.Server{
		Addr:    s.serverAddress,
		Handler: r,
	}
	s.srv = &srv

	return s, nil
}

func (s *Server) ListenAndServe() {
	s.srv.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}

func (s *Server) createRedirect(w http.ResponseWriter, r *http.Request) {
	headerContentType := r.Header.Get("Content-Type")
	userID, err := getUserID(r)
	if err != nil {
		s.error(w, http.StatusInternalServerError, "invalid token", err)
		return
	}

	var url string
	switch headerContentType {
	case "application/x-www-form-urlencoded":
		r.ParseForm()
		url = r.FormValue("url")
	case "application/x-gzip":
		urlBytes, err := io.ReadAll(r.Body)
		if err != nil {
			s.error(w, http.StatusInternalServerError, "invalid body", nil)
			return
		}
		url = strings.TrimSuffix(string(urlBytes), "\n")
	case "text/plain; charset=utf-8":
		urlBytes, err := io.ReadAll(r.Body)
		if err != nil {
			s.error(w, http.StatusInternalServerError, "invalid body", nil)
			return
		}
		url = strings.TrimSuffix(string(urlBytes), "\n")
	default:
		s.error(w, http.StatusUnsupportedMediaType, "invalid ContentType", nil)
		return
	}
	if url == "" {
		s.error(w, http.StatusBadRequest, "invalid url", nil)
		return
	}

	w.Header().Set("content-type", "text/plain")
	fmt.Println("Add url", url)
	key, err := s.service.CreateRedirect(s.context(r), url, userID)
	if err != nil {
		s.error(w, http.StatusInternalServerError, "invalid token", err)
		return
	}
	resultURL := fmt.Sprintf("%s/%s", s.serviceURL, key)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(resultURL))
}

func (s *Server) redirect(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "keyID")
	fmt.Println("Call redirect for", key)
	url, err := s.service.GetURLByKey(s.context(r), key)
	if err != nil {
		s.error(w, http.StatusBadRequest, "invalid key", err)
		return
	}
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (s *Server) createRedirectJSON(w http.ResponseWriter, r *http.Request) {
	headerContentType := r.Header.Get("Content-Type")
	userID, err := getUserID(r)
	if err != nil {
		s.error(w, http.StatusInternalServerError, "invalid token", err)
		return
	}

	var redirect Redirect
	switch headerContentType {
	case "application/json":
		dataBytes, err := io.ReadAll(r.Body)
		if err != nil {
			s.error(w, http.StatusInternalServerError, "invalid body", err)
			return
		}
		err = json.Unmarshal(dataBytes, &redirect)
		if err != nil || redirect.URL == "" {
			s.error(w, http.StatusBadRequest, "invalid body", err)
			return
		}
	default:
		s.error(w, http.StatusUnsupportedMediaType, "invalid ContentType", nil)
		return
	}
	fmt.Println("Create redirect for", redirect.URL)
	key, err := s.service.CreateRedirect(s.context(r), redirect.URL, userID)
	if err != nil {
		s.error(w, http.StatusInternalServerError, "internal error", err)
		return
	}
	result := ResultString{
		Result: fmt.Sprintf("%s/%s", s.serviceURL, key),
	}

	response, err := json.Marshal(result)
	if err != nil {
		s.error(w, http.StatusInternalServerError, "internal error", err)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(response))
}

func (s *Server) GetAllUserURLs(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		s.error(w, http.StatusInternalServerError, "invalid token", err)
		return
	}
	linksMap, err := s.service.GetAllUserURLs(s.context(r), userID)
	if err != nil {
		s.error(w, http.StatusInternalServerError, "internal error", err)
		return
	}

	result := make([]URLRow, len(linksMap))
	i := 0
	for key, url := range linksMap {
		result[i] = URLRow{
			OriginalURL: url,
			ShortURL:    fmt.Sprintf("%s/%s", s.serviceURL, key),
		}
		i = i + 1
	}
	response, err := json.Marshal(result)
	if err != nil {
		s.error(w, http.StatusInternalServerError, "internal server error", err)
		return
	}
	status := http.StatusOK
	if len(linksMap) == 0 {
		status = http.StatusNoContent
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(response))
}

func (s *Server) createRedirectByBatch(w http.ResponseWriter, r *http.Request) {
	headerContentType := r.Header.Get("Content-Type")
	if headerContentType != "application/json" {
		s.error(w, http.StatusUnsupportedMediaType, "invalid ContentType", nil)
	}
	userID, err := getUserID(r)
	if s.internalError(w, err) {
		return
	}
	// parse request
	dataBytes, err := io.ReadAll(r.Body)
	if s.internalError(w, err) {
		return
	}
	requestURLs := make([]URLRowOriginal, 0)
	err = json.Unmarshal(dataBytes, &requestURLs)
	if err != nil {
		s.error(w, http.StatusBadRequest, "invalid body", err)
		return
	}
	// transform request to internal format
	size := len(requestURLs)
	urls := make([]string, size)
	for i := range requestURLs {
		urls[i] = requestURLs[i].OriginalURL
	}

	urls, err = s.service.CreateRedirectByBatch(s.context(r), urls, userID)

	// transform result to responce format
	responseURLs := make([]URLRowShort, size)
	if s.internalError(w, err) {
		return
	}
	for i := range requestURLs {
		responseURLs[i] = URLRowShort{
			CorrelationID: requestURLs[i].CorrelationID,
			ShortURL:      fmt.Sprintf("%s/%s", s.serviceURL, urls[i]),
		}
	}
	response, err := json.Marshal(responseURLs)
	if s.internalError(w, err) {
		return
	}
	status := http.StatusOK
	if len(responseURLs) == 0 {
		status = http.StatusNoContent
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(response))
}

func (s *Server) PingStorage(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(s.pingTimeout))
	defer cancel()
	err := s.service.Ping(ctx)
	if err != nil {
		s.internalError(w, err)
		return
	}
	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *Server) error(w http.ResponseWriter, code int, msg string, err error) {
	if err != nil {
		fmt.Println(err)
	}
	w.WriteHeader(code)
	w.Header().Set("content-type", "text/plain")
	fmt.Println(msg)
	w.Write([]byte(msg))
}

func (s *Server) internalError(w http.ResponseWriter, err error) bool {
	if err != nil {
		s.error(w, http.StatusInternalServerError, "internal server error", err)
	}
	return err != nil
}

func (s Server) context(r *http.Request) context.Context {
	return r.Context()
}
