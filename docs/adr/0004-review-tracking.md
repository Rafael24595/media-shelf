# ADR-0004: Review tracking (`reviewed` flag and review queue)

- **Status:** Accepted; photo locations superseded by ADR-0010
- **Date:** 2026-11-09

## Context

The first cataloging run (6 photos) produced 432 rows, but:

- most metadata was researched by agents and only spot-checked against sources, and some
  identifications are inferred (box-set contents, which version of a film). Nothing in the data
  says which rows the owner has confirmed against the physical item;
- many items could not be added at all — unreadable or cut-off spines, ambiguous versions,
  box sets with unknown contents, items needing a schema decision — and they are only listed in
  chat reports, where they get lost.

## Decision

### 1. `reviewed` column (all data files)

- New **required** boolean core column `reviewed`, placed after `notes` (last core column,
  before the type-specific columns).
- `true` = the owner has confirmed that the row's identification and data are correct for the
  owned copy. `false` = not confirmed yet (default for every new row).
- Only the owner sets it to `true` (directly or by telling Claude). Any later change to a row's
  identifying data (`original_title`, `edition_title`, `creator`, `release_date`, `formats`)
  that the owner did not dictate resets it to `false`.

### 2. Review queue: `data/review_queue.csv`

A file of things that need a manual check by the owner, with ids `REV-NNNN`.

| # | Field | Type | Required | Rules |
|---|---|---|---|---|
| 1 | `id` | text | Yes | `REV-NNNN`, sequential, never reused. |
| 2 | `status` | enum | Yes | `open`, `resolved`, `discarded`. |
| 3 | `reason` | enum | Yes | `unidentified` (cannot tell what it is), `ambiguous` (several candidates, e.g. original vs remake), `box_set_contents` (contents of a set not visible), `needs_check` (a row's data is inferred and must be confirmed on the physical item), `schema_gap` (does not fit the schema; needs an ADR), `needs_photo` (not visible enough; needs a new photo). |
| 4 | `media_type` | enum | No | `movie`, `video_game`, `book`, `series`; empty if unknown. |
| 5 | `description` | text | Yes | What is visible or known (spine text, label, distributor…). |
| 6 | `question` | text | Yes | What the owner must check or decide. |
| 7 | `candidates` | list | No | Possible identifications, `\|`-separated. |
| 8 | `photo` | list | No | Photo file name(s) where the item appears. |
| 9 | `location` | text | No | Position in the photo (shelf, row, neighbours). |
| 10 | `related_ids` | list | No | Ids of data rows affected (e.g. `MOV-0123`). Each must exist. |
| 11 | `created` | date | Yes | `YYYY-MM-DD`. |
| 12 | `resolved` | date | No | `YYYY-MM-DD`; required when `status` is not `open`. |
| 13 | `resolution` | text | No | What was decided / which rows were added or changed; required when `status` is not `open`. |
| 14 | `notes` | text | No | Free notes. |

Entries are never deleted: they are closed as `resolved` or `discarded`.

## Consequences

- The owner can work through `review_queue.csv` and mark rows `reviewed=true` as they check
  their shelves; progress is measurable (open entries, % of reviewed rows).
- Cataloging skills/agents must write every unidentified or uncertain item to the queue, and
  resolving an entry must update the related rows.
- A photo can move to `.pics/processed/` once all its items are either in the data files or in
  the review queue (the queue keeps the photo name).
- The validator checks the queue (enums, dates, `related_ids` exist, resolution when closed).

## Migration

1. Add `reviewed` after `notes` in `movies.csv`, `video_games.csv`, `books.csv`,
   `series.csv`, with value `false` for every existing row (archive the originals in
   `data/archive/` with a `.pre-ADR-0004` suffix).
2. Create `data/review_queue.csv` and fill it with the pending items of the first cataloging
   run (unidentified spines, ambiguous versions, box sets, inferred contents, schema gaps,
   cut-off shelf rows).
3. Validate.
