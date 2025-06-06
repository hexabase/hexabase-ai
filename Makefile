# Hexabase KaaS Makefile
# Common development tasks

.PHONY: help setup clean dev-api dev-ui test lint build deploy-dev deploy-staging deploy-prod

# Default target
help:
	@echo "Hexabase KaaS Development Commands"
	@echo "=================================="
	@echo "Setup & Cleanup:"
	@echo "  make setup          - Set up complete development environment"
	@echo "  make clean          - Clean up development environment"
	@echo ""
	@echo "Development:"
	@echo "  make dev-api        - Run API server in development mode"
	@echo "  make dev-ui         - Run UI development server"
	@echo "  make dev            - Run both API and UI (requires tmux)"
	@echo ""
	@echo "Testing:"
	@echo "  make test           - Run all tests"
	@echo "  make test-api       - Run API tests"
	@echo "  make test-ui        - Run UI tests"
	@echo "  make test-e2e       - Run end-to-end tests"
	@echo "  make coverage       - Generate test coverage report"
	@echo ""
	@echo "Code Quality:"
	@echo "  make lint           - Run linters"
	@echo "  make fmt            - Format code"
	@echo ""
	@echo "Building:"
	@echo "  make build          - Build all components"
	@echo "  make docker         - Build Docker images"
	@echo ""
	@echo "Deployment:"
	@echo "  make deploy-dev     - Deploy to local development"
	@echo "  make deploy-staging - Deploy to staging"
	@echo "  make deploy-prod    - Deploy to production"
	@echo ""
	@echo "Utilities:"
	@echo "  make status         - Show environment status"
	@echo "  make logs-api       - Tail API logs"
	@echo "  make db-shell       - Access database shell"

# Setup development environment
setup:
	@echo "Setting up development environment..."
	@./scripts/dev-setup.sh

# Clean up development environment
clean:
	@echo "Cleaning up development environment..."
	@./scripts/dev-cleanup.sh

# Run API server
dev-api:
	@echo "Starting API server..."
	@cd api && go run cmd/api/main.go

# Run UI development server
dev-ui:
	@echo "Starting UI development server..."
	@cd ui && npm install && npm run dev

# Run both API and UI using tmux
dev:
	@if command -v tmux >/dev/null 2>&1; then \
		tmux new-session -d -s hexabase-dev 'make dev-api' \; \
			split-window -h 'make dev-ui' \; \
			attach-session -t hexabase-dev; \
	else \
		echo "tmux not found. Please install tmux or run 'make dev-api' and 'make dev-ui' in separate terminals."; \
	fi

# Run all tests
test: test-api test-ui

# Run API tests
test-api:
	@echo "Running API tests..."
	@cd api && go test ./... -v

# Run UI tests
test-ui:
	@echo "Running UI tests..."
	@cd ui && npm test

# Run end-to-end tests
test-e2e:
	@echo "Running end-to-end tests..."
	@cd ui && npm run test:e2e

# Generate coverage report
coverage:
	@echo "Running API tests with coverage..."
	@cd api && go test ./... -coverprofile=coverage.out
	@cd api && go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated at api/coverage.html"

# Run linters
lint:
	@echo "Running Go linter..."
	@cd api && golangci-lint run
	@echo "Running UI linters..."
	@cd ui && npm run lint

# Format code
fmt:
	@echo "Formatting Go code..."
	@cd api && go fmt ./...
	@echo "Formatting UI code..."
	@cd ui && npm run format

# Build all components
build:
	@echo "Building API binary..."
	@cd api && go build -o bin/hexabase-api cmd/api/main.go
	@echo "Building UI for production..."
	@cd ui && npm run build

# Build Docker images
docker:
	@echo "Building Docker images..."
	@docker build -t hexabase/hexabase-kaas-api:latest -f api/Dockerfile api/
	@docker build -t hexabase/hexabase-kaas-ui:latest -f ui/Dockerfile ui/

# Deploy to development
deploy-dev:
	@echo "Deploying to development..."
	@helm upgrade --install hexabase-kaas ./deployments/helm/hexabase-kaas \
		--namespace hexabase-dev \
		--create-namespace \
		--values deployments/helm/values-local.yaml \
		--wait

# Deploy to staging
deploy-staging:
	@echo "Deploying to staging..."
	@helm upgrade --install hexabase-kaas ./deployments/helm/hexabase-kaas \
		--namespace hexabase-staging \
		--create-namespace \
		--values deployments/helm/values-staging.yaml \
		--wait

# Deploy to production
deploy-prod:
	@echo "⚠️  WARNING: You are about to deploy to PRODUCTION!"
	@echo -n "Are you sure? Type 'yes' to continue: "
	@read confirm && [ "$$confirm" = "yes" ] || (echo "Deployment cancelled"; exit 1)
	@echo "Deploying to production..."
	@helm upgrade --install hexabase-kaas ./deployments/helm/hexabase-kaas \
		--namespace hexabase-system \
		--create-namespace \
		--values deployments/helm/values-production.yaml \
		--wait

# Show environment status
status:
	@echo "Environment Status:"
	@echo "=================="
	@echo "Docker Compose:"
	@docker-compose ps
	@echo ""
	@echo "Kind Cluster:"
	@kind get clusters
	@echo ""
	@echo "Kubernetes Pods:"
	@kubectl get pods -A | grep hexabase || echo "No hexabase pods found"

# Tail logs
logs-api:
	@docker-compose logs -f postgres redis nats

# Database shell
db-shell:
	@docker-compose exec postgres psql -U hexabase -d hexabase_kaas

# Quick start (alias for setup)
start: setup

# Quick stop (alias for clean)
stop: clean