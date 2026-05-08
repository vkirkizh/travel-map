# ADR 004: Reference Data Import

## Context

Flights require a reliable airport catalog with IATA codes and coordinates.
The MVP should support real airport routes without manually maintaining thousands of SQL inserts.

## Decision

Use OurAirports CSV datasets as reference data.
Countries and airports are imported through idempotent Go CLI commands:
- `cmd/import-countries`
- `cmd/import-airports`

Raw CSV files are stored locally in `data` directory and ignored by Git.

## Consequences

Positive:
- Real airport catalog for MVP
- Idempotent imports
- No large generated SQL files in the repository
- Clear separation between schema migrations, dev seed data, and reference data

Negative:
- Developers must download CSV files locally before importing
- Data freshness depends on re-running imports with updated source files
