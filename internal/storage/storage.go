package storage

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"sync"
)

type Row struct {
	UserID    string
	OriginURL string
	Key       string
}

type PersistentStorageExpected interface {
	Load() ([]Row, error)
	Add(val Row) error
	Close() error
}

type Storage struct {
	sync.RWMutex
	links   map[string]Row
	counter int
	storage PersistentStorageExpected
	db      *sql.DB
}

func New(persistent PersistentStorageExpected, db *sql.DB) (*Storage, error) {
	data, err := persistent.Load()
	dataMap := make(map[string]Row)
	for i := range data {
		dataMap[data[i].Key] = data[i]
	}
	if err != nil {
		return nil, err
	}

	return &Storage{
		counter: getMaxKeyToInt(data),
		links:   dataMap,
		storage: persistent,
		db:      db,
	}, nil
}

func (c *Storage) Ping(ctx context.Context) error {
	return c.db.PingContext(ctx)
}

func (c *Storage) Add(url string, userID string) string {
	c.Lock()
	defer c.Unlock()

	c.counter++
	key := strconv.Itoa(c.counter)
	row := Row{OriginURL: url, UserID: userID, Key: key}
	c.links[key] = row
	c.storage.Add(row)

	return key
}

func (c *Storage) Get(key string) (string, error) {
	c.RLock()
	defer c.RUnlock()

	row, ok := c.links[key]
	if !ok {
		return "", errors.New("key not exist")
	}

	return row.OriginURL, nil
}

func (c *Storage) GetAllUserURLs(userID string) map[string]string {
	data := make(map[string]string)
	for _, r := range c.links {
		if r.UserID == userID {
			data[r.Key] = r.OriginURL
		}
	}
	return data
}

func getMaxKeyToInt(data []Row) int {
	maxVal := 1
	for i := range data {
		keyStr := data[i].Key
		keyInt, err := strconv.Atoi(keyStr)
		if err == nil && maxVal < keyInt {
			maxVal = keyInt
		}
	}
	return maxVal
}
