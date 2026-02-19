package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/hibiken/asynq"
	"github.com/storage-gateway/src/config"
	"github.com/storage-gateway/src/queue"
	"github.com/storage-gateway/src/storage"
	"github.com/storage-gateway/src/storage/firebase_store"
	"github.com/storage-gateway/src/storage/s3_store"
)

func processS3Backup(ctx context.Context, original *storage.PutObject, bucket string, key string) error {
	s3ConfigPath, err := config.GetS3ConfigFromPath(bucket)
	if err != nil {
		return err
	}
	s3Client, err := s3_store.CreateClient(ctx, s3ConfigPath)
	if err != nil {
		return err
	}
	return s3Client.Put(ctx, bucket, key, original.Body, &storage.PutOptions{ContentType: original.ContentType, Metadata: original.Metadata})
}

func processFirebaseBackup(ctx context.Context, original *storage.PutObject, bucket string, key string) error {
	firebaseConfigPath, projectId, bucketStr, err := config.GetFirebaseConfigFromPath(bucket)
	if err != nil {
		return err
	}
	firebaseClient, err := firebase_store.CreateClient(ctx, firebaseConfigPath, projectId)
	if err != nil {
		return err
	}

	return firebaseClient.Put(ctx, bucketStr, key, original.Body, storage.PutOptions{ContentType: original.ContentType, Metadata: original.Metadata})
}

func HandleBackupTask(ctx context.Context, t *asynq.Task) error {
	var payload queue.BackupJob
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}
	key, bucket := payload.Key, payload.Bucket
	primaryStore := s3_store.GetPrimaryStore()

	fmt.Println("Starting backup: ", key)

	original, err := primaryStore.Get(ctx, bucket, key)
	if err != nil {
		return err
	}

	defer original.Body.Close()

	// Buffer the body content to allow multiple reads
	bodyBytes, err := io.ReadAll(original.Body)
	if err != nil {
		return err
	}
	creds, err := config.GetAvailableSecrets(bucket)
	if err != nil {
		return err
	}

	for _, method := range creds {
		obj := &storage.PutObject{
			Body:        bytes.NewReader(bodyBytes),
			ContentType: original.ContentType,
			Metadata:    original.Metadata,
		}
		if method == "firebase" {
			if err = processFirebaseBackup(ctx, obj, bucket, key); err != nil {
				fmt.Printf("Error processing firebase backup: %s", err.Error())
			}
		} else {
			if err = processS3Backup(ctx, obj, bucket, key); err != nil {
				fmt.Printf("Error processing s3 backup: %s", err.Error())
			}
		}
	}

	fmt.Println("Backup done: ", key)

	return nil
}
