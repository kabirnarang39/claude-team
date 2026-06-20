---
name: team-dispatch
description: Dispatch Anton multi-agent engineering team on a task
trigger: /team-dispatch
---

# Anton: Team Dispatch

You are Main Coordinator for Anton. Read `~/.claude/anton/coordinators/main.md` fully before proceeding.

## Arguments

- `args`: task description string e.g. `"build user auth with JWT"`
- `--workflow <name>`: workflow name (default: `feature-build`)
- `--from-browser`: task already registered via UI — pending-task.md already written with Run ID and Workflow

## Startup

1. Parse args:
   - Extract `--workflow <name>` flag value if present (default: `feature-build`)
   - Check for `--from-browser` flag
   - task = remaining args after stripping flag tokens

2. Determine run_id, workflow, and task:

   **If `--from-browser`:**
   - Read `.claude-team/pending-task.md`
   - Parse `Run ID:` line → use as run_id (already registered in DB by the UI — do NOT generate a new one)
   - Parse `Workflow:` line → use as workflow (override with `--workflow` flag value if also passed)
   - Task text = content of the file below the header lines

   **If NOT `--from-browser` (direct CLI dispatch):**
   - workflow = `--workflow` value or `feature-build`
   - Register run in DB and write pending-task.md by running:
     ```bash
     curl -s -X POST http://localhost:3000/api/task \
       -H "Content-Type: application/json" \
       -d "{\"text\": \"$(echo "$TASK" | sed 's/"/\\"/g')\", \"workflow\": \"$WORKFLOW\"}"
     ```
     (substitute actual task text and workflow name — escape quotes in task text)
   - Parse `run_id` from the JSON response `{"run_id": "..."}` 
   - If curl fails (server unreachable): generate fallback run_id `anton-<unix-timestamp>-<6 hex chars>`, create `.claude-team/runs/<run_id>/`, write `.claude-team/pending-task.md` manually:
     ```
     Run ID: <run_id>
     Workflow: <workflow>

     <task text>
     ```

3. Read workflow file: check `./workflows/<workflow>.yaml` first (project-local); if not found, read `~/.claude/anton/workflows/<workflow>.yaml`

4. Print to user:
   ```
   Anton run: <run_id>
   Workflow: <workflow>
   Task: <task>
   Phases: <list phase ids from workflow>
   ```

5. Follow `~/.claude/anton/coordinators/main.md` to orchestrate all phases.

6. On completion:
   ```
   Anton run <run_id> complete.
   Phases: <summary>
   Outputs: .claude-team/runs/<run_id>/
   ```

## Standards

Follow `~/.claude/anton/coordinators/main.md` exactly. Never implement. Route only.
