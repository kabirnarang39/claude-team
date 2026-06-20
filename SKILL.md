---
name: anton
description: Anton — multi-agent engineering team coordinator. Runs planning, architecture, engineering, QA, and DevOps agents in sequence using Claude Code's Agent tool. Anti-hallucination by design. Search-first. No AppleScript. No terminal windows.
version: 1.0.0
skills:
  - team-dispatch
  - team-status
  - team-stop
---

# Anton

Multi-agent engineering orchestration for Claude Code.

## What it does

Anton runs a full engineering team as sub-agents — no OS terminals, no AppleScript. The current Claude Code session becomes the Coordinator. Specialist agents (planner, architect, engineers, QA, security) run as isolated Agent tool calls and write structured results to SQLite. A Go server serves a live browser dashboard.

## Workflows

| Workflow | Description |
|----------|-------------|
| `feature-build` | Full cycle: planning → architecture → engineering → QA → DevOps |
| `code-review` | Architecture + security + quality review in parallel |
| `bug-fix` | Root cause analysis → fix → verify |
| `incident-response` | Triage → hotfix → verify → post-mortem |
| `architecture-review` | Design doc review → ADR |

## Skills

| Skill | Trigger | What it does |
|-------|---------|-------------|
| team-dispatch | `/team-dispatch <task>` | Runs a full workflow end-to-end as coordinator |
| team-status | `/team-status` | Shows live run status from SQLite |
| team-stop | `/team-stop` | Halts current run gracefully |

## Quick start

```bash
# 1. Install MCP deps
cd mcp && npm install && cd ..

# 2. Start dashboard
go run main.go
# → Anton running at http://localhost:3000

# 3. Dispatch a task (in Claude Code)
/team-dispatch build user authentication with JWT and refresh tokens
```

## Architecture

```
User → /team-dispatch skill → Main Coordinator (Claude Code session)
                                    │
                    ┌───────────────┼───────────────┐
                    ▼               ▼               ▼
             Planning Sub-    Engineering    QA Sub-Coordinator
             Coordinator      Sub-Coordinator
                    │               │               │
             requirements-  senior-architect   qa-engineer
             analyst        api-designer       security-reviewer
             tech-writer     backend-engineer  e2e-tester
                             frontend-engineer
                             dba
                                    │
                              coordinator MCP
                              (writes to SQLite)
                                    │
                              Go HTTP server
                              (reads SQLite, WebSocket)
                                    │
                              Browser dashboard
                              (live agent tree SVG)
```

## Engineering standards (all agents)

All 14 specialist agents enforce:
1. Search-first (brave-search + tavily mandatory)
2. Ask-before-assume (never guess requirements)
3. Anti-bias (cite sources for all recommendations)
4. Anti-hallucination (no invented APIs or packages)
5. Caveman mode (minimal tokens)
6. YAGNI (no scope creep)
7. Read-before-write
8. Tests-before-DONE

## References

- Pied Piper built Son of Anton. This is the real thing.
- Silicon Valley (HBO, S3) — the canonical inspiration
