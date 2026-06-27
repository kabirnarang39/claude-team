package main

import (
	"context"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/kabirnarang39/claude-team/internal/api"
	"github.com/kabirnarang39/claude-team/internal/mcp"
	"github.com/kabirnarang39/claude-team/internal/store"
	"github.com/kabirnarang39/claude-team/internal/workflow"
)

//go:embed all:ui
var uiFS embed.FS

var version = "dev"

func mustSubFS(f embed.FS, dir string) fs.FS {
	sub, err := fs.Sub(f, dir)
	if err != nil {
		panic(err)
	}
	return sub
}

// mcpDir returns the expected location of the installed MCP coordinator.
func mcpDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude", "anton-mcp")
}

// skillsDir returns the expected location of installed Anton skills.
func skillsDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude", "skills")
}

// ensureProjectMCP writes (or merges) the Anton MCP entry into
// .claude/settings.json in the current project directory, so that
// Claude Code picks it up automatically when the project is opened.
func ensureProjectMCP(projectPath, dbPath string) {
	mcpJS := filepath.Join(mcpDir(), "team-coordinator.js")
	if _, err := os.Stat(mcpJS); os.IsNotExist(err) {
		return // MCP not installed globally — skip silently
	}

	claudeDir := filepath.Join(projectPath, ".claude")
	settingsFile := filepath.Join(claudeDir, "settings.json")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return
	}

	// Read existing settings (if any)
	var settings map[string]interface{}
	if data, err := os.ReadFile(settingsFile); err == nil {
		json.Unmarshal(data, &settings) //nolint:errcheck
	}
	if settings == nil {
		settings = map[string]interface{}{}
	}

	// Ensure mcpServers map exists
	mcpServers, _ := settings["mcpServers"].(map[string]interface{})
	if mcpServers == nil {
		mcpServers = map[string]interface{}{}
	}

	// Write (or overwrite) the anton-coordinator entry with the project DB path
	mcpServers["anton-coordinator"] = map[string]interface{}{
		"command": "node",
		"args":    []string{mcpJS},
		"env":     map[string]string{"ANTON_DB_PATH": dbPath},
	}
	settings["mcpServers"] = mcpServers

	// Register Anton skills with absolute paths so Claude Code picks them up as
	// slash commands regardless of which directory the user runs `claude` from.
	sDir := skillsDir()
	antonSkills := []string{
		filepath.Join(sDir, "team-dispatch"),
		filepath.Join(sDir, "team-resume"),
		filepath.Join(sDir, "team-status"),
		filepath.Join(sDir, "team-stop"),
	}
	existingSkills, _ := settings["skills"].([]interface{})
	skillSet := make(map[string]bool)
	for _, s := range existingSkills {
		if sv, ok := s.(string); ok {
			skillSet[sv] = true
		}
	}
	for _, sk := range antonSkills {
		if !skillSet[sk] {
			existingSkills = append(existingSkills, sk)
		}
	}
	settings["skills"] = existingSkills

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return
	}
	if err := os.WriteFile(settingsFile, data, 0644); err != nil {
		log.Printf("warn: could not write .claude/settings.json: %v", err)
	}
}

func antonSkillFiles(root string) []string {
	names := []string{"team-dispatch", "team-resume", "team-status", "team-stop"}
	files := make([]string, 0, len(names))
	for _, name := range names {
		files = append(files, filepath.Join(root, name, "SKILL.md"))
	}
	return files
}

// runCheck prints a setup health report and exits.
func runCheck(projectPath string, _ string) {
	ok := true
	check := func(label, detail string, pass bool) {
		if pass {
			fmt.Printf("  ✓  %-30s %s\n", label, detail)
		} else {
			fmt.Printf("  ✗  %-30s %s\n", label, detail)
			ok = false
		}
	}

	fmt.Println("Anton setup check")
	fmt.Println()

	// Binary on PATH
	_, pathErr := os.Stat("/proc/self/exe") // just a reachability check; binary is running
	check("anton binary", "running", pathErr == nil || true)

	// Node.js
	nodeOK := false
	if _, err := os.Stat("/usr/local/bin/node"); err == nil {
		nodeOK = true
	} else if _, err := os.Stat("/opt/homebrew/bin/node"); err == nil {
		nodeOK = true
	} else if _, err := os.Stat("/usr/bin/node"); err == nil {
		nodeOK = true
	}
	check("node installed", "required for MCP coordinator", nodeOK)

	// MCP coordinator
	mcpJS := filepath.Join(mcpDir(), "team-coordinator.js")
	_, mcpErr := os.Stat(mcpJS)
	check("MCP coordinator installed", mcpJS, mcpErr == nil)

	// MCP node_modules
	mcpMods := filepath.Join(mcpDir(), "node_modules")
	_, modsErr := os.Stat(mcpMods)
	check("MCP dependencies", "node_modules present", modsErr == nil)

	// Skills
	for _, skillFile := range antonSkillFiles(skillsDir()) {
		_, skillErr := os.Stat(skillFile)
		name := filepath.Base(filepath.Dir(skillFile))
		check(name+" skill", skillFile, skillErr == nil)
	}

	// Project .claude/settings.json MCP entry
	settingsFile := filepath.Join(projectPath, ".claude", "settings.json")
	mcpConfigured := false
	if data, err := os.ReadFile(settingsFile); err == nil {
		var s map[string]interface{}
		if json.Unmarshal(data, &s) == nil {
			if servers, ok := s["mcpServers"].(map[string]interface{}); ok {
				_, mcpConfigured = servers["anton-coordinator"]
			}
		}
	}
	check("MCP registered in project", ".claude/settings.json", mcpConfigured)

	// Global anton data dirs
	home, _ := os.UserHomeDir()
	globalAntonDir := filepath.Join(home, ".claude", "anton")

	wfDir := filepath.Join(globalAntonDir, "workflows")
	_, wfErr := os.Stat(wfDir)
	check("workflows/ directory", wfDir, wfErr == nil)

	coordDir := filepath.Join(globalAntonDir, "coordinators")
	_, coordErr := os.Stat(coordDir)
	check("coordinators/ directory", coordDir, coordErr == nil)

	rolesDir := filepath.Join(globalAntonDir, "roles")
	_, rolesErr := os.Stat(rolesDir)
	check("roles/ directory", rolesDir, rolesErr == nil)

	registryFile := filepath.Join(globalAntonDir, "mcp-registry.yaml")
	_, registryErr := os.Stat(registryFile)
	check("MCP registry", registryFile, registryErr == nil)

	fmt.Println()
	if ok {
		fmt.Println("All checks passed. Run `anton` then open http://localhost:3000")
	} else {
		fmt.Println(checkRetryMessage())
		fmt.Printf("To register MCP for this project: anton (just run it — auto-configures on start)\n")
		os.Exit(1)
	}
}

func checkRetryMessage() string {
	return "Fix the items above. Re-run `anton --check` to verify."
}

// seedDemoRun inserts a realistic completed run so first-time visitors
// can see what the dashboard looks like without running a real workflow.
func seedDemoRun(db *store.Store, runtimeDir string) error {
	runID := "demo-feature-build-jwt-auth"
	if err := db.CreateRunWithID(runID, "feature-build"); err != nil {
		return err
	}
	runDir := filepath.Join(runtimeDir, "runs", runID)
	if err := os.MkdirAll(runDir, 0755); err != nil {
		return err
	}
	if err := writeDemoFiles(runDir); err != nil {
		return err
	}
	type agent struct {
		phase        string
		name         string
		summary      string
		conf         string
		tokens       int
		deliverables []string
		sources      []string
		testsRun     string
	}
	agents := []agent{
		{"planning", "requirements-analyst", "Defined scope, non-goals, and acceptance criteria for email/password auth with rotating refresh tokens.", "HIGH", 3200, []string{"prd.md", "acceptance-criteria.md"}, nil, ""},
		{"planning", "tech-writer", "Turned the requirements into a reviewer-friendly PRD and captured open product decisions.", "HIGH", 2800, []string{"prd.md"}, []string{"acceptance-criteria.md"}, ""},
		{"architecture", "senior-architect", "Selected short-lived JWT access tokens plus server-side hashed refresh tokens, with key-rotation and revocation notes.", "HIGH", 4100, []string{"adr.md", "architecture.md"}, []string{"prd.md", "acceptance-criteria.md"}, ""},
		{"architecture", "api-designer", "Documented the auth API surface, response shapes, rate-limit headers, and error model.", "HIGH", 3500, []string{"api-contract.md"}, []string{"acceptance-criteria.md", "adr.md"}, ""},
		{"engineering", "backend-engineer", "Prepared the backend implementation plan for token services, middleware, refresh rotation, and audit events.", "HIGH", 5200, []string{"implementation-plan.md", "api-contract.md"}, []string{"adr.md", "architecture.md"}, ""},
		{"engineering", "frontend-engineer", "Prepared the client integration plan for login, session refresh, protected routes, and logout behavior.", "HIGH", 3900, []string{"implementation-plan.md"}, []string{"api-contract.md", "prd.md"}, ""},
		{"engineering", "dba", "Drafted the relational schema for users, refresh-token families, roles, and audit events.", "MEDIUM", 2900, []string{"database-schema.sql"}, []string{"adr.md", "acceptance-criteria.md"}, ""},
		{"qa", "qa-engineer", "Prepared a QA report with integration, API, and regression test coverage for the auth flow.", "HIGH", 3800, []string{"qa-report.md"}, []string{"acceptance-criteria.md", "api-contract.md"}, "Demo fixture only: no project tests were executed. Recommended commands are listed in qa-report.md."},
		{"qa", "security-reviewer", "Reviewed the design against common auth risks and recorded concrete mitigations and open checks.", "HIGH", 4200, []string{"security-report.md"}, []string{"adr.md", "architecture.md", "database-schema.sql"}, ""},
		{"qa", "e2e-tester", "Added browser-flow scenarios to the QA plan for register, login, refresh, protected route, and logout.", "HIGH", 3100, []string{"qa-report.md"}, []string{"acceptance-criteria.md", "api-contract.md"}, "Demo fixture only: Playwright scenarios were specified, not executed."},
		{"devops", "code-reviewer", "Reviewed the planned deliverables against the PRD, ADR, security notes, and QA coverage.", "HIGH", 2600, []string{"review-report.md"}, []string{"prd.md", "adr.md", "security-report.md", "qa-report.md"}, "Demo fixture only: review confirms planned test coverage, not executed project tests."},
		{"devops", "devops-engineer", "Documented runtime and deployment considerations for Redis, secrets, and CI gates.", "HIGH", 3000, []string{"deployment-notes.md"}, []string{"architecture.md", "security-report.md"}, ""},
	}
	pairs := []store.PhaseAgentPair{
		{PhaseID: "planning", Agents: []string{"requirements-analyst", "tech-writer"}},
		{PhaseID: "architecture", Agents: []string{"senior-architect", "api-designer"}},
		{PhaseID: "engineering", Agents: []string{"backend-engineer", "frontend-engineer", "dba"}},
		{PhaseID: "qa", Agents: []string{"qa-engineer", "security-reviewer", "e2e-tester"}},
		{PhaseID: "devops", Agents: []string{"code-reviewer", "devops-engineer"}},
	}
	db.PrePopulateAgents(runID, pairs)
	for _, a := range agents {
		db.UpsertAgentResult(store.AgentResult{ //nolint:errcheck
			RunID:        runID,
			PhaseID:      a.phase,
			Agent:        a.name,
			Status:       "DONE",
			Confidence:   a.conf,
			Summary:      a.summary,
			Deliverables: a.deliverables,
			Sources:      a.sources,
			TestsRun:     a.testsRun,
			TokensUsed:   a.tokens,
		})
	}
	return db.CompleteRun(runID)
}

func writeDemoFiles(runDir string) error {
	files := map[string]string{
		"approach.md": `# Demo Approach

## Context Source
- Local demo fixture seeded by anton --demo

## Output Configuration
- Deliverable destination: .claude-team/runs/demo-feature-build-jwt-auth/

## Clarifications
- Product scope: Email/password auth only; OAuth and MFA stay out of scope.
- Token transport: Use HTTP-only cookies in the implementation plan.
- Demo status: Dashboard preview data; no external project was modified.

### Option 1 (chosen): JWT access token plus rotating refresh token
**Why this fits:** keeps normal API authorization stateless while preserving server-side session revocation.
**Trade-off:** refresh requires database or cache coordination.

### Option 2: Opaque session token
**Why this fits:** simpler revocation and less JWT key-management work.
**Trade-off:** every protected request needs a session lookup.
`,
		"prd.md": `# Product Requirements: JWT Auth

## Goal

Add email/password authentication for a web application that needs short-lived API authorization and revocable long-lived sessions.

## Users

- New users can create an account.
- Returning users can sign in and stay signed in across browser refreshes.
- Operators can identify suspicious refresh-token reuse in audit logs.

## In Scope

- Registration and login endpoints.
- Short-lived JWT access tokens.
- Rotating refresh tokens stored only as hashes.
- Logout and session revocation.
- Rate limits for login and refresh.
- Audit events for auth-sensitive actions.

## Out Of Scope

- OAuth providers.
- MFA enrollment.
- Password reset email delivery.
- Admin user management.

## Open Decisions

- Exact access-token TTL.
- Remembered-device lifetime.
- Production key-rotation runbook owner.
`,
		"acceptance-criteria.md": `# Acceptance Criteria

1. Given a new user submits a valid email and password, when registration succeeds, then the API creates the user and returns an authenticated session.
2. Given an existing user submits valid credentials, when login succeeds, then the API returns a short-lived access token and a refresh token cookie.
3. Given an access token expires and the refresh token is valid, when refresh is requested, then the API rotates the refresh token and returns a new token pair.
4. Given a refresh token has already been used, when it is submitted again, then the API rejects it and records a security event.
5. Given a user logs out, when the current refresh token is revoked, then later refresh attempts fail.
6. Given invalid credentials are submitted repeatedly, when the threshold is reached, then the API returns a rate-limit response.
7. Given a protected endpoint receives no valid access token, when middleware evaluates the request, then the response is unauthorized.
8. Given JWT keys rotate, when old tokens are past the grace window, then validation fails.
`,
		"adr.md": `# ADR: JWT Access Tokens With Rotating Refresh Tokens

## Status

Accepted for the demo plan.

## Context

The application needs stateless authorization for normal API requests while retaining server-side control over long-lived sessions.

## Decision

Use short-lived signed JWT access tokens and store hashed refresh tokens server-side. Refresh tokens rotate on every use. Reuse of an old refresh token is treated as a possible theft signal and revokes the token family.

## Consequences

Positive:

- API authorization can remain stateless for normal requests.
- Refresh-token rotation gives the server a revocation point.
- Token theft is easier to detect than with non-rotating refresh tokens.

Trade-offs:

- The refresh path requires a database or cache lookup.
- Key rotation and token-family invalidation need operational runbooks.
- Clients must handle refresh failure by returning to login.
`,
		"architecture.md": `# Auth Architecture

## Components

- Auth API: registration, login, refresh, logout, and current-user endpoints.
- JWT signer/verifier: RS256 keys with key IDs and a rotation grace window.
- Refresh token store: hashed tokens grouped by token family.
- Audit log: append-only auth events with request correlation IDs.
- Client session layer: protected-route guard and refresh-on-401 behavior.

## Flow

1. Login verifies credentials and writes a refresh-token family record.
2. The API returns a short-lived access token and an HTTP-only refresh cookie.
3. Protected requests validate the JWT signature, issuer, audience, expiry, and subject.
4. Refresh verifies the current token hash, rotates it transactionally, and returns a fresh pair.
5. Reuse of a rotated token revokes the family and records a security event.

## Operational Notes

- Keep signing keys out of the repository.
- Monitor refresh-token reuse events.
- Run key rotation in a staged grace window.
`,
		"api-contract.md": `# Auth API Contract

| Method | Path | Purpose |
| --- | --- | --- |
| POST | /auth/register | Create a user and session |
| POST | /auth/login | Create a session for valid credentials |
| POST | /auth/refresh | Rotate refresh token and issue fresh access token |
| DELETE | /auth/logout | Revoke current refresh token |
| GET | /auth/me | Return current user profile |

## Error Shape

` + "```json" + `
{
  "error": {
    "code": "invalid_credentials",
    "message": "Email or password is incorrect"
  }
}
` + "```" + `

## Headers

- X-RateLimit-Limit
- X-RateLimit-Remaining
- X-RateLimit-Reset
`,
		"implementation-plan.md": `# Implementation Plan

## Backend

- Add password hashing with bcrypt cost configured per environment.
- Add JWT signer/verifier with key ID support.
- Add refresh-token family persistence and transactional rotation.
- Add auth middleware for protected routes.
- Add audit events for login, logout, refresh, token reuse, and failed verification.

## Frontend

- Add login and registration forms.
- Add a session provider that refreshes on 401 once before redirecting.
- Add protected-route wrappers.
- Add logout action that calls the API and clears local session state.

## Review Gates

- Product reviews token lifetime policy.
- Security reviews cookie attributes and CSRF posture.
- QA reviews refresh-token race coverage.
`,
		"database-schema.sql": `-- Demo schema for planned JWT auth work.
CREATE TABLE users (
  id TEXT PRIMARY KEY,
  email TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  created_at INTEGER NOT NULL
);

CREATE TABLE refresh_token_families (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES users(id),
  revoked_at INTEGER,
  created_at INTEGER NOT NULL
);

CREATE TABLE refresh_tokens (
  id TEXT PRIMARY KEY,
  family_id TEXT NOT NULL REFERENCES refresh_token_families(id),
  token_hash TEXT NOT NULL UNIQUE,
  expires_at INTEGER NOT NULL,
  used_at INTEGER,
  created_at INTEGER NOT NULL
);

CREATE TABLE auth_audit_events (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id TEXT,
  event_type TEXT NOT NULL,
  request_id TEXT,
  created_at INTEGER NOT NULL
);
`,
		"qa-report.md": `# QA Report

## Demo Status

This is seeded demo evidence for dashboard review. No project repository was tested by anton --demo.

## Test Matrix

| Layer | Method | Result |
| --- | --- | --- |
| Unit | JWT signer/verifier, password policy, refresh-token hashing | Specified |
| Integration | register, login, refresh, logout, protected route | Specified |
| Security regression | refresh-token reuse, revoked family, expired token | Specified |
| Browser | register to protected route to logout | Specified |

## Recommended Commands

` + "```bash" + `
go test ./internal/auth ./internal/http
npm test -- auth
npx playwright test auth.spec.ts
` + "```" + `

## Risks To Test Carefully

- Concurrent refresh requests must not create two valid child tokens.
- Cookie attributes must match the chosen CSRF strategy.
- Rate-limit behavior should be deterministic enough for automated tests.
`,
		"security-report.md": `# Security Report

## Findings

| Severity | Area | Recommendation |
| --- | --- | --- |
| Medium | Refresh rotation race | Use a transaction or compare-and-swap update for token rotation |
| Medium | CSRF posture | Pair HTTP-only refresh cookies with SameSite and CSRF review |
| Low | Audit schema | Add event type enum and request correlation ID |

## Positive Notes

- Refresh tokens are stored only as hashes.
- Token-family revocation is defined for reuse detection.
- OAuth, MFA, and password reset are explicitly out of scope for this slice.

## Open Checks

- Confirm production key storage and rotation ownership.
- Confirm access-token TTL and remembered-device lifetime.
`,
		"review-report.md": `# Code Review Report

## Summary

The planned deliverables are internally consistent for a first auth slice. The main launch blocker before implementation is choosing cookie attributes and documenting CSRF expectations.

## Review Notes

| Area | Status | Note |
| --- | --- | --- |
| Requirements | Pass | Acceptance criteria map to API endpoints and security checks |
| Architecture | Pass | ADR names both benefits and trade-offs |
| QA | Follow-up | Test commands are specified, but this demo did not execute project tests |
| Security | Follow-up | CSRF and refresh race checks need implementation proof |

## Required Before Merge

- Add automated tests for refresh-token reuse.
- Add automated tests for rate-limit behavior.
- Document key rotation procedure.
`,
		"deployment-notes.md": `# Deployment Notes

## Runtime Dependencies

- Database tables for users, refresh-token families, refresh tokens, and audit events.
- Redis or database-backed rate limiting.
- Secret storage for JWT private keys.

## CI Gates

- Backend unit tests for token logic.
- API integration tests for auth endpoints.
- Browser smoke test for login and logout.
- Static checks for accidental key material in the repository.
`,
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(runDir, name), []byte(strings.TrimSpace(content)+"\n"), 0644); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	port := flag.Int("port", 3000, "HTTP port")
	registryPath := flag.String("registry", "mcp-registry.yaml", "Path to mcp-registry.yaml")
	versionFlag := flag.Bool("version", false, "Print version and exit")
	checkFlag := flag.Bool("check", false, "Check Anton setup and exit")
	demoFlag := flag.Bool("demo", false, "Pre-populate dashboard with a sample completed run")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("Anton v%s\n", version)
		os.Exit(0)
	}

	projectPath, _ := filepath.Abs(".")
	runtimeDir := filepath.Join(projectPath, ".claude-team")
	dbPath := filepath.Join(runtimeDir, "state.db")

	if *checkFlag {
		runCheck(projectPath, dbPath)
		return
	}

	_ = demoFlag // demo mode handled after DB open

	workflowsDir := filepath.Join(runtimeDir, "workflows")
	globalHome, _ := os.UserHomeDir()
	globalWorkflowsDir := filepath.Join(globalHome, ".claude", "anton", "workflows")

	for _, d := range []string{
		runtimeDir,
		filepath.Join(runtimeDir, "runs"),
		filepath.Join(runtimeDir, "uploads"),
		workflowsDir,
	} {
		_ = os.MkdirAll(d, 0755)
	}

	// Auto-register MCP in project .claude/settings.json so Claude Code picks it up.
	ensureProjectMCP(projectPath, dbPath)

	db, err := store.Open(dbPath)
	if err != nil {
		log.Fatal("open db:", err)
	}
	defer func() { _ = db.Close() }()

	if *demoFlag {
		if err := seedDemoRun(db, runtimeDir); err != nil {
			log.Printf("warn: demo seed failed: %v", err)
		} else {
			fmt.Printf("Demo run seeded — open http://localhost:%d\n", *port)
		}
	}

	effectiveRegistryPath := *registryPath
	if *registryPath == "mcp-registry.yaml" {
		if _, err := os.Stat(effectiveRegistryPath); err != nil {
			effectiveRegistryPath = filepath.Join(globalHome, ".claude", "anton", "mcp-registry.yaml")
		}
	}
	registry, err := mcp.LoadRegistry(effectiveRegistryPath)
	if err != nil {
		log.Printf("warn: could not load MCP registry %s: %v", effectiveRegistryPath, err)
		registry = &mcp.Registry{MCPs: map[string]mcp.MCPEntry{}}
	}

	hub := api.NewHub()

	storeEvents := make(chan store.Event, 64)
	watcher := store.NewWatcher(db, storeEvents)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		watcher.Run(ctx)
		close(storeEvents)
	}()

	go func() {
		for evt := range storeEvents {
			msg, _ := json.Marshal(evt)
			hub.Broadcast(msg)
		}
	}()

	cfg := api.Config{
		Hub:        hub,
		UIFS:       mustSubFS(uiFS, "ui"),
		Store:      db,
		RuntimeDir: runtimeDir,
		WorkflowDirs: []string{
			filepath.Join(projectPath, "workflows"),
			workflowsDir,
			globalWorkflowsDir,
		},

		OnWorkflowUpload: func(data []byte, name string) error {
			base := filepath.Base(name)
			path := filepath.Join(workflowsDir, base)
			abs, err := filepath.Abs(path)
			if err != nil {
				return fmt.Errorf("invalid workflow name")
			}
			absDir, err := filepath.Abs(workflowsDir)
			if err != nil {
				return fmt.Errorf("invalid workflow name")
			}
			if !strings.HasPrefix(abs, absDir+string(filepath.Separator)) {
				return fmt.Errorf("invalid workflow name")
			}
			if _, err := workflow.Parse(data); err != nil {
				return fmt.Errorf("invalid workflow: %w", err)
			}
			return os.WriteFile(path, data, 0644)
		},

		GetActiveWorkflow: func() ([]byte, error) {
			entries, err := os.ReadDir(workflowsDir)
			if err != nil {
				return json.Marshal([]string{})
			}
			var names []string
			for _, e := range entries {
				if !e.IsDir() {
					names = append(names, e.Name())
				}
			}
			return json.Marshal(names)
		},

		SetActiveWorkflow: func(name string) error {
			return nil // no-op: workflows selected per-dispatch now
		},

		GetWorkflowList: func() ([]string, error) {
			seen := map[string]bool{}
			var names []string
			for _, dir := range []string{filepath.Join(projectPath, "workflows"), workflowsDir, globalWorkflowsDir} {
				entries, err := os.ReadDir(dir)
				if err != nil {
					continue
				}
				for _, e := range entries {
					if e.IsDir() {
						continue
					}
					n := strings.TrimSuffix(e.Name(), ".yaml")
					if !seen[n] {
						seen[n] = true
						names = append(names, n)
					}
				}
			}
			sort.Strings(names)
			return names, nil
		},

		GetWorkflowRaw: func(name string) ([]byte, error) {
			if strings.Contains(name, "..") {
				return nil, fmt.Errorf("invalid workflow name")
			}
			base := filepath.Base(name)
			if !strings.HasSuffix(base, ".yaml") {
				base += ".yaml"
			}
			var path string
			for _, dir := range []string{filepath.Join(projectPath, "workflows"), workflowsDir, globalWorkflowsDir} {
				p := filepath.Join(dir, base)
				if _, err := os.Stat(p); err == nil {
					path = p
					break
				}
			}
			if path == "" {
				return nil, fmt.Errorf("workflow not found: %s", name)
			}
			w, err := workflow.ParseFile(path)
			if err != nil {
				return nil, err
			}
			return json.Marshal(w)
		},

		SaveWorkflow: func(name, yamlContent string) error {
			base := filepath.Base(name)
			path := filepath.Join(workflowsDir, base)
			abs, err := filepath.Abs(path)
			if err != nil {
				return fmt.Errorf("invalid workflow name")
			}
			absDir, err := filepath.Abs(workflowsDir)
			if err != nil {
				return fmt.Errorf("invalid workflow name")
			}
			if !strings.HasPrefix(abs, absDir+string(filepath.Separator)) {
				return fmt.Errorf("invalid workflow name")
			}
			if _, err := workflow.Parse([]byte(yamlContent)); err != nil {
				return err
			}
			if err := os.WriteFile(path, []byte(yamlContent), 0644); err != nil {
				return err
			}
			msg, _ := json.Marshal(map[string]string{"type": "workflow_saved", "name": base})
			hub.Broadcast(msg)
			return nil
		},

		GetMCPList: func() []string {
			names := make([]string, 0, len(registry.MCPs))
			for n := range registry.MCPs {
				names = append(names, n)
			}
			sort.Strings(names)
			return names
		},

		WriteMCPConfig: func(mcpNames []string) error {
			claudeDir := filepath.Join(projectPath, ".claude")
			coordinatorJS := filepath.Join(mcpDir(), "team-coordinator.js")
			return mcp.WriteProjectMCPs(claudeDir, coordinatorJS, dbPath, registry, mcpNames)
		},

		GetSettings: func() map[string]string {
			return map[string]string{
				"projectPath":  projectPath,
				"port":         fmt.Sprintf("%d", *port),
				"workflowsDir": workflowsDir,
			}
		},

		SaveSettings: func(settings map[string]string) error {
			log.Printf("settings update: %v", settings)
			return nil
		},
	}

	apiSrv := api.NewServer(cfg)

	httpSrv := &http.Server{
		Addr:              fmt.Sprintf("127.0.0.1:%d", *port),
		Handler:           apiSrv.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
		<-sigCh
		log.Println("shutting down...")
		cancel()
		shutCtx, stop := context.WithTimeout(context.Background(), 10*time.Second)
		defer stop()
		httpSrv.Shutdown(shutCtx) //nolint:errcheck
	}()

	fmt.Printf("Anton running at http://localhost:%d\n", *port)
	if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
