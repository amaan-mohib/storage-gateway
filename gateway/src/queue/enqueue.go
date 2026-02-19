package queue

import (
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
)

func EnqueueBackup(job BackupJob) error {
	payload, err := json.Marshal(job)
	if err != nil {
		return err
	}
	task := asynq.NewTask(TypeBackupFile, payload)

	_, err = asynqClient.Enqueue(task, asynq.MaxRetry(2), asynq.Timeout(10*time.Minute))
	return err
}

func EnqueueUpload(job UploadJob) error {
	payload, err := json.Marshal(job)
	if err != nil {
		return err
	}
	task := asynq.NewTask(TypeUploadFile, payload)

	_, err = asynqClient.Enqueue(task, asynq.MaxRetry(2), asynq.Timeout(10*time.Minute))
	return err
}
