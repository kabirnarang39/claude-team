# Anton

**One command. A full engineering team.**

Anton gives your Claude Code session a coordinated squad of specialist AI agents — planner, architect, backend and frontend engineers, DBA, QA, security reviewer, DevOps. They coordinate through structured workflows and report live to a browser dashboard. You stay in control; they do the work.

[![CI](https://github.com/kabirnarang39/claude-team/actions/workflows/ci.yml/badge.svg)](https://github.com/kabirnarang39/claude-team/actions/workflows/ci.yml)
[![Go 1.22+](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/kabirnarang39/claude-team)](https://github.com/kabirnarang39/claude-team/releases)

![Anton demo — 11 agents completing a feature-build workflow phase by phase](docs/demo.gif)

---

## Table of Contents

- [Quick Start](#quick-start)
- [What You Get](#what-you-get)
- [Workflows](#workflows)
- [Why Anton?](#why-anton)
- [How It Works](#how-it-works)
- [Add a Workflow](#add-a-workflow)
- [Add a Role](#add-a-role)
- [CLI Reference](#cli-reference)
- [Build From Source](#build-from-source)
- [Contributing](#contributing)

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

**Dispatch a task** — open Claude Code in your project directory:

```
/team-dispatch build user authentication with JWT and refresh tokens
```

Open `http://localhost:3000` and watch 11 specialist agents work through planning, architecture, engineering, QA, and DevOps — live.

> **Browser dispatch:** Go to `http://localhost:3000`, enter your task, pick a workflow, click **▶ Dispatch**, then paste the generated command into Claude Code.

---

## What You Get

After a `feature-build` run on *"user auth with JWT"*, your agents produce structured outputs in `.claude-team/runs/<run_id>/`:

| Agent | Produces | Example |
|-------|----------|---------|
| `requirements-analyst` | `acceptance-criteria.md` | 12 requirements, edge cases, user stories |
| `tech-writer` | `prd.md` | Product requirements doc, scope, open questions |
| `senior-architect` | `adr.md` | JWT vs sessions decision, Redis architecture |
| `api-designer` | `openapi.yaml` | 5 endpoints, request/response schemas |
| `backend-engineer` | `backend-report.md` | Implementation plan, key decisions, risks |
| `frontend-engineer` | `frontend-report.md` | Component structure, auth flow, state handling |
| `dba` | `dba-report.md` | Schema design, indices, migration strategy |
| `qa-engineer` | `qa-report.md` | Test cases, coverage plan, integration scenarios |
| `security-reviewer` | `security-report.md` | OWASP checklist, findings, mitigations |
| `e2e-tester` | `e2e-report.md` | Playwright test plan, edge cases |
| `devops-engineer` | `devops-report.md` | Dockerfile, CI/CD, Helm chart plan |

Every output is viewable in the dashboard — click any agent node to read its full report.

See [`docs/examples/`](docs/examples/) for real sample outputs.

---

## Workflows

| Workflow | What it does | Phases | Agents |
|----------|-------------|--------|--------|
| `feature-build` | Full cycle: planning → architecture → engineering → QA → DevOps | 5 | 11 |
| `code-review` | Architecture + security + quality review in parallel | 1 | 3 |
| `bug-fix` | Root cause analysis → fix plan → verify | 2 | 3 |
| `incident-response` | Triage → hotfix → verify → post-mortem | 3 | 4 |
| `architecture-review` | Design doc review → ADR | 1 | 2 |

Each workflow is a plain YAML file — [add your own](#add-a-workflow) in under 5 minutes.

---

## Why Anton?

**Anton is for Claude Code users who want their AI session to do more without leaving their terminal.**

### vs. CrewAI / AutoGen

| | Anton | CrewAI | AutoGen |
|--|-------|--------|---------|
| Runs inside Claude Code | ✅ | ❌ | ❌ |
| Uses your Claude Code subscription | ✅ | ❌ | ❌ |
| No Python, no venv, no LangChain | ✅ | ❌ | ❌ |
| Live browser dashboard | ✅ | ❌ | ❌ |
| Workflows in plain YAML | ✅ | ⚠️ code | ⚠️ code |
| Custom agent roles in Markdown | ✅ | ⚠️ code | ⚠️ code |
| SQLite state (survives restarts) | ✅ | ❌ | ❌ |
| Local / offline-first | ✅ | ✅ | ✅ |
| Open source | ✅ | ✅ | ✅ |

### vs. GitHub Copilot Workspace / Devin

Anton doesn't try to replace your judgment. Agents produce structured reports — you review, decide, and implement. No surprise commits, no mystery diffs. You own the process.

### The Core Idea

Most multi-agent frameworks require you to write orchestration code. Anton's orchestration *is* a Claude Code session — your existing subscription, your existing tools, your existing workflow. You add one slash command.

---

## How It Works

```
You → /team-dispatch → Main Coordinator (your Claude Code session)
                              │
             ┌────────────────┼────────────────┐
             ▼                ▼                ▼
       Planning           Engineering       QA
       Coordinator        Coordinator       Coordinator
             │                │                │
       requirements-   senior-architect   qa-engineer
       analyst         api-designer       security-reviewer
       tech-writer     backend-engineer   e2e-tester
                       frontend-engineer
                       dba
                              │
                    MCP coordinator tool    ← agents report results
                              │
                       SQLite state.db      ← source of truth
                              │
                     Go HTTP + WebSocket    ← streams to browser
                              │
                    Browser dashboard       ← live agent tree
```

1. **You dispatch** a task from Claude Code with `/team-dispatch`.
2. **The coordinator** reads the workflow YAML and spins up sub-coordinators for each phase.
3. **Each agent** reads its role system prompt (`roles/<name>.md`), does the work, and reports back via the MCP tool.
4. **Anton's Go server** writes results to SQLite and streams updates over WebSocket.
5. **The dashboard** shows live progress — click any agent to read its full output.

All agent outputs land in `.claude-team/runs/<run_id>/` as plain Markdown files.

---

## Add a Workflow

Drop a `.yaml` file into `workflows/`. Anton picks it up immediately — no restart:

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

**`sequential`** — agents run one after another, each receiving the previous agent's output context.  
**`parallel`** — agents run concurrently in the same phase.

**Available roles:** `requirements-analyst` · `tech-writer` · `senior-architect` · `api-designer` · `backend-engineer` · `frontend-engineer` · `dba` · `qa-engineer` · `e2e-tester` · `security-reviewer` · `code-reviewer` · `debugger` · `devops-engineer` · `performance-engineer` · `mobile-engineer`

---

## Add a Role

Each role is a Markdown system prompt in `roles/`. To add a specialist:

1. Create `roles/my-specialist.md` — write the agent's purpose, what it receives, what it must output.
2. Follow `roles/_standards.md` — all agents report in the same JSON format so the dashboard can display them.
3. Reference your role in any workflow YAML.

---

## CLI Reference

```
anton [flags]

Flags:
  -port int       HTTP port (default 3000)
  -registry str   Path to mcp-registry.yaml (default "mcp-registry.yaml")
  -version        Print version and exit
```

Anton listens on `127.0.0.1` (localhost only). Do not expose it on a public interface.

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

**Found a bug?** [Open an issue](https://github.com/kabirnarang39/claude-team/issues/new?template=bug_report.md)  
**Have an idea?** [Request a feature](https://github.com/kabirnarang39/claude-team/issues/new?template=feature_request.md)

## Security

See [SECURITY.md](SECURITY.md) for the vulnerability disclosure process.

## License

MIT — see [LICENSE](LICENSE).
