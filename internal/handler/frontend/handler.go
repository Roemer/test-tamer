package frontend

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/roemer/test-tamer/internal/component"
	"github.com/roemer/test-tamer/internal/store"
)

type FrontendHandler struct {
	templatesFS fs.FS
	version     string
	store       store.Client
}

func NewFrontendHandler(templatesFS fs.FS, version string, store store.Client) *FrontendHandler {
	return &FrontendHandler{
		templatesFS: templatesFS,
		version:     version,
		store:       store,
	}
}

func (h *FrontendHandler) Home(w http.ResponseWriter, r *http.Request) {
	type pageData struct {
		Breadcrumbs []component.BreadcrumbItem
	}
	data := pageData{
		Breadcrumbs: []component.BreadcrumbItem{},
	}
	h.renderPage(w, r, "page/home.html", data)
}

func (h *FrontendHandler) renderPartial(w http.ResponseWriter, templateName string, data any) {
	tmpl, err := template.New("").Funcs(h.templateFuncs()).ParseFS(h.templatesFS, "component/*.html", templateName)
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

func (h *FrontendHandler) renderPage(w http.ResponseWriter, r *http.Request, pageTemplate string, data any) {
	_ = r
	tmpl, err := template.New("").Funcs(h.templateFuncs()).ParseFS(h.templatesFS, "layout/base.html", "component/*.html", pageTemplate)
	if err != nil {
		slog.Error("failed to parse templates", "error", err)
		h.renderErrorPage(w, http.StatusInternalServerError, "Template Error", "Failed to render page.", "")
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		slog.Error("failed to render template", "error", err)
	}
}

func (h *FrontendHandler) renderErrorPage(w http.ResponseWriter, statusCode int, statusText, message string, details string) {
	tmpl2, _ := json.MarshalIndent(h.templatesFS, "", "  ")
	fmt.Println("Templates FS:", string(tmpl2)) // Debugging line to print the templatesFS structure
	tmpl, err := template.New("").Funcs(h.templateFuncs()).ParseFS(h.templatesFS, "layout/base.html", "page/error.html")
	if err != nil {
		slog.Error("failed to parse error template", "error", err)
		http.Error(w, statusText, statusCode)
		return
	}

	type pageData struct {
		Breadcrumbs []component.BreadcrumbItem
		StatusCode  int
		StatusText  string
		Message     string
		Details     string
	}
	data := pageData{
		Breadcrumbs: []component.BreadcrumbItem{},
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

func (h *FrontendHandler) templateFuncs() template.FuncMap {
	return template.FuncMap{
		"appVersion": func() string {
			return h.version
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
