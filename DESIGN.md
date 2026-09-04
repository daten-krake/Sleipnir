# Sleipnir design guidelines

- **Status:** normative for every builder (human or agent), same as
  `AGENTS.md`. Companion to `SPEC.md`. Where this file and an ADR
  disagree, **the ADR wins** and this file must be fixed.
- **Scope:** how the code is organized and how code is written. Product
  behavior lives in `SPEC.md`; decisions live in `adr/`.

## 1. Program structure

- One Go module `github.com/daten-krake/sleipnir`, pinned to the latest
  stable Go release (≥1.26), committed `vendor/` (`go mod vendor`).
  `go.mod` carries only the ADR-0010 exception list (currently pgx).
- **`cmd/` — thin binaries, one directory per binary:** `platform`
  (core service: UI + API + broker + gateway), `orchestrator` (per-job
  mastermind), `worker` (task runtime inside tool images), `remote-agent`
  (Pi node). A `main` does only flag/env parsing, wiring, and graceful
  shutdown. No logic in `cmd/`.
- **`internal/` — one concern per package:**
  - Foundation: `errs`, `log` (ADR-0019 primitives; zero internal imports).
  - Domain types: `events`, `graph`, `tools` — types + invariants, no I/O.
  - Services: `store` (interfaces) + `store/postgres` (the **only** place
    importing pgx), `auth`, `policy` (scope/blacklist/approval/hard-stop
    enforcement), `broker`, `llm` (provider contract + gateway + masking),
    `agentloop`, `notify`.
  - Application edges: `api` (`/api/v1` + SSE), `ui` (HTMX handlers +
    embedded templates/static), `orchestrator` (planning logic consumed by
    `cmd/orchestrator`).
- **Layering — dependencies point downward only:**
  `cmd/` → application edges → services → domain types → foundation.
  `api`/`ui` are never imported by services; `policy` imports no other
  service (it is the chokepoint others call); no import cycles.
- Deployment assets live under `deploy/`: `deploy/docker/<image>.Dockerfile`
  (SPEC §10 image list) and `deploy/compose/`.
- Tests are colocated (`*_test.go` next to the code); cross-package
  contract tests live with the contract owner.

## 2. Simplicity first — do not over-engineer

- Solve the problem at hand. No abstraction for a hypothetical future, no
  speculative config knobs, no "we might need this" code. YAGNI beats
  flexibility.
- A small amount of duplication is cheaper than a premature abstraction.
  Extract when the same pattern appears the third time, not the first.
- Every interface, indirection layer, and wrapper must justify itself with
  a concrete second consumer **or** an ADR that mandates swappability
  (store, LLM provider, notify, spawn). One implementation + one consumer
  = call the concrete thing directly.
- Prefer the boring stdlib solution. Cleverness is a review blocker, not
  an asset — the next reader is an agent.
- If a function needs a long comment to explain why it exists, delete it
  or inline it.

## 3. Functions

- One job, one level of abstraction, guideline ≤ ~50 lines. Split when a
  block can be named.
- Early returns over nesting; no `else` after a return.
- Few parameters (≤3–4). More than that: pass an options/config struct.
- Signature order: `ctx context.Context` first (where there is IO or
  cancellation), results before `error`, `error` always last.
- No naked returns; named results only when they genuinely clarify.
- Do not mutate inputs; if unavoidable, say so in the doc comment.
- Every exported function carries a doc comment starting with the function
  name: what it does, and non-obvious failure modes.

## 4. Interfaces and types

- Accept interfaces, return structs.
- Interfaces are small (1–3 methods) and defined **at the consumer**, not
  the producer.
- Domain types carry their invariants: construct via `NewX(...)` that
  validates; never hand out half-built values.
- No package-level mutable state. Dependencies are injected through
  constructors.

## 5. Errors and logging

Per ADR-0019, concretely:

- Create/wrap errors only through `internal/errs` (`errs.New`, `errs.Newf`,
  `errs.Wrap`, `errs.Wrapf`); never `errors.New`/`fmt.Errorf` for returned
  errors. Attach an error kind at creation.
- Every returned error reads
  `component.Function: what was attempted: key identifiers: cause`.
- Log-or-return, never both; the decision point logs exactly once, with
  correlation attributes (`engagement`, `run`, `job`, `node`).
- Redact secrets before they reach any error string or log line.

## 6. Concurrency

- Every goroutine has an owner and a termination path via `context.Context`;
  no leaked goroutines, no unbounded channels.
- Shared state is protected explicitly; document which lock guards what.
- Channels for ownership transfer, mutexes for state — pick one, keep it
  local to the owning package.

## 7. Configuration

- Environment variables, parsed once at the `cmd/` edge into typed config
  structs and passed down. No `os.Getenv` below `cmd/`. No config files in
  v1 unless an ADR says otherwise.

## 8. Tests

- Table-driven, colocated, deterministic — no `time.Sleep` synchronization,
  inject clocks.
- Hermetic unit tests: no network, no Docker, no database. Tests that need
  real infrastructure are integration tests gated behind an explicit opt-in
  (environment variable) and say so in a comment.
- Every safety/enforcement rule needs a test proving it holds **and** a
  test proving the bypass attempt fails.
- RFC/OWASP vectors for anything on the AGENTS.md high-review list.

## 9. Naming and style

- Effective Go / Go Code Review Comments are the review bar.
- MixedCaps, acronyms keep case (`ID`, `HTTP`, `URL`, `TOTP`).
- No stutter: `policy.Check`, not `policy.PolicyCheck`; `errs.New`, not
  `errs.NewErr`.
- Short names in small scopes (`ctx`, `i`, `r`); descriptive names grow
  with scope size.
- Package comments state the package's responsibility and its layer (§1).
