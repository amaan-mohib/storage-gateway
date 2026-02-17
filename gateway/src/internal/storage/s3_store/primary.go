package s3_store

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/storage-gateway/src/internal/config"
)

func GetPrimaryStore() *Filer {
	cfg := config.LoadConfig()

	options := s3.Options{
		Region:       cfg.Region,
		Credentials:  aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")),
		BaseEndpoint: aws.String(cfg.Endpoint),
		UsePathStyle: true,
	}

	s3Client := s3.New(options)

	return NewClient(s3Client)
}
