package database

import (
	"context"
	"embed"
	"fmt"
	"sort"
)

func (db *DB) RunMigrations(ctx context.Context, migrationFS embed.FS) error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
	version TEXT PRIMARY KEY,
	applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
	);`

	if _, err := db.DB.ExecContext(ctx, createTableSQL); err != nil {
		return fmt.Errorf("creating schema_migrations table: %w", err)
	}

	entries, err := migrationFS.ReadDir(".")
	if err != nil {
		return fmt.Errorf("reading embedded migraiotns directoy: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && entry.Name() != "embed.go" {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)

	for _, file := range files {
		var exists bool
		checkSQL := `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1);`
		if err := db.DB.QueryRowContext(ctx, checkSQL, file).Scan(&exists); err != nil {
			return fmt.Errorf("checking migration status for %s: %w", file, err)
		}

		if exists {
			continue
		}

		content, err := migrationFS.ReadFile(file)
		if err != nil {
			return fmt.Errorf("reading migration file %s: %w", file, err)
		}

		tx, err := db.DB.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("beginning transaction for %s: %w", file, err)
		}

		if _, err := tx.ExecContext(ctx, string(content)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("executing migration %s: %w", file, err)
		}

		insertSQL := `INSERT INTO schema_migrations (version) VALUES ($1);`
		if _, err := tx.ExecContext(ctx, insertSQL, file); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("recording migration %s: %w", file, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("committing migration transaction for %s: %w", file, err)
		}
		fmt.Printf("Applied migration: %s\n", file)
	}

	return nil
}
