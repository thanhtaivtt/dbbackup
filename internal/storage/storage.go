package storage

import (
	"context"
	"io"
	"time"
)

type Object struct {
	Key          string
	Size         int64
	LastModified time.Time
}

type Storage interface {
	Upload(ctx context.Context, key string, reader io.Reader) error
	List(ctx context.Context, prefix string) ([]Object, error)
	Delete(ctx context.Context, keys []string) error
	Name() string
}
