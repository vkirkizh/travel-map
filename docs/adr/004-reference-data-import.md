# ADR 004: Reference Data Import

## Context

Places use country codes for travel statistics.
The MVP needs a reliable country catalog without manually maintaining SQL inserts.

## Decision

Use the [OurAirports `countries.csv` dataset](https://ourairports.com/data/) as country reference data.
The country dataset is imported through an idempotent Go CLI command:
- `cmd/import-countries`

The raw CSV file is stored locally in the `data` directory and ignored by Git.

## Consequences

Positive:
- Idempotent imports
- No large generated SQL file in the repository
- Clear separation between schema migrations, dev seed data, and reference data

Negative:
- Developers must download the CSV file locally before importing
- Data freshness depends on re-running imports with updated source files
