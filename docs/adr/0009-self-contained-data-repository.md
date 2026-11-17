# ADR-0009: Self-contained data repository without archive copies

- **Status:** Accepted; partially superseded by ADR-0010
- **Date:** 2026-11-10

## Context

ADR-0008 gave `media-shelf-data` an `archive/` folder (frozen copies made by migrations) and a
`CLAUDE.md` pointing to this repository. `archive/` was introduced when the data was not under
version control; now every migration is a commit in the data repository, so the previous
state is already kept by git and the copies only duplicate it, add weight and are one more
folder that must not be touched. The owner also wants the data repository decoupled from any
particular tool: a plain data store described by its own README.

## Decision

- `media-shelf-data` contains only the data files, `shelf.json`, `README.md`, `.gitignore` and
  the git-ignored photo inbox `.pics/`.
- **No `archive/`.** The state before a migration is the previous commit of the data
  repository; a migration is one commit with the transformed files and the new
  `schema_version` in `shelf.json`.
- **No tool-specific files** (no `CLAUDE.md`). `README.md` is the only description of the
  repository and must state everything needed to change the data safely: what each file is,
  the manifest, where the definition lives, how to validate, and the basic rules (ids, one row
  per title, review queue, photos).
- This partially supersedes ADR-0008 (repository contents and migration step).

## Consequences

- Smaller repository, nothing duplicated; diffs of a migration show exactly what changed.
- Tools and agents learn the rules from `media-shelf` (they read its `CLAUDE.md` and schema)
  or from the data repository's README; opening a coding assistant in the data folder alone no
  longer loads the rules automatically.

## Migration

1. Delete `archive/` and `CLAUDE.md` from `media-shelf-data`; expand its `README.md`.
2. Update `CLAUDE.md`, `docs/csv-schema.md`, the ADR index and the skills in `media-shelf`.
