package api

import (
	"net/http"

	"claude-team/internal/store"
)

// Config holds all dependencies the server needs.
type Config struct {
	Hub        *Hub
	UIDir      string
	Store        *store.Store
	RuntimeDir   string
	WorkflowDirs []string

	// Callbacks injected by main.go
	OnWorkflowUpload  func(data []byte, filename string) error
	GetActiveWorkflow func() ([]byte, error)
	SetActiveWorkflow func(name string) error
	GetWorkflowList   func() ([]string, error)

	// Builder API
	GetWorkflowRaw func(name string) ([]byte, error)
	SaveWorkflow   func(name, yamlContent string) error
	GetMCPList     func() []string

	// GetSettings returns current server settings as key→value map.
	GetSettings func() map[string]string
	// SaveSettings persists updated settings.
	SaveSettings func(settings map[string]string) error
}

type Server struct {
	cfg *Config
}

func NewServer(cfg Config) *Server {
	return &Server{cfg: &cfg}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/", noCache(http.FileServer(http.Dir(s.cfg.UIDir))))
	mux.HandleFunc("GET /api/status", s.handleStatus)
	mux.HandleFunc("POST /api/task", s.handleTask)
	mux.HandleFunc("GET /api/runs", s.handleRuns)
	mux.HandleFunc("GET /api/runs/{id}", s.handleRunDetail)
	mux.HandleFunc("GET /api/runs/{id}/files", s.handleRunFiles)
	mux.HandleFunc("GET /api/runs/{id}/files/{filename}", s.handleRunFile)
	mux.HandleFunc("POST /api/workflow/upload", s.handleWorkflowUpload)
	mux.HandleFunc("GET /api/workflow/active", s.handleWorkflowActive)
	mux.HandleFunc("PUT /api/workflow/active", s.handleWorkflowSetActive)
	mux.HandleFunc("GET /api/workflows", s.handleWorkflowList)
	mux.HandleFunc("GET /api/workflow/raw", s.handleWorkflowRaw)
	mux.HandleFunc("POST /api/workflow/save", s.handleWorkflowSave)
	mux.HandleFunc("GET /api/mcp-registry", s.handleMCPRegistry)
	mux.HandleFunc("POST /api/files/upload", s.handleFileUpload)
	mux.HandleFunc("GET /api/settings", s.handleGetSettings)
	mux.HandleFunc("POST /api/settings", s.handleSettings)
	mux.HandleFunc("POST /api/ingest-result", s.handleIngestResult)
	mux.HandleFunc("GET /ws", s.cfg.Hub.ServeWS)
	return mux
}

func noCache(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		h.ServeHTTP(w, r)
	})
}
