# Anton: Coordinator Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace terminal-spawning multi-agent architecture with a Coordinator pattern using Claude Code's Agent tool — rename project to Anton, eliminate AppleScript/terminal windows, implement 20 specialist roles with anti-hallucination standards, slim Go backend to pure dashboard server.

**Architecture:** Current Claude Code session becomes the Coordinator; sub-agents spawned via Agent tool with fresh isolated context. Coordinator MCP writes structured JSON to SQLite. Go backend reads SQLite, broadcasts via WebSocket, serves browser dashboard. Browser shows live agent tree SVG.

**Tech Stack:** Go 1.21+ (net/http, modernc.org/sqlite), Node.js 18+ + @modelcontextprotocol/sdk + better-sqlite3 (coordinator MCP), Vanilla HTML/CSS/JS (zero build step)

## Global Constraints

- Tool name: `Anton` — all references to "claude-team" in user-facing strings replaced with "Anton"
- All role prompts embed all 8 engineering standards verbatim from `roles/_standards.md`
- All agents receive `brave-search` + `tavily` in their MCP list (mandatory, hardcoded)
- All agent output is structured JSON per AgentResult schema (see Task 1)
- All agent communication uses caveman mode (no articles/filler/hedging/pleasantries)
- SQLite WAL mode enabled — shared between Go (modernc.org/sqlite) and Node.js (better-sqlite3)
- Go backend spawns zero processes — no AppleScript, no osascript, no terminal windows
- DB path: `.claude-team/state.db` (unchanged — coordinator MCP already uses this path)
- Runtime dir: `.claude-team/` (unchanged)
- No backwards-compatibility shims — delete dead code cleanly

---

## File Map

### Create
```
roles/_standards.md
coordinators/main.md
coordinators/planning.md
coordinators/engineering.md
coordinators/qa.md
coordinators/devops.md
roles/requirements-analyst.md
roles/tech-writer.md
roles/api-designer.md
roles/dba.md
roles/mobile-engineer.md
roles/e2e-tester.md
roles/devops-engineer.md
roles/performance-engineer.md
roles/code-reviewer.md
roles/debugger.md
workflows/feature-build.yaml
workflows/code-review.yaml
workflows/bug-fix.yaml
workflows/incident-response.yaml
workflows/architecture-review.yaml
skills/team-dispatch.md
skills/team-status.md
skills/team-stop.md
internal/store/watcher.go
.claude-team/pending-task.md   (empty placeholder, gitignored)
```

### Modify
```
mcp/team-coordinator.js        add report tool, update ask tool, WAL mode
internal/store/store.go        schema migration v2: phases + agent_results tables
internal/store/runs.go         add new CRUD methods
internal/api/server.go         add /api/task, /api/runs, /api/runs/:id routes; remove dead routes
internal/api/handlers.go       add handleTask, handleRuns, handleRunDetail; remove handleAgentComplete, handleOrchestrate, handleLogsTail
internal/api/ws.go             switch from status.Event to store.Event
main.go                        remove terminal/status imports; wire SQLite watcher; add WAL pragma
ui/index.html                  agent tree layout
ui/app.js                      WebSocket + live agent tree SVG
ui/styles.css                  dark theme + agent tree styles
mcp-registry.yaml              full MCP catalog per spec
roles/senior-architect.md      add standards block + structured output
roles/senior-engineer.md       rename → roles/backend-engineer.md, add standards
roles/qa.md                    rename → roles/qa-engineer.md, add standards
roles/engineer.md              rename → roles/frontend-engineer.md, add standards
```

### Delete
```
internal/terminal/             whole package (spawner.go, spawner_test.go)
internal/status/               whole package (watcher.go, watcher_test.go)
internal/workflow/executor.go
internal/workflow/executor_test.go
```

---

## Task 1: Extend SQLite schema + store methods

**Files:**
- Modify: `internal/store/store.go`
- Modify: `internal/store/runs.go`
- Test: `internal/store/store_test.go`

**Interfaces:**
- Produces: `store.AgentResult` struct, `store.Phase` struct, `store.RunDetail` struct
- Produces: `(*Store).InsertAgentResult`, `(*Store).UpsertPhase`, `(*Store).GetAgentResultsSince`, `(*Store).GetRuns`, `(*Store).GetRunDetail`, `(*Store).WriteTask`

- [ ] **Step 1: Write failing tests for new schema**

```go
// internal/store/store_test.go — add these tests
func TestAgentResultInsertAndFetch(t *testing.T) {
    s := openTestStore(t)
    runID, _ := s.CreateRun("feature-build")

    err := s.UpsertPhase(runID, "planning", "running")
    if err != nil {
        t.Fatal(err)
    }

    result := AgentResult{
        RunID:      runID,
        PhaseID:    "planning",
        Agent:      "requirements-analyst",
        Status:     "DONE",
        Confidence: "high",
        Summary:    "Acceptance criteria written.",
        Sources:    []string{"https://example.com"},
        TestsRun:   "n/a",
        TokensUsed: 1200,
    }
    err = s.InsertAgentResult(result)
    if err != nil {
        t.Fatal(err)
    }

    results, maxID, err := s.GetAgentResultsSince(0)
    if err != nil {
        t.Fatal(err)
    }
    if len(results) != 1 {
        t.Fatalf("want 1 result, got %d", len(results))
    }
    if results[0].Agent != "requirements-analyst" {
        t.Errorf("want requirements-analyst, got %s", results[0].Agent)
    }
    if maxID == 0 {
        t.Error("maxID should be > 0")
    }
}

func TestGetRunDetail(t *testing.T) {
    s := openTestStore(t)
    runID, _ := s.CreateRun("feature-build")
    s.UpsertPhase(runID, "planning", "done")
    s.InsertAgentResult(AgentResult{
        RunID: runID, PhaseID: "planning", Agent: "tech-writer",
        Status: "DONE", Confidence: "high", Summary: "PRD written.",
    })

    detail, err := s.GetRunDetail(runID)
    if err != nil {
        t.Fatal(err)
    }
    if len(detail.Phases) != 1 {
        t.Errorf("want 1 phase, got %d", len(detail.Phases))
    }
    if len(detail.Results) != 1 {
        t.Errorf("want 1 result, got %d", len(detail.Results))
    }
}

func TestWriteTask(t *testing.T) {
    s := openTestStore(t)
    dir := t.TempDir()
    err := s.WriteTask(dir, "build user auth")
    if err != nil {
        t.Fatal(err)
    }
    data, err := os.ReadFile(filepath.Join(dir, "pending-task.md"))
    if err != nil {
        t.Fatal(err)
    }
    if !strings.Contains(string(data), "build user auth") {
        t.Error("pending-task.md missing task text")
    }
}
```

- [ ] **Step 2: Run tests — verify they fail**

```bash
cd /Users/kabir/Workspace/claude-team && go test ./internal/store/... -run "TestAgentResult|TestGetRunDetail|TestWriteTask" -v
```
Expected: compile error or FAIL — types/methods not defined yet.

- [ ] **Step 3: Add types and migration to store.go**

```go
// internal/store/store.go — replace migrate() and add types

type AgentResult struct {
    ID           int64    `json:"id"`
    RunID        string   `json:"run_id"`
    PhaseID      string   `json:"phase_id"`
    Agent        string   `json:"agent"`
    Status       string   `json:"status"`
    Confidence   string   `json:"confidence"`
    Summary      string   `json:"summary"`
    Deliverables []string `json:"deliverables"`
    Sources      []string `json:"sources"`
    Concerns     []string `json:"concerns"`
    Questions    []string `json:"questions"`
    TestsRun     string   `json:"tests_run"`
    TokensUsed   int      `json:"tokens_used"`
    CreatedAt    int64    `json:"created_at"`
}

type Phase struct {
    RunID       string `json:"run_id"`
    PhaseID     string `json:"phase_id"`
    Status      string `json:"status"`
    StartedAt   int64  `json:"started_at"`
    CompletedAt int64  `json:"completed_at"`
}

type RunDetail struct {
    ID           string        `json:"id"`
    WorkflowName string        `json:"workflow_name"`
    Status       string        `json:"status"`
    StartedAt    int64         `json:"started_at"`
    CompletedAt  int64         `json:"completed_at"`
    Phases       []Phase       `json:"phases"`
    Results      []AgentResult `json:"results"`
}

func (s *Store) migrate() error {
    // schema_version table
    s.db.Exec(`CREATE TABLE IF NOT EXISTS schema_version (version INTEGER NOT NULL)`)
    var version int
    s.db.QueryRow(`SELECT COALESCE(MAX(version), 0) FROM schema_version`).Scan(&version)

    if version < 1 {
        _, err := s.db.Exec(`
            CREATE TABLE IF NOT EXISTS runs (
                id TEXT PRIMARY KEY,
                workflow_name TEXT NOT NULL,
                status TEXT NOT NULL DEFAULT 'running',
                started_at INTEGER NOT NULL,
                completed_at INTEGER
            );
            CREATE TABLE IF NOT EXISTS agent_statuses (
                run_id TEXT NOT NULL,
                agent TEXT NOT NULL,
                status TEXT NOT NULL DEFAULT 'pending',
                updated_at INTEGER NOT NULL,
                PRIMARY KEY (run_id, agent)
            );
            CREATE TABLE IF NOT EXISTS messages (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                from_agent TEXT NOT NULL,
                to_agent TEXT NOT NULL,
                content TEXT NOT NULL,
                created_at INTEGER NOT NULL,
                read_at INTEGER
            );
        `)
        if err != nil {
            return err
        }
        s.db.Exec(`INSERT INTO schema_version VALUES (1)`)
    }

    if version < 2 {
        _, err := s.db.Exec(`
            CREATE TABLE IF NOT EXISTS phases (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                run_id TEXT NOT NULL,
                phase_id TEXT NOT NULL,
                status TEXT NOT NULL DEFAULT 'pending',
                started_at INTEGER,
                completed_at INTEGER
            );
            CREATE TABLE IF NOT EXISTS agent_results (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                run_id TEXT NOT NULL,
                phase_id TEXT NOT NULL,
                agent TEXT NOT NULL,
                status TEXT NOT NULL,
                confidence TEXT DEFAULT 'medium',
                summary TEXT DEFAULT '',
                deliverables_json TEXT DEFAULT '[]',
                sources_json TEXT DEFAULT '[]',
                concerns_json TEXT DEFAULT '[]',
                questions_json TEXT DEFAULT '[]',
                tests_run TEXT DEFAULT '',
                tokens_used INTEGER DEFAULT 0,
                created_at INTEGER NOT NULL
            );
        `)
        if err != nil {
            return err
        }
        // Best-effort column additions to messages (ignore errors if columns exist)
        s.db.Exec(`ALTER TABLE messages ADD COLUMN run_id TEXT`)
        s.db.Exec(`ALTER TABLE messages ADD COLUMN response TEXT`)
        s.db.Exec(`INSERT INTO schema_version VALUES (2)`)
    }

    // Enable WAL mode for concurrent Go + Node.js access
    _, err := s.db.Exec(`PRAGMA journal_mode=WAL`)
    return err
}
```

- [ ] **Step 4: Add new CRUD methods to runs.go**

```go
// internal/store/runs.go — append these methods

import (
    "encoding/json"
    "os"
    "path/filepath"
    "time"
)

func (s *Store) UpsertPhase(runID, phaseID, status string) error {
    _, err := s.db.Exec(`
        INSERT INTO phases (run_id, phase_id, status, started_at)
        VALUES (?, ?, ?, ?)
        ON CONFLICT(run_id, phase_id) DO UPDATE SET
            status=excluded.status,
            completed_at=CASE WHEN excluded.status IN ('done','failed') THEN ? ELSE completed_at END
    `, runID, phaseID, status, time.Now().Unix(), time.Now().Unix())
    return err
}

func (s *Store) InsertAgentResult(r AgentResult) error {
    deliverables, _ := json.Marshal(r.Deliverables)
    sources, _ := json.Marshal(r.Sources)
    concerns, _ := json.Marshal(r.Concerns)
    questions, _ := json.Marshal(r.Questions)
    if r.CreatedAt == 0 {
        r.CreatedAt = time.Now().Unix()
    }
    _, err := s.db.Exec(`
        INSERT INTO agent_results
        (run_id, phase_id, agent, status, confidence, summary,
         deliverables_json, sources_json, concerns_json, questions_json,
         tests_run, tokens_used, created_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `, r.RunID, r.PhaseID, r.Agent, r.Status, r.Confidence, r.Summary,
        string(deliverables), string(sources), string(concerns), string(questions),
        r.TestsRun, r.TokensUsed, r.CreatedAt)
    return err
}

func (s *Store) GetAgentResultsSince(lastID int64) ([]AgentResult, int64, error) {
    rows, err := s.db.Query(`
        SELECT id, run_id, phase_id, agent, status, confidence, summary,
               deliverables_json, sources_json, concerns_json, questions_json,
               tests_run, tokens_used, created_at
        FROM agent_results WHERE id > ? ORDER BY id ASC
    `, lastID)
    if err != nil {
        return nil, lastID, err
    }
    defer rows.Close()
    var results []AgentResult
    var maxID int64 = lastID
    for rows.Next() {
        var r AgentResult
        var deliverables, sources, concerns, questions string
        err := rows.Scan(&r.ID, &r.RunID, &r.PhaseID, &r.Agent, &r.Status,
            &r.Confidence, &r.Summary, &deliverables, &sources, &concerns,
            &questions, &r.TestsRun, &r.TokensUsed, &r.CreatedAt)
        if err != nil {
            return nil, maxID, err
        }
        json.Unmarshal([]byte(deliverables), &r.Deliverables)
        json.Unmarshal([]byte(sources), &r.Sources)
        json.Unmarshal([]byte(concerns), &r.Concerns)
        json.Unmarshal([]byte(questions), &r.Questions)
        if r.ID > maxID {
            maxID = r.ID
        }
        results = append(results, r)
    }
    return results, maxID, rows.Err()
}

func (s *Store) GetRuns(limit int) ([]RunDetail, error) {
    rows, err := s.db.Query(`
        SELECT id, workflow_name, status, started_at, COALESCE(completed_at,0)
        FROM runs ORDER BY started_at DESC LIMIT ?
    `, limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var runs []RunDetail
    for rows.Next() {
        var r RunDetail
        rows.Scan(&r.ID, &r.WorkflowName, &r.Status, &r.StartedAt, &r.CompletedAt)
        runs = append(runs, r)
    }
    return runs, rows.Err()
}

func (s *Store) GetRunDetail(runID string) (*RunDetail, error) {
    var r RunDetail
    err := s.db.QueryRow(`
        SELECT id, workflow_name, status, started_at, COALESCE(completed_at,0)
        FROM runs WHERE id = ?
    `, runID).Scan(&r.ID, &r.WorkflowName, &r.Status, &r.StartedAt, &r.CompletedAt)
    if err != nil {
        return nil, err
    }

    phaseRows, err := s.db.Query(`
        SELECT run_id, phase_id, status, COALESCE(started_at,0), COALESCE(completed_at,0)
        FROM phases WHERE run_id = ? ORDER BY id ASC
    `, runID)
    if err != nil {
        return nil, err
    }
    defer phaseRows.Close()
    for phaseRows.Next() {
        var p Phase
        phaseRows.Scan(&p.RunID, &p.PhaseID, &p.Status, &p.StartedAt, &p.CompletedAt)
        r.Phases = append(r.Phases, p)
    }

    results, _, err := s.GetAgentResultsSince(0)
    if err != nil {
        return nil, err
    }
    for _, res := range results {
        if res.RunID == runID {
            r.Results = append(r.Results, res)
        }
    }
    return &r, nil
}

func (s *Store) WriteTask(runtimeDir, text string) error {
    content := "# Pending Task\n\n" + text + "\n"
    return os.WriteFile(filepath.Join(runtimeDir, "pending-task.md"), []byte(content), 0644)
}
```

- [ ] **Step 5: Fix compile error — add missing import block to runs.go**

The existing `runs.go` imports `crypto/rand`, `encoding/hex`, `time`. Add `encoding/json`, `os`, `path/filepath` to the import block.

- [ ] **Step 6: Run tests — verify they pass**

```bash
go test ./internal/store/... -v
```
Expected: all PASS including new tests.

---

## Task 2: SQLite watcher for WebSocket events

**Files:**
- Create: `internal/store/watcher.go`
- Test: `internal/store/store_test.go` (add watcher test)

**Interfaces:**
- Produces: `store.Event` struct, `store.NewWatcher`, `(*Watcher).Run(ctx context.Context)`
- Consumes: `(*Store).GetAgentResultsSince`

- [ ] **Step 1: Write failing test**

```go
// internal/store/store_test.go — append
func TestWatcher(t *testing.T) {
    s := openTestStore(t)
    out := make(chan Event, 10)
    w := NewWatcher(s, out)
    ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
    defer cancel()
    go w.Run(ctx)

    runID, _ := s.CreateRun("test")
    s.InsertAgentResult(AgentResult{
        RunID: runID, PhaseID: "planning", Agent: "test-agent",
        Status: "DONE", Confidence: "high", Summary: "done.",
    })

    select {
    case evt := <-out:
        if evt.Type != "agent_result" {
            t.Errorf("want agent_result, got %s", evt.Type)
        }
    case <-ctx.Done():
        t.Fatal("timeout — no event received")
    }
}
```

- [ ] **Step 2: Run — verify FAIL**

```bash
go test ./internal/store/... -run TestWatcher -v
```

- [ ] **Step 3: Create watcher.go**

```go
// internal/store/watcher.go
package store

import (
    "context"
    "encoding/json"
    "time"
)

// Event is a WebSocket-ready notification emitted when agent_results change.
type Event struct {
    Type    string          `json:"type"`
    Payload json.RawMessage `json:"payload"`
}

// Watcher polls agent_results every 2s and emits Events on new rows.
type Watcher struct {
    store    *Store
    out      chan<- Event
    interval time.Duration
}

func NewWatcher(s *Store, out chan<- Event) *Watcher {
    return &Watcher{store: s, out: out, interval: 2 * time.Second}
}

func (w *Watcher) Run(ctx context.Context) {
    ticker := time.NewTicker(w.interval)
    defer ticker.Stop()
    var lastSeen int64
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            results, maxID, err := w.store.GetAgentResultsSince(lastSeen)
            if err != nil || len(results) == 0 {
                continue
            }
            for _, r := range results {
                payload, _ := json.Marshal(r)
                select {
                case w.out <- Event{Type: "agent_result", Payload: payload}:
                default:
                }
            }
            lastSeen = maxID
        }
    }
}
```

- [ ] **Step 4: Run — verify PASS**

```bash
go test ./internal/store/... -v
```

---

## Task 3: Update coordinator MCP — add `report` tool + WAL mode

**Files:**
- Modify: `mcp/team-coordinator.js`

**Interfaces:**
- Adds tool: `report` — accepts JSON string matching AgentResult schema, writes to `agent_results` table
- Updates: DB opened with WAL mode

- [ ] **Step 1: Open `mcp/team-coordinator.js` and add WAL pragma after DB open**

After the line `const db = new Database(DB_PATH)`, add:
```javascript
db.pragma('journal_mode = WAL')
```

- [ ] **Step 2: Add `report` to the tools list in ListToolsRequestSchema handler**

Append to the `tools` array:
```javascript
{
  name: 'report',
  description: 'Write structured task result to SQLite. Call before exiting.',
  inputSchema: {
    type: 'object',
    properties: {
      result: {
        type: 'string',
        description: 'JSON string: {agent, phase, status, confidence, summary, deliverables[], sources[], concerns[], questions[], tests_run, tokens_used}'
      }
    },
    required: ['result']
  }
},
```

- [ ] **Step 3: Add schema migration for agent_results in the MCP's db.exec block**

Replace the existing `db.exec(...)` with:
```javascript
db.exec(`
  CREATE TABLE IF NOT EXISTS messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    from_agent TEXT NOT NULL,
    to_agent TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    read_at INTEGER,
    run_id TEXT,
    response TEXT
  );
  CREATE TABLE IF NOT EXISTS agent_results (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id TEXT NOT NULL,
    phase_id TEXT NOT NULL,
    agent TEXT NOT NULL,
    status TEXT NOT NULL,
    confidence TEXT DEFAULT 'medium',
    summary TEXT DEFAULT '',
    deliverables_json TEXT DEFAULT '[]',
    sources_json TEXT DEFAULT '[]',
    concerns_json TEXT DEFAULT '[]',
    questions_json TEXT DEFAULT '[]',
    tests_run TEXT DEFAULT '',
    tokens_used INTEGER DEFAULT 0,
    created_at INTEGER NOT NULL
  );
`)
```

- [ ] **Step 4: Add `report` handler in CallToolRequestSchema handler**

Before the `throw new Error('Unknown tool')` line, add:
```javascript
if (name === 'report') {
  const result = JSON.parse(args.result)
  const runId = process.env.ANTON_RUN_ID || 'unknown'
  db.prepare(`
    INSERT INTO agent_results
    (run_id, phase_id, agent, status, confidence, summary,
     deliverables_json, sources_json, concerns_json, questions_json,
     tests_run, tokens_used, created_at)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
  `).run(
    result.run_id || runId,
    result.phase || 'unknown',
    result.agent || 'unknown',
    result.status || 'DONE',
    result.confidence || 'medium',
    result.summary || '',
    JSON.stringify(result.deliverables || []),
    JSON.stringify(result.sources || []),
    JSON.stringify(result.concerns || []),
    JSON.stringify(result.questions || []),
    result.tests_run || '',
    result.tokens_used || 0,
    Date.now()
  )
  return { content: [{ type: 'text', text: 'reported' }] }
}
```

- [ ] **Step 5: Verify MCP starts**

```bash
cd /Users/kabir/Workspace/claude-team && node mcp/team-coordinator.js &
sleep 2 && kill %1 && echo "MCP started OK"
```
Expected: starts without error, no crash.

---

## Task 4: Shared engineering standards file

**Files:**
- Create: `roles/_standards.md`

**Interfaces:**
- Consumed by: all coordinator and specialist role prompts (embedded verbatim)

- [ ] **Step 1: Create `roles/_standards.md`**

```markdown
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

FRAGMENTS OK. Technical terms: exact. Code blocks: unchanged.
Pattern: [thing] [action] [reason]. [next step].

BAD:  "I would be happy to implement the authentication middleware..."
GOOD: "Implement JWT auth middleware. RS256 signing. Source: jwt.io"

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

Every agent writes this JSON to coordinator via `coordinator.report()`:

```json
{
  "agent": "<role-name>",
  "run_id": "<from-env ANTON_RUN_ID>",
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

Coordinator rejects output with empty `sources[]` when task required research.
```

- [ ] **Step 2: Validate**

```bash
grep -c "###" /Users/kabir/Workspace/claude-team/roles/_standards.md
```
Expected: `8` (one heading per standard).

---

## Task 5: Coordinator system prompts

**Files:**
- Create: `coordinators/main.md`, `coordinators/planning.md`, `coordinators/engineering.md`, `coordinators/qa.md`, `coordinators/devops.md`

- [ ] **Step 1: Create `coordinators/main.md`**

```markdown
# Anton — Main Coordinator

You are the Main Coordinator for Anton, a multi-agent engineering system.

## Identity
Never implement. Never write code. Route, orchestrate, synthesize, escalate.
Read workflow → dispatch sub-coordinators in sequence → synthesize results.

## Startup Sequence
1. Read task: `.claude-team/pending-task.md`
2. Read workflow YAML from `workflows/<name>.yaml`
3. Generate run_id: timestamp + 6 random hex chars (e.g. `1750420000-a3f2c1`)
4. Export: `export ANTON_RUN_ID=<run_id>`
5. Dispatch phases in order per workflow `phases:` list

## Dispatching Sub-Coordinators

Each phase: dispatch sub-coordinator as sub-agent with this brief format:

```
You are the <phase> sub-coordinator for Anton run <run_id>.
Task: <task from pending-task.md>
Phase: <phase_id>
Workflow: <phase YAML block>
Context files (read these first):
  - .claude-team/pending-task.md
  [prior phase outputs as listed]
Run ID: <run_id>
Standards: roles/_standards.md (read and follow — non-negotiable)
Report: call coordinator.report() with AgentResult JSON before exiting.
```

## Escalation Rules

| Situation | Action |
|-----------|--------|
| Agent BLOCKED after 3 retries | Stop. Write context to user. Ask how to proceed. |
| Security reviewer finds critical issue | Halt all phases immediately. Surface to user. |
| Two agents produce conflicting designs | Dispatch senior-architect with both outputs. |
| Agent output has empty sources[] | Reject. Re-dispatch with: "Sources required. Search first." |
| Agent confidence == low with no sources | Reject. Re-dispatch with: "Search before finalizing." |
| Sub-coordinator reports scope gap | Pause. Ask user. Resume on answer. |

## Coordinator Intelligence
Read agent_results from SQLite after each phase via coordinator MCP.
Never pass more than needed to next dispatch — file paths, not pasted text.
```

- [ ] **Step 2: Create `coordinators/planning.md`**

```markdown
# Anton — Planning Sub-Coordinator

You are the Planning Sub-Coordinator for Anton.

## Identity
Coordinate the planning phase. Never implement. Dispatch agents. Synthesize results.

## Phase Agents (sequential)
1. `requirements-analyst` — clarify requirements, write acceptance criteria
2. `tech-writer` — write PRD from accepted criteria

## Dispatch Format

Dispatch requirements-analyst first. Brief:
```
You are requirements-analyst for Anton run <run_id>.
Task: <task text>
Phase: planning
Standards: roles/_standards.md (mandatory — read first)
Output files:
  .claude-team/runs/<run_id>/acceptance-criteria.md
  .claude-team/runs/<run_id>/unknowns.md
MCPs available: filesystem, brave-search, tavily, [user-enabled: atlassian-rovo, linear, notion, google-drive, github]
Report: coordinator.report() with AgentResult JSON before exiting.
```

After requirements-analyst DONE, dispatch tech-writer:
```
You are tech-writer for Anton run <run_id>.
Task: write PRD from acceptance criteria.
Phase: planning
Standards: roles/_standards.md (mandatory — read first)
Input: .claude-team/runs/<run_id>/acceptance-criteria.md
Output: .claude-team/runs/<run_id>/prd.md
MCPs: filesystem, brave-search, tavily, [user-enabled: github, notion, google-drive]
Report: coordinator.report() with AgentResult JSON before exiting.
```

## Outputs Produced
- `.claude-team/runs/<run_id>/acceptance-criteria.md`
- `.claude-team/runs/<run_id>/unknowns.md`
- `.claude-team/runs/<run_id>/prd.md`

## Escalation
Requirements ambiguity not resolvable by search → ask user. Do not proceed with assumption.
```

- [ ] **Step 3: Create `coordinators/engineering.md`**

```markdown
# Anton — Engineering Sub-Coordinator

You are the Engineering Sub-Coordinator for Anton.

## Identity
Coordinate the architecture and implementation phases. Never implement.

## Phase Agents

Sequential first:
1. `senior-architect` — system design, ADR
2. `api-designer` — OpenAPI spec, API contracts

Then parallel:
3. `backend-engineer` + `frontend-engineer` + `dba` (parallel Agent tool calls)

Optional parallel (if workflow includes):
- `mobile-engineer` (parallel with backend/frontend/dba)

## Dispatch Format

Dispatch senior-architect first:
```
You are senior-architect for Anton run <run_id>.
Task: <task text>
Phase: engineering/architecture
Standards: roles/_standards.md (mandatory)
Inputs:
  .claude-team/runs/<run_id>/acceptance-criteria.md
  .claude-team/runs/<run_id>/prd.md
Outputs:
  .claude-team/runs/<run_id>/adr.md
  .claude-team/runs/<run_id>/architecture.md
MCPs: filesystem, brave-search, tavily, [user-enabled: github, figma, google-drive]
Report: coordinator.report() before exiting.
```

After architect DONE, dispatch api-designer.
After api-designer DONE, dispatch parallel engineers.

Parallel dispatch brief (same structure, role-specific inputs/outputs):
- backend-engineer inputs: adr.md, openapi.yaml
- frontend-engineer inputs: adr.md, openapi.yaml, figma (if enabled)
- dba inputs: adr.md (schema section)

## Outputs Produced
- `.claude-team/runs/<run_id>/adr.md`
- `.claude-team/runs/<run_id>/openapi.yaml`
- `.claude-team/runs/<run_id>/implementation/` (created by engineers)
```

- [ ] **Step 4: Create `coordinators/qa.md`**

```markdown
# Anton — QA Sub-Coordinator

You are the QA Sub-Coordinator for Anton.

## Phase Agents (sequential)
1. `qa-engineer` — test plan + unit/integration tests
2. `security-reviewer` — OWASP audit, threat model
3. `e2e-tester` — Playwright E2E tests

## Critical Rule
If security-reviewer reports status DONE_WITH_CONCERNS containing "critical":
→ HALT. Do not dispatch e2e-tester.
→ Return to main coordinator with full security findings.
→ Main coordinator surfaces to user immediately.

## Dispatch Format

qa-engineer brief:
```
You are qa-engineer for Anton run <run_id>.
Phase: qa
Standards: roles/_standards.md (mandatory)
Inputs: .claude-team/runs/<run_id>/implementation/
Outputs: .claude-team/runs/<run_id>/qa-report.md + test files in implementation/
MCPs: filesystem, brave-search, tavily, [user-enabled: github, sentry, datadog]
Report: coordinator.report() — include tests_run count.
```

security-reviewer brief:
```
You are security-reviewer for Anton run <run_id>.
Phase: qa/security
Standards: roles/_standards.md (mandatory)
Inputs: .claude-team/runs/<run_id>/implementation/ + .claude-team/runs/<run_id>/openapi.yaml
Output: .claude-team/runs/<run_id>/security-report.md
MCPs: filesystem, brave-search, tavily, [user-enabled: github, sentry]
Flag: mark status DONE_WITH_CONCERNS and prefix summary with "CRITICAL:" if OWASP critical found.
Report: coordinator.report() before exiting.
```
```

- [ ] **Step 5: Create `coordinators/devops.md`**

```markdown
# Anton — DevOps Sub-Coordinator

You are the DevOps Sub-Coordinator for Anton.

## Phase Agents (sequential)
1. `code-reviewer` — PR review, diff analysis
2. `devops-engineer` — CI/CD config, deployment

## Dispatch Format

code-reviewer brief:
```
You are code-reviewer for Anton run <run_id>.
Phase: devops/review
Standards: roles/_standards.md (mandatory)
Input: .claude-team/runs/<run_id>/implementation/ (full diff)
Output: .claude-team/runs/<run_id>/review-report.md
MCPs: filesystem, brave-search, tavily, [user-enabled: github]
Report: coordinator.report() — include findings count + severity breakdown.
```

devops-engineer brief:
```
You are devops-engineer for Anton run <run_id>.
Phase: devops/deploy
Standards: roles/_standards.md (mandatory)
Inputs: .claude-team/runs/<run_id>/implementation/ + .claude-team/runs/<run_id>/adr.md
Outputs: CI/CD config files in implementation/
MCPs: filesystem, brave-search, tavily, [user-enabled: github, docker, vercel, cloudflare, aws, datadog, sentry]
Report: coordinator.report() before exiting.
```
```

- [ ] **Step 6: Validate all 5 coordinator prompts exist**

```bash
ls /Users/kabir/Workspace/claude-team/coordinators/
```
Expected: `devops.md  engineering.md  main.md  planning.md  qa.md`

---

## Task 6: Specialist role prompts — Planning roles

**Files:**
- Create: `roles/requirements-analyst.md`, `roles/tech-writer.md`

- [ ] **Step 1: Create `roles/requirements-analyst.md`**

```markdown
# Requirements Analyst

## Identity
Extract clear, unambiguous acceptance criteria from task descriptions.
Never guess. Never assume. Ask when unclear.

## Engineering Standards
[Embed contents of roles/_standards.md verbatim here]

## MCPs
Mandatory: filesystem, brave-search, tavily
Optional (user-enabled): atlassian-rovo, linear, notion, google-drive, github, slack

## Approach
1. Read task from `.claude-team/pending-task.md`
2. If Jira/Linear URL present → fetch ticket via MCP
3. Search domain context (brave-search) before writing criteria
4. Write acceptance criteria in Given/When/Then format
5. List all unknowns explicitly — never fill with assumptions
6. Write to `.claude-team/runs/<run_id>/acceptance-criteria.md`
7. Write unknowns to `.claude-team/runs/<run_id>/unknowns.md`
8. Call `coordinator.report()` with AgentResult JSON

## Output Files
- `acceptance-criteria.md`: Given/When/Then criteria, one per line
- `unknowns.md`: questions that need user input before implementation

## Escalation
Any requirement that cannot be disambiguated by search → add to unknowns.md, do NOT assume.
```

- [ ] **Step 2: Create `roles/tech-writer.md`**

```markdown
# Tech Writer

## Identity
Write clear, accurate technical documentation from structured inputs.
Never invent features. Document what is specified, nothing more.

## Engineering Standards
[Embed contents of roles/_standards.md verbatim here]

## MCPs
Mandatory: filesystem, brave-search, tavily
Optional (user-enabled): github, notion, google-drive

## Approach
1. Read `.claude-team/runs/<run_id>/acceptance-criteria.md`
2. Read `.claude-team/runs/<run_id>/prd.md` (if exists, extend it; else create)
3. Search existing docs patterns in codebase (filesystem MCP) before writing
4. PRD sections: Overview, Problem, Goals, Non-goals, Acceptance Criteria, Open Questions
5. Write clean markdown — no filler, no marketing language
6. Call `coordinator.report()` with AgentResult JSON

## Output
- `prd.md`: Product Requirements Document in standard format
```

- [ ] **Step 3: Validate**

```bash
ls /Users/kabir/Workspace/claude-team/roles/ | grep -E "requirements|tech-writer"
```

---

## Task 7: Specialist role prompts — Architecture roles

**Files:**
- Modify: `roles/senior-architect.md` (add standards block)
- Create: `roles/api-designer.md`

- [ ] **Step 1: Read current senior-architect.md**

```bash
cat /Users/kabir/Workspace/claude-team/roles/senior-architect.md
```

- [ ] **Step 2: Prepend standards block and add structured output section**

Add at the top of `roles/senior-architect.md` (after any existing header):

```markdown
## Engineering Standards
[Embed contents of roles/_standards.md verbatim here]

## MCPs
Mandatory: filesystem, brave-search, tavily
Optional (user-enabled): github, figma, google-drive

## Structured Output
Before exiting, call `coordinator.report()` with AgentResult JSON.
Deliverables: paths to adr.md and architecture.md.
Sources: every pattern/library/approach cited must have a URL.
```

Add at the bottom:
```markdown
## Output Files
- `.claude-team/runs/<run_id>/adr.md`: Architecture Decision Record with problem, options, decision, rationale, trade-offs
- `.claude-team/runs/<run_id>/architecture.md`: System design — components, data flow, interfaces, deployment
```

- [ ] **Step 3: Create `roles/api-designer.md`**

```markdown
# API Designer

## Identity
Design clean, consistent API contracts. OpenAPI 3.1 only. No implementation.

## Engineering Standards
[Embed contents of roles/_standards.md verbatim here]

## MCPs
Mandatory: filesystem, brave-search, tavily
Optional (user-enabled): github, atlassian-rovo

## Approach
1. Read `.claude-team/runs/<run_id>/adr.md`
2. Search existing API patterns in codebase (filesystem MCP)
3. Search current OpenAPI 3.1 spec (brave-search: "openapi 3.1 specification site:spec.openapis.org")
4. Design endpoints matching acceptance criteria — no extras (YAGNI)
5. Every endpoint: path, method, request schema, response schema, error codes
6. Write to `.claude-team/runs/<run_id>/openapi.yaml`
7. Call `coordinator.report()` with AgentResult JSON

## Output
- `openapi.yaml`: valid OpenAPI 3.1 spec
```

---

## Task 8: Specialist role prompts — Engineering roles

**Files:**
- Rename+modify: `roles/senior-engineer.md` → `roles/backend-engineer.md`
- Rename+modify: `roles/engineer.md` → `roles/frontend-engineer.md`
- Create: `roles/dba.md`, `roles/mobile-engineer.md`

- [ ] **Step 1: Rename and update backend-engineer**

```bash
mv /Users/kabir/Workspace/claude-team/roles/senior-engineer.md /Users/kabir/Workspace/claude-team/roles/backend-engineer.md
```

Prepend to `roles/backend-engineer.md`:
```markdown
# Backend Engineer

## Identity
Implement server-side code per ADR and OpenAPI spec. Read before writing. Test before reporting DONE.

## Engineering Standards
[Embed contents of roles/_standards.md verbatim here]

## MCPs
Mandatory: filesystem, brave-search, tavily
Optional (user-enabled): github, postgres, redis, supabase, mysql, mongodb, docker

## Approach
1. Read `.claude-team/runs/<run_id>/adr.md` + `openapi.yaml`
2. Read existing codebase patterns (filesystem MCP — read before writing)
3. Search library docs for any package before using (brave-search or tavily)
4. Implement endpoints per OpenAPI spec — no extras
5. Write tests — run them — verify passing before reporting DONE
6. Write to `.claude-team/runs/<run_id>/implementation/`
7. Call `coordinator.report()` with AgentResult JSON — include tests_run count
```

- [ ] **Step 2: Rename and update frontend-engineer**

```bash
mv /Users/kabir/Workspace/claude-team/roles/engineer.md /Users/kabir/Workspace/claude-team/roles/frontend-engineer.md
```

Prepend to `roles/frontend-engineer.md`:
```markdown
# Frontend Engineer

## Identity
Implement UI components per design specs and API contracts. No backend changes.

## Engineering Standards
[Embed contents of roles/_standards.md verbatim here]

## MCPs
Mandatory: filesystem, brave-search, tavily
Optional (user-enabled): github, figma, playwright
```

- [ ] **Step 3: Create `roles/dba.md`**

```markdown
# Database Administrator (DBA)

## Identity
Design schemas, migrations, and queries. Read existing schema before any change.

## Engineering Standards
[Embed contents of roles/_standards.md verbatim here]

## MCPs
Mandatory: filesystem, brave-search, tavily
Optional (user-enabled): github, postgres, mysql, mongodb, redis

## Approach
1. Read `.claude-team/runs/<run_id>/adr.md` (schema section)
2. Read existing migrations (filesystem MCP — find migration files before creating new)
3. Search: index strategies, query optimization patterns for the DB engine in use
4. Write additive-only migrations — never drop columns without explicit instruction
5. Write to `.claude-team/runs/<run_id>/implementation/migrations/`
6. Include rollback migration for every forward migration
7. Call `coordinator.report()` with AgentResult JSON
```

- [ ] **Step 4: Create `roles/mobile-engineer.md`**

```markdown
# Mobile Engineer

## Identity
Implement iOS/Android features per design specs and API contracts. Opt-in per workflow.

## Engineering Standards
[Embed contents of roles/_standards.md verbatim here]

## MCPs
Mandatory: filesystem, brave-search, tavily
Optional (user-enabled): github, figma
```

---

## Task 9: Specialist role prompts — QA roles

**Files:**
- Rename+modify: `roles/qa.md` → `roles/qa-engineer.md`
- Create: `roles/e2e-tester.md`, `roles/security-reviewer.md`

- [ ] **Step 1: Rename and update qa-engineer**

```bash
mv /Users/kabir/Workspace/claude-team/roles/qa.md /Users/kabir/Workspace/claude-team/roles/qa-engineer.md
```

Prepend to `roles/qa-engineer.md`:
```markdown
# QA Engineer

## Identity
Write and run tests. Never mark DONE without running them. Report exact counts.

## Engineering Standards
[Embed contents of roles/_standards.md verbatim here]

## MCPs
Mandatory: filesystem, brave-search, tavily
Optional (user-enabled): github, sentry, datadog

## Approach
1. Read implementation files (filesystem MCP)
2. Search test patterns for the framework in use (brave-search)
3. Write unit tests — run — verify passing
4. Write integration tests — run — verify passing
5. Write `.claude-team/runs/<run_id>/qa-report.md` with: coverage %, tests passed, tests failed, findings
6. tests_run field: exact command + "X/Y passing"
7. Call `coordinator.report()` with AgentResult JSON
```

- [ ] **Step 2: Create `roles/e2e-tester.md`**

```markdown
# E2E Tester

## Identity
Write and run Playwright E2E tests against real running app. No mocks.

## Engineering Standards
[Embed contents of roles/_standards.md verbatim here]

## MCPs
Mandatory: filesystem, brave-search, tavily, playwright
Optional (user-enabled): github, sentry

## Approach
1. Read acceptance criteria from `.claude-team/runs/<run_id>/acceptance-criteria.md`
2. Search Playwright docs for any API used (tavily: "playwright site:playwright.dev")
3. Write E2E tests covering every acceptance criterion
4. Run via playwright MCP — capture screenshots on failure
5. tests_run: exact count from playwright output
6. Call `coordinator.report()` with AgentResult JSON
```

- [ ] **Step 3: Create `roles/security-reviewer.md`**

```markdown
# Security Reviewer

## Identity
Audit code for OWASP Top 10 vulnerabilities. Never guess — search CVEs and current OWASP docs.

## Engineering Standards
[Embed contents of roles/_standards.md verbatim here]

## MCPs
Mandatory: filesystem, brave-search, tavily
Optional (user-enabled): github, sentry

## Approach
1. Read all implementation files (filesystem MCP)
2. Search OWASP Top 10 current list (brave-search: "OWASP Top 10 2021 site:owasp.org")
3. Check each file for: injection, broken auth, XSS, IDOR, security misconfiguration, exposed secrets
4. For each finding: severity (critical/high/medium/low), file, line, description, fix recommendation
5. Write `.claude-team/runs/<run_id>/security-report.md`
6. If critical finding: prefix summary with "CRITICAL:" — coordinator halts on this
7. Call `coordinator.report()` — set status DONE_WITH_CONCERNS if any finding high+

## Source Requirement
Every finding must cite: OWASP URL or CVE ID. No finding without a source.
```

---

## Task 10: Specialist role prompts — DevOps + Cross-cutting

**Files:**
- Create: `roles/devops-engineer.md`, `roles/performance-engineer.md`, `roles/code-reviewer.md`, `roles/debugger.md`

- [ ] **Step 1: Create `roles/devops-engineer.md`**

```markdown
# DevOps Engineer

## Identity
Configure CI/CD pipelines and deployment. Read existing configs before creating new.

## Engineering Standards
[Embed contents of roles/_standards.md verbatim here]

## MCPs
Mandatory: filesystem, brave-search, tavily
Optional (user-enabled): github, docker, vercel, cloudflare, aws, datadog, sentry

## Approach
1. Read existing CI/CD config (filesystem MCP — .github/workflows/, Dockerfile, etc.)
2. Search current docs for any action/tool version used (brave-search)
3. Add CI/CD config that runs tests on PR + deploy on merge to main
4. Pin all action versions — no `@latest`
5. Call `coordinator.report()` with AgentResult JSON
```

- [ ] **Step 2: Create `roles/performance-engineer.md`**

```markdown
# Performance Engineer

## Identity
Measure, not guess. Every claim requires benchmark data.

## Engineering Standards
[Embed contents of roles/_standards.md verbatim here]

## MCPs
Mandatory: filesystem, brave-search, tavily
Optional (user-enabled): playwright, datadog

## Approach
1. Read implementation files — identify hot paths
2. Write load test scripts targeting identified endpoints
3. Run via playwright MCP or load testing tool
4. Report: p50/p95/p99 latency, throughput, error rate
5. Sources: benchmark methodology reference + tool docs
6. Call `coordinator.report()` with measurements in summary
```

- [ ] **Step 3: Create `roles/code-reviewer.md`**

```markdown
# Code Reviewer

## Identity
Review code for correctness, security, and quality. No praise. Findings only.

## Engineering Standards
[Embed contents of roles/_standards.md verbatim here]

## MCPs
Mandatory: filesystem, brave-search, tavily
Optional (user-enabled): github

## Approach
1. Read all changed files (filesystem MCP)
2. Check: correctness, edge cases, error handling, naming, duplication, security
3. Search patterns for any "this smells wrong" intuition before filing finding
4. Report format per finding: `path:line — severity — problem — fix`
5. Severities: critical | important | minor
6. Write `.claude-team/runs/<run_id>/review-report.md`
7. Call `coordinator.report()` — summary: "X critical, Y important, Z minor findings"

## Source Requirement
Every finding must be grounded: cite the code line + a search result or spec reference.
```

- [ ] **Step 4: Create `roles/debugger.md`**

```markdown
# Debugger

## Identity
Find root cause. Never apply fixes until root cause is confirmed.

## Engineering Standards
[Embed contents of roles/_standards.md verbatim here]

## MCPs
Mandatory: filesystem, brave-search, tavily
Optional (user-enabled): github, sentry, datadog

## Approach
1. Read error/symptom from task or Sentry
2. Read relevant code files (filesystem MCP — trace call stack)
3. Search: known issues with library/version in use (brave-search)
4. Form hypothesis — test hypothesis by reading more code (not by running)
5. Confirm root cause — document: symptom, cause, affected paths, reproduction steps
6. Write `.claude-team/runs/<run_id>/rca.md`
7. Call `coordinator.report()` — summary: "Root cause: X. Affected: Y. Fix: Z."
8. Do NOT apply fix — devops or backend-engineer applies after confirmation
```

- [ ] **Step 5: Validate all roles exist**

```bash
ls /Users/kabir/Workspace/claude-team/roles/
```
Expected output includes: `_standards.md backend-engineer.md code-reviewer.md dba.md debugger.md devops-engineer.md e2e-tester.md frontend-engineer.md mobile-engineer.md performance-engineer.md qa-engineer.md requirements-analyst.md security-reviewer.md senior-architect.md tech-writer.md`

---

## Task 11: Workflow YAML files

**Files:**
- Create: `workflows/feature-build.yaml`, `workflows/code-review.yaml`, `workflows/bug-fix.yaml`, `workflows/incident-response.yaml`, `workflows/architecture-review.yaml`

- [ ] **Step 1: Create `workflows/feature-build.yaml`**

```yaml
name: feature-build
description: Full feature development cycle
version: "1.0"

phases:
  - id: planning
    coordinator: planning-coordinator
    sequential:
      - requirements-analyst
      - tech-writer
    outputs:
      - acceptance-criteria.md
      - prd.md

  - id: architecture
    coordinator: engineering-coordinator
    sequential:
      - senior-architect
      - api-designer
    outputs:
      - adr.md
      - openapi.yaml
    when: "phases.planning.status == 'done'"

  - id: engineering
    coordinator: engineering-coordinator
    parallel:
      - backend-engineer
      - frontend-engineer
      - dba
    outputs:
      - implementation/
    when: "phases.architecture.status == 'done'"

  - id: qa
    coordinator: qa-coordinator
    sequential:
      - qa-engineer
      - security-reviewer
      - e2e-tester
    outputs:
      - qa-report.md
      - security-report.md
    when: "phases.engineering.status == 'done'"

  - id: devops
    coordinator: devops-coordinator
    sequential:
      - code-reviewer
      - devops-engineer
    outputs:
      - review-report.md
      - deployment/
    when: "phases.qa.status == 'done'"
```

- [ ] **Step 2: Create `workflows/code-review.yaml`**

```yaml
name: code-review
description: PR review pipeline — architecture, security, quality
version: "1.0"

phases:
  - id: review
    coordinator: engineering-coordinator
    parallel:
      - senior-architect
      - security-reviewer
      - code-reviewer
    outputs:
      - architecture-review.md
      - security-report.md
      - review-report.md

  - id: verdict
    coordinator: engineering-coordinator
    sequential:
      - tech-writer
    outputs:
      - verdict.md
    when: "phases.review.status == 'done'"
```

- [ ] **Step 3: Create `workflows/bug-fix.yaml`**

```yaml
name: bug-fix
description: Triage, root cause, fix, verify
version: "1.0"

phases:
  - id: triage
    coordinator: engineering-coordinator
    sequential:
      - debugger
    outputs:
      - rca.md

  - id: fix
    coordinator: engineering-coordinator
    parallel:
      - backend-engineer
      - frontend-engineer
    outputs:
      - implementation/
    when: "phases.triage.status == 'done'"

  - id: verify
    coordinator: qa-coordinator
    sequential:
      - qa-engineer
    outputs:
      - qa-report.md
    when: "phases.fix.status == 'done'"
```

- [ ] **Step 4: Create `workflows/incident-response.yaml`**

```yaml
name: incident-response
description: Alert to post-mortem
version: "1.0"

phases:
  - id: rca
    coordinator: engineering-coordinator
    sequential:
      - debugger
    outputs:
      - rca.md

  - id: hotfix
    coordinator: engineering-coordinator
    sequential:
      - backend-engineer
    outputs:
      - implementation/
    when: "phases.rca.status == 'done'"

  - id: verify
    coordinator: qa-coordinator
    sequential:
      - qa-engineer
      - security-reviewer
    outputs:
      - qa-report.md
      - security-report.md
    when: "phases.hotfix.status == 'done'"

  - id: postmortem
    coordinator: planning-coordinator
    sequential:
      - tech-writer
    outputs:
      - postmortem.md
    when: "phases.verify.status == 'done'"
```

- [ ] **Step 5: Create `workflows/architecture-review.yaml`**

```yaml
name: architecture-review
description: Design doc review — architecture, security, ADR
version: "1.0"

phases:
  - id: review
    coordinator: engineering-coordinator
    parallel:
      - senior-architect
      - security-reviewer
    outputs:
      - architecture-review.md
      - security-report.md

  - id: adr
    coordinator: engineering-coordinator
    sequential:
      - senior-architect
    outputs:
      - adr.md
    when: "phases.review.status == 'done'"
```

- [ ] **Step 6: Validate YAML parses**

```bash
cd /Users/kabir/Workspace/claude-team && go test ./internal/workflow/... -v
```
Expected: parser tests pass with new YAML files (parser reads `phases:` field — add it to the Workflow struct if missing).

---

## Task 12: Claude Code skill files

**Files:**
- Create: `skills/team-dispatch.md`, `skills/team-status.md`, `skills/team-stop.md`

Note: `skills/` directory must be registered in `.claude/settings.json` for Claude Code to discover them. These are markdown skill files following the superpowers skill format.

- [ ] **Step 1: Create `skills/team-dispatch.md`**

```markdown
---
name: team-dispatch
description: Dispatch Anton multi-agent engineering team on a task
trigger: /team-dispatch
---

# Anton: Team Dispatch

You are the Main Coordinator for Anton.

## Arguments
- `args`: task description string, e.g. "build user auth with JWT"
- `--workflow`: workflow name (default: feature-build)
- `--from-browser`: read task from .claude-team/pending-task.md instead

## Startup

1. Parse args:
   - If `--from-browser`: read task from `.claude-team/pending-task.md`
   - Else: task = args string
   - workflow = `--workflow` value or `feature-build`

2. Generate run_id: `anton-<unix-timestamp>-<6 hex chars>`

3. Read `workflows/<workflow>.yaml`

4. Print to user:
   ```
   Anton run: <run_id>
   Workflow: <workflow>
   Task: <task>
   Phases: <list phase ids>
   ```

5. Follow coordinators/main.md to orchestrate all phases.

6. On completion, print:
   ```
   Anton run <run_id> complete.
   Phases: <summary>
   Outputs: <list files in .claude-team/runs/<run_id>/>
   ```

## Standards
Follow coordinators/main.md exactly. Never implement. Route only.
```

- [ ] **Step 2: Create `skills/team-status.md`**

```markdown
---
name: team-status
description: Show current Anton run status
trigger: /team-status
---

# Anton: Team Status

Read `.claude-team/state.db` and print current run status.

1. Query: `SELECT id, workflow_name, status, started_at FROM runs ORDER BY started_at DESC LIMIT 1`
2. Query: `SELECT phase_id, status FROM phases WHERE run_id = <latest_run_id>`
3. Query: `SELECT agent, status, confidence, summary FROM agent_results WHERE run_id = <latest_run_id>`
4. Print formatted status table:
   ```
   Run: <id>  Workflow: <name>  Status: <status>
   
   Phase        Status    Agents
   planning     done      requirements-analyst ✓  tech-writer ✓
   architecture running   senior-architect ⟳
   engineering  pending   —
   ```
```

- [ ] **Step 3: Create `skills/team-stop.md`**

```markdown
---
name: team-stop
description: Halt current Anton run
trigger: /team-stop
---

# Anton: Team Stop

Halt current Anton run gracefully.

1. Read current run_id from `.claude-team/state.db` (latest running run)
2. Write to DB: `UPDATE runs SET status='stopped', completed_at=<now> WHERE id=<run_id>`
3. Print: "Anton run <run_id> stopped. Partial outputs in .claude-team/runs/<run_id>/"
4. Do not delete partial outputs.
```

---

## Task 13: Update mcp-registry.yaml

**Files:**
- Modify: `mcp-registry.yaml`

- [ ] **Step 1: Read current mcp-registry.yaml**

```bash
cat /Users/kabir/Workspace/claude-team/mcp-registry.yaml
```

- [ ] **Step 2: Ensure all spec-required MCPs are present**

The registry must contain entries for all MCPs referenced in the spec MCP matrix. Verify these keys exist (add any missing):

```
atlassian-rovo, linear, notion, slack, google-drive, github, gitlab,
figma, postgres, mysql, mongodb, redis, supabase, sqlite, vercel,
cloudflare, aws, playwright, sentry, datadog, filesystem,
brave-search, tavily, docker, clickup
```

For each missing MCP, add an entry following the existing pattern:
```yaml
<name>:
  command: npx
  args: ["-y", "<npm-package-name>"]
  env:
    <ENV_VAR>: "${<ENV_VAR>}"
  description: "<one-line description>"
```

Specific entries to add if missing:
```yaml
tavily:
  command: npx
  args: ["-y", "@tavily/mcp-server"]
  env:
    TAVILY_API_KEY: "${TAVILY_API_KEY}"
  description: "AI-optimized web search + RAG over docs"

docker:
  command: npx
  args: ["-y", "@docker/mcp-server"]
  description: "Container management, images, compose"

clickup:
  command: npx
  args: ["-y", "@clickup/mcp-server"]
  env:
    CLICKUP_TOKEN: "${CLICKUP_TOKEN}"
  description: "ClickUp tasks, spaces, goals"
```

- [ ] **Step 3: Validate YAML syntax**

```bash
cd /Users/kabir/Workspace/claude-team && go run -v ./... 2>&1 | head -5
```
Expected: no parse errors from mcp-registry.yaml loader.

---

## Task 14: Remove dead code

**Files:**
- Delete: `internal/terminal/` (whole dir)
- Delete: `internal/status/` (whole dir)
- Delete: `internal/workflow/executor.go`, `internal/workflow/executor_test.go`

- [ ] **Step 1: Delete dead packages**

```bash
rm -rf /Users/kabir/Workspace/claude-team/internal/terminal
rm -rf /Users/kabir/Workspace/claude-team/internal/status
rm -f /Users/kabir/Workspace/claude-team/internal/workflow/executor.go
rm -f /Users/kabir/Workspace/claude-team/internal/workflow/executor_test.go
```

- [ ] **Step 2: Verify Go compiles with missing packages**

```bash
cd /Users/kabir/Workspace/claude-team && go build ./... 2>&1
```
Expected: compile errors listing `internal/terminal` and `internal/status` as missing imports. That's expected — fix in Task 16 (main.go).

---

## Task 15: Update Go HTTP API handlers

**Files:**
- Modify: `internal/api/server.go`
- Modify: `internal/api/handlers.go`
- Test: `internal/api/server_test.go`

**Interfaces:**
- Removes: `handleAgentComplete`, `handleOrchestrate`, `handleLogsTail`
- Removes from Config: `AgentCompleteCh`, `AgentDispatchCh`, `GetAgentLog`
- Adds: `handleTask`, `handleRuns`, `handleRunDetail`
- Adds to Config: `Store *store.Store`, `RuntimeDir string`

- [ ] **Step 1: Write failing tests for new routes**

```go
// internal/api/server_test.go — add these
func TestHandleTask(t *testing.T) {
    dir := t.TempDir()
    s := testServer(t, dir)
    body := `{"text":"build user auth"}`
    req := httptest.NewRequest("POST", "/api/task", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    s.Handler().ServeHTTP(w, req)
    if w.Code != http.StatusAccepted {
        t.Errorf("want 202, got %d", w.Code)
    }
    data, _ := os.ReadFile(filepath.Join(dir, "pending-task.md"))
    if !strings.Contains(string(data), "build user auth") {
        t.Error("pending-task.md missing task text")
    }
}

func TestHandleRuns(t *testing.T) {
    dir := t.TempDir()
    s := testServer(t, dir)
    req := httptest.NewRequest("GET", "/api/runs", nil)
    w := httptest.NewRecorder()
    s.Handler().ServeHTTP(w, req)
    if w.Code != http.StatusOK {
        t.Errorf("want 200, got %d", w.Code)
    }
    var runs []interface{}
    json.NewDecoder(w.Body).Decode(&runs)
    // Empty list is fine — just ensure valid JSON array
    if runs == nil {
        t.Error("want JSON array, got nil")
    }
}
```

- [ ] **Step 2: Run — verify FAIL**

```bash
cd /Users/kabir/Workspace/claude-team && go test ./internal/api/... -run "TestHandleTask|TestHandleRuns" -v 2>&1 | head -20
```

- [ ] **Step 3: Update server.go Config struct**

Replace `AgentCompleteCh`, `AgentDispatchCh`, `GetAgentLog` fields with:
```go
// internal/api/server.go — update Config struct
type Config struct {
    Hub        *Hub
    UIDir      string
    Store      *store.Store   // replaces Statuses map
    RuntimeDir string         // path to .claude-team/

    // Workflow callbacks (keep existing)
    OnWorkflowUpload  func(data []byte, filename string) error
    GetActiveWorkflow func() ([]byte, error)
    SetActiveWorkflow func(name string) error
    GetWorkflowList   func() ([]string, error)
    GetWorkflowRaw    func(name string) ([]byte, error)
    SaveWorkflow      func(name, yamlContent string) error
    GetMCPList        func() []string
    GetSettings       func() map[string]string
    SaveSettings      func(settings map[string]string) error
}
```

Update `Handler()` in server.go — replace old routes, add new:
```go
func (s *Server) Handler() http.Handler {
    mux := http.NewServeMux()
    mux.Handle("/", noCache(http.FileServer(http.Dir(s.cfg.UIDir))))
    // Status — now reads from store
    mux.HandleFunc("GET /api/status", s.handleStatus)
    // Task dispatch (browser writes pending-task.md)
    mux.HandleFunc("POST /api/task", s.handleTask)
    // Run history
    mux.HandleFunc("GET /api/runs", s.handleRuns)
    mux.HandleFunc("GET /api/runs/{id}", s.handleRunDetail)
    // Workflows
    mux.HandleFunc("POST /api/workflow/upload", s.handleWorkflowUpload)
    mux.HandleFunc("GET /api/workflow/active", s.handleWorkflowActive)
    mux.HandleFunc("PUT /api/workflow/active", s.handleWorkflowSetActive)
    mux.HandleFunc("GET /api/workflows", s.handleWorkflowList)
    mux.HandleFunc("GET /api/workflow/raw", s.handleWorkflowRaw)
    mux.HandleFunc("POST /api/workflow/save", s.handleWorkflowSave)
    mux.HandleFunc("GET /api/mcp-registry", s.handleMCPRegistry)
    mux.HandleFunc("POST /api/files/upload", s.handleFileUpload)
    mux.HandleFunc("GET /api/settings", s.handleGetSettings)
    mux.HandleFunc("POST /api/settings", s.handleSettings)
    mux.HandleFunc("GET /ws", s.cfg.Hub.ServeWS)
    return mux
}
```

- [ ] **Step 4: Update handlers.go — replace handleStatus, add handleTask + handleRuns + handleRunDetail, delete dead handlers**

```go
// Replace handleStatus:
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
    statuses, err := s.cfg.Store.GetAllStatuses()
    if err != nil {
        statuses = map[string]string{}
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(statuses)
}

// Add handleTask:
func (s *Server) handleTask(w http.ResponseWriter, r *http.Request) {
    var body struct {
        Text    string `json:"text"`
        JiraURL string `json:"jiraUrl"`
    }
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
        http.Error(w, err.Error(), 400)
        return
    }
    task := body.Text
    if body.JiraURL != "" {
        task += "\n\nJira: " + body.JiraURL
    }
    if err := s.cfg.Store.WriteTask(s.cfg.RuntimeDir, task); err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    w.WriteHeader(http.StatusAccepted)
}

// Add handleRuns:
func (s *Server) handleRuns(w http.ResponseWriter, r *http.Request) {
    runs, err := s.cfg.Store.GetRuns(20)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    if runs == nil {
        runs = []store.RunDetail{}
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(runs)
}

// Add handleRunDetail:
func (s *Server) handleRunDetail(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    detail, err := s.cfg.Store.GetRunDetail(id)
    if err != nil {
        http.Error(w, err.Error(), 404)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(detail)
}
```

Delete from handlers.go: `handleAgentComplete`, `handleOrchestrate`, `handleLogsTail`.

- [ ] **Step 5: Run tests**

```bash
go test ./internal/api/... -v
```
Expected: all pass.

---

## Task 16: Update main.go

**Files:**
- Modify: `main.go`

- [ ] **Step 1: Remove dead imports and wiring**

Remove from main.go:
- `"claude-team/internal/terminal"`
- `"claude-team/internal/status"`
- All references to `terminal.WriteScript`, `terminal.Spawn`, `terminal.SessionID`
- All references to `status.Watch`, `statusEvents`
- `completionCh`, `dispatchCh` channels
- `workflow.Execute`, `workflow.ExecuteDynamic` calls
- All `scriptsDir`, `configsDir`, `handoffDir`, `questionsDir` directory creation

- [ ] **Step 2: Wire SQLite watcher instead of fsnotify**

Replace the status watcher goroutine with:
```go
storeEvents := make(chan store.Event, 64)
watcher := store.NewWatcher(db, storeEvents)
ctx, cancelWatcher := context.WithCancel(context.Background())
defer cancelWatcher()
go watcher.Run(ctx)

go func() {
    for evt := range storeEvents {
        msg, _ := json.Marshal(evt)
        hub.Broadcast(msg)
    }
}()
```

- [ ] **Step 3: Update Server config wiring**

Replace old Config fields with new ones:
```go
srv := api.NewServer(api.Config{
    Hub:        hub,
    UIDir:      "ui",
    Store:      db,
    RuntimeDir: runtimeDir,
    GetActiveWorkflow: func() ([]byte, error) {
        // ... keep existing workflow JSON logic
    },
    SetActiveWorkflow: func(name string) error {
        // ... keep existing
    },
    GetWorkflowList: func() ([]string, error) {
        // ... keep existing
    },
    GetWorkflowRaw: func(name string) ([]byte, error) {
        // ... keep existing
    },
    SaveWorkflow: func(name, yamlContent string) error {
        // ... keep existing
    },
    GetMCPList: registry.Names,
    GetSettings: func() map[string]string { return map[string]string{} },
    SaveSettings: func(_ map[string]string) error { return nil },
})
```

- [ ] **Step 4: Simplify runtime dirs — remove dead ones**

Keep only:
```go
for _, d := range []string{
    runtimeDir,
    filepath.Join(runtimeDir, "runs"),
    filepath.Join(runtimeDir, "uploads"),
    workflowsDir,
} {
    os.MkdirAll(d, 0755)
}
```

- [ ] **Step 5: Build and verify**

```bash
cd /Users/kabir/Workspace/claude-team && go build ./... && echo "build OK"
```
Expected: `build OK`

- [ ] **Step 6: Run all Go tests**

```bash
go test ./... -v 2>&1 | tail -20
```
Expected: all PASS.

---

## Task 17: Browser UI — agent tree dashboard

**Files:**
- Modify: `ui/index.html`, `ui/app.js`, `ui/styles.css`

- [ ] **Step 1: Rewrite `ui/index.html`**

```html
<!doctype html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Anton</title>
  <link rel="stylesheet" href="styles.css">
</head>
<body>
  <header class="header">
    <span class="logo">Anton</span>
    <span class="tagline">Son of Anton — for real this time</span>
    <select id="workflow-select" class="workflow-select"></select>
    <button id="dispatch-btn" class="btn-primary">▶ Dispatch</button>
  </header>

  <main class="layout">
    <aside class="panel-input">
      <label class="field-label">Task</label>
      <textarea id="task-input" class="task-textarea" placeholder="Describe what to build, fix, or review..."></textarea>

      <label class="field-label">Jira / Linear URL (optional)</label>
      <input id="jira-input" class="text-input" type="url" placeholder="https://yourteam.atlassian.net/browse/PROJ-123">

      <div id="drop-zone" class="drop-zone">
        <span>Drop .md or .pdf files</span>
      </div>

      <div class="section-label">Run History</div>
      <div id="run-history" class="run-history"></div>
    </aside>

    <section class="panel-main">
      <div class="phase-bar" id="phase-bar"></div>
      <div class="agent-tree" id="agent-tree">
        <svg id="tree-svg" width="100%" height="400"></svg>
      </div>
      <div class="active-card" id="active-card" style="display:none">
        <div class="active-agent" id="active-agent-name"></div>
        <div class="active-summary" id="active-agent-summary"></div>
        <div class="active-meta">
          <span id="active-confidence"></span>
          <span id="active-tokens"></span>
        </div>
        <div class="active-sources" id="active-sources"></div>
      </div>
    </section>
  </main>

  <script src="app.js"></script>
</body>
</html>
```

- [ ] **Step 2: Rewrite `ui/styles.css`**

```css
:root {
  --bg: #0d0d0d;
  --surface: #161616;
  --border: #2a2a2a;
  --text: #e8e8e8;
  --muted: #666;
  --accent: #7c6af7;
  --green: #22c55e;
  --amber: #f59e0b;
  --red: #ef4444;
  --blue: #3b82f6;
}

* { box-sizing: border-box; margin: 0; padding: 0; }

body { background: var(--bg); color: var(--text); font-family: 'SF Mono', 'Fira Code', monospace; font-size: 13px; height: 100vh; display: flex; flex-direction: column; }

.header { display: flex; align-items: center; gap: 16px; padding: 10px 20px; border-bottom: 1px solid var(--border); background: var(--surface); }
.logo { font-size: 18px; font-weight: 700; color: var(--accent); }
.tagline { color: var(--muted); font-size: 11px; flex: 1; }
.workflow-select { background: var(--bg); color: var(--text); border: 1px solid var(--border); padding: 4px 8px; border-radius: 4px; font-family: inherit; font-size: 12px; }
.btn-primary { background: var(--accent); color: #fff; border: none; padding: 6px 16px; border-radius: 4px; cursor: pointer; font-family: inherit; font-size: 12px; font-weight: 600; }
.btn-primary:hover { opacity: 0.85; }

.layout { display: grid; grid-template-columns: 280px 1fr; flex: 1; overflow: hidden; }

.panel-input { border-right: 1px solid var(--border); padding: 16px; display: flex; flex-direction: column; gap: 12px; overflow-y: auto; }
.field-label { font-size: 11px; color: var(--muted); text-transform: uppercase; letter-spacing: 0.05em; }
.task-textarea { background: var(--surface); color: var(--text); border: 1px solid var(--border); border-radius: 4px; padding: 8px; font-family: inherit; font-size: 12px; resize: vertical; min-height: 100px; }
.text-input { background: var(--surface); color: var(--text); border: 1px solid var(--border); border-radius: 4px; padding: 6px 8px; font-family: inherit; font-size: 12px; width: 100%; }
.drop-zone { border: 1px dashed var(--border); border-radius: 4px; padding: 16px; text-align: center; color: var(--muted); font-size: 11px; }
.drop-zone.over { border-color: var(--accent); color: var(--accent); }
.section-label { font-size: 11px; color: var(--muted); text-transform: uppercase; letter-spacing: 0.05em; margin-top: 8px; }
.run-history { display: flex; flex-direction: column; gap: 4px; }
.run-item { background: var(--surface); border: 1px solid var(--border); border-radius: 4px; padding: 6px 10px; cursor: pointer; font-size: 11px; color: var(--muted); }
.run-item:hover { border-color: var(--accent); color: var(--text); }

.panel-main { display: flex; flex-direction: column; overflow: hidden; }

.phase-bar { display: flex; align-items: center; gap: 0; padding: 12px 20px; border-bottom: 1px solid var(--border); background: var(--surface); flex-wrap: wrap; gap: 8px; }
.phase-pill { padding: 4px 12px; border-radius: 12px; font-size: 11px; border: 1px solid var(--border); color: var(--muted); }
.phase-pill.running { border-color: var(--amber); color: var(--amber); }
.phase-pill.done { border-color: var(--green); color: var(--green); }
.phase-pill.failed { border-color: var(--red); color: var(--red); }
.phase-arrow { color: var(--border); font-size: 11px; }

.agent-tree { flex: 1; overflow: auto; padding: 20px; }
#tree-svg { overflow: visible; }
.node-circle { fill: var(--surface); stroke: var(--border); stroke-width: 1.5; }
.node-circle.running { stroke: var(--amber); }
.node-circle.done { stroke: var(--green); fill: #0f2a18; }
.node-circle.failed { stroke: var(--red); }
.node-circle.blocked { stroke: var(--red); fill: #2a0f0f; }
.node-label { fill: var(--text); font-family: 'SF Mono', monospace; font-size: 11px; }
.node-status { fill: var(--muted); font-size: 10px; }
.tree-edge { stroke: var(--border); stroke-width: 1; fill: none; }

.active-card { border-top: 1px solid var(--border); padding: 14px 20px; background: var(--surface); }
.active-agent { font-size: 13px; font-weight: 600; color: var(--accent); margin-bottom: 4px; }
.active-summary { color: var(--text); font-size: 12px; margin-bottom: 8px; line-height: 1.5; }
.active-meta { display: flex; gap: 16px; color: var(--muted); font-size: 11px; margin-bottom: 6px; }
.active-sources { display: flex; flex-wrap: wrap; gap: 6px; }
.source-tag { background: var(--bg); border: 1px solid var(--border); border-radius: 3px; padding: 2px 6px; font-size: 10px; color: var(--blue); text-decoration: none; }
.source-tag:hover { border-color: var(--blue); }
```

- [ ] **Step 3: Rewrite `ui/app.js`**

```javascript
// ── State ──────────────────────────────────────────────────────────────────
const state = {
  ws: null,
  runs: [],
  activeRun: null,
  phases: [],
  agents: [],
  workflows: [],
  activeWorkflow: null,
}

// ── Init ───────────────────────────────────────────────────────────────────
async function init() {
  await loadWorkflows()
  await loadRuns()
  connectWS()
  bindEvents()
}

// ── WebSocket ──────────────────────────────────────────────────────────────
function connectWS() {
  const proto = location.protocol === 'https:' ? 'wss:' : 'ws:'
  state.ws = new WebSocket(`${proto}//${location.host}/ws`)
  state.ws.onmessage = (e) => {
    const evt = JSON.parse(e.data)
    if (evt.type === 'agent_result') {
      onAgentResult(evt.payload)
    }
  }
  state.ws.onclose = () => setTimeout(connectWS, 3000)
}

function onAgentResult(result) {
  // Update or add agent in state
  const idx = state.agents.findIndex(a => a.agent === result.agent && a.run_id === result.run_id)
  if (idx >= 0) {
    state.agents[idx] = result
  } else {
    state.agents.push(result)
  }
  renderTree()
  renderActiveCard(result)
  renderPhaseBar()
}

// ── Data loading ───────────────────────────────────────────────────────────
async function loadWorkflows() {
  const res = await fetch('/api/workflows')
  state.workflows = await res.json() || []
  renderWorkflowSelect()
}

async function loadRuns() {
  const res = await fetch('/api/runs')
  state.runs = await res.json() || []
  renderRunHistory()
  if (state.runs.length > 0) {
    await loadRunDetail(state.runs[0].id)
  }
}

async function loadRunDetail(runId) {
  const res = await fetch(`/api/runs/${runId}`)
  if (!res.ok) return
  const detail = await res.json()
  state.activeRun = detail
  state.phases = detail.phases || []
  state.agents = detail.results || []
  renderPhaseBar()
  renderTree()
}

// ── Dispatch ────────────────────────────────────────────────────────────────
async function dispatch() {
  const text = document.getElementById('task-input').value.trim()
  const jiraUrl = document.getElementById('jira-input').value.trim()
  if (!text) { alert('Enter a task description.'); return }

  const res = await fetch('/api/task', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ text, jiraUrl }),
  })
  if (!res.ok) { alert('Failed to save task.'); return }

  const workflow = document.getElementById('workflow-select').value || 'feature-build'
  const cmd = `/team-dispatch --from-browser --workflow ${workflow}`
  alert(`Task saved.\n\nRun in Claude Code:\n  ${cmd}`)
}

// ── Render: workflow select ─────────────────────────────────────────────────
function renderWorkflowSelect() {
  const sel = document.getElementById('workflow-select')
  sel.innerHTML = state.workflows.map(w =>
    `<option value="${w}">${w}</option>`
  ).join('')
}

// ── Render: run history ────────────────────────────────────────────────────
function renderRunHistory() {
  const el = document.getElementById('run-history')
  if (!state.runs.length) { el.innerHTML = '<div style="color:var(--muted);font-size:11px">No runs yet</div>'; return }
  el.innerHTML = state.runs.map(r => `
    <div class="run-item" onclick="loadRunDetail('${r.id}')">
      <div>${r.workflow_name}</div>
      <div style="color:var(--muted)">${r.status} · ${new Date(r.started_at * 1000).toLocaleTimeString()}</div>
    </div>
  `).join('')
}

// ── Render: phase bar ──────────────────────────────────────────────────────
function renderPhaseBar() {
  const bar = document.getElementById('phase-bar')
  if (!state.phases.length) { bar.innerHTML = ''; return }
  bar.innerHTML = state.phases.map((p, i) => `
    ${i > 0 ? '<span class="phase-arrow">→</span>' : ''}
    <span class="phase-pill ${p.status}">${p.phase_id}</span>
  `).join('')
}

// ── Render: agent tree SVG ─────────────────────────────────────────────────
function renderTree() {
  const svg = document.getElementById('tree-svg')
  if (!state.agents.length) { svg.innerHTML = ''; return }

  // Group agents by phase
  const byPhase = {}
  state.agents.forEach(a => {
    if (!byPhase[a.phase_id]) byPhase[a.phase_id] = []
    byPhase[a.phase_id].push(a)
  })

  const phases = Object.keys(byPhase)
  const NODE_W = 160, NODE_H = 48, GAP_X = 200, GAP_Y = 70
  const MARGIN = 40

  let svgContent = ''
  let svgHeight = MARGIN

  phases.forEach((phaseId, pi) => {
    const agents = byPhase[phaseId]
    const x = MARGIN + pi * GAP_X
    agents.forEach((agent, ai) => {
      const y = MARGIN + ai * GAP_Y
      const statusClass = agent.status === 'DONE' ? 'done' : agent.status === 'BLOCKED' ? 'blocked' : 'running'
      const dot = agent.status === 'DONE' ? '✓' : agent.status === 'BLOCKED' ? '✗' : '⟳'
      svgContent += `
        <rect class="node-circle ${statusClass}" x="${x}" y="${y}" width="${NODE_W}" height="${NODE_H}" rx="6"/>
        <text class="node-label" x="${x + NODE_W/2}" y="${y + 18}" text-anchor="middle">${agent.agent}</text>
        <text class="node-status" x="${x + NODE_W/2}" y="${y + 34}" text-anchor="middle">${dot} ${agent.confidence || ''}</text>
      `
      if (pi > 0) {
        const prevX = MARGIN + (pi - 1) * GAP_X + NODE_W
        svgContent += `<line class="tree-edge" x1="${prevX}" y1="${y + NODE_H/2}" x2="${x}" y2="${y + NODE_H/2}"/>`
      }
      svgHeight = Math.max(svgHeight, y + NODE_H + MARGIN)
    })
  })

  const svgWidth = MARGIN + phases.length * GAP_X + NODE_W
  svg.setAttribute('viewBox', `0 0 ${svgWidth} ${svgHeight}`)
  svg.setAttribute('height', svgHeight)
  svg.innerHTML = svgContent
}

// ── Render: active agent card ──────────────────────────────────────────────
function renderActiveCard(agent) {
  const card = document.getElementById('active-card')
  if (!agent) { card.style.display = 'none'; return }
  card.style.display = 'block'
  document.getElementById('active-agent-name').textContent = agent.agent
  document.getElementById('active-agent-summary').textContent = agent.summary || ''
  document.getElementById('active-confidence').textContent = `confidence: ${agent.confidence || '—'}`
  document.getElementById('active-tokens').textContent = agent.tokens_used ? `${agent.tokens_used} tokens` : ''
  const sources = agent.sources || []
  document.getElementById('active-sources').innerHTML = sources.map(s =>
    `<a class="source-tag" href="${s}" target="_blank" rel="noopener">${s.replace(/^https?:\/\//, '').slice(0, 40)}</a>`
  ).join('')
}

// ── Events ─────────────────────────────────────────────────────────────────
function bindEvents() {
  document.getElementById('dispatch-btn').addEventListener('click', dispatch)

  const dropZone = document.getElementById('drop-zone')
  dropZone.addEventListener('dragover', e => { e.preventDefault(); dropZone.classList.add('over') })
  dropZone.addEventListener('dragleave', () => dropZone.classList.remove('over'))
  dropZone.addEventListener('drop', async e => {
    e.preventDefault()
    dropZone.classList.remove('over')
    const file = e.dataTransfer.files[0]
    if (!file) return
    const data = await file.arrayBuffer()
    await fetch(`/api/files/upload?name=${encodeURIComponent(file.name)}`, {
      method: 'POST', body: data,
    })
    dropZone.textContent = `✓ ${file.name} uploaded`
  })
}

// ── Boot ────────────────────────────────────────────────────────────────────
document.addEventListener('DOMContentLoaded', init)
```

- [ ] **Step 4: Verify UI loads**

```bash
cd /Users/kabir/Workspace/claude-team && go run main.go &
sleep 2
curl -s http://localhost:3000/ | grep -c "Anton"
kill %1
```
Expected: `1` or more (title tag found).

---

## Task 18: Update workflow types + parser for new YAML format

**Files:**
- Modify: `internal/workflow/types.go`
- Modify: `internal/workflow/parser.go`
- Test: `internal/workflow/parser_test.go`

**Interfaces:**
- Adds to Workflow: `Phases []WorkflowPhase`
- Adds: `WorkflowPhase` struct

- [ ] **Step 1: Write failing test**

```go
// internal/workflow/parser_test.go — add
func TestParseFeatureBuildWorkflow(t *testing.T) {
    w, err := ParseFile("../../workflows/feature-build.yaml")
    if err != nil {
        t.Fatal(err)
    }
    if len(w.Phases) == 0 {
        t.Error("feature-build.yaml should have phases")
    }
    if w.Phases[0].ID != "planning" {
        t.Errorf("first phase should be planning, got %s", w.Phases[0].ID)
    }
}
```

- [ ] **Step 2: Run — verify FAIL**

```bash
go test ./internal/workflow/... -run TestParseFeatureBuildWorkflow -v
```

- [ ] **Step 3: Add WorkflowPhase to types.go**

```go
// Add to internal/workflow/types.go
type WorkflowPhase struct {
    ID          string   `yaml:"id" json:"id"`
    Coordinator string   `yaml:"coordinator" json:"coordinator"`
    Sequential  []string `yaml:"sequential,omitempty" json:"sequential,omitempty"`
    Parallel    []string `yaml:"parallel,omitempty" json:"parallel,omitempty"`
    Outputs     []string `yaml:"outputs,omitempty" json:"outputs,omitempty"`
    When        string   `yaml:"when,omitempty" json:"when,omitempty"`
}
```

Add `Phases []WorkflowPhase` field to the `Workflow` struct.

- [ ] **Step 4: Verify parser picks up Phases field**

The parser uses `gopkg.in/yaml.v3` which reads struct tags. Adding the field and tag is sufficient — no parser code changes needed.

- [ ] **Step 5: Run tests**

```bash
go test ./internal/workflow/... -v
```
Expected: all PASS including TestParseFeatureBuildWorkflow.

---

## Final verification

- [ ] **Full build**

```bash
cd /Users/kabir/Workspace/claude-team && go build ./... && echo "build OK"
```

- [ ] **Full test suite**

```bash
go test ./... 2>&1 | grep -E "^(ok|FAIL|---)"
```
Expected: all `ok`, zero `FAIL`.

- [ ] **Server starts cleanly**

```bash
go run main.go &
sleep 2 && curl -s http://localhost:3000/api/runs && kill %1
```
Expected: `[]` (empty JSON array, no crash).

- [ ] **MCP server starts cleanly**

```bash
node mcp/team-coordinator.js &
sleep 2 && kill %1 && echo "MCP OK"
```

- [ ] **All role files present**

```bash
ls roles/ | wc -l
```
Expected: `15` (14 specialist roles + `_standards.md`)

```bash
ls coordinators/ | wc -l
```
Expected: `5`

```bash
ls workflows/ | wc -l
```
Expected: `5`

```bash
ls skills/ | wc -l
```
Expected: `3`
