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
- **A1 completed** (architect child, this run): `contracts/A1-events.md`
  347 → 2469 lines. Wrote A1-4 (payload rules, 12 clauses), A1-5 (hash chain,
  10), A1-6 (verification + failure behaviour, 9), A1-7 (write path, 12), A1-8
  (read/stream guarantees, 9), §4 types (Go sketch + JSON examples + a
  **normative 3-event chain vector** whose SHA-256s were computed and are
  reproducible from the document bytes), §5 traceability, §6 (14 PO items +
  AM-1..AM-4 + cross-contract requests).
- Additive A1-3 changes for A2 (A0-6.5): kinds `quarantine_recomputed`
  (A2-8.5) and `graph_edge_retracted` (A2-3.9) → 37 → **39 kinds**;
  enum values `chain_break_detected.break_kind: engagement_mismatch` (A12
  splicing) and `action_blocked.reason: append_rejected`; new clause **A1-3.8**
  answers every A2 cross-contract request (confirmed / corrected / added),
  incl. **no** `node_superseded` kind (one fact, one encoding).
- A0 obligations discharged in A1: A0-2.12/2.13 (A1-5.2, A1-5.10), A0-2.14
  (A1-4.1), A0-2.16 (A1-5.7), A0-3.11 (A1-7.6/7.7), A0-4.3 (A1-8.1/8.2),
  A0-5.6 (A1-5.4, A1-6.1), A0-5.7 (A1-4.12, A1-7.3), A0-6.6 (A1-4.10),
  A0-7.7 (A1-4.5: mechanism **R** for every A1 class, T for none).
- A0 amendment requests raised by A1: AM-1 `usr_` (freeze blocker, same as A2
  AM-1) · AM-2 extend A0-7.7 to A1 + adopt the A1-local cap constants ·
  AM-3 record the `recorded_at` forward clamp in A0-5.4 · AM-4 add
  "cursor id not in this collection" to A0-4.8's `validation` cases.
- Minor consistency edits inside A1-1..A1-3 (all listed in the run report):
  header package name `internal/event` → `internal/events` (DESIGN §1) and the
  additive A1-3 changes above. No A1-1/A1-2 clause text changed.
- **Recovery from the child timeout:** the A1 architect child hit its 60-min
  limit *after* writing the whole document (2472 lines) and updating this
  tracker — only its final report was lost. Recovered from disk, verified, and
  committed as `73c2fa8`. Lesson recorded: a child's deliverable must be a
  **file written incrementally**, never the final message.
- **Integrator verification of A1 §4.3 (normative chain vector):** recomputed
  independently with a plain sorted-key compact JSON encoder + SHA-256 — all
  three rows are **byte-exact** (preimage 428/816/715 B, the three `hash`
  values, served 589/977/876 B and their digests), canonical form stable under
  re-encoding, `prev_hash` links chain from `ChainZero`, and `hash`/
  `prev_hash`/`seq` are absent from every preimage. Evidences C1 (stdlib-only
  feasibility of the chain) and A0-2 determinism for the event spine.
- Reviews relaunched (workflow `c7a3dc5d`, 50-min budget each, in parallel):
  `sleipnir-principal` (buildability, cross-doc consistency, AM rulings,
  work-package split, test coverage) and `sleipnir-architect` (adversarial).
  Both must write findings to `/tmp/a0a2-review-{principal,architect}.md`
  **incrementally**, starting in the first 10 minutes, so a timeout cannot lose
  the report again.
- PR body drafted at `/tmp/pr-body-a0a2.md` with the full PO checklist: A0 §6
  (14), A2 §6 (11), A1 §6 (14), and the deduplicated amendment requests
  (A1 AM-1 ≡ A2 AM-1 `usr_`; A1 AM-2 overlaps A2 AM-2; A1 AM-3, AM-4).
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
| 12087529 | 283113 | 11804416 | 0 | 144134 | 64146 |

_(run `sessions/update-usage.sh sessions/2026-09-11-contract-freeze-a1.md` at session end)_
