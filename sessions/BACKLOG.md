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
| A1 prompt injection | CRIT | flag | this session |
| A2 runtime escalation | CRIT | ✅ decided | ADR-0017 Accepted |
| A3 worker egress | CRIT | flag | this session → own ADR |
| A4 LLM data leak | HIGH | ✅ decided | ADR-0020 Accepted (masking + gateway) |
| A5 approval manipulation | HIGH | ✅ decided | ADR-0018 Accepted |
| A6 API credentials | HIGH | flag | this session + API design |
| A7 node theft/impersonation | HIGH | flag | this session + session 4 |
| A8 stored credentials | HIGH | flag | this session + session 6 |
| A9 own-web attacks | MED | flag | this session + AGENTS.md review bar |
| A10 tool supply chain | MED | flag | this session + tool-registry work |
| A11 evidence tampering | MED | flag | session 6 (hash-chained log) |
| A12 cross-engagement leak | MED | flag | API contract test suite |
| A13 jailbreak vs safety | MED | accepted | covered by design |
| A14 availability/DoS | LOW | flag | API design session |
| A15 insider abuse | LOW | flag | later (four-eyes option) |

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

## Standing open questions

- Offline capability of the Pi agent (session 4).
- Which AD attack techniques are in/out of v1 tool registry scope
  (session with tool baseline).
- Handling of quarantined out-of-scope discoveries (ADR-0016 follow-up).

## Resolved (kept for history)

- ~~Approval timeout~~ → 2h default, configurable per engagement (ADR-0012).
- ~~Notification channel priority~~ → signed webhooks only for v1 (ADR-0012).
- ~~Embedded coding-agent harness~~ → own loop confirmed (ADR-0015).
