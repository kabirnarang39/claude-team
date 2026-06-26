#!/usr/bin/env bash
# Usage: ./scripts/poll-stars.sh [interval_seconds]
# Default interval: 3600 (60 minutes)
# Appends timestamp,count rows to scripts/star-log.csv
# Requires: gh CLI authenticated (gh auth login)
set -euo pipefail

INTERVAL="${1:-3600}"

# Validate interval is numeric
if ! [[ "$INTERVAL" =~ ^[0-9]+$ ]]; then
  echo "ERROR: interval must be a positive integer (seconds), got: $INTERVAL"
  exit 1
fi

REPO="kabirnarang39/claude-team"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
LOG="$SCRIPT_DIR/star-log.csv"

# Write CSV header if file does not exist
if [ ! -f "$LOG" ]; then
  echo "timestamp,count" > "$LOG"
fi

# Check gh CLI is available
if ! command -v gh &>/dev/null; then
  echo "ERROR: gh CLI not found. Install from https://cli.github.com"
  exit 1
fi

# Check gh is authenticated
if ! gh auth status &>/dev/null; then
  echo "ERROR: gh CLI not authenticated. Run: gh auth login"
  exit 1
fi

echo "Polling $REPO every ${INTERVAL}s. Log: $LOG"
echo "Press Ctrl+C to stop."
echo ""

while true; do
  COUNT="$(timeout 15 gh api "repos/$REPO" --jq '.stargazers_count' 2>/dev/null)" || {
    TIMESTAMP="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
    echo "$TIMESTAMP  ERROR: gh api failed — check auth or network"
    sleep "$INTERVAL"
    continue
  }
  TIMESTAMP="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  echo "$TIMESTAMP,$COUNT" | tee -a "$LOG"
  sleep "$INTERVAL"
done
