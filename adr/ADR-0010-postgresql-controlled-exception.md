# ADR-0010: PostgreSQL via controlled dependency exception

- **Status:** Accepted
- **Date:** 2026-09-03
- **Deciders:** Product owner, architect

## Context

Product owner prefers PostgreSQL as the system of record (projects, runs,
events, evidence index, approvals, users). ADR-0001 forbids external Go
dependencies; `database/sql` has no stdlib driver. Options reviewed:
(A) one controlled exception for a pinned driver, (B) hand-rolled Postgres
wire protocol, (C) HTTP sidecar. Supersedes the open question in ADR-0001.

## Decision

1. **PostgreSQL** is the primary store.
2. Access via `pgx` under a new **controlled-exception rule**: the stdlib-only
   rule stands, but infrastructure-level drivers may be admitted by ADR if
   they are (a) pinned to an exact version, (b) vendored in-repo, (c) listed
   here. Current exception list: `pgx` (+ its module dependencies).
3. All persistence goes through an internal **store interface** so the
   driver remains swappable (keeps the door open to a pure-stdlib wire
   protocol implementation later without architectural churn).

## Consequences

- **Positive:** real SQL for audit queries/reporting; mature auth/TLS/
  transactions; fast to ship; evidence integrity backed by proven code.
- **Negative:** first supply-chain entry — mitigated by pinning + vendoring
  + this ADR; any future exception needs the same scrutiny.
- **Follow-ups:**
  - Schema design session (events table is the core; see ADR-0009).
  - Exception list lives here and must stay current.
