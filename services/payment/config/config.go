package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	App             AppConfig
	Database        DatabaseConfig
	PayOS           PayOSConfig
	OrderServiceURL string
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

type PayOSConfig struct {
	ClientID    string
	ApiKey      string
	ChecksumKey string
	ReturnURL   string
	CancelURL   string
}

func Load() (*Config, error) {
	cfg := &Config{
		App: AppConfig{
			Name:        getEnv("APP_NAME", "payment-service"),
			Environment: getEnv("APP_ENV", "development"),
			Port:        getEnv("APP_PORT", "8086"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5436),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "payment_db"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
			MaxConns: getEnvAsInt("DB_MAX_CONNS", 25),
			MaxIdle:  getEnvAsInt("DB_MAX_IDLE", 10),
		},
		PayOS: PayOSConfig{
			ClientID:    getEnv("PAYOS_CLIENT_ID", ""),
			ApiKey:      getEnv("PAYOS_API_KEY", ""),
			ChecksumKey: getEnv("PAYOS_CHECKSUM_KEY", ""),
			ReturnURL:   getEnv("PAYOS_RETURN_URL", "http://localhost:3000/en/payment/result"),
			CancelURL:   getEnv("PAYOS_CANCEL_URL", "http://localhost:3000/en/payment/result?state=cancel"),
		},
		OrderServiceURL: getEnv("ORDER_SERVICE_URL", "http://localhost:8085"),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) Validate() error {
	if c.Database.Host == "" || c.Database.User == "" || c.Database.DBName == "" {
		return fmt.Errorf("database config is required")
	}
	return nil
}

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
