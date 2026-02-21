package queue

const TypeBackupFile = "backup:file"
const TypeUploadFile = "upload:file"
const TypeDeleteFile = "delete:file"
const TypeGenerateThumb = "generate:thumb"

type BackupJob struct {
	Key    string `json:"key"`
	Bucket string `json:"bucket"`
}

type UploadJob struct {
	Key    string `json:"key"`
	Bucket string `json:"bucket"`
	Method string `json:"method,omitempty"`
}

type DeleteJob = BackupJob

type GenerateThumbJob = BackupJob
