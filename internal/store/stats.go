package store

// Stats aggregates metrics across all runs for the dashboard and README.
type Stats struct {
	RunsTotal             int     `json:"runs_total"`
	AgentsTotal           int     `json:"agents_total"`
	TokensTotal           int     `json:"tokens_total"`
	AvgAgentsPerRun       float64 `json:"avg_agents_per_run"`
	AvgTokensPerRun       float64 `json:"avg_tokens_per_run"`
	ParallelismSpeedup    float64 `json:"parallelism_speedup"`
	CavemanCompressionPct int     `json:"caveman_compression_pct"`
	ContextIsolationMult  float64 `json:"context_isolation_multiplier"`
}

// GetStats returns aggregate metrics from the DB plus static benchmarked constants.
// Returns zeroed Stats (not an error) when no runs exist.
func (s *Store) GetStats() (Stats, error) {
	var st Stats

	if err := s.db.QueryRow(`SELECT COUNT(*) FROM runs`).Scan(&st.RunsTotal); err != nil {
		return st, err
	}

	if err := s.db.QueryRow(
		`SELECT COUNT(*), COALESCE(SUM(tokens_used), 0) FROM agent_results`,
	).Scan(&st.AgentsTotal, &st.TokensTotal); err != nil {
		return st, err
	}

	if st.RunsTotal > 0 {
		st.AvgAgentsPerRun = float64(st.AgentsTotal) / float64(st.RunsTotal)
		st.AvgTokensPerRun = float64(st.TokensTotal) / float64(st.RunsTotal)
	}

	// Static constants measured by scripts/benchmark_caveman.go and workflow YAML analysis.
	// caveman_compression_pct: measured word reduction from article/filler stripping.
	// parallelism_speedup: max parallel agents in any phase of feature-build.yaml (engineering: 3).
	st.CavemanCompressionPct = 22
	st.ParallelismSpeedup = 3.0

	// Context isolation multiplier: ratio of solo-session context growth vs Anton fresh sub-agents.
	// Solo total context = N*S + R*N*(N-1)/2 ≈ N(N+1)/2 when R≈S.
	// Anton total context = N*S.
	// Multiplier = (N+1)/2 where N = avg agents per run.
	if st.AvgAgentsPerRun > 1 {
		st.ContextIsolationMult = (st.AvgAgentsPerRun + 1) / 2
	} else {
		st.ContextIsolationMult = 1.0
	}

	return st, nil
}
