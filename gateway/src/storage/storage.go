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

type PutObject struct {
	ContentType   string
	Metadata      map[string]string
	ContentLength int64
	Body          io.Reader
}

type GetObject struct {
	ContentType   string
	Metadata      map[string]string
	ContentLength int64
	Body          io.ReadCloser
	ETag          string
	LastModified  time.Time
}

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
