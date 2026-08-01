// Package config handles application configuration loading from environment variables
package config

import (
	"errors"
	"fmt"
	"os"
)

var ErrMissingDatabaseURL = errors.New("DATABASE_URL environment variable is required")

type Config struct {
	Port        string
	DatabaseURL string
	LogLevel    string
}

func Load() (*Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("configuration error %w", ErrMissingDatabaseURL)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	return &Config{
		Port:        port,
		DatabaseURL: dbURL,
		LogLevel:    logLevel,
	}, nil
}
