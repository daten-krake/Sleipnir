# Session 2026-09-29 — WP-08 `internal/paging` + WP-13 `internal/secretscan`

- **Date:** 2026-09-29
- **Session id:** 01a0ed28-ade6-75eb-9874-c8ef1710d346
- **Model:** qwen-token-plan-individual/qwen3.8-max
- **Goal:** Deliver WP-08 `internal/paging` (A0-4, the one list envelope and its
  cursors) and WP-13 `internal/secretscan` (A2-9.4's closed rule table, A2-9.5,
  A1-4.9) as two independently reviewed lanes — plus, because the product owner
  asked "how far are we from the first HackTheBox test?", a walking-skeleton
  plan that gets there materially sooner than WP-00…WP-22 does.

## Done this session

- **Housekeeping closed.** PR #6 merged (`cb8305f`), branch deleted locally
  **and** remotely, local `main` fast-forwarded, CI green on `main`
  (run `36569416883`, 3/3 jobs). `next_steps.md` §0 discharged.
- **WP-08 `internal/paging` delivered** — 3 files, 1508 lines (182 doc + 246
  code + 1080 test), the §4 sketch's surface plus the two functions A0-4.5 and
  A0-4.6 force (`ParseLimit`, `NewPage`). Exactly the seven test ids §4.1 and
  the WP-08 review row name, everything else a subtest. Imports `errs`, `ids`,
  `cjson` — proved by `go list -deps`, and the A0 §4 preamble's "foundation
  packages import nothing internal but `errs`" is recorded as governing the
  *other* five, since the sketch itself declares this dependency.
- **WP-13 `internal/secretscan` delivered** — 3 files, 1053 lines (193 doc +
  153 code + 707 test). Exported surface is exactly
  `func Scan(field, value string) error`; the ten-row A2-9.4 table transcribed
  byte-exactly (verified twice: the implementer programmatically against the
  markdown with its escaped `\|` un-escaped, the reviewer with an independent
  extraction), table order, first match wins, imports `errs` only.
- **`internal/errs.callerOp` fixed** — it did not strip a generic
  instantiation's type-argument list, so every error from a generic function
  rendered `paging.NewPage[...]` instead of `paging.NewPage`, breaking
  ADR-0019 §2's `component.Function` shape and A0-3.4's greppability. Four
  lines plus `TestOpOfStripsGenericTypeArguments`. The brackets are cut
  **before** the import path, because a type argument can itself contain a `/`.
- **Two independent reviews** (`sleipnir-architect`, writing forbidden, tree
  verified clean afterwards): 11 findings on `paging`, 11 on `secretscan`.
  **Not one was a code defect in the shipped logic** — the pattern from
  2026-09-24 held: every finding was a citation's accuracy, a doc claim's
  truth, or a test's non-vacuity. All 22 applied or ruled on; both fix lanes
  proved their fixes by mutation **in copies outside the repo**.
- **Non-vacuity gaps found and closed.** `secretscan`'s suite hand-transcribed
  the rule-*id* column but never the *regexp* column, so narrowing
  `SEC-KRB {1,128}→{1,8}` passed all four tests; two SEC-ENTROPY knobs
  (threshold 4.5→4.4, a space added to the alphabet) were likewise unpinned.
  `paging` had six rejection rows asserting the stage substring `"canonical"`,
  which `DecodeCursor`'s own wrapper prose also contains, so they passed
  whichever step fired. All four classes now have mutation-proven assertions.
- **A real Go behaviour discovered, not assumed.** Go 1.27's
  `base64.RawURLEncoding` is **lenient about the final quantum's unused low
  bits**: for a payload of length 1 mod 3 there are 16 final-character
  spellings decoding to byte-identical bytes (4 ways at 2 mod 3). The stdlib
  therefore does *not* enforce A0-4.4's "replay byte-for-byte" on its own —
  `DecodeCursor`'s re-encode-and-compare step is what does, and the suite now
  walks the whole equivalence class and fails if its size changes.
- **Contract errata ruled and applied** (product owner, this session): A0 §6
  items **19–23**, A1 §6 item **18**, A2 §6 item **15**. Detail under
  *Decisions*.
- **Walking-skeleton plan written** —
  `docs/2026-09-29-walking-skeleton-plan.md`, 569 lines, opened as its own PR
  because it is a proposal, not a ruling. Answer to the owner's question:
  **~21 sessions, first HTB run attempt 2026-12-08, band 2026-11-24 →
  2027-01-15**, versus February 2027 on the current road — bought with nine
  named deferrals (D10–D18), one blocking ADR that does not exist yet (worker
  egress, adversarial finding A3, CRIT), and WP-23…WP-41. It states plainly
  that WP-22 is the end of the *planned* road and produces no binary at all.
- **Gates, all green under `GOMEMLIMIT=512MiB` and `-timeout`:** `gofmt -l .`
  clean, `go vet ./...`, `go build ./...`, `go test ./...` (8 packages),
  `go test -race ./...`, `verify-vectors.py` **PASS 52 / FAIL 0**, `go.mod`
  still at **zero requires** (ADR-0010 holds trivially), CI's table-row control
  character scan clean across all three edited contracts.

## Decisions

- **A0 §6 item 19 — ERRATUM.** §4's paging JSON example was internally
  inconsistent: its published `next_cursor` decoded to
  `{"id":"evt_01m1y2whfhp17g0avdqztd2p3","k":4711}`, a **25**-character id
  body, while `items[0].event_id` in the same example carried the
  **26**-character body A0-1.1 requires — so the frozen contract's own example
  cursor was not a valid A0-1.2 id. Found by the WP-08 implementer, who pinned
  the published bytes rather than "fixing" them; the principal recomputed both
  directions independently in Python, and the reviewer a third time with a
  separate implementation. The literal is corrected; the defective form survives
  in `paging`'s suite as the rejection fixture.
- **A0 §6 item 20 — ERRATUM.** (a) A0-4.4 wrote "base64url-unpadded
  (**A0-8.5**)"; A0-8.5 is the closed-enum `snake_case` rule and the
  base64url-unpadded rule is A0-8.6 — and `internal/paging` had inherited the
  mis-citation in five places. (b) A0-8.6's `Tests:` line named
  `TestEncodeCursorRoundTrip`, which §4.1 never registers, while the WP-08
  review row names `TestCursorRoundTripIsCanonicalBase64URL` for the same round
  trip: one test, two names, shipped under the WP-08 id so §4.1 stays the
  single registry.
- **A0 §6 item 21 — ERRATUM, adds two registry rows.** A0-4.5 has fixed `limit`
  at default 100 / maximum 1000 since the freeze, but A0-7.1 carried no row for
  them while A0-7.2 forbids a second definition of a registry value anywhere —
  so `paging.DefaultLimit`/`MaxLimit` were the only Go definition of a contract
  constant the registry did not know. The registry now carries
  `DefaultPageLimit = 100` and `MaxPageLimit = 1000` with mechanism
  "reject → `validation`", following the precedent of its only other numeric
  range (`ExitCodeMin`/`ExitCodeMax`). Values unchanged.
- **A0 §6 item 22 — ERRATUM, splits one test id across two owners.** §4.1
  registered `TestCursorWithInconsistentKAndIDRejected` as the **only** oracle
  for A0-4.4's k/id rule across A0, A1 and A2, but the rule has two halves and
  only one is checkable at decode time. `DecodeCursor(s, k ids.Kind)` can see
  only whether the cursor's `id` is of the declared kind; A0-4.8's "the id does
  not resolve in this collection, or its `k` disagrees with the ordering value
  of the row it resolves to" needs a collection and a row lookup, and
  `internal/paging` is a foundation package with no store dependency (DESIGN §1
  forbids foundation → service). §4.1 now names the split: `internal/paging`
  owns the decode-time half, WP-19/WP-20 the row-resolution half, WP-22
  exercises it against PostgreSQL. **Unclaimed, the second half would have been
  enforced nowhere while a traceability audit read the clause as covered.**
- **A0 §6 item 23 — ERRATUM.** Two test-name vocabularies guarded one rule:
  A0-3.4 named `TestSecretScanNamesFieldNotValue` and
  `TestNoSecretValueOrDigestInError` while A2-9.4, which owns the rule table,
  named `TestErrorMessageNamesFieldAndRuleIDOnly`. WP-13 ships A2-9.4's four;
  A0-3.4's two are now marked as shared-suite ids (WP-19/WP-20) naming the
  package-level oracle they duplicate.
- **A2 §6 item 15 / A1 §6 item 18 — ERRATUM, the session's most consequential
  ruling.** A2-9.4's `SEC-ENTROPY` rule **rejects the platform's own
  identifiers**. Measured against the real packages: of 20 000 ids from
  `ids.New(ids.AgentNode)`, **5 671 (28.4 %)** are rejected — a re-run by the
  WP-13 fix lane gave 24.9–26.3 %, so the rate is sampling-dependent and the
  class is not in doubt. Cause: `slp_node_` + 26 = 35 characters is the only
  A0-1.2 form reaching the rule's ≥ 32 window, and `_` is inside its alphabet
  `[A-Za-z0-9+/=_-]`, so a whole node id is one high-entropy run
  (`slp_node_01m3pqy611zczgaf4ajy4qkbxq` → H = 4.5147). The class is wider:
  container and network names built from an id body trip it too
  (`sleipnir-worker-<26>` H = 4.6241, `slp-run-<26>-net` H = 4.5568). Since
  A1's envelope key 6 *is* `node_id` and A1-4.9 mandates the scan over every
  string field of an append payload, **roughly one Pi-node event in four would
  have been rejected at ingest** — not the collateral §6 item 10 accepts for a
  bare 32-hex token in prose, but the platform rejecting its own primary key.
  Three remedies were measured and refused: window ≥ 40 exempts `slp_node_`
  (35) and `slp-run-…-net` (38) but **not** `sleipnir-worker-<body>` (42) and
  loses every 32–39 character secret; dropping `_`/`-` from the alphabet fixes
  all three shapes but blinds the rule to bare base64url blobs, which is what it
  exists for; accepting it breaks the remote-agent path (ADR-0013, SPEC §4.4).
  **Ruled: a platform-minted exemption, applied by the CALLER.** A caller skips
  `Scan` for a field whose value passes `ids.Valid(k, v)` for the A0-1.2 kind
  its contract declares, and for a field the platform stamps and A1-2.3 forbids
  a caller to supply. The rule table stays byte-identical, `Scan`'s signature is
  unchanged, and the exemption is unreachable by a hostile worker. **Blocks
  WP-11 and WP-15**, which must implement the caller side. That `secretscan`
  itself cannot implement it is enforced mechanically rather than by a test:
  doing so would need `ids.Valid`, and its import set is `errs` + stdlib.
- **`ParseLimit` accepts leading zeros** (`"0007"` → 7). A0-4.5's four failure
  classes (unparseable, ≤ 0, non-integer, > 1000) do not name them, and A0-2.6's
  canonical-number rule governs canonical documents, not query parameters.
  Ruled in `paging/doc.go`; the stricter reading is one clause away if the
  product owner ever wants it.
- **`DefaultLimit`/`MaxLimit` stay exported** although the tests are in-package:
  the walking-skeleton plan's D10/D11 freeze an A4-min `/api/v1` in which every
  list endpoint is paged under A0-4 with a hard-capped limit, so there is a
  named second consumer. `doc.go` records the reason so the next reviewer does
  not re-flag them as speculative.

## Process notes (they cost real time to learn)

- **A reviewer's evidence must be checked like a child's code.** The `paging`
  review's only MUST FIX claimed `errs.callerOp` renders the generic *shape
  body* — `paging.NewPage[go.shape.struct { ID string "json:\"id\""; … }]` —
  into every `NewPage` error, leaking struct tags into a client-facing envelope.
  It said so honestly ("evidence is the symbol dump, not a live render"), and it
  was wrong: Go 1.27 renders `paging.NewPage[...]`. The residual defect was real
  but a NIT, not a leak. The principal reproduced it live before acting, and the
  fix is one quarter of what the finding implied. **A symbol table is not a
  runtime value.**
- **A fix lane will find the fix you briefed is insufficient.** F10 asked the
  `secretscan` lane to grep the decoded `env.Error.Message` as an extra leak
  surface. It did, then reported that this is `err.Error()` verbatim today
  (`envelope.go:60`), so the line closed nothing — the escaped JSON/slog
  surfaces still missed a newline-bearing value's whole and last-8 forms. It
  added JSON-escaped forms too and proved the detector now catches a leak it
  previously missed. Briefing the *symptom* and letting the child find the
  *cause* beat briefing a specific line.
- **Two writers in one checkout works if the rule is mechanical.** Both
  implement lanes and both fix lanes shared this working tree with a hard
  "only your directory, never `./...`" rule, because `go test ./...` in one lane
  compiles the other's half-written package. It held across four children: no
  file was touched outside its owner, and both reviewers confirmed the tree with
  `git status` before and after. Cheaper than three worktrees and one branch
  reconciliation.
- **A frozen contract's own example is a fixture, and fixtures lie.** The
  published cursor literal was wrong by one character for 18 days and three
  reviews. Every child that touched it was told "the published literal wins,
  report a disagreement rather than adjusting the fixture" — which is why it
  surfaced as a defect instead of being silently normalized.

## Open questions carried forward

- **D10–D18 need a decision session** (`docs/2026-09-29-walking-skeleton-plan.md`
  §2.4). The plan proposes first HTB run 2026-12-08; every one of the nine is a
  deferral of a frozen clause, a locked Q-decision or an Accepted ADR, and none
  is assumed. D17 is the one to read first: it drops TOTP, so the account that
  approves attacks and fires the hard stop has no second factor.
- **The worker-egress ADR does not exist** (adversarial finding A3, CRIT). The
  plan's §5 recommends E1 — per-run bridge networks from a pre-provisioned
  pool, egress allowlist in **host** nftables installed once by `deploy/`, no
  container gets `NET_ADMIN` — with E2 behind one
  `Allocate(runID, scope) → networkName` seam. It must be accepted before the
  broker spawns anything, i.e. before WP-29.
- **The A2-9.4 platform-minted exemption has no test yet.** It is caller-side by
  ruling, so it lands with WP-11 (`internal/events`) and WP-15
  (`internal/graph`). Until then `secretscan` correctly rejects ~28 % of
  `slp_node_` ids and nothing exercises the exemption. Recorded in both
  packages' `doc.go` and in A1/A2 §6.
- **A0 §4.1's k/id split must be read before WP-19 starts.** §4.1 now names
  WP-19/WP-20 as the owner of A0-4.8's row-resolution half; if that suite is
  written without it, the clause is enforced nowhere.
- **`secretscan`'s planted corpus is provisional.** A2-9.4 says the corpus
  "lives in the shared suite", which is WP-19/20/21 and does not exist yet. The
  implementer's recommendation: move one-value-per-rule-id into the shared-suite
  package, have `events` and `graph` import it (never the reverse —
  `secretscan` must not import its own oracle), and shrink the local copy then.
  Three packages inventing their own corpus is the drift risk.
- **`cjson`'s carried-forward items are unchanged**: `CanonicalValue` cannot
  enforce A0-2.3's invalid-UTF-8 rule on Go values (field-level validation owns
  it, still no owner); an integral-valued float is byte-indistinguishable from
  an integer so A0-2.6 is review-only; `With` canonicalizes three times.
- **`TestIDOrderingMatchesByteOrderCollateC`** → WP-22 (`store/postgres`),
  unchanged. **`TestCapsRejectWithSummaryTooLarge`** → the first caller of
  `caps.Fits` (WP-11 or WP-15), unchanged.
- **A0-7.10's real composition rule is still A3's** — but the skeleton plan's
  D11 proposes *not writing A3 before the first run*, which would leave D4's
  composition rule unowned for longer. If D11 is accepted, that debt needs an
  explicit note in the plan's contingency list.
- **Attribute-level redaction seam** → backlog §10, unchanged. **CI deferred
  items** → backlog §9 (image build/pin/sign, full-SHA action pinning),
  unchanged.

## Token usage

| total input | uncached input | cache read | cache write | output | reasoning |
|---|---|---|---|---|---|
| 19664812 | 406316 | 19258496 | 0 | 120475 | 50126 |

_(run `sessions/update-usage.sh sessions/2026-09-29-paging-secretscan.md` at session end)_
