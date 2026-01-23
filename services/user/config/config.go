package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	JWT      JWTConfig
	// Redis    RedisConfig // Commented out - add back when you want to use Redis
}

type AppConfig struct {
	Name        string
	Environment string
	Port        string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
	MaxConns int
	MaxIdle  int
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type JWTConfig struct {
	SecretKey            string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		App: AppConfig{
			Name:        getEnv("APP_NAME", "user-service"),
			Environment: getEnv("APP_ENV", "development"),
			Port:        getEnv("APP_PORT", "50051"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "user_db"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
			MaxConns: getEnvAsInt("DB_MAX_CONNS", 25),
			MaxIdle:  getEnvAsInt("DB_MAX_IDLE", 10),
		},
		JWT: JWTConfig{
			SecretKey:            getEnv("JWT_SECRET", "your-secret-key-change-this"),
			AccessTokenDuration:  getEnvAsDuration("JWT_ACCESS_TOKEN_EXPIRE", 15*time.Minute),
			RefreshTokenDuration: getEnvAsDuration("JWT_REFRESH_TOKEN_EXPIRE", 168*time.Hour), // 7 days
		},
	}

	// Validate required fields
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Database.User == "" {
		return fmt.Errorf("database user is required")
	}
	if c.Database.DBName == "" {
		return fmt.Errorf("database name is required")
	}
	if c.JWT.SecretKey == "" || c.JWT.SecretKey == "your-secret-key-change-this" {
		return fmt.Errorf("JWT secret key must be set and changed from default")
	}
	return nil
}

// GetDatabaseDSN returns the database connection string
func (c *Config) GetDatabaseDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.SSLMode,
	)
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
