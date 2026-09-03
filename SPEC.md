# Sleipnir Platform Specification

- **Version:** 0.1 (working draft)
- **Date:** 2026-09-03
- **Audience:** humans *and the agents that will build this platform*
- **Authority:** this spec consolidates `docs/VISION.md` and `adr/`. Where
  this file and an ADR disagree, **the ADR wins** and this file must be
  fixed.

## 1. Product

Sleipnir is a self-deployable, single-tenant, agent-driven pentest platform.
v1 attacks internal networks and Active Directory; web-app testing and a
purple-team mode follow later. It must run on local AI (Qwen via
OpenAI-compatible endpoints) and proves itself first on HackTheBox targets.
Differentiators: tiered human approval for dangerous actions, a complete
tamper-evident command/evidence log, revertable actions (client cleanup),
and remote on-site agent nodes (Raspberry Pi). See `docs/VISION.md`.

## 2. Fixed constraints

| # | Constraint | Source |
|---|-----------|--------|
| C1 | Go, standard library only. Exceptions only via ADR-0010 rule (current: pgx, pinned + vendored) | ADR-0001, ADR-0010 |
| C2 | UI = HTMX + server-rendered Go templates; no SPA frameworks | ADR-0001 |
| C3 | Everything runs in Docker; hostile workloads only in ephemeral containers | ADR-0002, ADR-0007 |
| C4 | Primary store: PostgreSQL | ADR-0010 |
| C5 | No action outside scope; blacklist beats allowlist; enforcement in platform core, never in agents | ADR-0005 |
| C6 | Exploitation / initial access / privilege escalation require human approval (2h expiry, configurable) | ADR-0005, ADR-0012 |
| C7 | Agent configs immutable while jobs run | ADR-0006 |
| C8 | Learning/memory strictly per-engagement; no cross-customer data flow | ADR-0016 |

## 3. Roles

- **Admin:** users, settings (LLM endpoints, model-role matrix,
  notifications, tool registry), global blacklist, hard stop.
- **Operator:** engagements they are assigned to — scopes, runs, approvals
  (only assigned operators may approve for their engagement), cleanup.
- **Viewer:** read-only access to assigned engagements/reports.
- Login: username + password (PBKDF2-SHA256, hand-rolled, RFC test vectors)
  + TOTP (RFC 6238). Session cookies, HttpOnly/SameSite/Secure.

## 4. System topology

```
                 ┌──────────────────────────── Platform (Docker) ───────────────────────────┐
   browser ─────►│  HTMX UI + /api/v1 JSON  (one service core, two faces — ADR-0011)        │
                 │  auth · scope/policy engine · approval service · notification service    │
                 │  spawn broker · event store · context graph · evidence store · tool reg. │
                 └───────▲──────────────▲────────────────────▲──────────────┬───────────────┘
                         │ JSON API     │ JSON API          │ JSON API     │ spawn requests
                         │ (mTLS token) │ (job token)       │ (job token)  │ (never a docker socket)
                 ┌───────┴──────┐  ┌────┴─────┐   ┌─────────┴───────┐      ▼
                 │ Remote agent │  │ Orchestr. │   │ Worker (hand)   │  worker containers
                 │ node (Pi)    │  │ (per job) │   │ (per task)      │  (ephemeral, tool images)
                 └──────────────┘  └────┬──────┘   └─────────────────┘
                                        │ LLM calls (OpenAI-compatible)
                                        ▼
                          Model endpoints (Qwen Cloud, LM Studio, ...)
```

Components:

1. **Platform core** — the only long-lived part. Owns all enforcement
   (scope, blacklist, approval, kill switch), all persistence, and the
   spawn broker. Exposes HTMX UI + JSON API + SSE.
2. **Orchestrator (mastermind)** — one container per job. LLM-driven
   planning over the engagement context graph; requests tools/workers via
   the API; never touches a container runtime or the database directly.
3. **Workers (hands)** — ephemeral containers per task, running tool-registry
   images. Execute, log centrally, return *summarized* results + evidence
   refs. Egress restricted to target scope.
4. **Remote agent node** — small Go binary (+Docker) on a Raspberry Pi at
   the client; connects outward (mesh/tunnel per ADR-0013); executes jobs
   locally; buffers when offline.
5. **Model endpoints** — external OpenAI-compatible servers; the platform
   never hosts models itself. Model-role matrix per ADR-0014.

## 5. Job lifecycle (normative flow)

1. Admin/operator creates an **engagement**: client, scopes (allowlist),
   blacklist, assigned operators, model-role config, rules of engagement.
2. Operator starts a **run** → platform validates scope/blacklist →
   **spawn broker** starts an orchestrator container with a job-scoped API
   token.
3. **Recon phase** (auto): orchestrator requests enumeration workers;
   results flow into the central event log and the context graph.
4. **Research phase** (auto): analysis workers/orchestrator reason over
   graph views; hypotheses recorded as graph nodes with provenance.
5. **Action proposal:** any dangerous action is proposed as a concrete,
   reviewable action (tool + exact arguments + target). Platform re-checks
   scope/blacklist, classifies risk tier, routes to the approval queue and
   notifies assigned operators (signed webhooks, SSE).
6. **Approval** binds to the exact action fingerprint; expiry 2h default.
   Expired → orchestrator gets `approval_expired` and replans.
7. **Execution:** worker runs the action, streams command log centrally,
   stores evidence (incl. screenshots via browser-enabled worker image),
   records a **revert record** for state-changing effects.
8. **Loop:** orchestrator queries the graph, self-corrects, iterates 3–7.
9. **Report:** HTML report (+PDF later) from graph, findings, log, evidence.
10. **Cleanup:** revert plan assembled from revert records → operator
    approval → execution → verification; non-revertable effects documented.

## 6. Safety model

- Risk tiers per tool/action in the tool registry drive approval routing.
- **Enforcement point is always the platform core** (scope, blacklist,
  approval, spawn, kill). Agents are never trusted to police themselves.
- **Blacklist beats allowlist beats approval.** Hard stop kills all
  containers of a run; platform-owned lifecycle prevents respawns.
- Out-of-scope discoveries are **quarantined** graph nodes: recorded,
  never actionable.
- All hostile/untrusted content (tool output, model prose) is *untrusted
  input*: sanitized for UI rendering, flagged in approval views, never
  executed as configuration. (Full injection-defense design = security
  hardening session.)

## 7. LLM layer

- Provider contract: OpenAI-compatible chat completions + tool calling,
  streamed. Provider endpoint configuration per ADR-0006/0014.
- Model-role matrix (configurable): orchestrator = strong reasoning;
  enumeration/routine = small fast; research = capable; summarizer = small.
- Every run snapshots its exact model config into the event log
  (reproducibility).
- Supported-model list records per model: context size, tool-calling
  reliability, allowed roles.

## 8. Data model spine

- **Event log** — append-only; every command, model call, approval,
  delivery, lifecycle change. The audit trail and evidence backbone.
  Integrity: hash-chaining (planned, security session).
- **Context graph** — per-engagement nodes/edges with provenance
  (ADR-0016); property-graph tables in PostgreSQL.
- **Evidence store** — files/screenshots/dumps referenced from events and
  graph nodes.
- **Approvals, users, engagements, tool registry, notification config.**

## 9. Interfaces

- **HTMX UI** — HTML fragments + SSE for live run view and notifications.
- **`/api/v1` JSON** — the control plane for orchestrator, workers, Pi
  nodes, automation. Machine credentials: short-lived, job-scoped tokens.
- **Signed webhooks** — outbound notifications (ADR-0012).
- **Remote-node transport** — mesh/tunnel (NetBird/Tailscale/chisel, choice
  open) carrying the same API over TLS.

## 10. Deployment

- Docker images: `platform`, `orchestrator`, `worker-base`,
  `worker-browser`, `remote-agent`. Versioned, pinned tool payloads
  (ADR-0008 supply-chain rules).
- docker-compose for the platform; spawn broker manages job containers on
  dedicated per-run networks (workers isolated from host and platform
  internals; egress to targets only).

## 11. Non-goals (v1)

Multi-tenant SaaS · cross-engagement learning · Kubernetes · hosting model
inference · Windows-based workers · inbound connectivity into client
networks.

## 12. Build model: this platform is built by agents

Consequences that shape the design and the repo:

1. **Contract-first.** Event schema, graph schema, `/api/v1` surface, and
   handover views are defined before the components that consume them, so
   builder agents can work in parallel against stubs and contract tests.
2. **This repo is agent context.** `SPEC.md`, `adr/`, `docs/`, and
   `AGENTS.md` are the build input; keep them normative and current.
3. **Small, testable seams.** Every component hides behind an interface
   (store, llm provider, notify, spawn); high-risk hand-rolled code
   (auth, crypto, protocol parsing, streaming) requires RFC test vectors
   and review gates — see `AGENTS.md`.
4. **Deterministic builds.** stdlib-only + pinned exception list keeps
   agent-produced builds reproducible and auditable.

## 13. Open items

See `sessions/BACKLOG.md` (session queue) and the gap lists in
`docs/enterprise-readiness.md` and `docs/adversarial-review-2026-09-03.md`.
