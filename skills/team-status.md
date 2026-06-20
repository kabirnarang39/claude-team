---
name: team-status
description: Show current Anton run status
trigger: /team-status
---

# Anton: Team Status

Read `.claude-team/state.db` and print current run status.

## Steps

1. Open `.claude-team/state.db` (SQLite)

2. Query latest run:
   ```sql
   SELECT id, workflow_name, status, started_at
   FROM runs ORDER BY started_at DESC LIMIT 1
   ```

3. Query phases for that run:
   ```sql
   SELECT phase_id, status, started_at, completed_at
   FROM phases WHERE run_id = '<latest_run_id>'
   ORDER BY id ASC
   ```

4. Query agent results:
   ```sql
   SELECT agent, phase_id, status, confidence, summary
   FROM agent_results WHERE run_id = '<latest_run_id>'
   ORDER BY id ASC
   ```

5. Print formatted status:
   ```
   Run: <id>  Workflow: <name>  Status: <status>
   
   Phase          Status    Agents
   planning       done      requirements-analyst ✓  tech-writer ✓
   architecture   running   senior-architect ⟳
   engineering    pending   —
   ```
