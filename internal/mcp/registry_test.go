package mcp_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kabirnarang39/claude-team/internal/mcp"
)

func TestLoadRegistry(t *testing.T) {
	r, err := mcp.LoadRegistry("../../mcp-registry.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if len(r.MCPs) == 0 {
		t.Error("registry should have MCPs")
	}
	if _, ok := r.MCPs["github"]; !ok {
		t.Error("registry should have github entry")
	}
	if _, ok := r.MCPs["playwright"]; !ok {
		t.Error("registry should have playwright entry")
	}
}

func TestWriteAgentConfig(t *testing.T) {
	r, err := mcp.LoadRegistry("../../mcp-registry.yaml")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("GITHUB_PERSONAL_ACCESS_TOKEN", "ghp_test")

	dir := t.TempDir()
	err = mcp.WriteAgentConfig(dir, r, []string{"github", "filesystem"})
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "settings.json"))
	if err != nil {
		t.Fatal(err)
	}

	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatal(err)
	}

	servers, ok := settings["mcpServers"].(map[string]any)
	if !ok {
		t.Fatal("settings.json missing mcpServers")
	}
	if _, ok := servers["github"]; !ok {
		t.Error("settings.json should include github")
	}
	if _, ok := servers["coordinator"]; !ok {
		t.Error("settings.json should always include coordinator")
	}
	if _, ok := servers["playwright"]; ok {
		t.Error("settings.json should NOT include playwright (not requested)")
	}
}

func TestWriteAgentConfigSkipsMissingEnv(t *testing.T) {
	r, err := mcp.LoadRegistry("../../mcp-registry.yaml")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("GITHUB_PERSONAL_ACCESS_TOKEN", "")

	dir := t.TempDir()
	err = mcp.WriteAgentConfig(dir, r, []string{"github", "filesystem"})
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatal(err)
	}
	servers := settings["mcpServers"].(map[string]any)
	if _, ok := servers["github"]; ok {
		t.Error("github should be skipped when GITHUB_PERSONAL_ACCESS_TOKEN is unset")
	}
	if _, ok := servers["filesystem"]; !ok {
		t.Error("filesystem should still be included")
	}
}

func TestWriteProjectMCPsDropsStaleServerWhenEnvMissing(t *testing.T) {
	r, err := mcp.LoadRegistry("../../mcp-registry.yaml")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("GITHUB_PERSONAL_ACCESS_TOKEN", "")

	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}
	settingsFile := filepath.Join(claudeDir, "settings.json")
	existing := []byte(`{"mcpServers":{"github":{"command":"npx","args":["old"]}},"other":true}`)
	if err := os.WriteFile(settingsFile, existing, 0644); err != nil {
		t.Fatal(err)
	}

	err = mcp.WriteProjectMCPs(claudeDir, "/tmp/team-coordinator.js", "/tmp/state.db", r, []string{"github", "filesystem"})
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(settingsFile)
	if err != nil {
		t.Fatal(err)
	}
	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatal(err)
	}
	servers := settings["mcpServers"].(map[string]any)
	if _, ok := servers["github"]; ok {
		t.Error("stale github server should be removed when required env is missing")
	}
	if _, ok := servers["filesystem"]; !ok {
		t.Error("filesystem should be written")
	}
	if _, ok := servers["anton-coordinator"]; !ok {
		t.Error("anton-coordinator should always be written")
	}
	if settings["other"] != true {
		t.Error("unrelated settings should be preserved")
	}
}

func TestRegistryUsesNPMVerifiedPackages(t *testing.T) {
	r, err := mcp.LoadRegistry("../../mcp-registry.yaml")
	if err != nil {
		t.Fatal(err)
	}

	verified := map[string]string{
		"@modelcontextprotocol/server-filesystem":   "2026.1.14",
		"@modelcontextprotocol/server-brave-search": "0.6.2",
		"@modelcontextprotocol/server-github":       "2025.4.8",
		"@modelcontextprotocol/server-gitlab":       "2025.4.25",
		"@modelcontextprotocol/server-postgres":     "0.6.2",
		"@playwright/mcp":                           "0.0.76",
		"@modelcontextprotocol/server-slack":        "2025.4.25",
	}

	for name, entry := range r.MCPs {
		if entry.Command != "npx" || len(entry.Args) < 2 || entry.Args[0] != "-y" {
			continue
		}
		pkg := entry.Args[1]
		if _, ok := verified[pkg]; !ok {
			t.Fatalf("%s uses unverified npm package %q; update the registry and this test after validating with npm view", name, pkg)
		}
	}
}
