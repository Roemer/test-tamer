package store

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Client is the main interface for interacting with the data store.
// It provides methods for executing transactions and accessing repositories directly.
type Client interface {
	// Transact executes fn in a single transaction. Do not run repository calls concurrently inside fn.
	Transact(ctx context.Context, fn func(ctx context.Context, repos *Repositories) error) error
	// TransactWithOptions executes fn in a single transaction with the provided transaction options. Do not run repository calls concurrently inside fn.
	TransactWithOptions(ctx context.Context, txOptions pgx.TxOptions, fn func(ctx context.Context, repos *Repositories) error) error
	// Repos returns a new Repositories instance with the underlying database connection.
	Repos() *Repositories
}

type dbClient struct {
	pool *pgxpool.Pool
}

// New creates a new Client instance with the provided pgxpool.Pool.
func New(pool *pgxpool.Pool) Client {
	return &dbClient{pool: pool}
}

func (c *dbClient) Transact(ctx context.Context, fn func(ctx context.Context, repos *Repositories) error) error {
	return c.TransactWithOptions(ctx, pgx.TxOptions{}, fn)
}

func (c *dbClient) TransactWithOptions(ctx context.Context, txOptions pgx.TxOptions, fn func(ctx context.Context, repos *Repositories) error) error {
	return pgx.BeginTxFunc(ctx, c.pool, txOptions, func(tx pgx.Tx) error {
		return fn(ctx, NewDbRepositories(tx))
	})
}

func (c *dbClient) Repos() *Repositories {
	return NewDbRepositories(c.pool)
}
