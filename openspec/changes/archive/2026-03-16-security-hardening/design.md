# Diseño: security-hardening

## Enfoque técnico

Cinco archivos modificados, sin cambios de contrato de API externo ni migraciones de BD.
Los cambios son quirúrgicos: se alteran las firmas de `config.Load()` y `jwt.NewJWT()`
para que retornen error, y se propagan esos errores hacia `main.go`. El resto son
reemplazos de llamadas de logging uno a uno.

---

## Decisiones de arquitectura

### Decisión: config.Load() retorna (Config, error)

**Elección**: `func Load() (Config, error)` con función interna `validate()` que itera
los campos obligatorios.

**Alternativas consideradas**:
- `log.Fatal` dentro de `config.Load()` — descartado: hace la función no testeable y
  mezcla responsabilidades (configuración vs arranque).
- Panics — descartado: los panics son para errores de programación, no de entorno.
- Mantener strings `"ENV_X_NOT_SET"` — descartado: es el problema que estamos resolviendo.

**Rationale**: Retornar `error` sigue el patrón idiomático de Go, permite testing y
desplaza la decisión de "qué hacer con el error" al caller (`main.go`).

---

### Decisión: jwt.NewJWT() retorna (*JWT, error)

**Elección**: `func NewJWT() (*JWT, error)` — retorna error si `JWT_SECRET` o
`JWT_RECOVERY_PASS_SECRET` están vacíos.

**Alternativas consideradas**:
- Pasar los secrets como parámetros de función — válido, pero rompe la firma actual
  usada en `main.go` y agrega acoplamiento entre config y jwt.
- Leer de `Config` struct en vez de `os.Getenv` — razonable para el futuro, pero
  requeriría un refactor más amplio. Fuera de scope de este change.

**Rationale**: Cambio mínimo de firma que elimina el fallback inseguro. Los TTLs
mantienen sus defaults (60/15/1440 min) porque son configuración operativa,
no secretos de seguridad.

---

### Decisión: Logs con campos explícitos en lugar de %+v

**Elección**: Reemplazar `log.Printf("... %+v", struct)` por logs con campos
individuales no sensibles: `id=`, `email=`, `provider=`.

**Alternativas consideradas**:
- Eliminar los logs completamente — descartado: tienen valor diagnóstico legítimo
  para debugging de flujos OAuth.
- Migrar a `slog` en este change — deseable a futuro, pero implicaría un refactor
  mayor de todo el logging. Fuera de scope. Se mantiene `log.Printf` con campos seguros
  para los logs en `repository.go`.

**Rationale**: Mínima intervención que elimina la exposición de datos sensibles.

---

### Decisión: fmt.Println reemplazado por slog

**Elección**: `slog.Info` / `slog.Error` en `auth/handler.go`.

**Rationale**: `slog` ya está en uso en el proyecto (middlewares). Consistencia.
`fmt.Println` no tiene nivel, no tiene campos estructurados, y va a stdout sin
pasar por el sistema de logs.

---

## Flujo de datos / Cambios de firma

```
Antes:
  config.Load() → Config
  jwt.NewJWT()  → *JWT

Después:
  config.Load() → (Config, error)
  jwt.NewJWT()  → (*JWT, error)

main.go:
  cfg, err := config.Load()
  if err != nil { log.Fatal(err) }

  jwt, err := jwtx.NewJWT()
  if err != nil { log.Fatal(err) }
```

---

## Cambios por archivo

| Archivo | Acción | Descripción |
|---------|--------|-------------|
| `internal/platform/config/config.go` | Modificar | `Load()` → `(Config, error)`. Agregar función `validate()` que retorna error si alguna var obligatoria está vacía. Eliminar `getEnv` (ya no es necesaria con validación). |
| `internal/security/jwt/jwt.go` | Modificar | `NewJWT()` → `(*JWT, error)`. Eliminar fallbacks `"dev-secret"` y `"dev-reset-secret"`. Retornar error si secrets vacíos. |
| `cmd/api/main.go` | Modificar | Propagar los nuevos errores de `config.Load()` y `jwtx.NewJWT()` con `log.Fatal`. |
| `internal/domain/entities/user/repository.go` | Modificar | Reemplazar 3 `log.Printf("%+v", newUser)` por logs con campos explícitos (`id`, `email`, `provider`). |
| `internal/security/auth/handler.go` | Modificar | (1) Eliminar `log.Printf("🧪 ... %+v", userInfo)` → `slog.Info` con campos seguros. (2) Eliminar `fmt.Println(user.Email, resetURL)`. (3) Reemplazar `fmt.Println("HUBO UN ERROR...")` → `slog.Error`. (4) `UnlockUser`: reemplazar `json.NewEncoder` por `utils.WriteSuccess`. |

---

## Implementación detallada

### config/config.go

```go
// Vars obligatorias a validar
var requiredEnvVars = []struct {
    key   string
    field string
}{
    {"HTTP_ADDR", "HTTPAddr"},
    {"DB_DRIVER", "Driver"},
    {"DSN", "DSN"},
    {"MERCAGO_PAGO_TOKEN", "MercagoPagoToken"},
    {"COFFEJI_KEY", "CoffejiKey"},
    {"COFFEJI_SECRET", "CoffejiSecret"},
    {"RESEND_API_KEY", "ResendKey"},
    {"JWT_REFRESH_HASH", "HashToken"},
}

func Load() (Config, error) {
    cfg := Config{
        HTTPAddr:         os.Getenv("HTTP_ADDR"),
        Driver:           os.Getenv("DB_DRIVER"),
        DSN:              os.Getenv("DSN"),
        MercagoPagoToken: os.Getenv("MERCAGO_PAGO_TOKEN"),
        CoffejiKey:       os.Getenv("COFFEJI_KEY"),
        CoffejiSecret:    os.Getenv("COFFEJI_SECRET"),
        ResendKey:        os.Getenv("RESEND_API_KEY"),
        HashToken:        os.Getenv("JWT_REFRESH_HASH"),
    }
    if err := cfg.validate(); err != nil {
        return Config{}, err
    }
    return cfg, nil
}

func (c Config) validate() error {
    // itera los campos y retorna error con el nombre de la var que falta
}
```

### jwt/jwt.go

```go
func NewJWT() (*JWT, error) {
    sec := os.Getenv("JWT_SECRET")
    if sec == "" {
        return nil, errors.New("JWT_SECRET es requerido")
    }
    resetSec := os.Getenv("JWT_RECOVERY_PASS_SECRET")
    if resetSec == "" {
        return nil, errors.New("JWT_RECOVERY_PASS_SECRET es requerido")
    }
    // ... TTL parsing igual que antes ...
    return &JWT{...}, nil
}
```

### user/repository.go — logs seguros

```go
// Antes:
log.Printf("ℹ️ Usuario ya existe con OAuth: %+v", newUser)
log.Printf("🔁 Usuario existente actualizado con OAuth: %+v", newUser)
log.Printf("✅ Usuario nuevo con OAuth creado: %+v", newUser)

// Después:
log.Printf("ℹ️ Usuario ya existe con OAuth: id=%s email=%s provider=%s", newUser.ID, newUser.Email, newUser.OAuthProvider)
log.Printf("🔁 Usuario existente actualizado con OAuth: id=%s email=%s", newUser.ID, newUser.Email)
log.Printf("✅ Usuario nuevo con OAuth creado: id=%s email=%s", newUser.ID, newUser.Email)
```

### auth/handler.go

```go
// Antes (línea 78):
log.Printf("🧪 Datos del usuario de Google: %+v\n", userInfo)
// Después:
slog.Info("OAuth Google login", "email", userInfo.Email, "provider", userInfo.Provider)

// Antes (línea 222):
fmt.Println(user.Email, resetURL)
// Después: eliminado completamente

// Antes (línea 227):
fmt.Println("HUBO UN ERROR PARA MANDAR EL MAIL")
// Después:
slog.Error("error al enviar email de recovery", "email", user.Email, "error", err)

// Antes (líneas 243-244):
w.WriteHeader(http.StatusOK)
json.NewEncoder(w).Encode("User unlocked")
// Después:
utils.WriteSuccess(w, http.StatusOK, map[string]any{"message": "User unlocked"})
```

---

## Estrategia de testing

| Capa | Qué testear | Cómo |
|------|-------------|------|
| Unit | `config.Load()` retorna error cuando falta una var | `t.Setenv` + `config.Load()` |
| Unit | `jwt.NewJWT()` retorna error si JWT_SECRET vacío | `t.Setenv("")` + `NewJWT()` |
| Build | Compilación de `main.go` con nuevas firmas | `go build ./...` |
| Manual | Ausencia de `fmt.Println` en el código | `grep -r "fmt.Println" ./internal` |

No hay tests de integración existentes para estos flujos. Los tests unitarios nuevos
son simples y rápidos de agregar para `config` y `jwt`.

---

## Migración / Rollout

Sin migración de base de datos. Sin feature flags.

En entornos de CI donde las variables de entorno no estén configuradas, el build
va a fallar al intentar arrancar. Asegurarse de que el pipeline de CI tenga las
vars necesarias (o use un `.env.test`).

---

## Preguntas abiertas

- Ninguna que bloquee la implementación.
