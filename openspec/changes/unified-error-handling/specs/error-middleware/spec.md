# Error Middleware Specification

## Purpose

Catch unexpected errors that reach the HTTP layer, format them consistently, and log them — preventing raw technical messages from reaching users.

## Requirements

### Requirement: Global catch-all error handler

The chi router MUST register a middleware that catches panics and handler errors not explicitly handled, returning a sanitized 500 response in `APIResponse` format.

#### Scenario: Unhandled error returns sanitized response

- GIVEN a handler that returns an unexpected `fmt.Errorf("connection refused: %w", err)`
- WHEN the error propagates past the handler without being caught
- THEN the middleware MUST respond with HTTP 500 and `{ success: false, error: { code: "ERR_INTERNAL", message: "Error interno del servidor" } }`
- AND the original error MUST be logged via `slog.Error` with the error details

#### Scenario: Timeout middleware uses APIResponse format

- GIVEN the existing Timeout middleware
- WHEN a request times out
- THEN it MUST respond with `{ success: false, error: { code: "ERR_TIMEOUT", message: "La solicitud tardó demasiado" } }`
- AND NOT return a raw `{"error":"timeout"}` JSON string
