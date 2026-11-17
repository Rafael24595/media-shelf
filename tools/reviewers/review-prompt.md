# Instructions for an unattended reviewer

You are answering one media-shelf review request, with nobody available to answer
questions. Your working directory is the `media-shelf` repository; the data repository
(`media-shelf-data`) is also readable. **You cannot and must not change any file**: your only
output is a proposal, returned as JSON in your final message. A script stores it in the
request and the owner approves it in media-shelf-ui.

## Read first

1. The request file whose path you were given (`kind`, `photos`, `entries`, `note`).
2. `CLAUDE.md` (all rules apply, in particular 16: propose, do not write).
3. `docs/review-requests.md` (the proposal format) and `docs/csv-schema.md` (every value must
   follow it; `schema/schema.json` holds the vocabularies and canonical genres and tags).
4. The data files in `media-shelf-data/data/` and the open entries of `review_queue.csv`.

Photos are in `media-shelf-data/.source/` (`pending/`, `pending/processed/`, `review/`,
`review/processed/`); find each by name.

## Do the work

- `catalog`: the photos are new items. Follow the cataloging flow of `CLAUDE.md`: one
  `media-cataloger` agent per photo or small batch (at most 5 at a time), deduplicate across
  photos, then `media-auditor` on the proposed rows. If those agents are not available, run
  general-purpose agents with `agents/media-cataloger.md` / `agents/media-auditor.md` as
  instructions. Approved rows become `add_row`; merges become `update_row` on the existing
  row; anything uncertain, unidentified or rejected becomes `add_entry` (review queue) with a
  concrete question and the photo name in `photo`.
- `review`: for each entry in `entries`, use a `media-reviewer` agent (or its file as
  instructions) with the entry and all its photos. Rows to add go through `media-auditor`.
  Close an entry (`update_entry` with `status` and `resolution`) only when the photos and
  sources answer its question; otherwise keep it open and sharpen its `question`.
- The owner's `note` in the request is part of the instructions.
- Never invent data and never set personal fields (`consumed` other than `false`, `rating`,
  `condition`, `missing_items`, `reviewed=true`) unless the note states them.

## Answer

Your final message must contain **only** this JSON object, nothing before or after it:

```json
{
  "summary": "One or two sentences: what you propose and why.",
  "operations": [ ... operations exactly as in docs/review-requests.md ... ]
}
```

- `update_row` / `update_entry`: every change has `from` (the current value, exactly as in the
  file) and `to`.
- Use `ref` on `add_row` and `${ref}` in later operations to link new rows (e.g. in an entry's
  `related_ids` or `resolution`).
- Every operation has `why` and `evidence` (photo name and position, source URLs).
- Every `add_row` and `update_row` based on photos has `photos`: the names of the photos in
  which that item can be seen (only those, not every photo of the request).
- Values are CSV texts: lists joined with `|`, booleans `true`/`false`, vocabulary values only.
- If there is nothing to propose, return an empty `operations` list and say why in `summary`.
- If you cannot do the work at all (unreadable photos, missing files), return
  `{"summary": "", "operations": [], "failure": "why"}`.
