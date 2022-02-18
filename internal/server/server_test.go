package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"github.com/stretchr/testify/assert"
	"github.com/zueve/go-shortener/internal/services"
	"github.com/zueve/go-shortener/internal/storage"
)

type TestServer struct {
	*httptest.Server
	storage *storage.Storage
	service services.Service
	db      *sqlx.DB
}

func NewTestServer(t *testing.T) TestServer {
	db, err := sqlx.Open("sqlite3", ":memory:")
	// db, err := sqlx.Open("pgx", "postgres://user:pass@localhost:5432/db")
	assert.Nil(t, err)

	err = storage.Migrate(db)
	assert.Nil(t, err)
	_, err = db.Exec("DELETE FROM link")
	assert.Nil(t, err)

	storageTest, err := storage.New(db)
	assert.Nil(t, err)
	serviceTest := services.New(storageTest)

	s, err := New(serviceTest)
	assert.Nil(t, err)

	r := chi.NewRouter()
	r.Use(ungzipHandle)
	r.Use(gzipHandle)
	r.Use(setCookieHandler)
	r.Post("/", s.createRedirect)
	r.Post("/api/shorten/batch", s.createRedirectByBatch)
	r.Post("/api/shorten", s.createRedirectJSON)
	r.Get("/{keyID}", s.redirect)
	r.Get("/user/urls", s.GetAllUserURLs)
	ts := httptest.NewServer(r)

	srv := TestServer{
		Server:  ts,
		storage: storageTest,
		service: serviceTest,
		db:      db,
	}

	return srv
}

func (s *TestServer) Close() {
	s.Server.Close()
	s.db.Close()
}

func TestServer_createRedirect(t *testing.T) {
	ts := NewTestServer(t)
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
	ts := NewTestServer(t)
	defer ts.Close()
	location := "https://example.com"
	validKey, err := ts.service.CreateRedirect(context.Background(), location, "1")
	assert.Nil(t, err)
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
	ts := NewTestServer(t)
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
			result:      response{Result: "http://localhost:8080/10"},
		},
		{
			name:        "positive test2",
			method:      http.MethodPost,
			contentType: "application/json",
			data:        request{URL: "http://example.com"},
			code:        201,
			result:      response{Result: "http://localhost:8080/11"},
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
			}
		})
	}
}

func TestServer_GetAllUserURLs(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	jar, _ := cookiejar.New(nil)

	client := http.Client{Jar: jar}
	assert := assert.New(t)

	type row struct {
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
	}
	type response struct {
		ShortURL string `json:"result"`
	}
	type request struct {
		URL string `json:"url"`
	}

	expected := []row{
		{
			ShortURL:    "http://localhost:8080/2",
			OriginalURL: "http://example.com",
		},
		{
			ShortURL:    "http://localhost:8080/3",
			OriginalURL: "http://example.com/3",
		},
	}

	// get empty list
	resp, err := client.Get(fmt.Sprintf("%s/user/urls", ts.URL))
	assert.Equal(http.StatusNoContent, resp.StatusCode, "invalid status")
	assert.Nil(err)

	bodyBytes, err := io.ReadAll(resp.Body)
	assert.Nil(err)
	defer resp.Body.Close()
	body := make([]row, 0)
	json.Unmarshal(bodyBytes, &body)
	assert.Equal(body, make([]row, 0), "body should be empty")

	for i := range expected {
		contentType := "application/json"
		url := fmt.Sprintf("%s/api/shorten", ts.URL)
		data := request{URL: expected[i].OriginalURL}
		dataByte, err := json.Marshal(data)
		assert.Nil(err)
		resp, err := client.Post(url, contentType, bytes.NewBuffer(dataByte))
		assert.Nil(err)
		assert.Equal(resp.StatusCode, 201, "statuses should be equal")
		bodyBytes, err = io.ReadAll(resp.Body)
		assert.Nil(err)
		defer resp.Body.Close()
		var body response
		err = json.Unmarshal(bodyBytes, &body)
		assert.Nil(err)
		expected[i].ShortURL = body.ShortURL
	}

	// get list
	resp, err = client.Get(fmt.Sprintf("%s/user/urls", ts.URL))
	assert.Nil(err)
	assert.Equal(http.StatusOK, resp.StatusCode, "invalid status")

	bodyBytes, err = io.ReadAll(resp.Body)
	assert.Nil(err)
	defer resp.Body.Close()

	body = make([]row, 0)
	json.Unmarshal(bodyBytes, &body)
	assert.Equal(body, expected, "body should be empty")
}
