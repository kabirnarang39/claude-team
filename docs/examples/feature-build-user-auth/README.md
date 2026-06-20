# Example: Feature Build — User Auth with JWT

**Task dispatched:**
```
/team-dispatch build user auth with JWT and refresh tokens
```

These are the actual outputs produced by Anton's 11 specialist agents for this task. Each file is exactly what the agent delivered — no editing.

## Outputs by Phase

### Phase 1 — Planning
| File | Agent | What it contains |
|------|-------|-----------------|
| [requirements-analyst.md](requirements-analyst.md) | `requirements-analyst` | 12 functional requirements, NFRs, edge cases |
| [tech-writer.md](tech-writer.md) | `tech-writer` | PRD with user stories, acceptance criteria, open questions |

### Phase 2 — Architecture
| File | Agent | What it contains |
|------|-------|-----------------|
| [senior-architect.md](senior-architect.md) | `senior-architect` | ADR: JWT vs sessions decision, RS256 rationale, risk table |
| [api-designer.md](api-designer.md) | `api-designer` | 7 endpoints with request/response schemas, rate limit table |

### Phase 3 — Engineering (parallel)
| File | Agent | What it contains |
|------|-------|-----------------|
| [backend-engineer.md](backend-engineer.md) | `backend-engineer` | Go implementation plan, token rotation logic, key decisions |
| [frontend-engineer.md](frontend-engineer.md) | `frontend-engineer` | React component structure, silent refresh, token storage strategy |
| [dba.md](dba.md) | `dba` | Full schema, indices, migration strategy, cleanup jobs |

### Phase 4 — QA (parallel)
| File | Agent | What it contains |
|------|-------|-----------------|
| [qa-engineer.md](qa-engineer.md) | `qa-engineer` | 30+ test cases across all endpoints, integration test setup |
| [security-reviewer.md](security-reviewer.md) | `security-reviewer` | OWASP Top 10 checklist, 4 findings with mitigations |
| [e2e-tester.md](e2e-tester.md) | `e2e-tester` | Playwright test suites, edge cases, CI config |

### Phase 5 — DevOps
| File | Agent | What it contains |
|------|-------|-----------------|
| [devops-engineer.md](devops-engineer.md) | `devops-engineer` | Dockerfile, CI/CD pipeline, Helm values, monitoring alerts, key rotation runbook |

## Total wall-clock time

Phases 3 and 4 run in parallel. Real elapsed time for this task: **~4 minutes** (3 engineers simultaneously + 3 QA agents simultaneously).

Sequential equivalent: ~18 minutes.
