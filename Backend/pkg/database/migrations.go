package database

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
)

func RunMigrations(ctx context.Context, db *sqlx.DB) error {
	directory, err := migrationsDirectory()
	if err != nil {
		return err
	}

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("configure migration dialect: %w", err)
	}
	if err := goose.UpContext(ctx, db.DB, directory); err != nil {
		return fmt.Errorf("run migrations from %s: %w", directory, err)
	}
	return nil
}

func migrationsDirectory() (string, error) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get migration working directory: %w", err)
	}

	directory := workingDirectory
	for range 8 {
		candidate := filepath.Join(directory, "migrations")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			break
		}
		directory = parent
	}
	return "", fmt.Errorf("migrations directory not found from %s", workingDirectory)
}
