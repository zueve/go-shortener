package server

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/zueve/go-shortener/internal/services"
	"io"
	"net/http"
	"strings"
)

type Server struct {
	service services.Service
}

func New(service services.Service) Server {
	return Server{service: service}
}

func (s *Server) Run() {
	r := chi.NewRouter()
	r.Post("/", s.createRedirect)
	r.Get("/{keyID}", s.redirect)
	http.ListenAndServe(":8080", r)
}

func (s *Server) createRedirect(w http.ResponseWriter, r *http.Request) {
	headerContentType := r.Header.Get("Content-Type")
	w.Header().Set("content-type", "text/plain")
	var url = ""
	if headerContentType == "application/x-www-form-urlencoded" {
		r.ParseForm()
		url = r.FormValue("url")
	} else if headerContentType == "text/plain; charset=utf-8" {
		url_bytes, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			fmt.Println("invalid parse body")
		}
		url = string(url_bytes)
	} else {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		fmt.Println("invalid ContentType")
		return
	}

	if url == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println("invalid url")
		return
	}

	key := s.service.CreateRedirect(url)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(key))
}

func (s *Server) redirect(w http.ResponseWriter, r *http.Request) {
	// TODO don't work
	// key := chi.URLParam(r, "keyID")
	key := strings.TrimPrefix(r.URL.Path, "/")
	url, err := s.service.GetUrlByKey(key)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println("invalid key")
		return
	}
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
