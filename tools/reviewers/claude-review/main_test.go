package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRunNeedsTheShelfAndOneRequest(t *testing.T) {
	for _, args := range [][]string{{}, {"REQ.json"}, {"-shelf", "x"}, {"-shelf", "x", "a.json", "b.json"}, {"-nope"}} {
		if code := run(args); code != 2 {
			t.Errorf("%q: exit code %d, want 2", args, code)
		}
	}
}

// fakeClaude writes an executable that prints answer and exits with code, standing in for
// the claude command.
func fakeClaude(t *testing.T, answer string, code int) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the fake claude is a shell script")
	}
	path := filepath.Join(t.TempDir(), "claude")
	script := "#!/bin/sh\ncat <<'END'\n" + answer + "\nEND\nexit " + string(rune('0'+code)) + "\n"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRunAnswersARequestWithTheClaudeCommand(t *testing.T) {
	path := setup(t, "pending")
	claude := fakeClaude(t, answer(`{"summary":"s","operations":[{"op":"add_entry","fields":{}}]}`), 0)
	if code := run([]string{"-shelf", t.TempDir(), "-claude", claude, path}); code != 0 {
		t.Fatalf("exit code %d", code)
	}
	if read(t, path)["status"] != "proposed" {
		t.Errorf("status: %v", read(t, path)["status"])
	}
}

func TestRunExitsWithOneWhenTheRequestFailed(t *testing.T) {
	path := setup(t, "pending")
	claude := fakeClaude(t, "not json", 1)
	if code := run([]string{"-shelf", t.TempDir(), "-claude", claude, path}); code != 1 {
		t.Fatalf("exit code %d, want 1", code)
	}
	if f, _ := read(t, path)["failure"].(string); !strings.Contains(f, "did not return JSON (exit code 1)") {
		t.Errorf("failure: %q", f)
	}
}

func TestRunFailsTheRequestWhenClaudeIsMissing(t *testing.T) {
	path := setup(t, "pending")
	if code := run([]string{"-shelf", t.TempDir(), "-claude", filepath.Join(t.TempDir(), "no-claude"), path}); code != 1 {
		t.Fatalf("exit code %d, want 1", code)
	}
	if f, _ := read(t, path)["failure"].(string); !strings.Contains(f, "Could not run Claude Code") {
		t.Errorf("failure: %q", f)
	}
}
