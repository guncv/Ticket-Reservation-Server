#!/bin/bash

# EXPLAIN ANALYZE Script
# Run SQL queries with EXPLAIN ANALYZE for performance debugging
#
# Usage:
#   ./scripts/explain_analyze.sh                              # Interactive menu
#   ./scripts/explain_analyze.sh tickets <event_id>            # SELECT SKIP LOCKED only
#   ./scripts/explain_analyze.sh reserve <event_id> [limit]    # Full UPDATE reserve (txn rollback)
#   ./scripts/explain_analyze.sh custom "SELECT..."           # Custom query

DB_CONTAINER="${DB_CONTAINER:-ticket-reservation-server-db}"
DB_NAME="${DB_NAME:-ticket-reservation-server}"
DB_USER="${DB_USER:-root}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

run_query() {
    local query="$1"
    local title="$2"

    echo ""
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
    echo -e "${YELLOW}$title${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
    echo ""
    echo -e "${GREEN}Query:${NC}"
    echo "$query" | sed 's/^/  /'
    echo ""
    echo -e "${GREEN}EXPLAIN ANALYZE Result:${NC}"
    echo ""

    docker exec -i "$DB_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" -c "
EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
$query
"

    echo ""
}

# EXPLAIN ANALYZE for UPDATE/DELETE/INSERT — wrapped in BEGIN/ROLLBACK so changes are undone.
run_mutating_explain() {
    local query="$1"
    local title="$2"

    echo ""
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
    echo -e "${YELLOW}$title${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
    echo ""
    echo -e "${YELLOW}Runs in BEGIN … ROLLBACK (no persisted changes).${NC}"
    echo ""
    echo -e "${GREEN}Query:${NC}"
    echo "$query" | sed 's/^/  /'
    echo ""
    echo -e "${GREEN}EXPLAIN ANALYZE Result:${NC}"
    echo ""

    docker exec -i "$DB_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" -c "
BEGIN;
EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
$query
ROLLBACK;
"

    echo ""
}

analyze_interpretation() {
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
    echo -e "${YELLOW}How to Read Results:${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
    echo ""
    echo -e "  ${GREEN}✅ Index Scan${NC}        = Using index (fast)"
    echo -e "  ${RED}❌ Seq Scan${NC}          = Full table scan (slow for big tables)"
    echo -e "  ${YELLOW}⚠️  Bitmap Heap Scan${NC} = Using index but fetching many rows"
    echo ""
    echo -e "  ${GREEN}actual time${NC}         = Real execution time (start..end) in ms"
    echo -e "  ${GREEN}rows${NC}                = Actual rows returned"
    echo -e "  ${GREEN}Buffers: shared hit${NC} = Pages read from cache (fast)"
    echo -e "  ${YELLOW}Buffers: shared read${NC}= Pages read from disk (slow)"
    echo ""
}

get_event_id() {
    # Only the final UUID may go to stdout — callers use: event_id="$(get_event_id)"
    echo -e "${YELLOW}Available events:${NC}" >&2
    docker exec -i "$DB_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" -c "
SELECT id, title, available_tickets, total_tickets
FROM events
ORDER BY created_at DESC
LIMIT 5;
" >&2
    echo "" >&2
    read -rp "Enter event_id (or press Enter for first one): " event_id

    event_id="${event_id//\"/}"
    event_id="${event_id//\'/}"

    if [ -z "$event_id" ]; then
        event_id=$(docker exec -i "$DB_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" -t -c "
SELECT id FROM events ORDER BY created_at DESC LIMIT 1;
" | tr -d ' \n')
    fi

    echo "$event_id"
}

# Predefined queries — inner SELECT used by ReserveTickets (see repo/ticket.go)
query_reserve_tickets_select() {
    local event_id="$1"
    local limit_n="${2:-10}"
    cat <<EOF
SELECT id FROM tickets
WHERE event_id = '$event_id' AND status = 'available'
LIMIT $limit_n
FOR UPDATE SKIP LOCKED
EOF
}

# Exact UPDATE shape from ReserveTickets after reservation INSERT (dummy reservation UUID for explain-only).
query_reserve_tickets_update() {
    local event_id="$1"
    local limit_n="$2"
    cat <<EOF
UPDATE tickets
SET reservation_id = '00000000-0000-0000-0000-000000000001'::uuid,
    status = 'sold',
    updated_at = NOW()
WHERE id IN (
    SELECT id FROM tickets
    WHERE event_id = '$event_id' AND status = 'available'
    LIMIT $limit_n
    FOR UPDATE SKIP LOCKED
)
RETURNING id;
EOF
}

query_reserve_tickets() {
    query_reserve_tickets_select "$1" "${2:-5}"
}

query_count_available() {
    local event_id="$1"
    cat <<EOF
SELECT COUNT(*) FROM tickets
WHERE event_id = '$event_id' AND status = 'available'
EOF
}

query_tickets_by_event() {
    local event_id="$1"
    cat <<EOF
SELECT id, status, reservation_id
FROM tickets
WHERE event_id = '$event_id'
LIMIT 100
EOF
}

query_reservations() {
    cat <<EOF
SELECT r.id, r.event_id, r.user_id, COUNT(t.id) as ticket_count
FROM reservations r
LEFT JOIN tickets t ON t.reservation_id = r.id
GROUP BY r.id
ORDER BY r.created_at DESC
LIMIT 20
EOF
}

query_events_with_stats() {
    cat <<EOF
SELECT
    e.id,
    e.title,
    e.total_tickets,
    e.available_tickets,
    COUNT(t.id) FILTER (WHERE t.status = 'available') as actual_available,
    COUNT(t.id) FILTER (WHERE t.status = 'sold') as actual_sold
FROM events e
LEFT JOIN tickets t ON t.event_id = e.id
GROUP BY e.id
ORDER BY e.created_at DESC
LIMIT 10
EOF
}

# Check table indexes
show_indexes() {
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
    echo -e "${YELLOW}Table Indexes${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
    echo ""

    docker exec -i "$DB_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" -c "
SELECT
    tablename,
    indexname,
    indexdef
FROM pg_indexes
WHERE schemaname = 'public'
ORDER BY tablename, indexname;
"
}

# Check table sizes
show_table_stats() {
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
    echo -e "${YELLOW}Table Statistics${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
    echo ""

    docker exec -i "$DB_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" -c "
SELECT
    relname as table_name,
    reltuples::bigint as row_count,
    pg_size_pretty(pg_total_relation_size(relid)) as total_size,
    pg_size_pretty(pg_relation_size(relid)) as table_size,
    pg_size_pretty(pg_indexes_size(relid)) as index_size
FROM pg_catalog.pg_statio_user_tables
ORDER BY pg_total_relation_size(relid) DESC;
"
}

# Run ANALYZE to update statistics
run_analyze() {
    echo -e "${YELLOW}Running ANALYZE on all tables...${NC}"
    docker exec -i "$DB_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" -c "ANALYZE;"
    echo -e "${GREEN}Done!${NC}"
}

# Interactive menu
show_menu() {
    echo ""
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
    echo -e "${YELLOW}   EXPLAIN ANALYZE - SQL Performance Debugger${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
    echo ""
    echo "  1) Reserve: inner SELECT only (SKIP LOCKED subquery)"
    echo "  1u) Reserve: full UPDATE (matches app ReserveTickets — rolled back)"
    echo "  2) Count Available Tickets"
    echo "  3) Get Tickets by Event"
    echo "  4) Get Reservations with Ticket Count"
    echo "  5) Events with Stats"
    echo "  6) Custom Query"
    echo ""
    echo "  i) Show Indexes"
    echo "  s) Show Table Stats"
    echo "  a) Run ANALYZE (update statistics)"
    echo "  h) How to Read Results"
    echo "  q) Quit"
    echo ""
    read -p "Select option: " choice

    case $choice in
        1)
            event_id=$(get_event_id)
            read -p "LIMIT rows [10]: " lim
            [ -z "$lim" ] && lim=10
            run_query "$(query_reserve_tickets_select "$event_id" "$lim")" "Reserve — SELECT (SKIP LOCKED)"
            ;;
        1u)
            event_id=$(get_event_id)
            read -p "LIMIT rows (quantity) [10]: " lim
            [ -z "$lim" ] && lim=10
            if ! [[ "$lim" =~ ^[0-9]+$ ]] || [ "$lim" -lt 1 ]; then
                echo -e "${RED}LIMIT must be a positive integer${NC}"
            else
                run_mutating_explain "$(query_reserve_tickets_update "$event_id" "$lim")" "ReserveTickets UPDATE (repo/ticket.go)"
            fi
            ;;
        2)
            event_id=$(get_event_id)
            run_query "$(query_count_available "$event_id")" "Count Available Tickets"
            ;;
        3)
            event_id=$(get_event_id)
            run_query "$(query_tickets_by_event "$event_id")" "Tickets by Event"
            ;;
        4)
            run_query "$(query_reservations)" "Reservations with Ticket Count"
            ;;
        5)
            run_query "$(query_events_with_stats)" "Events with Stats"
            ;;
        6)
            echo "Enter your SQL query (end with semicolon, press Enter twice to run):"
            query=""
            while IFS= read -r line; do
                [ -z "$line" ] && break
                query="$query $line"
            done
            run_query "$query" "Custom Query"
            ;;
        i)
            show_indexes
            ;;
        s)
            show_table_stats
            ;;
        a)
            run_analyze
            ;;
        h)
            analyze_interpretation
            ;;
        q)
            echo "Bye!"
            exit 0
            ;;
        *)
            echo -e "${RED}Invalid option${NC}"
            ;;
    esac

    show_menu
}

# Command line mode
if [ $# -gt 0 ]; then
    case $1 in
        tickets)
            event_id="${2:-$(docker exec -i "$DB_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT id FROM events LIMIT 1;" | tr -d ' \n')}"
            lim="${3:-10}"
            run_query "$(query_reserve_tickets_select "$event_id" "$lim")" "Reserve — SELECT (SKIP LOCKED), LIMIT=$lim"
            analyze_interpretation
            ;;
        reserve)
            if [ -z "${2:-}" ]; then
                echo "Usage: $0 reserve <event_id> [limit]"
                exit 1
            fi
            event_id="$2"
            lim="${3:-10}"
            if ! [[ "$lim" =~ ^[0-9]+$ ]] || [ "$lim" -lt 1 ]; then
                echo "limit must be a positive integer"
                exit 1
            fi
            run_mutating_explain "$(query_reserve_tickets_update "$event_id" "$lim")" "ReserveTickets UPDATE (LIMIT=$lim, rolled back)"
            analyze_interpretation
            ;;
        custom)
            run_query "$2" "Custom Query"
            ;;
        indexes)
            show_indexes
            ;;
        stats)
            show_table_stats
            ;;
        analyze)
            run_analyze
            ;;
        *)
            echo "Usage:"
            echo "  $0                                   # Interactive menu"
            echo "  $0 tickets [event_id] [limit]         # INNER SELECT SKIP LOCKED only"
            echo "  $0 reserve <event_id> [limit]         # Full UPDATE ReserveTickets (+ ROLLBACK)"
            echo "  $0 custom \"SELECT ...\""
            echo "  $0 indexes"
            echo "  $0 stats"
            echo "  $0 analyze"
            ;;
    esac
else
    show_menu
fi
