package server

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type Storage struct {
	mtx     sync.Mutex
	links   map[string]string
	counter int
}

func (c *Storage) Add(url string) string {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.counter++
	key := strconv.Itoa(c.counter)
	c.links[key] = url
	return key
}

func (c *Storage) Get(key string) (string, error) {
	url, ok := c.links[key]
	if ok == false {
		return "", errors.New("Key not exist")
	}
	return url, nil
}

var storage = Storage{
	counter: 1,
	links:   map[string]string{},
}


func CreateRedirect(w http.ResponseWriter, r *http.Request) {
	headerContentType := r.Header.Get("Content-Type")
	if headerContentType != "application/x-www-form-urlencoded" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		fmt.Println("invalid ContentType")
		return
	}
	r.ParseForm()
	url := r.FormValue("url")
	if url == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println("invalid url")
		return
	}
	key := storage.Add(url)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(key))
}

func Redirect(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Path, "/")
	url, err := storage.Get(key)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println("invalid key")
		return
	}
	http.Redirect(w, r, url, http.StatusMovedPermanently)
}

func RouteRedirect(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		Redirect(w, r)
	} else {
		CreateRedirect(w, r)
	}
}

func RunServer() {
	http.HandleFunc("/", RouteRedirect)
	http.ListenAndServe(":8080", nil)
}
