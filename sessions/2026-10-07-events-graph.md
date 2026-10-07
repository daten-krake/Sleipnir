# Session 2026-10-07 — WP-09 `internal/events` + WP-14 `internal/graph`

- **Date:** 2026-10-07
- **Session id:** ses_eea8e38abffe7NBrfp4hw5mBxx
- **Model:** alibaba-token-plan/qwen3.8-max#xhigh
- **Goal:** Walking-skeleton session 2 (`docs/2026-09-29-walking-skeleton-plan.md`
  §7): deliver WP-09 `internal/events` (envelope, taxonomy, actor, all 42
  payload tables) and WP-14 `internal/graph` (kinds, `Node`/`Edge`/`Provenance`/
  `Attrs`, `contentDoc`) as two independently reviewed lanes against the frozen
  A1/A2 contracts. Housekeeping first: PR for `chore/opencode-config`, stale
  remote branch deleted (done), `gh` auth on the new macOS environment.

## Done this session

- **Housekeeping (partial):** verified all three 2026-09-29 PRs merged (#7/#8/#9),
  local `main` fast-forwarded to `321e319`, CI green on main (3/3 checks);
  deleted stale remote branch `sessions/close-2026-09-29`. Installed `gh` via
  brew (new environment: macOS, was WSL; WORKFLOW §7's Windows-token recipe no
  longer applies — needs amendment or a macOS recipe; `gh auth login` pending,
  then the `chore/opencode-config` PR). Opened working branch
  `foundation/events-graph` from `main`.
- **P-82 is stale-good news:** `next_steps.md` warned `TestContentHashVector`
  needs a new normative vector published in A2 first — A2 §4.2 already carries
  F1/F1-R/F3/S1 (bytes, lens, digests, plus the `4a017e71…` defective-marker)
  and `verify-vectors.py` covers all four (PASS 52 / FAIL 0 re-verified this
  session). No contract edit needed for WP-14.
- **Two implementer lanes launched in parallel** (shared checkout, mechanical
  rules: own directory only, never `./...`): WP-09 `internal/events`
  (envelope, taxonomy, actor, 42 payload structs; seven named test ids; 42-not-39
  reconciliation carried into the brief) and WP-14 `internal/graph` (kinds,
  Node/Edge/Provenance/Attrs, contentDoc; nine named test ids; ADR-0022
  no-confidence rule and the stale `TestNodeDedupIncludesConfidence` name
  carried into the brief). Briefs pin the closed lists against the contract
  text, not the implementation (next_steps §3 definition of done).

## Decisions

- (none yet)

## Open questions carried forward

- (none yet — see `sessions/BACKLOG.md` standing questions and `next_steps.md`)

## Token usage

| total input | uncached input | cache read | cache write | output | reasoning |
|---|---|---|---|---|---|
| 566075 | 53801 | 512274 | 0 | 3484 | 5699 |

_(run `sessions/update-usage.sh sessions/2026-10-07-events-graph.md ses_eea8e38abffe7NBrfp4hw5mBxx` at session end — note: this branch still carries the old `.pi` script; until the `chore/opencode-config` PR merges, use that branch's OpenCode version: `git show chore/opencode-config:sessions/update-usage.sh`)_
