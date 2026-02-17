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

func getFirebaseConfigFromPath(secretsPath string, bucket string) (string, string, string, error) {
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

func getS3ConfigFromPath(secretsPath string, bucket string) (string, error) {
	s3ConfigPath := path.Join(secretsPath, bucket, "s3_credentials")
	if _, err := os.ReadFile(s3ConfigPath); err != nil {
		return "", err
	}
	return s3ConfigPath, nil
}

func processS3Backup(ctx context.Context, original storage.Object, bucket string, key string, secretsPath string) error {
	s3ConfigPath, err := getS3ConfigFromPath(secretsPath, bucket)
	if err != nil {
		return err
	}
	s3Client, err := s3_store.CreateClient(ctx, s3ConfigPath)
	if err != nil {
		return err
	}
	return s3Client.Put(ctx, bucket, key, original.Body, storage.PutOptions{ContentType: original.ContentType, Metadata: original.Metadata})
}

func processFirebaseBackup(ctx context.Context, original storage.Object, bucket string, key string, secretsPath string) error {
	firebaseConfigPath, projectId, bucketStr, err := getFirebaseConfigFromPath(secretsPath, bucket)
	if err != nil {
		return err
	}
	firebaseClient, err := firebase_store.CreateClient(ctx, firebaseConfigPath, projectId)
	if err != nil {
		return err
	}

	return firebaseClient.Put(ctx, bucketStr, key, original.Body, storage.PutOptions{ContentType: original.ContentType, Metadata: original.Metadata})
}

func getS3Backup(ctx context.Context, bucket string, key string, secretsPath string) (storage.Object, error) {
	s3ConfigPath, err := getS3ConfigFromPath(secretsPath, bucket)
	if err != nil {
		return storage.Object{}, err
	}
	s3Client, err := s3_store.CreateClient(ctx, s3ConfigPath)
	if err != nil {
		return storage.Object{}, err
	}
	return s3Client.Get(ctx, bucket, key)
}

func getFirebaseBackup(ctx context.Context, bucket string, key string, secretsPath string) (storage.Object, error) {
	firebaseConfigPath, projectId, bucketStr, err := getFirebaseConfigFromPath(secretsPath, bucket)
	if err != nil {
		return storage.Object{}, err
	}
	firebaseClient, err := firebase_store.CreateClient(ctx, firebaseConfigPath, projectId)
	if err != nil {
		return storage.Object{}, err
	}

	return firebaseClient.Get(ctx, bucketStr, key)
}

func ProcessBackup(job BackupJob) error {
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

	secretsPath := config.GetSafeEnv("SECRETS_PATH", "/home/amaan/projects/storage-gateway/secrets")

	// Create a copy of the object with the buffered body for S3 backup
	s3Obj := storage.Object{
		Body:        io.NopCloser(bytes.NewReader(bodyBytes)),
		ContentType: original.ContentType,
		Metadata:    original.Metadata,
	}
	if err = processS3Backup(ctx, s3Obj, bucket, key, secretsPath); err != nil {
		fmt.Printf("Error processing s3 backup: %s", err.Error())
	}

	// Create a copy of the object with the buffered body for Firebase backup
	fbObj := storage.Object{
		Body:        io.NopCloser(bytes.NewReader(bodyBytes)),
		ContentType: original.ContentType,
		Metadata:    original.Metadata,
	}
	if err = processFirebaseBackup(ctx, fbObj, bucket, key, secretsPath); err != nil {
		fmt.Printf("Error processing firebase backup: %s", err.Error())
	}

	return nil
}

func FetchFromBackup(ctx context.Context, job BackupJob) (storage.Object, error) {
	key, bucket := job.Key, job.Bucket
	secretsPath := config.GetSafeEnv("SECRETS_PATH", "/home/amaan/projects/storage-gateway/secrets")
	obj, err := getFirebaseBackup(ctx, bucket, key, secretsPath)
	if err == nil {
		// queue this
		if err2 := uploadToPrimary(ctx, job, "firebase"); err2 != nil {
			log.Printf("fb 2 p: %s", err2.Error())
		}
		return obj, nil
	} else {
		log.Printf("fb: %s", err.Error())

	}
	obj, err = getS3Backup(ctx, bucket, key, secretsPath)
	if err == nil {
		// queue this
		uploadToPrimary(ctx, job, "s3")
		return obj, nil
	}
	return storage.Object{}, err
}

func uploadToPrimary(ctx context.Context, job BackupJob, method string) error {
	key, bucket := job.Key, job.Bucket
	secretsPath := config.GetSafeEnv("SECRETS_PATH", "/home/amaan/projects/storage-gateway/secrets")
	primaryStore := s3_store.GetPrimaryStore()
	var object storage.Object
	var err error
	if method == "firebase" {
		object, err = getFirebaseBackup(ctx, bucket, key, secretsPath)
	} else {
		object, err = getS3Backup(ctx, bucket, key, secretsPath)
	}
	if err != nil {
		return err
	}
	data, err := io.ReadAll(object.Body)
	if err != nil {
		return err
	}
	defer object.Body.Close()

	return primaryStore.Put(ctx, bucket, key, bytes.NewReader(data), storage.PutOptions{
		ContentType:   object.ContentType,
		Metadata:      object.Metadata,
		ContentLength: int64(len(data)),
	})
}
