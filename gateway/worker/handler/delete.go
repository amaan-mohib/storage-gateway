package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/storage-gateway/src/config"
	"github.com/storage-gateway/src/queue"
	"github.com/storage-gateway/src/storage/firebase_store"
	"github.com/storage-gateway/src/storage/s3_store"
)

func processS3Delete(ctx context.Context, bucket string, key string) error {
	s3ConfigPath, err := config.GetS3ConfigFromPath(bucket)
	if err != nil {
		return err
	}
	s3Client, err := s3_store.CreateClient(ctx, s3ConfigPath)
	if err != nil {
		return err
	}
	return s3Client.Delete(ctx, bucket, key)
}

func processFirebaseDelete(ctx context.Context, bucket string, key string) error {
	firebaseConfigPath, projectId, bucketStr, err := config.GetFirebaseConfigFromPath(bucket)
	if err != nil {
		return err
	}
	firebaseClient, err := firebase_store.CreateClient(ctx, firebaseConfigPath, projectId)
	if err != nil {
		return err
	}

	return firebaseClient.Delete(ctx, bucketStr, key)
}

func HandleDeleteTask(ctx context.Context, t *asynq.Task) error {
	var payload queue.DeleteJob
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}
	key, bucket := payload.Key, payload.Bucket
	primaryStore := s3_store.GetPrimaryStore()

	fmt.Println("Starting delete: ", key)

	err := primaryStore.Delete(ctx, bucket, key)
	if err != nil {
		return err
	}

	creds, err := config.GetAvailableSecrets(bucket)
	if err != nil {
		return err
	}

	for _, method := range creds {
		if method == "firebase" {
			if err = processFirebaseDelete(ctx, bucket, key); err != nil {
				fmt.Printf("Error processing firebase delete: %s", err.Error())
			}
		} else {
			if err = processS3Delete(ctx, bucket, key); err != nil {
				fmt.Printf("Error processing s3 delete: %s", err.Error())
			}
		}
	}

	fmt.Println("Delete done: ", key)

	return nil
}
