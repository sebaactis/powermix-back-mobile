# Proposal: Go Naming Conventions Refactor

**Change ID**: C2  
**Project**: powermix-mobile-backend  
**Date**: 2026-03-17  
**Status**: Proposed  
**Author**: OpenCode SDD-Propose

---

## Intent

The codebase currently violates Go idiomatic naming conventions across multiple layers, reducing readability and maintainability for developers familiar with Go standards. This technical debt manifests as:

- **Struct fields** using underscore separators (e.g., `Locked_until`, `ID_MP`, `Date_Approved_MP`) instead of camelCase
- **Function names** using non-idiomatic patterns (e.g., `GetById` instead of `GetID`)
- **Acronym capitalization** inconsistent (e.g., `userId` vs `userID`, `HttpHandler` vs `HTTPHandler`)
- **Variable naming** using single-letter shortcuts where clarity is preferred

This change brings the codebase into alignment with [Effective Go](https://golang.org/doc/effective_go#names) naming conventions, improving:
- **Readability** for Go developers (onboarding friction reduced)
- **Consistency** across entity layers (models, DTOs, services, handlers)
- **Linting compliance** (golangci-lint naming rules)
- **Long-term maintainability** (reduces cognitive load when reading code)

---

## Scope

### In Scope

1. **Entity Model Renames (Phase A)**
   - `user.go`: `Locked_until` → `LockedUntil`
   - `proof.go`: `ID_MP`, `Date_Approved_MP`, `Status_MP`, `Amount_MP`, `Is_Revoked` → `IDMP`, `DateApprovedMP`, `StatusMP`, `AmountMP`, `IsRevoked`
   - `token.go`: Similar acronym patterns (if present)
   - `voucher.go`: Similar patterns (if present)
   - Associated JSON tags (standardized to snake_case for API compatibility)

2. **DTO Renames (Phase A)**
   - `user/dto.go`: Field renaming to match camelCase
   - `proof/dto.go`: Field renaming to match camelCase
   - `token/dto.go`: Field renaming (if DTOs exist)
   - `voucher/dto.go`: Field renaming (if DTOs exist)

3. **Service/Handler Layer Renames (Phase B)**
   - Repository methods: `GetById` → `GetByID`, `GetAllByUserId` → `GetAllByUserID`, etc. (11+ method renames)
   - Handler variable names: `u` → `user`, `s` → `service`, `r` → `repo` for clarity
   - Function parameter and local variable improvements where applicable

4. **GORM Tags Verification**
   - Verify and update GORM column tags to ensure database query compatibility
   - No schema changes to database; column names remain unchanged

### Out of Scope

- **External API contracts**: JSON response/request field names remain snake_case (via JSON tags)
- **Database schema changes**: Column names stay as-is; GORM tags ensure correct mapping
- **Configuration or infrastructure**: No changes to config.go, routes, or middleware
- **Test rewrites**: Existing tests updated only where necessary to pass; no new tests added unless required for verification
- **Documentation/comments**: Godoc comments deferred to separate change

### Estimated Impact

- **Files affected**: ~15 (across `user/`, `proof/`, `token/`, `voucher/` entity packages)
- **Lines changed**: 200–250 lines of modifications
- **Complexity**: Medium (primarily rename operations; low risk of logical changes)
- **Breaking changes**: None (internal code only; JSON API contracts unchanged)

---

## Approach

### Two-Phase Execution Strategy

**Phase A: Entity Models & DTOs**
- Rename struct fields in models and DTOs
- Update GORM column tags (if needed for edge cases)
- Ensure JSON tags remain snake_case for API contracts
- Commit to git after each entity (User → Proof → Voucher → Token)
- Run tests and verify build after each commit

**Phase B: Service & Handler Layers**
- Rename repository methods (GetById, GetAllByUserId, etc.)
- Rename handler variable names for clarity (u → user, s → service)
- Update all call sites using IDE refactoring tools
- Single commit for phase B (all service/handler changes)
- Run full test suite and verify build

### Entity-by-Entity Order

To minimize risk and enable incremental verification:
1. **User** — foundational entity; used by many others
2. **Proof** — depends on User; has most violations (ID_MP, Date_Approved_MP, etc.)
3. **Voucher** — depends on User; fewer violations expected
4. **Token** — depends on User; potentially fewer violations

### Refactoring Tools & Process

- **IDE-guided refactoring** (GoLand, VS Code with Go extensions) for safe rename operations
- **Build verification** after each commit: `go build ./...`
- **Test verification** after each commit: `go test ./...`
- **Manual inspection** of GORM tags and JSON marshaling to ensure correctness

---

## Affected Areas

| Area | Phase | Impact | Details |
|------|-------|--------|---------|
| `internal/domain/entities/user/` | A | Modified | `Locked_until` → `LockedUntil` in user.go; DTO fields updated |
| `internal/domain/entities/proof/` | A, B | Modified | Entity fields (ID_MP, Date_Approved_MP, etc.) renamed; service methods (GetById, etc.) updated |
| `internal/domain/entities/voucher/` | A, B | Modified | DTO and service method renames |
| `internal/domain/entities/token/` | A, B | Modified | Field and method renames (if applicable) |
| `internal/domain/entities/*/handler.go` | B | Modified | Variable names (u → user, s → service) |
| `internal/domain/entities/*/repository.go` | B | Modified | Method names (GetById, GetAllByUserId, etc.) |
| `internal/domain/entities/*/service.go` | B | Modified | Method signatures and call sites updated |
| Tests | A, B | Updated | Test code updated to use new names; no new tests added |

---

## Risks & Mitigations

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| **Database query mismatch** — GORM struct tags not updated, causing queries to fail | Medium | Verify GORM column tags for each renamed field; test queries with real database; schema inspection before/after |
| **Cross-entity dependency breakage** — Renamed methods not updated in all call sites | Medium | Use IDE-guided refactoring (safe rename) to ensure all call sites are updated; manual inspection of diffs |
| **Test failures** — Tests using old field/method names fail | Low | Update all test references; run `go test ./...` after each phase |
| **JSON marshaling errors** — JSON tags not properly aligned with struct fields | Low | Verify JSON tags are snake_case for API contracts; test JSON serialization/deserialization |
| **Large diff hard to review** — Phase B creates a large changeset | Medium | Split into two separate PRs (Phase A, Phase B) for easier review; include detailed commit messages |
| **Acronym inconsistency** — Some acronyms renamed inconsistently (e.g., MP vs Mp) | Low | Use consistent rule: all-caps for known acronyms (ID → ID, MP → MP, HTTP → HTTP) |

---

## Rollback Plan

**Per-Phase Rollback**:
- After Phase A: `git revert <phase-a-commit>` — safely reverts entity model renames
- After Phase B: `git revert <phase-b-commit>` — safely reverts service/handler renames

**Full Rollback** (if needed):
```bash
# Revert both phases
git revert <phase-b-commit>
git revert <phase-a-commit>
```

**Rollback Feasibility**: Very low risk — no data migrations, no config changes, no external API changes. The change can be reverted at any point without side effects.

---

## Dependencies

- **Go 1.25+** (already in use; no version bump needed)
- **IDE with Go refactoring** (GoLand, VS Code Go extension, or vim-go for safe renames)
- **Existing test suite** (must pass before and after refactoring)
- **GORM knowledge** (to verify column tags match database schema)

---

## Success Criteria

- [ ] All 50+ identified naming violations are fixed
- [ ] Phase A commit: Entity models and DTOs renamed; tests pass
- [ ] Phase B commit: Service/handler layer and methods renamed; tests pass
- [ ] Build succeeds: `go build ./...` with zero errors
- [ ] Full test suite passes: `go test ./...`
- [ ] Code follows [Effective Go naming conventions](https://golang.org/doc/effective_go#names)
- [ ] JSON API contracts unchanged (no breaking changes to external clients)
- [ ] No new linting violations introduced; existing violations resolved
- [ ] PR review and approval from team
- [ ] No performance degradation (benchmarks unchanged, if applicable)

---

## Next Steps

### Immediate (sdd-spec)
1. **Write formal specifications** for Phase A and Phase B
   - Detailed list of each rename (struct fields, methods, variables)
   - JSON tag mapping
   - GORM tag verification checklist
   - Test mapping

2. **Create technical design** (optional, if needed for complex aspects)
   - GORM column tag strategy
   - JSON marshaling verification process
   - Variable naming conventions (when to use single-letter vs full names)

3. **Break down tasks** for implementation
   - Per-entity task (User, Proof, Voucher, Token)
   - Per-layer task (Phase A models, Phase A DTOs, Phase B service, Phase B handler)
   - Verification task (test + build + linting)

### Implementation (sdd-apply)
- Follow task checklist
- Use IDE-guided refactoring for safe renames
- Verify build and tests after each commit

### Verification (sdd-verify)
- All criteria from "Success Criteria" section met
- Full diff review for correctness
- Integration testing with real database

### Archive (sdd-archive)
- Sync delta specs to main specs
- Move change to archive with completion date

---

## Acceptance Criteria Checklist

- [ ] **Spec Phase**: Detailed specifications written and reviewed
- [ ] **Design Phase**: Technical approach documented (GORM tags, JSON marshaling)
- [ ] **Implementation Phase**: All 50+ violations fixed per specifications
- [ ] **Build Phase**: `go build ./...` passes
- [ ] **Test Phase**: `go test ./...` passes (all tests pass)
- [ ] **Code Quality**: No new linting violations; Go conventions followed
- [ ] **API Compatibility**: JSON contracts unchanged; no breaking changes
- [ ] **Code Review**: PR approved by team
- [ ] **Integration**: Verified with real database and external clients (if applicable)
- [ ] **Archive**: Change moved to archive with completion report

---

**Ready for**: [sdd-spec](../../../skills/sdd-spec/SKILL.md) phase to write detailed specifications and task breakdown.
