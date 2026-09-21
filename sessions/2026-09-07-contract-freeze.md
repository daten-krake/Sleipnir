# Session 2026-09-07 — Contract freeze (A0–A2)

- **Date:** 2026-09-07
- **Session id:** 01a07cc9-ddd5-7ea1-9c10-b263c0fb1dbb
- **Model:** qwen-token-plan-individual/qwen3.8-max
- **Goal:** Freeze the contract core: `contracts/A0-conventions.md`,
  `A1-events.md`, `A2-graph.md` (+ `contracts/README.md`) as one PR, so the
  A3–A8 fan-out and the first code package can start. No product code this
  session (`next_steps.md` / Q15: no coding until contracts are frozen).

## Done this session

- Housekeeping: PR #1 verified **merged**; local `main` fast-forwarded to
  `7726564`; merged branch `layout/program-layout` deleted (local + remote).
- Carry-over resolved: `sleipnir-implementer` model `qwen3.8-flash` resolves
  and runs (smoke test returned `SMOKE-TEST-OK`, provider `qwen-token-plan`).
  Lesson: the subagent harness flags read-only tasks as "completed without
  making edits" → use `sleipnir-architect` for read-only smoke tests.
- PR tooling verified: Linux `gh` is unauthenticated; the Windows `gh.exe`
  credential helper is authenticated but cannot operate on this WSL checkout
  ("dubious ownership"). Working recipe = `GH_TOKEN=$(gh.exe auth token) gh …`
  → recorded in `WORKFLOW.md` §7.
- Branch protection verified live: ruleset **"protect main"** (id 22285859,
  `active`): `deletion`, `non_fast_forward`, `pull_request`.
- `contracts/` created with `README.md`: doc lifecycle
  (Draft → Frozen → Implemented), uniform document shape, additive-only
  versioning, and the shared contract-test suite as the fan-out merge gate.
- `contracts/A0-conventions.md` drafted (architect, 773 lines): closed id
  scheme (`A0-1`, 11 prefixes, ULID-layout Crockford base32, stdlib-only) ·
  canonical JSON (`A0-2`, RFC 8785 with two declared restrictions, 6
  normative test vectors, exclusion lists, store-the-canonical-bytes rule) ·
  13 error kinds → HTTP + envelope (`A0-3`) · cursors (`A0-4`) · ms-precision
  UTC (`A0-5`) · the read/write unknown-field asymmetry (`A0-6`) · Q4 size
  caps with mechanisms T/R (`A0-7`) · field/enum conventions (`A0-8`) ·
  traceability table · 14 PO-confirm items.
- A0 surfaced a real spec conflict (**A0-7.10**): the Q4 caps are not jointly
  satisfiable for a maximal stage view (500 × 512 B ≈ 250 KiB ≫ 64 KiB, so
  ~131 B per node) → A3 must define the composition rule; escalated to PO.
- Process lesson: a 30-minute child budget is too tight for a contract doc of
  this size — the A0 run timed out after writing the file but before
  reporting. Relaunch uses 60 min + a "write the skeleton to disk in the
  first 15 minutes" rule.
- (append as work happens)

## Decisions

- Contract documents live in top-level **`contracts/`** (PO, 2026-09-07);
  normative build input per SPEC §12.1. Authority: ADR > SPEC > DESIGN >
  contract.
- A0 + A1 + A2 ship as **one PR** (PO, 2026-09-07) on branch
  `contracts/a0-a2-conventions`, label `agent-built`.
- Option C (A3–A8 fan-out briefs) deferred to the next session: briefs
  derived from not-yet-frozen contracts would need rewriting after review.
- No product Go code this session — contracts only, Go appears as
  illustrative type sketches in fenced blocks.

## Open questions carried forward

- A3–A8 fan-out briefs (super-minimal, WORKFLOW §5) — next session.
- First code package: repo scaffold + `internal/errs` + `internal/logging`
  (DESIGN §1, ADR-0019) — once this PR is merged.
- Shared contract-test suite (specified in `contracts/README.md`) to be
  written against the frozen contracts.
- PO-confirm items raised inside the A0–A2 drafts → resolved at PR review.
- Standing: Pi offline capability (session 4); AD technique scope for the v1
  tool registry; backlog 9–11 (CI/CD, observability, model drift).

## Token usage

| total input | uncached input | cache read | cache write | output | reasoning |
|---|---|---|---|---|---|
| 669095 | 63015 | 606080 | 0 | 22175 | 11833 |

_(run `sessions/update-usage.sh sessions/2026-09-07-contract-freeze.md` at session end)_
