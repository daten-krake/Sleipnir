# Session 2026-09-04 — Program layout + build rules + session skill

- **Date:** 2026-09-04
- **Session id:** 01a06cf4-51a4-708f-a305-150a5d2ca1a5
- **Model:** qwen-token-plan-individual/qwen3.8-max
- **Goal:** Backlog session 2 — decide the Go program layout as a
  normative guideline file (`DESIGN.md`), extend the build rules in
  `AGENTS.md`, and add a pi skill (`start-session`) that codifies how
  sessions are started and closed.

## Done this session

- Reviewed all 20 ADRs: consistent, all Accepted; no conflicts with
  SPEC.md found (ADR-0017/0018 carry a stale "proposed" wording in the
  Deciders line — left untouched: ADRs are immutable once Accepted).
- Housekeeping from `next_steps.md`: unpushed-commits item resolved
  (branch is up to date with origin); `gh` CLI not available in this
  environment → PRs are created via `git push -o pull_request.create`.
- Drafted the program layout (backlog session 2). Product owner: the
  layout is a guideline document, **not an ADR** → `DESIGN.md` created
  (normative): module/cmd/internal structure, layering, plus concrete
  design guidelines — simplicity first / no over-engineering, function,
  interface, error, concurrency, config, test, and naming rules.
- `AGENTS.md`: `DESIGN.md` and `WORKFLOW.md` added to the read-before-
  acting list; DESIGN.md declared binding. Error/logging convention stays
  canonical in AGENTS.md (per ADR-0019); DESIGN.md §5 points at it.
- Architect review pass applied: binary-boundary security rule +
  `apiclient` + `evidence` packages, `internal/log`→`logging` rename,
  consumer-interface exception for mandated seams, supply-chain PR gate
  (`go.mod`/`vendor/` check in WORKFLOW.md), SPEC §12.2 context list,
  backlog supersede note, GitHub-correct PR creation in the skill.
- `WORKFLOW.md`: no-direct-push rule. Branch protection initially deemed
  unavailable, then **enabled by the product owner 2026-09-04** → rule is
  enforced mechanically (GitHub branch rules) plus by agent compliance
  with SPEC/AGENTS/DESIGN/WORKFLOW.
- New skill `.pi/skills/start-session/SKILL.md`: starting a session
  triggers the ritual — housekeeping check, read SPEC → ADRs → DESIGN →
  backlog → AGENTS/WORKFLOW in fixed order, collect **all open tasks**
  (backlog queue, standing questions, next_steps items, carried-forward
  questions, Proposed ADRs), report them, confirm the goal, then open the
  tracker; plus the closing ritual.
- PR #1 pushed + created via GitHub API, labeled `agent-built` (branch
  protection enabled by PO the same day; git credential helper wired to
  `gh`; push options confirmed NOT to work on GitHub).
- **Design interview (design day — no code):** architect + principal
  prepared researched briefs; product owner decided all backlog-session-1
  topics plus API/token/integrity contracts (Q1–Q15, full record below).
- `WORKFLOW.md`: added the super-minimal/junior-level work-package rule
  (PO directive).
- Backlog: session 1 marked decided; new queue items 9 CI/CD, 10
  monitoring/observability, 11 model benchmarking & drift detection.

## Decisions

- Program layout + design guidelines live in `DESIGN.md` (normative
  guideline file), not in an ADR — product owner directive 2026-09-04.
- Branch `layout/program-layout` → PR #1 (`agent-built`).
- Branch protection enabled by PO → WORKFLOW.md mechanically enforced.

### Design interview — locked decisions (Q1–Q15)

- **Q1 Handover view mechanics:** fixed server-computed stage views +
  capped 1-hop neighbor drill-down. **No free-form query endpoint in v1**
  (additive later, no data migration needed). Review trigger: after first
  real HTB engagement.
- **Q2 Findings/hypotheses:** two separate typed node kinds, `Finding`
  (severity, confidence, status, evidence refs) and `Hypothesis` (claim,
  basis, status); revision via `supersedes` edges, never overwrite.
- **Q3 Graph write strictness:** hard reject on schema violation +
  bounded `Attrs` escape hatch; platform enforces size caps (truncate +
  marker / `summary_too_large`), never trusts worker discipline.
- **Q4 Size budgets (contract constants, change only via ADR):** stage
  view ≤ 64 KiB JSON / ≤ 500 nodes; node summary ≤ 512 B; finding
  summary ≤ 2 KiB; stage summary ≤ 2 KiB.
- **Q5 Quarantined discoveries:** included in reporting handover, marked
  *not tested*, non-actionable; operator can remove them from the report.
- **Q6 Machine-token exclusion (permanent contract rule):** machine
  principals never mutate engagement/scope/blacklist, never decide
  approvals, never touch kill controls, user/settings/tool-registry
  writes, or anything cross-engagement. Three enforcement layers:
  scopes exclude the verbs → middleware hard-blocks the endpoints for
  any machine principal → per-request engagement/run binding.
- **Worker principal (addendum):** report-only — `events:append`,
  `evidence:upload`, `task:result`; no graph access at all; graph writes
  happen platform-side on ingest with platform-stamped provenance.
- **Q7 v1 API cuts:** no external automation keys in v1 (scope design
  stays additive for later); `GET /engagements/{id}/findings` JSON export
  **included** in v1.
- **Q8 Token rotation:** job-lifetime orchestrator token, stateful
  (hash-only storage), instant revocation on hard stop / run end. No
  rotation in v1.
- **Q9 Node identity:** `slp_node_…` token over mesh transport; encrypted
  local storage; per-engagement pairing; revoke/wipe commands. Node-theft
  residual risk **accepted by PO**.
- **Q10 Approval execution:** uniformly single-use — one approval =
  exactly one execution; failed/changed actions re-queue.
- **Q11 Event integrity:** hash chaining from day one (write-side),
  per-engagement chains with genesis events; verification at startup and
  before export; on failure: customer-facing exports blocked, internal
  views flagged, **explicit operator override allowed — logged as event,
  report stamped "integrity verification failed"**; no external timestamp
  anchoring in v1 (self-recorded timestamps suffice).
- **Q12 API surface:** full `/api/v1` from day one, gated by Q6 rule.
- **Q13 Versioning:** additive-only `/api/v1`, clients ignore unknown
  fields (contract-tested), breaking changes only via `/api/v2`; formal
  deprecation window deferred until automation keys ship.
- **Q14 Spawn request:** orchestrator sends tool id+version (registry
  allowlist), capped untrusted task description, resource ask; platform
  derives image digest + risk tier from registry; risky spawns wait on
  approval.
- **Q15 Build sequencing:** A0 conventions + A1 events + A2 graph frozen
  first (principal drafts A0), then fan-out; work packages super-minimal
  / junior-level (now a WORKFLOW.md rule); no coding until contracts
  frozen.

## Open questions carried forward

- Contract documents A0–A8 to be drafted (next session; principal A0).
- CI/CD, monitoring/observability, model benchmark & drift detection →
  backlog items 9–11 (added by PO 2026-09-04).
- `qwen3.8-flash` availability check for `sleipnir-implementer`.

## Resolved this session

- Architect review findings M1/M2 + S1–S9 + N1–N7 applied.
- Branch protection **enabled** by PO 2026-09-04; WORKFLOW.md enforced.
- PR #1 created via GitHub API (push options don't work on GitHub),
  labeled `agent-built`.

## Token usage

| total input | uncached input | cache read | cache write | output | reasoning |
|---|---|---|---|---|---|
| 6044882 | 272210 | 5772672 | 0 | 97009 | 28050 |

_(run `sessions/update-usage.sh sessions/2026-09-04-program-layout.md` at session end)_
