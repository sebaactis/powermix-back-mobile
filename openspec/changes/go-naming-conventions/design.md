# Design: Go Naming Conventions Refactor

**Change ID**: C2  
**Date**: 2026-03-17  
**Status**: Design Phase

---

## Technical Approach

This design document outlines the technical strategy for refactoring the powermix-mobile-backend codebase from non-idiomatic Go naming to comply with [Effective Go conventions](https://golang.org/doc/effective_go#names).

The refactor is executed in **two sequential phases** to minimize risk and enable incremental verification:

- **Phase A**: Entity Models & DTOs (rename struct fields, update JSON/GORM tags)
- **Phase B**: Service & Handler Layers (rename methods, update variables and call sites)

This phased approach ensures:
1. **Dependency ordering**: Models are foundational; services/handlers depend on model names
2. **Reviewability**: Smaller, focused PRs (one per phase) reduce review burden
3. **Rollback safety**: Each phase can be independently reverted
4. **Test isolation**: Phase A changes are orthogonal to Phase B logic

---

## Architecture Decisions

### Decision: Entity-First Refactoring (Phase A Before Phase B)

**Choice**: Refactor entity models and DTOs first (Phase A), then propagate renames to service/handler layers (Phase B).

**Alternatives Considered**:
- **Simultaneous refactor**: Change all layers at once → High risk, massive diff, difficult to review, harder to rollback selectively
- **Backwards-compat layer**: Keep old names as aliases → Adds technical debt, confuses developers, not sustainable

**Rationale**: 
- Entity models are the source of truth for struct field names
- DTOs directly translate models to API contracts
- Service/handler layers depend on model field names; renaming them first ensures downstream layers use correct names
- Phased approach enables verification and testing at each step, reducing integration risk
- Entity-first allows intermediate testing of GORM tags and JSON marshaling before touching business logic

---

### Decision: Per-Entity Commits in Phase A

**Choice**: Separate git commits for each entity (User, Proof, Voucher, Token) within Phase A.

**Alternatives Considered**:
- Single Phase A commit for all entities → Harder to bisect/debug if issues arise
- Granular commits per file per entity → Too fragmented, harder to understand change scope

**Rationale**:
- Enables **bisection** if bugs are discovered post-merge
- Each commit is **independently testable** (build + tests after each)
- **Clear semantic grouping**: "User entity refactor" is self-contained
- Allows team to provide feedback after each entity commit before proceeding to next

---

### Decision: IDE-Guided Refactoring (No Manual Search-Replace)

**Choice**: Use IDE refactoring tools (GoLand, VS Code Go extension) for all renames; no manual find-and-replace.

**Alternatives Considered**:
- Manual search-replace with git/grep → Error-prone, easy to miss call sites in comments or strings
- Automated script with AST parsing → Overkill for this project size; IDE tools sufficient

**Rationale**:
- IDE refactoring **understands Go semantics**: distinguishes between types, functions, variables, imports
- **Automatic call site updates**: All references across the codebase are updated simultaneously
- **Rename validation**: IDE detects conflicts (e.g., shadowing, scope issues) before applying
- **Testability**: Build and tests immediately after refactor reveal any missed sites
- **Go team recommendation**: IDE-guided refactoring is the Go community best practice

---

### Decision: JSON Tags Remain snake_case

**Choice**: Go struct fields are renamed to PascalCase; JSON tags continue using snake_case.

**Alternatives Considered**:
- Change JSON tags to camelCase → Breaks external API contracts; client impact
- No JSON tags (use default marshaling) → Loses snake_case format; client impact

**Rationale**:
- **API stability**: External clients depend on JSON field names (snake_case)
- **Go idiom**: Go structs are PascalCase; marshaling format (JSON) is independent
- **REST convention**: snake_case is standard for REST APIs; camelCase is JavaScript convention
- **GORM consistency**: Maintains column name consistency (snake_case in database, struct-level tagging)

---

### Decision: GORM Tags Explicit for Clarity

**Choice**: Ensure all GORM column tags are explicit and correct after field renames (no relying on field name inference).

**Alternatives Considered**:
- Rely on GORM's default field name → struct → snake_case mapping → Ambiguous after renames; fragile
- Remove explicit tags, use field names → Tight coupling between struct names and schema

**Rationale**:
- **Clarity**: Explicit `gorm:"column:..."` tags make the struct-to-database mapping obvious
- **Safety**: Refactoring is isolated from GORM defaults; won't break if GORM changes behavior
- **Testability**: Can verify mapping with a simple `SELECT * FROM table` test query
- **Maintenance**: Future developers immediately see which database column each field maps to

---

### Decision: Single Phase B Commit for Service/Handler

**Choice**: All Phase B renames (repository methods, service signatures, handler variables) committed together.

**Alternatives Considered**:
- Per-entity commits in Phase B → Easier to bisect, but Entity boundaries don't align with logic boundaries
- Per-layer commits (repository, service, handler) → Breaks consistency; handler rename requires service rename

**Rationale**:
- **Logical cohesion**: Service method rename and all its call sites must be updated together
- **Reduced risk**: Single commit ensures all dependencies are in one changeset; no partial state
- **Review clarity**: Reviewers see the complete picture of method renames and their propagation

---

## Data Flow

```
Entity Models
  ├─ user.go (LockedUntil, OAuthID, ...)
  ├─ proof.go (IDMP, DateApprovedMP, StatusMP, AmountMP, IsRevoked, ...)
  ├─ voucher.go (similar patterns)
  └─ token.go (similar patterns)
           ↓
           ↓ [Phase A: Rename fields + update JSON/GORM tags]
           ↓
DTOs
  ├─ user/dto.go
  ├─ proof/dto.go
  ├─ voucher/dto.go
  └─ token/dto.go
           ↓
           ↓ [Phase B: Repository methods renamed]
           ↓
Repositories
  ├─ user/repository.go (GetByID, GetAllByUserID, ...)
  ├─ proof/repository.go (GetByID, GetAllByUserID, GetLastThreeByUserID, ...)
  ├─ voucher/repository.go
  └─ token/repository.go
           ↓
           ↓ [Phase B: Service method calls updated]
           ↓
Services
  ├─ user/service.go
  ├─ proof/service.go
  ├─ voucher/service.go
  └─ token/service.go
           ↓
           ↓ [Phase B: Handler method calls updated]
           ↓
HTTP Handlers
  ├─ user/handler.go (GetAllByUserID, GetByID, ...)
  ├─ proof/handler.go
  ├─ voucher/handler.go
  └─ token/handler.go
           ↓
           ↓ [No change to external JSON API format]
           ↓
HTTP Responses (JSON payloads unchanged)
```

---

## File Changes

### Phase A: Entity Models & DTOs

| File | Action | Description |
|------|--------|-------------|
| `internal/domain/entities/user/user.go` | Modify | Rename `Locked_until` → `LockedUntil`; verify JSON tag is `locked_until`; verify GORM tag is `column:"locked_until"` |
| `internal/domain/entities/user/dto.go` | Modify | Update DTO struct fields to match model (if DTO fields mirror model fields) |
| `internal/domain/entities/proof/proof.go` | Modify | Rename: `ID_MP` → `IDMP`, `Date_Approved_MP` → `DateApprovedMP`, `Operation_Type_MP` → `OperationTypeMP`, `Status_MP` → `StatusMP`, `Amount_MP` → `AmountMP`, `Is_Revoked` → `IsRevoked`; verify all GORM and JSON tags |
| `internal/domain/entities/proof/dto.go` | Modify | Update DTO fields to match renamed Proof struct fields; verify JSON tags use snake_case |
| `internal/domain/entities/voucher/voucher.go` | Modify | Rename any underscore-separated fields (if present) |
| `internal/domain/entities/voucher/dto.go` | Modify | Update DTO fields if needed |
| `internal/domain/entities/token/token.go` | Modify | Rename any underscore-separated fields (if present) |
| `internal/domain/entities/token/dto.go` | Modify | Update DTO fields if needed |

### Phase B: Service & Handler Layers

| File | Action | Description |
|------|--------|-------------|
| `internal/domain/entities/user/repository.go` | Modify | Rename methods: `GetById` → `GetByID`, `GetAllByUserId` → `GetAllByUserID`; update parameter names to `userID` |
| `internal/domain/entities/user/service.go` | Modify | Update method calls to repository (use new method names); update local variable names (e.g., `userId` → `userID` where applicable) |
| `internal/domain/entities/user/handler.go` | Modify | Update service method calls; rename handler context variables (`u` → `user`, `s` → `service`); update parameter names |
| `internal/domain/entities/proof/repository.go` | Modify | Rename methods: `GetAllByUserId` → `GetAllByUserID`, `GetAllByUserIdPaginated` → `GetAllByUserIDPaginated`, `GetById` → `GetByID`; update parameter names |
| `internal/domain/entities/proof/service.go` | Modify | Update all repository calls with new method names; rename internal variables with `userID` pattern |
| `internal/domain/entities/proof/handler.go` | Modify | Rename methods: `GetAllByUserId` → `GetAllByUserID`, `GetAllByUserIdPaginated` → `GetAllByUserIDPaginated`, `GetById` → `GetByID`; update handler context variables; update service method calls |
| `internal/domain/entities/voucher/repository.go` | Modify | Similar renames as proof (if methods exist) |
| `internal/domain/entities/voucher/service.go` | Modify | Update repository calls |
| `internal/domain/entities/voucher/handler.go` | Modify | Rename methods; update calls |
| `internal/domain/entities/token/repository.go` | Modify | Similar renames (if methods exist) |
| `internal/domain/entities/token/service.go` | Modify | Update repository calls |
| `internal/domain/entities/token/handler.go` | Modify | Rename methods; update calls |
| Tests (all `*_test.go` files) | Modify | Update test calls to use new method names and field names |

---

## Interfaces / Contracts

### Before Phase A

**User Struct** (non-idiomatic):
```go
type User struct {
	ID            uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name          string    `gorm:"not null"`
	Email         string    `gorm:"not null;unique"`
	Password      string    `gorm:"not null"`
	StampsCounter int       `gorm:"default:0"`
	LoginAttempt  int       `json:"login_attempt" gorm:"default:0"`
	Locked_until  time.Time `json:"locked_until" gorm:"default:null"`  // ← VIOLATION
	OAuthProvider string    `gorm:"column:oauth_provider;type:varchar(20);default:null"`
	OAuthID       string    `gorm:"column:oauth_id;type:varchar(100);default:null"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
```

### After Phase A

**User Struct** (idiomatic):
```go
type User struct {
	ID            uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name          string    `gorm:"not null"`
	Email         string    `gorm:"not null;unique"`
	Password      string    `gorm:"not null"`
	StampsCounter int       `gorm:"default:0"`
	LoginAttempt  int       `json:"login_attempt" gorm:"default:0"`
	LockedUntil   time.Time `json:"locked_until" gorm:"column:locked_until;default:null"`  // ← FIXED
	OAuthProvider string    `gorm:"column:oauth_provider;type:varchar(20);default:null"`
	OAuthID       string    `gorm:"column:oauth_id;type:varchar(100);default:null"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
```

### Proof Struct (Before Phase A)

```go
type Proof struct {
	ID                uuid.UUID           `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID            uuid.UUID           `gorm:"not null"`
	ID_MP             string              `gorm:"unique, not null"`  // ← VIOLATION
	Date_Approved_MP  utils.FormattedTime `gorm:"not null"`          // ← VIOLATION
	Operation_Type_MP string              `gorm:"not null"`          // ← VIOLATION
	Status_MP         string              `gorm:"not null"`          // ← VIOLATION
	Amount_MP         float64             `gorm:"not null"`          // ← VIOLATION
	ProofDate         utils.FormattedTime `gorm:"not null"`
	Dni               *string
	CardId            *string
	CardType          *string
	Last4Card         *string
	ExternalID        *string
	ProductName       *string
}
```

### Proof Struct (After Phase A)

```go
type Proof struct {
	ID               uuid.UUID           `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID           uuid.UUID           `gorm:"not null"`
	IDMP             string              `gorm:"column:id_mp;unique,not null"`       // ← FIXED
	DateApprovedMP   utils.FormattedTime `gorm:"column:date_approved_mp;not null"`   // ← FIXED
	OperationTypeMP  string              `gorm:"column:operation_type_mp;not null"`  // ← FIXED
	StatusMP         string              `gorm:"column:status_mp;not null"`          // ← FIXED
	AmountMP         float64             `gorm:"column:amount_mp;not null"`          // ← FIXED
	ProofDate        utils.FormattedTime `gorm:"not null"`
	Dni              *string
	CardID           *string            // Bonus: CardId → CardID (acronym)
	CardType         *string
	Last4Card        *string
	ExternalID       *string
	ProductName      *string
}
```

### Repository Interface Pattern

**Before Phase B**:
```go
// proof/repository.go
func (r *Repository) GetById(ctx context.Context, id string) (*Proof, error) { ... }
func (r *Repository) GetAllByUserId(ctx context.Context, userId uuid.UUID) ([]*Proof, error) { ... }
func (r *Repository) GetAllByUserIdPaginated(ctx context.Context, userId uuid.UUID, page int, pageSize int, filters ProofFilters) ([]*Proof, int64, error) { ... }
```

**After Phase B**:
```go
// proof/repository.go
func (r *Repository) GetByID(ctx context.Context, id string) (*Proof, error) { ... }
func (r *Repository) GetAllByUserID(ctx context.Context, userID uuid.UUID) ([]*Proof, error) { ... }
func (r *Repository) GetAllByUserIDPaginated(ctx context.Context, userID uuid.UUID, page int, pageSize int, filters ProofFilters) ([]*Proof, int64, error) { ... }
```

### JSON Tag Verification

**Example API Response** (unchanged before/after):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "660e8400-e29b-41d4-a716-446655440001",
  "id_mp": "12345678-MP",
  "date_approved_mp": "2026-03-17T15:30:00Z",
  "operation_type_mp": "payment",
  "status_mp": "approved",
  "amount_mp": 99.99,
  "proof_date": "2026-03-17T14:00:00Z",
  "locked_until": "2026-03-18T14:00:00Z"
}
```

JSON marshaling field names are **controlled by JSON tags** and **independent of struct field names**. Renames are transparent to API clients.

---

## Testing Strategy

| Layer | What to Test | Approach | Files |
|-------|-------------|----------|-------|
| **Model** | Struct field renaming completeness | Grep for old names to ensure none remain | user.go, proof.go, etc. |
| **JSON Marshaling** | JSON tags produce correct snake_case keys | Write/run marshaling test: `json.Marshal(proof)` → verify keys are snake_case | proof_test.go (new or extend) |
| **GORM Query** | Database query works with renamed struct fields; column mapping correct | Test `repo.GetByID(...)` and other methods; verify fields populated from DB | proof_test.go, user_test.go |
| **Cross-layer Calls** | Repository method rename propagates to service/handler | Search for old method names in service/handler; ensure none remain | service.go, handler.go |
| **Integration** | Full request/response cycle with renamed entities | HTTP test (POST /proof, GET /proof/{id}) → verify response JSON unchanged | e2e tests (if any) |
| **Build** | No compilation errors after refactoring | `go build ./...` | All .go files |
| **Linting** | No naming violations per golangci-lint | `golangci-lint run ./...` (focus on naming rules) | All .go files |

### Phase A Testing

After each entity commit:
1. **Build verification**: `go build ./...` (must pass)
2. **Unit tests**: `go test ./... -v` (all must pass)
3. **JSON marshaling**: Manual or automated test of `json.Marshal(entity)` → verify tags
4. **GORM queries**: Execute a sample query to verify column mapping works

### Phase B Testing

After Phase B commit:
1. **Build verification**: `go build ./...`
2. **Full test suite**: `go test ./... -v`
3. **Integration tests**: HTTP calls to all renamed endpoints
4. **Grep verification**: Search for old method/variable names to ensure none remain
5. **Linting**: `golangci-lint run ./...` (verify naming rule compliance)

---

## Risk Mitigation

| Risk | Likelihood | Severity | Mitigation |
|------|------------|----------|-----------|
| **GORM column mapping breaks** | Medium | High | 1. Explicit GORM tags on all renamed fields; 2. Write integration test querying DB for each renamed entity; 3. Schema inspection before/after (table structure unchanged) |
| **JSON marshaling field names wrong** | Low | Medium | 1. Verify JSON tags are snake_case before committing; 2. Write marshaling test: `json.Marshal()` and verify keys; 3. Curl/Postman test of real API before merge |
| **Method call site missed** | Medium | High | 1. Use IDE "Find Usages" before renaming (verify all usages listed); 2. Apply IDE refactor (updates all in one pass); 3. Build immediately after to catch any missed sites; 4. Grep for old method names post-refactor |
| **Test failures due to old names** | Low | Medium | 1. Update all test files (*.go) with new method/field names; 2. Run `go test ./...` after each phase to catch failures immediately |
| **Cross-entity dependency breakage** | Low | High | 1. Phase B is a single commit (all dependencies in one changeset); 2. Dependency graph is well-understood (user → proof/voucher/token); 3. IDE refactor updates all call sites simultaneously |
| **Partial refactor (some files missed)** | Low | Medium | 1. Grep post-refactor: `grep -r "GetById\|ID_MP\|Locked_until" internal/` → must return zero results; 2. Build verification catches compile errors |
| **Acronym inconsistency** | Low | Low | 1. Define naming rule upfront: all-caps for multi-letter acronyms (ID, MP, HTTP); 2. Manual inspection of refactored code to verify consistency |
| **Large diff hard to review** | Medium | Low | 1. Split into two PRs (Phase A, Phase B) for easier review; 2. Provide detailed commit messages explaining each rename category; 3. Generate list of all renames before committing (helps reviewers understand scope) |
| **IDE refactor limitation** | Low | Medium | 1. Understand IDE limitations (some edge cases may not be refactored); 2. Grep verification post-refactor catches limitations; 3. Fallback to manual verification if needed |

---

## Implementation Order

**Phase A**: Entity Models & DTOs (incremental entity-by-entity)

1. **Commit 1 (Phase A.1)**: User entity field renames
   - Rename `Locked_until` → `LockedUntil` in user.go
   - Verify JSON tag is `locked_until`, GORM tag is `column:locked_until`
   - Update user/dto.go if DTO fields mirror model
   - Update tests that reference old field name
   - Verification: `go build ./...` and `go test ./...` must pass

2. **Commit 2 (Phase A.2)**: Proof entity field renames
   - Rename: `ID_MP` → `IDMP`, `Date_Approved_MP` → `DateApprovedMP`, `Operation_Type_MP` → `OperationTypeMP`, `Status_MP` → `StatusMP`, `Amount_MP` → `AmountMP`, `Is_Revoked` → `IsRevoked`
   - Verify JSON tags: `id_mp`, `date_approved_mp`, `operation_type_mp`, `status_mp`, `amount_mp`, `is_revoked`
   - Verify GORM tags reference correct database columns (unchanged)
   - Update proof/dto.go fields
   - Update tests
   - Verification: `go build ./...` and `go test ./...` must pass

3. **Commit 3 (Phase A.3)**: Voucher entity field renames
   - Rename underscore-separated fields (if present)
   - Update voucher/dto.go
   - Update tests
   - Verification: `go build ./...` and `go test ./...` must pass

4. **Commit 4 (Phase A.4)**: Token entity field renames
   - Rename underscore-separated fields (if present)
   - Update token/dto.go
   - Update tests
   - Verification: `go build ./...` and `go test ./...` must pass

**Phase B**: Service & Handler Layers (single commit for all changes)

5. **Commit 5 (Phase B)**: Repository, Service, and Handler layer renames
   - **Repository layer**: Rename methods globally (GetById → GetByID, GetAllByUserId → GetAllByUserID, GetAllByUserIdPaginated → GetAllByUserIDPaginated, etc.)
   - **Service layer**: Update all repository method calls to use new names
   - **Handler layer**: Rename handler methods; update service method calls; rename handler context variables (u → user, s → service)
   - **Update all tests**: Test files that call renamed methods
   - **Verification**: 
     - `go build ./...` must pass
     - `go test ./...` must pass (all tests)
     - `grep -r "GetById\|GetAllByUserId\|userId\s:=\|^userId" internal/` must return zero results (no old patterns)
     - Manual review of diffs to verify consistency

---

## Verification Checklist

After all commits:

- [ ] Build succeeds: `go build ./...` with zero errors
- [ ] All tests pass: `go test ./... -v` with 100% pass rate
- [ ] No old naming patterns remain:
  - [ ] `grep -r "Locked_until\|ID_MP\|Date_Approved_MP\|Operation_Type_MP\|Status_MP\|Amount_MP\|Is_Revoked" internal/ | grep -v "column:" | grep -v "json:" | wc -l` = 0
  - [ ] `grep -r "GetById\|GetAllByUserId\|GetLastThreeByUserId" internal/ | wc -l` = 0
- [ ] GORM column mapping verified:
  - [ ] Sample query test passes for each entity
  - [ ] Database query finds records by renamed fields
- [ ] JSON API contract unchanged:
  - [ ] Curl/Postman test of all endpoints returns valid JSON
  - [ ] JSON field names are snake_case (verified with response inspection)
- [ ] Linting passes: `golangci-lint run ./...` (zero naming violations)
- [ ] Go conventions followed: Manual review confirms PascalCase fields, GetByID pattern, all-caps acronyms
- [ ] Code review approved by team

---

## Rollback Plan

### Per-Commit Rollback

If a commit introduces issues:

```bash
# Revert Phase A.1 (User entity rename)
git revert <commit-hash-A1>

# Revert Phase A.2 (Proof entity rename)
git revert <commit-hash-A2>

# Etc.
```

### Full Rollback (All Changes)

If the entire change needs to be reverted:

```bash
# Revert Phase B
git revert <commit-hash-B>

# Revert Phase A (in reverse order)
git revert <commit-hash-A4>
git revert <commit-hash-A3>
git revert <commit-hash-A2>
git revert <commit-hash-A1>
```

### Rollback Safety

- **No data migrations**: Field renames are Go-only; database schema unchanged
- **No configuration changes**: config.go untouched
- **No external API changes**: JSON field names (via tags) unchanged; clients unaffected
- **No state corruption**: Rollback can happen at any time without side effects

---

## Technical Dependencies

- **Go 1.25+** (already in use)
- **IDE with Go refactoring** (GoLand 2024.x or VS Code with Go extension 0.43+)
- **GORM knowledge** (column tag syntax: `gorm:"column:field_name"`)
- **Existing test suite** (must pass before and after each commit)
- **Build tool**: `go build` (standard Go toolchain)
- **Testing tool**: `go test` (standard Go toolchain)
- **Linting** (optional but recommended): `golangci-lint` for naming rule verification

---

## Summary

This design establishes a **phased, IDE-guided refactoring strategy** that:

1. **Prioritizes safety**: Each phase is independently verifiable; each commit is buildable and testable
2. **Maintains API stability**: JSON tags and database columns unchanged; external clients unaffected
3. **Uses idiomatic Go patterns**: IDE-guided refactoring ensures Go community best practices
4. **Enables rollback**: Per-commit structure allows fine-grained rollback if needed
5. **Reduces review burden**: Two focused PRs (Phase A, Phase B) vs. one massive changeset

The **two-phase approach** is critical: Models first (Phase A) ensures consistency downstream; Service/Handler layer (Phase B) updates all call sites in a single cohesive commit.

