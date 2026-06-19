package repositories

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/roemer/test-tamer/internal/entities"
)

type ProjectRepository interface {
	List(ctx context.Context) ([]entities.Project, error)
	ListPaged(ctx context.Context, page, pageSize int) ([]entities.Project, error)
	Count(ctx context.Context) (int, error)
}

type projectRepository struct {
	db DBTX
}

func NewProjectRepository(db DBTX) ProjectRepository {
	return &projectRepository{db: db}
}

func (r *projectRepository) List(ctx context.Context) ([]entities.Project, error) {
	query := `
		SELECT id, created_at, updated_at, name
		FROM projects
		ORDER BY id ASC`
	return queryAndCollectRows[entities.Project](ctx, r.db, query, nil)
}

func (r *projectRepository) ListPaged(ctx context.Context, page, pageSize int) ([]entities.Project, error) {
	if page <= 0 || pageSize <= 0 {
		return nil, ErrInvalid
	}
	offset := (page - 1) * pageSize
	query := `
		SELECT id, created_at, updated_at, name
		FROM projects
		ORDER BY id ASC
		LIMIT @limit OFFSET @offset`
	return queryAndCollectRows[entities.Project](ctx, r.db, query, pgx.NamedArgs{
		"limit":  pageSize,
		"offset": offset,
	})
}

func (r *projectRepository) Count(ctx context.Context) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM projects`
	var total int
	if err := r.db.QueryRow(ctx, query).Scan(&total); err != nil {
		return 0, translatePgError(err)
	}
	return total, nil
}
