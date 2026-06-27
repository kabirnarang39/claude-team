# Changelog

All notable changes to this project are documented here.

Format: [Keep a Changelog](https://keepachangelog.com). Versioning: [Semantic Versioning](https://semver.org).

## [Unreleased]

### Changed
- Tightened README claims around default agents, MCP integrations, install behavior, and known limitations.
- Reduced the default MCP registry to npm-verified packages and skipped entries with missing required environment variables.
- Added launch assets and inspectable example outputs for prospective users.

## [1.4.5] - 2026-06-26

### Changed
- Refreshed README positioning and dashboard screenshots.

## [1.4.4] - 2026-06-26

### Changed
- Cleaned up code and release assets.

## [1.4.3] - 2026-06-26

### Changed
- Unified dashboard and landing-page styling around the amber/navy visual system.

## [1.4.2] - 2026-06-26

### Added
- Dashboard design-system pass with improved empty states, inspector polish, and public landing-page presentation.

## [1.4.1] - 2026-06-26

### Fixed
- Deduplicated token counts in reconciliation scripts and stats reporting.

## [1.4.0] - 2026-06-26

### Added
- Anti-hallucination framework in role standards.
- OWASP-oriented security reviewer flow.
- Security fix loop for high-risk findings.

## [1.3.1] - 2026-06-26

### Fixed
- Token reconciliation in workflow reporting.
- README refresh for v1.3.0 behavior.

## [1.3.0] - 2026-06-26

### Added
- Agent planning with skipped-agent status.
- Phase review gates and retry flow.
- Token reconciliation across workflows.

## [1.2.0] - 2026-06-26

### Added
- Human feedback loop.
- TDD enforcement language for engineering agents.
- QA circuit breaker.
- Launch asset scaffolding.

## [1.1.0] - 2026-06-26

### Added
- Resume-oriented checkpoint improvements.
- Dashboard deliverables and run-inspection polish.

## [1.0.0] - 2026-06-20

### Added
- 15 reusable specialist role prompts: requirements-analyst, tech-writer, senior-architect, api-designer, backend-engineer, frontend-engineer, dba, qa-engineer, security-reviewer, e2e-tester, code-reviewer, devops-engineer, debugger, performance-engineer, mobile-engineer.
- 5 workflows: feature-build, code-review, bug-fix, incident-response, architecture-review.
- Go HTTP server with SQLite state and WebSocket live updates.
- Browser dashboard with live agent execution tree.
- Coordinator MCP server for inter-agent communication and result reporting.
- Skills: `/team-dispatch`, `/team-status`, `/team-stop`.
- One-command install script for macOS arm64/amd64 and Linux amd64.
- `--version` flag.
- Agent output file viewer in dashboard.
- Server offline detection banner.
