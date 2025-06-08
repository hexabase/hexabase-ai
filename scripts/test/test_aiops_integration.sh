#!/bin/bash

# Test AI Ops Chat Integration
# This script tests the integration between the Go API and Python AI Ops service

set -e

echo "Testing AI Ops Chat Integration..."

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Test configuration
API_URL="http://localhost:8080"
AIOPS_URL="http://localhost:8000"
TOKEN="test-jwt-token"  # This should be a valid JWT token

# Function to print test results
print_result() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓ $2${NC}"
    else
        echo -e "${RED}✗ $2${NC}"
    fi
}

# Test 1: Check if AI Ops service is healthy
echo "1. Checking AI Ops service health..."
HEALTH_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" $AIOPS_URL/health)
if [ "$HEALTH_RESPONSE" = "200" ]; then
    print_result 0 "AI Ops service is healthy"
else
    print_result 1 "AI Ops service is not responding (HTTP $HEALTH_RESPONSE)"
    exit 1
fi

# Test 2: Test direct chat endpoint on AI Ops service
echo "2. Testing direct chat endpoint..."
CHAT_RESPONSE=$(curl -s -X POST $AIOPS_URL/v1/chat \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d '{
        "messages": [{"role": "user", "content": "Hello, test message"}],
        "model": "llama3.2"
    }' \
    -w "\n%{http_code}")

HTTP_CODE=$(echo "$CHAT_RESPONSE" | tail -n1)
RESPONSE_BODY=$(echo "$CHAT_RESPONSE" | head -n-1)

if [ "$HTTP_CODE" = "200" ]; then
    print_result 0 "Direct chat endpoint working (HTTP 200)"
    echo "   Response: $RESPONSE_BODY"
else
    print_result 1 "Direct chat endpoint failed (HTTP $HTTP_CODE)"
    echo "   Response: $RESPONSE_BODY"
fi

# Test 3: Test chat proxy through Go API
echo "3. Testing chat proxy through Go API..."
PROXY_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/ai/chat?workspace_id=test-workspace-123" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d '{
        "messages": [{"role": "user", "content": "Hello through proxy"}],
        "model": "llama3.2"
    }' \
    -w "\n%{http_code}")

HTTP_CODE=$(echo "$PROXY_RESPONSE" | tail -n1)
RESPONSE_BODY=$(echo "$PROXY_RESPONSE" | head -n-1)

if [ "$HTTP_CODE" = "200" ]; then
    print_result 0 "Chat proxy working (HTTP 200)"
    echo "   Response: $RESPONSE_BODY"
else
    print_result 1 "Chat proxy failed (HTTP $HTTP_CODE)"
    echo "   Response: $RESPONSE_BODY"
fi

echo -e "\nIntegration test complete!"