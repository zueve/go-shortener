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
	ServerAddress string
	ServiceURL    string
}

func New(service services.Service, opts ...ServerOption) Server {
	const (
		defaultServerAddress = ":8080"
		defaultServiceURL    = "http://localhost:8080"
	)

	s := Server{
		srv:           nil,
		service:       service,
		ServerAddress: defaultServerAddress,
		ServiceURL:    defaultServiceURL,
	}

	for _, opt := range opts {
		opt(&s)
	}

	r := chi.NewRouter()
	r.Use(ungzipHandle)
	r.Use(gzipHandle)
	r.Post("/", s.createRedirect)
	r.Post("/api/shorten", s.createRedirectJSON)
	r.Get("/{keyID}", s.redirect)

	srv := http.Server{
		Addr:    s.ServerAddress,
		Handler: r,
	}
	s.srv = &srv

	return s
}

func (s *Server) ListenAndServe() {
	s.srv.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}

func (s *Server) createRedirect(w http.ResponseWriter, r *http.Request) {
	headerContentType := r.Header.Get("Content-Type")

	var url string
	switch headerContentType {
	case "application/x-www-form-urlencoded":
		r.ParseForm()
		url = r.FormValue("url")
	case "application/x-gzip":
		urlBytes, err := io.ReadAll(r.Body)
		if err != nil {
			s.error(w, http.StatusInternalServerError, "invalid parse body")
			return
		}
		url = strings.TrimSuffix(string(urlBytes), "\n")
	case "text/plain; charset=utf-8":
		urlBytes, err := io.ReadAll(r.Body)
		if err != nil {
			s.error(w, http.StatusInternalServerError, "invalid parse body")
			return
		}
		url = strings.TrimSuffix(string(urlBytes), "\n")
	default:
		s.error(w, http.StatusUnsupportedMediaType, "invalid ContentType")
		return
	}
	if url == "" {
		s.error(w, http.StatusBadRequest, "invalid url")
		return
	}

	w.Header().Set("content-type", "text/plain")
	fmt.Println("Add url", url)
	key := s.service.CreateRedirect(url)
	resultURL := fmt.Sprintf("%s/%s", s.ServiceURL, key)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(resultURL))
}

func (s *Server) redirect(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "keyID")
	fmt.Println("Call redirect for", key)
	url, err := s.service.GetURLByKey(key)
	if err != nil {
		s.error(w, http.StatusBadRequest, "invalid key")
		return
	}
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (s *Server) createRedirectJSON(w http.ResponseWriter, r *http.Request) {
	headerContentType := r.Header.Get("Content-Type")

	var redirect Redirect
	switch headerContentType {
	case "application/json":
		dataBytes, err := io.ReadAll(r.Body)
		if err != nil {
			s.error(w, http.StatusInternalServerError, "invalid parse body")
			return
		}
		err = json.Unmarshal(dataBytes, &redirect)
		if err != nil || redirect.URL == "" {
			s.error(w, http.StatusBadRequest, "invalid parse body")
			return
		}
	default:
		s.error(w, http.StatusUnsupportedMediaType, "invalid ContentType")
		return
	}
	fmt.Println("Create redirect for", redirect.URL)
	key := s.service.CreateRedirect(redirect.URL)
	result := ResultString{
		Result: fmt.Sprintf("%s/%s", s.ServiceURL, key),
	}

	response, _ := json.Marshal(result)
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(response))
}

func (s *Server) error(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	w.Header().Set("content-type", "plain/text")
	fmt.Println(msg)
	w.Write([]byte(msg))
}
