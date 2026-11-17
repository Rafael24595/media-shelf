# ADR-0003: One CSV file per media type

- **Status:** Accepted
- **Date:** 2026-11-09

## Context

ADR-0001 and ADR-0002 keep every item in a single `data/inventory.csv` with a `category`
column. After the first cataloging run (432 items: 417 movies, 15 video games) two problems
appeared:

- Each media type needs its own fields (platform for games; ISBN, publisher and pages for
  books; seasons for series). In a single file every such field is an empty column for all
  other types.
- TV series and miniseries found on the shelves do not fit any `category`, and they need
  fields of their own.

Per-type rules (formats vocabulary, applicable `missing_items`) were already conditioned on
`category`, which is a sign that the types have different schemas.

## Decision

1. Replace `data/inventory.csv` with one file per media type:

   | File | Media type | `id` prefix |
   |---|---|---|
   | `data/movies.csv` | Movies | `MOV-` |
   | `data/video_games.csv` | Video games | `GAM-` |
   | `data/books.csv` | Books | `BOK-` |
   | `data/series.csv` | TV series and miniseries | `SER-` |

2. **`category` is removed**: the file determines the media type.
3. **`id`** becomes `<PREFIX>-<NNNN>` (four digits, zero-padded, sequential per file, never
   reused), e.g. `MOV-0001`. Ids are unique across the whole collection.
4. **Common core columns**, identical in name, meaning and order at the start of every file:
   `id, original_title, edition_title, synopsis, genre, creator, formats, release_date,
   consumed, condition, missing_items, rating, tags, notes`.
5. **Type-specific columns**, appended after the core (all optional):
   - `movies.csv`: none.
   - `video_games.csv`: `platform` (moved from the core; required for video games).
   - `books.csv`: `isbn`, `publisher`, `pages`.
   - `series.csv`: `series_type` (`series` | `miniseries`), `seasons` (owned season numbers,
     `|`-separated, e.g. `1|2`; empty for miniseries).
6. `creator` means: director (movies), developer studio (video games), author (books),
   creator/showrunner (series).
7. `formats` vocabulary and `missing_items` applicability are defined per file. Series use
   the movie vocabularies.
8. All files follow the existing conventions (UTF-8 without BOM, comma, standard quoting,
   `|` for multi-value fields). Every file exists with at least its header, even when empty.

## Consequences

- New per-type fields can be added without touching the other files (still via ADR).
- Series can be cataloged (pending items: *The Walking Dead* season 1, *Hitler: The Rise of
  Evil*, *The Triangle*, *Les Misérables* 2000, possibly *Vicente Ferrer*).
- Collection-wide queries must read several files; they rely on the common core columns.
- `scripts/validate.py` validates every file and checks id uniqueness across files.

## Migration

1. Copy `data/inventory.csv` to `data/archive/inventory.pre-ADR-0003.csv`.
2. Rows with `category=movie` → `data/movies.csv`; rows with `category=video_game` →
   `data/video_games.csv`, keeping their original order.
3. Drop `category`; move `platform` to the end (games only); renumber ids per file as
   `MOV-0001…`, `GAM-0001…` in that order.
4. Create `data/books.csv` and `data/series.csv` with headers only.
5. Remove `data/inventory.csv` from the active data set (moved into the archive).
6. Validate: total rows must equal the original 432 (417 + 15).
