package workflow

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func ParseFile(path string) (*Workflow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read workflow file: %w", err)
	}
	return Parse(data)
}

func Parse(data []byte) (*Workflow, error) {
	var w Workflow
	if err := yaml.Unmarshal(data, &w); err != nil {
		return nil, fmt.Errorf("parse workflow yaml: %w", err)
	}
	if w.Name == "" {
		return nil, fmt.Errorf("workflow must have a name")
	}
	return &w, nil
}

// PhaseIDs returns the ordered list of phase IDs from a phases-style workflow.
func PhaseIDs(w *Workflow) []string {
	ids := make([]string, 0, len(w.Phases))
	for _, p := range w.Phases {
		ids = append(ids, p.ID)
	}
	return ids
}

// PhaseAgentPair pairs a phase ID with its agent names for pre-population.
type PhaseAgentPair struct {
	PhaseID string
	Agents  []string
}

// AgentsByPhase extracts ordered (phase_id, agents) pairs.
// Supports both phases: format and steps:/agents: format.
func AgentsByPhase(w *Workflow) []PhaseAgentPair {
	var pairs []PhaseAgentPair
	for _, p := range w.Phases {
		agents := make([]string, 0, len(p.Sequential)+len(p.Parallel))
		agents = append(agents, p.Sequential...)
		agents = append(agents, p.Parallel...)
		if len(agents) > 0 {
			pairs = append(pairs, PhaseAgentPair{PhaseID: p.ID, Agents: agents})
		}
	}
	if len(pairs) > 0 {
		return pairs
	}
	for i, step := range w.Steps {
		phaseID := fmt.Sprintf("step-%d", i)
		var agents []string
		if step.Run != "" {
			agents = append(agents, step.Run)
		}
		for _, sub := range step.Parallel {
			if sub.Run != "" {
				agents = append(agents, sub.Run)
			}
		}
		if len(agents) > 0 {
			pairs = append(pairs, PhaseAgentPair{PhaseID: phaseID, Agents: agents})
		}
	}
	return pairs
}

// ToGraph converts a Workflow into a Graph for SVG rendering.
// Sequential steps become a chain; parallel sub-steps share a group.
func ToGraph(w *Workflow) *Graph {
	g := &Graph{Name: w.Name}
	var prevIDs []string

	for i, step := range w.Steps {
		if len(step.Parallel) > 0 {
			groupID := fmt.Sprintf("group-%d", i)
			var groupIDs []string
			for _, sub := range step.Parallel {
				nodeID := fmt.Sprintf("%s-%d", sub.Run, i)
				role := ""
				if a, ok := w.Agents[sub.Run]; ok {
					role = a.Role
				}
				g.Nodes = append(g.Nodes, GraphNode{
					ID:      nodeID,
					Label:   sub.Run,
					Role:    role,
					GroupID: groupID,
				})
				for _, prev := range prevIDs {
					g.Edges = append(g.Edges, GraphEdge{From: prev, To: nodeID})
				}
				groupIDs = append(groupIDs, nodeID)
			}
			prevIDs = groupIDs
		} else {
			nodeID := fmt.Sprintf("%s-%d", step.Run, i)
			role := ""
			if a, ok := w.Agents[step.Run]; ok {
				role = a.Role
			}
			g.Nodes = append(g.Nodes, GraphNode{
				ID:    nodeID,
				Label: step.Run,
				Role:  role,
			})
			for _, prev := range prevIDs {
				g.Edges = append(g.Edges, GraphEdge{From: prev, To: nodeID})
			}
			prevIDs = []string{nodeID}
		}
	}
	return g
}
