# Session 2026-09-03 — Architecture kickoff

- **Date:** 2026-09-03
- **Session id:** 01a067d5-b67a-7d18-bd80-e7614a99232e
- **Model:** qwen-token-plan-individual/qwen3.8-max
- **Goal:** Establish architecture working mode (ADRs, session tracking,
  architect persona) and build shared understanding of the platform.

## Done this session

- Created `adr/` (README, TEMPLATE), `sessions/` tracker + `update-usage.sh`,
  architect persona `.pi/agents/sleipnir-architect.md`, `docs/VISION.md`.
- Architecture round 1 answered by product owner; decisions recorded:
  - ADR-0001: Go stdlib-only + HTMX (fixed constraint)
  - ADR-0002: Docker deployment
  - ADR-0003: Self-deployable, single-tenant; projects per client/engagement
  - ADR-0004: Target focus AD/internal first; web later; purple team vision;
    HackTheBox as dev test ground
  - ADR-0005: Autonomy tiers (recon+research auto; exploit/initial access/
    privesc gated by human approval), mandatory scopes, blacklist-beats-
    allowlist, hard stop, OWASP + OWASP-LLM as design constraints
  - ADR-0006: OpenAI-compatible LLM layer, local-first (Qwen), supported-model
    list later, agent config locked during runs
  - ADR-0007: Topology: Platform → per-job Orchestrator container → ephemeral
    Worker containers; central logging; remote agent node (Raspberry Pi)
  - ADR-0008: Tool registry wrapping external open-source tools
  - ADR-0009: Evidence: central command log, screenshots, revert records +
    approved cleanup run, HTML report + PDF option

## Round 2 answers → decisions

- ADR-0010: PostgreSQL via controlled exception (pgx, pinned+vendored),
  store interface keeps driver swappable
- ADR-0011: dual interface — HTMX UI + JSON API `/api/v1` as one control
  plane (orchestrator/workers/Pi agent all consume it)
- ADR-0012 (Proposed): notifications via signed webhooks (+optional SMTP);
  approvals restricted to operators assigned to the engagement
- ADR-0013: remote agent connectivity via tunnel/mesh tooling
  (Tailscale/NetBird/chisel), tool choice deferred
- ADR-0014: task-based model mix; start with Qwen Cloud, LM Studio/local
  later; per-run model snapshot required
- ADR-0015 (Proposed): own minimal agent loop in stdlib Go instead of
  embedding pi/opencode in worker containers
- Confirmed: agent config immutable while jobs run (ADR-0006 closed-door);
  cleanup needs explicit approval, non-revertable effects documented
  (ADR-0009); screenshots/evidence via headless browser where possible;
  roles admin/operator/viewer + TOTP for v1
- New: headless-browser capability implies a separate browser-enabled
  worker image (image strategy → tool-registry follow-up)

## Round 3 confirmations → decisions

- ADR-0012 Accepted: notifications = signed webhooks only (v1); approval
  timeout **2h default, configurable per engagement in the web GUI**;
  expired actions return as `approval_expired` for replanning
- ADR-0015 Accepted: own agent loop confirmed — "without a coding agent";
  requirement added: loop must support iteration, self-correction and
  learning within the running engagement
- ADR-0016 Accepted (new): **engagement-scoped context graph** as agent
  working memory — nodes/edges with provenance, policy-aligned with
  scope/blacklist, learning strictly per-engagement, storage in PostgreSQL,
  handovers as graph views
- Backlog updated: handover session now anchored on graph views; resolved
  questions archived

See `sessions/BACKLOG.md`. Queue: handover contracts, program layout,
UI design, remote agent (Pi), agent loop design, persistence schema,
security hardening, SSO.

## Round 4 — spec, adversarial review, enterprise readiness

- Created `SPEC.md` (v0.1) + `AGENTS.md` (rules for agent builders)
- `docs/adversarial-review-2026-09-03.md`: 15 findings; CRITICAL: prompt
  injection from targets, orchestrator→runtime escalation, worker egress
  exfil; HIGH: LLM endpoint data leak, approval manipulation, node theft,
  stored credentials; verdict: build-by-agents feasible contract-first
- `docs/enterprise-readiness.md`: P0/P1/P2 gap list (SSO, audit export,
  retention, backups, migrations, secrets, air-gap...)
- ADR-0017 Proposed: spawn broker (orchestrator never touches runtime)
- ADR-0018 Proposed: approval binding (fingerprint + execution-time
  re-validation)
## Round 6 — build agents + findings tracking

- Backlog: adversarial findings A1–A15 now tracked with severity/class/
  owner in session 7 table
- Created build agents:
  - `.pi/agents/sleipnir-principal.md` (qwen3.8-max, decomposes, spawns
    implementers in parallel, reviews, integrates, runs gates)
  - `.pi/agents/sleipnir-implementer.md` (qwen3.8-flash, one scoped work
    package, tests, honest reporting)
- Awaiting product owner: the data-classification policy (cloud models
  lab-only vs. customer engagements)

## Round 7 — confirmations + error/logging + Go style

- ADR-0017 and ADR-0018 → Accepted (backlog tracker updated)
- New binding convention → ADR-0019 Accepted: descriptive errors +
  granular logging (`internal/errs` captures origin function via
  `runtime.Caller`; errors read `component.Function: what: ids: cause`;
  stdlib `slog` with correlation attributes; log-or-return; secret
  redaction); motivation: agent-built code must be agent-debuggable
- `AGENTS.md`: idiomatic Go section + error/logging section added;
  both build agents (principal/implementer) updated to enforce it;
  SPEC.md constraints C9/C10 added

## Token usage

| total input | uncached input | cache read | cache write | output | reasoning |
| 3148179 | 119571 | 3028608 | 0 | 69991 | 29213 |

_(run `sessions/update-usage.sh sessions/2026-09-03-architecture-kickoff.md` at session end)_
