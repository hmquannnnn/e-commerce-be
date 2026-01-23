package config

import (
	"fmt"
	"os"
)

type Config struct {
	App     AppConfig
	Storage StorageConfig
	JWT     JWTConfig
}

type AppConfig struct {
	Name        string
	Environment string
	Port        string
}

type StorageConfig struct {
	BucketName string
}

type JWTConfig struct {
	SecretKey string
}

func Load() (*Config, error) {
	cfg := &Config{
		App: AppConfig{
			Name:        getEnv("APP_NAME", "file-service"),
			Environment: getEnv("APP_ENV", "development"),
			Port:        getEnv("APP_PORT", "8082"),
		},
		Storage: StorageConfig{
			BucketName: getEnv("STORAGE_BUCKET_NAME", "app-files"),
		},
		JWT: JWTConfig{
			SecretKey: os.Getenv("JWT_SECRET"),
		},
	}

	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func validateConfig(cfg *Config) error {
	if cfg.JWT.SecretKey == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
