package middlewares

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/sebaactis/powermix-back-mobile/internal/utils"
)

const timeoutResponseBody = `{"success":false,"data":null,"error":{"code":"ERR_TIMEOUT","message":"La solicitud tardó demasiado"}}`

func JSONContentType() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			next.ServeHTTP(w, r)
		})
	}
}

// Recoverer atrapa los panics y devuelve un envelope de error estándar de la API.
func Recoverer(logger *slog.Logger) func(http.Handler) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.ErrorContext(r.Context(), "panic recovered",
						"panic", rec,
						"method", r.Method,
						"path", r.URL.Path,
					)
					utils.WriteError(w, http.StatusInternalServerError, utils.WriteErrorOpts{
						Code:    utils.ErrCodeInternal,
						Message: "Error interno del servidor",
					})
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// Timeout corta las requests lentas con el envelope de error estándar de la API.
func Timeout(d time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.TimeoutHandler(next, d, timeoutResponseBody)
	}
}
