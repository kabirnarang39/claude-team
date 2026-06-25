package store_test

import (
	"testing"

	"github.com/kabirnarang39/claude-team/internal/store"
)

func TestGetStats_empty(t *testing.T) {
	s := openTestStore(t)
	stats, err := s.GetStats()
	if err != nil {
		t.Fatal(err)
	}
	if stats.RunsTotal != 0 {
		t.Errorf("runs_total: want 0, got %d", stats.RunsTotal)
	}
	if stats.AgentsTotal != 0 {
		t.Errorf("agents_total: want 0, got %d", stats.AgentsTotal)
	}
	if stats.TokensTotal != 0 {
		t.Errorf("tokens_total: want 0, got %d", stats.TokensTotal)
	}
	if stats.ContextIsolationMult != 1.0 {
		t.Errorf("context_isolation_multiplier: want 1.0, got %f", stats.ContextIsolationMult)
	}
	if stats.ContextSavingsPct != 0 {
		t.Errorf("context_savings_pct: want 0 (no data), got %d", stats.ContextSavingsPct)
	}
	// No agent_results → falls back to 1.0
	if stats.ParallelismSpeedup != 1.0 {
		t.Errorf("parallelism_speedup: want 1.0 (no data), got %f", stats.ParallelismSpeedup)
	}
	if stats.AvgAgentsPerRun != 0.0 {
		t.Errorf("avg_agents_per_run: want 0.0, got %f", stats.AvgAgentsPerRun)
	}
	if stats.AvgTokensPerRun != 0.0 {
		t.Errorf("avg_tokens_per_run: want 0.0, got %f", stats.AvgTokensPerRun)
	}
}

func TestGetStats_withData(t *testing.T) {
	s := openTestStore(t)

	runID, err := s.CreateRun("feature-build")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.UpsertAgentResult(store.AgentResult{RunID: runID, PhaseID: "p1", Agent: "architect", Status: "done", TokensUsed: 3000}); err != nil {
		t.Fatal(err)
	}
	if err := s.UpsertAgentResult(store.AgentResult{RunID: runID, PhaseID: "p1", Agent: "engineer", Status: "done", TokensUsed: 5000}); err != nil {
		t.Fatal(err)
	}

	stats, err := s.GetStats()
	if err != nil {
		t.Fatal(err)
	}
	if stats.RunsTotal != 1 {
		t.Errorf("runs_total: want 1, got %d", stats.RunsTotal)
	}
	if stats.AgentsTotal != 2 {
		t.Errorf("agents_total: want 2, got %d", stats.AgentsTotal)
	}
	if stats.TokensTotal != 8000 {
		t.Errorf("tokens_total: want 8000, got %d", stats.TokensTotal)
	}
	if stats.AvgAgentsPerRun != 2.0 {
		t.Errorf("avg_agents_per_run: want 2.0, got %f", stats.AvgAgentsPerRun)
	}
	if stats.AvgTokensPerRun != 8000.0 {
		t.Errorf("avg_tokens_per_run: want 8000.0, got %f", stats.AvgTokensPerRun)
	}
	// (N+1)/2 = (2+1)/2 = 1.5
	if stats.ContextIsolationMult != 1.5 {
		t.Errorf("context_isolation_multiplier: want 1.5, got %f", stats.ContextIsolationMult)
	}
	// 2 agents in 1 phase → speedup = 2/1 = 2.0
	if stats.ParallelismSpeedup != 2.0 {
		t.Errorf("parallelism_speedup: want 2.0, got %f", stats.ParallelismSpeedup)
	}
}
