# Hexabase KaaS Local Testing Guide

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
docker exec -i hexabase-kaas-postgres-1 psql -U postgres -d hexabase < scripts/create_test_user.sql

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
docker exec -it hexabase-kaas-postgres-1 psql -U postgres -d hexabase

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
docker logs hexabase-kaas-api-1 -f

# View all service logs
docker-compose logs -f
```

### 2. Check Redis State
```bash
# Connect to Redis
docker exec -it hexabase-kaas-redis-1 redis-cli

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
docker logs hexabase-kaas-postgres-1

# Test connection
docker exec hexabase-kaas-postgres-1 pg_isready
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