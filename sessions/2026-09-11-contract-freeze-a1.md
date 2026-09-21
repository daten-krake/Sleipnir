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
- **Both reviews completed** (workflow `c7a3dc5d`) and committed as the audit
  trail in `docs/reviews/` (`203ee04`):
  - `2026-09-11-contract-review-principal.md` — 105 findings
    (**35 MUST** / 62 SHOULD / 8 NICE), verdict on buildability: *a junior
    implementer cannot build `internal/ids`/`cjson`/`events`/`graph` from the
    three documents alone* (31 MUST FIX gaps, none architectural); all **8
    amendment requests ruled** (AM-1 `usr_` ACCEPT = the single freeze blocker;
    A2 AM-3 accepted *in part* — reservation yes, blessing A2's field names no);
    work-package split WP-01… in §4.
  - `2026-09-11-contract-review-adversarial.md` — 57 findings
    (**28 MUST** / 20 SHOULD / 9 NICE). Most serious: **C-02** the approved
    action spec is never stored, so `fingerprint_hash` is unlinkable to any
    bytes (ADR-0018 §3 fails); **C-01** A1-6.5 override authority contradicts
    itself *and* narrows Q11 without a PO decision; **C-04** scope-change vs
    blacklist re-check; **S-02** tail truncation (A11); **F-01/F-02/F-03**
    canonical-JSON edge cases (floats, invalid UTF-8/lone surrogates,
    U+2028/9); **E-01…E-04** MUSTs with no negative test.
- **Integrator verification of A0-2.17 (V1–V6, normative canonical-JSON
  vectors):** all six reproduce exactly — canonical bytes, stated lengths and
  SHA-256 digests — and re-canonicalizing each input cell (sorted keys, compact
  separators, `ensure_ascii=False`) reproduces the canonical column byte for
  byte, V5 included after applying the `{seq, prev_hash, hash}` exclusion list.
  This is independent evidence that A0-2 is implementable with a plain
  stdlib-style encoder (C1) and that the declared UTF-8-byte-order deviation
  from RFC 8785 is what V6 locks.
- **New integrator finding I-01 (MUST FIX, not in either review):** in the
  *normative* A0-2.17 table the `\uXXXX` display form means two different
  things. V3's canonical cell needs `\u0001` read as **literal 6-char escape
  text** (raw length 57 = stated), while V6's canonical cell needs `\uFFFD`
  read as **the character U+FFFD** (decoded length 18 = stated; raw would be
  21). A builder transcribing the table into the contract-test suite therefore
  cannot apply one rule, and V6's printed cell contradicts its own `len`
  column. Fix: print V6 with the literal character (as V3 already does for
  `é中😀`) or add an explicit byte/hex column plus a note stating when an
  escape is text and when it is a character. → hand to the A0 writer.
- **Triage stage launched** (`490a1d06`, `sleipnir-principal`, 45 min): merge
  both reports into one authoritative `/tmp/fix-plan.md` — every MUST FIX
  assigned exactly once to one of three single-file writers, duplicates
  collapsed, reviewer conflicts resolved (adversarial wins on safety semantics,
  principal on notation/buildability), PO-only decisions separated from
  agent-fixable ones, and digest-critical changes flagged for vector
  recomputation. Rationale: three writers working from two overlapping reports
  would resolve the same conflict three different ways.
- **Triage stage completed** (`sleipnir-principal`, 649-line plan, committed as
  `docs/reviews/2026-09-11-contract-fix-plan.md`): all **63 MUST FIX assigned
  exactly once** (2 PO-only, 13 A0, 30 A1, 18 A2), 90 of 99 SHOULD/NICE
  assigned, 9 deferred with additive-safety verdicts, 14 duplicate findings
  collapsed, **17 paired rows** requiring byte-identical text in 2–3 files, and
  18 reviewer conflicts resolved by a stated precedence rule (adversarial wins
  on safety semantics, principal on notation/buildability/naming, stricter
  option + PO signature when both are safety-relevant).
- **All 12 vector values in the plan independently recomputed by the integrator
  before the writers ran** (V7 19 B, V8 28 B, A1 §4.3 row 3 preimage 702 B +
  served 863 B, A1-7.6 `PayloadHash` 293 B, A2 §4.2 F1 656 B, F1-R's forbidden
  unsorted digest, F3 656 B, S1 304 B) — every length and digest exact, so the
  writers publish verified values rather than recomputed guesses.
- **Three writers ran in parallel** (one file each, workflow `00ee86b3`); all
  three were cut short again, but the partial work survived on disk and was
  committed as WIP `9ab3a26` after integrator verification (A0-2.17 V1–V8 still
  byte-exact, A1 §4.3 rows 0–2 untouched).
- **A0 finished by the integrator** (`dfb54ac`): all 21 plan rows + PO-2 + PO-8
  verified present by grep audit, plus two integrator findings:
  - **I-01** (MUST FIX, missed by both reviewers): the canonical-bytes column of
    the *normative* A0-2.17 table had two contradictory readings of `\uXXXX` —
    V3 only reaches len 57 if `\u0001` is six literal characters, V6 only
    reaches len 18 if `\uFFFD` is the character. A notation rule now says which
    is which and that the `len` column decides.
  - **I-02**: raw U+2028/U+2029 in a Markdown table row made the file's line
    count reader-dependent (`str.splitlines` 1094 vs `grep`/`awk` 1092), so a
    normative vector parsed differently per tool — found because my own
    verifier silently skipped V7/V8. Line-break/control code points are now
    `<2028>`/`<2029>`/`<7F>` placeholders with a byte legend; decoded they still
    reproduce len 19 / len 28 and both digests.
- **`contracts/README.md` integrator-only items** (`7694c4a`): the A0 tie-break
  rule (P-68) and the seventh merge-gate bullet **safety-path pairing** (E-06) —
  the rule that makes the ~60 test ids the plan assigns enforceable.
- Continuation writers for A1 (11 rows + 5 PO rows) and A2 (9 rows + 5 PO rows)
  launched with the parent-verified remaining-work audit instead of their own
  stale logs (workflow `fe5bae3a`).
- **Continuation writers finished A1 and A2** (workflow `fe5bae3a`, committed
  `133fbdf`): A1 taxonomy **39 → 42 kinds** (the three new `Kind` consts were
  missing entirely), `BLOCK-A1-01/02/09`, kill outbox, `UnmarshalEvent`,
  `Served() = cjson.With`, §4.3 **row 3**, the A1-7.6 `PayloadHash` definition
  *and* vector (the first writer's log claimed it applied; it was not), the
  PAIR-K1 resolved-row seek rule, §4.4 test index, PO-1/3/4/6/9. A2: renames
  finished in the examples, §4.2 fingerprint vectors, the fabricated finding
  `content_hash` replaced by the verified F1 value, the other three recomputed
  and labelled illustrative, `provenance` an array in all four examples with
  `verified` only on independent re-observation, §4.3 test index, PO-5…PO-9.
- **Integrator pass 1** (`9638165`): stripped 33 fix-plan scaffolding references
  (`BLOCK-*`, `PAIR-*`, the pair symbol, `plan §n`) out of normative text and
  added a *Review provenance* header row to each document so the `P-nn`/`C-nn`/
  `S-nn`/`F-nn`/`D-nn`/`T-nn`/`E-nn` citations resolve to `docs/reviews/` and
  are marked audit trail, not normative references.
- **Integrator pass 2** (`a7562ea`): merged the three duplicate PO questions in
  A1 §6 (6/15, 7/16, 9/17) and one in A2 §6 (3/12) by cross-reference — the item
  numbers are cited from normative clauses, so renumbering would have created
  dangling references; fixed a stale value (§6.7 still said the head hash is
  emitted every 1000 `seq`, A1-5.8 rules 100); replaced the last plan-internal
  ids in A1's traceability table with §6 item numbers.
- **Verification: 52 mechanical checks, 0 failures**
  (`docs/reviews/2026-09-11-verify-vectors.py`, committed as review evidence).
  Every published vector recomputed independently: A0-2.17 **V1–V8**, A1 §4.3
  **rows 0–3** (preimage lengths, `hash`, served lengths, served digests,
  `prev_hash` links from `ChainZero`, canonical stability, the three-name
  exclusion list), A1-7.6 `PayloadHash`, A2 §4.2 **F1/F1-R/F3/S1** incl. the
  "defective implementation" digest, and all four A2 example `content_hash`
  values rebuilt from each example's own 20-key `contentDoc`. Plus: 42 `Kind`
  consts, no stale counts, six sections per document, the six byte-identical
  paired blocks still identical, **no dangling clause reference among ~2650
  cross-document citations**, no raw U+2028/U+2029/U+007F in any table row.
- **D1 decided by the product owner at session close** (2026-09-11) and applied:
  `adr/ADR-0021-integrity-override-authority-and-lifetime.md` (**Proposed**)
  records admin-only authority + a **not single-use** `export` override whose
  residual risk is accepted and transferred to the service owner, compensated by
  logging. A1-6.5, A1-6.6, A1-3.3, A1-3.5, A1-4.2, §4.1, §4.4 and §6 item 15
  updated consistently (11 edits); `TestSingleUseOverride` replaced by
  `TestEveryReleaseUnderOneOverrideIsChained` +
  `TestReleaseWithoutLiveOverrideRefused`; the `(engagement_id,
  override_event_id)` uniqueness constraint is gone. Verification re-run: 52
  checks, 0 failures.
- **PR #2 opened and verified**: <https://github.com/daten-krake/Sleipnir/pull/2>
  — OPEN, base `main`, 13 files, +9056/−1, label `agent-built`, body carrying the
  review trail, the integrator verification, the **nine decisions D1–D9**, the
  46-item confirm checklist and the residual-risk list.

## Decisions

- **No new ADRs this session.** Every decision the reviews forced is either a
  contract clause (authority: ADR > SPEC > DESIGN > contract, and no clause
  contradicts an Accepted ADR) or a product-owner item in a document's §6.
- **D1 (integrity override) is DECIDED → ADR-0021 (Proposed, 2026-09-11):**
  authority admin-only (SPEC §3, *not* the insider argument — the product owner
  does not treat A15 as a v1 concern); lifetime **not single-use**, because a
  per-artifact rule does not survive the long-term service vision; the residual
  risk is accepted and transferred to the service owner who compensates with
  logging. This narrows Q11's literal "operator override", so it is recorded as
  an ADR rather than a contract note. It also makes **D5** (out-of-band head
  anchoring) load-bearing: the "the owner can see everything" bargain requires
  that the `artifact_released` trail cannot be truncated.
- **D2 still needs the product owner's signature** (A2-2.7/A2-5.6 implement Q2's
  "Finding carries confidence" as the provenance evidence grade instead of a
  finding field — a change to a locked decision). To be answered in PR #2; if
  signed it gets its own ADR (proposed number ADR-0022) and A2 §6.13 is
  rewritten from "PO signature required" to "decided".
- **AM-1 (`usr_`) resolved by default** (PR #2 item D3) — it was the single hard
  freeze blocker, claimed independently by A1 §6.2 and A2 §6.2. Registered in
  A0-1.2 + `KindUser`, *not* delegated to A5 (that would split A0-1.5 validation
  across two contracts).
- **Conflict precedence for contract review** (used by the fix plan, worth
  keeping): adversarial reviewer wins on safety semantics, principal wins on
  notation/buildability/naming, and where both are safety-relevant and
  incompatible the **stricter** option is applied and escalated to the product
  owner. Applied to 18 conflicts, e.g. `HeadLogIntervalSeq` 1000 → **100**,
  cursor direction bit → resolved-row seek rule, `confidence` new-node fork →
  bounded provenance list, `operator_release` → **deleted**.
- **Process decisions:** review reports and the fix plan are committed under
  `docs/reviews/` (precedent: `docs/adversarial-review-2026-09-03.md`), because
  clause rationale cites finding ids and a citation that cannot be followed is
  worthless; a child's deliverable is always a **file written incrementally**,
  never its final message (three runs were cut short and only files survived);
  a child's own progress log is treated as untrusted input and re-audited
  against the file (the A0 writer's log understated its progress by nine rows,
  the A1 writer's claimed an applied row that was absent).

## Open questions carried forward

- **PR #2 needs the product owner**: **D2–D9** (D1 is decided → ADR-0021, which
  the reviewer accepts by comment or merge) plus the 46-item confirm checklist.
  On answers, flip A0/A1/A2 `Draft` → **Frozen** in `contracts/README.md` and
  each document header (fix plan calls this WP-00).
- **Shared contract-test suite** — now seven categories (`contracts/README.md`:
  the safety-path pairing bullet is new) and ~60 test ids are named inside the
  clauses they guard. This is the merge gate for every implementing package.
- **First code packages**, in the order the principal review proposes in its §4:
  repo scaffold + `internal/errs` + `internal/logging` (DESIGN §1, ADR-0019),
  then `internal/ids`, `internal/cjson`, `internal/paging`, `internal/timex`,
  `internal/caps`, then `internal/events` and `internal/graph`.
- **A3–A8 fan-out briefs** (super-minimal, WORKFLOW §5). **A3 is blocked on D4**
  (the A0-7.10 view-cap composition rule); A5 inherits the `usr_` shape, the
  machine-principal exclusion list from A1-7.4/A1-8.4, the admin gate on
  `integrity_override`, and `engagement_assignment_changed` (D9's debt); A7 owns
  the action-spec document whose canonical bytes `action_spec_evidence_id` holds.
- **Residual risk carried** (PR #2 "Known debt"): engagement assignment and
  credential revocation unaudited until A5; no `evidence_removed` kind, so
  A1-8.8 is unimplementable until one exists; tail truncation open unless D5
  approves the webhook anchor; `cvss_v3_x10` has no range (P-77 — the only
  deferred item that is **not** additive-safe, needs an ADR); no approval-path
  JSON example in A1 §4.2.
- **Store-seam duties named by A1/A2 for backlog §6** (persistence): the
  per-engagement append lock primitive (A1-5.4), the dedup table (A1-7.6), the
  ingest watermark (A1-7.7), `chain_head_trail` and the kill outbox with
  `REVOKE UPDATE, DELETE` (A1-5.8, A1-7.12), and any checkpoint/re-genesis
  mechanism (A1-6.9 — needs its own ADR).
- Standing: Pi offline capability (backlog §4); AD technique scope for the v1
  tool registry; backlog §9–§11 (CI/CD, observability, model drift).

## Token usage

| total input | uncached input | cache read | cache write | output | reasoning |
|---|---|---|---|---|---|
| 25660457 | 379817 | 25280640 | 0 | 161546 | 65508 |

_(run `sessions/update-usage.sh sessions/2026-09-11-contract-freeze-a1.md` at session end)_
