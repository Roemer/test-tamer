package db

import (
	"embed"
	"errors"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func (db *DB) RunMigrations() error {
	slog.Info("starting database migrations")
	src, err := iofs.New(migrationFiles, "migrations")
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", src, db.dbUrl)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer m.Close()

	// Get the current active version
	currentVersion, _, err := m.Version()
	if err != nil {
		if err == migrate.ErrNilVersion {
			slog.Info("Database has no schema version yet")
		} else {
			return fmt.Errorf("failed to get current version: %w", err)
		}
	} else {
		slog.Info(fmt.Sprintf("Current schema version: %d", currentVersion))
	}

	m.Log = DbMigrationConsoleLogger{}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			slog.Info("no migrations to apply")
			return nil
		}
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
