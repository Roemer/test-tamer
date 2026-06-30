package memory

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/roemer/test-tamer/internal/model"
	"github.com/roemer/test-tamer/internal/store"
)

type projectStore struct {
	projects []model.Project
}

func newProjectStore() store.ProjectStore {
	return &projectStore{}
}

func (s *projectStore) Create(ctx context.Context, project model.Project) (model.Project, error) {
	// Calculate the next ID
	maxId := int64(0)
	for _, p := range s.projects {
		if p.ID > maxId {
			maxId = p.ID
		}
		// Also check if the slug already exists
		if p.Slug == project.Slug {
			return model.Project{}, fmt.Errorf("project slug already exists")
		}
	}
	// Set the values
	project.ID = maxId + 1
	project.PublicID = uuid.Must(uuid.NewV7())
	project.CreatedAt = time.Now()
	project.UpdatedAt = time.Now()
	// Append the project
	s.projects = append(s.projects, project)
	return project, nil
}

func (s *projectStore) GetByPublicID(ctx context.Context, publicID uuid.UUID) (model.Project, error) {
	for _, project := range s.projects {
		if project.PublicID == publicID {
			return project, nil
		}
	}
	return model.Project{}, store.ErrNotFound
}

func (s *projectStore) Update(ctx context.Context, project model.Project) (model.Project, error) {
	for _, existing := range s.projects {
		if existing.Slug == project.Slug && existing.ID != project.ID {
			return model.Project{}, fmt.Errorf("project slug already exists")
		}
	}

	for i, existing := range s.projects {
		if existing.ID == project.ID {
			project.PublicID = existing.PublicID
			project.CreatedAt = existing.CreatedAt
			project.UpdatedAt = time.Now()
			s.projects[i] = project
			return project, nil
		}
	}

	return model.Project{}, store.ErrNotFound
}

func (s *projectStore) List(ctx context.Context, page, pageSize int) ([]model.Project, error) {
	start := (page - 1) * pageSize
	end := start + pageSize
	if start > len(s.projects) {
		return []model.Project{}, nil
	}
	if end > len(s.projects) {
		end = len(s.projects)
	}
	return s.projects[start:end], nil
}

func (s *projectStore) Count(ctx context.Context) (int, error) {
	return len(s.projects), nil
}
