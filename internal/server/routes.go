package server

import "net/http"

func (s *Server) registerRoutes() {
	// Frontend
	s.registerFrontendRoutes(s.router)

	// Api
	apiRouter := http.NewServeMux()
	s.registerAPIRoutes(apiRouter)
	s.router.Handle("/api/", http.StripPrefix("/api", CORS(apiRouter)))

	// Static
	s.router.Handle("GET /static/", http.StripPrefix("/static", http.FileServer(http.FS(s.staticFS))))
}

func (s *Server) registerFrontendRoutes(router *http.ServeMux) {
	router.HandleFunc("GET /{$}", s.frontend.Home)
	router.HandleFunc("GET /projects", s.frontend.Projects)
	router.HandleFunc("GET /projects/new", s.frontend.NewProjectForm)
	router.HandleFunc("POST /projects/new", s.frontend.CreateProject)
	router.HandleFunc("GET /projects/{public_id}/edit", s.frontend.EditProjectForm)
	router.HandleFunc("POST /projects/{public_id}/edit", s.frontend.UpdateProject)
}

func (s *Server) registerAPIRoutes(router *http.ServeMux) {
	router.HandleFunc("GET /ping", s.api.Ping)
}
