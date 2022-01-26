package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/stretchr/testify/assert"
	"github.com/zueve/go-shortener/internal/services"
	"github.com/zueve/go-shortener/internal/storage"
)

type TestServer struct {
	*httptest.Server
	storage           *storage.Storage
	persistentStorage *storage.FileStorage
	service           services.Service
	filename          string
}

func NewTestServer() (TestServer, error) {
	file, err := os.CreateTemp("", "go_shortener")
	if err != nil {
		return TestServer{}, err
	}
	os.Remove(file.Name())
	persistentStorage, _ := storage.NewFileStorage(file.Name())
	storageTest, err := storage.New(persistentStorage)
	if err != nil {
		return TestServer{}, err
	}
	serviceTest := services.New(storageTest)

	s, err := New(serviceTest)
	if err != nil {
		return TestServer{}, err
	}

	r := chi.NewRouter()
	r.Use(ungzipHandle)
	r.Use(gzipHandle)
	r.Use(setCookieHandler)
	r.Post("/", s.createRedirect)
	r.Post("/api/shorten", s.createRedirectJSON)
	r.Get("/{keyID}", s.redirect)
	ts := httptest.NewServer(r)

	srv := TestServer{
		Server:            ts,
		storage:           storageTest,
		persistentStorage: persistentStorage,
		service:           serviceTest,
		filename:          file.Name(),
	}

	return srv, nil
}

func (s *TestServer) Close() {
	s.Server.Close()
	s.persistentStorage.Close()
	os.Remove(s.filename)
}

func TestServer_createRedirect(t *testing.T) {
	ts, _ := NewTestServer()
	defer ts.Close()
	client := http.Client{}
	reqURL := fmt.Sprintf("%s/", ts.URL)

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
			method:      http.MethodPatch,
			contentType: "application/x-www-form-urlencoded",
			code:        405,
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
			data := url.Values{}
			data.Set(tt.urlKey, tt.urlVal)

			request, err := http.NewRequest(tt.method, reqURL, bytes.NewBufferString(data.Encode()))
			assert.Nil(t, err)

			request.Header.Set("Content-Type", tt.contentType)
			res, err := client.Do(request)

			assert.Nil(t, err)
			assert.Equal(t, res.StatusCode, tt.code, "statuses should be equal")

			defer res.Body.Close()
		})
	}
}

func TestServer_redirect(t *testing.T) {
	ts, _ := NewTestServer()
	defer ts.Close()
	location := "https://example.com"
	validKey := ts.service.CreateRedirect(location, "1")
	client := http.Client{}
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	tests := []struct {
		name     string
		method   string
		url      string
		code     int
		location string
	}{
		{
			name:     "positive test1",
			method:   http.MethodGet,
			url:      fmt.Sprintf("/%s", validKey),
			code:     307,
			location: location,
		},
		{
			name:     "negative test2",
			method:   http.MethodGet,
			url:      "/invalid",
			code:     400,
			location: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("%s%s", ts.URL, tt.url)
			res, err := client.Get(url)

			assert.Nil(t, err)
			defer res.Body.Close()

			assert.Equal(t, res.StatusCode, tt.code, "statuses should be equal")

			if tt.code == 307 {
				loc := res.Header.Get("location")
				assert.Equal(t, loc, tt.location, "statuses should be equal")
			}

		})
	}
}

func TestServer_createRedirectJSON(t *testing.T) {
	ts, _ := NewTestServer()
	defer ts.Close()

	client := http.Client{}
	url := fmt.Sprintf("%s/api/shorten", ts.URL)

	type request struct {
		URL string `json:"url"`
	}

	type response struct {
		Result string `json:"result"`
	}

	tests := []struct {
		name        string
		method      string
		contentType string
		data        request
		code        int
		result      response
	}{
		{
			name:        "positive test1",
			method:      http.MethodPost,
			contentType: "application/json",
			data:        request{URL: "http://example.com"},
			code:        201,
			result:      response{Result: "http://localhost:8080/2"},
		},
		{
			name:        "positive test2",
			method:      http.MethodPost,
			contentType: "application/json",
			data:        request{URL: "http://example.com"},
			code:        201,
			result:      response{Result: "http://localhost:8080/3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.data)
			assert.Nil(t, err)

			req, _ := http.NewRequest(tt.method, url, bytes.NewBuffer(data))
			req.Header.Set("Content-Type", tt.contentType)
			res, err := client.Do(req)
			assert.Nil(t, err)

			assert.Equal(t, res.StatusCode, tt.code, "statuses should be equal")

			defer res.Body.Close()
			if tt.code == 201 {
				bodyBytes, err := io.ReadAll(res.Body)
				assert.Nil(t, err)
				body := response{}
				assert.Nil(t, json.Unmarshal(bodyBytes, &body))
				assert.Equal(t, body, tt.result, "body should be equal")
			}
		})
	}
}
