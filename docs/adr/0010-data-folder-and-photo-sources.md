# ADR-0010: Data folder and photo sources in the data repository

- **Status:** Accepted
- **Date:** 2026-11-10

## Context

In `media-shelf-data` the CSV files sat at the root next to the README, and every photo went
to one inbox (`.pics/`, processed ones to `.pics/processed/`). Two needs appeared:

- Keep the dataset apart from the repository's other files, so a tool's data folder holds
  only data.
- Photos taken to answer review-queue entries (the back of a box set, a closer shot of a
  spine) are not new items and must not go through cataloging; they need their own place and
  a way to find the entry they answer.

## Decision

1. **Layout of `media-shelf-data`:**

   ```text
   README.md            the repository's only description (ADR-0009)
   .gitignore           ignores .source/
   data/                the dataset: the five CSV files and shelf.json
   .source/             photos, never committed
     pending/           new photos to catalog; processed/ once cataloged
     review/            photos that answer review-queue entries; processed/ once their entries are closed
   ```

   `data/` is what tools call the *data folder* (`media-shelf-ui`'s `data_dir`). The manifest
   `shelf.json` moves into it, so the data folder is self-contained: files plus the schema
   version they follow. File names inside it are still those of `schema.json`.

2. **Photo references.** The review queue's `photo` column holds photo *file names* (not
   paths), which are unique across `.source/` (camera names such as `PXL_20261110_123754895.jpg`).
   A file is found by name in any folder under `.source/`. Photos added to answer an entry are
   appended to that entry's `photo`.

3. **Linking review photos to entries.** A photo in `.source/review/` whose name starts with
   an entry id (`REV-0001.jpg`, `REV-0001_back.jpg`) belongs to that entry. Otherwise the
   `media-reviewer` agent proposes the entry from the photo's content and the owner confirms
   before anything is written. A review photo moves to `.source/review/processed/` when every
   entry it belongs to is closed.

4. **Review work.** A new read-only agent, `media-reviewer`, takes one open entry (with its
   photos, old and new, and any answer from the owner), researches it and proposes the
   resolution: rows to create or change, entry status and resolution text. The main
   conversation writes. Entries that only need the owner's answer (seen or not, which
   edition…) are resolved directly, without the agent (skill `media-shelf-manage`).

5. No column, vocabulary or rule of the CSV files changes, so `schema_version` stays
   `ADR-0008`; only descriptions of locations change.

This partially supersedes ADR-0008 and ADR-0009 (repository contents) and ADR-0004 (where
photos are kept).

## Alternatives considered

- *Keep `shelf.json` at the repository root and have tools look one folder up.* An implicit
  rule that surprises; rejected.
- *A new column for review photos.* `photo` already holds the photos of an entry; a second
  list would split the same information.
- *Only ids in file names, or only matching by content.* Ids in names are reliable but tedious
  to type; matching by content is convenient but makes mistakes with similar boxes. Accepting
  both, with confirmation for the second, keeps each where it is good.

## Consequences

- Tools point their data folder at `media-shelf-data/data` (configuration change, no code
  change).
- Cataloging reads `.source/pending/`; review work reads `.source/review/`; the two flows no
  longer mix.
- Existing `photo` values keep working: those photos are now in `.source/pending/processed/`.

## Migration

1. Move the CSV files into `data/` and `shelf.json` into `data/` (done by the owner and in
   this change); `.pics/` becomes `.source/pending/`; create `.source/review/`;
   `.gitignore` ignores `.source/`.
2. Point `media-shelf-ui`'s `data_dir` to `media-shelf-data/data`.
3. Update the data README, `CLAUDE.md`, `docs/csv-schema.md`, agents and skills.
