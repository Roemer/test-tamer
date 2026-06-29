package server

import "net/http"

type APIHandlers struct {
	server *Server
}

func NewAPIHandlers(server *Server) *APIHandlers {
	return &APIHandlers{server: server}
}

func (h *APIHandlers) registerRoutes(router *http.ServeMux) {
	apiRouter := http.NewServeMux()
	h.setupRoutes(apiRouter)
	router.Handle("/api/", http.StripPrefix("/api", h.server.corsMiddleware(apiRouter)))
}

func (h *APIHandlers) setupRoutes(router *http.ServeMux) {
	router.HandleFunc("GET /ping", h.handlePing)
}

func (h *APIHandlers) handlePing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
