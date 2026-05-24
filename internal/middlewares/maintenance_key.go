package middlewares

import (
	"crypto/subtle"
	"net/http"
)

// MaintenanceKeyConfig es la interfaz que necesita el middleware
// para decidir si aplicar la verificación de clave admin y qué clave aceptar.
type MaintenanceKeyConfig interface {
	IsProdeEnabled() bool
	IsMaintenanceEnabled() bool
	AdminAPIKey() string
}

// MaintenanceKey devuelve un middleware HTTP que exige el header
// X-Prode-Admin-Key cuando el modo mantenimiento de PRODE está activo.
//
// Si PRODE está deshabilitado o el mantenimiento deshabilitado, el middleware
// deja pasar sin verificar. Con mantenimiento activo, exige que el request
// incluya un header X-Prode-Admin-Key cuyo valor coincida con la clave
// configurada (comparación en tiempo constante). Claves faltantes o
// incorrectas reciben respuesta 401.
func MaintenanceKey(cfg MaintenanceKeyConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !cfg.IsProdeEnabled() || !cfg.IsMaintenanceEnabled() {
				next.ServeHTTP(w, r)
				return
			}

			key := r.Header.Get("X-Prode-Admin-Key")
			if key == "" || subtle.ConstantTimeCompare([]byte(key), []byte(cfg.AdminAPIKey())) != 1 {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
