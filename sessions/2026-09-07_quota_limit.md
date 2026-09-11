# Quota limit snapshot — contract freeze, A0–A2

> **SUPERSEDED 2026-09-11 — the resume plan below was executed in full.** A1 was
> completed (and A1-4 turned out to be a stub too), both reviews ran, all 63
> MUST FIX findings were applied, and the branch is open as
> [PR #2](https://github.com/daten-krake/Sleipnir/pull/2). See
> `sessions/2026-09-11-contract-freeze-a1.md` and `next_steps.md`. Kept as the
> record of what a quota cut costs and how to resume one.

- **Date**: 2026-09-07 (session `2026-09-07-contract-freeze`)
- **Branch**: `contracts/a0-a2-conventions` (pushed to origin — verify before
  resuming)
- **Why**: provider quota exhausted mid-workflow; all four children of workflow
  `2d8bf9eb` (A1, A2, review-principal, review-architect) failed. Earlier, the
  first A0 child was cut at 30 min (file survived, report lost).

## State on disk

| Artifact | State |
|---|---|
| `contracts/README.md` (lifecycle, doc shape, contract-test gate) | ✅ committed `7666332` |
| `contracts/A0-conventions.md` (773 lines, 8 clause groups, 14 PO items) | ✅ complete, my review pass done, committed `733eadb` |
| `contracts/A1-events.md` (347 lines) | ⚠️ **partial** — A1-1 envelope, A1-2 actor/identity, A1-3 taxonomy, A1-4 payload rules written; **A1-5 hash chain, A1-6 verification, A1-7 write path, A1-8 read/stream are empty stubs**, and §4 types / §5 traceability / §6 PO-confirm are empty |
| `contracts/A2-graph.md` (1200 lines) | ✅ looks complete (A2-1…A2-12 + types + traceability + 6+ PO items) — **uncommitted, not yet reviewed by anyone** |
| Principal + architect reviews of A0/A1/A2 | ❌ never ran (quota) |

## What is left (in order)

1. **Finish A1** (architect child, 60-min budget): write A1-5 (per-engagement
   hash chain, genesis, chainable field set, tamper-case contract tests),
   A1-6 (verification at startup/pre-export; `integrity_failed`; export
   blocked, override recorded), A1-7 (write path: platform-only composition,
   worker report-only, `events:append` dedup key per **A0-3.11**, append-only),
   A1-8 (read/stream guarantees; pagination per A0-4), then §4 types,
   §5 traceability (Q6/Q10/Q11/Q12/Q13/Q14; ADR-0009/0017/0018/0020; A11/A12),
   §6 PO-confirm. Must stay consistent with the already-written A1-1…A1-4.
2. **Commit A2** after my integrator pass over it (first read of all 1200 lines).
3. **Two reviews** (both failed, never ran):
   - `sleipnir-principal`: buildability, cross-document consistency, work-package
     split for implementation (per WORKFLOW §5), missing contract tests.
   - `sleipnir-architect`: adversarial — ADR/SPEC conflicts, safety holes
     (chain bypass, replay, cross-engagement splicing/leak, secret egress,
     untrusted-content laundering), stdlib-only feasibility of A0's
     canonical-JSON + id scheme, determinism traps, taxonomy completeness.
   Briefs survived in workflow `2d8bf9eb` (rerun-able verbatim).
4. **Apply MUST FIX / SHOULD FIX**, then cross-check traceability tables both
   ways (A0↔A1↔A2; A0 obligations A0-2.12/2.13/2.14/2.16, A0-3.11, A0-4.3,
   A0-7.7 must all be discharged).
5. **PO decision queue** (PR body will carry these as a checklist):
   - A0 §6: 14 items (id alphabet/length, `gn_`/`ge_`/`evi_`/`apr_` prefixes,
     integers-only canonical JSON, ms timestamps, unknown-field rejection scope,
     no `schema_version`, no list totals, `limit` hard-reject, kind→HTTP mapping,
     `internal/ids` + `internal/cjson` packages, …).
   - **A0-7.10 (blocks A3)**: Q4 caps are not jointly satisfiable (500 nodes ×
     512 B ≫ 64 KiB view). A3 must define the composition rule; recommended:
     views carry compact refs, full summaries only in the capped 1-hop
     drill-down.
   - A2 §6: confidence = evidence grade `observed`/`inferred`/`verified`;
     **`usr_` principal prefix — A0 amendment request on A2's critical path**;
     blacklisted discovery: record+quarantine vs refuse write; report exclusion
     by flag+event; mutable `retracted` on edges; `exploited_by` targeting
     `hypothesis`.
6. **Land the branch**: commit A1/A2 + review fixes, refresh this tracker,
   push, `gh pr create` with the PO checklist in the body (recipe: WORKFLOW §7).
7. **After the freeze** (new sessions): A3 stage views (needs the A0-7.10
   decision), A4 API/SSE envelope, A5 authn/session; then the backlog's
   contract-test skeleton and implementation per WORKFLOW §5.

## Resume plan (concrete)

1. `git status` — expect untracked: `contracts/A1-events.md`,
   `contracts/A2-graph.md`.
2. One `sleipnir-architect` child: "Complete A1-5…A1-8 + §4/§5/§6 of
   `contracts/A1-events.md`, consistent with A1-1…A1-4 and A0 (frozen). 60-min
   budget, write skeleton to disk early." (Same obligation list as the failed
   run: A0-2.12/2.13/2.14/2.16, A0-3.11, A0-4.3, A0-7.7, A0-5.6/5.7, A0-6.6.)
3. Rerun the two review children verbatim (briefs in workflow `2d8bf9eb`,
   outputs expected at `/tmp/a0a2-review-{principal,architect}.md`).
4. Integrator pass: apply findings, cross-check, commit, PR.

## Key refs

- ADRs: adr/ (0005, 0008, 0009, 0010, 0011, 0012, 0013, 0016, 0017, 0018,
  0019, 0020). Decisions: `sessions/2026-09-04-program-layout.md` Q1–Q15.
- Adversarial findings: `docs/adversarial-review-2026-09-03.md` (A1, A9, A11,
  A12 relevant here); `sessions/BACKLOG.md` items 5, 6, 7, 11.
- Style: em/en dash and quote-mark rules — `sessions/style-notes.md`.
