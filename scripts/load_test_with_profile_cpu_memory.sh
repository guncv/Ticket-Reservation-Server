#!/bin/bash

# Load test + pprof: CPU (/debug/pprof/profile) and memory-related profiles (heap,
# goroutine, allocs, block, mutex). No fgprof.
#
# Usage:
#   ./scripts/load_test_with_profile_cpu_memory.sh
#
# Environment (same knobs as load_test.sh):
#   BASE_URL          - API (default: http://localhost:8080)
#   PPROF_URL         - pprof (default: http://localhost:6060)
#   CONCURRENCY, REQUESTS, QUANTITY, TOTAL_TICKETS, TIMEOUT
#   PROFILE_DIR       - output root (default: ./profiles)
#   PROFILE_SECONDS   - CPU sample window in seconds (default: 30; set >= load duration)

set -e

BASE_URL="${BASE_URL:-http://localhost:8080}"
PPROF_URL="${PPROF_URL:-http://localhost:6060}"
TIMEOUT="${TIMEOUT:-30s}"
CONCURRENCY="${CONCURRENCY:-300}"
REQUESTS="${REQUESTS:-10000}"
QUANTITY="${QUANTITY:-5}"
TOTAL_TICKETS="${TOTAL_TICKETS:-100000}"
PROFILE_DIR="${PROFILE_DIR:-./profiles}"
PROFILE_SECONDS="${PROFILE_SECONDS:-30}"

TIMESTAMP=$(date +%Y%m%d_%H%M%S)
PROFILE_SUBDIR="$PROFILE_DIR/cpu-memory-$TIMESTAMP"

mkdir -p "$PROFILE_SUBDIR"

echo "=== Load test + pprof (CPU & memory) ==="
echo "Base URL: $BASE_URL"
echo "pprof URL: $PPROF_URL"
echo "Concurrency: $CONCURRENCY | Requests: $REQUESTS | CPU sample: ${PROFILE_SECONDS}s"
echo "Output: $PROFILE_SUBDIR"
echo ""

echo "Checking pprof..."
if ! curl -sf --max-time 5 "$PPROF_URL/debug/pprof/" > /dev/null 2>&1; then
  echo "ERROR: pprof not reachable at $PPROF_URL — run the API (make run) first."
  exit 1
fi

echo "Step 1: Login..."
LOGIN_RESPONSE=$(curl -s -c /tmp/cookies.txt -w "\n%{http_code}" \
  -X POST "$BASE_URL/api/v1/user/login" \
  -H "Content-Type: application/json" \
  -d '{"user_name": "admin", "password": "password"}')
HTTP_CODE=$(echo "$LOGIN_RESPONSE" | tail -n1)
BODY=$(echo "$LOGIN_RESPONSE" | sed '$d')
if [ "$HTTP_CODE" != "200" ]; then
  echo "Login failed: $HTTP_CODE — $BODY"
  exit 1
fi
ACCESS_TOKEN=$(echo "$BODY" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
REFRESH_TOKEN=$(grep "refresh_token" /tmp/cookies.txt | awk '{print $NF}')
[ -n "$ACCESS_TOKEN" ] || { echo "No access token"; exit 1; }

echo "Step 2: Create event ($TOTAL_TICKETS tickets)..."
EVENT_TITLE="LoadTest_CPU_${TIMESTAMP}"
CREATE_RESPONSE=$(curl -s -w "\n%{http_code}" --max-time 600 \
  -X POST "$BASE_URL/api/v1/event/" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Cookie: refresh_token=$REFRESH_TOKEN" \
  -d "{\"title\": \"$EVENT_TITLE\", \"description\": \"Load test event description\", \"price\": 99.99, \"total_tickets\": $TOTAL_TICKETS}")
HTTP_CODE=$(echo "$CREATE_RESPONSE" | tail -n1)
BODY=$(echo "$CREATE_RESPONSE" | sed '$d')
[ "$HTTP_CODE" = "201" ] || { echo "Create event failed: $HTTP_CODE $BODY"; exit 1; }
EVENT_ID=$(echo "$BODY" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
[ -n "$EVENT_ID" ] || { echo "No event id"; exit 1; }

echo "Step 3: Baseline heap / goroutine..."
curl -s "$PPROF_URL/debug/pprof/heap" > "$PROFILE_SUBDIR/heap_before.prof"
curl -s "$PPROF_URL/debug/pprof/goroutine" > "$PROFILE_SUBDIR/goroutine_before.prof"

echo "Step 4: Start CPU profile (${PROFILE_SECONDS}s)..."
curl -s "$PPROF_URL/debug/pprof/profile?seconds=$PROFILE_SECONDS" > "$PROFILE_SUBDIR/cpu.prof" &
CPU_PID=$!
sleep 2

echo "Step 5: Bombardier..."
LOAD_TEST_START=$(date +%s)
bombardier -c "$CONCURRENCY" -n "$REQUESTS" -t "$TIMEOUT" -l -m POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Cookie: refresh_token=$REFRESH_TOKEN" \
  -b "{\"event_id\": \"$EVENT_ID\", \"quantity\": $QUANTITY}" \
  "$BASE_URL/api/v1/event/ticket" 2>&1 | tee "$PROFILE_SUBDIR/bombardier_output.txt"
LOAD_TEST_END=$(date +%s)
echo "Load finished in $((LOAD_TEST_END - LOAD_TEST_START))s"

echo "Step 6: Post-load profiles..."
curl -s "$PPROF_URL/debug/pprof/heap" > "$PROFILE_SUBDIR/heap_after.prof"
curl -s "$PPROF_URL/debug/pprof/goroutine" > "$PROFILE_SUBDIR/goroutine_after.prof"
curl -s "$PPROF_URL/debug/pprof/allocs" > "$PROFILE_SUBDIR/allocs.prof"
curl -s "$PPROF_URL/debug/pprof/block" > "$PROFILE_SUBDIR/block.prof"
curl -s "$PPROF_URL/debug/pprof/mutex" > "$PROFILE_SUBDIR/mutex.prof"

echo "Waiting for CPU profile..."
wait "$CPU_PID" 2>/dev/null || true

cat > "$PROFILE_SUBDIR/analyze.sh" << 'EOF'
#!/bin/bash
D="$(dirname "$0")"
echo "1) CPU  2) heap after  3) goroutine after  4) allocs  5) block  6) mutex  7) heap diff"
read -p "Choice [1-7]: " c
case $c in
  1) go tool pprof -http=:8081 "$D/cpu.prof" ;;
  2) go tool pprof -http=:8081 "$D/heap_after.prof" ;;
  3) go tool pprof -http=:8081 "$D/goroutine_after.prof" ;;
  4) go tool pprof -http=:8081 "$D/allocs.prof" ;;
  5) go tool pprof -http=:8081 "$D/block.prof" ;;
  6) go tool pprof -http=:8081 "$D/mutex.prof" ;;
  7) go tool pprof -http=:8081 -diff_base="$D/heap_before.prof" "$D/heap_after.prof" ;;
  *) echo "invalid" ;;
esac
EOF
chmod +x "$PROFILE_SUBDIR/analyze.sh"

echo ""
echo "=== Done — $PROFILE_SUBDIR ==="
ls -lh "$PROFILE_SUBDIR"
echo ""
echo "go tool pprof -http=:8081 $PROFILE_SUBDIR/cpu.prof"
echo "go tool pprof -http=:8081 -diff_base=$PROFILE_SUBDIR/heap_before.prof $PROFILE_SUBDIR/heap_after.prof"
rm -f /tmp/cookies.txt
