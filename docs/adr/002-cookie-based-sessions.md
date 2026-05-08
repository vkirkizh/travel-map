# ADR 002: Cookie-Based Sessions

## Context

Travel Map is primarily a browser-based application. The app needs authentication for profile editing, places management and flight management.

## Decision

Use cookie-based sessions.
Session tokens are stored in HttpOnly cookies. The database stores only hashed session tokens.

## Consequences

Positive:
- Simple browser auth model
- Tokens are not exposed to JavaScript
- Server-side session invalidation is possible
- Good fit for a web-first MVP

Negative:
- Requires CSRF considerations for state-changing requests
- Requires correct cookie settings in production
