#!/bin/bash
# Dispatch to CPU/memory pprof or wall-clock (fgprof) scripts.
#
# Usage:
#   ./scripts/load_test_with_profile.sh cpu-memory
#   ./scripts/load_test_with_profile.sh wallclock
#
# Same env vars as the subscripts (BASE_URL, PPROF_URL, CONCURRENCY, …).

set -e
DIR="$(cd "$(dirname "$0")" && pwd)"
case "${1:-}" in
  cpu-memory|cpu|memory)
    shift || true
    exec "$DIR/load_test_with_profile_cpu_memory.sh" "$@"
    ;;
  wallclock|wall|fgprof)
    shift || true
    exec "$DIR/load_test_with_profile_wallclock.sh" "$@"
    ;;
  *)
    echo "Usage: $0 {cpu-memory|wallclock}  [extra args ignored by subshells]"
    echo ""
    echo "  cpu-memory   pprof CPU + heap/goroutine/allocs/block/mutex"
    echo "  wallclock    fgprof wall-clock (on- + off-CPU / I/O wait)"
    echo ""
    echo "Examples:"
    echo "  PROFILE_SECONDS=60 $0 cpu-memory"
    echo "  CONCURRENCY=500 REQUESTS=5000 $0 wallclock"
    exit 1
    ;;
esac
