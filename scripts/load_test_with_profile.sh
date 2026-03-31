#!/bin/bash

# Load Test with Profiling Script
# Automatically captures CPU, memory, goroutine profiles during load test
#
# Usage:
#   ./scripts/load_test_with_profile.sh
#
# Environment variables (same as load_test.sh):
#   BASE_URL      - Server URL (default: http://localhost:8080)
#   PPROF_URL     - pprof URL (default: http://localhost:6060)
#   CONCURRENCY   - Number of concurrent connections (default: 1000)
#   REQUESTS      - Total number of requests (default: 100000)
#   QUANTITY      - Tickets per reservation (default: 10)
#   TOTAL_TICKETS - Total tickets for the event (default: 1000000)
#   PROFILE_DIR   - Directory to save profiles (default: ./profiles)
#   OPEN_BROWSER  - Open pprof in browser after test (default: true)

set -e

BASE_URL="${BASE_URL:-http://localhost:8080}"
PPROF_URL="${PPROF_URL:-http://localhost:6060}"
TIMEOUT="${TIMEOUT:-30s}"
CONCURRENCY="${CONCURRENCY:-1000}"
REQUESTS="${REQUESTS:-100000}"
QUANTITY="${QUANTITY:-10}"
TOTAL_TICKETS="${TOTAL_TICKETS:-1000000}"
PROFILE_DIR="${PROFILE_DIR:-./profiles}"
OPEN_BROWSER="${OPEN_BROWSER:-true}"

TIMESTAMP=$(date +%Y%m%d_%H%M%S)
PROFILE_SUBDIR="$PROFILE_DIR/$TIMESTAMP"

mkdir -p "$PROFILE_SUBDIR"

echo "=== Load Test with Profiling ==="
echo "Base URL: $BASE_URL"
echo "pprof URL: $PPROF_URL"
echo "Concurrency: $CONCURRENCY"
echo "Total Requests: $REQUESTS"
echo "Profiles will be saved to: $PROFILE_SUBDIR"
echo ""

# Check if pprof is available (same process as make run — listens on :6060)
echo "Checking pprof availability..."
if ! curl -sf --max-time 5 "$PPROF_URL/debug/pprof/" > /dev/null 2>&1; then
    echo "ERROR: pprof not reachable at $PPROF_URL"
    echo ""
    echo "The API must be running before this script. From the repo root:"
    echo "  make run"
    echo "  # starts REST on your app port (default 8080) and pprof on 6060"
    echo ""
    echo "If the server is already running, check that nothing else is using 6060"
    echo "and that PPROF_URL matches (e.g. PPROF_URL=http://127.0.0.1:6060)."
    exit 1
fi
echo "pprof is available!"
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
  exit 1
fi
echo "Login successful!"
echo ""

# Step 2: Create event
echo "Step 2: Creating event with $TOTAL_TICKETS tickets..."
EVENT_TITLE="LoadTest_$TIMESTAMP"

CREATE_RESPONSE=$(curl -s -w "\n%{http_code}" --max-time 600 \
  -X POST "$BASE_URL/api/v1/event/" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Cookie: refresh_token=$REFRESH_TOKEN" \
  -d "{\"title\": \"$EVENT_TITLE\", \"description\": \"Load test event\", \"price\": 99.99, \"total_tickets\": $TOTAL_TICKETS}")

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
  exit 1
fi
echo "Event created: $EVENT_ID"
echo ""

# Step 3: Capture baseline profiles before load
echo "Step 3: Capturing baseline profiles..."
curl -s "$PPROF_URL/debug/pprof/heap" > "$PROFILE_SUBDIR/heap_before.prof"
curl -s "$PPROF_URL/debug/pprof/goroutine" > "$PROFILE_SUBDIR/goroutine_before.prof"
echo "Baseline captured"
echo ""

# Calculate profile duration (estimate based on requests and concurrency)
ESTIMATED_DURATION=$(( (REQUESTS / CONCURRENCY) + 30 ))
if [ "$ESTIMATED_DURATION" -lt 30 ]; then
    ESTIMATED_DURATION=30
fi
if [ "$ESTIMATED_DURATION" -gt 300 ]; then
    ESTIMATED_DURATION=300
fi

# Step 4: Start CPU profiling in background
echo "Step 4: Starting CPU profiling (${ESTIMATED_DURATION}s)..."
curl -s "$PPROF_URL/debug/pprof/profile?seconds=$ESTIMATED_DURATION" > "$PROFILE_SUBDIR/cpu.prof" &
CPU_PROFILE_PID=$!
sleep 2  # Give profiling time to start

# Step 5: Run load test
echo "Step 5: Running load test with bombardier..."
echo ""

LOAD_TEST_START=$(date +%s)

bombardier -c "$CONCURRENCY" -n "$REQUESTS" -t "$TIMEOUT" -l -m POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Cookie: refresh_token=$REFRESH_TOKEN" \
  -b "{\"event_id\": \"$EVENT_ID\", \"quantity\": $QUANTITY}" \
  "$BASE_URL/api/v1/event/ticket" 2>&1 | tee "$PROFILE_SUBDIR/bombardier_output.txt"

LOAD_TEST_END=$(date +%s)
LOAD_TEST_DURATION=$((LOAD_TEST_END - LOAD_TEST_START))

echo ""
echo "Load test completed in ${LOAD_TEST_DURATION}s"
echo ""

# Step 6: Capture post-load profiles
echo "Step 6: Capturing post-load profiles..."
curl -s "$PPROF_URL/debug/pprof/heap" > "$PROFILE_SUBDIR/heap_after.prof"
curl -s "$PPROF_URL/debug/pprof/goroutine" > "$PROFILE_SUBDIR/goroutine_after.prof"
curl -s "$PPROF_URL/debug/pprof/allocs" > "$PROFILE_SUBDIR/allocs.prof"
curl -s "$PPROF_URL/debug/pprof/block" > "$PROFILE_SUBDIR/block.prof"
curl -s "$PPROF_URL/debug/pprof/mutex" > "$PROFILE_SUBDIR/mutex.prof"
echo "Post-load profiles captured"
echo ""

# Wait for CPU profiling to complete
echo "Waiting for CPU profiling to complete..."
wait $CPU_PROFILE_PID 2>/dev/null || true
echo "CPU profiling complete"
echo ""

# Step 7: Generate summary
echo "=== Profile Summary ==="
echo "Profiles saved to: $PROFILE_SUBDIR"
echo ""
echo "Files:"
ls -lh "$PROFILE_SUBDIR"
echo ""

# Create analysis commands file
cat > "$PROFILE_SUBDIR/analyze.sh" << 'ANALYZE_EOF'
#!/bin/bash
PROFILE_DIR="$(dirname "$0")"

echo "Select profile to analyze:"
echo "1) CPU profile"
echo "2) Memory (heap) - after load"
echo "3) Goroutines - after load"
echo "4) Allocations"
echo "5) Block profile"
echo "6) Mutex profile"
echo "7) Compare heap before/after"
read -p "Choice [1-7]: " choice

case $choice in
    1) go tool pprof -http=:8081 "$PROFILE_DIR/cpu.prof" ;;
    2) go tool pprof -http=:8081 "$PROFILE_DIR/heap_after.prof" ;;
    3) go tool pprof -http=:8081 "$PROFILE_DIR/goroutine_after.prof" ;;
    4) go tool pprof -http=:8081 "$PROFILE_DIR/allocs.prof" ;;
    5) go tool pprof -http=:8081 "$PROFILE_DIR/block.prof" ;;
    6) go tool pprof -http=:8081 "$PROFILE_DIR/mutex.prof" ;;
    7) go tool pprof -http=:8081 -diff_base="$PROFILE_DIR/heap_before.prof" "$PROFILE_DIR/heap_after.prof" ;;
    *) echo "Invalid choice" ;;
esac
ANALYZE_EOF
chmod +x "$PROFILE_SUBDIR/analyze.sh"

echo "=== Quick Analysis Commands ==="
echo ""
echo "# CPU profile (what functions are slow)"
echo "go tool pprof -http=:8081 $PROFILE_SUBDIR/cpu.prof"
echo ""
echo "# Memory profile (where memory is allocated)"
echo "go tool pprof -http=:8081 $PROFILE_SUBDIR/heap_after.prof"
echo ""
echo "# Memory diff (what changed during load)"
echo "go tool pprof -http=:8081 -diff_base=$PROFILE_SUBDIR/heap_before.prof $PROFILE_SUBDIR/heap_after.prof"
echo ""
echo "# Goroutines (check for leaks)"
echo "go tool pprof -http=:8081 $PROFILE_SUBDIR/goroutine_after.prof"
echo ""
echo "# Or run interactive analyzer:"
echo "$PROFILE_SUBDIR/analyze.sh"
echo ""

# Cleanup
rm -f /tmp/cookies.txt

# Open browser if requested
if [ "$OPEN_BROWSER" = "true" ]; then
    echo "Opening CPU profile in browser..."
    go tool pprof -http=:8081 "$PROFILE_SUBDIR/cpu.prof" &
fi

echo "=== Load Test with Profiling Complete ==="
