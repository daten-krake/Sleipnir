# ADR-0002: Docker as the deployment vehicle

- **Status:** Accepted (details open)
- **Date:** 2026-09-03
- **Deciders:** Product owner, architect

## Context

Sleipnir must be shipped and run as containerized software. Pentest workloads
(scanners, exploit attempts) are inherently hostile; containers also give us
the isolation boundary for agent-driven attack execution.

## Decision

Docker is the fixed packaging/runtime target for Sleipnir.

## Consequences

- **Positive:** reproducible environments; per-scan disposable execution
  containers are a natural fit for isolating agent actions.
- **Negative:** orchestrating worker containers from inside Go means talking
  to the Docker daemon (socket/API over HTTP+JSON, feasible with stdlib,
  hand-rolled) or shelling out to the `docker` CLI.
- **Open (to be decided in follow-up ADRs):**
  - Single container vs. docker-compose topology (API/UI, workers, storage).
  - How scan workers are spawned and networked.
  - Orchestration story beyond Docker (Kubernetes?) — out of scope for now.
