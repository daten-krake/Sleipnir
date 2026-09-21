# Contracts

Normative interface contracts for the Sleipnir build. Together with
`SPEC.md`, `adr/`, `DESIGN.md`, `AGENTS.md`, and `WORKFLOW.md` these are the
build input for every builder, human or agent — SPEC §12.1 (contract-first):
schemas are defined before the components that consume them, so builder
agents can work in parallel against stubs and contract tests.

**Authority:** ADR > SPEC > DESIGN > contract. A contract never contradicts
an Accepted ADR; it makes ADRs concrete and testable. Changing a **frozen**
clause or constant requires a new ADR (or an explicit product owner decision
recorded in a session tracker) plus a PR. **A0 gates every later contract;
where A0 and A1–A8 disagree on a cross-cutting convention, A0 wins and the
later document is defective.**

## Documents

| ID | Document | Scope | Status |
|----|----------|-------|--------|
| A0 | [`A0-conventions.md`](A0-conventions.md) | cross-cutting conventions: ids, canonical JSON, error kind → HTTP, pagination, time, unknown fields, size caps, field conventions | **Frozen** (2026-09-21, PR #2) |
| A1 | [`A1-events.md`](A1-events.md) | event envelope, closed taxonomy, hash chaining, write path | **Frozen** (2026-09-21, PR #2) |
| A2 | [`A2-graph.md`](A2-graph.md) | closed node/edge kinds, provenance, quarantine, validation, no-secret-values | **Frozen** (2026-09-21, PR #2) |
| A3 | stage views _(not started)_ | fixed per-stage handover views + capped 1-hop drill-down (Q1) | — |
| A4 | `/api/v1` surface _(not started)_ | endpoints, scopes, SSE mapping (Q12) | — |
| A5 | token contract _(not started)_ | machine principals, job/node tokens, exclusion list (Q6, Q8, Q9) | — |
| A6 | bounded query _(deferred)_ | no free-form query endpoint in v1 — stub only (Q1) | — |
| A7 | spawn + fingerprint schemas _(not started)_ | spawn request, action fingerprint (Q14, ADR-0017, ADR-0018) | — |
| A8 | config convention _(not started)_ | env → typed config at the `cmd/` edge (DESIGN §7) | — |

The A1–A9 labels are contract ids, not ADR numbers; the Q1–Q15 references
are the product owner decisions recorded in
`sessions/2026-09-04-program-layout.md`.

## Lifecycle

`Draft` → `Frozen` → `Implemented`.

- **Draft** — in progress; anything may change; **no code may be written
  against it**.
- **Frozen** — merged and accepted by the product owner. Code is written
  against it and the shared contract-test suite (below) locks it.
- **Implemented** — the implementing package and its contract tests exist
  and pass.

**A0, A1 and A2 reached `Frozen` on 2026-09-21** (PR #2): the product owner
answered decisions **D1–D9** and accepted all three blocks of the 46-item
confirm checklist. D1 is recorded in ADR-0021, D2 in ADR-0022; the remaining
answers are recorded in `sessions/2026-09-21-freeze-and-first-code.md` and cited
inline by each document's §6, which is now a decision record rather than a
question list. A frozen clause changes only by a new ADR or an explicit product
owner decision recorded in a session tracker, plus a PR.

Versioning is **additive only** within `/api/v1` (Q13): new fields, new
kinds, new endpoints are allowed; removing or repurposing them is not.
Breaking changes require `/api/v2`.

## Document shape

Every contract document uses the same sections, in this order:

1. **Header** — status (per the lifecycle above), owner, and the sources it
   implements (ADRs, SPEC sections, Q-decisions).
2. **Scope** — what this contract fixes, and explicitly what it does not.
3. **Normative clauses** — numbered and citable (`A0-1`, `A1-4.2`, …) with
   MUST / MUST NOT / MAY wording, so reviews, work packages, and tests can
   point at a clause instead of paraphrasing it.
4. **Types** — Go type sketch (illustrative, **not compiled**) plus JSON
   examples for every shape.
5. **Traceability** — table mapping each source decision to the clause that
   implements it.
6. **Open for product owner** — recommendations marked **PO confirm**.

Contracts contain no compiled code. Type sketches live in fenced blocks and
are transcribed into `internal/` packages (DESIGN §1) by the work packages
that implement them; the sketch is the source of truth for field names and
JSON shapes until that package is merged.

## Merge gate: shared contract-test suite

No package implementing a contract is merged before the shared contract-test
suite exists and passes. It tests, at minimum:

- **JSON round-trip** — marshal → unmarshal → equal for every contract type,
  including boundary sizes (A0 size caps) and empty/absent optional fields.
- **Canonical JSON stability** — the A0 canonical form is byte-identical
  across runs and platforms (key order, numbers, unicode, escaping) and
  matches the published test vectors. The event hash chain (Q11) and
  approval fingerprints (ADR-0018) both depend on this.
- **Unknown-field tolerance** — payloads with injected unknown fields decode
  successfully and the unknown fields are ignored (Q13).
- **Hard-reject validation** — writes violating the closed kind lists or the
  size caps are rejected with the A0 `validation` error kind and a
  self-contained message (Q3, ADR-0019 §2).
- **Cross-engagement negatives** — a read scoped to engagement B never
  returns engagement A data: events, graph nodes/edges, and stage views
  (SPEC C8, adversarial finding A12).
- **Secret-free serialization** — no secret material appears in serialized
  events, graph nodes, error bodies, or logs (ADR-0019 §5, ADR-0020).
- **Safety-path pairing** — every MUST / MUST NOT in A0–A8 that guards
  authorization, integrity, approval, quarantine, egress or the kill path
  carries a test id in the clause itself (`Tests: <Name>, <Name>`); one id MUST
  be a positive test of the rule and one MUST be a negative test of the bypass
  attempt named in the clause. A contract reaches `Frozen` only when every such
  clause has both. The suite fails if a safety-path clause has no test id.

The suite lives with the contract owner package (DESIGN §1: cross-package
contract tests live with the contract owner) and is referenced by name in
every implementing work package's acceptance criteria (WORKFLOW §5).
