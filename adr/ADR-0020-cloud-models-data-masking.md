# ADR-0020: Cloud models allowed — mandatory data masking via LLM gateway

- **Status:** Accepted
- **Date:** 2026-09-03
- **Deciders:** Product owner, architect

## Context

Adversarial finding A4: recon data sent to cloud LLM endpoints leaks client
data. Earlier proposal was `local-models-only` for customer engagements.
Product owner decision: **cloud models are allowed, but customer data must
be masked** before it leaves the platform. Resolves the A4 policy item.

## Decision

1. **Per-engagement LLM data policy:** `local-only` · `cloud-masked`
   (default for customer engagements) · `cloud-raw` (default for lab/HTB).
   Stored on the engagement; enforced by the platform, never configurable
   inside agents.
2. **LLM gateway in the platform core:** *all* model traffic (orchestrator
   and worker calls) routes through a gateway service in the platform.
   The gateway is the single enforcement and logging point: it applies the
   engagement's policy (block / mask / pass-through) and records egress
   metadata (endpoint, model, payload size, policy applied). Agents never
   hold direct model-endpoint credentials for cloud providers.
3. **Pseudonymization, not deletion:** customer identifiers (hostnames,
   FQDNs, IPs, domains, account names, emails) are replaced with
   **consistent per-run tokens** (e.g. `DC01.corp.acme.com` → `HOST-004`)
   using a run-scoped mapping table. The model reasons over tokens; the
   gateway remaps model output back to real values at the boundary;
   reports show real values. The mapping table is stored encrypted at rest
   and never leaves the platform.
4. **Exclusion beats masking:** captured credentials, hashes, and tokens
   are **never sent to any cloud endpoint under any policy**. Agents work
   with opaque references (e.g. `CRED-003`) usable only through local tool
   invocations.
5. **Residual risk is declared:** masking cannot be perfect on free text
   (business data in documents, shares, descriptions). Mitigations:
   pattern-based scanner + identifier registry from the context graph
   (known entities masked deterministically), operator-visible egress log,
   and the policy option to still choose `local-only` per engagement.

## Consequences

- **Positive:** cloud model quality available for real engagements; egress
  becomes auditable at one chokepoint; local-models-only remains available
  for maximum-sensitivity clients; secrets never leave, period.
- **Negative:** gateway adds latency and a critical path (must be simple,
  well-tested); masking is best-effort on unstructured text; remapping must
  be robust or reports corrupt.
- **Follow-ups:**
  - Masking scanner design (graph-entity registry + patterns) → security
    hardening session.
  - Gateway credentials handling + egress log schema → API/persistence
    sessions.
  - `internal/errs` must redact payloads in its own log paths (ADR-0019).
