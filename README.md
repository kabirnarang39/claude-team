# Anton

Multi-agent engineering team coordinator for Claude Code. Dispatch a task — Anton runs a full team of specialist AI agents (planner, architect, engineers, QA, security) and shows live progress in your browser.

## Quick Start

**Install:**
```bash
curl -fsSL https://raw.githubusercontent.com/<org>/claude-team/main/install.sh | sh
```

**Start the dashboard:**
```bash
anton
# → Anton running at http://localhost:3000
```

**Dispatch a task** (in Claude Code, in the same directory):
```
/team-dispatch build user authentication with JWT and refresh tokens
```

Or use the browser — go to `http://localhost:3000`, enter your task, click Dispatch, then run the command shown.

## Workflows

| Workflow | What it does | Agents |
|----------|-------------|--------|
| `feature-build` | Full cycle: planning → architecture → engineering → QA → DevOps | 11 agents |
| `code-review` | Architecture + security + quality review in parallel | 3 agents |
| `bug-fix` | Root cause analysis → fix → verify | 3 agents |
| `incident-response` | Triage → hotfix → verify → post-mortem | 4 agents |
| `architecture-review` | Design doc review → ADR | 2 agents |

## Requirements

- [Claude Code](https://claude.ai/download) with an active subscription or `ANTHROPIC_API_KEY`
- Node.js 20+
- Go 1.22+ (for building from source)

## Build From Source

```bash
git clone https://github.com/<org>/claude-team
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

Agent results are written to `.claude-team/runs/<run_id>/` and visible in the dashboard.

## Troubleshooting

See [docs/troubleshooting.md](docs/troubleshooting.md).

## Contributing

PRs welcome. Run `go test ./...` before opening a PR.

## License

MIT
