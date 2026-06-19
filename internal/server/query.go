package server

import (
	"net/http"
	"strconv"
)

func queryParseInt(r *http.Request, key string, defaultValue int) int {
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
