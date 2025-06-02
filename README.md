# Hexabase KaaS Platform

An open-source, multi-tenant Kubernetes as a Service platform built on K3s and vCluster.

## Architecture Overview

Hexabase KaaS provides a user-friendly abstraction layer over Kubernetes, enabling developers to deploy and manage applications without dealing with Kubernetes complexity directly.

### Core Concepts

- **Organization**: Billing and user management unit
- **Workspace**: Isolated Kubernetes environment (vCluster)
- **Project**: Namespace within a Workspace
- **Users & Groups**: Hierarchical permission system

## Project Structure

```
hexabase-kaas/
├── api/                    # Go API service
│   ├── cmd/               # Entry points (api, worker)
│   ├── internal/          # Internal packages
│   │   ├── api/          # HTTP handlers
│   │   ├── auth/         # Authentication & OIDC
│   │   ├── billing/      # Stripe integration
│   │   ├── config/       # Configuration management
│   │   ├── db/           # Database models & repos
│   │   ├── k8s/          # vCluster management
│   │   ├── messaging/    # NATS pub/sub
│   │   └── service/      # Business logic
│   ├── Dockerfile
│   └── config.yaml
├── ui/                    # Next.js frontend (planned)
├── deployments/           # IaC and deployment configs
├── docs/                  # Documentation
├── docker-compose.yml     # Development environment
└── Makefile              # Development commands
```

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.21+ (for local development)
- PostgreSQL (included in docker-compose)

### Development Setup

1. **Initialize development environment:**
   ```bash
   make init
   ```

2. **Start all services:**
   ```bash
   make docker-up
   ```

3. **Verify the setup:**
   ```bash
   curl http://localhost:8080/health
   # Expected: {"status":"ok","timestamp":"..."}
   ```

### Available Services

- **API Server**: http://localhost:8080
- **PostgreSQL**: localhost:5433 (user: postgres, password: postgres)
- **Redis**: localhost:6380
- **NATS**: localhost:4223

### Testing the API

**Quick Test:**
```bash
# Run automated test script
./scripts/quick_test.sh
```

For detailed testing instructions, see [TESTING_GUIDE.md](TESTING_GUIDE.md) which covers:
- How to obtain authentication tokens
- Testing all API endpoints with curl
- Inspecting the database
- Debugging common issues

### Quick Database Access

```bash
# Connect to PostgreSQL
docker exec -it hexabase-kaas-postgres-1 psql -U postgres -d hexabase

# View all tables
\dt

# Exit
\q
```

## Development Commands

```bash
# Start development environment
make dev

# Run tests
make test
make api-test

# Build binaries
make build

# View logs
make docker-logs

# Clean up
make clean
```

## API Endpoints

### Authentication
- `POST /auth/login/{provider}` - Initiate OAuth login
- `GET /auth/callback/{provider}` - OAuth callback
- `POST /auth/logout` - Logout
- `GET /auth/me` - Get current user

### Organizations (✅ Implemented)
- `POST /api/v1/organizations` - Create organization
- `GET /api/v1/organizations` - List organizations
- `GET /api/v1/organizations/{orgId}` - Get organization details
- `PUT /api/v1/organizations/{orgId}` - Update organization
- `DELETE /api/v1/organizations/{orgId}` - Delete organization

### Workspaces (🚧 Coming Soon)
- `POST /api/v1/organizations/{orgId}/workspaces` - Create workspace
- `GET /api/v1/organizations/{orgId}/workspaces` - List workspaces
- `GET /api/v1/organizations/{orgId}/workspaces/{wsId}/kubeconfig` - Get kubeconfig

### Projects (🚧 Coming Soon)
- `POST /api/v1/workspaces/{wsId}/projects` - Create project
- `GET /api/v1/workspaces/{wsId}/projects` - List projects

### OIDC Provider
- `GET /.well-known/openid-configuration` - OIDC discovery
- `GET /.well-known/jwks.json` - JSON Web Key Set

📖 **For testing these endpoints locally, see [TESTING_GUIDE.md](TESTING_GUIDE.md)**

## Current Implementation Status

✅ **Completed:**
- Project structure and build system
- Database models with GORM
- Basic API server with Gin
- **Complete OAuth/OIDC authentication system**
  - JWT token management with RSA-256 signing
  - OAuth2 integration (Google, GitHub)
  - Redis state validation for CSRF protection
  - JWKS endpoint for token verification
  - Comprehensive test suite (21+ tests)
- Docker containerization
- Development environment with docker-compose
- Configuration management
- Health check endpoints
- TDD test coverage across all auth components

🚧 **In Progress:**
- Database migrations and advanced testing

📋 **Planned:**
- vCluster orchestration
- Core API endpoints (Organizations, Workspaces, Projects)
- Stripe billing integration
- Next.js UI
- Kubernetes operator for vCluster management
- Helm charts for deployment

## Testing

The project follows Test-Driven Development (TDD) principles with comprehensive test coverage across all components.

### Prerequisites for Tests

1. **Start the development environment** (required for database integration tests):
   ```bash
   make docker-up
   ```

2. **Create test database** (required for database integration tests):
   ```bash
   # Connect to the PostgreSQL container and create test database
   docker exec -it hexabase-kaas-postgres-1 createdb -U postgres hexabase_test
   
   # Alternative: Connect directly using psql
   docker exec -it hexabase-kaas-postgres-1 psql -U postgres -c "CREATE DATABASE hexabase_test;"
   ```

### Running Tests

#### Method 1: Using Makefile (Recommended)
```bash
# Run all API tests
make api-test
```

#### Method 2: Using Go Test Commands

**Run all tests:**
```bash
cd api
go test ./... -v
```

**Run tests for specific packages:**
```bash
# Authentication tests
go test ./internal/auth/... -v

# API endpoint tests
go test ./internal/api/... -v

# Service layer tests
go test ./internal/service/... -v

# Database tests (requires test DB)
go test ./internal/db/... -v
```

**Run tests with coverage:**
```bash
go test ./... -v -cover
```

**Generate detailed coverage report:**
```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
open coverage.html  # Opens coverage report in browser
```

#### Method 3: Run Specific Test Files

```bash
# JWT and token management tests
go test ./internal/auth/jwt_test.go -v

# RSA key management tests
go test ./internal/auth/keys_test.go -v

# OAuth client tests
go test ./internal/auth/oauth_client_test.go -v

# Redis state validation tests
go test ./internal/auth/oauth_redis_test.go -v

# API handler tests
go test ./internal/api/auth_test.go -v
```

#### Method 4: Advanced Test Options

**Run tests in parallel:**
```bash
go test ./... -v -parallel 4
```

**Run tests with timeout:**
```bash
go test ./... -v -timeout 30s
```

**Run only specific test functions:**
```bash
go test ./internal/auth -run TestTokenManager -v
go test ./internal/api -run TestAuthHandler -v
```

**Quick test status check:**
```bash
go test ./... | grep -E "(PASS|FAIL|SKIP)"
```

### Test Coverage

Current test coverage includes:

- **Authentication System**: JWT generation/validation, OAuth flows, state management
- **API Endpoints**: All auth-related HTTP handlers with proper status codes
- **Database Models**: GORM model validation and relationships
- **Security Features**: CSRF protection, token validation, key management
- **Redis Integration**: State storage and validation with Redis
- **Error Handling**: Comprehensive error scenarios and edge cases

**Test Statistics:**
- 21+ test functions across authentication components
- 100% API endpoint coverage for auth routes
- Mock-based testing for external dependencies
- Database integration tests (auto-skip if DB unavailable)
- Redis integration tests with proper mocking

### Troubleshooting Tests

**Database connection issues:**
```bash
# Ensure PostgreSQL is running
docker ps | grep postgres

# Check if test database exists
docker exec -it hexabase-kaas-postgres-1 psql -U postgres -l | grep hexabase_test

# Recreate test database if needed
docker exec -it hexabase-kaas-postgres-1 dropdb -U postgres hexabase_test --if-exists
docker exec -it hexabase-kaas-postgres-1 createdb -U postgres hexabase_test
```

**Port conflicts:**
- PostgreSQL: localhost:5433 (not 5432)
- Redis: localhost:6380 (not 6379)
- NATS: localhost:4223 (not 4222)

**Test database credentials:**
- Host: localhost:5433
- User: postgres
- Password: postgres
- Database: hexabase_test

## Configuration

Configuration is managed via YAML files and environment variables:

- Development: `api/config.yaml`
- Production: Environment variables (see `docker-compose.yml`)

### Key Configuration Sections

- **Database**: PostgreSQL connection settings
- **Redis**: Cache configuration
- **NATS**: Message queue settings
- **Auth**: JWT and OAuth provider settings
- **Stripe**: Payment processing
- **K8s**: Kubernetes client configuration

## Contributing

1. Follow TDD principles - write tests first
2. Use conventional commit messages
3. Ensure all tests pass before submitting PR
4. Update documentation for new features

## Architecture Decisions

- **Go**: Backend API for performance and K8s ecosystem compatibility
- **PostgreSQL**: Primary database for ACID compliance
- **Redis**: Caching and session storage
- **NATS**: Asynchronous task processing
- **vCluster**: Tenant isolation with full Kubernetes API
- **GORM**: ORM for type-safe database operations
- **Gin**: HTTP framework for API development

## Next Steps

1. Implement comprehensive test coverage
2. Build OAuth/OIDC authentication system
3. Develop vCluster orchestration layer
4. Create Next.js frontend
5. Add Stripe billing integration
6. Deploy monitoring and observability stack

## License

Open Source - License TBD