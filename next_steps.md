# NEXT STEPS — session 2026-09-04

Goal of the day: **backlog sessions 1 + 2 — handover contracts and program
layout**, executed by the build agents (`sleipnir-principal` spawning
`sleipnir-implementer` in parallel), delivered as the first real PR.

## 0. Housekeeping before anything else

- [ ] **Push is pending:** local `main` has unpushed commits (from the
      2026-09-03 session). Either push manually or set up credentials
      (SSH key / PAT / `gh auth login`) for this environment.
- [ ] **GitHub branch protection:** enable PR-only pushes on `main` in the
      repo settings so `WORKFLOW.md` is actually enforced.
- [ ] **Model check:** quick test spawn of `sleipnir-implementer` to
      confirm `qwen3.8-flash` resolves on the provider.

## 1. Context to load (read order)

1. `SPEC.md` (normative; constraints C1–C11)
2. `adr/` — especially ADR-0007 (topology), ADR-0016 (context graph),
   ADR-0017 (spawn broker), ADR-0019 (errors/logging)
3. `sessions/BACKLOG.md` — items 1 & 2 carry the open questions
4. `AGENTS.md` — build rules for agents

## 2. Session 1 — handover contracts (design first, ~decisions)

Decide, then record as ADR(s):
- Handover = **view over the engagement context graph** (ADR-0016): define
  the stage views — what `recon → research → exploitation → reporting`
  each receive (node/edge types, summaries, size limits).
- Summarization duty of workers: raw output stays in the event log;
  handovers carry references + summaries only.
- Schema for findings/hypotheses nodes (fields, severity, confidence,
  provenance).
- Contract shape: Go types + JSON representation (must serve both the
  internal API and agent prompts).

## 3. Session 2 — program layout

- Go module + package tree under stdlib-only + pgx exception:
  `cmd/` (platform, remote-agent), `internal/` (api, auth, policy,
  store, graph, events, orchestrator, agentloop, broker, notify, errs,
  llm, tools).
- `internal/errs` + slog setup as the **foundational first work package**
  (ADR-0019) — implementable immediately with tests (origin-function
  capture, redaction negative tests).
- Where Dockerfiles/compose live; image list per SPEC §10.

## 4. How to run it

1. Architect (this persona) + product owner agree on the contract
   decisions above (they are architecture, so decide them in chat first).
2. Delegate to `principal`:
   *"Implement backlog sessions 1+2: define the handover-contract types
   and the program layout; build `internal/errs` + logging setup with
   tests. Work on a branch, open a PR per WORKFLOW.md."*
3. Principal decomposes and fans out implementers (disjoint packages:
   e.g. `internal/errs`, contract types, layout scaffold).
4. Review the PR together; merge; ADRs for anything newly decided.

## 5. Definition of done for tomorrow

- [ ] Handover-contract ADR accepted
- [ ] Program-layout ADR accepted
- [ ] PR #1 open: layout scaffold + `internal/errs` + tests green
- [ ] Session tracker entry for 2026-09-04 (with token usage)

## Carry-forward open questions

- `qwen3.8-flash` availability check (see housekeeping)
- Approval-timeout notification edge cases → UI design session
- Masking scanner design → security hardening session
- Pi offline buffering → remote agent session
