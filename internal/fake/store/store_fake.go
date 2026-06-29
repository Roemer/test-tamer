package store

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/roemer/test-tamer/internal/fake/repositories"
	"github.com/roemer/test-tamer/internal/store"
)

type fakeClient struct {
	repos *store.Repositories
}

// New creates a store client backed by the provided repositories.
func New() store.Client {
	return &fakeClient{repos: repositories.NewFakeRepositories()}
}

func (c *fakeClient) Transact(ctx context.Context, fn func(ctx context.Context, repos *store.Repositories) error) error {
	return fn(ctx, c.repos)
}

func (c *fakeClient) TransactWithOptions(ctx context.Context, txOptions pgx.TxOptions, fn func(ctx context.Context, repos *store.Repositories) error) error {
	_ = txOptions
	return fn(ctx, c.repos)
}

func (c *fakeClient) Repos() *store.Repositories {
	return c.repos
}
