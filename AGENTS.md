# AGENTS.md — rules for AI agents building Sleipnir

This repository's platform code is (co-)built by coding agents. Every agent
working in this repo MUST follow this file.

## Read before acting

1. `SPEC.md` — normative platform specification.
2. `adr/` — binding decisions; never violate an Accepted ADR.
3. `sessions/BACKLOG.md` — current work queue and open questions.

## Hard constraints (same as SPEC.md §2)

- Go **standard library only**. The only permitted dependency is the
  pinned/vendored exception list in ADR-0010 (currently: pgx). Any change
  to that list requires a new ADR first.
- UI via HTMX + `html/template` only. No frontend frameworks, no Node
  build step.
- Do not introduce Docker-incompatible runtime assumptions.

## Go style — idiomatic Go is mandatory

- Write idiomatic Go per *Effective Go* and the *Go Code Review Comments*:
  errors are values (handled explicitly, never discarded with `_`), early
  returns over deep nesting, small interfaces (accept interfaces, return
  structs), `context.Context` as first parameter where cancellation/IO is
  involved, table-driven tests, no `init()` abuse, no panics for control
  flow, no cleverness for its own sake — the next reader is an agent.
- `gofmt`/`go vet` clean is a merge gate.
- Middleware = `func(http.Handler) http.Handler`; prefer stdlib patterns
  (`io.Reader`, `encoding/json`, `net/http`) over invented abstractions.

## Program structure (ADR-0021) — binding

- One module (`github.com/daten-krake/sleipnir`); thin `cmd/` binaries
  (`platform`, `orchestrator`, `worker`, `remote-agent`): a `main` does
  only flag/env parsing, wiring, and graceful shutdown — no logic.
- All logic lives in `internal/` (foundation → domain → services →
  application edges); dependencies point downward only; `api`/`ui` are
  never imported by services; `internal/policy` imports no other service.
- The pgx exception (ADR-0010) is imported **only** in
  `internal/store/postgres`; every other package talks to store interfaces.
- Deployment assets live under `deploy/` (Dockerfiles, compose); tests are
  colocated with the code they test.

## Build rules

- **No package-level mutable state.** Dependencies are injected through
  constructors (`NewX(deps...)`); accept interfaces, return structs.
- **Documentation:** every package, exported type, function, and constant
  carries a doc comment that starts with the identifier name. Every
  package has a package comment stating its responsibility and its layer.
- **Hermetic unit tests:** unit tests must not need network, Docker, or a
  database; integration/contract tests that do are gated behind an
  explicit opt-in (environment variable) and say so in a comment. Tests
  are deterministic — no `time.Sleep` synchronization; inject clocks.
- **Concurrency:** every goroutine has an owner and a termination path via
  `context.Context`; no leaked goroutines, no unbounded channels. Shared
  state is protected explicitly — document which lock guards what.
- **Configuration:** environment variables, parsed once at the `cmd/`
  edge and passed down as typed config structs; no `os.Getenv` below
  `cmd/`.

## Error and logging conventions (ADR-0019) — binding

- All errors are created/wrapped through `internal/errs` (captures the
  originating function via `runtime.Caller`). Never use bare `errors.New`/
  `fmt.Errorf` for returned errors.
- Every returned error must read as
  `component.Function: what was attempted: key identifiers: cause` and be
  self-contained for troubleshooting from logs alone.
- Structured logging via stdlib `log/slog` with subsystem loggers and
  correlation attributes (engagement, run, job, node).
- Log-or-return, never both: an error is logged exactly once, where it is
  handled/decided.
- Secrets (credentials, tokens) never appear in errors or logs; use the
  redaction helpers and write tests proving it.

## Safety-critical code = high review bar

The following areas are hand-rolled security code and are the most likely
place for agent-introduced vulnerabilities. They require: RFC/OWASP test
vectors in tests, table-driven edge cases, and a dedicated review pass
before merge.

- Authentication & sessions (PBKDF2, TOTP, cookies)
- Scope/blacklist/approval enforcement (the platform core safety path)
- Token handling for job/agent API credentials
- Streaming parsers (LLM SSE/JSON, Postgres rows) and all untrusted-input
  parsing (tool output, webhook payloads, file uploads)
- Anything touching the spawn broker or container lifecycle

## Working conventions

- Contract-first: implement against the interfaces in `internal/` as
  defined by SPEC.md; if a contract is missing or wrong, raise it (new ADR
  or backlog item) before improvising.
- Every new architectural decision → ADR via `adr/TEMPLATE.md`.
- Update `sessions/` records when you work; refresh token usage with
  `sessions/update-usage.sh`.
- Tests are mandatory for enforcement logic: for every safety rule there
  must be a test proving the rule holds *and* a test proving the bypass
  attempt fails.
- Keep changes reviewably small; one concern per change.
