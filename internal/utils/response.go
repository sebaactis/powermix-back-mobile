package utils

import (
	"encoding/json"
	"net/http"
)

type ErrBody struct {
	Error   string            `json:"error"`
	Details map[string]string `json:"details,omitempty"`
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func WriteError(w http.ResponseWriter, status int, msg string, details map[string]string) {
	WriteJSON(w, status, ErrBody{Error: msg, Details: details})
}
