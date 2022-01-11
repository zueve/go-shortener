package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/zueve/go-shortener/internal/config"
	"github.com/zueve/go-shortener/internal/services"
)

type Server struct {
	service services.Service
	ctx     *config.Context
	srv     *http.Server
}

func New(ctx *config.Context, service services.Service) Server {
	newServer := Server{ctx: ctx, service: service, srv: nil}

	r := chi.NewRouter()
	r.Post("/", newServer.createRedirect)
	r.Get("/{keyID}", newServer.redirect)
	loc := fmt.Sprintf(":%d", newServer.ctx.Port)

	srv := http.Server{
		Addr:    loc,
		Handler: r,
	}
	newServer.srv = &srv

	return newServer
}

func (s *Server) ListenAndServe() {
	s.srv.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
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
			return
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
	resultURL := fmt.Sprintf("%s/%s", s.ctx.ServiceURL, key)
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
