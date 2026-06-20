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
