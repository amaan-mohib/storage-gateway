package storage

import (
	"context"
	"io"
	"time"

	firebase "firebase.google.com/go"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type PutOptions struct {
	ContentType   string            `json:"contentType"`
	Metadata      map[string]string `json:"metadata"`
	ContentLength int64             `json:"contentLength"`
}

type Object[T any] struct {
	ContentType   string
	Metadata      map[string]string
	ContentLength int64
	Body          T
	ETag          string
	LastModified  time.Time
}

type PutObject = Object[io.Reader]

type GetObject = Object[io.ReadCloser]

type Client struct {
	S3       *s3.Client
	Firebase *firebase.App
}

type Storage interface {
	Put(ctx context.Context, bucket string, key string, r io.Reader, opts *PutOptions) error
	Get(ctx context.Context, bucket string, key string) (*GetObject, error)
	Delete(ctx context.Context, bucket string, key string) error
	Exists(ctx context.Context, bucket string, key string) bool
}
