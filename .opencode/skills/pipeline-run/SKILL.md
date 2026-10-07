---
description: Start, stop, or check the Sleipnir continuous build pipeline (24/7 small-LLM lane runs). Use when the user asks to run the pipeline, start an overnight/weekend run, "let it build", work the WP queue unattended, check on a running loop, or stop one. Always asks the user how long the run should last before starting anything.
---

# Skill: pipeline-run

Runs `pipeline/driver.sh` — the supervised continuous-build loop from
`docs/2026-10-07-continuous-build-pipeline.md`. The loop picks queue-ready
work packages, runs implementer/reviewer/fixer children in isolated
worktrees under watchdog caps, commits each stage (small + often), and parks
lanes at `principal_validation`. **It never merges, never opens PRs, never
edits frozen contracts** — those stay human gates (plan §4 H1/H2).

## Context hygiene (owner directive 2026-10-07 — binding)

**After each task, the context is cleared.** Three rules follow, and they
override any "efficiency" idea to the contrary:

1. **Fresh child context per stage.** The driver spawns one new
   `opencode run` per stage (implementer → reviewer → each fix round) and
   never passes `--session`/`--continue`. No child is ever resumed; nothing
   carries over between stages or tasks.
2. **Durable memory is files, not conversation.** `queue.json` (status),
   `pipeline/runs/wp-<id>/` (briefs, logs, report.md, review.md) and the lane
   branch's commits are the only handoff. The next task's brief is rendered
   from the queue entry and report paths — never from a conversation
   summary. If a fact must survive a task, it goes into the queue entry's
   `notes` or a report file before the task ends.
3. **No babysitting session.** Runs launch **detached** (`nohup`, §3): no
   LLM context exists at all while the loop works. Checking on a run is a
   file read — from any session, including a brand-new one. Never keep a
   session alive just to watch a loop, and never hold task details in your
   own context between checks: re-read the files each time.

## Start a run

### 1. Ask the duration FIRST (mandatory — do not skip)

Use the question tool before touching anything else:

> "How long should the pipeline run?"

Options to offer (recommend based on time of day and queue depth):

- **2 hours** — a supervised trial (first-ever run, or after a retier)
- **4 hours** — one lane stage, watched
- **8 hours** — overnight (the standard unattended run)
- **12 hours** — long day run
- custom — the user types their own (`--for` accepts `90m`, `8h`, `3600s`)

Also ask (one question, only if the previous run ended in violations or
timeouts): whether to retry parked lanes or skip them.

### 2. Preflight (all must pass; fix or report before starting)

```sh
cd /Users/uwe/Sleipnir                          # or the current checkout
pgrep -f "driver.sh loop" && echo "LOOP ALREADY ACTIVE — do not start a second one; report progress (§4)"
git status --short                              # clean tree required
git fetch origin && git log --oneline -1 main origin/main   # main current?
gh auth status                                  # needed for the principal later, not the loop
curl -s -m 5 "$(sed -n 's/.*baseURL": "\(http[^"]*\)".*/\1/p' ~/.config/opencode/opencode.jsonc | head -1)/models" >/dev/null && echo "LM Studio reachable"
pipeline/driver.sh status                       # queue state
pipeline/driver.sh next                         # something schedulable?
```

Checks that matter:

- **Local main must contain `pipeline/`** (the bootstrap PR merged) — lane
  worktrees are created from main and need the gitignored `pipeline/runs/`
  convention; on a pre-bootstrap main the loop's report files would pollute
  lane diffs.
- **The implementer model must be the pinned id** `lmstudio/qwen/qwen3.8-27b`
  (models.env). Its global config entry carries `limit.context: 64000`;
  the auto-discovered sibling ids (`lmstudio-community/…`, `unsloth/…`)
  report the 262144 architecture max instead — never point a tier at those.
- **Awaiting-human capacity:** if `driver.sh status` shows ≥ AWAITING_HUMAN_MAX
  lanes at `principal_validation`/`pr_open`, the loop will finish almost
  immediately (plan §7: never outrun H1). Tell the user to clear the review
  batch first, or run anyway if they insist.
- Nothing schedulable (`next` empty) → do not start; report what is blocking
  (unmerged deps or all lanes blocked-h3).

### 3. Launch (detached, timed — context hygiene rule 3)

Default for every run, so no session context is held while the loop works:

```sh
nohup pipeline/driver.sh loop --for <duration> \
  > pipeline/runs/loop-$(date +%Y%m%d-%H%M%S).log 2>&1 &
echo $! > pipeline/runs/loop.pid
```

Then tell the user: what was started, the deadline (local time), the loop
log path, the lane it will pick up first (`driver.sh next | head -1`), and
how to check on it — in ANY later session, even a fresh one: "check the
pipeline" (this skill, §4). End your turn; there is nothing to watch.

Only for a **supervised 2h trial** (first-ever run, or right after a
retier), launch attached instead, with the shell tool in background mode —
the harness then notifies on completion:

```sh
pipeline/driver.sh loop --for 2h
```

### 4. While it runs / checking on a run

All checks are file reads — never keep them in context between checks
(hygiene rule 3):

- Running? `pgrep -f "driver.sh loop"` (or `cat pipeline/runs/loop.pid`).
  **If a loop is already active, do NOT start a second one** — report
  progress instead. Concurrency beyond one loop is a P6 decision, not an
  improvisation.
- Progress: `pipeline/driver.sh status`; loop log tail:
  `tail -20 pipeline/runs/loop-*.log`.
- Per-lane artifacts: `git -C ~/Sleipnir-lanes/wp-<id> log --oneline` and
  `pipeline/runs/wp-<id>/`.

### 5. Early stop (only on user request)

```sh
pkill -f "driver.sh loop"                       # stops the supervisor
pkill -f "opencode run --agent sleipnir-"       # stops running children
```

Then commit any uncommitted lane work so nothing is lost:

```sh
pipeline/driver.sh commit <wp> "WP-<wp>: WIP at operator stop (pipeline)"
```

A stopped lane stays resumable: its worktree and branch survive; re-running
the loop continues with the next schedulable WP (a `timed_out`/parked lane
needs `cleanup` + `start` for a fresh attempt, or a manual `fix` round).

### 6. On completion — report to the user

A detached run sends no notification: completion means `pgrep -f "driver.sh
loop"` is empty and the loop log ends with `loop: run summary` + the status
table (`tail -40 pipeline/runs/loop-*.log`). The user will typically ask in a
**fresh session** — that is the point of the hygiene rules: reconstruct the
summary from files only, never from remembered context.

Summarize from the log's final status table plus:

- lanes delivered (`principal_validation`) and where their reports are
  (`pipeline/runs/wp-<id>/report.md` + `review.md`);
- lanes that stopped early (`gates_failed`, `timed_out`,
  `ownership_violation`, `reviewer_violation`) and the one-line cause;
- commits per lane branch (`git -C ~/Sleipnir-lanes/wp-<id> log --oneline main..`);
- **errata candidates** found (grep the review reports for
  `## ERRATUM CANDIDATE`) — these need product-owner rulings (H2);
- the suggested next step: a principal pass to validate the parked lanes and
  open PRs (H1 batch).

## Rules this skill never overrides

- **Context hygiene (above): fresh child per stage, files as the only
  memory, detached runs, no babysitting session.**
- WORKFLOW: merges are the product owner's; PRs get the `agent-built` label.
- Frozen contracts/ADRs change only via product-owner rulings (H2), applied
  by the principal — never by a lane child.
- Caps are not negotiable (AGENTS.md 2026-09-21): the driver's watchdogs
  stay on; a retier or a re-cut WP is the answer to timeouts, not a bigger
  cap.
