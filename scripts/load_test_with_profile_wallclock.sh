#!/bin/bash

# Load test + fgprof wall-clock profile (on- + off-CPU time, I/O wait, scheduling).
# Uses /debug/fgprof only — no standard pprof CPU profile.
#
# Usage:
#   ./scripts/load_test_with_profile_wallclock.sh
#
# Environment:
#   BASE_URL, PPROF_URL, CONCURRENCY, REQUESTS, QUANTITY, TOTAL_TICKETS, TIMEOUT, PROFILE_DIR
#   PROFILE_SECONDS - fgprof sample window (default: 30; should cover load test duration)

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
PROFILE_SUBDIR="$PROFILE_DIR/wallclock-$TIMESTAMP"

mkdir -p "$PROFILE_SUBDIR"

echo "=== Load test + fgprof (wall-clock) ==="
echo "Base URL: $BASE_URL"
echo "pprof URL: $PPROF_URL"
echo "Concurrency: $CONCURRENCY | Requests: $REQUESTS | fgprof: ${PROFILE_SECONDS}s"
echo "Output: $PROFILE_SUBDIR"
echo ""

echo "Checking pprof + fgprof endpoint..."
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
EVENT_TITLE="LoadTest_WC_${TIMESTAMP}"
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

echo "Step 3: Start fgprof (${PROFILE_SECONDS}s)..."
curl -s "$PPROF_URL/debug/fgprof?seconds=$PROFILE_SECONDS" > "$PROFILE_SUBDIR/fgprof.prof" &
FG_PID=$!
sleep 2

echo "Step 4: Bombardier..."
LOAD_TEST_START=$(date +%s)
bombardier -c "$CONCURRENCY" -n "$REQUESTS" -t "$TIMEOUT" -l -m POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Cookie: refresh_token=$REFRESH_TOKEN" \
  -b "{\"event_id\": \"$EVENT_ID\", \"quantity\": $QUANTITY}" \
  "$BASE_URL/api/v1/event/ticket" 2>&1 | tee "$PROFILE_SUBDIR/bombardier_output.txt"
LOAD_TEST_END=$(date +%s)
echo "Load finished in $((LOAD_TEST_END - LOAD_TEST_START))s"

echo "Waiting for fgprof..."
wait "$FG_PID" 2>/dev/null || true

cat > "$PROFILE_SUBDIR/analyze.sh" << 'EOF'
#!/bin/bash
D="$(dirname "$0")"
go tool pprof -http=:8081 "$D/fgprof.prof"
EOF
chmod +x "$PROFILE_SUBDIR/analyze.sh"

echo ""
echo "=== Done — $PROFILE_SUBDIR ==="
ls -lh "$PROFILE_SUBDIR"
echo ""
echo "go tool pprof -http=:8081 $PROFILE_SUBDIR/fgprof.prof"
rm -f /tmp/cookies.txt
