package store

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

type AgentResult struct {
	ID           int64    `json:"id"`
	RunID        string   `json:"run_id"`
	PhaseID      string   `json:"phase_id"`
	Agent        string   `json:"agent"`
	Status       string   `json:"status"`
	Confidence   string   `json:"confidence"`
	Summary      string   `json:"summary"`
	Deliverables []string `json:"deliverables"`
	Sources      []string `json:"sources"`
	Concerns     []string `json:"concerns"`
	Questions    []string `json:"questions"`
	TestsRun     string   `json:"tests_run"`
	TokensUsed   int      `json:"tokens_used"`
	CreatedAt    int64    `json:"created_at"`
}

type Phase struct {
	RunID       string `json:"run_id"`
	PhaseID     string `json:"phase_id"`
	Status      string `json:"status"`
	StartedAt   int64  `json:"started_at"`
	CompletedAt int64  `json:"completed_at"`
}

type HumanReview struct {
	ID         int64  `json:"id"`
	RunID      string `json:"run_id"`
	Gate       string `json:"gate"`
	Status     string `json:"status"`
	Summary    string `json:"summary"`
	Feedback   string `json:"feedback"`
	CreatedAt  int64  `json:"created_at"`
	ResolvedAt int64  `json:"resolved_at,omitempty"`
}

type RunDetail struct {
	ID           string        `json:"id"`
	WorkflowName string        `json:"workflow_name"`
	Status       string        `json:"status"`
	StartedAt    int64         `json:"started_at"`
	CompletedAt  int64         `json:"completed_at"`
	Phases       []Phase       `json:"phases"`
	Results      []AgentResult `json:"results"`
	Reviews      []HumanReview `json:"reviews"`
}

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	s := &Store{db: db}
	return s, s.migrate()
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate() error {
	if _, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS schema_version (version INTEGER NOT NULL)`); err != nil {
		return err
	}
	var version int
	if err := s.db.QueryRow(`SELECT COALESCE(MAX(version), 0) FROM schema_version`).Scan(&version); err != nil {
		return err
	}

	if version < 1 {
		_, err := s.db.Exec(`
			CREATE TABLE IF NOT EXISTS runs (
				id TEXT PRIMARY KEY,
				workflow_name TEXT NOT NULL,
				status TEXT NOT NULL DEFAULT 'running',
				started_at INTEGER NOT NULL,
				completed_at INTEGER
			);
			CREATE TABLE IF NOT EXISTS agent_statuses (
				run_id TEXT NOT NULL,
				agent TEXT NOT NULL,
				status TEXT NOT NULL DEFAULT 'pending',
				updated_at INTEGER NOT NULL,
				PRIMARY KEY (run_id, agent)
			);
			CREATE TABLE IF NOT EXISTS messages (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				from_agent TEXT NOT NULL,
				to_agent TEXT NOT NULL,
				content TEXT NOT NULL,
				created_at INTEGER NOT NULL,
				read_at INTEGER
			);
		`)
		if err != nil {
			return err
		}
		s.db.Exec(`INSERT INTO schema_version VALUES (1)`)
	}

	if version < 2 {
		_, err := s.db.Exec(`
			CREATE TABLE IF NOT EXISTS phases (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				run_id TEXT NOT NULL,
				phase_id TEXT NOT NULL,
				status TEXT NOT NULL DEFAULT 'pending',
				started_at INTEGER,
				completed_at INTEGER,
				UNIQUE(run_id, phase_id)
			);
			CREATE TABLE IF NOT EXISTS agent_results (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				run_id TEXT NOT NULL,
				phase_id TEXT NOT NULL,
				agent TEXT NOT NULL,
				status TEXT NOT NULL,
				confidence TEXT DEFAULT 'medium',
				summary TEXT DEFAULT '',
				deliverables_json TEXT DEFAULT '[]',
				sources_json TEXT DEFAULT '[]',
				concerns_json TEXT DEFAULT '[]',
				questions_json TEXT DEFAULT '[]',
				tests_run TEXT DEFAULT '',
				tokens_used INTEGER DEFAULT 0,
				created_at INTEGER NOT NULL
			);
		`)
		if err != nil {
			return err
		}
		// Best-effort — columns may already exist on upgraded DBs
		s.db.Exec(`ALTER TABLE messages ADD COLUMN run_id TEXT`)
		s.db.Exec(`ALTER TABLE messages ADD COLUMN response TEXT`)
		s.db.Exec(`INSERT INTO schema_version VALUES (2)`)
	}

	if version < 3 {
		_, err := s.db.Exec(`
			CREATE TABLE IF NOT EXISTS human_reviews (
				id          INTEGER PRIMARY KEY AUTOINCREMENT,
				run_id      TEXT    NOT NULL,
				gate        TEXT    NOT NULL,
				status      TEXT    NOT NULL DEFAULT 'pending',
				summary     TEXT    DEFAULT '',
				feedback    TEXT    DEFAULT '',
				created_at  INTEGER NOT NULL,
				resolved_at INTEGER
			);
		`)
		if err != nil {
			return err
		}
		s.db.Exec(`INSERT INTO schema_version VALUES (3)`)
	}

	_, err := s.db.Exec(`PRAGMA journal_mode=WAL`)
	return err
}
