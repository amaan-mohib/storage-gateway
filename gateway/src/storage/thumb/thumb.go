package thumb

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/storage-gateway/src/internal/service"
	"github.com/storage-gateway/src/queue"
	"github.com/storage-gateway/src/storage"
	"github.com/storage-gateway/src/storage/backup"
)

const ThumbExt string = ".thumb.jpg"

func GenerateThumb(ctx context.Context, fileHandler *service.FileService, job *queue.BackupJob) (*storage.GetObject, error) {
	key, bucket := job.Key, job.Bucket
	videoKey := strings.Replace(key, ThumbExt, "", 1)
	exists := fileHandler.Exists(ctx, bucket, videoKey)
	var videoFile *storage.GetObject
	var err error
	if exists {
		videoFile, err = fileHandler.GetFile(ctx, bucket, videoKey)
		if err != nil {
			return nil, err
		}
	} else {
		videoFile, err = backup.FetchFromBackup(ctx, &queue.BackupJob{Key: videoKey, Bucket: bucket})
		if err != nil {
			return nil, err
		}
	}

	now := time.Now().Unix()
	inFile, err := os.CreateTemp("", fmt.Sprintf("input-%d-*.mp4", now))
	if err != nil {
		return nil, err
	}
	defer os.Remove(inFile.Name())
	defer inFile.Close()

	inData, err := io.ReadAll(videoFile.Body)
	if err != nil {
		return nil, err
	}
	defer videoFile.Body.Close()

	if _, err := inFile.Write(inData); err != nil {
		return nil, err
	}

	outFile, err := os.CreateTemp("", fmt.Sprintf("output-%d-*.jpg", now))
	if err != nil {
		return nil, err
	}
	defer os.Remove(outFile.Name())
	defer outFile.Close()

	cmd := exec.Command("ffmpeg", "-y",
		"-ss", "00:00:01.000",
		"-i", inFile.Name(),
		"-vf", "scale='min(854,iw)':'min(480,ih)':force_original_aspect_ratio=decrease,scale=trunc(iw/2)*2:trunc(ih/2)*2",
		"-vframes", "1",
		outFile.Name(),
	)

	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(outFile.Name())
	if err != nil {
		return nil, err
	}

	size := int64(len(data))
	body := bytes.NewReader(data)
	contentType := "image/jpeg"

	err = fileHandler.Upload(ctx, bucket, key, body, &storage.PutOptions{
		ContentType:   contentType,
		ContentLength: size,
	})
	if err != nil {
		return nil, err
	}

	obj, err := fileHandler.GetFile(ctx, bucket, key)
	if err != nil {
		return nil, err
	}

	return obj, nil
}
