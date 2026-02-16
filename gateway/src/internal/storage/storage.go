package storage

import (
	"context"
	"io"
)

type PutOptions struct {
	ContentType string
	Metadata    map[string]string
}

type Storage interface {
	Put(ctx context.Context, bucket string, key string, r io.Reader, opts PutOptions) error
	Get(ctx context.Context, bucket string, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, bucket string, key string) error
	Exists(ctx context.Context, bucket string, key string) bool
}
