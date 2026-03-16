#!/bin/bash

# Load Test Script for Event Service
# Uses seeded admin user: admin/password
#
# Flow:
#   1. Login (single request)
#   2. Create event (single request)
#   3. Load test reserve ticket with bombardier
#
# Environment variables:
#   BASE_URL     - Server URL (default: http://localhost:8080)
#   CONCURRENCY  - Number of concurrent connections (default: 125)
#   REQUESTS     - Total number of requests (default: 10000)
#   QUANTITY     - Tickets per reservation (default: 1)
#   TOTAL_TICKETS - Total tickets for the event (default: 1000000)
#   TIMEOUT      - Per-request timeout for bombardier (default: 30s)

set -e

BASE_URL="${BASE_URL:-http://localhost:8080}"
TIMEOUT="${TIMEOUT:-30s}"
CONCURRENCY="${CONCURRENCY:-1000}"
REQUESTS="${REQUESTS:-100000}"
QUANTITY="${QUANTITY:-10}"
TOTAL_TICKETS="${TOTAL_TICKETS:-1000000}"

echo "=== Event Service Load Test ==="
echo "Base URL: $BASE_URL"
echo "Concurrency: $CONCURRENCY"
echo "Total Requests: $REQUESTS"
echo "Tickets per request: $QUANTITY"
echo "Total tickets in event: $TOTAL_TICKETS"
echo ""

# Step 1: Login
echo "Step 1: Logging in as admin..."
LOGIN_RESPONSE=$(curl -s -c /tmp/cookies.txt -w "\n%{http_code}" \
  -X POST "$BASE_URL/api/v1/user/login" \
  -H "Content-Type: application/json" \
  -d '{"user_name": "admin", "password": "password"}')

HTTP_CODE=$(echo "$LOGIN_RESPONSE" | tail -n1)
BODY=$(echo "$LOGIN_RESPONSE" | sed '$d')

if [ "$HTTP_CODE" != "200" ]; then
  echo "Login failed with status $HTTP_CODE"
  echo "Response: $BODY"
  exit 1
fi

ACCESS_TOKEN=$(echo "$BODY" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
REFRESH_TOKEN=$(grep "refresh_token" /tmp/cookies.txt | awk '{print $NF}')

if [ -z "$ACCESS_TOKEN" ]; then
  echo "Failed to extract access token"
  echo "Response: $BODY"
  exit 1
fi

echo "Login successful!"
echo ""

# Step 2: Create event
echo "Step 2: Creating event with $TOTAL_TICKETS tickets..."
EVENT_TITLE="LoadTest_$(date +%s)"

CREATE_RESPONSE=$(curl -s -w "\n%{http_code}" --max-time 600 \
  -X POST "$BASE_URL/api/v1/event/" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Cookie: refresh_token=$REFRESH_TOKEN" \
  -d "{\"title\": \"$EVENT_TITLE\", \"description\": \"Load test event description\", \"price\": 99.99, \"total_tickets\": $TOTAL_TICKETS}")

HTTP_CODE=$(echo "$CREATE_RESPONSE" | tail -n1)
BODY=$(echo "$CREATE_RESPONSE" | sed '$d')

if [ "$HTTP_CODE" != "201" ]; then
  echo "Create event failed with status $HTTP_CODE"
  echo "Response: $BODY"
  exit 1
fi

EVENT_ID=$(echo "$BODY" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)

if [ -z "$EVENT_ID" ]; then
  echo "Failed to extract event ID"
  echo "HTTP Status: $HTTP_CODE"
  echo "Response: $BODY"
  exit 1
fi

echo "Event created: $EVENT_ID"
echo ""

# Step 3: Load test
echo "Step 3: Running load test with bombardier..."
echo ""

bombardier -c "$CONCURRENCY" -n "$REQUESTS" -t "$TIMEOUT" -l -m POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Cookie: refresh_token=$REFRESH_TOKEN" \
  -b "{\"event_id\": \"$EVENT_ID\", \"quantity\": $QUANTITY}" \
  "$BASE_URL/api/v1/event/ticket"

# Cleanup
rm -f /tmp/cookies.txt

echo ""
echo "=== Load Test Complete ==="
