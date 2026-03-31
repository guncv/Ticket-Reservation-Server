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
