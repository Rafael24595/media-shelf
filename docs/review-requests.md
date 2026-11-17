# Review requests and proposals

Contract `media-shelf-request/1` (ADR-0011). It lets any reviewer — an assistant, a person,
a script — be asked to catalog or review photos, and return its work as a **proposal** that
the owner approves before anything changes in the data. Nothing in it is specific to one
reviewer.

## Where

One JSON file per request in the data repository's git-ignored `.source/requests/`, named
after its id: `.source/requests/REQ-20261110-154400.json`. UTF-8, LF.

## Lifecycle

```text
pending ──► in_progress ──► proposed ──► applied
   │             │              └──────► rejected
   └─────────────┴──► failed
```

| Status | Set by | Meaning |
|---|---|---|
| `pending` | whoever creates the request (the UI when notifying, or a reviewer working on the owner's direct instruction) | Waiting for a reviewer. |
| `in_progress` | reviewer | Claimed: `claimed_by`, `claimed_at` set. |
| `proposed` | reviewer | `proposal` written; waiting for the owner. |
| `applied` | the tool that applies it (media-shelf-ui) | The owner approved some or all operations; `review` says which and how they went. |
| `rejected` | the tool / the owner | Nothing applied; `review.note` says why. |
| `failed` | reviewer | The reviewer could not finish; `failure` says why. |

A request in `applied`, `rejected` or `failed` is final; further work is a new request. To
retry a failed request, a tool creates a new one with the same `kind`, `photos`, `entries` and
`note`, and sets `superseded_by` on the failed one to the new id. A final request may be
deleted by the owner (its file removed); open ones (`pending`, `in_progress`, `proposed`) may
not, because a reviewer may still be writing them or a proposal is still undecided.
Reviewers never change the data files: only the owner's approval does, through a tool that
validates every operation (media-shelf-ui).

## Fields

| Field | Type | Who | Description |
|---|---|---|---|
| `contract` | text | creator | Always `media-shelf-request/1`. |
| `id` | text | creator | `REQ-YYYYMMDD-HHMMSS`, with `-2`, `-3`… if taken; equals the file name. |
| `kind` | `catalog` \| `review` \| `other` | creator | Catalog new photos; work review-queue entries; anything else (explained in `note`). |
| `status` | see above | — | |
| `created` | RFC 3339 time | creator | |
| `created_by` | text | creator | Free label (e.g. `media-shelf-ui`, the reviewer's name). |
| `schema_version` | text | creator | `schema_version` of the data when the request was made. |
| `photos` | list of names | creator | Photo file names under `.source/` (as in the review queue). |
| `entries` | list of ids | creator | Review-queue entries concerned (`review`). |
| `note` | text | creator | Instructions from the owner. |
| `claimed_by`, `claimed_at` | text, time | reviewer | |
| `proposal` | object | reviewer | See below. |
| `failure` | text | reviewer | Why the request failed. |
| `notifications` | list | tool | `{notifier, at, ok, message}` for each notification sent. |
| `review` | object | tool | `{at, note, results: [{index, status, id, message}]}` after applying or rejecting. |
| `superseded_by` | text | tool | On a failed request: the id of the request that retried it. |

Tools must keep fields they do not know when they rewrite a request.

## Proposal

```json
"proposal": {
  "by": "reviewer name",
  "at": "2026-11-10T16:05:00+01:00",
  "summary": "Box set REV-0003 holds four films; two were already rows.",
  "operations": [
    { "op": "add_row", "media": "movie", "ref": "t1",
      "fields": { "original_title": "Tremors", "edition_title": "Temblores", "formats": "DVD", "consumed": "false", "reviewed": "false" },
      "why": "Listed on the back of the set.", "evidence": ["REV-0003_back.jpg, top left", "https://…"],
      "photos": ["REV-0003_back.jpg"] },
    { "op": "update_row", "media": "movie", "id": "MOV-0050",
      "changes": { "notes": { "from": "", "to": "Part of the Tremors Master Collection box set." } },
      "why": "…" },
    { "op": "add_entry",
      "fields": { "reason": "needs_check", "media_type": "movie", "description": "…", "question": "…", "related_ids": "${t1}" } },
    { "op": "update_entry", "id": "REV-0003",
      "changes": { "status": { "from": "open", "to": "resolved" },
                   "resolution": { "from": "", "to": "Four films; added Tremors (${t1})." },
                   "related_ids": { "from": "", "to": "MOV-0050|${t1}" },
                   "photo": { "from": "PXL_20261109_200958707.jpg", "to": "PXL_20261109_200958707.jpg|REV-0003_back.jpg" } } }
  ]
}
```

| `op` | Required | Effect |
|---|---|---|
| `add_row` | `media` (a key of `media_types`), `fields` | New row; the id is assigned when applied. `ref` names it for later operations. |
| `update_row` | `media`, `id`, `changes` | Changes fields of a row. |
| `add_entry` | `fields` | New review-queue entry (open; id assigned when applied). `ref` allowed. |
| `update_entry` | `id`, `changes` | Changes fields of an entry, including closing it (`status`, `resolution`). |

- **Values** are the CSV texts (lists with `|`, booleans `true`/`false`), and must follow the
  schema of `schema_version`.
- **`changes`** gives, per field, the value the reviewer saw (`from`) and the new one (`to`).
  When applying, an operation whose `from` values no longer match the data is a *conflict*
  and is not applied.
- **References:** `${ref}` inside a value or an `id` is replaced by the id assigned to the
  operation with that `ref`. An operation that refers to one that was not applied is skipped.
- **Order:** operations are applied in order; put `add_row` before the operations that refer
  to it.
- **`why`** (one sentence) and **`evidence`** (photo and position, source URLs) are shown to
  the owner next to each operation.
- **`photos`** (optional, `add_row` and `update_row`): the photo file names (as under
  `.source/`) in which the row's item can be seen. Once the operation is applied, the
  interface adds them to the item's images, if an images folder is configured
  (media-shelf-ui ADR-0015). A shelf photo can be listed in every row it shows.

## Notifying reviewers

How a reviewer learns about a request is the tool's configuration, not part of this
contract: media-shelf-ui can run a configured command or call a configured URL when the owner
asks for a review (its ADR-0008). Without any configured, requests are created by a reviewer
when the owner asks it directly.

## Reviewer adapters

`tools/reviewers/` holds adapters that turn a request into work for a specific reviewer.
They are optional and outside the tools' code; a notifier points to one.

| Adapter | What it does |
|---|---|
| `claude-review/` (+ `review-prompt.md`) | Go program (ADR-0012), Windows and Linux. Marks the request `in_progress`, runs Claude Code non-interactively in this repository with read-only tools (read, search the web, agents; no file writes) and the data repository readable, and writes the JSON proposal it returns into the request (`proposed`), or `failed` with the reason. |

### claude-review

Build it once, from `tools/reviewers/claude-review/` (Go 1.27):

```sh
go build -o dist/claude-review.exe .               # on Windows
go build -o dist/claude-review-linux-amd64 .       # on Linux
go test ./...
```

It needs Claude Code installed and signed in (the `claude` command); each run uses the
owner's Claude plan. Notifier in media-shelf-ui's configuration, with `timeout_seconds` 3600:

```json
{ "name": "Claude",
  "command": ["E:\\…\\media-shelf\\tools\\reviewers\\claude-review\\dist\\claude-review.exe",
              "-shelf", "E:\\…\\media-shelf", "{request}"],
  "timeout_seconds": 3600 }
```

On Linux, the same with `…/dist/claude-review-linux-amd64` and Linux paths. `-claude <command>`
sets the Claude Code command when it is not `claude` on the `PATH`. Exit code 0: answered (or
nothing to do: the request was not `pending`); 1: the request failed, with the reason in it;
2: the request could not be read or written.
