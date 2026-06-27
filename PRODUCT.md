# Product

## Register

product

## Users

Claude Code users who want larger coding tasks to be easier to inspect, review, and resume. They are usually working locally in an existing repository and need visible workflow state more than a hosted agent platform.

## Product Purpose

Anton adds a small local workflow layer around Claude Code. It starts structured runs from a slash command, stores workflow state locally, and shows progress in a browser dashboard so planning, architecture, engineering, QA/security, and DevOps passes are visible.

## Brand Personality

Direct, developer-native, and skeptical of magic. Anton should feel practical, local-first, and review-heavy rather than autonomous, inflated, or enterprise-polished.

## Anti-references

Avoid generic AI-agent hype, fake social proof, inflated autonomy claims, benchmark claims without proof, hosted-control-plane language, and framework-heavy positioning that sounds like LangChain or AutoGen. The interface should not look like a decorative SaaS landing page when the user is trying to inspect a run.

## Design Principles

- Show the work: make agent outputs, artifacts, and state transitions inspectable.
- Keep trust local: be explicit about SQLite state, localhost serving, and files Anton writes.
- Prefer boring mechanics over magic: YAML workflows, Markdown roles, Go, SQLite, WebSocket, vanilla JS.
- Make review feel first-class: planning, QA, security, and code review should be visible parts of the flow.
- Stay honest about limits: demo data, real Claude Code requirements, and non-guarantees should be clear.

## Accessibility & Inclusion

Aim for WCAG AA contrast and keyboard-reachable controls. Motion should respect reduced-motion preferences, and status should not rely on color alone where text labels can carry the same meaning.
