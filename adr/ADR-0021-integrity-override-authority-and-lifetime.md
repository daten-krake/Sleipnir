# ADR-0021: Integrity-override authority and lifetime

- **Status:** Proposed
- **Date:** 2026-09-11
- **Deciders:** product owner (decision, session 2026-09-11); architect
  (adversarial review findings C-01, S-07); principal build engineer (fix-plan
  ruling PO-1)

## Context

Q11 (2026-09-04) decided that the event log is hash-chained from day one, per
engagement, that verification failure **blocks exports**, and that an "explicit
operator override [is] logged as an event". `contracts/A1-events.md` A1-6.5 makes
that concrete, and the adversarial contract review of 2026-09-11
(`docs/reviews/2026-09-11-contract-review-adversarial.md`) found two defects in
it:

- **C-01 — the clause contradicted itself about *who*.** It said "only a `user`
  principal with the admin role may override" and then "an operator-scoped user
  MAY NOT override **another engagement's** break", which only makes sense if
  operators may override their own. Two builders would have implemented two
  different authorization rules on the one control that releases a customer
  artifact from a damaged chain. It also silently **narrowed Q11**, whose literal
  wording says "operator".
- **S-07 — the clause said nothing about *how long*.** One override authorized an
  unbounded stream of artifacts: a single human decision, once recorded, released
  every later export from a chain known to be broken.

Forces and constraints: SPEC §3 places integrity-class controls next to the hard
stop, which is admin-owned; ADR-0003 makes v1 self-deployable and single-tenant,
so in the common deployment one human holds both the admin and the operator role;
ADR-0009 §5 requires non-revertable effects and integrity state to be documented
in the report; SPEC §6 keeps enforcement in the platform core; and the long-term
product vision includes offering Sleipnir as a service, where a service owner
releases artifacts on behalf of customers (multi-tenancy is a v1 non-goal,
SPEC §11, but the vision is stated in `docs/VISION.md`).

## Options considered

1. **Operator-scoped override for the operator's own engagement** (Q11's literal
   reading) — fastest in a team deployment; weaker separation of duties; leaves
   C-01's ambiguity half-resolved, because "assigned operator" still needs an
   assignment rule; does not address S-07 at all.
2. **Admin-only, valid until the next verification or break** — one unambiguous
   rule, the control sits next to the hard stop, and in a single-tenant
   deployment it costs nothing operationally (the same human switches role).
   Leaves S-07 open: several artifacts can ride one decision.
3. **Admin-only and single-use per artifact** (what the freeze draft applied) —
   the strictest reading and a real answer to S-07; but one legitimate release
   decision (HTML report + findings JSON + evidence bundle) becomes three human
   decisions, and per-artifact gating does not survive a service offering where
   releases are routine and batched.

## Decision

**Authority: admin-only** (option 2). Only a `user` principal holding the admin
role may compose `integrity_override`; an operator-scoped user — including one
assigned to the engagement — MUST NOT override any break, and A5 gates the
endpoint accordingly. Q11's "operator" is read as "human" and is narrowed here
deliberately, because SPEC §3 puts integrity-class controls next to the hard
stop.

**Lifetime: not single-use** (option 3 rejected). An `integrity_override` with
`scope:"export"` is valid from its own `seq` until the next `chain_verified` or
`chain_break_detected` event on that chain, and MAY authorize more than one
artifact release. A new break always needs a new human decision; a later
verification supersedes the override because the chain is sound again.

**The residual risk is accepted and transferred to the service owner** who runs
the deployment (product owner, 2026-09-11): one admin decision may release
several customer artifacts from a chain known to be broken, and the platform does
not gate each of them. The platform's side of the bargain is that the trail is
complete, immutable and exported — every release composes `artifact_released`
naming `override_event_id`, `head_seq`, `head_hash`, `artifact_evidence_id` and
`recipient_ref` (A1-3.3, A1-6.6); the export carries the A1-6.6 integrity
metadata and stamp; the override event itself carries a non-empty `reason`, is
attributed to a `usr_` principal, and is inside the digest (A1-6.5, A1-5.2). An
owner who needs a stricter rule compensates with their own review and logging of
`artifact_released`.

## Consequences

- **Positive:**
  - One authorization rule instead of two contradictory readings (C-01 closed);
    the control sits next to the hard stop, where SPEC §3 puts it.
  - A legitimate multi-artifact release needs exactly one human decision, which
    is what a service owner can actually operate.
  - The compensating control is cryptographic rather than procedural: what left
    the platform, under which decision, from which chain state, to whom, is
    reconstructable from the chain alone and cannot be edited afterwards.
  - No clock-driven expiry inside a safety path (the override dies on a chain
    event, not on a timer), so no `time`-dependent branch in enforcement.
- **Negative:**
  - Adversarial finding **A15 (insider abuse) is not mitigated by the platform**
    for this control: an admin who overrides once can release several artifacts,
    and detection is after the fact. The product owner accepted this explicitly.
  - Q11's literal "operator override" is narrowed to admin; recorded here so the
    narrowing is a decision and not a drift.
  - The compensating-logging argument is only as strong as the chain's own
    integrity: if the tail can be truncated, `artifact_released` rows can
    disappear with it (see follow-ups).
- **Follow-ups:**
  - **A5** MUST gate `integrity_override` on the admin role and return
    `forbidden` (A0-3.1, principal-level, naming no object) to an operator; the
    machine-principal exclusion list already forbids it (Q6, A1-2.7, A1-3.4).
  - **D5 / A1 §6.16** (out-of-band head anchoring over the ADR-0012 §3 signed
    webhook) is now load-bearing for this decision, not merely nice to have: the
    accepted risk is "the owner can see everything", which requires that the
    `artifact_released` trail cannot be truncated by an attacker with store
    access. If D5 is declined, the residual risk wording in A1-6.6 must say so.
  - **A SaaS/multi-tenant offering MUST reopen this ADR**: per-customer
    separation of override authority, and four-eyes for `scope:"export"`
    (adversarial A15), are not expressible in a single-tenant admin role.
  - Contract-test suite (the `contracts/README.md` merge gate) must carry
    `TestOverrideIsAdminOnly`, `TestOperatorCannotOverrideAnyBreak`,
    `TestEveryReleaseUnderOneOverrideIsChained`,
    `TestReleaseWithoutLiveOverrideRefused`, `TestOverrideDiesAtNextBreak`,
    `TestArtifactReleasedNamesItsOverride` — the positive **and** negative half
    per AGENTS.md and the README's safety-path pairing bullet.
