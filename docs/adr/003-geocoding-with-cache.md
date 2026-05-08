# ADR 003: Geocoding with PostgreSQL Cache

## Context

Users add places by typing free-form text such as "Berlin Germany" or "Stonehenge".
The system needs to resolve this text into coordinates and country metadata.
The MVP should avoid paid geocoding providers and minimize external API calls.

## Decision

Use Nominatim as the initial geocoding provider.
Before calling the external provider, the backend checks the `geocoding_cache` table using a normalized query.
Successful geocoding responses are stored in PostgreSQL together with coordinates and country code.
Countries are upserted automatically based on geocoding results.

## Consequences

Positive:
- No paid provider required for MVP
- Works well with OpenStreetMap-based UI
- External calls are minimized
- Provider can be replaced later behind the geocoding service

Negative:
- Public Nominatim usage has strict limits
- Geocoding quality depends on OpenStreetMap data
- Production-scale usage would likely require a paid provider or self-hosted service
