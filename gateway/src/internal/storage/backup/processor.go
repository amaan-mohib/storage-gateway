package backup

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path"

	"github.com/storage-gateway/src/internal/config"
	"github.com/storage-gateway/src/internal/storage"
	"github.com/storage-gateway/src/internal/storage/firebase_store"
	"github.com/storage-gateway/src/internal/storage/s3_store"
)

var secretsPath = config.GetSafeEnv(config.SecretsPath)

func getAvailableSecrets(bucket string) ([]string, error) {
	entries, err := os.ReadDir(path.Join(secretsPath, bucket))
	if err != nil {
		return nil, err
	}

	secrets := []string{}
	for _, dir := range entries {
		if dir.IsDir() {
			continue
		}
		switch dir.Name() {
		case "firebase.json":
			secrets = append(secrets, "firebase")
		case "s3_credentials":
			secrets = append(secrets, "s3")
		}
	}
	return secrets, err
}

func getFirebaseConfigFromPath(bucket string) (string, string, string, error) {
	firebaseConfigPath := path.Join(secretsPath, bucket, "firebase.json")
	firebaseFile, err := os.ReadFile(firebaseConfigPath)
	if err != nil {
		return "", "", "", err
	}
	var m map[string]any
	if err = json.Unmarshal(firebaseFile, &m); err != nil {
		return "", "", "", err
	}
	var projectId string = m["project_id"].(string)
	bucketStr := "default"
	if bucketName, ok := m["bucket"]; ok && bucketName != nil {
		bucketStr = bucketName.(string)
	}
	return firebaseConfigPath, projectId, bucketStr, err
}

func getS3ConfigFromPath(bucket string) (string, error) {
	s3ConfigPath := path.Join(secretsPath, bucket, "s3_credentials")
	if _, err := os.ReadFile(s3ConfigPath); err != nil {
		return "", err
	}
	return s3ConfigPath, nil
}

func processS3Backup(ctx context.Context, original *storage.PutObject, bucket string, key string) error {
	s3ConfigPath, err := getS3ConfigFromPath(bucket)
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
	firebaseConfigPath, projectId, bucketStr, err := getFirebaseConfigFromPath(bucket)
	if err != nil {
		return err
	}
	firebaseClient, err := firebase_store.CreateClient(ctx, firebaseConfigPath, projectId)
	if err != nil {
		return err
	}

	return firebaseClient.Put(ctx, bucketStr, key, original.Body, storage.PutOptions{ContentType: original.ContentType, Metadata: original.Metadata})
}

func getS3Backup(ctx context.Context, bucket string, key string) (*storage.GetObject, error) {
	s3ConfigPath, err := getS3ConfigFromPath(bucket)
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
	firebaseConfigPath, projectId, bucketStr, err := getFirebaseConfigFromPath(bucket)
	if err != nil {
		return nil, err
	}
	firebaseClient, err := firebase_store.CreateClient(ctx, firebaseConfigPath, projectId)
	if err != nil {
		return nil, err
	}

	return firebaseClient.Get(ctx, bucketStr, key)
}

func ProcessBackup(job *BackupJob) error {
	ctx := context.Background()
	key, bucket := job.Key, job.Bucket
	primaryStore := s3_store.GetPrimaryStore()

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
	creds, err := getAvailableSecrets(bucket)
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

	return nil
}

func getBackup(ctx context.Context, method string, bucket string, key string) (*storage.GetObject, error) {
	if method == "firebase" {
		return getFirebaseBackup(ctx, bucket, key)
	}
	if method == "s3" {
		return getS3Backup(ctx, bucket, key)
	}
	return nil, fmt.Errorf("Not a valid credential file: %s", method)
}

func FetchFromBackup(ctx context.Context, job *BackupJob) (*storage.GetObject, error) {
	key, bucket := job.Key, job.Bucket
	creds, err := getAvailableSecrets(bucket)
	if err != nil {
		return nil, err
	}
	for _, method := range creds {
		obj, err := getBackup(ctx, method, bucket, key)
		if err == nil {
			// queue this
			if err2 := UploadToPrimary(ctx, job, method); err2 != nil {
				log.Printf("fb 2 p: %s", err2.Error())
			}
			return obj, nil
		}
	}
	return nil, err
}

func UploadToPrimary(ctx context.Context, job *BackupJob, method string) error {
	key, bucket := job.Key, job.Bucket
	primaryStore := s3_store.GetPrimaryStore()

	object, err := getBackup(ctx, method, bucket, key)
	if err != nil {
		return err
	}
	data, err := io.ReadAll(object.Body)
	if err != nil {
		return err
	}
	defer object.Body.Close()

	return primaryStore.Put(ctx, bucket, key, bytes.NewReader(data), &storage.PutOptions{
		ContentType:   object.ContentType,
		Metadata:      object.Metadata,
		ContentLength: int64(len(data)),
	})
}
