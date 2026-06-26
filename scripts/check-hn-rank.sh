#!/usr/bin/env bash
# Usage: ./scripts/check-hn-rank.sh <HN_ITEM_ID>
# Fetches https://hacker-news.firebaseio.com/v0/topstories.json
# and reports the item's rank in positions 1-10, 11-20, 21-30, or "not in top 30".
# Requires: curl, python3
set -euo pipefail

ITEM_ID="${1:?Usage: $0 <HN_ITEM_ID>}"

# Validate item ID is numeric
if ! [[ "$ITEM_ID" =~ ^[0-9]+$ ]]; then
  echo "ERROR: HN item ID must be numeric (e.g. 40123456)"
  exit 1
fi

# Check curl is available
if ! command -v curl &>/dev/null; then
  echo "ERROR: curl not found"
  exit 1
fi

# Check python3 is available
if ! command -v python3 &>/dev/null; then
  echo "ERROR: python3 not found"
  exit 1
fi

# Fetch top stories list
TOP_JSON="$(curl -fsSL --max-time 10 'https://hacker-news.firebaseio.com/v0/topstories.json' 2>/dev/null)" || {
  echo "ERROR: Failed to fetch HN top stories — check internet connection"
  exit 1
}

# Determine rank using python3; pass JSON via environment variable to avoid stdin conflict
RANK="$(ITEM_ID="$ITEM_ID" TOP_JSON="$TOP_JSON" python3 - <<'EOF'
import os, json, sys
item_id = int(os.environ["ITEM_ID"])
try:
    ids = json.loads(os.environ["TOP_JSON"])
except Exception as e:
    print("ERROR: " + str(e))
    sys.exit(1)
if item_id in ids:
    rank = ids.index(item_id) + 1
    print(rank)
else:
    print("not_found")
EOF
)" || {
  echo "ERROR: Failed to parse HN response"
  exit 1
}

TIMESTAMP="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

if [ "$RANK" = "not_found" ]; then
  echo "$TIMESTAMP  item=$ITEM_ID  not in top 30"
else
  if [ "$RANK" -le 10 ]; then
    BUCKET="top 1-10"
  elif [ "$RANK" -le 20 ]; then
    BUCKET="top 11-20"
  elif [ "$RANK" -le 30 ]; then
    BUCKET="top 21-30"
  else
    BUCKET="not in top 30 (rank $RANK)"
  fi
  echo "$TIMESTAMP  item=$ITEM_ID  rank=#$RANK  bucket=$BUCKET"
fi
