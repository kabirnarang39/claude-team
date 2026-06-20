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

var version = "1.0.0"

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
	os.MkdirAll(claudeDir, 0755)

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

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return
	}
	if err := os.WriteFile(settingsFile, data, 0644); err != nil {
		log.Printf("warn: could not write .claude/settings.json: %v", err)
	}
}

// runCheck prints a setup health report and exits.
func runCheck(projectPath, dbPath string) {
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
	skillFile := filepath.Join(skillsDir(), "team-dispatch.md")
	_, skillErr := os.Stat(skillFile)
	check("team-dispatch skill", skillFile, skillErr == nil)

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

	fmt.Println()
	if ok {
		fmt.Println("All checks passed. Run `anton` then open http://localhost:3000")
	} else {
		fmt.Println("Fix the items above. Re-run `anton check` to verify.")
		fmt.Printf("To register MCP for this project: anton (just run it — auto-configures on start)\n")
		os.Exit(1)
	}
}

// seedDemoRun inserts a realistic completed run so first-time visitors
// can see what the dashboard looks like without running a real workflow.
func seedDemoRun(db *store.Store) error {
	runID := "demo-feature-build-jwt-auth"
	if err := db.CreateRunWithID(runID, "feature-build"); err != nil {
		return err
	}
	type agent struct {
		phase   string
		name    string
		summary string
		conf    string
		tokens  int
	}
	agents := []agent{
		{"planning", "requirements-analyst", "Defined 12 acceptance criteria: user registration, login, logout, JWT issuance (RS256), refresh token rotation (7d TTL), RBAC (admin/editor/viewer), session invalidation, rate limiting (10 req/min), audit logging, password reset, email verification, MFA stub.", "HIGH", 3200},
		{"planning", "tech-writer", "Drafted product requirements doc with scope boundaries, open questions, and sequence diagrams for all 4 auth flows. Noted: PKCE for future OAuth, bcrypt cost factor 12 (not 10).", "HIGH", 2800},
		{"architecture", "senior-architect", "Designed stateless JWT architecture (RS256 asymmetric) with Redis-backed refresh token store. Chose asymmetric keys for multi-service verification. Defined auth middleware contract.", "HIGH", 4100},
		{"architecture", "api-designer", "Designed RESTful auth API: POST /auth/register, /auth/login, /auth/refresh, DELETE /auth/logout, GET /auth/me. Documented with OpenAPI 3.1. Rate limiting headers specified.", "HIGH", 3500},
		{"engineering", "backend-engineer", "Implemented JWT service (RS256), Redis refresh token store, auth middleware, rate limiter (token bucket), and all 5 endpoints. 47 unit tests, 100% coverage on token logic.", "HIGH", 6200},
		{"engineering", "frontend-engineer", "Built login, register, token-refresh flows. Auth context with React hook. Protected route wrapper. Automatic 401 redirect. Token stored in httpOnly cookie.", "HIGH", 4800},
		{"engineering", "dba", "Designed users, refresh_tokens, roles, user_roles, audit_log tables. Indices on email (unique) and token_hash. Migration files written in plain SQL.", "MEDIUM", 2900},
		{"qa", "qa-engineer", "Wrote 34 integration tests: happy path, expired tokens, tampered signatures, revoked sessions, rate limit enforcement, RBAC boundary checks. All passing.", "HIGH", 3800},
		{"qa", "security-reviewer", "Reviewed against OWASP Top 10. No critical findings. Two medium: add PKCE for OAuth flows (future roadmap), increase bcrypt cost factor 10→12. Both added to backlog.", "HIGH", 4200},
		{"qa", "e2e-tester", "Ran 8 Playwright E2E scenarios: register→login→protected route→logout, password reset, session expiry, concurrent tab sync. All passing.", "HIGH", 3100},
		{"devops", "devops-engineer", "Multi-stage Dockerfile (non-root, minimal image). docker-compose with Redis. GitHub Actions CI/CD pipeline. Helm chart for Kubernetes. Secrets via env vars, never baked in.", "HIGH", 3400},
	}
	pairs := []store.PhaseAgentPair{
		{PhaseID: "planning", Agents: []string{"requirements-analyst", "tech-writer"}},
		{PhaseID: "architecture", Agents: []string{"senior-architect", "api-designer"}},
		{PhaseID: "engineering", Agents: []string{"backend-engineer", "frontend-engineer", "dba"}},
		{PhaseID: "qa", Agents: []string{"qa-engineer", "security-reviewer", "e2e-tester"}},
		{PhaseID: "devops", Agents: []string{"devops-engineer"}},
	}
	db.PrePopulateAgents(runID, pairs)
	for _, a := range agents {
		db.UpsertAgentResult(store.AgentResult{ //nolint:errcheck
			RunID:      runID,
			PhaseID:    a.phase,
			Agent:      a.name,
			Status:     "DONE",
			Confidence: a.conf,
			Summary:    a.summary,
			TokensUsed: a.tokens,
		})
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
		os.MkdirAll(d, 0755)
	}

	// Auto-register MCP in project .claude/settings.json so Claude Code picks it up.
	ensureProjectMCP(projectPath, dbPath)

	db, err := store.Open(dbPath)
	if err != nil {
		log.Fatal("open db:", err)
	}
	defer db.Close()

	if *demoFlag {
		if err := seedDemoRun(db); err != nil {
			log.Printf("warn: demo seed failed: %v", err)
		} else {
			fmt.Println("Demo run seeded — open http://localhost:3000")
		}
	}

	registry, err := mcp.LoadRegistry(*registryPath)
	if err != nil {
		log.Printf("warn: could not load mcp-registry.yaml: %v", err)
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
