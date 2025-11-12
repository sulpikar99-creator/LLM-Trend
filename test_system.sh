#!/bin/bash

echo "=========================================="
echo "LLM-TREND SYSTEM FUNCTIONALITY TEST"
echo "=========================================="
echo ""

# Test 1: Check .env file
echo "1. Checking .env configuration..."
if [ ! -f ".env" ]; then
    echo "✗ FAIL: .env file does NOT exist"
    echo "   FIX: cp .env.example .env"
    exit 1
else
    echo "✓ .env file exists"
fi

# Test 2: Check if containers running
echo ""
echo "2. Checking Docker containers..."
docker compose ps

# Test 3: Check backend logs for startup
echo ""
echo "3. Checking backend startup logs..."
docker compose logs backend | tail -20

# Test 4: Check if traders loaded
echo ""
echo "4. Checking traders in database..."
docker compose logs backend | grep "Trader loading complete"

# Test 5: Test API endpoint
echo ""
echo "5. Testing backend API..."
curl -s http://localhost:8080/api/health | jq .

# Test 6: Check for errors
echo ""
echo "6. Checking for errors in logs..."
docker compose logs backend 2>&1 | grep -i "error\|fatal" | tail -10

echo ""
echo "=========================================="
echo "If you see errors above, that's the problem!"
echo "=========================================="
