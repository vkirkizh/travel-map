# Travel Map

A personal travel map for sharing visited places, flights and travel statistics.

## Stack

- Go
- PostgreSQL
- React
- TypeScript
- Leaflet
- Docker Compose

## Local development

Start database:
```bash
make dev-env
```

Run migrations:
```bash
make migrate-up
```

Seed demo data:
```bash
make seed-dev
```

Run backend:
```bash
make backend-run
```

Run frontend:
```bash
make frontend-run
```

Open:
```bash
http://localhost:5173/
http://localhost:5173/vkirkizh/
```

Health checks:
```bash
curl http://localhost:8080/healthz
curl http://localhost:8080/readyz
```

Development checks:
```bash
make backend-lint
make backend-test
```

---

Author: Valery Kirkizh (valery@kirkizh.com)
