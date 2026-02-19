package backup

import (
	"context"
	"fmt"

	"github.com/storage-gateway/src/config"
	"github.com/storage-gateway/src/queue"
	"github.com/storage-gateway/src/storage"
	"github.com/storage-gateway/src/storage/firebase_store"
	"github.com/storage-gateway/src/storage/s3_store"
)

func getS3Backup(ctx context.Context, bucket string, key string) (*storage.GetObject, error) {
	s3ConfigPath, err := config.GetS3ConfigFromPath(bucket)
	if err != nil {
		return nil, err
	}
	s3Client, err := s3_store.CreateClient(ctx, s3ConfigPath)
	if err != nil {
		return nil, err
	}
	return s3Client.Get(ctx, bucket, key)
}

func getFirebaseBackup(ctx context.Context, bucket string, key string) (*storage.GetObject, error) {
	firebaseConfigPath, projectId, bucketStr, err := config.GetFirebaseConfigFromPath(bucket)
	if err != nil {
		return nil, err
	}
	firebaseClient, err := firebase_store.CreateClient(ctx, firebaseConfigPath, projectId)
	if err != nil {
		return nil, err
	}

	return firebaseClient.Get(ctx, bucketStr, key)
}

func GetBackup(ctx context.Context, method string, bucket string, key string) (*storage.GetObject, error) {
	if method == "firebase" {
		return getFirebaseBackup(ctx, bucket, key)
	}
	if method == "s3" {
		return getS3Backup(ctx, bucket, key)
	}
	return nil, fmt.Errorf("Not a valid credential file: %s", method)
}

func FetchFromBackup(ctx context.Context, job *queue.BackupJob) (*storage.GetObject, error) {
	key, bucket := job.Key, job.Bucket
	creds, err := config.GetAvailableSecrets(bucket)
	if err != nil {
		return nil, err
	}
	for _, method := range creds {
		obj, err := GetBackup(ctx, method, bucket, key)
		if err == nil {
			queue.EnqueueUpload(queue.UploadJob{
				Key:    key,
				Bucket: bucket,
				Method: method,
			})
			return obj, nil
		}
	}
	return nil, err
}
