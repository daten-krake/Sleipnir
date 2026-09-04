# Development workflow

Effective 2026-09-03 (decision by the product owner):

1. **Everything goes through a pull request.** The `main` branch is
   protected (GitHub branch rules enabled 2026-09-04). No direct pushes —
   code, ADRs, docs, spec changes alike. Agents additionally bind
   themselves to `SPEC.md`, `AGENTS.md`, `DESIGN.md`, and this file; a
   review-then-merge by the product owner is the gate.
2. Branch naming: `<topic>/<short-slug>` (e.g. `layout/internal-errs`).
3. One PR per work package or backlog session; keep them reviewably small.
4. PR requirements:
   - links the backlog item / ADR(s) it implements,
   - `gofmt -l`, `go vet ./...`, `go build ./...`, `go test ./...` green,
   - `go.mod` contains only the ADR-0010 exception list and `vendor/` is
     consistent (`go mod verify`); any diff there requires a new ADR first,
   - tests included; safety/enforcement logic additionally needs negative
     tests (see `AGENTS.md`).
5. Agent-built work: the **principal engineer** aggregates implementer
   output on a branch, runs the gates, and opens the PR; a human reviews
   and merges. Agent-authored PRs must be labeled `agent-built`.
6. ADRs follow the same PR flow: proposed in the PR, accepted by the
   product owner's review comment/merge.

**One-time exception:** the 2026-09-03 architecture kickoff session was
pushed directly to `main` (explicitly allowed by the product owner).
