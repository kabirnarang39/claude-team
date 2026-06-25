package workflow_test

import (
	"testing"

	"github.com/kabirnarang39/claude-team/internal/workflow"
)

func TestParse(t *testing.T) {
	data := []byte(`
name: feature-build
agents:
  manager:
    role: manager
    clarify: true
    mcps:
      - atlassian-rovo
      - slack
  engineer:
    role: senior-engineer
    mcps:
      - github
      - filesystem
steps:
  - run: manager
  - parallel:
      - run: engineer
  - run: qa
    when: 'steps["manager"].status == "done"'
`)
	w, err := workflow.Parse(data)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if w.Name != "feature-build" {
		t.Errorf("got name %q, want %q", w.Name, "feature-build")
	}
	if len(w.Agents) != 2 {
		t.Errorf("got %d agents, want 2", len(w.Agents))
	}
	if !w.Agents["manager"].Clarify {
		t.Error("manager.clarify should be true")
	}
	if len(w.Steps) != 3 {
		t.Errorf("got %d steps, want 3", len(w.Steps))
	}
	if len(w.Steps[1].Parallel) != 1 {
		t.Errorf("step[1] should have 1 parallel sub-step")
	}
}

func TestParseFeatureBuildWorkflow(t *testing.T) {
	w, err := workflow.ParseFile("../../workflows/feature-build.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if len(w.Phases) == 0 {
		t.Error("feature-build.yaml should have phases")
	}
	if w.Phases[0].ID != "planning" {
		t.Errorf("first phase should be planning, got %s", w.Phases[0].ID)
	}
	ids := workflow.PhaseIDs(w)
	if len(ids) != len(w.Phases) {
		t.Errorf("PhaseIDs returned %d, want %d", len(ids), len(w.Phases))
	}
}

func TestParseFileNotFound(t *testing.T) {
	_, err := workflow.ParseFile("/no/such/file.yaml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestParseNoName(t *testing.T) {
	_, err := workflow.Parse([]byte(`agents: {}\nsteps: []`))
	if err == nil {
		t.Error("expected error for workflow with no name")
	}
}

func TestAgentsByPhase_Phases(t *testing.T) {
	data := []byte(`
name: feature-build
phases:
  - id: planning
    sequential:
      - requirements-analyst
      - tech-writer
  - id: engineering
    parallel:
      - backend-engineer
      - frontend-engineer
`)
	w, err := workflow.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	pairs := workflow.AgentsByPhase(w)
	if len(pairs) != 2 {
		t.Fatalf("want 2 pairs, got %d", len(pairs))
	}
	if pairs[0].PhaseID != "planning" {
		t.Errorf("want planning, got %s", pairs[0].PhaseID)
	}
	if len(pairs[0].Agents) != 2 {
		t.Errorf("want 2 agents in planning, got %d", len(pairs[0].Agents))
	}
	if pairs[1].PhaseID != "engineering" {
		t.Errorf("want engineering, got %s", pairs[1].PhaseID)
	}
	if len(pairs[1].Agents) != 2 {
		t.Errorf("want 2 agents in engineering, got %d", len(pairs[1].Agents))
	}
}

func TestAgentsByPhase_Steps(t *testing.T) {
	data := []byte(`
name: simple
agents:
  manager:
    role: manager
  engineer:
    role: engineer
steps:
  - run: manager
  - parallel:
      - run: engineer
`)
	w, err := workflow.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	pairs := workflow.AgentsByPhase(w)
	if len(pairs) != 2 {
		t.Fatalf("want 2 pairs, got %d", len(pairs))
	}
	if pairs[0].Agents[0] != "manager" {
		t.Errorf("want manager, got %s", pairs[0].Agents[0])
	}
	if pairs[1].Agents[0] != "engineer" {
		t.Errorf("want engineer, got %s", pairs[1].Agents[0])
	}
}

func TestToGraph(t *testing.T) {
	w := &workflow.Workflow{
		Name: "test",
		Agents: map[string]workflow.Agent{
			"manager":  {Role: "manager"},
			"engineer": {Role: "senior-engineer"},
		},
		Steps: []workflow.Step{
			{Run: "manager"},
			{Parallel: []workflow.Step{{Run: "engineer"}}},
		},
	}
	g := workflow.ToGraph(w)
	if g.Name != "test" {
		t.Errorf("got name %q", g.Name)
	}
	if len(g.Nodes) == 0 {
		t.Error("graph should have nodes")
	}
	if len(g.Edges) == 0 {
		t.Error("graph should have edges")
	}
}
