# Session 2026-09-21 — Contract freeze (WP-00) + the first line of code

- **Date:** 2026-09-21
- **Session id:** 01a0c2df-436a-7794-b45a-5aaf1c1516b0
- **Model:** qwen-token-plan-individual/qwen3.8-max
- **Goal:** Land WP-00 (A0/A1/A2 `Draft` → `Frozen`) as a follow-up commit in
  PR #2 now that the product owner has answered D1–D9 and the 46-item
  checklist, then write **the first product code**: repo scaffold +
  `internal/errs` + `internal/logging` with their contract tests.

## Product-owner answers this session (2026-09-21)

Recorded here because they are the authority for every edit below
(`contracts/README.md`: a frozen clause changes only by ADR **or an explicit
product owner decision recorded in a session tracker**).

- **D1–D9 and all three checklist blocks ticked in the PR #2 body** (edited
  2026-09-21T07:29:19Z; the PR still has 0 formal reviews — the body edit is
  the answer record). Ticking accepts the stated default for each.
- **D2 is SIGNED** — the provenance evidence grade `observed · inferred ·
  verified` on every node/edge replaces Q2's finding-level `confidence`.
  Changes a locked decision ⇒ **ADR-0022**.
- **D4 confirmed** — the interim fail-safe (where two A0-7.1 caps cannot both
  hold, the smaller governs and the builder truncates, mechanism T) stands.
  **A3 is unblocked**, but A3 still owes the real composition rule (~131 B per
  node is not a usable view).
- **D5 APPROVED, not declined** (asked explicitly; answer: "webhook is ok") —
  the out-of-band head anchor ships: push `(engagement_id, head_seq,
  head_hash)` over the ADR-0012 §3 signed webhook at every head-hash emission
  point. This is *added* contract text (A1-5.8), and it discharges ADR-0021's
  load-bearing follow-up: the `artifact_released` trail can no longer be
  truncated by an attacker with store + log access on one host without also
  contradicting an anchor the platform does not hold.
- **`cvss_v3_x10` stays UNDECLARED** (asked explicitly; answer: "undeclare").
  Accepted debt; declaring `[0,100]` later is **not additive-safe** (A0-6.5)
  and requires an ADR (A0-7.2).
- **CI backlog item = yes** — add `.github/workflows/` enforcement of the
  WORKFLOW §4 gates *before* the first code package lands, not after.
- **Nit §4.2 resolved by the PO's own tick**: the merge-gate box reads
  "follow-up commit **in this PR**", so the freeze lands inside PR #2 and there
  is no separate WP-00 PR.

## Branch

`contracts/a0-a2-conventions` (PR #2, head `ec69799`) for WP-00; a fresh
`<topic>/<short-slug>` branch for the code package once PR #2 is merged.

## Done this session

### WP-00 — the freeze (follow-up commit in PR #2)

- **A0/A1/A2 are `Frozen`.** Statuses flipped in `contracts/README.md` (3 rows)
  and each document's §1 header; A2's §1 `Depends on` row no longer calls A0
  "Draft at the time of writing".
- **All 51 marker sites de-asked** (A0 ×10 as briefed, but **12 real** — two wrap
  the phrase across a line break and were invisible to single-line grep; A1 ×25,
  one of them split as `PO\nconfirm`; A2 ×16). Every one now cites the decision
  instead of requesting it. **Zero** occurrences of `PO confirm` / `PO decision`
  / `PO signature` — or even a bare `PO` — survive in any of the three files,
  verified whitespace-normalized.
- **Each document's §6 is now a decision record**, headed by the same
  `**All items answered**` line in all three files.
- **D5 made normative, not just recorded** (the only decision that *added*
  contract text): A1-5.8 gains mitigation **(4)** — the platform MUST deliver
  `(engagement_id, head_seq, head_hash, integrity_state, chain_spec)` over the
  ADR-0012 §3 signed webhook at exactly the emission points of (2), to a
  recipient outside the deployment. Payload carries no event body, evidence or
  secret; the URL is never stored (A0-3.7), only `target_name`. Delivery is
  **best-effort and MUST NOT gate the safety path** — no webhook outage may
  stall the event log — and composing the delivery record MUST NOT itself
  trigger another anchor. A1-6.6's residual paragraph rewritten from "the PO
  declined" to **narrowed but not eliminated**; the closing "does not survive
  unrestricted write access" bullet now says the forgery is *detectable from
  outside*.
- **Additive taxonomy consequence of D5:** `notification_sent.notification_kind`
  gains a seventh value `chain_head_anchor` (ADR-0012 §6 already requires
  deliveries to be audited events, so an unchained anchor would contradict it).
  It is a **notification** kind, not an A1 `Kind` — the 42 `Kind` constants are
  untouched, verified. Updated consistently in A1-3.3 (enum + rationale note),
  A1-4.2 (payload rule), §6 item 14 (the additive-change record) and two §5
  traceability rows.
- **Three new test ids** for the anchor, keeping the README's safety-path
  pairing: `TestHeadAnchorWebhookCarriesTuple` (positive) +
  `TestWebhookAnchorFailureNeverBlocksAppend` and
  `TestAnchorDeliveryNeverTriggersAnotherAnchor` (negatives); mirrored into A1's
  §5 test-mapping row.
- **ADR-0021 → Accepted**, and its D5 follow-up rewritten from "load-bearing,
  if declined…" to **DISCHARGED**.
- **ADR-0022 written (new, Accepted)** — `adr/ADR-0022-provenance-grade-replaces-confidence.md`.
  D2's signature: the closed provenance grade `observed · inferred · verified`
  on every node and edge replaces Q2's `Finding.confidence`. Three options
  weighed; records that it closes P-42/S-05 (`verified` had been unreachable)
  and the fix plan's `confidence` new-node fork; follow-ups for A3 (budget bytes
  for the grade), the report/UI sessions (render the grade, never a re-invented
  adjective), the named test ids, and that a *numeric* confidence needs its own
  ADR.
- **A2 §6 items 1/13, A2-2.7 and A2-5.6** rewritten from "PO signature
  required" to the signature record citing ADR-0022.
- **`cvss_v3_x10` recorded as deliberately undeclared** in A2-2.5 and A2 §6.8,
  with the reason it is not additive-safe and that closing it later needs an ADR.
- **Integrator fixes on top of the writers' work** (five sites the A1 writer
  flagged as outside its brief, all legitimate): §6 item 2 was still written in
  pre-AM-1 tense ("A0-1.2 registers no prefix…", "until AM-1 lands,
  user-composed events MUST be refused") — false in a Frozen document, so it is
  now a RESOLVED record keeping the normative "a username or e-mail MUST NOT be
  used"; §6 item 15 no longer says ADR-0021 "is Proposed and awaits review";
  A1-2.2's actor table reads "resolved and confirmed"; A1-3.3's source cell no
  longer claims the enum is only "the six" ADR-0012 kinds; §5's ADR-0012 row
  names `chain_head_anchor` as A1's seventh value. Plus one A2 site my own audit
  caught that no grep in the brief could see: §6 item 11 still called D4 "the PO
  escalation".
- **Verification: two independent mechanical gates, both green.**
  `docs/reviews/2026-09-11-verify-vectors.py` → **PASS 52 / FAIL 0** (every
  published vector still byte-exact; `git diff -U0` proves **no** vector row,
  digest, `Kind = "` or `PayloadHash` line changed). A new 40-check integration
  audit (`/tmp/wp00-audit.py`) → **AUDIT PASS, 0 failures**: marker residue,
  status headers, ten cross-file paired sentences byte-identical, all twelve D5
  assertions, D2/D4/CVSS, six-section structure, provenance row, no raw
  U+2028/9/7F in table rows, 42 `Kind` constants, and `chain_head_anchor` proven
  *not* to have leaked into the `Kind` block.

### PR #2 merged, WP-01 delegated

- **PR #2 MERGED** 2026-09-21T08:33:33Z (merge commit `7ad14e2`, 18 files).
  A0/A1/A2 are `Frozen` **on `main`**, ADR-0021 and ADR-0022 are Accepted. Local
  `main` fast-forwarded; `contracts/a0-a2-conventions` deleted locally and on the
  remote. This is what unblocks code (`contracts/README.md`: no code against a
  `Draft`).
- Review note posted on the PR before the merge, recording how each tick was
  read and what the three explicit answers changed. The PO's body text was left
  untouched.
- **WP-01 cut as three super-minimal packages** (WORKFLOW §5) and delegated to
  `sleipnir-implementer`, one directory each: **WP-01.1** `internal/errs`
  (verify the draft, write `redact.go` + 18 named test ids) · **WP-01.2**
  `internal/logging` (3 files, 11 test ids including two that mechanically prove
  the DESIGN §1 zero-internal-imports boundary) · **WP-01.3**
  `.github/workflows/ci.yml`.

### WP-01 review pass, and the OOM that ended the session

- The three implementers reviewed **my** drafts and found real defects — which
  is what WORKFLOW §5's independent check is for. Two of them (the `slog`
  JSON-handler/`fmt`-verb redaction rule and the nil-logger fallback) are
  binding conventions, so they were promoted into `AGENTS.md` rather than left
  in a package comment.
- Review fixes were dispatched at 09:25:13Z as three one-writer-per-directory
  lanes: `fix-errs` (bound the `Error()` walk), `fix-logging`, `fix-ci`
  (dependency-allowlist findings). They did not complete.
- **Incident: the session was lost to a kernel OOM kill, not to an agent
  error.** At 11:36:58 CEST the `fix-errs` implementer, proving that
  `TestErrorChainWalkIsBounded` can fail, short-circuited the guard to
  `if false && depth == maxChainDepth { // BYPASS PROOF` in the working tree,
  backed the original up to `/tmp/errs.go.bak`, and ran the test. With the cap
  gone the cyclic-chain cases appended to a `strings.Builder` without bound:
  `errs.test` reached **11.4 GB anon-rss / 21 GB total-vm** on an 11 GiB WSL2
  VM, `all_unreclaimable? yes`, and the kernel killed it. The VM went down at
  11:37:54 (`init.scope: Stopping timed out. Killing.`) and both in-flight
  children took SIGTERM (`exit=143`). Reboot 11:40:48.
- **Recovery (this commit).** `/tmp` did not survive the reboot, so the backup
  was gone and the disabled guard was still in the tree — one `go test
  ./internal/errs/` away from repeating the crash. The cap is restored, no
  other mutation marker survives a repo-wide grep, and the gates were re-run
  under a hard cap (`ulimit -v 3G`, `GOMEMLIMIT=1GiB`, `-timeout 120s`):
  `gofmt -l`, `go vet ./...`, `go build ./...`, `go test ./...` all green
  (`errs` 0.006s, `logging` 0.005s).
- Lost with `/tmp`: the three lanes' scratch output (`/tmp/wp01/*-log.md`,
  the `fix-ci` negative/positive vector fixtures). The `fix-errs` and
  `fix-ci` lanes must be re-run; their findings are summarised above and in
  `AGENTS.md`, but the code changes were never applied.
- Rule added to `AGENTS.md`: a negative test is proved by asserting the
  bound's *effect*, never by removing the bound and executing the path.

### Process

- **Product-owner catch (2026-09-21): the principal drifted into implementer
  work.** After the merge I began typing `internal/errs` myself — four files,
  ~355 lines — instead of decomposing and delegating per WORKFLOW §5. Two
  defects, not one: the principal was doing an implementer's job, **and** the
  package I had briefed bundled four concerns (scaffold + `errs` + `logging` +
  CI), which is exactly what "super-minimal" forbids. Corrected by splitting
  WP-01 into three one-concern packages with disjoint directories.
- The drafted files were **not** laundered into done work: they are handed to
  the implementer explicitly labelled "a principal's draft, not accepted work",
  with instructions to verify every line against A0-3/ADR-0019 and fix or
  rewrite what is wrong, and to report defects as a success. This keeps the two
  architectural rulings I made (`New(kind, msg)`; `OpOf` existing so `logging`
  needs no internal import) under independent check instead of self-approved.
  Baseline before handover: `gofmt -l`, `go vet ./...`, `go build ./...` all
  clean, so any breakage is attributable.
- Session ritual completed: SPEC → `adr/` (README + ADR-0021, the only
  Proposed; other 20 Accepted) → DESIGN → BACKLOG → AGENTS/WORKFLOW.
- Housekeeping: tree clean, branch in sync with origin, local `main` ==
  `origin/main` == `7726564`; PR #2 OPEN/MERGEABLE, 16 files, label
  `agent-built` applied (its merge-gate box was stale-unticked); Linux `gh`
  still unauthenticated, WORKFLOW §7 token recipe verified working.
- Read the PO's PR #2 body edit and parsed all 9 decisions + 3 blocks.
- Sized WP-00: not 4 status lines but **51 live `PO confirm` / `PO decision` /
  `PO signature` markers** (A0 ×10, A1 ×25, A2 ×16) that are now answered and
  must cite the decision instead of requesting one (precedent: A1-6.5, which
  D1's application already rewrote as "product owner decision 2026-09-11").

## Decisions

- **A0/A1/A2 are Frozen** (this commit). A frozen clause now changes only by a
  new ADR or an explicit product-owner decision recorded in a session tracker,
  plus a PR (`contracts/README.md`).
- **ADR-0021 Accepted**; **ADR-0022 new and Accepted** (D2).
- **D5 approved ⇒ the out-of-band webhook head anchor is normative** (A1-5.8
  (4)). This discharges ADR-0021's load-bearing follow-up and narrows
  adversarial A11's tail-truncation residual from *open* to *detectable from
  outside*. It also adds `notification_kind:chain_head_anchor` — the first
  additive taxonomy change made after the reviews, and the reason a Frozen A1
  has seven notification kinds while ADR-0012 §2 lists six.
- **Attribute-level redaction seam** (new, → backlog §10): DESIGN §1 gives
  `logging` zero internal imports, so it cannot call `errs`' redaction helpers.
  Ruling for WP-01: `errs` owns redaction of error strings, `logging` takes only
  caller-supplied correlation ids, and the `api` handler is the one place a kind
  and a redacted attribute meet a log record. Enforcing redaction *inside*
  `logging` would need a DESIGN §1 amendment or a third foundation package.
- **`errs.New` takes the kind first** (`errs.New(kind, msg)`): ADR-0019 §1 writes
  `errs.New(msg)` but also requires kinds "attached at creation", and a kind
  parameter is the only way both hold. `Wrap`/`Wrapf` keep the ADR's signature
  and **inherit** the cause's kind, so wrapping can never silently reclassify an
  error. Recorded in WP-01's brief and to be recorded in the package doc comment.
- **CI rides in WP-01's PR**, not its own: it is the PR that first needs it, and
  the workflow's Go steps are guarded by `hashFiles('go.mod')` so doc-only and
  contract-only branches stay green. It gates the five WORKFLOW §4 commands, an
  **ADR-0010 dependency check that fails closed** on any non-pgx require, the
  contract-vector recomputation for PRs touching `contracts/`, and the I-02
  U+2028/9 check.
- **Attribute spelling: A0-3.6 beats ADR-0019 §3** (ruled for WP-01.2, to be
  ratified). ADR-0019 §3 names the correlation attributes `engagement`, `run`,
  `job`, `node`; frozen A0-3.6 rules `engagement_id`, `run_id`, `job_id`,
  `node_id` and says why — `node_id` is the *remote agent node* (`slp_node_`)
  while a graph node is `graph_node_id` (`gn_`), and the two MUST NOT be
  conflated. A0-3.6 already notes that ADR-0019 §3's "node" predates ADR-0016's
  graph vocabulary. Treated as a **naming clarification of an Accepted ADR by a
  frozen contract**, not a new decision, so no new ADR; the deviation is recorded
  in the `logging` package doc comment. **Open for the product owner:** whether
  ADR-0019 should carry an amendment note, since ADRs are immutable once
  Accepted and the literal reading still says `node`.
- **Brief defect found by a child, ruling issued mid-flight:** my brief's C.5
  tense fix re-inserted the literal token `**PO confirm**` that J.1 forbids. The
  A2 writer surfaced it instead of silently picking a side; its resolution is
  binding on all three files, and the replacement sentence is now a cross-file
  paired sentence verified byte-identical. Lesson: a mechanical definition of
  done (J.1) outranks illustrative wording in a brief, and writers should be
  told that explicitly.

## Open questions carried forward

- (append)

## Token usage

| total input | uncached input | cache read | cache write | output | reasoning |
|---|---|---|---|---|---|
| 0 | 0 | 0 | 0 | 0 | 0 |

_(run `sessions/update-usage.sh sessions/2026-09-21-freeze-and-first-code.md` at session end)_
