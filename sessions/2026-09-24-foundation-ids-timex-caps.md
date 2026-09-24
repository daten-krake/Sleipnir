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

## Open questions carried forward

- (append)

## Token usage

| total input | uncached input | cache read | cache write | output | reasoning |
|---|---|---|---|---|---|
| 789074 | 134098 | 654976 | 0 | 10165 | 5409 |

_(run `sessions/update-usage.sh sessions/2026-09-24-foundation-ids-timex-caps.md` at session end)_
