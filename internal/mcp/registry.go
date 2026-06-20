package mcp

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Registry struct {
	MCPs map[string]MCPEntry `yaml:"mcps"`
}

type MCPEntry struct {
	Command     string            `yaml:"command"`
	Args        []string          `yaml:"args"`
	Env         map[string]string `yaml:"env,omitempty"`
	Description string            `yaml:"description"`
}

func LoadRegistry(path string) (*Registry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var r Registry
	return &r, yaml.Unmarshal(data, &r)
}
