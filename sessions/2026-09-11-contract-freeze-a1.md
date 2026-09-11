# Session 2026-09-11 — Contract freeze A1 completion + A0–A2 reviews + PR

- **Date:** 2026-09-11
- **Session id:** 01a090d6-349c-74e0-9039-9baebc6458e1
- **Model:** qwen-token-plan-individual/qwen3.8-max
- **Goal:** Resume the quota-cut session of 2026-09-07
      (`sessions/2026-09-07_quota_limit.md`): finish `contracts/A1-events.md`
      (A1-4…A1-8 + §4/§5/§6), run the principal + architect reviews over
      A0/A1/A2, apply the fixes, and land the branch as a PR carrying the
      product-owner decision checklist. No product code this session.

## Branch

`contracts/a0-a2-conventions` (existing, pushed; A0 + A2 already committed at
`733eadb` / `461200a`). One PR for A0 + A1 + A2 (PO decision 2026-09-07).

## Done this session

- Session ritual completed: SPEC → ADRs (all 20 Accepted, none Proposed, none
  changed since 2026-09-07) → DESIGN → BACKLOG → AGENTS/WORKFLOW.
- Housekeeping verified: tree clean; PR #1 merged; **no PR existed** for
  `contracts/a0-a2-conventions`; Linux `gh` still unauthenticated, the
  WORKFLOW §7 recipe (`GH_TOKEN=$(gh.exe auth token)`) works and lists PRs.
- Correction to the quota snapshot: **A1-4 (payload rules) is also an empty
  stub**, not only A1-5…A1-8. Real A1 content = A1-1…A1-3 (347 lines).
- Missing artifact noted: `sessions/style-notes.md` (em/en dash + quote rules)
  is referenced by the quota snapshot but does not exist in the repo.
- (append as work happens)

## Decisions

- (record as they happen; ADR-worthy outcomes go to `adr/` as **Proposed**)

## Open questions carried forward

- PO decision queue: A0 §6 (14 items, incl. **A0-7.10** which blocks A3) and
  A2 §6 (11 items + **AM-1…AM-4**, where AM-1 `usr_` is on A2's critical path).
- A3–A8 fan-out briefs (super-minimal, WORKFLOW §5) — after the freeze.
- First code package (scaffold + `internal/errs` + `internal/logging`) — after
  the PR merges.
- Shared contract-test suite (JSON round-trip, unknown-field tolerance,
  negative cross-engagement tests) — the fan-out merge gate.
- Standing: Pi offline capability (backlog 4); AD technique scope for the v1
  tool registry; backlog 9–11 (CI/CD, observability, model drift).

## Token usage

| total input | uncached input | cache read | cache write | output | reasoning |
|---|---|---|---|---|---|
| 320308 | 45236 | 275072 | 0 | 8359 | 4371 |

_(run `sessions/update-usage.sh sessions/2026-09-11-contract-freeze-a1.md` at session end)_
