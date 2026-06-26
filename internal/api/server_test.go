package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/kabirnarang39/claude-team/internal/api"
	"github.com/kabirnarang39/claude-team/internal/store"
)

func TestStatusEndpointNoStore(t *testing.T) {
	hub := api.NewHub()
	srv := api.NewServer(api.Config{
		Hub:  hub,
		UIFS: fstest.MapFS{},
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
		Hub:  hub,
		UIFS: fstest.MapFS{},
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
		Hub:  hub,
		UIFS: fstest.MapFS{},
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
		Hub:  hub,
		UIFS: fstest.MapFS{},
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
		Hub:  hub,
		UIFS: fstest.MapFS{},
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

func TestHandleRunFilesRich(t *testing.T) {
	dir := t.TempDir()
	runID := "test-run-001"
	runDir := filepath.Join(dir, "runs", runID)
	if err := os.MkdirAll(runDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runDir, "prd.md"), []byte("# PRD\ncontent"), 0644); err != nil {
		t.Fatal(err)
	}

	dbPath := filepath.Join(dir, "state.db")
	st, err := store.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	hub := api.NewHub()
	srv := api.NewServer(api.Config{
		Hub:        hub,
		UIFS:       fstest.MapFS{},
		Store:      st,
		RuntimeDir: dir,
	})

	req := httptest.NewRequest("GET", "/api/runs/test-run-001/files", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", rec.Code)
	}

	type FileEntry struct {
		Name  string `json:"name"`
		Size  int64  `json:"size"`
		Mtime int64  `json:"mtime"`
		Ext   string `json:"ext"`
		Agent string `json:"agent"`
		Phase string `json:"phase"`
	}
	var entries []FileEntry
	if err := json.NewDecoder(rec.Body).Decode(&entries); err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("want 1 entry, got %d", len(entries))
	}
	if entries[0].Name != "prd.md" {
		t.Errorf("want name prd.md, got %q", entries[0].Name)
	}
	if entries[0].Ext != "md" {
		t.Errorf("want ext md, got %q", entries[0].Ext)
	}
	if entries[0].Size != int64(len("# PRD\ncontent")) {
		t.Errorf("want size %d, got %d", len("# PRD\ncontent"), entries[0].Size)
	}
}

func TestSaveSettingsEndpoint(t *testing.T) {
	hub := api.NewHub()
	var saved map[string]string
	srv := api.NewServer(api.Config{
		Hub:  hub,
		UIFS: fstest.MapFS{},
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
		UIFS:  fstest.MapFS{},
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
		UIFS:  fstest.MapFS{},
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
		UIFS:       fstest.MapFS{},
	})
	handler := srv.Handler()

	t.Run("list files", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/runs/test-run-123/files", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body)
		}
		type FileEntry struct {
			Name  string `json:"name"`
			Size  int64  `json:"size"`
			Mtime int64  `json:"mtime"`
			Ext   string `json:"ext"`
			Agent string `json:"agent"`
			Phase string `json:"phase"`
		}
		var files []FileEntry
		json.NewDecoder(w.Body).Decode(&files)
		if len(files) != 1 || files[0].Name != "adr.md" {
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

	t.Run("path traversal blocked (encoded dotdot in run id, files list)", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/runs/%2e%2e%2f%2e%2e%2fetc/files", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != 400 {
			t.Fatalf("expected 400 for dotdot in run id, got %d", w.Code)
		}
	})

	t.Run("path traversal blocked (encoded dotdot in run id, single file)", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/runs/%2e%2e%2f%2e%2e%2fetc/files/passwd", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != 400 {
			t.Fatalf("expected 400 for dotdot in run id, got %d", w.Code)
		}
	})
}

func TestRunDetailEndpoint(t *testing.T) {
	dir := t.TempDir()
	db, err := store.Open(filepath.Join(dir, "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := db.CreateRunWithID("anton-detail-test", "feature-build"); err != nil {
		t.Fatal(err)
	}

	hub := api.NewHub()
	srv := api.NewServer(api.Config{
		Hub:   hub,
		UIFS:  fstest.MapFS{},
		Store: db,
	})
	handler := srv.Handler()

	t.Run("existing run returns 200", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/runs/anton-detail-test", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body)
		}
		var result map[string]any
		if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
			t.Fatal(err)
		}
		if result["id"] != "anton-detail-test" {
			t.Errorf("got id %v, want anton-detail-test", result["id"])
		}
	})

	t.Run("path traversal blocked in run id", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/runs/%2e%2e%2f%2e%2e%2fetc", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != 400 {
			t.Fatalf("expected 400 for dotdot in run id, got %d", w.Code)
		}
	})
}

func TestTaskEndpoint(t *testing.T) {
	dir := t.TempDir()
	db, err := store.Open(filepath.Join(dir, "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	hub := api.NewHub()
	srv := api.NewServer(api.Config{
		Hub:        hub,
		UIFS:       fstest.MapFS{},
		Store:      db,
		RuntimeDir: dir,
	})
	handler := srv.Handler()

	t.Run("valid task returns 202 with run_id", func(t *testing.T) {
		body := `{"text":"build user auth","workflow":"feature-build"}`
		req := httptest.NewRequest("POST", "/api/task", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusAccepted {
			t.Fatalf("expected 202, got %d: %s", w.Code, w.Body)
		}
		var result map[string]string
		if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
			t.Fatal(err)
		}
		if result["run_id"] == "" {
			t.Error("expected non-empty run_id in response")
		}
	})

	t.Run("empty body returns 400", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/task", strings.NewReader("{bad json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != 400 {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})
}

func TestRunsEndpointNoStore(t *testing.T) {
	hub := api.NewHub()
	srv := api.NewServer(api.Config{
		Hub:  hub,
		UIFS: fstest.MapFS{},
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

func TestHandleStats_empty(t *testing.T) {
	dir := t.TempDir()
	s, err := store.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	srv := api.NewServer(api.Config{Store: s})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/stats", nil)
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: want 200, got %d", rec.Code)
	}
	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["runs_total"] != float64(0) {
		t.Errorf("runs_total: want 0, got %v", body["runs_total"])
	}
	if body["context_savings_pct"] != float64(0) {
		t.Errorf("context_savings_pct: want 0, got %v", body["context_savings_pct"])
	}
}

func TestHandleStats_nilStore(t *testing.T) {
	srv := api.NewServer(api.Config{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/stats", nil)
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: want 200, got %d", rec.Code)
	}
	if strings.TrimSpace(rec.Body.String()) != "{}" {
		t.Errorf("want {}, got %q", rec.Body.String())
	}
}

func openTestStoreForAPI(t *testing.T) *store.Store {
	t.Helper()
	s, err := store.Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestSignalReview(t *testing.T) {
	db := openTestStoreForAPI(t)
	runID := "run-signal-test"
	if err := db.CreateRunWithID(runID, "feature-build"); err != nil {
		t.Fatal(err)
	}

	hub := api.NewHub()
	srv := api.NewServer(api.Config{
		Hub:   hub,
		UIFS:  fstest.MapFS{},
		Store: db,
	})

	body := `{"gate":"plan-review","summary":"PRD covers auth and payments"}`
	req := httptest.NewRequest("POST", "/api/runs/"+runID+"/signal-review", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("got %d, want 204", rec.Code)
	}

	detail, err := db.GetRunDetail(runID)
	if err != nil {
		t.Fatal(err)
	}
	if len(detail.Reviews) != 1 {
		t.Fatalf("expected 1 review, got %d", len(detail.Reviews))
	}
	if detail.Reviews[0].Status != "pending" {
		t.Errorf("status: got %q, want pending", detail.Reviews[0].Status)
	}
	if detail.Reviews[0].Gate != "plan-review" {
		t.Errorf("gate: got %q, want plan-review", detail.Reviews[0].Gate)
	}
}

func TestSignalReviewInvalidGate(t *testing.T) {
	db := openTestStoreForAPI(t)
	runID := "run-invalid-gate"
	db.CreateRunWithID(runID, "feature-build")

	hub := api.NewHub()
	srv := api.NewServer(api.Config{Hub: hub, UIFS: fstest.MapFS{}, Store: db})

	// Any non-empty gate name is now valid — gates are open-ended (agent-question, qa-fail, etc.)
	// Only empty string or >128 chars should return 400.
	cases := []struct {
		gate string
		want int
	}{
		{"", 400},
		{strings.Repeat("x", 129), 400},
		{"bad-gate", 204},        // previously invalid, now accepted
		{"agent-question", 204},  // new gate type
		{"plan-review", 204},     // existing gate type still works
	}
	for _, c := range cases {
		body := `{"gate":"` + c.gate + `","summary":"test"}`
		req := httptest.NewRequest("POST", "/api/runs/"+runID+"/signal-review", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, req)
		if rec.Code != c.want {
			t.Errorf("gate %q: got %d, want %d", c.gate, rec.Code, c.want)
		}
	}
}

func TestResolveReview(t *testing.T) {
	db := openTestStoreForAPI(t)
	runID := "run-resolve-test"
	if err := db.CreateRunWithID(runID, "feature-build"); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateReview(runID, "task-review", "ADR summary"); err != nil {
		t.Fatal(err)
	}

	hub := api.NewHub()
	srv := api.NewServer(api.Config{Hub: hub, UIFS: fstest.MapFS{}, Store: db})

	body := `{"gate":"task-review","status":"rejected","feedback":"Add caching layer detail"}`
	req := httptest.NewRequest("POST", "/api/runs/"+runID+"/resolve-review", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("got %d, want 204", rec.Code)
	}

	detail, err := db.GetRunDetail(runID)
	if err != nil {
		t.Fatal(err)
	}
	if detail.Reviews[0].Status != "rejected" {
		t.Errorf("status: got %q, want rejected", detail.Reviews[0].Status)
	}
	if detail.Reviews[0].Feedback != "Add caching layer detail" {
		t.Errorf("feedback: got %q", detail.Reviews[0].Feedback)
	}
}

func TestResolveReviewInvalidStatus(t *testing.T) {
	db := openTestStoreForAPI(t)
	runID := "run-invalid-status"
	db.CreateRunWithID(runID, "feature-build")
	db.CreateReview(runID, "plan-review", "summary")

	hub := api.NewHub()
	srv := api.NewServer(api.Config{Hub: hub, UIFS: fstest.MapFS{}, Store: db})

	body := `{"gate":"plan-review","status":"maybe","feedback":""}`
	req := httptest.NewRequest("POST", "/api/runs/"+runID+"/resolve-review", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", rec.Code)
	}
}

func TestWorkflowUploadEndpoint(t *testing.T) {
	var uploaded []byte
	hub := api.NewHub()
	srv := api.NewServer(api.Config{
		Hub:  hub,
		UIFS: fstest.MapFS{},
		OnWorkflowUpload: func(data []byte, name string) error {
			uploaded = data
			return nil
		},
	})

	body := `name: my-workflow
agents: {}
steps: []`
	req := httptest.NewRequest("POST", "/api/workflow/upload?name=my-workflow.yaml", strings.NewReader(body))
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("got %d, want 201", rec.Code)
	}
	if string(uploaded) != body {
		t.Errorf("uploaded content mismatch")
	}
}

func TestWorkflowUploadNotConfigured(t *testing.T) {
	srv := api.NewServer(api.Config{UIFS: fstest.MapFS{}})
	req := httptest.NewRequest("POST", "/api/workflow/upload", strings.NewReader("data"))
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("got %d, want 500", rec.Code)
	}
}

func TestWorkflowActiveEndpoint(t *testing.T) {
	hub := api.NewHub()
	srv := api.NewServer(api.Config{
		Hub:  hub,
		UIFS: fstest.MapFS{},
		GetActiveWorkflow: func() ([]byte, error) {
			return []byte(`{"name":"feature-build"}`), nil
		},
	})

	req := httptest.NewRequest("GET", "/api/workflow/active", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "feature-build") {
		t.Errorf("body missing workflow name: %s", rec.Body.String())
	}
}

func TestWorkflowSetActiveEndpoint(t *testing.T) {
	var activated string
	hub := api.NewHub()
	srv := api.NewServer(api.Config{
		Hub:  hub,
		UIFS: fstest.MapFS{},
		SetActiveWorkflow: func(name string) error {
			activated = name
			return nil
		},
	})

	body := `{"name":"security-audit"}`
	req := httptest.NewRequest("PUT", "/api/workflow/active", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got %d, want 200", rec.Code)
	}
	if activated != "security-audit" {
		t.Errorf("want security-audit activated, got %q", activated)
	}
}

func TestWorkflowListEndpoint(t *testing.T) {
	hub := api.NewHub()
	srv := api.NewServer(api.Config{
		Hub:  hub,
		UIFS: fstest.MapFS{},
		GetWorkflowList: func() ([]string, error) {
			return []string{"feature-build", "security-audit"}, nil
		},
	})

	req := httptest.NewRequest("GET", "/api/workflows", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got %d, want 200", rec.Code)
	}
	var names []string
	if err := json.NewDecoder(rec.Body).Decode(&names); err != nil {
		t.Fatal(err)
	}
	if len(names) != 2 {
		t.Errorf("want 2 workflows, got %d", len(names))
	}
}

func TestWorkflowListNilCallback(t *testing.T) {
	srv := api.NewServer(api.Config{UIFS: fstest.MapFS{}})
	req := httptest.NewRequest("GET", "/api/workflows", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("got %d, want 200", rec.Code)
	}
	var names []string
	if err := json.NewDecoder(rec.Body).Decode(&names); err != nil {
		t.Fatal(err)
	}
	if len(names) != 0 {
		t.Errorf("want empty list, got %v", names)
	}
}

func TestFileUploadEndpoint(t *testing.T) {
	dir := t.TempDir()
	hub := api.NewHub()
	srv := api.NewServer(api.Config{
		Hub:        hub,
		UIFS:       fstest.MapFS{},
		RuntimeDir: dir,
	})

	req := httptest.NewRequest("POST", "/api/files/upload?name=spec.md", strings.NewReader("# spec content"))
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got %d, want 200: %s", rec.Code, rec.Body.String())
	}
	var result map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result["path"] == "" {
		t.Error("expected path in response")
	}
	// Verify file on disk
	data, err := os.ReadFile(filepath.Join(dir, "uploads", "spec.md"))
	if err != nil {
		t.Fatalf("uploaded file not found: %v", err)
	}
	if string(data) != "# spec content" {
		t.Errorf("file content mismatch: %q", string(data))
	}
}

func TestFileUploadMissingName(t *testing.T) {
	dir := t.TempDir()
	srv := api.NewServer(api.Config{UIFS: fstest.MapFS{}, RuntimeDir: dir})
	req := httptest.NewRequest("POST", "/api/files/upload", strings.NewReader("data"))
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", rec.Code)
	}
}

func TestRunFileRawEndpoint(t *testing.T) {
	dir := t.TempDir()
	runDir := filepath.Join(dir, "runs", "run-raw-test")
	if err := os.MkdirAll(runDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runDir, "plan.md"), []byte("# Plan"), 0644); err != nil {
		t.Fatal(err)
	}

	srv := api.NewServer(api.Config{UIFS: fstest.MapFS{}, RuntimeDir: dir})
	handler := srv.Handler()

	t.Run("returns file content with correct content type", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/runs/run-raw-test/files/plan.md/raw", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("got %d, want 200", rec.Code)
		}
		if rec.Body.String() != "# Plan" {
			t.Errorf("got %q, want '# Plan'", rec.Body.String())
		}
		if !strings.Contains(rec.Header().Get("Content-Type"), "text/plain") {
			t.Errorf("want text/plain content type, got %q", rec.Header().Get("Content-Type"))
		}
	})

	t.Run("missing file returns 404", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/runs/run-raw-test/files/missing.md/raw", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("got %d, want 404", rec.Code)
		}
	})

	t.Run("path traversal blocked", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/runs/run-raw-test/files/..%2Fetc%2Fpasswd/raw", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("got %d, want 400", rec.Code)
		}
	})
}

func TestNoCacheHeader(t *testing.T) {
	srv := api.NewServer(api.Config{UIFS: fstest.MapFS{"index.html": &fstest.MapFile{Data: []byte("hi")}}})
	req := httptest.NewRequest("GET", "/index.html", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Errorf("want Cache-Control: no-store, got %q", rec.Header().Get("Cache-Control"))
	}
}

func TestHandleSettingsBadJSON(t *testing.T) {
	srv := api.NewServer(api.Config{UIFS: fstest.MapFS{}})
	req := httptest.NewRequest("POST", "/api/settings", strings.NewReader("{bad json"))
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", rec.Code)
	}
}

func TestHandleSettingsNilCallback(t *testing.T) {
	// nil SaveSettings should still return 200
	srv := api.NewServer(api.Config{UIFS: fstest.MapFS{}})
	req := httptest.NewRequest("POST", "/api/settings", strings.NewReader(`{"port":"3000"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("got %d, want 200", rec.Code)
	}
}

func TestHandleRunFilesNotConfigured(t *testing.T) {
	srv := api.NewServer(api.Config{UIFS: fstest.MapFS{}})
	req := httptest.NewRequest("GET", "/api/runs/some-run/files", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("got %d, want 500", rec.Code)
	}
}

func TestHandleRunFilesEmptyDir(t *testing.T) {
	// Non-existent run dir returns empty array, not error
	srv := api.NewServer(api.Config{UIFS: fstest.MapFS{}, RuntimeDir: t.TempDir()})
	req := httptest.NewRequest("GET", "/api/runs/no-such-run/files", nil)
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
		t.Errorf("want empty array, got %v", result)
	}
}

func TestHandleRunFileNotConfigured(t *testing.T) {
	srv := api.NewServer(api.Config{UIFS: fstest.MapFS{}})
	req := httptest.NewRequest("GET", "/api/runs/some-run/files/plan.md", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("got %d, want 500", rec.Code)
	}
}

func TestHandleRunFileMissing(t *testing.T) {
	srv := api.NewServer(api.Config{UIFS: fstest.MapFS{}, RuntimeDir: t.TempDir()})
	req := httptest.NewRequest("GET", "/api/runs/no-such-run/files/plan.md", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("got %d, want 404", rec.Code)
	}
}

func TestHandleRunsWithData(t *testing.T) {
	db := openTestStoreForAPI(t)
	if err := db.CreateRunWithID("run-a", "feature-build"); err != nil {
		t.Fatal(err)
	}
	srv := api.NewServer(api.Config{UIFS: fstest.MapFS{}, Store: db})
	req := httptest.NewRequest("GET", "/api/runs", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", rec.Code)
	}
	var runs []map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&runs); err != nil {
		t.Fatal(err)
	}
	if len(runs) != 1 {
		t.Errorf("want 1 run, got %d", len(runs))
	}
}

func TestHandleStatusWithStore(t *testing.T) {
	db := openTestStoreForAPI(t)
	id, err := db.CreateRun("feature-build")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateAgentStatus(id, "engineer", "running"); err != nil {
		t.Fatal(err)
	}
	srv := api.NewServer(api.Config{UIFS: fstest.MapFS{}, Store: db})
	req := httptest.NewRequest("GET", "/api/status", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", rec.Code)
	}
	var result map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result["engineer"] != "running" {
		t.Errorf("got %q, want running", result["engineer"])
	}
}

func TestWorkflowRawNotConfigured(t *testing.T) {
	srv := api.NewServer(api.Config{UIFS: fstest.MapFS{}})
	req := httptest.NewRequest("GET", "/api/workflow/raw", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("got %d, want 500", rec.Code)
	}
}

func TestWorkflowRawNotFound(t *testing.T) {
	srv := api.NewServer(api.Config{
		UIFS: fstest.MapFS{},
		GetWorkflowRaw: func(name string) ([]byte, error) {
			return nil, fmt.Errorf("workflow not found: %s", name)
		},
	})
	req := httptest.NewRequest("GET", "/api/workflow/raw?name=missing", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("got %d, want 404", rec.Code)
	}
}

func TestWorkflowSaveBadJSON(t *testing.T) {
	srv := api.NewServer(api.Config{UIFS: fstest.MapFS{}, SaveWorkflow: func(n, y string) error { return nil }})
	req := httptest.NewRequest("POST", "/api/workflow/save", strings.NewReader("{bad"))
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", rec.Code)
	}
}

func TestWorkflowSaveEmptyName(t *testing.T) {
	srv := api.NewServer(api.Config{UIFS: fstest.MapFS{}, SaveWorkflow: func(n, y string) error { return nil }})
	req := httptest.NewRequest("POST", "/api/workflow/save", strings.NewReader(`{"name":"","yaml":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", rec.Code)
	}
}

func TestWorkflowSaveNilCallback(t *testing.T) {
	srv := api.NewServer(api.Config{UIFS: fstest.MapFS{}})
	req := httptest.NewRequest("POST", "/api/workflow/save", strings.NewReader(`{"name":"x.yaml","yaml":"y"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("got %d, want 500", rec.Code)
	}
}

func TestWorkflowSetActiveNilCallback(t *testing.T) {
	srv := api.NewServer(api.Config{UIFS: fstest.MapFS{}})
	req := httptest.NewRequest("PUT", "/api/workflow/active", strings.NewReader(`{"name":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("got %d, want 500", rec.Code)
	}
}

func TestWorkflowSetActiveBadJSON(t *testing.T) {
	srv := api.NewServer(api.Config{UIFS: fstest.MapFS{}, SetActiveWorkflow: func(n string) error { return nil }})
	req := httptest.NewRequest("PUT", "/api/workflow/active", strings.NewReader("{bad"))
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", rec.Code)
	}
}

func TestWorkflowActiveNilCallback(t *testing.T) {
	srv := api.NewServer(api.Config{UIFS: fstest.MapFS{}})
	req := httptest.NewRequest("GET", "/api/workflow/active", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("got %d, want 500", rec.Code)
	}
}

func TestMCPRegistryNilCallback(t *testing.T) {
	srv := api.NewServer(api.Config{UIFS: fstest.MapFS{}})
	req := httptest.NewRequest("GET", "/api/mcp-registry", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("got %d, want 200", rec.Code)
	}
	var names []string
	if err := json.NewDecoder(rec.Body).Decode(&names); err != nil {
		t.Fatal(err)
	}
	if len(names) != 0 {
		t.Errorf("want empty, got %v", names)
	}
}

func TestGetSettingsNilCallback(t *testing.T) {
	srv := api.NewServer(api.Config{UIFS: fstest.MapFS{}})
	req := httptest.NewRequest("GET", "/api/settings", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("got %d, want 200", rec.Code)
	}
}

func TestSignalReviewBadJSON(t *testing.T) {
	db := openTestStoreForAPI(t)
	srv := api.NewServer(api.Config{Hub: api.NewHub(), UIFS: fstest.MapFS{}, Store: db})
	req := httptest.NewRequest("POST", "/api/runs/run-x/signal-review", strings.NewReader("{bad"))
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", rec.Code)
	}
}

func TestSignalReviewNilStore(t *testing.T) {
	srv := api.NewServer(api.Config{Hub: api.NewHub(), UIFS: fstest.MapFS{}})
	body := `{"gate":"plan-review","summary":"x"}`
	req := httptest.NewRequest("POST", "/api/runs/run-x/signal-review", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("got %d, want 500", rec.Code)
	}
}

func TestResolveReviewBadJSON(t *testing.T) {
	db := openTestStoreForAPI(t)
	srv := api.NewServer(api.Config{Hub: api.NewHub(), UIFS: fstest.MapFS{}, Store: db})
	req := httptest.NewRequest("POST", "/api/runs/run-x/resolve-review", strings.NewReader("{bad"))
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", rec.Code)
	}
}

func TestRunDetailNilStore(t *testing.T) {
	srv := api.NewServer(api.Config{UIFS: fstest.MapFS{}})
	req := httptest.NewRequest("GET", "/api/runs/some-run", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("got %d, want 500", rec.Code)
	}
}

func TestRunDetailNotFound(t *testing.T) {
	db := openTestStoreForAPI(t)
	srv := api.NewServer(api.Config{UIFS: fstest.MapFS{}, Store: db})
	req := httptest.NewRequest("GET", "/api/runs/no-such-run", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("got %d, want 404", rec.Code)
	}
}

func TestIngestResultNilStore(t *testing.T) {
	srv := api.NewServer(api.Config{UIFS: fstest.MapFS{}})
	req := httptest.NewRequest("POST", "/api/ingest-result", strings.NewReader(`{"agent":"x","run_id":"r","status":"DONE"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("got %d, want 500", rec.Code)
	}
}

func TestHandleTaskNilStore(t *testing.T) {
	srv := api.NewServer(api.Config{UIFS: fstest.MapFS{}})
	req := httptest.NewRequest("POST", "/api/task", strings.NewReader(`{"text":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("got %d, want 500", rec.Code)
	}
}

func TestHandleTaskWithJiraURL(t *testing.T) {
	dir := t.TempDir()
	db := openTestStoreForAPI(t)
	srv := api.NewServer(api.Config{
		UIFS:       fstest.MapFS{},
		Store:      db,
		RuntimeDir: dir,
	})
	body := `{"text":"build auth","jiraUrl":"https://jira.example.com/PROJ-123","workflow":"feature-build"}`
	req := httptest.NewRequest("POST", "/api/task", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("got %d, want 202: %s", rec.Code, rec.Body.String())
	}
}

func TestResolveReviewNilStore(t *testing.T) {
	srv := api.NewServer(api.Config{Hub: api.NewHub(), UIFS: fstest.MapFS{}})
	body := `{"gate":"plan-review","status":"approved","feedback":""}`
	req := httptest.NewRequest("POST", "/api/runs/run-x/resolve-review", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("got %d, want 500", rec.Code)
	}
}

func TestWorkflowUploadCallbackError(t *testing.T) {
	srv := api.NewServer(api.Config{
		UIFS: fstest.MapFS{},
		OnWorkflowUpload: func(data []byte, name string) error {
			return fmt.Errorf("disk full")
		},
	})
	req := httptest.NewRequest("POST", "/api/workflow/upload?name=x.yaml", strings.NewReader("data"))
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("got %d, want 500", rec.Code)
	}
}

func TestWorkflowActiveCallbackError(t *testing.T) {
	srv := api.NewServer(api.Config{
		UIFS: fstest.MapFS{},
		GetActiveWorkflow: func() ([]byte, error) {
			return nil, fmt.Errorf("no active workflow")
		},
	})
	req := httptest.NewRequest("GET", "/api/workflow/active", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("got %d, want 500", rec.Code)
	}
}

func TestWorkflowListCallbackError(t *testing.T) {
	srv := api.NewServer(api.Config{
		UIFS: fstest.MapFS{},
		GetWorkflowList: func() ([]string, error) {
			return nil, fmt.Errorf("read error")
		},
	})
	req := httptest.NewRequest("GET", "/api/workflows", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("got %d, want 500", rec.Code)
	}
}

func TestHandleSettingsSaveError(t *testing.T) {
	srv := api.NewServer(api.Config{
		UIFS: fstest.MapFS{},
		SaveSettings: func(s map[string]string) error {
			return fmt.Errorf("write error")
		},
	})
	req := httptest.NewRequest("POST", "/api/settings", strings.NewReader(`{"port":"9999"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("got %d, want 500", rec.Code)
	}
}

func TestFileUploadNotConfigured(t *testing.T) {
	srv := api.NewServer(api.Config{UIFS: fstest.MapFS{}})
	req := httptest.NewRequest("POST", "/api/files/upload?name=x.md", strings.NewReader("data"))
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("got %d, want 500", rec.Code)
	}
}

func TestAgentsByPhaseStepsPath(t *testing.T) {
	// Exercise agentsByPhase steps branch via handleTask with a steps-based workflow.
	dir := t.TempDir()
	wfDir := filepath.Join(dir, "workflows")
	if err := os.MkdirAll(wfDir, 0755); err != nil {
		t.Fatal(err)
	}
	// steps-based workflow (no phases:)
	wfYAML := `name: steps-workflow
agents:
  manager:
    role: manager
  engineer:
    role: engineer
steps:
  - run: manager
  - parallel:
      - run: engineer
`
	if err := os.WriteFile(filepath.Join(wfDir, "steps-workflow.yaml"), []byte(wfYAML), 0644); err != nil {
		t.Fatal(err)
	}

	db := openTestStoreForAPI(t)
	srv := api.NewServer(api.Config{
		UIFS:         fstest.MapFS{},
		Store:        db,
		RuntimeDir:   dir,
		WorkflowDirs: []string{wfDir},
	})

	body := `{"text":"build something","workflow":"steps-workflow"}`
	req := httptest.NewRequest("POST", "/api/task", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("got %d, want 202: %s", rec.Code, rec.Body.String())
	}
	var result map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	detail, err := db.GetRunDetail(result["run_id"])
	if err != nil {
		t.Fatal(err)
	}
	// steps-based workflow should pre-populate 2 agents (manager + engineer)
	if len(detail.Results) != 2 {
		t.Errorf("want 2 pre-populated agents, got %d", len(detail.Results))
	}
}

func TestTaskEndpointWithWorkflowDir(t *testing.T) {
	// Tests agentsByPhase code path via handleTask when WorkflowDirs is set.
	dir := t.TempDir()
	wfDir := filepath.Join(dir, "workflows")
	if err := os.MkdirAll(wfDir, 0755); err != nil {
		t.Fatal(err)
	}
	wfYAML := `name: feature-build
phases:
  - id: planning
    sequential:
      - requirements-analyst
      - tech-writer
  - id: engineering
    parallel:
      - backend-engineer
      - frontend-engineer
`
	if err := os.WriteFile(filepath.Join(wfDir, "feature-build.yaml"), []byte(wfYAML), 0644); err != nil {
		t.Fatal(err)
	}

	db, err := store.Open(filepath.Join(dir, "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	srv := api.NewServer(api.Config{
		UIFS:         fstest.MapFS{},
		Store:        db,
		RuntimeDir:   dir,
		WorkflowDirs: []string{wfDir},
	})

	body := `{"text":"build user auth","workflow":"feature-build"}`
	req := httptest.NewRequest("POST", "/api/task", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("got %d, want 202: %s", rec.Code, rec.Body.String())
	}
	var result map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result["run_id"] == "" {
		t.Error("expected non-empty run_id")
	}
	// Verify PrePopulateAgents was called — run detail should have placeholder agents
	detail, err := db.GetRunDetail(result["run_id"])
	if err != nil {
		t.Fatal(err)
	}
	if len(detail.Results) == 0 {
		t.Error("expected placeholder agent results from PrePopulateAgents")
	}
}
