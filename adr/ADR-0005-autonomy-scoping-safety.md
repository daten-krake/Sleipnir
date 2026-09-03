# ADR-0005: Autonomy tiers, scoping, and safety controls

- **Status:** Accepted
- **Date:** 2026-09-03
- **Deciders:** Product owner, architect

## Context

Sleipnir runs hostile actions driven by LLM agents. Autonomy must be tiered
and bounded: scopes, deny-lists, human approval gates, and a hard stop.
The platform must also respect **OWASP Top 10** (for its own web surface)
and the **OWASP Top 10 for LLM Applications** (prompt injection, excessive
agency, etc.) as design constraints.

## Decision

1. **Autonomy tiers per action class:**
   - Recon / enumeration: **fully automatic**.
   - Research / analysis of findings: **fully automatic**.
   - Exploitation, initial access, privilege escalation (and any
     state-changing/dangerous action): **requires explicit human approval**
     from the operator before execution.
2. **Scoping:** every engagement has a mandatory allowlist scope (hosts,
   ranges, domains). No action may target anything outside scope.
3. **Blacklist:** global/per-engagement deny-list of forbidden targets and
   ranges. **Blacklist beats allowlist** — it is the last line of defense
   and cannot be overridden by scopes or approvals.
4. **Hard stop:** an immediate global kill switch halts all running jobs,
   plus per-job abort. Stopped agents must not be able to restart themselves.
5. Scope check + blacklist check happen in the **platform core** (enforcement
   point), not inside the agent — agents cannot be trusted to police
   themselves.

## Consequences

- **Positive:** auditable, sellable to cautious organizations; a compromised
  or hallucinating agent cannot reach outside its cage.
- **Negative:** approval queue + notification UX is required early; scans
  stall when no operator is present (acceptable by design).
- **Follow-ups:**
  - ADR on approval UX: timeout behavior, notification channels.
  - Define the risk-tier taxonomy of actions/tools (drives approval routing).
  - Prompt-injection defenses per OWASP LLM Top 10 (sanitize tool output
    shown to models) — needs dedicated design discussion.
