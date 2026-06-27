# Launch Checklist

## Before Posting

- [ ] `go test -race -coverprofile=/tmp/anton-coverage.out -covermode=atomic ./...`
- [ ] `go vet ./...`
- [ ] `node scripts/validate-mcp-registry.mjs`
- [ ] `bash -n install.sh`
- [ ] `cd mcp && npm ci --omit=dev`
- [ ] `go run main.go --demo --port 3100`
- [ ] Open dashboard and inspect demo run, inspector tabs, deliverables, and mobile-ish width.
- [ ] Record a short GIF/video using `docs/demo/recording-script.md`.
- [ ] Verify README install command against the latest release.
- [ ] Confirm release binaries exist for macOS arm64/amd64 and Linux amd64.

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
