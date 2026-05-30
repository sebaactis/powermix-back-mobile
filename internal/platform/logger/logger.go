package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

// ctxKey es el tipo de key para los valores del contexto.
type ctxKey string

const requestIDKey ctxKey = "request_id"
const serviceKey ctxKey = "service"

// Logger envuelve slog extrayendo automáticamente request_id y service.
type Logger struct {
	inner *slog.Logger
}

// New crea un logger JSON para producción (compatible con Render.com).
func New() *Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	return &Logger{inner: slog.New(handler)}
}

// Info loguea a nivel INFO. Extrae request_id y service del contexto.
func (l *Logger) Info(ctx context.Context, msg string, attrs ...any) {
	l.inner.InfoContext(ctx, msg, l.withContext(ctx, attrs...)...)
}

// Warn loguea a nivel WARN. Extrae request_id y service del contexto.
func (l *Logger) Warn(ctx context.Context, msg string, attrs ...any) {
	l.inner.WarnContext(ctx, msg, l.withContext(ctx, attrs...)...)
}

// Error loguea a nivel ERROR. Extrae request_id y service del contexto.
func (l *Logger) Error(ctx context.Context, msg string, attrs ...any) {
	l.inner.ErrorContext(ctx, msg, l.withContext(ctx, attrs...)...)
}

// withContext le agrega request_id y service del contexto a los atributos.
func (l *Logger) withContext(ctx context.Context, attrs ...any) []any {
	result := make([]any, 0, len(attrs)+4)

	if id := RequestIDFromContext(ctx); id != "" {
		result = append(result, "request_id", id)
	}
	if svc := ServiceFromContext(ctx); svc != "" {
		result = append(result, "service", svc)
	}

	return append(result, attrs...)
}

// SetRequestID guarda el request ID en el contexto.
func SetRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// RequestIDFromContext recupera el request ID del contexto.
func RequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

// SetService guarda el nombre del service en el contexto.
func SetService(ctx context.Context, service string) context.Context {
	return context.WithValue(ctx, serviceKey, service)
}

// ServiceFromContext recupera el nombre del service del contexto.
func ServiceFromContext(ctx context.Context) string {
	if svc, ok := ctx.Value(serviceKey).(string); ok {
		return svc
	}
	return ""
}

// ServiceFromPath extrae el nombre del service desde el path HTTP.
// ej: "/api/v1/proof" → "proof", "/api/v1/user" → "user"
func ServiceFromPath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 3 && parts[0] == "api" {
		return parts[2] // /api/v1/{service}/...
	}
	if len(parts) >= 1 {
		return parts[0]
	}
	return "unknown"
}
