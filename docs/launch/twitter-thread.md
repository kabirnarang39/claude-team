# X / Twitter Thread

## Thread

1. I built Anton: a local 12-agent workflow dashboard for Claude Code.

2. The problem: big Claude Code tasks often become one long conversation. Planning, architecture, implementation, QA, and review blur together.

3. Anton splits that into a repeatable workflow: planning -> architecture -> engineering -> QA/security -> DevOps.

4. The stack is intentionally small: Go server, SQLite state, WebSocket updates, vanilla JS dashboard, Node MCP bridge.

5. Workflows are YAML. Agent roles are Markdown. Outputs are saved under `.claude-team/runs/<run_id>/`.

6. It does not call Anthropic APIs directly. It runs inside your existing Claude Code session.

7. Optional MCP presets are deliberately small and npm-verified. Missing env vars are skipped instead of writing broken config.

8. Try the dashboard without spending tokens:

```bash
anton --demo
```

9. Repo:
https://github.com/kabirnarang39/claude-team

10. I am especially interested in feedback on the workflow format and whether the dashboard makes agent work easier to trust.
