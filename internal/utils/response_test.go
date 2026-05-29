package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteError_IncludesCodeInJSON(t *testing.T) {
	rr := httptest.NewRecorder()
	WriteError(rr, http.StatusBadRequest, WriteErrorOpts{
		Code:    ErrCodeValidation,
		Message: "Dato inválido",
	})

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}

	var resp APIResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Success {
		t.Fatal("expected success=false")
	}

	errMap, ok := resp.Error.(map[string]interface{})
	if !ok {
		t.Fatalf("error type = %T, want map[string]interface{}", resp.Error)
	}
	if got := errMap["code"]; got != ErrCodeValidation {
		t.Fatalf("code = %v, want %s", got, ErrCodeValidation)
	}
	if got := errMap["message"]; got != "Dato inválido" {
		t.Fatalf("message = %v, want Dato inválido", got)
	}
}

func TestWriteError_ConflictCode(t *testing.T) {
	rr := httptest.NewRecorder()
	WriteError(rr, http.StatusConflict, WriteErrorOpts{
		Code:    ErrCodeConflict,
		Message: "Conflicto de estado",
	})

	var resp APIResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	errMap := resp.Error.(map[string]interface{})
	if got := errMap["code"]; got != ErrCodeConflict {
		t.Fatalf("code = %v, want %s", got, ErrCodeConflict)
	}
}

func TestWriteError_AllErrorConstantsExist(t *testing.T) {
	codes := []string{
		ErrCodeValidation,
		ErrCodeNotFound,
		ErrCodeDuplicateEntry,
		ErrCodeConflict,
		ErrCodeInvalidCreds,
		ErrCodeUnauthorized,
		ErrCodeTimeout,
		ErrCodeInternal,
		ErrCodeExternalService,
	}
	if len(codes) != 9 {
		t.Fatalf("expected 9 error code constants, got %d", len(codes))
	}
	for _, code := range codes {
		if code == "" {
			t.Fatal("error code constant must not be empty")
		}
	}
}
