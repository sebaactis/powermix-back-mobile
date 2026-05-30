package middlewares

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/sebaactis/powermix-back-mobile/internal/platform/logger"
)

// RequestID genera un UUID por request, lo guarda en el contexto
// y lo devuelve en el header X-Request-ID.
func RequestID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := uuid.New().String()
			ctx := logger.SetRequestID(r.Context(), id)
			w.Header().Set("X-Request-ID", id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
