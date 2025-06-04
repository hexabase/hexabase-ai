#!/bin/bash

# Test OAuth Security Implementation
# This script tests the enhanced OAuth flow with PKCE, refresh tokens, and security features

set -e

API_URL="${API_URL:-http://localhost:8080}"
FRONTEND_URL="${FRONTEND_URL:-http://localhost:3000}"

echo "=== OAuth Security Integration Test ==="
echo "API URL: $API_URL"
echo "Frontend URL: $FRONTEND_URL"
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print test results
print_result() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓ $2${NC}"
    else
        echo -e "${RED}✗ $2${NC}"
    fi
}

# Test 1: Check OIDC Discovery endpoint
echo "1. Testing OIDC Discovery endpoint..."
DISCOVERY=$(curl -s "$API_URL/.well-known/openid-configuration")
if echo "$DISCOVERY" | grep -q "jwks_uri"; then
    print_result 0 "OIDC Discovery endpoint is working"
else
    print_result 1 "OIDC Discovery endpoint failed"
fi

# Test 2: Check JWKS endpoint
echo -e "\n2. Testing JWKS endpoint..."
JWKS=$(curl -s "$API_URL/.well-known/jwks.json")
if echo "$JWKS" | grep -q "keys"; then
    print_result 0 "JWKS endpoint is working"
else
    print_result 1 "JWKS endpoint failed"
fi

# Test 3: Test OAuth login with PKCE
echo -e "\n3. Testing OAuth login with PKCE..."

# Generate PKCE values
CODE_VERIFIER=$(openssl rand -base64 32 | tr -d "=+/" | cut -c 1-43)
CODE_CHALLENGE=$(echo -n "$CODE_VERIFIER" | openssl dgst -sha256 -binary | base64 | tr -d "=+/" | tr "+/" "-_")

# Initiate OAuth flow
AUTH_RESPONSE=$(curl -s -X POST "$API_URL/auth/login/google" \
  -H "Content-Type: application/json" \
  -d "{\"code_challenge\":\"$CODE_CHALLENGE\",\"code_challenge_method\":\"S256\"}")

if echo "$AUTH_RESPONSE" | grep -q "auth_url"; then
    print_result 0 "OAuth login initiated with PKCE"
    AUTH_URL=$(echo "$AUTH_RESPONSE" | jq -r '.auth_url')
    STATE=$(echo "$AUTH_RESPONSE" | jq -r '.state')
    echo "  Auth URL: $AUTH_URL"
    echo "  State: $STATE"
else
    print_result 1 "OAuth login failed"
fi

# Test 4: Test invalid refresh token
echo -e "\n4. Testing invalid refresh token..."
REFRESH_RESPONSE=$(curl -s -X POST "$API_URL/auth/refresh" \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"invalid_token"}' \
  -w "\n%{http_code}")

HTTP_CODE=$(echo "$REFRESH_RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "401" ]; then
    print_result 0 "Invalid refresh token correctly rejected"
else
    print_result 1 "Invalid refresh token not properly handled"
fi

# Test 5: Check rate limiting
echo -e "\n5. Testing rate limiting..."
RATE_LIMIT_HIT=0
for i in {1..20}; do
    RESPONSE=$(curl -s -X POST "$API_URL/auth/login/google" \
      -H "Content-Type: application/json" \
      -d '{}' \
      -w "\n%{http_code}")
    
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "429" ]; then
        RATE_LIMIT_HIT=1
        break
    fi
done

if [ $RATE_LIMIT_HIT -eq 1 ]; then
    print_result 0 "Rate limiting is working"
else
    print_result 1 "Rate limiting not detected"
fi

# Test 6: Check security headers
echo -e "\n6. Testing security headers..."
HEADERS=$(curl -s -I "$API_URL/auth/me")

check_header() {
    if echo "$HEADERS" | grep -qi "$1"; then
        print_result 0 "  $1 header present"
    else
        print_result 1 "  $1 header missing"
    fi
}

check_header "X-Content-Type-Options"
check_header "X-Frame-Options"
check_header "X-XSS-Protection"
check_header "Strict-Transport-Security"

# Test 7: Frontend integration test
echo -e "\n7. Testing frontend OAuth integration..."

# Check if frontend is running
FRONTEND_CHECK=$(curl -s -o /dev/null -w "%{http_code}" "$FRONTEND_URL")
if [ "$FRONTEND_CHECK" = "200" ]; then
    print_result 0 "Frontend is accessible"
    
    # Check if auth context is updated
    echo "  Checking for PKCE implementation in frontend..."
    if curl -s "$FRONTEND_URL" | grep -q "generateCodeVerifier"; then
        print_result 0 "  PKCE implementation found in frontend"
    else
        print_result 1 "  PKCE implementation not found in frontend"
    fi
else
    print_result 1 "Frontend is not accessible (HTTP $FRONTEND_CHECK)"
fi

# Test 8: Database integration
echo -e "\n8. Testing database integration..."
if docker ps | grep -q "postgres"; then
    print_result 0 "PostgreSQL is running"
    
    # Check if auth tables exist
    AUTH_TABLES=$(docker exec hexabase-postgres psql -U hexabase -d hexabase -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_name IN ('users', 'oauth_states', 'refresh_tokens', 'security_logs');")
    if [ "$AUTH_TABLES" -gt 0 ]; then
        print_result 0 "  Auth tables exist in database"
    else
        print_result 1 "  Auth tables missing in database"
    fi
else
    print_result 1 "PostgreSQL is not running"
fi

# Test 9: Redis integration
echo -e "\n9. Testing Redis integration..."
if docker ps | grep -q "redis"; then
    print_result 0 "Redis is running"
    
    # Test Redis connectivity
    if docker exec hexabase-redis redis-cli ping | grep -q "PONG"; then
        print_result 0 "  Redis is responsive"
    else
        print_result 1 "  Redis is not responsive"
    fi
else
    print_result 1 "Redis is not running"
fi

# Summary
echo -e "\n${YELLOW}=== Test Summary ===${NC}"
echo "OAuth security implementation test completed."
echo ""
echo "Next steps:"
echo "1. Run the API server with: cd api && go run cmd/api/main.go"
echo "2. Run the frontend with: cd ui && npm run dev"
echo "3. Test the complete OAuth flow manually:"
echo "   - Visit $FRONTEND_URL"
echo "   - Click 'Login with Google' or 'Login with GitHub'"
echo "   - Verify PKCE flow in browser developer tools"
echo "   - Check for secure token storage"
echo "   - Test token refresh"
echo "   - Verify session management"
echo ""
echo "To run comprehensive tests:"
echo "  cd api && go test ./... -v"
echo ""