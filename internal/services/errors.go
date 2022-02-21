package services

import (
	"errors"
	"fmt"
)

var ErrRowDeleted = errors.New("deleted record")

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
