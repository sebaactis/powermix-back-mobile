# Design: unified-error-handling

## Technical Approach

Three incremental phases: (1) backend error codes + middlewares, (2) error wrapping across layers + GORM→domain mapping, (3) frontend Error Boundary + Sentry. Each phase independently deployable. TDD-first: write test, then code.

## Architecture Decisions

### D1: Error code format

| Option | Tradeoff | Decision |
|--------|----------|----------|
| HTTP-status-derived (400, 404) | Not descriptive enough | ❌ |
| `ERR_` + SCREAMING_SNAKE | Machine-readable, grep-able, unambiguous | **✅ Adopted** |

### D2: Error codes location

| Option | Tradeoff | Decision |
|--------|----------|----------|
| New `internal/errors/` package | Extra import, over-engineered for 7 constants | ❌ |
| `internal/utils/response.go` | Co-located with APIError, zero new deps | **✅ Adopted** |

### D3: Middleware ordering (chi stack)

| Position | Role |
|----------|------|
| 1. **Recoverer** | Catches panics from ALL downstream layers |
| 2. Logger | Logs request info |
| 3. Timeout | Enforces deadline, returns APIResponse format |
| 4. Auth | Validates JWT, sets context |
| 5. Handler | Business logic |

### D4: Wrapping strategy

- **Repositories**: wrap GORM errors with `fmt.Errorf("entity: action id=%d: %w", id, err)` — add context, preserve cause chain
- **Services**: propagate wrapped errors directly, add business context only for branching logic
- **Handlers**: use `errors.Is()` to match sentinels → map to error codes → call `WriteError(code: ...)`
- **Raw GORM errors NEVER reach the response** — logged at source, replaced with `ErrInternal`

### D5: Error Boundary placement

Wrap `MainNavigator` in `app/index.tsx` (inside `AuthProvider`, not outside — so retry has valid auth state). Catches render crashes in any screen.

## Data Flow

```
DB (GORM) ──→ Repository ──→ Service ──→ Handler ──→ Response ──→ Frontend
   │             │              │            │             │            │
   │             %w wrap        %w wrap      errors.Is()   writeError()  parseError()
   │             + map sentinel + propagate  + code+msg    {code,msg}    check code→show UI
   │             + log raw err               + log details               fallback to msg
   └─ sql.ErrNoRows ──→ ErrRecordNotFound ──→ err ──→ ERR_NOT_FOUND ──→ "Usuario no encontrado"
```

## Interfaces / Contracts

### Error code constants (`internal/utils/response.go`)

```go
const (
    ErrCodeValidation     = "ERR_VALIDATION"
    ErrCodeNotFound       = "ERR_NOT_FOUND"
    ErrCodeDuplicateEntry = "ERR_DUPLICATE_ENTRY"
    ErrCodeInvalidCreds   = "ERR_INVALID_CREDENTIALS"
    ErrCodeUnauthorized   = "ERR_UNAUTHORIZED"
    ErrCodeTimeout        = "ERR_TIMEOUT"
    ErrCodeInternal       = "ERR_INTERNAL"
    ErrCodeExternalService = "ERR_EXTERNAL_SERVICE"
)
```

### APIError (modified)

```go
type APIError struct {
    Code    string      `json:"code"`
    Message string      `json:"message"`
    Fields  interface{} `json:"fields,omitempty"`
}

type WriteErrorOpts struct {
    Code    string
    Message string
    Fields  interface{}
}
func WriteError(w http.ResponseWriter, statusCode int, opts WriteErrorOpts) { ... }
```

### Middleware signatures

```go
// Recoverer — catches panics, returns 500 APIResponse
func Recoverer(logger *slog.Logger) func(http.Handler) http.Handler

// RequestID — injects request_id into context
func RequestID() func(http.Handler) http.Handler
```

### Frontend API response type (`src/helpers/apiHelper.ts`)

```ts
type ApiError = {
    code: string;          // NEW
    message: string;
    fields?: Record<string, string> | null;
} | null;
```

### handleApiError utility

```ts
function handleApiError<T>(res: ApiResponse<T>, fallback: string): string {
    return res?.error?.code
        ? mapCodeToMessage(res.error.code, res.error.message, fallback)
        : fallback;
}
```

## File Changes

### Phase 1 — Backend error codes + middlewares

| File | Action | Description |
|------|--------|-------------|
| `internal/utils/response.go` | Modify | Add `code` to `APIError`, define error constants, update `WriteError` signature |
| `internal/utils/response_test.go` | Create | Test `WriteError` produces `code` field in JSON |
| `internal/middlewares/mw.go` | Modify | Add `Recoverer()` middleware; fix `Timeout` to return `APIResponse` format |
| `internal/middlewares/mw_test.go` | Create | Test Recoverer catches panics, logs, returns correct JSON |
| `internal/domain/entities/user/handler.go` | Modify | Replace `err.Error()` leaks with Spanish messages + error codes |
| `internal/domain/entities/token/handler.go` | Modify | Same pattern |
| `internal/domain/entities/proof/handler.go` | Modify | Same pattern |
| `internal/domain/entities/voucher/handler.go` | Modify | Same pattern |
| `internal/domain/entities/prode/handler.go` | Modify | Add error codes to `WriteError` calls |
| `internal/middlewares/auth_middleware.go` | Modify | Add error codes |
| `internal/routes/router.go` | Modify | Register `Recoverer` middleware first |

### Phase 2 — Error wrapping

| File | Action | Description |
|------|--------|-------------|
| `internal/domain/entities/user/repository.go` | Modify | `%w` wrapping + GORM→domain mapping |
| `internal/domain/entities/user/service.go` | Modify | Add `ErrInternal` sentinel, `%w` through layers |
| `internal/domain/entities/proof/repository.go` | Modify | Same wrapping pattern |
| `internal/domain/entities/token/repository.go` | Modify | Same |
| `internal/domain/entities/voucher/repository.go` | Modify | Same |
| `internal/domain/entities/prode/repository.go` | Modify | Unify wrapping pattern (already maps ErrRecordNotFound) |
| `src/helpers/apiHelper.ts` | Modify | Add `code` to `ApiError` type, add `handleApiError` |

### Phase 3 — Frontend error handling

| File | Action | Description |
|------|--------|-------------|
| `src/components/ErrorBoundary.tsx` | Create | React Error Boundary with fallback UI + "Reintentar" |
| `app/index.tsx` | Modify | Wrap `MainNavigator` with `ErrorBoundary` |
| `internal/middlewares/mw.go` | Modify | Add `RequestID()` middleware |
| `internal/routes/router.go` | Modify | Register `RequestID` middleware |
| `cmd/api/main.go` | Modify | Register RequestID middleware (or already via router) |

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Utils unit | `WriteError` produces `code` field | Test JSON output, verify envelope shape |
| Middleware unit | Recoverer catches panic, logs, returns valid JSON | Chi `httptest` + `slog` buffer |
| Handler unit | Each handler maps error sentinels → correct code+HTTP status | Table-driven with `httptest.NewRecorder` |
| Repository unit | GORM errors wrapped + sentinel mapped | Mock `*gorm.DB` or in-memory SQLite |
| Frontend unit | `handleApiError` returns correct fallback | Jest/RNTL, test code→message mapping |
| Frontend render | Error Boundary catches throws, shows fallback | RNTL render + simulate throw |

## Migration / Rollout

Phase 1 is safe to ship alone — adding `code` to response is non-breaking (frontend uses optional chaining). Phase 2 changes internal error flow but keeps HTTP response shape identical. Phase 3 is purely additive (new component + dependency). Each phase must pass `go test ./...` independently. No feature flags needed.

## Open Questions

- [ ] Confirm Sentry org slug and DSN for Phase 3
- [ ] Decide if backend Sentry `slog` handler is Phase 3 or post-MVP
