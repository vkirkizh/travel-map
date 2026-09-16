# Travel Map

A personal travel map for visualizing visited places.

**Live demo:** https://map.kirkizh.com/valery/

![Travel Map](docs/screenshots/travel-map-screenshot.png)

## Features

-   Public travel maps with visited places and statistics
-   User registration and cookie-based authentication
-   Private dashboard and profile settings
-   Places management with Nominatim geocoding
-   PostgreSQL-backed geocoding cache
-   Automatic country detection from geocoding results
-   Gravatar-based profile avatars

## Architecture

Travel Map is a small full-stack web application with a Go backend, React frontend and PostgreSQL database.

The backend is structured as a modular monolith with explicit domain boundaries.
It exposes an HTTP API for authentication, profile management, places and public maps.
External geocoding is handled through Nominatim and cached locally to reduce repeated requests.

The production deployment runs with Docker Compose behind Caddy,
which serves the frontend, reverse-proxies API requests and manages HTTPS.

## Stack

**Backend:** Go, PostgreSQL\
**Frontend:** React, TypeScript, Leaflet\
**Infrastructure:** Docker, Docker Compose, Caddy, DigitalOcean,
Cloudflare\
**External services:** Nominatim, CARTO, Gravatar

## Local development

Start PostgreSQL:

```bash
make dev-env
```

Run migrations and seed development data:

```bash
make migrate-up
make seed-dev
```

Run the backend and frontend:

```bash
make backend-run
make frontend-run
```

The application will be available at:

```text
http://localhost:5173/
http://localhost:5173/valery/
http://localhost:5173/app/
```

Test user:

```text
Login: valery@kirkizh.com
Password: 123456
```

Run all development checks:

```bash
make check
```

## Author

Valery Kirkizh

<valery@kirkizh.com>
