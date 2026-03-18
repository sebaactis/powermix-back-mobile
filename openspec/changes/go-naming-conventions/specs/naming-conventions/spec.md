# Go Naming Conventions Specification

## Purpose

This specification defines idiomatic Go naming conventions for the powermix-mobile-backend codebase. 
The codebase currently contains 50+ naming violations that reduce readability and maintainability 
for Go developers. This domain brings the codebase into compliance with 
[Effective Go naming conventions](https://golang.org/doc/effective_go#names), improving code clarity 
and consistency across all entity layers (models, DTOs, services, handlers).

The system MUST follow Go naming standards, which include:
- PascalCase (exported identifiers) and camelCase (unexported identifiers)
- No underscores in struct field names
- ALL-CAPS for multi-letter acronyms (ID, MP, HTTP, not Id, Mp, Http)
- Full words over single-letter variables (user vs u, service vs s) in public APIs
- Consistent JSON tags using snake_case for API contracts

---

## Requirements

### REQ-1: Struct Field Names Use PascalCase Without Underscores

The system MUST rename all struct fields using snake_case or mixed case with underscores 
to idiomatic PascalCase (camelCase for exported fields).

No struct field name SHALL contain underscores. Each renamed field MUST have a corresponding 
snake_case JSON tag for API compatibility.

#### Scenario: User.LockedUntil field rename

- **GIVEN** the User struct has a field named `Locked_until` (underscore, non-idiomatic)
- **WHEN** code is refactored for Go idioms
- **THEN** the field is renamed to `LockedUntil` (PascalCase, no underscores)
- **AND** the JSON tag is set to `json:"locked_until"` to preserve API contract

#### Scenario: Proof MercadoPago fields rename

- **GIVEN** the Proof struct contains fields:
  - `ID_MP` (underscore + mixed case)
  - `Date_Approved_MP` (underscores + mixed case)
  - `Status_MP` (underscores + mixed case)
  - `Amount_MP` (underscores + mixed case)
  - `Is_Revoked` (underscore + mixed case)
- **WHEN** code is refactored for Go idioms
- **THEN** fields are renamed to:
  - `IDMP` (all-caps acronyms)
  - `DateApprovedMP` (PascalCase + all-caps acronym)
  - `StatusMP` (PascalCase + all-caps acronym)
  - `AmountMP` (PascalCase + all-caps acronym)
  - `IsRevoked` (PascalCase)
- **AND** JSON tags are set to snake_case variants for API marshaling:
  - `json:"id_mp"`
  - `json:"date_approved_mp"`
  - `json:"status_mp"`
  - `json:"amount_mp"`
  - `json:"is_revoked"`

#### Scenario: DTO field consistency

- **GIVEN** DTOs in `user/dto.go` and `proof/dto.go` have underscore-separated fields
- **WHEN** DTOs are refactored to match entity naming
- **THEN** all DTO struct fields follow PascalCase convention
- **AND** JSON tags remain snake_case for API request/response marshaling

---

### REQ-2: Repository Methods Use GetByID Pattern (Not GetById)

The system MUST rename repository methods to use ALL-CAPS for acronyms. 
The pattern `GetByID` is idiomatic Go; `GetById` is not.

All repository method names containing acronyms SHALL use all-caps for multi-letter acronyms 
(ID, not Id; MP, not Mp; HTTP, not Http).

#### Scenario: Proof repository GetByID method

- **GIVEN** the ProofRepository has a method named `GetById(id string)` 
- **WHEN** repository methods are refactored for Go idioms
- **THEN** the method is renamed to `GetByID(id string)`
- **AND** all call sites (service, handler, tests) are updated to use `GetByID`

#### Scenario: Plural methods with UserID acronym

- **GIVEN** ProofRepository has methods:
  - `GetAllByUserId(userID string)` (mixed case acronym)
  - `GetLastThreeByUserId(userID string)` (mixed case acronym)
- **WHEN** repository methods are refactored
- **THEN** methods are renamed to:
  - `GetAllByUserID(userID string)` (ALL-CAPS ID)
  - `GetLastThreeByUserID(userID string)` (ALL-CAPS ID)
- **AND** parameter names are updated to `userID` (all-caps acronym)
- **AND** all call sites are updated to use new method names

#### Scenario: VoucherRepository method consistency

- **GIVEN** VoucherRepository methods follow the same pattern as ProofRepository
- **WHEN** service and handler layers call these methods
- **THEN** all method calls use the new `GetByID`, `GetAllByUserID` naming
- **AND** parameter passing uses `userID` (all-caps)

---

### REQ-3: Acronym Capitalization Must Be ALL-CAPS

The system MUST use ALL-CAPS for all multi-letter acronyms in variable, parameter, 
and field names. Single-letter acronyms use PascalCase (A, not a).

Function parameters and local variables SHALL be updated to use all-caps acronyms 
for consistency with struct fields.

#### Scenario: Function parameters with userID

- **GIVEN** functions have parameters named `userId` (lowercase acronym)
- **WHEN** code is refactored for naming consistency
- **THEN** parameter names are updated to `userID` (all-caps ID)
- **AND** function signatures are updated across service, handler, and repository layers

#### Scenario: Local variable acronyms in handler functions

- **GIVEN** handler functions assign user IDs to variables like `uid := req.UserID` 
- **WHEN** code is refactored
- **THEN** variable names use `userID`, not `uid` or `userId`
- **AND** the pattern is applied consistently to all handler methods

#### Scenario: MercadoPago acronym in variables

- **GIVEN** local variables reference MercadoPago fields like `mp_id := proof.ID_MP`
- **WHEN** code is refactored
- **THEN** variable names become `mpID := proof.IDMP` (all-caps for both acronyms)
- **AND** the assignment uses the renamed struct field

---

### REQ-4: Variable Names Use Full Words in Public APIs

The system SHOULD use full words over single-letter variable names in function parameters 
and handler context where code clarity is important.

In handlers and services, single-letter receiver variables are acceptable per Go idiom 
(e.g., `u *User` for methods), but parameters and context variables SHOULD be explicit.

#### Scenario: Handler context variables

- **GIVEN** handler methods use single-letter context variables:
  - `u, err := h.service.Create(...)` (u for user)
  - `s := h.service` (s for service)
  - `r := h.repo` (r for repo)
- **WHEN** code is reviewed for clarity in HTTP handlers
- **THEN** handler context variables are renamed to full names:
  - `user, err := h.service.Create(...)`
  - `service := h.service`
  - `repo := h.repo`
- **AND** the pattern is applied to new handler code

#### Scenario: Service function parameters

- **GIVEN** service methods have parameters named with single letters
- **WHEN** code is refactored
- **THEN** function parameters use full descriptive names
- **AND** this improves code documentation without needing additional comments

#### Scenario: Test context variables (optional)

- **GIVEN** test functions use context like `u := &User{...}`
- **WHEN** test code is updated
- **THEN** test variables may continue using single-letter receivers per Go idiom
- **AND** this is acceptable if it does not reduce test readability

---

### REQ-5: Method Receiver Consistency

The system MUST maintain consistency with Go idioms for method receivers. 
Single-letter receivers are idiomatic Go and SHOULD be retained when consistent with project patterns.

Method receivers (e.g., `func (u *User) Method()`) follow established Go conventions 
and need not be changed if the project uses them consistently.

#### Scenario: User method receiver

- **GIVEN** the User struct has methods with receiver `u *User`
- **WHEN** code is reviewed for naming consistency
- **THEN** single-letter receiver `u` is retained (idiomatic Go)
- **AND** this matches the project's established receiver pattern

#### Scenario: Proof method receiver consistency

- **GIVEN** Proof and other entity methods use single-letter receivers consistently
- **WHEN** the codebase is refactored
- **THEN** method receivers are kept as-is (not renamed to `proof` or `p`)
- **AND** the pattern remains consistent across all entities

#### Scenario: Receiver abbreviation validation

- **GIVEN** all methods have receivers matching entity first letter or standard abbreviation
- **WHEN** naming conventions are verified
- **THEN** receiver names match project conventions and Go idioms
- **AND** no changes are required to receiver names

---

### REQ-6: JSON Tag Consistency Must Use snake_case

The system MUST ensure all JSON tags use snake_case format for API contracts. 
This is independent of struct field naming and MUST NOT be changed.

When struct fields are renamed from underscore format to PascalCase, 
the JSON tags MUST be updated to use snake_case equivalents.

#### Scenario: JSON tag mapping after field rename

- **GIVEN** a struct field `Locked_until` with JSON tag `json:"locked_until"`
- **WHEN** the field is renamed to `LockedUntil`
- **THEN** the JSON tag remains `json:"locked_until"` (unchanged, snake_case)
- **AND** API responses continue to use `locked_until` in JSON payloads

#### Scenario: MercadoPago field JSON tags

- **GIVEN** Proof fields are renamed:
  - `ID_MP` → `IDMP` with tag `json:"id_mp"`
  - `Date_Approved_MP` → `DateApprovedMP` with tag `json:"date_approved_mp"`
  - `Status_MP` → `StatusMP` with tag `json:"status_mp"`
  - `Amount_MP` → `AmountMP` with tag `json:"amount_mp"`
- **WHEN** the struct is marshaled for API responses
- **THEN** JSON payloads use exact snake_case keys (id_mp, date_approved_mp, status_mp, amount_mp)
- **AND** the mapping is verified with JSON marshaling tests

#### Scenario: Null/blank JSON tag handling

- **GIVEN** some struct fields may have JSON tags with omitempty or default behavior
- **WHEN** fields are renamed
- **THEN** all JSON tag options (omitempty, default values) are preserved
- **AND** marshaling behavior remains unchanged

#### Scenario: DTO JSON tag consistency

- **GIVEN** DTOs in `user/dto.go` and `proof/dto.go` have JSON tags
- **WHEN** DTOs are refactored with PascalCase fields
- **THEN** JSON tags are updated to snake_case equivalents
- **AND** API request/response contracts remain unchanged

---

### REQ-7: GORM Tag Verification

The system MUST verify that GORM struct tags (column mappings) are correct after renaming fields.
No database schema changes are required; only struct tags must be verified for correct mapping.

GORM column tags SHALL reference existing database column names. 
If a struct field name doesn't match the database column, the GORM tag MUST specify the correct column.

#### Scenario: GORM column tag for LockedUntil field

- **GIVEN** the User struct field is renamed from `Locked_until` to `LockedUntil`
- **WHEN** the database column is still named `locked_until`
- **THEN** the GORM tag MUST be `gorm:"column:locked_until"`
- **AND** the field rename is verified with a database query test

#### Scenario: GORM tags for MercadoPago fields

- **GIVEN** Proof struct fields are renamed (ID_MP → IDMP, Date_Approved_MP → DateApprovedMP, etc.)
- **WHEN** database columns maintain their original naming (id_mp, date_approved_mp, etc.)
- **THEN** GORM tags are verified or added:
  - `IDMP` with tag `gorm:"column:id_mp"`
  - `DateApprovedMP` with tag `gorm:"column:date_approved_mp"`
  - `StatusMP` with tag `gorm:"column:status_mp"`
  - `AmountMP` with tag `gorm:"column:amount_mp"`
- **AND** no database schema migration is required

#### Scenario: GORM tag validation with test queries

- **GIVEN** struct field renames are complete with GORM tags
- **WHEN** a test queries the database (e.g., `repo.GetByID(...)`)
- **THEN** the query successfully maps database rows to renamed struct fields
- **AND** all fields are correctly populated from database columns

---

### REQ-8: Cross-Layer Method Rename Propagation

The system MUST ensure all method renames are propagated across repository, service, 
and handler layers. No call site SHALL be missed.

When a method is renamed in the repository layer (e.g., `GetById` → `GetByID`), 
all callers in service and handler layers MUST be updated.

#### Scenario: Repository method rename propagates to service

- **GIVEN** ProofRepository method is renamed from `GetById(id string)` to `GetByID(id string)`
- **WHEN** the ProofService calls this method
- **THEN** the service layer call is updated to `repo.GetByID(id)`
- **AND** the service method signature remains compatible

#### Scenario: Service method rename propagates to handler

- **GIVEN** ProofService calls repository method
- **WHEN** the ProofHandler calls the service method
- **THEN** all handler call sites are updated to use renamed methods
- **AND** no old method names remain in the handler

#### Scenario: Parameter renames propagate through all layers

- **GIVEN** a method parameter is renamed from `userId` to `userID`
- **WHEN** the method is called across service and handler layers
- **THEN** all call sites pass parameters using the new name `userID`
- **AND** variable assignments and references are updated consistently

#### Scenario: Test call sites updated

- **GIVEN** tests call repository methods by old names (e.g., `repo.GetById(...)`)
- **WHEN** the method is renamed to `GetByID`
- **THEN** all test call sites are updated to use `GetByID(...)`
- **AND** test cases continue to pass with new names

---

## Acceptance Criteria

The implementation SHALL be considered complete when ALL of the following criteria are met:

- [ ] **All 50+ naming violations identified in exploration are fixed** — Struct fields, method names, parameters, and variables follow Go naming standards
- [ ] **Build succeeds** — `go build ./...` produces zero errors or warnings related to naming
- [ ] **All tests pass** — `go test ./...` shows 100% pass rate; no test failures due to naming changes
- [ ] **Code follows Effective Go conventions** — Manual review confirms compliance with [Effective Go naming guide](https://golang.org/doc/effective_go#names)
- [ ] **JSON API contracts unchanged** — External API clients receive identical JSON payloads; snake_case tags verified
- [ ] **GORM tags verified** — Database queries continue working; column mappings are correct
- [ ] **No breaking changes** — Internal interfaces renamed; external API contracts (JSON, database) unchanged
- [ ] **Linting clean** — `golangci-lint run ./...` shows zero naming-related violations
- [ ] **Cross-layer consistency** — Repository, service, and handler methods use consistent naming
- [ ] **Code review approved** — Team review confirms naming changes are correct and complete

---

## Implementation Phases

### Phase A: Entity Models & DTOs
1. Rename struct fields in models (user.go, proof.go, voucher.go, token.go)
2. Update JSON tags to snake_case
3. Verify GORM column mappings
4. Update DTO struct fields
5. Commit per entity; verify build and tests after each commit

### Phase B: Service & Handler Layers
1. Rename repository methods (GetById → GetByID, etc.)
2. Rename function parameters (userId → userID)
3. Rename handler context variables (u → user, s → service)
4. Update all call sites using IDE-guided refactoring
5. Single commit for phase B; verify full test suite and build

### Phase C: Verification
1. Full test suite passes
2. Build succeeds with zero errors
3. Linting checks pass
4. Manual code review confirms correctness
5. Integration testing with real database

---

## Related Documents

- **Proposal**: `openspec/changes/go-naming-conventions/proposal.md`
- **Design**: TBD (sdd-design phase)
- **Tasks**: TBD (sdd-tasks phase)
- **Reference**: [Effective Go - Names](https://golang.org/doc/effective_go#names)

---

## Not in Scope

- **Database schema changes** — Column names remain unchanged; only struct tags are updated
- **External API contract changes** — JSON field names (via tags) remain snake_case; no client impact
- **Configuration or infrastructure naming** — config.go, routes, middleware unchanged
- **Test rewrites** — Existing tests updated only to use new names; no new test coverage added
- **Documentation/Godoc comments** — Comment updates deferred to separate change
- **Third-party library integration** — No external packages renamed; only internal code
- **Performance optimization** — Naming changes do not affect performance; no benchmarking required
