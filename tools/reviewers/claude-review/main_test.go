package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Rafael24595/go-assert/assert"
	"github.com/Rafael24595/go-assert/require"
)

func TestRunNeedsTheShelfAndOneRequest(t *testing.T) {
	for _, args := range [][]string{{}, {"REQ.json"}, {"-shelf", "x"}, {"-shelf", "x", "a.json", "b.json"}, {"-nope"}} {
		assert.Equal(t, 2, run(args), "%q: exit code", args)
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
	require.Nil(t, os.WriteFile(path, []byte(script), 0o755))
	return path
}

func TestRunAnswersARequestWithTheClaudeCommand(t *testing.T) {
	path := setup(t, "pending")
	claude := fakeClaude(t, answer(`{"summary":"s","operations":[{"op":"add_entry","fields":{}}]}`), 0)
	require.Equal(t, 0, run([]string{"-shelf", t.TempDir(), "-claude", claude, path}), "exit code")
	assert.DeepEqual(t, "proposed", read(t, path)["status"])
}

func TestRunExitsWithOneWhenTheRequestFailed(t *testing.T) {
	path := setup(t, "pending")
	claude := fakeClaude(t, "not json", 1)
	require.Equal(t, 1, run([]string{"-shelf", t.TempDir(), "-claude", claude, path}), "exit code")
	failure, _ := read(t, path)["failure"].(string)
	assert.Inside(t, "did not return JSON (exit code 1)", failure)
}

func TestRunFailsTheRequestWhenClaudeIsMissing(t *testing.T) {
	path := setup(t, "pending")
	require.Equal(t, 1, run([]string{"-shelf", t.TempDir(), "-claude", filepath.Join(t.TempDir(), "no-claude"), path}), "exit code")
	failure, _ := read(t, path)["failure"].(string)
	assert.Inside(t, "Could not run Claude Code", failure)
}
