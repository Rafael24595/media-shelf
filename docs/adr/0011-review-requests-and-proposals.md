# ADR-0011: Review requests and proposals

- **Status:** Accepted
- **Date:** 2026-11-10

## Context

Cataloging and review were done by an assistant writing straight into the data files during
a conversation. The owner wants to be able to hand photos over to *a* reviewer (an
assistant, a person, a script) without the tools depending on any of them, and to approve
the result before it changes the data.

## Decision

1. **A request/proposal contract** (`media-shelf-request/1`), defined in
   [`docs/review-requests.md`](../review-requests.md): one JSON file per request in the data
   repository's git-ignored `.source/requests/`, with a lifecycle (`pending`, `in_progress`,
   `proposed`, `applied`, `rejected`, `failed`) and a **proposal**: an ordered list of
   operations (`add_row`, `update_row`, `add_entry`, `update_entry`), each with the values the
   reviewer saw, its reason and its evidence.
2. **Reviewers propose; the owner applies.** By default a reviewer never writes the data
   files: it writes a proposal, and a tool (media-shelf-ui) shows it and applies the
   operations the owner approves, validating each one and refusing those whose data changed
   in between (conflicts). Writing straight to the data is only allowed when the owner asks
   for it explicitly in that conversation (typically a one-row fix).
3. **How reviewers are notified is configuration of the tools**, not part of the contract.
   No reviewer is named in the specification or in code.
4. Requests live in `.source/` because they belong with the photos they are about and are
   working material, not data: they are not committed. The outcome is: rows and queue
   entries changed through the normal rules, which are committed.

This extends ADR-0010 (new `.source/requests/` folder) and ADR-0004 (resolving entries
through proposals).

## Alternatives considered

- *Reviewers keep writing directly; git as the safety net.* Simpler, but changes are only
  checked after the fact, and partial acceptance means editing by hand.
- *Proposals as edited copies of the CSV files.* Easy to produce, hard to review and to
  apply partially; operations with `from`/`to` make conflicts detectable.

## Consequences

- One more step for the owner (approve), in exchange for seeing every change before it is
  made and being able to accept part of it.
- The data repository's README, `CLAUDE.md` and the skills describe the proposal flow.
