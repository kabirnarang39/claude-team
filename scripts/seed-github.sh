#!/usr/bin/env bash
# seed-github.sh — Create GitHub issue labels and seed good-first-issues for Anton launch
# AC-026 (labels) + AC-027 (seeded issues)
# Run at T-7 before HN submission.
#
# Prerequisites: gh CLI installed and authenticated (gh auth login)
# Idempotent: --force on label creates; issues are created once (no duplicate guard — run only once)
#
# Usage: ./scripts/seed-github.sh

set -euo pipefail

REPO="kabirnarang39/claude-team"

# Safety check — confirm before creating issues
echo "Anton launch: seed GitHub labels and good-first-issues"
echo "Repo: https://github.com/$REPO"
echo ""
echo "This script will:"
echo "  1. Create/update 4 issue labels (idempotent)"
echo "  2. Create 3 seeded good-first-issues (NOT idempotent — run only once)"
echo ""
read -r -p "Proceed? [y/N] " CONFIRM
if [[ "$CONFIRM" != "y" && "$CONFIRM" != "Y" ]]; then
  echo "Aborted."
  exit 0
fi

echo ""
echo "--- Step 1: Create issue labels (--force makes this idempotent) ---"

gh label create "good first issue" \
  --repo "$REPO" \
  --color "7057ff" \
  --description "Good for newcomers" \
  --force
echo "DONE: label 'good first issue'"

gh label create "help wanted" \
  --repo "$REPO" \
  --color "008672" \
  --description "Extra attention is needed" \
  --force
echo "DONE: label 'help wanted'"

gh label create "bug" \
  --repo "$REPO" \
  --color "d73a4a" \
  --description "Something isn't working" \
  --force
echo "DONE: label 'bug'"

gh label create "enhancement" \
  --repo "$REPO" \
  --color "a2eeef" \
  --description "New feature or request" \
  --force
echo "DONE: label 'enhancement'"

echo ""
echo "--- Step 2: Create seeded good-first-issues (run once) ---"

# Issue 1: Linux arm64 binary support
gh issue create \
  --repo "$REPO" \
  --title "Add Linux arm64 binary to release artifacts" \
  --label "good first issue" \
  --label "enhancement" \
  --body "$(cat <<'EOF'
## Summary

The Anton `install.sh` currently supports darwin-arm64, darwin-amd64, and linux-amd64. Linux arm64 (AWS Graviton, Raspberry Pi 4/5, Oracle Ampere instances) is excluded.

## What needs to change

1. Add `linux-arm64` to the build matrix in the release workflow (`.github/workflows/release.yml` or equivalent).
   - Build target: `GOOS=linux GOARCH=arm64 go build -o anton-linux-arm64 .`
2. Update `install.sh` to detect `linux-arm64` via `uname -m` returning `aarch64` and map it to the correct binary name.
   - Current detection block starts around the `PLATFORM` variable assignment — add an `aarch64` arm to the case statement.
3. Add `linux-arm64` to the release artifact upload step.
4. Update the `README.md` platform support note to include Linux arm64.

## Acceptance criteria

- `curl install.sh | sh` completes without error on an Ubuntu 22.04 arm64 machine (Graviton or Raspberry Pi).
- `anton --version` prints a version string.
- CI builds the `linux-arm64` binary without warnings.

## Notes

- Go cross-compilation for arm64 requires no additional toolchain — `GOARCH=arm64` works out of the box.
- The MCP layer (`mcp/`) is Node.js — Node.js arm64 binaries are available for Ubuntu 22.04+ via the NodeSource setup script.
- If you need an arm64 test environment: Oracle Cloud Free Tier includes always-free Ampere (arm64) instances.

This is a good first issue — the change is isolated to the build/release pipeline and `install.sh`. No Go code changes needed.
EOF
)"
echo "DONE: issue 1 — Linux arm64 binary support"

# Issue 2: Workflow template for API-only projects
gh issue create \
  --repo "$REPO" \
  --title "Add workflow template: api-only (no frontend engineer)" \
  --label "good first issue" \
  --label "enhancement" \
  --body "$(cat <<'EOF'
## Summary

The default `feature-build` workflow runs all 12 agents including frontend-engineer and dba. For backend-only or API-only tasks (REST endpoint, CLI tool, library), these agents add noise and token spend. An `api-only` workflow template would improve the experience for this common case.

## What needs to change

Create a new file `workflows/api-only.yaml` with this structure:

```yaml
name: api-only
description: Backend API feature — planning, architecture, backend engineering, QA, DevOps. No frontend or DBA agents.

phases:
  - id: planning
    sequential:
      - requirements-analyst
      - tech-writer

  - id: architecture
    sequential:
      - senior-architect
      - api-designer

  - id: engineering
    sequential:
      - backend-engineer

  - id: review
    sequential:
      - qa-engineer
      - security-reviewer
      - code-reviewer

  - id: devops
    sequential:
      - devops-engineer
```

## Acceptance criteria

- `workflows/api-only.yaml` parses without error: `go run main.go` starts and the workflow is listed.
- `/team-dispatch --workflow api-only build a rate-limiter middleware` dispatches only the agents listed above.
- No frontend-engineer or dba agent is spawned.
- The workflow is documented in the README workflows table.

## Notes

- Workflow YAML format is documented in `README.md` under "Add a Workflow".
- Anton picks up new workflow files at startup without any Go code changes.
- You can test locally by adding the file, restarting `go run main.go`, and verifying the workflow appears in the server response.

This is a good first issue — it requires only creating a new YAML file and updating one table in README.md.
EOF
)"
echo "DONE: issue 2 — api-only workflow template"

# Issue 3: Progress bar / step counter for install.sh
gh issue create \
  --repo "$REPO" \
  --title "Add step counter to install.sh (e.g. [1/5] Downloading binary...)" \
  --label "good first issue" \
  --label "enhancement" \
  --body "$(cat <<'EOF'
## Summary

`install.sh` runs silently for several steps (platform detection, binary download, xattr strip, codesign, npm install). On a slow connection, the script appears to hang with no feedback. Adding a step counter makes progress visible and reduces abort-and-retry behavior.

## What needs to change

Replace silent steps in `install.sh` with a `[N/TOTAL] message` format. For example:

```
[1/5] Detecting platform (darwin-arm64)...
[2/5] Downloading anton v1.2.0...
[3/5] Stripping quarantine attribute (macOS)...
[4/5] Verifying binary...
[5/5] Installing MCP server (npm install)...

Anton installed successfully. Run: anton --check
```

Implementation notes:
- Add a `STEP` counter variable initialized to 1.
- Wrap each major step in a `log_step() { echo "[${STEP}/${TOTAL_STEPS}] $1"; STEP=$((STEP+1)); }` helper function.
- `TOTAL_STEPS` can be hardcoded to the number of steps (currently ~5 depending on platform).
- On macOS, xattr + codesign are one step. On Linux, skip that step (adjust `TOTAL_STEPS` accordingly or count it conditionally).

## Acceptance criteria

- Running `curl install.sh | sh` on macOS arm64 prints one `[N/TOTAL]` line per major step.
- No steps are silent for more than 5 seconds without a progress message.
- `TOTAL_STEPS` is accurate (counter does not show `[4/3]`).
- Existing error handling (`set -euo pipefail`, prereq checks) is preserved.

## Notes

- The script is at `install.sh` in the repo root.
- Test by running `bash install.sh` locally (does not require a clean machine for this change).
- Avoid color codes — keep output plain text for compatibility with pipes and CI logs.

This is a good first issue — the change is isolated to `install.sh` and requires only shell scripting.
EOF
)"
echo "DONE: issue 3 — install.sh progress bar"

echo ""
echo "All done. Labels created (idempotent). 3 good-first-issues created."
echo "Verify at: https://github.com/$REPO/issues?q=label%3A%22good+first+issue%22"
