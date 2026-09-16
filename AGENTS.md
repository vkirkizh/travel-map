# Travel Map

Travel Map is a small pet project web application.

## Goal

Keep the project simple, production-ready and maintainable.
Avoid scope creep and unnecessary abstractions.

## Architecture

- Go backend
- React/TypeScript frontend
- PostgreSQL
- REST API
- Docker Compose
- Modular monolith; do not introduce microservices

## Development

Backend:

    make backend-test
    make backend-lint

Frontend:

    make frontend-lint
    make frontend-build

## Rules

- Prefer minimal changes.
- Do not refactor unrelated code.
- Do not add dependencies unless necessary.
- Preserve existing architecture unless the task requires otherwise.
- Never commit secrets or credentials.
- Keep migrations and development seed data consistent with the schema.
- Run relevant tests and linters after changes.
