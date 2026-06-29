package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	dbUrl string
}

func New(dbUrl string) *DB {
	db := &DB{
		dbUrl: dbUrl,
	}
	return db
}

func (db *DB) CreateDbPool() (*pgxpool.Pool, error) {
	dbpool, err := pgxpool.New(context.Background(), db.dbUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection pool: %w", err)
	}
	return dbpool, nil
}

func (db *DB) GetDbUrl() string {
	return db.dbUrl
}
