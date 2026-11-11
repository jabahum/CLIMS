.PHONY: help build-backend run-backend build-frontend run-frontend install-frontend docker-up docker-down docker-build clean test

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
	@echo "    make docker-up       - Start the entire multi-container stack"
	@echo "    make docker-down     - Stop and remove Docker Compose containers"
	@echo "    make docker-build    - Rebuild and start Docker stack with version args"
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
	cd backend && go run ./cmd/server

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
	docker compose up -d

docker-down:
	@echo "Stopping Docker Compose environment..."
	docker compose down

docker-build:
	@echo "Building and starting Docker stack with version args..."
	VERSION=$(VERSION) COMMIT_HASH=$(COMMIT_HASH) BUILD_TIME=$(BUILD_TIME) docker compose up --build -d

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