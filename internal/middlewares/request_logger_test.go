package middlewares

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sebaactis/powermix-back-mobile/internal/platform/logger"
)

// captureHandler records slog Records for inspection.
type captureHandler struct {
	records []slog.Record
}

func (h *captureHandler) Enabled(context.Context, slog.Level) bool { return true }
func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	h.records = append(h.records, r)
	return nil
}
func (h *captureHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h *captureHandler) WithGroup(string) slog.Handler      { return h }

func attrsToMap(r slog.Record) map[string]any {
	m := make(map[string]any)
	r.Attrs(func(a slog.Attr) bool {
		m[a.Key] = a.Value.Any()
		return true
	})
	return m
}

func TestRequestLogger_LogsExpectedFields(t *testing.T) {
	cap := &captureHandler{}
	log := slog.New(logger.NewContextHandler(cap))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"ok":true}`))
	})

	mw := RequestLogger(log)(handler)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/proof", strings.NewReader(`{"foo":"bar"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-For", "203.0.113.42")
	rec := httptest.NewRecorder()

	mw.ServeHTTP(rec, req)

	if len(cap.records) != 1 {
		t.Fatalf("expected 1 log record, got %d", len(cap.records))
	}

	attrs := attrsToMap(cap.records[0])

	if attrs["method"] != "POST" {
		t.Errorf("method = %v, want POST", attrs["method"])
	}
	if attrs["path"] != "/api/v1/proof" {
		t.Errorf("path = %v, want /api/v1/proof", attrs["path"])
	}
	if attrs["status"] != int64(http.StatusCreated) {
		t.Errorf("status = %v, want %d", attrs["status"], http.StatusCreated)
	}
	if attrs["service"] != "proof" {
		t.Errorf("service = %v, want proof", attrs["service"])
	}
	if attrs["ip"] != "203.0.113.42" {
		t.Errorf("ip = %v, want 203.0.113.42", attrs["ip"])
	}
	if _, ok := attrs["request_id"]; !ok {
		t.Error("expected request_id in log attrs")
	}
	if _, ok := attrs["duration_ms"]; !ok {
		t.Error("expected duration_ms in log attrs")
	}

	body, ok := attrs["body"].(map[string]any)
	if !ok {
		t.Fatalf("expected body map, got %T", attrs["body"])
	}
	if body["foo"] != "bar" {
		t.Errorf("body.foo = %v, want bar", body["foo"])
	}
}

func TestRequestLogger_SanitizesSensitiveFields(t *testing.T) {
	cap := &captureHandler{}
	log := slog.New(logger.NewContextHandler(cap))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := RequestLogger(log)(handler)

	payload := map[string]any{
		"email":    "test@example.com",
		"password": "secret123",
		"token":    "abc",
	}
	bodyBytes, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	mw.ServeHTTP(rec, req)

	if len(cap.records) != 1 {
		t.Fatalf("expected 1 log record, got %d", len(cap.records))
	}

	attrs := attrsToMap(cap.records[0])
	body, ok := attrs["body"].(map[string]any)
	if !ok {
		t.Fatalf("expected body map, got %T", attrs["body"])
	}

	if body["email"] != "test@example.com" {
		t.Errorf("email should be preserved, got %v", body["email"])
	}
	if body["password"] != "[REDACTED]" {
		t.Errorf("password should be redacted, got %v", body["password"])
	}
	if body["token"] != "[REDACTED]" {
		t.Errorf("token should be redacted, got %v", body["token"])
	}
}

func TestRequestLogger_SkipsBodyForGet(t *testing.T) {
	cap := &captureHandler{}
	log := slog.New(logger.NewContextHandler(cap))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := RequestLogger(log)(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user/me", nil)
	rec := httptest.NewRecorder()

	mw.ServeHTTP(rec, req)

	attrs := attrsToMap(cap.records[0])
	if _, ok := attrs["body"]; ok {
		t.Error("GET request should not log body")
	}
}

func TestRequestLogger_ContextPropagation(t *testing.T) {
	cap := &captureHandler{}
	log := slog.New(logger.NewContextHandler(cap))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate downstream handler reading request_id from context
		id := logger.RequestIDFromContext(r.Context())
		if id == "" {
			t.Error("request_id missing in handler context")
		}
		svc := logger.ServiceFromContext(r.Context())
		if svc != "voucher" {
			t.Errorf("service = %q, want voucher", svc)
		}
		w.WriteHeader(http.StatusOK)
	})

	// Chain RequestID → RequestLogger → handler
	mw := RequestID()(RequestLogger(log)(handler))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/voucher/me", nil)
	rec := httptest.NewRecorder()

	mw.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}

	// Verify response header
	if rec.Header().Get("X-Request-ID") == "" {
		t.Error("expected X-Request-ID response header")
	}
}

func TestExtractClientIP(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		remote   string
		expected string
	}{
		{
			name:     "X-Forwarded-For single",
			headers:  map[string]string{"X-Forwarded-For": "203.0.113.1"},
			remote:   "192.168.1.1",
			expected: "203.0.113.1",
		},
		{
			name:     "X-Forwarded-For multiple",
			headers:  map[string]string{"X-Forwarded-For": "203.0.113.1, 198.51.100.2"},
			remote:   "192.168.1.1",
			expected: "203.0.113.1",
		},
		{
			name:     "X-Real-Ip",
			headers:  map[string]string{"X-Real-Ip": "203.0.113.2"},
			remote:   "192.168.1.1",
			expected: "203.0.113.2",
		},
		{
			name:     "fallback to RemoteAddr",
			headers:  map[string]string{},
			remote:   "192.168.1.1:1234",
			expected: "192.168.1.1:1234",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remote
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			got := extractClientIP(req)
			if got != tt.expected {
				t.Errorf("extractClientIP() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestResponseWriter_CapturesStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := newResponseWriter(rec)

	if rw.status != http.StatusOK {
		t.Errorf("default status = %d, want 200", rw.status)
	}

	rw.WriteHeader(http.StatusNotFound)
	if rw.status != http.StatusNotFound {
		t.Errorf("status after WriteHeader = %d, want 404", rw.status)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("recorder code = %d, want 404", rec.Code)
	}
}

func TestShouldLogBody(t *testing.T) {
	if !shouldLogBody(http.MethodPost) {
		t.Error("POST should log body")
	}
	if !shouldLogBody(http.MethodPut) {
		t.Error("PUT should log body")
	}
	if !shouldLogBody(http.MethodPatch) {
		t.Error("PATCH should log body")
	}
	if shouldLogBody(http.MethodGet) {
		t.Error("GET should not log body")
	}
	if shouldLogBody(http.MethodDelete) {
		t.Error("DELETE should not log body")
	}
}

func TestRequestLogger_DurationIsPositive(t *testing.T) {
	cap := &captureHandler{}
	log := slog.New(logger.NewContextHandler(cap))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	mw := RequestLogger(log)(handler)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	mw.ServeHTTP(rec, req)

	attrs := attrsToMap(cap.records[0])
	ms, ok := attrs["duration_ms"].(int64)
	if !ok {
		t.Fatalf("duration_ms type = %T, want int64", attrs["duration_ms"])
	}
	if ms < 1 {
		t.Errorf("duration_ms = %d, expected > 0", ms)
	}
}
