# Proposal: unified-error-handling

## Intent

Backend (Go + chi + GORM) and frontend (React Native) have 16 error-handling gaps: no error codes, no panic recovery, raw errors leak to users, no `%w` wrapping, mixed Spanish/English, no frontend Error Boundaries, duplicated error logic per screen. Fix incrementally in 3 phases.

## Scope

### In Scope
- **Phase 1 (backend)**: Add `code` field to `APIError`, fix handlers that leak `err.Error()`, add panic recovery middleware, fix Timeout middleware format, standardize error messages to Spanish
- **Phase 2 (backend + frontend)**: `%w` wrapping in all services/repos, map GORM→domain errors in repositories, add frontend `handleApiError` utility
- **Phase 3 (frontend)**: React Error Boundary, crash reporting (Sentry), request ID correlation

### Out of Scope
- Error sentinel refactor to new package structure
- Frontend i18n translation framework
- Changing response envelope shape (only adding `code` field)
- Database schema changes

## Capabilities

### New Capabilities
- `error-codes`: Error code taxonomy + `code` field in `APIError`
- `panic-recovery`: Global `recover()` middleware for chi
- `error-middleware`: Global catch-all error handler middleware
- `domain-error-wrapping`: `%w` wrapping across repository→service→handler
- `error-boundary`: React Error Boundary component
- `error-monitoring`: Crash reporting (Sentry) integration

### Modified Capabilities
None — no existing specs in `openspec/specs/`.

## Approach

3-phase incremental per exploration recommendation:

**Phase 1** — `internal/utils/response.go`: add `code string` to `APIError`. `internal/middlewares/mw.go`: add `Recoverer` middleware, fix `Timeout` to use `APIResponse`. All `*handler.go` files: replace `err.Error()` with descriptive Spanish messages, add `code` to `WriteError` calls.

**Phase 2** — All `repository.go` files: wrap GORM errors with `fmt.Errorf("domain context: %w", err)`, map to domain sentinels. All `service.go` files: `%w` wrapping through layers. Frontend `src/helpers/apiHelper.ts`: consume `code` field, create `handleApiError(res, fallback)` centralized utility.

**Phase 3** — `src/components/ErrorBoundary.tsx` (new): React Error Boundary. `src/App.tsx`: wrap navigation tree. Integrate Sentry via `@sentry/react-native`. Add request ID propagation via `context.Context` and `slog` attributes.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/utils/response.go` | Modified | Add `code` field to `APIError` |
| `internal/middlewares/mw.go` | Modified | Add `Recoverer`, fix `Timeout` format |
| `internal/domain/entities/*/handler.go` | Modified | Fix leaks, add codes |
| `internal/domain/entities/*/repository.go` | Modified | Map GORM→domain errors (Phase 2) |
| `internal/domain/entities/*/service.go` | Modified | `%w` wrapping (Phase 2) |
| `src/helpers/apiHelper.ts` | Modified | Consume `code`, add `handleApiError` |
| `src/components/ErrorBoundary.tsx` | New | React Error Boundary (Phase 3) |
| `src/App.tsx` | Modified | Wrap with ErrorBoundary (Phase 3) |
| `cmd/api/main.go` | Modified | Request ID middleware (Phase 3) |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|-------------|
| Adding `code` field breaks frontend parsing | Low | Frontend consumes field with optional chaining — non-breaking |
| GORM error code inspection fragile across PG versions | Med | Use `errors.Is(gorm.ErrDuplicatedKey)` not raw PG codes |
| Phase 2 touches 5+ repositories — regression risk | Med | Per-repository tests + `go test ./...` after each |
| Sentry SDK adds bundle size and startup cost | Low | Lazy init, feature-flag guarded |

## Rollback Plan

- **Phase 1**: Revert `code` field addition → frontend unaffected. Remove `Recoverer` middleware. Restore old handler error strings.
- **Phase 2**: Revert repository/service error wrapping changes per entity. Restore old frontend error handling.
- **Phase 3**: Remove Error Boundary from `App.tsx`, delete `ErrorBoundary.tsx`, remove Sentry SDK. Revert request ID changes.

Each phase is independently rollbackable. Phase N+1 does NOT depend on Phase N being stable — but each phase builds on the previous one.

## Dependencies

- Phase 3: `@sentry/react-native` package install

## Success Criteria

- [ ] Every `WriteError` call includes a `code` field
- [ ] No raw `err.Error()` or `map[string]string{"error": err.Error()}` in any handler
- [ ] Panic recovery returns `APIResponse` format, not raw text
- [ ] Timeout middleware returns `APIResponse` with error code
- [ ] All user-facing error messages in Spanish
- [ ] GORM errors mapped to domain sentinels in every repository
- [ ] Frontend has a single `handleApiError(res, fallback)` utility used by all screens
- [ ] React Error Boundary wraps navigation tree
- [ ] `go test ./...` passes after each phase
