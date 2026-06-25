#!/bin/bash
# Wipes all Anton run data and restarts server clean

set -e
REPO="$(cd "$(dirname "$0")/.." && pwd)"
STATE="$REPO/.claude-team"

echo "Stopping server..."
lsof -ti:3000 | xargs kill -9 2>/dev/null || true
sleep 1

echo "Clearing database..."
rm -f "$STATE/state.db" "$STATE/state.db-shm" "$STATE/state.db-wal"

echo "Clearing runs..."
rm -rf "$STATE/runs"
mkdir -p "$STATE/runs"

echo "Clearing pending task..."
echo "" > "$STATE/pending-task.md"

echo "Restarting server..."
cd "$REPO"
go run main.go &>/tmp/anton.log &
sleep 3

if curl -s -o /dev/null -w "%{http_code}" http://localhost:3000 | grep -q "200"; then
  echo "Server live at http://localhost:3000"
else
  echo "ERROR: server failed to start. Check /tmp/anton.log"
  exit 1
fi
