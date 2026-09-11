# Fix plan — A0/A1/A2 contract freeze

Author: principal build engineer. Inputs: `docs/reviews/2026-09-11-contract-review-principal.md`
(P-01…P-105), `docs/reviews/2026-09-11-contract-review-adversarial.md` (C/S/F/D/T/E), read in full.
Repo **untouched** (read-only run; the only file written is this plan).
Three writers execute this plan mechanically: **A0** → `contracts/A0-conventions.md`,
**A1** → `contracts/A1-events.md`, **A2** → `contracts/A2-graph.md`. §8 is their shared rule set.

## 0. Summary

| | MUST FIX | SHOULD FIX | NICE | total | assigned here | deferred (§7) |
|---|---|---|---|---|---|---|
| principal (P-*) | 35 | 62 | 8 | 105 | 35 MUST + 69 SHOULD/NICE | **1** (P-77) |
| adversarial (C/S/F/D/T/E) | 28 | 20 | 9 | 57 | 28 MUST + 21 SHOULD/NICE | **8** |
| **both** | **63** | 82 | 17 | **162** | **63 MUST + 90 SHOULD/NICE = 153** | **9** |

- **All 63 MUST FIX ids are assigned exactly once** (machine-checked: no id missing, no id in two
  rows): §1 PO-1 carries 2 (C-01, S-07), §2 (A0) 13, §3 (A1) 30, §4 (A2) 18. Row counts: **9 PO
  rows + 21 A0 rows + 34 A1 rows + 25 A2 rows**. §7 defers 9 SHOULD/NICE ids plus two
  integrator-only `contracts/README.md` sentences.
- **Agent-fixable vs PO-only:** 61 of 63 MUST FIX ids are agent-fixable (a writer inserts the text
  this plan specifies). **2 are PO-only** — `C-01` and `S-07` (both A1-6.5) — and land as §1 PO-1
  with a SAFE DEFAULT the A1 writer inserts now. Nine §1 PO rows exist in total; seven carry **no**
  finding id (they are §6 open questions: A0 §6.14, A1 §6.7, A1 §6.9, A2 §6.3, reject-vs-redact,
  Q2 deviation, AM-1) and one (PO-9) is a PO-*confirm* of a decision the writers still apply.
- **Deduplicated findings: 14 pairs/groups** (one row, both ids cited, one writer owns):
  P-03+S-01 · P-04+F-03 · P-05+F-02 · P-06+F-06 · P-42+S-05 · P-47+S-04 · P-63+C-06+T-06 ·
  P-64+T-04 · P-82+D-01 · P-43+E-07 · P-44+P-45+P-54 · P-19+D-06+F-01(+F-09) · P-23+F-04 ·
  P-31+S-08. Overlaps the task asked me to check and my ruling on each:
  **P-03/S-01** same clause, different facet → merged. **P-04/F-01** *not* duplicates (F-01 is
  number handling; P-04 pairs with **F-03**) → separate rows A0-03/A0-04. **P-05/F-02** duplicates
  → merged. **P-06/D-02** *not* duplicates (D-02 is A2-owned: which type is fingerprinted; P-06
  pairs with **F-06**) → separate rows A0-06/A2-16. **P-41/D-01** adjacent, not duplicates
  (array order vs missing vector) → rows A2-02/A2-03, vector computed after the sort rule.
  **P-42/S-05** duplicates with incompatible fixes → merged, adversarial wins (§5 conflict 3).
  **P-46 with A1 §6.4/A2 §6.10** → P-46 is the rule table (A2-07); the reject-vs-redact question is
  PO-6, ruled once, cited from both. **P-63/C-06/S-04** → P-63+C-06+T-06 merged (vocabulary,
  A1-20/A2-12); S-04 merged with P-47 (matched fields, A2-08) and its A1 enum half is A1-28.
  **C-01/S-07** → one PO row (PO-1). **E-01…E-04** → enforceability roll-ups: E-01→A1-25,
  E-02→A2-13, E-03→A1-24 (with P-97, S-13), E-04→A1-26 (test names assigned explicitly; the
  underlying clauses A1-2.3/2.6/2.7/3.4/7.4, A1-5.3/6.2/7.11, A2-8.3/8.5/8.8 need no other change).
- **Paired rows (identical wording required in 2–3 files): 17 pair ids spanning 31 rows** (each pair
  has one owner row and one or two partner rows).
  PAIR-U1 (`usr_`, A0+A1+A2) · PAIR-Q1 (quarantine mapping/release, A1+A2) · PAIR-Q2 (quarantine
  recompute triggers + `target_quarantined`, A1+A2) · PAIR-G1 (`content_hash`/`dedup_hit` on graph
  kinds, A1+A2) · PAIR-M1 (A1↔A2 field mapping, A1+A2) · PAIR-A1 (actor↔principal_kind mapping,
  A1+A2) · PAIR-A2 (`principal_kind` ≠ `actor`, A1+A2) · PAIR-SEC1 (secret rule table ownership +
  filter-not-guarantee, A1+A2) · PAIR-N1 (cap registry adoption, A0+A1+A2) · PAIR-N2 (no `null`,
  A0+A1+A2) · PAIR-K1 (cursor validity, A0+A1) · PAIR-C1 (`cjson.With` / `Served`, A0+A1) ·
  PAIR-T1 (`recorded_at` clamp, A0+A1) · PAIR-R1 (retryable `internal`, A0+A1) · PAIR-V5
  (V5 relabel, A0+A2) · PAIR-V1 (U+2028/9 literal, A0+A1) · PAIR-X1 (no secret echo, A0+A1+A2).
  (17 pair ids; 11 of them require *byte-identical* insert text — marked ⧉ in the tables.)
- **Id-appearance convention (enforces the exactly-once rule):** every finding id appears in the
  `finding ids` cell of exactly **one** row in §1–§4, or in §7. (§1's table has the ids in its
  *second* column.) A paired partner row's id cell reads
  `PAIR-xx (partner; decision owned by §N row YY-nn)` and names **no** finding id of its own, so a
  mechanical scan of id cells yields each id once.
- **Verified by recomputation while writing this plan** (python3, recipe in §8): A0-2.17 **V1–V6
  all reproduce byte-exactly** (lengths 13/26/57/45/113/18 and all six published digests are
  correct); A1 §4.3 **rows 0–2 all reproduce byte-exactly** (428/816/715, 589/977/876, six
  digests). Therefore every new vector in §6 is python3-computable and the values published there
  are final.

## 1. PO-DECISION-ONLY (agents must not decide these)

Each row: the writer applies the **SAFE DEFAULT** text now, and adds the **§6 sentence** verbatim.
Nothing here is decided by a writer; the PO signs before the documents flip to `Frozen`.

| # | finding ids | clause | question | SAFE DEFAULT the writer inserts now | exact sentence to add to that document's §6 |
|---|---|---|---|---|---|
| PO-1 | **C-01, S-07** | A1-6.5 (+A1-6.4, A1-6.6) | Who may override a chain break, and for how long? A1-6.5 says "only a user principal with the admin role" *and* "an operator-scoped user MAY NOT override **another engagement's** break" (implies operators may override their own); Q11 says "explicit operator override allowed". Separately, one override currently authorizes an unbounded stream of artifacts. | **A1 writer.** In A1-6.5 delete the "another engagement's" sentence and insert BLOCK-PO1-A (admin-only, operator MUST NOT, `forbidden` naming no object) and BLOCK-PO1-B (an `integrity_override` with `scope:"export"` is **single-use**: uniqueness on `(engagement_id, override_event_id)`, bound atomically to one `artifact_released` event — added by §3 A1-03/T-02; a second release → `conflict` + `integrity_failed`). Add `Tests: TestOverrideIsAdminOnly, TestOperatorCannotOverrideAnyBreak, TestSingleUseOverride, TestOverrideDiesAtNextBreak`. | `15. **A1-6.5 — override authority and lifetime (PO decision, not confirmation).** The clause now narrows Q11's "explicit operator override allowed" to **admin-role only** (an operator-scoped user, including one assigned to the engagement, MUST NOT override any break) and makes an `scope:"export"` override **single-use**, one override per released artifact. Both narrowings are stricter than the literal Q11 wording; the product owner MUST sign them before A1 flips to Frozen (adversarial C-01, S-07; insider threat A15).` |
| PO-2 | — | A0-7.10 / A0 §6.14 | How is a stage view's byte budget composed from per-node caps (500 × 512 B ≈ 250 KiB ≫ 64 KiB)? | **A0 writer.** Keep A0-7.10's "declared conflict, not resolved here" and append BLOCK-PO2 (fail-safe interim rule: where two registry values cannot both hold, the **smaller** governs; a view builder MUST stop at the first cap it reaches and set the truncation marker; A3 owns the composition rule). | Replace the tail of A0 §6 item 14 with: `**Interim rule applied for the Freeze (BLOCK-PO2, A0-7.10): where two A0-7.1 caps cannot both hold for one composed document, the smaller governs and the builder truncates with mechanism T. The product owner MUST confirm this fail-safe default or replace it with A3's composition rule before A3 is drafted; ~131 B per node is not a usable view budget.` |
| PO-3 | — (the fix is assigned in §3 row A1-09) | A1 §6.7 / A1-5.8 | Is `(engagement_id, head_seq, head_hash)` also pushed over the ADR-0012 §3 signed webhook (the only v1 mechanism that closes tail truncation against a store-capable attacker)? | **A1 writer.** Apply §3 A1-09 (S-02: `head_regression` break kind + append-only `chain_head_trail` table with `REVOKE UPDATE, DELETE` + emission at every verification, every 100-`seq` crossing and every `run_ended`/`hard_stop_fired`). Do **not** add the webhook push. | `16. **A1-5.8 / §6.7 — out-of-band head anchoring (PO decision).** The Freeze ships the in-platform anchor only: an append-only `chain_head_trail` table (`REVOKE UPDATE, DELETE`) plus `break_kind:"head_regression"`. Residual risk accepted unless the PO also approves pushing `(engagement_id, head_seq, head_hash)` over the ADR-0012 §3 signed webhook: an attacker with both store-write and log-write access on the same host can forge history (default Docker deployment, SPEC §10). This residual MUST be printed into A1-6.6's export wording if the webhook anchor is declined.` |
| PO-4 | C-08 (SHOULD, rides here) | A1 §6.9 | Who owns user/session and engagement-assignment audit — A1's per-engagement chain or A5's platform-scoped store? | **A1 writer.** Keep the §6.9 gap as written, but insert BLOCK-PO4 into A1 §6.9: authentication/session audit stays **out of A1** (A5, platform-scoped store); **blacklist mutation is not part of this gap** (it is engagement-reachable safety state and is chained per §3 A1-19/C-04); engagement **assignment/role** changes are engagement-scoped and A5 MUST add kind `engagement_assignment_changed` additively (A1-3.5) at its own freeze. | `17. **A1 §6.9 — user/session audit ownership (PO decision).** A1 chains engagement-scoped facts only; authentication, session and role audit belong to A5 in a platform-scoped store. **Known debt accepted at Freeze:** engagement operator assignment (who may approve, ADR-0012 §1) is unaudited in v1 until A5 adds `engagement_assignment_changed` additively (A1-3.5); an insider admin self-assigning and then approving is detectable only in A5's store (adversarial C-08).` |
| PO-5 | — | A2 §6.3 / A2-8.5 / A2-8.10 | Is a blacklisted discovery **recorded** (quarantined node + event) or **refused** (never stored)? | **A2 writer.** Record it: keep A2-8.5's recommendation as normative text — a blacklisted discovery is stored with `quarantine_reason:"blacklisted"`, permanently non-releasable, absent from planning views (A2-12.5), reported as "not tested", and MUST emit `graph_node_quarantined{blacklist_match}` (A2-8.10). | `12. **A2-8.5 — blacklisted discovery: recorded, not refused (PO confirm).** The Freeze stores the node with `quarantine_reason:"blacklisted"`, never releasable, never in a planning view, and chains `graph_node_quarantined{blacklist_match}` — "we saw the forbidden target and did not touch it". The alternative reading of ADR-0016 §2 (refuse the write, store nothing about a forbidden system) is defensible and minimizes stored data; the product owner MUST confirm before Frozen, because refusing the write makes the near-miss unprovable in a customer report.` |
| PO-6 | — ⧉ | A1 §6.4 ≡ A2 §6.10 | Secret material found in a write payload: **reject** the write, or **redact** the value and keep the record? Asked twice, in two documents. | **Ruled once here: REJECT.** Both writers insert BLOCK-PO6 (byte-identical) into their own §6 item, and both must confirm the normative clauses already say reject (`validation` naming field + rule id, value never echoed, rejection chained as `action_blocked`). | BLOCK-PO6 (byte-identical in A1 §6.4 and A2 §6.10): `**Ruled once for both contracts (PO confirm): reject, never redact.** A secret-pattern hit (A2-9.4 rule ids) is a hard reject — `validation` (400) naming the field and the rule id, the value never echoed in whole, in part or as a digest (A0-3.4) — and the rejection is chained (`action_blocked`). Redaction was rejected: a false positive would silently destroy a worker's only report of what it ran, and `redacted:true` (A1-4.6) means platform redaction, never rejection. The false-positive risk is controlled by shipping the A2-9.4 rule table with the planted-secret corpus. A1 §6.4 and A2 §6.10 are the same question and MUST NOT be answered differently.` |
| PO-7 | — | A2-2.7 / A2-5.6 (A2 §6.1) | A2 deviates from Q2's literal "Finding carries confidence": `confidence` is the provenance evidence grade (`observed`·`inferred`·`verified`), one per node/edge, not a finding field. | **A2 writer.** Keep A2's design unchanged (it is the safer reading: a worker cannot assert confidence in its own claim without an evidence grade). Ensure A2-2.7 and A2 §6.1 both say the deviation is **a change to a locked PO decision and requires signature, not confirmation**. | `13. **A2-2.7 / A2-5.6 — deviation from Q2 (PO signature required, not confirmation).** Q2 records "Finding carries confidence". A2 implements that as the mandatory provenance grade `observed · inferred · verified` on every node and edge instead of a finding field, so a grade is always tied to a referenced event (ADR-0016 §1: evidence, not opinion). This **changes a locked decision**; the product owner MUST sign it before A2 flips to Frozen. With §4 A2-04 (provenance set) `verified` is now reachable: it requires a second, independent observation.` |
| PO-8 | AM-1, P-78, C-10 ⧉ | A0-1.2 + A0 §4 (`ids.Kind`) | Register a human-principal id prefix? A1 §6.2 and A2 §6.2 both declare themselves freeze-blocked on it. | **Principal's §3 ruling is the default: ACCEPT.** A0 writer registers the row `\| user (human principal, SPEC §3) \| usr_ \| usr_01m1y2whfhv3x6z9b2d5f8h1jk \| ^usr_B{26}$ \| 30 \|` in A0-1.2 and adds `KindUser Kind = "usr_"` to the A0 §4 `ids` const block; A1 and A2 writers insert the identical citation sentence (BLOCK-PO8) in A1-2.2 and A2-5.3. Do **not** delegate the shape to A5. | BLOCK-PO8 (byte-identical, A0 §6 / A1 §6.2 / A2 §6.2): `**AM-1 — resolved by default for the Freeze (PO confirm):** A0-1.2 registers the human-principal prefix `usr_` (`^usr_B{26}$`, 30 B) and A0 §4 adds `KindUser`. A0 owns id *shapes*; delegating the spelling to A5 would split A0-1.5 validation across two contracts. Every user-composed A1 kind (`actor.principal_id`, A1-2.2) and every A2 operator write (`user_id`, A2-5.3) validates against it. The product owner MUST confirm the prefix spelling before Frozen; it is additive-only afterwards (A0-1.10).` |
| PO-9 | — (ids owned by §3 A1-20 / §4 A2-12) ⧉ | A1-3.3 `quarantine_kind` / A2-8.1–8.5 | May an operator manually **release** a quarantined node? P-63 offers "add the operation to A2-8"; C-06 demands deleting `operator_release` (A2-8.5: `out_of_scope` changes only via a scope change; ADR-0016 §2: an out-of-scope node can never be a target). Both readings are safety-relevant and incompatible → stricter chosen. | **Stricter option applied: `operator_release` is DELETED** from A1's `quarantine_kind` enum (pre-Freeze, additive-only afterwards); `operator_quarantine` is **kept** (tightening only). Release happens solely as `quarantine_recomputed{scope_changed}` + per-node recomputation. Both writers apply BLOCK-PO9's decision content (⧉ byte-identical) and add the §6 sentence. | BLOCK-PO9 §6 sentence (byte-identical in A1 §6 and A2 §6): `**PO confirm — operator release of quarantine is removed.** `operator_release` is deleted from A1's `quarantine_kind` enum and A2 provides no release operation: an `out_of_scope` node is released **only** by an operator scope change and the recomputation it causes (A2-8.5, ADR-0016 §2 — an out-of-scope node can never be a target of a planned action). `operator_quarantine` (tightening) is kept, and a `blacklisted` node is never releasable. If the product owner wants a manual release it MUST be a new ADR amending ADR-0016 §2 and MUST require the target to be inside the widened allowlist at release time.` |

**§1 counts:** 9 PO rows; 2 MUST FIX ids (C-01, S-07), 3 non-MUST ids (C-08, P-78, C-10) and the
AM-1 ruling are assigned here; the remaining seven PO rows carry no finding id.

## 2. Assignment — contracts/A0-conventions.md (writer: A0)

Apply in row order. Rows marked ⧉ carry text that another writer inserts byte-identically.

| finding ids | clause(s) | exact change to make | paired with |
|---|---|---|---|
| P-02, P-35 | A0 §4 intro + caps block | Replace the §4 preamble with BLOCK-A0-01: A0 requires **six** foundation packages (zero internal imports except `errs`): `internal/ids` (A0-1), `internal/cjson` (A0-2), `internal/errs` (A0-3), `internal/paging` (A0-4), `internal/timex` (A0-5), `internal/caps` (A0-7). Give the three homeless sketch blocks a `package` line (`package paging`, `package timex` — name avoids shadowing stdlib `time`, same reasoning DESIGN §1 gives for `logging` —, `package caps`). Change `Truncate(s string, cap int)` → `func Truncate(s string, limit int) (out string, truncated bool)` (`cap` shadows the builtin) and state: `limit < len(TruncationMarker)` is a platform defect → return `"", true`, caller surfaces `internal`. (The report's prose says "five" and lists six: **six** is correct.) | — |
| P-03, S-01 (+AM-4, P-22, F-08) | A0-4.3, A0-4.4, A0-4.5, A0-4.8, §4 `DecodeCursor` | BLOCK-A0-02. (a) `func DecodeCursor(s string, k ids.Kind) (Cursor, error)` — the caller passes the id kind of the collection paged; `DecodeCursor` validates `id` against it (A0-1.5) and returns `validation` on mismatch. (b) **Seek rule (S-01):** the platform MUST derive the seek position from the row its lookup of the cursor's `id` resolves to and MUST ignore `k` for seeking; if `k` disagrees with that row's ordering value the response is `validation` (A0-4.8); replace "at worst an empty page or `validation`" with "an empty page, `validation`, or a page whose ordering value is the one the platform resolved — never a silently skipped range. Cursors are not integrity-protected and grant nothing." (c) **AM-4 (MUST):** add to A0-4.8's `validation` list: "the cursor's `id` does not resolve in this collection". (d) A0-4.5: "`limit` absent → 100; present but unparseable, ≤ 0, non-integer or > 1000 → `validation`, never clamped" (delete "absent-and-unparseable"). (e) A0-4.3: "every paginated collection MUST have an integer primary ordering key so A0-4.4's `{k,id}` cursor is expressible; a text key is only the tie-breaker carried in `id`. A collection whose order key is text MUST declare its own cursor shape in its own contract; A0-4.4's two-key set is closed." `Tests: TestDecodeCursorRejects, TestCursorWithInconsistentKAndIDRejected, TestHasMoreDetection, TestLimitValidation`. | ⧉ PAIR-K1 → §3 A1-15 |
| P-04, F-03 (+P-20) | A0-2.7, A0-2.17 | BLOCK-A0-03. A0-2.7 gains: "`Canonical` MUST decode every JSON string escape in the input (`\uXXXX` incl. well-formed surrogate pairs, `\n`, `\"`, `\\`) and re-emit per this clause; it MUST NOT pass an input escape through. **U+2028 and U+2029 are emitted literally (raw UTF-8, 3 B each), not as `\u2028`/`\u2029`; U+007F is literal (1 B).** `encoding/json` escapes U+2028/9 unconditionally and HTML-escapes `<>&`, so it MUST NOT be the canonical emitter — it MAY produce the intermediate bytes that the canonicalizer re-parses (`CanonicalValue`), never the final ones." Add vectors **V7** and **V8** with the exact bytes/len/digest from §6.2. In V6's "Canonical bytes (exact)" cell print the **literal** U+FFFD and add "(the U+FFFD key is the three bytes `EF BF BD`; the escape appears only in the input column; len 18 and the published digest are correct)". | ⧉ PAIR-V1 → §3 A1-07 |
| F-01 (+P-19, D-06, F-09) | A0-2.5, A0-2.6, A0-2.11 | BLOCK-A0-04. A0-2.5: "the `Token()` walk MUST call `Decoder.UseNumber()` and MUST re-validate every number's **literal token text** against `^-?(0\|[1-9][0-9]{0,15})$` before range-checking with `strconv.ParseInt` against A0-2.6's `[-(2^53-1), 2^53-1]`; a token whose text is not already canonical (`.`, `e`/`E`, `+`, leading zeros, `-0`) is `validation`. The canonicalizer emits the validated literal text verbatim. The walk MUST call `Decoder.More()` after the top-level value (A0-2.3 no trailing data) and MUST count nesting depth itself (`encoding/json` has none)." A0-2.11: "depth counts container boundaries — the top-level object is level 1, each nested object or array adds 1, scalars are not levels; a document at depth 32 MUST be accepted, at 33 rejected (`validation`); size is `len(doc)` of the input; both are checked before canonicalization. This walk is AGENTS.md high-review untrusted-input parsing." Add rejection vectors `{"a":1E3}`, `{"a":-0.0}`, `{"a":9007199254740993}`, `{"a":10000000000000000000}`, a 32-deep accept / 33-deep reject pair and a 1 MiB boundary pair. `Tests: TestRejections, TestDepthLimit, TestSizeLimit, TestCanonicalEmitsNumberLiteralText`. | — |
| P-05, F-02 | A0-2.3, A0-2.7, A0-8.9 | BLOCK-A0-05. A0-2.3: "implementations MUST NOT rely on `encoding/json` for UTF-8 validation — it substitutes U+FFFD. The canonicalizer MUST (a) reject the whole document when `utf8.Valid(doc)` is false and (b) scan every `\u` escape and reject a surrogate code point (`D800`–`DFFF`) not followed by a complementary surrogate forming a valid pair; both `validation`." A0-8.9: "the handler MUST call `utf8.Valid` on the raw body before decoding and reject with `validation`; `cjson.Canonical` MUST call `utf8.Valid(doc)` (A0-2.3). Relying on the decoder is a defect." Add rejection vectors `{"a":"\ud800"}`, `{"a":"\xff"}` and accept vector `{"a":"\ud83d\ude00"}` (literal 😀). `Tests: TestInvalidUTF8Rejected, TestLoneSurrogateRejected`. | — |
| P-06, F-06 (+P-94) | A0-2.14, A0-2.8, A0-8.3, §4 | BLOCK-A0-06 ⧉. A0-2.14: "constructors MUST initialize every slice and map field of a canonicalized type to a non-nil empty value (`events.NewEvent`, `graph.NewNode`, `cjson.CanonicalValue`); `json.Marshal` emits `null` for a nil slice/map/pointer, which changes every digest. **`cjson.Canonical` and `cjson.CanonicalValue` MUST reject a `null` at any depth with `validation`.** A canonical document containing `null` is a platform defect → `internal`." Amend A0-2.8: "`null` is recognized when scanning input and rejected by `Canonical` (A0-2.14); no contract document contains it (A0-8.3)." Add rejection vector `{"a":null}`. `Tests: TestNilCollectionNeverSerializesAsNull` (reflection: a zero-valued instance of every canonicalized type marshals with no `null` token). | ⧉ PAIR-N2 → §3 A1-31, §4 A2-16 |
| P-11 | A0 §4 `cjson` sketch | BLOCK-A0-07 ⧉. Add to the sketch: `// With adds top-level fields to an already-canonical document and re-canonicalizes; it never` / `// decodes into a typed struct. A key already present in doc is an error (A1-1.8, A1-5.7).` / `func With(doc []byte, add map[string]any) ([]byte, error)`. `Tests: TestWithAddsKeys, TestWithRejectsExistingKey`. | ⧉ PAIR-C1 → §3 A1-07 |
| P-62, AM-2 (+P-72, P-70, P-34) | A0-7.1, A0-7.2, A0-7.7, A0-3.1, §4 const block | BLOCK-A0-08 ⧉. (a) A0-7.7: extend the mechanism-assignment duty from "A2 and A3" to "**A1, A2 and A3**" (every contract with capped fields). (b) A0-7.1: adopt into the single registry/const block **A1's ten local constants** (`EventMaxCanonicalBytes`, `ExitCodeMin`, `ExitCodeMax`, `IdempotencyKeyMaxBytes`, `ProseLongMaxBytes`, `LabelMaxBytes`, `EvidenceRefsMax`, …as listed in A1-4.5/4.7/§4.1) and **A2's nine** (`NodeSummaryMaxBytes`, `FindingSummaryMaxBytes`, `MaxSupersedeChain`, `AttrsMaxKeys`, `AttrKeyMaxBytes`, `AttrValueMaxBytes`, `AttrsTotalMaxBytes`, `AddressesMax`, `EvidenceIDsMax`→merged per (c)), each with its mechanism. (c) **One constant per value:** `ToolVersionMaxBytes = 64` (A2-7.1's 32 is a defect — P-62) and `EvidenceRefsMax = 8` (replacing A2's `EvidenceIDsMax` name; both documents cite the A0 name). (d) A0-3.1 `summary_too_large`: "a capped field **or count** exceeded a budget declared by its owning contract under mechanism R (A0-7.6) — the Q4 constants of A0-7.1 and the per-contract caps of A1-4.5 / A2-7.1." `Tests: TestConstantsMatchA0Table` (one assertion per registry row). | ⧉ PAIR-N1 → §3 A1-31, §4 A2-11 |
| P-66 | A0-2.17 V5 row | BLOCK-A0-09 ⧉. Annotate V5 (no byte change — see §6.1): "a canonicalization vector with a **synthetic key set**: `tool_invoked` is not an A1-3.1 kind and this is not a valid event document (six keys, not the 17-key envelope of A1-1.1). The normative event vector is A1 §4.3; the envelope field names are A1-1.1." | ⧉ PAIR-V5 → §4 A2-15 |
| F-04, P-23 (+P-91) | A0-5.3 | BLOCK-A0-10. Replace the implementation note: "parsing is three checks in order — (1) the byte-exact A0-5.1 regex `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z$` with the seconds group additionally range-checked `00`–`59` (this is what forces `Z`, exactly three digits, and rejects `:60`); (2) `time.Parse(TimeLayout, s)` for calendar validity; (3) the year window `[MinYear, MaxYear)`. `time.Parse` alone MUST NOT be used: `Z07:00` accepts numeric offsets and Go normalizes leap seconds. A parsed value MUST round-trip: `FormatTime(ParseTime(s)) == s` byte-exactly, else `validation`." `Tests: TestParseTimeRejects` (table over the six rejection classes, incl. a round-trip subtest). | — |
| AM-3, T-07 | A0-5.4 (+A0-5.6) | BLOCK-A0-11 ⧉. Record the platform-time rule in A0: "`recorded_at` is platform-stamped and MUST be non-decreasing per chain: the writer applies the forward clamp `max(clock_now, prev_recorded_at)`. **The clamp MUST be bounded to 1000 ms**: beyond that the platform MUST use the true clock reading, MUST log at error level with the engagement/run correlation attributes (ADR-0019 §3), and `recorded_at` MAY then be non-monotone — A1-8.1's rule that only `seq` orders anything is the compensating control. Any client-claimed time (`*_claimed_at`) is clamped by the same bound so the two never diverge by more than it." (T-05's `clock_anomaly` kind is deferred, §7.) | ⧉ PAIR-T1 → §3 A1-08 |
| P-18 | A0-1.1 (+A0-1.5) | "The 48-bit millisecond value is encoded big-endian in 10 characters, whose first is therefore in `0`–`7`; the 80 random bits are encoded in 16 characters with no restriction. Validation (A0-1.5) is the regex only: a body whose first character is `8`–`z` is accepted but never generated." | — |
| P-36 | A0-6.2 (+A0-8.1) | "Write-side key matching is byte-exact: a key differing from the declared `json` tag only by case is an unknown field and MUST be rejected with `validation` naming it. Because `encoding/json` matches case-insensitively, the decoder MUST verify observed key spelling (the same `Token()` walk A0-2.5 requires) rather than rely on `DisallowUnknownFields` alone." `Tests: TestAppendRejectsUnknownField` (incl. case variants), `TestCaseDuplicateKeyRejected`. | — |
| P-21 | A0-3.10 + §4 `errs` sketch | "A rate-limit response MUST carry `retry_after_ms ≥ 1` and `Retry-After ≥ 1` (`ceil(retry_after_ms/1000)`); the `omitempty` tag on `RetryAfterMS` in the §4 sketch is wrong for this field and MUST be replaced by an explicit presence rule (custom `MarshalJSON`) or by the `≥ 1` floor." `Tests: TestRetryAfterPresence, TestRetryAfterAgreement`. | — |
| P-24 | A0-8.7 | "…64 characters for SHA-256. Container image digests are the exception: `image_digest` carries the registry's `<algorithm>:<hex>` form, validated as `^[a-z0-9]+(?:[._-][a-z0-9]+)*:[0-9a-f]{64}$` and capped by the owning contract (A1-4.5: 256 B)." | — |
| P-68 | A0-3.4 | BLOCK-A0-12 ⧉. Amend A0-3.4: "…MAY echo untrusted request material (target, tool name, field name) **except a value rejected by a secret-pattern rule, which MUST NOT be echoed in whole, in part or as a digest (A2-9.5, A1-4.9)**." | ⧉ PAIR-X1 → §3 A1-27, §4 A2-07 |
| PAIR-R1 (partner; decision owned by §3 A1-05) | A0-3.11 | Insert: "an owning contract MAY declare one specific write retryable under `internal` where that contract defines a deduplication key (A1-7.6); the retry MUST reuse that key. Every other kind remains terminal." This is the A0 half of the A1-7.5/A1-7.11 fix (BLOCK-A1-01). | ⧉ PAIR-R1 → §3 A1-05 |
| P-105, D-08 | A0-1.9 | "The `COLLATE \"C\"` ordering is declared **on the column** (DDL); queries MUST NOT re-specify it (a per-query `COLLATE` silently disables index use on every paginated read, A0-4.6). `Tests: TestIDOrderingMatchesByteOrderCollateC` — integration, opt-in per DESIGN §8." | — |
| P-86, P-87, P-88, P-89, P-90, P-92, P-93, E-05, P-39, P-75, D-07 | new **§4.1 Contract tests** subsection; A0-8.2; A0-4.4 example | Add "### 4.1 Contract tests (A0)" listing, per clause, the ids this plan names: A0-1.4 `TestIDGenerationUsesCryptoRand`, `TestUniquenessViolationIsInternal`; A0-1.5 `TestValidRejectsNormalization`, `TestValidIsByteExact`; A0-2.5 `TestDuplicateKeyRejected`, `TestCaseDuplicateKeyRejected`; A0-2.15 `TestDigestEqual`, `TestGatingComparisonsAreConstantTime`; A0-3.10 `TestRetryAfterAgreement`; A0-4.6 `TestHasMoreDetection`, `TestExactFullPageHasNoNextCursor`(alias of the former — keep **one** name, `TestHasMoreDetection`); A0-4.5 `TestLimitAboveMaxRejectedNotClamped`; A0-7.4/7.5 `TestTruncateRuneBoundary`, `TestNoSilentTruncationInHashedRecords`; A0-8.6 `TestCursorRejectsStandardAlphabetAndPadding`. A0-8.2: add suffix entries `*_ids` (array of identifiers, sorted+deduped per the owning contract) and `sha256` (artifact-integrity digest, ADR-0009 §2); record `evidence_refs` as the one approved exception (digest-locked, A1-1.1). A0-4.4 prose: write the cursor as `{"id":…,"k":…}` and add "(canonical order puts `id` first, A0-2.4)". | — |
| PO-2 (§1) | A0-7.10, A0 §6 item 14 | Apply §1 PO-2 (BLOCK-PO2 + the §6 sentence). | — |
| PO-8 (§1, AM-1) | A0-1.2, A0 §4 `ids`, A0 §6 | Apply §1 PO-8: register `usr_` (row + `KindUser`), add BLOCK-PO8 to A0 §6. Update WP-facing text: A0-1.2's table now has **13** prefixes. | ⧉ PAIR-U1 → §3 A1-32, §4 A2-10 |

### 2.1 Blocks (A0) — verbatim insert text

```
BLOCK-A0-01 (A0 §4 preamble)
A0 requires six foundation packages; none imports another internal package except
`errs` (DESIGN §1 layering): `internal/ids` (A0-1), `internal/cjson` (A0-2),
`internal/errs` (A0-3), `internal/paging` (A0-4: `Page`, `Cursor`, `EncodeCursor`,
`DecodeCursor`), `internal/timex` (A0-5: `Clock`, `Now`, `FormatTime`, `ParseTime` —
the name avoids shadowing stdlib `time`, the same reasoning DESIGN §1 gives for
`logging`), `internal/caps` (A0-7: every cap constant of the A0-7.1 registry,
`TruncationMarker`, `Truncate`, `Fits`). The sketches below are illustrative and not
compiled (contracts/README.md §4); each carries its `package` line.
```

```
BLOCK-A0-02 (A0-4.4, replacing the "at worst" sentence)
A cursor is a position hint, not a grant: authorization is re-derived per page
(A0-4.4), cursors are not integrity-protected, and a forged cursor yields an empty
page, `validation`, or a page whose ordering value is the one the platform resolved —
never a silently skipped range. The platform MUST derive the seek position from the
row its lookup of the cursor's `id` resolves to and MUST ignore `k` for seeking; if
`k` disagrees with that row's ordering value the response MUST be `validation`
(A0-4.8). `DecodeCursor(s string, k ids.Kind)` takes the id kind of the collection
being paged and validates `id` against it (A0-1.5).
Tests: TestDecodeCursorRejects, TestCursorWithInconsistentKAndIDRejected,
TestHasMoreDetection.
```

```
BLOCK-PO2 (A0-7.10, appended)
Interim composition rule for the Freeze (escalated, §6.14): where two A0-7.1 caps
apply to one composed document and cannot both hold, the **smaller** governs; the
builder MUST stop at the first cap it reaches, in its declared deterministic order
(A0-4.3), and set the truncation marker (mechanism T). A3 owns the composition rule
and MAY raise it only via ADR (A0-7.2).
```

## 3. Assignment — contracts/A1-events.md (writer: A1)

Apply in row order; rows A1-03/A1-20/A1-21 change vocabularies, so do them **before** any row
that quotes an enum or a kind count.

| finding ids | clause(s) | exact change to make | paired with |
|---|---|---|---|
| P-01 | A1-3.3 (all nine tables), A1-2.1, A1-2.2 | Rename the "Composed by" column to **`actor (type, component)`** and give the literal pair in every row: `run_started` → `(user, "")`; `chain_genesis` → `(platform, "event_store")`; `graph_node_quarantined` → `(user, "")` **or** `(platform, "graph")`; `action_blocked` → `(requesting principal's type, "")`; every other platform row → `(platform, "<the subsystem named in its Source column>")`. Add under the tables: "The subsystem named in the Source column is the code path that writes the row; it appears in `actor.component` **only** when `actor.type=\"platform\"` (A1-2.1). For every non-platform row `actor.component` is `\"\"`. `actor` is inside the digest (A1-5.2)." Delete the word "varying". `Tests: TestActorComponentIsEmptyForNonPlatform`. **No §4.3 byte change** (verified: rows 0–2 already carry `(platform,"event_store")`, `(worker,"")`, `(platform,"graph")`). | — |
| P-07 | A1-7.3 (+A1-3.3 `evidence_stored`, A1-8.3) | Replace derivation-by-suffix with a **normative per-kind table**: each A1-3.3 row lists its evidence-reference fields; `evidence_refs` is the union of that list, sorted and deduplicated (A1-4.7). The table MUST include `evidence_stored` → `evidence_id` (do **not** rename the payload field; the per-kind list is the fix), `command_executed` → `output_evidence_id`, `llm_call` → `request_evidence_id`, `response_evidence_id`, `task_spawned`/`job_spawned`/`container_started` → their `*_evidence_id` fields, `approval_requested`/`approval_executed` → `action_spec_evidence_id` (A1-04), and `[]` for every kind with none. Keep the `_evidence_id` suffix rule as the *default* for kinds added later. `Tests: TestEvidenceRefsDerivation` (one subtest per kind, incl. `evidence_stored`). | — |
| T-01, T-02 | A1-3.1, A1-3.3, A1-3.5, A1-3.7, A1 §4.1 `Kind` consts | Add **three kinds** (additive, A1-3.5): `engagement_created` (user · `scope`) `client_ref:string(128)`, `roe_evidence_id:string`, `policy_evidence_id:string`, `operator_count:int` — at `seq` 1, `chain_genesis` stays the integrity anchor at `seq` 0; `engagement_closed` (user · `runtime`) `close_reason:enum{completed,cancelled,abandoned}`, `report_evidence_id:string`; `artifact_released` (user or platform · `integrity`) `artifact_kind:enum{report_html,report_pdf,findings_json,evidence_bundle,verification_bundle}`, `artifact_evidence_id:string`, `head_seq:int`, `head_hash:64hex`, `integrity_state:enum{verified,failed_overridden}`, `override_event_id:string`, `recipient_ref:string(128)`. **Then update every count in the document: the taxonomy is 42 kinds, not 39** (A1-3.1 prose, A1-3.7 coverage table rows for SPEC §5 steps 1 and 9, §4.1 comment "39 kinds", A1-4.5's maximal-payload scope, and `TestKindListIs39AndClosed` → **`TestKindListIs42AndClosed`**). A1-6.6's export path MUST compose `artifact_released` (it is the enforcement point for PO-1/S-07). | — |
| C-02, T-03 (+P-28, S-09) | A1-3.3 `approval_requested`/`approval_executed`/`spawn_requested`/`task_spawned`, A1-4.2, A1-3.6 | Reserve now (pre-Freeze, additive): on `approval_requested` **and** `approval_executed` add `action_spec_evidence_id:string` (`evi_`, MUST be non-empty — the platform stores the exact canonical bytes (A0-2) of the A7 action spec as a write-once evidence artifact at request time and records its id here and in `evidence_refs`), `target_graph_node_id:string` (`gn_`, `""` when the target is not a graph node), `argv_hash:64hex`; on `spawn_requested`/`task_spawned` add `target_graph_node_id:string`. A1-4.2 obligations: `approval_executed.action_spec_evidence_id` MUST equal the `approval_requested` value for the same `approval_id`; `fingerprint_hash` MUST be the A0-2.15 digest of exactly those artifact bytes; **`expires_at` on `approval_granted`, `approval_expired` and `approval_executed` MUST be byte-equal to the `approval_requested` value for the same `approval_id`** (divergence = platform defect → `internal`, abort the execution, record `action_blocked{reason:"approval_metadata_mismatch"}` — new enum value, additive; a timeout policy change MUST NOT affect an already-requested approval); a fingerprint that diverged at execution MUST be refused with **`conflict`** (409) naming `approval_id` and the first 8 hex chars of both digests, and MUST record `action_blocked{fingerprint_mismatch}`. A1-3.6: "A7 owns the *content* of these fields, A1 their *presence*." `Tests: TestApprovalFingerprintAndExpiryAreEqual, TestApprovalMetadataMismatchIsConflict, TestActionSpecArtifactIsChained`. | — |
| P-09 | A1-7.5, A1-7.11 | BLOCK-A1-01 ⧉. A1-7.5: "an append refused because the chain's startup verification has not completed returns **`timeout`** (504, retryable per A0-3.11); `internal` is reserved for the no-valid-genesis case." A1-7.11: "a retry MUST reuse the same `idempotency_key` (A1-7.6); a kind declared terminal by A0-3.11 MUST NOT be retried except where its owning contract declares the write retryable under `internal` with a deduplication key." | ⧉ PAIR-R1 → §2 A0 row (A0-3.11) |
| P-10, D-04 | A1-7.10, A1 §4.1 `NewEvent` | "Steps 3, 5, 7, 8, 9, 10, 12, 13 are pure and MUST run in `internal/events` (`NewEvent` / `Payload.Validate`); steps 1, 2, 6, 11, 14 need I/O or request context and run in the caller (api / policy / store layer) **in the same numbered order** around the pure part. The order is normative; the split is an implementation fact recorded here so work packages are disjoint." Fix `NewEvent`'s doc comment to name the pure subset. Add D-04's note: "(12) `untrusted` and (13) `evidence_refs` are order-independent because neither is a `*`-marked field; a future amendment that makes a derived field untrusted MUST move step 12 last." `Tests: TestValidationOrderIsDeterministic`. | — |
| P-12 | A1-4.12 (+A1 §4.1) | "A `timestamp` field is a Go **`string`** holding an A0-5.1 value, in the envelope and in every payload. `time.Time` MUST NOT appear in any canonicalized type: its `MarshalJSON` drops trailing zero fractional digits and silently changes digests. Transcription rule for `:int` — `int64` for every `*_ms`, `*_bytes`, count and `seq` field; `int` only where A1-4.2 declares a range (`exit_code`)." `Tests: TestNoTimeTimeInCanonicalizedTypes`. | — |
| P-13, P-32 | A1-5.4, A1-5.6 | BLOCK-A1-02 ⧉. Add to `ChainHead` (A1-5.6, mutable bookkeeping, never hashed, updated inside the append transaction): `LastRecordedAt string` and `LastBreakSeq int64`, `LastBreakKind BreakKind`, `LastBreakEventID string`. A1-5.4: "the forward clamp compares A0-5.1 strings **byte-wise** — for a fixed-format UTC millisecond timestamp, byte order equals time order — and is bounded to 1000 ms by A0-5.4." A1-6.3: "the break-dedup state is `ChainHead.LastBreakSeq`/`LastBreakKind`, written in the same transaction as the break event." `Tests: TestRecordedAtClampIsMonotone, TestOneBreakEventPerRun`. | ⧉ PAIR-T1 → §2 A0-11 |
| S-02, P-37 | A1-5.8, A1-6.3, A1-6.2, A1 §4.1 | Add `head_regression` to `break_kind`: "the stored head `(head_seq, head_hash)` is lower or different from the highest head this platform recorded out-of-band for that engagement." A1-5.8: the head hash MUST be emitted at every verification, at every append crossing a **100**-`seq` boundary, and at every `run_ended`/`hard_stop_fired`, and MUST be written to a store the event-store role cannot UPDATE or DELETE — a separate `chain_head_trail` table with `REVOKE UPDATE, DELETE` (A1-7.9's rule). Startup verification MUST compare against that trail and report `head_regression` as a break with A1-6.4's consequences. Add `HeadLogIntervalSeq = 100 // A1-5.8` to §4.1 (supersedes 1000 — §5 conflict 10). If PO-3 declines the webhook anchor, print the residual into A1-6.6. `Tests: TestHeadRegressionDetected, TestHeadTrailIsAppendOnly`. | §1 PO-3 |
| S-08, P-31 | A1-6.2, A1-6.4 | "Startup verification runs asynchronously, one goroutine per engagement chain, owned and cancellable per DESIGN §6; it MUST NOT block startup of the platform process. While a walk is incomplete `integrity_state` is `unverified`: a read served before the walk completes MUST carry `integrity_state:"unverified"` on the same carrier A4 uses for `failed` (A1-6.4); appends return `timeout` (A1-05); **no customer-facing artifact MUST be produced from an `unverified` chain** (`integrity_failed` 409, A0-3.1); other engagements are unaffected." `Tests: TestReadDuringStartupWalkIsFlaggedUnverified, TestExportFromUnverifiedChainRefused`. | — |
| P-14 | A1-6.6 | "`head_seq`/`head_hash` are the values of the **last row the walk verified** — identical to the `head_seq`/`head_hash` payload of the event named by `chain_verified_event_id` — not the head after that event was appended; `head_seq` therefore equals the post-append `ChainHead.HeadSeq` minus one." Fix the §4.2 export example so the two agree. | — |
| P-15 | A1-6.6 (+A1-5.6) | "The export `integrity_state` is a **distinct** enum (`verified \| failed_overridden`) derived from A1-5.6's (`unverified \| verified \| failed \| overridden`): `unverified`/`failed` → not exportable (A1-6.4); `verified` → `verified`; `overridden` → `failed_overridden`. The two lists MUST NOT be used interchangeably." Give the mapping as a 4-row table. | — |
| P-16 | A1-5.6 (+A1-8.9) | "…the chain-head row MUST NOT appear in an event document (A1-1.6) and is never hashed. The **row** is internal; the **projection** `(head_seq, head_hash, integrity_state, verified_at)` is served to authorized readers and to every export, with a wire shape A4 owns (A1-8.9, A1-6.4)." Delete the unqualified "never served". | — |
| P-17 | A1-8.1 | Delete "a cursor is only valid in the direction it was issued for" (unenforceable: A0-4.4's key set is closed at `{id, k}` and carries no direction) and replace with: "direction is a request parameter; a cursor issued for the other direction is interpreted in the requested direction and yields a well-defined (possibly empty) page. The cursor shape is **not** extended (§5 conflict 9): its integrity problem is solved by A0-4.4's resolved-row seek rule, not by a MAC." | ⧉ PAIR-K1 (context) |
| PAIR-K1 (partner; decision owned by §2 A0-02), P-71, AM-4 (A1 side) | A1-8.2, A1-8.6 | BLOCK-A1-03 ⧉ (partner text; A0 owns the decision). A1-8.2: "the platform MUST resolve the cursor's `id` **in the requested engagement** and MUST derive the seek position from that row's `seq`, ignoring `k` for seeking; a cursor whose `k` disagrees with the resolved row's `seq`, or whose `id` does not resolve in this collection, is `validation` (400) telling the client to restart from the first page (A0-4.4, A0-4.8)." **One oracle:** A1-8.2 and A2-11.4 both say `validation` (P-71 resolved). `Tests: TestCursorFromEngagementARejectedInB, TestCursorWithInconsistentKAndIDRejected`. | ⧉ PAIR-K1 → §2 A0-02, §4 A2-23 |
| S-03 | A1-4.2 `approval_executed`, A1-7.6 | "The platform MUST consume the approval and authorize the execution **in one transaction** (or under the same per-engagement lock as A1-5.4's append), guarded by a uniqueness constraint on `(engagement_id, approval_id)` in the consumption table. A second attempt MUST fail **before** any container is created, with `conflict` (A0-3.1) and `action_blocked{reason:"approval_consumed"}` (new enum value, additive), and MUST NOT compose a second `approval_executed`." `Tests: TestSingleUseApprovalRaceConsumesOnce` (N concurrent executions → exactly one `approval_executed`, N−1 `action_blocked{approval_consumed}`), `TestConsumedApprovalCannotSpawnAgain`. | — |
| S-06, P-25 | A1-4.2 `approval_requested`, A1-4.4, A1 §4.1 | "`untrusted_context` is **platform-computed**: it MUST be `true` iff this payload has any non-empty `*`-marked field, or iff the event referenced by `request_event_id`/`spawn_request_event_id` (transitively) carries `untrusted:true`. It MUST NOT be a caller value. `action_summary` MUST be composed only from platform vocabulary and the A7 action spec's own fields; copying model or tool prose into it is a laundering violation of A1-4.4." Add to A1 §4.1: `// UntrustedFields returns k's A1-4.4 "*" field names in canonical order. Normative:` / `// transcribed from A1-3.3, part of the digest definition (A1-5.2).` / `func UntrustedFields(k Kind) []string`, plus a kind → `*`-fields table in A1-4.4. `Tests: TestUntrustedContextComputedNotSupplied, TestNoUntrustedTextInUnmarkedFields` (per-kind source→target allowlist table), `TestApprovalViewFlagsUntrustedContext, TestUntrustedFlagMatchesStarredFields`. | — |
| C-03, P-27 | A1-4.2 `container_killed`, A1-7.12 | Replace "MUST be preceded in the chain by a `hard_stop_fired` event" with: "`container_killed{kill_reason:\"hard_stop\"}` MUST carry `stop_event_id:string` (`evt_`, the `hard_stop_fired` event it answers). The platform MUST commit `hard_stop_fired` **before** it issues any kill, in its own transaction, and MUST NOT block the kill on that commit (A1-7.12): if the commit fails the platform MUST kill anyway and MUST retry the append until it lands. The obligation is correlation by `stop_event_id`, not `seq` order; a `container_killed{hard_stop}` whose `stop_event_id` is empty or unresolvable is a platform defect → `internal` (A0-3.1) and MUST NOT delay the kill. The validator reads the run's hard-stop state held by `internal/policy`, never the chain (A1-7.12)." | — |
| C-04 | A1-3.3 `scope_changed`, A1-4.2, A1 §6.9 | "A global-blacklist change MUST also be composed as `scope_changed` into **every affected engagement chain**, with `change_kind:blacklist_added`/`blacklist_removed` (additive enum values) and `entry`/`entry_hash` of the global entry; the composing subsystem is `scope`, `actor` is the admin user (`user`, `""`, AM-1 `usr_` id), `run_id` is `\"\"`. Each engagement's `quarantine_recomputed{blacklist_changed}` MUST reference **that engagement's own** `scope_changed` event as `trigger_event_id` (A1-4.2 stays satisfiable). A1 §6.9's deferral MUST NOT cover blacklist mutation: it is engagement-reachable safety state, not session audit." `Tests: TestGlobalBlacklistChangeIsChainedPerEngagement, TestQuarantineRecomputedTriggerResolvesInEngagement`. | §1 PO-4 |
| P-63, C-06, T-06 | A1-3.3 `graph_node_quarantined`, A1-3.6, A1-7.4 | BLOCK-A1-04 ⧉. **Delete `operator_release`** from the `quarantine_kind` enum (pre-Freeze; additive-only afterwards) and keep `operator_quarantine`. A1-3.6 gains the ownership + mapping rule: "A2 owns the quarantine **state** vocabulary (`quarantine_reason`); A1 owns the **occurrence** vocabulary (`quarantine_kind`). Mapping: `out_of_scope_discovery → out_of_scope` · `blacklist_match → blacklisted` · `operator_quarantine → (the reason already in force)`. The stored reason MUST be derived by the platform from this mapping, never copied from an event string. A release is not an A1 enum value: it is `quarantine_recomputed{scope_changed}` plus the per-node recomputation, stored as `quarantined:false` with `quarantine_reason` absent; the A1 event is the only record of the previous state." A1-7.4: `graph_node_quarantined` stays user-composable for `operator_quarantine` only. | ⧉ PAIR-Q1 → §4 A2-12, §1 PO-9 |
| PAIR-Q2 (partner; decision owned by §4 A2-08) | A1-3.3 `action_blocked`, `quarantine_recomputed` | BLOCK-A1-05 ⧉. Add to `action_blocked.reason` (additive): `target_quarantined` — "a request cited a `gn_` id whose node is quarantined; refused **before** approval routing (A2-8.3)". Extend `quarantine_recomputed.trigger` (additive) with `node_written` and `edge_written`, so the full list is `scope_changed`, `blacklist_changed`, `node_written`, `edge_written`. | ⧉ PAIR-Q2 → §4 A2-08 |
| P-64, T-04 | A1-3.3 `graph_node_written`/`graph_edge_written`, A1-4.2 | BLOCK-A1-06 ⧉. Add to `graph_node_written`: `content_hash:64hex` and `dedup_hit:bool`; add to `graph_edge_written`: `dedup_hit:bool`. A1-4.2: "`content_hash` MUST equal the A2-4.6 fingerprint of the written node (edges have no fingerprint — A2-4.6 is node-only); `dedup_hit` is `true` when the write collapsed into an existing row under A2-4.7, and a collapse MUST still emit the event." **No §4.3 change** (row 2 is `graph_node_quarantined`). | ⧉ PAIR-G1 → §4 A2-04 |
| P-08, P-33 | A1-7.6 | "The A1-7.6 uniqueness constraint applies to rows appended through `events:append` only. A platform-composed row MUST carry a non-empty deterministic key: `approval_*` → `approval_id` + decision · `evidence_stored` → `evidence_id` · `job_spawned`/`task_spawned`/`container_started` → `spawn_request_event_id` · `graph_*` → the written `gn_`/`ge_` id · `chain_verified` → `trigger` + `head_seq` · `chain_break_detected` → `break_seq` + `break_kind` · `cleanup_*` → `revert_event_id` · `notification_sent` → `related_event_id` + `attempt` · `llm_call` → the gateway's per-call id · otherwise the composing subsystem's operation id. `\"\"` MUST NOT be used." And: "`PayloadHash = cjson.SHA256Hex(cjson.CanonicalValue(struct{Kind Kind \\`json:\"kind\"\\`; Payload Payload \\`json:\"payload\"\\`}{…}))`, computed **after** the A1-7.10 normalization steps, so a retry whose array order differs is a dedup hit. No other field participates." Publish the vector from §6.3. `Tests: TestPlatformDedupKeyIsNonEmptyPerKind, TestIdempotencyKeyReuseWithDifferentPayloadIsConflict, TestDedupHitWritesNoEvent`. | — |
| P-80 | A1 §4.1 | Add: `// UnmarshalEvent decodes a served event document in two passes, both with` / `// DisallowUnknownFields: the envelope keys give kind, then payload decodes into` / `// that kind's concrete type (A1-4.1). No map[string]any intermediate (A0-2.5).` / `func UnmarshalEvent(b []byte) (Event, error)`. `Tests: TestEventRoundTrip` (the README merge-gate item, currently unwritable). | — |
| P-97, E-03, S-13 | A1-7.12, A1-6.4, A1-6.5 | A1-7.12: "the pending kill record MUST be written to a durable **outbox** (append-only table, `REVOKE UPDATE, DELETE`, A1-7.9) **before** the kill is issued, and startup MUST drain the outbox into the chain; a kill whose event cannot be composed MUST surface in the UI as an unresolved integrity warning on A1-6.4's carrier, not only in `slog`." Add the E-03 test names to the clauses they guard: A1-6.4 `Tests: TestExportBlockedOnFailedChain` (each of the four artifact classes → `integrity_failed` 409 naming `break_seq`/`break_kind`), `TestInternalViewOverrideDoesNotAuthorizeExport`; A1-6.5 `Tests: TestOverrideDiesAtNextBreak, TestSingleUseOverride`; A1-7.12 `Tests: TestKillPathDoesNotBlockOnEventStore, TestHardStopProceedsWhenEventStoreUnavailable, TestKillPathDoesNotBlockOnAppendLatency`. | — |
| E-01 | A1-2.3, A1-2.6, A1-2.7, A1-3.4, A1-7.4 | Add "Tests:" lines naming, verbatim: `TestWorkerCannotAppendNonCKind` (each non-C kind → `forbidden` + `action_blocked{append_not_permitted}`), `TestClientCannotSupplyEnvelopeFields` (table over all 12 A1-2.3 fields → `validation` naming the field), `TestActorCannotBeForged`, `TestOrchestratorCannotClaimCommandExecuted`, `TestMachinePrincipalCannotReachIntegrityKinds`, `TestUntrustedFlagCannotBeSupplied`. State in A1-7.4: "these six are the negative half of AGENTS.md's safety-test rule for the write path; A1 does not reach `Frozen` without them." | — |
| E-04 | A1-5.3, A1-6.2, A1-7.11 | Add "Tests:" lines naming, verbatim: `TestSecondGenesisIsRejected`, `TestAppendRefusedWithoutValidGenesis`, `TestAppendRefusedBeforeStartupVerificationCompletes`, `TestFailedAppendConsumesNoSeq` (rollback leaves no gap, A1-5.4), `TestNoGlobalSequenceSharedAcrossEngagements`. | — |
| P-69, C-07 (+PAIR-SEC1 partner text; rule-table decision owned by §4 A2-07) | A1-4.9 | BLOCK-A1-07 ⧉ (A2 owns the rule table). A1-4.9: **delete its own pattern enumeration** and cite A2-9.4's rule ids as the only source ("the platform MUST run the `internal/secretscan` rules of A2-9.4; the rule ids live in exactly one document"). Then insert the identical filter sentence: "The scan is a **filter, not a guarantee**: a hostile worker can encode or split a secret past any pattern set. Egress exclusion (ADR-0020 §4) is the enforcement point and MUST be applied independently at the gateway to every string that leaves the platform; for any engagement whose policy is not `local_only` the gateway MUST exclude by **kind** — no `credential` node, no node with `credential_kind` set, no `attrs` of such a node — and `llm_call` MUST record the exclusion in `excluded_secret_count`." Delete the "because secrets never enter an event…" deduction. `Tests: TestEventSecretFreeSerialization, TestSecretScanNamesFieldNotValue, TestNoSecretValueOrDigestInError`. | ⧉ PAIR-SEC1 → §4 A2-07; PAIR-X1 → §2 A0-12 |
| P-83, P-95, P-96, P-98 (+ the A1 test names listed under §2 row A0-17) | new **§4.4 Contract tests (A1)** subsection | List per clause: A1-4.10 `TestServedEventIgnoresUnknownFields`, `TestAppendRejectsUnknownField`; A1-4.6 `TestRedactedIsPlatformSetOnly`, `TestRedactedNotUsedForTruncation`; A1-4.7 `TestArraysSortedDedupedAtComposition`, `TestCapsRejectWithSummaryTooLarge`; A1-6.8 `TestVerificationIdempotent`, `TestVerificationWritesNoOtherKind`; A1-7.6 `TestIdempotencyKeyReuseWithDifferentPayloadIsConflict`, `TestRetryCannotShiftClaimedTime`; A1-8.4 `TestWorkerAndNodeHaveNoReadScope`, `TestSSENotReachableByMachinePrincipal`; A1-8.5 `TestSSEEmitAfterCommitInSeqOrder`, `TestSSEReplayDedupesByEventID`; A1-8.6 the six cross-engagement negatives; A0-2.15 `TestGatingComparisonsAreConstantTime`. | — |
| P-26, P-30, P-38 | A1-4.5, A1-4.2, A1-4.7, A1-4.12 | A1-4.5: "a kind's **maximal payload** sets every string field to exactly its cap length in bytes of U+0001 (worst case `\u0001` = 6 canonical bytes per input byte, A0-2.7), every integer to its declared maximum, every array to its count cap filled with maximum-length ids, every bool to `true`. `TestMaximalPayloadFitsCanonicalBound` asserts `len(Preimage(e)) ≤ EventMaxCanonicalBytes` for all **42** kinds under that construction." A1-4.2: "every `error_kind` field (`task_result`, `agent_error`, `llm_call`) MUST be `\"\"` or a byte-exact A0-3.1 kind; the per-kind non-empty obligations are additional; an unknown value is `validation` (A0-6.3)." A1-4.7: drop "or of integers"; A1-4.12: "Five payload fields across three groups…". | — |
| P-29, P-74, S-10 | A1-4.11, A1-4.2 (`graph_*_written`), A1-8.5 | A1-4.11: "a client MUST upload its artifacts (`evidence:upload`) before appending the event that references them; an append whose `*_evidence_id` does not resolve in this engagement is `notfound` (404) and MUST be retried by the buffering client after the upload (ADR-0013). Write-time resolution is what makes A1-8.8's dangling reference an archival state, never a normal one." A1-4.2 `graph_node_written`/`graph_edge_written`: "`node_kind` MUST be an A2-2.1 kind and `edge_kind` an A2-3.1 kind; because only the platform composes these kinds after a successful graph write, a violation is a platform defect → `internal`." A1-8.5: "the resume position MUST be validated exactly like a cursor (A1-8.2): resolve the requested `seq` **in the stream's own engagement**, resync from that engagement's earliest retained row when it cannot, and send an explicit resync control frame rather than silently starting elsewhere. A `Last-Event-ID` that is not a decimal integer in `[0, head_seq]` MUST be ignored (restart from head) and MUST NOT be echoed. `Tests: TestSSEResumeFromForeignSeqNeverServesUnrelatedPosition`." | — |
| PAIR-A1, PAIR-A2, PAIR-M1 (partner; tables owned by §4 A2-10 and §4 A2-14) | A1-2.1, A1-3.3 `graph_node_written`, A1-3.6 | BLOCK-A1-08 ⧉ (partner; A2 owns the tables). A1-2.1 gains the identity mapping table A1 `actor.type` ↔ A2 `principal_kind` (verbatim from §4 A2-10). A1-3.3's `graph_node_written` row gains: "`actor` is `(platform, \"graph\")`; the originating worker/orchestrator is carried by the node's `provenance.principal_kind` (A2-5.3). The two are expected to differ — copying one into the other is a defect." A1-3.6 gains the A1↔A2 field mapping table (verbatim from §4 A2-14) and cites A2-1.6. | ⧉ PAIR-A1, PAIR-A2, PAIR-M1 → §4 A2-10, A2-14 |
| D-03 + PAIR-N2/N1/V1/C1 (partner; those decisions owned by §2 A0-03, A0-06, A0-07, A0-08) | A1-1.2, A1-4.8, A1-4.5/4.7, A1-5.7, **A1 §4.3** | BLOCK-A1-09 ⧉. (a) A1-1.2/A1-4.8: "every slice and map field of a canonicalized type MUST be non-nil before marshaling; the constructors initialize them to empty (A0-2.14). A canonical event document containing `null` is a platform defect → `internal`." (b) A1-4.5/A1-4.7: replace A1's local constant definitions with citations of the A0-7.1 registry names (`ToolVersionMaxBytes = 64`, `EvidenceRefsMax = 8`, …) — do not restate values. (c) A1-5.7: "`Served(preimage, c)` = `cjson.With(preimage, map[string]any{\"seq\": c.Seq, \"prev_hash\": c.PrevHash, \"hash\": c.Hash})`, operating on the generic document only; it MUST decode with `UseNumber()` and re-emit every number's literal text verbatim, and MUST reject a preimage that fails A0-2 (`preimage_mismatch`, A1-6.3). U+2028/U+2029 are literal in the served bytes (A0-2.7)." `Tests: TestServedBytesReproducible, TestServedRoundTripIsIdentity`. (d) **A1 §4.3 (D-03):** add **row 3** to the normative chain vector with the exact bytes/lengths/digests published in §6.1 (a `task_result` carrying a 16-digit `duration_ms`, a literal U+2028, U+2029, U+007F and a non-BMP 😀 inside `result_summary`); rows 0–2 are unchanged (verified). Recompute with the §8 script, self-check rows 0–2 first, and show both in the progress report. | ⧉ PAIR-N2/N1/V1/C1 → §2 A0-06, A0-08, A0-03, A0-07 |
| PO-1, PO-3, PO-4, PO-6, PO-8 (§1) | A1-6.5, A1-5.8, A1 §6.9, A1 §6.4, A1-2.2 | Apply the §1 rows: PO-1 (BLOCK-PO1-A/B in A1-6.5 + §6 item 15), PO-3 (§6 item 16), PO-4 (BLOCK-PO4 in §6.9 + item 17), PO-6 (BLOCK-PO6 in §6.4), PO-8 (BLOCK-PO8 citation in A1-2.2 + §6.2). | ⧉ PO-6/PO-8 pairs |

### 3.1 Blocks (A1) — verbatim insert text

```
BLOCK-A1-01 (A1-7.5 row "chain not verified at startup")
| chain not verified at startup (A1-6.2 walk in flight) | `timeout` (504, retryable
per A0-3.11 and A1-7.11) | none — nothing was appended |
`internal` (500) is reserved for the no-valid-genesis case (A1-5.3), which is a
platform defect (A0-3.1) and fail-closed (A1-7.11).
```

```
BLOCK-PO1-A (A1-6.5, replacing the two contradictory sentences)
Only a `user` principal holding the **admin** role may compose `integrity_override`
(SPEC §3 places integrity-class controls next to the hard stop; Q11's "operator"
reads as "human", and this clause narrows it — §6 item 15). An operator-scoped user —
including one assigned to the engagement — MUST NOT override any break. A5 MUST gate
the endpoint on the admin role and MUST return `forbidden` (A0-3.1, principal-level,
naming no object) to an operator.
Tests: TestOverrideIsAdminOnly, TestOperatorCannotOverrideAnyBreak.
```

```
BLOCK-PO1-B (A1-6.5, new paragraph)
An `integrity_override` with `scope:"export"` is **single-use** (mirroring Q10): it
authorizes exactly one artifact release. The platform MUST bind the release to it
atomically — a uniqueness constraint on `(engagement_id, override_event_id)` in the
release record — and MUST compose an `artifact_released` event naming
`override_event_id`, `head_seq`, `head_hash` and the artifact's `evi_` id. A second
release requires a second human decision and MUST fail with `conflict` +
`integrity_failed` (A0-3.1). An override with any other `scope` still dies at the
next `chain_verified` or `chain_break_detected`.
Tests: TestSingleUseOverride, TestOverrideDiesAtNextBreak,
TestArtifactReleasedNamesItsOverride.
```

```
BLOCK-PO4 (A1 §6.9, appended)
Ownership split applied for the Freeze: authentication, session, credential and role
audit are **not** A1 kinds — they have no natural `engagement_id` (A1-5.1) and belong
to A5's platform-scoped audit store. Two exceptions are engagement-scoped and are
chained here: a global-blacklist mutation is composed as `scope_changed` into every
affected engagement chain (A1-3.3, adversarial C-04), and engagement operator
assignment is a known gap that A5 MUST close by adding
`engagement_assignment_changed` additively (A1-3.5) — see §6 item 17.
```

```
BLOCK-A1-04 (A1-3.6, new bullet — ⧉ identical decision content to A2 BLOCK-A2-12)
Quarantine vocabulary: A2 owns the **state** vocabulary (`quarantine_reason`:
`out_of_scope`, `blacklisted`); A1 owns the **occurrence** vocabulary
(`quarantine_kind`: `out_of_scope_discovery`, `blacklist_match`,
`operator_quarantine`). Mapping — `out_of_scope_discovery → out_of_scope` ·
`blacklist_match → blacklisted` · `operator_quarantine → (the reason already in
force)`. There is no `operator_release` value: a release is
`quarantine_recomputed{scope_changed}` plus the per-node recomputation, stored as
`quarantined:false` with `quarantine_reason` absent. The stored reason MUST be
derived by the platform from this mapping, never copied from an event string.
```

```
BLOCK-A1-05 (A1-3.3 — ⧉ identical decision content to A2 BLOCK-A2-08)
`action_blocked.reason` gains `target_quarantined` (additive): a spawn or action
request cited a `gn_` id whose node is quarantined; refused **before** approval
routing (A2-8.3, ADR-0018 §2). `quarantine_recomputed.trigger` is the closed list
`scope_changed · blacklist_changed · node_written · edge_written` — the last two are
the recomputations a new node or a new edge touching a quarantined node causes
(A2-8.2).
```

```
BLOCK-A1-06 (A1-3.3 graph rows — ⧉ identical decision content to A2 BLOCK-A2-04)
`graph_node_written` carries `content_hash:64hex` (the A2-4.6 fingerprint of the
written node) and `dedup_hit:bool`; `graph_edge_written` carries `dedup_hit:bool`
(edges have no fingerprint). A dedup collapse (A2-4.7) MUST still emit the event with
`dedup_hit:true`, so the chain distinguishes "new evidence recorded" from "duplicate
absorbed" — including every offline-node replay (ADR-0013).
```

```
BLOCK-A1-07 (A1-4.9 — ⧉ identical decision content to A2 BLOCK-A2-07)
The pattern set is A2-9.4's rule table and nothing else: rule ids live in exactly
one document, and an error MUST name the field and the rule id, never the value, a
prefix of it, or a digest of it (A0-3.4, A2-9.5). The scan is a **filter, not a
guarantee** — a hostile worker can encode, split or re-format a secret past any
pattern set. Egress exclusion (ADR-0020 §4) is the enforcement point and MUST be
applied independently at the gateway to every string that leaves the platform; for an
engagement whose policy is not `local_only` the gateway MUST exclude by **kind** (no
`credential` node, no node with `credential_kind` set, no `attrs` of such a node) and
`llm_call` MUST record the exclusion in `excluded_secret_count`.
```

```
BLOCK-A1-08 (A1-2.1 — ⧉ byte-identical to A2 BLOCK-A2-10's mapping table)
| A1 `actor.type` | A2 `principal_kind` | id shape (A0-1.2) |
|---|---|---|
| `platform` | `platform` | `""` |
| `orchestrator` | `orchestrator` | `job_` |
| `worker` | `worker` | `task_` |
| `node` | `node` | `slp_node_` |
| `user` | `user` | `usr_` (AM-1) |
One vocabulary, two documents: A2-5.3 adopts A1-2.1's list verbatim; `operator` is
renamed `user` and `operator_id` is renamed `user_id` platform-wide. The prose word
"operator" (a human role, SPEC §3) is unaffected — only the enum value changes.
```

## 4. Assignment — contracts/A2-graph.md (writer: A2)

Apply A2-01 (rename), A2-14 (rename) and A2-17 (`attrs` wire shape) **first**: §4.1's examples and
the new §4.2 vectors depend on them.

| finding ids | clause(s) | exact change to make | paired with |
|---|---|---|---|
| P-40 | A2-1.4, A2-8.9, A2-5.4, A2-5.5, §4, §4.1 examples | New clause **A2-1.4a**: "`seq` is assigned by the platform inside the transaction that inserts the row, from **one per-engagement graph sequence shared by nodes and edges**, strictly increasing by 1, dense, and independent of the A1 event `seq` (A1-5.4). To keep the two apart in code, logs and views the graph field is named **`graph_seq`** in Go fields, store columns **and JSON**; A2-8.9's ordering tuple is `(graph_seq, graph_node_id)`/`(graph_seq, graph_edge_id)` and A0-4.4's cursor `k` is that value." Rename every occurrence (clauses, §4 sketch `Seq int64 \`json:"graph_seq"\``, §4.1 examples, A2-11 tests). **`content_hash` is unaffected** (A2-4.6's 20 keys contain no `seq`). | — |
| P-65 | A2-1.6, A2-1.2, §4, §4.1 examples | BLOCK-A2-14 ⧉. Rename A2's served `id` to **`graph_node_id`** (nodes) and **`graph_edge_id`** (edges) so one value has one name platform-wide (A0-3.6 already reserves those spellings), and publish the mapping table: `graph_node_written.graph_node_id → Node.ID` · `graph_edge_written.graph_edge_id → Edge.ID` · `from_graph_node_id → Edge.SourceID (source_id)` · `to_graph_node_id → Edge.TargetID (target_id)` · `supersedes_graph_node_id → Node.SupersedesID (supersedes_id)`. Update §4.1's five `"id":` keys and every clause that cites the served name. **`content_hash` is unaffected** (no `id` key in A2-4.6). | ⧉ PAIR-M1 → §3 A1-32 |
| F-07, P-51 | A2-6.1, A2-6.2, §4 `Attrs`/`AttrValue` | "The wire and canonical form of `attrs` is `{\"<key>\": <JSON string \| integer \| boolean>}` — depth exactly one, no wrapper object. `AttrValue` MUST carry `json:\"-\"` on its Go fields and a hand-written `MarshalJSON`/`UnmarshalJSON` emitting the bare scalar of the live field; the `Type` discriminator exists only in Go and is never serialized. `UnmarshalJSON` accepts a JSON string, integer or boolean only and rejects `null`, floats, arrays and objects with `validation` (A2-6.1). Round-trip MUST preserve `Type`." Add the sketch lines. `Tests: TestAttrValueRoundTrip, TestAttrsRejectNestedFloatNull`. | — |
| P-41 | A2-4.6 | "`addresses` and `evidence_ids` MUST be sorted ascending by unsigned byte value and deduplicated **before** the content document is canonicalized (A1-4.7's rule), so two observations of the same content in a different input order produce the same `content_hash`. `attrs` keys are ordered by A0-2.4." `Tests: TestContentHashStableAcrossArrayOrder`. | §4 A2-03 (vector proves it) |
| D-01, P-82 | new **§4.2 Normative content fingerprint vector** | Insert the four rows published verbatim in §6.4 (F1 finding, F1-R reordered-input, F3 attrs-differ, S1 service) with exact canonical bytes, lengths and SHA-256, declared normative like A0-2.17 and A1 §4.3: "the shared contract-test suite MUST reproduce these bytes and digests byte-exactly; a `content_hash` computed over unsorted arrays (F1-R's forbidden digest) is a defect." `Tests: TestContentHashVector, TestContentDocFixedKeySet`. | — |
| P-42, S-05 (+P-67) | A2-4.7, A2-5.6, A2-5.3, §4 `Provenance` | BLOCK-A2-04 ⧉. `provenance` becomes a **bounded list (≤ 8 entries)**, each with its own `event_id`, `principal_kind`, `run_id`, `job_id`, `task_id`, `agent_node_id`, `tool_id`, `tool_version`, `recorded_at`, `observed_claimed_at`, `confidence`; list order MUST be by the `seq` of `provenance[].event_id`, never by arrival, so the node's bytes are deterministic. A dedup collapse MUST NOT discard the observation: the platform appends the new provenance entry, and **a provenance entry whose `event_id` is already present is not appended** (replay idempotence, ADR-0013 — this guard is required, it is not in either report). `confidence` MUST be raised to `verified` **only** when the new entry's `event_id` differs from every existing one **and** its `task_id`/`agent_node_id` differ (independence, A2-5.6); that is the only path by which `verified` is stored, so Q2's grade is now reachable. The collapse MUST emit A1 `graph_node_written{dedup_hit:true}` (BLOCK-A1-06). `content_hash` remains computed over content only (A2-4.6), so the collapse key `(engagement_id, kind, content_hash)` is unchanged. A2-4.7 also gains: "content dedup is the **only** protection against a rewound ingest watermark (A1-7.7 item 4); `content_hash` MUST therefore be stable across platform releases (A2-4.8, A0-2.16)." `Tests: TestNodeDedupKeepsEveryObservation, TestVerifiedRequiresIndependentObservation, TestReplayedEventAddsNoProvenanceEntry, TestProvenanceListIsOrderedByEventSeq`. | ⧉ PAIR-G1 → §3 A1-21 |
| P-43, E-07 | A2-10.6 | Keep `Node`/`Edge` fields **exported** with the sketched `json` tags; delete the "fields are unexported" claim and delete the "`go vet`-visible exported-field audit" sentence (no such `go vet` check exists; `encoding/json` cannot marshal unexported fields). Replace the enforcement with: "(a) `NewNode`/`NewEdge` are the only documented construction path; (b) the store seam re-validates every value it is given (A2-10.7); (c) `TestNodeHasNoExportedContentSetter` — reflection over `graph.Node`/`graph.Edge` asserts no exported method mutates a content field (the four mutation methods of A2-54 and the flags of A2-1.3/A2-3.9 are the declared exceptions); (d) review per AGENTS.md. If the product owner prefers unexported fields, A2 MUST specify `MarshalJSON`/`UnmarshalJSON` for `Node` and `Edge`." | — |
| P-44, P-45, P-54 | §4 `NodeDraft`/`PendingNode`/`NewNode`/`WriteNode`, A2-10.6 | BLOCK-A2-06. Spell out `NodeDraft` field-by-field (the report's listing, verbatim) and state: "`ID`, `EngagementID`, `GraphSeq`, `ContentHash`, `Quarantined`, `QuarantineReason`, `ReportExcluded`, `SupersededByID` and `Provenance` are platform-set and absent from the draft; a request body carrying one is an unknown field on a write (A0-6.2, A2-1.5, A2-5.2)." `NewNode(in NodeDraft, prov Provenance, q QuarantineState) (PendingNode, error)` returns validated content + provenance + quarantine state + `content_hash`, **no id, no seq** (DESIGN §4: never hand out a half-built value). The seam is `WriteNode(ctx context.Context, engagementID string, n PendingNode) (Node, error)` — assigns `graph_node_id` (A0-1.4) and `graph_seq` (A2-1.4) inside the insert transaction, applies A2-3.4/A2-4.7 dedup, returns the complete `Node`; same shape for `NewEdge`/`WriteEdge`. A2-10.6 applies to `PendingNode` and `Node` alike. A2-10.6/A2-11.1 also gain: "the graph store seam exposes exactly four mutation methods, each taking an engagement id and each returning `notfound`/`conflict` per A2-10.3 — `SetSupersededBy`, `SetQuarantine`, `SetReportExcluded`, `SetEdgeRetracted`. No other update or delete method MAY exist (A1-7.2's rule applied to the graph seam)." `Tests: TestGraphSeamHasExactlyFourMutationMethods, TestNodeDraftRejectsPlatformFields`. | — |
| P-46 (PAIR-SEC1 owner; the A1-side ids are owned by §3 A1-27) | A2-9.4, A2-9.5, A2-9.7 | BLOCK-A2-07 ⧉. A2-9.4 MUST publish a **normative, closed, additive-only (A0-6.5) rule table** `rule id \| Go regexp \| fields scanned \| note` with at least: `SEC-PEM` (`-----BEGIN [A-Z ]*PRIVATE KEY-----`) · `SEC-NTLM` (32-hex NTLM/LM shapes) · `SEC-KRB` (`krbtgt` ticket material) · `SEC-AWSKEY` (`(AKIA\|ASIA)[0-9A-Z]{16}`) · `SEC-GCPKEY` · `SEC-AZUREKEY` · `SEC-JWT` (`eyJ[0-9A-Za-z_-]+\.[0-9A-Za-z_-]+\.[0-9A-Za-z_-]+`) · `SEC-BEARER` (`(?i)(bearer\|token\|api[_-]?key\|password\|passwd\|secret)\s*[:=]\s*\S{8,}`) · `SEC-URLCRED` (`[a-z][a-z0-9+.-]*://[^/\s:@]{1,64}:[^/\s:@]{1,64}@`) · `SEC-ENTROPY` (Shannon entropy ≥ 4.5 bits/char over a window of ≥ 32 characters drawn from a base64/hex alphabet; the formula and the window are part of the rule). Plus: "the corpus planted by `TestEventSecretFreeSerialization` and `TestGraphSecretFreeSerialization` is exactly one value per rule id and lives in the shared suite; a rule added later adds a corpus entry." Also insert BLOCK-A1-07's filter-not-guarantee text into A2-9.3 (⧉ identical). `Tests: TestEveryRuleIDMatchesItsCorpusValue, TestNoFalsePositiveOnBenignCorpus, TestErrorMessageNamesFieldAndRuleIDOnly, TestEntropyRuleIsDeterministic, TestGraphSecretFreeSerialization`. | ⧉ PAIR-SEC1 → §3 A1-27; PAIR-X1 → §2 A0-12 |
| P-47, S-04 | A2-8.2, A2-8.3, A2-8.5, A2-10.2 step 13 | BLOCK-A2-08 ⧉. (a) Publish the **per-kind evaluated field set**: `host` → `label` + `addresses` · `network` → `label` + `cidr` + `addresses` · `identity`/`group` → `label` + `domain` + `sid` · `credential` → `label` + `domain` · `share` → `label` + `domain` · `service` → `label` + `protocol` · `evidence_ref`/`finding`/`hypothesis` → derived (below). `attrs`, `summary`, `claim` and `basis` MUST NOT be matched — they are prose (A2-6.7). (b) **Derivation:** a node whose kind has no identity field (`finding`, `hypothesis`, `evidence_ref`) is quarantined by derivation from its edges — if any non-retracted edge connects it to a quarantined node it is quarantined with the same reason; and quarantine propagates **one hop** along `reachable`, `authenticates_to`, `grants_access` from a quarantined `host`/`network` to the attached `service`/`share`. Neither rule is transitive beyond what is stated. (c) **Recomputation triggers:** on a scope/blacklist change (A2-8.5), on a revision that changes any identity field, and on a new node or edge touching a quarantined node — `quarantine_recomputed.trigger` is the closed list of BLOCK-A1-05. (d) A2-8.3 gains: "the target of a spawn or action is **never** taken from a graph field. The scope engine resolves the target itself (A7 action spec); the graph node a request cites MUST be named by `gn_` id so the quarantine check is on the id, not on a worker-supplied string. A request citing a quarantined node id is refused **before** approval routing with `action_blocked{reason:\"target_quarantined\"}` (A1-3.3)." `Tests: TestPerKindMatchedFields, TestOneHopPropagation, TestIdentitylessNodeQuarantinedByDerivation, TestQuarantinedNodeNotTargetableWithValidApproval, TestTargetResolvedByIDNotByString`. | ⧉ PAIR-Q2 → §3 A1-28 |
| P-49 (+P-102) | A2-6.3 | Publish the closed list as a normative constant `ReservedAttrKeys = {id, engagement_id, seq, graph_seq, kind, label, summary, attrs, evidence_id, evidence_ids, addresses, cidr, port, transport, protocol, sid, domain, credential_kind, media_kind, size_bytes, severity, claim, basis, status, content_hash, quarantined, quarantine_reason, report_excluded, supersedes_id, superseded_by_id, provenance, source_id, target_id, source_kind, target_kind, retracted, principal_kind, run_id, job_id, task_id, agent_node_id, user_id, tool_id, tool_version, event_id, recorded_at, observed_claimed_at, confidence, graph_node_id, graph_edge_id}` (note: `operator_id`→`user_id` per A2-10, plus the two new id spellings per A2-14; keep `seq` reserved even though the graph field is `graph_seq`). Require a **byte-exact membership test** — no prefix or substring matching. Replace the ellipsis in A2-6.3. `Tests: TestReservedAttrKeysRejected`. | — |
| P-60, P-61 | A2-5.3, A2-2.8, A2-7.1 | BLOCK-A2-10 ⧉. A2-5.3 adopts A1-2.1's `principal_kind` list verbatim (`platform`, `orchestrator`, `worker`, `node`, `user`), renames the enum value `operator` → `user` and the field `operator_id` → `user_id` (a `usr_` id, AM-1), inserts the mapping table of BLOCK-A1-08 byte-identically, and adds: "**`node` is required** — a Q9/ADR-0013 remote-agent observation must be attributable." A2-5.3 also gains: "`principal_kind` is the principal whose **work produced the content**; `provenance.event_id`'s event `actor` is the platform subsystem that wrote the row (A1-3.3, always `(platform, \"graph\")`). The two are expected to differ, and copying one into the other is a defect." Update A2-2.8's "operator correction is a new node with `principal_kind: operator`" → `user`. Fix `ToolVersionMaxBytes` to the A0 registry value (64). `Tests: TestPrincipalKindIsNotCopiedFromActor, TestFieldNameMappingIsTotal`. | ⧉ PAIR-A1, PAIR-A2 → §3 A1-32 |
| P-50, D-05 + PAIR-N1 (partner; registry decision owned by §2 A0-08) | A2-7.1, A2-6.4 | Replace A2's local constant block with citations of the A0-7.1 registry (PAIR-N1): `ToolVersionMaxBytes = 64` (**A2-7.1's 32 is a defect — delete it**), `EvidenceRefsMax = 8` (replacing `EvidenceIDsMax`), `MaxSupersedeChain`, `AttrsMaxKeys`, `AttrKeyMaxBytes`, `AttrValueMaxBytes`, `AttrsTotalMaxBytes`, `AddressesMax`, `NodeSummaryMaxBytes`, `FindingSummaryMaxBytes`, each with mechanism **R** and a row in A2-7.1's table. A2-6.4: "document-class caps in A2 are measured on the **A0-2 canonical form** of the field's own object — the same bytes that enter `content_hash` (A2-4.6) — computed once at ingest and stored alongside it; the API representation is never a measurement input, so the cap check and the fingerprint can never disagree." | ⧉ PAIR-N1 → §2 A0-08 |
| PAIR-Q1 (partner; decision owned by §3 A1-20) | A2-8.1, A2-8.2, A2-2.8, A2-8.5, A2-8.7 | BLOCK-A2-12 ⧉. Insert BLOCK-A1-04's mapping decision verbatim into A2-8.1. A2-8 gains: "an admin or the assigned operator MAY **tighten** quarantine (`SetQuarantine` with `quarantined:true`, preserving the reason in force) and MUST be accompanied by `graph_node_quarantined{operator_quarantine}`. There is **no release operation**: an `out_of_scope` node is released only by an operator scope change and the recomputation it causes (A2-8.5, ADR-0016 §2); releasing a `blacklisted` node is `conflict` and MUST NOT be offered (A2-8.5). A release is stored as `quarantined:false` with `quarantine_reason` absent; the A1 event is the only record of the previous state." Update A2-2.8's authority table row for `quarantined`/`quarantine_reason` accordingly. | ⧉ PAIR-Q1 → §3 A1-20, §1 PO-9 |
| C-05 (+E-02) | A2-2.8, A2-8.7, A2-8.8, A2-12.5 | "`report_excluded` MUST be settable to `true` **only** on a node with `quarantined:true` (Q5 authorizes removal of quarantined discoveries from the report only). An attempt on a non-quarantined node is `conflict` (A0-3.1, immutable state) and MUST be recorded as `action_blocked{reason:\"graph_write_rejected\", action_kind:\"graph_write\"}`. Removing a confirmed finding from a report is expressed by a revision (A2-4) to `status:\"refuted\"` or `severity:\"info\"`, never by a flag." Add the E-02 test names to the clauses they guard: A2-8.3 `Tests: TestQuarantinedNodeNotTargetableWithValidApproval`; A2-8.5 `Tests: TestBlacklistedNodeSurvivesScopeWidening`; A2-8.7 `Tests: TestMachinePrincipalCannotSetReportExcluded, TestReportExcludedRequiresQuarantine`; A2-8.8 `Tests: TestQuarantineFlagCannotBeSuppliedOnWrite`; A2-12.5 `Tests: TestQuarantinedNodeAbsentFromPlanningView, TestRetractedEdgeAbsentFromPlanningView`. | — |
| D-02 (+PAIR-N2 half) | A2-4.6, §4 `contentDoc` | "The fingerprint MUST be computed from `contentDoc` **only**; `Node` MUST NOT be passed to `cjson` for fingerprinting (`Node` carries `omitempty` tags and A0-8.3 absence semantics, `contentDoc` carries the fixed 20-key set). `contentDoc.Attrs`, `.EvidenceIDs` and `.Addresses` MUST be non-nil before marshaling (A0-2.14): a canonical content document containing `null` is a platform defect → `internal`." `Tests: TestContentDocFixedKeySet` (reflection: the canonical bytes of a zero-valued `contentDoc` contain exactly the 20 keys of A2-4.6 in byte order), `TestContentHashVector`. | ⧉ PAIR-N2 → §2 A0-06 |
| S-11 | A2-10.2 | Insert into the 13-step order: "(10a) compute `content_hash` over the A2-4.6 document; (10b) evaluate policy and quarantine (current step 13); (10c) **then** attempt the dedup collapse of A2-3.4/A2-4.7", and state: "a collapse still emits the A1 `graph_node_written` (with `dedup_hit:true`) and still emits `graph_node_quarantined` when the recomputed quarantine state differs from the stored one — so an agent that re-observes a blacklisted host 500 times produces 500 chained signals, not one." `Tests: TestValidationOrderIsDeterministic, TestCollapseStillEmitsQuarantineEvent`. | — |
| P-48 | A2-8.2, A2-10.2 step 13, §4 | "`internal/graph` declares and consumes exactly one interface: `type QuarantineDecider interface { Classify(ctx context.Context, engagementID string, n NodeDraft) (QuarantineState, error) }`. `internal/policy` provides the implementation; `internal/graph` never imports `internal/policy` (DESIGN §1 layering, DESIGN §4: interfaces are defined at the consumer)." | — |
| P-52 | A2-3.6 (+A2-3.5) | "For `contradicts` the platform normalizes the stored direction to `source_id < target_id` byte-wise (A0-1.9) at composition, which makes A2-3.5's uniqueness constraint on `(engagement_id, kind, source_id, target_id)` sufficient and makes a replayed write converge; endpoint roles carry no meaning for this kind." `Tests: TestContradictsDirectionNormalized`. | — |
| P-53 | A2-4.5 | "A `supersedes` write whose target chain already holds `MaxSupersedeChain` (64) revisions is `conflict` (A0-3.1) naming the bound and the chain's first `gn_`; the bound is enforced at **write** time, and A2-4.5's read bound is the consequence." `Tests: TestSupersedeChainBoundAtWrite`. | — |
| P-55 | A2-6.6 | "The distinct `attrs` keys and their per-kind occurrence counts MUST be **derivable from stored rows** (`attrs` stored per key, never as an opaque blob), ordered by key byte-wise (A0-1.9). The operator-facing surface is the UI's (ADR-0011), not a `/api/v1` query endpoint (Q1); A4 decides whether it exists in v1." (Removes an unowned MUST that reads like the free-form query Q1 excludes.) | — |
| P-56, P-57, P-58, P-73, P-76 | A2-2.3, A2-1.7, §4 header, A2 header "Depends on", A2-7.1 last row | A2-2.3: "`size_bytes` MUST be ≥ 1 — a zero-byte artifact is not an artifact — so absence and zero cannot be confused in the served representation (A0-8.3/8.4); `contentDoc` still carries `0` for kinds where the field does not apply (A2-4.6)." A2-1.7: "Floats, `null` (A0-8.3) and nested objects MUST NOT appear in any node or edge field. `attrs` (A2-6) is the only nested structure and is itself restricted to flat scalar values: no float, no `null`, no array, no deeper object." §4 header: "**Domain layer** (DESIGN §1); imports foundation only: `internal/ids`, `internal/cjson`, `internal/errs`." Document header: "Depends on A0 (**Draft** at the time of writing; every work package below is gated on the A0/A1/A2 Freeze PR)." A2-7.1 last row: "not assigned here — A3 per A0-7.7." | — |
| PAIR-V5 (partner; decision owned by §2 A0-09) | A2-5.4, A2 §6.9 | Change A2-5.4's citation from "envelope shape per A0-2.17 vector V5" to "**envelope shape per A1-1.1** (A0-2.17 V5 is a canonicalization vector with a synthetic key set, not a valid event — see A0-2.17's annotation)" and delete A2 §6 item 9's "A1 was not on disk when A2 was drafted" caveat, keeping only the list of event kinds A2 needs from A1 (all now exist: A1-3.3 `graph_*`, `quarantine_recomputed`, `report_inclusion_changed`, `action_blocked{graph_write_rejected}`). | ⧉ PAIR-V5 → §2 A0-09 |
| P-59, P-79, P-81, P-84, P-85, P-99, P-100, P-101, P-103, P-104 (+ the A2 test names listed under §2 row A0-17, and the A2 unknown-field names of §3 row A1-29, and the PAIR-K1 cursor oracle) | new **§4.3 Contract tests (A2)** subsection; A2-10.3, A2-11.4 | One subsection listing per clause, verbatim: A2-11.4 `TestGraphNodeIDFromAIsNotFoundInB` (**one oracle:** the node's `attrs.graph_node_id` is the requested id — ids are not secrets, A0-1.7 — and the `message` does not distinguish "absent" from "elsewhere", A0-3.9), `TestCursorFromEngagementARejectedInB` (**`validation` only**, not "empty page or `validation`" — PAIR-K1/AM-4), `TestNodeDedupDoesNotSpanEngagements`, `TestEdgeDedupDoesNotSpanEngagements`, `TestNoCrossEngagementEdge`, `TestNoBareNodeIDInGraphDocuments`; A2-10.3 `TestHardRejectMatrix` (one subtest per rejection: unknown node kind, unknown edge kind, wrong endpoint kind, not-applicable field, unparseable `cidr`, over-cap field, over-cap `attrs`, reserved key, nested/float/null `attrs`, self-edge, supersede of non-current, supersede cycle, cross-engagement endpoint, missing provenance); A2-1.3 `TestImmutableFieldUpdateIsConflict`; A2-3.8 `TestDenormalizedEndpointKindsEqual`; A2-4.8 `TestContentHashRecomputedFromStoredBytes`; A2-8.9 `TestQuarantineFlagFlipDoesNotSkipRows`; A2-10.8 `TestRejectedWriteIsStillChained`; A2-10.2 `TestNodeWriteRejectsUnknownField`, A2-1.5 `TestNodeReadIgnoresUnknownFields`; A2-7.1 `TestMaximalNodeFitsCaps` (a maximal node per kind: every field at its cap, `attrs` 16 keys × 512 B, `addresses` 16 × 64 B, `evidence_ids` 8); A2-3.2 `TestEdgeEndpointMatrixRejects`; A2-4.3 `TestSupersedeCycleRejected`, `TestSupersedeNonCurrentRejected`. | ⧉ PAIR-K1 |
| PO-5, PO-7, PO-8, PO-9 (§1) | A2-8.5, A2-2.7/5.6, A2-5.3, A2 §6 | Apply the §1 rows: PO-5 (A2-8.5 normative "recorded, not refused" + §6 item 12), PO-7 (§6 item 13 signature wording), PO-8 (BLOCK-PO8 in A2-5.3 + §6 item 2), PO-9 (BLOCK-PO9 §6 sentence). | ⧉ PO-8/PO-9 pairs |

### 4.1 Blocks (A2) — verbatim insert text

```
BLOCK-A2-04 (A2-4.7 / A2-5.3 / A2-5.6 — ⧉ identical decision content to A1 BLOCK-A1-06)
`provenance` is a bounded list of at most 8 entries, each carrying its own
`event_id`, `principal_kind`, `run_id`, `job_id`, `task_id`, `agent_node_id`,
`tool_id`, `tool_version`, `recorded_at`, `observed_claimed_at` and `confidence`.
The list is ordered by the `seq` of `provenance[].event_id`, never by arrival, so a
node's bytes are deterministic. A dedup collapse (A2-4.7) appends the new
observation's entry instead of discarding it — except that an entry whose `event_id`
is already present is **not** appended, which is what keeps a replayed A1 event
idempotent (ADR-0013 offline buffering). The node's `confidence` is the highest
grade in the list and MUST be raised to `verified` only when the new entry's
`event_id` differs from every existing one **and** its `task_id`/`agent_node_id`
differ: a second, independent observation (A2-5.6). That is the only path by which
`verified` is stored. Every collapse emits the A1 `graph_node_written` event with
`dedup_hit:true`.
Tests: TestNodeDedupKeepsEveryObservation, TestVerifiedRequiresIndependentObservation,
TestReplayedEventAddsNoProvenanceEntry, TestProvenanceListIsOrderedByEventSeq.
```

```
BLOCK-A2-06 (§4 sketch — replaces NodeDraft/NewNode/WriteNode)
// NodeDraft is the ingest path's input: content only, no platform-set field.
type NodeDraft struct {
	Kind                     NodeKind
	Label, Summary           string
	Attrs                    Attrs
	EvidenceIDs, Addresses   []string
	CIDR                     string
	Port                     int
	Transport, Protocol      string
	SID, Domain              string
	CredentialKind           CredentialKind
	EvidenceID               string
	MediaKind                MediaKind
	SizeBytes                int64
	Severity                 Severity
	Claim, Basis             string
	Status                   string
	SupersedesID             string // A2-4.2: set by the revising write, not by the platform
}

// PendingNode is validated content + provenance + quarantine state + content_hash,
// with no id and no graph_seq (assigned at insert, A2-1.2/A2-1.4a). DESIGN §4: a
// constructor never returns a half-built value.
type PendingNode struct { /* unexported; NewNode is the only way to build one */ }

func NewNode(in NodeDraft, prov Provenance, q QuarantineState) (PendingNode, error)

// WriteNode assigns graph_node_id (A0-1.4) and graph_seq (A2-1.4a) inside the insert
// transaction, applies A2-3.4/A2-4.7 dedup, and returns the complete Node. A2-10.6:
// the seam accepts PendingNode and nothing else.
func WriteNode(ctx context.Context, engagementID string, n PendingNode) (Node, error)
```

```
BLOCK-A2-14 (A2-1.6 — ⧉ byte-identical table to A1 BLOCK-A1-08's mapping bullet)
One value, one name platform-wide (A0-3.6): a graph node is `graph_node_id` and a
graph edge is `graph_edge_id` in every JSON document, payload, error body and log
attribute; `node_id` remains the remote agent node (`slp_node_`, Q9). A2's served
field `id` is renamed accordingly (pre-Freeze, additive-only afterwards).

| A1 event payload field | A2 graph field (Go / JSON) |
|---|---|
| `graph_node_written.graph_node_id` | `Node.ID` / `graph_node_id` |
| `graph_edge_written.graph_edge_id` | `Edge.ID` / `graph_edge_id` |
| `from_graph_node_id` | `Edge.SourceID` / `source_id` |
| `to_graph_node_id` | `Edge.TargetID` / `target_id` |
| `supersedes_graph_node_id` | `Node.SupersedesID` / `supersedes_id` |
```

## 5. Conflicts between the two reviewers, and your resolution

**Precedence rule applied (stated as required):** for **safety semantics** the adversarial
reviewer wins; for **notation, buildability, naming and test names** the principal wins; where
both are safety-relevant and incompatible, the **stricter** option is chosen and recorded as a
PO-confirm item in §1.

| # | clause | principal wanted | adversarial wanted | resolution | one-line reason |
|---|---|---|---|---|---|
| 1 | A1-6.5 authority | (no finding; §3 lists it as PO-open) | C-01: admin-only, delete the operator sentence | **C-01**, as §1 PO-1 SAFE DEFAULT | safety + Q11 narrowing must be explicit and signed |
| 2 | A1-6.5 lifetime | (none) | S-07: single-use override | **S-07**, §1 PO-1 (stricter) | one human decision must not release an unbounded artifact stream |
| 3 | A2-4.7 / A2-5.6 | P-42: `confidence` in the dedup key; a higher-grade re-observation becomes a **new node** + `supersedes` edge | S-05: keep the dedup key, append a **provenance entry**, raise `verified` on independence, `dedup_hit` on the event | **S-05** (with the replay guard I added) | safety wins: P-42 forks one fact into two nodes, breaking A2-4.7's replay idempotence (ADR-0013) and losing the first observer's attribution |
| 4 | A2-8.2 propagation | P-47: one hop only, never transitive; `finding`/`hypothesis`/`evidence_ref` match `label` | S-04: identity-less kinds are quarantined **by derivation from their edges**; recompute on node/edge writes | **both, composed**: S-04's derivation for identity-less kinds + P-47's one-hop rule for `service`/`share`; neither transitive beyond what is stated | stricter union; a `finding` about a blacklisted host must not stay in a planning view |
| 5 | A2-8.3 target resolution | (none) | S-04: target never from a graph field, cited by `gn_` id, new `target_quarantined` reason | **S-04** | safety: a worker-supplied string must not be the thing the blacklist matches |
| 6 | A2-10.6 field visibility | P-43: fields stay exported; delete the claim; `TestNodeHasNoExportedContentSetter` | E-07: fields unexported; `TestNodeFieldsUnexported` | **P-43** | buildability + naming: `encoding/json` cannot marshal unexported fields and no `go vet` audit exists; E-07's mechanism is unimplementable as sketched |
| 7 | A0-2.17 new vector | P-04's V7 = `{"s":"<a href=\"x\">&é\u2028"}` (escape decoding + HTML chars) | F-03's V7 = `{"s":"a\u2028b\u2029c\u007fd"}` (line separators + DEL) | **publish both**: V7 = F-03's, V8 = P-04's; exact values in §6.2 | additive, no published digest changes, both escaping paths get locked |
| 8 | A0-2.14 `null` | P-06: `CanonicalValue` rejects `null`; byte-level `Canonical` MAY accept it (A0-2.8 lists `null` as a literal) | F-06: `cjson.Canonical` MUST reject `null` in input | **F-06** (blanket reject) + amend A0-2.8 to "`null` is recognized when scanning input and rejected by `Canonical`" | stricter, and no contract caller needs `null` (A0-8.3); one rule beats two entry points |
| 9 | A0-4.4 cursor shape | P-17 option B: add `"d":1\|-1` to the cursor | S-01: no MAC, derive the seek from the resolved row, ignore `k` for seeking | **S-01**, and P-17's unenforceable MUST is deleted | S-01 closes the skip hole without a breaking cursor-shape change; a direction bit would still be forgeable |
| 10 | A1-5.8 emission | P-37: name the constant `HeadLogIntervalSeq = 1000` | S-02: every verification, every **100**-`seq` crossing, every `run_ended`/`hard_stop_fired`, into an append-only trail | **S-02**; `HeadLogIntervalSeq = 100` | safety: 999 silently truncatable rows is not a mitigation |
| 11 | A1-5.4 / A0-5.4 clamp | AM-3: record the unbounded forward clamp in A0 | T-07: bound the clamp to 1000 ms, log it, clamp `occurred_at` too | **T-07's bound inside AM-3's A0 home**; the `clock_anomaly` kind (T-05) is deferred to §7 | safety/timeline integrity; a new kind is not needed to get the bound and the error log |
| 12 | A1-3.3 approval fields | (P-28: fingerprint divergence → `conflict`) | C-02: add `action_spec_evidence_id`; T-03: also `target_graph_node_id`, `argv_hash` | **T-03 (superset of C-02) + P-28's `conflict`** | pre-Freeze additive; A7 cannot be written without the reservations |
| 13 | A2-5.3 principal vocabulary | P-60: adopt A1's five values, `operator`→`user` | (implicit: `node` must exist for Q9 attribution) | **P-60**, and A2-2.8's prose is updated in the same edit | one vocabulary; without `node` a remote-agent observation is unattributable |
| 14 | A1-4.9 secret scan | P-46: A2-9.4 publishes the rule table; P-69: A1 deletes its own list | C-07: the scan is a filter, egress-by-kind is the enforcement point | **all three**: A2 owns the table, A1 cites it, both carry the identical filter sentence (PAIR-SEC1) | one source of truth for the rules, and no unsounded "therefore no secret can reach a model" deduction |
| 15 | reject vs redact | P-46's table implies reject | A2 §6.10 / A1 §6.4 both ask | **reject**, ruled once in §1 PO-6, cited from both | Q3 hard-reject posture; `redacted:true` means platform redaction, never rejection |
| 16 | `TestCursorFromEngagementARejectedInB` oracle | P-71: `validation` everywhere | (A2 text allowed "empty page or `validation`") | **`validation` (400)** in A1-8.2 and A2-11.4 | one test name, one oracle (AM-4) |
| 17 | test names | `TestReservedAttrKeysRejected` (P-49/P-102) · `TestNilCollectionNeverSerializesAsNull` (P-06) · `TestHasMoreDetection` (P-90) · `TestParseTimeRejects` (P-91) | `TestReservedAttrsKeyRejected` (E-05) · `TestNilCollectionsNeverCanonicalizeToNull` (F-06) · `TestExactFullPageHasNoNextCursor` (E-05) · `TestTimeParseRejectsNotNormalizes` (F-04) | **principal's names win**; the adversarial intents become subtests (round-trip subtest inside `TestParseTimeRejects`; `TestExactFullPageHasNoNextCursor` recorded as an alias of `TestHasMoreDetection`) | naming precedence; one name per test keeps the shared suite grep-able |
| 18 | A2-4.6 measurement | P-50: measure `attrs` on the A0-2 canonical form | D-05: same, and store the measurement next to `content_hash` | **D-05** (superset of P-50) | identical intent; storing the value removes any read-back disagreement |

## 6. Digest-critical changes (vector recomputation required)

**Verified recipe (used to produce every value below, and the one writers MUST use):**
`python3` + `json.dumps(obj, sort_keys=True, separators=(',',':'), ensure_ascii=False).encode('utf-8')`
+ `hashlib.sha256(...).hexdigest()`. This recipe **reproduces all six A0-2.17 vectors and all three
A1 §4.3 rows byte-exactly** (checked during this run) — so it *is* the A0-2 canonical form for
these documents, and every new vector below is writer-recomputable. No vector in this plan needs
the "to be regenerated by the contract-test suite" escape hatch.

### 6.1 A1 §4.3 — chain vector: rows 0–2 unchanged, **row 3 added** (D-03)

- **Unchanged (verified by recomputation):** row 0 `chain_genesis` 428 B / `d65ade15…`, served
  589 B / `137b750d…`; row 1 `command_executed` 816 B / `a5641155…`, served 977 B / `752bd834…`;
  row 2 `graph_node_quarantined` 715 B / `05a06eab…`, served 876 B / `9d55ad66…`.
  **None of the assigned changes touches these bytes:** P-01 only makes the notation match what is
  already published (`(platform,"event_store")`, `(worker,"")`, `(platform,"graph")`); C-06 deletes
  `operator_release`, not row 2's `blacklist_match`; P-64/T-04 add fields to `graph_node_written`,
  which is not in the vector; T-01/T-02 add kinds, not envelope keys (still 17).
- **Row 3 (new, normative).** Kind `task_result`, worker-composed, `seq` 3,
  `prev_hash` = `05a06eabc89552dd1798ac918c65f10ab2c8c778cbda08d626929b6770785317`.
  Preimage (**702 bytes**, exact — note the `result_summary` contains a literal U+2028 (E2 80 A8),
  a literal U+2029 (E2 80 A9), a literal 0x7F and a literal 😀 (F0 9F 98 80); they are invisible in
  a terminal, so verify by length and digest, not by eye):

```
{"actor":{"component":"","principal_id":"task_01m1y2whfh1txm57x8dn41r9hg","type":"worker"},"engagement_id":"eng_01m1y2whfhgbz06ays6dxnvyws","event_id":"evt_01m1y2whfhz8k3p5r7t9v1x3z5","evidence_refs":[],"job_id":"job_01m1y2whfhbt69j0h0fbxepw90","kind":"task_result","node_id":"","occurred_at":"2026-09-07T14:20:11.380Z","occurred_claimed_at":"2026-09-07T14:20:09.120Z","payload":{"command_count":3,"duration_ms":9007199254740991,"error_kind":"","result_summary":"Relay confirmed.\u2028Second line.\u2029DEL:\u007f done 😀","revert_event_ids":[],"status":"succeeded"},"recorded_at":"2026-09-07T14:20:11.400Z","run_id":"run_01m1y2whfhnjx2am9103w0pnqw","task_id":"task_01m1y2whfh1txm57x8dn41r9hg","untrusted":true}
```

  (`\u2028`, `\u2029`, `\u007f` above denote the **literal code points in the canonical bytes**, per
  A0-2.7 after BLOCK-A0-03 — the writer MUST write the real characters into the fenced block, not
  the escapes, and MUST state that in the row's note.)
  - preimage len **702** · `hash` **`1cae22e3bba2c7cf2b15fc920aadbd1c45f18ce07675680c96a4ce46538215c1`**
  - served len **863** · served SHA-256 **`0c79020c4ebce6315d2d29eeef75b3d301e929769de719804a078f730217b614`**
  - `duration_ms` = 9007199254740991 = 2^53−1, the A0-2.6 maximum (16 digits — the point of the
    row: it locks number literal text through preimage → served → preimage). A1-4.2 declares no
    upper bound for `duration_ms`; add the note "the 16-digit value is deliberately implausible: it
    exercises the A0-2.6 bound and F-01's literal-text rule."
  - **Writer duty:** run the §8 script, assert rows 0–2 reproduce the published values *with the
    same script*, then emit row 3's four numbers. If rows 0–2 do not reproduce, stop and report —
    do not publish row 3.
- Also add the served-envelope JSON block for row 3 in the same style as rows 0–2 (pretty-printed,
  with `seq`/`prev_hash`/`hash`).

### 6.2 A0-2.17 — V1–V6 unchanged, V7 and V8 added

- **V1–V6: no change.** Recomputed during this run: lengths 13/26/57/45/113/18 and all six digests
  match the published values exactly (`d3626ac3…`, `735f90d3…`, `63fc2ef8…`, `d66e67ee…`,
  `3695e846…`, `9fbfff35…`). P-20 is **documentation-only** (print the literal U+FFFD in the
  "exact" column; len 18 and the digest are already right).
- **V5: RELABEL, do not fix** (P-66 ruling). Changing V5's key set would invalidate a published
  digest for no gain — V5's job is the A0-2.12 exclusion mechanism, not event shape. Annotate it
  (BLOCK-A0-09) and move A2-5.4's citation to A1-1.1.
- **V7 (new, F-03).** Input `{"s":"a\u2028b\u2029c\u007fd"}` → canonical bytes
  `{"s":"a<2028>b<2029>c<7F>d"}` where `<2028>`=`E2 80 A8`, `<2029>`=`E2 80 A9`, `<7F>`=`7F`
  (all literal; U+007F is **not** escaped — F-03's prose showing `\u007f` in the canonical column
  is wrong and A0-2.7 decides it). **len 19 · SHA-256
  `aedd6df88cc462fdbdc5788549d753c9b8c21ac8b9e16c51c72016cf564c3e85`.**
- **V8 (new, P-04).** Input `{"s":"<a href=\"x\">&é\u2028"}` → canonical bytes
  `{"s":"<a href=\"x\">&é<2028>"}` (literal `<`, `>`, `&`; `é` = `C3 A9`; U+2028 literal; the `\"`
  stays escaped because A0-2.7 requires it). **len 28 · SHA-256
  `63995ca86de5cce6f4d74df8e1d90aed78918aad912cc7c7feaff4d89fb15b2d`.**
- New **rejection** entries (no digests): `{"a":1E3}`, `{"a":-0.0}`, `{"a":9007199254740993}`,
  `{"a":10000000000000000000}`, `{"a":"\ud800"}`, `{"a":"\xff"}`, `{"a":null}`, a 33-deep
  document, a >1 MiB document. Add one **accept** pair: a 32-deep document and `{"a":"\ud83d\ude00"}`.
- Writer duty: recompute V7/V8 with the §8 script and paste the script output into the progress
  report (they must equal the values above).

### 6.3 A1-7.6 `PayloadHash` vector (P-33, new)

Preimage = the A0-2 canonical form of `{"kind":…,"payload":…}` only (no envelope field). Worked
example, using §4.3 row 1's kind and payload: **len 293 · SHA-256
`19d976a83cc9d36ac160313a20b80c0745fff805526b7f43f88d05e33c7be5e5`** over

```
{"kind":"command_executed","payload":{"command":"nmap -sV -p 445 10.20.0.14","duration_ms":48210,"exit_code":0,"output_bytes":18432,"output_evidence_id":"evi_01m1y2whfh3ca875z2x8v8h7qt","redacted":false,"target":"10.20.0.14","tool_id":"tool_01m1y2whfhfjdvwqp9pfxqekmf","tool_version":"1.4.2"}}
```

### 6.4 A2 §4.2 — new normative content-fingerprint vectors (D-01/P-82), and §4.1 example fixes

The 20-key `contentDoc` of A2-4.6 is **unchanged** by every assigned A2 edit (verified against the
clause): `graph_seq` (P-40) and `graph_node_id` (P-65) are not content keys, `provenance` (S-05) is
excluded, and `attrs`' flat wire shape (F-07) is what the examples already show. So these values are
final. Apply P-40/P-65/F-07/P-41 **before** transcribing them.

| row | node | canonical bytes (exact) | len | SHA-256 |
|---|---|---|---|---|
| **F1** | `finding` (the §4.1 example) | `{"addresses":[],"attrs":{"cvss_v3_x10":88,"first_seen_task":"task_01m1y2whfh1txm57x8dn41r9hg","relay_tool":"ntlmrelayx"},"basis":"","cidr":"","claim":"","credential_kind":"","domain":"","evidence_id":"","evidence_ids":["evi_01m1y2whfh3ca875z2x8v8h7qt","evi_01m1y2whfh7kq2m4c8x1z9vb3n"],"kind":"finding","label":"SMB relay to SYSVOL on dc01","media_kind":"","port":0,"protocol":"","severity":"high","sid":"","size_bytes":0,"status":"confirmed","summary":"Captured NTLM authentication from 10.20.0.14 was relayed to the SYSVOL share on dc01, yielding read access to group policy preferences. Secret material is referenced, not stored (evi_).","transport":""}` | **656** | `ad8f188e63b2563f9adad88df585d94df9c662e4082895e974b0b31e8a65076e` |
| **F1-R** | same node, `evidence_ids` **given** in reverse order | sorted per A2-4.6 before canonicalization → **identical bytes and digest to F1**. Forbidden variant: canonicalizing unsorted yields `4a017e71658de7ebb2d3a5429f90818b8a69302ff1909badbc782d9049ce9cae` — an implementation that produces this digest is defective (P-41). | 656 | `ad8f188e…076e` (must equal F1) |
| **F3** | F1 with `attrs.cvss_v3_x10` = 87 (one `attrs` value differs) | identical to F1 except `"cvss_v3_x10":87` | **656** | `1748b813a0d28d9f17f4d89dff2a08536741bf45e8fe0defbb3c4363d3f9b983` |
| **S1** | `service` (inapplicable keys at zero values) | `{"addresses":["10.20.0.14"],"attrs":{},"basis":"","cidr":"","claim":"","credential_kind":"","domain":"","evidence_id":"","evidence_ids":[],"kind":"service","label":"microsoft-ds","media_kind":"","port":445,"protocol":"smb","severity":"","sid":"","size_bytes":0,"status":"","summary":"","transport":"tcp"}` | **304** | `3955d82160284d3e76c9be1b06981530df50e1635a1bad7ceb7b50c5f73c2b67` |

**A2 §4.1 example fixes required by the renames (digest-critical):**
- Rename the five `"id":` keys to `"graph_node_id"`/`"graph_edge_id"` (P-65) and every `"seq":`
  to `"graph_seq":` (P-40).
- The `finding` example's `content_hash` (currently `7f3c1a92de48b06f5ac7d1e8b93042fa6c5d7e81b2a39f04c6d81e5b7a290c34`,
  line ~919) MUST become **`ad8f188e63b2563f9adad88df585d94df9c662e4082895e974b0b31e8a65076e`**
  (= vector F1) so the example and §4.2 agree.
- The other three `content_hash` values (lines ~955, ~978, ~1016) MUST each be recomputed from that
  example's own `contentDoc` with the §8 script; if an example lacks the fields to build one,
  annotate the value `(illustrative — not a §4.2 vector)`. Do not leave a fabricated digest that
  looks normative.
- The `finding` example's `provenance` MUST become a **one-or-two-element array** (S-05). To keep
  `confidence:"verified"` legal it MUST show **two** entries from different `task_id`/`agent_node_id`
  values (BLOCK-A2-04's independence rule); a single-entry node MUST show `inferred` or `observed`.
- Writer duty: recompute all four rows with the §8 script and paste the output into the progress
  report.

### 6.5 Changes that do **not** touch any digest (confirmed, so writers do not recompute blindly)

Envelope key set stays 17 (no assigned change adds an envelope field) · A0-2.17 V1–V6 · A1 §4.3
rows 0–2 · A2-4.6's 20-key set · A2 `content_hash` under the `graph_seq`/`graph_node_id` renames ·
`evidence_refs` for every §4.3 row (row 1's `output_evidence_id` still yields one entry under
P-07's per-kind table; rows 0 and 2 have none).

## 7. Deferred (not applied before Freeze)

**Integrator-only (no writer may touch these files — §8):**
- `contracts/README.md`: P-68's tie-break sentence ("A0 gates every later contract; where A0 and
  A1–A8 disagree on a cross-cutting convention, A0 wins and the later document is defective") and
  E-06's seventh merge-gate bullet (safety-path pairing: one positive + one negative test id per
  safety clause; a contract reaches `Frozen` only when every such clause has both). Reason: outside
  the three writers' file scope. **Record in the PR body as a required integrator commit** — E-06 is
  the rule that makes the ~60 test ids in §2–§4 enforceable.
- `sessions/` record + `sessions/update-usage.sh`, and flipping A0/A1/A2 status `Draft`→`Frozen`
  (WP-00): integrator, after the PO signs §1.

**SHOULD FIX / NICE not assigned — additive-after-freeze-safe (A0-6.5) unless marked PR-body debt:**

| ids | why deferred | additive-safe? |
|---|---|---|
| T-05 (`clock_anomaly` kind) | T-07's bound is applied via A0-5.4 + an error-level slog record; a new kind is not needed for the safety property | yes — new kind, A1-3.5 |
| S-12 (`evidence_removed` kind) | A1-8.8's "deletion is itself recorded" has no kind, but **no deletion mechanism ships in v1**; must land before one does | yes — **PR-body debt** ("A1-8.8 is unimplementable until `evidence_removed` exists") |
| C-09 (`credential_revoked`, `node_paired`, `node_unpaired`) | Q8/Q9 lifecycle audit is A5's token contract; the kinds are additive and A5 is not started | yes — **PR-body debt** (revocation is currently unprovable from the chain) |
| C-08 (`engagement_assignment_changed`) | folded into §1 PO-4 as an A5 duty; not added to the closed list now (keeps the count at 42) | yes |
| T-08 (`run_ended.cancelled_by`) | attribution nicety; `actor` is platform for the cancel path | yes |
| S-14 (`seq` publication note) | documentation of an accepted posture (A0-1.7) | yes |
| F-05 (advisory-lock wording) | A1-5.4's `SELECT … FOR UPDATE` on the chain-head row is already normative; the pgx detail is WP-22 | yes |
| D-07 / P-39 (cursor prose order) | assigned inside §2 A0-17 (one clause, one edit) — listed here only to show it is not lost | n/a |
| P-77 (`cvss_v3_x10` has no range) | the only principal finding left unassigned: adding a `[0,100]` range **narrows** an accepted value set, which is not additive after Freeze | **no — PR-body debt**, needs an ADR (A0-7.2) or a pre-Freeze PO nod; a wrong CVSS ×10 is a data-quality bug, not a safety hole |
| P-38, P-39, P-75, P-76, P-79, P-81, P-83…P-105, D-07 | **all assigned** in §2–§4 (P-79 in §4 A2-23, P-81 in §4 A2-23, P-105/D-08 in §2, D-07/P-39/P-75 in §2 A0-17, P-76 in §4 A2-22) — listed here only to show nothing is lost | n/a |

| E-08 (approval-path JSON example in A1 §4.2) | documentation, not a rule: the `fingerprint_hash`/`expires_at` equality obligations it would illustrate are already normative (§3 A1-04) | yes — an example can be added at any time |
| E-06 (README seventh merge-gate bullet) | **integrator-only** — no writer may touch `contracts/README.md`; see the bullet above this table | n/a — required integrator commit |

**Adversarial SHOULD/NICE assigned anyway** (all are entailed by a MUST row or are one-line
wording): C-07, C-10 (PO-8), D-04, D-05, D-06, D-07, D-08, E-05, E-07, F-05→**deferred**, F-06,
F-07, F-08, F-09, S-09, S-10, S-11, S-13, T-04, T-06, T-07. **Deferred:** C-09, E-06
(integrator-only), E-08, F-05, S-12, S-14, T-05, T-08 — reasons in the table above.

**Residual risk to print in the PR body:** (1) engagement assignment and credential revocation are
unaudited in v1 (C-08, C-09) until A5 lands; (2) evidence deletion has no kind (S-12); (3) tail
truncation by an attacker with both store and log access remains open unless PO-3 approves the
webhook anchor; (4) `cvss_v3_x10` has no range (P-77); (5) A1 §4.2 has no approval-path example
(E-08); (6) A1-8.8's "a deletion is itself recorded" is unimplementable until `evidence_removed`
exists (S-12).

## 8. Writer instructions (identical text for all three)

You are one of three writers. Your only input is this plan. Do not read the review reports.

1. **Edit only your own file** — `contracts/A0-conventions.md` **or** `contracts/A1-events.md`
   **or** `contracts/A2-graph.md`. Never another contract, never `contracts/README.md`, never
   anything else under `/home/wtadmin/Sleipnir`. No git write command of any kind
   (no `add`, `commit`, `checkout`, `stash`, `mv`, `rm`).
2. **Keep the document shape** of `contracts/README.md` §"Document shape": 1 Header · 2 Scope ·
   3 Normative clauses · 4 Types · 5 Traceability · 6 Open for product owner — in that order. New
   subsections (e.g. "4.1 Contract tests") are added **inside** the right section; never add a
   seventh top-level section.
3. **Wording:** MUST / MUST NOT / MAY exactly as RFC 2119 uses them. Every clause stays **citable**
   (`A0-2.7`, `A1-4.2`, `A2-8.2`): if you add a clause, give it the next free number in its group
   (or a letter suffix like `A2-1.4a` where this plan names one) and never renumber an existing
   clause — other documents cite them.
4. **Typography — match the existing text of your file, do not invent a new style:** ASCII
   hyphen-minus `-` for compounds and ranges; the middle dot `·` as the list/row separator the
   documents already use; the em dash `—` only where the surrounding prose already uses one (the
   three contracts all use it); no smart quotes, no en dashes, no new punctuation. Keep the existing
   line-wrap style (~78 columns) and the existing table formats. When you copy a BLOCK from this
   plan, copy its punctuation unchanged — the blocks are already written in the documents' style.
5. **No compiled code.** Go appears only in fenced sketches marked illustrative (README §4). Copy
   the blocks in §2.1/§3.1/§4.1 of this plan verbatim, including their comment lines.
6. **Closed vocabularies are additive-only** (A0-6.5, A1-3.5, Q13): you may add enum values and
   kinds; the only deletions this plan authorizes are `operator_release` (A1-3.3, §1 PO-9 / §3
   A1-20) and the two contradictory sentences named in §1 PO-1. Renames are authorized only where
   this plan says "pre-Freeze" (`graph_seq`, `graph_node_id`/`graph_edge_id`, `operator`→`user`,
   `operator_id`→`user_id`, `EvidenceIDsMax`→`EvidenceRefsMax`, `Truncate`'s `cap`→`limit`).
7. **Every new MUST on the safety path names a positive test AND a negative test** (AGENTS.md):
   write `Tests: <Positive>, <Negative>` in the clause itself, using the exact names this plan
   gives. If a plan row gives only one name for a safety rule, add the missing counterpart as
   `Test<Rule>BypassRejected` and list it in your progress report.
8. **Vectors:** recompute any vector your change invalidates with `python3`
   (`json.dumps(obj, sort_keys=True, separators=(',',':'), ensure_ascii=False)` +
   `hashlib.sha256`) and **show the script and its output in your progress report**. Before
   publishing a new vector, assert that your script reproduces the already-published ones (A0:
   V1–V6; A1: §4.3 rows 0–2) — if it does not, stop and report instead of publishing. Use the
   values in §6 of this plan; they are final.
9. **Cross-references:** if your change alters a clause another document cites, **do not edit the
   other file** — list it in your progress report under "cross-references to verify" as
   `<other doc> <clause> — what changed — what the other writer must check`. Rows marked ⧉ /
   `PAIR-xx` are the exception: there you insert the given block **byte-identically** into your own
   file only; the partner writer does the same into theirs.
10. **Progress log:** create `/tmp/fix-report-<a0|a1|a2>.md` **in your first 10 minutes** and update
    it after every few findings. It lists: applied ids (in plan-row order) · skipped ids with a
    reason · recomputed vectors (script + output) · cross-references to verify · any place where
    this plan was ambiguous and what you chose. Never leave it empty at the end.
11. **Order of work:** apply your §1 PO rows first (they are quoted by other rows), then the rows in
    the order printed, then the digest/vector rows last (A1: §6.1/§6.3; A0: §6.2; A2: §6.4 — A2 must
    do the renames and the `attrs` shape before the vectors).
12. **Stop conditions:** if a row would require an architectural decision, a new dependency, or a
    change this plan does not authorize, do not improvise — write the question into your progress
    report under "BLOCKED" and continue with the remaining rows.

## NOT TRIAGED

Nothing. All 63 MUST FIX ids from both reports are assigned exactly once (§1: 2 — C-01, S-07; §2
A0: 13; §3 A1: 30; §4 A2: 18), 90 of the 99 SHOULD/NICE ids are assigned too, all amendment
requests AM-1…AM-4 are ruled, and every unassigned id (P-77, C-09, E-06, E-08, F-05, S-12, S-14,
T-05, T-08) is listed in §7 with a reason and an additive-safety verdict.

Two items are flagged for the integrator rather than a writer (§7): the two
`contracts/README.md` sentences (P-68 tie-break, E-06 safety-pairing gate) and the WP-00 status
flip + session record.
