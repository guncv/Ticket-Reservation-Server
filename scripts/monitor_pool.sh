#!/bin/bash

# Pool Monitor Script
# Monitors database connection pool stats during load testing
#
# Usage:
#   ./scripts/monitor_pool.sh
#
# Environment variables:
#   BASE_URL - Server URL (default: http://localhost:8080)
#   INTERVAL - Polling interval in seconds (default: 2)

BASE_URL="${BASE_URL:-http://localhost:8080}"
INTERVAL="${INTERVAL:-2}"

echo "=== Pool Monitor ==="
echo "URL: $BASE_URL/api/v1/debug/pool-stats"
echo "Interval: ${INTERVAL}s"
echo "Press Ctrl+C to stop"
echo ""
echo "Legend:"
echo "  Total    = Total connections in pool"
echo "  Acquired = Currently in use"
echo "  Idle     = Available for use"
echo "  Empty    = Requests that had to wait (pool was full)"
echo "  Util%    = Pool utilization percentage"
echo ""
printf "%-10s | %-6s | %-8s | %-6s | %-8s | %-6s | %s\n" \
    "Time" "Total" "Acquired" "Idle" "Empty" "Util%" "Status"
echo "-----------|--------|----------|--------|----------|--------|------------------"

while true; do
    RESPONSE=$(curl -s "$BASE_URL/api/v1/debug/pool-stats" 2>/dev/null)

    if [ $? -ne 0 ] || [ -z "$RESPONSE" ]; then
        printf "%-10s | %s\n" "$(date +%H:%M:%S)" "ERROR: Cannot reach server"
    else
        TOTAL=$(echo "$RESPONSE" | grep -o '"total_conns":[0-9]*' | cut -d: -f2)
        ACQUIRED=$(echo "$RESPONSE" | grep -o '"acquired_conns":[0-9]*' | cut -d: -f2)
        IDLE=$(echo "$RESPONSE" | grep -o '"idle_conns":[0-9]*' | cut -d: -f2)
        EMPTY=$(echo "$RESPONSE" | grep -o '"empty_acquire_count":[0-9]*' | cut -d: -f2)
        UTIL=$(echo "$RESPONSE" | grep -o '"pool_utilization":[0-9.]*' | cut -d: -f2)
        STATUS=$(echo "$RESPONSE" | grep -o '"status":"[^"]*"' | cut -d'"' -f4)

        # Calculate percentage
        UTIL_PCT=$(echo "$UTIL * 100" | bc 2>/dev/null | cut -d. -f1)
        [ -z "$UTIL_PCT" ] && UTIL_PCT="0"

        printf "%-10s | %-6s | %-8s | %-6s | %-8s | %-5s%% | %s\n" \
            "$(date +%H:%M:%S)" "$TOTAL" "$ACQUIRED" "$IDLE" "$EMPTY" "$UTIL_PCT" "$STATUS"
    fi

    sleep "$INTERVAL"
done
