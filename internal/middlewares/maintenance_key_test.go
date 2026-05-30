package middlewares

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sebaactis/powermix-back-mobile/internal/utils"
)

type mockMaintenanceKeyConfig struct {
	prodeEnabled       bool
	maintenanceEnabled bool
	adminKey           string
}

func (m *mockMaintenanceKeyConfig) IsProdeEnabled() bool       { return m.prodeEnabled }
func (m *mockMaintenanceKeyConfig) IsMaintenanceEnabled() bool { return m.maintenanceEnabled }
func (m *mockMaintenanceKeyConfig) AdminAPIKey() string         { return m.adminKey }

func TestMaintenanceKey_MissingKey_Returns401Envelope(t *testing.T) {
	cfg := &mockMaintenanceKeyConfig{
		prodeEnabled:       true,
		maintenanceEnabled: true,
		adminKey:           "secret123",
	}

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	})

	handler := MaintenanceKey(cfg)(inner)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/prode/admin/matches", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
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
	if got := errMap["code"]; got != utils.ErrCodeUnauthorized {
		t.Fatalf("code = %v, want %s", got, utils.ErrCodeUnauthorized)
	}
	if got := errMap["message"]; got != "No autorizado" {
		t.Fatalf("message = %v, want 'No autorizado'", got)
	}
}

func TestMaintenanceKey_WrongKey_Returns401Envelope(t *testing.T) {
	cfg := &mockMaintenanceKeyConfig{
		prodeEnabled:       true,
		maintenanceEnabled: true,
		adminKey:           "secret123",
	}

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	})

	handler := MaintenanceKey(cfg)(inner)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/prode/admin/matches", nil)
	req.Header.Set("X-Prode-Admin-Key", "wrongkey")
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}

	var resp utils.APIResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Success {
		t.Fatal("expected success=false")
	}
}

func TestMaintenanceKey_CorrectKey_CallsNextHandler(t *testing.T) {
	cfg := &mockMaintenanceKeyConfig{
		prodeEnabled:       true,
		maintenanceEnabled: true,
		adminKey:           "secret123",
	}

	called := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	handler := MaintenanceKey(cfg)(inner)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/prode/admin/matches", nil)
	req.Header.Set("X-Prode-Admin-Key", "secret123")
	handler.ServeHTTP(rr, req)

	if !called {
		t.Fatal("next handler was not called")
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestMaintenanceKey_Disabled_CallsNextWithoutKey(t *testing.T) {
	cfg := &mockMaintenanceKeyConfig{
		prodeEnabled:       true,
		maintenanceEnabled: false,
		adminKey:           "secret123",
	}

	called := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	handler := MaintenanceKey(cfg)(inner)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/prode/admin/matches", nil)
	handler.ServeHTTP(rr, req)

	if !called {
		t.Fatal("next handler was not called when maintenance disabled")
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}
