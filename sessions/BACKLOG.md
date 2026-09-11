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
| A11 evidence tampering | MED | **contracted** | A1-5 (per-engagement chain, genesis, `chain_spec`), A1-6 (verification, `integrity_failed`, single-use admin override), `chain_head_trail` + `head_regression`; **residual:** tail truncation needs the webhook anchor (PR #2 **D5**) |
| A12 cross-engagement leak | MED | **contracted** | A1-8.4/8.6, A2-11 + named negatives (`TestCursorFromEngagementARejectedInB`, `TestNoBulkEventReadSpansEngagements`, …); suite still to write |
| A13 jailbreak vs safety | MED | accepted | covered by design |
| A14 availability/DoS | LOW | flag | API design session; A1-6.8 rate-limits on-demand verification |
| A15 insider abuse | LOW | flag | PR #2 **D1** (admin-only override), **D9** (assignment audit gap → A5) |

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
- **PR #2 decisions D1–D9** (2026-09-11): override authority + lifetime (narrows
  Q11), the Q2 confidence deviation (needs a signature), `usr_`, the A0-7.10
  view-cap composition rule (**blocks A3**), out-of-band head anchoring, no
  operator release of quarantine, blacklisted discoveries recorded, reject vs
  redact, and user/session audit ownership.
- Known debt accepted at the freeze: engagement-assignment and
  credential-revocation audit belong to A5; no `evidence_removed` kind, so
  A1-8.8 stays unimplementable until one exists; `cvss_v3_x10` has no range
  (P-77 — the only deferred item that is **not** additive-safe, needs an ADR).
- Store-seam duties the contracts hand to backlog §6: the per-engagement append
  lock (A1-5.4), dedup table (A1-7.6), ingest watermark (A1-7.7),
  `chain_head_trail` and the kill outbox with `REVOKE UPDATE, DELETE`, and any
  checkpoint/re-genesis mechanism (A1-6.9 — needs its own ADR).

## Resolved (kept for history)

- ~~Approval timeout~~ → 2h default, configurable per engagement (ADR-0012).
- ~~Notification channel priority~~ → signed webhooks only for v1 (ADR-0012).
- ~~Embedded coding-agent harness~~ → own loop confirmed (ADR-0015).
- ~~Handling of quarantined out-of-scope discoveries~~ → Q5 (2026-09-04):
  included in the reporting handover marked *not tested*, non-actionable,
  operator can exclude them from the report; contract clause in A2.
- ~~`qwen3.8-flash` availability for `sleipnir-implementer`~~ → verified
  2026-09-07 (smoke test ran; provider `qwen-token-plan`).
