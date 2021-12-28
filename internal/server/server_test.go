package server

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/zueve/go-shortener/internal/services"
	"github.com/zueve/go-shortener/internal/storage"
)

func TestServer_createRedirect(t *testing.T) {
	var storage_ = storage.New()
	var service_ = services.New(&storage_)
	tests := []struct {
		name        string
		method      string
		contentType string
		code        int
		url_key     string
		url_val     string
	}{
		{
			name:        "positive test1",
			method:      http.MethodPost,
			contentType: "application/x-www-form-urlencoded",
			code:        201,
			url_key:     "url",
			url_val:     "http://example.com/...",
		},
		{
			name:        "negative data",
			method:      http.MethodPost,
			contentType: "application/x-www-form-urlencoded",
			code:        400,
			url_key:     "url0",
			url_val:     "http://example.com/...",
		},
		{
			name:        "negative empty url",
			method:      http.MethodPost,
			contentType: "application/x-www-form-urlencoded",
			code:        400,
			url_key:     "url0",
			url_val:     "",
		},
		{
			name:        "negative invalid method",
			method:      http.MethodGet,
			contentType: "application/x-www-form-urlencoded",
			code:        400,
			url_key:     "url",
			url_val:     "http://example.com/...",
		},
		{
			name:        "negative invalid content type",
			method:      http.MethodPost,
			contentType: "application/json",
			code:        415,
			url_key:     "url",
			url_val:     "http://example.com/...",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{service: service_}
			data := url.Values{}
			data.Set(tt.url_key, tt.url_val)

			request := httptest.NewRequest(tt.method, "/", bytes.NewBufferString(data.Encode()))
			request.Header.Set("Content-Type", tt.contentType)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(s.routeRedirect)

			h.ServeHTTP(w, request)
			res := w.Result()
			if res.StatusCode != tt.code {
				t.Errorf("Expected status code %d, got %d", tt.code, w.Code)
			}
			defer res.Body.Close()
		})
	}
}

func TestServer_redirect(t *testing.T) {
	var storage_ = storage.New()
	var service_ = services.New(&storage_)
	var valid_key = service_.CreateRedirect("https://example.com")
	tests := []struct {
		name   string
		method string
		code   int
		url    string
	}{
		{
			name:   "positive test1",
			method: http.MethodGet,
			code:   307,
			url:    fmt.Sprintf("/%s", valid_key),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{service: service_}
			request := httptest.NewRequest(tt.method, tt.url, nil)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(s.routeRedirect)

			h.ServeHTTP(w, request)
			res := w.Result()
			if res.StatusCode != tt.code {
				t.Errorf("Expected status code %d, got %d", tt.code, w.Code)
			}
			defer res.Body.Close()
		})
	}
}
