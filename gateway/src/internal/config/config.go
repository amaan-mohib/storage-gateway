package config

import "os"

type Config struct {
	Endpoint  string `json:"endpoint"`
	Region    string `json:"region"`
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
}

func GetSafeEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func LoadConfig() Config {
	return Config{
		Endpoint:  GetSafeEnv("STORAGE_ENDPOINT", "http://localhost:8333"),
		Region:    GetSafeEnv("STORAGE_REGION", "us-east-1"),
		AccessKey: GetSafeEnv("STORAGE_ACCESS_KEY", "seaweed_access"),
		SecretKey: GetSafeEnv("STORAGE_SECRET_KEY", "seaweed_secret"),
	}
}
