package service

import (
	"context"
	"io"

	"github.com/storage-gateway/src/internal/storage"
)

type FileService struct {
	store storage.Storage
}

func NewFileService(store storage.Storage) *FileService {
	return &FileService{store: store}
}

func (s *FileService) Upload(ctx context.Context, bucket string, key string, r io.Reader, contentType string, metadata map[string]string) error {
	return s.store.Put(ctx, bucket, key, r, storage.PutOptions{
		ContentType: contentType,
		Metadata:    metadata,
	})
}

func (s *FileService) GetFile(ctx context.Context, bucket string, key string) (io.ReadCloser, error) {
	return s.store.Get(ctx, bucket, key)
}

func (s *FileService) Exists(ctx context.Context, bucket string, key string) bool {
	return s.store.Exists(ctx, bucket, key)
}
