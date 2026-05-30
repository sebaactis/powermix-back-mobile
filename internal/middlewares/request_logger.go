package middlewares

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/sebaactis/powermix-back-mobile/internal/platform/logger"
)

// sensitiveFields son campos que nunca se loguean en el body del request.
var sensitiveFields = map[string]bool{
	"password": true, "currentPassword": true, "newPassword": true,
	"confirmPassword": true, "token": true, "refreshToken": true,
	"accessToken": true, "access_token": true, "refresh_token": true,
}

// responseWriter envuelve http.ResponseWriter para capturar el status code.
type responseWriter struct {
	http.ResponseWriter
	status int
	wrote  bool
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, status: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.wrote {
		rw.status = code
		rw.wrote = true
	}
	rw.ResponseWriter.WriteHeader(code)
}

// RequestLogger loguea cada request como JSON estructurado con request_id,
// service, method, path, status, duration_ms y body sanitizado.
func RequestLogger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			service := logger.ServiceFromPath(r.URL.Path)
			ctx := logger.SetService(r.Context(), service)
			r = r.WithContext(ctx)

			// Capturamos el body para loguear (solo POST/PUT/PATCH)
			var bodyMap map[string]any
			if shouldLogBody(r.Method) {
				bodyMap = readAndRestoreBody(r)
			}

			rw := newResponseWriter(w)
			next.ServeHTTP(rw, r)

			duration := time.Since(start).Milliseconds()
			reqID := logger.RequestIDFromContext(ctx)

			attrs := []any{
				"request_id", reqID,
				"service", service,
				"method", r.Method,
				"path", r.URL.Path,
				"status", rw.status,
				"duration_ms", duration,
				"ip", extractClientIP(r),
			}

			if bodyMap != nil {
				sanitized := sanitizeBody(bodyMap)
				if len(sanitized) > 0 {
					attrs = append(attrs, "body", sanitized)
				}
			}

			log.InfoContext(ctx, "request", attrs...)
		})
	}
}

func shouldLogBody(method string) bool {
	return method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch
}

// readAndRestoreBody lee el body del request para loguear y lo restaura
// así los handlers de abajo pueden leerlo de nuevo.
func readAndRestoreBody(r *http.Request) map[string]any {
	if r.Body == nil || r.ContentLength == 0 {
		return nil
	}

	// Limitamos la lectura del body a 10KB
	const maxBodySize = 10 * 1024
	body, err := io.ReadAll(io.LimitReader(r.Body, maxBodySize))
	if err != nil {
		return nil
	}
	r.Body.Close()
	r.Body = io.NopCloser(bytes.NewReader(body))

	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		return nil
	}
	return m
}

// sanitizeBody saca los campos sensibles del body que se loguea.
func sanitizeBody(m map[string]any) map[string]any {
	result := make(map[string]any, len(m))
	for k, v := range m {
		if sensitiveFields[k] {
			result[k] = "[REDACTED]"
		} else {
			result[k] = v
		}
	}
	return result
}

// extractClientIP extrae la IP del cliente, respetando X-Forwarded-For (para Render.com).
func extractClientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		// X-Forwarded-For puede ser una lista separada por comas
		if idx := strings.Index(fwd, ","); idx != -1 {
			return strings.TrimSpace(fwd[:idx])
		}
		return fwd
	}
	if real := r.Header.Get("X-Real-Ip"); real != "" {
		return real
	}
	return r.RemoteAddr
}
