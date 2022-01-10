package storage

import (
	"errors"
	"strconv"
	"sync"
)

type Storage struct {
	mtx     sync.Mutex
	links   map[string]string
	counter int
}

func New() *Storage {
	return &Storage{
		counter: 1,
		links:   map[string]string{},
	}
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
	if !ok {
		return "", errors.New("key not exist")
	}

	return url, nil
}
