# Apply Progress: unified-error-handling

**Last updated:** 2026-05-29  
**Phase:** 3 (slog/requestID) — **100% complete**  
**Status:** Complete — all backend tasks done; ready for verify/PR

## Session config

| Setting | Value |
|---------|--------|
| Mode | Interactive (show plan before write) |
| Artifacts | Hybrid (`openspec/` + Engram topic keys) |
| Delivery | ask-on-risk, PRs chicos |
| Backend branch | Feature branch from `develop` |
| Frontend branch | `feature/prode` at `D:\Descargas\Seba\Proyects\React Native\powermix` |

## Repos

- **Backend:** `powermix-mobile-backend` (Go 1.25, chi, GORM)
- **Frontend:** `powermix` (React Native) — not started in apply

## SDD artifacts (planning complete)

All under `openspec/changes/unified-error-handling/`:

- `exploration.md`, `proposal.md`, `design.md`, `tasks.md`, `specs/` (6 domains)
- Engram keys: `sdd/unified-error-handling/{explore,proposal,spec,design,tasks,apply-progress}`

## Phase 1 — completed tasks

| Task | File | Summary |
|------|------|---------|
| 1.1 | `internal/utils/response_test.go` | Tests PASS |
| 1.2 | `internal/utils/response.go` | `code` on `APIError`, 8 `ERR_*` constants, `WriteErrorOpts`, `WriteErrorMessage` bridge |
| 1.3 | `internal/middlewares/mw_test.go` | Recoverer + Timeout tests PASS |
| 1.4 | `internal/middlewares/mw.go` | `Recoverer()`, standard timeout envelope |
| 1.5 | `internal/domain/entities/user/handler.go` | Explicit codes, no leaks, safe UUID parse |
| 1.6 | `internal/security/auth/handler.go` | Login, OAuth, refresh, recovery migrated |
| 1.7 | `internal/domain/entities/proof/handler.go` | Helpers + business vs internal errors |
| 1.8 | `internal/domain/entities/voucher/handler.go` | Helpers, explicit codes, 401 on missing user, no err leaks |
| 1.9 | `internal/domain/entities/prode/handler.go` | Helpers, explicit validation messages (no err.Error()) |
| 1.8b | `internal/middlewares/ratelimit_middleware.go` | `WriteError` with `ERR_INTERNAL` on 429 |
| 1.10 | `internal/middlewares/auth_middleware.go` | `ERR_UNAUTHORIZED` on auth failures |
| 1.11 | `internal/routes/router.go` | `Recoverer` first in middleware stack |

**Note 1.6:** `token/handler.go` is only a stub (struct + constructor). Token HTTP lives in `auth/handler.go`.

**Verify last run:** `go build ./...` OK; utils, middlewares, proof, auth tests OK. Pre-existing failure in `internal/clients/coffeeji` (unrelated).

## Phase 1 — changes in this session (1.8–1.8b)

### voucher/handler.go
- Added `writeVoucher*` helpers (Unauthorized, Validation, NotFound, Forbidden, Conflict, Internal)
- Missing user context: 400 → **401 ERR_UNAUTHORIZED**
- `GetAllByUserID`: removed `err` from response fields; added slog + ERR_INTERNAL
- `DeleteVoucher` / `GetAvailableCount`: removed `err.Error()` leaks

### prode/handler.go
- Added `writeProde*` helpers
- `CreateOrUpdatePrediction`: explicit Spanish for `ErrInvalidScore` / `ErrScoreOutOfRange` instead of `err.Error()`
- Removed redundant slog on unauthenticated prediction (401 via helper)

### ratelimit_middleware.go
- Single `WriteError` with explicit `ERR_INTERNAL` code

## Phase 2 — next tasks

| Task | Scope | Status |
|------|-------|--------|
| 2.1 | [RED] Repository mapping tests + sentinels + pass-through stubs | ✅ RED verified |
| 2.2 | [GREEN] `user/repository.go` — implement `mapRepoErr` + wire all methods | ✅ GREEN |
| 2.3 | [GREEN] `user/service.go` — `wrapServiceErr` + wire all methods | ✅ GREEN |
| 2.4 | [GREEN] `proof/repository.go` — `mapProofRepoErr` + wire all methods | ✅ GREEN |
| 2.5 | [GREEN] `token/repository.go` — `mapTokenRepoErr` + `mapResetTokenRepoErr` | ✅ GREEN |
| 2.6 | [GREEN] `voucher/repository.go` — `mapVoucherAssignErr` + `mapVoucherRepoErr` | ✅ GREEN |
| 2.7 | [GREEN] `prode/repository.go` — match/prediction/repo helpers + `errors.Is` | ✅ GREEN |
| 2.8–2.9 | Frontend `apiHelper.ts` | Next |

### 2.1 RED evidence (2026-05-29)

New files:
- `user/errors.go`, `token/errors.go`, `proof/errors.go`
- `*/repository_test.go` ×5 (table-driven `map*RepoErr` tests)
- `ErrInternal` added to `voucher/repository.go`, `prode/errors.go`
- Pass-through stubs in each `repository.go` (return raw `err`)

```text
go test ./internal/domain/entities/{user,token,proof,voucher,prode}/...
→ 5 FAIL (mapping tests red); nil-input subtests pass
```

### 2.2 GREEN evidence (2026-05-29)

- `mapRepoErr` implemented with `%w` + `slog.Error` on unexpected DB errors
- All repository methods wired; `RowsAffected == 0` → `ErrNotFound`
- `ErrDuplicateEmail` moved to `errors.go`
- `handler.go`: `gorm.ErrRecordNotFound` → `ErrNotFound` (GetByID, Me)

```text
go test ./internal/domain/entities/user/... → PASS (4/4 subtests)
go build ./... → OK
```

### 2.3 GREEN evidence (2026-05-29)

- `wrapServiceErr` adds service-layer context with `%w` (sentinels still detectable via `errors.Is`)
- All service methods that call repository/mailer wrap errors on return
- `service_test.go`: sentinel preservation through repo → service chain

```text
go test ./internal/domain/entities/user/... → PASS (6 subtests)
```

### 2.7 GREEN evidence (2026-05-29)

- `mapProdeMatchErr`, `mapProdePredictionErr`, `mapProdeRepoErr`, `mapProdeUserPredictionLookupErr`
- Replaced `err == gorm.ErrRecordNotFound` with `errors.Is` throughout
- `GetUserPrediction` / `GetRewardByPredictionID`: not found → `nil` (business rule)
- All CRUD methods wired

```text
go test ./internal/domain/entities/{user,proof,token,voucher,prode}/... → all PASS
```

### 2.6 GREEN evidence (2026-05-29)

- `mapVoucherAssignErr`: not found → `ErrNoAvailableVouchers`; DB errors → log + `ErrInternal`
- `mapVoucherRepoErr`: generic DB errors → log + `ErrInternal`
- All methods wired; `DeleteUsedVoucher` business sentinels wrapped with `%w`

```text
go test ./internal/domain/entities/voucher/... → PASS (5 subtests)
```

### 2.5 GREEN evidence (2026-05-29)

- `mapTokenRepoErr`: not found → `ErrTokenNotFound`; DB errors → log + `ErrInternal`
- `mapResetTokenRepoErr`: not found → `ErrTokenInvalid` (reset password flow)
- All repository methods wired; `Update` with `RowsAffected == 0` → `ErrTokenNotFound`

```text
go test ./internal/domain/entities/token/... → PASS (6 subtests)
```

### 2.4 GREEN evidence (2026-05-29)

- `mapProofRepoErr` logs unexpected DB errors and wraps with `ErrInternal`
- All list/create methods wired; `GetByID` keeps `nil, nil` for not found (business rule)
- `handler.go`: removed dead `gorm.ErrRecordNotFound` check (not found handled via `proof == nil`)

```text
go test ./internal/domain/entities/proof/... → PASS (mapping + service tests)
```

### 2.7b Phase — Error contract cleanup evidence (2026-05-29)

| Task | Fix | Tests |
|------|-----|-------|
| 2.7b.1 | Register duplicate → `ErrDuplicateEmail` directo, sin tx abortada | `TestIsDuplicateKeyError` (10 casos) |
| 2.7b.2 | Update nil panic → validation error | `TestService_Update_nilName` |
| 2.7b.3 | ChangePassword classification (val/404/500) | — |
| 2.7b.4 | Recovery token invalid → 401 (privacy) | — |
| 2.7b.6 | Maintenance key envelope | 4 tests middleware |
| 2.7b.7 | `ErrCodeConflict` nuevo | `TestWriteError_ConflictCode` |
| 2.7b.8 | Proof sentinels vs `strings.Contains` | existentes + `errors.Is` |
| 2.7b.10 | Contract tests | 6 tests nuevos |

**Build:** OK. **Tests:** todos PASS except `coffeeji` preexistente.

## Phase 3 — completed tasks

| Task | File | Summary |
|------|------|---------|
| 3.1 | `src/components/ErrorBoundary.test.tsx` | Tests PASS |
| 3.2 | `src/components/ErrorBoundary.tsx` | React Error Boundary with fallback UI + retry |
| 3.3 | `app/index.tsx` | Wrapped MainNavigator with ErrorBoundary |
| 3.4 | `internal/middlewares/request_id.go` | UUID generation + context injection + X-Request-ID header |
| 3.5 | `internal/routes/router.go` | Register RequestID + RequestLogger + Recoverer middlewares |
| 3.6 | `internal/middlewares/request_logger.go` | Structured JSON request logging with body sanitization |
| 3.7 | `internal/platform/logger/{logger,handler}.go` | ContextHandler injects request_id/service automatically; Logger wrapper |
| 3.8 | `cmd/api/main.go` | JSON slog with ContextHandler configured as default |
| 3.9 | All domain handlers/services/repositories | Migrated to `slog.*Context` for automatic context field injection |

### Phase 3 fixes (this session)

- **RequestLogger**: changed `log.Info(...)` → `log.InfoContext(ctx, ...)` so ContextHandler injects request_id/service
- **Recoverer**: changed `logger.Error(...)` → `logger.ErrorContext(r.Context(), ...)` so panics carry request context
- **Cleaned obsolete `Logger()` middleware** from `mw.go` and removed unused `Deps.Logger` from router
- **Fixed repository_test.go build errors** ×5: added missing `context.Background()` to `map*RepoErr` calls in prode/proof/token/voucher tests
- **New tests:**
  - `internal/platform/logger/handler_test.go` — ContextHandler injects request_id + service (2 tests)
  - `internal/middlewares/request_logger_test.go` — request logging, sanitization, context propagation, IP extraction, status capture, body skip (7 tests)

### Phase 3 verify evidence

```text
go build ./... → OK
go test ./internal/domain/entities/... → PASS (user, proof, token, voucher, prode)
go test ./internal/middlewares/... → PASS (mw, auth, maintenance_key, request_logger)
go test ./internal/platform/logger/... → PASS (handler)
```

**Pre-existing failure:** `internal/clients/coffeeji` (unrelated to this change).

## How to resume

1. Say: *"Retomemos unified-error-handling Fase 2"*
2. Or run `/sdd-continue unified-error-handling`
3. Read this file + `tasks.md` for checklist state

### Post-apply cleanup (2026-05-29)

- **RateLimiter connected**: added `d.RateLimiter.Middleware()` to router middleware stack
- **All comments translated to Spanish**: ~40 comments across 15+ files
- **Logger context fix**: `Logger.Info/Warn/Error` now pass `ctx` to `inner.InfoContext/etc`
- **Dead code**: none found beyond the unused RateLimiter (now connected)

**Gate before PR:** verify chained-PR strategy. Recommended: 2 PRs — Phase 1+2 (error codes + wrapping) and Phase 3 (logging + ErrorBoundary).
