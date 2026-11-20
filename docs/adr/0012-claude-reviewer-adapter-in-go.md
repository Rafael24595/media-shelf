# ADR-0012: Claude reviewer adapter in Go

- **Status:** Accepted (partially superseded by ADR-0014)
- **Date:** 2026-11-16

## Context

The Claude Code reviewer adapter (`tools/reviewers/claude-review.ps1`, ADR-0011) runs only
with Windows PowerShell: its paths use backslashes, the prompt carries a Windows path, and
the notifier calls `powershell`. On Linux it would need PowerShell 7 installed and changes to
work, and it has no tests. The owner's tools are written in Go (media-shelf-ui), and the
owner wants nothing tied to one operating system.

## Decision

1. **The adapter is a Go program**, `tools/reviewers/claude-review/`, a module of its own
   that uses only the standard library. It does what the PowerShell script did: marks the
   request `in_progress`, runs Claude Code non-interactively with read-only tools in this
   repository and the data repository readable, and writes the proposal (`proposed`) or the
   reason it could not (`failed`), keeping every other field of the request.
2. **Locations are arguments, not guesses**: `-shelf <media-shelf>` (rules, schema and
   `review-prompt.md`) and the request file; the data repository is the one holding the
   request (`<data>/.source/requests/`). `-claude` names the Claude Code command when it is
   not `claude`.
3. **Executables are built, not versioned**: `go build` into `claude-review/dist/`
   (git-ignored), for Windows and Linux; the notifier runs the one of its system.
4. **`claude-review.ps1` is removed**; `review-prompt.md` stays, shared by any adapter.
5. Running Claude Code and the clock are injected, so the adapter is tested without Claude
   Code (a fake that answers like it, and a fake command for the program as a whole).

## Consequences

- The same adapter, with the same behaviour, on Windows and Linux; notifier:
  `["<…>/claude-review/dist/claude-review.exe", "-shelf", "<media-shelf>", "{request}"]`
  (Windows) or `…/claude-review-linux-amd64` (Linux).
- Building it needs Go once per machine (or copying a built executable); running it needs
  only Claude Code.
- This repository holds Go code for the first time, limited to `tools/`; it does not change
  the data definition.
- As before, if the notifier's timeout stops the adapter, the request stays `in_progress`.

## Migration

None: no data changes. Notifiers that ran `claude-review.ps1` are changed to run the
executable.
