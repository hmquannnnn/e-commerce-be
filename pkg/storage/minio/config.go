package minio

import (
	"os"
)

type Config struct {
	Endpoint        string
	PublicEndpoint  string // externally reachable endpoint for presigned URLs (e.g. localhost:9000)
	AccessKey       string
	SecretAccessKey string
	UseSSL          bool
	Region          string
	AuthType        string // "static" (access key — MinIO/local) | "iam" (AWS S3 — credential tạm từ IAM role qua IMDS)
}

func NewConfigFromEnv() Config {
	endpoint := getEnv("MINIO_ENDPOINT", "localhost:9000")
	return Config{
		Endpoint:        endpoint,
		PublicEndpoint:  getEnv("MINIO_PUBLIC_ENDPOINT", "localhost:9000"),
		AccessKey:       getEnv("MINIO_ACCESS_KEY", "minio"),
		SecretAccessKey: getEnv("MINIO_SECRET_KEY", "minio123"),
		UseSSL:          getEnvAsBool("MINIO_USE_SSL", false),
		Region:          getEnv("MINIO_REGION", "us-east-1"),
		AuthType:        getEnv("MINIO_AUTH_TYPE", "static"),
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
