package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

var fixedNow = time.Date(2026, 11, 16, 9, 30, 0, 0, time.FixedZone("CET", 3600))

// setup writes a request in <tmp>/data/.source/requests/ and returns its path.
func setup(t *testing.T, status string) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "data", ".source", "requests")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "REQ-20261116-093000.json")
	body := `{"contract":"media-shelf-request/1","id":"REQ-20261116-093000","kind":"catalog","status":"` + status +
		`","photos":["a.jpg"],"notifications":[{"notifier":"Claude","ok":true}],"custom":"kept"}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
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
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("request is not JSON: %v\n%s", err, data)
	}
	return m
}

func TestAProposalIsWrittenIntoTheRequest(t *testing.T) {
	path := setup(t, "pending")
	var calls []call
	msg := "Here it is:\n" + `{"summary":"One film.","operations":[{"op":"add_row","media":"movie","fields":{"edition_title":"Alien"}}]}`
	out, err := job(path, fake(answer(msg), "", 0, nil, &calls)).do(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if out.failed || out.message != "REQ-20261116-093000: proposal with 1 operation(s)." {
		t.Errorf("outcome: %+v", out)
	}
	r := read(t, path)
	if r["status"] != "proposed" || r["claimed_by"] != reviewer || r["claimed_at"] != "2026-11-16T09:30:00+01:00" {
		t.Errorf("claim: %v %v %v", r["status"], r["claimed_by"], r["claimed_at"])
	}
	p := r["proposal"].(map[string]any)
	if p["summary"] != "One film." || p["by"] != reviewer || p["at"] != "2026-11-16T09:30:00+01:00" || len(p["operations"].([]any)) != 1 {
		t.Errorf("proposal: %v", p)
	}
	if r["custom"] != "kept" || r["notifications"] == nil {
		t.Errorf("fields the adapter does not know were lost: %v", r)
	}
}

func TestClaudeCodeIsRunReadOnlyWithTheDataRepository(t *testing.T) {
	path := setup(t, "pending")
	var calls []call
	if _, err := job(path, fake(answer(`{"operations":[]}`), "", 0, nil, &calls)).do(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 1 {
		t.Fatalf("calls: %v", calls)
	}
	c := calls[0]
	data := filepath.Dir(filepath.Dir(filepath.Dir(path)))
	wantPrompt := "Read " + filepath.Join("/repos/media-shelf", "tools", "reviewers", "review-prompt.md") + " and follow it for the request file " + path
	want := []string{"-p", wantPrompt, "--output-format", "json", "--permission-mode", "dontAsk",
		"--allowedTools", "Read,Glob,Grep,WebSearch,WebFetch,Agent", "--add-dir", data}
	if c.dir != "/repos/media-shelf" || c.name != "claude" || !slices.Equal(c.args, want) {
		t.Errorf("run in %s: %s %q\nwant %q", c.dir, c.name, c.args, want)
	}
	for _, tool := range strings.Split(readOnlyTools, ",") {
		if tool == "Write" || tool == "Edit" || tool == "Bash" || tool == "NotebookEdit" {
			t.Errorf("a tool that writes is allowed: %s", tool)
		}
	}
}

func TestAnEmptyProposalIsStillAProposal(t *testing.T) {
	path := setup(t, "pending")
	var calls []call
	out, _ := job(path, fake(answer(`{"summary":"Nothing new.","operations":[]}`), "", 0, nil, &calls)).do(context.Background())
	if out.failed || read(t, path)["status"] != "proposed" {
		t.Errorf("outcome %+v, status %v", out, read(t, path)["status"])
	}
}

func TestARequestThatIsNotPendingIsLeftAlone(t *testing.T) {
	for _, status := range []string{"in_progress", "proposed", "applied", "failed"} {
		path := setup(t, status)
		before, _ := os.ReadFile(path)
		var calls []call
		out, err := job(path, fake("", "", 0, nil, &calls)).do(context.Background())
		after, _ := os.ReadFile(path)
		if err != nil || out.failed || len(calls) != 0 || string(before) != string(after) {
			t.Errorf("%s: err %v, outcome %+v, calls %d, changed %v", status, err, out, len(calls), string(before) != string(after))
		}
		if !strings.Contains(out.message, "is "+status+": nothing to do") {
			t.Errorf("%s: message %q", status, out.message)
		}
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
			if err != nil {
				t.Fatal(err)
			}
			r := read(t, path)
			failure, _ := r["failure"].(string)
			if !out.failed || r["status"] != "failed" || !strings.Contains(failure, c.why) {
				t.Errorf("outcome %+v, status %v, failure %q; want it to contain %q", out, r["status"], failure, c.why)
			}
			if r["proposal"] != nil || r["custom"] != "kept" {
				t.Errorf("request after failing: %v", r)
			}
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
	if _, err := job(path, run).do(context.Background()); err != nil {
		t.Fatal(err)
	}
	if n := len(read(t, path)["notifications"].([]any)); n != 2 {
		t.Errorf("notifications: %d, want 2", n)
	}
}

func TestAMissingRequestIsAnError(t *testing.T) {
	var calls []call
	if _, err := job(filepath.Join(t.TempDir(), "nope.json"), fake("", "", 0, nil, &calls)).do(context.Background()); err == nil {
		t.Error("no error")
	}
}

func TestTailKeepsTheEndWithoutCuttingCharacters(t *testing.T) {
	if got := tail("  short  ", 10); got != "short" {
		t.Errorf("short: %q", got)
	}
	if got := tail("abcdefghij", 4); got != "...ghij" {
		t.Errorf("long: %q", got)
	}
	// "ñ" is two bytes: a cut in its middle moves after it
	if got := tail("xxñyy", 3); got != "...yy" {
		t.Errorf("multibyte: %q", got)
	}
}

func TestDataRepoIsTwoFoldersAboveRequests(t *testing.T) {
	path := filepath.Join("base", "data", ".source", "requests", "REQ-1.json")
	if got := dataRepo(path); got != filepath.Join("base", "data") {
		t.Errorf("dataRepo: %s", got)
	}
}
