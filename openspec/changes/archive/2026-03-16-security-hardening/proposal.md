# Propuesta: security-hardening

## Intención

El backend expone datos sensibles en logs de producción, arranca silenciosamente con
secrets JWT hardcodeados si las variables de entorno no están configuradas, y tiene
llamadas `fmt.Println` de debug que nunca deberían llegar a producción.
Estos problemas constituyen riesgos de seguridad reales: un atacante con acceso a logs
puede obtener hashes de contraseñas e IDs de OAuth; un deploy mal configurado puede
arrancar con secrets predecibles.

Este change elimina todas esas superficies de riesgo sin alterar el comportamiento
funcional del sistema.

## Alcance

### En Scope
- Reemplazar `log.Printf("%+v", newUser)` por logs estructurados que **no** incluyan campos sensibles (`password`, `oauth_id`) — `user/repository.go` líneas 109, 120, 126
- Reemplazar `log.Printf("🧪 Datos del usuario de Google: %+v\n", userInfo)` por log no sensible — `auth/handler.go` línea 78
- Eliminar `fmt.Println(user.Email, resetURL)` — `auth/handler.go` línea 222
- Eliminar `fmt.Println("HUBO UN ERROR PARA MANDAR EL MAIL")` y reemplazar por `slog.Error` — `auth/handler.go` línea 227
- Agregar validación de vars obligatorias en `config.Load()` que retorne `error` y haga fallar el arranque si alguna falta — `config/config.go`
- Hacer que `jwt.NewJWT()` retorne `error` si `JWT_SECRET` o `JWT_RECOVERY_PASS_SECRET` están vacíos, en lugar de usar fallbacks — `jwt/jwt.go` líneas 30-38
- Actualizar `main.go` para manejar los nuevos errores de `config.Load()` y `jwt.NewJWT()`
- Usar `utils.WriteSuccess` en `UnlockUser` en lugar de `json.NewEncoder` directo — `auth/handler.go` líneas 243-244

### Fuera de Scope
- Rotación de secrets en runtime
- Sistema de secretos externos (Vault, AWS Secrets Manager)
- Cambios en la estructura de `Config` más allá de la validación
- Cambios en otros handlers o repositorios que no sean los listados

## Enfoque

1. **Config**: agregar una función `validate() error` en el paquete config que itere los campos obligatorios. `Load()` pasa a retornar `(Config, error)`. El `main.go` hace `log.Fatal` si hay error.
2. **JWT**: `NewJWT()` pasa a retornar `(*JWT, error)`. Si alguno de los dos secrets está vacío, retorna `errors.New("JWT_SECRET requerido")` (ídem para recovery). El `main.go` hace `log.Fatal` si hay error.
3. **Logs sensibles**: reemplazar `%+v` por campos explícitos: `id=`, `email=`, `provider=`. Nunca loggear `password` ni `oauth_id`.
4. **fmt.Println**: reemplazar por `slog.Info` / `slog.Error` con campos clave-valor. Eliminar el que loggea `resetURL` completa (contiene el token de recovery).
5. **UnlockUser**: reemplazar `json.NewEncoder` por `utils.WriteSuccess`.

## Áreas afectadas

| Archivo | Impacto | Descripción |
|---------|---------|-------------|
| `internal/platform/config/config.go` | Modificado | `Load()` retorna `(Config, error)`, validación de obligatorias |
| `internal/security/jwt/jwt.go` | Modificado | `NewJWT()` retorna `(*JWT, error)`, falla si secrets vacíos |
| `cmd/api/main.go` | Modificado | Maneja nuevos errores de config y jwt |
| `internal/domain/entities/user/repository.go` | Modificado | Reemplaza 3 `log.Printf(%+v)` por logs sin campos sensibles |
| `internal/security/auth/handler.go` | Modificado | Elimina `fmt.Println` (×2), reemplaza log de OAuth, unifica `UnlockUser` |

## Riesgos

| Riesgo | Probabilidad | Mitigación |
|--------|-------------|------------|
| `main.go` no compila si la firma de `NewJWT()` o `Load()` cambia | Alta (es intencional) | Actualizar `main.go` en el mismo change |
| Algún test existente instancia `NewJWT()` sin env vars | Baja | Revisar `*_test.go` antes de aplicar; ajustar si es necesario |
| Pérdida de trazabilidad de OAuth al reducir los logs | Baja | Los nuevos logs con campos explícitos son igual de útiles para debugging |

## Plan de rollback

```bash
git revert HEAD  # si el change fue en un único commit
# o
git checkout origin/develop -- internal/platform/config/config.go \
  internal/security/jwt/jwt.go \
  cmd/api/main.go \
  internal/domain/entities/user/repository.go \
  internal/security/auth/handler.go
```

No hay migraciones de base de datos ni cambios de contrato de API. El rollback es trivial.

## Dependencias

- Ninguna externa. El change es autocontenido.

## Criterios de éxito

- [ ] `go build ./...` pasa sin errores
- [ ] `go test ./...` pasa sin errores
- [ ] Ningún archivo del proyecto contiene `fmt.Println`
- [ ] Ningún `log.Printf` loggea un struct completo con `%+v` que incluya `User`
- [ ] `config.Load()` retorna error si `DSN` o `JWT_SECRET` no están seteados
- [ ] `jwt.NewJWT()` retorna error si `JWT_SECRET` o `JWT_RECOVERY_PASS_SECRET` están vacíos
- [ ] `UnlockUser` usa `utils.WriteSuccess`
