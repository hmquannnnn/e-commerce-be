package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	App        AppConfig
	Database   DatabaseConfig
	ProductSvc ProductServiceConfig
	UserSvc    UserServiceConfig
}

type AppConfig struct {
	Name               string
	Environment        string
	Port               string
	GRPCPort           string
	OrderExpiryEnabled bool
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

type ProductServiceConfig struct {
	URL         string
	GRPCAddress string
}

type UserServiceConfig struct {
	URL string
}

func Load() (*Config, error) {
	cfg := &Config{
		App: AppConfig{
			Name:               getEnv("APP_NAME", "order-service"),
			Environment:        getEnv("APP_ENV", "development"),
			Port:               getEnv("APP_PORT", "8085"),
			GRPCPort:           getEnv("GRPC_PORT", "9085"),
			OrderExpiryEnabled: getEnvAsBool("ORDER_EXPIRY_WORKER_ENABLED", true),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5435),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "order_db"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
			MaxConns: getEnvAsInt("DB_MAX_CONNS", 25),
			MaxIdle:  getEnvAsInt("DB_MAX_IDLE", 10),
		},
		ProductSvc: ProductServiceConfig{
			URL:         getEnv("PRODUCT_SERVICE_URL", "http://localhost:8083"),
			GRPCAddress: getEnv("PRODUCT_SERVICE_GRPC_ADDRESS", "localhost:9083"),
		},
		UserSvc: UserServiceConfig{
			URL: getEnv("USER_SERVICE_URL", "http://localhost:8081"),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

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
	if c.ProductSvc.URL == "" {
		return fmt.Errorf("PRODUCT_SERVICE_URL is required")
	}
	if c.ProductSvc.GRPCAddress == "" {
		return fmt.Errorf("PRODUCT_SERVICE_GRPC_ADDRESS is required")
	}
	if c.UserSvc.URL == "" {
		return fmt.Errorf("USER_SERVICE_URL is required")
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

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
