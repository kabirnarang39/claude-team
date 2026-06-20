package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"claude-team/internal/api"
	"claude-team/internal/store"
)

func TestStatusEndpointNoStore(t *testing.T) {
	hub := api.NewHub()
	srv := api.NewServer(api.Config{
		Hub:   hub,
		UIDir: "../../ui",
		// Store is nil — should return empty object
	})

	req := httptest.NewRequest("GET", "/api/status", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got %d, want 200", rec.Code)
	}

	var result map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty map, got %v", result)
	}
}

func TestWorkflowRawEndpoint(t *testing.T) {
	hub := api.NewHub()
	srv := api.NewServer(api.Config{
		Hub:   hub,
		UIDir: "../../ui",
		GetWorkflowRaw: func(name string) ([]byte, error) {
			return []byte(`{"name":"test","agents":{},"steps":[]}`), nil
		},
	})

	req := httptest.NewRequest("GET", "/api/workflow/raw", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got %d, want 200", rec.Code)
	}
	var result map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result["name"] != "test" {
		t.Errorf("got name %q, want test", result["name"])
	}
}

func TestWorkflowSaveEndpoint(t *testing.T) {
	hub := api.NewHub()
	saved := ""
	srv := api.NewServer(api.Config{
		Hub:   hub,
		UIDir: "../../ui",
		SaveWorkflow: func(name, yaml string) error {
			saved = name + ":" + yaml
			return nil
		},
	})

	body := `{"name":"test.yaml","yaml":"name: test\nagents: {}\nsteps: []"}`
	req := httptest.NewRequest("POST", "/api/workflow/save", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("got %d, want 201", rec.Code)
	}
	if saved == "" {
		t.Error("SaveWorkflow callback was not called")
	}
}

func TestMCPRegistryEndpoint(t *testing.T) {
	hub := api.NewHub()
	srv := api.NewServer(api.Config{
		Hub:   hub,
		UIDir: "../../ui",
		GetMCPList: func() []string {
			return []string{"github", "slack"}
		},
	})

	req := httptest.NewRequest("GET", "/api/mcp-registry", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got %d, want 200", rec.Code)
	}
	var result []string
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if len(result) != 2 {
		t.Errorf("got %d MCPs, want 2", len(result))
	}
}

func TestGetSettingsEndpoint(t *testing.T) {
	hub := api.NewHub()
	srv := api.NewServer(api.Config{
		Hub:   hub,
		UIDir: "../../ui",
		GetSettings: func() map[string]string {
			return map[string]string{
				"projectPath":  "/home/user/myapp",
				"port":         "3000",
				"workflowsDir": ".claude-team/workflows",
			}
		},
	})

	req := httptest.NewRequest("GET", "/api/settings", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got %d, want 200", rec.Code)
	}
	var result map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result["projectPath"] != "/home/user/myapp" {
		t.Errorf("got projectPath %q, want /home/user/myapp", result["projectPath"])
	}
}

func TestSaveSettingsEndpoint(t *testing.T) {
	hub := api.NewHub()
	var saved map[string]string
	srv := api.NewServer(api.Config{
		Hub:   hub,
		UIDir: "../../ui",
		SaveSettings: func(settings map[string]string) error {
			saved = settings
			return nil
		},
	})

	body := `{"projectPath":"/home/user/myapp","port":"3001"}`
	req := httptest.NewRequest("POST", "/api/settings", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got %d, want 200", rec.Code)
	}
	if saved["port"] != "3001" {
		t.Errorf("got port %q, want 3001", saved["port"])
	}
}

func TestIngestResultEndpoint(t *testing.T) {
	// Open a real SQLite store so we can verify the row is persisted.
	s, err := store.Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	// Pre-create the run so the FK-like relationships are consistent.
	if err := s.CreateRunWithID("anton-1781940885-4f3d7a", "feature-build"); err != nil {
		t.Fatal(err)
	}

	hub := api.NewHub()
	srv := api.NewServer(api.Config{
		Hub:   hub,
		UIDir: "../../ui",
		Store: s,
	})

	body := `{
		"agent": "requirements-analyst",
		"run_id": "anton-1781940885-4f3d7a",
		"phase": "planning",
		"status": "DONE",
		"confidence": "high",
		"deliverables": ["docs/requirements.md"],
		"summary": "Acceptance criteria written.",
		"sources": ["https://example.com"],
		"concerns": [],
		"questions": [],
		"tests_run": "N/A",
		"tokens_used": 0
	}`

	req := httptest.NewRequest("POST", "/api/ingest-result", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got %d, want 200 — body: %s", rec.Code, rec.Body.String())
	}

	// Verify the row was written to agent_results.
	results, _, err := s.GetAgentResultsSince(0)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) == 0 {
		t.Fatal("no rows in agent_results after ingest")
	}
	got := results[len(results)-1]
	if got.RunID != "anton-1781940885-4f3d7a" {
		t.Errorf("run_id: got %q, want anton-1781940885-4f3d7a", got.RunID)
	}
	if got.Agent != "requirements-analyst" {
		t.Errorf("agent: got %q, want requirements-analyst", got.Agent)
	}
	if got.Status != "DONE" {
		t.Errorf("status: got %q, want DONE", got.Status)
	}
}

func TestIngestResultBadJSON(t *testing.T) {
	s, err := store.Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	hub := api.NewHub()
	srv := api.NewServer(api.Config{
		Hub:   hub,
		UIDir: "../../ui",
		Store: s,
	})

	req := httptest.NewRequest("POST", "/api/ingest-result", strings.NewReader("{bad json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", rec.Code)
	}
}

func TestRunFilesEndpoints(t *testing.T) {
	dir := t.TempDir()
	runDir := filepath.Join(dir, ".claude-team", "runs", "test-run-123")
	os.MkdirAll(runDir, 0755)
	os.WriteFile(filepath.Join(runDir, "adr.md"), []byte("# ADR content"), 0644)

	db, _ := store.Open(filepath.Join(dir, "state.db"))
	defer db.Close()
	hub := api.NewHub()
	srv := api.NewServer(api.Config{
		Hub:        hub,
		Store:      db,
		RuntimeDir: filepath.Join(dir, ".claude-team"),
		UIDir:      dir,
	})
	handler := srv.Handler()

	t.Run("list files", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/runs/test-run-123/files", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body)
		}
		var files []string
		json.NewDecoder(w.Body).Decode(&files)
		if len(files) != 1 || files[0] != "adr.md" {
			t.Errorf("expected [adr.md], got %v", files)
		}
	})

	t.Run("get file content", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/runs/test-run-123/files/adr.md", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body)
		}
		var result map[string]string
		json.NewDecoder(w.Body).Decode(&result)
		if result["content"] != "# ADR content" {
			t.Errorf("unexpected content: %v", result)
		}
	})

	t.Run("path traversal blocked (encoded slash)", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/runs/test-run-123/files/..%2F..%2Fetc%2Fpasswd", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != 400 {
			t.Fatalf("expected 400 for path traversal, got %d", w.Code)
		}
	})

	t.Run("path traversal blocked (dotdot in name)", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/runs/test-run-123/files/foo..bar", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != 400 {
			t.Fatalf("expected 400 for dotdot in filename, got %d", w.Code)
		}
	})
}

func TestRunsEndpointNoStore(t *testing.T) {
	hub := api.NewHub()
	srv := api.NewServer(api.Config{
		Hub:   hub,
		UIDir: "../../ui",
		// Store is nil — should return empty array
	})

	req := httptest.NewRequest("GET", "/api/runs", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got %d, want 200", rec.Code)
	}
	var result []any
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty array, got %v", result)
	}
}
