package frontend

import (
	"errors"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"github.com/roemer/test-tamer/internal/component"
	"github.com/roemer/test-tamer/internal/model"
	"github.com/roemer/test-tamer/internal/store"
)

func (h *FrontendHandler) Projects(w http.ResponseWriter, r *http.Request) {
	page := queryParseInt(r, "page", 1)
	pageSize := queryParseInt(r, "page-size", component.DefaultPageSize)
	pageSize = min(pageSize, component.MaxPageSize)

	if r.Header.Get("HX-Request") == "true" {
		type partialData struct {
			ProjectsTable *component.Table
			Paging        *component.Paging
		}

		totalProjects, err := h.store.Stores().Projects.Count(r.Context())
		if err != nil {
			slog.Error("failed to count projects", "error", err)
			h.renderErrorPage(w, http.StatusInternalServerError, "Internal Server Error", "Failed to load projects.", "")
			return
		}
		paging := component.NewPaging("/projects", "#project-list", page, pageSize, totalProjects)
		projects, err := h.store.Stores().Projects.List(r.Context(), paging.Page, paging.PageSize)
		if err != nil {
			slog.Error("failed to list projects", "error", err)
			h.renderErrorPage(w, http.StatusInternalServerError, "Internal Server Error", "Failed to load projects.", "")
			return
		}
		rows := make([]component.TableRow, 0, len(projects))
		for _, project := range projects {
			rows = append(rows, component.TableRow{
				Cells: []component.TableCell{
					{Value: project.PublicID.String()},
					{Value: project.Name},
					{Value: project.Slug},
				},
				Actions: []component.TableAction{
					{
						Label: "Edit",
						Class: "btn btn-sm btn-outline-secondary",
						Attrs: template.HTMLAttr(`type="button" onclick="window.location.href='/projects/` + project.PublicID.String() + `/edit'"`),
					},
				},
			})
		}

		data := partialData{
			ProjectsTable: &component.Table{
				Columns: []component.TableColumn{
					{Header: "ID"},
					{Header: "Name"},
					{Header: "Slug"},
				},
				Rows: rows,
			},
			Paging: paging,
		}
		h.renderPartial(w, "partial/projects/list.html", data)
		return
	}

	type pageData struct {
		Breadcrumbs []component.BreadcrumbItem
		Page        int
		PageSize    int
	}
	data := pageData{
		Breadcrumbs: []component.BreadcrumbItem{
			{Name: "Home", Link: "/", Active: false},
			{Name: "Projects", Link: "/projects", Active: true},
		},
		Page:     page,
		PageSize: pageSize,
	}
	h.renderPage(w, r, "page/projects/list.html", data)
}

func (h *FrontendHandler) NewProjectForm(w http.ResponseWriter, r *http.Request) {
	h.renderProjectForm(w, r, projectFormPageData{
		Breadcrumbs: []component.BreadcrumbItem{
			{Name: "Home", Link: "/", Active: false},
			{Name: "Projects", Link: "/projects", Active: false},
			{Name: "New", Link: "/projects/new", Active: true},
		},
		Title:       "New Project",
		Heading:     "Create Project",
		FormAction:  "/projects/new",
		SubmitLabel: "Create Project",
	})
}

func (h *FrontendHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		slog.Error("failed to parse project create form", "error", err)
		h.renderProjectForm(w, r, projectFormPageData{
			Breadcrumbs: []component.BreadcrumbItem{
				{Name: "Home", Link: "/", Active: false},
				{Name: "Projects", Link: "/projects", Active: false},
				{Name: "New", Link: "/projects/new", Active: true},
			},
			Title:        "New Project",
			Heading:      "Create Project",
			FormAction:   "/projects/new",
			SubmitLabel:  "Create Project",
			ErrorMessage: "Failed to process form submission.",
		})
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	slug := normalizeSlug(r.FormValue("slug"))
	if slug == "" {
		slug = normalizeSlug(name)
	}
	if name == "" {
		h.renderProjectForm(w, r, projectFormPageData{
			Breadcrumbs: []component.BreadcrumbItem{
				{Name: "Home", Link: "/", Active: false},
				{Name: "Projects", Link: "/projects", Active: false},
				{Name: "New", Link: "/projects/new", Active: true},
			},
			Title:        "New Project",
			Heading:      "Create Project",
			FormAction:   "/projects/new",
			SubmitLabel:  "Create Project",
			Name:         name,
			Slug:         slug,
			ErrorMessage: "Project name is required.",
		})
		return
	}

	_, err := h.store.Stores().Projects.Create(r.Context(), model.Project{
		Name: name,
		Slug: slug,
	})
	if err != nil {
		slog.Error("failed to create project", "error", err)
		h.renderProjectForm(w, r, projectFormPageData{
			Breadcrumbs: []component.BreadcrumbItem{
				{Name: "Home", Link: "/", Active: false},
				{Name: "Projects", Link: "/projects", Active: false},
				{Name: "New", Link: "/projects/new", Active: true},
			},
			Title:        "New Project",
			Heading:      "Create Project",
			FormAction:   "/projects/new",
			SubmitLabel:  "Create Project",
			Name:         name,
			Slug:         slug,
			ErrorMessage: "Failed to create project: " + err.Error(),
		})
		return
	}

	http.Redirect(w, r, "/projects", http.StatusSeeOther)
}

func (h *FrontendHandler) EditProjectForm(w http.ResponseWriter, r *http.Request) {
	projectPublicID, err := projectPublicIDFromRequest(r)
	if err != nil {
		h.renderErrorPage(w, http.StatusBadRequest, "Bad Request", "Invalid project ID.", "")
		return
	}

	project, err := h.store.Stores().Projects.GetByPublicID(r.Context(), projectPublicID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			h.renderErrorPage(w, http.StatusNotFound, "Not Found", "Project not found.", "")
			return
		}
		slog.Error("failed to fetch project", "public_id", projectPublicID.String(), "error", err)
		h.renderErrorPage(w, http.StatusInternalServerError, "Internal Server Error", "Failed to load project.", "")
		return
	}

	h.renderProjectForm(w, r, projectFormPageData{
		Breadcrumbs: []component.BreadcrumbItem{
			{Name: "Home", Link: "/", Active: false},
			{Name: "Projects", Link: "/projects", Active: false},
			{Name: project.Name, Link: "/projects/" + projectPublicID.String() + "/edit", Active: true},
		},
		Title:       "Edit Project",
		Heading:     "Edit Project",
		FormAction:  "/projects/" + projectPublicID.String() + "/edit",
		SubmitLabel: "Save Changes",
		Name:        project.Name,
		Slug:        project.Slug,
	})
}

func (h *FrontendHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	projectPublicID, err := projectPublicIDFromRequest(r)
	if err != nil {
		h.renderErrorPage(w, http.StatusBadRequest, "Bad Request", "Invalid project ID.", "")
		return
	}

	project, err := h.store.Stores().Projects.GetByPublicID(r.Context(), projectPublicID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			h.renderErrorPage(w, http.StatusNotFound, "Not Found", "Project not found.", "")
			return
		}
		slog.Error("failed to fetch project", "public_id", projectPublicID.String(), "error", err)
		h.renderErrorPage(w, http.StatusInternalServerError, "Internal Server Error", "Failed to load project.", "")
		return
	}

	if err := r.ParseForm(); err != nil {
		h.renderProjectForm(w, r, projectFormPageData{
			Breadcrumbs: []component.BreadcrumbItem{
				{Name: "Home", Link: "/", Active: false},
				{Name: "Projects", Link: "/projects", Active: false},
				{Name: project.Name, Link: "/projects/" + projectPublicID.String() + "/edit", Active: true},
			},
			Title:        "Edit Project",
			Heading:      "Edit Project",
			FormAction:   "/projects/" + projectPublicID.String() + "/edit",
			SubmitLabel:  "Save Changes",
			Name:         project.Name,
			Slug:         project.Slug,
			ErrorMessage: "Failed to process form submission.",
		})
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	slug := normalizeSlug(r.FormValue("slug"))
	if slug == "" {
		slug = normalizeSlug(name)
	}
	if name == "" {
		h.renderProjectForm(w, r, projectFormPageData{
			Breadcrumbs: []component.BreadcrumbItem{
				{Name: "Home", Link: "/", Active: false},
				{Name: "Projects", Link: "/projects", Active: false},
				{Name: project.Name, Link: "/projects/" + projectPublicID.String() + "/edit", Active: true},
			},
			Title:        "Edit Project",
			Heading:      "Edit Project",
			FormAction:   "/projects/" + projectPublicID.String() + "/edit",
			SubmitLabel:  "Save Changes",
			Name:         name,
			Slug:         slug,
			ErrorMessage: "Project name is required.",
		})
		return
	}

	project.Name = name
	project.Slug = slug
	_, err = h.store.Stores().Projects.Update(r.Context(), project)
	if err != nil {
		slog.Error("failed to update project", "public_id", projectPublicID.String(), "error", err)
		h.renderProjectForm(w, r, projectFormPageData{
			Breadcrumbs: []component.BreadcrumbItem{
				{Name: "Home", Link: "/", Active: false},
				{Name: "Projects", Link: "/projects", Active: false},
				{Name: project.Name, Link: "/projects/" + projectPublicID.String() + "/edit", Active: true},
			},
			Title:        "Edit Project",
			Heading:      "Edit Project",
			FormAction:   "/projects/" + projectPublicID.String() + "/edit",
			SubmitLabel:  "Save Changes",
			Name:         name,
			Slug:         slug,
			ErrorMessage: "Failed to update project: " + err.Error(),
		})
		return
	}

	http.Redirect(w, r, "/projects", http.StatusSeeOther)
}

type projectFormPageData struct {
	Breadcrumbs  []component.BreadcrumbItem
	Title        string
	Heading      string
	FormAction   string
	SubmitLabel  string
	Name         string
	Slug         string
	ErrorMessage string
}

func (h *FrontendHandler) renderProjectForm(w http.ResponseWriter, r *http.Request, data projectFormPageData) {
	h.renderPage(w, r, "page/projects/form.html", data)
}

func projectPublicIDFromRequest(r *http.Request) (uuid.UUID, error) {
	projectPublicID, err := uuid.Parse(r.PathValue("public_id"))
	if err != nil {
		return uuid.Nil, errors.New("invalid project public id")
	}
	return projectPublicID, nil
}

func normalizeSlug(raw string) string {
	raw = strings.ToLower(strings.TrimSpace(raw))
	if raw == "" {
		return ""
	}

	var b strings.Builder
	lastDash := false
	for _, r := range raw {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			lastDash = false
		default:
			if !lastDash {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}

	return strings.Trim(b.String(), "-")
}

func queryParseInt(r *http.Request, key string, defaultValue int) int {
	rawValue := r.URL.Query().Get(key)
	if rawValue == "" {
		return defaultValue
	}
	parsed, err := strconv.Atoi(rawValue)
	if err != nil || parsed <= 0 {
		return defaultValue
	}
	return parsed
}
