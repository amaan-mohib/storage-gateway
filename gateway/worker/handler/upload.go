package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/hibiken/asynq"
	"github.com/storage-gateway/src/processing"
	"github.com/storage-gateway/src/queue"
	"github.com/storage-gateway/src/storage"
	"github.com/storage-gateway/src/storage/s3_store"
)

func HandleUploadTask(ctx context.Context, t *asynq.Task) error {
	var payload queue.UploadJob
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}
	key, bucket, method := payload.Key, payload.Bucket, payload.Method
	primaryStore := s3_store.GetPrimaryStore()

	fmt.Println("Starting copy upload: ", key)

	object, err := processing.GetBackup(ctx, method, bucket, key)
	if err != nil {
		return err
	}
	data, err := io.ReadAll(object.Body)
	if err != nil {
		return err
	}
	defer object.Body.Close()

	if object.Metadata == nil {
		object.Metadata = map[string]string{}
	}
	object.Metadata["original-upload-date"] = object.LastModified.Format(time.RFC1123)

	err = primaryStore.Put(ctx, bucket, key, bytes.NewReader(data), &storage.PutOptions{
		ContentType:   object.ContentType,
		Metadata:      object.Metadata,
		ContentLength: int64(len(data)),
	})

	if err != nil {
		fmt.Println("!!! Copy upload failed: ", key, " Error: ", err.Error())
	} else {
		fmt.Println("Copy upload done: ", key)
	}

	return err
}
