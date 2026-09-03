# ADR-0017: Spawn broker — orchestrators never touch the container runtime

- **Status:** Proposed (arisen from adversarial review 001, finding A2)
- **Date:** 2026-09-03
- **Deciders:** architect proposing, product owner to confirm

## Context

ADR-0007 states the orchestrator spawns workers. If orchestrator containers
are given Docker API/socket access, a prompt-injected orchestrator
(adversarial review A1) can escalate to host compromise. Refines the
*mechanism* of ADR-0007 without changing its topology.

## Decision (proposed)

1. Only the **platform** talks to the container runtime.
2. Orchestrators request workers through the platform API (**spawn
   broker**): each request carries engagement, task description, desired
   tool image, and is validated against scope, risk tier, image allowlist,
   and per-run resource quotas.
3. The platform creates the worker on the run's isolated network, hands the
   orchestrator the worker's job-scoped endpoint, and owns the full
   lifecycle (monitor, kill, cleanup) — making the hard stop (ADR-0005)
   enforceable and respawn-proof.
4. Remote agent nodes implement the same broker contract locally.

## Consequences

- **Positive:** removes the single worst escalation path; quotas and image
  allowlisting come free; hard stop is platform-owned; same contract for
  local and remote execution.
- **Negative:** one extra hop in worker startup; broker becomes critical
  path (must be simple and well-tested).
- **Follow-ups:** spawn request schema (API session); quota defaults.
