# Tasks: security-hardening

## Fase 1: Infraestructura — cambios de firma y validación

- [ ] 1.1 Modificar `internal/platform/config/config.go`: cambiar `Load()` a `(Config, error)`, agregar método `validate()` que retorna error si alguna var obligatoria está vacía o ausente. Eliminar la función `getEnv()`.
- [ ] 1.2 Modificar `internal/security/jwt/jwt.go`: cambiar `NewJWT()` a `(*JWT, error)`. Eliminar los fallbacks `"dev-secret"` y `"dev-reset-secret"`. Retornar `errors.New("JWT_SECRET es requerido")` si el secret está vacío. Ídem para `JWT_RECOVERY_PASS_SECRET`.
- [ ] 1.3 Modificar `cmd/api/main.go`: actualizar la llamada a `config.Load()` para recibir `(cfg, err)` y hacer `log.Fatal(err)` si hay error. Ídem para `jwtx.NewJWT()`.

## Fase 2: Limpieza de logs sensibles

- [ ] 2.1 Modificar `internal/domain/entities/user/repository.go` línea 109: reemplazar `log.Printf("ℹ️ Usuario ya existe con OAuth: %+v", newUser)` por `log.Printf("ℹ️ Usuario ya existe con OAuth: id=%s email=%s provider=%s", newUser.ID, newUser.Email, newUser.OAuthProvider)`.
- [ ] 2.2 Modificar `internal/domain/entities/user/repository.go` línea 120: reemplazar `log.Printf("🔁 Usuario existente actualizado con OAuth: %+v", newUser)` por `log.Printf("🔁 Usuario existente actualizado con OAuth: id=%s email=%s", newUser.ID, newUser.Email)`.
- [x] 2.3 Modificar `internal/domain/entities/user/repository.go` línea 126: reemplazar `log.Printf("✅ Usuario nuevo con OAuth creado: %+v", newUser)` por `log.Printf("✅ Usuario nuevo con OAuth creado: id=%s email=%s", newUser.ID, newUser.Email)`.
- [x] 2.4 Modificar `internal/security/auth/handler.go` línea 78: reemplazar `log.Printf("🧪 Datos del usuario de Google: %+v\n", userInfo)` por `slog.Info("OAuth Google login", "email", userInfo.Email, "provider", userInfo.Provider)`. Agregar `"log/slog"` al bloque de imports si no está.
- [x] 2.5 Modificar `internal/security/auth/handler.go` línea 222: eliminar completamente `fmt.Println(user.Email, resetURL)`.
- [x] 2.6 Modificar `internal/security/auth/handler.go` línea 227: reemplazar `fmt.Println("HUBO UN ERROR PARA MANDAR EL MAIL")` por `slog.Error("error al enviar email de recovery", "email", user.Email, "error", err)`.
- [x] 2.7 Verificar y limpiar imports en `auth/handler.go`: eliminar `"fmt"` si ya no se usa, eliminar `"log"` si ya no se usa, agregar `"log/slog"` si se necesita.

## Fase 3: Consistencia de respuestas HTTP

- [ ] 3.1 Modificar `internal/security/auth/handler.go` función `UnlockUser` (líneas 242-244): reemplazar `w.WriteHeader(http.StatusOK)` + `json.NewEncoder(w).Encode("User unlocked")` por `utils.WriteSuccess(w, http.StatusOK, map[string]any{"message": "User unlocked"})`.
- [ ] 3.2 Verificar imports en `auth/handler.go`: eliminar `"encoding/json"` si ya no se usa en ningún otro lugar del archivo.

## Fase 4: Tests

- [ ] 4.1 Crear `internal/platform/config/config_test.go`: test que verifica que `Load()` retorna error cuando `DSN` no está seteada (`t.Setenv("DSN", "")`).
- [ ] 4.2 Agregar test en `config_test.go`: verifica que `Load()` retorna error cuando `JWT_REFRESH_HASH` no está seteada.
- [ ] 4.3 Agregar test en `config_test.go`: verifica que `Load()` retorna `nil` como error cuando todas las vars están presentes.
- [ ] 4.4 Crear `internal/security/jwt/jwt_test.go`: test que verifica que `NewJWT()` retorna error cuando `JWT_SECRET` está vacío.
- [ ] 4.5 Agregar test en `jwt_test.go`: verifica que `NewJWT()` retorna error cuando `JWT_RECOVERY_PASS_SECRET` está vacío.
- [ ] 4.6 Agregar test en `jwt_test.go`: verifica que `NewJWT()` retorna `*JWT` válido cuando ambos secrets están seteados.

## Fase 5: Verificación final

- [ ] 5.1 Ejecutar `go build ./...` y confirmar que compila sin errores.
- [ ] 5.2 Ejecutar `go test ./...` y confirmar que todos los tests pasan.
- [ ] 5.3 Verificar ausencia de `fmt.Println` en el código: `grep -r "fmt.Println" ./internal` debe retornar vacío.
- [ ] 5.4 Verificar ausencia de logs con `%+v` sobre structs User: `grep -rn "%+v" ./internal` no debe mostrar ninguna línea que loggee structs completos de entidades de dominio.
- [ ] 5.5 Verificar que `"dev-secret"` no aparece en ningún archivo Go: `grep -r "dev-secret" .` debe retornar vacío.
