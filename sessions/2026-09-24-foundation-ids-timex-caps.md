# Session 2026-09-24 — A0 foundation packages: ids, timex, caps (WP-04/05/06)

- **Date:** 2026-09-24
- **Session id:** 01a0d430-3a10-7fc2-90a0-929f6dbca0b1
- **Model:** qwen-token-plan-individual/qwen3.8-max
- **Goal:** Land WP-04 `internal/ids` (A0-1), WP-05 `internal/timex` (A0-5) and
  WP-06 `internal/caps` (A0-7) with their named acceptance tests as one PR, and
  record the product owner's ruling on the ADR-0019 correlation-attribute
  spelling as an amendment note.

## Product-owner decisions this session (2026-09-24)

- **Session goal = WP-04/05/06** (`ids` + `timex` + `caps`), one PR. `cjson`
  (WP-07) stays out: it carries the high review bar and gets its own PR.
- **ADR-0019 gets an amendment note** for A0-3.6's attribute spelling
  (`engagement_id`/`run_id`/`job_id`/`node_id`, `graph_node_id` for graph
  nodes). The decision text is unchanged — this closes the standing backlog
  question by ruling it a clarification, recorded in the ADR itself so the
  literal reading of §3 no longer contradicts a frozen contract. No new ADR.
- **Execution model: three parallel implementer lanes, capped.** One writer per
  directory, cross-review lane afterwards, principal runs the gates under
  `ulimit -v` / `GOMEMLIMIT` / `-timeout` (AGENTS.md: never prove a negative by
  removing the bound).

## Housekeeping at session start

- Local `main` fast-forwarded `7ad14e2` → `4bdaec8` (was 7 behind); merged
  branch `layout/scaffold-errs-logging` deleted locally and on the remote.
- Gates re-run on `main`: `gofmt -l` clean, `go vet ./...`, `go build ./...`,
  `go test ./...` green (`errs` 0.006s, `logging` 0.004s); CI green on `main`
  (run `36025667809`, 2026-09-24T16:11:55Z). No open PRs. All 22 ADRs Accepted,
  none Proposed.
- **The 2026-09-21 session's close ritual was incomplete** (the OOM ended it):
  `next_steps.md` still says "after session 2026-09-11" and describes work that
  is now merged; that tracker's usage table is zeros and its "Open questions
  carried forward" is still `(append)`. Both are corrected at this session's
  close.

## Branch

`foundation/ids-timex-caps` (from `main` @ `4bdaec8`).

## Done this session

- **ADR-0019 carries the amendment note** (PO ruling above): §3's
  correlation-attribute list now points at frozen A0-3.6 for the spelling, with
  the note marked "spelling only, decision unchanged" and the Status line
  saying the ADR is not superseded. Verified against the shipped code before
  writing it: `internal/logging` emits `engagement_id`/`run_id`/`job_id`/
  `node_id` and has no graph-node parameter, `errs.Attrs` carries
  `graph_node_id`. Commit `ccce912`.
- **The 2026-09-21 close ritual completed retroactively** (commit `fcbc5cc`):
  usage recomputed from that session's log (16 530 624 total input — a lower
  bound, the log ends at the OOM kill) and its "Open questions carried
  forward" transcribed from its own Done/Decisions instead of the `(append)`
  placeholder.
- **WP-04/05/06 dispatched** as three parallel `sleipnir-implementer` lanes
  with a read-only `reviewer` stage behind each (`runs.lanes`, one writer per
  directory, shared checkout). Every brief quotes its clause range in
  `contracts/A0-conventions.md` as the only contract input, names the exact
  files, the exact test ids and what is out of scope, and carries the same
  binding rules block: capped test runs (`ulimit -v 2G`, `GOMEMLIMIT=1GiB`,
  `-timeout 60s`, own package only), never `go test ./...` while siblings
  write, and never proving a negative by disabling a bound (AGENTS.md).
- **A0-7.5 erratum landed** (PO ruling below): the truncation marker is 11 B,
  not 12 B, corrected at the clause and in the §4 sketch comment, plus a new
  §6 item 16 recording it. `docs/reviews/2026-09-11-verify-vectors.py` still
  reports **PASS 52 / FAIL 0** and no raw U+2028/9/7F entered the file, so the
  CI `contracts` job stays green on a branch that touched `contracts/`.
- **WP-04/05/06 delivered**: `internal/ids` (788 lines), `internal/timex`
  (581), `internal/caps` (676) — 9 files, 2045 lines, **15 top-level test ids
  and exactly those 15** (6 + 4 + 5, verbatim from A0 §4.1 and the principal
  review's WP-04/05/06 rows; everything else is a subtest), 239 passing
  tests/subtests. Import boundary verified mechanically rather than trusted:
  `go list -deps` shows `caps` with **zero** internal imports (it returns no
  errors, so it needs no `errs`) and `ids`/`timex` importing only
  `internal/errs`, per the A0 §4 preamble. `go.mod` still has zero requires.
- **Gates green, capped** (`ulimit -v 3G`, `GOMEMLIMIT=1GiB`, `-timeout`):
  `gofmt -l` clean, `go vet ./...`, `go build ./...`, `go test ./...` and
  `go test -race ./...` all pass on all five packages. After three contract
  edits, `docs/reviews/2026-09-11-verify-vectors.py` still reports **PASS 52 /
  FAIL 0** and no raw U+2028/9/7F entered the file.
- **Independent review per package, then the fixes applied by the principal**
  (commits `555ced2`+`a852f07` folded, `76cd3d1`+`a0ebed0` separate):
  `ids` 7 findings (1 SHOULD FIX: the crypto/rand import check was an
  allow-list only, so the positive half of A0-1.4's pair was vacuous), `caps` 6
  (3 SHOULD FIX: a fixture named `mechanismTCaps` restated a per-field
  mechanism registry the package deliberately does not own, three passages
  called the erratum still outstanding, and the `limit == len(marker)+1`
  boundary — the only path where the rune walk-back reaches zero with a
  non-zero budget — was untested), `timex` 5 (1 MUST FIX, doc-only). **No
  finding in any of the three was a code defect**: every one was a test's
  non-vacuity, a comment's accuracy or a coverage gap in a package doc.
- **The implementers found four defects in the principal's own input**, which
  is what WORKFLOW §5's independent check is for: the marker's byte count
  (contract, → A0 §6 item 16), Go's `.000` directive *truncating* rather than
  rounding sub-millisecond digits (brief — re-verified here), a demanded
  pre-1970 *positive* timestamp row that A0-5.3's `[2020, 2100)` window makes
  impossible (brief — implemented as a rejection row plus a formatting row),
  and A0-5.3's leap-second rationale plus §4.1's unnamed owner for the two
  clamp test ids (contract, → A0 §6 item 17).
- **One review lane lost, and the cause is a tooling mismatch, not an agent
  error.** The builtin `reviewer` has no shell: two of the three review lanes
  reported "not run by me, the supervisor must run the gates", and the third
  spent four minutes in `find /` and had to be interrupted, its report
  unrecoverable. Re-run with a brief that states the absence of a shell and
  forbids searching outside the repo, it completed clean. Both review
  conventions are now in `AGENTS.md`.
- **Child reports were recovered from the session transcripts, not the
  completion previews** (which truncate mid-finding): the durable copy of each
  child's final message is
  `~/.pi/agent/sessions/<parent>/<child-run>/run-0/session.jsonl`. All five
  reports were extracted before any fix was applied, so no finding was acted on
  from a truncated preview.

## Decisions

- **The `[truncated]` literal governs; A0-7.5's "12 B" was an arithmetic
  error** (PO decision 2026-09-24, recorded as A0 §6 item 16). Found by the
  WP-06 implementer, which surfaced it instead of silently picking a side —
  the same behaviour the 2026-09-21 brief-defect catch established. Ruled an
  **erratum, not an ADR**: no cap value, mechanism or vector changes, and
  A0-7.2 governs the constants, not that prose. `caps` derives every boundary
  from `len(TruncationMarker)` rather than a hardcoded number, so a future
  literal change by ADR cannot leave a stale constant behind.
- **`TestTruncationMarkerIsTwelveBytes` is renamed
  `TestTruncationMarkerIsCountedAgainstTheCap`** (principal ruling). The id
  came from the principal review report's WP-06 row, not from the frozen A0
  §4.1 registry — which lists only `TestTruncateRuneBoundary` and
  `TestNoSilentTruncationInHashedRecords` for A0-7.4/7.5 — so renaming it is
  not a contract change, and a test whose name asserts a wrong byte count is a
  trap for the next agent. Review reports are audit trail, never normative.
- **`internal/timex` does not carry A0-5.4's `recorded_at` clamp.** The clause
  assigns the clamp to the writer, the principal review assigns
  `TestRecordedAtClampIsMonotone`/`…BypassRejected` to the event-chain package
  (WP-10), and the error-level log with correlation attributes needs a logger
  `timex` must not import (DESIGN §1: foundation packages import only `errs`).
  No stub, no constant, no helper — the consumer does not exist yet
  (DESIGN §2). `timex/doc.go` must name the clauses it serves but does not
  implement.
- **`internal/caps` gets no mechanism-R error helper.** A0-7.6's
  `summary_too_large` rejection names field, cap and actual, and both `events`
  (WP-11) and `graph` (WP-15) will need it — but neither exists yet, so this
  package ships constants plus the two measurement/truncation primitives the
  A0 §4 sketch names, and `doc.go` records the one-line helper
  (`errs.SummaryTooLarge`) as the follow-up for whichever package needs it
  second. Extract on the third use, not the first (DESIGN §2).
- **`ids` owns only the classification half of A0-1.4.** "A uniqueness
  violation at insert MUST surface as `internal`" is a store rule (WP-22);
  `TestUniquenessViolationIsInternal` in `internal/ids` therefore proves what
  this package can: a failing entropy reader is attempted **exactly once** (no
  retry loop) and the error's kind is `errs.Internal` with the reader's error
  as its cause. `doc.go` says which half lives where.
- **A0-5.3's rationale and §4.1's clamp-id ownership corrected** (PO decision
  2026-09-24, recorded as A0 §6 item 17). `time.Parse` does not normalize leap
  seconds — it rejects `:60` with `second out of range`, re-verified here on Go
  1.27 — so the clause now gives the one true hazard (`Z07:00` accepts
  `+02:00`) and keeps the separate seconds range check for the reason that
  survives: it names the rejection class. §4.1's naming-rulings paragraph now
  says the two clamp ids belong to the event-chain package (WP-10), not to
  `internal/timex`, because A0-5.4 assigns the clamp to the writer.
- **A documented exception must be true** (timex finding 2, principal ruling).
  `Now(nil)` still panics — the §4 sketch fixes the signature, a `time.Now`
  fallback would violate A0-5.4's one-injected-clock rule and put an
  unreproducible timestamp on an audit chain, and the panic is pinned by a
  test. What was wrong is the *premise*: the doc argued no request can reach
  `Now(nil)` because the clock is wired at the `cmd/` edge and boot would crash
  first, which is false — `Now` takes the clock as a parameter, nothing
  dereferences it at boot, and a half-built caller struct reaches it inside a
  request. Reworded to the true ground (no client input can make a `Clock` nil;
  the only route in is a half-built value, which DESIGN §4 forbids; the panic
  is fail-loud, not error transport). A future agent reads that sentence to
  decide whether to add a fallback, so a false premise there is a latent
  defect even though the behaviour was right.
- **A test fixture must not restate a registry the package deliberately does
  not own** (caps finding 1). `mechanismTCaps` listed 15 caps as "the entries a
  caller applies mechanism T to" while A0-7.1's Mechanism column assigns 14 of
  them **R** — a second, wrong source of truth that the events and graph lanes
  would have read, in the one package whose doc says no such registry exists.
  Renamed `byteCaps` with the comment stating what it actually proves (every
  byte cap is ≥ `len(TruncationMarker)`, so mechanism T is well-defined for it)
  and who assigns mechanisms (A0-7.7, the owning contract).
- **A non-vacuity gap is a finding even when the test passes** (ids finding 1,
  caps finding 3). An import allow-list that never *requires* `crypto/rand`
  stays green if `New` switches to a deterministic reader, and a truncation
  corpus whose every row leaves budget ≥ 4 stays green if the rune walk-back's
  `cut > 0` becomes `cut > 1` and returns a continuation byte. Both halves are
  now asserted, and the two new normalization rows in `ids` carry a subtest
  pinning that their body is exactly 26 bytes — otherwise they would die in the
  length check and never reach the alphabet scan they exist to test.

## Open questions carried forward

- **Next code: WP-07 `internal/cjson`** (A0-2, the high-review-bar package:
  six vectors + V7 + the rejection list), then **WP-08 `internal/paging`**
  (needs `ids` + `cjson`), and **WP-13 `internal/secretscan`** which is
  parallel to both (it imports only `errs`). After those, WP-09…WP-12
  (`events`) and WP-14…WP-16 (`graph`), then the shared contract-test suite
  (WP-19 is A0's owner package and needs `cjson` + `paging`).
- **`TestIDOrderingMatchesByteOrderCollateC` has no owner until WP-22.** A0
  §4.1 registers it against A0-1.9, it is a PostgreSQL integration test
  (`COLLATE "C"` declared on the column, opt-in per DESIGN §8), and `ids`
  deliberately does not ship it as a skipped test. Recorded in backlog §6 as an
  explicit `store/postgres` acceptance item so the A0-1.9 pairing is not lost.
- **`TestCapsRejectWithSummaryTooLarge` has no owner either.** A0-7.1's own
  "Tests:" line names it, but mechanism R is the ingest path's (A0-7.6/7.7) and
  `caps` returns no errors. It belongs to the first caller of `caps.Fits`
  (WP-11 `events` or WP-15 `graph`), and the one-line rejection helper belongs
  in `caps` on the **third** use (DESIGN §2), not the first.
- **A0-7.10's real composition rule is still A3's** (interim fail-safe frozen;
  ~131 B per node is not a usable view) and must budget bytes for ADR-0022's
  provenance grade.
- **Attribute-level redaction seam** → backlog §10, unchanged.
- **CI deferred items** → backlog §9: container image build/pin/sign (A10) and
  full-SHA action pinning.
- **Process, for the next session that fans out:** brief read-only reviewers
  with their tool limits stated (no shell ⇒ the principal runs the gates; never
  search outside the repo), and recover a child's full report from its session
  jsonl because completion previews truncate mid-finding. Both are now
  `AGENTS.md` conventions.

## Token usage

| total input | uncached input | cache read | cache write | output | reasoning |
|---|---|---|---|---|---|
| 12553822 | 425950 | 12127872 | 0 | 95918 | 45652 |

_(run `sessions/update-usage.sh sessions/2026-09-24-foundation-ids-timex-caps.md` at session end)_
