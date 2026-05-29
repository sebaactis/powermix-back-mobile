package middlewares

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sebaactis/powermix-back-mobile/internal/utils"
)

func TestRecoverer_PanicReturnsStandardEnvelope(t *testing.T) {
	logger := slog.Default()
	handler := Recoverer(logger)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	}))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/panic", nil))

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}

	var resp utils.APIResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Success {
		t.Fatal("expected success=false")
	}

	errMap, ok := resp.Error.(map[string]interface{})
	if !ok {
		t.Fatalf("error type = %T", resp.Error)
	}
	if got := errMap["code"]; got != utils.ErrCodeInternal {
		t.Fatalf("code = %v, want %s", got, utils.ErrCodeInternal)
	}
}

func TestTimeout_ReturnsTimeoutEnvelope(t *testing.T) {
	handler := Timeout(10 * time.Millisecond)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/slow", nil))

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusServiceUnavailable)
	}

	var resp utils.APIResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Success {
		t.Fatal("expected success=false")
	}

	errMap, ok := resp.Error.(map[string]interface{})
	if !ok {
		t.Fatalf("error type = %T", resp.Error)
	}
	if got := errMap["code"]; got != utils.ErrCodeTimeout {
		t.Fatalf("code = %v, want %s", got, utils.ErrCodeTimeout)
	}
	if got := errMap["message"]; got != "La solicitud tardó demasiado" {
		t.Fatalf("message = %v", got)
	}
}
