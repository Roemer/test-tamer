package postgres

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

type Migrator struct {
	dbUrl string
	fs    fs.FS
}

func NewMigrator(dbUrl string, fs fs.FS) *Migrator {
	if fs == nil {
		fs = migrationFiles
	}
	return &Migrator{dbUrl: dbUrl, fs: fs}
}

func (m *Migrator) MigrateUp() error {
	migrator, err := m.createMigrator()
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer migrator.Close()
	if err := migrator.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			slog.Info("no migrations to apply")
			return nil
		}
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	return nil
}

func (m *Migrator) MigrateDown() error {
	migrator, err := m.createMigrator()
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer migrator.Close()
	if err := migrator.Down(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			slog.Info("no migrations to apply")
			return nil
		}
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	return nil
}

func (m *Migrator) createMigrator() (*migrate.Migrate, error) {
	src, err := iofs.New(m.fs, "migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to load migrations: %w", err)
	}
	migrator, err := migrate.NewWithSourceInstance("iofs", src, m.dbUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrator: %w", err)
	}
	migrator.Log = DbMigrationConsoleLogger{}
	slog.Info("starting database migrations")
	// Get the current active version
	currentVersion, _, err := migrator.Version()
	if err != nil {
		if err == migrate.ErrNilVersion {
			slog.Info("Database has no schema version yet")
		} else {
			return nil, fmt.Errorf("failed to get current version: %w", err)
		}
	} else {
		slog.Info(fmt.Sprintf("Current schema version: %d", currentVersion))
	}

	return migrator, nil
}
