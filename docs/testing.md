# OAuth Integration Testing Guide

This document describes the OAuth integration testing system for Hexabase KaaS.

## Overview

The OAuth integration tests verify the complete authentication flow including:
- OAuth provider login initiation (Google & GitHub)
- State validation for CSRF protection
- JWT token generation and validation
- Protected endpoint access control
- OIDC discovery and JWKS endpoints

## Test Results

✅ **All OAuth Integration Tests Passing (12/12)**

### Test Coverage

| Test | Status | Description |
|------|--------|-------------|
| **TestGoogleOAuthLoginFlow** | ✅ PASS | Google OAuth URL generation and parameters |
| **TestGitHubOAuthLoginFlow** | ✅ PASS | GitHub OAuth URL generation and parameters |
| **TestOAuthStateValidation** | ✅ PASS | CSRF state validation in callbacks |
| **TestUnsupportedOAuthProvider** | ✅ PASS | Proper handling of unsupported providers |
| **TestJWTTokenGeneration** | ✅ PASS | JWT token creation for authenticated users |
| **TestAuthMiddleware** | ✅ PASS | Protected endpoint access control |
| **TestInvalidTokenHandling** | ✅ PASS | Invalid/expired token rejection |
| **TestOAuthProviderConfiguration** | ✅ PASS | OAuth provider config validation |
| **TestOIDCDiscoveryEndpoint** | ✅ PASS | OIDC discovery metadata |
| **TestJWKSEndpoint** | ✅ PASS | JSON Web Key Set for token validation |
| **TestLogoutEndpoint** | ✅ PASS | User logout functionality |
| **TestCompleteOAuthFlowSimulation** | ✅ PASS | End-to-end OAuth flow simulation |

## Key Validations Confirmed

### 1. OAuth Provider Support
- **Google OAuth**: ✅ Proper URL generation with correct scopes
- **GitHub OAuth**: ✅ Proper URL generation with correct scopes
- **State Management**: ✅ CSRF protection via state validation
- **Error Handling**: ✅ Graceful handling of unsupported providers

### 2. Security Features
- **JWT Validation**: ✅ Proper token validation in middleware
- **Authorization Required**: ✅ Protected endpoints require valid tokens
- **State Validation**: ✅ OAuth callbacks validate CSRF state
- **Token Expiration**: ✅ Expired tokens are properly rejected

### 3. OIDC Compliance
- **Discovery Endpoint**: ✅ Proper OIDC discovery metadata
- **JWKS Endpoint**: ✅ RSA public keys for token validation
- **Standard Claims**: ✅ JWT tokens include required claims

### 4. API Integration
- **Auth Middleware**: ✅ Protects Organizations API endpoints
- **Error Responses**: ✅ Consistent error format across endpoints
- **CORS Support**: ✅ Proper CORS headers for browser integration

## Sample Test Output

```
=== RUN   TestOAuthIntegrationSuite
=== RUN   TestOAuthIntegrationSuite/TestGoogleOAuthLoginFlow
    oauth_integration_test.go:160: ✅ Google OAuth login URL generated successfully
    oauth_integration_test.go:161: State: J_D7IYBoaIxUSjYkJeyi0qstom9V2zjnsjZ1l6bYvm8=
    oauth_integration_test.go:162: Auth URL: https://accounts.google.com/o/oauth2/auth?...

=== RUN   TestOAuthIntegrationSuite/TestCompleteOAuthFlowSimulation
    oauth_integration_test.go:385: 🔄 Starting complete OAuth flow simulation...
    oauth_integration_test.go:400: ✅ Step 1: OAuth login initiated
    oauth_integration_test.go:418: ✅ Step 2: OAuth callback state validation passed

--- PASS: TestOAuthIntegrationSuite (0.40s)
    --- PASS: TestOAuthIntegrationSuite/TestAuthMiddleware (0.00s)
    --- PASS: TestOAuthIntegrationSuite/TestCompleteOAuthFlowSimulation (0.12s)
    [... all 12 tests passing ...]
PASS
ok  	github.com/hexabase/kaas-api/internal/api	0.644s
```

## Running OAuth Tests

### Prerequisites
1. Test database available at `localhost:5433`
2. Redis available at `localhost:6380`

### Execute Tests
```bash
cd api
go test ./internal/api -run TestOAuthIntegrationSuite -v
```

### Individual Test Cases
```bash
# Test specific OAuth provider
go test ./internal/api -run TestGoogleOAuthLoginFlow -v
go test ./internal/api -run TestGitHubOAuthLoginFlow -v

# Test security features  
go test ./internal/api -run TestAuthMiddleware -v
go test ./internal/api -run TestOAuthStateValidation -v

# Test OIDC compliance
go test ./internal/api -run TestOIDCDiscoveryEndpoint -v
go test ./internal/api -run TestJWKSEndpoint -v
```

## OAuth Flow Verification

The integration tests verify these OAuth flow steps:

1. **Login Initiation**: `/auth/login/{provider}` generates proper OAuth URLs
2. **State Generation**: Secure CSRF state tokens are created and stored
3. **URL Validation**: OAuth URLs contain correct client IDs, scopes, and parameters
4. **Callback Handling**: `/auth/callback/{provider}` validates state and processes codes
5. **Token Generation**: Valid JWT tokens are created for authenticated users
6. **Access Control**: Protected endpoints require valid authentication

## Next Steps

With OAuth integration tests complete, the authentication system is verified and ready for:

1. **Frontend Integration**: Build Next.js UI with OAuth login buttons
2. **End-to-End Testing**: Test complete user journeys with real browsers
3. **Production Setup**: Configure real OAuth app credentials for deployment

## Configuration Examples

### Google OAuth Configuration
```yaml
auth:
  external_providers:
    google:
      client_id: "your-google-client-id"
      client_secret: "your-google-client-secret"
      redirect_url: "https://your-domain.com/auth/callback/google"
      scopes: ["openid", "profile", "email"]
```

### GitHub OAuth Configuration
```yaml
auth:
  external_providers:
    github:
      client_id: "your-github-client-id" 
      client_secret: "your-github-client-secret"
      redirect_url: "https://your-domain.com/auth/callback/github"
      scopes: ["user:email"]
```

## OAuth Test Architecture

The OAuth integration tests use:
- **Test Database**: Isolated test data without affecting development DB
- **Mock Configuration**: Test OAuth credentials for safe testing
- **HTTP Test Server**: In-memory server for fast test execution
- **State Validation**: Real Redis integration for CSRF protection testing
- **JWT Verification**: Actual RSA key generation and validation

This comprehensive test suite ensures the OAuth system is production-ready and secure.# Testing Documentation

This directory contains guides and best practices for testing the Hexabase AI platform. 

**Important**: All test results and coverage reports have been moved to `/api/testresults/` to maintain a clean project structure.

## Directory Structure

```
testing/
├── README.md              # This file
├── testing-guide.md       # General testing guide and best practices
└── oauth-testing.md       # OAuth integration testing guide
```

## Test Result Locations

Test execution results are now centrally located:

- **Test Results**: `/api/testresults/`
- **Coverage Reports**: `/api/testresults/coverage/`
- **Test Summaries**: `/api/testresults/summary/`
- **Comprehensive Report**: `/api/testresults/COMPREHENSIVE_TEST_REPORT.md`
- **Historical Reports**: `/api/testresults/reports/`

## Test Organization

### API Tests (Go)
- **Unit Tests**: Located alongside source files (e.g., `auth/jwt_test.go`)
- **Integration Tests**: Located in `/api/tests/integration/`
- **E2E Tests**: Located in `/api/tests/e2e/`
- **Test Results**: All results in `/api/testresults/`

### UI Tests (TypeScript)
- **Playwright Tests**: Located in `/ui/tests/`
- **Unit Tests**: Located alongside components
- **Screenshots**: Located in `/ui/screenshots/`

### Test Scripts
- Located in `/scripts/test/`
- Includes database setup, token generation, and test runners

## Running Tests

### Go Tests
```bash
cd api
# Run all tests with coverage and reporting
./run_tests_with_coverage.sh

# Run all tests
go test ./...

# Run with coverage
go test ./... -cover

# Run specific package tests
go test -v ./internal/auth/...

# Run with race detection
go test -race ./...
```

### UI Tests
```bash
cd ui
# Run unit tests
npm test

# Run Playwright tests
npm run test:e2e

# Run specific test file
npm test -- workspace.spec.ts

# Run in headed mode for debugging
npm test -- --headed
```

## Testing Best Practices

### Unit Testing
- Write tests alongside implementation code
- Aim for >80% coverage on critical paths
- Use table-driven tests for multiple scenarios
- Mock external dependencies

### Integration Testing
- Test API endpoints with real HTTP requests
- Use test database for data persistence tests
- Verify error handling and edge cases
- Test authentication and authorization

### End-to-End Testing
- Use Playwright for UI testing
- Test complete user workflows
- Capture screenshots for visual validation
- Run against staging environment

## Coverage Requirements

- **Critical Paths**: >80% coverage required
- **Service Layer**: >70% coverage required
- **Handlers**: >60% coverage required
- **Utilities**: >50% coverage required

## Generating Coverage Reports

The test script automatically generates comprehensive reports:

```bash
cd api
./run_tests_with_coverage.sh
```

This will create:
- Coverage reports in `/api/testresults/coverage/[timestamp]/`
- Test summaries in `/api/testresults/summary/`
- HTML coverage visualization
- Package-level coverage details

For manual coverage generation:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## Additional Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Playwright Documentation](https://playwright.dev/)
- [Test Coverage in Go](https://blog.golang.org/cover)
- [Table Driven Tests](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)

---

For actual test results and coverage reports, see `/api/testresults/`.# Hexabase AI Local Testing Guide

This guide helps you test the API locally and inspect the database.

## Prerequisites

1. Ensure Docker services are running:
```bash
make docker-up
# or
docker-compose up -d
```

2. Check services are healthy:
```bash
docker ps
# Should show: postgres, redis, nats, api, worker
```

## Testing the API Locally

### 1. Health Check
```bash
# Test if API is running
curl http://localhost:8080/health
```

### 2. OAuth Login Flow

Since we're using OAuth, you need to simulate the login flow:

#### Option A: Get OAuth URL (for testing the flow)
```bash
# Get Google OAuth URL
curl http://localhost:8080/auth/login/google

# Get GitHub OAuth URL  
curl http://localhost:8080/auth/login/github
```

#### Option B: Create Test Token (for API testing)

For testing purposes, we provide scripts to create test data and generate JWT tokens:

```bash
# Step 1: Create test user in database
docker exec -i hexabase-ai-postgres-1 psql -U postgres -d hexabase < scripts/create_test_user.sql

# Step 2: Generate test JWT token
cd api
go run ../scripts/generate_test_token.go

# Or with custom parameters:
go run ../scripts/generate_test_token.go -user test-user-001 -email test@hexabase.local -org test-org-001

# To see token details:
go run ../scripts/generate_test_token.go -pretty
```

The test user created by the script:
- User ID: `test-user-001`
- Email: `test@hexabase.local`
- Organization: `test-org-001` (as admin)

### 3. Testing Organization APIs

Save the token from above as `TOKEN`:
```bash
export TOKEN="Bearer eyJhbGc..."  # Use the output from generate_test_token.go
```

#### Create Organization
```bash
curl -X POST http://localhost:8080/api/v1/organizations \
  -H "Authorization: $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Test Organization"
  }'
```

#### List Organizations
```bash
curl http://localhost:8080/api/v1/organizations \
  -H "Authorization: $TOKEN"
```

#### Get Organization
```bash
# Replace {org-id} with actual ID from create response
curl http://localhost:8080/api/v1/organizations/{org-id} \
  -H "Authorization: $TOKEN"
```

#### Update Organization
```bash
curl -X PUT http://localhost:8080/api/v1/organizations/{org-id} \
  -H "Authorization: $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Organization Name"
  }'
```

#### Delete Organization
```bash
curl -X DELETE http://localhost:8080/api/v1/organizations/{org-id} \
  -H "Authorization: $TOKEN"
```

## Database Inspection

### 1. Connect to PostgreSQL

```bash
# Connect to the database container
docker exec -it hexabase-ai-postgres-1 psql -U postgres -d hexabase

# Common psql commands:
\dt                    # List all tables
\d+ users             # Show users table structure
\d+ organizations     # Show organizations table structure
\d+ organization_users # Show organization_users table structure
```

### 2. Useful SQL Queries

```sql
-- List all users
SELECT id, email, display_name, provider, created_at FROM users;

-- List all organizations
SELECT id, name, created_at FROM organizations;

-- Show organization membership
SELECT 
    o.name as org_name,
    u.email as user_email,
    ou.role,
    ou.joined_at
FROM organization_users ou
JOIN organizations o ON o.id = ou.organization_id
JOIN users u ON u.id = ou.user_id;

-- Show all tables with row counts
SELECT 
    schemaname,
    tablename,
    n_live_tup as row_count
FROM pg_stat_user_tables
ORDER BY n_live_tup DESC;
```

### 3. Create Test Data

```sql
-- Create a test user
INSERT INTO users (id, external_id, provider, email, display_name)
VALUES (
    'test-user-123',
    'external-123', 
    'test',
    'test@example.com',
    'Test User'
);

-- Create a test organization
INSERT INTO organizations (id, name)
VALUES ('test-org-123', 'Test Organization');

-- Link user to organization as admin
INSERT INTO organization_users (organization_id, user_id, role)
VALUES ('test-org-123', 'test-user-123', 'admin');
```

### 4. Clean Test Data

```sql
-- Remove test data
DELETE FROM organization_users WHERE user_id = 'test-user-123';
DELETE FROM organizations WHERE id = 'test-org-123';
DELETE FROM users WHERE id = 'test-user-123';
```

## Debugging Tips

### 1. View API Logs
```bash
# View API logs
docker logs hexabase-ai-api-1 -f

# View all service logs
docker-compose logs -f
```

### 2. Check Redis State
```bash
# Connect to Redis
docker exec -it hexabase-ai-redis-1 redis-cli

# Commands:
KEYS *              # List all keys
GET oauth_state:*   # Get OAuth state
TTL oauth_state:*   # Check TTL
```

### 3. Test Database Connection
```bash
# From outside container
psql -h localhost -p 5433 -U postgres -d hexabase

# Password: postgres
```

### 4. API Debug Mode
The API runs in debug mode by default in development. Check logs for detailed information about:
- SQL queries being executed
- Request/response data
- Authentication details
- Error stack traces

## Common Issues

### Port Already in Use
```bash
# Check what's using the ports
lsof -i :5433  # PostgreSQL
lsof -i :6380  # Redis
lsof -i :4223  # NATS
lsof -i :8080  # API

# Stop conflicting services or change ports in docker-compose.yml
```

### Database Connection Failed
```bash
# Ensure PostgreSQL is running
docker ps | grep postgres

# Check PostgreSQL logs
docker logs hexabase-ai-postgres-1

# Test connection
docker exec hexabase-ai-postgres-1 pg_isready
```

### Authentication Issues
- Ensure the JWT token hasn't expired
- Check that the user exists in the database
- Verify the Authorization header format: `Bearer <token>`

## Testing Checklist

- [ ] API health check working
- [ ] Can generate/obtain valid JWT token
- [ ] Create organization successful
- [ ] List organizations shows created org
- [ ] Get specific organization works
- [ ] Update organization (as admin) works
- [ ] Delete organization (as admin) works
- [ ] Database shows correct data
- [ ] Redis shows OAuth states (if testing OAuth flow)
- [ ] Logs show expected behavior