# Architecture Decision Records

Every change to the definition of the inventory files (in `media-shelf-data`) — columns, names, order, types,
required flags, vocabularies, scales, file conventions or location — **must** be recorded
in an ADR before it is applied.

## Index

| ADR | Title | Status | Date |
|---|---|---|---|
| [0001](0001-initial-csv-schema.md) | Initial CSV schema | Accepted | 2026-11-09 |
| [0002](0002-split-title-into-original-and-edition.md) | Split `title` into `original_title` and `edition_title` | Accepted | 2026-11-09 |
| [0003](0003-one-file-per-media-type.md) | One CSV file per media type | Accepted | 2026-11-09 |
| [0004](0004-review-tracking.md) | Review tracking (`reviewed` flag and review queue) | Accepted | 2026-11-09 |
| [0005](0005-data-conventions.md) | Data conventions (platform names, original titles, tags and genres) | Accepted | 2026-11-10 |
| [0006](0006-machine-readable-schema-and-go-validator.md) | Machine-readable schema and Go validator | Accepted | 2026-11-10 |
| [0007](0007-rename-to-media-shelf.md) | Rename the project to media-shelf | Accepted | 2026-11-10 |
| [0008](0008-separate-data-repository.md) | Separate data repository with a schema-version manifest | Accepted (partially superseded by ADR-0009) | 2026-11-10 |
| [0009](0009-self-contained-data-repository.md) | Self-contained data repository without archive copies | Accepted (partially superseded by ADR-0010) | 2026-11-10 |
| [0010](0010-data-folder-and-photo-sources.md) | Data folder and photo sources in the data repository | Accepted | 2026-11-10 |
| [0011](0011-review-requests-and-proposals.md) | Review requests and proposals | Accepted | 2026-11-10 |
| [0012](0012-claude-reviewer-adapter-in-go.md) | Claude reviewer adapter in Go | Accepted (partially superseded by ADR-0014) | 2026-11-16 |
| [0013](0013-actors-column.md) | `actors` column for movies, video games and series | Accepted | 2026-11-16 |
| [0014](0014-go-assert-in-go-tests.md) | go-assert in the tests of the Go programs | Accepted | 2026-11-18 |

## Process

1. Copy [`template.md`](template.md) to `NNNN-short-kebab-title.md` with the next number.
2. Fill it in with status **Proposed** and share it for review.
3. Once the owner accepts it, set the status to **Accepted** and, in the same change:
   - update [`docs/csv-schema.md`](../csv-schema.md) (including its "Current version" line)
     and [`schema/schema.json`](../../schema/schema.json) (including `version`);
   - if the change needs new validation logic (not just vocabulary or columns), update the
     validator in the `media-shelf-ui` project, with tests;
   - in `media-shelf-data`, migrate the files as described in the ADR's *Migration* section
     and set `schema_version` in `shelf.json` to the new version (same commit there);
   - add the ADR to the index above.
4. Run `media-shelf-ui validate` — it must report no errors.
5. ADRs are immutable once accepted. To reverse or amend one, write a new ADR and set the
   old one's status to **Superseded by ADR-NNNN**.
6. Rejected proposals are kept with status **Rejected**.

All ADRs are written in English.
