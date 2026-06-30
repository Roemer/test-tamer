package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roemer/test-tamer/internal/store"
)

type DBTX interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type dbClient struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) store.Client {
	return &dbClient{pool: pool}
}

func (c *dbClient) Transact(ctx context.Context, fn func(ctx context.Context, stores *store.Stores) error) error {
	return pgx.BeginTxFunc(ctx, c.pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		return fn(ctx, newStores(tx))
	})
}

func (c *dbClient) Stores() *store.Stores {
	return newStores(c.pool)
}

func newStores(db DBTX) *store.Stores {
	return &store.Stores{
		Projects: newProjectStore(db),
	}
}
