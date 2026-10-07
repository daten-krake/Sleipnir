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

- (none yet — see `sessions/BACKLOG.md` standing questions and `next_steps.md`)

## Token usage

| total input | uncached input | cache read | cache write | output | reasoning |
|---|---|---|---|---|---|
| 566075 | 53801 | 512274 | 0 | 3484 | 5699 |

_(run `sessions/update-usage.sh sessions/2026-10-07-events-graph.md ses_eea8e38abffe7NBrfp4hw5mBxx` at session end — note: this branch still carries the old `.pi` script; until the `chore/opencode-config` PR merges, use that branch's OpenCode version: `git show chore/opencode-config:sessions/update-usage.sh`)_
