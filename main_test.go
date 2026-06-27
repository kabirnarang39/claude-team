package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kabirnarang39/claude-team/internal/store"
)

func TestVersionFlag(t *testing.T) {
	out, err := exec.Command("go", "run", ".", "--version").CombinedOutput()
	if err != nil {
		t.Fatalf("--version errored: %v\noutput: %s", err, out)
	}
	if !strings.Contains(string(out), "Anton v") {
		t.Errorf("expected 'Anton v' in output, got: %s", out)
	}
}

func TestAntonSkillFilesUseInstalledSkillDirectories(t *testing.T) {
	root := filepath.Join(t.TempDir(), "skills")

	got := antonSkillFiles(root)
	want := []string{
		filepath.Join(root, "team-dispatch", "SKILL.md"),
		filepath.Join(root, "team-resume", "SKILL.md"),
		filepath.Join(root, "team-status", "SKILL.md"),
		filepath.Join(root, "team-stop", "SKILL.md"),
	}

	if len(got) != len(want) {
		t.Fatalf("got %d skill files, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("skill file %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCheckFailureMessageUsesCheckFlag(t *testing.T) {
	text := checkRetryMessage()
	if !strings.Contains(text, "anton --check") {
		t.Fatalf("expected retry message to mention anton --check, got: %s", text)
	}
	if strings.Contains(text, "anton check") {
		t.Fatalf("retry message still mentions anton check: %s", text)
	}
}

func TestSeedDemoRunWritesFilesAndMetadata(t *testing.T) {
	dir := t.TempDir()
	db, err := store.Open(filepath.Join(dir, "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	runtimeDir := filepath.Join(dir, ".claude-team")
	if err := seedDemoRun(db, runtimeDir); err != nil {
		t.Fatal(err)
	}

	runDir := filepath.Join(runtimeDir, "runs", "demo-feature-build-jwt-auth")
	for _, name := range []string{
		"prd.md",
		"acceptance-criteria.md",
		"adr.md",
		"architecture.md",
		"qa-report.md",
		"security-report.md",
		"review-report.md",
	} {
		path := filepath.Join(runDir, name)
		if info, err := os.Stat(path); err != nil {
			t.Fatalf("expected demo file %s: %v", name, err)
		} else if info.Size() == 0 {
			t.Fatalf("expected demo file %s to be non-empty", name)
		}
	}

	detail, err := db.GetRunDetail("demo-feature-build-jwt-auth")
	if err != nil {
		t.Fatal(err)
	}
	if len(detail.Results) == 0 {
		t.Fatal("expected demo agent results")
	}
	for _, result := range detail.Results {
		if result.Status == "DONE" && len(result.Deliverables) == 0 {
			t.Fatalf("agent %s has no deliverables", result.Agent)
		}
		if strings.Contains(strings.ToLower(result.Summary), "tests") && result.TestsRun == "" {
			t.Fatalf("agent %s summary mentions tests but tests_run is empty", result.Agent)
		}
	}
}
