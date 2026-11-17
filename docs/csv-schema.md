# Inventory CSV schema

This document is the human-readable definition of the inventory files, which live in the
`media-shelf-data` repository (ADR-0008); its
machine-readable twin is [`schema/schema.json`](../schema/schema.json), which every tool reads
(ADR-0006). Both state the same rules, reflect every accepted ADR in
[`docs/adr/`](adr/README.md) and are only changed together, through an accepted ADR (see
[ADR process](adr/README.md#process)).

Current version: **ADR-0013** (2026-11-16)

## Files

All files sit in the `data/` folder of the data repository (`media-shelf-data/data/`, the
*data folder*, ADR-0010). One file per media type; every file must exist, with at least its
header.

| File | Media type | `id` prefix | Type-specific columns |
|---|---|---|---|
| `movies.csv` | Movies | `MOV-` | `actors` |
| `video_games.csv` | Video games | `GAM-` | `platform`, `actors` |
| `books.csv` | Books | `BOK-` | `isbn`, `publisher`, `pages` |
| `series.csv` | TV series and miniseries | `SER-` | `series_type`, `seasons`, `actors` |
| `review_queue.csv` | Items pending manual review (not media rows) | `REV-` | see [Review queue](#review-queue) |
| `shelf.json` | Manifest: the schema version the data conforms to | — | see [Manifest](#manifest) |

Photos are not part of the data. They live in the repository's git-ignored `.source/` folder
(ADR-0010): `.source/pending/` for new photos to catalog (`processed/` once cataloged) and
`.source/review/` for photos that answer review-queue entries (`processed/` once their entries
are closed). The repository keeps no archive copies: the state before a migration is its
previous commit (ADR-0009).

## File conventions

| Rule | Value |
|---|---|
| Encoding | UTF-8, no BOM |
| Delimiter | `,` (comma) |
| Quoting | Fields containing commas, double quotes or line breaks are wrapped in double quotes; inner double quotes are doubled (`""`). |
| Multi-value separator | `\|` (pipe), no surrounding spaces, no empty items. |
| Header | Mandatory, exactly as defined below, in this order. |
| One row per title | The same title owned in several editions/formats is a single row with several `formats`. `condition` describes the worst copy; `missing_items` lists what is missing from any copy (detail in `notes`). |

## Headers

```csv
# movies.csv
id,original_title,edition_title,synopsis,genre,creator,formats,release_date,consumed,condition,missing_items,rating,tags,notes,reviewed,actors

# video_games.csv
id,original_title,edition_title,synopsis,genre,creator,formats,release_date,consumed,condition,missing_items,rating,tags,notes,reviewed,platform,actors

# books.csv
id,original_title,edition_title,synopsis,genre,creator,formats,release_date,consumed,condition,missing_items,rating,tags,notes,reviewed,isbn,publisher,pages

# series.csv
id,original_title,edition_title,synopsis,genre,creator,formats,release_date,consumed,condition,missing_items,rating,tags,notes,reviewed,series_type,seasons,actors
```

## Common core columns (all files)

| # | Field | Type | Required | Rules |
|---|---|---|---|---|
| 1 | `id` | text | Yes | `<PREFIX>-<NNNN>` (prefix of the file, four digits, zero-padded), e.g. `MOV-0001`. Sequential per file, unique across all files. Never reused or renumbered. |
| 2 | `original_title` | text | Yes | Title of the original release **in its original language** — never an English or other international title for a non-English work. Non-Latin scripts are romanized (see [Original titles](#original-titles)). |
| 3 | `edition_title` | text | Yes | Title as printed on the owned copy, with standard capitalization (as in the Spanish distributor/FilmAffinity listing, not ALL CAPS). Equal to `original_title` when they match. For an item inside a box set whose own title is not visible, the box-set title. |
| 4 | `synopsis` | text | No | 1–2 sentences, in English, written in our own words (never copied). |
| 5 | `genre` | list | No | English, lowercase, hyphens instead of spaces, e.g. `sci-fi\|martial-arts`. Canonical forms only (see [Tags and genres](#tags-and-genres)). |
| 6 | `creator` | list | No | Movies: director. Video games: developer studio. Books: author. Series: creator/showrunner. |
| 7 | `formats` | list | Yes | Owned formats, from the file's [formats vocabulary](#formats-vocabulary). |
| 8 | `release_date` | date | No | Original first release: `YYYY-MM-DD`, or `YYYY` if only the year is known. |
| 9 | `consumed` | boolean | Yes | `true` = watched/played/read; `false` = pending (default when unknown). |
| 10 | `condition` | integer 1–5 | No | Physical wear (see [scale](#condition-scale)). Empty for digital-only items. |
| 11 | `missing_items` | list | No | Missing pieces from the [vocabulary](#missing_items-vocabulary). Empty = complete. |
| 12 | `rating` | integer 1–10 | No | Personal rating. Only when `consumed` is `true`. |
| 13 | `tags` | list | No | Free tags, English, lowercase, hyphens instead of spaces, e.g. `cyberpunk\|oscar-winner`. Canonical forms only; never repeat a genre of the same row (see [Tags and genres](#tags-and-genres)). |
| 14 | `notes` | text | No | Free notes, in English. |
| 15 | `reviewed` | boolean | Yes | `true` = the owner confirmed the row's identification and data against the owned copy; `false` = not confirmed (default for new rows). Only the owner sets `true`. A change to `original_title`, `edition_title`, `creator`, `release_date` or `formats` not dictated by the owner resets it to `false`. |

## Type-specific columns

### `actors` (`movies.csv`, `video_games.csv`, `series.csv`)

| Field | Type | Required | Rules |
|---|---|---|---|
| `actors` | list | No | Main original cast in billing order (up to about eight), full names as credited, romanized like `original_title` for non-Latin scripts. In animation and video games, the original voice actors. Original performances only: no dubbing actors (ADR-0013). A person known under another name is listed under both, the credited one first (`Dwayne Johnson\|The Rock`, `Elliot Page\|Ellen Page`); accents need no second form, since searches ignore them. Not an identifying field: changing it does not reset `reviewed`. |

It is the last column of each of these files, after their other type-specific columns.

### `video_games.csv`

| Field | Type | Required | Rules |
|---|---|---|---|
| `platform` | list | Yes | Values from the [platform vocabulary](#platform-vocabulary). The same game on several platforms is one row. |

### `books.csv`

| Field | Type | Required | Rules |
|---|---|---|---|
| `isbn` | text | No | ISBN-10 or ISBN-13 of the owned edition, digits only (final `X` allowed in ISBN-10). |
| `publisher` | text | No | Publisher of the owned edition. |
| `pages` | integer | No | Page count of the owned edition, positive. |

### `series.csv`

| Field | Type | Required | Rules |
|---|---|---|---|
| `series_type` | enum | Yes | `series` or `miniseries`. |
| `seasons` | list | No | Owned season numbers, positive integers, e.g. `1\|2`. Empty for miniseries. |

## Data language

Documentation and all free-text data (`synopsis`, `genre`, `tags`, `notes`) are in **English**.
Titles are kept as published: `original_title` in the original language and `edition_title`
as printed on the owned copy.

## Vocabularies

### `formats` vocabulary

| File | Allowed values |
|---|---|
| `movies.csv`, `series.csv` | `DVD`, `Blu-ray`, `4K UHD`, `VHS`, `Digital` |
| `video_games.csv` | `Physical`, `Digital` |
| `books.csv` | `Paperback`, `Hardcover`, `eBook`, `Audiobook` |

Digital-only formats: `Digital`, `eBook`, `Audiobook`.

### Platform vocabulary

`PS1`, `PS2`, `PS3`, `PS4`, `PS5`, `PSP`, `PS Vita`, `Xbox`, `Xbox 360`, `Xbox One`, `Xbox Series X/S`, `NES`, `SNES`, `N64`, `GameCube`, `Wii`, `Wii U`, `Switch`, `Switch 2`, `Game Boy`, `Game Boy Color`, `Game Boy Advance`, `DS`, `3DS`, `Master System`, `Mega Drive`, `Saturn`, `Dreamcast`, `PC`, `Mac`.

A value never contains `|`; where a vendor name does, `/` is used instead (`Xbox Series X/S`).

### `condition` scale

| Value | Meaning |
|---|---|
| 5 | New / sealed |
| 4 | Like new |
| 3 | Good (light wear) |
| 2 | Fair (visible wear, complete) |
| 1 | Poor (damaged or incomplete) |

### `missing_items` vocabulary

| Value | Meaning | Applies to |
|---|---|---|
| `box` | Original box/case | movies, series, video games |
| `manual` | Instruction manual | video games |
| `cover_art` | Cover / inner artwork | movies, series, video games |
| `disc` | Disc or cartridge (case only) | movies, series, video games |
| `extra_disc` | Bonus disc | movies, series, video games |
| `slipcover` | Cardboard slipcover | movies, series, video games |
| `inserts` | Leaflets, maps, posters, stickers | movies, series, video games |
| `code` | Download/DLC code (missing or redeemed) | video games |
| `dust_jacket` | Dust jacket | books |
| `pages` | Loose or torn pages | books |
| `other` | Anything else (must be detailed in `notes`) | all |

## Original titles

`original_title` is the title of the original release in its original language. Non-Latin
scripts use the standard romanization of that title:

| Language | Romanization |
|---|---|
| Japanese | Modified Hepburn (macrons allowed: *Rashōmon*) |
| Mandarin | Hanyu Pinyin |
| Cantonese (Hong Kong productions) | As listed by the Hong Kong Movie Database / Hong Kong Film Archive (usually as on IMDb) |
| Korean | Revised Romanization |
| Other (Russian, Thai…) | The romanized form used by IMDb / Wikipedia |

Video games keep the title of their first release in its original market (Western games keep
their English title).

## Tags and genres

- Canonical forms only: the synonyms below are validation errors.
- People are tagged with their full name in kebab case (`steven-spielberg`), except actors: the
  cast goes in `actors`, never in `tags` (ADR-0013).
- Decades are written `1980s`, `1990s`…
- A tag never repeats one of the row's genres.
- Forbidden tags: `box-set` (box-set membership goes in `edition_title`/`notes`), and the tags
  that named actors before ADR-0013 (`adam-sandler`, `al-pacino`, `arnold-schwarzenegger`, `bill-murray`, `bob-odenkirk`, `brad-pitt`, `bruce-willis`, `bud-spencer`, `cantinflas`, `charlie-chaplin`, `charlie-sheen`, `charlton-heston`, `christian-bale`, `christopher-lee`, `clint-eastwood`, `daniel-day-lewis`, `dean-martin`, `denzel-washington`, `eddie-murphy`, `jackie-chan`, `jerry-lewis`, `jet-li`, `jim-carrey`, `john-wayne`, `julia-roberts`, `keanu-reeves`, `kevin-costner`, `kurt-russell`, `leslie-nielsen`, `martin-lawrence`, `mel-gibson`, `michael-douglas`, `nicolas-cage`, `rob-schneider`, `robert-de-niro`, `robin-williams`, `ryan-gosling`, `sacha-baron-cohen`, `sean-connery`, `stephen-chow`, `sylvester-stallone`, `terence-hill`, `tom-cruise`, `tom-hanks`, `van-damme`, `viggo-mortensen`, `vin-diesel`, `will-ferrell`): those people go in `actors`.

The canonical tables live in `schema/schema.json` (`vocabularies`); new synonyms found later are
added there and here through an ADR.

| Genre synonym | Canonical |
|---|---|
| `action-role-playing` | `action-rpg` |
| `history` | `historical` |
| `sport` | `sports` |

| Tag synonym | Canonical |
|---|---|
| `50s` | `1950s` |
| `60s` | `1960s` |
| `70s` | `1970s` |
| `80s` | `1980s` |
| `90s` | `1990s` |
| `90s-action` | `1990s` |
| `alien` | `aliens` |
| `antihero` | `anti-hero` |
| `asimov` | `isaac-asimov` |
| `assassin` | `assassins` |
| `based-on-true-events` | `based-on-true-story` |
| `biographical` | `biopic` |
| `biography` | `biopic` |
| `british` | `british-cinema` |
| `comic-book` | `comic-adaptation` |
| `con-artist` | `con-artists` |
| `creature` | `monsters` |
| `creatures` | `monsters` |
| `cult-classic` | `cult` |
| `dc-comics` | `dc` |
| `dickens` | `charles-dickens` |
| `dumas` | `alexandre-dumas` |
| `ensemble` | `ensemble-cast` |
| `french` | `french-cinema` |
| `gangster` | `gangsters` |
| `hong-kong-cinema` | `hong-kong` |
| `kubrick` | `stanley-kubrick` |
| `lovecraft` | `h-p-lovecraft` |
| `lovecraftian` | `h-p-lovecraft` |
| `monster` | `monsters` |
| `nolan` | `christopher-nolan` |
| `palme-d'or` | `palme-d-or` |
| `period` | `period-piece` |
| `spielberg` | `steven-spielberg` |
| `spoof` | `parody` |
| `stevenson` | `robert-louis-stevenson` |
| `tarantino` | `quentin-tarantino` |
| `tolkien` | `j-r-r-tolkien` |
| `true-story` | `based-on-true-story` |
| `wwii` | `world-war-ii` |

## Examples

```csv
# movies.csv
id,original_title,edition_title,synopsis,genre,creator,formats,release_date,consumed,condition,missing_items,rating,tags,notes,reviewed,actors
MOV-0001,Blade Runner,Blade Runner,"A blade runner must hunt down four replicants who have returned to Earth.",sci-fi|noir,Ridley Scott,Blu-ray,1982-06-25,true,4,slipcover,9,cyberpunk|classic,"Final Cut edition",true,Harrison Ford|Rutger Hauer|Sean Young

# video_games.csv
id,original_title,edition_title,synopsis,genre,creator,formats,release_date,consumed,condition,missing_items,rating,tags,notes,reviewed,platform,actors
GAM-0001,Zelda no Densetsu: Toki no Ocarina,The Legend of Zelda: Ocarina of Time,"Link travels through time to stop Ganondorf.",action-adventure,Nintendo EAD,Physical,1998-11-21,true,3,box|manual,10,nintendo|retro,"Loose cartridge",true,N64,

# books.csv
id,original_title,edition_title,synopsis,genre,creator,formats,release_date,consumed,condition,missing_items,rating,tags,notes,reviewed,isbn,publisher,pages
BOK-0001,Dune,Dune,"Paul Atreides and his family take control of the desert planet Arrakis.",sci-fi,Frank Herbert,Hardcover,1965-08-01,false,4,dust_jacket,,classic|saga,,false,,,

# series.csv
id,original_title,edition_title,synopsis,genre,creator,formats,release_date,consumed,condition,missing_items,rating,tags,notes,reviewed,series_type,seasons,actors
SER-0001,The Walking Dead,The Walking Dead: Primera temporada,"A sheriff's deputy wakes from a coma into a world overrun by the dead.",horror|drama,Frank Darabont,DVD,2010,false,,,,zombies|post-apocalyptic,,false,series,1,Andrew Lincoln|Jon Bernthal|Sarah Wayne Callies
```

## Review queue

`review_queue.csv` lists things the owner must check by hand: items that could not be
identified, ambiguous versions, box sets with unknown contents, inferred data, schema gaps and
shelves that need a new photo. Entries are never deleted; they are closed as `resolved` or
`discarded`.

```csv
id,status,reason,media_type,description,question,candidates,photo,location,related_ids,created,resolved,resolution,notes
```

| # | Field | Type | Required | Rules |
|---|---|---|---|---|
| 1 | `id` | text | Yes | `REV-NNNN`, sequential, never reused. |
| 2 | `status` | enum | Yes | `open`, `resolved`, `discarded`. |
| 3 | `reason` | enum | Yes | `unidentified`, `ambiguous`, `box_set_contents`, `needs_check`, `schema_gap`, `needs_photo` (see below). |
| 4 | `media_type` | enum | No | `movie`, `video_game`, `book`, `series`; empty if unknown. |
| 5 | `description` | text | Yes | What is visible or known (spine text, label, distributor…). |
| 6 | `question` | text | Yes | What the owner must check or decide. |
| 7 | `candidates` | list | No | Possible identifications, `\|`-separated. |
| 8 | `photo` | list | No | File names (not paths) of the photos where the item appears and of those taken to answer the entry; found by name anywhere under `.source/` (ADR-0010). |
| 9 | `location` | text | No | Position in the photo (shelf, row, neighbours). |
| 10 | `related_ids` | list | No | Ids of affected data rows (e.g. `MOV-0123`). Each must exist. |
| 11 | `created` | date | Yes | `YYYY-MM-DD`. |
| 12 | `resolved` | date | No | `YYYY-MM-DD`; required when `status` is not `open`. |
| 13 | `resolution` | text | No | What was decided / which rows were added or changed; required when `status` is not `open`. |
| 14 | `notes` | text | No | Free notes. |

| `reason` | Meaning |
|---|---|
| `unidentified` | Cannot tell what the item is. |
| `ambiguous` | Several candidates (e.g. original vs remake). |
| `box_set_contents` | The contents of a set are not visible. |
| `needs_check` | A row's data is inferred and must be confirmed on the physical item. |
| `schema_gap` | Does not fit the schema; needs an ADR. |
| `needs_photo` | Not visible enough; needs a new photo. |

## Manifest

`shelf.json`, in the data folder next to the CSV files (ADR-0010), declares the version of this schema the data
conforms to (ADR-0008):

```json
{ "schema_version": "ADR-0013" }
```

It is required. The validator reports an error when it is missing, invalid, or when
`schema_version` differs from `version` in `schema/schema.json`; while they differ, tools must
not write data. A schema change is completed by migrating the data and updating this file in
the data repository.

## Validation

`media-shelf-ui validate` (repository `media-shelf-ui`) checks every file of the data
repository, including the review queue and the manifest, against `schema/schema.json`. Errors
make it exit with code 1; warnings do not.

Where the data folder and this schema are is configuration of `media-shelf-ui`
(its ADR-0002), not a convention:

| Where | Command |
|---|---|
| Windows (configured with `media-shelf-ui init`) | `media-shelf-ui.exe validate` |
| Any system, explicit paths (run from `media-shelf-data`) | `media-shelf-ui validate -data data -schema <media-shelf>/schema/schema.json` |
| From the Go source | `go run ./cmd/media-shelf-ui validate -data <data folder> -schema <schema.json>` |
