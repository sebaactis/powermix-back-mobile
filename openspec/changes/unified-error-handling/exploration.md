## Exploration: unified-error-handling

### Current State

Error handling exists in both projects but is **inconsistent** across domains. The backend has a shared response utility (`utils.WriteError`) and domain-specific `var Err*` sentinels, but there's **no global error handling middleware**, **no panic recovery**, **no standardized error codes**, and **no contextual wrapping** of errors through the layers. The frontend has a well-structured API helper (`apiHelper.ts`) with automatic 401 refresh logic and toast display, but **no structured error code mapping** and **mixed handling** of the `error.fields` field.

---

### Backend Error Flow

#### Error Response Structure

**File**: `internal/utils/response.go`

- `APIResponse`: `{ success: bool, data: any, error: { message: string, fields?: any } }`
- `WriteSuccess(w, status, data)` — sets 200/201, writes JSON
- `WriteError(w, statusCode, message, fields)` — sets status, writes JSON with error field

**Good**: Consistent envelope across all handlers.

**Problems**:
1. No `code` field (only `message` + optional `fields`) — frontend can't programmatically decide what to display
2. `fields` is typed as `interface{}` — sometimes it's `map[string]string` (validation), sometimes `map[string]string{"error": err.Error()}`, sometimes just a raw string passed to `err` param (as in `SendEmailContact` handler)
3. `WriteError` silently swallows JSON encoding errors (`_ = json.NewEncoder(w).Encode(resp)`)
4. No function signature for domain-specific typed errors — every handler does its own `errors.Is` chain

#### Error Middleware

**Files in `internal/middlewares/`**: `mw.go` (Logger, JSONContentType, Timeout), `auth_middleware.go`, `ratelimit_middleware.go`, `maintenance_key.go`

**Missing**:
- **No panic recovery middleware** — a goroutine panic crashes the server. `http.Server` has no `PanicHandler` set.
- **No global error handler** — if a middleware panics or an unexpected error reaches the handler, there's no catch-all.
- **No structured logging middleware** that captures response status codes (the Logger only logs request, not response status).
- The `Timeout` middleware just returns `{"error":"timeout"}` — doesn't use the standard `APIResponse` format.

#### Error Types by Domain

**Prode** (`internal/domain/entities/prode/errors.go`):
- 13 sentinel errors defined — best-practice domain error module
- All use `errors.New(...)` with `prode:` prefix
- Some errors *leak to handler* via `err.Error()` (e.g., `ErrInvalidScore`, `ErrScoreOutOfRange`)
- Others are mapped to specific HTTP status codes in handler switch
- **Good**: Rich domain errors with clear semantic meaning
- **Missing**: Error wrapping through service→handler. Repository methods return raw `err` (GORM errors leak)

**User** (`internal/domain/entities/user/`):
- `ErrDuplicateEmail` — set in repository, checked in handler via `errors.Is`
- `ErrSameName` — set in service, checked in handler
- Repository methods return raw `errors.New("Usuario no encontrado")` / `errors.New("user not found")` — **inconsistent Spanish/English**, some leak to handler via `err.Error()`
- `isDuplicateKeyError()` inspects PostgreSQL error code 23505, GORM `ErrDuplicatedKey`, string contains — **fragile approach**

**Voucher** (`internal/domain/entities/voucher/`):
- 4 sentinel errors — clean, well-named (`ErrNoAvailableVouchers`, `ErrVoucherNotFound`, `ErrVoucherNotBelongsToUser`, `ErrVoucherNotUsed`)
- Handler correctly maps via `errors.Is` with specific HTTP codes
- **Good**: One of the cleanest error handling patterns in the codebase

**Auth** (`internal/security/auth/`):
- `ErrInvalidCredentials`, `ErrAccountLocked` — defined at package level
- Handler `handleLoginError` uses `switch` on sentinel values — correct pattern
- JWT client errors are generic (`"token invalido"`, `"firma invalida"`) — no wrapping

**Token** (`internal/domain/entities/token/`):
- `ErrRefreshReuseDetected`, `ErrRefreshInvalid` — good domain semantics
- Repository returns `errors.New("no se encontró el token proporcionado")` / `errors.New("error inesperado")` — **generic**, masks actual GORM errors

**Proof** (`internal/domain/entities/proof/`):
- No sentinel errors — uses `fmt.Errorf` inline throughout service
- Repository returns raw `result.Error` — GORM errors leak directly to service
- Service creates inline `fmt.Errorf("ya tenes guardado un comprobante con este ID: %s", ...)` — **no sentinels, no wrapping**
- Handler catches validation errors via `AsValidationError`, otherwise passes `err.Error()` to the response — **user sees raw DB/internal error messages**

**Clients**:
- **Coffeeji**: Returns `fmt.Errorf("status no exitoso: %s", ...)` — includes raw body. "respuesta no exitosa: code=%d, msg=%s" — OK for logging, but shouldn't reach user
- **MercadoPago**: Returns `fmt.Errorf("mercado pago error: %s", ...)` — raw API errors leak
- **Mailer**: Uses Resend SDK, returns raw SDK errors

#### Error Propagation

**Layer tracing** (Repository → Service → Handler):

```
Repository: returns raw gorm.Error or domain Err*
    ↓
Service: may wrap with context (but mostly doesn't), returns to handler
    ↓
Handler: uses errors.Is() chain, maps to WriteError with message + status
```

**Key findings**:
- **NO error wrapping** using `fmt.Errorf("...: %w", err)` in most places — context is lost
- Repository errors that aren't `gorm.ErrRecordNotFound` leak to the handler as generic "Error interno"
- `slog.Error` is called inconsistently in handlers — sometimes before WriteError, sometimes not
- No structured `slog` attributes for correlating errors (trace IDs, request IDs)

#### Logging

- Uses `log/slog` — standard Go structured logger
- `slog.Error()` is used in ~20 places across handler/service files
- No request ID propagation — can't correlate log lines to specific requests
- No error grouping or severity levels beyond slog's built-in
- `log.Printf` and `log.Fatal` are used in main.go and `proof/service.go` (mixed with slog)

---

### Frontend Error Flow

#### API Client Setup

**File**: `src/helpers/apiHelper.ts`

- Custom `fetch` wrapper with typed `ApiResponse<T>` = `{ success, data, error: { message, fields? }, status? }`
- Automatic `Bearer` token injection from runtime storage
- **Automatic 401 handling**: catches 401 → tries token refresh → retries original request. If refresh fails → clears auth and calls `onAuthFailed` callback
- **Refresh deduplication**: `refreshInFlight` singleton prevents parallel refresh calls
- **Network error catch**: Returns a generic `"Error de red al llamar a la API"` message

**Good**: Solid architecture with refresh token rotation, typed responses, and auth failure callback.

**Problems**:
1. `error` is typed as `{ message: string, fields?: Record<string, string> | null } | null` — but the actual backend response sends `fields` as `interface{}`, so it can be `string`, `map[string]string`, etc.
2. Network errors are caught with a bare `catch` — no distinction between timeout vs. DNS vs. HTTP errors
3. Fallback parsing when backend doesn't send standard envelope: creates `{ message: "HTTP error {status}" }` — no body

#### Error Interception

**File**: `src/helpers/authApi.ts` (thin wrapper over `ApiHelper` with `signOut` callback)

- No centralized error logging/capture (no Sentry, no Datadog, no error boundary telemetry)
- No global toast for HTTP errors — each screen decides when to show Toast
- No interceptor that could do global error side-effects (e.g., logging to analytics)

#### Error Display Mechanisms

**Toast system** (`components/toast/`):
- Custom `AppToast` component with `success`, `error`, `warning` variants
- Uses `react-native-toast-message` — registered in `toastConfig.tsx`
- Displays `text1` (title) and `text2` (message)
- **Good**: Visual design is clean, with icon circle, accent bars, colors

**Pattern in screens**:
```
const res = await AuthApi<PaginatedProofs>(url, "GET", signOut)
if (!res.success) {
    Toast.show({ type: "appError", text1: "Ocurrió un error", text2: res.error?.message })
    return
}
```

**Problems**:
1. Each screen duplicates the same error handling pattern — no abstraction
2. Some screens use `catch (e)` block with separate Toast call — mixing expected API errors with unexpected exceptions
3. `text2: res.error?.message` — displays raw backend message (could be technical/internal)
4. Error messages are **not translated** — mix of Spanish from backend and English from Toast defaults
5. No error visibility tracking — errors disappear after toast timeout with no audit trail

#### Error Boundaries
- **No React Error Boundaries** found in the project
- No error boundary wrapper in the navigation tree
- A React render crash would show the white screen of death

#### Error Code Mapping
- **No error code mapping** exists — frontend doesn't check specific `error.code` (because backend doesn't send one)
- All errors are treated as string messages
- No mapping of error codes to localized user-facing messages
- No distinction between validation errors (`error.fields`) and application errors (`error.message`)

#### Auth Error Handling
- **Login**: Shows Toast on failure with backend message
- **Google OAuth**: Shows `setError` inline text on login screen
- **Token refresh**: On failure → `clearAuthRuntime()` + `signOut()` → user gets logged out
- **401 retry**: Handled transparently in `ApiHelper`

---

### Gaps Found

1. **No standard error codes (backend + frontend)**: Backend sends `{ message, fields }` with no `code` field. Frontend can't programmatically decide UI action based on error type. Frontend screens duplicate error-handling logic.

2. **No panic recovery middleware (backend)**: A panic in any handler crashes the server. No `recover()` middleware exists.

3. **No global error handler (backend)**: No catch-all for unexpected errors. No centralized place to log + format all errors consistently.

4. **Raw error messages leak to users (backend)**: Several handlers pass `err.Error()` or `map[string]string{"error": err.Error()}` directly to `WriteError`. Users see technical DB/client error messages (MercadoPago, PostgreSQL, fmt.Errorf output).

5. **Inconsistent error wrapping (backend)**: Most services don't use `%w` to wrap errors. Context is lost when errors propagate from repository → service → handler.

6. **GORM errors leak (backend)**: Repository methods return `result.Error` directly without mapping to domain errors. `isDuplicateKeyError()` inspects PostgreSQL error codes — fragile.

7. **Mixed Spanish/English error messages (backend)**: Some errors are Spanish ("Usuario no encontrado"), others English ("user not found", "invalid json"). The frontend's toast defaults are also English.

8. **`fields` field type inconsistency (backend)**: `APIError.Fields` is `interface{}`. It's used as `map[string]string` (validation), `map[string]string{"error": ...}` (user handler), and sometimes as raw string (contact handler). Frontend can't safely consume it.

9. **No response status code in logging middleware (backend)**: The Logger middleware logs request method, path, and duration, but NOT the response status code. You can't build error rate dashboards from current logs.

10. **Missing request ID for correlation (backend)**: No request-scoped ID is generated. Errors can't be correlated across log lines or between backend and frontend.

11. **`slog.Error` is called inconsistently (backend)**: Some handlers log before `WriteError`, some don't. Some services log errors, others rely on the handler to log them.

12. **No error boundaries (frontend)**: No React Error Boundary wrapping the app. A render crash kills the app silently.

13. **No error visibility/monitoring (frontend)**: No Sentry/Crashlytics/Datadog integration. Errors shown in toasts disappear with no audit trail.

14. **catch (e) pattern with unknown error type (frontend)**: Several screens have `catch (error: any)` blocks with `error?.message` — no type safety, `Toast.show` with potentially undefined values.

15. **`Timeout` middleware returns non-standard format (backend)**: `http.TimeoutHandler(next, d, '{"error":"timeout"}')` returns a raw JSON string that doesn't match the `APIResponse` envelope format.

16. **Mixed `log` vs `slog` usage (backend)**: `log.Printf` and `log.Fatal` in main.go and proof/service.go, while handlers use `slog`. Cron job uses `log.Printf`.

---

### Approaches

1. **Minimal — Add error codes + frontend error mapping** — Add a `code` field to `APIError`, keep current error handling patterns, add frontend error code map.
   - Pros: Low effort, immediate improvement, no architecture changes
   - Cons: Doesn't fix error wrapping, leaky errors, missing middleware
   - Effort: Low

2. **Structured — Standard error middleware + domain error wrapping** — Create global panic recovery middleware, chi error handler middleware, standardize `%w` wrapping in all services, map GORM errors to domain errors in repositories, add `code` field to APIError.
   - Pros: Addresses the core issues, consistent error flow, protects against panics
   - Cons: Requires touching every repository and many service files
   - Effort: High

3. **Full observability — Approach 2 + frontend error boundary + telemetry** — All of approach 2 plus React error boundaries, Sentry/Crashlytics integration, request tracing via context, structured error logging with correlation IDs.
   - Pros: Production-grade error observability, full panic safety, debugging capability
   - Cons: Highest effort, requires third-party service setup
   - Effort: High

4. **Incremental — Fix leaks first, then standardize** — Phase 1: Fix all handlers that leak raw errors, add error codes. Phase 2: Add panic recovery middleware + error wrapping. Phase 3: Frontend error boundaries + telemetry.
   - Pros: Lower risk per phase, immediate wins, adaptable
   - Cons: Requires discipline to complete all phases
   - Effort: Medium (per phase) / High (total)

### Recommendation

**Approach 4 (Incremental)** — Do it in phases:

**Phase 1** (backend): 
- Add `code` field to `APIError`
- Fix every handler that passes `err.Error()` to the user
- Standardize error messages to Spanish
- Add panic recovery middleware
- Change `Timeout` middleware to use standard `APIResponse` format

**Phase 2** (backend + frontend):
- Wrap all repository errors with domain context (`fmt.Errorf("user not found: %w", err)`)
- Map GORM errors to domain sentinels in all repositories
- Add `code` field consumption in frontend `ApiResponse` type
- Create a centralized `handleApiError` utility on the frontend

**Phase 3** (frontend):
- Add React Error Boundary
- Integrate crash reporting (Sentry or similar)
- Add request ID header propagation and logging

### Risks

- **Risk 1**: Adding error codes requires updating both frontend and backend simultaneously — API contract change requires coordination
- **Risk 2**: The `fields` field type inconsistency (`interface{}`) means frontend error parsing can break if the backend sends unexpected field shapes
- **Risk 3**: GORM error handling depends on specific PostgreSQL error codes — this breaks if migrating to a different database
- **Risk 4**: Phase 2 requires touching every repository (prode, user, voucher, proof, token) — risk of introducing regressions under transaction logic

### Ready for Proposal

Yes — the analysis is clear enough to write a proposal. The main debate is whether to go all-in (Approach 2/3) or phased (Approach 4). Recommendation is phased.
