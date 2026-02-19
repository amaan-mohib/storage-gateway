package s3_store

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/storage-gateway/src/storage"
	"github.com/storage-gateway/src/storage/optimizer"
)

type Filer struct {
	S3 *s3.Client
}

func NewClient(client *s3.Client) *Filer {
	return &Filer{
		S3: client,
	}
}

func (s *Filer) Put(ctx context.Context, bucket string, key string, r io.Reader, opts *storage.PutOptions) error {
	_, err := s.S3.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: &bucket,
	})
	if err != nil {
		println("Creating bucket: ", bucket)
		_, err := s.S3.CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket: &bucket,
		})
		if err != nil {
			return err
		}
	}
	object := &storage.PutObject{
		ContentType:   opts.ContentType,
		Metadata:      opts.Metadata,
		ContentLength: opts.ContentLength,
		Body:          r,
	}
	object, err = optimizer.Optimize(object)
	if err != nil {
		return err
	}
	_, err = s.S3.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(key),
		Body:          object.Body,
		ContentType:   aws.String(object.ContentType),
		Metadata:      object.Metadata,
		ContentLength: &object.ContentLength,
	})
	return err
}

func (s *Filer) Get(ctx context.Context, bucket string, key string) (*storage.GetObject, error) {
	out, err := s.S3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return &storage.GetObject{
		ContentType: *out.ContentType,
		Metadata:    out.Metadata,
		Body:        out.Body,
	}, nil
}

func (s *Filer) Delete(ctx context.Context, bucket string, key string) error {
	_, err := s.S3.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return err
}

func (s *Filer) Exists(ctx context.Context, bucket string, key string) bool {
	_, err := s.S3.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return false
	}
	return true
}

func CreateClient(ctx context.Context, configPath string) (*Filer, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedCredentialsFiles([]string{configPath}))
	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(cfg)
	return NewClient(client), nil
}
