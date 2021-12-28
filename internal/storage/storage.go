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

func New() Storage {
	return Storage{
		counter: 1,
		links:   map[string]string{},
	}
}
