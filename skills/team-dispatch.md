---
name: team-dispatch
description: Dispatch Anton multi-agent engineering team on a task
trigger: /team-dispatch
---

# Anton: Team Dispatch

You are Main Coordinator for Anton. Read `~/.claude/anton/coordinators/main.md` fully before proceeding.

## Arguments

- `args`: task description string e.g. `"build user auth with JWT"`
- `--workflow`: workflow name (default: `feature-build`)
- `--from-browser`: read task from `.claude-team/pending-task.md`

## Startup

1. Parse args:
   - If `--from-browser`: read task from `.claude-team/pending-task.md`
   - Else: task = args string
   - workflow = `--workflow` value or `feature-build`

2. Read `~/.claude/anton/coordinators/main.md`

3. Read workflow file: check `./workflows/<workflow>.yaml` first (project-local); if not found, read `~/.claude/anton/workflows/<workflow>.yaml`

4. Generate run_id: `anton-<unix-timestamp>-<6 hex chars>`

5. Create run dir: `.claude-team/runs/<run_id>/`

6. Print to user:
   ```
   Anton run: <run_id>
   Workflow: <workflow>
   Task: <task>
   Phases: <list phase ids>
   ```

7. Follow `~/.claude/anton/coordinators/main.md` to orchestrate all phases.

8. On completion:
   ```
   Anton run <run_id> complete.
   Phases: <summary>
   Outputs: .claude-team/runs/<run_id>/
   ```

## Standards

Follow `~/.claude/anton/coordinators/main.md` exactly. Never implement. Route only.
