package logger

import (
	"context"
	"log/slog"
)

// ContextHandler envuelve un slog.Handler e inyecta request_id y service
// desde el contexto en cada log record automáticamente.
type ContextHandler struct {
	slog.Handler
}

func NewContextHandler(h slog.Handler) *ContextHandler {
	return &ContextHandler{Handler: h}
}

func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if id := RequestIDFromContext(ctx); id != "" {
		r.AddAttrs(slog.String("request_id", id))
	}
	if svc := ServiceFromContext(ctx); svc != "" {
		r.AddAttrs(slog.String("service", svc))
	}
	return h.Handler.Handle(ctx, r)
}

func (h *ContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewContextHandler(h.Handler.WithAttrs(attrs))
}

func (h *ContextHandler) WithGroup(name string) slog.Handler {
	return NewContextHandler(h.Handler.WithGroup(name))
}
