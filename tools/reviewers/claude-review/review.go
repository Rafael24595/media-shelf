package main

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"
)

// reviewer is the name written in claimed_by and proposal.by.
const reviewer = "Claude Code"

// tools Claude Code may use: reading, searching the web and agents; none writes files.
const readOnlyTools = "Read,Glob,Grep,WebSearch,WebFetch,Agent"

// runner runs a program in dir and returns what it wrote and its exit code; err is set only
// when it could not be started.
type runner func(ctx context.Context, dir, name string, args ...string) (stdout, stderr string, code int, err error)

// reviewJob answers one request: where things are, how to run Claude Code and what time it is.
type reviewJob struct {
	request string // the request file
	shelf   string // the media-shelf repository (rules, schema, prompt)
	claude  string // the Claude Code command
	run     runner
	now     func() time.Time
}

// outcome is what the job did, for the notifier's log.
type outcome struct {
	message string
	failed  bool
}

// dataRepo is the data repository of a request: <data>/.source/requests/REQ-….json.
func dataRepo(requestPath string) string {
	return filepath.Dir(filepath.Dir(filepath.Dir(requestPath)))
}

// prompt is the instruction given to Claude Code: the steps live in a file of this
// repository, so the command line only carries two paths.
func prompt(shelf, requestPath string) string {
	return fmt.Sprintf("Read %s and follow it for the request file %s",
		filepath.Join(shelf, "tools", "reviewers", "review-prompt.md"), requestPath)
}

// claudeArgs runs Claude Code once, non-interactively, with read-only tools and the data
// repository readable.
func claudeArgs(shelf, requestPath string) []string {
	return []string{
		"-p", prompt(shelf, requestPath),
		"--output-format", "json",
		"--permission-mode", "dontAsk",
		"--allowedTools", readOnlyTools,
		"--add-dir", dataRepo(requestPath),
	}
}

// do claims a pending request, asks Claude Code for a proposal and writes it into the
// request (proposed), or writes why it could not (failed). A request that is not pending is
// left alone.
func (j reviewJob) do(ctx context.Context) (outcome, error) {
	r, err := readRequest(j.request)
	if err != nil {
		return outcome{}, fmt.Errorf("cannot read the request: %w", err)
	}
	id := r.text("id")
	if status := r.text("status"); status != "pending" {
		return outcome{message: fmt.Sprintf("%s is %s: nothing to do.", id, status)}, nil
	}
	if err := j.update(func(r request) error {
		return setAll(r, "status", "in_progress", "claimed_by", reviewer, "claimed_at", j.stamp())
	}); err != nil {
		return outcome{}, err
	}

	proposal, why := j.ask(ctx)
	if why != "" {
		if err := j.update(func(r request) error { return setAll(r, "status", "failed", "failure", why) }); err != nil {
			return outcome{}, err
		}
		return outcome{message: fmt.Sprintf("%s: failed. %s", id, why), failed: true}, nil
	}
	if err := setAll(proposal, "by", reviewer, "at", j.stamp()); err != nil {
		return outcome{}, err
	}
	if err := j.update(func(r request) error {
		if err := r.set("proposal", proposal); err != nil {
			return err
		}
		return r.set("status", "proposed")
	}); err != nil {
		return outcome{}, err
	}
	return outcome{message: fmt.Sprintf("%s: proposal with %d operation(s).", id, operations(proposal))}, nil
}

// update reads the request again, changes it and writes it: the owner's tool may have
// changed other fields meanwhile.
func (j reviewJob) update(change func(request) error) error {
	r, err := readRequest(j.request)
	if err != nil {
		return fmt.Errorf("cannot read the request: %w", err)
	}
	if err := change(r); err != nil {
		return err
	}
	if err := writeRequest(j.request, r); err != nil {
		return fmt.Errorf("cannot write the request: %w", err)
	}
	return nil
}

// ask runs Claude Code and returns its proposal, or why there is none.
func (j reviewJob) ask(ctx context.Context) (request, string) {
	stdout, stderr, code, err := j.run(ctx, j.shelf, j.claude, claudeArgs(j.shelf, j.request)...)
	if err != nil {
		return nil, "Could not run Claude Code (is the claude command installed and signed in?): " + err.Error()
	}
	var res struct {
		IsError bool   `json:"is_error"`
		Result  string `json:"result"`
	}
	if json.Unmarshal([]byte(stdout), &res) != nil {
		return nil, fmt.Sprintf("Claude Code did not return JSON (exit code %d). %s", code, tail(stdout+" "+stderr, 600))
	}
	if res.IsError || code != 0 {
		return nil, fmt.Sprintf("Claude Code reported an error (exit code %d): %s", code, tail(res.Result+" "+stderr, 600))
	}
	return parseProposal(res.Result)
}

// parseProposal takes the JSON object of the answer (from its first { to its last }) and
// checks it is a proposal: it has an operations list and no failure.
func parseProposal(text string) (request, string) {
	start, end := strings.Index(text, "{"), strings.LastIndex(text, "}")
	if start < 0 || end <= start {
		return nil, "The answer holds no JSON proposal: " + tail(text, 600)
	}
	p, err := parseRequest([]byte(text[start : end+1]))
	if err != nil {
		return nil, "The proposal is not valid JSON: " + err.Error() + " " + tail(text, 400)
	}
	var list []json.RawMessage
	if p["operations"] == nil || json.Unmarshal(p["operations"], &list) != nil {
		return nil, "The proposal has no operations list: " + tail(text, 400)
	}
	if failure := p.text("failure"); failure != "" {
		return nil, "Claude Code could not complete the request: " + failure
	}
	return p, ""
}

func operations(p request) int {
	var list []json.RawMessage
	_ = json.Unmarshal(p["operations"], &list)
	return len(list)
}

// setAll sets text fields given as name, value pairs.
func setAll(r request, pairs ...string) error {
	for i := 0; i+1 < len(pairs); i += 2 {
		if err := r.set(pairs[i], pairs[i+1]); err != nil {
			return err
		}
	}
	return nil
}

// stamp is the current time as the contract writes it (RFC 3339, local offset).
func (j reviewJob) stamp() string { return j.now().Format(time.RFC3339) }

// tail keeps the end of a long message, where errors usually are, without cutting a
// character in half.
func tail(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	cut := len(s) - n
	for cut < len(s) && !utf8.RuneStart(s[cut]) {
		cut++
	}
	return "..." + s[cut:]
}
