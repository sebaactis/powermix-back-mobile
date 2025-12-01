package utils

import (
	"encoding/json"
	"net/http"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Error   interface{} `json:"error"`
}

type APIError struct {
	Message string      `json:"message"`
	Fields  interface{} `json:"fields,omitempty"`
}

func WriteSuccess(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := APIResponse{
		Success: true,
		Data:    data,
		Error:   nil,
	}

	_ = json.NewEncoder(w).Encode(resp)
}

func WriteError(w http.ResponseWriter, statusCode int, message string, fields interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	apiErr := APIError{
		Message: message,
		Fields:  fields,
	}

	resp := APIResponse{
		Success: false,
		Data:    nil,
		Error:   apiErr,
	}

	_ = json.NewEncoder(w).Encode(resp)
}