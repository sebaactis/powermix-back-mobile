# Tasks: unified-error-handling

## Phase 1 — Error codes + middlewares

- [x] 1.1 [RED] `internal/utils/response_test.go` — WriteError outputs `code` in JSON, all 7 constants present in envelope
- [x] 1.2 [GREEN] `internal/utils/response.go` — add `code` to APIError, 8 error constants, `WriteError(w, status, WriteErrorOpts)`, `WriteErrorMessage` bridge
- [x] 1.3 [RED] `internal/middlewares/mw_test.go` — Recoverer catches panic → 500 JSON; Timeout returns APIResponse format
- [x] 1.4 [GREEN] `internal/middlewares/mw.go` — add Recoverer(logger); fix Timeout to write APIResponse with ERR_TIMEOUT + Spanish
- [x] 1.5 [GREEN] `internal/domain/entities/user/handler.go` — replace err.Error() with Spanish messages + error codes
- [x] 1.6 [GREEN] `auth/handler.go` (login, OAuth, refresh, recovery) — token/handler.go is stub only; HTTP lives in auth
- [x] 1.7 [GREEN] `internal/domain/entities/proof/handler.go` — WriteError + codes, helpers, no err.Error() leak
- [x] 1.8 [GREEN] `internal/domain/entities/voucher/handler.go` — same pattern; fix 2× err.Error() leaks in delete/count
- [x] 1.9 [GREEN] `internal/domain/entities/prode/handler.go` — add error codes; fix 1× err.Error() leak
- [x] 1.8b [GREEN] `internal/middlewares/ratelimit_middleware.go` — migrate WriteErrorMessage → WriteError
- [x] 1.10 [GREEN] `internal/middlewares/auth_middleware.go` — add error codes to auth failure responses
- [x] 1.11 [GREEN] `internal/routes/router.go` — register Recoverer as first middleware

## Phase 2 — Error wrapping + frontend utility

- [x] 2.1 [RED] Repository test files (table-driven, per entity) — verify GORM→domain sentinel mapping
- [x] 2.2 [GREEN] `internal/domain/entities/user/repository.go` — %w wrap GORM errors, map to domain sentinels, log raw
- [x] 2.3 [GREEN] `internal/domain/entities/user/service.go` — %w propagate, add ErrInternal sentinel, business context
- [x] 2.4 [GREEN] `internal/domain/entities/proof/repository.go` — same wrapping pattern
- [x] 2.5 [GREEN] `internal/domain/entities/token/repository.go` — same
- [x] 2.6 [GREEN] `internal/domain/entities/voucher/repository.go` — same
- [x] 2.7 [GREEN] `internal/domain/entities/prode/repository.go` — unify wrapping pattern

## Phase 2b — Error contract cleanup (discovered during apply)

- [x] 2.7b.1 [GREEN] `user/repository.go` — fix register duplicate email returning ERR_INTERNAL (PostgreSQL tx abort bug)
- [x] 2.7b.2 [GREEN] `user/service.go` — fix Update nil panic on req.Name, add validation error
- [x] 2.7b.3 [GREEN] `user/handler.go` — UpdatePassword error classification (validation/not-found/internal)
- [x] 2.7b.4 [GREEN] `auth/handler.go` — UpdatePasswordByRecovery error classification (token invalid → 401)
- [x] 2.7b.6 [GREEN] `maintenance_key.go` — use utils.WriteError with API envelope instead of http.Error
- [x] 2.7b.7 [GREEN] `response.go` — add ErrCodeConflict, migrate prode/voucher conflict helpers
- [x] 2.7b.8 [GREEN] `proof/{errors,service,handler}.go` — refactor business errors from strings.Contains to sentinels + errors.Is
- [x] 2.7b.10 [RED/GREEN] Contract tests — maintenance_key (4), response conflict, isDuplicateKeyError (10), Update nil name

## Phase 2c — Frontend error utility

- [x] 2.8 [RED] `src/helpers/apiHelper.test.ts` — test handleApiError returns correct message/fallback per code
- [x] 2.9 [GREEN] `src/helpers/apiHelper.ts` — add `code` to ApiError type, create `handleApiError<T>(res, fallback)`

## Phase 3 — ErrorBoundary + Sentry + RequestID

- [x] 3.1 [RED] `src/components/ErrorBoundary.test.tsx` — fallback shows "Algo salió mal" + "Reintentar" + retry works
- [x] 3.2 [GREEN] `src/components/ErrorBoundary.tsx` — React Error Boundary with fallback UI + retry state
- [x] 3.3 [GREEN] `app/index.tsx` — wrap MainNavigator with ErrorBoundary inside AuthProvider
- [x] 3.4 [GREEN] `internal/middlewares/mw.go` — add RequestID() middleware (UUID in context)
- [x] 3.5 [GREEN] `internal/routes/router.go` — register RequestID middleware
- [ ] 3.6 [GREEN] `internal/middlewares/request_logger.go` — RequestLogger middleware (method, path, service, duration, status, body)
- [ ] 3.7 [GREEN] `internal/platform/logger/{logger,handler}.go` — JSON slog handler with ContextHandler for request_id injection
- [ ] 3.8 [GREEN] `cmd/api/main.go` — configure JSON slog with ContextHandler, replace log. calls with slog
- [ ] 3.9 [GREEN] All handler/service/repository files — migrate slog calls to context-aware versions

## Review Workload Forecast

| Metric | Value |
|--------|-------|
| Estimated additions + deletions | ~650–750 lines |
| Exceeds 400 lines | **YES** |
| Chained PRs recommended | **YES** — 3 chained PRs (one per phase), each independently verifiable |
| Decision needed before apply | **YES** — confirm chained-PR splitting strategy, verify per-phase `go test ./...` pass |

## Implementation Order

Phase 1 → Phase 2 → Phase 3. Each phase is independently deployable and must pass `go test ./...` before moving to the next. Phase 1 foundation (codes + middleware) is prerequisite for Phase 2 error mapping. Phase 3 is purely additive on frontend + RequestID on backend.
