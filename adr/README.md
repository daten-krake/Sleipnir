# Architecture Decision Records (ADR)

This folder tracks every significant architectural decision for Sleipnir,
the automated agentic pentest platform.

## Rules

- One decision per file, numbered sequentially: `ADR-<NNNN>-<short-slug>.md`.
- ADRs are immutable once **Accepted**; changes get a new ADR that marks the
  old one **Superseded by ADR-XXXX**.
- Every ADR states its **Status**: `Proposed` → `Accepted` | `Rejected` | `Superseded`.
- Record decisions early, even small ones. A rejected option with reasoning is
  as valuable as an accepted one.

## Fixed constraints (not up for debate)

- Implementation language: **Go**, **standard library only** — no external Go
  dependencies, ever (see ADR-0001).
- UI: **HTMX** + server-rendered Go templates.
- Packaging/runtime: **Docker** (see ADR-0002).

## Template

Copy `TEMPLATE.md` for new records. Sections: Status, Date, Context,
Decision, Consequences (positive / negative / follow-ups).
