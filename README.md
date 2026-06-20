# Anton

**Describe a task. Get a full engineering team.**

Anton spins up a coordinated squad of specialist AI agents — planner, architect, engineers, QA, security reviewer — and shows their live progress in a browser dashboard. One slash command kicks off the whole thing from Claude Code.

[![CI](https://github.com/kabirnarang39/claude-team/actions/workflows/ci.yml/badge.svg)](https://github.com/kabirnarang39/claude-team/actions/workflows/ci.yml)
[![Go 1.22+](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/kabirnarang39/claude-team)](https://github.com/kabirnarang39/claude-team/releases)

![Anton demo — agents completing phase by phase](docs/demo.gif)

---

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

**Dispatch a task** — open Claude Code in your project directory, then run:
```
/team-dispatch build user authentication with JWT and refresh tokens
```

Or use the browser: open `http://localhost:3000`, enter your task, pick a workflow, click **Dispatch**, and paste the generated command into Claude Code.

---

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

Agent outputs land in `.claude-team/runs/<run_id>/` and are viewable in the dashboard by clicking any agent node.

---

## Workflows

| Workflow | What it does | Agents |
|----------|-------------|--------|
| `feature-build` | Full cycle: planning → architecture → engineering → QA → DevOps | 11 agents |
| `code-review` | Architecture + security + quality review in parallel | 3 agents |
| `bug-fix` | Root cause analysis → fix → verify | 3 agents |
| `incident-response` | Triage → hotfix → verify → post-mortem | 4 agents |
| `architecture-review` | Design doc review → ADR | 2 agents |

---

## Adding a Workflow

Drop a YAML file into `workflows/`. Anton picks it up immediately — no restart needed.

```yaml
name: my-workflow
description: What this workflow does

phases:
  - id: planning
    sequential:
      - requirements-analyst
  - id: engineering
    parallel:
      - backend-engineer
      - frontend-engineer
  - id: review
    sequential:
      - code-reviewer
      - security-reviewer
```

**Sequential** agents run one after another. **Parallel** agents run concurrently in the same phase.

Available roles: `requirements-analyst`, `tech-writer`, `senior-architect`, `api-designer`, `backend-engineer`, `frontend-engineer`, `dba`, `qa-engineer`, `e2e-tester`, `security-reviewer`, `code-reviewer`, `debugger`, `devops-engineer`, `performance-engineer`, `mobile-engineer`.

---

## Adding a Role

Each role is a Markdown system prompt in `roles/`. To add a specialist:

1. Create `roles/my-specialist.md` — write the agent's purpose, what inputs it receives, and what it must output.
2. Follow `roles/_standards.md` — all agents report results in the same JSON format so the dashboard can display them.
3. Reference your role name in any workflow YAML.

---

## CLI Reference

```
Usage of anton:
  -port int
        HTTP port (default 3000)
  -registry string
        Path to mcp-registry.yaml (default "mcp-registry.yaml")
  -version
        Print version and exit
```

---

## Requirements

- [Claude Code](https://claude.ai/download) — active subscription
- Node.js 20+
- Go 1.22+ (build from source only)

---

## Build From Source

```bash
git clone https://github.com/kabirnarang39/claude-team
cd claude-team
cd mcp && npm install && cd ..
go run main.go
```

Run tests:
```bash
go test ./...
```

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for dev setup, how to add workflows and roles, and PR guidelines.

## Security

See [SECURITY.md](SECURITY.md) for the vulnerability disclosure process.

## License

MIT
