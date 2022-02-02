package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/zueve/go-shortener/internal/services"
)

type Server struct {
	service       services.Service
	srv           *http.Server
	serverAddress string
	serviceURL    string
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
	)

	s := Server{
		srv:           nil,
		service:       service,
		serverAddress: defaultServerAddress,
		serviceURL:    defaultServiceURL,
	}

	for _, opt := range opts {
		if err := opt(&s); err != nil {
			return Server{}, err
		}
	}

	r := chi.NewRouter()
	r.Use(ungzipHandle)
	r.Use(gzipHandle)
	r.Post("/", s.createRedirect)
	r.Post("/api/shorten", s.createRedirectJSON)
	r.Get("/{keyID}", s.redirect)
	r.Get("/user/urls", s.GetAllUserURLs)

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
		s.error(w, http.StatusInternalServerError, "invalid parse token", err)
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
	key := s.service.CreateRedirect(url, userID)
	resultURL := fmt.Sprintf("%s/%s", s.serviceURL, key)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(resultURL))
}

func (s *Server) redirect(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		s.error(w, http.StatusInternalServerError, "invalid parse token", err)
	}
	key := chi.URLParam(r, "keyID")
	fmt.Println("Call redirect for", key)
	url, err := s.service.GetURLByKey(key, userID)
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
		s.error(w, http.StatusInternalServerError, "invalid parse token", err)
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
	key := s.service.CreateRedirect(redirect.URL, userID)
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
		fmt.Println(err)
		s.error(w, http.StatusInternalServerError, "invalid parse token", err)
	}
	linksMap := s.service.GetAllUserURLs(userID)

	result := make([]URLRow, len(linksMap))
	i := 0
	for key, url := range linksMap {
		result[i] = URLRow{
			OriginalURL: url,
			ShortURL:    fmt.Sprintf("%s/%s", s.serviceURL, key),
		}
		i = i + 1
	}
	fmt.Println(result)
	response, err := json.Marshal(result)
	if err != nil {
		s.error(w, http.StatusInternalServerError, "internal server error", err)
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response))
}

func (s *Server) error(w http.ResponseWriter, code int, msg string, err error) {
	if err != nil {
		fmt.Println(err)
	}
	w.WriteHeader(code)
	w.Header().Set("content-type", "plain/text")
	fmt.Println(msg)
	w.Write([]byte(msg))
}
