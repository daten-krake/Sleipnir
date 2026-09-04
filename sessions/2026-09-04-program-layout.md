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
- Drafted **ADR-0021 (Proposed)**: program layout — one Go module,
  thin `cmd/` binaries (platform, orchestrator, worker, remote-agent),
  `internal/` package map with layering rules, pgx quarantined to
  `internal/store/postgres`, deployment assets under `deploy/`.
- `AGENTS.md`: added binding sections — program structure & layering,
  state/DI rule, documentation rule, hermetic-tests rule, concurrency
  rule, config rule.
- New skill `.pi/skills/start-session/SKILL.md`: session start/close
  ritual (read order, tracker template, usage snapshots, workflow).

## Decisions

- ADR-0021 **Proposed** — awaiting product owner acceptance.
- Branch `layout/program-layout` → PR per WORKFLOW.md (`agent-built`).

## Open questions carried forward

- ADR-0021 acceptance (flip Status, adjust if owner objects).
- Handover-contract package placement (backlog session 1 feeds it).
- GitHub branch protection still to be enabled by the product owner.
- `qwen3.8-flash` availability check for `sleipnir-implementer`.

## Token usage

| total input | uncached input | cache read | cache write | output | reasoning |
|---|---|---|---|---|---|
| 0 | 0 | 0 | 0 | 0 | 0 |

_(run `sessions/update-usage.sh sessions/2026-09-04-program-layout.md` at session end)_
