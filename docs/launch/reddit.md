# Reddit

## Suggested communities

- r/ClaudeAI
- r/LocalLLaMA, only if positioned as local orchestration rather than model hosting
- r/programming, only after the install and README are stable

## Post

Title:

```text
I built Anton: a local 12-agent workflow dashboard for Claude Code
```

Body:

```text
I have been using Claude Code for larger tasks, but I kept losing the shape of the work in one long conversation. Anton is my attempt to make that work more explicit.

It wraps Claude Code with:

- a 12-agent default workflow
- YAML workflows
- Markdown role prompts
- SQLite run state
- a local browser dashboard
- saved artifacts in .claude-team/runs/<run_id>/

It does not call the Anthropic API directly and it is not an autonomous deployer. You still review the plan, outputs, and code. Optional MCP integrations are small and npm-verified; missing tokens are skipped rather than added as broken config.

Repo: https://github.com/kabirnarang39/claude-team

Try the UI without a real Claude Code run:

anton --demo

I would love feedback on whether this workflow model is useful or whether it adds too much ceremony.
```

## Comment to add if asked about cost

Anton itself does not make Anthropic API calls. It orchestrates Claude Code, so usage goes through your existing Claude Code plan. Optional external MCP tools may require their own provider tokens.
