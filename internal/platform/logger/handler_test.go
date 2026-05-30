package logger

import (
	"context"
	"log/slog"
	"testing"
)

// captureHandler records all handled slog Records for inspection.
type captureHandler struct {
	records []slog.Record
	attrs   []slog.Attr
}

func (h *captureHandler) Enabled(context.Context, slog.Level) bool { return true }
func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	for _, a := range h.attrs {
		r.AddAttrs(a)
	}
	h.records = append(h.records, r)
	return nil
}
func (h *captureHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)
	return &captureHandler{records: h.records, attrs: newAttrs}
}
func (h *captureHandler) WithGroup(string) slog.Handler { return h }

func TestContextHandler_InjectsRequestIDAndService(t *testing.T) {
	base := &captureHandler{}
	h := NewContextHandler(base)
	logger := slog.New(h)

	ctx := SetRequestID(SetService(context.Background(), "proof"), "req-123")
	logger.InfoContext(ctx, "hello")

	if len(base.records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(base.records))
	}

	rec := base.records[0]
	attrs := attrsToMap(rec)

	if got := attrs["request_id"]; got != "req-123" {
		t.Errorf("request_id = %q, want req-123", got)
	}
	if got := attrs["service"]; got != "proof" {
		t.Errorf("service = %q, want proof", got)
	}
}

func TestContextHandler_NoContextValues(t *testing.T) {
	base := &captureHandler{}
	h := NewContextHandler(base)
	logger := slog.New(h)

	logger.InfoContext(context.Background(), "hello")

	if len(base.records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(base.records))
	}

	rec := base.records[0]
	attrs := attrsToMap(rec)

	if _, ok := attrs["request_id"]; ok {
		t.Error("expected no request_id when context empty")
	}
	if _, ok := attrs["service"]; ok {
		t.Error("expected no service when context empty")
	}
}

func attrsToMap(r slog.Record) map[string]string {
	m := make(map[string]string)
	r.Attrs(func(a slog.Attr) bool {
		m[a.Key] = a.Value.String()
		return true
	})
	return m
}
