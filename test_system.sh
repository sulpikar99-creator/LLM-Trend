#!/bin/bash

echo "====================================="
echo "LLM Trend System Test"
echo "====================================="
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counters
PASSED=0
FAILED=0

# Function to test endpoint
test_endpoint() {
    local name=$1
    local url=$2
    local expected_status=$3
    local method=${4:-GET}
    local data=${5:-}

    echo -n "Testing $name... "

    if [ -z "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X $method "$url")
    else
        response=$(curl -s -w "\n%{http_code}" -X $method -H "Content-Type: application/json" -d "$data" "$url")
    fi

    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n-1)

    if [ "$http_code" -eq "$expected_status" ]; then
        echo -e "${GREEN}✓ PASS${NC} (HTTP $http_code)"
        PASSED=$((PASSED + 1))
        return 0
    else
        echo -e "${RED}✗ FAIL${NC} (Expected $expected_status, got $http_code)"
        echo "Response: $body"
        FAILED=$((FAILED + 1))
        return 1
    fi
}

echo "1. Testing Backend API"
echo "-------------------------------------"

# Health check
test_endpoint "Backend Health Check" "http://localhost:8080/api/health" 200

# Register new test user
TIMESTAMP=$(date +%s)
TEST_USER="testuser_$TIMESTAMP"
echo ""
echo "2. Testing User Registration"
echo "-------------------------------------"
test_endpoint "Register User" "http://localhost:3000/api/auth/register" 200 POST \
    "{\"username\":\"$TEST_USER\",\"email\":\"$TEST_USER@test.com\",\"password\":\"test123456\",\"beta_code\":\"BETA2024\"}"

# Login
echo ""
echo "3. Testing User Login"
echo "-------------------------------------"
LOGIN_RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" \
    -d "{\"username\":\"$TEST_USER\",\"password\":\"test123456\"}" \
    "http://localhost:3000/api/auth/login")

echo -n "Login Test... "
if echo "$LOGIN_RESPONSE" | grep -q "token"; then
    echo -e "${GREEN}✓ PASS${NC}"
    PASSED=$((PASSED + 1))
    TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    echo "Token obtained: ${TOKEN:0:20}..."
else
    echo -e "${RED}✗ FAIL${NC}"
    echo "Response: $LOGIN_RESPONSE"
    FAILED=$((FAILED + 1))
fi

# Test authenticated endpoints
if [ ! -z "$TOKEN" ]; then
    echo ""
    echo "4. Testing Authenticated Endpoints"
    echo "-------------------------------------"

    # Get traders
    echo -n "Get Traders List... "
    TRADERS_RESPONSE=$(curl -s -H "Authorization: Bearer $TOKEN" "http://localhost:3000/api/traders")
    if echo "$TRADERS_RESPONSE" | grep -q "data"; then
        echo -e "${GREEN}✓ PASS${NC}"
        PASSED=$((PASSED + 1))
    else
        echo -e "${RED}✗ FAIL${NC}"
        echo "Response: $TRADERS_RESPONSE"
        FAILED=$((FAILED + 1))
    fi

    # Get config
    echo -n "Get Config... "
    CONFIG_RESPONSE=$(curl -s -H "Authorization: Bearer $TOKEN" "http://localhost:3000/api/config")
    if echo "$CONFIG_RESPONSE" | grep -q "data"; then
        echo -e "${GREEN}✓ PASS${NC}"
        PASSED=$((PASSED + 1))
    else
        echo -e "${RED}✗ FAIL${NC}"
        echo "Response: $CONFIG_RESPONSE"
        FAILED=$((FAILED + 1))
    fi

    # Get analytics
    echo -n "Get Performance Analytics... "
    PERF_RESPONSE=$(curl -s -H "Authorization: Bearer $TOKEN" "http://localhost:3000/api/analytics/performance")
    if echo "$PERF_RESPONSE" | grep -q "data"; then
        echo -e "${GREEN}✓ PASS${NC}"
        PASSED=$((PASSED + 1))
    else
        echo -e "${RED}✗ FAIL${NC}"
        echo "Response: $PERF_RESPONSE"
        FAILED=$((FAILED + 1))
    fi
fi

echo ""
echo "5. Testing Frontend"
echo "-------------------------------------"

# Test frontend pages
test_endpoint "Frontend Homepage" "http://localhost:3000/" 200
test_endpoint "Frontend Login Page" "http://localhost:3000/login" 200
test_endpoint "Frontend Assets (JS)" "http://localhost:3000/assets/index-D5rnY638.js" 200
test_endpoint "Frontend Assets (CSS)" "http://localhost:3000/assets/index-49tiw3XF.css" 200

echo ""
echo "6. Testing WebSocket Endpoints"
echo "-------------------------------------"

if [ ! -z "$TOKEN" ]; then
    # WebSocket stats
    echo -n "WebSocket Stats... "
    WS_STATS=$(curl -s -H "Authorization: Bearer $TOKEN" "http://localhost:3000/api/ws/stats")
    if echo "$WS_STATS" | grep -q "total_clients"; then
        echo -e "${GREEN}✓ PASS${NC}"
        PASSED=$((PASSED + 1))
        echo "WebSocket Stats: $WS_STATS"
    else
        echo -e "${YELLOW}⚠ WARNING${NC} (WebSocket may not be ready)"
        echo "Response: $WS_STATS"
    fi
fi

echo ""
echo "====================================="
echo "Test Summary"
echo "====================================="
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ ALL TESTS PASSED!${NC}"
    echo ""
    echo "System is working correctly!"
    echo "Frontend: http://localhost:3000"
    echo "Backend API: http://localhost:8080"
    echo ""
    echo "Test user created:"
    echo "Username: $TEST_USER"
    echo "Password: test123456"
    echo ""
    exit 0
else
    echo -e "${RED}✗ SOME TESTS FAILED!${NC}"
    echo ""
    echo "Please check the errors above."
    exit 1
fi
