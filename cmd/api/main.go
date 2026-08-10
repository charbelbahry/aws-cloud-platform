package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charbelbahry/aws-cloud-platform/internal/config"
	"github.com/charbelbahry/aws-cloud-platform/internal/database"
	"github.com/charbelbahry/aws-cloud-platform/internal/handlers"
	"github.com/charbelbahry/aws-cloud-platform/internal/middleware"
	"github.com/charbelbahry/aws-cloud-platform/migrations"
)

func main() {
	// Configure structured JSON logging to stdout
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer func() { _ = db.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.Ping(ctx); err != nil {
		slog.Error("database ping failed", "error", err)
		os.Exit(1)
	}

	if err := db.RunMigrations(ctx, migrations.FS); err != nil {
		slog.Error("database migration failed", "error", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()

	serviceHandler := handlers.NewServiceHandler(db)
	serviceHandler.RegisterRoutes(mux)

	deploymentHandler := handlers.NewDeploymentHandler(db)
	deploymentHandler.RegisterRoutes(mux)

	healthHandler := handlers.NewHealthHandler(db)
	healthHandler.RegisterRoutes(mux)

	// Wrap mux with middleware chain
	handler := middleware.Logging(middleware.Recovery(mux))

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Listen for SIGINT / SIGTERM signals for graceful shutdown
	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("starting HTTP server", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-shutdownCtx.Done()
	slog.Info("shutting down HTTP server gracefully...")

	shutdownTimeoutCtx, cancelTimeout := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelTimeout()

	if err := server.Shutdown(shutdownTimeoutCtx); err != nil {
		slog.Error("server shutdown failed", "error", err)
	}

	slog.Info("server stopped cleanly")
}
