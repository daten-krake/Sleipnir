# Session backlog / open items

Queued topics for upcoming architecture sessions, plus open questions that
travel with them. Reorder as needed; add new items at the bottom of the queue.

## Session queue

### 1. Handover contracts between agent stages
How recon → research → exploitation → reporting hand off data to each other
without bloating context.
- Schema of handover artifacts (findings list, hypotheses, evidence refs).
- What the orchestrator sees vs. what stays in the event store.
- Serialization format + size limits; summarization duties of workers.

### 2. Program layout (Go module structure)
Package layout of the monorepo given stdlib-only + pgx exception.
- cmd/ tree: platform, remote-agent; internal/ boundaries
  (api, auth, store, orchestration, agents, tools, notify).
- Where the Dockerfile/compose topology lives.

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
Depends on ADR-0015 confirmation.
- Agent profile schema (prompt, model, allowed tools, risk tiers).
- Loop mechanics: context window management, retries, streaming.
- Relation of existing `Pen_Enumeration`/`Pen_Researcher` prototypes to
  platform agent profiles.

### 6. Persistence schema
PostgreSQL schema (ADR-0010): projects, runs, events, evidence, approvals,
users, tool registry. Event table is the spine.

## Standing open questions

- Approval timeout behavior when no operator responds: pause forever vs.
  auto-expire after N hours? (belongs to session 3 / approval UX)
- Notification channel priority: signed-webhook-only for v1, plus email?
  (ADR-0012 proposed)
- Offline capability of the Pi agent (session 4).
- Which AD attack techniques are in/out of v1 tool registry scope
  (session with tool baseline).
