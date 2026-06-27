# Launch Checklist

## Before Posting

- [x] `go test -race -coverprofile=/tmp/anton-coverage.out -covermode=atomic ./...`
- [x] `go vet ./...`
- [x] `node scripts/validate-mcp-registry.mjs`
- [x] `bash -n install.sh`
- [x] `cd mcp && npm ci --omit=dev`
- [x] `go run main.go --demo --port 3100`
- [x] Open dashboard and inspect demo run, inspector tabs, deliverables, and mobile-ish width.
- [x] Add README demo GIF: `docs/assets/anton-demo.gif`
- [x] Add dashboard screenshot with phase progress: `docs/assets/anton-dashboard.png`
- [x] Add Docs tab screenshot: `docs/assets/anton-docs.png`
- [x] Add Deliverables tab screenshot: `docs/assets/anton-deliverables.png`
- [x] Add mobile-width screenshot: `docs/assets/anton-mobile.png`
- [x] Verify README install command against the latest release. GitHub latest release API returned `v1.4.5`.
- [x] Confirm release binaries exist for macOS arm64/amd64 and Linux amd64: `anton-darwin-arm64`, `anton-darwin-amd64`, `anton-linux-amd64`.
- [ ] Create and pin a GitHub issue asking for workflow/role feedback. Local `gh` is not logged in; issue copy is staged in `docs/launch/feedback-issue.md`.

## Post Order

1. GitHub release notes
2. README screenshots or demo GIF
3. Hacker News
4. X/Twitter thread
5. Reddit follow-up
6. Product Hunt only after first external feedback

## Guardrails

- Do not claim benchmarked speedups unless a benchmark is linked.
- Do not claim 25 integrations.
- Do not imply Anton works without Claude Code for real runs.
- Do not describe sample outputs as user results.
- Answer skeptical comments with implementation details, not hype.

## Proof Assets

- Demo GIF: `docs/assets/anton-demo.gif`
- Dashboard screenshot: `docs/assets/anton-dashboard.png`
- Docs screenshot: `docs/assets/anton-docs.png`
- Deliverables screenshot: `docs/assets/anton-deliverables.png`
- Mobile screenshot: `docs/assets/anton-mobile.png`

Regenerate the GIF from the three desktop screenshots with:

```bash
ffmpeg -y \
  -loop 1 -t 1.8 -i docs/assets/anton-dashboard.png \
  -loop 1 -t 1.8 -i docs/assets/anton-docs.png \
  -loop 1 -t 1.8 -i docs/assets/anton-deliverables.png \
  -filter_complex "[0:v]scale=960:-1:flags=lanczos,fps=5,setsar=1[v0];[1:v]scale=960:-1:flags=lanczos,fps=5,setsar=1[v1];[2:v]scale=960:-1:flags=lanczos,fps=5,setsar=1[v2];[v0][v1][v2]concat=n=3:v=1:a=0,split[s0][s1];[s0]palettegen=max_colors=96[p];[s1][p]paletteuse=dither=bayer:bayer_scale=5" \
  -loop 0 docs/assets/anton-demo.gif
```
