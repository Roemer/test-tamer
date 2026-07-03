package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
	query := `
		INSERT INTO projects (name)
		VALUES (@name)
		RETURNING id, public_id, created_at, updated_at, name`
	created, err := queryAndCollectOne[model.Project](ctx, s.db, query, pgx.NamedArgs{
		"name": project.Name,
	})
	if err != nil {
		return model.Project{}, err
	}
	return *created, nil
}

func (s *projectStore) GetByPublicID(ctx context.Context, publicID uuid.UUID) (model.Project, error) {
	query := `
		SELECT id, public_id, created_at, updated_at, name
		FROM projects
		WHERE public_id = @public_id`
	project, err := queryAndCollectOne[model.Project](ctx, s.db, query, pgx.NamedArgs{
		"public_id": publicID,
	})
	if err != nil {
		return model.Project{}, err
	}
	return *project, nil
}

func (s *projectStore) Update(ctx context.Context, project model.Project) (model.Project, error) {
	query := `
		UPDATE projects
		SET name = @name
		WHERE id = @id
		RETURNING id, public_id, created_at, updated_at, name`
	updated, err := queryAndCollectOne[model.Project](ctx, s.db, query, pgx.NamedArgs{
		"id":   project.ID,
		"name": project.Name,
	})
	if err != nil {
		return model.Project{}, err
	}
	return *updated, nil
}

func (s *projectStore) DeleteByPublicID(ctx context.Context, publicID uuid.UUID) error {
	query := `
		DELETE FROM projects
		WHERE public_id = @public_id`
	return execDelete(ctx, s.db, query, pgx.NamedArgs{
		"public_id": publicID,
	})
}

func (s *projectStore) List(ctx context.Context, page, pageSize int) ([]model.Project, error) {
	if page <= 0 || pageSize <= 0 {
		return nil, store.ErrInvalid
	}
	offset := (page - 1) * pageSize
	query := `
		SELECT id, public_id, created_at, updated_at, name
		FROM projects
		ORDER BY id ASC
		LIMIT @limit OFFSET @offset`
	return queryAndCollectRows[model.Project](ctx, s.db, query, pgx.NamedArgs{
		"limit":  pageSize,
		"offset": offset,
	})
}

func (s *projectStore) Count(ctx context.Context) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM projects`
	var total int
	if err := s.db.QueryRow(ctx, query).Scan(&total); err != nil {
		return 0, translatePgError(err)
	}
	return total, nil
}
