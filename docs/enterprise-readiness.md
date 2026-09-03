# Enterprise readiness — gap list (2026-09-03)

What separates "works for us / HTB" from "sellable to and deployable by
enterprises". Priorities: P0 = blocks enterprise deals, P1 = expected,
P2 = differentiator/later. Items already covered by ADRs are marked.

## Identity & access

| Gap | P | Notes |
|-----|---|-------|
| SSO: OIDC (and later SAML) | P0 | stdlib-feasible (hand-rolled JWT/OIDC client — high-review-bar code) but real effort; enterprises will demand it |
| API keys/scopes for automation consumers | P0 | extends ADR-0011 machine credentials |
| Fine-grained RBAC + per-client operator groups | P1 | v1 roles (ADR/spec §3) are coarse |
| Four-eyes approval policy for sensitive tiers | P1 | optional per engagement |
| Session/device management, revocation | P1 | |
| SCIM provisioning | P2 | |

## Compliance & legal

| Gap | P | Notes |
|-----|---|-------|
| Rules-of-engagement templates + signed scope records | P0 | legally grounding what we may attack; ties to ADR-0005 |
| Audit export (JSON/CSV + SIEM webhook) | P0 | event log already exists; needs export + retention proofs |
| Evidence chain of custody (hash-chained log, signed reports) | P0 | adversarial review A11 |
| Data retention & destruction policies per engagement (incl. captured creds) | P0 | adversarial review A8; GDPR-adjacent for EU clients |
| Per-engagement data classification (local-only / cloud-masked / cloud-raw) | ✅ | ADR-0020: masking + LLM gateway instead of local-only |
| SBOM + CVE policy for bundled tool images | P1 | adversarial review A10 |
| Licensing review of bundled offensive tools | P1 | ADR-0008 follow-up |

## Operations & resilience

| Gap | P | Notes |
|-----|---|-------|
| HA story: stateless platform replicas + PG failover | P1 | v1 single-host is accepted |
| Backup/restore + disaster-recovery runbook | P0 | evidence + DB; testable restores |
| Health/metrics endpoints + alerting hooks | P1 | JSON metrics, /healthz; no Prometheus dep — scrape-friendly format |
| Upgrade/migration path (schema migrations in-repo) | P0 | from day one, else painful later |
| Secrets management for platform operation (key storage/rotation) | P0 | adversarial review A8 |
| Resource quotas & scheduling for concurrent engagements | P1 | shared worker pools vs per-engagement |
| Offline/air-gapped install (images + models bundled) | P1 | fits local-AI story; matters for classified clients |

## Product surface

| Gap | P | Notes |
|-----|---|-------|
| Report branding/white-label for consultants | P1 | |
| Multi-client engagement management UX (portfolio view) | P1 | extends ADR-0003 projects |
| Purple-team integration hooks (detection validation) | P2 | vision item |
| Cross-engagement learning with consent+anonymization | P2 | explicitly deferred (ADR-0016) |

## Suggested sequencing

P0 items fold into the planned sessions: security hardening (classification,
secrets, custody), API design (machine credentials, audit export),
persistence session (migrations, retention), remote-agent session
(air-gap). SSO is large enough to deserve its own backlog entry once the
platform core exists.
