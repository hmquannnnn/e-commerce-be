package minio

import (
	"os"
)

type Config struct {
	Endpoint        string
	AccessKey       string
	SecretAccessKey string
	UseSSL          bool
	Region          string
}

func NewConfigFromEnv() Config {
	return Config{
		Endpoint:        getEnv("MINIO_ENDPOINT", "localhost:9000"),
		AccessKey:       getEnv("MINIO_ACCESS_KEY", "minio"),
		SecretAccessKey: getEnv("MINIO_SECRET_KEY", "minio123"),
		UseSSL:          getEnvAsBool("MINIO_USE_SSL", false),
		Region:          getEnv("MINIO_REGION", "us-east-1"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true"
	}
	return defaultValue
}
