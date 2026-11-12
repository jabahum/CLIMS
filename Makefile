.PHONY: help build-backend run-backend build-frontend run-frontend install-frontend docker-up docker-down docker-build clean test

ENV_FILE ?= infra/env/dev.env
COMPOSE = docker compose --env-file $(ENV_FILE) -f infra/compose/docker-compose.yml

# Project variables
VERSION ?= 1.0.0
COMMIT_HASH ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

help:
	@echo "Available commands:"
	@echo "  Backend:"
	@echo "    make build-backend   - Build the backend binary with version ldflags"
	@echo "    make run-backend     - Run the backend server locally"
	@echo ""
	@echo "  Frontend:"
	@echo "    make install-frontend- Install frontend dependencies (npm install)"
	@echo "    make run-frontend    - Run the Next.js frontend development server"
	@echo "    make build-frontend  - Build the Next.js frontend for production"
	@echo ""
	@echo "  Docker Compose:"
	@echo "    make docker-up       - Start the PostgreSQL and API stack"
	@echo "    make docker-down     - Stop and remove Docker Compose containers"
	@echo "    make docker-build    - Rebuild and start the Docker stack"
	@echo ""
	@echo "    make setup           - Create the local environment file if missing"
	@echo "    make db-up           - Start PostgreSQL only"
	@echo "    make db-seed         - Import legacy data into an empty migrated database"
	@echo "    make sql-convert     - Regenerate PostgreSQL SQL from the MySQL dump"
	@echo ""
	@echo "  General:"
	@echo "    make clean           - Remove build binaries and artifacts"
	@echo "    make test            - Run backend tests"

# --- Backend Targets ---
build-backend:
	@echo "Building backend binary (Version: $(VERSION), Commit: $(COMMIT_HASH))..."
	cd backend && CGO_ENABLED=0 GOOS=linux go build \
		-ldflags="-s -w -X main.Version=$(VERSION) -X main.CommitHash=$(COMMIT_HASH) -X main.BuildTime=$(BUILD_TIME)" \
		-o bin/server ./cmd/server

run-backend:
	@echo "Starting backend server locally..."
	set -a; . ./$(ENV_FILE); set +a; cd backend && go run ./cmd/server

# --- Frontend Targets ---
install-frontend:
	@echo "Installing frontend dependencies..."
	cd frontend && npm install

run-frontend:
	@echo "Starting Next.js frontend dev server..."
	cd frontend && npm run dev

build-frontend:
	@echo "Building Next.js frontend production bundle..."
	cd frontend && npm run build

# --- Docker Compose Targets ---
docker-up:
	@echo "Starting Docker Compose environment..."
	$(COMPOSE) up -d --build --wait

docker-down:
	@echo "Stopping Docker Compose environment..."
	$(COMPOSE) down

docker-build:
	@echo "Building and starting Docker stack with version args..."
	VERSION=$(VERSION) COMMIT_HASH=$(COMMIT_HASH) BUILD_TIME=$(BUILD_TIME) $(COMPOSE) up --build -d --wait

# --- Utility Targets ---
clean:
	@echo "Cleaning build artifacts..."
	rm -rf backend/bin
	rm -f backend/server
	rm -rf frontend/.next
	rm -rf frontend/out

test:
	@echo "Running backend tests..."
	cd backend && go test -v ./...
.PHONY: setup db-up db-seed sql-convert
setup:
	@test -f $(ENV_FILE) || cp infra/env/dev.env.example $(ENV_FILE)
	@echo "Environment ready: $(ENV_FILE)"

db-up:
	$(COMPOSE) up -d --wait postgres

db-seed:
	$(COMPOSE) exec -T postgres sh -c 'psql -U "$$POSTGRES_USER" -d "$$POSTGRES_DB" -v ON_ERROR_STOP=1' < infra/database/clims_postgres_seed.sql

sql-convert:
	python3 scripts/convert_mysql_dump.py
