package services

import (
	"context"
)

type StorageExpected interface {
	Get(ctx context.Context, key string) (string, error)
	Add(ctx context.Context, url string, userID string) (string, error)
	AddByBatch(ctx context.Context, urls []string, userID string) ([]string, error)
	GetAllUserURLs(ctx context.Context, userID string) (map[string]string, error)
	Ping(ctx context.Context) error
}
