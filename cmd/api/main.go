package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/charbelbahry/aws-cloud-platform/internal/config"
	"github.com/charbelbahry/aws-cloud-platform/internal/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.Ping(ctx); err != nil {
		log.Fatalf("Database connection check failed: %v", err)
	}

	fmt.Println("Successfully connected and pinged PostgreSQL!")
}
