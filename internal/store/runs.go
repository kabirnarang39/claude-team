package store

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// PhaseAgentPair is the store-layer type for workflow phase+agent pre-population.
type PhaseAgentPair struct {
	PhaseID string
	Agents  []string
}

func (s *Store) CreateRun(workflowName string) (string, error) {
	id := randomID()
	_, err := s.db.Exec(
		`INSERT INTO runs (id, workflow_name, status, started_at) VALUES (?, ?, 'running', ?)`,
		id, workflowName, time.Now().Unix(),
	)
	return id, err
}

func (s *Store) UpdateAgentStatus(runID, agent, status string) error {
	_, err := s.db.Exec(`
		INSERT INTO agent_statuses (run_id, agent, status, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(run_id, agent) DO UPDATE SET status=excluded.status, updated_at=excluded.updated_at
	`, runID, agent, status, time.Now().Unix())
	return err
}

func (s *Store) GetRunStatuses(runID string) (map[string]string, error) {
	rows, err := s.db.Query(`SELECT agent, status FROM agent_statuses WHERE run_id = ?`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := map[string]string{}
	for rows.Next() {
		var agent, status string
		if err := rows.Scan(&agent, &status); err != nil {
			return nil, err
		}
		result[agent] = status
	}
	return result, rows.Err()
}

func (s *Store) GetAllStatuses() (map[string]string, error) {
	rows, err := s.db.Query(`
		SELECT agent, status FROM agent_statuses
		WHERE run_id = (SELECT id FROM runs ORDER BY started_at DESC LIMIT 1)
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := map[string]string{}
	for rows.Next() {
		var agent, status string
		if err := rows.Scan(&agent, &status); err != nil {
			return nil, err
		}
		result[agent] = status
	}
	return result, rows.Err()
}

func (s *Store) SendMessage(fromAgent, toAgent, content string) error {
	_, err := s.db.Exec(
		`INSERT INTO messages (from_agent, to_agent, content, created_at) VALUES (?, ?, ?, ?)`,
		fromAgent, toAgent, content, time.Now().Unix(),
	)
	return err
}

func (s *Store) ReadPendingMessages(agent string) ([]string, error) {
	rows, err := s.db.Query(
		`SELECT id, content FROM messages WHERE to_agent = ? AND read_at IS NULL ORDER BY created_at`,
		agent,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int64
	var msgs []string
	for rows.Next() {
		var id int64
		var content string
		if err := rows.Scan(&id, &content); err != nil {
			return nil, err
		}
		ids = append(ids, id)
		msgs = append(msgs, content)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for _, id := range ids {
		s.db.Exec(`UPDATE messages SET read_at = ? WHERE id = ?`, time.Now().Unix(), id)
	}
	return msgs, nil
}

func (s *Store) UpsertPhase(runID, phaseID, status string) error {
	_, err := s.db.Exec(`
		INSERT INTO phases (run_id, phase_id, status, started_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(run_id, phase_id) DO UPDATE SET
			status=excluded.status,
			completed_at=CASE WHEN excluded.status IN ('done','failed') THEN ? ELSE completed_at END
	`, runID, phaseID, status, time.Now().Unix(), time.Now().Unix())
	return err
}

func (s *Store) InsertAgentResult(r AgentResult) error {
	deliverables, _ := json.Marshal(r.Deliverables)
	sources, _ := json.Marshal(r.Sources)
	concerns, _ := json.Marshal(r.Concerns)
	questions, _ := json.Marshal(r.Questions)
	if r.CreatedAt == 0 {
		r.CreatedAt = time.Now().Unix()
	}
	_, err := s.db.Exec(`
		INSERT INTO agent_results
		(run_id, phase_id, agent, status, confidence, summary,
		 deliverables_json, sources_json, concerns_json, questions_json,
		 tests_run, tokens_used, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, r.RunID, r.PhaseID, r.Agent, r.Status, r.Confidence, r.Summary,
		string(deliverables), string(sources), string(concerns), string(questions),
		r.TestsRun, r.TokensUsed, r.CreatedAt)
	return err
}

func (s *Store) GetAgentResultsSince(lastID int64) ([]AgentResult, int64, error) {
	rows, err := s.db.Query(`
		SELECT id, run_id, phase_id, agent, status, confidence, summary,
		       deliverables_json, sources_json, concerns_json, questions_json,
		       tests_run, tokens_used, created_at
		FROM agent_results WHERE id > ? ORDER BY id ASC
	`, lastID)
	if err != nil {
		return nil, lastID, err
	}
	defer rows.Close()
	var results []AgentResult
	var maxID int64 = lastID
	for rows.Next() {
		var r AgentResult
		var deliverables, sources, concerns, questions string
		err := rows.Scan(&r.ID, &r.RunID, &r.PhaseID, &r.Agent, &r.Status,
			&r.Confidence, &r.Summary, &deliverables, &sources, &concerns,
			&questions, &r.TestsRun, &r.TokensUsed, &r.CreatedAt)
		if err != nil {
			return nil, maxID, err
		}
		json.Unmarshal([]byte(deliverables), &r.Deliverables)
		json.Unmarshal([]byte(sources), &r.Sources)
		json.Unmarshal([]byte(concerns), &r.Concerns)
		json.Unmarshal([]byte(questions), &r.Questions)
		if r.ID > maxID {
			maxID = r.ID
		}
		results = append(results, r)
	}
	return results, maxID, rows.Err()
}

func (s *Store) GetRuns(limit int) ([]RunDetail, error) {
	rows, err := s.db.Query(`
		SELECT id, workflow_name, status, started_at, COALESCE(completed_at,0)
		FROM runs ORDER BY started_at DESC LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var runs []RunDetail
	for rows.Next() {
		var r RunDetail
		if err := rows.Scan(&r.ID, &r.WorkflowName, &r.Status, &r.StartedAt, &r.CompletedAt); err != nil {
			return nil, err
		}
		runs = append(runs, r)
	}
	return runs, rows.Err()
}

func (s *Store) GetRunDetail(runID string) (*RunDetail, error) {
	var r RunDetail
	err := s.db.QueryRow(`
		SELECT id, workflow_name, status, started_at, COALESCE(completed_at,0)
		FROM runs WHERE id = ?
	`, runID).Scan(&r.ID, &r.WorkflowName, &r.Status, &r.StartedAt, &r.CompletedAt)
	if err != nil {
		return nil, err
	}

	phaseRows, err := s.db.Query(`
		SELECT run_id, phase_id, status, COALESCE(started_at,0), COALESCE(completed_at,0)
		FROM phases WHERE run_id = ? ORDER BY rowid ASC
	`, runID)
	if err != nil {
		return nil, err
	}
	defer phaseRows.Close()
	for phaseRows.Next() {
		var p Phase
		phaseRows.Scan(&p.RunID, &p.PhaseID, &p.Status, &p.StartedAt, &p.CompletedAt)
		r.Phases = append(r.Phases, p)
	}

	resultRows, err := s.db.Query(`
		SELECT id, run_id, phase_id, agent, status, confidence, summary,
		       deliverables_json, sources_json, concerns_json, questions_json,
		       tests_run, tokens_used, created_at
		FROM agent_results
		WHERE run_id = ?
		  AND id IN (
		      SELECT MAX(id) FROM agent_results
		      WHERE run_id = ?
		      GROUP BY phase_id, agent
		  )
		ORDER BY id ASC
	`, runID, runID)
	if err != nil {
		return nil, err
	}
	defer resultRows.Close()
	for resultRows.Next() {
		var res AgentResult
		var deliverables, sources, concerns, questions string
		if err := resultRows.Scan(&res.ID, &res.RunID, &res.PhaseID, &res.Agent,
			&res.Status, &res.Confidence, &res.Summary,
			&deliverables, &sources, &concerns, &questions,
			&res.TestsRun, &res.TokensUsed, &res.CreatedAt); err != nil {
			continue
		}
		json.Unmarshal([]byte(deliverables), &res.Deliverables)
		json.Unmarshal([]byte(sources), &res.Sources)
		json.Unmarshal([]byte(concerns), &res.Concerns)
		json.Unmarshal([]byte(questions), &res.Questions)
		r.Results = append(r.Results, res)
	}
	return &r, nil
}

func (s *Store) CreateRunWithID(id, workflowName string) error {
	_, err := s.db.Exec(
		`INSERT OR IGNORE INTO runs (id, workflow_name, status, started_at) VALUES (?, ?, 'pending', ?)`,
		id, workflowName, time.Now().Unix(),
	)
	return err
}

// PrePopulateAgents inserts PENDING placeholder rows for all expected agents.
// GetRunDetail deduplicates via MAX(id) per (phase_id, agent), so a later DONE
// row from the real agent automatically supersedes the PENDING placeholder.
func (s *Store) PrePopulateAgents(runID string, pairs []PhaseAgentPair) {
	now := time.Now().Unix()
	for _, pa := range pairs {
		s.db.Exec(`INSERT OR IGNORE INTO phases (run_id, phase_id, status, started_at) VALUES (?,?,'pending',?)`,
			runID, pa.PhaseID, now)
		for _, agent := range pa.Agents {
			s.db.Exec(`INSERT OR IGNORE INTO agent_results
				(run_id, phase_id, agent, status, confidence, summary,
				 deliverables_json, sources_json, concerns_json, questions_json,
				 tests_run, tokens_used, created_at)
				VALUES (?,?,?,'PENDING','','','[]','[]','[]','[]','',0,?)`,
				runID, pa.PhaseID, agent, now)
		}
	}
}

// UpsertAgentResult inserts or replaces an agent result row.
// It maps the agent JSON "phase" field to the DB "phase_id" column.
// After writing the result it also upserts the phase row so the dashboard
// reflects the agent's terminal status.
func (s *Store) UpsertAgentResult(r AgentResult) error {
	// Normalise: agents write "phase" in their JSON; store uses phase_id.
	if r.PhaseID == "" {
		r.PhaseID = "unknown"
	}
	if err := s.InsertAgentResult(r); err != nil {
		return err
	}
	// Ensure the run is marked running once the first agent result arrives.
	s.db.Exec(`UPDATE runs SET status='running' WHERE id=? AND status='pending'`, r.RunID)

	// Map agent status → phase status.
	phaseStatus := "running"
	switch r.Status {
	case "DONE", "DONE_WITH_CONCERNS":
		phaseStatus = "done"
	case "BLOCKED", "FAILED":
		phaseStatus = "failed"
	}
	return s.UpsertPhase(r.RunID, r.PhaseID, phaseStatus)
}

func (s *Store) WriteTask(runtimeDir, text string) error {
	content := "# Pending Task\n\n" + text + "\n"
	return os.WriteFile(filepath.Join(runtimeDir, "pending-task.md"), []byte(content), 0644)
}

func randomID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}
