# ADR-0014: go-assert in the tests of the Go programs

- **Status:** Accepted
- **Date:** 2026-11-18
- **Amends:** ADR-0012, decision 1 (claude-review uses only the standard library)

## Context

ADR-0012 made claude-review a Go module that uses only the standard library, tests included.
Its tests check every result by hand (`if got != want { t.Errorf(…) }`), composing each
failure message themselves. `media-shelf-ui`, the other Go program of the owner's tools,
writes its tests with the owner's assertion library, `github.com/Rafael24595/go-assert`
(its ADR-0010): shorter checks whose failures say what was expected and what came.

## Decision

1. **Tests of the Go programs in this repository use go-assert** (`assert` for checks that
   let the test go on, `require` for those it cannot go on without). It is a test-only
   dependency: the programs themselves still use only the standard library, as ADR-0012
   decided.
2. **claude-review's tests are rewritten with it**, keeping exactly what they check.
3. **Migrations keep their tests as they are.** A migration runs once and is kept as history
   (`scripts/migrations/`); `adr_0013_actors` is not rewritten. New migrations use go-assert.
4. The version is pinned in each module's `go.mod` and `go.sum`, as published (no `replace`).

## Consequences

- One way of writing Go tests across media-shelf and media-shelf-ui.
- The first `go test` of claude-review (and possibly the first build, which reads the
  requirements of `go.mod`) downloads go-assert; the executable does not contain it.
- go-assert is maintained by the owner: a breaking change there is updated here on purpose,
  never pulled in by accident, since the version is pinned.

## Migration

None: no data changes.
