package workflow

// Workflow is the root structure parsed from workflow.yaml.
type Workflow struct {
	Name         string           `yaml:"name" json:"name"`
	Description  string           `yaml:"description,omitempty" json:"description,omitempty"`
	Version      string           `yaml:"version,omitempty" json:"version,omitempty"`
	Type         string           `yaml:"type,omitempty" json:"type,omitempty"`
	Orchestrator string           `yaml:"orchestrator,omitempty" json:"orchestrator,omitempty"`
	MaxCycles    int              `yaml:"max_cycles,omitempty" json:"max_cycles,omitempty"`
	Agents       map[string]Agent `yaml:"agents" json:"agents"`
	Steps        []Step           `yaml:"steps,omitempty" json:"steps,omitempty"`
	Phases       []WorkflowPhase  `yaml:"phases,omitempty" json:"phases,omitempty"`
}

// WorkflowPhase defines a coordinator-driven phase in the new Anton format.
type WorkflowPhase struct {
	ID          string   `yaml:"id" json:"id"`
	Coordinator string   `yaml:"coordinator" json:"coordinator"`
	Sequential  []string `yaml:"sequential,omitempty" json:"sequential,omitempty"`
	Parallel    []string `yaml:"parallel,omitempty" json:"parallel,omitempty"`
	Outputs     []string `yaml:"outputs,omitempty" json:"outputs,omitempty"`
	When        string   `yaml:"when,omitempty" json:"when,omitempty"`
}

// DispatchEvent is sent by the orchestrator via coordinator_dispatch / coordinator_finish MCP tools.
type DispatchEvent struct {
	Agents []string `json:"agents"`
	Done   bool     `json:"done"`
}

// Agent defines a role and its MCP scope for a workflow run.
type Agent struct {
	Role    string   `yaml:"role" json:"role"`
	Clarify bool     `yaml:"clarify,omitempty" json:"clarify,omitempty"`
	MCPs    []string `yaml:"mcps,omitempty" json:"mcps,omitempty"`
}

// Step is either a single agent run or a parallel group.
type Step struct {
	Run      string `yaml:"run,omitempty" json:"run,omitempty"`
	Parallel []Step `yaml:"parallel,omitempty" json:"parallel,omitempty"`
	When     string `yaml:"when,omitempty" json:"when,omitempty"`
}

// AgentCompletion carries the result of an agent terminal session.
type AgentCompletion struct {
	Agent    string
	ExitCode int
}

// Graph is the JSON-serialisable form sent to the UI for SVG rendering.
type Graph struct {
	Name  string      `json:"name"`
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

type GraphNode struct {
	ID      string `json:"id"`
	Label   string `json:"label"`
	Role    string `json:"role"`
	GroupID string `json:"groupId,omitempty"`
}

type GraphEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
}
