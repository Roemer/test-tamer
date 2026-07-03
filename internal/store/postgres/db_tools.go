package postgres

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func BuildDbUrlFromEnv() (string, error) {
	if v, ok := os.LookupEnv("TEST_TAMER_POSTGRES_URL"); ok && v != "" {
		return v, nil
	}
	dbHost := os.Getenv("TEST_TAMER_POSTGRES_HOST")
	dbPort := os.Getenv("TEST_TAMER_POSTGRES_PORT")
	dbUser := os.Getenv("TEST_TAMER_POSTGRES_USER")
	dbPassword := os.Getenv("TEST_TAMER_POSTGRES_PASSWORD")
	dbName := os.Getenv("TEST_TAMER_POSTGRES_DB")
	sslMode := os.Getenv("TEST_TAMER_POSTGRES_SSLMODE")
	return BuildDbUrl(dbHost, dbPort, dbUser, dbPassword, dbName, sslMode)
}

func BuildDbUrl(dbHost string, dbPort string, dbUser string, dbPassword string, dbName string, sslMode string) (string, error) {
	if dbHost == "" {
		dbHost = "localhost"
	}
	if dbPort == "" {
		dbPort = "5432"
	}
	if dbUser == "" {
		return "", fmt.Errorf("dbUser is required")
	}
	if dbPassword == "" {
		return "", fmt.Errorf("dbPassword is required")
	}
	if dbName == "" {
		return "", fmt.Errorf("dbName is required")
	}
	if sslMode == "" {
		sslMode = "disable"
	}

	dbURL := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(dbUser, dbPassword),
		Host:   fmt.Sprintf("%s:%s", dbHost, dbPort),
		Path:   "/" + dbName,
	}

	queryValues := url.Values{}
	queryValues.Set("sslmode", sslMode)
	dbURL.RawQuery = queryValues.Encode()

	return dbURL.String(), nil
}

func CreateDbPool(dbUrl string) (*pgxpool.Pool, error) {
	dbpool, err := pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection pool: %w", err)
	}
	return dbpool, nil
}
