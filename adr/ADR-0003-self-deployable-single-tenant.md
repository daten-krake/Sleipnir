# ADR-0003: Audience and deployment model — self-deployable, single-tenant

- **Status:** Accepted
- **Date:** 2026-09-03
- **Deciders:** Product owner, architect

## Context

Sleipnir's users are internal security teams and pentest consultants. SaaS is
explicitly deferred. Consultants need to run engagements for different
clients from one installation.

## Decision

- Sleipnir is a **self-deployable, single-tenant platform** (one install per
  organization). No multi-tenant isolation in scope for now.
- Work is organized into **projects/engagements** (one per client or internal
  target environment), each with its own scopes, runs, evidence, and reports.
- SaaS/multi-tenancy is a possible later phase; architecture must not paint
  us into a corner (keep project-level data separation clean).

## Consequences

- **Positive:** no tenant-isolation machinery; simpler auth and storage.
- **Negative:** later SaaS phase may require retrofitting tenancy boundaries.
- **Follow-ups:** ADR on platform authentication/roles (operator, admin, viewer?).
