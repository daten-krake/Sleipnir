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
