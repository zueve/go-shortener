package storage

import (
	"errors"
	"strconv"
	"sync"
)

type PersistentStorageExpected interface {
	Load() (map[string]string, error)
	Add(key string, val string) error
}

type Storage struct {
	sync.RWMutex
	links   map[string]string
	counter int
	storage PersistentStorageExpected
}

func New(persistent PersistentStorageExpected) (*Storage, error) {
	data, err := persistent.Load()
	if err != nil {
		return nil, err
	}
	return &Storage{
		counter: getMaxKyeToInt(data),
		links:   data,
		storage: persistent,
	}, nil
}

func (c *Storage) Add(url string) string {
	c.Lock()
	defer c.Unlock()

	c.counter++
	key := strconv.Itoa(c.counter)
	c.links[key] = url
	c.storage.Add(key, url)

	return key
}

func (c *Storage) Get(key string) (string, error) {
	c.RLock()
	defer c.RUnlock()

	url, ok := c.links[key]
	if !ok {
		return "", errors.New("key not exist")
	}

	return url, nil
}

func getMaxKyeToInt(data map[string]string) int {
	maxVal := 1
	for keyStr := range data {
		keyInt, err := strconv.Atoi(keyStr)
		if err == nil && maxVal < keyInt {
			maxVal = keyInt
		}
	}
	return maxVal
}
