# media-shelf

Personal inventory of the multimedia items the owner has: movies, video games, books and
TV series. This repository holds the **specification and the process**: the data definition,
its decisions (ADRs), the rules below and the agents. The data itself lives in a separate
repository, **`media-shelf-data`** (ADR-0008), described by its own README; the tooling in
**`media-shelf-ui`**. [`README.md`](README.md) is the overview for people (what each
repository holds, how to build the tools); this file holds the rules for working here.

The three repositories sit in one common folder on the owner's computer, but no tool assumes
it: locations are configuration.

## Layout of this repository

| Path | Purpose |
|---|---|
| `schema/schema.json` | **Machine-readable definition** of the data files; every tool reads it. |
| `docs/csv-schema.md` | Human-readable definition (same rules as `schema.json`). |
| `docs/adr/` | Architecture Decision Records for every change to that definition. |
| `docs/review-requests.md` | Request/proposal contract between the owner's tools and reviewers (ADR-0011). |
| `tools/reviewers/` | Optional reviewer adapters a media-shelf-ui notifier can run (e.g. `claude-review/`: a Go program, ADR-0012, that answers a request with a proposal using Claude Code read-only). |
| `README.md` | Overview for people: the three repositories, the definition, the agents, building claude-review. |
| `agents/` | Source of the project agents (`media-cataloger`, `media-auditor`, `media-reviewer`). |
| `scripts/migrations/` | Data migrations, one Go program per schema change (`adr_0013_actors/`, ADR-0013); run once, kept as history. The earlier Python ones (ADR-0003 to ADR-0005) and the retired Python validator (ADR-0006) were removed. |

## Rules

1. **English only for documentation.** Every document in these repositories — README, schema,
   ADRs, skills, agents, code comments, commit messages — is written in English, even when the
   conversation is in another language.
2. **No schema change without an ADR.** Any change to the definition of the data files
   (files, columns, names, order, types, required flags, vocabularies, scales, conventions,
   manifest) needs an ADR in `docs/adr/`, accepted by the owner, *before* it is applied. Follow
   the process in `docs/adr/README.md` (skill: `media-shelf-schema-change`). If an item does not
   fit the current schema, propose an ADR — never improvise new files, columns or values.
3. **Two-step schema changes.** (a) Here: the ADR, `schema.json` (new `version`) and
   `csv-schema.md`. (b) In `media-shelf-data`: migrate the data and set `shelf.json`'s
   `schema_version` to the new version, in one commit (the previous commit is the pre-migration
   state; there are no archive copies, ADR-0009). Until (b), the validator flags the data
   and the UI refuses to write. Keep `media-shelf-data/README.md` accurate: it is that
   repository's only documentation.
4. **The schema wins.** Read `docs/csv-schema.md` (and `schema/schema.json` for exact
   vocabularies) before writing data. If they, the validator and the data disagree, stop and
   report it.
5. **Right file, right id.** Each item goes to the file of its media type, with the next id of
   that file (`MOV-0714`, `GAM-0056`…). Ids are never reused or renumbered.
6. **One row per title.** Several editions/formats of the same title share one row. Before
   adding, look for the title in its file (by `original_title` and by `edition_title`).
7. **Canonical values only.** `original_title` in the original language (romanized for
   non-Latin scripts), platforms from the vocabulary, canonical tags and genres
   (`schema/schema.json`; see `docs/csv-schema.md`). If a new value is needed, extend the
   vocabulary through an ADR.
8. **Never invent data.** If a value cannot be confirmed (from the photo, the owner or a reliable
   source), leave it empty and flag it to the owner. Synopses are written in our own words.
9. **Personal fields come from the owner.** `consumed`, `rating`, `condition` and
   `missing_items` are only filled from what the owner says or what is clearly visible in a
   photo. `consumed` defaults to `false` when unknown.
10. **`reviewed` belongs to the owner.** New rows always get `reviewed=false`. Only the owner
    sets `true`. Changing `original_title`, `edition_title`, `creator`, `release_date` or
    `formats` without the owner dictating it resets the row to `false`.
11. **Nothing uncertain gets lost.** Every item that cannot be added with confidence
    (unidentified, ambiguous, box set with unknown contents, inferred data, schema gap,
    not visible enough) gets an entry in the review queue. Resolving an entry updates the
    related rows and closes the entry (`resolved`/`discarded` + date + resolution); entries are
    never deleted.
12. **Write CSVs through the web UI/API or a proper CSV library** (Go's `encoding/csv`), never
    by hand-typing rows, and never rewrite rows that are not being changed. No new Python
    scripts (ADR-0006).
13. **Validate after every change** with `media-shelf-ui validate`, passing both locations, e.g.
    from `media-shelf-data`:
    `<media-shelf-ui>/dist/media-shelf-ui-linux-amd64 validate -data data -schema <media-shelf>/schema/schema.json`
    (Linux) or `media-shelf-ui.exe validate -data data -schema <media-shelf>\schema\schema.json`
    (Windows); with Go: `go run ./cmd/media-shelf-ui validate -data <data> -schema <schema.json>`
    from the `media-shelf-ui` source. It must report no errors.
14. **Working alongside the web UI is safe:** it refuses a write when the file changed on disk
    since it was loaded. After bulk changes made outside the UI, reload it.
15. **Photos** (ADR-0010) live in `media-shelf-data/.source/` and are never committed.
    - `.source/pending/`: new items to catalog. Move a photo to `.source/pending/processed/`
      once every item in it is either in a data file or in the review queue (the queue keeps
      the photo name). Several photos may show the same shelf: deduplicate across photos
      before writing.
    - `.source/review/`: photos that answer review-queue entries. A name starting with an
      entry id (`REV-0001_back.jpg`) binds it to that entry; otherwise `media-reviewer`
      proposes the entry and the owner confirms. Append the photo name to the entry's
      `photo`; move the photo to `.source/review/processed/` when all its entries are closed
      (media-shelf-ui does both when entries are edited or closed from its queue screen, if
      its `photos_dir` is configured).
    - `photo` holds file names, not paths; names are unique across `.source/`.
16. **Propose, do not write** (ADR-0011). Cataloging and review results go to the owner as a
    proposal: a request file in `media-shelf-data/.source/requests/` following
    `docs/review-requests.md` (create the request yourself, `created_by` = your name, when the
    owner asked in the conversation; or take a `pending` one, set `in_progress`, then
    `proposed`). Each operation carries the values you saw (`from`), its reason and its
    evidence; use `${ref}` to refer to rows added earlier in the same proposal. The owner
    applies it in media-shelf-ui, which validates every operation. Write the data files
    directly only when the owner explicitly asks for a direct change in that conversation,
    then validate (rule 13). Before writing a proposal, check the data with the validator and
    make sure every value follows the schema, so operations do not fail when applied.

## Skills

| Skill | Use it to |
|---|---|
| `media-shelf-add-from-images` | Add new items from the photos in `media-shelf-data/.source/pending/`. |
| `media-shelf-manage` | Update existing rows, mark rows as reviewed, work through the review queue (including the photos in `.source/review/`), answer questions about the collection, validate the data. |
| `media-shelf-schema-change` | Propose and apply a change to the data definition through an ADR. |

## Agents

Source files live in `agents/`; copy them to `.claude/agents/` so Claude Code loads them
(re-copy after editing). All are read-only: **only the main conversation writes data**, so
ids and rows never conflict, and by default the main conversation does not write data
either: it writes a proposal (rule 16).

| Agent | Role |
|---|---|
| `media-cataloger` | Identifies the items in one photo (or a small batch) and researches their metadata. Returns proposed rows grouped by media type, with evidence and flags. Run one per photo, in parallel. |
| `media-auditor` | Independently verifies proposed rows (facts, photo, schema, duplicates) before they are written. Run once per batch (split large batches, ~90 rows each). |
| `media-reviewer` | Works one open review-queue entry with its photos (including new ones in `.source/review/`) and the owner's answers; researches it and proposes the resolution (rows to create or change, how to close the entry). Can also match unnamed review photos to entries. Run one per entry, in parallel. |

Flow for new photos: `media-cataloger` × N → main conversation deduplicates across photos →
`media-auditor` × M → main conversation writes a proposal with the approved rows and the
review-queue entries (rule 16); applying it in media-shelf-ui moves the photos to
`.source/pending/processed/`.

Flow for review photos: main conversation binds each photo in `.source/review/` to its entry
(by name, or `media-reviewer` proposes and the owner confirms) → `media-reviewer` × entries →
rows that come out of a resolution go through `media-auditor` → main conversation writes a
proposal (new rows, changed rows, entries closed, photos added to entries); applying it in
media-shelf-ui moves photos whose entries are all closed to `.source/review/processed/`.
Entries that only need the owner's answer go in a proposal too, unless the owner asks for a
direct change.

Queries and ADRs are handled directly in the main conversation; single-row updates too when
the owner asks for a direct change.
