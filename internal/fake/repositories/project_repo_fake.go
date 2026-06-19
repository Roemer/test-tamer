package repositories

import (
	"context"

	"github.com/roemer/test-tamer/internal/entities"
	"github.com/roemer/test-tamer/internal/repositories"
)

type fakeProjectRepository struct {
	projects []entities.Project
}

func NewFakeProjectRepository(projects []entities.Project) repositories.ProjectRepository {
	repoProjects := make([]entities.Project, len(projects))
	copy(repoProjects, projects)
	return &fakeProjectRepository{projects: repoProjects}
}

func (r *fakeProjectRepository) List(ctx context.Context) ([]entities.Project, error) {
	_ = ctx
	items := make([]entities.Project, len(r.projects))
	copy(items, r.projects)
	return items, nil
}

func (r *fakeProjectRepository) ListPaged(ctx context.Context, page, pageSize int) ([]entities.Project, error) {
	addRandomDelay()
	_ = ctx
	if pageSize <= 0 || page <= 0 {
		return []entities.Project{}, nil
	}

	offset := (page - 1) * pageSize
	if offset >= len(r.projects) {
		return []entities.Project{}, nil
	}

	end := min(offset+pageSize, len(r.projects))

	items := make([]entities.Project, end-offset)
	copy(items, r.projects[offset:end])
	return items, nil
}

func (r *fakeProjectRepository) Count(ctx context.Context) (int, error) {
	_ = ctx
	return len(r.projects), nil
}
