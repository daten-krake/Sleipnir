# ADR-0004: Target focus — Active Directory / internal network first

- **Status:** Accepted
- **Date:** 2026-09-03
- **Deciders:** Product owner, architect

## Context

Primary goal is internal pentesting with focus on **Active Directory**.
Web-app pentesting comes later; the long-term vision includes use as a
**purple team platform**.

## Decision

- v1 target domain: **internal networks and Active Directory** (network
  recon, AD enumeration, Kerberos/LDAP/SMB attacks, credential attacks,
  lateral movement, privilege escalation paths).
- Web application testing: later phase.
- Purple team mode: vision item, kept in mind (attack telemetry vs.
  detection validation), not designed yet.
- **HackTheBox machines are the primary test ground** during development;
  scoping rules still apply to HTB ranges (they are simply easy to declare).

## Consequences

- **Positive:** clear v1 scope; AD attack tooling runs well from Linux workers.
- **Negative:** AD-focused design must later absorb web targets without a
  rewrite (keep target/tool abstractions generic).
- **Follow-ups:**
  - Worker tooling baseline for AD (nmap, netexec/crackmapexec, impacket,
    bloodhound collector, kerberos tooling...) → tool registry ADR.
  - Evidence model for AD-specific artifacts (credential dumps, attack paths).
