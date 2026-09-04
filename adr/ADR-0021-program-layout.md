# ADR-0021: Program layout — one module, thin binaries, layered `internal/` packages

- **Status:** Proposed
- **Date:** 2026-09-04
- **Deciders:** architect proposing, product owner to confirm

## Context

Backlog session 2. SPEC §12 build model is contract-first and parallel:
builder agents work on disjoint packages at the same time. That requires a
fixed, boring, compiler-checked repository structure decided *before* the
first code lands. Forces:

- Four v1 binaries (platform, orchestrator, worker, remote-agent) share
  most of their code (agent loop, contracts, errs/log) — ADR-0007/0015.
- stdlib-only with one vendored exception (pgx, ADR-0001/0010): the
  exception must stay quarantined so a later swap (e.g. hand-rolled wire
  protocol) touches exactly one package.
- Safety core (scope/blacklist/approval, ADR-0005) must be a clearly
  isolatable, separately reviewable unit — `AGENTS.md` high-review bar.
- ADR-0019 makes `internal/errs` + slog setup the foundational first work
  package; everything else builds on it.

## Options considered

1. **One Go module, standard `cmd/` + `internal/` layout** — compiler
   enforces `internal/` visibility; single `go.mod` keeps the vendored
   exception auditable; one build/test pipeline for all binaries.
2. Multi-module repo (module per component) — cross-module versioning
   churn with no benefit at this size; rejected.
3. Flat package layout (`pkg/…` or root-level packages) — no enforced
   boundaries, invites import spaghetti between agent components; rejected.

## Decision

1. **One module** `github.com/daten-krake/sleipnir`, pinned to the latest
   stable Go release (≥1.26). `go.sum` + committed `vendor/`
   (`go mod vendor`); `go.mod` carries only the ADR-0010 exception list
   (currently pgx).
2. **`cmd/` holds thin binaries — one directory per binary:**
   `platform` (core service: UI + API + broker + gateway), `orchestrator`
   (per-job mastermind), `worker` (task runtime in tool images),
   `remote-agent` (Pi node). A `main` does only flag/env parsing, wiring,
   and graceful shutdown; no logic.
3. **`internal/` package map** (one concern per package):
   - Foundation: `errs`, `log` — ADR-0019 primitives; zero internal imports.
   - Domain types: `events`, `graph`, `tools` — types + invariants, no I/O.
   - Services: `store` (interfaces) + `store/postgres` (the **only** place
     that imports pgx), `auth`, `policy` (scope/blacklist/approval/hard-stop
     enforcement), `broker`, `llm` (provider contract + gateway + masking),
     `agentloop`, `notify`.
   - Application edges: `api` (`/api/v1` + SSE), `ui` (HTMX handlers +
     embedded templates/static), `orchestrator` (planning logic consumed by
     `cmd/orchestrator`).
4. **Layering rule — dependencies point downward only:**
   `cmd/` → application edges → services → domain types → foundation.
   `api`/`ui` are never imported by services; `policy` imports no other
   service (it is the chokepoint others call). No import cycles.
5. **Deployment assets live under `deploy/`:**
   `deploy/docker/<image>.Dockerfile` (platform, orchestrator, worker-base,
   worker-browser, remote-agent — SPEC §10 image list) and
   `deploy/compose/` for the platform stack.
6. **Tests are colocated** (`*_test.go` next to code); cross-package
   contract tests live with the contract owner.

## Consequences

- **Positive:** compiler-enforced boundaries; disjoint packages enable
  parallel implementer agents; pgx swap cost is one package; the safety
  core (`policy`) is isolatable for the high-review bar; boring and
  grep-able for agent builders.
- **Negative:** package splits are a bet — some boundaries will prove
  wrong and need rework (cheap while the tree is small); layering beyond
  import cycles is enforced by review, not tooling, until an import linter
  exists.
- **Follow-ups:**
  - Handover-contract types (backlog session 1) get their own package
    (likely `internal/handoff`) once defined.
  - Config handling convention (env-based, parsed in `cmd/`) — confirm in
    the first platform work package; promote to this ADR or a new one.
  - Promote Status to Accepted with product owner confirmation; first
    implementation PR is the layout scaffold + `internal/errs` + `internal/log`.
