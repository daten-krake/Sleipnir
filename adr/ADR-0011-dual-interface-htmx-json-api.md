# ADR-0011: Dual interface — HTMX UI + JSON API on one control plane

- **Status:** Accepted
- **Date:** 2026-09-03
- **Deciders:** Product owner, architect

## Context

The UI is HTMX/server-rendered, but orchestrators, workers, remote agent
nodes, and future automation all need programmatic access to the same
capabilities.

## Decision

- Every platform capability is exposed twice: as HTML endpoints for the
  HTMX UI and as a structured **JSON API under `/api/v1`**, both backed by
  the same core services, same auth/session model.
- The JSON API is the **control plane** for: orchestrators/workers
  (event reporting, approval requests, scope checks), remote agent nodes,
  scripting/CI, and later purple-team integrations.
- Live streaming: SSE for the UI; the same event log streamable as JSON
  over the API.

## Consequences

- **Positive:** one implementation, many consumers; the platform is
  automatable from day one; remote-agent protocol becomes "just the API".
- **Negative:** API surface is part of the attack surface — needs its own
  auth hardening, rate limiting, and input validation discipline.
- **Follow-ups:** API auth for non-humans (per-node/per-job credentials),
  versioning policy.
