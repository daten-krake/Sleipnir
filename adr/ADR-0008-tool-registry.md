# ADR-0008: Tool registry wrapping external open-source tools

- **Status:** Accepted
- **Date:** 2026-09-03
- **Deciders:** Product owner, architect

## Context

Re-implementing mature offensive tooling (nmap, impacket, netexec, ...) in
stdlib Go would be wasted effort. The no-external-Go-dependency rule applies
to our own Go code, not to the offensive tooling the platform orchestrates.

## Decision

- Offensive/security capabilities come from **external open-source tools**,
  wrapped and catalogued in a **tool registry**.
- Registry entries declare: name, version, category, invocation contract
  (args/stdin/stdout), output parsing hints, **risk tier** (drives the
  approval gate of ADR-0005), resource needs, and the worker image it lives in.
- Tools execute **only inside worker containers** (or on a remote agent
  node), never in the platform process.
- The platform's own Go code remains stdlib-only (ADR-0001).

## Consequences

- **Positive:** leverage the entire open-source offensive ecosystem; clear
  supply-chain boundary (tools live in versioned images, our Go stays clean).
- **Negative:** tool output parsing/normalization is real work; image size;
  licensing of bundled tools must be checked per tool.
- **Follow-ups:**
  - Registry schema + how agents discover/request tools.
  - Baseline AD toolset list for v1.
  - Worker image build pipeline (one fat image vs. per-category images).
