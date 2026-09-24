# AGENTS.md — rules for AI agents building Sleipnir

This repository's platform code is (co-)built by coding agents. Every agent
working in this repo MUST follow this file.

## Read before acting

1. `SPEC.md` — normative platform specification.
2. `adr/` — binding decisions; never violate an Accepted ADR.
3. `DESIGN.md` — normative program layout and design guidelines
   (simplicity, functions, interfaces, tests, naming).
4. `sessions/BACKLOG.md` — current work queue and open questions.
5. `WORKFLOW.md` — PR / no-direct-push rules (mechanically enforced by
   GitHub branch rules since 2026-09-04).

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

## Design guidelines (DESIGN.md) — binding

The program layout and the rules for writing code (simplicity first,
function/interface/naming/test/concurrency/config conventions) are
normative in `DESIGN.md`. Read it before writing any Go; when SPEC.md,
an ADR, and DESIGN.md conflict, ADR > SPEC > DESIGN.

## Error and logging conventions (ADR-0019) — binding

- All errors are created/wrapped through `internal/errs` (captures the
  originating function via `runtime.Caller`). Never use bare `errors.New`/
  `fmt.Errorf` for returned errors.
- Every returned error must read as
  `component.Function: what was attempted: key identifiers: cause` and be
  self-contained for troubleshooting from logs alone.
- Structured logging via stdlib `log/slog` with subsystem loggers and
  correlation attributes. **Spelling is frozen A0-3.6's**, not ADR-0019 §3's
  short forms: `engagement_id`, `run_id`, `job_id`, `node_id`. `node_id` always
  means the **remote agent node** (`slp_node_`, Q9); a graph node is
  `graph_node_id` (`gn_`). The two must never be conflated. (A0-3.6 records that
  ADR-0019 §3's "node" predates ADR-0016's graph vocabulary; the frozen
  contract governs.)
- Log-or-return, never both: an error is logged exactly once, where it is
  handled/decided.
- Secrets (credentials, tokens) never appear in errors or logs; use the
  redaction helpers and write tests proving it.
- **A redaction type must implement `MarshalJSON` and `MarshalText`, never
  `String()` alone.** `slog`'s JSON handler marshals a `KindAny` attribute value
  with `encoding/json`, which **ignores `fmt.Stringer`** — so a String-only
  redaction helper silently leaks every exported field of the value into the log
  record it was supposed to protect. It must also survive *every* `fmt` verb:
  `%T` and `%p` are handled before `fmt.Formatter`, and `fmt`'s bad-verb path
  prints struct fields by reflection with `Stringer` and `Formatter` suppressed,
  so a secret held in a plain `string` field renders verbatim under `%p`. Hold
  it behind a closure (`internal/errs.Secret` is the reference implementation)
  and pin the whole verb set with a test.
- Never drop an audit record and never panic in a request or job path
  (ADR-0019 §6): a logging helper given a nil logger falls back to
  `slog.Default()` rather than panicking or discarding the record.

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
- **Never prove a negative test by removing the bound and running it.**
  Deleting or short-circuiting a guard (depth cap, size limit, timeout) in the
  working tree turns a bounded loop into an unbounded one. On 2026-09-21 an
  agent disabled `internal/errs`' `maxChainDepth` check to show
  `TestErrorChainWalkIsBounded` could fail, then ran it: the cyclic-chain cases
  appended to a `strings.Builder` forever, `errs.test` reached 11.4 GB RSS, the
  kernel OOM-killed it, and the whole WSL VM went down — taking the session and
  every sibling agent with it. The killed agent left the disabled guard in the
  tree and its `/tmp` backup did not survive the reboot. Assert the bound's
  *effect* instead (the truncation notice, the rendered layer count); no
  mutation needed. If a mutation proof is truly unavoidable: mutate a copy
  outside the repo, restore via `defer`/trap that survives a failed run, never
  rely on a `/tmp` backup, and always run capped (`ulimit -v`, `GOMEMLIMIT`,
  short `-timeout`).
- Keep changes reviewably small; one concern per change.
- **Delegation: brief a child for the tools it actually has.** A read-only
  reviewer has no shell and cannot run `gofmt`/`go vet`/`go test` — say so in
  the brief, run the gates yourself, and forbid filesystem-wide searches: on
  2026-09-24 a reviewer spent four minutes in `find /` and was interrupted,
  losing its whole report. A child's completion preview truncates mid-finding,
  so recover the full text from
  `~/.pi/agent/sessions/<parent>/<child-run>/run-0/session.jsonl` before acting
  on it. And treat your own brief as untrusted input: that day three
  implementers found four defects in the principal's briefs and in the frozen
  contract (two became A0 §6 errata 16 and 17) — a child that surfaces a bad
  instruction instead of following it silently is succeeding, not failing.
