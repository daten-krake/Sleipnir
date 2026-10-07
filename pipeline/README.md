# pipeline/ — the continuous build pipeline

Supervised 24/7 execution of the walking-skeleton work packages by tiered
models, per `docs/2026-10-07-continuous-build-pipeline.md` (the plan; its §9
P-decisions govern tuning). **The driver is a scheduler, not an authority:**
it never merges, never opens PRs, never edits frozen contracts. Human gates
(plan §4): H1 merges, H2 errata rulings, H3 the D10–D18/egress decision
evening, H4 safety-critical deep review.

Start/stop runs through the **`pipeline-run` skill** (it asks the operator
how long to run). Manual operation below.

## Layout

| Path | What |
|---|---|
| `driver.sh` | the supervisor: worktrees, capped child runs, gates, ownership diff, lane commits, queue transitions |
| `gates.sh` | the WORKFLOW §4 gates for one worktree (five gates + `-race` + verify-vectors) |
| `queue.json` | the work queue (seeded from the principal review §4 rows + walking-skeleton plan §4/§7 + the 2026-10-07 handoff notes) |
| `models.env` | model tiers, watchdog caps, throughput caps, diff-size warning |
| `templates/` | brief templates: implementer, reviewer, fixer (placeholders rendered from queue.json) |
| `errata-queue.md` | contract/ADR defects surfaced by lanes, awaiting product-owner rulings (H2) |
| `runs/` | gitignored run artifacts: briefs, child logs, reports, gate logs |

Worktrees live **outside** the repo: `~/Sleipnir-lanes/wp-<id>` on branch
`wp/<id>-<slug>`, created from local `main`.

## State machine (queue.json `status`)

```
ready → implementing → gates_passed → reviewed → principal_validation
              │                              ↘ (MUST FIX) fix round → …
              ├→ gates_failed / ownership_violation / timed_out / reviewer_violation
              └ (principal) → pr_open → merged     [merged unlocks dependents]
blocked-h3 → (H3 rulings) → ready
```

`ready` + all deps `merged` = schedulable (`driver.sh next`).

## Commands

```sh
pipeline/driver.sh status                 # queue table
pipeline/driver.sh next                   # what would run now
pipeline/driver.sh start  <wp>            # worktree + implementer + gates + commit
pipeline/driver.sh review <wp>            # read-only reviewer (worktree must be clean after)
pipeline/driver.sh fix    <wp> <findings> # fix round + gates + commit (≤ MAX_FIX_ROUNDS)
pipeline/driver.sh commit <wp> [msg]      # manual checkpoint commit (ownership-checked)
pipeline/driver.sh gates  <wp>            # re-run gates only
pipeline/driver.sh mark   <wp> <status> [pr]
pipeline/driver.sh collect <wp>           # artifacts + review report
pipeline/driver.sh cleanup <wp>           # remove worktree (branch kept)
pipeline/driver.sh loop --for 8h          # the 24/7 mode (via the pipeline-run skill)
```

Small-change + commit-often discipline (owner directive 2026-10-07): the
driver commits after every stage that produced changes (delivery, each fix
round, WIP-on-timeout), warns when a lane diff exceeds `DIFF_WARN_LINES`,
and the briefs forbid drive-by changes — a lane that cannot stay small
should be re-cut in queue.json, not pushed through.

## Machine prerequisites

- `opencode` on PATH; **global config defines the `lmstudio` provider**
  (`settings.baseURL` → the LM Studio host) with the model
  `qwen/qwen3.8-27b` carrying `limit.context: 64000`. ⚠️ Tier configs must
  pin that **exact id** (`lmstudio/qwen/qwen3.8-27b`): the auto-discovered
  sibling ids report the 262144 architecture max, which is not the loaded
  runtime window.
- Go toolchain matching `go.mod`, python3, perl (process-group watchdog),
  git, `gh` (principal steps only).
- The bootstrap PR merged into `main` **before the first dry run** — lane
  worktrees branch from main and rely on the gitignored `pipeline/runs/`.
- Wave B additionally needs the P3 Linux host (PostgreSQL, host nftables,
  Docker, the dry-run target).

## Dry run (plan §6.4 — mandatory before parallel lanes)

1. `pipeline/driver.sh start 10` (WP-10 chain primitives, small tier).
2. `pipeline/driver.sh review 10`.
3. Principal reads `pipeline/runs/wp-10/{report.md,review.md}` + gate logs,
   spot-checks evidence, opens the PR, `mark 10 pr_open <n>`.
4. Measure: wall time, findings by class, how much the principal had to
   correct. If the implementer output needed rewriting → retier
   (`models.env`: flash → 27b → strong) or re-cut the WP, then repeat.

## Recovery

- `timed_out` / stopped lane: worktree + branch survive; WIP is committed.
  Fresh attempt: `cleanup <wp>` + `start <wp>`. Continue: manual
  `opencode run` in the worktree, then `gates`/`commit`, or a `fix` round.
- `ownership_violation`: the lane touched files outside its queue entry —
  inspect `runs/wp-<id>/ownership.log`; either the entry's file list is too
  narrow (fix queue.json, principal's call) or the child strayed (fresh
  attempt with the violation quoted into the brief).
- `reviewer_violation`: the worktree changed during review — the review is
  discarded; re-run `review`.
- Loop died mid-stage: re-run `loop`; `implementing` entries with an
  existing worktree need `cleanup` + `start` (the driver refuses to reuse a
  dirty state it did not create).

## What never changes (plan §7/§10)

Merges are the product owner's (WORKFLOW §1). Contracts/ADRs change only
through H2 rulings recorded as §6 items / erratum notes. Every delivery
keeps its independent review. Caps stay on (AGENTS.md). The pipeline is a
scheduler; SPEC/ADR/DESIGN/AGENTS/WORKFLOW remain the authority for
everything it schedules.
