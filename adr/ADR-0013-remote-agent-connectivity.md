# ADR-0013: Remote agent connectivity via tunnel/mesh tooling

- **Status:** Accepted (transport approach); specific tooling open
- **Date:** 2026-09-03
- **Deciders:** Product owner, architect

## Context

For client-site engagements a Raspberry Pi agent sits inside the client
network while operators run the platform elsewhere (ADR-0007). Client
networks are typically outbound-only; Sleipnir should not reinvent NAT
traversal.

## Decision

- Remote connectivity is delegated to **established tunnel/mesh tooling**,
  candidates: **Tailscale**, **NetBird**, **chisel**.
- The Sleipnir agent protocol (JSON API over TLS, ADR-0011) then rides on a
  normal reachable network — no custom NAT/firewall punching.
- Direction: agent connects outward / mesh establishes the path; platform
  never requires inbound ports at the client.

## Consequences

- **Positive:** huge complexity removed; proven crypto; works behind
  outbound-only firewalls.
- **Negative:** dependency on third-party connectivity tool (its own supply
  chain + control-plane considerations); tool choice affects self-hosting
  posture.
- **Follow-ups — decide in remote-agent session:**
  - Tool choice: NetBird (self-hostable control plane, mesh) vs. Tailscale
    (easiest; self-host via Headscale) vs. chisel (minimal point-to-point).
  - Offline/buffered operation of the Pi when the tunnel is down.
  - Whether tool binaries ship inside the agent image.
