---
name: start-session
description: Start a new Sleipnir working session. Use when the user wants to start the next session, says "start session", or asks to pick up work in this repository. Triggers the session ritual: check housekeeping, read the normative context in the fixed order, collect all open tasks and questions, report them, then open the session tracker. Also defines how to close a session.
---

# Sleipnir session lifecycle

Starting a session is a ritual: load context **in the fixed order**, collect
**every open task**, report it, then begin. Do not start any work before
steps 1–4 are done and the goal is confirmed.

## Start

### 1. Housekeeping check

- `git status` + `git fetch`: clean tree? unpushed commits or branches?
  Pending PRs? Deal with leftovers before new work.
- Read `next_steps.md` if present — it is the handover from the last
  session (its housekeeping items come first).

### 2. Load normative context — fixed order, skip nothing

1. `SPEC.md` — the platform specification.
2. `adr/` — at minimum `README.md` plus every ADR written or changed
   since the last session; skim statuses of the rest.
3. `DESIGN.md` — program layout + design guidelines.
4. `sessions/BACKLOG.md` — the work queue.
5. `AGENTS.md`, `WORKFLOW.md` — build rules and PR flow.

Authority when they disagree: **ADR > SPEC > DESIGN**.

### 3. Collect open tasks

Gather into one list, each item with its source:

- Session queue items in `sessions/BACKLOG.md` (numbered, in order).
- "Standing open questions" from the backlog.
- Housekeeping items from `next_steps.md` (pushes, credentials, checks).
- Unfinished items in the latest session tracker file under `sessions/`
  ("Open questions carried forward").
- ADRs still `Proposed` (they need product owner decisions).

### 4. Report and confirm the goal

Present the open-task list compactly to the product owner, recommend the
next item (top of the backlog queue unless `next_steps.md` says
otherwise), and get the goal confirmed **before** creating files or
branches.

### 5. Open the session tracker

Create `sessions/YYYY-MM-DD-<topic>.md` from the template below:

- **Session id:** `$PI_SESSION_ID` · **Model:** `$PI_PROVIDER/$PI_MODEL`
- **Goal:** the confirmed goal in one or two sentences.
- Snapshot initial usage: `./sessions/update-usage.sh sessions/<file>.md`
- Then set up the working branch per `WORKFLOW.md`
  (`<topic>/<short-slug>`); new ADRs start as **Proposed**, numbered after
  the highest existing number.

## Template

```markdown
# Session YYYY-MM-DD — <topic>

- **Date:** YYYY-MM-DD
- **Session id:** <PI_SESSION_ID>
- **Model:** <provider/model>
- **Goal:** <one or two sentences>

## Done this session

- (append as work happens)

## Decisions

- <decision> → ADR-NNNN (Proposed/Accepted) or DESIGN.md/SPEC.md change

## Open questions carried forward

- <question>

## Token usage

| total input | uncached input | cache read | cache write | output | reasoning |
|---|---|---|---|---|---|
| 0 | 0 | 0 | 0 | 0 | 0 |

_(run `sessions/update-usage.sh sessions/<file>.md` at session end)_
```

## Close (before ending a session)

1. Update the tracker: Done / Decisions / carried-forward questions.
2. Refresh `sessions/BACKLOG.md` (progress, new queue items, archive
   resolved questions).
3. Write `next_steps.md` for the next session: housekeeping first, then
   the plan with a definition of done.
4. Refresh usage: `./sessions/update-usage.sh sessions/<file>.md`
5. Commit, push the branch, ensure the PR exists
   (`git push -o pull_request.create -o pull_request.title="..."` when no
   `gh` CLI is available); agent-authored PRs get the `agent-built` label.
