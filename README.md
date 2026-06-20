# Anton

Multi-agent engineering team coordinator for Claude Code. Describe a task — Anton spins up a full team of specialist AI agents (planner, architect, engineers, QA, security reviewer) and shows live progress in your browser.

[![CI](https://github.com/kabirnarang39/claude-team/actions/workflows/ci.yml/badge.svg)](https://github.com/kabirnarang39/claude-team/actions/workflows/ci.yml)

## Quick Start

**Install (macOS / Linux):**
```bash
curl -fsSL https://raw.githubusercontent.com/kabirnarang39/claude-team/main/install.sh | sh
```

**Start the dashboard:**
```bash
anton
# Anton running at http://localhost:3000
```

**Dispatch a task** — open Claude Code in the same project directory, then run:
```
/team-dispatch build user authentication with JWT and refresh tokens
```

Or use the browser: go to `http://localhost:3000`, enter your task, select a workflow, click **Dispatch**, then paste the command shown into Claude Code.

## Workflows

| Workflow | What it does | Agents |
|----------|-------------|--------|
| `feature-build` | Full cycle: planning → architecture → engineering → QA → DevOps | 11 agents |
| `code-review` | Architecture + security + quality review in parallel | 3 agents |
| `bug-fix` | Root cause analysis → fix → verify | 3 agents |
| `incident-response` | Triage → hotfix → verify → post-mortem | 4 agents |
| `architecture-review` | Design doc review → ADR | 2 agents |

## Requirements

- [Claude Code](https://claude.ai/download) — active subscription or `ANTHROPIC_API_KEY`
- Node.js 20+
- Go 1.22+ (build from source only)

## Build From Source

```bash
git clone https://github.com/kabirnarang39/claude-team
cd claude-team
cd mcp && npm install && cd ..
go run main.go
```

## How It Works

```
You → /team-dispatch → Main Coordinator (your Claude Code session)
                              │
             ┌────────────────┼────────────────┐
             ▼                ▼                ▼
       Planning Sub-    Engineering      QA Sub-Coordinator
       Coordinator      Sub-Coordinator
             │                │                │
       requirements-   senior-architect   qa-engineer
       analyst         api-designer       security-reviewer
       tech-writer     backend-engineer   e2e-tester
                        frontend-engineer
                        dba
                              │
                        coordinator MCP     ← writes results
                              │
                        SQLite state.db     ← source of truth
                              │
                        Go HTTP server      ← reads + WebSocket
                              │
                        Browser dashboard   ← live agent tree
```

Agent outputs land in `.claude-team/runs/<run_id>/` and are viewable in the dashboard.

## Contributing

PRs welcome. Run `go test ./...` before opening a PR.

## License

MIT
