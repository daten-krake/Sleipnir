# NEXT STEPS — session 2026-09-05

Goal: **freeze the contracts** so the build can fan out. Still no product
code until contracts A0–A2 are frozen and PR #1 is merged.

## 0. Housekeeping

- [ ] **Review + merge PR #1** (`layout/program-layout`): DESIGN.md,
      AGENTS.md/WORKFLOW.md rules, start-session skill. Merge makes the
      design guidelines binding.
- [ ] Model check: spawn `sleipnir-implementer` once to confirm
      `qwen3.8-flash` resolves (carry-over).

## 1. Context to load

Use the `start-session` skill ritual (fixed read order). New since last
session: the **Q1–Q15 decision record** in
`sessions/2026-09-04-program-layout.md` — it is the input to every
contract document.

## 2. Contract documents (principal leads)

Build order (approved by product owner 2026-09-04, Q15):

1. **A0 — cross-cutting conventions** (principal writes himself): ID
   scheme (`eng_`, `run_`, `evt_`, `node_`), canonical-JSON rules (used by
   BOTH hash chains and approval fingerprints), error-kind → HTTP map,
   pagination/cursor, RFC3339 UTC, ignore-unknown-fields rule.
2. **A1 — events contract** (envelope + taxonomy + day-one chain fields,
   per-engagement chains) and **A2 — graph contract** (closed node/edge
   kinds incl. Finding + Hypothesis, provenance, quarantine, no-secret-
   values rule) — the critical pair, provenance ties them.
3. Fan out (super-minimal, junior-level packages per WORKFLOW.md):
   A3 stage views · A4 `/api/v1` surface · A5 token contract ·
   A6 bounded query (deferred — contract stub only) · A7 spawn/fingerprint
   schemas · A8 config convention.
4. Merge gate: shared contract-test suite (JSON round-trip, unknown-field
   tolerance, negative cross-engagement tests).

Each contract doc: one reviewable doc + Go type sketch + JSON schema,
delivered as PR(s), labeled `agent-built`.

## 3. New design topics queued by product owner (backlog 9–11)

- **CI/CD pipeline design** (backlog 9): PR gates as pipeline steps,
  reproducible vendored builds, image build/pin/sign (A10), release +
  rollback.
- **Monitoring & observability** (backlog 10): slog JSON export, health
  endpoints, per-run metrics, alerting via signed webhooks, SIEM export.
- **Model benchmarking & drift detection** (backlog 11): fixed benchmark
  suite scored per model, drift thresholds + reaction (alert / quarantine
  from role matrix); data sources = per-run model snapshots (ADR-0014) +
  gateway egress metadata (ADR-0020).

Schedule these after contracts are frozen; all three are design sessions
(ADRs/docs), not code.

## 4. Definition of done for 2026-09-05

- [ ] PR #1 merged
- [ ] A0 written; A1 + A2 drafted as PRs
- [ ] Fan-out packages briefed (super-minimal) or already fanned out
- [ ] Session tracker + next_steps updated

## Carry-forward open questions

- `qwen3.8-flash` availability check
- Handover-contract package placement (`internal/handoff`?) — falls out
  of A3
- CI/CD, observability, model benchmarks → backlog 9–11
