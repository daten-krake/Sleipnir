# Session 2026-10-07 — WP-09 `internal/events` + WP-14 `internal/graph`

- **Date:** 2026-10-07
- **Session id:** ses_eea8e38abffe7NBrfp4hw5mBxx
- **Model:** alibaba-token-plan/qwen3.8-max#xhigh
- **Goal:** Walking-skeleton session 2 (`docs/2026-09-29-walking-skeleton-plan.md`
  §7): deliver WP-09 `internal/events` (envelope, taxonomy, actor, all 42
  payload tables) and WP-14 `internal/graph` (kinds, `Node`/`Edge`/`Provenance`/
  `Attrs`, `contentDoc`) as two independently reviewed lanes against the frozen
  A1/A2 contracts. Housekeeping first: PR for `chore/opencode-config`, stale
  remote branch deleted (done), `gh` auth on the new macOS environment.

## Done this session

- **Housekeeping done:** verified all three 2026-09-29 PRs merged (#7/#8/#9),
  local `main` fast-forwarded to `321e319`, CI green on main (3/3 checks);
  deleted stale remote branch `sessions/close-2026-09-29`. Installed and
  authenticated `gh` via brew (new environment: macOS, was WSL; WORKFLOW §7's
  Windows-token recipe no longer applies — needs amendment to the plain
  `gh auth login` + `gh pr create` flow). Opened working branch
  `foundation/events-graph` from `main`.
- **PR #10 opened** for `chore/opencode-config` (`.pi` → OpenCode migration:
  agent definitions, start-session skill, `update-usage.sh`), label
  `agent-built`, URL verified: https://github.com/daten-krake/Sleipnir/pull/10.
  `next_steps.md` §0 is fully discharged.
- **P-82 is stale-good news:** `next_steps.md` warned `TestContentHashVector`
  needs a new normative vector published in A2 first — A2 §4.2 already carries
  F1/F1-R/F3/S1 (bytes, lens, digests, plus the `4a017e71…` defective-marker)
  and `verify-vectors.py` covers all four (PASS 52 / FAIL 0 re-verified this
  session). No contract edit needed for WP-14.
- **Two implementer lanes launched in parallel** (shared checkout, mechanical
  rules: own directory only, never `./...`): WP-09 `internal/events`
  (envelope, taxonomy, actor, 42 payload structs; seven named test ids; 42-not-39
  reconciliation carried into the brief) and WP-14 `internal/graph` (kinds,
  Node/Edge/Provenance/Attrs, contentDoc; nine named test ids; ADR-0022
  no-confidence rule and the stale `TestNodeDedupIncludesConfidence` name
  carried into the brief). Briefs pin the closed lists against the contract
  text, not the implementation (next_steps §3 definition of done).
- **WP-09 lane complete and committed** (`a1438b9`, 21 files, 2422 lines).
  Principal re-verified independently: gofmt/vet/build/test/`-race` green
  (`-count=1`), 42 kind consts with zero duplicates, 121 RUN lines (7 ids +
  114 subtests), `go list -deps` → zero internal imports, tests parse
  `contracts/A1-events.md` as the oracle, no `"time"` import in production
  files. Implementer's mutation proof: 8/8 kills on an out-of-repo copy
  (script preserved in the scratch dir).
- **Independent WP-09 review launched** (read-only, writing forbidden, scoped
  commands only, `git status` pinned before/after, own scratch extractor
  required so the oracle is not shared with the code under test).
- **WP-14 lane complete and committed** (`be357a5`, 12 files, 3505 lines).
  Principal re-verified independently: gofmt/vet/build/test/`-race` green
  (`-count=1`), `go list -deps` → exactly {errs, cjson, ids, timex}, and the
  three §4.2 published digests (F1/F3/S1, lens 656/656/304) recomputed from
  the contract's own byte strings with a separate python3+hashlib oracle —
  all match. Nine test ids, 97 subtests, closed lists and the 50-key reserved
  set parsed from the contract text.
- **The WP-14 brief contained a defect the lane correctly surfaced:** its
  "no `confidence` field anywhere" line overstates ADR-0022, which removed
  only the *finding-level* confidence — frozen A2 §4 declares
  `type Confidence string` (observed/inferred/verified) **on Provenance**,
  A2-5.3 requires the field, A2-6.3 reserves the key. The lane followed the
  contract over the brief (as briefed), shipped the provenance grade, has no
  finding-level confidence, and rejects the low/medium/high scale in tests.
- **Independent WP-14 review launched** (same read-only discipline; own
  extractor for the endpoint matrix, reserved-key set and vectors; must
  re-run and invent mutations).
- **WP-09 review: ACCEPT, no MUST FIX** (report: scratchpad
  `wp09-events-review.md`, 426 lines). Independent extractor: **1178 checks,
  0 failures**; enum pass 33/33; all 8 implementer mutations re-run (8/8
  killed) plus 9 reviewer-invented (8 killed, 1 survivor → NIT, now pinned);
  all five contract-defect claims confirmed from the text; P-25 signature
  deviation judged defensible; `TestTimestampFieldsAreStrings` behaviour
  confirmed covered by shipped subtests; `Validate()` deferral confirmed
  clean (no `var _ Payload`, nothing half-stubbed). Applied (`6568a4d`):
  doc.go dispositions for every row clause (SHOULD FIX-1), `ChainHead` ruled
  WP-10's (NIT-4), the non-nil `UntrustedFields` promise pinned with a
  regression assertion — mutation-proven killed on an out-of-repo copy
  (NIT-1). NIT-2/3 deferred to WP-11/20 with notes; NIT-5 no action.
- **WP-14 review: APPROVE, no MUST FIX** (report: scratchpad
  `wp14-graph-review.md`, 477 lines). Independently reproduced: all gates,
  vectors via own python oracle (incl. the defective marker as exactly the
  unsorted canonicalization), 1960-triple endpoint-matrix cross-check
  (exactly 48 allowed), 50-key reserved set, 12-field provenance shape, 45
  conditional-id probes, 37 attrs probes; 4 implementer mutation proofs
  re-run + 7 new package-side mutations, every one killed; all five
  implementer contract-disagreement claims confirmed (① the brief's
  "no confidence anywhere" overstated ADR-0022 — provenance grade is
  contract; ②–⑤ corroborate the errata the owner had already ruled
  mid-review). Applied (`9a02788`): **S1** attrs pass-2 rewrap now
  `errs.Wrapf` with an accurate message — chain survives `errors.Unwrap`,
  scratch-proven; **S2** rejected provenance values (principal_kind,
  confidence, failed ids) are never echoed — field + byte length only,
  regression-pinned by `rejected_values_are_never_echoed` with an
  `sk-live-` sentinel; **S3** doc.go names WP-15 as the owner of A2-5.7's
  caller-dependent internal/validation reclassification (type-level kinds
  are defaults); NITs 1–5 applied (comment accuracy, `./internal/graph`
  literal, dispositions for A2-1.1/1.3/1.9/5.5/6.4, `Provenance.Validate`
  and `walkAttrsStructure` split into named helpers preserving A2-5.3 table
  order, subtest rename); N6/N7 recorded, no change.
- **A2 errata 16–18 + the ADR-0022 fingerprint erratum note applied**
  (`311faac`) after the WP-14 review finished (oracle discipline): stale
  "(A2-local)" markers struck, sketch import header gains timex,
  ReservedAttrKeys shape note; ADR-0022 carries the erratum note (decision
  text unchanged, nothing superseded, ADR-0019 §3 precedent).
- **E5 ruled and applied** (`6d7d6d5`): `node_id` joins the A2-6.3 reserved
  set (50 → 51 keys, A2 §6 item 19), contract and package in lockstep, count
  pin updated. Ruling path: the question tool aborted; the owner's "go on"
  was taken as confirmation of the recommended option and is recorded here
  and in the PR body so it can be vetoed at PR review.
- **Full-tree gates green at session end:** `gofmt -l .` clean, `go vet ./...`,
  `go build ./...`, `go test ./...` and `-race` (10 packages) under
  `GOMEMLIMIT=512MiB`/`-timeout`, `go.mod` still **zero requires**,
  `verify-vectors.py` **PASS 52 / FAIL 0**, A1/A2 control-character scans
  clean (A1's two raw U+2028/29/7F are the pre-existing §4.2 example bytes in
  code blocks, not table rows).

## Decisions

- **WP-09 implementation rulings (principal, recorded in `internal/events/doc.go`):**
  `UntrustedFields(Kind) ([]string, bool)` — the bool separates "no starred
  fields" (23 kinds) from "unknown kind", fresh-copy return; `AllKinds()`
  exported (closedness untestable without an enumerator); `Payload.Validate()`
  declared-but-unimplemented by design (WP-11 owns it; a nil-returning stub
  would fake a safety path); `exit_code` is the only Go `int`, every other
  `:int` → `int64` per A1-4.12's literal rule; no cap constants in `events`
  (A1-4.5: the registry is `internal/caps`, WP-11 imports it); no constructors
  in this package (`NewEvent` is WP-11's per the §4.1 sketch).
- **A1 errata ruled by the product owner 2026-10-07 → A1 §6 items 19–23,
  applied and committed (`3218c30`):** ① stale sketch enum comment
  (`policy_changed` → the normative 4-value trigger list); ② §4.1 now
  declares `UntrustedFields(k Kind) (fields []string, ok bool)` (shipped
  two-value form supersedes P-25's proposal); ③ sketch's A1-local
  cap-constants block struck (A1-4.5 + `internal/caps`/A0-7.1 govern;
  WP-11 imports `caps`); ④ A1-4.12 int rule clarified — `int` only for
  `exit_code`, `attempt`/`http_status`/`containers_killed` stay `int64`;
  ⑤ `TestTimestampFieldsAreStrings` struck (registered nowhere; behaviour is
  `TestNoTimeTimeInCanonicalizedTypes`' shipped subtests; §4.4 stays the
  single registry). Gates after the edit: `verify-vectors.py` PASS 52 /
  FAIL 0, events suite green against the edited oracle, no control characters
  introduced. The stale "39" in the principal-review row/P-26 is handled per
  precedent: the dated review document stays untouched, the reconciliation
  travels in the tracker/next_steps and the WP-11 brief.
- **ADR-0022 erratum ruled by the product owner 2026-10-07 (application
  pending until the WP-14 reviewer finishes reading the ADR — do not move a
  reviewer's oracle mid-review):** an erratum note inside the Accepted ADR
  (ADR-0019 §3 precedent: decision text unchanged, nothing superseded)
  recording that the provenance grade lives in the append-only provenance
  entry, **outside** the content fingerprint: frozen A2-4.6 excludes
  provenance from the 20-key document, A2-4.7's dedup collapse appends
  provenance without changing `content_hash`, and §4.2's published vectors
  (independently recomputed this session) carry no `confidence` key. The
  bullet's intent — a grade cannot be silently changed — is discharged by
  provenance being append-only, not by digest inclusion. WP-14's shipped
  A2-4.6 behaviour stands.
- **A2 editorial errata ruled by the product owner 2026-10-07 (application
  pending, same reason):** ① A2-7.1's stale "(A2-local)" markers on the four
  constants the A0-7.1 registry/`internal/caps` owns (only
  `ProvenanceMaxEntries` is genuinely A2-local); ② §4 sketch import header
  gains `timex` (A2-5.3 needs A0-5.3; one-implementation principle); ③ §4
  sketch records that `ReservedAttrKeys` ships unexported (an exported
  package-level map is mutable cross-package state, DESIGN §4).
- **A0-2.3 UTF-8 field validation, events half → WP-11** (ingest validation):
  `cjson.CanonicalValue` cannot enforce it (encoding/json substitutes U+FFFD
  first), the WP-09 types-only package has no validation surface, and WP-11
  is the first code that rejects a caller-supplied string.
- **RESOLVED — ADR-0022 vs frozen A2-4.6 (ruled, see below):**
  ADR-0022's Decision (lines ~90–91) says the provenance grade is "**inside**
  the content fingerprint (A2-4.8), so a grade cannot be changed without a
  revision". Frozen A2-4.6 says provenance "is not part of the document at
  all, so none of them can influence the digest"; A2-4.7's dedup collapse
  appends provenance entries **without** changing `content_hash`; and §4.2's
  published vectors (independently recomputed this session) contain no
  `confidence` key. The two cannot both hold. WP-14 shipped A2-4.6's
  behaviour — the only reading under which the published vectors and the
  dedup semantics survive — and the product owner ruled 2026-10-07: an
  erratum note inside the Accepted ADR (ADR-0019 §3 precedent — decision
  text unchanged, no new ADR, nothing superseded); see the ruling bullet
  below.
- **A2 errata candidates surfaced by WP-14 (editorial, none blocking):**
  ① A2-7.1's "(A2-local)" markers on four constants already owned by the
  A0-7.1 registry/`internal/caps` (stale against AM-2's acceptance; A0-7.2
  second-definition hazard — not redefined in `graph`); ② §4 sketch's
  import-header comment omits `timex` although A2-5.3 timestamp validation
  needs A0-5.3 (one-implementation principle); ③ §4 sketch shows an exported
  `var ReservedAttrKeys` map — mutable package-level exported state, DESIGN §4
  forbids; shipped unexported `reservedAttrKeys`, same 50-key set.
- **A0-2.3 UTF-8 ownership, graph half:** WP-14's `CanonicalContent` owns the
  field-level UTF-8 gate for graph documents (reject-not-normalize,
  `validation`, names field + byte length, never echoes the value per A2-9.5
  discipline). Together with the events half (WP-11, above) this closes the
  BACKLOG 2026-09-24 "who validates UTF-8" question once WP-11 lands.
- **WP-15 handoff notes (from the WP-14 lane, recorded so they survive):**
  `AllowsEndpoints` (A2-3.2 as data), `ValidateProvenance`,
  `CanonicalContent`, in-package `reservedAttrKeys` are ready to consume;
  foreign `*json.SyntaxError` → A0-3 mapping at the request-decode boundary
  is WP-15's (paging precedent); A2-9.4's caller-side platform-minted
  exemption stays WP-15's; the stale row name
  `TestNodeDedupIncludesConfidence` should be renamed when WP-15 ships.

## Open questions carried forward

- **E5 veto window:** the `node_id` reservation (A2 §6 item 19) was ruled via
  the owner's "go on" after the question tool aborted — recorded as a
  confirmation of the recommended option. Vetoable at PR review; reverting
  means A2 §6 item 19 + the lockstep package/test edits (`6d7d6d5`).
- **WP-10 (next session):** `ChainHead` (A1-5.6) is WP-10's to declare
  (NIT-4 ruling); WP-12 consumes it. A1 §4.3's normative chain vector is the
  gate: `TestChainVectorDigests` byte-exact.
- **WP-11 handoff (events validation):** implement `Payload.Validate()` on
  all 42 structs + `NewEvent` (incl. A1-1.2's non-nil `EvidenceRefs` init) —
  until then no struct satisfies `Payload` and `Event.Payload` cannot be
  populated (deliberate, documented). Import `internal/caps` (never redeclare
  caps; the sketch block is struck — A1 §6 item 21). Own the events half of
  A0-2.3 field-level UTF-8 validation. Wire the A1-4.9 secret scan **with
  the A2-9.4/A1 §6 item 18 caller-side platform-minted exemption** (blocks
  otherwise: ~28 % of `slp_node_` ids rejected). `attempt`/`http_status`/
  `containers_killed` are `int64` (A1 §6 item 22). If WP-11 exports a
  production kind→payload registry, switch `payload_test.go`'s
  `allPayloadTypes` to it (drift risk noted by the implementer). NIT-2: pin
  the "enums are plain strings except BreakKind/VerifyTrigger" ruling with a
  `PkgPath()`-level assertion (here or WP-20). NIT-3: the 31 comment-only
  enum value lists get behavioural pins through WP-11's validation tables.
  `TestMaximalPayloadFitsCanonicalBound` runs over **42** kinds (the review
  row's "39" is stale — A1 §6 item context; row not edited, precedent).
- **WP-15 handoff (graph validation + write path):** own A2-5.7's
  caller-dependent internal/validation reclassification at A2-10.2 step 12
  (type-level kinds are defaults — S3); own the raw-body UTF-8 rejection at
  the request-decode boundary (H1: encoding/json normalizes invalid UTF-8 at
  decode, so wire strings arrive pre-normalized — without this an
  invalid-UTF-8 body yields a U+FFFD `content_hash` instead of a rejection);
  map foreign `*json.SyntaxError` to A0-3 kinds (H2 — else every malformed
  client body 500s; paging precedent); apply the 64-byte `ToolVersionMaxBytes`
  cap at step 8 (H3 — `Provenance.Validate` deliberately does not); wire the
  A2-9.4 caller-side platform-minted exemption; consume `AllowsEndpoints`,
  `ValidateProvenance`, `CanonicalContent`, in-package `reservedAttrKeys`;
  rename the stale row id `TestNodeDedupIncludesConfidence` when shipping
  (ADR-0022 — surface, don't implement a confidence field); discharge
  `TestCapsRejectWithSummaryTooLarge` (first caller of `caps.Fits`).
- **Still standing from BACKLOG (unchanged):** D10–D18 decision session
  (S8, 2026-10-23 — gates WP-23 onward; read D17 first); the worker-egress
  ADR (A3 CRIT, must exist before WP-29); A0 §4.1's k/id row-resolution half
  (WP-19/20); `secretscan`'s provisional planted corpus → shared suite
  (WP-19/20/21, `events`/`graph` import it, never the reverse); A0-7.10
  composition rule ↔ D11 contingency note.
- **Environment follow-ups:** WORKFLOW §7's PR recipe is WSL-era and needs
  amendment to the macOS/OpenCode flow (`gh auth login` + `gh pr create`) —
  proposed as a follow-up PR, not mixed into this one. AGENTS.md's
  subagent-artifacts path (`~/.pi/agent/sessions/...`) is stale under
  OpenCode; this session's child reports were routed through scratchpad files
  by brief instead — worth an AGENTS.md delegation-bullet update in the same
  follow-up.

## Token usage

| total input | uncached input | cache read | cache write | output | reasoning |
|---|---|---|---|---|---|
| 21310303 | 1143201 | 20167102 | 0 | 59710 | 58472 |

_(run `sessions/update-usage.sh sessions/2026-10-07-events-graph.md ses_eea8e38abffe7NBrfp4hw5mBxx` at session end — note: this branch still carries the old `.pi` script; until the `chore/opencode-config` PR merges, use that branch's OpenCode version: `git show chore/opencode-config:sessions/update-usage.sh`)_
