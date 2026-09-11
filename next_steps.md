# NEXT STEPS — after session 2026-09-11

Goal of the next session: **get PR #2 merged, flip A0–A2 to `Frozen`, and start
the first code package.** Still no product code until the contracts are frozen
(`contracts/README.md` lifecycle: *no code may be written against a Draft*).

## 0. Housekeeping (do these first)

- [ ] **PR #2** — <https://github.com/daten-krake/Sleipnir/pull/2>
      (`contracts/a0-a2-conventions` → `main`, 13 files, +9056/−1, label
      `agent-built`, verified OPEN at session close). The product owner must
      answer **decisions D1–D9** and the 46-item confirm checklist in the PR
      body; then the follow-up commit flips the statuses (WP-00 below).
- [ ] Branch is pushed and the tree was clean at session close
      (`a58ba46` + the session-record commit). `git status` / `git fetch` first;
      local `main` may need a fast-forward after the merge.
- [ ] `gh` on Linux is **unauthenticated**; the working recipe is WORKFLOW §7:
      `GH_TOKEN=$('/mnt/c/Program Files/GitHub CLI/gh.exe' auth token) gh …`.
      Never write the token to a file or a log (ADR-0019 §5).
- [ ] After the merge: delete the branch locally and on the remote, as in the
      2026-09-07 session.

## 1. Context to load

Use the `start-session` skill ritual (fixed read order: SPEC → adr/ → DESIGN →
`sessions/BACKLOG.md` → AGENTS/WORKFLOW). **New since the last session:**

- `contracts/A0-conventions.md` (1101 lines, 8 normative vectors),
  `contracts/A1-events.md` (3035 lines, 42 kinds, 4-row chain vector),
  `contracts/A2-graph.md` (1808 lines, 4-row fingerprint vector),
  `contracts/README.md` (7 merge-gate categories incl. **safety-path pairing**).
- `docs/reviews/2026-09-11-contract-review-principal.md` (105 findings),
  `…-contract-review-adversarial.md` (57 findings),
  `…-contract-fix-plan.md` (the triage: every MUST FIX assigned once, 17 paired
  rows, 18 resolved conflicts, §7 deferred list),
  `…-verify-vectors.py` (52 checks — run it after any contract edit).
- `sessions/2026-09-11-contract-freeze-a1.md` — this session's record,
  including the process lessons in §4 below.

Contract clause rationale may cite finding ids (`P-nn`, `C-nn`, `S-nn`, `F-nn`,
`D-nn`, `T-nn`, `E-nn`); each document's *Review provenance* header row says
where they resolve. They are audit trail, never normative references.

## 2. Plan once PR #2 is merged

1. **WP-00 — flip the freeze switch** (tiny PR): `Draft` → `Frozen` for A0/A1/A2
   in `contracts/README.md` **and** in each document's §1 header; record the
   D1–D9 answers in the session tracker (and in an ADR if D1 or D2 is signed as
   a change to Q11/Q2 — both narrow or change a locked decision, so an ADR is
   the honest record; propose ADR-0021 for whichever is signed).
2. **Shared contract-test suite** (`contracts/README.md`, seven categories). It
   is the merge gate for every implementing package, so it comes first. The ~60
   test ids are already named inside the clauses they guard; the vectors are
   published and verified (`docs/reviews/2026-09-11-verify-vectors.py` is the
   reference implementation of the arithmetic in Python).
3. **First code packages**, in the order the principal review's §4 proposes:
   repo scaffold + `internal/errs` + `internal/logging` (DESIGN §1, ADR-0019) →
   `internal/ids` (A0-1) → `internal/cjson` (A0-2, high-review: untrusted-input
   parsing, RFC vectors V1–V8 + the rejection list) → `internal/paging` (A0-4) →
   `internal/timex` (A0-5) → `internal/caps` (A0-7) → `internal/events` (A1) →
   `internal/graph` (A2). Each package = one super-minimal work package per
   WORKFLOW §5: the exact contract excerpt, files to touch, named acceptance
   tests, no cross-package assumptions.
4. **A3–A8 fan-out briefs** (super-minimal). **A3 is blocked on D4** — the
   A0-7.10 view-cap composition rule; do not brief A3 before that answer. A5
   inherits: the `usr_` shape (D3), the machine-principal exclusion list
   (A1-7.4/A1-8.4), the admin gate on `integrity_override` (D1), and
   `engagement_assignment_changed` (D9's debt). A7 owns the action-spec document
   whose canonical bytes `action_spec_evidence_id` holds, and the
   `fingerprint_hash`/`risk_tier` vocabulary A1 stores opaquely.
5. Design sessions still queued: backlog §9 CI/CD, §10 observability, §11 model
   benchmarking/drift.

## 3. Definition of done for the next session

- [ ] PR #2 merged, branch deleted, local `main` fast-forwarded.
- [ ] WP-00 merged: A0/A1/A2 read `Frozen`; D1–D9 answers recorded (ADR if a
      locked decision changed).
- [ ] Contract-test suite drafted as its own PR (or the first package landed
      with its tests) — `gofmt -l`, `go vet ./...`, `go build ./...`,
      `go test ./...` green, `go.mod` empty except pgx (ADR-0010).
- [ ] Session tracker + `sessions/BACKLOG.md` + this file updated, usage
      refreshed, PR labelled `agent-built`.

## 4. Process lessons to keep (they cost three failed runs to learn)

- **A child's deliverable must be a file written incrementally**, never its
  final message. Three runs were cut short by timeouts; only the files survived.
- Budget 45–60 min per child and require a skeleton on disk in the first 10
  minutes plus a "stop and make it coherent at minute N−10" rule.
- **Treat a child's own progress log as untrusted**: re-audit the file state
  before briefing the next pass. The A0 log understated progress by nine rows;
  the A1 log claimed a row applied that was absent.
- **One writer per file.** Three parallel writers on three documents worked;
  paired blocks had to be byte-identical and were verified as such afterwards.
- Verify every published vector mechanically before merging
  (`docs/reviews/2026-09-11-verify-vectors.py`), and check that normative
  documents stay tool-safe: raw U+2028/U+2029 in a Markdown table made the file's
  line count reader-dependent (integrator finding I-02).

## 5. Carry-forward open questions

- PR #2 **D1–D9** and the 46-item checklist (see the PR body).
- Residual risk accepted at the freeze: engagement-assignment and
  credential-revocation audit → A5; no `evidence_removed` kind, so A1-8.8 is
  unimplementable until one exists; tail truncation open unless D5 approves the
  signed-webhook head anchor; `cvss_v3_x10` has no range (P-77, **not**
  additive-safe — needs an ADR); no approval-path JSON example in A1 §4.2.
- Store-seam duties handed to backlog §6: append lock (A1-5.4), dedup table
  (A1-7.6), ingest watermark (A1-7.7), `chain_head_trail` + kill outbox with
  `REVOKE UPDATE, DELETE`, checkpoint/re-genesis (A1-6.9 → own ADR).
- Standing: Pi offline capability (backlog §4); AD technique scope for the v1
  tool registry; backlog §9–§11.
- `sessions/style-notes.md` is referenced by the 2026-09-07 quota snapshot but
  does not exist — either write it or drop the reference.
