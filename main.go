package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"claude-team/internal/api"
	"claude-team/internal/mcp"
	"claude-team/internal/store"
	"claude-team/internal/workflow"
)

var version = "1.0.0"

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

func main() {
	port := flag.Int("port", 3000, "HTTP port")
	registryPath := flag.String("registry", "mcp-registry.yaml", "Path to mcp-registry.yaml")
	versionFlag := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("Anton v%s\n", version)
		os.Exit(0)
	}

	checkPrereqs()

	projectPath, _ := filepath.Abs(".")
	runtimeDir := filepath.Join(projectPath, ".claude-team")
	workflowsDir := filepath.Join(runtimeDir, "workflows")

	for _, d := range []string{
		runtimeDir,
		filepath.Join(runtimeDir, "runs"),
		filepath.Join(runtimeDir, "uploads"),
		workflowsDir,
	} {
		os.MkdirAll(d, 0755)
	}

	db, err := store.Open(filepath.Join(runtimeDir, "state.db"))
	if err != nil {
		log.Fatal("open db:", err)
	}
	defer db.Close()

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
	go watcher.Run(ctx)

	go func() {
		for evt := range storeEvents {
			msg, _ := json.Marshal(evt)
			hub.Broadcast(msg)
		}
	}()

	cfg := api.Config{
		Hub:        hub,
		UIDir:      "ui",
		Store:      db,
		RuntimeDir: runtimeDir,
		WorkflowDirs: []string{
			filepath.Join(projectPath, "workflows"),
			workflowsDir,
		},

		OnWorkflowUpload: func(data []byte, name string) error {
			path := filepath.Join(workflowsDir, name)
			if err := os.WriteFile(path, data, 0644); err != nil {
				return err
			}
			return nil
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
			for _, dir := range []string{filepath.Join(projectPath, "workflows"), workflowsDir} {
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
			for _, dir := range []string{filepath.Join(projectPath, "workflows"), workflowsDir} {
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
			if _, err := workflow.Parse([]byte(yamlContent)); err != nil {
				return err
			}
			path := filepath.Join(workflowsDir, name)
			if err := os.WriteFile(path, []byte(yamlContent), 0644); err != nil {
				return err
			}
			msg, _ := json.Marshal(map[string]string{"type": "workflow_saved", "name": name})
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

	srv := api.NewServer(cfg)

	fmt.Printf("Anton running at http://localhost:%d\n", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), srv.Handler()))
}
