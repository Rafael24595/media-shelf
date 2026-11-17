# ADR-0007: Rename the project to media-shelf

- **Status:** Accepted
- **Date:** 2026-11-10

## Context

The project was named `media-track`, which suggests tracking progress (what was watched or
played). It has grown into a catalogue of a physical collection: editions, formats, condition,
missing pieces, a review queue. The owner chose a name that reflects the shelves the data comes
from.

## Decision

- The project is renamed `media-track` → `media-shelf`; its companion tool
  `media-track-ui` → `media-shelf-ui` (ADR-0003 there).
- The data files, their schema and every id stay exactly as they are.
- Accepted ADRs (0001–0006) and the historical migration scripts keep the old name; they
  record decisions made under it.
- The owner-facing skills are renamed `media-shelf-add-from-images`, `media-shelf-manage` and
  `media-shelf-schema-change`.

## Consequences

- Paths, commands and documentation refer to `media-shelf` / `media-shelf-ui`.
- The old skills (`media-track-*`) must be removed by the owner once the new ones are saved.

## Migration

1. Rename the folders `media-track` → `media-shelf` and `media-track-ui` → `media-shelf-ui`.
2. Update `CLAUDE.md`, `docs/csv-schema.md`, the ADR index, the agents and the `$comment` of
   `schema/schema.json` (no change to the data definition, so `version` stays `ADR-0006`).
3. Point the `media-shelf-ui` configuration to the renamed folder and validate.
