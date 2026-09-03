# ADR-0001: Core technology stack — Go (stdlib only) + HTMX

- **Status:** Accepted
- **Date:** 2026-09-03
- **Deciders:** Product owner, architect

## Context

Sleipnir needs an implementation language and UI approach. The product owner
fixed the following constraints up front.

## Decision

1. The platform is implemented in **Go**.
2. **No external Go dependencies are allowed** — everything must be buildable
   with the Go standard library only (`go.mod` has zero `require`s).
3. The web UI uses **HTMX** served from server-rendered Go `html/template`
   pages over `net/http`. No SPA frontend framework, no Node build step.

## Consequences

- **Positive:**
  - Trivially reproducible builds, no supply-chain surface from Go deps —
    appropriate for a security product.
  - Single static binary; ideal for Docker (ADR-0002).
  - Server-rendered HTMX keeps UI logic in the backend; simple mental model.
- **Negative / forces we must design around:**
  - No third-party drivers: no native SQL driver, no YAML, no JWT, no gRPC.
    Persistence, auth tokens, config parsing etc. must be solved with
    stdlib means or out-of-process sidecars. → dedicated ADRs required.
  - No established web framework: routing, middleware, sessions are hand-rolled.
  - Rich interactive UI (live scan output) must be done via HTMX + SSE
    (`text/event-stream`, supported by stdlib).
- **Follow-ups:**
  - ADR needed: persistence strategy without external Go drivers.
  - ADR needed: HTTP routing/middleware approach.
