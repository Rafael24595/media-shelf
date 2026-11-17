# ADR-0001: Initial CSV schema

- **Status:** Accepted
- **Date:** 2026-11-09

## Context

The owner wants to keep a record of every multimedia item they own — movies, video games
and books — in a single CSV file that is easy to edit by hand, to process with scripts and
to fill in from photos of the items.

## Decision

- The inventory lives in `data/inventory.csv`: UTF-8, comma-delimited, standard CSV quoting,
  `|` as the separator for multi-value fields.
- One row per title. Several editions/formats of the same title share one row with several
  `formats`.
- Columns, in order:
  `id, title, synopsis, category, genre, creator, platform, formats, release_date, consumed,
  condition, missing_items, rating, tags, notes`.
- `category` is one of `movie`, `video_game`, `book`.
- `consumed` is a boolean flag (`true` / `false`) marking whether the item has been
  watched/played/read or is pending.
- Physical state is described by two complementary fields:
  - `condition`: wear on a 1–5 scale (5 = new/sealed … 1 = poor);
  - `missing_items`: controlled vocabulary of missing pieces (`box`, `manual`, `cover_art`,
    `disc`, `extra_disc`, `slipcover`, `inserts`, `code`, `dust_jacket`, `pages`, `other`).
    A percentage or a single scale was rejected because it cannot express *what* is missing
    (e.g. "no manual", "no box").
- `rating` is a personal 1–10 score; `release_date` is ISO 8601 (`YYYY-MM-DD` or `YYYY`).
- A controlled `formats` vocabulary per category.

## Consequences

- Items can be filtered by status (`consumed`), completeness (`missing_items`) and wear
  (`condition`) independently.
- Controlled vocabularies keep data consistent but require an ADR to extend.

## Migration

None (initial version).
