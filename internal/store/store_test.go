package store_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kabirnarang39/claude-team/internal/store"
)

func TestOpenAndMigrate(t *testing.T) {
	path := t.TempDir() + "/test.db"
	s, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
}

func TestRunCRUD(t *testing.T) {
	s, err := store.Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	id, err := s.CreateRun("feature-build")
	if err != nil {
		t.Fatal(err)
	}
	if id == "" {
		t.Error("run id should not be empty")
	}

	err = s.UpdateAgentStatus(id, "manager", "running")
	if err != nil {
		t.Fatal(err)
	}

	statuses, err := s.GetRunStatuses(id)
	if err != nil {
		t.Fatal(err)
	}
	if statuses["manager"] != "running" {
		t.Errorf("got status %q, want running", statuses["manager"])
	}
}

func openTestStore(t *testing.T) *store.Store {
	t.Helper()
	s, err := store.Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestAgentResultInsertAndFetch(t *testing.T) {
	s := openTestStore(t)
	runID, err := s.CreateRun("feature-build")
	if err != nil {
		t.Fatal(err)
	}

	err = s.UpsertPhase(runID, "planning", "running")
	if err != nil {
		t.Fatal(err)
	}

	result := store.AgentResult{
		RunID:      runID,
		PhaseID:    "planning",
		Agent:      "requirements-analyst",
		Status:     "DONE",
		Confidence: "high",
		Summary:    "Acceptance criteria written.",
		Sources:    []string{"https://example.com"},
		TestsRun:   "n/a",
		TokensUsed: 1200,
	}
	err = s.InsertAgentResult(result)
	if err != nil {
		t.Fatal(err)
	}

	results, maxID, err := s.GetAgentResultsSince(0)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("want 1 result, got %d", len(results))
	}
	if results[0].Agent != "requirements-analyst" {
		t.Errorf("want requirements-analyst, got %s", results[0].Agent)
	}
	if maxID == 0 {
		t.Error("maxID should be > 0")
	}
}

func TestGetRunDetail(t *testing.T) {
	s := openTestStore(t)
	runID, err := s.CreateRun("feature-build")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.UpsertPhase(runID, "planning", "done"); err != nil {
		t.Fatal(err)
	}
	if err := s.InsertAgentResult(store.AgentResult{
		RunID: runID, PhaseID: "planning", Agent: "tech-writer",
		Status: "DONE", Confidence: "high", Summary: "PRD written.",
	}); err != nil {
		t.Fatal(err)
	}

	detail, err := s.GetRunDetail(runID)
	if err != nil {
		t.Fatal(err)
	}
	if len(detail.Phases) != 1 {
		t.Errorf("want 1 phase, got %d", len(detail.Phases))
	}
	if len(detail.Results) != 1 {
		t.Errorf("want 1 result, got %d", len(detail.Results))
	}
}

func TestWriteTask(t *testing.T) {
	s := openTestStore(t)
	dir := t.TempDir()
	err := s.WriteTask(dir, "build user auth")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "pending-task.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "build user auth") {
		t.Error("pending-task.md missing task text")
	}
}

func TestWatcher(t *testing.T) {
	s := openTestStore(t)
	out := make(chan store.Event, 10)
	w := store.NewWatcher(s, out)
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()
	go w.Run(ctx)

	runID, err := s.CreateRun("test")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.InsertAgentResult(store.AgentResult{
		RunID: runID, PhaseID: "planning", Agent: "test-agent",
		Status: "DONE", Confidence: "high", Summary: "done.",
	}); err != nil {
		t.Fatal(err)
	}

	select {
	case evt := <-out:
		if evt.Type != "agent_result" {
			t.Errorf("want agent_result, got %s", evt.Type)
		}
	case <-ctx.Done():
		t.Fatal("timeout — no event received within 6s")
	}
}
