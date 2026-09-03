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

## Decisions pending / next sessions

See `sessions/BACKLOG.md`. Queue: handover contracts, program layout,
UI design, remote agent (Pi), agent loop design, persistence schema.
Awaiting product owner: approval-timeout behavior, notification channel
priority, confirmation of ADR-0012 and ADR-0015.

## Token usage

| total input | uncached input | cache read | cache write | output | reasoning |
| 545395 | 51315 | 494080 | 0 | 27618 | 11910 |
| 0 | 0 | 0 | 0 | 0 | 0 |

_(run `sessions/update-usage.sh sessions/2026-09-03-architecture-kickoff.md` at session end)_
