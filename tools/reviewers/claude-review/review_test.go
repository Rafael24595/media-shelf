package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Rafael24595/go-assert/assert"
	"github.com/Rafael24595/go-assert/require"
)

var fixedNow = time.Date(2026, 11, 16, 9, 30, 0, 0, time.FixedZone("CET", 3600))

// setup writes a request in <tmp>/data/.source/requests/ and returns its path.
func setup(t *testing.T, status string) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "data", ".source", "requests")
	require.Nil(t, os.MkdirAll(dir, 0o755))
	path := filepath.Join(dir, "REQ-20261116-093000.json")
	body := `{"contract":"media-shelf-request/1","id":"REQ-20261116-093000","kind":"catalog","status":"` + status +
		`","photos":["a.jpg"],"notifications":[{"notifier":"Claude","ok":true}],"custom":"kept"}`
	require.Nil(t, os.WriteFile(path, []byte(body), 0o644))
	return path
}

// call records how Claude Code was run.
type call struct {
	dir, name string
	args      []string
}

// fake answers like Claude Code would, and records the call.
func fake(stdout, stderr string, code int, err error, calls *[]call) runner {
	return func(_ context.Context, dir, name string, args ...string) (string, string, int, error) {
		*calls = append(*calls, call{dir, name, args})
		return stdout, stderr, code, err
	}
}

// answer is Claude Code's --output-format json envelope around a final message.
func answer(message string) string {
	out, _ := json.Marshal(map[string]any{"type": "result", "is_error": false, "result": message})
	return string(out)
}

func job(path string, run runner) reviewJob {
	return reviewJob{request: path, shelf: "/repos/media-shelf", claude: "claude", run: run, now: func() time.Time { return fixedNow }}
}

func read(t *testing.T, path string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	require.Nil(t, err)
	var m map[string]any
	require.Nil(t, json.Unmarshal(data, &m), "request is not JSON:\n%s", data)
	return m
}

func TestAProposalIsWrittenIntoTheRequest(t *testing.T) {
	path := setup(t, "pending")
	var calls []call
	msg := "Here it is:\n" + `{"summary":"One film.","operations":[{"op":"add_row","media":"movie","fields":{"edition_title":"Alien"}}]}`
	out, err := job(path, fake(answer(msg), "", 0, nil, &calls)).do(context.Background())
	require.Nil(t, err)
	assert.False(t, out.failed)
	assert.Equal(t, "REQ-20261116-093000: proposal with 1 operation(s).", out.message)

	r := read(t, path)
	assert.DeepEqual(t, "proposed", r["status"])
	assert.DeepEqual(t, reviewer, r["claimed_by"])
	assert.DeepEqual(t, "2026-11-16T09:30:00+01:00", r["claimed_at"])

	p := r["proposal"].(map[string]any)
	assert.DeepEqual(t, "One film.", p["summary"])
	assert.DeepEqual(t, reviewer, p["by"])
	assert.DeepEqual(t, "2026-11-16T09:30:00+01:00", p["at"])
	assert.Size(t, 1, p["operations"], "operations")

	assert.DeepEqual(t, "kept", r["custom"], "fields the adapter does not know are kept")
	assert.NotNil(t, r["notifications"], "fields the adapter does not know are kept")
}

func TestClaudeCodeIsRunReadOnlyWithTheDataRepository(t *testing.T) {
	path := setup(t, "pending")
	var calls []call
	_, err := job(path, fake(answer(`{"operations":[]}`), "", 0, nil, &calls)).do(context.Background())
	require.Nil(t, err)
	require.Size(t, 1, calls, "calls: %v", calls)

	c := calls[0]
	data := filepath.Dir(filepath.Dir(filepath.Dir(path)))
	wantPrompt := "Read " + filepath.Join("/repos/media-shelf", "tools", "reviewers", "review-prompt.md") + " and follow it for the request file " + path
	want := []string{"-p", wantPrompt, "--output-format", "json", "--permission-mode", "dontAsk",
		"--allowedTools", "Read,Glob,Grep,WebSearch,WebFetch,Agent", "--add-dir", data}
	assert.Equal(t, "/repos/media-shelf", c.dir)
	assert.Equal(t, "claude", c.name)
	assert.DeepEqual(t, want, c.args)

	for _, tool := range strings.Split(readOnlyTools, ",") {
		assert.NotInside(t, tool, []string{"Write", "Edit", "Bash", "NotebookEdit"}, "a tool that writes is allowed")
	}
}

func TestAnEmptyProposalIsStillAProposal(t *testing.T) {
	path := setup(t, "pending")
	var calls []call
	out, _ := job(path, fake(answer(`{"summary":"Nothing new.","operations":[]}`), "", 0, nil, &calls)).do(context.Background())
	assert.False(t, out.failed, "outcome %+v", out)
	assert.DeepEqual(t, "proposed", read(t, path)["status"])
}

func TestARequestThatIsNotPendingIsLeftAlone(t *testing.T) {
	for _, status := range []string{"in_progress", "proposed", "applied", "failed"} {
		path := setup(t, status)
		before, _ := os.ReadFile(path)
		var calls []call
		out, err := job(path, fake("", "", 0, nil, &calls)).do(context.Background())
		after, _ := os.ReadFile(path)
		assert.Nil(t, err, status)
		assert.False(t, out.failed, "%s: outcome %+v", status, out)
		assert.Size(t, 0, calls, "%s: Claude Code is not run", status)
		assert.Equal(t, string(before), string(after), "%s: the request is not changed", status)
		assert.Inside(t, "is "+status+": nothing to do", out.message, status)
	}
}

func TestFailuresAreWrittenIntoTheRequest(t *testing.T) {
	cases := []struct {
		name           string
		stdout, stderr string
		code           int
		err            error
		why            string
	}{
		{"not installed", "", "", -1, errors.New("executable file not found"), "Could not run Claude Code"},
		{"not JSON", "Please log in", "", 1, nil, "did not return JSON (exit code 1). Please log in"},
		{"reported error", `{"is_error":true,"result":"Credit balance is too low"}`, "", 1, nil, "reported an error (exit code 1): Credit balance is too low"},
		{"non-zero exit", answer(`{"operations":[]}`), "warning", 3, nil, "reported an error (exit code 3)"},
		{"no JSON in the answer", answer("I could not find the photos."), "", 0, nil, "holds no JSON proposal: I could not find the photos."},
		{"broken JSON", answer(`{"operations": [}`), "", 0, nil, "is not valid JSON"},
		{"no operations", answer(`{"summary":"x"}`), "", 0, nil, "has no operations list"},
		{"operations not a list", answer(`{"operations":{}}`), "", 0, nil, "has no operations list"},
		{"reviewer gave up", answer(`{"summary":"","operations":[],"failure":"photos unreadable"}`), "", 0, nil, "could not complete the request: photos unreadable"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			path := setup(t, "pending")
			var calls []call
			out, err := job(path, fake(c.stdout, c.stderr, c.code, c.err, &calls)).do(context.Background())
			require.Nil(t, err)
			r := read(t, path)
			failure, _ := r["failure"].(string)
			assert.True(t, out.failed, "outcome %+v", out)
			assert.DeepEqual(t, "failed", r["status"])
			assert.Inside(t, c.why, failure)
			assert.Nil(t, r["proposal"], "no proposal after failing")
			assert.DeepEqual(t, "kept", r["custom"], "fields the adapter does not know are kept")
		})
	}
}

func TestChangesMadeMeanwhileAreKept(t *testing.T) {
	path := setup(t, "pending")
	// the owner's tool adds a notification while Claude Code works
	run := func(_ context.Context, _, _ string, _ ...string) (string, string, int, error) {
		r, _ := readRequest(path)
		_ = r.set("notifications", []map[string]any{{"notifier": "Claude", "ok": true}, {"notifier": "Mail", "ok": false}})
		_ = writeRequest(path, r)
		return answer(`{"operations":[]}`), "", 0, nil
	}
	_, err := job(path, run).do(context.Background())
	require.Nil(t, err)
	assert.Size(t, 2, read(t, path)["notifications"], "notifications")
}

func TestAMissingRequestIsAnError(t *testing.T) {
	var calls []call
	_, err := job(filepath.Join(t.TempDir(), "nope.json"), fake("", "", 0, nil, &calls)).do(context.Background())
	assert.NotNil(t, err)
}

func TestTailKeepsTheEndWithoutCuttingCharacters(t *testing.T) {
	assert.Equal(t, "short", tail("  short  ", 10))
	assert.Equal(t, "...ghij", tail("abcdefghij", 4))
	// "ñ" is two bytes: a cut in its middle moves after it
	assert.Equal(t, "...yy", tail("xxñyy", 3))
}

func TestDataRepoIsTwoFoldersAboveRequests(t *testing.T) {
	path := filepath.Join("base", "data", ".source", "requests", "REQ-1.json")
	assert.Equal(t, filepath.Join("base", "data"), dataRepo(path))
}
