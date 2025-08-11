# GOTRS Makefile - Docker/Podman compatible development

# Detect if we should use podman-compose or docker-compose
COMPOSE_CMD := $(shell command -v podman-compose 2> /dev/null || echo docker-compose)
CONTAINER_CMD := $(shell command -v podman 2> /dev/null || echo docker)

.PHONY: help up down logs restart clean setup test build

# Default target
help:
	@echo "GOTRS Container Development Commands (Docker/Podman):"
	@echo "  Using: $(COMPOSE_CMD)"
	@echo "  make up       - Start all services"
	@echo "  make down     - Stop all services"
	@echo "  make logs     - View logs"
	@echo "  make restart  - Restart all services"
	@echo "  make clean    - Clean everything (including volumes)"
	@echo "  make setup    - Initial project setup"
	@echo "  make test     - Run tests in containers"
	@echo "  make build    - Build production images"
	@echo ""
	@echo "Service-specific commands:"
	@echo "  make backend-logs  - View backend logs"
	@echo "  make frontend-logs - View frontend logs"
	@echo "  make db-shell      - PostgreSQL shell"
	@echo "  make valkey-cli    - Valkey CLI"
	@echo ""
	@echo "Database migration commands:"
	@echo "  make db-migrate    - Run all pending migrations"
	@echo "  make db-rollback   - Rollback last migration"
	@echo "  make db-reset      - Reset database (down all, up all)"
	@echo "  make db-status     - Show current migration version"
	@echo "  make db-force      - Force database to specific version"

# Initial setup
setup:
	@cp -n .env.example .env || true
	@cp -n docker-compose.override.yml.example docker-compose.override.yml || true
	@echo "Setup complete. Edit .env if needed."
	@echo "Run 'make up' to start development environment."

# Start all services
up:
	$(COMPOSE_CMD) up --build

# Start in background
up-d:
	$(COMPOSE_CMD) up -d --build

# Stop all services
down:
	$(COMPOSE_CMD) down

# Restart services
restart:
	$(COMPOSE_CMD) restart

# View logs
logs:
	$(COMPOSE_CMD) logs -f

# Service-specific logs
backend-logs:
	$(COMPOSE_CMD) logs -f backend

frontend-logs:
	$(COMPOSE_CMD) logs -f frontend

db-logs:
	$(COMPOSE_CMD) logs -f postgres

# Clean everything (including volumes)
clean:
	$(COMPOSE_CMD) down -v
	rm -rf tmp/

# Reset database
reset-db:
	$(COMPOSE_CMD) down -v postgres
	$(COMPOSE_CMD) up -d postgres
	@echo "Database reset. Waiting for initialization..."
	@sleep 5

# Load environment variables from .env file
include .env
export

# Database operations
db-shell:
	$(COMPOSE_CMD) exec postgres psql -U $(DB_USER) -d $(DB_NAME)

db-migrate:
	@echo "Running database migrations..."
	$(COMPOSE_CMD) exec -e PGPASSWORD=$(DB_PASSWORD) backend psql -h postgres -U $(DB_USER) -d $(DB_NAME) -f /app/migrations/000001_initial_schema.up.sql
	$(COMPOSE_CMD) exec -e PGPASSWORD=$(DB_PASSWORD) backend psql -h postgres -U $(DB_USER) -d $(DB_NAME) -f /app/migrations/000002_initial_data.up.sql
	@echo "Migrations completed successfully!"

db-migrate-schema-only:
	@echo "Running schema migration only..."
	$(COMPOSE_CMD) exec -e PGPASSWORD=$(DB_PASSWORD) backend psql -h postgres -U $(DB_USER) -d $(DB_NAME) -f /app/migrations/000001_initial_schema.up.sql

db-rollback:
	$(COMPOSE_CMD) exec backend migrate -path /app/migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@postgres:5432/$(DB_NAME)?sslmode=disable" down 1

db-reset:
	$(COMPOSE_CMD) exec backend migrate -path /app/migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@postgres:5432/$(DB_NAME)?sslmode=disable" down -all
	$(COMPOSE_CMD) exec backend migrate -path /app/migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@postgres:5432/$(DB_NAME)?sslmode=disable" up

db-status:
	$(COMPOSE_CMD) exec backend migrate -path /app/migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@postgres:5432/$(DB_NAME)?sslmode=disable" version

db-force:
	@read -p "Force migration to version: " version; \
	$(COMPOSE_CMD) exec backend migrate -path /app/migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@postgres:5432/$(DB_NAME)?sslmode=disable" force $$version

# Valkey CLI
valkey-cli:
	$(COMPOSE_CMD) exec valkey valkey-cli

# Run tests
test:
	$(COMPOSE_CMD) exec backend go test -v ./...

test-short:
	$(COMPOSE_CMD) exec backend go test -short ./...

# Build for production
build:
	$(CONTAINER_CMD) build -f Dockerfile -t gotrs:latest .
	$(CONTAINER_CMD) build -f Dockerfile.frontend -t gotrs-frontend:latest ./web

# Check service health
health:
	@echo "Checking service health..."
	@curl -f http://localhost/health || echo "Backend not healthy"
	@curl -f http://localhost/ || echo "Frontend not healthy"

# Open services in browser
open:
	@echo "Opening services..."
	@open http://localhost || xdg-open http://localhost || echo "Open http://localhost"

open-mail:
	@open http://localhost:8025 || xdg-open http://localhost:8025 || echo "Open http://localhost:8025"

open-db:
	@open http://localhost:8090 || xdg-open http://localhost:8090 || echo "Open http://localhost:8090"

# Development shortcuts
dev: up

stop: down

reset: clean setup up-d
	@echo "Environment reset and restarted"

# Show running services
ps:
	$(COMPOSE_CMD) ps

# Execute commands in containers
exec-backend:
	$(COMPOSE_CMD) exec backend sh

exec-frontend:
	$(COMPOSE_CMD) exec frontend sh

# Podman-specific: Generate systemd units
podman-systemd:
	@echo "Generating systemd units for podman..."
	podman generate systemd --new --files --name gotrs-postgres
	podman generate systemd --new --files --name gotrs-valkey
	podman generate systemd --new --files --name gotrs-backend
	podman generate systemd --new --files --name gotrs-frontend
	@echo "Systemd unit files generated. Move to ~/.config/systemd/user/"

# Generate migration file pair
gen-migration:
	@read -p "Migration name: " name; \
	timestamp=$$(date +%Y%m%d%H%M%S); \
	touch migrations/$$timestamp\_$$name.up.sql; \
	touch migrations/$$timestamp\_$$name.down.sql; \
	echo "-- Migration: $$name" > migrations/$$timestamp\_$$name.up.sql; \
	echo "" >> migrations/$$timestamp\_$$name.up.sql; \
	echo "-- Rollback: $$name" > migrations/$$timestamp\_$$name.down.sql; \
	echo "" >> migrations/$$timestamp\_$$name.down.sql; \
	echo "Created migration files:"; \
	echo "  migrations/$$timestamp\_$$name.up.sql"; \
	echo "  migrations/$$timestamp\_$$name.down.sql"