package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type claudeSettings struct {
	MCPServers map[string]mcpServer `json:"mcpServers"`
}

type mcpServer struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env,omitempty"`
}

// WriteAgentConfig writes a per-agent settings.json to configDir.
// The coordinator MCP is always injected regardless of mcpNames.
func WriteAgentConfig(configDir string, registry *Registry, mcpNames []string) error {
	settings := claudeSettings{MCPServers: map[string]mcpServer{}}

	for _, name := range mcpNames {
		entry, ok := registry.MCPs[name]
		if !ok {
			continue
		}
		env := make(map[string]string, len(entry.Env))
		for k, v := range entry.Env {
			env[k] = expandVar(v)
		}
		settings.MCPServers[name] = mcpServer{
			Command: entry.Command,
			Args:    expandArgs(entry.Args),
			Env:     env,
		}
	}

	settings.MCPServers["coordinator"] = mcpServer{
		Command: "node",
		Args:    []string{".claude-team/coordinator-mcp.js"},
	}

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configDir+"/settings.json", data, 0644)
}

func expandVar(v string) string {
	if strings.HasPrefix(v, "${") && strings.HasSuffix(v, "}") {
		return os.Getenv(v[2 : len(v)-1])
	}
	return v
}

func expandArgs(args []string) []string {
	result := make([]string, len(args))
	for i, a := range args {
		result[i] = expandVar(a)
	}
	return result
}

// WriteProjectMCPs merges selected MCP entries (plus the coordinator) into
// the project's .claude/settings.json, preserving all other keys.
// coordinatorJS is the absolute path to team-coordinator.js; dbPath is the
// SQLite state file path passed to the coordinator as ANTON_DB_PATH.
func WriteProjectMCPs(claudeDir, coordinatorJS, dbPath string, registry *Registry, mcpNames []string) error {
	settingsFile := filepath.Join(claudeDir, "settings.json")

	var raw map[string]json.RawMessage
	if data, err := os.ReadFile(settingsFile); err == nil {
		_ = json.Unmarshal(data, &raw)
	}
	if raw == nil {
		raw = make(map[string]json.RawMessage)
	}

	servers := map[string]mcpServer{}
	if v, ok := raw["mcpServers"]; ok {
		_ = json.Unmarshal(v, &servers)
	}

	servers["anton-coordinator"] = mcpServer{
		Command: "node",
		Args:    []string{coordinatorJS},
		Env:     map[string]string{"ANTON_DB_PATH": dbPath},
	}

	for _, name := range mcpNames {
		entry, ok := registry.MCPs[name]
		if !ok {
			continue
		}
		var env map[string]string
		if len(entry.Env) > 0 {
			env = make(map[string]string, len(entry.Env))
			for k, v := range entry.Env {
				env[k] = expandVar(v)
			}
		}
		servers[name] = mcpServer{
			Command: entry.Command,
			Args:    expandArgs(entry.Args),
			Env:     env,
		}
	}

	merged, err := json.Marshal(servers)
	if err != nil {
		return err
	}
	raw["mcpServers"] = json.RawMessage(merged)

	data, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return err
	}
	return os.WriteFile(settingsFile, data, 0644)
}
