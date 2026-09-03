# ADR-0007: Execution topology — Platform / Orchestrator / Workers + remote agent

- **Status:** Accepted
- **Date:** 2026-09-03
- **Deciders:** Product owner, architect

## Context

Everything hostile must be isolated. The platform runs centrally (or at
home); for on-site client engagements a small device (e.g. Raspberry Pi)
should be deployable at the client, while operators stay remote.

## Decision

Three-tier Docker execution topology:

1. **Platform (the brain):** long-lived container(s) — UI, API, scoping &
   approval enforcement, persistence, central log store, orchestrator
   lifecycle management.
2. **Orchestrator ("mastermind"):** a **new container spawned per job** by
   the platform. Owns the engagement plan, spawns/steers workers, decides
   next steps with the LLM.
3. **Workers ("hands"):** ephemeral containers spawned by the orchestrator
   (one per task/tool as needed). They execute tools, **log centrally**, and
   return **summarized results** to the orchestrator so raw output never
   bloats the orchestrator's LLM context.

Plus:

4. **Remote agent node:** a small binary/service (target: Raspberry Pi)
   deployed inside a client network. It connects **outbound** to the
   platform and provides local execution capability there (job dispatch,
   tool execution, evidence upload).
5. Per-scan Docker networks isolate workers; workers cannot reach the
   platform brain or the host beyond the defined control channel.

## Consequences

- **Positive:** blast radius per job; context hygiene by design; remote
  engagements without exposing the platform inward.
- **Negative:** significant distributed-systems surface: container lifecycle
  management, remote-agent protocol, reliable result/log transport, offline
  tolerance of the Pi agent.
- **Follow-ups:**
  - ADR: platform↔agent communication protocol (mTLS, message format,
    offline queueing on the Pi).
  - ADR: how the Pi executes workers (Docker on the Pi vs. plain processes).
  - Central log/event schema (also the evidence backbone, see ADR-0009).
