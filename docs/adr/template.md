# ADR-NNNN: Title

- **Status:** Proposed | Accepted | Rejected | Superseded by ADR-NNNN
- **Date:** YYYY-MM-DD

## Context

Why is a change needed? What problem or limitation of the current schema does it address?

## Decision

The exact change to the CSV definition (columns, types, rules, vocabularies, conventions).

## Consequences

What becomes easier or harder; impact on existing data, skills and the validator.

## Migration

Step-by-step transformation of the existing data in `media-shelf-data`, including the new
`schema_version` in `shelf.json` (or "None"). Migration programs are not written in Python
(ADR-0006).
