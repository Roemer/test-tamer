package main

import (
	"log/slog"
	"os"

	"github.com/roemer/test-tamer/internal/server"
	"github.com/roemer/test-tamer/internal/store"
	"github.com/roemer/test-tamer/internal/store/memory"
	"github.com/roemer/test-tamer/internal/store/postgres"
)

// Default if not build via CI and the -ldflags
var version = "dev"

func main() {
	dbType, ok := os.LookupEnv("TEST_TAMER_DB_TYPE")
	if !ok {
		dbType = "memory"
	}

	var serverStore store.Client
	switch dbType {
	case "memory":
		serverStore = memory.New()
	case "postgres":
		// Create the database instance
		dbUrl, err := postgres.BuildDbUrlFromEnv()
		if err != nil {
			slog.Error("failed to build database URL", "error", err)
			return
		}
		// Run the migrations
		if err := postgres.NewMigrator(dbUrl, nil).MigrateUp(); err != nil {
			slog.Error("failed to run migrations", "error", err)
			return
		}
		// Create the pool
		pgPool, err := postgres.CreateDbPool(dbUrl)
		if err != nil {
			slog.Error("failed to create Postgres connection pool", "error", err)
			return
		}
		// Initialize the store
		serverStore = postgres.New(pgPool)
	default:
		serverStore = memory.New()
	}

	config := server.Config{
		Address: ":8080",
		Version: version,
		Store:   serverStore,
	}
	// Start HTTP server
	srv := server.New(config)
	srv.Start()
}
