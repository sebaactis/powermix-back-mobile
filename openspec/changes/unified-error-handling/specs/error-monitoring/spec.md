# Error Monitoring Specification

## Purpose

Capture runtime errors and crashes in production for debugging, with Sentry integration for both backend and frontend.

## Requirements

### Requirement: Sentry crash reporting

The frontend MUST integrate `@sentry/react-native` and capture unhandled promise rejections, render errors, and manual `Sentry.captureException` calls. The backend SHOULD integrate Sentry via `slog`-compatible transport for error-level log forwarding.

#### Scenario: Unhandled promise rejection captured

- GIVEN a frontend async function that throws without a catch block
- WHEN the rejection reaches Sentry's global handler
- THEN the error MUST appear in the Sentry dashboard with stack trace, device info, and app version

#### Scenario: Backend error logged to Sentry

- GIVEN a backend handler that logs `slog.Error("db connection failed", "err", err)`
- WHEN Sentry integration is active
- THEN the error MUST be forwarded to Sentry with the request path and error details

### Requirement: Request ID correlation

The backend MUST generate a unique request ID per HTTP request, propagate it via `context.Context`, and include it in Sentry events and log lines.

#### Scenario: Request ID links log and Sentry event

- GIVEN a request that triggers an error
- WHEN the error is logged and captured in Sentry
- THEN both the log line and Sentry event MUST include the same `request_id` value
