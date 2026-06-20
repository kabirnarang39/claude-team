package main

import (
	"os/exec"
	"strings"
	"testing"
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
