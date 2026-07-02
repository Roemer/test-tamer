package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/roemer/test-tamer/internal/model"
	"github.com/roemer/test-tamer/internal/store"
)

type projectStore struct {
	db DBTX
}

func newProjectStore(db DBTX) store.ProjectStore {
	return &projectStore{db: db}
}

func (s *projectStore) Create(ctx context.Context, project model.Project) (model.Project, error) {
	return project, nil
}

func (s *projectStore) GetByPublicID(ctx context.Context, publicID uuid.UUID) (model.Project, error) {
	return model.Project{}, store.ErrNotFound
}

func (s *projectStore) Update(ctx context.Context, project model.Project) (model.Project, error) {
	return model.Project{}, store.ErrNotFound
}

func (s *projectStore) DeleteByPublicID(ctx context.Context, publicID uuid.UUID) error {
	return store.ErrNotFound
}

func (s *projectStore) List(ctx context.Context, page, pageSize int) ([]model.Project, error) {
	return nil, nil
}

func (s *projectStore) Count(ctx context.Context) (int, error) {
	return 0, nil
}
