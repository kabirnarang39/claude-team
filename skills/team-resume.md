---
name: team-resume
description: Resume a halted or incomplete Anton run from checkpoint
trigger: /team-resume
---

# Anton: Team Resume

Resume an incomplete or halted Anton run using its checkpoint.json.

## Arguments

- No args: list all incomplete runs
- `<run_id>`: resume the specified run

## Steps

### No-argument mode (list incomplete runs)

If no run_id is provided:

1. List all run directories:
   ```bash
   ls .claude-team/runs/
   ```

2. For each directory, check for checkpoint.json and print status:
   ```bash
   for d in .claude-team/runs/*/; do
     run_id=$(basename "$d")
     ckpt="$d/checkpoint.json"
     if [ -f "$ckpt" ]; then
       current=$(python3 -c "import json; d=json.load(open('$ckpt')); print(d.get('current_phase','?'))")
       done=$(python3 -c "import json; d=json.load(open('$ckpt')); print(', '.join(d.get('completed_phases',[])) or 'none')")
       halted=$(python3 -c "import json; d=json.load(open('$ckpt')); print(d.get('halted_reason',''))")
       echo "$run_id | current_phase: $current | completed: $done | halted: $halted"
     else
       echo "$run_id | no checkpoint.json"
     fi
   done
   ```

3. Print:
   ```
   Incomplete Anton runs:
   <run_id>  current_phase: <phase>  completed: <phases>  halted: <reason or none>

   To resume: /team-resume <run_id>
   ```

4. Stop here — do not resume automatically.

### Resume mode (run_id provided)

#### Step 1 — Locate checkpoint.json

```bash
cat .claude-team/runs/<run_id>/checkpoint.json
```

If the file does not exist, fall back to database state:

```bash
sqlite3 .claude-team/state.db "
  SELECT r.id, r.workflow_name, r.status,
         GROUP_CONCAT(p.phase_id || ':' || p.status, ',') as phases
  FROM runs r
  LEFT JOIN phases p ON p.run_id = r.id
  WHERE r.id = '<run_id>'
  GROUP BY r.id;
"
```

Construct a synthetic checkpoint from DB output:
- `current_phase` = first phase with status not `completed`
- `completed_phases` = list of phase_ids with status `completed`
- `completed_agents` = `{}`
- `workflow_name` = from DB `workflow_name` column

Write the reconstructed checkpoint to disk so future resumes don't need to fall back again:
```bash
python3 -c "
import json
data = {
  'run_id': 'ACTUAL_RUN_ID',
  'workflow_name': 'ACTUAL_WORKFLOW',
  'current_phase': 'ACTUAL_CURRENT_PHASE',
  'completed_phases': ['COMPLETED_PHASE_1', 'COMPLETED_PHASE_2'],
  'completed_agents': {}
}
with open('.claude-team/runs/ACTUAL_RUN_ID/checkpoint.json', 'w') as f:
    json.dump(data, f, indent=2)
"
```

Print: `Note: checkpoint.json not found — reconstructed from DB state and written to disk.`

#### Step 2 — Halted-reason confirmation gate

If `halted_reason` is present and non-empty in checkpoint.json:

Print exactly:
```
── HALTED RUN ─────────────────────────────────────────────────────────────
This run was halted.
Reason: <halted_reason>

Run ID:          <run_id>
Current phase:   <current_phase>
Completed:       <completed_phases joined by ", " or "none">

Type  resume    to continue from the current phase.
Type  abort     to mark this run as stopped and exit.
───────────────────────────────────────────────────────────────────────────
```

Wait for user response:
- `resume` → clear `halted_reason` from checkpoint.json then continue to Step 3:
  ```bash
  python3 -c "
  import json
  with open('.claude-team/runs/<run_id>/checkpoint.json') as f:
      ck = json.load(f)
  ck.pop('halted_reason', None)
  with open('.claude-team/runs/<run_id>/checkpoint.json', 'w') as f:
      json.dump(ck, f, indent=2)
  "
  ```
- `abort` → mark run stopped and exit:
  ```bash
  sqlite3 .claude-team/state.db "UPDATE runs SET status='stopped' WHERE id='<run_id>';"
  ```
  Print: `Anton run <run_id> marked as stopped.`
  Stop here.
- Any other response: re-print the prompt above and wait again.

If `halted_reason` is absent or empty, skip this gate and proceed to Step 3.

#### Step 3 — Dispatch to main coordinator in RESUME MODE

Read `.claude-team/runs/<run_id>/approach.md` if it exists (for context).

Write a resume brief to `.claude-team/pending-task.md`:
```
RESUME MODE
Run ID: <run_id>
```

Then invoke the main coordinator exactly as `team-dispatch` does, but pass `RESUME MODE` so it reads from `~/.claude/anton/coordinators/main.md` Resume Mode section.

Print:
```
Resuming Anton run <run_id>.
Current phase: <current_phase>
Completed:     <completed_phases joined by ", " or "none">
```

Then read `~/.claude/anton/coordinators/main.md` fully and follow the **Resume Mode** section from the top of that file.
