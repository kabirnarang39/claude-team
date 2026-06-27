# Hacker News

## Title

Show HN: Anton - a 12-agent workflow dashboard for Claude Code

## Body

I built Anton because my longer Claude Code tasks kept turning into a long single conversation: requirements mixed with architecture, implementation mixed with QA, and no easy way to inspect what happened afterward.

Anton is a local Go app that adds a repeatable 12-agent workflow around Claude Code:

- planning -> architecture -> engineering -> QA/security -> DevOps
- workflows are YAML
- role prompts are Markdown
- state is SQLite
- progress streams to a local browser dashboard
- outputs are saved under `.claude-team/runs/<run_id>/`

It is not a hosted autonomous dev platform and it does not call the Anthropic API directly. It runs through your existing Claude Code session. Optional MCP presets are deliberately small and npm-verified; entries with missing env vars are skipped instead of writing broken config.

Repo: https://github.com/kabirnarang39/claude-team

Install:

```bash
curl -fsSL https://raw.githubusercontent.com/kabirnarang39/claude-team/main/install.sh | bash
```

Try the dashboard without spending tokens:

```bash
anton --demo
```

Feedback I would value most: is the YAML workflow shape useful, and does the dashboard make multi-agent work easier to trust or just easier to watch?
