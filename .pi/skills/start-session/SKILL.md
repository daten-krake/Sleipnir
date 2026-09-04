---
name: start-session
description: Start (and close) a Sleipnir working session. Use when the user says they want to start the next session, pick up the backlog, or open a new working session in this repository. Loads the normative context, creates the session tracker record, and ends sessions with usage snapshot and next_steps handover.
---

# Sleipnir session lifecycle

Every working session in this repo is tracked and leaves the repo ready for
the next one. Follow the steps below in order.

## Start

1. **Housekeeping.**
   - `git status` / `git fetch` — confirm a clean tree; note unpushed
     commits and deal with them first.
   - Read `next_steps.md` if it exists: it is the handover from the last
     session (housekeeping items first, then the plan).

2. **Load normative context, in this order** (skip nothing):
   1. `SPEC.md`
   2. `adr/` — at minimum README + every ADR newer than the last session
   3. `sessions/BACKLOG.md`
   4. `AGENTS.md`, `WORKFLOW.md`

3. **Create the session tracker file** `sessions/YYYY-MM-DD-<topic>.md`
   (see template below). Fill:
   - **Session id:** `$PI_SESSION_ID`
   - **Model:** `$PI_PROVIDER/$PI_MODEL`
   - **Goal:** one or two sentences, taken from the top of the
     `sessions/BACKLOG.md` queue or `next_steps.md`; confirm the goal with
     the product owner before starting work.
   - Snapshot initial token usage:
     `./sessions/update-usage.sh sessions/<file>.md`

4. **Set up the working branch** per `WORKFLOW.md`: everything (code,
   ADRs, docs, spec) goes through a PR; branch name `<topic>/<short-slug>`.
   New ADRs are numbered sequentially after the highest existing number and
   start as **Proposed**.

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

- <decision> → ADR-NNNN (Proposed/Accepted)

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
2. Refresh `sessions/BACKLOG.md` (mark progress, add new queue items,
   archive resolved questions).
3. Write `next_steps.md` for the next session: housekeeping first
   (pending pushes, credentials, model checks), then the plan with a
   definition of done.
4. Refresh usage: `./sessions/update-usage.sh sessions/<file>.md`
5. Commit, push the branch, and make sure the PR exists
   (`git push -o pull_request.create -o pull_request.title="..."` when no
   `gh` CLI is available); agent-authored PRs get the `agent-built` label.
