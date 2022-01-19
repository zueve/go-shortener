package server

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/zueve/go-shortener/internal/config"
	"github.com/zueve/go-shortener/internal/services"
	"github.com/zueve/go-shortener/internal/storage"
)

const filename = "/tmp/go-shortener-test.txt"

func getCtx() *config.Context {
	return config.NewContext(
		config.WithServiceURL("http://localhost:8080"),
		config.WithServerAddress(":8080"),
	)
}

func TestServer_createRedirect(t *testing.T) {
	os.Remove(filename)
	persistentStorage := storage.NewFileStorage(filename)
	defer persistentStorage.Close()
	storageTest := storage.New(persistentStorage)
	serviceTest := services.New(storageTest)
	tests := []struct {
		name        string
		method      string
		contentType string
		code        int
		urlKey      string
		urlVal      string
	}{
		{
			name:        "positive test1",
			method:      http.MethodPost,
			contentType: "application/x-www-form-urlencoded",
			code:        201,
			urlKey:      "url",
			urlVal:      "http://example.com/...",
		},
		{
			name:        "negative data",
			method:      http.MethodPost,
			contentType: "application/x-www-form-urlencoded",
			code:        400,
			urlKey:      "url0",
			urlVal:      "http://example.com/...",
		},
		{
			name:        "negative empty url",
			method:      http.MethodPost,
			contentType: "application/x-www-form-urlencoded",
			code:        400,
			urlKey:      "url0",
			urlVal:      "",
		},
		{
			name:        "negative invalid method",
			method:      http.MethodGet,
			contentType: "application/x-www-form-urlencoded",
			code:        400,
			urlKey:      "url",
			urlVal:      "http://example.com/...",
		},
		{
			name:        "negative invalid content type",
			method:      http.MethodPost,
			contentType: "application/json",
			code:        415,
			urlKey:      "url",
			urlVal:      "http://example.com/...",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(getCtx(), serviceTest)
			data := url.Values{}
			data.Set(tt.urlKey, tt.urlVal)

			request := httptest.NewRequest(tt.method, "/", bytes.NewBufferString(data.Encode()))
			request.Header.Set("Content-Type", tt.contentType)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(s.createRedirect)

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
	os.Remove(filename)
	persistentStorage := storage.NewFileStorage(filename)
	storageTest := storage.New(persistentStorage)
	serviceTest := services.New(storageTest)
	location := "https://example.com"
	validKey := serviceTest.CreateRedirect(location)
	client := http.Client{}
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	tests := []struct {
		name     string
		method   string
		code     int
		url      string
		location string
	}{
		{
			name:     "positive test1",
			method:   http.MethodGet,
			code:     307,
			url:      fmt.Sprintf("/%s", validKey),
			location: location,
		},
		{
			name:     "negative test2",
			method:   http.MethodGet,
			code:     400,
			url:      "/invalid",
			location: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(getCtx(), serviceTest)

			r := chi.NewRouter()
			r.Get("/{keyID}", s.redirect)
			ts := httptest.NewServer(r)
			defer ts.Close()
			url := fmt.Sprintf("%s%s", ts.URL, tt.url)
			fmt.Println("Url - ", url)
			res, err := client.Get(url)
			if err != nil {
				t.Errorf("Problem with server")
			}
			defer res.Body.Close()
			if res.StatusCode != tt.code {
				t.Errorf("Expected status code %d, got %d", tt.code, res.StatusCode)
			}
			if tt.code == 307 {
				loc := res.Header.Get("location")
				if loc != tt.location {
					t.Errorf("Expected location %s, got %s", tt.location, loc)
				}
			}

		})
	}
	persistentStorage.Close()
}

func TestServer_createRedirectJSON(t *testing.T) {
	os.Remove(filename)
	persistentStorage := storage.NewFileStorage(filename)
	storageTest := storage.New(persistentStorage)
	serviceTest := services.New(storageTest)
	tests := []struct {
		name        string
		method      string
		contentType string
		code        int
		data        string
		result      string
	}{
		{
			name:        "positive test1",
			method:      http.MethodPost,
			contentType: "application/json",
			code:        201,
			data:        "{\"url\": \"http://example.com\"}",
			result:      "{\"result\":\"http://localhost:8080/2\"}",
		},
		{
			name:        "positive test2",
			method:      http.MethodPost,
			contentType: "application/json",
			code:        201,
			data:        "{\"url\": \"http://example.com\"}",
			result:      "{\"result\":\"http://localhost:8080/3\"}",
		},
		{
			name:        "negative test3",
			method:      http.MethodPost,
			contentType: "application/json",
			code:        400,
			data:        "{\"param\": 123}",
			result:      "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(getCtx(), serviceTest)
			request := httptest.NewRequest(tt.method, "/", bytes.NewBufferString(tt.data))
			request.Header.Set("Content-Type", tt.contentType)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(s.createRedirectJSON)

			h.ServeHTTP(w, request)
			res := w.Result()
			if res.StatusCode != tt.code {
				t.Errorf("Expected status code %d, got %d", tt.code, w.Code)
			}
			defer res.Body.Close()
			if tt.code == 201 {
				payloadBytes, _ := io.ReadAll(res.Body)
				payload := string(payloadBytes)
				if payload != tt.result {
					t.Errorf("Expected result %s, got %s", tt.result, payload)
				}
			}
		})
	}
}
