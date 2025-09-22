package config

import (
	"os"
)

// Config holds application configuration
type Config struct {
	Port     string
	LogLevel string
	Database DatabaseConfig
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	DSN string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Port:     getEnv("PORT", "8080"),
		LogLevel: getEnv("LOG_LEVEL", "debug"),
		Database: DatabaseConfig{
			DSN: getEnv("MYSQL_DSN", "app:app@tcp(127.0.0.1:3306)/servicesdb?parseTime=true&charset=utf8mb4&collation=utf8mb4_0900_ai_ci"),
		},
	}
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
