package workflow_test

import (
	"testing"

	"claude-team/internal/workflow"
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
