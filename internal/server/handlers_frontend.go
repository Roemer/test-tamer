package server

import (
	"html/template"
	"log/slog"
	"net/http"
)

func (s *Server) handleFrontendIndex(w http.ResponseWriter, r *http.Request) {
	type pageData struct {
		Breadcrumbs []breadcrumbItem
	}
	data := pageData{
		Breadcrumbs: []breadcrumbItem{},
	}
	s.renderPage(w, r, "pages/home.html", data)
}

func (s *Server) handleFrontendProjects(w http.ResponseWriter, r *http.Request) {
	page := queryParseInt(r, "page", 1)
	pageSize := queryParseInt(r, "page-size", defaultPageSize)

	if r.Header.Get("HX-Request") == "true" {
		type partialData struct {
			ProjectsTable tableComponent
			Paging        *pagingData
		}

		totalProjects, err := s.config.Store.Repos().Project.Count(r.Context())
		if err != nil {
			slog.Error("failed to count projects", "error", err)
			http.Error(w, "Failed to load projects.", http.StatusInternalServerError)
			return
		}

		paging := newPaging("/projects", "#project-list", page, pageSize, totalProjects)

		projects, err := s.config.Store.Repos().Project.ListPaged(r.Context(), paging.Page, paging.PageSize)
		if err != nil {
			slog.Error("failed to list projects", "error", err)
			http.Error(w, "Failed to load projects.", http.StatusInternalServerError)
			return
		}

		data := partialData{
			ProjectsTable: tableComponent{
				Columns: []tableComponentColumn{
					{Header: "ID"},
					{Header: "Name"},
				},
				Rows: []tableComponentRow{},
			},
			Paging: paging,
		}
		for _, project := range projects {
			data.ProjectsTable.Rows = append(data.ProjectsTable.Rows, tableComponentRow{
				Cells: []tableComponentCell{
					{Value: project.PublicID.String()},
					{Value: project.Name},
				},
			})
		}
		s.renderPartial(w, "partials/projects/list.html", data)
		return
	}

	type pageData struct {
		Breadcrumbs []breadcrumbItem
		Page        int
		PageSize    int
	}
	data := pageData{
		Breadcrumbs: []breadcrumbItem{
			{Name: "Home", Link: "/", Active: false},
			{Name: "Projects", Link: "/projects", Active: true},
		},
		Page:     page,
		PageSize: pageSize,
	}
	s.renderPage(w, r, "pages/projects/list.html", data)
}

func (s *Server) handlePageNotFound(w http.ResponseWriter, r *http.Request) {
	s.renderErrorPage(w, http.StatusNotFound, "Not Found",
		"The page you're looking for doesn't exist.",
		"Requested path: "+r.URL.Path)
}

func (s *Server) renderPartial(w http.ResponseWriter, templateName string, data any) {
	tmpl, err := template.ParseFS(s.templatesFS, "components/*.html", templateName)
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
	tmpl, err := template.New("").Funcs(template.FuncMap{
		"appVersion": func() string {
			return s.config.Version
		},
	}).ParseFS(s.templatesFS, "layouts/base.html", "components/*.html", pageTemplate)
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
	tmpl, err := template.New("").Funcs(template.FuncMap{
		"appVersion": func() string {
			return s.config.Version
		},
	}).ParseFS(s.templatesFS, "layouts/base.html", "pages/error.html")
	if err != nil {
		slog.Error("failed to parse error template", "error", err)
		http.Error(w, statusText, statusCode)
		return
	}

	type pageData struct {
		Breadcrumbs []breadcrumbItem
		StatusCode  int
		StatusText  string
		Message     string
		Details     string
	}
	data := pageData{
		Breadcrumbs: []breadcrumbItem{},
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
