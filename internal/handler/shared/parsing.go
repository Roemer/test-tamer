package shared

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

func QueryParsePublicID(r *http.Request) (uuid.UUID, error) {
	projectPublicID, err := uuid.Parse(r.PathValue("public_id"))
	if err != nil {
		return uuid.Nil, errors.New("invalid project public id")
	}
	return projectPublicID, nil
}

func QueryParsePositiveInt(r *http.Request, key string, defaultValue int) int {
	rawValue := r.URL.Query().Get(key)
	if rawValue == "" {
		return defaultValue
	}
	parsed, err := strconv.Atoi(rawValue)
	if err != nil || parsed <= 0 {
		return defaultValue
	}
	return parsed
}
