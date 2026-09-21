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
   **Work packages must be super-minimal** (product owner directive
   2026-09-04): one narrow concern each, the exact contract excerpt to
   implement, files to touch, acceptance tests named explicitly, no
   cross-package assumptions — executable by a junior developer without
   architectural judgment.
6. ADRs follow the same PR flow: proposed in the PR, accepted by the
   product owner's review comment/merge.
7. **PR creation from the agent environment** (verified 2026-09-07): the
   Linux `gh` binary is not authenticated here, and the configured git
   credential helper — the Windows GitHub CLI at
   `/mnt/c/Program Files/GitHub CLI/gh.exe` — is authenticated but cannot
   operate on this WSL checkout (`detected dubious ownership`). Push with
   git, then create/label the PR with the Linux `gh` using the Windows
   token:

   ```sh
   git push -u origin <branch>
   GH_TOKEN=$('/mnt/c/Program Files/GitHub CLI/gh.exe' auth token) \
     gh pr create --repo daten-krake/Sleipnir --base main --head <branch> \
       --title '<title>' --body-file <body.md> --label agent-built
   ```

   The token is read per command and never written to a file, committed, or
   logged (ADR-0019 §5). GitHub does not support push options
   (`-o pull_request.create`); verify the returned PR URL exists before
   closing the session.

**One-time exception:** the 2026-09-03 architecture kickoff session was
pushed directly to `main` (explicitly allowed by the product owner).
