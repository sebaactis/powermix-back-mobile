# Error Codes Specification

## Purpose

Define a consistent error code taxonomy so the frontend can programmatically act on errors instead of parsing message strings.

## Requirements

### Requirement: Error code taxonomy

The APIError response MUST include a `code` field with a string from a predefined taxonomy. The system MUST use machine-readable codes (e.g., `ERR_VALIDATION`, `ERR_NOT_FOUND`, `ERR_DUPLICATE_ENTRY`, `ERR_INTERNAL`) and human-readable `message` in Spanish.

#### Scenario: Error code present on failed response

- GIVEN a request that fails (e.g., POST /api/login with wrong password)
- WHEN the server responds with 401
- THEN the JSON body MUST contain `error.code` set to `ERR_INVALID_CREDENTIALS`
- AND `error.message` MUST be `"Credenciales inválidas"`

#### Scenario: Unknown error uses generic code

- GIVEN an unexpected server error with no matching code
- WHEN the server responds with 500
- THEN `error.code` MUST be `ERR_INTERNAL`
- AND `error.message` MUST NOT expose raw error details

### Requirement: Frontend consumes error code

The frontend SHOULD check `error.code` before falling back to `error.message` when displaying errors.

#### Scenario: Validation error shows field highlights

- GIVEN a frontend form submission that returns `{ error: { code: "ERR_VALIDATION", fields: { email: "Email inválido" } } }`
- WHEN the response is parsed
- THEN the component MUST highlight the `email` field with the fields message
