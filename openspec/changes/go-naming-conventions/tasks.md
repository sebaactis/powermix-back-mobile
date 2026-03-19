# Tasks: Go Naming Conventions Refactor

## Overview

This task breakdown implements the **two-phase refactoring strategy** defined in the design document. The change targets 50+ naming violations across entity models, DTOs, and service layers.

**Total Tasks**: 58 tasks organized into 5 commits + final verification.

---

## Commit 1: User Entity Phase (A.1)

### Phase Overview
Refactor the User entity to idiomatic Go naming conventions. User is the foundational entity; fixing it enables dependent entities to follow suit. Tasks focus on renaming `Locked_until` → `LockedUntil` in the model and DTO, updating tags, and verifying the change with build and tests.

### Tasks

**Model Renames**
- [ ] 1.1: Read `internal/domain/entities/user/user.go` — identify all underscore-separated fields
- [ ] 1.2: Rename struct field `Locked_until` → `LockedUntil` in `internal/domain/entities/user/user.go`
- [ ] 1.3: Verify JSON tag on `LockedUntil` field is `json:"locked_until"` (snake_case)
- [ ] 1.4: Verify GORM tag on `LockedUntil` field is `gorm:"column:locked_until"` (unchanged)
- [ ] 1.5: Confirm no other underscore-separated fields exist in User struct (legacy cleanup check)

**DTO Updates**
- [ ] 1.6: Read `internal/domain/entities/user/dto.go` — check if DTO fields mirror model
- [ ] 1.7: Update DTO struct field(s) if DTO contains mirrored `Locked_until` field
- [ ] 1.8: Verify JSON tags on DTO fields are snake_case

**Test Updates**
- [ ] 1.9: Search `internal/domain/entities/user/` for test references to old field name `Locked_until` (case-sensitive)
- [ ] 1.10: Update any test assertions or field assignments using `Locked_until`
- [ ] 1.11: Update any test JSON unmarshaling that references the old field name

**Verification**
- [ ] 1.12: Build verification: `go build ./...` (zero errors)
- [ ] 1.13: Test verification: `go test ./internal/domain/entities/user/... -v` (all tests pass)
- [ ] 1.14: Grep verification: `grep -r "Locked_until" internal/domain/entities/user/` (should return zero results in non-tag code)

### Verification Steps
```bash
go build ./...
go test ./internal/domain/entities/user/... -v
grep -r "Locked_until" internal/domain/entities/user/ | grep -v "json:\|gorm:" | wc -l
# Expected: 0 (after Commit 1)
```

---

## Commit 2: Proof Entity Phase (A.2)

### Phase Overview
Refactor the Proof entity, which contains the most naming violations. The Proof entity has 6 MercadoPago-related fields using underscore and mixed-case naming. This commit renames all of them, updates tags, propagates the change to DTOs, updates test files, and verifies the changes across repository, service, and handler layers (if they reference struct fields directly).

### Tasks

**Model Renames**
- [x] 2.1: Read `internal/domain/entities/proof/proof.go` — identify all underscore-separated fields
- [x] 2.2: Rename struct field `ID_MP` → `IDMP` in `internal/domain/entities/proof/proof.go`
- [x] 2.3: Rename struct field `Date_Approved_MP` → `DateApprovedMP`
- [x] 2.4: Rename struct field `Operation_Type_MP` → `OperationTypeMP`
- [x] 2.5: Rename struct field `Status_MP` → `StatusMP`
- [x] 2.6: Rename struct field `Amount_MP` → `AmountMP`
- [x] 2.7: Rename struct field `Is_Revoked` → `IsRevoked` (bonus: CardId → CardID if present)

**Tag Updates**
- [x] 2.8: Verify JSON tag on `IDMP` is `json:"id_mp"` (snake_case)
- [x] 2.9: Verify JSON tag on `DateApprovedMP` is `json:"date_approved_mp"`
- [x] 2.10: Verify JSON tag on `OperationTypeMP` is `json:"operation_type_mp"`
- [x] 2.11: Verify JSON tag on `StatusMP` is `json:"status_mp"`
- [x] 2.12: Verify JSON tag on `AmountMP` is `json:"amount_mp"`
- [x] 2.13: Verify JSON tag on `IsRevoked` is `json:"is_revoked"`
- [x] 2.14: Verify GORM column tags for all renamed fields (should remain unchanged, e.g., `gorm:"column:id_mp"`)

**DTO Updates**
- [x] 2.15: Update `internal/domain/entities/proof/dto.go` — rename DTO fields to match model
- [x] 2.16: Update JSON tags in DTO to snake_case

**Service/Handler/Repository Field References**
- [x] 2.17: Search `internal/domain/entities/proof/service.go` for direct struct field accesses (e.g., `p.ID_MP`)
- [x] 2.18: Update all struct field references in proof service (e.g., `p.ID_MP` → `p.IDMP`)
- [x] 2.19: Search `internal/domain/entities/proof/handler.go` for direct struct field accesses
- [x] 2.20: Update all struct field references in proof handler
- [x] 2.21: Search `internal/domain/entities/proof/repository.go` for direct struct field accesses or SQL query references
- [x] 2.22: Update struct field references in proof repository (GORM column tags remain unchanged)

**Test Updates**
- [x] 2.23: Search test files in `internal/domain/entities/proof/` for references to old field names
- [x] 2.24: Update test assertions and field assignments for all renamed fields
- [x] 2.25: Update test JSON unmarshaling/marshaling that references old field names
- [x] 2.26: Update test setup/fixtures that use old field names

**Verification**
- [x] 2.27: Build verification: `go build ./...` (zero errors)
- [x] 2.28: Test verification: `go test ./internal/domain/entities/proof/... -v` (all tests pass)
- [x] 2.29: Grep verification: `grep -r "ID_MP\|Date_Approved_MP\|Operation_Type_MP\|Status_MP\|Amount_MP\|Is_Revoked" internal/domain/entities/proof/` (should return zero in non-tag code)

### Verification Steps
```bash
go build ./...
go test ./internal/domain/entities/proof/... -v
grep -r "ID_MP\|Date_Approved_MP\|Operation_Type_MP\|Status_MP\|Amount_MP\|Is_Revoked" internal/domain/entities/proof/ | grep -v "json:\|gorm:\|column:" | wc -l
# Expected: 0 (after Commit 2)
```

---

## Commit 3: Voucher Entity Phase (A.3)

### Phase Overview
Refactor the Voucher entity, following the same pattern as User and Proof. Check for underscore-separated fields, rename them to PascalCase, update tags, DTOs, and tests.

### Tasks

**Model Inspection & Renames**
- [ ] 3.1: Read `internal/domain/entities/voucher/voucher.go` — identify all underscore-separated fields
- [ ] 3.2: If underscore fields exist, rename to PascalCase (update each field)
- [ ] 3.3: Verify all JSON tags are snake_case
- [ ] 3.4: Verify all GORM column tags are correct (unchanged from database schema)

**DTO Updates**
- [ ] 3.5: Update `internal/domain/entities/voucher/dto.go` — rename fields to match model
- [ ] 3.6: Update JSON tags in DTO

**Cross-Layer Field References**
- [ ] 3.7: Search `internal/domain/entities/voucher/service.go` for direct struct field accesses
- [ ] 3.8: Update field references in voucher service
- [ ] 3.9: Search `internal/domain/entities/voucher/handler.go` for direct struct field accesses
- [ ] 3.10: Update field references in voucher handler
- [ ] 3.11: Search `internal/domain/entities/voucher/repository.go` for field references
- [ ] 3.12: Update field references in voucher repository

**Test Updates**
- [ ] 3.13: Update test files in `internal/domain/entities/voucher/` for old field names
- [ ] 3.14: Update test assertions, fixtures, and JSON tests

**Verification**
- [ ] 3.15: Build verification: `go build ./...` (zero errors)
- [ ] 3.16: Test verification: `go test ./internal/domain/entities/voucher/... -v` (all tests pass)
- [ ] 3.17: Grep verification: confirm no underscore field patterns remain in voucher code

### Verification Steps
```bash
go build ./...
go test ./internal/domain/entities/voucher/... -v
grep -r "_[a-z]" internal/domain/entities/voucher/ | grep "\.go:" | grep -v "json:\|gorm:" | wc -l
# Expected: 0 (minimal underscore patterns in non-tag code)
```

---

## Commit 4: Token Entity Phase (A.4)

### Phase Overview
Refactor the Token entity, following the same pattern as previous entities.

### Tasks

**Model Inspection & Renames**
- [x] 4.1: Read `internal/domain/entities/token/token.go` — identify all underscore-separated fields
- [x] 4.2: If underscore fields exist, rename to PascalCase
- [x] 4.3: Verify JSON tags use snake_case
- [x] 4.4: Verify GORM column tags are correct

**DTO Updates**
- [x] 4.5: Update `internal/domain/entities/token/dto.go` — rename fields to match model
- [x] 4.6: Update JSON tags in DTO

**Cross-Layer Field References**
- [x] 4.7: Update struct field accesses in `internal/domain/entities/token/service.go`
- [x] 4.8: Update struct field accesses in `internal/domain/entities/token/handler.go`
- [x] 4.9: Update struct field accesses in `internal/domain/entities/token/repository.go`

**Test Updates**
- [x] 4.10: Update test files in `internal/domain/entities/token/` for old field names

**Verification**
- [x] 4.11: Build verification: `go build ./...` (zero errors)
- [x] 4.12: Test verification: `go test ./internal/domain/entities/token/... -v` (all tests pass)
- [x] 4.13: Grep verification: confirm no underscore field patterns remain

### Verification Steps
```bash
go build ./...
go test ./internal/domain/entities/token/... -v
grep -r "_[a-z]" internal/domain/entities/token/ | grep "\.go:" | grep -v "json:\|gorm:" | wc -l
# Expected: 0 (minimal underscore patterns in non-tag code)
```

---

## Commit 5: Service/Handler Layer Phase (B)

### Phase Overview
Rename repository methods, service signatures, and handler methods across all entities. This is a **single commit** for all Phase B changes, ensuring all dependencies are updated together. Tasks are organized by entity and layer.

### Repository Method Renames (All Entities)

**User Repository**
- [x] 5.1: Rename method `GetById` → `GetByID` in `internal/domain/entities/user/repository.go`
- [x] 5.2: Rename method `GetAllByUserId` → `GetAllByUserID` (if exists)
- [x] 5.3: Update parameter name `userId` → `userID` in user repository methods

**Proof Repository**
- [x] 5.4: Rename method `GetById` → `GetByID` in `internal/domain/entities/proof/repository.go`
- [x] 5.5: Rename method `GetAllByUserId` → `GetAllByUserID`
- [x] 5.6: Rename method `GetAllByUserIdPaginated` → `GetAllByUserIDPaginated` (if exists)
- [x] 5.7: Rename method `GetLastThreeByUserId` → `GetLastThreeByUserID` (if exists)
- [x] 5.8: Update parameter name `userId` → `userID` in proof repository methods

**Voucher Repository**
- [x] 5.9: Rename methods matching user/proof pattern in `internal/domain/entities/voucher/repository.go`
- [x] 5.10: Update parameter names `userId` → `userID`

**Token Repository**
- [x] 5.11: Rename methods matching user/proof pattern in `internal/domain/entities/token/repository.go`
- [x] 5.12: Update parameter names `userId` → `userID`

### Service Layer Updates (All Entities)

**User Service**
- [x] 5.13: Update all repository method calls in `internal/domain/entities/user/service.go` to use new method names
- [x] 5.14: Rename function parameters from `userId` → `userID` in user service (if applicable)
- [x] 5.15: Update local variable names using `userID` pattern

**Proof Service**
- [x] 5.16: Update all repository method calls in `internal/domain/entities/proof/service.go` to use new names
- [x] 5.17: Rename function parameters from `userId` → `userID`
- [x] 5.18: Update local variable names using `userID` pattern

**Voucher Service**
- [x] 5.19: Update all repository method calls in `internal/domain/entities/voucher/service.go`
- [x] 5.20: Rename parameters/variables to use `userID` pattern

**Token Service**
- [x] 5.21: Update all repository method calls in `internal/domain/entities/token/service.go`
- [x] 5.22: Rename parameters/variables to use `userID` pattern

### Handler Layer Updates (All Entities)

**User Handler**
- [x] 5.23: Update all service method calls in `internal/domain/entities/user/handler.go` to use renamed methods
- [x] 5.24: Rename handler context variables (e.g., `u` → `user` where unclear; `s` → `service`)
- [x] 5.25: Update query parameter names `userId` → `userID` in HTTP handlers
- [x] 5.26: Update handler method names if they follow old pattern (e.g., `GetById` → `GetByID`)

**Proof Handler**
- [x] 5.27: Update all service method calls in `internal/domain/entities/proof/handler.go`
- [x] 5.28: Rename handler context variables for clarity
- [x] 5.29: Update query parameter names and HTTP method signatures
- [x] 5.30: Rename handler methods to use `GetByID`, `GetAllByUserID` patterns

**Voucher Handler**
- [x] 5.31: Update all service method calls in `internal/domain/entities/voucher/handler.go`
- [x] 5.32: Rename handler context variables and method names

**Token Handler**
- [x] 5.33: Update all service method calls in `internal/domain/entities/token/handler.go`
- [x] 5.34: Rename handler context variables and method names

### Test File Updates (All Entities)

**User Tests**
- [x] 5.35: Update `internal/domain/entities/user/*_test.go` — all repository method calls
- [x] 5.36: Update test assertions for parameter names (`userId` → `userID`)

**Proof Tests**
- [x] 5.37: Update `internal/domain/entities/proof/*_test.go` — all repository/service method calls
- [x] 5.38: Update test assertions for parameter names

**Voucher Tests**
- [x] 5.39: Update `internal/domain/entities/voucher/*_test.go` — all method calls and parameters

**Token Tests**
- [x] 5.40: Update `internal/domain/entities/token/*_test.go` — all method calls and parameters

### Global Cross-Entity Updates

**Route Handlers/Middleware**
- [x] 5.41: Search codebase for route definitions or middleware that may call renamed handler methods
- [x] 5.42: Update any route-level references to old method names

**Other Layer Integration**
- [x] 5.43: Search `internal/` for any other references to old repository method names or `userId` parameters
- [x] 5.44: Update cross-layer calls not covered by entity-specific tasks

### Verification

- [x] 5.45: Build verification: `go build ./...` (zero errors)
- [x] 5.46: Full test verification: `go test ./... -v` (all tests pass, including integration tests)
- [x] 5.47: Grep verification: `grep -r "GetById\|GetAllByUserId\|GetLastThreeByUserId" internal/` (should return zero)
- [x] 5.48: Parameter name verification: `grep -r "\busterId\b" internal/` (should return zero occurrences of old pattern)
- [x] 5.49: Manual diff review: Check that all renaming is consistent and complete
- [x] 5.50: Verify no breaking changes to method signatures (parameters, return types)

### Verification Steps
```bash
go build ./...
go test ./... -v
grep -r "GetById\|GetAllByUserId\|GetLastThreeByUserId\|GetLastThreeByUserId" internal/ | grep -v "test" | wc -l
# Expected: 0 (after Commit 5)

grep -r "\buserId\b" internal/ | grep -v "test" | wc -l
# Expected: 0 (after Commit 5)

grep -r "func.*userId" internal/ | wc -l
# Expected: 0 (renamed to userID)
```

---

## Final Verification Phase (Post-Commit 5)

### Comprehensive Testing & Linting

**Build & Test Suite**
- [x] 5.51: Run `go build ./...` one final time (zero errors expected)
- [x] 5.52: Run full test suite: `go test ./... -v` (100% pass rate expected, report count)
- [x] 5.53: Run integration tests if available (e.g., database connection tests)

**Linting & Code Quality**
- [x] 5.54: Linting check: `golangci-lint run ./...` (if available; zero naming violations expected)
- [x] 5.55: Manual naming review: Confirm all struct fields follow PascalCase (exported) convention
- [x] 5.56: Manual review: Confirm all acronyms are ALL-CAPS (ID, MP, HTTP, not Id, Mp, Http)

**API Contract Verification**
- [x] 5.57: Manual curl/Postman test of a few endpoints (e.g., `GET /proof/{id}`, `POST /user`) to verify JSON responses use snake_case field names (unchanged from before refactoring)

**GORM Integration**
- [x] 5.58: Verify GORM integration — sample query test for each entity (User, Proof, Voucher, Token) to ensure field renames don't break database queries

**Documentation & Checklist**
- [x] 5.59: Record all renamed fields and methods in commit message for reference
- [x] 5.60: Create final verification report with all results

### Final Verification Commands
```bash
# Build
go build ./...

# Full test suite
go test ./... -v

# Linting (if available)
golangci-lint run ./...

# Grep final verification
echo "=== Checking for old field patterns ==="
grep -r "Locked_until\|ID_MP\|Date_Approved_MP\|Operation_Type_MP\|Status_MP\|Amount_MP\|Is_Revoked" internal/ | grep -v "json:\|gorm:" | wc -l
# Expected: 0

echo "=== Checking for old method patterns ==="
grep -r "GetById\|GetAllByUserId\|GetLastThreeByUserId" internal/ | wc -l
# Expected: 0

echo "=== Checking for old parameter patterns ==="
grep -r "\busterId\b" internal/ | wc -l
# Expected: 0
```

---

## Task Summary

| Phase | Commit | Tasks | Focus |
|-------|--------|-------|-------|
| A.1 | Commit 1 | 1.1–1.14 (14 tasks) | User entity field rename + tests |
| A.2 | Commit 2 | 2.1–2.29 (29 tasks) | Proof entity fields + cross-layer refs + tests |
| A.3 | Commit 3 | 3.1–3.17 (17 tasks) | Voucher entity fields + tests |
| A.4 | Commit 4 | 4.1–4.13 (13 tasks) | Token entity fields + tests |
| B | Commit 5 | 5.1–5.50 (50 tasks) | Repository methods, service, handler renames |
| — | Final | 5.51–5.60 (10 tasks) | Verification, testing, linting |
| **Total** | **5 commits + verification** | **~143 tasks** | **All naming violations** |

---

## Implementation Notes

### IDE-Guided Refactoring Approach

For all rename operations, use IDE refactoring tools (GoLand, VS Code Go extension) to ensure:
- All call sites are updated automatically
- No references are missed
- Type safety is maintained

**Steps**:
1. Right-click on field/method name
2. Select "Refactor" → "Rename"
3. Enter new name
4. Review all usages
5. Apply rename

### Build & Test After Each Commit

After committing each phase:
```bash
go build ./...
go test ./... -v
```

Do NOT move to the next commit until build and tests pass.

### Grep Verification

Use grep to confirm old patterns are gone (focus on code, not comments/tags):
```bash
grep -r "OLD_PATTERN" internal/ | grep -v "json:\|gorm:\|column:" | wc -l
# Should be zero
```

### Cross-Layer Dependencies

Phase B updates are in a **single commit** to ensure:
- Repository method rename → Service call updates → Handler call updates all happen together
- No partial state (all dependencies resolved in one changeset)
- Easier rollback if needed

---

## Risk Mitigation

| Risk | Mitigation |
|------|-----------|
| GORM column mapping breaks | Verify tags match database schema; write query test per entity |
| JSON API contract changes | Verify JSON tags remain snake_case (only struct fields change) |
| Missed call sites | Use IDE refactoring; manual grep verification post-refactor |
| Test failures | Run `go test ./...` after each phase |
| Partial refactor | Build verification catches compile errors immediately |

---

## Next Steps (After Implementation)

1. **Code Review**: Request team review of all 5 commits
2. **Integration Testing**: Test with real database and API clients
3. **Merge**: Merge all 5 commits to main branch
4. **Archive**: Move change to `openspec/archive/` with completion report

