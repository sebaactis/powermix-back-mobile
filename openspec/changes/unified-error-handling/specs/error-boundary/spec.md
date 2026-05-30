# Error Boundary Specification

## Purpose

Prevent React Native render crashes from showing a white screen by catching rendering errors and displaying a fallback UI.

## Requirements

### Requirement: React Error Boundary component

The app MUST have a React Error Boundary component wrapping the navigation tree. On a render crash, it MUST display a fallback screen with a "Reintentar" button and log the error.

#### Scenario: Render crash shows fallback

- GIVEN a screen component that throws during render
- WHEN the Error Boundary catches the error
- THEN the boundary MUST display a fallback UI with the message "Algo salió mal" and a "Reintentar" button
- AND the original error and component stack MUST be logged

#### Scenario: Retry re-renders children

- GIVEN the Error Boundary fallback UI is displayed
- WHEN the user taps "Reintentar"
- THEN the boundary MUST reset its state and re-render the children
- AND the fallback UI MUST disappear
