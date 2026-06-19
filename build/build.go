package main

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/roemer/gotaskr"
	"github.com/roemer/test-tamer/internal/db"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	os.Exit(gotaskr.Execute())
}

func init() {
	gotaskr.Task("init:vendor", initVendor)
	gotaskr.Task("db-down", dbDown)
	gotaskr.Task("db-up", dbUp)
}

////////////////////////////////////////////////////////////
// Tasks
////////////////////////////////////////////////////////////

func initVendor() error {
	vendorDir := "internal/server/static/vendor"
	if err := os.MkdirAll(vendorDir, 0o755); err != nil {
		return fmt.Errorf("failed to create vendor dir: %w", err)
	}

	// Bootstrap
	{
		version := "5.3.8"
		cssFile := fmt.Sprintf("https://cdn.jsdelivr.net/npm/bootstrap@%s/dist/css/bootstrap.min.css", version)
		jsFile := fmt.Sprintf("https://cdn.jsdelivr.net/npm/bootstrap@%s/dist/js/bootstrap.bundle.min.js", version)

		// Download the files
		if err := downloadFile(cssFile, fmt.Sprintf("%s/bootstrap.min.css", vendorDir)); err != nil {
			return fmt.Errorf("failed to download CSS: %w", err)
		}
		if err := downloadFile(jsFile, fmt.Sprintf("%s/bootstrap.bundle.min.js", vendorDir)); err != nil {
			return fmt.Errorf("failed to download JS: %w", err)
		}
	}

	// htmx
	{
		version := "2.0.10"
		jsFile := fmt.Sprintf("https://cdn.jsdelivr.net/npm/htmx.org@%s/dist/htmx.min.js", version)

		// Download the file
		if err := downloadFile(jsFile, fmt.Sprintf("%s/htmx.min.js", vendorDir)); err != nil {
			return fmt.Errorf("failed to download htmx: %w", err)
		}
	}

	return nil
}

func dbDown() error {
	m, err := createMigrator()
	if err != nil {
		return err
	}
	if err := m.Down(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			slog.Info("no migrations to apply")
			return nil
		}
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	return nil
}

func dbUp() error {
	m, err := createMigrator()
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			slog.Info("no migrations to apply")
			return nil
		}
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	return nil
}

////////////////////////////////////////////////////////////
// Internal functions
////////////////////////////////////////////////////////////

func downloadFile(url, dest string) error {
	client := http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download file '%s': %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download file '%s'. Status code: %d", url, resp.StatusCode)
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read body: %w", err)
	}
	if err := os.WriteFile(dest, bodyBytes, 0644); err != nil {
		return fmt.Errorf("failed to write file '%s': %w", dest, err)
	}
	return nil
}

func createMigrator() (*migrate.Migrate, error) {
	dbUrl, err := db.BuildDbUrlFromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to build database URL: %w", err)
	}

	m, err := migrate.New("file://internal/db/migrations", dbUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrator: %w", err)
	}
	m.Log = db.DbMigrationConsoleLogger{}

	return m, nil
}
