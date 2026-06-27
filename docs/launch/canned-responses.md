# Canned Responses

Use these as starting points. Personalize before posting.

## Why not CrewAI or AutoGen?

Anton is not trying to be a general Python agent framework. It is specifically a local workflow layer for Claude Code users. The design goal is: no separate API key, no Python runtime, no framework DSL, and visible artifacts for each phase.

## Is this autonomous coding?

No. Anton coordinates specialist prompts and stores their outputs. You still review plans, code, test output, and security findings. It is designed to make the work more inspectable, not to remove judgment.

## Does this cost extra?

Anton itself does not call the Anthropic API. Real runs go through your existing Claude Code session. Optional MCP integrations can require their own provider tokens, such as `BRAVE_API_KEY` or `GITHUB_PERSONAL_ACCESS_TOKEN`.

## Why only a few MCP integrations?

Earlier versions listed more integrations than the repo could verify. The default registry is now intentionally small and npm-verified. Users can add more through a custom `mcp-registry.yaml`.

## How is state stored?

Run state is local SQLite in `.claude-team/state.db`. Agent outputs and checkpoints are written under `.claude-team/runs/<run_id>/`.

## Can I run it without Claude Code?

You can run `anton --demo` to preview the dashboard. Real agent dispatch requires Claude Code.

## Is the dashboard public?

No. Anton binds to `127.0.0.1` and is intended for local developer machines only. Do not expose it to the public internet.
