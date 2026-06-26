# Social Copy — Anton Demo Launch

---

## X / Twitter

**Primary post (attach demo-dashboard.mp4):**

```
I gave Claude Code a team of 12 AI specialists.

One slash command. They work in parallel. You watch them live.

/team-dispatch build user auth with JWT

→ requirements analyst writes acceptance criteria
→ architect designs the RS256 token strategy  
→ backend + frontend + DBA run in parallel
→ security reviewer audits OWASP Top 10
→ devops engineer ships Dockerfile + Helm chart

No new API key. No LangChain. Runs inside your Claude Code subscription.

github.com/kabirnarang39/claude-team
```

**Thread follow-up (reply to above):**

```
What's different from CrewAI / AutoGen:

✅ Runs inside Claude Code (no new subscription)
✅ Live browser dashboard — click any agent to read its output
✅ Plain YAML workflows + Markdown roles (fork in 5 min)
✅ Resume on crash — checkpoint per phase
✅ Cost routing: haiku → sonnet → opus by complexity

Agents produce structured outputs in .claude-team/runs/ — yours to read and version-control.
```

---

## Reddit — r/ClaudeAI

**Title:**
```
I built a multi-agent team for Claude Code — 12 specialists, live browser dashboard, no new subscription
```

**Body:**
```
I got tired of context bloat in long Claude Code sessions, so I built Anton.

One slash command dispatches a full engineering team:

```
/team-dispatch build user auth with JWT and refresh tokens
```

12 specialists run across 5 phases — planning → architecture → engineering (3 parallel) → QA → DevOps. You watch them work in a live browser dashboard.

**What makes this different:**

- **Context isolation**: each agent starts fresh (~2-4k tokens). Solo 10-agent session grows as N(N+1)/2. Anton is 5.5x less context overhead.
- **Parallel speedup**: backend, frontend, and DBA run concurrently. Real 3x speedup on the engineering phase.
- **Observable**: click any agent node to read its full output, confidence score, and deliverables.
- **Human review gates**: Anton pauses after planning and architecture for your approval. Reject with feedback → re-runs the phase.
- **Resume on crash**: `/team-resume` picks up from last checkpoint.
- **Cost routing**: haiku for boilerplate, sonnet for implementation, opus for architect and security reviewer.

No new API key. No venv. No LangChain. Runs inside your existing Claude Code subscription.

Workflows are plain YAML. Agent roles are plain Markdown. Fork either in under 10 minutes.

GitHub: https://github.com/kabirnarang39/claude-team

Happy to answer questions on how the MCP coordination layer works.
```

---

## Reddit — r/LocalLLaMA

**Title:**
```
Anton: multi-agent engineering team inside Claude Code — 12 specialists, parallel phases, live DAG dashboard
```

**Body:**
```
Built a multi-agent orchestration layer on top of Claude Code's native agent capabilities.

Architecture: coordinator reads a YAML workflow, spins up specialist sub-agents (planner, architect, 3 engineers, QA, security, DevOps) via Claude Code's Agent tool, each with fresh context. Results flow through a Node.js MCP server → SQLite → Go HTTP/WebSocket → browser dashboard.

Key design decisions:

**Context isolation over shared context.** Each agent sees only: its role prompt + prior phase outputs. At 10 agents, shared context grows as N(N+1)/2 turns. Fresh agents = 5.5x less overhead. Practically: agents don't hallucinate about what other agents said.

**Parallel execution on independent tasks.** Backend, frontend, DBA are independent — they run concurrently in the same phase. Wall-clock 3x vs sequential.

**Structured output contract.** Every agent reports a JSON envelope (status, confidence, summary, deliverables, concerns, tokens_used). Dashboard reads this contract — not free-text parsing.

**Cost routing.** Haiku handles boilerplate (tech-writer, requirements extraction), Sonnet handles implementation, Opus handles architect/security/code-reviewer where reasoning quality matters.

25 MCP integrations auto-registered from mcp-registry.yaml — agents have access to Jira, GitHub, Postgres, Sentry, Slack, etc. during their phase.

GitHub: https://github.com/kabirnarang39/claude-team

YAML workflows + Markdown roles if you want to fork/extend.
```

---

## Hacker News — Show HN

**Title:**
```
Show HN: Anton – 12 AI specialists work in parallel inside Claude Code
```

**Body:**
```
Anton gives Claude Code a team of specialist sub-agents that run in parallel and report to a live browser dashboard.

One slash command dispatches the team:

  /team-dispatch build user auth with JWT and refresh tokens

Twelve agents run across 5 phases — requirements analyst, tech writer, senior architect, API designer, three engineers (parallel), QA engineer, security reviewer, E2E tester, code reviewer, DevOps engineer.

Each agent starts with fresh context and its own role prompt. Results are written to SQLite, streamed over WebSocket, and displayed in a DAG view with a per-agent inspector panel.

Technical notes:
- Coordination via a Node.js MCP stdio server providing `report`, `ask`, `inbox`, `reply` tools
- Go HTTP server (SQLite WAL, WebSocket hub) — no external dependencies
- Plain YAML workflows, plain Markdown role prompts — fork either in under 10 min
- Human review gates after planning and architecture phases
- Checkpoint + resume via /team-resume
- Cost-smart model routing (haiku → sonnet → opus by task)
- 25 MCP integrations in mcp-registry.yaml (Jira, GitHub, Postgres, Sentry, Datadog, etc.)

No new API key, no venv, no LangChain. Runs inside your existing Claude Code subscription.

Install: curl -fsSL https://raw.githubusercontent.com/kabirnarang39/claude-team/main/install.sh | sh

Source: https://github.com/kabirnarang39/claude-team
```

---

## Timing guidance

- **Post X first** — if it gets traction, post HN same day while momentum is up
- **Best HN time**: 8–9am EST weekday (Mon–Wed highest traffic)
- **Best X time**: 9–11am EST or 7–9pm EST weekday
- **Reddit**: post to r/ClaudeAI first, then r/LocalLLaMA 2 hours later (avoid duplicate flag)
- **Don't** post all at once — stagger by 2-3 hours so each post has breathing room
