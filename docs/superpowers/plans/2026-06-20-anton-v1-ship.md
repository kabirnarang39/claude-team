# Anton Stable v1 — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship Anton as a stable open source v1 — any engineer can install with one command and run all 5 workflows end-to-end.

**Architecture:** Validate the existing Go + SQLite + WebSocket + Claude Code skill execution chain, fix core reliability issues, then add polish (install script, docs, CI/CD, UI improvements). Tasks 1–2 are sequential and blocking. Tasks 3–8 can run in parallel with each other after Task 1 is complete.

**Tech Stack:** Go 1.22+, SQLite (better-sqlite3), Node.js 20+ (MCP server), vanilla JS/CSS/HTML (UI), GitHub Actions, GoReleaser

## Global Constraints

- Go version: 1.22+ (uses `r.PathValue()` from `net/http`)
- Never modify `.claude-team/` contents directly — always through the store API
- All new API endpoints must be registered in `internal/api/server.go`
- Run `go test ./...` after every Go change — must pass before committing
- No external JS dependencies in the UI — inline only
- `main` branch must always be in a releasable state after each task

---

### Task 1: Validate End-to-End Workflow Execution

**Files:**
- Create: `docs/validation-findings.md`

**Interfaces:**
- Produces: `docs/validation-findings.md` — structured failure log consumed by Task 2

- [ ] **Step 1: Start the server**

```bash
go run main.go
```
Expected output: `Anton running at http://localhost:3000`
If it errors: check Go version (`go version` — needs 1.22+), check `go.mod` module path.

- [ ] **Step 2: Open dashboard and dispatch a test task**

Open `http://localhost:3000`. Enter this task:
```
Create a health check endpoint GET /health that returns {"status":"ok"}
```
Select workflow `feature-build`. Click Dispatch.

Expected: cmd banner appears with `/team-dispatch --from-browser --workflow feature-build`. Run history shows new run with amber dot.

- [ ] **Step 3: Check pre-populated agents in SQLite**

```bash
sqlite3 .claude-team/state.db "SELECT run_id, phase_id, agent, status FROM agent_results ORDER BY id;"
```
Expected: rows with all agents at status `PENDING`, run_id matching the banner.
If run_id is `unknown`: the `CreateRunWithID` → `PrePopulateAgents` path is broken.

- [ ] **Step 4: Execute the workflow in Claude Code**

Open a new Claude Code terminal in this project directory. Run:
```
/team-dispatch --from-browser --workflow feature-build
```
Watch the terminal output for coordinator activity. Let it run for at least one phase.

- [ ] **Step 5: Check agent results in SQLite**

After at least one agent completes (or fails), run:
```bash
sqlite3 .claude-team/state.db "SELECT run_id, phase_id, agent, status, summary FROM agent_results ORDER BY id;"
```
Record: what run_id did results land in? Is it the same as the pre-populated run_id?

- [ ] **Step 6: Check WebSocket delivery**

Watch the dashboard during Step 4. Do agents update from PENDING to RUNNING/DONE in real time?
If not: open browser DevTools → Network → WS → check if messages arrive. If messages arrive but UI doesn't update, bug is in `onAgentResult`. If no messages, watcher or hub is broken.

- [ ] **Step 7: Run all 5 workflows and document failures**

Repeat Steps 2–6 for each workflow: `code-review`, `bug-fix`, `incident-response`, `architecture-review`.
For each: note whether agents transition PENDING → DONE, whether run_id matches, whether outputs land in `.claude-team/runs/<run_id>/`.

- [ ] **Step 8: Document findings**

Create `docs/validation-findings.md`:
```markdown
# Validation Findings — <date>

## Summary
[Pass/Fail per workflow]

## Failures

### [Failure name]
- Symptom: [what was observed]
- Location: [file:line or tool]
- Severity: blocking | polish
- Hypothesis: [why it fails]
```

- [ ] **Step 9: Commit findings**

```bash
git add docs/validation-findings.md
git commit -m "docs: add validation findings from end-to-end test run"
```

---

### Task 2: Fix run_id Propagation and Report Reliability

**Files:**
- Modify: `roles/_standards.md` (clarify run_id source)
- Modify: `coordinators/main.md` (make run_id propagation explicit in every brief)
- Modify: `mcp/team-coordinator.js` (defensive run_id handling)

**Interfaces:**
- Consumes: `docs/validation-findings.md` from Task 1
- Produces: reliable `run_id` in every `agent_results` row matching the active run

> **Before starting this task:** read `docs/validation-findings.md`. If the validation found a different root cause than run_id propagation, fix that root cause instead and skip steps that don't apply.

- [ ] **Step 1: Verify the run_id propagation chain**

The chain: `handleTask` creates run in DB → coordinator reads `pending-task.md` (which has `Run ID: xxx`) → coordinator briefs sub-coordinator with `Run ID: xxx` → sub-coordinator briefs agents with `Run ID: xxx` → agent includes `run_id` in report JSON → MCP `report` writes it to DB.

Check `roles/_standards.md` line 106: confirm `"run_id": "<from-env ANTON_RUN_ID>"`. The current standard says `from-env` which is wrong — sub-agents don't inherit env. The value must come from the brief's `Run ID:` line.

- [ ] **Step 2: Update `_standards.md` to clarify run_id source**

In `roles/_standards.md`, replace the run_id line in the output format:

Old:
```json
"run_id": "<from-env ANTON_RUN_ID>",
```

New:
```json
"run_id": "<Run ID from your brief — e.g. anton-1750420000-a3f2c1>",
```

Also add a note below the JSON block:
```markdown
**run_id rule:** Read it from the `Run ID:` line in your brief. Do NOT use env var — sub-agents do not inherit environment. If no Run ID in brief, ask your coordinator before calling report.
```

- [ ] **Step 3: Update `coordinators/main.md` brief template to be explicit**

In `coordinators/main.md`, find the sub-coordinator brief template and ensure the run_id instruction is unambiguous. Add after the brief block:

```markdown
## run_id Propagation Rule

Pass the EXACT same run_id to every sub-coordinator and every agent you brief.
Format the Run ID line as: `Run ID: <exact-value-you-received>`
Never generate a new run_id. Always forward the one you were given.
Sub-agents must include this value verbatim in their report JSON `run_id` field.
```

- [ ] **Step 4: Add defensive run_id validation in MCP report handler**

In `mcp/team-coordinator.js`, in the `report` handler, add a warning when run_id falls back to 'unknown':

```javascript
if (name === 'report') {
    const result = JSON.parse(args.result)
    const runId = result.run_id || process.env.ANTON_RUN_ID || 'unknown'
    if (runId === 'unknown') {
      console.error('[coordinator] WARNING: report called with no run_id — results will be orphaned. Agent must include run_id in result JSON.')
    }
    const phaseId = result.phase_id || result.phase || 'unknown'
    // ... rest of handler unchanged
```

- [ ] **Step 5: Re-run validation for feature-build**

```bash
# Start server
go run main.go &

# Dispatch via browser, then in Claude Code:
/team-dispatch --from-browser --workflow feature-build
```

After at least one agent completes:
```bash
sqlite3 .claude-team/state.db "SELECT run_id, agent, status FROM agent_results;"
```
Expected: run_id matches the active run (not 'unknown').

- [ ] **Step 6: Commit**

```bash
git add roles/_standards.md coordinators/main.md mcp/team-coordinator.js
git commit -m "fix: clarify run_id propagation — read from brief not env var"
```

---

### Task 3: Add --version Flag and Startup Prerequisites Check

**Files:**
- Modify: `main.go` (add version flag + prereq check)

**Interfaces:**
- Produces: `anton --version` prints `Anton v1.0.0`, startup prints warnings for missing prereqs

- [ ] **Step 1: Write the test**

In `main_test.go` (create if not exists):
```go
package main

import (
    "os/exec"
    "strings"
    "testing"
)

func TestVersionFlag(t *testing.T) {
    out, err := exec.Command("go", "run", ".", "--version").CombinedOutput()
    if err != nil {
        t.Fatalf("--version errored: %v\noutput: %s", err, out)
    }
    if !strings.Contains(string(out), "Anton v") {
        t.Errorf("expected 'Anton v' in output, got: %s", out)
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test -run TestVersionFlag ./...
```
Expected: FAIL (--version flag not defined, program exits with error)

- [ ] **Step 3: Add version constant and --version flag to main.go**

At the top of `main.go`, after the package declaration and before imports, add:
```go
const version = "1.0.0"
```

In `main()`, before `flag.Parse()`, add:
```go
versionFlag := flag.Bool("version", false, "Print version and exit")
```

After `flag.Parse()`, add:
```go
if *versionFlag {
    fmt.Printf("Anton v%s\n", version)
    os.Exit(0)
}
```

- [ ] **Step 4: Add startup prerequisites check**

After `flag.Parse()` and version check, add:
```go
checkPrereqs()
```

Add this function before `main()`:
```go
func checkPrereqs() {
    if os.Getenv("ANTHROPIC_API_KEY") == "" {
        log.Println("warn: ANTHROPIC_API_KEY not set — agents will fail to run")
    }
    if _, err := os.Stat("mcp/node_modules"); os.IsNotExist(err) {
        log.Println("warn: mcp/node_modules not found — run: cd mcp && npm install")
    }
    if _, err := os.Stat("workflows"); os.IsNotExist(err) {
        log.Println("warn: workflows/ directory not found — no workflows available")
    }
}
```

- [ ] **Step 5: Run test to verify it passes**

```bash
go test -run TestVersionFlag ./...
```
Expected: PASS

- [ ] **Step 6: Manual verification**

```bash
go run main.go --version
```
Expected: `Anton v1.0.0`

```bash
go run main.go
```
Expected: startup prints any applicable warnings, then `Anton running at http://localhost:3000`

- [ ] **Step 7: Run full test suite**

```bash
go test ./...
```
Expected: all PASS

- [ ] **Step 8: Commit**

```bash
git add main.go main_test.go
git commit -m "feat: add --version flag and startup prerequisites check"
```

---

### Task 4: Add Agent Output File Viewer API

**Files:**
- Modify: `internal/api/handlers.go` (add two new handlers)
- Modify: `internal/api/server.go` (register two new routes)

**Interfaces:**
- Produces:
  - `GET /api/runs/{id}/files` → `["adr.md","prd.md",...]` (file list)
  - `GET /api/runs/{id}/files/{filename}` → `{"name":"adr.md","content":"# Architecture..."}` (file content)

- [ ] **Step 1: Write the tests**

In `internal/api/server_test.go`, add:
```go
func TestRunFilesEndpoints(t *testing.T) {
    dir := t.TempDir()
    runDir := filepath.Join(dir, ".claude-team", "runs", "test-run-123")
    os.MkdirAll(runDir, 0755)
    os.WriteFile(filepath.Join(runDir, "adr.md"), []byte("# ADR content"), 0644)

    db, _ := store.Open(filepath.Join(dir, "state.db"))
    defer db.Close()
    srv := NewServer(Config{
        Store:      db,
        RuntimeDir: filepath.Join(dir, ".claude-team"),
        UIDir:      dir,
    })
    handler := srv.Handler()

    t.Run("list files", func(t *testing.T) {
        req := httptest.NewRequest("GET", "/api/runs/test-run-123/files", nil)
        w := httptest.NewRecorder()
        handler.ServeHTTP(w, req)
        if w.Code != 200 {
            t.Fatalf("expected 200, got %d: %s", w.Code, w.Body)
        }
        var files []string
        json.NewDecoder(w.Body).Decode(&files)
        if len(files) != 1 || files[0] != "adr.md" {
            t.Errorf("expected [adr.md], got %v", files)
        }
    })

    t.Run("get file content", func(t *testing.T) {
        req := httptest.NewRequest("GET", "/api/runs/test-run-123/files/adr.md", nil)
        w := httptest.NewRecorder()
        handler.ServeHTTP(w, req)
        if w.Code != 200 {
            t.Fatalf("expected 200, got %d: %s", w.Code, w.Body)
        }
        var result map[string]string
        json.NewDecoder(w.Body).Decode(&result)
        if result["content"] != "# ADR content" {
            t.Errorf("unexpected content: %v", result)
        }
    })

    t.Run("path traversal blocked", func(t *testing.T) {
        req := httptest.NewRequest("GET", "/api/runs/test-run-123/files/..%2F..%2Fetc%2Fpasswd", nil)
        w := httptest.NewRecorder()
        handler.ServeHTTP(w, req)
        if w.Code != 400 {
            t.Fatalf("expected 400 for path traversal, got %d", w.Code)
        }
    })
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/api/... -run TestRunFilesEndpoints -v
```
Expected: FAIL — routes not found (404)

- [ ] **Step 3: Add handlers to handlers.go**

At the end of `internal/api/handlers.go`, add:
```go
func (s *Server) handleRunFiles(w http.ResponseWriter, r *http.Request) {
    runID := r.PathValue("id")
    if s.cfg.RuntimeDir == "" {
        http.Error(w, "not configured", 500)
        return
    }
    runDir := filepath.Join(s.cfg.RuntimeDir, "runs", runID)
    entries, err := os.ReadDir(runDir)
    if err != nil {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode([]string{})
        return
    }
    var names []string
    for _, e := range entries {
        if !e.IsDir() {
            names = append(names, e.Name())
        }
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(names)
}

func (s *Server) handleRunFile(w http.ResponseWriter, r *http.Request) {
    runID := r.PathValue("id")
    filename := r.PathValue("filename")
    if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
        http.Error(w, "invalid filename", 400)
        return
    }
    if s.cfg.RuntimeDir == "" {
        http.Error(w, "not configured", 500)
        return
    }
    path := filepath.Join(s.cfg.RuntimeDir, "runs", runID, filepath.Base(filename))
    data, err := os.ReadFile(path)
    if err != nil {
        http.Error(w, "file not found", 404)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "name":    filename,
        "content": string(data),
    })
}
```

- [ ] **Step 4: Register routes in server.go**

In `internal/api/server.go`, in `Handler()`, after the existing run route, add:
```go
mux.HandleFunc("GET /api/runs/{id}/files", s.handleRunFiles)
mux.HandleFunc("GET /api/runs/{id}/files/{filename}", s.handleRunFile)
```

- [ ] **Step 5: Run tests to verify they pass**

```bash
go test ./internal/api/... -run TestRunFilesEndpoints -v
```
Expected: PASS (3 subtests)

- [ ] **Step 6: Run full test suite**

```bash
go test ./...
```
Expected: all PASS

- [ ] **Step 7: Commit**

```bash
git add internal/api/handlers.go internal/api/server.go internal/api/server_test.go
git commit -m "feat: add run output file viewer API endpoints"
```

---

### Task 5: UI Improvements

**Files:**
- Modify: `ui/app.js` (offline banner, error surfacing, output viewer, fix empty state)
- Modify: `ui/index.html` (offline banner element, output viewer panel)
- Modify: `ui/styles.css` (styles for offline banner, output viewer)

**Interfaces:**
- Consumes: `GET /api/runs/{id}/files` and `GET /api/runs/{id}/files/{filename}` from Task 4
- Produces: visible offline banner when server unreachable; clickable agent opens output file panel; meaningful error messages instead of silent failures; first-time empty state with instructions

- [ ] **Step 1: Add offline banner HTML to index.html**

In `ui/index.html`, after the opening `<div class="shell">` and before `<header class="hdr">`, add:
```html
<div class="offline-banner" id="offline-banner" style="display:none">
  Server offline — start Anton with <code>go run main.go</code> then refresh
</div>
```

- [ ] **Step 2: Add output viewer panel HTML to index.html**

In `ui/index.html`, replace the `<div class="active-card" ...>` block with:
```html
<div class="active-card" id="active-card" style="display:none">
  <div class="ac-agent" id="ac-name"></div>
  <div class="ac-summary" id="ac-summary"></div>
  <div class="ac-meta">
    <span id="ac-conf"></span>
    <span id="ac-tokens"></span>
  </div>
  <div class="ac-sources" id="ac-sources"></div>
  <div class="ac-output" id="ac-output" style="display:none">
    <div class="ac-output-label">Output file</div>
    <pre class="ac-output-content" id="ac-output-content"></pre>
  </div>
</div>
```

- [ ] **Step 3: Add styles to styles.css**

At the end of `ui/styles.css`, add:
```css
/* ── Offline banner ── */
.offline-banner{
  background:#1a0808;
  border-bottom:1px solid var(--red);
  color:var(--red);
  font-size:11px;
  padding:7px 18px;
  text-align:center;
  flex-shrink:0;
}
.offline-banner code{
  background:rgba(239,68,68,.15);
  padding:1px 5px;
  border-radius:3px;
  font-family:var(--font);
}

/* ── Output viewer ── */
.ac-output{margin-top:8px}
.ac-output-label{
  font-size:9px;
  color:var(--muted);
  text-transform:uppercase;
  letter-spacing:.06em;
  margin-bottom:4px;
}
.ac-output-content{
  background:var(--bg);
  border:1px solid var(--border);
  border-radius:4px;
  padding:8px 10px;
  font-family:var(--font);
  font-size:10px;
  color:var(--text);
  max-height:180px;
  overflow-y:auto;
  white-space:pre-wrap;
  word-break:break-word;
}
```

- [ ] **Step 4: Add offline detection and banner toggle to app.js**

In `app.js`, update the `loadRuns()` function to show the offline banner on fetch failure:
```javascript
async function loadRuns() {
  try {
    const res = await fetch('/api/runs')
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    state.runs = await res.json() || []
    document.getElementById('offline-banner').style.display = 'none'
  } catch (_) {
    state.runs = []
    document.getElementById('offline-banner').style.display = 'block'
  }
  renderRunHistory()
  if (state.runs.length > 0) await loadRunDetail(state.runs[0].id)
}
```

Also update `dispatch()` to show errors in the banner rather than silently failing. Replace the inner `catch (e)` in `dispatch()`:
```javascript
  } catch (e) {
    document.getElementById('offline-banner').style.display = 'block'
    showCmd(null, `Server unreachable: ${e.message}`, true)
    return
  }
```

- [ ] **Step 5: Fix the first-time empty state in renderTreeSimple**

In `app.js`, in `renderTreeSimple()`, update the empty-state message:
```javascript
  const msg = isRunning
    ? `${state.activeRun.id} — agents working, polling every 6s…`
    : state.runs.length === 0
      ? 'Enter a task in the sidebar and click ▶ Dispatch to start.'
      : 'Select a run from the sidebar to view its agent tree.'
```

- [ ] **Step 6: Add output file loading to selectAgentData**

In `app.js`, at the end of `selectAgentData(ag)`:
```javascript
  // Load output file if run has output files
  const outputEl = document.getElementById('ac-output')
  const outputContent = document.getElementById('ac-output-content')
  outputEl.style.display = 'none'
  if (state.activeRun) {
    fetch(`/api/runs/${state.activeRun.id}/files`)
      .then(r => r.ok ? r.json() : [])
      .then(files => {
        // Match agent name to a file (e.g. backend-engineer → implementation/)
        const match = files.find(f =>
          f.toLowerCase().includes(ag.agent.replace(/-/g, '')) ||
          ag.agent.includes(f.replace(/\.md$/, '').replace(/-/g, ''))
        )
        if (!match) return
        return fetch(`/api/runs/${state.activeRun.id}/files/${encodeURIComponent(match)}`)
      })
      .then(r => r && r.ok ? r.json() : null)
      .then(data => {
        if (!data) return
        outputContent.textContent = data.content
        outputEl.style.display = 'block'
      })
      .catch(() => {})
  }
```

- [ ] **Step 7: Manual verification**

```bash
go run main.go
```
Open `http://localhost:3000`.
- Kill the server. Refresh. Confirm red "Server offline" banner appears.
- Restart server, dispatch a task. After agents complete, click an agent in the tree. If output files exist, confirm the output panel appears.
- First visit (empty DB): confirm helpful empty state text appears instead of blank SVG.

- [ ] **Step 8: Commit**

```bash
git add ui/app.js ui/index.html ui/styles.css
git commit -m "feat: add offline banner, output viewer, and empty state guidance to UI"
```

---

### Task 6: Write install.sh

**Files:**
- Create: `install.sh`

**Interfaces:**
- Produces: single-command install that sets up binary, skills, and MCP for macOS arm64/amd64 + Linux amd64

- [ ] **Step 1: Write install.sh**

Create `install.sh`:
```bash
#!/usr/bin/env bash
set -euo pipefail

REPO="https://github.com/<org>/claude-team"
VERSION="v1.0.0"
INSTALL_DIR="$HOME/.local/bin"
SKILL_DIR="$HOME/.claude/skills"

# ── Detect platform ──────────────────────────────────────────────────────────
OS="$(uname -s)"
ARCH="$(uname -m)"
case "$OS-$ARCH" in
  Darwin-arm64)  PLATFORM="darwin-arm64"  ;;
  Darwin-x86_64) PLATFORM="darwin-amd64"  ;;
  Linux-x86_64)  PLATFORM="linux-amd64"   ;;
  *)
    echo "ERROR: Unsupported platform $OS-$ARCH"
    echo "Supported: macOS arm64/amd64, Linux amd64"
    exit 1
    ;;
esac

# ── Check prerequisites ──────────────────────────────────────────────────────
check() {
  command -v "$1" &>/dev/null || {
    echo "ERROR: $1 is required but not installed."
    echo "Install: $2"
    exit 1
  }
}
check claude  "https://claude.ai/download"
check node    "https://nodejs.org"
check npm     "https://nodejs.org"

if [[ -z "${ANTHROPIC_API_KEY:-}" ]]; then
  echo "WARN: ANTHROPIC_API_KEY is not set — agents will fail to run."
  echo "      Set it with: export ANTHROPIC_API_KEY=sk-ant-..."
fi

# ── Download binary ──────────────────────────────────────────────────────────
echo "Installing Anton $VERSION for $PLATFORM..."
mkdir -p "$INSTALL_DIR"
BINARY_URL="$REPO/releases/download/$VERSION/anton-$PLATFORM"
curl -fsSL "$BINARY_URL" -o "$INSTALL_DIR/anton"
chmod +x "$INSTALL_DIR/anton"

# ── Install skills ───────────────────────────────────────────────────────────
echo "Installing Anton skills..."
mkdir -p "$SKILL_DIR"
SKILLS_URL="$REPO/releases/download/$VERSION/skills.tar.gz"
curl -fsSL "$SKILLS_URL" | tar -xz -C "$SKILL_DIR"

# ── Install MCP server ───────────────────────────────────────────────────────
echo "Installing MCP server dependencies..."
MCP_DIR="$HOME/.claude/anton-mcp"
mkdir -p "$MCP_DIR"
MCP_URL="$REPO/releases/download/$VERSION/mcp.tar.gz"
curl -fsSL "$MCP_URL" | tar -xz -C "$MCP_DIR"
(cd "$MCP_DIR" && npm install --silent)

# ── PATH check ───────────────────────────────────────────────────────────────
if ! command -v anton &>/dev/null; then
  echo ""
  echo "NOTE: Add $INSTALL_DIR to your PATH:"
  echo "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.zshrc && source ~/.zshrc"
fi

echo ""
echo "Anton $VERSION installed successfully."
echo "Start the dashboard: anton"
echo "Then in Claude Code: /team-dispatch <your task>"
```

- [ ] **Step 2: Make it executable**

```bash
chmod +x install.sh
```

- [ ] **Step 3: Dry-run test (don't curl | sh — read it)**

```bash
bash -n install.sh
```
Expected: no syntax errors (exits 0)

- [ ] **Step 4: Test on a clean path with fake binary dir**

```bash
INSTALL_DIR=/tmp/anton-test-install bash -c '
  # Stub out curl and npm to avoid network
  curl() { echo "stubbed"; }
  npm() { echo "stubbed"; }
  export -f curl npm
  # Run with stubs — will fail at curl but validates prereq checks
  ANTHROPIC_API_KEY="" bash -n install.sh
  echo "syntax OK"
'
```
Expected: `syntax OK`

- [ ] **Step 5: Commit**

```bash
git add install.sh
git commit -m "feat: add one-command install script for macOS and Linux"
```

---

### Task 7: Write README.md and CHANGELOG.md

**Files:**
- Create: `README.md`
- Create: `CHANGELOG.md`
- Create: `docs/troubleshooting.md`

**Interfaces:**
- Produces: complete project documentation consumable by all three user audiences (power users, engineers, leads)

- [ ] **Step 1: Write README.md**

Create `README.md`:
```markdown
# Anton

Multi-agent engineering team coordinator for Claude Code. Dispatch a task — Anton runs a full team of specialist AI agents (planner, architect, engineers, QA, security) and shows live progress in your browser.

## Quick Start

**Install:**
```bash
curl -fsSL https://raw.githubusercontent.com/<org>/claude-team/main/install.sh | sh
```

**Start the dashboard:**
```bash
anton
# → Anton running at http://localhost:3000
```

**Dispatch a task** (in Claude Code, in the same directory):
```
/team-dispatch build user authentication with JWT and refresh tokens
```

Or use the browser — go to `http://localhost:3000`, enter your task, click Dispatch, then run the command shown.

## Workflows

| Workflow | What it does | Agents |
|----------|-------------|--------|
| `feature-build` | Full cycle: planning → architecture → engineering → QA → DevOps | 11 agents |
| `code-review` | Architecture + security + quality review in parallel | 3 agents |
| `bug-fix` | Root cause analysis → fix → verify | 3 agents |
| `incident-response` | Triage → hotfix → verify → post-mortem | 4 agents |
| `architecture-review` | Design doc review → ADR | 2 agents |

## Requirements

- [Claude Code](https://claude.ai/download) with an active subscription or `ANTHROPIC_API_KEY`
- Node.js 20+
- Go 1.22+ (for building from source)

## Build From Source

```bash
git clone https://github.com/<org>/claude-team
cd claude-team
cd mcp && npm install && cd ..
go run main.go
```

## How It Works

```
You → /team-dispatch → Main Coordinator (your Claude Code session)
                              │
             ┌────────────────┼────────────────┐
             ▼                ▼                ▼
       Planning Sub-    Engineering      QA Sub-Coordinator
       Coordinator      Sub-Coordinator
             │                │                │
       requirements-   senior-architect   qa-engineer
       analyst         api-designer       security-reviewer
       tech-writer     backend-engineer   e2e-tester
                        frontend-engineer
                        dba
                              │
                        coordinator MCP     ← writes results
                              │
                        SQLite state.db     ← source of truth
                              │
                        Go HTTP server      ← reads + WebSocket
                              │
                        Browser dashboard   ← live agent tree
```

Agent results are written to `.claude-team/runs/<run_id>/` and visible in the dashboard.

## Troubleshooting

See [docs/troubleshooting.md](docs/troubleshooting.md).

## Contributing

PRs welcome. Run `go test ./...` before opening a PR.

## License

MIT
```

- [ ] **Step 2: Write CHANGELOG.md**

Create `CHANGELOG.md`:
```markdown
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
```

- [ ] **Step 3: Write docs/troubleshooting.md**

Create `docs/troubleshooting.md`:
```markdown
# Troubleshooting

## Agents stuck at PENDING

**Symptom:** Dashboard shows agents but they never transition from PENDING.
**Cause:** The browser saves the task but agents only run when you execute `/team-dispatch` in Claude Code.
**Fix:** Open Claude Code in the same directory as Anton, then run:
```
/team-dispatch --from-browser --workflow feature-build
```

## run_id shows as "unknown" in SQLite

**Symptom:** `sqlite3 .claude-team/state.db "SELECT run_id FROM agent_results;"` returns `unknown`.
**Cause:** Agent did not receive the Run ID in its brief, or did not include it in the report JSON.
**Fix:** Ensure the coordinator brief includes `Run ID: <exact-id>`. See `roles/_standards.md` for the required report format.

## Dashboard shows blank / no connection

**Symptom:** Opening `http://localhost:3000` shows the offline banner or blank page.
**Cause:** The Anton server is not running.
**Fix:** Start it with `go run main.go` or `anton` if installed.

## MCP coordinator not found

**Symptom:** Agents fail with "coordinator tool not available" or similar.
**Cause:** MCP npm dependencies not installed.
**Fix:**
```bash
cd mcp && npm install
```

## ANTHROPIC_API_KEY not set

**Symptom:** Agents immediately fail with authentication errors.
**Fix:**
```bash
export ANTHROPIC_API_KEY=sk-ant-...
```
Add to your shell profile (`~/.zshrc` or `~/.bashrc`) to persist.

## Workflow not found

**Symptom:** Error "workflow not found: feature-build".
**Cause:** Running from the wrong directory, or `workflows/` not present.
**Fix:** Run Anton from the root of the `claude-team` directory (where `workflows/` lives).
```

- [ ] **Step 4: Commit**

```bash
git add README.md CHANGELOG.md docs/troubleshooting.md
git commit -m "docs: add README, CHANGELOG, and troubleshooting guide for v1"
```

---

### Task 8: GoReleaser and GitHub Actions

**Files:**
- Create: `.goreleaser.yaml`
- Create: `.github/workflows/ci.yml`
- Create: `.github/workflows/release.yml`

**Interfaces:**
- Produces: automated CI on PRs, automated binary releases on `v*` tags

- [ ] **Step 1: Write .goreleaser.yaml**

Create `.goreleaser.yaml`:
```yaml
version: 2

before:
  hooks:
    - go mod tidy

builds:
  - id: anton
    main: .
    binary: anton
    env:
      - CGO_ENABLED=0
    goos:
      - darwin
      - linux
    goarch:
      - amd64
      - arm64
    ignore:
      - goos: linux
        goarch: arm64
    ldflags:
      - -s -w -X main.version={{.Version}}

archives:
  - id: binary
    format: binary
    name_template: "anton-{{ .Os }}-{{ .Arch }}"

checksum:
  name_template: "checksums.txt"

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"
      - "^chore:"
```

- [ ] **Step 2: Write .github/workflows/ci.yml**

Create `.github/workflows/ci.yml`:
```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"
      - name: Install dependencies
        run: go mod download
      - name: Run tests
        run: go test ./...
      - name: Build
        run: go build -o /dev/null .
```

- [ ] **Step 3: Write .github/workflows/release.yml**

Create `.github/workflows/release.yml`:
```yaml
name: Release

on:
  push:
    tags:
      - "v*"

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"
      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v6
        with:
          distribution: goreleaser
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

- [ ] **Step 4: Verify GoReleaser config locally**

Install goreleaser if not present: `brew install goreleaser`

```bash
goreleaser check
```
Expected: `config is valid` (or similar — no errors)

```bash
goreleaser build --snapshot --clean
```
Expected: binaries appear in `dist/` directory for darwin-arm64, darwin-amd64, linux-amd64.

- [ ] **Step 5: Commit**

```bash
git add .goreleaser.yaml .github/workflows/ci.yml .github/workflows/release.yml
git commit -m "ci: add GitHub Actions CI and GoReleaser release pipeline"
```

---

### Task 9: Pre-release Gate Verification

**Files:**
- Modify: `docs/validation-findings.md` (add final verification results)

- [ ] **Step 1: Verify all 5 workflows run end-to-end**

For each workflow: `feature-build`, `code-review`, `bug-fix`, `incident-response`, `architecture-review`:
```bash
# Start server
go run main.go &

# Dispatch via browser, then in Claude Code:
/team-dispatch <task> --workflow <workflow-name>

# Check results
sqlite3 .claude-team/state.db "SELECT agent, status FROM agent_results WHERE run_id = '<id>';"
```
Expected for each: all agents reach status `DONE` or `DONE_WITH_CONCERNS` (not stuck at PENDING).

- [ ] **Step 2: Verify fresh install on a clean machine**

On a machine with only Claude Code and Node.js installed (no Go, no clone):

```bash
curl -fsSL https://raw.githubusercontent.com/<org>/claude-team/main/install.sh | sh
```
Follow the quickstart in README.md verbatim. Dispatch a task. Confirm agents run and dashboard updates.

- [ ] **Step 3: Verify dashboard live updates**

During an active run, confirm the dashboard updates without manual refresh. Agent nodes should change color as they transition PENDING → RUNNING → DONE.

- [ ] **Step 4: Verify run history persists**

```bash
go run main.go   # start server, dispatch a task, let it run
# Kill server with Ctrl-C
go run main.go   # restart
```
Open `http://localhost:3000`. Confirm completed run appears in Run History sidebar.

- [ ] **Step 5: Verify --version**

```bash
go run main.go --version
```
Expected: `Anton v1.0.0`

- [ ] **Step 6: Run full test suite**

```bash
go test ./...
```
Expected: all PASS, 0 failures

- [ ] **Step 7: Verify README quickstart**

Follow the README quickstart section verbatim on a fresh checkout. Every step should work without requiring knowledge not in the README.

- [ ] **Step 8: Tag and release**

When all above pass:
```bash
git tag v1.0.0
git push origin v1.0.0
```
This triggers `.github/workflows/release.yml` → GoReleaser → GitHub Release with binaries.

Verify on GitHub: release appears, binaries are attached for darwin-arm64, darwin-amd64, linux-amd64, checksums.txt is present.

- [ ] **Step 9: Final commit**

```bash
git add docs/validation-findings.md
git commit -m "docs: final pre-release gate verification results"
```
