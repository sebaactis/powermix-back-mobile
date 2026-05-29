package utils

import (
	"encoding/json"
	"net/http"
)

const (
	ErrCodeValidation      = "ERR_VALIDATION"
	ErrCodeNotFound        = "ERR_NOT_FOUND"
	ErrCodeDuplicateEntry  = "ERR_DUPLICATE_ENTRY"
	ErrCodeConflict        = "ERR_CONFLICT"
	ErrCodeInvalidCreds    = "ERR_INVALID_CREDENTIALS"
	ErrCodeUnauthorized    = "ERR_UNAUTHORIZED"
	ErrCodeTimeout         = "ERR_TIMEOUT"
	ErrCodeInternal        = "ERR_INTERNAL"
	ErrCodeExternalService = "ERR_EXTERNAL_SERVICE"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Error   interface{} `json:"error"`
}

type APIError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Fields  interface{} `json:"fields,omitempty"`
}

type WriteErrorOpts struct {
	Code    string
	Message string
	Fields  interface{}
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

// WriteErrorMessage maps HTTP status to a default error code during handler migration.
func WriteErrorMessage(w http.ResponseWriter, statusCode int, message string, fields interface{}) {
	WriteError(w, statusCode, WriteErrorOpts{
		Code:    codeFromHTTPStatus(statusCode),
		Message: message,
		Fields:  fields,
	})
}

func codeFromHTTPStatus(statusCode int) string {
	switch statusCode {
	case http.StatusBadRequest:
		return ErrCodeValidation
	case http.StatusUnauthorized:
		return ErrCodeUnauthorized
	case http.StatusForbidden:
		return ErrCodeUnauthorized
	case http.StatusNotFound:
		return ErrCodeNotFound
	case http.StatusConflict:
		return ErrCodeDuplicateEntry
	case http.StatusLocked:
		return ErrCodeInvalidCreds
	case http.StatusGatewayTimeout, http.StatusRequestTimeout:
		return ErrCodeTimeout
	default:
		return ErrCodeInternal
	}
}

func WriteError(w http.ResponseWriter, statusCode int, opts WriteErrorOpts) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	apiErr := APIError{
		Code:    opts.Code,
		Message: opts.Message,
		Fields:  opts.Fields,
	}

	resp := APIResponse{
		Success: false,
		Data:    nil,
		Error:   apiErr,
	}

	_ = json.NewEncoder(w).Encode(resp)
}
