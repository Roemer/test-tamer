package server

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

//go:embed templates
var embeddedTemplates embed.FS

//go:embed static
var embeddedStatic embed.FS

type Server struct {
	mux         *http.ServeMux
	config      Config
	templatesFS fs.FS
	staticFS    fs.FS
}

func New(config Config) *Server {
	// Prepare FS for templates
	tmplFs, err := fs.Sub(embeddedTemplates, "templates")
	if err != nil {
		log.Fatalf("failed to init embedded templates FS: %v", err)
	}
	// Overwrite with a local fs if the directory exists (for easier development without rebuilding)
	tmplDir := "internal/server/templates/"
	if info, err := os.Stat(tmplDir); err == nil && info.IsDir() {
		if tmplFs, err = fs.Sub(os.DirFS(tmplDir), "."); err != nil {
			log.Fatalf("failed to init templates FS from %s: %v", tmplDir, err)
		}
	}

	// Prepare FS for static files
	staticFs, err := fs.Sub(embeddedStatic, "static")
	if err != nil {
		log.Fatalf("failed to init embedded static FS: %v", err)
	}
	// Overwrite with a local fs if the directory exists (for easier development without rebuilding)
	staticDir := "internal/server/static/"
	if info, err := os.Stat(staticDir); err == nil && info.IsDir() {
		if staticFs, err = fs.Sub(os.DirFS(staticDir), "."); err != nil {
			log.Fatalf("failed to init static FS from %s: %v", staticDir, err)
		}
	}

	// Create the server
	return &Server{
		mux:         http.NewServeMux(),
		config:      config,
		templatesFS: tmplFs,
		staticFS:    staticFs,
	}
}

func (s *Server) Start() {
	s.setupRoutes()

	httpServer := &http.Server{
		Addr:              s.config.Address,
		Handler:           s.mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Create a cancelable context that listens for interrupt signals
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Run the server in a worker
	go func() {
		slog.Info("starting HTTP server", "addr", s.config.Address)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("listen error", "error", err)
			stop()
		}
	}()

	// Wait for a shutdown signal
	<-ctx.Done()
	// Create a context with timeout for the shutdown process
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("HTTP server shutdown error", "error", err)
	}
}

func (s *Server) setupRoutes() {
	// Static resources
	s.mux.Handle("GET /static/", http.StripPrefix("/static", http.FileServer(http.FS(s.staticFS))))

	// Frontend
	s.mux.HandleFunc("GET /{$}", s.handleFrontendIndex)
	s.mux.HandleFunc("GET /projects", s.handleFrontendProjects)

	// Catch-all for 404 errors (must be last - matches anything not matched above)
	s.mux.HandleFunc("/", s.handlePageNotFound)
}
