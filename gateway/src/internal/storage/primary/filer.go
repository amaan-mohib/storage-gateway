package primary

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/storage-gateway/src/internal/storage"
)

type Filer struct {
	client *s3.Client
}

func NewClient(client *s3.Client) *Filer {
	return &Filer{
		client: client,
	}
}

func (s *Filer) Put(ctx context.Context, bucket string, key string, r io.Reader, opts storage.PutOptions) error {
	_, bucketErr := s.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: &bucket,
	})
	if bucketErr != nil {
		println("Creating bucket: ", bucket)
		_, err := s.client.CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket: &bucket,
		})
		if err != nil {
			return err
		}
	}
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        r,
		ContentType: aws.String(opts.ContentType),
		Metadata:    opts.Metadata,
	})
	return err
}

func (s *Filer) Get(ctx context.Context, bucket string, key string) (io.ReadCloser, error) {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return out.Body, nil
}

func (s *Filer) Delete(ctx context.Context, bucket string, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return err
}

func (s *Filer) Exists(ctx context.Context, bucket string, key string) bool {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return false
	}
	return true
}
