package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppPort            string
	AppEnv             string
	DBHost             string
	DBPort             string
	DBUser             string
	DBPassword         string
	DBName             string
	DBSSLMode          string
	JWTSecret          string
	JWTExpiresIn       time.Duration
	CORSAllowedOrigins []string
}

func Load() (*Config, error) {
	cfg := &Config{
		AppPort:      getEnv("APP_PORT", "8080"),
		AppEnv:       getEnv("APP_ENV", "development"),
		DBHost:       getEnv("DB_HOST", "localhost"),
		DBPort:       getEnv("DB_PORT", "5432"),
		DBUser:       getEnv("DB_USER", "postgres"),
		DBPassword:   getEnv("DB_PASSWORD", "postgres"),
		DBName:       getEnv("DB_NAME", "inventory_db"),
		DBSSLMode:    getEnv("DB_SSLMODE", "disable"),
		JWTSecret:    os.Getenv("JWT_SECRET"),
		JWTExpiresIn: 24 * time.Hour,
	}

	if _, err := strconv.Atoi(cfg.DBPort); err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %s", cfg.DBPort)
	}

	if len(cfg.JWTSecret) < 32 {
		return nil, errors.New("JWT_SECRET must be at least 32 characters long for security")
	}

	if expStr := os.Getenv("JWT_EXPIRES_IN"); expStr != "" {
		d, err := time.ParseDuration(expStr)
		if err != nil {
			return nil, fmt.Errorf("invalid JWT_EXPIRES_IN format: %w", err)
		}
		cfg.JWTExpiresIn = d
	}

	corsOrigins := getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:3000")
	parts := strings.Split(corsOrigins, ",")
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			cfg.CORSAllowedOrigins = append(cfg.CORSAllowedOrigins, trimmed)
		}
	}

	return cfg, nil
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}

