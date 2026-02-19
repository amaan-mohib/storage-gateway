package config

import "os"

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
