# Session backlog / open items

Queued topics for upcoming architecture sessions, plus open questions that
travel with them. Reorder as needed; add new items at the bottom of the queue.

## Session queue

### 1. Handover contracts between agent stages
How recon → research → exploitation → reporting hand off data to each other
without bloating context.
- Handovers are **views over the engagement context graph** (ADR-0016):
  define which subgraph/summary each stage receives.
- Schema of handover artifacts (findings list, hypotheses, evidence refs).
- What the orchestrator sees vs. what stays in the event store.
- Size limits; summarization duties of workers.
- 2026-09-11: the seam is contracted — **A2-12** (what A3 may rely on) plus
  A1-8 (read/SSE guarantees) and the A0-7 cap registry. **A3 is blocked on the
  A0-7.10 composition rule** (PR #2 decision **D4**: 500 × 512 B ≈ 250 KiB ≫ the
  64 KiB view cap; interim fail-safe = the smaller cap governs and the builder
  truncates). Q1's fixed-stage-views + capped-1-hop design stands; the
  `internal/handoff` placement question falls out of A3, not A1/A2.
- 2026-09-21: **A3 is unblocked.** D4 was confirmed by the product owner
  (PR #2): the interim fail-safe stands as the frozen rule, and **A3 owns the
  real composition rule** — ~131 B per node is not a usable view, so A3 MUST
  design compact refs with full summaries only in the capped 1-hop drill-down.
  A3 must additionally budget bytes for the provenance evidence grade that
  ADR-0022 puts on every node and edge.

### 2. Program layout (Go module structure)
Package layout of the monorepo given stdlib-only + pgx exception.
- `cmd/`: `platform`, `orchestrator`, `worker`, `remote-agent`; `internal/`
  map per `DESIGN.md` §1 (earlier sketches in this item are superseded).
- Where the Dockerfile/compose topology lives.
- `internal/errs` + slog setup as a foundational first work package
  (ADR-0019).
- 2026-09-04: layout + design guidelines are normative in **`DESIGN.md`**
  (product owner: layout is a guideline doc, not an ADR); scaffold +
  `internal/errs` = first PR.
- 2026-09-07: contract documents live in top-level **`contracts/`**
  (`README.md` = lifecycle, document shape, shared contract-test merge
  gate); A0 conventions + A1 events + A2 graph drafted as one PR
  (`contracts/a0-a2-conventions`). A3–A8 fan-out briefs next session.
- 2026-09-11: **A0–A2 completed, reviewed twice, fixed and opened as PR #2**
  (13 files, +9056). A1 grew to 42 event kinds, A0 to eight canonical-JSON
  vectors and six foundation packages (`ids`, `cjson`, `errs`, `paging`,
  `timex`, `caps`), A2 gained the closed secret-scan rule table and normative
  content-fingerprint vectors. 162 review findings, **63 MUST FIX applied**;
  review reports + fix plan committed under `docs/reviews/`. Nine product-owner
  decisions (**D1–D9**) and a 46-item confirm checklist are in the PR body;
  the documents stay `Draft` until they are answered, then flip to `Frozen`
  (WP-00). Next code work: the shared contract-test suite (now **seven**
  categories, ~60 test ids named in the clauses they guard), then scaffold +
  `internal/errs` + `internal/logging`, then the foundation packages in the
  order the principal review's §4 proposes.
- 2026-09-21: **A0/A1/A2 are `Frozen`** (PR #2, product owner decisions D1–D9 +
  all three blocks of the 46-item confirm checklist). WP-00 flipped the
  statuses, turned every one of the 51 `PO confirm`/`PO decision`/`PO signature`
  markers into a decision record, made the **D5 signed-webhook head anchor
  normative** (A1-5.8 mitigation (4), new `notification_kind:chain_head_anchor`),
  and added **ADR-0022** (D2's signature: the provenance evidence grade replaces
  Q2's finding-level `confidence`); ADR-0021 became Accepted. Next: **WP-01**,
  the repo scaffold + `internal/errs` + `internal/logging` with their contract
  tests (`next_steps.md` §3 allows the first package to land with its tests).
- 2026-09-21: **WP-01 delivered** — `go.mod` (module
  `github.com/daten-krake/sleipnir`, Go 1.27, zero requires so ADR-0010 holds
  trivially), `internal/errs` (6 files, 22 test ids) and `internal/logging`
  (3 files, 15 test ids). Two rulings the implementers forced, both recorded in
  the session tracker: **E-a** `Error()` bounds its cause-chain walk at
  A0-2.11's 32 layers and names the truncation — the exported fields make a
  cycle reachable and `Error()` rides every log record and every `/api/v1`
  envelope, so an unbounded walk is an adversarial-A14 availability hazard, and
  the principal overruled "a depth cap is speculative under DESIGN §2"; **L-d**
  a nil logger falls back to `slog.Default()`, never panicking (ADR-0019 §6) and
  never dropping the record (ADR-0019 §3). `logging`'s zero-internal-imports
  boundary is enforced mechanically, by two tests that parse its own source
  rather than trusting a reviewer. Next: the remaining foundation packages in
  the principal review §4's order (`ids`, `cjson`, `paging`, `timex`, `caps`),
  then the shared contract-test suite.
- 2026-09-24: **WP-04/05/06 delivered** — `internal/ids` (A0-1),
  `internal/timex` (A0-5) and `internal/caps` (A0-7): 9 files, 2045 lines, and
  the 15 test ids A0 §4.1 plus the principal review name — no others,
  everything else a subtest. The A0 §4 import rule is verified, not trusted:
  `go list -deps` shows `caps` with **zero** internal imports (it returns no
  errors, so it needs no `errs`) and `ids`/`timex` with only `errs`. Five gates
  green plus `-race`; `go.mod` still at zero requires; contract vectors still
  **PASS 52 / FAIL 0**. Three independent reviews found **no code defect at
  all** — all 18 findings were a test's non-vacuity, a comment's accuracy or a
  gap in a package doc's coverage trail (e.g. an import allow-list that never
  *required* `crypto/rand`; a truncation corpus that left the rune walk-back's
  zero-budget path untested). Two **contract errata** came out of it, both
  product-owner decisions recorded as A0 §6 items 16 and 17: the truncation
  marker is 11 B, not the 12 B A0-7.5 claimed, and A0-5.3's rationale "Go
  normalizes leap seconds" is false (`time.Parse` rejects `:60`; the real
  hazard is `Z07:00` accepting `+02:00`), with §4.1 now naming WP-10 as the
  owner of the two clamp test ids. Next: **WP-07 `internal/cjson`** (A0-2, the
  high-review-bar package), then **WP-08 `internal/paging`**, with **WP-13
  `internal/secretscan`** parallel to both.
- 2026-09-24 (second half): **WP-07 delivered and merged** (PR #5, `997bbcc`) —
  `internal/cjson`, 4 files, 2094 lines, the §4 sketch's surface exactly and
  nothing more. All eight A0-2.17 vectors byte-exact on canonical bytes, the
  published `len` **and** the published SHA-256, with the angle-bracket
  placeholders expanded to real bytes; all 21 rejection rows plus 13 extras and
  both accept cases; 18 test ids, 163 subtests. The vector arithmetic was
  checked three independent ways (implementer in Python before writing Go,
  reviewer from the table alone, reviewer re-parsing the test's Go literals), so
  the fixtures are provably the contract's and not re-derived from the code.
  The independent review found the package holding **two notions of "the same
  key"**: `strings.ToLower` is a weaker equivalence than the `encoding/json`
  matching A0-2.5's own rationale cites (`{"s":1,"ſ":2}` was accepted, yet
  `json.Unmarshal` of `{"ſ":7}` matches a field tagged `json:"s"`), and exclusion
  matched byte-exactly while duplicates matched case-insensitively, so an
  excluded field could influence a digest (A0-2.12). One `foldKey`
  (orbit-minimum `unicode.SimpleFold`, ASCII lowercased first) now governs all
  three key-identity sites; pairwise `EqualFold` was **measured and refused**
  (quadratic, ~11 s for the ~40k keys a 1 MiB hostile document carries —
  adversarial A14), and A0-8.1's ASCII key rule keeps every platform key on a
  zero-allocation fast path pinned by `AllocsPerRun` rather than by timing. Two
  more A0 errata came out of it (**§6 item 18**: when each A0-2.11 bound bites,
  and who A0-2.14's `internal` is for — A0-8.3 settles it). Next: **WP-08
  `internal/paging`** (A0-4 — the first package needing both `ids` and `cjson`),
  with **WP-13 `internal/secretscan`** as the parallel option.
- 2026-09-04 (design interview): **all session-1 decisions locked** —
  fixed stage views + capped 1-hop (no query endpoint v1); two node types
  Finding/Hypothesis; hard-reject validation; size budgets as contract
  constants; quarantined discoveries reported "not tested", removable;
  permanent machine-token exclusion list; worker = report-only; no
  automation keys v1 but findings JSON export yes; job-lifetime stateful
  tokens; node tokens over mesh; single-use approvals; day-one hash
  chaining per engagement with export block + logged override; full
  `/api/v1` surface; additive-only versioning; spawn schemas
  platform-derived. See session tracker 2026-09-04. Next: contract docs
  A0–A8 (principal drafts A0; super-minimal fan-out).

### 3. UI design (HTMX)
Screens and flows.
- Engagement/project management; scope + blacklist editing.
- Live run view (SSE event stream, agent activity).
- Approval queue UX (the operator's daily driver).
- Evidence browser + report view; cleanup run UI.
- Login/roles; settings (LLM endpoints, model-role matrix, notifications).

### 4. Remote agent node (Raspberry Pi)
- Tooling choice: NetBird vs. Tailscale/Headscale vs. chisel (ADR-0013).
- Docker-on-Pi worker strategy; offline buffering; image distribution
  to sites without internet.

### 5. Agent model & loop design
ADR-0015 accepted; now concrete.
- Agent profile schema (prompt, model, allowed tools, risk tiers).
- Loop mechanics: context window management, retries, **self-correction /
  reflection against the context graph (ADR-0016)**.
- Bounded graph query API available to agents.
- Relation of existing `Pen_Enumeration`/`Pen_Researcher` prototypes to
  platform agent profiles.

### 6. Persistence schema
PostgreSQL schema (ADR-0010): projects, runs, events, evidence, approvals,
users, tool registry — plus the **context graph tables** (ADR-0016).
Event table is the spine. Includes: **hash-chained event log** (adversarial
review A11), schema migrations strategy, retention columns.
- **A0-1.9's `TestIDOrderingMatchesByteOrderCollateC` is an explicit acceptance
  item here** (2026-09-24): `internal/ids` deliberately does not ship it — it
  is a PostgreSQL integration test, opt-in per DESIGN §8, and a skipped test is
  worse than an absent one — so the A0-1.9 pairing is owned by
  `store/postgres`: `COLLATE "C"` declared **on the column** in the DDL, never
  re-specified per query (a per-query `COLLATE` silently disables index use on
  every paginated read, A0-4.6).

### 7. Security hardening (from adversarial review 001)
Findings tracker (details in `docs/adversarial-review-2026-09-03.md`):

| Finding | Severity | Class | Tracked in |
|---------|----------|-------|------------|
| A1 prompt injection | CRIT | flag | contracted: A1-4.4 untrusted marking (platform-computed, `UntrustedFields`), A1-4.9 + A2-9 secret scan, ADR-0018 §4 approval-view flag; rendering rules still open (UI session) |
| A2 runtime escalation | CRIT | ✅ decided | ADR-0017 Accepted; write path contracted in A1-7.1/7.4 (platform-only composition) |
| A3 worker egress | CRIT | flag | this session → own ADR; A1-4.9/A2-9.3 make gateway exclusion the enforcement point |
| A4 LLM data leak | HIGH | ✅ decided | ADR-0020 Accepted (masking + gateway); `llm_call` is the egress log, metadata only (A1-4.2) |
| A5 approval manipulation | HIGH | ✅ decided | ADR-0018 Accepted; A1-3.3/4.2 store the action spec as a chained artifact and bind `fingerprint_hash` + `expires_at` |
| A6 API credentials | HIGH | flag | A5 (machine-principal exclusion list derived from A1-7.4/A1-8.4) |
| A7 node theft/impersonation | HIGH | flag | this session + session 4; `slp_node_` id shape and buffering/replay contracted (A1-7.6/7.11) |
| A8 stored credentials | HIGH | flag | this session + session 6; A2-9 no-secret-values + the closed rule table |
| A9 own-web attacks | MED | flag | this session + AGENTS.md review bar; untrusted content never becomes configuration (A1-4.4, A2-6.7) |
| A10 tool supply chain | MED | flag | this session + tool-registry work; `image_digest` is registry-derived, never orchestrator-supplied |
| A11 evidence tampering | MED | **contracted** | A1-5 (per-engagement chain, genesis, `chain_spec`), A1-6 (verification, `integrity_failed`, admin-only override per ADR-0021), `chain_head_trail` + `head_regression`, **and the out-of-band signed-webhook head anchor of A1-5.8 (4)** (D5 approved 2026-09-21); **residual, narrowed not closed:** an attacker with store-write *and* log-write on one host can still forge history and suppress a delivery, but suppression is visible to the webhook recipient as a head that stops advancing — detection depends on that recipient retaining and comparing its anchors (A1-6.6) |
| A12 cross-engagement leak | MED | **contracted** | A1-8.4/8.6, A2-11 + named negatives (`TestCursorFromEngagementARejectedInB`, `TestNoBulkEventReadSpansEngagements`, …); suite still to write |
| A13 jailbreak vs safety | MED | accepted | covered by design |
| A14 availability/DoS | LOW | flag | API design session; A1-6.8 rate-limits on-demand verification |
| A15 insider abuse | LOW | **accepted** | ADR-0021: not mitigated by the platform for the integrity-override control; the service owner compensates with the chained `artifact_released` trail |

Work items:
- Worker egress policy = target scope only (A3) → ADR.
- Untrusted-content policy: marking, sanitization, UI rendering rules (A1/A9).
- Job-scoped credential design (A6); secrets/credential handling (A8).
- Node hardening: encrypted storage, mTLS identity, revoke/wipe (A7).
- Data classification resolved (A4 → ADR-0020); remaining masking work:
  scanner design (graph-entity registry + patterns) in this session.
- Tool image build/pin/sign pipeline (A10).
- Platform-host-compromise scenario review.

### 8. SSO / OIDC
Own session after platform core exists (enterprise P0; hand-rolled
OIDC/JWT client, high-review bar).

### 9. CI/CD pipeline design
From PR gate to release: build (stdlib + vendored, reproducible), the
WORKFLOW.md gates as pipeline steps, container image build/pin/sign
pipeline (adversarial A10), versioning, rollback. Added 2026-09-04.
- **2026-09-21, product owner: add CI *before* the first code package, not
  after.** The repo has no `.github/workflows/`, so the WORKFLOW §4 merge gate
  (`gofmt -l`, `go vet ./...`, `go build ./...`, `go test ./...`, `go mod
  verify`) is currently unenforced and becomes load-bearing the moment
  `internal/errs` lands. Scope for the first pipeline: those five gates on a
  pinned Go toolchain, stdlib-only + `vendor/` consistency, and the
  `docs/reviews/*-verify-vectors.py` contract-vector check for any PR touching
  `contracts/`. Image build/pin/sign (A10) stays a later item.
- **2026-09-21: the first pipeline landed with WP-01**
  (`.github/workflows/ci.yml`, three jobs), meeting the "before, not after"
  directive. `detect` asserts the preconditions so the gates cannot silently
  no-op on a branch that removed them. `go-gates` runs the five §4 commands plus
  `go test -race` on the toolchain `go.mod` pins, with an ADR-0010 §2 dependency
  check that fails closed on any import or require outside the exception list
  (prefixes match exactly or as a subpackage, so `pgxfoo` cannot ride in on
  `pgx`) and `go mod verify` + `vendor/` consistency. `contracts` recomputes
  every published vector and asserts **both** a clean exit and a
  `PASS n>=1 / FAIL 0` summary line — the grep is load-bearing, because a
  verifier whose regexes stopped matching computes zero checks, fails nothing
  and exits 0, which is how I-02 got introduced — rejects raw U+2028/9/7F in
  table rows, and fails on a tracked credential-looking file *name* (never
  contents, which would copy secret material into the log ADR-0019 §5 protects).
  Triggers are `pull_request`/`push` on `main` only, so a feature branch does
  not run it and the pipeline's own first execution is the PR that adds it.
  `docs/reviews/2026-09-21-ci-gate-fixtures.sh` (16 fixtures, step bodies
  transcribed verbatim, mutating only copies under `$TMPDIR`) is the negative
  proof that those assertions bite. Image build/pin/sign stays A10, and so does
  full-SHA action pinning — the actions say so rather than leaving it implicit.

### 10. Monitoring & observability
Slog JSON export (ADR-0019 structure), health endpoints, per-run
metrics, alerting over the existing signed-webhook channel, audit/SIEM
export (enterprise gap list). Added 2026-09-04.

### 11. Model benchmarking & drift detection
Standing benchmark suite (fixed HTB-based task set) scored per model;
per-run model snapshot (ADR-0014) + gateway egress metadata (ADR-0020)
as the data source; define drift thresholds and reaction (alert,
quarantine model from role matrix). Added 2026-09-04.

## Standing open questions

- Offline capability of the Pi agent (session 4).
- Which AD attack techniques are in/out of v1 tool registry scope
  (session with tool baseline).
- **Who validates UTF-8 in contract string fields** (new 2026-09-24, → WP-09/
  WP-14): `cjson.CanonicalValue` cannot enforce A0-2.3 on a Go value, because
  `encoding/json` replaces an invalid string with U+FFFD before the canonicalizer
  sees any bytes. A0-2.3 governs documents; field-level validation (A0-8.1) is
  the only place left, and no package owns it yet. Related, and enforceable only
  by review: `json.Marshal(float64(2))` emits `2`, byte-identical to an integer,
  so A0-2.6's "no float field in a canonicalized type" needs a merge-gate line
  when the domain types land.
- **Where attribute-level redaction lives** (new 2026-09-21, → §10): DESIGN §1
  gives `logging` zero internal imports, so it cannot call `errs`' redaction
  helpers. WP-01's ruling: `errs` owns redaction of error strings (`errs.Secret`
  is leak-proof under every `fmt` verb and both encoders), `logging` takes only
  caller-supplied correlation ids, and the `api` handler — which may import both
  — is the only place a kind and a redacted attribute meet a log record. If the
  observability session wants redaction enforced *inside* `logging`, that needs
  a DESIGN §1 amendment or a third foundation package.
- Known debt accepted at the freeze (2026-09-21): engagement-assignment and
  credential-revocation audit belong to A5; no `evidence_removed` kind, so
  A1-8.8 stays unimplementable until one exists; A1 §4.2 has no approval-path
  JSON example (E-08); **`cvss_v3_x10` has no declared range** — the product
  owner explicitly declined to declare one (2026-09-21), and because adding
  `[0,100]` later *narrows* an accepted value set it is **not additive-safe**
  (A0-6.5) and will need an ADR (A0-7.2).
- **ADR-0022 follow-ups** (2026-09-21): the report and UI sessions MUST render
  the provenance evidence grade itself, never a re-invented adjective, and MUST
  show the provenance entries behind a `verified` grade; a *numeric* confidence
  would reopen the two-sources problem ADR-0022 closes and needs its own ADR.
- Store-seam duties the contracts hand to backlog §6: the per-engagement append
  lock (A1-5.4), dedup table (A1-7.6), ingest watermark (A1-7.7),
  `chain_head_trail` and the kill outbox with `REVOKE UPDATE, DELETE`, and any
  checkpoint/re-genesis mechanism (A1-6.9 — needs its own ADR).

## Resolved (kept for history)

- ~~**Whether ADR-0019 needs an amendment note for A0-3.6's attribute
  spelling**~~ → **ruled by the product owner 2026-09-24:** yes — a
  spelling-only amendment note inside the Accepted ADR, decision text
  unchanged, so no new ADR and nothing superseded (`adr/ADR-0019-…` §3 plus its
  Status line). A0-3.6 governs: `engagement_id`/`run_id`/`job_id`/`node_id`,
  with `graph_node_id` for a graph node and `node_id` always the remote agent
  node (`slp_node_`, Q9).

- ~~**PR #2 decisions D1–D9**~~ → all answered by the product owner 2026-09-21
  (PR #2 body edit + session), A0/A1/A2 `Frozen`: **D1** admin-only integrity
  override, *not* single-use → **ADR-0021 Accepted**; **D2 signed** →
  **ADR-0022** (the provenance evidence grade `observed·inferred·verified`
  replaces Q2's finding-level `confidence`); **D3** `usr_` registered in A0-1.2 +
  `KindUser`; **D4** the interim fail-safe stands, A3 owns the real composition
  rule; **D5 approved** — the ADR-0012 §3 signed-webhook head anchor is
  normative (A1-5.8 (4)), which discharges ADR-0021's load-bearing follow-up;
  **D6** no operator release of quarantine; **D7** blacklisted discoveries
  recorded, not refused; **D8** reject, never redact; **D9** user/session audit
  → A5 in a platform-scoped store. Plus all three blocks of the 46-item confirm
  checklist (A0 §6.1–13, A1 §6.1,3,5–14, A2 §6.4–11).
- ~~Approval timeout~~ → 2h default, configurable per engagement (ADR-0012).
- ~~Notification channel priority~~ → signed webhooks only for v1 (ADR-0012).
- ~~Embedded coding-agent harness~~ → own loop confirmed (ADR-0015).
- ~~Handling of quarantined out-of-scope discoveries~~ → Q5 (2026-09-04):
  included in the reporting handover marked *not tested*, non-actionable,
  operator can exclude them from the report; contract clause in A2.
- ~~`qwen3.8-flash` availability for `sleipnir-implementer`~~ → verified
  2026-09-07 (smoke test ran; provider `qwen-token-plan`).
