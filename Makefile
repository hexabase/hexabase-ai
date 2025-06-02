# Hexabase KaaS Development Makefile

.PHONY: help build test run clean docker-up docker-down docker-logs api-test ui-test

# Default target
help:
	@echo "Available targets:"
	@echo "  build       - Build API binaries"
	@echo "  test        - Run all tests"
	@echo "  api-test    - Run API tests only"
	@echo "  ui-test     - Run UI tests only"
	@echo "  run         - Run API server locally"
	@echo "  docker-up   - Start all services with docker-compose"
	@echo "  docker-down - Stop all services"
	@echo "  docker-logs - Show logs from all services"
	@echo "  clean       - Clean build artifacts"

# Build API binaries
build:
	@echo "Building API binaries..."
	cd api && go build -o bin/api ./cmd/api
	cd api && go build -o bin/worker ./cmd/worker

# Run all tests
test: api-test ui-test

# Run API tests
api-test:
	@echo "Running API tests..."
	cd api && go test -v ./...

# Run UI tests (placeholder for when UI is implemented)
ui-test:
	@echo "UI tests not implemented yet"

# Run API server locally
run:
	@echo "Starting API server..."
	cd api && go run ./cmd/api

# Start all services with docker compose
docker-up:
	@echo "Starting all services..."
	docker compose up -d
	@echo "Services started. API available at http://localhost:8080"
	@echo "Health check: curl http://localhost:8080/health"

# Stop all services
docker-down:
	@echo "Stopping all services..."
	docker compose down

# Show logs from all services
docker-logs:
	docker compose logs -f

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	cd api && rm -rf bin/
	docker compose down --volumes --remove-orphans
	docker system prune -f

# Initialize development environment
init:
	@echo "Initializing development environment..."
	@echo "Creating .env file..."
	@echo "# Development environment variables" > .env
	@echo "DATABASE_HOST=localhost" >> .env
	@echo "DATABASE_PORT=5433" >> .env
	@echo "DATABASE_USER=postgres" >> .env
	@echo "DATABASE_PASSWORD=postgres" >> .env
	@echo "DATABASE_DBNAME=hexabase" >> .env
	@echo "AUTH_JWT_SECRET=dev-jwt-secret-change-in-production" >> .env
	@echo "STRIPE_SECRET_KEY=sk_test_your_key_here" >> .env
	@echo "STRIPE_WEBHOOK_SECRET=whsec_your_webhook_secret" >> .env
	@echo "Development environment initialized!"
	@echo "Run 'make docker-up' to start all services"

# Quick development setup
dev: docker-up
	@echo "Development environment is ready!"
	@echo "API: http://localhost:8080"
	@echo "PostgreSQL: localhost:5433"
	@echo "Redis: localhost:6380"
	@echo "NATS: localhost:4223"