# Anton Demo — Storyboard & Recording Guide

Use this if you want to re-record manually or record the terminal portion.

---

## Automated dashboard recording

**File:** `docs/demo-dashboard.mp4`  
**Duration:** ~46 seconds | 1280×720 H.264 | 2.1MB  
**Regenerate:** `node /path/to/record-demo.mjs` (requires Anton server on :3000 with demo data)

---

## Shot-by-shot breakdown

| Time | What's on screen | Why it matters |
|------|-----------------|---------------|
| 0:00–0:03 | Black terminal overlay: `❯ /team-dispatch build user auth with JWT and refresh tokens` types character by character | Shows the UX — one command |
| 0:03–0:05 | Dispatch output: "12 specialists across 5 phases…" with phase breakdown | Shows what fires |
| 0:05–0:06 | Fade to dashboard — sidebar has 8 runs | Shows multiple past runs = this is a real workflow tool |
| 0:06–0:08 | Runs visible in sidebar (rate-limiting shows `running` badge) | Runs persist, dashboard is live |
| 0:08–0:11 | Click rate-limiting run → DAG renders with 5 phases, color-coded | The visual "wow" — pipeline in a browser |
| 0:11–0:15 | Engineering phase visible: 3 agents side by side (parallel) | The 3× speedup story |
| 0:15–0:18 | Click Ragnar (backend-engineer-1) → inspector opens right panel | Click any node = full output |
| 0:18–0:21 | Inspector shows: agent summary, confidence HIGH, 9,200 tokens, deliverables list | Quality + cost visibility |
| 0:21–0:24 | Click Jon Snow (security-reviewer) in QA phase (running) | Shows QA running while engineering done |
| 0:24–0:26 | Switch to Docs tab → planning docs visible | Structured deliverables per phase |
| 0:26–0:28 | Switch back to Agent tab | |
| 0:28–0:31 | Click JWT auth run in sidebar → full 5-phase DAG, all done | Shows a completed run |
| 0:31–0:34 | All 5 phases green ✓ — 10 agents visible | Full pipeline completed |
| 0:34–0:37 | Click Jon Snow (security-reviewer) in JWT run → "OWASP Top 10, no criticals" | Security built-in, not an afterthought |
| 0:37–0:40 | Click Floki (devops-engineer) → "Multi-stage Dockerfile, Helm chart, GitHub Actions" | Full stack coverage |
| 0:40–0:43 | Switch to Deliverables tab → list of files produced | Real files, not just summaries |
| 0:43–0:46 | Stripe-webhook run loads — 3 backend engineers visible in parallel | Shows different workflow, 3× engineers |

---

## If you want to add a terminal recording

Record the terminal portion separately, then edit together with `ffmpeg`:

```bash
# Record terminal with asciinema or QuickTime
# Then merge:
ffmpeg -i terminal-clip.mp4 -i docs/demo-dashboard.mp4 \
  -filter_complex "[0:v][1:v]concat=n=2:v=1:a=0[v]" \
  -map "[v]" docs/demo-final.mp4
```

Terminal clip to record (20 seconds):
```
$ anton
⚡ Anton v3 — localhost:3000
[hit enter, open browser]

> /team-dispatch build user auth with JWT and refresh tokens
⚡ Anton dispatching 12 specialists across 5 phases…
  Planning    → requirements-analyst, tech-writer
  Architecture → senior-architect, api-designer
  Engineering  → backend × 3, frontend, dba  ↺ parallel
  QA           → qa-engineer, security-reviewer, e2e-tester
  DevOps       → code-reviewer, devops-engineer
Dashboard → http://localhost:3000
[cut to browser]
```

---

## Adding music

Free no-copyright tracks that fit the vibe:
- Search "lofi chill instrumental no copyright" on YouTube Audio Library
- Target BPM: 80–100 (matches the pacing of the agent walkthrough)
- Fade in at 0:00, fade out at end

```bash
# Add music track:
ffmpeg -i docs/demo-dashboard.mp4 -i music.mp3 \
  -map 0:v -map 1:a -c:v copy -c:a aac \
  -shortest docs/demo-with-music.mp4
```

---

## Upload checklist

- [ ] X/Twitter: upload `demo-dashboard.mp4` directly (≤512MB, ≤2:20)
- [ ] YouTube: upload as unlisted first, check thumbnail, then publish
- [ ] Reddit: use video upload on r/ClaudeAI (not a link — native video gets more views)
- [ ] HN: paste the Show HN text from `docs/social-copy.md`
- [ ] README: replace `docs/demo.gif` reference with the new video embed (use a GitHub-hosted thumbnail linking to YouTube)
