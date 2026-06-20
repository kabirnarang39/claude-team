package api

import (
	crand "crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"claude-team/internal/store"
	wflow "claude-team/internal/workflow"
)

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if s.cfg.Store == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{}"))
		return
	}
	statuses, err := s.cfg.Store.GetAllStatuses()
	if err != nil {
		statuses = map[string]string{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(statuses)
}

func (s *Server) handleTask(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Text     string `json:"text"`
		JiraURL  string `json:"jiraUrl"`
		Workflow string `json:"workflow"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	wfName := body.Workflow
	if wfName == "" {
		wfName = "feature-build"
	}

	b := make([]byte, 3)
	crand.Read(b)
	runID := fmt.Sprintf("anton-%d-%x", time.Now().Unix(), b)

	task := body.Text
	if body.JiraURL != "" {
		task += "\n\nJira: " + body.JiraURL
	}
	taskContent := fmt.Sprintf("Run ID: %s\nWorkflow: %s\n\n%s", runID, wfName, task)

	if s.cfg.Store == nil || s.cfg.RuntimeDir == "" {
		http.Error(w, "not configured", 500)
		return
	}

	if err := s.cfg.Store.CreateRunWithID(runID, wfName); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if wf := s.findWorkflow(wfName); wf != nil {
		pairs := agentsByPhase(wf)
		s.cfg.Store.PrePopulateAgents(runID, pairs)
	}

	if err := s.cfg.Store.WriteTask(s.cfg.RuntimeDir, taskContent); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"run_id": runID})
}

func (s *Server) findWorkflow(name string) *wflow.Workflow {
	base := name
	if !strings.HasSuffix(base, ".yaml") {
		base += ".yaml"
	}
	for _, dir := range s.cfg.WorkflowDirs {
		p := filepath.Join(dir, base)
		if wf, err := wflow.ParseFile(p); err == nil {
			return wf
		}
	}
	return nil
}

func agentsByPhase(w *wflow.Workflow) []store.PhaseAgentPair {
	var pairs []store.PhaseAgentPair
	for _, p := range w.Phases {
		agents := append(append([]string{}, p.Sequential...), p.Parallel...)
		if len(agents) > 0 {
			pairs = append(pairs, store.PhaseAgentPair{PhaseID: p.ID, Agents: agents})
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
			pairs = append(pairs, store.PhaseAgentPair{PhaseID: phaseID, Agents: agents})
		}
	}
	return pairs
}

func (s *Server) handleRuns(w http.ResponseWriter, r *http.Request) {
	if s.cfg.Store == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
		return
	}
	runs, err := s.cfg.Store.GetRuns(20)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if runs == nil {
		runs = []store.RunDetail{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(runs)
}

func (s *Server) handleRunDetail(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if s.cfg.Store == nil {
		http.Error(w, "not configured", 500)
		return
	}
	detail, err := s.cfg.Store.GetRunDetail(id)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(detail)
}

func (s *Server) handleWorkflowUpload(w http.ResponseWriter, r *http.Request) {
	if s.cfg.OnWorkflowUpload == nil {
		http.Error(w, "not configured", 500)
		return
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	name := r.URL.Query().Get("name")
	if err := s.cfg.OnWorkflowUpload(data, name); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) handleWorkflowActive(w http.ResponseWriter, r *http.Request) {
	if s.cfg.GetActiveWorkflow == nil {
		http.Error(w, "not configured", 500)
		return
	}
	data, err := s.cfg.GetActiveWorkflow()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (s *Server) handleWorkflowSetActive(w http.ResponseWriter, r *http.Request) {
	if s.cfg.SetActiveWorkflow == nil {
		http.Error(w, "not configured", 500)
		return
	}
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if err := s.cfg.SetActiveWorkflow(body.Name); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleWorkflowList(w http.ResponseWriter, r *http.Request) {
	if s.cfg.GetWorkflowList == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]string{})
		return
	}
	names, err := s.cfg.GetWorkflowList()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(names)
}

func (s *Server) handleWorkflowRaw(w http.ResponseWriter, r *http.Request) {
	if s.cfg.GetWorkflowRaw == nil {
		http.Error(w, "not configured", 500)
		return
	}
	name := r.URL.Query().Get("name")
	data, err := s.cfg.GetWorkflowRaw(name)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (s *Server) handleWorkflowSave(w http.ResponseWriter, r *http.Request) {
	if s.cfg.SaveWorkflow == nil {
		http.Error(w, "not configured", 500)
		return
	}
	var body struct {
		Name string `json:"name"`
		YAML string `json:"yaml"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if body.Name == "" {
		http.Error(w, "name required", 400)
		return
	}
	if err := s.cfg.SaveWorkflow(body.Name, body.YAML); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"name": body.Name})
}

func (s *Server) handleMCPRegistry(w http.ResponseWriter, r *http.Request) {
	if s.cfg.GetMCPList == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]string{})
		return
	}
	names := s.cfg.GetMCPList()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(names)
}

func (s *Server) handleFileUpload(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "name required", 400)
		return
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	uploadDir := ".claude-team/uploads"
	os.MkdirAll(uploadDir, 0755)
	path := filepath.Join(uploadDir, filepath.Base(name))
	if err := os.WriteFile(path, data, 0644); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"path": path})
}

func (s *Server) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	if s.cfg.GetSettings == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{}"))
		return
	}
	settings := s.cfg.GetSettings()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

// handleIngestResult accepts an AgentResult JSON body posted by coordinators
// reading fallback JSON files and forwards it to the store. This is the
// HTTP-based alternative to the MCP report tool, which is unavailable to
// sub-agents running in Agent tool context.
func (s *Server) handleIngestResult(w http.ResponseWriter, r *http.Request) {
	if s.cfg.Store == nil {
		http.Error(w, "not configured", 500)
		return
	}

	// ingestPayload maps the fallback JSON written by agents. Agents use "phase"
	// (not "phase_id") so we accept both field names.
	var payload struct {
		Agent        string   `json:"agent"`
		RunID        string   `json:"run_id"`
		Phase        string   `json:"phase"`
		PhaseID      string   `json:"phase_id"`
		Status       string   `json:"status"`
		Confidence   string   `json:"confidence"`
		Deliverables []string `json:"deliverables"`
		Summary      string   `json:"summary"`
		Sources      []string `json:"sources"`
		Concerns     []string `json:"concerns"`
		Questions    []string `json:"questions"`
		TestsRun     string   `json:"tests_run"`
		TokensUsed   int      `json:"tokens_used"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	// Resolve phase_id: accept either field name.
	phaseID := payload.PhaseID
	if phaseID == "" {
		phaseID = payload.Phase
	}
	if phaseID == "" {
		phaseID = "unknown"
	}

	result := store.AgentResult{
		RunID:        payload.RunID,
		PhaseID:      phaseID,
		Agent:        payload.Agent,
		Status:       payload.Status,
		Confidence:   payload.Confidence,
		Summary:      payload.Summary,
		Deliverables: payload.Deliverables,
		Sources:      payload.Sources,
		Concerns:     payload.Concerns,
		Questions:    payload.Questions,
		TestsRun:     payload.TestsRun,
		TokensUsed:   payload.TokensUsed,
	}

	if err := s.cfg.Store.UpsertAgentResult(result); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	var settings map[string]string
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if s.cfg.SaveSettings != nil {
		if err := s.cfg.SaveSettings(settings); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleRunFiles(w http.ResponseWriter, r *http.Request) {
	runID := r.PathValue("id")
	if s.cfg.RuntimeDir == "" {
		http.Error(w, "not configured", 500)
		return
	}
	runDir := filepath.Join(s.cfg.RuntimeDir, "runs", runID)
	entries, err := os.ReadDir(runDir)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]string{})
		return
	}
	names := []string{}
	for _, e := range entries {
		if !e.IsDir() {
			names = append(names, e.Name())
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(names)
}

func (s *Server) handleRunFile(w http.ResponseWriter, r *http.Request) {
	runID := r.PathValue("id")
	filename := r.PathValue("filename")
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		http.Error(w, "invalid filename", 400)
		return
	}
	if s.cfg.RuntimeDir == "" {
		http.Error(w, "not configured", 500)
		return
	}
	path := filepath.Join(s.cfg.RuntimeDir, "runs", runID, filepath.Base(filename))
	data, err := os.ReadFile(path)
	if err != nil {
		http.Error(w, "file not found", 404)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"name":    filename,
		"content": string(data),
	})
}
