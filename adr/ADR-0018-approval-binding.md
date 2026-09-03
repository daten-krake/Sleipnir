# ADR-0018: Approval binding — fingerprint-based approvals with execution-time re-validation

- **Status:** Accepted (confirmed by product owner 2026-09-03; arisen from adversarial review 001, findings A1/A5)
- **Date:** 2026-09-03
- **Deciders:** architect proposing, product owner to confirm

## Context

A prompt-injected or hallucinating agent could phrase approval requests
deceptively, or the executed action could drift from what was approved
(TOCTOU). Operators approve dangerous actions (ADR-0005/0012); the approval
must be unambiguous and tamper-proof.

## Decision (proposed)

1. Approvals are granted to an **action fingerprint**: tool id + exact
   arguments + target(s) + hash of the full action spec. Agent prose is
   displayed *in addition to*, never instead of, the concrete action.
2. At execution time the platform **re-validates**: fingerprint unchanged,
   scope still allows it, blacklist still clear, approval unexpired. Any
   mismatch ⇒ abort and re-queue for approval.
3. The executed action and its fingerprint are recorded in the event log
   next to the approval record (auditable pairing).
4. Approval requests originating from untrusted-content-influenced context
   are flagged as such in the UI (content quarantine marker from the
   security hardening session).

## Consequences

- **Positive:** deceptive summarization cannot change what actually runs;
  TOCTOU closed; audit pairing is evidence-grade.
- **Negative:** minor UX complexity (operators see raw commands — which is
  what they should see anyway); any action modification costs a new
  approval round (by design).
- **Follow-ups:** fingerprint schema; UI mock in UI design session; later
  four-eyes option for sensitive tiers.
