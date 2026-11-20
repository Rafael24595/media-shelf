# media-shelf

📚 Schema, decisions and AI agents to catalog a personal collection of movies, video games,
books and TV series.

media-shelf keeps an inventory of every physical (and digital) item the owner has, in plain
CSV files. This repository holds the **specification and the process**: what the files look
like, why, and how items get into them. The data and the tooling live next to it:

| Repository | What it holds |
|---|---|
| **media-shelf** (this one) | The data definition, its decisions (ADRs), the agents and the reviewer adapter |
| [media-shelf-data](https://github.com/Rafael24595/media-shelf-data) | The data: one CSV file per media type, the review queue and the manifest |
| [media-shelf-ui](https://github.com/Rafael24595/media-shelf-ui) | A local web interface to browse and edit the data, and the reference validator |

The three repositories are usually checked out in one common folder, but no tool assumes it:
every location is configuration.

## Layout

```text
schema/schema.json          the definition of the data files that every tool reads
docs/csv-schema.md          the same definition, for people
docs/adr/                   architecture decision records: why the data looks as it does
docs/review-requests.md     the contract between the owner's tools and reviewers
agents/                     Claude Code agents: cataloger, auditor, reviewer
tools/reviewers/            reviewer adapters (claude-review) and the prompt they share
scripts/migrations/         data migrations, one Go program per schema change
CLAUDE.md                   the rules Claude follows when working on the inventory
```

## The data definition

[`docs/csv-schema.md`](docs/csv-schema.md) describes every file and column: ids, titles in
their original language, formats, condition, review status, actors and the vocabularies of
genres, tags and platforms. [`schema/schema.json`](schema/schema.json) states the same rules
for programs, and `media-shelf-ui validate` checks the data against it.

The definition only changes through an ADR, accepted before it is applied (see
[`docs/adr/README.md`](docs/adr/README.md)). A change takes two steps: here, the ADR plus a new
`version` in the schema and its description; in media-shelf-data, the migrated data plus the
new version in its manifest (`shelf.json`). While the two versions differ, the validator
reports it and the web interface refuses to write.

## Cataloging with Claude

New items are cataloged from photos of the shelves. The agents in [`agents/`](agents) do the
research and never write anything:

- **media-cataloger** identifies the items in a photo and proposes their rows.
- **media-auditor** checks those rows independently against sources, the photo and the schema.
- **media-reviewer** works one entry of the review queue: what could not be identified or
  confirmed.

Their results reach the data as a **proposal** that the owner approves in media-shelf-ui
([`docs/review-requests.md`](docs/review-requests.md)), never as a direct write. To use the
agents, copy them to `.claude/agents/`.

### claude-review

[`tools/reviewers/claude-review`](tools/reviewers/claude-review) answers a review request
without anyone at the keyboard: when the owner asks Claude for a catalog or a review from
media-shelf-ui, it runs this program, which writes the proposal into the request. It is a Go
program using only the standard library, and needs Claude Code installed and signed in (each
run uses the owner's Claude plan).

Build it into `dist/` (git-ignored), from `tools/reviewers/claude-review`:

```sh
go test ./...
go build -o dist/claude-review.exe .                              # on Windows
GOOS=linux GOARCH=amd64 go build -o dist/claude-review-linux-amd64 .  # for Linux
```

and configure it as a notifier of media-shelf-ui, with `{request}` as the last argument:

```json
"notifiers": [{
  "name": "Claude",
  "command": ["<media-shelf>\\tools\\reviewers\\claude-review\\dist\\claude-review.exe",
              "-shelf", "<media-shelf>", "{request}"],
  "timeout_seconds": 3600
}]
```

## Conventions

- Everything written here (documents, code comments, commit messages) is in English.
- ADRs are immutable once accepted: a later ADR amends or supersedes them.
- Tools are written in Go; there are no Python scripts (ADR-0006).
