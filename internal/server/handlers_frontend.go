package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/roemer/test-tamer/internal/repositories"
	"github.com/roemer/test-tamer/internal/server/components"
)

type FrontendHandlers struct {
	server *Server
}

func NewFrontendHandlers(server *Server) *FrontendHandlers {
	return &FrontendHandlers{server: server}
}

func (h *FrontendHandlers) registerRoutes(router *http.ServeMux) {
	router.HandleFunc("GET /{$}", h.handleFrontendIndex)
	router.HandleFunc("GET /projects", h.handleFrontendProjects)
	router.HandleFunc("DELETE /projects/{publicID}", h.handleFrontendProjectDelete)

	// Catch-all for 404 errors (must be last - matches anything not matched above)
	router.HandleFunc("/", h.handlePageNotFound)
}

func (h *FrontendHandlers) handleFrontendIndex(w http.ResponseWriter, r *http.Request) {
	type pageData struct {
		Breadcrumbs []components.BreadcrumbItem
	}
	data := pageData{
		Breadcrumbs: []components.BreadcrumbItem{},
	}
	h.server.renderPage(w, r, "pages/home.html", data)
}

func (h *FrontendHandlers) handleFrontendProjects(w http.ResponseWriter, r *http.Request) {
	page := queryParseInt(r, "page", 1)
	pageSize := queryParseInt(r, "page-size", components.DefaultPageSize)
	pageSize = min(pageSize, components.MaxPageSize)

	if r.Header.Get("HX-Request") == "true" {
		type partialData struct {
			ProjectsTable *components.Table
			Paging        *components.Paging
		}

		totalProjects, err := h.server.config.Store.Repos().Project.Count(r.Context())
		if err != nil {
			slog.Error("failed to count projects", "error", err)
			http.Error(w, "Failed to load projects.", http.StatusInternalServerError)
			return
		}

		paging := components.NewPaging("/projects", "#project-list", page, pageSize, totalProjects)

		projects, err := h.server.config.Store.Repos().Project.ListPaged(r.Context(), paging.Page, paging.PageSize)
		if err != nil {
			slog.Error("failed to list projects", "error", err)
			http.Error(w, "Failed to load projects.", http.StatusInternalServerError)
			return
		}

		data := partialData{
			ProjectsTable: &components.Table{
				Columns: []components.TableColumn{
					{Header: "ID"},
					{Header: "Name"},
				},
				Rows: []components.TableRow{},
			},
			Paging: paging,
		}
		for _, project := range projects {
			data.ProjectsTable.Rows = append(data.ProjectsTable.Rows, components.TableRow{
				Cells: []components.TableCell{
					{Value: project.PublicID.String()},
					{Value: project.Name},
				},
				Actions: []components.TableAction{
					{
						Label: "Delete",
						Class: "btn btn-danger btn-sm",
						Attrs: HTMXAttrs(map[string]string{
							"hx-delete":  fmt.Sprintf("/projects/%s?%s", project.PublicID.String(), paging.PageQueryPart(paging.Page)),
							"hx-target":  "#project-list",
							"hx-swap":    "innerHTML",
							"hx-confirm": fmt.Sprintf("Delete project '%s'?", project.Name),
						}),
					},
				},
			})
		}
		h.server.renderPartial(w, "partials/projects/list.html", data)
		return
	}

	type pageData struct {
		Breadcrumbs []components.BreadcrumbItem
		Page        int
		PageSize    int
	}
	data := pageData{
		Breadcrumbs: []components.BreadcrumbItem{
			{Name: "Home", Link: "/", Active: false},
			{Name: "Projects", Link: "/projects", Active: true},
		},
		Page:     page,
		PageSize: pageSize,
	}
	h.server.renderPage(w, r, "pages/projects/list.html", data)
}

func (h *FrontendHandlers) handleFrontendProjectDelete(w http.ResponseWriter, r *http.Request) {
	httpError := func(message string, code int) {
		http.Error(w, "Delete failed: "+message, code)
	}
	publicID, err := uuid.Parse(r.PathValue("publicID"))
	if err != nil {
		httpError("Invalid project identifier.", http.StatusBadRequest)
		return
	}
	if err := h.server.config.Store.Repos().Project.DeleteByPublicID(r.Context(), publicID); err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			httpError("Project not found.", http.StatusNotFound)
			return
		}
		slog.Error("failed to delete project", "public_id", publicID, "error", err)
		httpError("Error.", http.StatusInternalServerError)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		h.server.triggerToast(w, "success", "Project deleted successfully.")
		h.handleFrontendProjects(w, r)
		return
	}

	http.Redirect(w, r, "/projects", http.StatusSeeOther)
}

func (h *FrontendHandlers) handlePageNotFound(w http.ResponseWriter, r *http.Request) {
	h.server.renderErrorPage(w, http.StatusNotFound, "Not Found",
		"The page you're looking for doesn't exist.",
		"Requested path: "+r.URL.Path)
}

func (s *Server) renderPartial(w http.ResponseWriter, templateName string, data any) {
	tmpl, err := template.New("").Funcs(s.templateFuncs()).ParseFS(s.templatesFS, "components/*.html", templateName)
	if err != nil {
		slog.Error("failed to parse partial template", "template_name", templateName, "error", err)
		http.Error(w, "Failed to render content.", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, templateName, data); err != nil {
		slog.Error("failed to render partial template", "template_name", templateName, "error", err)
	}
}

func (s *Server) renderPage(w http.ResponseWriter, r *http.Request, pageTemplate string, data any) {
	tmpl, err := template.New("").Funcs(s.templateFuncs()).ParseFS(s.templatesFS, "layouts/base.html", "components/*.html", pageTemplate)
	if err != nil {
		slog.Error("failed to parse templates", "error", err)
		s.renderErrorPage(w, http.StatusInternalServerError, "Template Error", "Failed to render page.", "")
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		slog.Error("failed to render template", "error", err)
	}
}

func (s *Server) renderErrorPage(w http.ResponseWriter, statusCode int, statusText, message string, details string) {
	tmpl, err := template.New("").Funcs(s.templateFuncs()).ParseFS(s.templatesFS, "layouts/base.html", "pages/error.html")
	if err != nil {
		slog.Error("failed to parse error template", "error", err)
		http.Error(w, statusText, statusCode)
		return
	}

	type pageData struct {
		Breadcrumbs []components.BreadcrumbItem
		StatusCode  int
		StatusText  string
		Message     string
		Details     string
	}
	data := pageData{
		Breadcrumbs: []components.BreadcrumbItem{},
		StatusCode:  statusCode,
		StatusText:  statusText,
		Message:     message,
		Details:     details,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		slog.Error("failed to render error page", "error", err)
	}
}

func (s *Server) templateFuncs() template.FuncMap {
	return template.FuncMap{
		"appVersion": func() string {
			return s.config.Version
		},
		"subtitle": func() string {
			return s.config.Subtitle
		},
		"iterate": func(n int) []int {
			s := make([]int, n)
			for i := range s {
				s[i] = i + 1
			}
			return s
		},
	}
}

func (s *Server) triggerToast(w http.ResponseWriter, toastType, message string) {
	payload := map[string]any{
		"tt:toast": map[string]string{
			"type":    toastType,
			"message": message,
		},
	}
	b, _ := json.Marshal(payload)
	w.Header().Set("HX-Trigger", string(b))
}
