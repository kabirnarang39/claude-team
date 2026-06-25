#!/usr/bin/env bash
# Seed the Anton dashboard with realistic demo data across all 5 workflows.
# Run after `go run main.go` (or `anton`) has initialized the DB.
# Safe to re-run — uses INSERT OR IGNORE.
set -euo pipefail

DB="${1:-.claude-team/state.db}"
RUNS_DIR=".claude-team/runs"

if [ ! -f "$DB" ]; then
  echo "ERROR: DB not found at $DB — start Anton first: go run main.go"
  exit 1
fi

q() { sqlite3 "$DB" "$@"; }
mkdir -p "$RUNS_DIR"

echo "Seeding demo runs into $DB..."

# ── Helpers ──────────────────────────────────────────────────────────────────
NOW="strftime('%s','now')"

# ── Run 1: JWT Auth (feature-build) ──────────────────────────────────────────
R="demo-jwt-auth-1782385638"; mkdir -p "$RUNS_DIR/$R"
q "INSERT OR IGNORE INTO runs (id,workflow_name,status,started_at,completed_at) VALUES ('$R','feature-build','done',strftime('%s','now')-7200,strftime('%s','now')-3600);"
q "INSERT OR IGNORE INTO phases VALUES ('$R','planning','done',strftime('%s','now')-7200,strftime('%s','now')-6600),('$R','engineering','done',strftime('%s','now')-6600,strftime('%s','now')-4800),('$R','qa','done',strftime('%s','now')-4800,strftime('%s','now')-3900),('$R','devops','done',strftime('%s','now')-3900,strftime('%s','now')-3600);"
q "INSERT OR IGNORE INTO agent_results (run_id,phase_id,agent,status,confidence,summary,deliverables_json,sources_json,concerns_json,questions_json,tests_run,tokens_used,created_at) VALUES
('$R','planning','requirements-analyst','done','high','Defined 12 acceptance criteria for JWT auth: registration, login, RS256 token issuance, refresh rotation (7d TTL), RBAC (3 roles), rate limiting, audit logging.','[\"prd.md\"]','[]','[]','[]','',5100,strftime('%s','now')-7000),
('$R','planning','tech-writer','done','high','Drafted PRD with scope boundaries, sequence diagrams for 4 auth flows.','[\"api-spec.yaml\"]','[]','[]','[]','',3200,strftime('%s','now')-6800),
('$R','engineering','senior-architect','done','high','Stateless JWT (RS256) with Redis refresh token store. Asymmetric keys for multi-service verification.','[\"architecture.md\"]','[]','[]','[]','',6800,strftime('%s','now')-6200),
('$R','engineering','backend-engineer-1','done','high','JWT service, Redis token store, rate limiter, all 5 auth endpoints. 47 unit tests.','[]','[]','[]','[]','47/47',8900,strftime('%s','now')-5600),
('$R','engineering','frontend-engineer','done','high','Login/register/refresh flows. Auth context hook. Protected route wrapper. httpOnly cookie.','[]','[]','[]','[]','',5400,strftime('%s','now')-5400),
('$R','engineering','dba','done','high','Schema: users, refresh_tokens, roles, user_roles, audit_log. Indices on email and token_hash.','[\"schema.sql\"]','[]','[]','[]','',3900,strftime('%s','now')-5200),
('$R','qa','qa-engineer','done','high','34 integration tests: happy path, expired tokens, tampered signatures, RBAC boundaries. All passing.','[\"test-report.md\"]','[]','[]','[]','34/34',7200,strftime('%s','now')-4600),
('$R','qa','security-reviewer','done','high','OWASP Top 10 audit. No criticals. Two mediums: PKCE, bcrypt cost factor 10→12.','[\"security-review.md\"]','[]','[\"bcrypt cost factor should be 12 not 10\"]','[]','',5800,strftime('%s','now')-4400),
('$R','qa','e2e-tester','done','high','8 Playwright E2E scenarios. All passing.','[]','[]','[]','[]','8/8 e2e',4600,strftime('%s','now')-4200),
('$R','devops','devops-engineer','done','high','Multi-stage Dockerfile, docker-compose+Redis, GitHub Actions CI, Helm chart. Secrets via env vars only.','[\"deployment-notes.md\"]','[]','[]','[]','',4300,strftime('%s','now')-3700);"
echo "  ✓ jwt-auth (feature-build)"

# ── Run 2: Stripe Webhook Retry (feature-build) ───────────────────────────────
R="demo-stripe-webhook-1782300000"; mkdir -p "$RUNS_DIR/$R"
q "INSERT OR IGNORE INTO runs (id,workflow_name,status,started_at,completed_at) VALUES ('$R','feature-build','done',strftime('%s','now')-86400-3200,strftime('%s','now')-86400-200);"
q "INSERT OR IGNORE INTO phases VALUES ('$R','planning','done',strftime('%s','now')-86400-3200,strftime('%s','now')-86400-2700),('$R','engineering','done',strftime('%s','now')-86400-2700,strftime('%s','now')-86400-900),('$R','qa','done',strftime('%s','now')-86400-900,strftime('%s','now')-86400-300),('$R','devops','done',strftime('%s','now')-86400-300,strftime('%s','now')-86400-200);"
q "INSERT OR IGNORE INTO agent_results (run_id,phase_id,agent,status,confidence,summary,deliverables_json,sources_json,concerns_json,questions_json,tests_run,tokens_used,created_at) VALUES
('$R','planning','requirements-analyst','done','high','Scoped retry logic: idempotency via Stripe event_id, 5-attempt exponential backoff, dead-letter queue.','[\"prd.md\"]','[]','[]','[]','',5100,strftime('%s','now')-86400-2900),
('$R','planning','tech-writer','done','high','API spec drafted for webhook endpoint and retry status query endpoint.','[\"api-spec.yaml\"]','[]','[]','[]','',3200,strftime('%s','now')-86400-2800),
('$R','engineering','senior-architect','done','high','DB-backed retry queue with background goroutine. ON CONFLICT idempotency.','[\"architecture.md\"]','[]','[]','[]','',6800,strftime('%s','now')-86400-2200),
('$R','engineering','backend-engineer-1','done','high','Webhook handler with HMAC validation and idempotent event insertion.','[]','[]','[]','[]','',8900,strftime('%s','now')-86400-1800),
('$R','engineering','backend-engineer-2','done','high','Retry worker goroutine with exponential backoff and context cancellation.','[]','[]','[]','[]','',7600,strftime('%s','now')-86400-1700),
('$R','engineering','backend-engineer-3','done','high','Dead-letter queue handler with PagerDuty alert on 5th failure.','[]','[]','[]','[]','',6400,strftime('%s','now')-86400-1600),
('$R','engineering','dba','done','high','Schema with retry_log, partial index on pending events, JSONB payload column.','[\"schema.sql\"]','[]','[]','[]','',3900,strftime('%s','now')-86400-1500),
('$R','qa','qa-engineer','done','high','63 unit + 18 integration tests pass. 10k events/min load: 99.8% within 60s.','[\"test-report.md\"]','[]','[]','[]','63/63',7200,strftime('%s','now')-86400-600),
('$R','qa','security-reviewer','done','high','APPROVED. HMAC timing-safe, replay window enforced. PII encryption recommended.','[\"security-review.md\"]','[]','[\"PII in payload — recommend at-rest encryption\"]','[]','',5800,strftime('%s','now')-86400-500),
('$R','qa','e2e-tester','done','high','Real Stripe sandbox events processed correctly across all retry scenarios.','[]','[]','[]','[]','18/18 e2e',4600,strftime('%s','now')-86400-400),
('$R','devops','code-reviewer','done','high','APPROVED. Minor: goroutine leak on ctx cancel — fixed before merge.','[\"code-review.md\"]','[]','[]','[]','',4300,strftime('%s','now')-86400-280),
('$R','devops','devops-engineer','done','high','Zero-downtime migration. Feature-flagged rollout. Env vars documented.','[\"deployment-notes.md\"]','[]','[]','[]','',3100,strftime('%s','now')-86400-220);"
echo "  ✓ stripe-webhook-retry (feature-build)"

# ── Run 3: Rate Limiting (feature-build, running) ────────────────────────────
R="demo-rate-limiting-1782360000"; mkdir -p "$RUNS_DIR/$R"
q "INSERT OR IGNORE INTO runs (id,workflow_name,status,started_at,completed_at) VALUES ('$R','feature-build','running',strftime('%s','now')-3800,NULL);"
q "INSERT OR IGNORE INTO phases VALUES ('$R','planning','done',strftime('%s','now')-3800,strftime('%s','now')-3200),('$R','engineering','done',strftime('%s','now')-3200,strftime('%s','now')-1400),('$R','qa','running',strftime('%s','now')-1400,NULL),('$R','devops','pending',NULL,NULL);"
q "INSERT OR IGNORE INTO agent_results (run_id,phase_id,agent,status,confidence,summary,deliverables_json,sources_json,concerns_json,questions_json,tests_run,tokens_used,created_at) VALUES
('$R','planning','requirements-analyst','done','high','Sliding window rate limiting: 100 req/min default, 1000 premium, per-IP and per-user.','[\"prd.md\"]','[]','[]','[]','',4800,strftime('%s','now')-3500),
('$R','planning','tech-writer','done','high','OpenAPI updated with 429 responses and X-RateLimit-* header schemas.','[\"api-spec.yaml\"]','[]','[]','[]','',2900,strftime('%s','now')-3400),
('$R','engineering','senior-architect','done','high','Redis ZADD sliding window primary, in-memory token bucket fallback. YAML config per endpoint group.','[\"architecture.md\"]','[]','[]','[]','',7100,strftime('%s','now')-2900),
('$R','engineering','backend-engineer-1','done','high','Redis sliding window: ZADD/ZREMRANGEBYSCORE/ZCARD in MULTI/EXEC. p99 overhead: 1.3ms.','[]','[]','[]','[]','',9200,strftime('%s','now')-2200),
('$R','engineering','backend-engineer-2','done','high','HTTP middleware injecting X-RateLimit-* headers. 429 with Retry-After on limit breach.','[]','[]','[]','[]','',6700,strftime('%s','now')-2100),
('$R','engineering','backend-engineer-3','done','high','YAML config loader for per-route limits. Hot-reload via SIGHUP.','[]','[]','[]','[]','',5400,strftime('%s','now')-2000),
('$R','engineering','dba','done','high','Redis key schema: ratelimit:{user_id}:{window}. TTL = window_ms.','[]','[]','[]','[]','',2600,strftime('%s','now')-1900),
('$R','qa','qa-engineer','running','','Running 54 unit tests + load simulation (10k req burst)...','[]','[]','[]','[]','',0,strftime('%s','now')-1200),
('$R','qa','security-reviewer','running','','Reviewing Redis key enumeration risk and IP spoofing via X-Forwarded-For...','[]','[]','[]','[]','',0,strftime('%s','now')-1100);"
echo "  ✓ rate-limiting (feature-build, running)"

# ── Run 4: Query Optimisation (feature-build) ─────────────────────────────────
R="demo-query-optimization-1782200000"; mkdir -p "$RUNS_DIR/$R"
q "INSERT OR IGNORE INTO runs (id,workflow_name,status,started_at,completed_at) VALUES ('$R','feature-build','done',strftime('%s','now')-172800-4100,strftime('%s','now')-172800-400);"
q "INSERT OR IGNORE INTO phases VALUES ('$R','planning','done',strftime('%s','now')-172800-4100,strftime('%s','now')-172800-3500),('$R','engineering','done',strftime('%s','now')-172800-3500,strftime('%s','now')-172800-1600),('$R','qa','done',strftime('%s','now')-172800-1600,strftime('%s','now')-172800-700),('$R','devops','done',strftime('%s','now')-172800-700,strftime('%s','now')-172800-400);"
q "INSERT OR IGNORE INTO agent_results (run_id,phase_id,agent,status,confidence,summary,deliverables_json,sources_json,concerns_json,questions_json,tests_run,tokens_used,created_at) VALUES
('$R','planning','requirements-analyst','done','high','Perf baseline: 3 queries at 4-12s. Target: all <100ms. Root cause: missing indexes and N+1.','[\"prd.md\"]','[]','[]','[]','',4400,strftime('%s','now')-172800-3900),
('$R','planning','tech-writer','done','high','Slow query report with EXPLAIN ANALYZE for all 3 offenders.','[\"slow-query-report.md\"]','[]','[]','[]','',3700,strftime('%s','now')-172800-3700),
('$R','engineering','senior-architect','done','high','Composite index strategy + monthly partitioning plan for transactions table.','[\"architecture.md\"]','[]','[]','[]','',6200,strftime('%s','now')-172800-3100),
('$R','engineering','dba','done','high','3 CONCURRENT index migrations. Monthly range partition on transactions. All zero-downtime.','[]','[]','[]','[]','',5100,strftime('%s','now')-172800-2800),
('$R','engineering','backend-engineer-1','done','high','N+1 dashboard query rewritten to single CTE. Invoice history uses new index.','[]','[]','[]','[]','',7800,strftime('%s','now')-172800-2400),
('$R','engineering','backend-engineer-2','done','high','Redis result cache with 30s TTL for monthly summary. Cache-aside, write-through on mutation.','[]','[]','[]','[]','',5900,strftime('%s','now')-172800-2200),
('$R','qa','qa-engineer','done','high','31/31 unit + 147/147 regression tests pass. Data integrity verified.','[\"test-report.md\"]','[]','[]','[]','31/31 unit',6100,strftime('%s','now')-172800-1200),
('$R','qa','e2e-tester','done','high','350x improvement on invoice history. 1087x on monthly summary. p99 from 14.1s to 47ms.','[\"perf-results.md\"]','[]','[]','[]','load test passed',5300,strftime('%s','now')-172800-900),
('$R','devops','code-reviewer','done','high','APPROVED. CONCURRENT index creation correct. No lock contention risk.','[\"code-review.md\"]','[]','[]','[]','',3800,strftime('%s','now')-172800-600),
('$R','devops','devops-engineer','done','high','Staged migration: indexes first (0 downtime), partition in maintenance window.','[\"deployment-notes.md\"]','[]','[]','[]','',3200,strftime('%s','now')-172800-450);"
echo "  ✓ query-optimisation (feature-build)"

# ── Run 5: Memory Leak Bug Fix (bug-fix) ─────────────────────────────────────
R="demo-bug-memory-leak-1782280000"; mkdir -p "$RUNS_DIR/$R"
q "INSERT OR IGNORE INTO runs (id,workflow_name,status,started_at,completed_at) VALUES ('$R','bug-fix','done',strftime('%s','now')-259200-2100,strftime('%s','now')-259200-900);"
q "INSERT OR IGNORE INTO phases VALUES ('$R','triage','done',strftime('%s','now')-259200-2100,strftime('%s','now')-259200-1600),('$R','fix','done',strftime('%s','now')-259200-1600,strftime('%s','now')-259200-1100),('$R','verify','done',strftime('%s','now')-259200-1100,strftime('%s','now')-259200-900);"
q "INSERT OR IGNORE INTO agent_results (run_id,phase_id,agent,status,confidence,summary,deliverables_json,sources_json,concerns_json,questions_json,tests_run,tokens_used,created_at) VALUES
('$R','triage','debugger','done','high','Goroutine leak: time.NewTicker not stopped on ctx cancel. Frontend: setInterval not cleared on unmount.','[\"rca.md\"]','[]','[]','[]','',6800,strftime('%s','now')-259200-1800),
('$R','fix','backend-engineer','done','high','Added defer ticker.Stop() and ctx.Done() select. Zero goroutine leak confirmed via pprof.','[]','[]','[]','[]','',5200,strftime('%s','now')-259200-1300),
('$R','fix','frontend-engineer','done','high','useEffect cleanup returns clearInterval. No interval accumulation confirmed with DevTools.','[]','[]','[]','[]','',3100,strftime('%s','now')-259200-1200),
('$R','verify','qa-engineer','done','high','Memory flat after 2h load test. Goroutines stable at 12 (was growing to 800+). 89/89 pass.','[\"qa-report.md\"]','[]','[]','[]','89/89',4900,strftime('%s','now')-259200-950);"
echo "  ✓ memory-leak (bug-fix)"

# ── Run 6: Payment API Incident (incident-response) ──────────────────────────
R="demo-incident-payment-1782150000"; mkdir -p "$RUNS_DIR/$R"
q "INSERT OR IGNORE INTO runs (id,workflow_name,status,started_at,completed_at) VALUES ('$R','incident-response','done',strftime('%s','now')-345600-2800,strftime('%s','now')-345600-400);"
q "INSERT OR IGNORE INTO phases VALUES ('$R','rca','done',strftime('%s','now')-345600-2800,strftime('%s','now')-345600-2300),('$R','hotfix','done',strftime('%s','now')-345600-2300,strftime('%s','now')-345600-2000),('$R','verify','done',strftime('%s','now')-345600-2000,strftime('%s','now')-345600-1300),('$R','postmortem','done',strftime('%s','now')-345600-1300,strftime('%s','now')-345600-400);"
q "INSERT OR IGNORE INTO agent_results (run_id,phase_id,agent,status,confidence,summary,deliverables_json,sources_json,concerns_json,questions_json,tests_run,tokens_used,created_at) VALUES
('$R','rca','debugger','done','high','Root cause: Stripe key rotated at 15:40 UTC, not propagated to AWS Secrets Manager. All pods had revoked key.','[\"rca.md\"]','[]','[]','[]','',7100,strftime('%s','now')-345600-2500),
('$R','hotfix','backend-engineer','done','high','Updated STRIPE_SECRET_KEY in Secrets Manager. Rolling restart complete. Payments restored in 2 min.','[]','[]','[]','[]','',3800,strftime('%s','now')-345600-2100),
('$R','verify','qa-engineer','done','high','Payment success rate 100%. No dropped transactions during restart.','[\"qa-report.md\"]','[]','[]','[]','smoke tests pass',4200,strftime('%s','now')-345600-1600),
('$R','verify','security-reviewer','done','high','New key scopes correct. Old key fully revoked. No replay risk.','[\"security-report.md\"]','[]','[]','[]','',3600,strftime('%s','now')-345600-1500),
('$R','postmortem','tech-writer','done','high','Postmortem: 14 min outage. 4 action items — automate key rotation, add healthcheck, canary deploy.','[\"postmortem.md\"]','[]','[]','[]','',5400,strftime('%s','now')-345600-600);"
echo "  ✓ payment-incident (incident-response)"

# ── Run 7: Billing Refactor Code Review (code-review) ────────────────────────
R="demo-code-review-billing-1782100000"; mkdir -p "$RUNS_DIR/$R"
q "INSERT OR IGNORE INTO runs (id,workflow_name,status,started_at,completed_at) VALUES ('$R','code-review','done',strftime('%s','now')-432000-3600,strftime('%s','now')-432000-1800);"
q "INSERT OR IGNORE INTO phases VALUES ('$R','review','done',strftime('%s','now')-432000-3600,strftime('%s','now')-432000-2200),('$R','verdict','done',strftime('%s','now')-432000-2200,strftime('%s','now')-432000-1800);"
q "INSERT OR IGNORE INTO agent_results (run_id,phase_id,agent,status,confidence,summary,deliverables_json,sources_json,concerns_json,questions_json,tests_run,tokens_used,created_at) VALUES
('$R','review','senior-architect','done','high','APPROVE. Clean domain separation. BillingManager 1847→94 lines. Recommend StripeAdapter interface.','[\"architecture-review.md\"]','[]','[\"Consider StripeAdapter for testability\"]','[]','',8200,strftime('%s','now')-432000-2800),
('$R','review','security-reviewer','done','high','APPROVE. Auth checks preserved, no new injection vectors. Minor: PII filter tracking.','[\"security-report.md\"]','[]','[]','[]','',5900,strftime('%s','now')-432000-2700),
('$R','review','code-reviewer','done','high','APPROVE with fix: swallowed error at payment_service.go:134. Minor: magic number, missing godoc.','[\"review-report.md\"]','[]','[\"Swallowed error must be fixed before merge\"]','[]','',6700,strftime('%s','now')-432000-2600),
('$R','verdict','tech-writer','done','high','Consolidated verdict: APPROVE, one required fix (swallowed error). Two minor follow-ups.','[\"verdict.md\"]','[]','[]','[]','',4100,strftime('%s','now')-432000-1900);"
echo "  ✓ billing-refactor (code-review)"

# ── Run 8: Search Architecture Review (architecture-review) ──────────────────
R="demo-arch-review-search-1782050000"; mkdir -p "$RUNS_DIR/$R"
q "INSERT OR IGNORE INTO runs (id,workflow_name,status,started_at,completed_at) VALUES ('$R','architecture-review','done',strftime('%s','now')-518400-5200,strftime('%s','now')-518400-2800);"
q "INSERT OR IGNORE INTO phases VALUES ('$R','review','done',strftime('%s','now')-518400-5200,strftime('%s','now')-518400-3600),('$R','adr','done',strftime('%s','now')-518400-3600,strftime('%s','now')-518400-2800);"
q "INSERT OR IGNORE INTO agent_results (run_id,phase_id,agent,status,confidence,summary,deliverables_json,sources_json,concerns_json,questions_json,tests_run,tokens_used,created_at) VALUES
('$R','review','senior-architect','done','high','APPROVE with modifications. ES for invoice/customer search. Mandate tsvector benchmark first.','[\"architecture-review.md\"]','[]','[\"Benchmark tsvector before ES commit\"]','[]','',9100,strftime('%s','now')-518400-4600),
('$R','review','security-reviewer','done','high','CONDITIONAL APPROVE. Multi-tenant isolation and PII field security required before implementation.','[\"security-report.md\"]','[]','[\"Multi-tenant isolation is critical\",\"PII encryption mandatory\"]','[]','',7300,strftime('%s','now')-518400-4400),
('$R','adr','senior-architect','done','high','ADR-014: ES accepted for invoice/customer search. tsvector rejected. Typesense deferred to Q3.','[\"adr.md\"]','[]','[]','[]','',6800,strftime('%s','now')-518400-3000);"
echo "  ✓ search-architecture (architecture-review)"

echo ""
echo "Done. 8 runs seeded across all 5 workflows."
echo "Open http://localhost:3000"
