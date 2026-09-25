# NEXT STEPS — after session 2026-09-24

Goal of the next session: **WP-08 `internal/paging`** (A0-4, the one list
envelope and its cursors) — the first package that needs two siblings, `ids`
and `cjson`, both merged today. If the session has room, **WP-13
`internal/secretscan`** runs in a parallel lane (it imports only `errs`); it is
the one lane whose brief must quote **A2** verbatim, because its closed rule
table exists nowhere else.

## 0. Housekeeping (do these first)

- [ ] **PR #6** — <https://github.com/daten-krake/Sleipnir/pull/6>
      (`sessions/close-2026-09-24` → `main`, label `agent-built`): the close
      ritual for 2026-09-24 — both trackers marked merged, `sessions/BACKLOG.md`
      §2 + two standing questions, and this file. Docs only.
- [ ] Nothing else is pending. PR #4 merged as `dd5837c`, PR #5 as `997bbcc`,
      both feature branches deleted locally **and** on the remote, local `main`
      fast-forwarded, and the gates re-run on `main`: all six foundation
      packages green (`errs`, `logging`, `ids`, `timex`, `caps`, `cjson`),
      `verify-vectors.py` **PASS 52 / FAIL 0**, A0 §6 items 16–18 present.
- [ ] `gh` on Linux is still unauthenticated; WORKFLOW §7's recipe works:
      `GH_TOKEN=$('/mnt/c/Program Files/GitHub CLI/gh.exe' auth token) gh …`.
      Never write the token to a file or a log (ADR-0019 §5).
- [ ] CI triggers only on `pull_request` and on pushes to `main`, so a feature
      branch's first pipeline run is the PR that opens it.

## 1. Context to load

`start-session` ritual, fixed order: SPEC → `adr/` (README + anything changed
since 2026-09-24: **ADR-0019 §3 now carries an amendment note** — A0-3.6 governs
the correlation-attribute spelling) → DESIGN → `sessions/BACKLOG.md` →
AGENTS/WORKFLOW. **New since the last session:**

- `internal/ids`, `internal/timex`, `internal/caps`, `internal/cjson` — and
  **their `doc.go` files are the entry point**. Each lists the clauses it
  implements, the clauses it deliberately does not *with the package that owns
  them*, and its rulings. Read `cjson`'s before writing anything that hashes:
  key identity is `foldKey` (orbit-minimum `unicode.SimpleFold`), exclusion
  folds both sides, a null at any depth is `validation` (A0-8.3 settles
  A0-2.14), `With`'s collision is `internal`, and `DigestEqual` needs its
  `sha256.Size` width check because `subtle.ConstantTimeCompare` reports two
  empty slices as equal.
- `contracts/A0-conventions.md` §6 items **16, 17, 18** — the three errata ruled
  2026-09-24 (the truncation marker is 11 B; A0-5.3's leap-second rationale was
  false and §4.1 now names WP-10 as the clamp ids' owner; A0-2.11 says when each
  bound bites and A0-2.14 says who its `internal` is for).
- `sessions/2026-09-24-foundation-ids-timex-caps.md` and
  `sessions/2026-09-24-cjson.md` — two trackers for one session (the
  2026-09-07 precedent): 25 review findings across four packages, **not one a
  code defect in the first three**, and six defects the children found in the
  principal's own briefs and in the frozen contract.
- `AGENTS.md`'s delegation bullet (new 2026-09-24): brief a child for the tools
  it actually has, recover full child reports from
  `~/.pi/agent/sessions/<parent>/<child-run>/run-0/session.jsonl` because
  completion previews truncate mid-finding, and treat your own brief as untrusted
  input.

## 2. Plan

1. **WP-08 `internal/paging`** — A0-4.1…A0-4.8 (`contracts/A0-conventions.md`
   lines 396–455), A0-8.6's base64url rule (lines 680–718), the §4 `paging`
   sketch (around line 901). `Page[T]`, `Cursor`, `EncodeCursor`,
   `DecodeCursor(s string, k ids.Kind)`. It is the first package to import two
   siblings (`ids` + `cjson`), so its brief must state that A0 §4's
   "foundation packages import nothing internal but `errs`" applies to the
   *other* five and that the sketch itself declares this dependency
   (`cjson.CanonicalValue` + `RawURLEncoding`; `ids.Kind` for the cursor's id).
   Test ids: A0 §4.1's rows A0-4.4/A0-4.5/A0-4.6/A0-8.6 —
   `TestDecodeCursorRejects`, `TestCursorWithInconsistentKAndIDRejected`,
   `TestLimitValidation`, `TestLimitAboveMaxRejectedNotClamped`,
   `TestHasMoreDetection`, `TestCursorRejectsStandardAlphabetAndPadding` — plus
   the WP-08 review row's `TestCursorRoundTripIsCanonicalBase64URL`.
   §4.1's naming ruling binds: `TestExactFullPageHasNoNextCursor` **is**
   `TestHasMoreDetection`, not a second id; and an inconsistent `k`/`id`
   cursor has exactly one oracle across A0/A1/A2 — `validation` (400), never
   "empty page or `validation`".
2. **WP-13 `internal/secretscan`** (∥, imports only `errs`) — A2-9
   (`contracts/A2-graph.md` lines 716–804: the closed rule table, the P-46 rule
   ids, A2-9.5's "name the field and the rule id, never the value, a prefix or
   its digest") plus A1-4.9 (`contracts/A1-events.md`, from line 439). Test ids:
   `TestEveryRuleIDMatchesItsCorpusValue`,
   `TestNoFalsePositiveOnBenignCorpus` (hostnames, CIDRs, argv, hex digests of
   artifacts), `TestErrorMessageNamesFieldAndRuleIDOnly`,
   `TestEntropyRuleIsDeterministic`. It unblocks six named secret-free-
   serialization tests in A1/A2 and A0-3.4's two ids. Its brief is the only one
   that quotes A2, and it must not read A0's other sections.
3. Then the domain packages in the principal review §4's order: WP-09…WP-12
   (`events`: envelope/taxonomy, chain primitives, validation, verification
   walk — WP-10 owns A0-5.4's `recorded_at` clamp and the two clamp test ids),
   WP-14…WP-16 (`graph`), WP-17 store seams, WP-18 ingest mapping, WP-19…WP-21
   the shared contract suite per owner package, WP-22 `store/postgres` (pgx
   vendored per ADR-0010, integration opt-in per DESIGN §8 — it owns
   `TestIDOrderingMatchesByteOrderCollateC`, backlog §6).
4. Two debts to place when the domain types land: **A0-2.6's "no float field in
   a canonicalized type"** is enforceable only by review (`json.Marshal(float64(2))`
   emits `2`), and **UTF-8 validation of string fields** has no owner yet
   (`CanonicalValue` cannot do it — `encoding/json` substitutes U+FFFD first).
   Both are standing questions in `sessions/BACKLOG.md`.
5. Design sessions still queued: **A3 stage views** (unblocked by D4; owns the
   real A0-7.10 composition rule and must budget bytes for ADR-0022's
   provenance grade), backlog §3 UI, §4 Pi node, §5 agent loop, §6 persistence,
   §7 security work items, §8 SSO, §9 CI remainder (image build/pin/sign,
   full-SHA action pinning), §10 observability (incl. the attribute-redaction
   seam), §11 model benchmarking/drift.

## 3. Definition of done for the next session

- [ ] PR #6 merged, branch deleted locally and remotely, local `main`
      fast-forwarded, CI green on `main`.
- [ ] WP-08 reviewed independently and merged or opened as its own PR: every
      A0-4 MUST covered, cursor round-trip byte-stable, the standard alphabet
      and `=` padding rejected (A0-8.6), and an over-max `limit` **rejected,
      never clamped** (A0-4.5).
- [ ] `gofmt -l`, `go vet ./...`, `go build ./...`, `go test ./...`,
      `go test -race ./...` green under `ulimit -v` / `GOMEMLIMIT` / `-timeout`;
      `go.mod` still empty except pgx (ADR-0010); `verify-vectors.py` still
      `PASS 52 / FAIL 0` if `contracts/` was touched.
- [ ] Session tracker + `sessions/BACKLOG.md` + this file updated, usage
      refreshed, PR labelled `agent-built`, PR URL verified to exist.

## 4. Process lessons to keep (they cost real sessions to learn)

- **A child's deliverable must be a file written incrementally**, never its
  final message; require a compiling skeleton in the first 10 minutes and a
  "stop and make it coherent at minute N−10" rule.
- **Treat a child's own progress log as untrusted** — re-audit the file state
  before briefing the next pass.
- **Treat your own brief as untrusted too.** On 2026-09-24 children found six
  defects in the principal's input: a wrong byte count and a false rationale in
  the frozen contract (→ errata 16 and 17), Go's `.000` directive truncating
  rather than rounding, an impossible test row, an unachievable "before
  canonicalization" reading (→ erratum 18), and a **self-contradictory remedy**
  whose two properties disagreed on `{A, a}` and would have broken A0-2.5's own
  example. A specification that names two properties must be checked for a case
  where they conflict, and a reviewer's suggested code is a proposal, not a
  ruling.
- **A child's completion preview truncates mid-finding.** Read the full report
  from `~/.pi/agent/sessions/<parent>/<child-run>/run-0/session.jsonl`; `/tmp`
  is not durable (2026-09-21).
- **Brief reviewers for the tools they have.** The builtin `reviewer` has no
  shell (it cannot run the gates) and once burned its whole budget in `find /`.
  Its launch guard also rejects dense review briefs as "implementation work" —
  it fired three times on the cjson review and not once on the shorter ones.
  `sleipnir-architect` has a shell, passes the guard, and can recompute vectors
  independently; it can write, so the brief must forbid it and `git status`
  must be checked afterwards.
- **An assertion that cannot fail is not a test.** Three separate findings today
  were the same defect: an import allow-list that never required `crypto/rand`,
  a truncation corpus that left the zero-budget walk-back path uncovered, and a
  source scan hardcoding the parameter names `a`/`b` so a rename would vacate
  it. Ask of every guard: what edit would make this pass while breaking the rule?
- **One writer per file.** Two PRs appending to the same numbered list (A0 §6)
  is a guaranteed conflict — erratum 18 was ruled during WP-07 and still went to
  the branch that already owned §6.
- **Never prove a negative test by removing the bound and running it**
  (2026-09-21: 11.4 GB RSS, kernel OOM, the whole WSL VM down). Assert the
  bound's effect, or mutate a copy outside the repo and delete it — which is
  what the cjson fix lane did, and reported.
- Verify every published vector mechanically before merging, and prefer an
  oracle the code under test did not produce.
