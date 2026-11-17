# ADR-0002: Split `title` into `original_title` and `edition_title`

- **Status:** Accepted
- **Date:** 2026-11-09

## Context

ADR-0001 defines a single `title` column. Owned copies are often localized editions
(e.g. a Spanish edition of an English book, or a western release of a Japanese game), so a
single title either loses the original work's name or does not match what is printed on the
physical copy, which makes identifying items from photos and searching harder.

## Decision

- Replace `title` with two required columns, in this position:
  `id, original_title, edition_title, synopsis, …`
  - `original_title`: title of the original release in its original language; for non-Latin
    scripts, the official Latin-script title.
  - `edition_title`: title exactly as printed on the owned copy. Equal to `original_title`
    when they match.
- Free-text data (`synopsis`, `genre`, `tags`, `notes`) is written in English.
- `release_date` refers to the original first release, not the owned edition.

## Consequences

- Items can be found by either title.
- Each new row requires researching the original title when the copy is a localized edition.

## Migration

For each existing row: copy `title` into both `original_title` and `edition_title`, then
correct `original_title` where the copy is a localized edition. (No rows existed when this
ADR was accepted.)
