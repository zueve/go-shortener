package server

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/zueve/go-shortener/internal/services"
)

type Server struct {
	service    services.Service
	serviceURL string
	port       int
}

func New(service services.Service, serviceURL string, port int) Server {
	return Server{service: service, serviceURL: serviceURL, port: port}
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
	var url string
	switch headerContentType {
	case "application/x-www-form-urlencoded":
		r.ParseForm()
		url = r.FormValue("url")
	case "text/plain; charset=utf-8":
		urlBytes, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Println("invalid parse body")
		}
		url = strings.TrimSuffix(string(urlBytes), "\n")
	default:
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
	resultURL := fmt.Sprintf("%s/%s", s.serviceURL, key)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(resultURL))
}

func (s *Server) redirect(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "keyID")
	fmt.Println("Call redirect for", key)
	url, err := s.service.GetURLByKey(key)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println("invalid key", key)
		return
	}
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
