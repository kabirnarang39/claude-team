# Changelog

All notable changes to this project will be documented here.
Format: [Keep a Changelog](https://keepachangelog.com). Versioning: [Semantic Versioning](https://semver.org).

## [1.0.0] — 2026-06-20

### Added
- 14 specialist agent roles: requirements-analyst, tech-writer, senior-architect, api-designer, backend-engineer, frontend-engineer, dba, qa-engineer, security-reviewer, e2e-tester, code-reviewer, devops-engineer, debugger, performance-engineer
- 5 workflows: feature-build, code-review, bug-fix, incident-response, architecture-review
- Go HTTP server with SQLite state and WebSocket live updates
- Browser dashboard with live agent execution tree
- Coordinator MCP server for inter-agent communication and result reporting
- Skills: /team-dispatch, /team-status, /team-stop
- One-command install script for macOS arm64/amd64 and Linux amd64
- `--version` flag
- Agent output file viewer in dashboard
- Server offline detection banner
