package service

import (
	"context"
	"io"

	"github.com/storage-gateway/src/storage"
)

type FileService struct {
	store storage.Storage
}

func NewFileService(store storage.Storage) *FileService {
	return &FileService{store: store}
}

func (s *FileService) Upload(ctx context.Context, bucket string, key string, r io.Reader, opts *storage.PutOptions) error {
	return s.store.Put(ctx, bucket, key, r, opts)
}

func (s *FileService) GetFile(ctx context.Context, bucket string, key string) (*storage.GetObject, error) {
	return s.store.Get(ctx, bucket, key)
}

func (s *FileService) Exists(ctx context.Context, bucket string, key string) bool {
	return s.store.Exists(ctx, bucket, key)
}
