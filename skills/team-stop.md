---
name: team-stop
description: Halt current Anton run
trigger: /team-stop
---

# Anton: Team Stop

Halt current Anton run gracefully.

## Steps

1. Open `.claude-team/state.db`

2. Find latest running run:
   ```sql
   SELECT id FROM runs WHERE status = 'running'
   ORDER BY started_at DESC LIMIT 1
   ```

3. If no running run: print "No active Anton run."

4. Mark stopped:
   ```sql
   UPDATE runs SET status='stopped', completed_at=<now>
   WHERE id = '<run_id>'
   ```

5. Print:
   ```
   Anton run <run_id> stopped.
   Partial outputs in: .claude-team/runs/<run_id>/
   ```

Note: Do not delete partial outputs. They remain for inspection.
