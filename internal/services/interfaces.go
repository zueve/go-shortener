package services

import (
	"context"
	"fmt"
)

type StorageExpected interface {
	Get(ctx context.Context, key string) (string, error)
	Add(ctx context.Context, url string, userID string) (string, error)
	AddByBatch(ctx context.Context, urls []string, userID string) ([]string, error)
	GetAllUserURLs(ctx context.Context, userID string) (map[string]string, error)
	Ping(ctx context.Context) error
}

type LinkExistError struct {
	Key string
	Err error
}

func NewLinkExistError(key string, err error) error {
	return &LinkExistError{
		Key: key,
		Err: err,
	}
}

func (e *LinkExistError) Error() string {
	return fmt.Sprintf("Link Already Exist %s, %s", e.Key, e.Err)
}

func (e *LinkExistError) Unwrap() error {
	return e.Err
}
