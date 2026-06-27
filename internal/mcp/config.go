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
		server, ok := buildServer(entry)
		if !ok {
			continue
		}
		settings.MCPServers[name] = server
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

func expandTemplate(v string) (string, bool) {
	if strings.HasPrefix(v, "${") && strings.HasSuffix(v, "}") {
		value := os.Getenv(v[2 : len(v)-1])
		return value, value != ""
	}
	return v, true
}

func buildServer(entry MCPEntry) (mcpServer, bool) {
	result := make([]string, len(entry.Args))
	for i, a := range entry.Args {
		value, ok := expandTemplate(a)
		if !ok {
			return mcpServer{}, false
		}
		result[i] = value
	}

	var env map[string]string
	if len(entry.Env) > 0 {
		env = make(map[string]string, len(entry.Env))
		for k, v := range entry.Env {
			value, ok := expandTemplate(v)
			if !ok {
				return mcpServer{}, false
			}
			env[k] = value
		}
	}
	return mcpServer{Command: entry.Command, Args: result, Env: env}, true
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
		delete(servers, name)
		entry, ok := registry.MCPs[name]
		if !ok {
			continue
		}
		server, ok := buildServer(entry)
		if !ok {
			continue
		}
		servers[name] = server
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
