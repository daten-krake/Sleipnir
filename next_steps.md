# NEXT STEPS — after session 2026-09-29

Goal of the next session: **WP-09 `internal/events` (envelope, taxonomy, actor,
payload structs)** in parallel with **WP-14 `internal/graph` (kinds,
`Node`/`Edge`/`Provenance`/`Attrs`, `contentDoc`)** — session 2 of the
walking-skeleton schedule (`docs/2026-09-29-walking-skeleton-plan.md` §7).
Both implement already-frozen contract text, so **neither is blocked on
D10–D18**; the decision session is S8 (2026-10-23) and only gates WP-23
onward.

If the product owner would rather decide the roadmap first, the alternative next
session is the **D10–D18 + worker-egress ADR decision session** — it is cheap,
it unblocks nothing before WP-23, and it is the one thing that can move the
2026-12-08 date.

## 0. Housekeeping (do these first)

- [ ] **Three PRs from 2026-09-29**, all `agent-built`, all branching from
      `main` @ `cb8305f` and touching disjoint files, so they merge in any
      order:
      - `foundation/paging-secretscan` — WP-08 + WP-13 + the `errs.callerOp`
        generic-strip fix + the six contract errata (A0 §6 items 19–23, A1 §6
        item 18, A2 §6 item 15).
      - `plan/walking-skeleton` — `docs/2026-09-29-walking-skeleton-plan.md`
        only. A **proposal**, not a ruling: nothing in it is decided until
        D10–D18 are answered.
      - `sessions/2026-09-29-paging-secretscan` — the tracker, `BACKLOG.md`,
        this file.
- [ ] After each merge: delete the branch locally **and** remotely, fast-forward
      local `main`, and confirm CI went green on `main`.
- [ ] `gh` on Linux is still unauthenticated; WORKFLOW §7's recipe works:
      `GH_TOKEN=$('/mnt/c/Program Files/GitHub CLI/gh.exe' auth token) gh …`.
      Never write the token to a file or a log (ADR-0019 §5).
- [ ] CI triggers only on `pull_request` and on pushes to `main`, so a feature
      branch's first pipeline run is the PR that opens it.

## 1. Context to load

`start-session` ritual, fixed order: SPEC → `adr/` (README + anything changed
since 2026-09-29 — **nothing in `adr/` changed**; the ADR-0019 §3 amendment
note of 2026-09-24 still governs correlation-attribute spelling) → DESIGN →
`sessions/BACKLOG.md` → AGENTS/WORKFLOW. **New since the last session:**

- `internal/paging`, `internal/secretscan` — and **their `doc.go` files are the
  entry point**, same convention as `cjson`. Each lists the clauses it
  implements, the clauses it deliberately does not *with the package that owns
  them*, and its rulings. Read `secretscan`'s before wiring a scan anywhere: the
  A2-9.4 platform-minted exemption is **caller-side** and this package cannot
  implement it (it would need an `ids` import; `go list -deps` is the guard).
- `contracts/A0-conventions.md` §6 items **19–23**, `contracts/A1-events.md` §6
  item **18**, `contracts/A2-graph.md` §6 item **15** — the six errata ruled
  2026-09-29. Two of them change what a builder must do:
  - **§4.1's k/id split**: `TestCursorWithInconsistentKAndIDRejected` covers
    only the decode-time half. WP-19/WP-20 own A0-4.8's row-resolution half.
  - **A2-9.4's platform-minted exemption**: WP-11 and WP-15 MUST skip `Scan`
    for a value passing `ids.Valid` for the field's declared kind and for A1-2.3
    platform-stamped fields. Without it ~28 % of `slp_node_` ids are rejected
    and roughly one Pi-node event in four fails ingest.
- `contracts/A0-conventions.md` A0-7.1 now carries `DefaultPageLimit = 100` and
  `MaxPageLimit = 1000`; A0-4.5 cites them; `internal/paging` stays their one Go
  definition (A0-7.2).
- `sessions/2026-09-29-paging-secretscan.md` — one tracker, two packages, 22
  review findings of which **zero** were code defects in the shipped logic, and
  six errata.
- `docs/2026-09-29-walking-skeleton-plan.md` — the roadmap answer. §2.4 is the
  D10–D18 decision request, §4 is WP-23…WP-41, §5 is the egress ADR options,
  §7.2 is the falsifiable definition of done for "first HTB run".

## 2. Plan

1. **WP-09 `internal/events`** — envelope, taxonomy, actor, payload structs.
   A1-1.1…1.8, A1-2.1…2.8, A1-3.1…3.8 (**all 42 payload tables**),
   A1-4.1/4.8/4.12, A1 §4.1. Files `internal/events/{kind.go,actor.go,event.go,
   payload_*.go}`. Test ids: `TestEnvelopeKeySetIsSeventeen`,
   **`TestKindListIs42AndClosed`**, `TestActorComponentIsEmptyForNonPlatform`,
   `TestPayloadStructsMatchA1Tables`, `TestPayloadsAreFlat`,
   `TestNoTimeTimeInCanonicalizedTypes`,
   `TestUntrustedFieldsCoversEveryStarredField`. Depends on `ids`, `timex`,
   `cjson` — all merged.
   **Already reconciled, do not re-litigate:** the 2026-09-11 principal review's
   WP-09 row says "all 39 tables" and `TestKindListIs39AndClosed`, and it is
   **stale**. A1-3.3 has 42 kind rows (counted 2026-09-29), A1 §4.1's sketch
   says "the closed event taxonomy (A1-3.1, 42 kinds)", and A1's own §5
   traceability row names `TestKindListIs42AndClosed` plus "three new `Kind`
   consts, 42 total" — the three lifecycle kinds (T-01/T-02) were added after
   the review was written. **42 is the number**; A1-4.5's
   `TestMaximalPayloadFitsCanonicalBound` also runs over 42 kinds.
2. **WP-14 `internal/graph`** (∥, disjoint files) — kinds, `Node`/`Edge`/
   `Provenance`/`Attrs`, `contentDoc`. A2-1…A2-6 per its review row. Files
   `internal/graph/{kind.go,node.go,edge.go,provenance.go,attrs.go,content.go}`.
   Test ids include `TestContentHashVector`, which the principal review flags
   **P-82: needs a new normative vector** — that vector must be published in A2
   and added to `verify-vectors.py` before the test can be written, so this lane
   may need a contract edit. **One writer owns `contracts/`** — if both lanes
   need errata, serialize them or give A2 to WP-14 and A1 to WP-09 and let the
   principal merge.
   Note `TestNodeDedupIncludesConfidence` in that row: ADR-0022 **removed**
   finding-level confidence, so the row's name is stale — surface it rather than
   implementing a `confidence` field.
3. Then WP-10 ∥ WP-15, WP-11 ∥ WP-16, WP-12 ∥ WP-17, WP-18 ∥ WP-19/20/21,
   WP-22 ∥ WP-31 — the schedule in the plan's §7 is the queue, and each row
   names the gate that must be green before the session ends.
4. **Two debts WP-11 and WP-15 must discharge** (both new 2026-09-29):
   the caller-side A2-9.4 platform-minted exemption, and A0-2.3's UTF-8
   validation of string fields, which still has no owner. Also
   `TestCapsRejectWithSummaryTooLarge`, which belongs to the first caller of
   `caps.Fits` — that is WP-11 or WP-15.
5. Design sessions still queued: **A3 stage views** (the plan's D11 proposes
   *not* writing it before the first run — if accepted, D4's composition rule
   stays unowned longer and needs a note in the contingency list), backlog §3 UI,
   §4 Pi node, §5 agent loop, §6 persistence, §7 security work items, §8 SSO,
   §9 CI remainder, §10 observability, §11 model benchmarking.

## 3. Definition of done for the next session

- [ ] All three 2026-09-29 PRs merged, branches deleted locally and remotely,
      local `main` fast-forwarded, CI green on `main`.
- [ ] WP-09 and WP-14 each reviewed independently and merged or opened as their
      own PR: every clause in their review row covered, the payload structs
      matching A1-3.3's tables by reflection, and the closed kind lists pinned
      against the contract rather than against the implementation.
- [ ] `gofmt -l .`, `go vet ./...`, `go build ./...`, `go test ./...`,
      `go test -race ./...` green under `GOMEMLIMIT` / `-timeout`; `go.mod`
      still zero requires except pgx once WP-22 lands (ADR-0010);
      `verify-vectors.py` still `PASS 52 / FAIL 0` — or a higher count with the
      new P-82 vector added and every published byte recomputed.
- [ ] Session tracker + `sessions/BACKLOG.md` + this file updated, usage
      refreshed, PRs labelled `agent-built`, PR URLs verified to exist.

## 4. Process lessons to keep (they cost real sessions to learn)

- **Check a reviewer's evidence the way you check a child's code.** The
  `paging` review's only MUST FIX claimed `errs.callerOp` leaks a generic shape
  body — struct tags and all — into every `NewPage` error. It said honestly that
  its evidence was a symbol dump, not a live render, and it was wrong: Go 1.27
  renders `paging.NewPage[...]`. The real defect was a NIT. Reproduce before
  acting; **a symbol table is not a runtime value.**
- **Brief the symptom, not the line.** The `secretscan` fix lane was told to add
  one leak surface; it reported that the surface it was given is `err.Error()`
  verbatim and so closed nothing, found the surfaces that did matter
  (JSON-escaped forms), and proved the detector now catches a leak it previously
  missed. A child that argues with your remedy is succeeding.
- **An assertion that cannot fail is not a test — and neither is one that
  asserts the wrong thing.** Four separate findings this session were the same
  defect: a hand-transcribed rule-*id* column with no regexp column beside it,
  two unpinned entropy knobs, six rejection rows asserting a substring the
  wrapper prose also contains, and a locally re-declared `SEC-NTLM` pattern that
  tested Go's `regexp` instead of the package. Ask of every guard: what edit
  would make this pass while breaking the rule?
- **A frozen contract's own example is a fixture, and fixtures lie.** A0 §4's
  published `next_cursor` had a 25-character id body for 18 days and three
  reviews. It surfaced only because every child was told "the published literal
  wins; report a disagreement rather than adjusting the fixture".
- **Two writers can share one checkout if the rule is mechanical**: only your
  own directory, and never `./...` — because `go test ./...` in one lane compiles
  the other's half-written package. It held across four children this session,
  and both reviewers confirmed the tree with `git status` before and after.
- **A child's deliverable must be a file written incrementally**, never its
  final message; require a compiling skeleton in the first 10 minutes and a
  "stop and make it coherent at minute N−10" rule.
- **A child's completion preview truncates mid-finding.** This session's full
  reports were recovered from
  `~/.pi/agent/sessions/--home-wtadmin-Sleipnir--/subagent-artifacts/<run-id>_<agent>_0_output.md`
  — note that path, **not** the `<parent>/<child-run>/run-0/session.jsonl` shape
  AGENTS.md records, which did not exist here. `/tmp` is not durable
  (2026-09-21). Worth correcting AGENTS.md's delegation bullet.
- **Brief reviewers for the tools they have.** `sleipnir-architect` has a shell,
  passes the `reviewer` launch guard that rejects dense review briefs, and can
  recompute vectors independently; it can write, so the brief must forbid it and
  `git status` must be checked afterwards. Both reviewers obeyed.
- **Never prove a negative test by removing the bound and running it**
  (2026-09-21: 11.4 GB RSS, kernel OOM, the whole WSL VM down). Both fix lanes
  ran their mutation sweeps in copies outside the repo and restored them — which
  is what the brief demanded and what they reported.
- Verify every published vector mechanically before merging, and prefer an
  oracle the code under test did not produce. This session the paging literal was
  checked three independent ways (implementer in Python before writing Go,
  reviewer with a separate implementation, principal with a third).
