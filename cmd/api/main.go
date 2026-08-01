package main

import (
	"fmt"
	"log"

	"github.com/charbelbahry/aws-cloud-platform/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	fmt.Printf("Configuration loaded successfully:\n")
	fmt.Printf("	Port: %s\n", cfg.Port)
	fmt.Printf("	LogLevel: %s\n", cfg.LogLevel)
	fmt.Printf("	DatabaseURL: %s\n", cfg.DatabaseURL)
}
