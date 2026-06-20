# Anton: Coordinator Redesign

> *Named after Son of Anton — Pied Piper's AI from Silicon Valley. This time it's real.*

**Date:** 2026-06-20  
**Status:** Approved  
**Replaces:** 2026-06-17-claude-team-redesign.md  
**CLI:** `anton dispatch "build user auth"`

---

## Overview

Replace terminal-spawning multi-agent architecture with a native Coordinator pattern using Claude Code's Agent tool. Every "agent" becomes a fresh, task-scoped sub-agent invocation — no OS terminal windows, no AppleScript, no macOS dependency. One Coordinator session drives the entire engineering lifecycle via structured Agent tool calls.

---

## Problem With Current Design

| Problem | Impact |
|---------|--------|
| N OS terminal windows per run (AppleScript) | macOS-only, heavy, opaque |
| Status via fsnotify log file watching | Fragile, race conditions |
| Agents are persistent sessions | Context pollution, bias accumulates |
| No anti-hallucination guarantees | Agents invent APIs, packages |
| Verbose agent output | Wastes tokens |
| No structured escalation | Blocked agents go silent |

---

## Core Architecture

### Fundamental Shift

```
OLD: N OS terminal windows (AppleScript) → one claude CLI process per agent
NEW: 1 Coordinator Claude Code session → Agent tool calls for all sub-agents
```

### System Overview

```
Entry Points:
  /team-dispatch "task" --workflow feature-build   ← Claude Code skill (primary)
  /team-dispatch --from-browser                    ← reads .claude-team/pending-task.md
  Browser UI → POST /api/task → writes pending-task.md → user runs /team-dispatch

Coordinator (current Claude Code session)
  │  reads workflow YAML
  │  writes status to SQLite
  │  dispatches via Agent tool
  │
  ├── Planning Sub-Coordinator        (Agent tool call)
  │     ├── requirements-analyst      (Agent tool call)
  │     └── tech-writer               (Agent tool call)
  │
  ├── Engineering Sub-Coordinator     (Agent tool call)
  │     ├── senior-architect          (Agent tool call)
  │     ├── api-designer              (Agent tool call)
  │     ├── backend-engineer          (Agent tool call, parallel)
  │     ├── frontend-engineer         (Agent tool call, parallel)
  │     └── dba                       (Agent tool call, parallel)
  │
  ├── QA Sub-Coordinator              (Agent tool call)
  │     ├── qa-engineer               (Agent tool call)
  │     ├── security-reviewer         (Agent tool call)
  │     └── e2e-tester                (Agent tool call)
  │
  └── DevOps Sub-Coordinator          (Agent tool call)
        ├── code-reviewer             (Agent tool call)
        └── devops-engineer           (Agent tool call)

Go Backend (thin dashboard server — no spawning)
  ├── Serves browser UI (static files)
  ├── Reads SQLite → WebSocket → browser
  └── Writes pending-task.md on browser dispatch
        Note: browser dispatch is intentionally two-step.
        Browser writes the task file; user runs /team-dispatch --from-browser
        in their Claude Code session. Go never spawns Claude processes.

SQLite (single source of truth)
  ← coordinator writes status
  ← agents write structured results
  → Go backend reads + broadcasts
  → browser displays
```

### Agent Communication

All inter-agent communication routes through coordinator via `coordinator` MCP tool.

**What it is:** `mcp/team-coordinator.js` — updated from current message-bus implementation to write structured JSON to SQLite. Injected into every agent automatically (not user-configurable).

**Tool schema:**
```json
{
  "report": { "result": "<structured JSON string>" },
  "ask":    { "question": "<string>", "context": "<string>" }
}
```

- Agent calls `coordinator.report(result_json)` → writes to SQLite `agent_results`
- Agent calls `coordinator.ask(question)` → coordinator routes to correct specialist
- No direct agent-to-agent calls — all routing through coordinator

---

## Agent Hierarchy

### 3 Levels

```
Level 0 — Main Coordinator
  Reads workflow YAML. Dispatches sub-coordinators. Never implements.
  Routes escalations. Resolves conflicts. Halts on critical security findings.

Level 1 — Phase Sub-Coordinators
  planning-coordinator     Manages requirements phase
  engineering-coordinator  Manages architecture + implementation phase
  qa-coordinator           Manages testing + security phase
  devops-coordinator       Manages deployment + review phase

Level 2 — Specialist Agents (leaf nodes, fresh per task)
  Planning:
    requirements-analyst   Acceptance criteria, edge cases, unknowns
    tech-writer            PRDs, changelogs, API docs

  Architecture:
    senior-architect       System design, ADRs, trade-off analysis
    api-designer           OpenAPI specs, REST/GraphQL contracts

  Engineering:
    backend-engineer       Server-side implementation
    frontend-engineer      UI/component implementation
    dba                    Schema design, migrations, query optimization
    mobile-engineer        iOS/Android (opt-in per workflow)

  QA:
    qa-engineer            Test plans, unit + integration tests
    e2e-tester             Playwright, browser automation
    security-reviewer      OWASP, threat modeling, CVE lookup

  DevOps:
    devops-engineer        CI/CD, Docker, K8s, deployment
    performance-engineer   Load testing, profiling

  Cross-cutting (any workflow):
    code-reviewer          PR review, diff analysis
    debugger               Root cause analysis, reproduction
```

---

## Workflow Templates

### YAML Format

```yaml
name: feature-build
description: Full feature development cycle
version: 1.0

phases:
  - id: planning
    coordinator: planning-coordinator
    sequential:
      - requirements-analyst
      - tech-writer
    outputs: [acceptance-criteria.md, prd.md]

  - id: architecture
    coordinator: engineering-coordinator
    sequential:
      - senior-architect
      - api-designer
    outputs: [adr.md, openapi.yaml]
    when: "phases.planning.status == 'done'"

  - id: engineering
    coordinator: engineering-coordinator
    parallel:
      - backend-engineer
      - frontend-engineer
      - dba
    outputs: [implementation]
    when: "phases.architecture.status == 'done'"

  - id: qa
    coordinator: qa-coordinator
    sequential:
      - qa-engineer
      - security-reviewer
      - e2e-tester
    outputs: [qa-report.md]
    when: "phases.engineering.status == 'done'"

  - id: devops
    coordinator: devops-coordinator
    sequential:
      - code-reviewer
      - devops-engineer
    outputs: [deployment]
    when: "phases.qa.status == 'done'"
```

### Available Workflows

| Workflow | Phases |
|----------|--------|
| `feature-build` | Planning → Architecture → Engineering (parallel) → QA → DevOps |
| `code-review` | Intake → Architecture review → Security → Code quality → Verdict |
| `bug-fix` | Triage → Debugger → Fix (backend/frontend) → QA verify |
| `incident-response` | Alert → RCA → Hotfix → QA → Post-mortem |
| `architecture-review` | Design doc → Architect → Security → ADR output |

---

## MCP Tool Assignments

All MCPs defined in `mcp-registry.yaml`. User enables per-project. Agents receive only what's enabled.

`brave-search` and `tavily` are **mandatory** for every agent — not optional, not user-configured. Baked into every role prompt.

| Agent | Mandatory MCPs | Optional (user-enabled) MCPs |
|-------|---------------|------------------------------|
| **main-coordinator** | `coordinator`, `filesystem`, `brave-search`, `tavily` | — |
| **planning-coordinator** | `coordinator`, `filesystem`, `brave-search`, `tavily` | `atlassian-rovo`, `linear`, `notion`, `slack` |
| **engineering-coordinator** | `coordinator`, `filesystem`, `brave-search`, `tavily` | `github` |
| **qa-coordinator** | `coordinator`, `filesystem`, `brave-search`, `tavily` | `github`, `sentry` |
| **devops-coordinator** | `coordinator`, `filesystem`, `brave-search`, `tavily` | `github`, `datadog` |
| **requirements-analyst** | `filesystem`, `brave-search`, `tavily` | `atlassian-rovo`, `linear`, `notion`, `google-drive`, `github` |
| **tech-writer** | `filesystem`, `brave-search`, `tavily` | `github`, `notion`, `google-drive` |
| **senior-architect** | `filesystem`, `brave-search`, `tavily` | `github`, `figma`, `google-drive` |
| **api-designer** | `filesystem`, `brave-search`, `tavily` | `github`, `atlassian-rovo` |
| **backend-engineer** | `filesystem`, `brave-search`, `tavily` | `github`, `postgres`, `redis`, `supabase`, `docker` |
| **frontend-engineer** | `filesystem`, `brave-search`, `tavily` | `github`, `figma`, `playwright` |
| **dba** | `filesystem`, `brave-search`, `tavily` | `github`, `postgres`, `mysql`, `mongodb`, `redis` |
| **mobile-engineer** | `filesystem`, `brave-search`, `tavily` | `github`, `figma` |
| **qa-engineer** | `filesystem`, `brave-search`, `tavily` | `github`, `sentry`, `datadog` |
| **e2e-tester** | `filesystem`, `brave-search`, `tavily`, `playwright` | `github`, `sentry` |
| **security-reviewer** | `filesystem`, `brave-search`, `tavily` | `github`, `sentry` |
| **devops-engineer** | `filesystem`, `brave-search`, `tavily` | `github`, `docker`, `vercel`, `cloudflare`, `aws`, `datadog`, `sentry` |
| **performance-engineer** | `filesystem`, `brave-search`, `tavily` | `playwright`, `datadog` |
| **code-reviewer** | `filesystem`, `brave-search`, `tavily` | `github` |
| **debugger** | `filesystem`, `brave-search`, `tavily` | `github`, `sentry`, `datadog` |

---

## Non-Negotiable Engineering Standards

All 8 rules baked into **every** role system prompt. No exceptions.

### 1. Search-First Protocol

```
MANDATORY ORDER:
1. Read existing code/context    (filesystem MCP)
2. Search current docs           (brave-search or tavily)
3. Ask coordinator if blocked    (coordinator MCP)
4. Implement only when confident

USE brave-search FOR: broad web, general docs, comparisons
USE tavily FOR: technical deep-dive, RAG over docs, current API specs
```

### 2. Ask-Before-Assume Protocol

```
IF uncertain about ANY of:
  • Requirements intent
  • API behavior or signature
  • Best practice currency
  • Architecture decision rationale
  • Whether something exists in codebase

THEN:
  1. Search first (brave-search or tavily)
  2. If search insufficient → write question to coordinator MCP
  3. STOP and wait — never proceed on assumption

ESCALATION ROUTING (coordinator handles):
  • Requirements ambiguity     → requirements-analyst
  • Architecture decision      → senior-architect
  • Security concern           → security-reviewer (immediate, halts phase)
  • Cross-cutting unknown      → main coordinator routes
```

### 3. Anti-Bias Rules

```
NEVER:
  • Favor technology because familiar from training
  • Recommend library without searching: last commit date, open issues, downloads
  • State best practice without ≤2 year old source URL

ALWAYS:
  • Search current benchmarks before recommending
  • Cite sources in output JSON sources[] field
  • State confidence level: "high | medium | low"
  • Low confidence → mandatory search before finalizing
```

### 4. Anti-Hallucination Rules

```
NEVER:
  • Invent package names, function signatures, API endpoints, config keys
  • Guess version numbers — search or read package files
  • Fill uncertainty gaps with plausible-sounding content
  • Say "typically" or "usually" without a source
  • Complete code example you're unsure about — leave TODO + question

IF don't know → output exactly: "UNKNOWN — searched, not found: <query>"
IF conflicting search results → escalate, cite both sources

VERIFICATION BEFORE OUTPUT:
  • Every external package     → verify exists (npm/pip/go pkg search)
  • Every API call             → verify signature from current docs
  • Every config format        → read from file or official source
  • Every best practice claim  → source URL ≤2 years old in sources[]
```

### 5. Caveman Mode (Token Efficiency)

```
DROP: articles (a/an/the), filler (just/really/basically/actually),
      pleasantries (sure/happy to/of course), hedging (might/could/perhaps)

USE: fragments, short synonyms, exact technical terms
CODE BLOCKS: write normal, unchanged

Pattern: [thing] [action] [reason]. [next step].

BAD:  "I would be happy to implement the authentication middleware..."
GOOD: "Implement JWT auth middleware. RS256 signing. Source: jwt.io/introduction"

Output JSON only for structured results. No prose wrappers.
Summary field: max 3 sentences. Fragments OK.
```

### 6. YAGNI

```
Implement exactly what was asked. Nothing more.
No extra endpoints, no bonus abstractions, no "while I'm here" changes.
Scope creep = rejected by coordinator.
```

### 7. Read Before Write

```
ALWAYS read existing files before editing (filesystem MCP).
Never assume file contents.
Check existing patterns before introducing new ones.
```

### 8. Tests Before DONE

```
Run tests before reporting status DONE.
Report: command run + output in tests_run field.
Never mark DONE without test evidence.
```

---

## Structured Agent Output

Every agent writes structured JSON result — coordinator rejects free text.

```json
{
  "agent": "backend-engineer",
  "run_id": "uuid",
  "phase": "engineering",
  "status": "DONE | DONE_WITH_CONCERNS | NEEDS_CONTEXT | BLOCKED",
  "deliverables": ["src/auth/middleware.ts", "src/auth/middleware.test.ts"],
  "summary": "JWT auth middleware. RS256. Validates exp + iss claims. Rejects on missing header.",
  "confidence": "high | medium | low",
  "sources": ["https://jwt.io/introduction", "src/auth/existing-middleware.ts"],
  "concerns": [],
  "questions": [],
  "tests_run": "12/12 passing — npm test src/auth",
  "tokens_used": 4821,
  "escalation": null
}
```

Coordinator rejects output with:
- Empty `sources[]` when task required research
- `status: DONE` with failing tests
- Missing required fields
- Prose instead of JSON

---

## Coordinator Intelligence

Coordinator routes, never assumes:

| Situation | Action |
|-----------|--------|
| Agent BLOCKED | Re-dispatch with more context; if 3rd retry → escalate to user |
| Agent output contradicts requirements | Return with specific correction |
| Security reviewer finds critical issue | Halt all phases immediately, surface to user |
| Two agents produce conflicting designs | Route conflict to senior-architect |
| Sub-coordinator scope gap | Pause workflow, ask user, resume |
| Low confidence output (no sources) | Reject, require search + re-submit |

---

## Coordinator Dispatch Format

Terse briefs — no pasted session history. File paths only.

```
Dispatch to backend-engineer:
  Task: implement POST /api/auth/login
  Brief: .claude-team/runs/<id>/briefs/backend-engineer-1.md
  Report: .claude-team/runs/<id>/reports/backend-engineer-1.json
  Context files:
    - .claude-team/runs/<id>/adr.md          (architect decisions)
    - .claude-team/runs/<id>/openapi.yaml    (API contract)
  Constraints: fastify only, pg driver, no ORMs
```

---

## Skill Entrypoints

```bash
/team-dispatch "build user auth" --workflow feature-build
/team-dispatch --from-browser      # reads .claude-team/pending-task.md
/team-dispatch                     # interactive: prompts task + workflow

/team-status                       # current run status in terminal
/team-stop                         # halt run, write partial results
```

Registered in `.claude/settings.json`. Primary entry point is always a Claude Code session — the current session becomes the coordinator.

---

## Go Backend (Thin Dashboard Server)

Zero spawning. Zero AppleScript. No process management.

### Responsibilities

```
Serve browser UI     Static files from ui/
Write pending task   POST /api/task → .claude-team/pending-task.md
Read SQLite          Poll for status changes → WebSocket broadcast
Input processing     PDF text extraction, md pass-through, Jira URL pass-through
```

### API Routes

```
GET  /                    Serve UI
POST /api/task            Write pending-task.md (browser dispatch)
GET  /api/runs            Run history
GET  /api/runs/:id        Run detail + phase/agent status
GET  /api/workflows       List workflow YAMLs
WS   /ws                  Live status stream
```

### SQLite Schema

```sql
runs          (id, workflow, status, started_at, completed_at)
phases        (run_id, phase_id, status, started_at, completed_at)
agent_results (run_id, phase_id, agent, status, confidence, sources_json,
               deliverables_json, summary, concerns_json, tests_run,
               tokens_used, created_at)
messages      (id, run_id, from_agent, to_agent, content, created_at)
```

---

## Browser UI

Vanilla HTML/CSS/JS. Zero build step. Dark theme. Go serves `ui/` via `net/http.FileServer`.

### Layout

```
┌──────────────────────────────────────────────────────────────┐
│  claude-team                  [feature-build ▾]  [▶ Dispatch]│
├───────────────────────┬──────────────────────────────────────┤
│  INPUT                │  AGENT TREE (live SVG)               │
│                       │                                       │
│  [Task description]   │  ⟳ Coordinator                       │
│                       │    ├── ✓ Planning Sub-Coord           │
│  [drag files / PDF]   │    │     ├── ✓ requirements-analyst  │
│  [Jira / Linear URL]  │    │     └── ⟳ tech-writer           │
│                       │    ├── ○ Engineering Sub-Coord        │
│  Workflow:            │    │     ├── ○ senior-architect       │
│  [feature-build ▾]    │    │     ├── ○ backend-engineer       │
│                       │    │     └── ○ frontend-engineer      │
│  MCPs enabled:        │    └── ○ QA Sub-Coord                 │
│  [github][figma][pg]  │                                       │
├───────────────────────┴──────────────────────────────────────┤
│  PHASE PROGRESS                                               │
│  [Planning ✓] → [Architecture ⟳] → [Engineering ○] →        │
│  [QA ○] → [DevOps ○]                                         │
├──────────────────────────────────────────────────────────────┤
│  ACTIVE: tech-writer                    confidence: high      │
│  Writing PRD from acceptance criteria                         │
│  Sources: notion.so/… · drive.google.com/…                   │
│  Tokens: 4,821                                               │
└──────────────────────────────────────────────────────────────┘
```

Status dots: ○ pending · ⟳ running · ✓ done · ✗ failed · ⚠ blocked

Agent card expands on click → full JSON output, sources list, concerns, test results.

### Components

- **Agent Tree** — pure SVG DAG, WebSocket updates node state live
- **Phase Progress** — step bar, highlights active phase
- **Input Panel** — textarea, HTML5 drag-drop (pdf/md/yaml), Jira URL field
- **Active Agent Card** — name, summary, confidence, sources, token count
- **MCP Selector** — checkboxes per available MCP, persisted to localStorage
- **Run History** — sidebar list, click to replay any past run's agent tree

---

## File Structure

```
claude-team/
├── coordinators/
│   ├── main.md                    Main coordinator system prompt
│   ├── planning.md
│   ├── engineering.md
│   ├── qa.md
│   └── devops.md
├── roles/
│   ├── requirements-analyst.md
│   ├── tech-writer.md
│   ├── senior-architect.md
│   ├── api-designer.md
│   ├── backend-engineer.md
│   ├── frontend-engineer.md
│   ├── dba.md
│   ├── mobile-engineer.md
│   ├── qa-engineer.md
│   ├── e2e-tester.md
│   ├── security-reviewer.md
│   ├── devops-engineer.md
│   ├── performance-engineer.md
│   ├── code-reviewer.md
│   └── debugger.md
├── workflows/
│   ├── feature-build.yaml
│   ├── code-review.yaml
│   ├── bug-fix.yaml
│   ├── incident-response.yaml
│   └── architecture-review.yaml
├── skills/
│   ├── team-dispatch.md
│   ├── team-status.md
│   └── team-stop.md
├── internal/
│   ├── api/
│   │   ├── server.go
│   │   ├── handlers.go
│   │   └── ws.go
│   ├── store/
│   │   ├── store.go
│   │   └── watcher.go
│   └── input/
│       └── processor.go
├── ui/
│   ├── index.html
│   ├── app.js
│   └── styles.css
├── mcp-registry.yaml
├── main.go
└── .claude-team/                  (gitignored, runtime)
    ├── pending-task.md
    ├── claude-team.db
    └── runs/
        └── <run-id>/
            ├── context.md
            ├── briefs/
            └── reports/
```

---

## What Gets Removed

| Removed | Replaced by |
|---------|-------------|
| `internal/terminal/spawner.go` | Agent tool calls in coordinator |
| AppleScript / osascript | Native Claude Code Agent spawning |
| `internal/status/watcher.go` (fsnotify) | SQLite writes from coordinator |
| `.claude-team/scripts/` | `.claude-team/runs/<id>/briefs/` |
| `.claude-team/logs/` | SQLite `agent_results` table |
| `.claude-team/handoff/` | Structured JSON in SQLite |
| `internal/workflow/executor.go` (terminal-based) | Coordinator role prompt + Agent tool |
| Per-agent `settings.json` with `CLAUDE_CONFIG_DIR` | MCP list per-agent via Agent tool `agentType` |
| `roles/manager.md`, `roles/peer-programmer.md` | Full role set in `roles/` + `coordinators/` |

## What Gets Kept

| Kept | Notes |
|------|-------|
| `mcp-registry.yaml` | Extended with full MCP catalog |
| SQLite (modernc.org/sqlite) | Schema extended |
| WebSocket hub | Driver changes from fsnotify to SQLite polling |
| Vanilla HTML/JS UI | Layout redesigned |
| `workflow.yaml` YAML format | Extended with `when:` conditions + coordinator fields |
| Go backend | Slimmed to dashboard server only |

---

## LinkedIn Launch Story

### Hook

> "I replaced my entire sprint planning + dev + QA cycle with a self-orchestrating multi-agent system built on Claude Code.
>
> Not a chatbot. A Coordinator that spawns specialist agents — architect, backend, frontend, QA, security reviewer — each with their own tools, each required to search docs before writing a line, each forced to ask questions instead of guessing.
>
> Anti-hallucination by design: every agent cites sources. Low confidence = mandatory search. If stuck = escalates. Never assumes.
>
> Open source. Plugs into Jira, GitHub, Figma, Notion, Postgres, Vercel."

### Demo Flow (90 seconds)

1. Type task in browser: *"Add Stripe subscription billing"*
2. Select `feature-build` workflow, enable GitHub + Stripe MCPs
3. Hit Dispatch → run `/team-dispatch --from-browser` in Claude Code
4. Agent tree lights up live in browser
5. requirements-analyst completes — sources cited, acceptance criteria written
6. senior-architect designs — ADR written, trade-offs documented
7. backend + frontend + dba run in parallel
8. security-reviewer flags one issue → coordinator halts, routes fix to backend
9. All green → code-reviewer approves → PR opened on GitHub
10. Show outputs: `adr.md`, `openapi.yaml`, test results, PR link

### Differentiators

- Not terminal windows — a live agent tree in the browser
- Not a wrapper — genuine coordinator intelligence with escalation routing
- Anti-hallucination baked in — sources required, not optional
- Caveman mode — agents use minimal tokens, fast execution
- Workflow library — feature build, code review, bug fix, incident response, architecture review
- Open source — fork + customize roles + add MCPs
