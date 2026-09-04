# Session 2026-09-04 — Program layout + build rules + session skill

- **Date:** 2026-09-04
- **Session id:** 01a06cf4-51a4-708f-a305-150a5d2ca1a5
- **Model:** qwen-token-plan-individual/qwen3.8-max
- **Goal:** Backlog session 2 — decide the Go program layout (ADR-0021),
  extend the build rules in `AGENTS.md`, and add a pi skill
  (`start-session`) that codifies how sessions are started and closed.

## Done this session

- Reviewed all 20 ADRs: consistent, all Accepted; no conflicts with
  SPEC.md found (ADR-0017/0018 carry a stale "proposed" wording in the
  Deciders line — cosmetic only).
- Housekeeping from `next_steps.md`: unpushed-commits item resolved
  (branch is up to date with origin); `gh` CLI not available in this
  environment → PRs are created via `git push -o pull_request.create`.
- Drafted the program layout (backlog session 2). Product owner: the
  layout is a guideline document, **not an ADR** → `DESIGN.md` created
  (normative): module/cmd/internal structure, layering, plus concrete
  design guidelines — simplicity first / no over-engineering, function,
  interface, error, concurrency, config, test, and naming rules.
- `AGENTS.md`: `DESIGN.md` added to the read-before-acting list and
  referenced as binding; detailed rules live in DESIGN.md, not duplicated.
- `WORKFLOW.md`: GitHub branch protection is not possible on our hosting;
  the no-direct-push rule is enforced by discipline + agent compliance
  with SPEC/AGENTS/DESIGN/WORKFLOW.
- New skill `.pi/skills/start-session/SKILL.md`: session start/close
  ritual (read order, tracker template, usage snapshots, workflow).

## Decisions

- Program layout + design guidelines live in `DESIGN.md` (normative
  guideline file), not in an ADR — product owner directive 2026-09-04.
- Branch `layout/program-layout` → PR per WORKFLOW.md (`agent-built`).

## Open questions carried forward

- Architect review pass over DESIGN.md/AGENTS.md/WORKFLOW.md (this PR).
- Handover-contract package placement (backlog session 1 feeds it).
- GitHub branch protection still to be enabled by the product owner.
- `qwen3.8-flash` availability check for `sleipnir-implementer`.

## Token usage

| total input | uncached input | cache read | cache write | output | reasoning |
|---|---|---|---|---|---|
| 0 | 0 | 0 | 0 | 0 | 0 |

_(run `sessions/update-usage.sh sessions/2026-09-04-program-layout.md` at session end)_
