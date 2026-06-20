package mcp

import (
	"encoding/json"
	"os"
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
