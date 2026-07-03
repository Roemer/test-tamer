package store

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/roemer/test-tamer/internal/model"
)

var (
	ErrNotFound         = errors.New("not found")
	ErrTooMany          = errors.New("too many records found")
	ErrInvalid          = errors.New("invalid argument")
	ErrUniqueConstraint = errors.New("unique constraint violation")
)

type Stores struct {
	Projects ProjectStore
}

type Client interface {
	Transact(ctx context.Context, fn func(ctx context.Context, stores *Stores) error) error
	Stores() *Stores
}

type ProjectStore interface {
	Create(ctx context.Context, project model.Project) (model.Project, error)
	GetByPublicID(ctx context.Context, publicID uuid.UUID) (model.Project, error)
	Update(ctx context.Context, project model.Project) (model.Project, error)
	DeleteByPublicID(ctx context.Context, publicID uuid.UUID) error
	List(ctx context.Context, page, pageSize int) ([]model.Project, error)
	Count(ctx context.Context) (int, error)
}
