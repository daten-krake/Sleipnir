# NEXT STEPS — after session 2026-09-24

Goal of the next session: **get PR #4 merged, then write WP-07
`internal/cjson`** — the one canonical-JSON form (A0-2). It is the last
foundation package with a high review bar, and A1's hash chain, ADR-0018's
approval fingerprints and WP-08 `internal/paging` all depend on it.

## 0. Housekeeping (do these first)

- [ ] **PR #4** — <https://github.com/daten-krake/Sleipnir/pull/4>
      (`foundation/ids-timex-caps` → `main`, label `agent-built`). The product
      owner reviews three new foundation packages, **two A0 errata** (§6 items
      16 and 17 — both already ruled, both recorded in the session tracker),
      the **ADR-0019 §3 amendment note** and the new `AGENTS.md` delegation
      bullet.
- [ ] Branch pushed, tree clean at session close, CI green on the PR
      (`detect` + `go-gates` + `contracts`; the last runs because this branch
      touches `contracts/`). `git status` / `git fetch` first; local `main`
      needs a fast-forward after the merge.
- [ ] After the merge: delete the branch locally **and** on the remote, as in
      the 2026-09-21 and 2026-09-24 sessions.
- [ ] `gh` on Linux is still **unauthenticated**; the working recipe is
      WORKFLOW §7: `GH_TOKEN=$('/mnt/c/Program Files/GitHub CLI/gh.exe' auth token) gh …`.
      Never write the token to a file or a log (ADR-0019 §5).
- [ ] No open question is waiting on the product owner: the ADR-0019 spelling
      ruling and both errata were answered on 2026-09-24 and are recorded in
      `sessions/BACKLOG.md` → Resolved.

## 1. Context to load

Use the `start-session` ritual (fixed order: SPEC → `adr/` → DESIGN →
`sessions/BACKLOG.md` → AGENTS/WORKFLOW). **New since the last session:**

- `internal/ids`, `internal/timex`, `internal/caps` — and **their `doc.go`
  files are the entry point**. Each has a "Clauses implemented here" list, a
  "Clauses deliberately not implemented here, and where they live" list, and a
  "Rulings" list. Read them before writing a consumer: they say which clause
  belongs to which later package (A0-5.4's clamp → WP-10, mechanism R → the
  first ingest caller of `caps.Fits`, A0-1.9's collation test → `store/postgres`).
- `contracts/A0-conventions.md` §6 items 16 and 17 (the two errata), A0-5.3's
  corrected rationale, and §4.1's naming-rulings paragraph (which now names the
  owner of the two clamp test ids).
- `adr/ADR-0019-descriptive-errors-logging.md` §3's amendment note: A0-3.6
  governs the correlation-attribute spelling.
- `sessions/2026-09-24-foundation-ids-timex-caps.md` — the session record:
  18 review findings (none a code defect), four defects the implementers found
  in the principal's own briefs and in the frozen contract, and the rulings.
- `AGENTS.md`'s new delegation bullet (read-only reviewers have no shell; child
  completion previews truncate — recover the full report from the child's
  `session.jsonl`; treat your own brief as untrusted input).

## 2. Plan once PR #4 is merged

1. **WP-07 `internal/cjson`** (A0-2.1…A0-2.17 + the A0 §4 `cjson` sketch).
   `AGENTS.md` high-review bar: untrusted-input parsing on the integrity path.
   One lane, one directory, then an independent review lane. Excerpt: the six
   clauses that make it hard — `Decoder.Token()` walk with `UseNumber()` and
   the number-literal re-validation (A0-2.5), integers only (A0-2.6), the
   string-escape re-emission rules incl. literal U+2028/9 and U+007F (A0-2.7),
   UTF-8 and lone-surrogate rejection that must **not** rely on
   `encoding/json` (A0-2.3), its own depth/size bounds (A0-2.11: 32 / 1 MiB,
   the same 32 `internal/errs` reuses for its chain walk), the top-level-only
   exclusion list (A0-2.12), constant-time digest comparison (A0-2.15), and the
   **V1–V8 normative vectors** (A0-2.17) asserted byte-exactly.
   Test ids: A0 §4.1's rows for A0-2.3/2.5/2.11/2.14/2.15 and §4's `With`
   (`TestInvalidUTF8Rejected`, `TestLoneSurrogateRejected`,
   `TestDuplicateKeyRejected`, `TestCaseDuplicateKeyRejected`,
   `TestCanonicalEmitsNumberLiteralText`, `TestRejections`, `TestDepthLimit`,
   `TestSizeLimit`, `TestNilCollectionNeverSerializesAsNull`, `TestDigestEqual`,
   `TestGatingComparisonsAreConstantTime`, `TestWithAddsKeys`,
   `TestWithRejectsExistingKey`) plus the principal review's WP-07 row
   (`TestCanonicalIsStableAcrossRuns`, `TestCanonicalDecodesEscapes`,
   `TestExclusionListDropsTopLevelOnly`, `TestSHA256HexLowercase64`). The
   vectors are **data, not a test id** (§4.1): name their test for V1–V8 — the
   review report's `TestVectorsV1ToV7ByteExact` predates V8.
   `docs/reviews/2026-09-11-verify-vectors.py` is the reference implementation
   of the arithmetic in Python and CI recomputes every vector on a PR touching
   `contracts/`; if cjson disagrees with it, one of them is wrong — find out
   which before merging. `cjson` imports only `errs` (A0 §4 preamble).
2. **WP-13 `internal/secretscan`** is parallel to WP-07 (it imports only
   `errs`): A2-9.4's closed rule table + the P-46 rule ids, A2-9.5's
   "name the field and the rule id, never the value" message rule, A1-4.9. It
   is the one lane whose brief must quote **A2** verbatim — the rule table does
   not exist anywhere else, and six named secret-free-serialization tests
   cannot be written until it does.
3. **WP-08 `internal/paging`** (A0-4.1…A0-4.8) after WP-07: `Page[T]`,
   `Cursor`, `EncodeCursor`/`DecodeCursor` over `cjson.CanonicalValue` +
   `RawURLEncoding`, validating the decoded cursor's id against `ids.Kind`
   (A0-4.4, A0-1.5).
4. Then the domain packages in the principal review §4's order: WP-09…WP-12
   (`events`: envelope/taxonomy, chain primitives, validation, verification
   walk), WP-14…WP-16 (`graph`), WP-17 store seams, WP-18 ingest mapping,
   WP-19…WP-21 the shared contract suite per owner package, WP-22
   `store/postgres` (pgx vendored, ADR-0010, integration opt-in per DESIGN §8 —
   it owns `TestIDOrderingMatchesByteOrderCollateC`).
5. Design sessions still queued: **A3 stage views** (unblocked by D4; owns the
   real A0-7.10 composition rule and must budget bytes for ADR-0022's
   provenance grade), backlog §3 UI, §4 Pi node, §5 agent loop, §6 persistence,
   §7 security work items, §8 SSO, §9 CI remainder (image build/pin/sign, full
   SHA action pinning), §10 observability (incl. the attribute-redaction seam),
   §11 model benchmarking/drift.

## 3. Definition of done for the next session

- [ ] PR #4 merged, branch deleted locally and remotely, local `main`
      fast-forwarded, CI green on `main`.
- [ ] WP-07 reviewed independently and merged or opened as its own PR: every
      A0-2.17 vector byte-exact, the whole rejection list covered by
      `TestRejections`, both bounds (32 / 1 MiB) tested **by asserting their
      effect** (AGENTS.md: never by removing them).
- [ ] `gofmt -l`, `go vet ./...`, `go build ./...`, `go test ./...`,
      `go test -race ./...` green under `ulimit -v` / `GOMEMLIMIT` / `-timeout`;
      `go.mod` still empty except pgx (ADR-0010); `verify-vectors.py` still
      `PASS 52 / FAIL 0`.
- [ ] Session tracker + `sessions/BACKLOG.md` + this file updated, usage
      refreshed, PR labelled `agent-built`, PR URL verified to exist.

## 4. Process lessons to keep (they cost real sessions to learn)

- **A child's deliverable must be a file written incrementally**, never its
  final message; require a compiling skeleton in the first 10 minutes and a
  "stop and make it coherent at minute N−10" rule.
- **Treat a child's own progress log as untrusted**: re-audit the file state
  before briefing the next pass.
- **Treat your own brief as untrusted too** (new 2026-09-24): three
  implementers found four defects in the principal's briefs and in the frozen
  contract in one session — a wrong byte count, a wrong claim about Go's
  formatting, an impossible test row, a false rationale. A child that surfaces
  a bad instruction instead of following it is succeeding.
- **A child's completion preview truncates mid-finding.** Read the full report
  from `~/.pi/agent/sessions/<parent>/<child-run>/run-0/session.jsonl`; `/tmp`
  is not durable (2026-09-21).
- **Brief reviewers for the tools they have** (new 2026-09-24): the builtin
  `reviewer` has no shell, so it cannot run the gates — say so, run them
  yourself, and forbid filesystem-wide searches (one reviewer burned its whole
  budget in `find /` and lost its report).
- **One writer per file/directory.** Three parallel lanes on three disjoint
  directories worked twice; paired blocks must be byte-identical and verified
  afterwards.
- **Never prove a negative test by removing the bound and running it**
  (2026-09-21: 11.4 GB RSS, kernel OOM, the whole WSL VM down). Assert the
  bound's effect. Now an `AGENTS.md` rule.
- Verify every published vector mechanically before merging.
