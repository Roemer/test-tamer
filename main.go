package main

import (
	"flag"
	"log/slog"

	"github.com/roemer/test-tamer/internal/db"
	fstore "github.com/roemer/test-tamer/internal/fake/store"
	"github.com/roemer/test-tamer/internal/server"
	"github.com/roemer/test-tamer/internal/store"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
)

// Default if not build via CI and the -ldflags
var version = "dev"

func main() {
	verbose := flag.Bool("v", false, "Enable verbose logging")
	useFake := flag.Bool("fake", false, "Use fake in-memory store")
	flag.Parse()

	if *verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	slog.Info("starting test-tamer", "version", version)
	slog.Debug("verbose logging enabled")

	var serverStore store.Client
	if *useFake {
		slog.Info("using fake in-memory store")
		serverStore = fstore.New()
	} else {
		// Create the database instance
		dbUrl, err := db.BuildDbUrlFromEnv()
		if err != nil {
			slog.Error("failed to build database URL", "error", err)
			return
		}
		dbInst := db.New(dbUrl)

		// Run the migrations
		if err := dbInst.RunMigrations(); err != nil {
			slog.Error("failed to run migrations", "error", err)
			return
		}

		// Prepare the db-pool
		dbpool, err := dbInst.CreateDbPool()
		if err != nil {
			slog.Error("failed to create database connection pool", "error", err)
			return
		}
		defer dbpool.Close()

		serverStore = store.New(dbpool)
	}

	// Configure the server
	config := server.Config{
		Address: ":8080",
		Version: version,
		Store:   serverStore,
	}
	// Start HTTP server
	ser := server.New(config)
	ser.Start()
}
