# Demo Recording Script

Use this to record a short launch GIF or video without relying on a live Claude Code run.

## Setup

```bash
go run main.go --demo --port 3100
```

Open `http://localhost:3100`.

## Shot List

1. Start on the empty/dashboard view and show the live connection indicator.
2. Select the `demo-feature-build-jwt-auth` run.
3. Click through the phase timeline.
4. Click `requirements-analyst`, `senior-architect`, `security-reviewer`, and `code-reviewer`.
5. Open the Docs tab and show PRD, ADR, QA, security, and review content.
6. Open the Deliverables tab and show the seeded files under `.claude-team/runs/demo-feature-build-jwt-auth/`.
7. End on the run summary and the README install command.

## Voiceover

```text
Anton is a local coordinator for Claude Code. One slash command starts a 12-agent workflow, stores every result in SQLite and Markdown, and streams progress to this dashboard. It is not a hosted agent platform; it is a way to make multi-step Claude Code work visible and repeatable.
```

## Notes

- Keep the video under 90 seconds.
- Do not claim speed benchmarks unless a real timed run is shown.
- Do not claim external integrations unless the relevant env vars are configured during the recording.
