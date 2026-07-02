package frontend

import (
	"fmt"
	"html"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"

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

// Renders the page not found error
func (h *FrontendHandler) PageNotFound(w http.ResponseWriter, r *http.Request) {
	h.renderErrorPage(w, http.StatusNotFound, "Not Found",
		"The page you're looking for doesn't exist.",
		"Requested path: "+r.URL.Path)
}

// Renders an internal server error page with an optional error message
func (h *FrontendHandler) InternalServerError(w http.ResponseWriter, r *http.Request, text string, details string) {
	h.renderErrorPage(w, http.StatusInternalServerError, "Internal Server Error", text, details)
}

// Renders a partial template
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

// Renders a full page template with the base layout
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

// Renders an error page
func (h *FrontendHandler) renderErrorPage(w http.ResponseWriter, statusCode int, statusText, message string, details string) {
	tmpl, err := template.New("").Funcs(h.templateFuncs()).ParseFS(h.templatesFS, "layout/base.html", "page/error.html")
	if err != nil {
		slog.Error("failed to parse error template", "error", err)
		http.Error(w, statusText, statusCode)
		return
	}

	type pageData struct {
		Breadcrumbs component.Breadcrumbs
		StatusCode  int
		StatusText  string
		Message     string
		Details     string
	}
	data := pageData{
		Breadcrumbs: component.NewBreadcrumbs(),
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

// Functions to be used in templates
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

// HTMXAttrs generates a string of HTML attributes for HTMX requests.
func HTMXAttrs(attrs map[string]string) template.HTMLAttr {
	var b strings.Builder
	for k, v := range attrs {
		fmt.Fprintf(&b, `%s="%s" `,
			html.EscapeString(k),
			html.EscapeString(v),
		)
	}
	return template.HTMLAttr(b.String())
}
