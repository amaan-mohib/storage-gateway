package firebase_store

import (
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
	firebase "firebase.google.com/go"
	internal "github.com/storage-gateway/src/storage"
	"github.com/storage-gateway/src/storage/optimizer"
	"google.golang.org/api/option"
)

type Filer struct {
	client    *firebase.App
	projectId string
}

func NewClient(client *firebase.App, projectId string) *Filer {
	return &Filer{
		client:    client,
		projectId: projectId,
	}
}

func (s *Filer) GetBucket(ctx context.Context, bucketStr string) (*storage.BucketHandle, error) {
	client, clientError := s.client.Storage(ctx)
	if clientError != nil {
		return nil, clientError
	}

	if bucketStr == "default" {
		return client.Bucket(fmt.Sprintf("%s.appspot.com", s.projectId))
	} else {
		bucket, err := client.Bucket(bucketStr)
		if err != nil {
			return nil, err
		}
		_, err = bucket.Attrs(ctx)
		if err != nil {
			err = bucket.Create(ctx, s.projectId, nil)
		}
		return bucket, err
	}
}

func (s *Filer) Put(ctx context.Context, bucketStr string, key string, r io.Reader, opts internal.PutOptions) error {
	bucket, err := s.GetBucket(ctx, bucketStr)

	if err != nil {
		return err
	}

	object := &internal.PutObject{
		ContentType: opts.ContentType,
		Metadata:    opts.Metadata,
		Body:        r,
	}
	object, err = optimizer.Optimize(object)
	if err != nil {
		return err
	}

	wc := bucket.Object(key).NewWriter(ctx)
	wc.ObjectAttrs.Metadata = object.Metadata

	if _, err = io.Copy(wc, object.Body); err != nil {
		return fmt.Errorf("io.Copy: %w", err)
	}
	// Data can continue to be added to the file until the writer is closed.
	if err := wc.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %w", err)
	}

	return nil
}

func (s *Filer) Get(ctx context.Context, bucketStr string, key string) (*internal.GetObject, error) {
	bucket, err := s.GetBucket(ctx, bucketStr)
	if err != nil {
		return nil, err
	}
	o := bucket.Object(key)
	attrs, err := o.Attrs(ctx)
	if err != nil {
		return nil, fmt.Errorf("object.Attrs: %w", err)
	}
	rc, err := o.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("Object(%q).NewReader: %w", key, err)
	}
	defer rc.Close()

	return &internal.GetObject{
		ContentType:   attrs.ContentType,
		Metadata:      attrs.Metadata,
		Body:          rc,
		ContentLength: attrs.Size,
		ETag:          attrs.Etag,
		LastModified:  attrs.Updated,
	}, nil
}

func (s *Filer) Delete(ctx context.Context, bucketStr string, key string) error {
	bucket, err := s.GetBucket(ctx, bucketStr)
	if err != nil {
		return err
	}
	o := bucket.Object(key)

	// Optional: set a generation-match precondition to avoid potential race
	// conditions and data corruptions. The request to delete the file is aborted
	// if the object's generation number does not match your precondition.
	attrs, err := o.Attrs(ctx)
	if err != nil {
		return fmt.Errorf("object.Attrs: %w", err)
	}
	o = o.If(storage.Conditions{GenerationMatch: attrs.Generation})

	return o.Delete(ctx)
}

func (s *Filer) Exists(ctx context.Context, bucketStr string, key string) bool {
	bucket, err := s.GetBucket(ctx, bucketStr)
	if err != nil {
		return false
	}
	o := bucket.Object(key)

	_, err = o.Attrs(ctx)
	if err != nil {
		return false
	}
	return true
}

func CreateClient(ctx context.Context, configPath string, projectId string) (*Filer, error) {
	opt := option.WithCredentialsFile(configPath)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, err
	}
	return NewClient(app, projectId), nil
}
