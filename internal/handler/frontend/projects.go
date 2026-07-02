package frontend

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/roemer/test-tamer/internal/component"
	"github.com/roemer/test-tamer/internal/handler/shared"
	"github.com/roemer/test-tamer/internal/model"
	"github.com/roemer/test-tamer/internal/store"
)

func (h *FrontendHandler) Projects(w http.ResponseWriter, r *http.Request) {
	page := shared.QueryParsePositiveInt(r, "page", 1)
	pageSize := shared.QueryParsePositiveInt(r, "page-size", component.DefaultPageSize)
	pageSize = min(pageSize, component.MaxPageSize)

	// Handle partial
	if r.Header.Get("HX-Request") == "true" {
		h.renderProjectsListPartial(w, r, page, pageSize)
		return
	}

	// Handle full page
	type pageData struct {
		Breadcrumbs component.Breadcrumbs
		Page        int
		PageSize    int
	}
	data := pageData{
		Breadcrumbs: component.NewBreadcrumbs(
			component.NewBreadcrumbItem("Home", "/"),
			component.NewBreadcrumbItem("Projects", ""),
		),
		Page:     page,
		PageSize: pageSize,
	}
	h.renderPage(w, r, "page/projects/list.html", data)
}

func (h *FrontendHandler) NewProjectForm(w http.ResponseWriter, r *http.Request) {
	h.renderProjectForm(w, r, projectFormPageData{
		Breadcrumbs: component.NewBreadcrumbs(
			component.NewBreadcrumbItem("Home", "/"),
			component.NewBreadcrumbItem("Projects", "/projects"),
			component.NewBreadcrumbItem("New", ""),
		),
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
			Breadcrumbs: component.NewBreadcrumbs(
				component.NewBreadcrumbItem("Home", "/"),
				component.NewBreadcrumbItem("Projects", "/projects"),
				component.NewBreadcrumbItem("New", ""),
			),
			Title:        "New Project",
			Heading:      "Create Project",
			FormAction:   "/projects/new",
			SubmitLabel:  "Create Project",
			ErrorMessage: "Failed to process form submission.",
		})
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		h.renderProjectForm(w, r, projectFormPageData{
			Breadcrumbs: component.NewBreadcrumbs(
				component.NewBreadcrumbItem("Home", "/"),
				component.NewBreadcrumbItem("Projects", "/projects"),
				component.NewBreadcrumbItem("New", ""),
			),
			Title:        "New Project",
			Heading:      "Create Project",
			FormAction:   "/projects/new",
			SubmitLabel:  "Create Project",
			Name:         name,
			ErrorMessage: "Project name is required.",
		})
		return
	}

	_, err := h.store.Stores().Projects.Create(r.Context(), model.Project{
		Name: name,
	})
	if err != nil {
		slog.Error("failed to create project", "error", err)
		h.renderProjectForm(w, r, projectFormPageData{
			Breadcrumbs: component.NewBreadcrumbs(
				component.NewBreadcrumbItem("Home", "/"),
				component.NewBreadcrumbItem("Projects", "/projects"),
				component.NewBreadcrumbItem("New", ""),
			),
			Title:        "New Project",
			Heading:      "Create Project",
			FormAction:   "/projects/new",
			SubmitLabel:  "Create Project",
			Name:         name,
			ErrorMessage: "Failed to create project: " + err.Error(),
		})
		return
	}

	http.Redirect(w, r, "/projects", http.StatusSeeOther)
}

func (h *FrontendHandler) EditProjectForm(w http.ResponseWriter, r *http.Request) {
	projectPublicID, err := shared.QueryParsePublicID(r)
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
		Breadcrumbs: component.NewBreadcrumbs(
			component.NewBreadcrumbItem("Home", "/"),
			component.NewBreadcrumbItem("Projects", "/projects"),
			component.NewBreadcrumbItem(project.Name, ""),
		),
		Title:       "Edit Project",
		Heading:     "Edit Project",
		FormAction:  "/projects/" + projectPublicID.String() + "/edit",
		SubmitLabel: "Save Changes",
		Name:        project.Name,
	})
}

func (h *FrontendHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	projectPublicID, err := shared.QueryParsePublicID(r)
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
			Breadcrumbs: component.NewBreadcrumbs(
				component.NewBreadcrumbItem("Home", "/"),
				component.NewBreadcrumbItem("Projects", "/projects"),
				component.NewBreadcrumbItem(project.Name, ""),
			),
			Title:        "Edit Project",
			Heading:      "Edit Project",
			FormAction:   "/projects/" + projectPublicID.String() + "/edit",
			SubmitLabel:  "Save Changes",
			Name:         project.Name,
			ErrorMessage: "Failed to process form submission.",
		})
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		h.renderProjectForm(w, r, projectFormPageData{
			Breadcrumbs: component.NewBreadcrumbs(
				component.NewBreadcrumbItem("Home", "/"),
				component.NewBreadcrumbItem("Projects", "/projects"),
				component.NewBreadcrumbItem(project.Name, ""),
			),
			Title:        "Edit Project",
			Heading:      "Edit Project",
			FormAction:   "/projects/" + projectPublicID.String() + "/edit",
			SubmitLabel:  "Save Changes",
			Name:         name,
			ErrorMessage: "Project name is required.",
		})
		return
	}

	project.Name = name
	_, err = h.store.Stores().Projects.Update(r.Context(), project)
	if err != nil {
		slog.Error("failed to update project", "public_id", projectPublicID.String(), "error", err)
		h.renderProjectForm(w, r, projectFormPageData{
			Breadcrumbs: component.NewBreadcrumbs(
				component.NewBreadcrumbItem("Home", "/"),
				component.NewBreadcrumbItem("Projects", "/projects"),
				component.NewBreadcrumbItem(project.Name, ""),
			),
			Title:        "Edit Project",
			Heading:      "Edit Project",
			FormAction:   "/projects/" + projectPublicID.String() + "/edit",
			SubmitLabel:  "Save Changes",
			Name:         name,
			ErrorMessage: "Failed to update project: " + err.Error(),
		})
		return
	}

	http.Redirect(w, r, "/projects", http.StatusSeeOther)
}

func (h *FrontendHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	projectPublicID, err := shared.QueryParsePublicID(r)
	if err != nil {
		h.renderErrorPage(w, http.StatusBadRequest, "Bad Request", "Invalid project ID.", "")
		return
	}

	err = h.store.Stores().Projects.DeleteByPublicID(r.Context(), projectPublicID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			h.renderErrorPage(w, http.StatusNotFound, "Not Found", "Project not found.", "")
			return
		}
		slog.Error("failed to delete project", "public_id", projectPublicID.String(), "error", err)
		h.renderErrorPage(w, http.StatusInternalServerError, "Internal Server Error", "Failed to delete project.", "")
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		page := shared.QueryParsePositiveInt(r, "page", 1)
		pageSize := shared.QueryParsePositiveInt(r, "page-size", component.DefaultPageSize)
		pageSize = min(pageSize, component.MaxPageSize)
		h.renderProjectsListPartial(w, r, page, pageSize)
		return
	}

	http.Redirect(w, r, "/projects", http.StatusSeeOther)
}

func (h *FrontendHandler) renderProjectsListPartial(w http.ResponseWriter, r *http.Request, page, pageSize int) {
	totalProjects, err := h.store.Stores().Projects.Count(r.Context())
	if err != nil {
		slog.Error("failed to count projects", "error", err)
		h.InternalServerError(w, r, "Failed to count projects", "")
		return
	}

	paging := component.NewPaging("/projects", "#project-list", page, pageSize, totalProjects)
	projects, err := h.store.Stores().Projects.List(r.Context(), paging.Page, paging.PageSize)
	if err != nil {
		slog.Error("failed to list projects", "error", err)
		h.InternalServerError(w, r, "Failed to list projects.", "")
		return
	}

	rows := make([]component.TableRow, 0, len(projects))
	for _, project := range projects {
		editURL := "/projects/" + project.PublicID.String() + "/edit"
		deleteURL := fmt.Sprintf("/projects/%s/delete?page=%d&page-size=%d", project.PublicID.String(), paging.Page, paging.PageSize)
		rows = append(rows, component.TableRow{
			Cells: []component.TableCell{
				{Value: project.PublicID.String()},
				{Value: project.Name},
			},
			Actions: []component.TableAction{
				{
					Label: "Edit",
					Class: "btn btn-sm btn-outline-secondary",
					Attrs: HTMXAttrs(map[string]string{
						"type":    "button",
						"onclick": "window.location.href='" + editURL + "'",
					}),
				},
				{
					Label: "Delete",
					Class: "btn btn-sm btn-outline-danger",
					Attrs: HTMXAttrs(map[string]string{
						"type":       "button",
						"hx-delete":  deleteURL,
						"hx-confirm": "Delete project " + project.Name + "?",
						"hx-target":  "#project-list",
						"hx-swap":    "innerHTML",
					}),
				},
			},
		})
	}

	type partialData struct {
		ProjectsTable *component.Table
		Paging        *component.Paging
	}
	data := partialData{
		ProjectsTable: &component.Table{
			Columns: []component.TableColumn{
				{Header: "ID"},
				{Header: "Name"},
			},
			Rows: rows,
		},
		Paging: paging,
	}
	h.renderPartial(w, "partial/projects/list.html", data)
}

type projectFormPageData struct {
	Breadcrumbs  component.Breadcrumbs
	Title        string
	Heading      string
	FormAction   string
	SubmitLabel  string
	Name         string
	ErrorMessage string
}

func (h *FrontendHandler) renderProjectForm(w http.ResponseWriter, r *http.Request, data projectFormPageData) {
	h.renderPage(w, r, "page/projects/form.html", data)
}
