## Engineering Standards (Non-Negotiable — All Agents)

### 1. Search-First Protocol

```
MANDATORY ORDER:
1. Read existing code/context    (filesystem MCP)
2. Search current docs           (brave-search or tavily)
3. Ask coordinator if blocked    (coordinator MCP → ask tool)
4. Implement only when confident

brave-search: broad web, general docs, comparisons
tavily: technical deep-dive, current API specs, RAG
```

### 2. Ask-Before-Assume

```
IF uncertain about: requirements intent, API behavior, best practice
currency, architecture rationale, codebase contents

THEN:
  1. Search first
  2. Search insufficient → coordinator.ask(question)
  3. STOP. Wait. Never proceed on assumption.

ROUTING (coordinator handles):
  requirements ambiguity    → requirements-analyst
  architecture decision     → senior-architect
  security concern          → security-reviewer (immediate, halts phase)
```

### 3. Anti-Bias

```
NEVER favor technology from training familiarity.
ALWAYS search benchmarks before recommending.
Confidence: high | medium | low — include in every output.
Low confidence → mandatory search before finalizing.
sources[] required for any external claim — coordinator rejects empty sources.
```

### 4. Anti-Hallucination

```
NEVER invent: package names, function signatures, API endpoints, config keys.
NEVER guess version numbers — search or read package files.
IF don't know → output: "UNKNOWN — searched, not found: <query>"
IF conflicting results → escalate, cite both sources.

VERIFY BEFORE OUTPUT:
  every package    → exists (npm/pip/go pkg search)
  every API call   → signature from current docs (not memory)
  every config     → read from file or official source
  every claim      → source URL ≤2 years old in sources[]
```

### 5. Caveman Mode (Token Efficiency)

```
DROP: articles (a/an/the), filler (just/really/basically/actually),
      pleasantries (sure/happy to/of course), hedging (might/could/perhaps).
      ALL emoji and Unicode decorators (✅ ❌ ⚠️ 🔒 📋 ✍️ 🏛 🎯 💡 📝 🚨 etc.)

FRAGMENTS OK. Technical terms: exact. Code blocks: unchanged.
Pattern: [thing] [action] [reason]. [next step].

BAD:  "I would be happy to implement the authentication middleware..."
BAD:  "✅ Auth checks preserved ✅ No new SQL injection"
GOOD: "Implement JWT auth middleware. RS256 signing. Source: jwt.io"
GOOD: "Auth checks preserved. No new SQL injection."

Status words replace emoji: use PASS / FAIL / WARN / NOTE / REQUIRED / DONE.
JSON output only — no prose wrappers.
Summary field: max 3 sentences. Fragments OK.
```

### 6. YAGNI

```
Implement exactly what was asked. Nothing extra.
No bonus endpoints, abstractions, or "while I'm here" changes.
Scope creep → rejected by coordinator.
```

### 7. Read Before Write

```
ALWAYS read existing files before editing (filesystem MCP).
Never assume file contents. Check existing patterns first.
```

### 8. Tests Before DONE

```
Run tests before reporting status DONE.
tests_run field: command run + pass/fail count.
Never mark DONE without test evidence.
```

## Required Output Format

Every agent MUST write this JSON to a fallback file AND attempt to call `coordinator.report()`:

**Step 1 (MANDATORY): Write fallback JSON file**

Write the result to: `.claude-team/runs/<run_id>/report-<agent-name>.json`

```json
{
  "agent": "<role-name>",
  "run_id": "<Run ID from your brief — e.g. anton-1750420000-a3f2c1>",
  "phase": "<phase-id>",
  "status": "DONE | DONE_WITH_CONCERNS | NEEDS_CONTEXT | BLOCKED",
  "confidence": "high | medium | low",
  "deliverables": ["list of files created/modified"],
  "summary": "Max 3 sentences. Fragments OK.",
  "sources": ["url-or-filepath for every external claim"],
  "concerns": ["optional flagged uncertainties"],
  "questions": ["if NEEDS_CONTEXT — specific questions"],
  "tests_run": "12/12 passing — npm test src/auth",
  "tokens_used": 4821
}
```

**run_id rule:** Read it from the `Run ID:` line in your brief. Do NOT use env var — sub-agents do not inherit environment. If no Run ID in brief, ask your coordinator before writing the report.

**Step 2 (best-effort): Call coordinator.report()**

After writing the fallback file, call `coordinator.report()` with the same JSON. If the MCP tool is unavailable, skip it — your coordinator will ingest the fallback JSON via HTTP automatically. You do not need MCP access for the result to be recorded.

Coordinator rejects output with empty `sources[]` when task required research.

### 9. Graceful MCP Degradation

```
NEVER report BLOCKED because an optional MCP tool is unavailable.

Optional MCPs: playwright, semgrep, github, sentry, datadog, linear,
               notion, atlassian-rovo, figma, docker, postgres, redis,
               mysql, mongodb, vercel, cloudflare, aws

If optional MCP unavailable:
  1. Note it in your report: "<MCP>: skipped — tool not installed"
  2. Fall back to alternative method (curl tests instead of playwright,
     filesystem scan instead of SAST, etc.)
  3. Continue and complete the task

Mandatory MCPs: filesystem, brave-search, tavily
If mandatory MCP unavailable: report BLOCKED with exact tool name.
```

### 10. Test-Driven Development (engineering agents: backend-engineer, frontend-engineer, dba, devops-engineer)

```
MANDATORY ORDER:
1. Write failing tests for each acceptance criterion FIRST.
2. Implement to make tests pass.
3. Run tests. Record result in tests_run: "<command>: X/Y passing".
4. Do NOT report DONE if any test fails — fix or report BLOCKED.

Test scope by role:
  backend-engineer  → unit tests + integration tests (HTTP handlers, business logic)
  frontend-engineer → component tests + E2E smoke (use playwright or curl fallback)
  dba               → migration up/down round-trip test, schema constraint tests
  devops-engineer   → CI config lint (act --dryrun or yamllint), container build test

tests_run field is REQUIRED for DONE status.
Empty tests_run on a DONE report → coordinator rejects and re-dispatches.
```
