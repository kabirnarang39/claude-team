# Anton — Claude Code Context

## What This Repo Is

Anton is a multi-agent engineering team coordinator. It runs specialist AI agents (planner, architect, engineers, QA, security reviewer, DevOps) inside a Claude Code session and shows live progress in a browser dashboard.

## Architecture

- **`main.go`** — HTTP server entry point, config wiring, MCP auto-registration
- **`internal/api/`** — HTTP handlers, WebSocket hub, server routing
- **`internal/store/`** — SQLite store (runs, phases, agent results)
- **`internal/workflow/`** — YAML workflow parser
- **`internal/mcp/`** — MCP registry loader
- **`mcp/team-coordinator.js`** — Stdio MCP server providing `report`, `ask`, `inbox`, `reply` tools
- **`workflows/`** — Workflow YAML files (one per use case)
- **`roles/`** — Agent system prompts in Markdown (one per specialist)
- **`skills/`** — Claude Code slash-command skills (`/team-dispatch`, `/team-status`, `/team-stop`)
- **`ui/`** — Vanilla JS dashboard (no build step — edit and reload)
- **`coordinators/`** — Main coordinator logic read by the team-dispatch skill

## Key Invariants

- `go test -race ./...` must pass before any commit
- `go vet ./...` must be clean
- Module path: `github.com/kabirnarang39/claude-team`
- Go version: 1.22
- SQLite WAL mode, `INSERT OR IGNORE` for dedup
- All HTTP handlers use `http.MaxBytesReader` before reading body
- Server binds to `127.0.0.1` only (localhost)
- Path traversal: reject filenames with `..` or `/`; verify `filepath.Abs` against allowed dirs
- Agent results deduplicated via `MAX(id)` per `(run_id, phase_id, agent)`

## Development

```bash
go run main.go           # server on :3000
go test -race ./...      # full test suite
go vet ./...             # static analysis
```

The UI is at `ui/` — plain HTML/CSS/JS, no build step. Edit and reload.

## Adding a Workflow

Drop a `.yaml` file in `workflows/`. See existing files for the format. Phases can be `sequential` or `parallel`.

## Adding a Role

Add a `.md` file in `roles/`. Follow `roles/_standards.md` for the output format. Reference the role name in a workflow YAML.
