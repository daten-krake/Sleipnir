# ADR-0022: Provenance evidence grade replaces finding-level confidence

- **Status:** Accepted
- **Date:** 2026-09-21 (signed by the product owner in PR #2, item D2; Q2
  dates from 2026-09-04, the A2 deviation from 2026-09-11)
- **Deciders:** product owner (signature, PR #2 item D2); architect (A2-2.7,
  A2-5.6); principal build engineer (fix-plan ruling on the `confidence`
  new-node fork); principal review findings P-42, S-05

## Context

Q2 (2026-09-04, `sessions/2026-09-04-program-layout.md`) locked the shape of the
two analysis node kinds: *"two separate typed node kinds, `Finding` (severity,
**confidence**, status, evidence refs) and `Hypothesis` (claim, basis, status);
revision via `supersedes` edges, never overwrite."*

ADR-0016 §1, Accepted and older, independently requires that *"every node/edge
carries **provenance** (agent, tool, event-id, timestamp, **confidence**) — graph
content is evidence, not opinion."*

When A2 made the graph contract concrete, the two uses of the word collided.
Q2's `confidence` is a property of a **finding**: how sure the writer is of its
own claim. ADR-0016's is a property of an **observation**: what kind of evidence
backs this graph content, tied to the event that recorded it. A2-2.7 and A2-5.6
implemented the second reading — a closed evidence grade `observed · inferred ·
verified` carried on every provenance entry of every node and edge — and
therefore dropped `confidence` from the `finding` node type. That is a **change
to a locked product-owner decision**, so A2 §6 item 13 refused to treat it as a
confirmation and required a signature before A2 could freeze.

Two review findings made the choice urgent rather than cosmetic:

- **P-42 / S-05** — as first drafted, the grade `verified` was **unreachable**:
  nothing defined what would make an observation independent, so no writer could
  ever legitimately produce one and every finding would have looked like a
  single unverified claim.
- The fix plan's `confidence` new-node fork — whether a re-observation with a
  different confidence is a *new node* or an *amended* one — had to be resolved
  before the content fingerprint (A2-4.8) could be defined at all, because the
  answer decides whether the grade is inside the fingerprint.

Forces and constraints: ADR-0016 §1's evidence-not-opinion principle is
Accepted and outranks a session decision (authority: ADR > SPEC > DESIGN >
contract, `contracts/README.md`); A0-6.5 makes versioning **additive only**, so
removing a field after Freeze is not available — this was the last cheap moment
to decide; the report is a customer-facing legal artifact (SPEC §5 step 9,
ADR-0009 §5), so any grade it prints must be defensible; and A2-4.7 already
bounds provenance to a list of at most 8 entries ordered by event `seq`, which
is the structure the grade has to live in.

## Options considered

1. **A `confidence` field on the `finding` node type, exactly as Q2 recorded**
   (adjective scale `low`/`medium`/`high`) — matches Q2 literally and needs no
   signature; but the value is asserted by the same principal that asserts the
   claim, is tied to no event, and cannot be checked. A worker could raise its
   own confidence, which is precisely what ADR-0016 §1 forbids in spirit. It
   also leaves a second meaning of "confidence" next to ADR-0016's provenance
   confidence: two sources of truth for one fact, with no rule for when they
   disagree.
2. **Both: the provenance grade *and* a finding-level adjective** — a report
   could show whichever is convenient; but one fact in two encodings that can
   contradict each other, and the contradiction has no resolution rule. A2-2.6
   already rejects exactly this shape for supersession ("a second encoding would
   be a second source of truth"), so accepting it here would be inconsistent
   within the same document.
3. **The provenance evidence grade only** — a closed three-value scale on every
   provenance entry of every node and edge, with the node's grade defined as the
   highest in its list and `verified` reachable only through a second,
   independent observation. Q2's intent is discharged (every finding *does*
   carry a confidence — on its mandatory provenance), but the value is a
   statement about the evidence rather than about the writer's certainty.

## Decision

**Option 3, signed by the product owner (PR #2 item D2).** Q2's
`Finding (severity, confidence, status, evidence refs)` is amended to
`Finding (severity, status, evidence refs)`: there is **no** `confidence` field
on the `finding` node type and **no** adjective certainty scale anywhere in A2.

"Confidence" in the Sleipnir graph means the **provenance evidence grade** of
ADR-0016 §1, and it is:

- a **closed** three-value scale — `observed` (directly seen in tool output),
  `inferred` (derived by reasoning from what was seen), `verified` (independently
  re-observed) — with no numeric form and no per-contract extension (A0-6.4);
- carried **once per provenance entry**, on **every node and every edge**, not
  only on findings — hypotheses, which Q2's finding-centric wording never
  covered, get the same treatment;
- **inside** the content fingerprint (A2-4.8), so a grade cannot be changed
  without a revision and a `supersedes` edge;
- raisable to `verified` **only** when the new provenance entry's `event_id`
  differs from every existing one **and** its `task_id`/`agent_node_id` differ —
  a second, independent observation (A2-5.6). A node that only ever observed
  itself stays `inferred`.

This narrows Q2 deliberately and is recorded here so the narrowing is a decision
and not a drift; ADR-0016 §1 is unchanged and is the authority the grade
implements.

## Consequences

- **Positive:**
  - One fact, one encoding: no `confidence` field can disagree with a
    provenance grade, and ADR-0016 §1's "evidence, not opinion" becomes
    mechanically enforceable instead of aspirational.
  - **P-42 / S-05 closed** — `verified` is reachable and is reachable *only* by
    the independence rule, so the strongest grade in the system cannot be
    self-asserted by a single worker or a single run.
  - Every grade cites an `event_id`, so a customer report can defend
    "verified" as "independently observed twice, here are the two events" —
    which an adjective never could.
  - Uniform across nodes and edges, so A3 stage views and the report builder
    need one rendering rule, not one per node kind.
  - Decided before Freeze, so no additive-only violation (A0-6.5) is created and
    no migration is owed.
- **Negative:**
  - Q2's literal node shape changed. Anyone reading the 2026-09-04 tracker will
    look for `Finding.confidence` and not find it; this ADR and A2 §6 item 13
    are the trail.
  - Three values are **coarser** than a number. A model's calibrated
    probability, or a scanner's certainty score, has no home in the schema; it
    may ride in `attrs` as an uninterpreted value (A2-2.5's `cvss_v3_x10` is the
    precedent) and the platform MUST NOT route on it.
  - The report must explain the scale to a customer, because "inferred" reads as
    weaker than it is — an inferred finding may be entirely correct, it was
    simply not directly observed. That wording is the report session's, not A2's.
- **Follow-ups:**
  - **Report and UI sessions** MUST render the grade itself, never a
    re-invented adjective, and MUST show the provenance entries behind a
    `verified` grade.
  - **A3** stage views carry the grade in their compact node refs; the D4
    composition rule (A0-7.10, confirmed 2026-09-21) must budget bytes for it —
    at ~131 B per node there is no room for a per-entry provenance list, which is
    an argument for A3's "compact refs, full detail in the capped 1-hop
    drill-down" recommendation.
  - **Contract-test suite** (`contracts/README.md` merge gate) carries
    `TestVerifiedRequiresIndependentObservation`,
    `TestSelfObservedNodeStaysInferred`, `TestNodeDedupKeepsEveryObservation`,
    `TestReplayedEventAddsNoProvenanceEntry` and
    `TestProvenanceListIsOrderedByEventSeq` — the positive and the bypass half
    per AGENTS.md.
  - Introducing a **numeric** confidence later reopens the two-sources problem
    this ADR closes and requires a new ADR, not an additive field.
