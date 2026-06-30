package main

import (
	"github.com/roemer/test-tamer/internal/server"
	"github.com/roemer/test-tamer/internal/store"
	"github.com/roemer/test-tamer/internal/store/memory"
)

// Default if not build via CI and the -ldflags
var version = "dev"

func main() {
	var serverStore store.Client
	serverStore = memory.New()

	config := server.Config{
		Address: ":8080",
		Version: version,
		Store:   serverStore,
	}
	// Start HTTP server
	srv := server.New(config)
	srv.Start()
}
