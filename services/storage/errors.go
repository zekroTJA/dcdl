package storage

import "errors"

var (
	ErrNotFound     = errors.New("not found")
	ErrNotDirectory = errors.New("location is not a directory")
)
