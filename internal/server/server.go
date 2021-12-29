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
	service     services.Service
	service_url string
	port        int
}

func New(service services.Service, service_url string, port int) Server {
	return Server{service: service, service_url: service_url, port: port}
}

func (s *Server) Run() {
	r := chi.NewRouter()
	r.Post("/", s.createRedirect)
	r.Get("/{keyID}", s.redirect)
	loc := fmt.Sprintf(":%d", s.port)
	http.ListenAndServe(loc, r)
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
		url = strings.TrimSuffix(string(url_bytes), "\n")
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

	fmt.Println("Add url", url)
	key := s.service.CreateRedirect(url)
	result_url := fmt.Sprintf("%s/%s", s.service_url, key)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(result_url))
}

func (s *Server) redirect(w http.ResponseWriter, r *http.Request) {
	// TODO don't work
	// key := chi.URLParam(r, "keyID")
	key := strings.TrimPrefix(r.URL.Path, "/")
	url, err := s.service.GetUrlByKey(key)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println("invalid key", key)
		return
	}
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
