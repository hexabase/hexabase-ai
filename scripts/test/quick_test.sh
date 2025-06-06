#!/bin/bash

# Quick test script for Hexabase KaaS API
# This script creates test data and shows how to test the API

set -e

echo "🚀 Hexabase KaaS Quick Test Script"
echo "=================================="
echo ""

# Check if services are running
echo "1. Checking services..."
if ! docker ps | grep -q hexabase-kaas-postgres-1; then
    echo "❌ PostgreSQL not running. Please run: make docker-up"
    exit 1
fi
echo "✅ Services are running"
echo ""

# Create test user
echo "2. Creating test user in database..."
docker exec -i hexabase-kaas-postgres-1 psql -U postgres -d hexabase << 'EOF' 2>/dev/null || true
-- Create test user if not exists
INSERT INTO users (id, external_id, provider, email, display_name, created_at, updated_at)
VALUES (
    'test-user-001',
    'external-test-001', 
    'test',
    'test@hexabase.local',
    'Test User',
    NOW(),
    NOW()
) ON CONFLICT (email) DO NOTHING;

-- Show user
SELECT id, email, display_name FROM users WHERE id = 'test-user-001';
EOF
echo "✅ Test user created/verified"
echo ""

# Generate token
echo "3. Generating test JWT token..."
cd api
TOKEN=$(go run ../scripts/generate_test_token.go 2>/dev/null | grep "Bearer" | head -1)
cd ..

if [ -z "$TOKEN" ]; then
    echo "❌ Failed to generate token"
    exit 1
fi

echo "✅ Token generated"
echo ""

# Test health endpoint
echo "4. Testing health endpoint..."
HEALTH=$(curl -s http://localhost:8080/health)
echo "Response: $HEALTH"
echo ""

# Test organizations endpoint
echo "5. Testing Organizations API..."
echo ""

echo "Creating organization..."
CREATE_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/organizations \
  -H "Authorization: $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Organization from Script"
  }')

echo "Response: $CREATE_RESPONSE"
ORG_ID=$(echo $CREATE_RESPONSE | grep -o '"id":"[^"]*' | cut -d'"' -f4)
echo ""

if [ ! -z "$ORG_ID" ]; then
    echo "✅ Organization created with ID: $ORG_ID"
    echo ""
    
    echo "Listing organizations..."
    curl -s http://localhost:8080/api/v1/organizations \
      -H "Authorization: $TOKEN" | jq '.' 2>/dev/null || \
    curl -s http://localhost:8080/api/v1/organizations \
      -H "Authorization: $TOKEN"
    echo ""
fi

echo ""
echo "6. Database check..."
docker exec -it hexabase-kaas-postgres-1 psql -U postgres -d hexabase -c "
SELECT o.id, o.name, u.email, ou.role 
FROM organizations o 
JOIN organization_users ou ON o.id = ou.organization_id 
JOIN users u ON u.id = ou.user_id 
WHERE u.id = 'test-user-001';"

echo ""
echo "✅ Test completed!"
echo ""
echo "To manually test with the token:"
echo "export TOKEN=\"$TOKEN\""
echo 'curl http://localhost:8080/api/v1/organizations -H "Authorization: $TOKEN"'
echo ""
echo "To connect to the database:"
echo "docker exec -it hexabase-kaas-postgres-1 psql -U postgres -d hexabase"