# ADR 001: Modular Monolith

## Context

Travel Map is an MVP with a Go backend, React frontend, PostgreSQL database and a small number of business domains: users, places, flights, geocoding and public map rendering.

## Decision

Use a modular monolith for the backend.
The project should be simple enough to build quickly.
The backend is a single Go service with internal packages organized by domain and responsibility.

## Consequences

Positive:
- Simple local development
- Simple deployment
- Easy refactoring
- Lower operational overhead
- Clear path to extracting services later if needed

Negative:
- Module boundaries are enforced by convention, not network boundaries
- The codebase needs discipline to avoid becoming a big ball of mud
