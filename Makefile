.PHONY: up down logs dev-env migrate-up migrate-down seed-dev backend-run backend-test backend-lint frontend-run frontend-lint frontend-build check

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

dev-env:
	docker compose up -d postgres

migrate-up:
	migrate -path backend/migrations -database "postgres://travel_map:travel_map@localhost:5432/travel_map?sslmode=disable" up

migrate-down:
	migrate -path backend/migrations -database "postgres://travel_map:travel_map@localhost:5432/travel_map?sslmode=disable" down

seed-dev:
	psql "postgres://travel_map:travel_map@localhost:5432/travel_map?sslmode=disable" -f backend/seeds/dev_seed.sql

backend-run:
	cd backend && go run ./cmd/api

backend-test:
	cd backend && go test ./...

backend-lint:
	cd backend && golangci-lint run

frontend-run:
	cd frontend && npm run dev

frontend-lint:
	cd frontend && npm run lint

frontend-build:
	cd frontend && npm run build

check: backend-test backend-lint frontend-lint frontend-build
