package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/storage-gateway/src/processing"
	"github.com/storage-gateway/src/queue"
	"github.com/storage-gateway/src/storage"
	"github.com/storage-gateway/src/storage/s3_store"
)

func HandleGenerateThumbTask(ctx context.Context, t *asynq.Task) error {
	var payload queue.GenerateThumbJob
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}
	key, bucket := payload.Key, payload.Bucket
	thumbKey := fmt.Sprintf("%s%s", key, processing.ThumbExt)
	primaryStore := s3_store.GetPrimaryStore()

	fmt.Println("Starting thumbnail generation: ", thumbKey)

	videoFile, err := primaryStore.Get(ctx, bucket, key)
	if err != nil {
		return err
	}

	thumb, err := processing.GenerateThumb(ctx, videoFile)
	if err != nil {
		return err
	}
	defer videoFile.Body.Close()

	err = primaryStore.Put(
		ctx,
		bucket,
		thumbKey,
		thumb.Body,
		&storage.PutOptions{
			ContentType:   thumb.ContentType,
			Metadata:      thumb.Metadata,
			ContentLength: thumb.ContentLength,
		},
	)

	if err != nil {
		fmt.Println("!!! Thumbnail generation failed: ", thumbKey, " Error: ", err.Error())
	} else {
		fmt.Println("Thumbnail generation done: ", thumbKey)
	}

	return err
}
