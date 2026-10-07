# NEXT STEPS — after session 2026-10-07

Goal of the next session: **WP-10 `internal/events` chain primitives** in
parallel with **WP-15 `internal/graph` validation + pure write path** —
session 3 of the walking-skeleton schedule
(`docs/2026-09-29-walking-skeleton-plan.md` §7, row S3, gate:
`TestChainVectorDigests` byte-exact against A1 §4.3). Both implement
frozen contract text; neither is blocked on D10–D18 (decision session S8,
2026-10-23, gates WP-23 onward only).

## 0. Housekeeping (do these first)

- [ ] **Two PRs from 2026-10-07**, both `agent-built`, both from
      `foundation/events-graph`-era branches off `main` @ `321e319`:
      - **PR #10** `chore/opencode-config` — the `.pi` → OpenCode migration
        (agent definitions, start-session skill, `update-usage.sh` on the
        OpenCode session API). Merging it makes the usage script on `main`
        the working one.
      - **PR #11** `foundation/events-graph` — WP-09 + WP-14 + both review
        applications + ten ruled errata (A1 §6 19–23, A2 §6 16–19, the
        ADR-0022 fingerprint erratum note) + the session records. **Note for
        the reviewer:** the E5 ruling (`node_id` reserved, A2 §6 item 19) was
        taken from the owner's "go on" after the question tool aborted —
        recorded as confirmation of the recommended option and explicitly
        vetoable in the PR; a veto means reverting `6d7d6d5`.
- [ ] After each merge: delete the branch locally **and** remotely,
      fast-forward local `main`, confirm CI green on `main`.
- [ ] **Follow-up PR (small, own branch):** WORKFLOW §7 still carries the
      WSL-era Windows-token recipe — replace with the macOS/OpenCode flow
      (`brew install gh`, `gh auth login`, `gh pr create … --label
      agent-built`, verified 2026-10-07). In the same or a sibling PR:
      AGENTS.md's delegation bullet still names `~/.pi/agent/sessions/…` for
      subagent-artifact recovery; under OpenCode the working pattern is
      routing every child's report through a scratchpad file named in its
      brief (`$TMPDIR/opencode/…`), which this session used successfully.
- [ ] CI triggers only on `pull_request` and pushes to `main`; a feature
      branch's first pipeline run is the PR that opens it.

## 1. Context to load

`start-session` ritual, fixed order: SPEC → `adr/` (README + **ADR-0022's new
erratum note** — the grade is OUTSIDE the content fingerprint; the rest
unchanged) → DESIGN → `sessions/BACKLOG.md` → AGENTS/WORKFLOW. **New since
2026-09-29:**

- `internal/events` (WP-09) and `internal/graph` (WP-14) — **their `doc.go`
  files are the entry points**, same convention as the other packages. Each
  lists clauses implemented, clauses deferred *with the owning WP*, and every
  ruling. WP-11/WP-15 briefs MUST read the sibling package's doc.go handoff
  notes before starting.
- Contract deltas: A1 §6 items **19–23**, A2 §6 items **16–19**, ADR-0022's
  erratum note. Consequential for builders:
  - **A1 §4.1 now declares `UntrustedFields(k Kind) (fields []string, ok
    bool)`** (two-value form; item 20) — WP-11 codes against it.
  - **A1's sketch cap-constant block is struck** (item 21): caps live only in
    `internal/caps`; WP-11 imports it, never redeclares.
  - **A1-4.12's int rule is "`int` only for `exit_code`"** (item 22):
    `attempt`, `http_status`, `containers_killed` are `int64`.
  - **A2-6.3's reserved set is 51 keys incl. `node_id`** (item 19).
  - **A2 §4 sketch imports timex** (item 17); `ReservedAttrKeys` ships
    unexported (item 18).
- `sessions/2026-10-07-events-graph.md` — one tracker, two packages, two
  independent reviews (ACCEPT + APPROVE, zero MUST FIX, zero shipped-logic
  defects), ten errata, and the full WP-10/11/15 handoff lists under *Open
  questions carried forward*.

## 2. Plan

1. **WP-10 `internal/events` — hash chain primitives.** A1-5.1…5.10 + the
   §4.3 normative chain vector. Files `internal/events/{chain.go,served.go}`.
   Test ids: `TestChainVectorDigests` (§4.3 rows 0–3: preimage bytes, lens,
   hashes, served lens+digests, `prev_hash` links — byte-exact, the session
   gate), `TestChainZeroIsSixtyFourZeros`, `TestChainSpecConstant`,
   `TestPreimageExcludesExactlyThreeNames`, `TestServedBytesReproducible`,
   `TestRecordedAtClampIsMonotone` (P-13, clock stepped backwards). Depends
   on WP-09 (merged). **Ruled:** `ChainHead` (A1-5.6) is WP-10's to declare;
   WP-12 consumes it. Note: no payload struct satisfies `Payload` until
   WP-11 — chain tests use test-local payload types or the raw canonical
   bytes; do NOT stub `Validate` (AGENTS.md safety-path rule).
2. **WP-15 `internal/graph` — validation + pure write path** (∥, disjoint
   files). A2-10.2 steps 3…11, A2-10.3…10.7, A2-7.1/7.2, A2-3.3…3.7,
   A2-4.3/4.5/4.7. Files `internal/graph/{validate.go,pending.go}`. Test ids:
   `TestValidationOrderIsDeterministic`, `TestEndpointMatrix` (one subtest per
   A2-3.2 row; consume `AllowsEndpoints`), `TestNotApplicableFieldRejected`,
   `TestCIDRMustParse`, `TestCapsRejectWithSummaryTooLarge` (the first caller
   of `caps.Fits` — discharges the carried debt), `TestReservedAttrKeysRejected`,
   `TestSelfEdgeRejected`, `TestSupersedeConstraints` (same kind, current
   target, no cycle, `MaxSupersedeChain` per P-53), `TestDuplicateEdgeCollapses`,
   `TestContradictsDirectionNormalized` (P-52). **The row's
   `TestNodeDedupIncludesConfidence` is stale (ADR-0022): surface it, rename
   it (e.g. `TestNodeDedupKeepsEveryObservation`), never implement a
   confidence field.** WP-15's debts from the 2026-10-07 reviews (all in the
   tracker's carried-forward list): the A2-9.4 caller-side platform-minted
   exemption (secretscan + ids.Valid; without it ~28 % of `slp_node_` ids are
   rejected); A2-5.7's caller-dependent internal/validation reclassification
   at A2-10.2 step 12 (type-level kinds from WP-14 are defaults); H1 the
   raw-body UTF-8 rejection at the request-decode boundary (encoding/json
   normalizes at decode); H2 foreign `*json.SyntaxError` → A0-3 kind mapping
   (paging precedent — else every malformed body 500s); H3 the 64-byte
   `ToolVersionMaxBytes` cap at step 8.
3. **One writer owns `contracts/`** — if both lanes surface errata, serialize
   them or split ownership (A1 → WP-10 lane, A2 → WP-15 lane) and let the
   principal merge. **Never edit a file a running reviewer is reading** —
   this session deferred the A2/ADR-0022 errata until the WP-14 review
   finished; keep that rule.
4. Then session 4: WP-11 events validation ∥ WP-16 quarantine (gate
   `TestQuarantineIsNeverCallerSettable`). WP-11's handoff list is long —
   read the tracker's carried-forward section before briefing it.
5. Design sessions still queued: **A3 stage views** (D11 interaction note),
   backlog §3 UI, §4 Pi node, §5 agent loop, §6 persistence, §7 security work
   items, §8 SSO, §9 CI remainder, §10 observability, §11 model benchmarking;
   **S8 (2026-10-23): D10–D18 + the worker-egress ADR** — unchanged, and
   still the one thing that can move the 2026-12-08 date.

## 3. Definition of done for the next session

- [ ] PRs #10 and #11 merged, branches deleted locally and remotely, local
      `main` fast-forwarded, CI green on `main`.
- [ ] WP-10 and WP-15 each independently reviewed and merged or opened as
      their own PR: every clause in their review row covered or dispositioned
      in doc.go, `TestChainVectorDigests` byte-exact against A1 §4.3, the
      A2-9.4 exemption implemented **and tested** (positive: a minted
      `slp_node_` id passes; negative: a high-entropy non-id string still
      fails), A2-5.7's reclassification tested for both caller contexts.
- [ ] `gofmt -l .`, `go vet ./...`, `go build ./...`, `go test ./...`,
      `go test -race ./...` green under `GOMEMLIMIT`/`-timeout`; `go.mod`
      still zero requires (pgx only once WP-22 lands, ADR-0010);
      `verify-vectors.py` still `PASS 52 / FAIL 0` (or higher with any new
      vector published and every published byte recomputed).
- [ ] Session tracker + `sessions/BACKLOG.md` + this file updated, usage
      refreshed, PRs labelled `agent-built`, PR URLs verified to exist.

## 4. Process lessons to keep (they cost real sessions to learn)

- **Do not move a reviewer's oracle mid-review.** Contract/ADR edits that a
  running reviewer reads as its source of truth wait until it reports (this
  session: A2/ADR-0022 errata applied only after the WP-14 review finished;
  A1 errata were safe because no active reviewer read A1).
- **A brief is untrusted input — and this session's own brief had the
  defect.** The WP-14 brief's "no `confidence` field anywhere" overstated
  ADR-0022 (which removes only the *finding-level* field; frozen A2 §4
  declares the provenance grade). The lane followed the contract, surfaced
  the conflict, and was right. Keep telling children to do exactly that.
- **Record how a ruling was obtained.** The question tool aborted and "go on"
  was taken as confirmation of the recommended option; the interpretation is
  written into the tracker, the §6 item's context and the PR body so the
  owner can veto at review. Never let an ambiguous confirmation become an
  unrecorded decision on a frozen contract.
- **Check a reviewer's evidence the way you check a child's code** (kept from
  2026-09-29): reproduce before acting; a symbol table is not a runtime
  value. This session both directions held: the reviews re-ran the
  implementers' mutations, and the principal re-ran both reviews' key claims
  (vectors via a third python oracle, the N9 survivor, the S1 chain, the E5
  evidence) before applying anything.
- **Pin the promise, not just the behaviour.** Two findings this session were
  documented-but-unpinned promises (the non-nil empty slice; the enum-typing
  ruling). If doc.go or a report promises a property, a test must kill the
  mutation that breaks it — `slices.Equal` treating nil == empty is how N9
  survived.
- **Echo discipline is a moving target — grep for `%q`/`%.40q` on rejected
  values in every new validation code.** S2 (provenance value echoes) is the
  third echo finding in three sessions; the rule is field name + byte length
  for anything that failed validation, everywhere (A2-5.8, A2-9.5, A0-3.4's
  bounds are a floor, not a license).
- **Two writers (and two reviewers) can share one checkout if the rule is
  mechanical**: only your own directory, never `./...`, `git status` before
  and after. Held across four children again this session.
- **A child's deliverable is a file written incrementally**, named in its
  brief (scratchpad path), never its final message; completion previews
  truncate.
- **Never prove a negative test by removing the bound and running it**
  (2026-09-21: 11.4 GB RSS, kernel OOM). Mutations only in copies outside
  the repo, capped, restored — both lanes and both reviewers obeyed.
- Verify every published vector mechanically before merging, and prefer an
  oracle the code under test did not produce: this session the A2 §4.2
  digests were computed three independent ways (implementer's python scratch,
  reviewer's own oracle, principal's third) before anyone trusted them.
