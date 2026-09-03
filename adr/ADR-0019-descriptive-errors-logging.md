# ADR-0019: Descriptive errors and granular logging (agent-debuggable by design)

- **Status:** Accepted
- **Date:** 2026-09-03
- **Deciders:** Product owner, architect

## Context

Sleipnir is built and maintained by coding agents. Troubleshooting quality
stands or falls with how errors read: an agent (or human) given a log line
must immediately know **what is wrong and where it happened**. Go's bare
`errors.New` strings and silent wrapping do not provide that. Requirement
from the product owner: errors and logs must be detailed and granular;
every error carries the function it originated in.

## Decision

1. **All errors flow through an internal `errs` package** (stdlib-only)
   that captures origin automatically via `runtime.Caller`:
   - `errs.New(msg)`, `errs.Newf(format, ...)` — create an error tagged
     with the fully-qualified originating function.
   - `errs.Wrap(err, msg)`, `errs.Wrapf(...)` — add the current function +
     what it was doing, preserving the cause chain (`errors.Is/As` intact).
   - Error kinds (`validation`, `auth`, `notfound`, `conflict`, `timeout`,
     `upstream`, `internal`, ...) attached at creation so handling is
     programmatic while messages stay descriptive.
2. **Every returned error is self-contained and greppable:**
   `component.Function: what was attempted: key identifiers: cause`.
   Example:
   `store.InsertEvent: insert event for engagement=e12 run=r3: pq: duplicate key value violates unique constraint`.
   Public functions wrap lower-level errors with what they were doing and
   the relevant identifiers (engagement/run/job ids, target, tool).
3. **Structured logging with stdlib `log/slog`** (Go ≥1.21):
   - Subsystem loggers (`platform.policy`, `broker`, `api`, `notify`, ...).
   - Every error-level record includes: function/op, error chain, and
     correlation attributes (`engagement`, `run`, `job`, `node`).
   - **Log-or-return, never both:** an error is logged once, at the point
     where it is handled/decided; intermediate layers only wrap and return.
4. **Granularity:** failures of external calls (LLM endpoint, Docker
   runtime, webhook delivery, DB) log request context (endpoint, status,
   duration, retry count) without request bodies that may contain secrets.
5. **Redaction is mandatory:** captured credentials, tokens, and secret
   material never appear in error strings or logs; helpers exist to
   redact, and tests assert their behavior.
6. **Panic policy:** panics only for programmer-invariant violations at
   startup; never as error transport in request/job paths.
7. Convention enforcement: `AGENTS.md` makes this binding for builder
   agents; negative tests required (e.g. "error contains origin function",
   "secret not present in error string").

## Consequences

- **Positive:** agent (and human) troubleshooting from logs alone becomes
  viable; error chains double as execution traces; consistent machine-
  parseable structure (slog JSON) feeds monitoring and audit export later.
- **Negative:** small runtime cost of `runtime.Caller` (negligible at our
  error rates); discipline required — enforced via AGENTS.md + review.
- **Follow-ups:**
  - `internal/errs` + logging setup are part of the program-layout work
    package (backlog session 2).
  - Log export surface (SIEM webhook / audit) inherits this structure
    (enterprise gap list).
