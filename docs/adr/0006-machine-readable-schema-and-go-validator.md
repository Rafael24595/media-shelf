# ADR-0006: Machine-readable schema and Go validator

- **Status:** Accepted
- **Date:** 2026-11-10

## Context

The data definition was written three times: for people (`docs/csv-schema.md`) and for Python
(`scripts/validate.py`, `scripts/conventions.py`). A new tool, the `media-track-ui` web
interface (a separate project, Go standard library only, see its ADR-0001), also needs the
definition: columns, vocabularies, synonyms, id prefixes. Copying it into a fourth place would
let the copies drift apart. The owner also wants the project free of Python.

## Decision

1. **`schema/schema.json`** is the machine-readable definition of every data file: file
   conventions, core and type-specific columns (type, required, vocabulary), formats per
   media type, controlled vocabularies (platforms, missing items, condition scale, series
   types), genre and tag synonyms, forbidden tags, identifying fields and the review-queue
   format. Tools read it; none hard-codes the data model.
2. **`docs/csv-schema.md`** remains the human-readable specification and states the same
   rules. Both are changed together, through an ADR.
3. **The reference validator is `media-track-ui validate`** (project `media-track-ui`, next to
   this one). It reads `schema/schema.json` and implements the same rules and messages as the
   former Python validator. Parity was verified on 2026-11-10: identical results on the real
   data (769 rows, 0 issues) and on a deliberately broken copy (44 identical errors and
   warnings).
4. **`scripts/validate.py` and `scripts/conventions.py` are retired** (moved to
   `scripts/retired/`, kept only as history). The migration scripts in `scripts/migrations/`
   are historical records of ADR-0003 to ADR-0005 and are not run again.
5. Future migrations do not use Python.

## Consequences

- One definition feeds every tool; vocabulary or column changes need no code changes in the
  UI or the validator.
- Validating requires the `media-track-ui` executable (`dist/media-track-ui.exe` on Windows,
  `dist/media-track-ui-linux-amd64` on Linux) or `go run` from its source.
- New cross-field rules are implemented in Go, with tests, in `media-track-ui`.

## Migration

1. Create `schema/schema.json` from the current definition (ADR-0005).
2. Move `scripts/validate.py` and `scripts/conventions.py` to `scripts/retired/`.
3. Update `CLAUDE.md`, `docs/csv-schema.md`, the agents and the skills to the new validator.
4. Validate with `media-track-ui validate`.
