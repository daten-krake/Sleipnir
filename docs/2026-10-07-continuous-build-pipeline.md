# Continuous build pipeline — small-LLM 24/7 execution plan

- **Date:** 2026-10-07
- **Status:** PROPOSAL — nothing in it is decided until §9's P1–P6 are
  answered (same shape as the walking-skeleton plan's D10–D18).
- **Input:** `docs/2026-09-29-walking-skeleton-plan.md` (the schedule this
  accelerates), `sessions/2026-10-07-events-graph.md` (the calibration data),
  AGENTS.md / WORKFLOW.md (governance this plan must not weaken).
- **Target:** first HackTheBox run attempt **2026-11-17**, band
  **2026-11-10 → 2026-11-28** — versus the walking-skeleton plan's
  2026-12-08 milestone and 2026-11-24 → 2027-01-15 band.

## 1. The honest bottleneck

The walking-skeleton schedule spends 20 sessions at two per week. Each
session is a human-supervised unit: brief lanes, watch them, review, rule,
merge. Between sessions **nothing moves**, and the local models that could
do the transcription work sit idle ~95 % of the time. The plan's own words:
"the scarcest resource in the schedule is your attention, not compute."

This proposal therefore does two things and only two:

1. **Decouple execution from sessions.** A supervised pipeline (driver
   script + tiered agents) runs work packages continuously, 24/7, producing
   review-ready PRs.
2. **Convert owner attention from sessions to gates.** Instead of two
   3-hour sessions a week: one daily batch (~30–60 min) clearing a PR queue
   and an errata queue, plus two decision evenings.

What it does **not** do: let agents merge, let agents rule on frozen
contracts, or let small models solo on safety-critical code. Those gates are
the platform's product — a pentest tool whose approval path was built and
reviewed carelessly is worse than no tool.

## 2. Calibration: why small models can carry Wave A

The last three deliveries are the evidence base:

| Session | Package class | Review outcome |
|---|---|---|
| 2026-09-24 | foundation (ids/timex/caps, cjson) | 18 findings, **zero** code defects in shipped logic |
| 2026-09-29 | paging + secretscan | 22 findings, **zero** shipped-logic defects |
| 2026-10-07 | events + graph | ACCEPT + APPROVE, **zero** MUST FIX, zero shipped-logic defects |

The work that keeps arriving is *transcription under mechanical oracles*:
frozen contract text → Go structs → tests that parse the contract markdown
and mutation-prove their own pins. Every defect three sessions of reviewers
found was a citation, a doc claim, or an unpinned promise — the class a
second independent reader catches, not the class that needs a big model.
The mechanical backstops (contract-parsing tests, `verify-vectors.py`, CI's
five gates, `go list -deps` import proofs) are exactly what makes small-model
output *checkable* instead of *trustable*. This plan leans on that.

Where small models are **not** sufficient: anything on AGENTS.md's
high-review list (auth, tokens, enforcement, streaming parsers, broker,
container lifecycle). Those WPs get the strongest available model plus a
human who reads the diff, not just the reports. §5 assigns every WP a tier.

## 3. Pipeline architecture

### 3.1 Roles

| Role | Implementation | Writes to repo? | Model tier |
|---|---|---|---|
| **Driver** | `pipeline/driver.sh` + launchd/cron. No LLM. Owns the worktree lifecycle, the caps, the watchdog, artifact collection | never | — |
| **Principal** | one long-lived OpenCode session per batch (or `opencode run`), the only agent that touches more than one lane | briefs, validation notes, errata *applications after human ruling*, PR creation, queue state | strong (e.g. qwen3.8-max) |
| **Implementer** | one child per WP, per §3.2 isolation | its own directory in its own worktree only | small (e.g. local qwen3.8-27b / flash) |
| **Reviewer** | one child per delivered WP, read-only | nothing — findings to a report file | small, *different provider/model than the implementer where possible* |
| **Fixer** | one child per accepted finding batch | the reviewed worktree only | small |

The principal is the same role this session played — the difference is that
its inputs (reports) and outputs (briefs, PRs, queue updates) are files in
`pipeline/`, so the driver can run it unattended and the owner can audit
every step.

### 3.2 Isolation: worktrees, not shared checkouts

The shared-checkout rule ("own directory only, never `./...`") held across
eight supervised children in two sessions. **Unattended, it is the wrong
tool**: one runaway child can poison a sibling's build. Instead:

- one `git worktree` per active lane under `~/Sleipnir-lanes/<wp-id>`,
  branch `wp/<nn>-<slug>` off current `main`;
- the lane's brief forbids touching anything outside its file list (same
  rule, now mechanically enforceable — the driver diffs the worktree against
  the file list before the principal aggregates);
- the principal merges lane branches into one integration branch per PR
  family (precedent: `foundation/events-graph` = two lanes, one PR);
- contracts/ and adr/ are **read-only to every child**. Only the principal
  edits them, only to apply human-ruled errata, one writer at a time.

### 3.3 The queue

`pipeline/queue.json`, seeded from the walking-skeleton plan §7's rows:

```json
{ "wp": 10, "family": "events",  "deps": [9],
  "tier": "small", "review": "one",
  "clauses": "A1-5.1..5.10 + §4.3 vector",
  "files": ["internal/events/chain.go", "internal/events/served.go"],
  "gate": "TestChainVectorDigests byte-exact",
  "status": "ready" }
```

Status machine per WP:
`ready → briefed → implementing → gates → reviewing → fixing → gates →
principal-validated → PR_OPEN → MERGED (human) → (unlocks deps)`

A WP whose review produces MUST FIX goes back to `implementing` with the
findings file appended to its brief. A WP that surfaces a contract defect
gets `errata-blocked` and moves to §4 H2 — **the lane stops at the defect,
it never rules**. (Both 2026-10-07 lanes did exactly this and were right.)

### 3.4 Driver guardrails (mechanical, from repo incidents)

The driver enforces, per lane, without any LLM in the loop:

- `GOMEMLIMIT=512MiB` on every go command; `go test -timeout 300s`;
- wall-clock kill: 4 h per implementer lane, 2 h per reviewer, then SIGKILL
  and status `timed-out` (AGENTS.md 2026-09-21: an unbounded loop OOM-killed
  a whole VM — unattended operation makes the watchdog non-optional);
- `ulimit -v` on mutation-proof subprocesses, which the brief confines to
  copies **outside** the worktree;
- no network egress except the model endpoints and `gh` (macOS seatbelt or
  the driver's environment; WP-30's probe work excepted, explicitly);
- artifact discipline: every child's report is a file it appends to
  (`pipeline/reports/<wp>-<role>.md`), named in its brief — completion
  previews truncate and final messages get lost (AGENTS.md delegation
  bullet);
- secrets: `gh` via keychain only; the driver never writes a token to a
  file, env dump or log (ADR-0019 §5);
- concurrency cap (P6): 2–3 lanes — RAM, disk and the merge-gate cadence all
  bound it, not ambition.

### 3.5 Why the tests keep small-model drift visible

Wave A packages inherit the oracle discipline WP-09/WP-14 established:
closed lists, key sets and vectors are **parsed from the contract markdown**,
mutation-proven, and CI recomputes every published vector. A small model
that hallucinates a field fails its own package's tests; a small model that
weakens a test fails the mutation proofs or the reviewer's re-run. That is
the property that makes tiering safe — it must be a standing requirement in
every brief, not a lucky accident.

## 4. Human gates (irreducible)

| # | Gate | Cadence | What it costs |
|---|---|---|---|
| **H1** | **PR merges.** CI green + principal validation note + review reports linked in the body; owner reads notes, merges or bounces | daily batch, ~30–60 min | the pipeline's speed limit. PRs opened faster than H1 clears them just queue up — the driver is capped to H1's measured rate (§8) |
| **H2** | **Errata rulings.** Frozen-contract and ADR changes queue in `pipeline/errata-queue.md` with evidence + recommendation; owner rules in batch; principal applies | daily batch (same sitting as H1) | 10 errata were ruled in one sitting on 2026-10-07 — the batch works |
| **H3** | **Decision evening: D10–D18 + the worker-egress ADR** — pulled *forward* from plan-S8 (2026-10-23) to **week of 2026-10-06** | one evening, once | everything from WP-23 on is gated on it; also decides P3's host question (egress option E1 assumes host nftables — a Linux host, not Docker Desktop on macOS) |
| **H4** | **Safety-critical deep review.** For every `tier: strong+human` WP (§5): the owner reads the diff and the negative tests, not just the reports | per WP, ~1 h each, ~10 WPs | cannot be delegated; this is the AGENTS.md high-review bar and C5/C6 enforcement |

H2/H4 rulings stay recorded where they belong: contract §6 items, ADR notes,
session trackers. The pipeline changes *when* decisions happen, never *how
they are recorded*.

## 5. Waves and tiering

### Wave A — contract packages (start immediately after PR #11 merges)

All frozen-contract transcription; small-model suitable; the §7 schedule's
S3–S7 compressed into one continuous flow:

| Stage | Lanes (∥) | Tier | Gate that ends the stage |
|---|---|---|---|
| A1 | WP-10 chain primitives ∥ WP-15 graph validation | small + one review each | `TestChainVectorDigests` byte-exact; A2-9.4 caller-side exemption tested (positive + negative) |
| A2 | WP-11 events validation ∥ WP-16 quarantine | small; WP-11's secret-scan wiring gets a second review (enforcement path) | `TestQuarantineIsNeverCallerSettable` |
| A3 | WP-12 verification walk ∥ WP-17 store seams | small; WP-12 tamper matrix is safety-adjacent → strong-model reviewer | `TestTamperMatrixT1ToT17` all rows detect **and** name `break_kind` |
| A4 | WP-18 ingest ∥ WP-19/20/21 contract suites (3 lanes) | small | `contracts/README.md` merge gate passes for A0, A1, A2 |
| A5 | WP-22 store/postgres ∥ WP-31 tools registry | **WP-22: strong+human** (pgx vendoring, `REVOKE UPDATE, DELETE`, append-only at DB level); WP-31 small | `TestDBLevelAppendOnly`, `TestIDOrderingMatchesByteOrderCollateC`; `go.mod` = ADR-0010 list only |

WP-22 is the first WP needing a real PostgreSQL (integration tests are
opt-in per DESIGN §8) — P3's host question again.

### Wave B — the skeleton (starts as H3 rulings land; overlap with Wave A tail)

**Small-model suitable** (still one independent review + H1):
IR-4 DDL · WP-27 A7-min · WP-28 A4-min/A5-min (api surface is large but
contracted; the authorization negatives are the review focus) · WP-36
evidence · WP-37 notify (the D5 head-anchor webhook gets a strong-model
reviewer — it is an integrity control) · WP-38 UI (HTMX; approval-queue PR
gets H4-adjacent attention: fingerprint hash + CSRF on the button) · WP-39
worker · WP-41 compose+runbook · WP-33 apiclient (SSE parser → strong-model
reviewer; streaming parsers are on the high-review list).

**Strong model + H4 human deep review — never small-model solo:**
WP-23 store schema+impls (audit spine) · WP-24 auth (PBKDF2 RFC 6070
vectors, sessions, CSRF) · WP-25 machine tokens · WP-26 policy
(blacklist-beats-everything — the platform's reason to exist) · WP-29 spawn
broker (container lifecycle, no docker socket to orchestrator) · WP-30
netpolicy + probe (host nftables; needs P3's Linux host and a real worker
container) · WP-32 llm provider+gateway (secret exclusion under every
policy, SSE framing) · WP-34 agentloop (the first unprecedented unknown;
`TestLoopStopsAtMaxIterations` is a bound — watchdog discipline applies
doubly) · WP-40 images (digest pinning = supply chain).

**Move R1 early:** the 20-call model tool-calling smoke test (plan-S8) runs
at H3, not later — if the chosen orchestrator model cannot hold ≥90 %
well-formed tool calls, WP-34's design changes and that must be known before
Wave B's second half, not after.

### Wave C — first contact (human in the loop, by design)

WP-38 completion → S21-equivalent: dry run against a throwaway container on
the Linux host, then HTB attempt 1. No automation value here; this is the
point of the whole build. Buffer: one week (band, §9).

## 6. Bootstrap — building the pipeline itself (2 days, before Wave A)

1. `pipeline/` skeleton: `queue.json` (seeded from plan §7 rows S3–S22 with
   §5's tiers), `driver.sh`, `reports/`, `errata-queue.md`, `gates.sh`
   (scoped per-worktree five gates + deps check).
2. OpenCode agent definitions: extend `.opencode/agents/sleipnir-implementer`
   and `-principal` with pipeline-mode sections (report-file discipline,
   worktree rules); add `sleipnir-reviewer` (read-only permissions).
3. Driver: worktree create/destroy, `opencode run --model <tier>` per lane,
   watchdog + caps (§3.4), artifact collection, `queue.json` status
   transitions, `gh pr create` through the principal only.
4. **Dry run = WP-10**, the next WP anyway: one small-model implementer, one
   small-model reviewer, principal validation, human merge. Measure: wall
   time, findings by class, gate failures, how much the principal had to
   correct. If the dry run needs the principal to rewrite the lane's output,
   retier (flash → 27b → strong) before opening parallel lanes.
5. Only then: parallelize stages A1's two lanes, and keep concurrency at
   what H1's measured daily merge rate supports (§8).

Bootstrap lives on its own branch/PR (`pipeline/…` is tooling, not platform
code — DESIGN §1's layout is untouched; the driver is never imported by Go
code).

## 7. What the pipeline must never do

- **Merge.** `main` stays protected; H1 is a human (WORKFLOW §1).
- **Edit frozen text.** Contracts and ADRs change only via H2 rulings,
  applied by the principal, recorded as §6 items / erratum notes with the
  ruling date and path (the 2026-10-07 set is the format).
- **Skip the independent review.** One reviewer minimum, a *different* model
  than the implementer where possible, read-only, own extractors, mutation
  proofs out-of-repo and capped. Two reviewers for enforcement-path WPs.
- **Trust a final message.** Reports are files; the principal spot-checks
  reviewer evidence like child code ("a symbol table is not a runtime
  value").
- **Run uncapped.** Every subprocess has memory, timeout and wall-clock
  bounds; mutation proofs never disable a bound in any tree (AGENTS.md).
- **Outrun H1.** The driver opens at most N PRs ahead of the merge queue
  (N=2 recommended); a full queue pauses lane starts. Throughput beyond the
  owner's review rate is inventory, not progress.

## 8. Timeline arithmetic

Assumptions: PRs #10/#11 merged 2026-10-08; bootstrap 10-08→10-09; H3
evening ≤ 2026-10-10; H1 daily from 10-09; P3's Linux host exists by
Wave B start; concurrency 2–3 lanes.

| Phase | Calendar | Content |
|---|---|---|
| Bootstrap + dry run | 10-08 → 10-10 | driver, queue, WP-10 through the pipeline |
| Wave A | 10-10 → **10-21** | stages A1–A5; ~1.5–2.5 days per stage (24/7 vs two-per-week: S3–S7 took 5 calendar weeks on the old plan) |
| Wave B first half | 10-21 → **11-03** | WP-27/28, 23∥24, 26∥25 — H4 reviews are the pacing item (~1/day sustainable) |
| Wave B second half | 11-03 → **11-13** | 29∥36, 30∥37, 32, 33∥35, 34∥38, 39, 40∥41 |
| Wave C | 11-13 → **11-17** | dry run + HTB attempt 1; band to 11-28 with buffer |

**Compression achieved:** 20 sessions / 10 weeks → ~6 weeks, without
touching any gate. The three unprecedented unknowns (agentloop, broker +
netpolicy, LLM tool-calling reliability) sit in Wave B and keep their risk —
they are why the band exists; 24/7 does not make unknowns smaller, it makes
the *queue* shorter.

**What breaks the date, in order of likelihood:**
1. H1/H4 cadence slips (one skipped day ≈ one day on the end date; a
   skipped week ≈ a skipped week).
2. H3 slips past 10-13 (Wave B cannot start; Wave A tail idles).
3. P3's Linux host isn't ready (WP-22 integration tests, WP-30 probe, Wave C
   dry run all need it; Docker Desktop on macOS cannot do host nftables).
4. A tier misjudgment discovered late (mitigation: dry run + retier rule,
   §6.4; the queue records findings-by-class per lane so retiering is data,
   not vibes).
5. Errata rate much higher than 10/session (H2 batches grow; the frozen
   contracts have been stable — 6 + 10 errata in the last two sessions, all
   editorial or clarifying, none invalidating shipped code).

## 9. Decision requests (P1–P6)

| # | Question | Recommendation |
|---|---|---|
| **P1** | H1/H2 cadence you will actually hold: daily? twice daily? fixed evenings? | daily, fixed time — the date is a direct function of it |
| **P2** | Which models run which tier: local qwen3.8-27b (free, slower) vs cloud small (cost, faster) for implementer/reviewer lanes; strong tier for principal? | implementers local 27b, reviewers cloud small (independence + speed), principal strong; cap cloud spend per week |
| **P3** | Linux host/VM for Wave B + Wave C (nftables, PostgreSQL, Docker, the dry-run "box") — provision now? | yes, before Wave A ends; it gates H3's egress-ADR feasibility too (E1 assumes host nftables) |
| **P4** | H3 (D10–D18 + egress ADR + R1 smoke test) this week? | yes — 2026-10-09/10; it is the cheapest critical-path item in the whole plan |
| **P5** | Confirm the tiering rule: AGENTS.md high-review-bar WPs are strong-model + H4, never small-model solo | confirm; the alternative saves ~1 week and spends it back on the first incident |
| **P6** | Concurrency cap and kill-switch: max simultaneous lanes, max open PRs ahead of merges | 3 lanes / 2 open PRs; the driver pauses, never queues deeper |

## 10. What this plan does not change

SPEC, the ADRs, DESIGN, AGENTS.md and WORKFLOW govern the pipeline exactly
as they govern sessions: stdlib-only (+pgx per ADR-0010), HTMX-only UI,
enforcement in the platform core, tests proving both the rule and the
bypass, RFC vectors on the high-review list, `agent-built` labels, protected
`main`, session records and usage accounting per delivered unit (the
principal writes one tracker per merged PR family, not one per calendar
session). The pipeline is a *scheduler*, not a new authority.
