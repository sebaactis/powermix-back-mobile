# Domain Error Wrapping Specification

## Purpose

Preserve error context across layers by wrapping errors with `%w` and mapping GORM/database errors to domain sentinels.

## Requirements

### Requirement: Error wrapping across layers

Every repository method MUST wrap errors with domain context using `fmt.Errorf("domain context: %w", err)`. Every service MUST propagate wrapped errors without discarding the cause chain.

#### Scenario: Repository wraps GORM error

- GIVEN a repository method that calls `db.First(&user, id)`
- WHEN `gorm.ErrRecordNotFound` is returned
- THEN the repository MUST return `fmt.Errorf("usuario con id %d no encontrado: %w", id, gorm.ErrRecordNotFound)`
- AND the handler MUST be able to `errors.Is(err, gorm.ErrRecordNotFound)`

### Requirement: GORM error mapping

Each repository MUST map GORM driver errors (e.g., `ErrDuplicatedKey`, `ErrRecordNotFound`) to domain sentinels before returning them to the service layer.

#### Scenario: Duplicate key maps to domain sentinel

- GIVEN a repository insert that violates a unique constraint
- WHEN GORM returns `ErrDuplicatedKey`
- THEN the repository MUST return a domain-specific sentinel (e.g., `user.ErrDuplicateEmail`)
- AND the handler MUST identify it via `errors.Is(err, user.ErrDuplicateEmail)`

#### Scenario: Raw database error does not leak

- GIVEN an unexpected database error (e.g., connection dropped)
- WHEN the repository catches it
- THEN the repository MUST log the raw error and return a generic `ErrInternal` sentinel
- AND the raw GORM error MUST NOT reach the HTTP response
