package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
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

func queryAndCollectOne[T any](ctx context.Context, db DBTX, query string, args pgx.NamedArgs) (*T, error) {
	rows, err := db.Query(ctx, query, args)
	if err != nil {
		return nil, translatePgError(err)
	}
	item, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[T])
	if err != nil {
		return nil, translatePgError(err)
	}
	return item, nil
}

func queryAndCollectRows[T any](ctx context.Context, db DBTX, query string, args pgx.NamedArgs) ([]T, error) {
	rows, err := db.Query(ctx, query, args)
	if err != nil {
		return nil, translatePgError(err)
	}
	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[T])
	if err != nil {
		return nil, translatePgError(err)
	}
	return items, nil
}

func execDelete(ctx context.Context, db DBTX, query string, args pgx.NamedArgs) error {
	result, err := db.Exec(ctx, query, args)
	if err != nil {
		return translatePgError(err)
	} else if result.RowsAffected() == 0 {
		return store.ErrNotFound
	}
	return nil
}

func translatePgError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return store.ErrNotFound
	} else if errors.Is(err, pgx.ErrTooManyRows) {
		return store.ErrTooMany
	} else if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		if pgErr.Code == pgerrcode.UniqueViolation {
			return fmt.Errorf("%w: %s", store.ErrUniqueConstraint, pgErr.ConstraintName)
		}
	}
	return err
}
