// Command claude-review is a reviewer adapter (docs/review-requests.md): it answers one
// media-shelf review request with Claude Code, running read-only, and writes the proposal
// into the request for the owner to approve in media-shelf-ui (ADR-0012).
//
//	claude-review -shelf <media-shelf> [-claude <command>] <request file>
//
// Configure it as a media-shelf-ui notifier: the command with "{request}" as last argument.
// It needs Claude Code installed and signed in; each run uses the owner's Claude plan.
package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"time"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

// run returns the exit code: 0 when the request was answered or had nothing to do, 1 when
// it failed (the reason is in the request), 2 when it could not be read or written.
func run(args []string) int {
	flags := flag.NewFlagSet("claude-review", flag.ContinueOnError)
	shelf := flags.String("shelf", "", "the media-shelf repository (rules, schema and review-prompt.md)")
	claude := flags.String("claude", "claude", "the Claude Code command")
	flags.Usage = func() {
		fmt.Fprintln(flags.Output(), "usage: claude-review -shelf <media-shelf> [-claude <command>] <request file>")
		flags.PrintDefaults()
	}
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *shelf == "" || flags.NArg() != 1 {
		flags.Usage()
		return 2
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	job := reviewJob{request: flags.Arg(0), shelf: *shelf, claude: *claude, run: execute, now: time.Now}
	out, err := job.do(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	fmt.Println(out.message)
	if out.failed {
		return 1
	}
	return 0
}

// execute runs a program and collects what it writes. A non-zero exit is not an error here:
// the caller reads Claude Code's answer and its code.
func execute(ctx context.Context, dir, name string, args ...string) (string, string, int, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return stdout.String(), stderr.String(), exit.ExitCode(), nil
	}
	if err != nil {
		return "", "", -1, err
	}
	return stdout.String(), stderr.String(), 0, nil
}
