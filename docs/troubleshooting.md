# Troubleshooting

## Agents stuck at PENDING

**Symptom:** Dashboard shows agents but they never transition from PENDING.
**Cause:** The browser saves the task but agents only run when you execute `/team-dispatch` in Claude Code.
**Fix:** Open Claude Code in the same directory as Anton, then run:
```
/team-dispatch --from-browser --workflow feature-build
```

## run_id shows as "unknown" in SQLite

**Symptom:** `sqlite3 .claude-team/state.db "SELECT run_id FROM agent_results;"` returns `unknown`.
**Cause:** Agent did not receive the Run ID in its brief, or did not include it in the report JSON.
**Fix:** Ensure the coordinator brief includes `Run ID: <exact-id>`. See `roles/_standards.md` for the required report format.

## Dashboard shows blank / no connection

**Symptom:** Opening `http://localhost:3000` shows the offline banner or blank page.
**Cause:** The Anton server is not running.
**Fix:** Start it with `go run main.go` or `anton` if installed.

## MCP coordinator not found

**Symptom:** Agents fail with "coordinator tool not available" or similar.
**Cause:** MCP npm dependencies not installed.
**Fix:**
```bash
cd mcp && npm install
```

## ANTHROPIC_API_KEY not set

**Symptom:** Agents immediately fail with authentication errors.
**Fix:**
```bash
export ANTHROPIC_API_KEY=sk-ant-...
```
Add to your shell profile (`~/.zshrc` or `~/.bashrc`) to persist.

## Workflow not found

**Symptom:** Error "workflow not found: feature-build".
**Cause:** Running from the wrong directory, or `workflows/` not present.
**Fix:** Run Anton from the root of the `claude-team` directory (where `workflows/` lives).
