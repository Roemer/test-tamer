package db

import (
	"fmt"
	"net/url"
	"os"
)

func BuildDbUrlFromEnv() (string, error) {
	if v, ok := os.LookupEnv("DB_URL"); ok && v != "" {
		return v, nil
	}
	dbHost := os.Getenv("POSTGRES_HOSTNAME")
	dbPort := os.Getenv("POSTGRES_PORT")
	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")
	sslMode := os.Getenv("POSTGRES_SSLMODE")
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
