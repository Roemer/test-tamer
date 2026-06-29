package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/roemer/test-tamer/internal/entities"
)

var (
	ErrNotFound         = errors.New("record not found")
	ErrTooMany          = errors.New("too many records found")
	ErrInvalid          = errors.New("invalid argument")
	ErrUniqueConstraint = errors.New("unique constraint violation")
)

type DBTX interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func validatePositiveID(id int64, field string) error {
	if id <= 0 {
		return errors.Join(ErrInvalid, errors.New(field+" must be > 0"))
	}
	return nil
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

func createBaseEntity(ctx context.Context, db DBTX, entity *entities.BaseEntity, query string, args pgx.NamedArgs) error {
	row := db.QueryRow(ctx, query, args)
	err := row.Scan(
		&entity.ID,
		&entity.PublicID,
		&entity.CreatedAt,
		&entity.UpdatedAt,
	)
	return translatePgError(err)
}

func updateBaseEntity(ctx context.Context, db DBTX, entity *entities.BaseEntity, query string, args pgx.NamedArgs) error {
	row := db.QueryRow(ctx, query, args)
	err := row.Scan(
		&entity.UpdatedAt,
	)
	return translatePgError(err)
}

func execDelete(ctx context.Context, db DBTX, query string, args pgx.NamedArgs) error {
	result, err := db.Exec(ctx, query, args)
	if err != nil {
		return translatePgError(err)
	} else if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func translatePgError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	} else if errors.Is(err, pgx.ErrTooManyRows) {
		return ErrTooMany
	} else if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		if pgErr.Code == pgerrcode.UniqueViolation {
			return fmt.Errorf("%w: %s", ErrUniqueConstraint, pgErr.ConstraintName)
		}
	}
	return err
}
