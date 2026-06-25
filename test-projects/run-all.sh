#!/bin/bash
# Dispatches all 5 workflows sequentially via Claude Code CLI
# Each runs to completion before the next starts
# Monitor progress at http://localhost:3000

set -e
REPO="$(cd "$(dirname "$0")/.." && pwd)"
TASKS="$(dirname "$0")/tasks"

# Ensure server is up
if ! curl -s -o /dev/null -w "%{http_code}" http://localhost:3000 | grep -q "200"; then
  echo "Server not running. Run: bash test-projects/reset.sh"
  exit 1
fi

run_workflow() {
  local workflow="$1"
  local task_file="$2"
  local task
  task="$(cat "$task_file")"

  echo ""
  echo "════════════════════════════════════════"
  echo " STARTING: $workflow"
  echo "════════════════════════════════════════"

  cd "$REPO"
  claude --print "/team-dispatch --workflow $workflow $task"

  echo ""
  echo "✓ DONE: $workflow"
  echo ""
  sleep 3
}

echo "Anton — Sequential Workflow Test"
echo "Dashboard: http://localhost:3000"
echo ""

run_workflow "feature-build"       "$TASKS/feature-build.md"
run_workflow "bug-fix"             "$TASKS/bug-fix.md"
run_workflow "code-review"         "$TASKS/code-review.md"
run_workflow "architecture-review" "$TASKS/architecture-review.md"
run_workflow "incident-response"   "$TASKS/incident-response.md"

echo ""
echo "All 5 workflows complete."
echo "Results at: http://localhost:3000"
