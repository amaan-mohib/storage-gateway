package config

import (
	"os"
	"path"
)

func getHomeDir() string {
	dir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return dir
}

var (
	StorageEndpoint = map[string]string{
		"key":          "STORAGE_ENDPOINT",
		"defaultValue": "http://localhost:8333",
	}
	StorageRegion = map[string]string{
		"key":          "STORAGE_REGION",
		"defaultValue": "us-east-1",
	}
	StorageAccessKey = map[string]string{
		"key":          "MINIO_ROOT_USER",
		"defaultValue": "admin",
	}
	StorageSecretKey = map[string]string{
		"key":          "MINIO_ROOT_PASSWORD",
		"defaultValue": "admin@123",
	}
	AdminAccessToken = map[string]string{
		"key":          "ADMIN_ACCESS_TOKEN",
		"defaultValue": "admin@123",
	}
	SecretsPath = map[string]string{
		"key":          "SECRETS_PATH",
		"defaultValue": path.Join(getHomeDir(), "projects", "storage-gateway", "secrets"),
	}
	MinioVolumeLocation = map[string]string{
		"key":          "MINIO_VOLUME_LOCATION",
		"defaultValue": path.Join(getHomeDir(), "minio", "data"),
	}
)
