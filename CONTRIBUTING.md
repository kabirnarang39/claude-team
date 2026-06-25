# Contributing to Anton

## Dev Setup

```bash
git clone https://github.com/kabirnarang39/claude-team
cd claude-team
cd mcp && npm install && cd ..
go run main.go          # Anton running at http://localhost:3000
```

**Requirements:** Go 1.25+, Node.js 20+, Claude Code

Run tests before opening a PR:

```bash
go test -race ./...
```

## Project Layout

```
main.go                  # Entry point, config wiring
internal/
  api/                   # HTTP handlers, WebSocket hub, server
  store/                 # SQLite store (runs, agents, phases)
  workflow/              # Workflow YAML parser
  mcp/                   # MCP registry loader
mcp/
  team-coordinator.js    # Stdio MCP server (report tool)
workflows/               # Built-in workflow YAML files
roles/                   # Agent system prompt files
skills/                  # Claude Code slash-command skills
ui/                      # Vanilla JS dashboard (no build step)
```

## Adding a Workflow

Workflows live in `workflows/`. Each is a YAML file that defines phases and the agents that run in each phase.

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
```

**Sequential** agents run one after another (each receives the previous agent's output). **Parallel** agents run concurrently.

Available agent roles: `requirements-analyst`, `tech-writer`, `senior-architect`, `api-designer`, `backend-engineer`, `frontend-engineer`, `dba`, `qa-engineer`, `e2e-tester`, `security-reviewer`, `code-reviewer`, `debugger`, `devops-engineer`, `performance-engineer`, `mobile-engineer`.

## Adding a Role

Roles are agent system prompts in `roles/`. Each file defines how a specialist agent behaves.

1. Create `roles/my-specialist.md` — write the system prompt. Include: the agent's purpose, what inputs it receives, and what it must deliver.
2. Reference `roles/_standards.md` — all roles must follow the output standards defined there (JSON report format, confidence scores, deliverables).
3. Add the new role name to any workflow YAML that should use it.

## Pull Request Guidelines

- One logical change per PR
- `go test -race ./...` must pass
- New handlers need a corresponding test in `internal/api/server_test.go`
- Keep the dashboard vanilla JS — no build toolchain

## Reporting Bugs

Use [GitHub Issues](https://github.com/kabirnarang39/claude-team/issues). Include your platform, Anton version (`anton --version`), and steps to reproduce.
