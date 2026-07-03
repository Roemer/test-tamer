package main

import (
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/roemer/gotaskr"
	"github.com/roemer/test-tamer/internal/store/postgres"
)

func main() {
	os.Exit(gotaskr.Execute())
}

func init() {
	gotaskr.Task("vendor:init", vendorInit)
	gotaskr.Task("db-down", dbDown)
	gotaskr.Task("db-up", dbUp)
}

////////////////////////////////////////////////////////////
// Tasks
////////////////////////////////////////////////////////////

func vendorInit() error {
	vendorDir := "internal/server/static/vendor"
	if err := os.MkdirAll(vendorDir, 0o755); err != nil {
		return fmt.Errorf("failed to create vendor dir: %w", err)
	}

	bootstrapVersion := "5.3.8"
	htmxVersion := "2.0.10"
	type file struct {
		Url  string
		Dest string
	}

	files := []file{
		// Bootstrap CSS and JS
		{
			Url:  fmt.Sprintf("https://cdn.jsdelivr.net/npm/bootstrap@%s/dist/css/bootstrap.min.css", bootstrapVersion),
			Dest: "bootstrap.min.css",
		},
		{
			Url:  fmt.Sprintf("https://cdn.jsdelivr.net/npm/bootstrap@%s/dist/js/bootstrap.bundle.min.js", bootstrapVersion),
			Dest: "bootstrap.bundle.min.js",
		},
		// htmx JS
		{
			Url:  fmt.Sprintf("https://cdn.jsdelivr.net/npm/htmx.org@%s/dist/htmx.min.js", htmxVersion),
			Dest: "htmx.min.js",
		},
	}

	// Process the files
	for _, f := range files {
		targetPath := fmt.Sprintf("%s/%s", vendorDir, f.Dest)
		if err := downloadFile(f.Url, targetPath); err != nil {
			return fmt.Errorf("failed to download file '%s': %w", f.Url, err)
		}
		// Calculate the SHA384 checksum of the downloaded file
		data, err := os.ReadFile(targetPath)
		if err != nil {
			return fmt.Errorf("failed to read file '%s': %w", targetPath, err)
		}
		hash := sha512.Sum384(data)
		shaString := "sha384-" + base64.StdEncoding.EncodeToString(hash[:])

		// Verify the integrity checksum inside the templates
		if err := vendorCheckIntegrity("internal/server/template/layout/base.html", f.Dest, shaString); err != nil {
			return err
		}
	}

	return nil
}

func dbDown() error {
	dbUrl, err := postgres.BuildDbUrlFromEnv()
	if err != nil {
		return fmt.Errorf("failed to build database URL: %w", err)
	}
	m := postgres.NewMigrator(dbUrl, os.DirFS("internal/store/postgres"))
	return m.MigrateDown()
}

func dbUp() error {
	dbUrl, err := postgres.BuildDbUrlFromEnv()
	if err != nil {
		return fmt.Errorf("failed to build database URL: %w", err)
	}
	m := postgres.NewMigrator(dbUrl, os.DirFS("internal/store/postgres"))
	return m.MigrateUp()
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

func vendorCheckIntegrity(templateFile string, dependency string, expectedChecksum string) error {
	// src or href = the dependency, integrity="<the sha checksum>"
	regexPattern := regexp.MustCompile(fmt.Sprintf(`(?:src|href)\s*=\s*["'].*%s["'].*integrity\s*=\s*["']([^"']+)["']`, dependency))
	fileContent, err := os.ReadFile(templateFile)
	if err != nil {
		return fmt.Errorf("failed to read template file '%s': %w", templateFile, err)
	}

	matches := regexPattern.FindAllStringSubmatch(string(fileContent), -1)
	if len(matches) == 0 {
		return fmt.Errorf("no integrity attribute found for dependency '%s' in template '%s'", dependency, templateFile)
	}

	actualChecksum := matches[0][1]
	if actualChecksum != expectedChecksum {
		return fmt.Errorf("integrity checksum mismatch for dependency '%s' in template '%s': expected '%s', got '%s'", dependency, templateFile, expectedChecksum, actualChecksum)
	}

	slog.Info("integrity checksum verified", "dependency", dependency, "template", templateFile)
	return nil
}

func createMigrator() (*migrate.Migrate, error) {
	dbUrl, err := postgres.BuildDbUrlFromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to build database URL: %w", err)
	}

	m, err := migrate.New("file://internal/store/postgres/migrations", dbUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrator: %w", err)
	}
	m.Log = postgres.DbMigrationConsoleLogger{}

	return m, nil
}
