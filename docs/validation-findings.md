# Validation Findings — 2026-06-20

## Summary

| Workflow | Result |
|----------|--------|
| feature-build (planning phase) | PARTIAL — output files produced, SQLite not updated |
| architecture, engineering, qa, devops | Not tested (blocked by planning issue) |

## Failures

### BLOCKING: Coordinator MCP `report` tool unavailable to sub-agents

**Symptom:** Agents write fallback JSON report files to `.claude-team/runs/<id>/report-<agent>.json` but SQLite `agent_results` table stays PENDING. Dashboard shows no progress.

**Observed:** After running planning phase (requirements-analyst + tech-writer):
- `acceptance-criteria.md`, `prd.md`, `unknowns.md` produced correctly
- `report-requirements-analyst.json`, `report-tech-writer.json` written as fallback
- `SELECT * FROM agent_results WHERE status != 'PENDING'` returns zero rows
- Dashboard shows all 12 agents at PENDING

**Root cause:** The coordinator MCP (`mcp/team-coordinator.js`) is configured in `.claude/settings.json` but uses `path.join(process.cwd(), '.claude-team', 'state.db')` for the DB path. When sub-agents are spawned via the Agent tool, their `process.cwd()` may differ from the project root, causing the MCP to fail on startup silently. Sub-agents see the tool as unavailable and write fallback files.

**Severity:** BLOCKING — dashboard is completely static without this.

**Fix:** Add `POST /api/ingest-result` HTTP endpoint to Go server. Coordinator reads fallback JSON files and POSTs them to localhost:3000. Reliable regardless of MCP availability in sub-agent context.

### POLISH: run_id in fallback JSON is correct

**Observed:** Fallback JSON files correctly contain `"run_id": "anton-1781940885-4f3d7a"`. The run_id propagation from brief → agent → report JSON works. The only issue is the MCP write path.

### POLISH: Watcher fires correctly

**Observed:** When tested by manually inserting a row into agent_results, the watcher detects it within 2s and broadcasts via WebSocket. Watcher itself is not broken.

## What Works

- Pre-population of agents with PENDING status ✓
- pending-task.md format (Run ID, Workflow, task text) ✓
- Run dir creation ✓
- Agent output file production (md files) ✓
- Fallback JSON report format including correct run_id ✓
- SQLite watcher → WebSocket events ✓
- Dashboard real-time updates (when rows change) ✓

## Action Items

1. Add `POST /api/ingest-result` endpoint — coordinator ingests fallback JSON after each agent
2. Update `coordinators/main.md` — add curl step after each agent dispatch
3. Update `roles/_standards.md` — make fallback JSON writing explicit and required
