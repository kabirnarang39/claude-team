# Anton Stable v1 — Ship Design

**Date:** 2026-06-20
**Status:** Approved
**Audience:** Open source, public release targeting Claude Code power users, general software engineers, and engineering leads

---

## Goal

Ship Anton as a stable v1 open source release. Anton must work reliably end-to-end — dispatch a task, agents run through all phases, results appear live in the dashboard — for users who have never seen the codebase.

---

## Phase 1: Validate

**Objective:** Discover every failure mode before any user does.

**What to run:**
- All 5 workflows: `feature-build`, `code-review`, `bug-fix`, `incident-response`, `architecture-review`
- Full path for each: skill dispatch → coordinator spawns agents → agents write to SQLite via MCP → WebSocket pushes → dashboard updates
- Error paths: missing API key, MCP not installed, malformed task, server not running

**Success criteria per workflow:**
- Agents transition PENDING → RUNNING → DONE without manual intervention
- Results land in SQLite with summary, confidence, token count populated
- Dashboard reflects live updates without requiring manual refresh
- Run history persists and loads correctly after server restart

**Output:** A failure log — every broken path with root cause and severity (blocking vs. polish). This log drives Phase 2 directly.

**Estimate:** 1–2 days.

---

## Phase 2: Fix

### Tier 1 — Blocking (nothing ships until resolved)

**Agent execution stuck at PENDING**
Root cause hypothesis: `team-coordinator.js` MCP tool not registered in sub-agent context, so results never write back to SQLite. Must verify MCP tool is available to every spawned agent and that the write path works.

**Phase transitions never fire**
The `when:` conditions in workflow YAML (e.g. `phases.planning.status == 'done'`) must be evaluated by the coordinator after each phase completes. This evaluator must be tested against real run data — not just unit tested against fixture YAML.

**Watcher never notifies**
The SQLite watcher polls for state changes and should push WebSocket events. If broken, the dashboard is a static snapshot. Must verify events fire on every agent status change.

**Fix order:** MCP wiring → phase transition evaluator → watcher. Sequential — each depends on the previous.

### Tier 2 — Quality (bad experience, does not block ship)

- **Errors swallowed silently** — 8+ bare `catch (_) {}` in UI, `log.Printf` in Go for recoverable errors. Both must surface meaningful messages to the user.
- **No graceful server-offline state** — dashboard shows blank when server unreachable. Needs an explicit "server offline" banner.
- **Run outputs not viewable in dashboard** — agents write markdown files to `.claude-team/runs/<id>/` but there is no way to read them from the UI.

---

## Phase 3: Polish

Four tracks run in parallel.

### Install Story

Single command installs everything:
```bash
curl -fsSL https://raw.githubusercontent.com/<org>/claude-team/main/install.sh | sh
```

(Domain `get.anton.dev` is a v1.1+ nice-to-have; install script lives at the GitHub raw URL for v1.)

Script responsibilities:
- Detect platform (macOS arm64/amd64, Linux amd64)
- Download pre-built binary from GitHub Release
- Install Claude Code skill (copy skills/ to Claude Code skills directory)
- Run `npm install` in `mcp/`
- Check prerequisites: Claude Code installed, `ANTHROPIC_API_KEY` set
- Fail loudly with a fix instruction on any step failure — no silent broken state

### Documentation

- `README.md` — what Anton is, 60-second quickstart, workflow reference table, architecture diagram
- `CHANGELOG.md` — versioned release notes starting at v1.0.0
- Per-workflow docs: example tasks, expected output artifacts, time estimate
- Troubleshooting section: top 5 failure modes with diagnosis steps

### UI

- **Server offline banner** — explicit message when API unreachable (currently: blank, no explanation)
- **Run output viewer** — clicking an agent shows its full markdown output from `.claude-team/runs/<id>/`
- **Empty state guidance** — first-time user sees actionable text ("Enter a task above and click Dispatch") not an empty SVG
- **Error banner** — surfaces errors currently swallowed in `catch (_) {}` blocks

### Distribution + CI

- GoReleaser configuration producing macOS arm64/amd64 and Linux amd64 binaries
- `anton --version` prints semver
- GitHub Actions:
  - On PR: `go test ./...`, `go build`, lint
  - On tag push: GoReleaser → GitHub Release with binaries attached

---

## Phase 4: Ship

Anton ships when every item in this checklist passes:

### Pre-release Gate

- [ ] All 5 workflows run end-to-end without manual intervention
- [ ] Fresh install via `install.sh` works on macOS arm64, macOS amd64, Linux amd64 — tested on a clean machine
- [ ] Dashboard shows live agent progress during an active run
- [ ] Run history persists across server restarts
- [ ] `anton --version` returns correct semver
- [ ] `go test ./...` passes
- [ ] README quickstart followed verbatim produces a working install

### Versioning

- First release: `v1.0.0`
- Semver strictly thereafter: breaking = major, new features = minor, fixes = patch
- `main` branch always in releasable state

### GitHub Release

- Tag `v1.0.0` triggers GoReleaser → binaries attached to GitHub Release
- Release notes = CHANGELOG entry for v1.0.0

---

## Out of Scope (v1.1+)

- Homebrew tap
- Demo video / GIF in README
- Example tasks library
- Multi-user / auth on dashboard
- Hosted / SaaS mode
