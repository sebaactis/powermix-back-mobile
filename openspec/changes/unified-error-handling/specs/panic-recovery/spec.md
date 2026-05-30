# Panic Recovery Specification

## Purpose

Prevent server crashes from unhandled panics in request handlers by recovering gracefully and returning a standard API response.

## Requirements

### Requirement: Global panic recovery middleware

The chi router MUST register a recovery middleware that catches panics from any handler and returns a 500 response in the standard `APIResponse` format.

#### Scenario: Handler panics, server stays up

- GIVEN a chi handler that panics (e.g., nil pointer dereference)
- WHEN the panic occurs during request processing
- THEN the middleware MUST recover the panic
- AND respond with HTTP 500 and `{ success: false, error: { code: "ERR_INTERNAL", message: "Error interno del servidor" } }`
- AND the server MUST continue serving subsequent requests

#### Scenario: Panic logged with stack trace

- GIVEN a recovered panic
- WHEN the middleware catches it
- THEN the middleware MUST log the panic message and stack trace via `slog.Error`
