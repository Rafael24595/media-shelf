# ADR-0008: Separate data repository with a schema-version manifest

- **Status:** Accepted; partially superseded by ADR-0009 and ADR-0010
- **Date:** 2026-11-10

## Context

`media-shelf` holds two things that change at very different rates: the definition and
process (schema, ADRs, rules, agents), which changes rarely and through explicit decisions,
and the content (the CSV files and the review queue), which changes daily. Keeping both in one
repository mixes decisions with routine data edits in the history, and ties the data's
location to this project. The owner wants the data to have its own repository so that both
can evolve in parallel; the `media-shelf-ui` tool already takes the data folder and the schema
file from separate configuration settings (its ADR-0002).

## Decision

1. **Three repositories:**

   | Repository | Contents |
   |---|---|
   | `media-shelf` | Specification and process: `schema/schema.json`, `docs/csv-schema.md`, ADRs, `CLAUDE.md`, agents. |
   | `media-shelf-data` | Content only: the data CSV files and `review_queue.csv` at its root, `archive/` (frozen copies made by migrations), the manifest `shelf.json`, and the git-ignored photo inbox `.pics/`. |
   | `media-shelf-ui` | Tool: web interface and reference validator. |

2. **Manifest.** The data repository declares the schema version it conforms to in
   `shelf.json` at its root:

   ```json
   { "schema_version": "ADR-0008" }
   ```

   The file is required. The validator reports an error when it is missing, malformed, or
   when `schema_version` differs from `version` in `schema/schema.json`; tools must not write
   data while the versions differ.
3. **Schema changes become two coordinated steps:** (a) in `media-shelf`, the ADR and the new
   `schema.json`/`csv-schema.md`; (b) in `media-shelf-data`, the migration of the data and
   the new `schema_version` in `shelf.json`, in one commit. Between (a) and (b) the validator
   flags the data as needing migration.
4. **Photos** are raw input, not data: they live in `.pics/` inside the data repository folder
   and are never committed. `.pics/processed/` keeps the processed ones; review-queue entries
   refer to photos by file name.
5. Files inside the data repository are named by `schema.json` (`media_types.*.file`,
   `review_queue.file`, `manifest.file`); its location on disk is configuration of each tool.

## Consequences

- Each repository has a clean history: decisions in `media-shelf`, content in
  `media-shelf-data`, code in `media-shelf-ui`. The data repository can be private, backed up
  or synced on its own.
- Agents and skills work with two folders: the specification (read) and the data (write).
- `schema.json` gains a `manifest` section; its version becomes `ADR-0008`.

## Migration

1. Create `media-shelf-data` with the current `data/*.csv` at its root, `data/archive/` as
   `archive/`, `shelf.json` (`schema_version: ADR-0008`), a README, a `CLAUDE.md` pointing to
   the specification, and a `.gitignore` for `.pics/`.
2. Move `.pics/` (with `processed/`) from `media-shelf` to `media-shelf-data`.
3. Remove `data/` and `.pics/` from `media-shelf`; update its `CLAUDE.md`, `csv-schema.md`,
   agents and the ADR index; add `manifest` to `schema.json` and set `version` to `ADR-0008`.
4. `media-shelf-ui` checks the manifest (its ADR-0004); point its configuration to the new
   data folder and validate.
5. The data repository starts a new git history with the current state (the previous
   repository only had frozen archive copies of earlier data, which move to `archive/`).
