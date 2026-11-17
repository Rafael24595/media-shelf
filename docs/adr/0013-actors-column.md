# ADR-0013: `actors` column for movies, video games and series

- **Status:** Accepted
- **Date:** 2026-11-16

## Context

The owner wants to find titles by the people who act in them. The only place for people
besides `creator` is `tags` (ADR-0005: "people by full name"), which mixes them with themes and
is not a list of the cast. Books have no cast.

## Decision

A new optional column `actors` (type `list`, `|`-separated) at the end of `movies.csv`,
`video_games.csv` and `series.csv`, after their other type-specific columns:

- the **main original cast**, in billing order, full names as credited (romanized as for
  `original_title` when the original script is not Latin);
- in animation and video games, the original voice actors: it is the only way they appear;
- **original performances only**: dubbing actors are not recorded;
- a handful of names (up to about eight), not the full cast. This is guidance for the
  catalogers, not a validation rule.

It is not an identifying field: changing it does not reset `reviewed`.

**Actors are no longer tags.** ADR-0005 tags people by full name; from now on that applies to
people other than the cast (directors, writers, composers…). The tags that named actors move
to `actors` with the name as credited (`van-damme` → `Jean-Claude Van Damme`) and become
forbidden tags, so the validator rejects them if they are used again; the synonyms that
pointed to them (`chaplin`, `eastwood`, `schwarzenegger`, `stallone`) are removed. The 48
tags were checked row by row: every tagged person acts in that title, actor-directors
included (Clint Eastwood, Kevin Costner, Jerry Lewis, Charlie Chaplin, Stephen Chow, Jackie
Chan).

## Consequences

- The interface searches and filters by the column without changes (it reads the schema).
- Existing rows get the column empty except where an actor was tagged (186 rows); completing
  the cast is a cataloging task (review requests), not part of the migration.
- A person who directs a title without acting in it stays out of `actors` (the director is
  in `creator`). The forbidden list only knows the 48 names that were tags; for anyone else
  the rule is followed by the catalogers, not checked.
- `books.csv` does not change.

## Migration

In `media-shelf-data`, with `shelf.json` at `ADR-0008`:

```sh
cd scripts/migrations/adr_0013_actors && go run . <media-shelf-data>/data
```

It adds `,actors` to the header and an empty last field to every record of the three files,
keeping every other byte; then moves the actor tags (`actors.go`: the tag → name table, and
the billing order of the three rows whose tags were not in that order), encoding the files as
the data tools do (`encoding/csv`, which reproduces the current files byte for byte). Each
step is checked: only `tags` and `actors` may change, and no actor tag may remain. It sets
`schema_version` to `ADR-0013` and writes nothing unless all three files can be migrated.
