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
- `WORKFLOW.md`: GitHub branch protection is not possible on our hosting;
  the no-direct-push rule is enforced by discipline + agent compliance
  with SPEC/AGENTS/DESIGN/WORKFLOW.
- New skill `.pi/skills/start-session/SKILL.md`: starting a session
  triggers the ritual — housekeeping check, read SPEC → ADRs → DESIGN →
  backlog → AGENTS/WORKFLOW in fixed order, collect **all open tasks**
  (backlog queue, standing questions, next_steps items, carried-forward
  questions, Proposed ADRs), report them, confirm the goal, then open the
  tracker; plus the closing ritual.

## Decisions

- Program layout + design guidelines live in `DESIGN.md` (normative
  guideline file), not in an ADR — product owner directive 2026-09-04.
- Branch `layout/program-layout` → PR per WORKFLOW.md (`agent-built`).

## Open questions carried forward

- Architect review pass over DESIGN.md/AGENTS.md/WORKFLOW.md — done this
  session (findings M1/M2 + S1–S9 + N1–N7 applied).
- Handover-contract package placement (backlog session 1 feeds it).
- Branch protection: PO directive 2026-09-04 — not possible on current
  hosting; no-direct-push enforced by agent discipline (WORKFLOW.md §1).
  Architect note: GitHub free plans DO support branch protection on
  private repos since 2024 — verify repo-admin permissions; if available,
  enable and restore enforced wording.
- PR #1 to be pushed/created once `gh` credentials exist (PR creation via
  `gh pr create` or GitHub UI — push options do not work on GitHub).
- `qwen3.8-flash` availability check for `sleipnir-implementer`.

## Token usage

| total input | uncached input | cache read | cache write | output | reasoning |
|---|---|---|---|---|---|
| 0 | 0 | 0 | 0 | 0 | 0 |

_(run `sessions/update-usage.sh sessions/2026-09-04-program-layout.md` at session end)_
