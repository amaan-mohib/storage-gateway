package config

import (
	"encoding/json"
	"os"
	"path"
)

type Config struct {
	Endpoint  string `json:"endpoint"`
	Region    string `json:"region"`
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
}

func GetSafeEnv(obj map[string]string) string {
	key := obj["key"]
	defaultValue := obj["defaultValue"]
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func LoadConfig() Config {
	return Config{
		Endpoint:  GetSafeEnv(StorageEndpoint),
		Region:    GetSafeEnv(StorageRegion),
		AccessKey: GetSafeEnv(StorageAccessKey),
		SecretKey: GetSafeEnv(StorageSecretKey),
	}
}

func GetAvailableSecrets(bucket string) ([]string, error) {
	secretsPath := GetSafeEnv(SecretsPath)
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

func GetFirebaseConfigFromPath(bucket string) (string, string, string, error) {
	secretsPath := GetSafeEnv(SecretsPath)
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

func GetS3ConfigFromPath(bucket string) (string, error) {
	secretsPath := GetSafeEnv(SecretsPath)
	s3ConfigPath := path.Join(secretsPath, bucket, "s3_credentials")
	if _, err := os.ReadFile(s3ConfigPath); err != nil {
		return "", err
	}
	return s3ConfigPath, nil
}
