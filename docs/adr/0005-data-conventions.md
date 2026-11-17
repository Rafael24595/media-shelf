# ADR-0005: Data conventions (platform names, original titles, tags and genres)

- **Status:** Accepted
- **Date:** 2026-11-10

## Context

The second cataloging run (7 photos, 769 rows in total) exposed three inconsistencies:

- `platform` is free text, and "Xbox Series X|S" contains the `|` list separator.
- `original_title` sometimes held an English international title for non-English productions
  (e.g. "Kung Fu Hustle" for 功夫), and Chinese titles mixed Cantonese and Mandarin
  romanizations.
- `tags` and `genre` accumulated synonyms written by different agents: `based-on-true-story`,
  `based-on-true-events` and `true-story`; `action-rpg` and `action-role-playing`; `history` and
  `historical`; `spielberg` and `stanley-kubrick` (short vs full names); tags repeating the row's
  genre (`comedy`, `sci-fi`…); a `box-set` tag duplicating `notes`.

## Decision

### 1. Platform vocabulary (`video_games.csv`)

`platform` becomes a controlled vocabulary. A value never contains `|`; where a vendor name
does, `/` is used instead.

`PS1`, `PS2`, `PS3`, `PS4`, `PS5`, `PSP`, `PS Vita`, `Xbox`, `Xbox 360`, `Xbox One`,
`Xbox Series X/S`, `NES`, `SNES`, `N64`, `GameCube`, `Wii`, `Wii U`, `Switch`, `Switch 2`,
`Game Boy`, `Game Boy Color`, `Game Boy Advance`, `DS`, `3DS`, `Master System`, `Mega Drive`,
`Saturn`, `Dreamcast`, `PC`, `Mac`.

New platforms are added to the list through an ADR.

### 2. Original titles

`original_title` is the title of the original release **in its original language** — never an
English or other international title for a non-English work. For non-Latin scripts, it is the
standard romanization of that title:

| Language | Romanization |
|---|---|
| Japanese | Modified Hepburn (macrons allowed: *Rashōmon*) |
| Mandarin | Hanyu Pinyin |
| Cantonese (Hong Kong productions) | As listed by the Hong Kong Movie Database / Hong Kong Film Archive (usually as on IMDb) |
| Korean | Revised Romanization |
| Other (Russian, Thai…) | The romanized form used by IMDb / Wikipedia |

Video games keep the title of their first release in its original market (Western games keep
their English title).

### 3. Tags and genres

- **Canonical forms:** synonyms are mapped to one canonical value (table in
  `docs/csv-schema.md`, enforced by `scripts/validate.py`). Using a non-canonical synonym is a
  validation error.
- **People** are tagged with their full name in kebab case (`steven-spielberg`, not
  `spielberg`).
- **Decades** are written as `1980s`, `1990s`…
- **No tag repeats one of the row's genres.**
- **No `box-set` tag**: box-set membership is recorded in `edition_title`/`notes`.

## Consequences

- Filtering by tag, genre or platform gives complete results.
- Agents and skills must use the canonical values; the validator catches deviations.
- New synonyms found in future runs are added to the canonical table (a data-convention change,
  recorded by amending this table through a new ADR).

## Migration

1. `scripts/migrations/adr_0005_data_conventions.py`: map genre/tag synonyms to their canonical
   form, remove tags that repeat a genre of the same row, remove `box-set` tags, check that every
   `platform` value is in the vocabulary. Originals archived as `.pre-ADR-0005.csv`.
2. Correct non-original-language `original_title` values after verifying each one against a
   source; unverifiable cases go to the review queue (`needs_check`).
3. Validate.
