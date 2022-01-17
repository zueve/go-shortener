package server

import (
	"context"
	"encoding/json"
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
	r.Post("/api/shorten", newServer.createRedirectJSON)
	r.Get("/{keyID}", newServer.redirect)

	srv := http.Server{
		Addr:    newServer.ctx.ServerAddress,
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

	var url string
	switch headerContentType {
	case "application/x-www-form-urlencoded":
		r.ParseForm()
		url = r.FormValue("url")
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
	resultURL := fmt.Sprintf("%s/%s", s.ctx.ServiceURL, key)
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
		Result: fmt.Sprintf("%s/%s", s.ctx.ServiceURL, key),
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
