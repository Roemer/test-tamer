package api

import (
	"encoding/json"
	"net/http"
)

type APIHandler struct {
}

func NewAPIHandler() *APIHandler {
	return &APIHandler{}
}

func (h *APIHandler) Ping(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
