# Product Hunt

## Tagline

A local 12-agent workflow dashboard for Claude Code.

## Short Description

Anton wraps Claude Code with YAML workflows, Markdown specialist roles, SQLite run state, and a live browser dashboard so multi-step engineering work is easier to inspect and resume.

## First Comment

I built Anton after using Claude Code for larger engineering tasks and wanting more structure than one long conversation.

Anton runs locally:

- Go HTTP/WebSocket server
- SQLite run state
- Node MCP coordinator
- vanilla JS dashboard
- YAML workflows
- Markdown role prompts

It does not call the Anthropic API directly and it is not a hosted agent platform. It uses your existing Claude Code session, writes artifacts to `.claude-team/runs/<run_id>/`, and keeps optional MCP integrations explicit.

Try the dashboard without a real run:

```bash
anton --demo
```

I would love feedback on onboarding and whether the workflow feels useful or too heavy.

## Gallery Checklist

- Dashboard screenshot with a selected run
- Inspector screenshot showing agent output
- Deliverables tab screenshot
- Short demo GIF under 90 seconds
