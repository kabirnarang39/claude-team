# Anton

**11 specialist AI agents. One slash command. Live in your browser.**

```
/team-dispatch build user auth with JWT and refresh tokens
```

A planner maps requirements. An architect writes the ADR. Three engineers tackle backend, frontend, and database in parallel. QA writes test cases. Security checks OWASP. DevOps writes the Dockerfile. You watch every agent's reasoning live in a browser dashboard.

[![CI](https://github.com/kabirnarang39/claude-team/actions/workflows/ci.yml/badge.svg)](https://github.com/kabirnarang39/claude-team/actions/workflows/ci.yml)
[![Go 1.22+](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/kabirnarang39/claude-team)](https://github.com/kabirnarang39/claude-team/releases)

**Works inside your existing Claude Code session. No new API key. No Python. No venv.**

![Anton demo — 11 agents completing a feature-build workflow phase by phase](docs/demo.gif)

The workflow YAML files (`workflows/`) and agent role prompts (`roles/`) are plain text — read them, fork them, make them yours.

> If Anton saves you time, ⭐ this repo — it helps others find it.

---

## Table of Contents

- [Quick Start](#quick-start)
- [What You Get](#what-you-get)
- [The Agent Roles](#the-agent-roles)
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
cd your-project
anton
# Anton running at http://localhost:3000
# MCP coordinator auto-registered in .claude/settings.json
```

**Open Claude Code in the same directory:**

```bash
claude
```

**Dispatch a task:**

```
/team-dispatch build user authentication with JWT and refresh tokens
```

Open `http://localhost:3000` and watch 11 specialist agents work through planning, architecture, engineering, QA, and DevOps — live.

> **Verify setup:** Run `anton --check` to confirm everything is wired correctly.  
> **Browser dispatch:** Enter a task at `http://localhost:3000`, click **▶ Dispatch**, paste the command into Claude Code.

---

## What You Get

Anton agents produce **structured analysis and planning outputs** — acceptance criteria, architecture decisions, API specs, security reports, test plans, infrastructure configs. Every file lands in `.claude-team/runs/<run_id>/` and is readable in the dashboard.

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

See [`docs/examples/`](docs/examples/) for real sample outputs.

---

## The Agent Roles

Every agent is a plain Markdown system prompt in [`roles/`](roles/). Here's what they actually say:

**`roles/security-reviewer.md`** (excerpt):
```
Audit code for OWASP Top 10 vulnerabilities. Never guess — search CVEs and
current OWASP docs. Halt the entire phase on critical findings.

Approach:
1. Read all implementation files (filesystem MCP)
2. Search OWASP Top 10 current list (brave-search: "OWASP Top 10 site:owasp.org")
3. Check each file for: injection, broken auth, XSS, IDOR, security misconfiguration
4. For each finding: severity, file, line, description, fix, OWASP/CVE reference
```

**`roles/senior-architect.md`** (excerpt):
```
You design systems. You do NOT write application code or tests.
Never hallucinate API signatures or package names — if unsure, search before stating.
Before recommending any library: check its current maintenance status.
Read existing code structure before proposing architecture.
```

**`roles/requirements-analyst.md`** (excerpt):
```
Extract clear, unambiguous acceptance criteria. Never guess. Never assume.
If a Jira/Linear URL is present, fetch the ticket. Search domain context before writing.
```

Read and fork the full prompts in [`roles/`](roles/). Add your own specialist in under 10 minutes.

---

## Workflows

| Workflow | What it does | Phases | Agents |
|----------|-------------|--------|--------|
| `feature-build` | Full cycle: planning → architecture → engineering → QA → DevOps | 5 | 11 |
| `code-review` | Architecture + security + quality review in parallel | 1 | 3 |
| `bug-fix` | Root cause analysis → fix plan → verify | 2 | 3 |
| `incident-response` | Triage → hotfix → verify → post-mortem | 3 | 4 |
| `architecture-review` | Design doc review → ADR | 1 | 2 |

Each workflow is a plain YAML file in [`workflows/`](workflows/) — [add your own](#add-a-workflow) in under 5 minutes.

---

## Why Anton?

**Anton is for Claude Code users who want parallelism, specialization, and observability without leaving their terminal.**

### Claude Code is one agent. Anton is a team.

- **Parallelism**: your architect is writing the ADR while your backend engineer is planning the API while your DBA is designing the schema — simultaneously.
- **Specialization**: your security reviewer only does security. It has the OWASP docs in context. It doesn't context-switch.
- **Observability**: you watch every agent's reasoning live. Click any node in the dashboard to read its full output.
- **Zero orchestration code**: the workflows are YAML. The roles are Markdown. You read them, edit them, own them.

### vs. CrewAI / AutoGen / MetaGPT

| | Anton | CrewAI | AutoGen | MetaGPT |
|--|-------|--------|---------|---------|
| Runs inside Claude Code | ✅ | ❌ | ❌ | ❌ |
| Uses your Claude Code subscription | ✅ | ❌ | ❌ | ❌ |
| No Python, no venv, no LangChain | ✅ | ❌ | ❌ | ❌ |
| Live browser dashboard | ✅ | ❌ | ❌ | ❌ |
| Workflows in plain YAML | ✅ | ⚠️ code | ⚠️ code | ⚠️ code |
| Agent roles in plain Markdown | ✅ | ⚠️ code | ⚠️ code | ⚠️ code |
| SQLite state (survives restarts) | ✅ | ❌ | ❌ | ❌ |
| Local / offline-first | ✅ | ✅ | ✅ | ✅ |
| Open source | ✅ | ✅ | ✅ | ✅ |

### vs. GitHub Copilot Workspace / Devin

Anton doesn't try to replace your judgment or write code autonomously. Agents produce structured analysis, plans, and specifications — you review, decide, and implement. No surprise commits, no mystery diffs. You own the process.

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
2. **The coordinator** reads the workflow YAML and spins up sub-coordinators per phase.
3. **Each agent** reads its role prompt (`roles/<name>.md`), does its work, and reports via the MCP tool.
4. **Anton's Go server** writes results to SQLite and streams updates over WebSocket.
5. **The dashboard** shows live progress — click any agent to read its full output.

All outputs land in `.claude-team/runs/<run_id>/` as plain Markdown files.

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

**`sequential`** — agents run one after another, each receiving the previous agent's output.  
**`parallel`** — agents run concurrently in the same phase.

**Available roles:** `requirements-analyst` · `tech-writer` · `senior-architect` · `api-designer` · `backend-engineer` · `frontend-engineer` · `dba` · `qa-engineer` · `e2e-tester` · `security-reviewer` · `code-reviewer` · `debugger` · `devops-engineer` · `performance-engineer` · `mobile-engineer`

---

## Add a Role

Each role is a Markdown system prompt in `roles/`. To add a specialist:

1. Create `roles/my-specialist.md` — write the agent's purpose, what it receives, what it must output.
2. Follow `roles/_standards.md` — all agents report in the same JSON format so the dashboard can display them.
3. Reference your role name in any workflow YAML.

---

## CLI Reference

```
Usage of anton:
  -check
        Check Anton setup and exit
  -demo
        Pre-populate dashboard with a sample completed run
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

MIT — see [LICENSE](LICENSE).
