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

	"github.com/roemer/test-tamer/internal/handler/api"
	"github.com/roemer/test-tamer/internal/handler/frontend"
)

//go:embed static
var embeddedStatic embed.FS

//go:embed template
var embeddedTemplate embed.FS

type Server struct {
	router     *http.ServeMux
	config     Config
	api        *api.APIHandler
	frontend   *frontend.FrontendHandler
	staticFS   fs.FS
	templateFS fs.FS
}

func New(config Config) *Server {
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

	// Prepare FS for template
	tmplFs, err := fs.Sub(embeddedTemplate, "template")
	if err != nil {
		log.Fatalf("failed to init embedded template FS: %v", err)
	}
	// Overwrite with a local fs if the directory exists (for easier development without rebuilding)
	tmplDir := "internal/server/template/"
	if info, err := os.Stat(tmplDir); err == nil && info.IsDir() {
		if tmplFs, err = fs.Sub(os.DirFS(tmplDir), "."); err != nil {
			log.Fatalf("failed to init template FS from %s: %v", tmplDir, err)
		}
	}

	// Create the server
	srv := &Server{
		router:     http.NewServeMux(),
		config:     config,
		api:        api.NewAPIHandler(),
		frontend:   frontend.NewFrontendHandler(tmplFs, config.Version, config.Store),
		staticFS:   staticFs,
		templateFS: tmplFs,
	}

	srv.registerRoutes()

	return srv
}

func (s *Server) Start() {
	httpServer := &http.Server{
		Addr:              s.config.Address,
		Handler:           s.router,
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
