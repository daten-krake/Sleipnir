# A0/A1/A2 contract review — principal build engineer

Run: **READ-ONLY** review. Repo `/home/wtadmin/Sleipnir`, branch `contracts/a0-a2-conventions`.
Nothing under the repo was created, edited, moved or deleted; no git write command was run
(`git branch --show-current` only). This file is the deliverable.

Inputs read in full: `contracts/A0-conventions.md` (773 L), `contracts/A1-events.md` (2472 L),
`contracts/A2-graph.md` (1200 L), `contracts/README.md`, `SPEC.md` §2 C1–C11, `DESIGN.md`,
`AGENTS.md`, `WORKFLOW.md` §5, `sessions/2026-09-04-program-layout.md` (Q1–Q15),
ADR-0016 §1–§2, ADR-0012 §2, ADR-0019 §2. Not re-done: the A1 §4.3 chain vector
(parent-verified).

Counts: **105 findings — MUST FIX 35 · SHOULD FIX 62 · NICE 8** (ids P-01…P-105; P-01…P-79
are headed findings in §1–§2, P-80…P-105 are the test-coverage findings of §4.1–§4.2; the §3
rulings are not findings). Repository untouched: `git status --porcelain` is empty.

Freeze blockers (must land in the Freeze PR): P-01, P-04, P-05, P-06, P-07, P-08, P-11, P-12,
P-40, P-41, P-42, P-43, P-44, P-45, P-46, P-47, P-49, P-60, P-62, P-63, P-64, P-65, P-66, P-78,
P-80, P-82, P-97 + amendment AM-1 (`usr_`).

Three most serious: **P-46** (the secret-scan rule table both A1 and A2 depend on does not
exist), **P-42** (`confidence: verified` is unreachable — A2's dedup swallows every
re-observation), **P-01** (`actor.component` is digest-critical and A1-3.3's notation
contradicts A1-2.1).

---

## 1. Buildability

Question: can a junior implementer, exercising no architectural judgment, build `internal/ids`,
`internal/cjson`, `internal/events`, `internal/graph` from these three documents alone?

**Verdict: no.** 31 MUST FIX gaps. They cluster into five shapes: (a) digest-critical facts
stated in prose notation instead of a normative table (P-01, P-25, P-41); (b) a mandate whose
implementing mechanism does not exist in the sketched API (P-03, P-11, P-43, P-45); (c) a rule
delegated to a document that never states it (P-46, P-47, P-48); (d) Go-level facts the
documents get wrong or omit, where the stdlib silently does the opposite of the clause (P-04,
P-05, P-06, P-12, P-36, P-43); (e) a state field a clause requires but no type carries (P-08,
P-13, P-32). None is architectural; all are one-to-five-sentence fixes, and all are cheaper
before Freeze than after (A0-2.13, A0-6.5, A1-4.10).

### A0 / A1

#### P-01 · MUST FIX · A1-3.3 (all nine tables) + A1-2.1/A1-2.2
The "Composed by" column (`user · scope`, `orchestrator · spawn_broker`, `user or platform ·
graph`, `requesting principal · varying`) reads as `(actor.type, actor.component)`, but A1-2.1
says `component` is `""` unless `type="platform"`. `actor` is inside the digest (A1-5.2), so an
implementer who stamps `component:"runtime"` on a user-composed `run_started` produces a
different `hash` than one who stamps `""`, and every such row later verifies as
`preimage_mismatch`. `action_blocked`'s "varying" is undefined outright.
**Wording:** rename the column to `actor (type, component)` and give the literal pair per row —
`run_started` → `(user, "")`; `chain_genesis` → `(platform, "event_store")`;
`graph_node_quarantined` → `(user, "")` **or** `(platform, "graph")`; `action_blocked` →
`(requesting principal's type, "")`. Add: "The subsystem named in the Source column is the code
path that writes the row; it appears in `actor.component` only when `actor.type="platform"`
(A1-2.1). For every non-platform row `actor.component` is `""`." Delete "varying".

#### P-02 · MUST FIX · A0 §4 (the three unnamed sketch blocks)
The `Page`/`Cursor`/`EncodeCursor`/`DecodeCursor` block, the `Clock`/`Now`/`FormatTime`/
`ParseTime` block and the caps/`Truncate`/`Fits` block have **no `package` line and no home**.
A0 §4 names only `internal/ids`, `internal/cjson`, `internal/errs`; DESIGN §1 lists only `errs`
and `logging` as foundation. No work package can list "files to touch".
**Wording:** "A0 requires five foundation packages (zero internal imports except `errs`):
`internal/ids` (A0-1), `internal/cjson` (A0-2), `internal/errs` (A0-3), `internal/paging`
(A0-4: `Page`, `Cursor`, `EncodeCursor`, `DecodeCursor`), `internal/timex` (A0-5: `Clock`,
`Now`, `FormatTime`, `ParseTime`; the name avoids shadowing stdlib `time`, the same reasoning
DESIGN §1 gives for `logging`), `internal/caps` (A0-7: every cap constant of the A0-7.1
registry, `TruncationMarker`, `Truncate`, `Fits`)."

#### P-03 · MUST FIX · A0-4.4 + A0-4.8 + A0 §4 `DecodeCursor`
A0-4.8 requires rejecting a cursor whose `id` is "of the wrong type per A0-1", but
`DecodeCursor(s string)` has no parameter from which the expected type could be known — the
clause assigns a check to a function with no basis to perform it. A0-4.3 also permits a **text**
ordering key while the sketch hard-codes `K int64`, so a text-ordered collection cannot encode a
cursor at all.
**Wording:** `func DecodeCursor(s string, k ids.Kind) (Cursor, error)` — "the caller passes the
id kind of the collection being paged; `DecodeCursor` validates `id` against it (A0-1.5) and
returns `validation` on mismatch." Plus: "`k` is an integer for every v1 collection (A1: `seq`;
A2: `seq`). A collection whose order key is text MUST declare its own cursor shape in its own
contract; A0-4.4's two-key set is closed."

#### P-04 · MUST FIX · A0-2.7 + A0 §4 `CanonicalValue` (missing vector)
`CanonicalValue` marshals with `encoding/json`, which HTML-escapes `<`, `>`, `&` (and U+2028/9)
by default; A0-2.7 forbids HTML escaping in the output. No clause says `Canonical` **decodes**
input escapes and re-emits per A0-2.7, and no vector covers escaped input — yet A1's only digest
path (`Preimage` → `CanonicalValue`) always goes through `json.Marshal`, and `<` is common in
tool output. Two implementations, two chains.
**Wording (A0-2.7):** "`Canonical` MUST decode every JSON string escape in the input (`\uXXXX`
including well-formed surrogate pairs, `\n`, `\"`, `\\`) and re-emit the value per this clause;
it MUST NOT pass an input escape sequence through. `CanonicalValue` MUST NOT rely on
`encoding/json`'s escaping decisions: the canonicalizer is the only authority on the emitted
form." **Add vector V7:** input `{"s":"<a href=\"x\">&é\u2028"}` → canonical
`{"s":"<a href=\"x\">&é\u2028"}` (U+2028 literal, 3 B) with length and SHA-256.

#### P-05 · MUST FIX · A0-2.3 / A0-8.9 (nobody is told to check UTF-8)
Both clauses require rejecting invalid UTF-8, but the stdlib path silently coerces:
`json.Decoder` replaces invalid bytes in a string with U+FFFD and `json.Marshal` does the same on
output. The default implementation never rejects and the A0-2.3 test can never fail.
**Wording (A0-8.9):** "The UTF-8 check is explicit: the handler MUST call `utf8.Valid` on the raw
body before decoding and reject with `validation`; `cjson.Canonical` MUST call `utf8.Valid(doc)`
and return `validation` (A0-2.3). `encoding/json` substitutes U+FFFD instead of failing, so
relying on the decoder is a defect."

#### P-06 · MUST FIX · A0-2.14 / A0-8.3 vs Go marshaling (nil → `null`)
A0-2.14 requires `[]`/`{}` and forbids `null`, but `json.Marshal` of a nil slice/map/pointer
emits `null` — the easiest way to break every event digest (A1 §4.1 `EvidenceRefs []string`).
It is also unstated whether `cjson.Canonical` accepts or rejects a `null` it is given (A0-2.8
lists `null` as a legal literal; A0-8.3 forbids it in contract JSON).
**Wording (A0-2.14):** "Constructors MUST initialize every collection field to a non-nil empty
value. `CanonicalValue` MUST reject a `null` at any depth with `validation` (contract documents
contain no `null`, A0-8.3); the byte-level `Canonical` entry point MAY accept `null` for
non-contract callers and emits the literal `null` (A0-2.8)." Add a rejection vector `{"a":null}`
and `TestNilCollectionNeverSerializesAsNull`.

#### P-07 · MUST FIX · A1-7.3 (`evidence_refs` derivation) vs A1-3.3 `evidence_stored`
`evidence_refs` is the union of "every non-empty payload field whose name ends in
`_evidence_id`", but `evidence_stored.evidence_id` does not match that suffix — so the one kind
whose whole purpose is an artifact reference yields `evidence_refs: []`, and A1-8.3's
`evidence_ref` filter can never find an `evidence_stored` event. `evidence_refs` is inside the
digest: unfixable after Freeze.
**Wording:** rename the payload field to `stored_evidence_id` (pre-Freeze, zero cost, no vector
changes — §4.3 row 2 is `graph_node_quarantined`), **or** amend A1-7.3 to "…plus
`evidence_stored.evidence_id`". Either way replace derivation-by-naming with a normative
per-kind list: each A1-3.3 row declares its evidence-reference fields.

#### P-08 · MUST FIX · A1-7.6 (platform-composed dedup keys)
The key is `(engagement_id, actor.principal_id, kind, idempotency_key)` behind a uniqueness
constraint, and platform-composed events get only "a deterministic key … (e.g. …)". With
`principal_id=""` for `type="platform"` and no key rule, the **second** `llm_call` (or
`graph_node_written`, `chain_verified`, `container_killed`, …) in an engagement collides with the
first and the append fails.
**Wording:** "The A1-7.6 uniqueness constraint applies to rows appended through `events:append`
only. A platform-composed row MUST carry a non-empty deterministic key: `approval_*` →
`approval_id` + decision · `evidence_stored` → `evidence_id` · `job_spawned`/`task_spawned`/
`container_started` → `spawn_request_event_id` · `graph_*` → the written `gn_`/`ge_` id ·
`chain_verified` → `trigger` + `head_seq` · `chain_break_detected` → `break_seq` + `break_kind` ·
`cleanup_*` → `revert_event_id` · `notification_sent` → `related_event_id` + `attempt` ·
`llm_call` → the gateway's per-call id · otherwise the composing subsystem's operation id.
`""` MUST NOT be used."

#### P-09 · MUST FIX · A1-7.11 / A1-7.5 vs A0-3.11 (retrying a terminal kind)
A1-7.11 tells the caller to retry a failed append "with the same `idempotency_key`" for `timeout`
**and** `internal`; A0-3.11 declares every kind but `timeout`/`upstream`/`rate_limited`
terminal. A1-7.5 also maps "chain not verified at startup" — a normal transient boot state — to
`internal` (500), which A0-3.1 defines as a platform defect.
**Wording:** A1-7.5 — "an append refused because the chain's startup verification has not
completed returns `timeout` (504, retryable per A0-3.11); `internal` is reserved for the
no-valid-genesis case." A0-3.11 — "an owning contract MAY declare one specific write retryable
under `internal` where it defines a deduplication key (A1-7.6); the retry MUST reuse that key."

#### P-10 · MUST FIX · A1-7.10 vs DESIGN §1 / A1 §4.1 `NewEvent`
A1-7.10 mandates one pass of 14 steps including (1) token binding, (6) reference resolution,
(11) secret scan and (14) dedup — all I/O — while `NewEvent` is documented as running "the
A1-7.10 pass" inside a domain package with no I/O. The validator's package boundary is undefined,
so no work package can say what `internal/events` owns.
**Wording (A1-7.10):** "Steps 3, 5, 7, 8, 9, 10, 12, 13 are pure and MUST run in
`internal/events` (`NewEvent` / `Payload.Validate`). Steps 1, 2, 6, 11, 14 need I/O or request
context and run in the caller (api / policy / store layer) **in the same numbered order** around
the pure part. The order is normative; the split is an implementation fact recorded here so work
packages are disjoint." Fix `NewEvent`'s doc comment to name the pure subset.

#### P-11 · MUST FIX · A1-1.8 / A1-5.7 vs A0 §4 `cjson` (no API to add keys)
The served form is "decode the stored preimage, add the three chain fields, re-canonicalize", but
`cjson.Canonical(doc, exclude...)` can only remove keys, and decoding into `Event` needs the
per-kind payload type (A1-4.1) that the serve path does not have.
**Wording (A0 §4):** add
`// With adds top-level fields to an already-canonical document and re-canonicalizes; it never
// decodes into a typed struct. A key already present in doc is an error (A1-1.8, A1-5.7).
func With(doc []byte, add map[string]any) ([]byte, error)`
and in A1-5.7: "`Served(preimage, c)` = `cjson.With(preimage, map[string]any{\"seq\": c.Seq,
\"prev_hash\": c.PrevHash, \"hash\": c.Hash})`, operating on the generic document only."

#### P-12 · MUST FIX · A1-4.12 / A1 §4.1 (Go type of `timestamp` and `int` payload fields)
A1-3.3 declares `expires_at:timestamp` but never the Go type; the envelope uses `string`. A
junior will use `time.Time`, whose `MarshalJSON` emits RFC 3339 with *variable* fractional digits
(`T14:03:22Z` for a whole second) — violating A0-5.1 and silently changing digests.
**Wording (A1-4.12):** "A `timestamp` field is a Go `string` holding an A0-5.1 value, in the
envelope and in every payload. `time.Time` MUST NOT appear in any canonicalized type: its
`MarshalJSON` drops trailing zeros. Transcription rule for `:int` — `int64` for every `*_ms`,
`*_bytes`, count and `seq` field; `int` only where A1-4.2 declares a range (`exit_code`)."

#### P-13 · MUST FIX · A1-5.4 / A1-5.6 (`prev_recorded_at` has no home)
The forward clamp `max(clock_now, prev_recorded_at)` is normative, but `ChainHead` carries no
last-`recorded_at` field and the clause does not say the predecessor row is read.
**Wording:** add `LastRecordedAt string` to `ChainHead` (A1-5.6: mutable bookkeeping, never
hashed, updated in the append transaction) and to A1-5.4: "the clamp compares A0-5.1 strings
byte-wise — for a fixed-format UTC millisecond timestamp, byte order equals time order."

#### P-14 · MUST FIX · A1-6.6 vs A1-6.2 (which head an export carries)
A1-6.6 says `head_seq`/`head_hash` are "the chain head at release time", but the `pre_export`
walk appends a `chain_verified` event *before* release and the §4.2 example shows the walk-end
head (`seq` 4313) while `chain_verified_event_id` names a later row. Two readings of a
customer-facing legal artifact.
**Wording (A1-6.6):** "`head_seq`/`head_hash` are the values of the **last row the walk
verified** — identical to the `head_seq`/`head_hash` payload of the event named by
`chain_verified_event_id` — not the head after that event was appended; `head_seq` therefore
equals the post-append `ChainHead.HeadSeq` minus one."

#### P-15 · MUST FIX · A1-6.6 vs A1-5.6 (two vocabularies named `integrity_state`)
The export block's `integrity_state` is `verified | failed_overridden`; the chain-head
`IntegrityState` is `unverified | verified | failed | overridden`. One field name, two closed
lists, no mapping — a client switching on it mis-handles three of four values.
**Wording (A1-6.6):** "The export `integrity_state` is a **distinct** enum
(`verified | failed_overridden`) derived from A1-5.6's: `unverified`/`failed` → not exportable
(A1-6.4); `verified` → `verified`; `overridden` → `failed_overridden`. The two lists MUST NOT be
used interchangeably."

#### P-16 · MUST FIX · A1-8.9 vs A1-5.6 (chain head "never served" yet "available to any reader")
A1-5.6 says the chain-head state "is never served"; A1-8.9 guarantees `(head_seq, head_hash)` and
`integrity_state` "are available to any authorized reader and to every export"; A1-6.4's
"internal views are flagged" depends on the served half.
**Wording (A1-5.6):** "…it MUST NOT appear in an event document (A1-1.6) and is never hashed. The
row is internal; the projection `(head_seq, head_hash, integrity_state, verified_at)` is served to
authorized readers with a wire shape A4 owns (A1-8.9, A1-6.4)."

#### P-17 · MUST FIX · A1-8.1 (an unenforceable MUST about cursor direction)
"A cursor is only valid in the direction it was issued for" cannot be enforced: A0-4.4's cursor
key set is closed at `{id, k}` and carries no direction, so the platform cannot distinguish an
ascending cursor from a descending one.
**Wording:** either delete it and state "direction is a request parameter; a cursor issued for the
other direction is interpreted in the requested direction and yields a well-defined (possibly
empty) page", or amend A0-4.4 to add `"d":1|-1` — a breaking change to the cursor shape that must
be decided **before** Freeze.

#### P-18 · SHOULD FIX · A0-1.1 (128 bits in 26 base32 characters)
26 Crockford characters carry 130 bits; the clause never says the leading character is restricted
(ULID restricts it to `0`–`7`), so a generator can be written two ways and the accepted-id set is
undefined (A0-1.5's regex accepts a leading `z`).
**Wording:** "The 48-bit millisecond value is encoded big-endian in 10 characters, whose first is
therefore in `0`–`7`; the 80 random bits are encoded in 16 characters with no restriction.
Validation (A0-1.5) is the regex only: a body whose first character is `8`–`z` is accepted but
never generated."

#### P-19 · SHOULD FIX · A0-2.11 (depth and size semantics)
"nesting depth ≤ 32" never says what counts as a level (objects and arrays? top level = 1?), and
"size ≤ 1 MiB" never says input or output.
**Wording:** "Depth counts container boundaries: the top-level object is depth 1, each nested
object or array adds 1, scalars are not levels. Size is `len(doc)` of the input. Both are checked
before canonicalization."

#### P-20 · SHOULD FIX · A0-2.17 V6 (the "exact" bytes column shows an escape)
The column is labelled exact but prints the key as `"\uFFFD"`, which A0-2.7 forbids in output; the
`len` 18 only works for literal U+FFFD (`EF BF BD`). A test-writer transcribing the cell writes
the wrong bytes.
**Wording:** print the literal character and add "(the U+FFFD key is the three bytes
`EF BF BD`; the escape appears only in the input column)". Better: add a hex-dump column for V3
and V6.

#### P-21 · SHOULD FIX · A0-3.10 vs A0 §4 (`retry_after_ms` with `omitempty`)
The envelope MUST carry `retry_after_ms` for `rate_limited`, but the sketch tags it `omitempty`,
so a limiter emitting 0 drops the key and violates A0-3.10.
**Wording:** "A rate-limit response MUST carry `retry_after_ms ≥ 1` and `Retry-After ≥ 1`; the
`omitempty` tag in the §4 sketch is wrong for this field and MUST be replaced by an explicit
presence rule (custom `MarshalJSON`) or by the `≥ 1` floor."

#### P-22 · SHOULD FIX · A0-4.5 ("absent-and-unparseable")
The phrase contradicts the same clause's "default 100" for an absent `limit`.
**Wording:** "`limit` absent → 100. `limit` present but unparseable, ≤ 0, non-integer, or > 1000
→ `validation`, never clamped."

#### P-23 · SHOULD FIX · A0-5.3 (the recommended layout accepts offsets)
`2006-01-02T15:04:05.000Z07:00` **parses** `+02:00`; only formatting is UTC-safe. A junior
following the implementation note accepts the offsets A0-5.3 forbids.
**Wording:** "Parsing is three checks in order: (1) the byte-exact A0-5.1 regex — this is what
forces `Z` and exactly three digits, (2) `time.Parse(TimeLayout, s)`, (3) the year window
`[MinYear, MaxYear)`. `time.Parse` alone is insufficient: the layout accepts numeric offsets."

#### P-24 · SHOULD FIX · A0-8.7 vs A1-3.3 `image_digest`
A0-8.7 says digests are lowercase hex, 64 characters for SHA-256; `image_digest` is an OCI digest
(`sha256:<64hex>`, 71 B) capped at 256 B by A1-4.5. As written a strict A0-8.7 validator rejects
every real registry digest.
**Wording (A0-8.7):** "…64 characters for SHA-256. Container image digests are the exception:
`image_digest` carries the registry's `<algorithm>:<hex>` form, validated as
`^[a-z0-9]+(?:[._-][a-z0-9]+)*:[0-9a-f]{64}$` and capped by the owning contract (A1-4.5: 256 B)."

#### P-25 · SHOULD FIX · A1-4.4 (the `*` set is prose, but digest-critical)
`untrusted` is computed from the `*`-marked fields and is inside the digest, yet the marking
exists only as a glyph in the A1-3.3 tables and as comments in §4.1. Two transcribers will not
produce the same set, and the shared suite has nothing to check per kind.
**Wording:** add to A1 §4.1 —
`// UntrustedFields returns k's A1-4.4 "*" field names in canonical order. Normative: transcribed
// from A1-3.3, part of the digest definition (A1-5.2).
func UntrustedFields(k Kind) []string`
plus a kind → `*`-fields table in A1-4.4 and `TestUntrustedFlagMatchesStarredFields` over the
P-26 corpus.

#### P-26 · SHOULD FIX · A1-4.5 (the 32768 B bound has no runnable derivation)
"6 × the largest decoded prose cap plus envelope overhead" is a hand-wave, and the mandated test
("a maximal payload for every kind fits") has no definition of *maximal*.
**Wording:** "A kind's maximal payload sets every string field to exactly its cap length in bytes
of U+0001 (worst case `\u0001` = 6 canonical bytes per input byte, A0-2.7), every integer to its
declared maximum, every array to its count cap filled with maximum-length ids, every bool to
`true`. `TestMaximalPayloadFitsCanonicalBound` asserts `len(Preimage(e)) ≤
EventMaxCanonicalBytes` for all 39 kinds under that construction."

#### P-27 · SHOULD FIX · A1-4.2 (`container_killed{hard_stop}` needs a chain lookback)
"MUST be preceded in the chain by a `hard_stop_fired` event" implies an unbounded scan inside the
append path, which A1-7.12 forbids for kill paths, and no error kind is named for a violation.
**Wording:** "The check reads the run's hard-stop state held by `internal/policy`, not the chain;
a violation is a platform defect → `internal` (A0-3.1) and MUST NOT delay the kill (A1-7.12)."

#### P-28 · SHOULD FIX · A1-4.2 (fingerprint divergence: which error kind?)
"The platform MUST refuse to compose a decision event whose fingerprint diverged" — the A1-4.2
preamble defaults to `validation`, blaming a human approver for a platform-side inconsistency.
**Wording:** "…MUST refuse with `conflict` (409, A0-3.1: state prevents the request) naming
`approval_id` and the first 8 hex characters of both digests, and MUST record
`action_blocked{fingerprint_mismatch}` (A1-3.3)."

#### P-29 · SHOULD FIX · A1-4.11 vs ADR-0013 / A1-8.8 (offline nodes and `evi_` resolution)
A1-4.11 requires every `evi_` reference to resolve at composition (`notfound` otherwise) while
A1-8.8 makes a dangling `evi_` a benign archival state and Q9/ADR-0013 nodes buffer offline; the
required upload-then-append order is never stated.
**Wording (A1-4.11):** "A client MUST upload its artifacts (`evidence:upload`) before appending
the event that references them; an append whose `*_evidence_id` does not resolve in this
engagement is `notfound` (404) and MUST be retried by the buffering client after the upload
(ADR-0013). Write-time resolution is what makes A1-8.8's dangling reference an archival state,
never a normal one."

#### P-30 · SHOULD FIX · A1-4.12 (`error_kind` validated for two kinds of three)
`task_result` and `agent_error` require an A0-3.1 kind; `llm_call` requires only non-empty. Same
field name, three kinds, two rules.
**Wording:** "Every `error_kind` field (`task_result`, `agent_error`, `llm_call`) MUST be `""` or
a byte-exact A0-3.1 kind; the per-kind non-empty obligations of A1-4.2 are additional. An unknown
value is `validation` (A0-6.3)."

#### P-31 · SHOULD FIX · A1-6.2 (startup-window semantics)
"MUST NOT block startup of the platform process for engagements already verified in this boot" is
not a rule, and read behaviour during the window is undefined.
**Wording:** "Startup verification runs asynchronously, one goroutine per engagement chain, owned
and cancellable per DESIGN §6. While a walk is incomplete `integrity_state` is `unverified`:
reads are served and flagged (A1-6.4), appends return `timeout` (P-09), exports are refused with
`integrity_failed`. Other engagements are unaffected."

#### P-32 · SHOULD FIX · A1-6.3 (break-dedup state has no home)
"MUST NOT append a duplicate `chain_break_detected` for the same `(break_seq, break_kind)`" needs
remembered state; `ChainHead` has none.
**Wording:** add `LastBreakSeq int64`, `LastBreakKind BreakKind`, `LastBreakEventID string` to
`ChainHead` (A1-5.6), updated in the same transaction as the break event.

#### P-33 · SHOULD FIX · A1-7.6 (the dedup payload hash preimage is undefined)
"the SHA-256 of the accepted canonical payload … covers `kind` + `payload` only" never names the
preimage, so two implementations hash different bytes and dedup silently stops working
(duplicate evidence rows). Whether A1-4.7 normalization happens before the hash is also unstated.
**Wording:** "`PayloadHash = cjson.SHA256Hex(cjson.CanonicalValue(struct{Kind Kind
`json:\"kind\"`; Payload Payload `json:\"payload\"`}{…}))`, computed **after** the A1-7.10
normalization steps, so a retry whose array order differs is a dedup hit. No other field
participates." Add a vector (kind + payload → 64 hex).

#### P-34 · SHOULD FIX · A0-3.1 `summary_too_large` scope vs A1-4.5/A1-4.7/A2-6.4
A0-3.1 scopes the kind to "a capped field exceeded its **Q4** budget", but A1 uses it for
A1-local caps and array **counts** and A2 for `attrs` counts — none are Q4 constants.
**Wording (A0-3.1):** "a capped field or count exceeded a budget declared by its owning contract
under mechanism R (A0-7.6) — the Q4 constants of A0-7.1 and the per-contract caps of A1-4.5 /
A2-7.1."

#### P-35 · SHOULD FIX · A0-7.5 / A0 §4 `Truncate`
`Truncate(s string, cap int)` shadows the builtin `cap` (AGENTS.md / DESIGN §9 review bar) and its
behaviour for `cap < len("[truncated]")` is undefined.
**Wording:** `func Truncate(s string, limit int) (out string, truncated bool)` plus "`limit <
len(TruncationMarker)` is a platform defect (no cap in the A0-7.1 registry is that small): return
`\"\", true` and let the caller surface `internal`."

#### P-36 · SHOULD FIX · A0-6.2 / A0-8.1 (case-insensitive key matching on writes)
`encoding/json` matches field names case-insensitively, so `DisallowUnknownFields` does **not**
reject `{"Kind":…}` or `{"KIND":…}`: the accepted-input set is wider than A0-8.1 implies and
A0-2.5's case-duplicate concern is only half enforced.
**Wording (A0-6.2):** "Write-side key matching is byte-exact: a key differing from the declared
`json` tag only by case is an unknown field and MUST be rejected with `validation` naming it.
Because `encoding/json` matches case-insensitively, the decoder MUST verify observed key spelling
(the same `Token()` walk A0-2.5 requires) rather than rely on `DisallowUnknownFields` alone."

#### P-37 · NICE · A1-5.8 (the 1000-`seq` constant is unnamed)
**Wording:** add `HeadLogIntervalSeq = 1000 // A1-5.8: emit the head hash to slog at every
crossing` to A1 §4.1 and cite it from A1-5.8.

#### P-38 · NICE · A1-4.7 (integer arrays) / A1-4.12 ("three fields" is five)
A1-4.7 allows "arrays of integers" but no A1 kind has one and the sort/dedup rule is stated for
strings only; A1-4.12 says "Three payload fields are not A1 enums" and lists five.
**Wording:** drop "or of integers" (or extend the rule to it); "Five payload fields across three
groups…".

#### P-39 · NICE · A0-4.4 example order / A0-8.2 `sha256`
A0-4.4 writes the cursor as `{"k":…,"id":…}` while the canonical form (and the §4 example) puts
`id` first — add "(canonical order puts `id` first, A0-2.4)". A0-8.2's suffix list has no entry
for `evidence_stored.sha256`, a digest that is not `*_hash`: add "`sha256` (the artifact-integrity
digest, ADR-0009 §2)" or rename the field.

### A2

#### P-40 · MUST FIX · A2-1.4 (+ A2-8.9, A2-5.4)
The graph `seq` has no assignment clause and no stated relationship to A1's `seq`: A1-5.4 fixes
assignment (inside the insert transaction, per-engagement lock, dense, no gaps) for events; A2
says only "platform-assigned, strictly increasing, immutable, unique". Unstated: which sequence,
shared between nodes and edges or per collection, dense or not, and how it differs from the A1
`seq` that the same ingest path reads (A2-5.4). Both are named `seq` in JSON, and A3 stage views
will contain both.
**Wording:** "A2-1.4a: `seq` is assigned by the platform inside the transaction that inserts the
row, from **one per-engagement graph sequence shared by nodes and edges**, strictly increasing by
1, dense, and independent of the A1 event `seq` (A1-5.4). To keep the two apart in code, logs and
views the graph field is named `graph_seq` in Go fields, store columns **and JSON** (pre-Freeze,
zero cost); A2-8.9's `(graph_seq, id)` tuple and A0-4.4's `k` are that value."

#### P-41 · MUST FIX · A2-4.6 / A2-4.7 (array order inside the fingerprint)
`content_hash` covers `addresses` and `evidence_ids`, but nothing requires them sorted or
deduplicated (A1-4.7 does for events). The same content submitted in a different order yields a
different fingerprint, so the node dedup of A2-4.7 fails and the graph fills with duplicate
evidence nodes.
**Wording (A2-4.6):** "`addresses` and `evidence_ids` MUST be sorted ascending by unsigned byte
value and deduplicated before the content document is canonicalized (A1-4.7's rule), so two
observations of the same content in a different input order produce the same `content_hash`.
`attrs` keys are ordered by A0-2.4."

#### P-42 · MUST FIX · A2-5.6 vs A2-4.6/A2-4.7 (`confidence: verified` is unreachable)
A2-5.6 defines `verified` as "reproduced by a second, independent observation" and forbids raising
confidence by a later write (revision only). But a re-observation with identical content collapses
into the existing node (A2-4.7), and a revision with identical content has an identical
`content_hash` — because A2-4.6 deliberately excludes provenance — so it collapses too. No
sequence of writes can ever store `verified`, yet both §4.1 examples do.
**Wording (A2-4.7):** "The node dedup key is `(engagement_id, kind, content_hash, confidence)`. A
re-observation whose content is identical but whose evidence grade is higher (`inferred` →
`verified`, A2-5.6) MUST create a new node with a `supersedes` edge to the previous one; that is
the only path by which `verified` is stored, and it is why `confidence` participates in dedup
although it is not part of `content_hash` (A2-4.6)."

#### P-43 · MUST FIX · A2-10.6 + A2 §4 `Node` (unimplementable as written)
"their fields are unexported so no package can build an unvalidated node literal" contradicts the
sketch's exported fields and `json` tags: `encoding/json` cannot marshal unexported fields, and Go
cannot prevent `graph.Node{Label:"x"}` from another package when the fields are exported. The
claimed "`go vet`-visible exported-field audit" does not exist in `go vet`.
**Wording:** keep the fields exported with the sketched tags, delete the unexported claim and the
`go vet` sentence, and enforce with: "(a) `NewNode`/`NewEdge` are the only documented construction
path; (b) the store seam re-validates every value it is given (A2-10.7); (c)
`TestNodeHasNoExportedContentSetter` — reflection over `graph.Node`/`graph.Edge` asserts no
exported method mutates a content field (the three A2-8/A2-3.9 flag setters and P-54's
`SetSupersededBy` are the declared exceptions); (d) review per AGENTS.md. If the PO prefers
unexported fields, A2 MUST specify `MarshalJSON`/`UnmarshalJSON` for `Node` and `Edge`."

#### P-44 · MUST FIX · A2 §4 `NodeDraft` (the write contract is a comment)
`type NodeDraft struct{ /* Node minus platform-set fields */ }` — the draft is what the ingest
path builds and what every A2 acceptance test constructs; its field set is undefined (does it
carry `supersedes_id`? `attrs`? `status`?).
**Wording:** spell it out:
```go
type NodeDraft struct {
	Kind NodeKind
	Label, Summary string
	Attrs Attrs
	EvidenceIDs []string
	Addresses []string
	CIDR string
	Port int
	Transport, Protocol, SID, Domain string
	CredentialKind CredentialKind
	EvidenceID string
	MediaKind MediaKind
	SizeBytes int64
	Severity Severity
	Claim, Basis string
	Status string
	SupersedesID string // A2-4.2: set by the revising write, not by the platform
}
```
and state: "`ID`, `EngagementID`, `Seq`, `ContentHash`, `Quarantined`, `QuarantineReason`,
`ReportExcluded`, `SupersededByID` and `Provenance` are platform-set and absent from the draft; a
request body carrying one is an unknown field on a write (A0-6.2, A2-1.5, A2-5.2)."

#### P-45 · MUST FIX · A2 §4 `NewNode` returns a half-built `Node`
`NewNode(in NodeDraft, prov Provenance, q QuarantineState) (Node, error)` cannot set `ID` or `Seq`
(assigned at insert, A2-1.2/A2-1.4), so it hands out a half-built value (DESIGN §4 forbids it)
while A2-10.6 requires the store seam to accept only constructor output — the seam contract is
incoherent.
**Wording:** "`NewNode` returns `PendingNode` (validated content + provenance + quarantine state +
`content_hash`; no id, no seq). The store seam's
`WriteNode(ctx context.Context, engagementID string, n PendingNode) (Node, error)` assigns `id`
(A0-1.4) and `seq` (A2-1.4) inside the insert transaction, applies A2-3.4/A2-4.7 dedup, and
returns the complete `Node`. A2-10.6 applies to `PendingNode` and `Node` alike: the seam accepts
no other input." Same shape for `NewEdge`/`WriteEdge`.

#### P-46 · MUST FIX · A2-9.4 (the secret-scan rule table both contracts depend on does not exist)
A1-4.9 delegates: "the platform MUST run the stdlib pattern scan **A2-9.4 defines**". A2-9.4
defines five prose categories — no rule ids, no patterns, no entropy definition — while A1-4.9 and
A2-9.5 both require the error to name "the field and the **rule id**", and the
`contracts/README.md` merge gate ("secret-free serialization") plus A1-4.9/A2-9.7's six named
tests need a planted-secret corpus. This is the AGENTS.md high-review area with the largest
undefined surface in all three documents, and it blocks A1's and A2's Freeze simultaneously.
**Wording:** A2-9.4 MUST publish a normative, closed, additive-only (A0-6.5) rule table
`rule id | Go regexp | fields scanned | note`, at minimum: `SEC-PEM`
(`-----BEGIN [A-Z ]*PRIVATE KEY-----`) · `SEC-NTLM` (32-hex NTLM/LM shapes) · `SEC-KRB`
(`krbtgt` ticket material) · `SEC-AWSKEY` (`(AKIA|ASIA)[0-9A-Z]{16}`) · `SEC-GCPKEY`,
`SEC-AZUREKEY` · `SEC-JWT` (`eyJ[0-9A-Za-z_-]+\.[0-9A-Za-z_-]+\.[0-9A-Za-z_-]+`) · `SEC-BEARER`
(`(?i)(bearer|token|api[_-]?key|password|passwd|secret)\s*[:=]\s*\S{8,}`) · `SEC-URLCRED`
(`[a-z][a-z0-9+.-]*://[^/\s:@]{1,64}:[^/\s:@]{1,64}@`) · `SEC-ENTROPY` (Shannon entropy ≥ 4.5
bits/char over a window of ≥ 32 characters drawn from a base64/hex alphabet; the formula and the
window are part of the rule). Plus: "The corpus planted by `TestEventSecretFreeSerialization` and
`TestGraphSecretFreeSerialization` is exactly one value per rule id and lives in the shared
suite; a rule added later adds a corpus entry."

#### P-47 · MUST FIX · A2-8.2 (which fields are matched, and whether quarantine propagates)
The clause lists `label, addresses, cidr, domain, sid` for every kind, but `service`, `share`,
`credential`, `evidence_ref`, `finding` and `hypothesis` carry at most `domain`/`sid` — a
`service` node listening on a blacklisted `host` is **not** quarantined by A2-8.2, and A2-8.4
propagates only to *edges*, never to nodes. This is the platform's core safety path (SPEC C5,
ADR-0016 §2) and it is the one place where an implementer's guess decides whether an
out-of-scope target is actionable.
**Wording:** add a per-kind matched-field table (`host`: `label` + `addresses` · `network`:
`label` + `cidr` + `addresses` · `identity`/`group`: `label` + `domain` + `sid` · `credential`:
`label` + `domain` · `share`: `label` + `domain` · `service`: `label` + `protocol` ·
`evidence_ref`/`finding`/`hypothesis`: `label` only) and: "Quarantine propagates **one hop** along
`reachable`, `authenticates_to` and `grants_access` from a quarantined `host`/`network` to the
`service`/`share` attached to it, at ingest and on every recomputation (A1
`quarantine_recomputed`); it is never transitive beyond one hop. `attrs`, `summary`, `claim` and
`basis` are MUST NOT be matched — they are prose (A2-6.7)."

#### P-48 · SHOULD FIX · A2-8.2 / A2-10.2 step 13 (the policy seam has no shape)
A2 defers matching to `internal/policy` but never names the interface it consumes, so WP-graph
cannot be built against a stub (DESIGN §4: interfaces are defined at the consumer).
**Wording:** "`internal/graph` declares and consumes exactly one interface:
`type QuarantineDecider interface { Classify(ctx context.Context, engagementID string, n NodeDraft) (QuarantineState, error) }`.
`internal/policy` provides the implementation; `internal/graph` never imports `internal/policy`
(DESIGN §1 layering)."

#### P-49 · MUST FIX · A2-6.3 (the reserved-key set is open-ended)
"A key MUST NOT equal or shadow a schema field name of the node's kind or of A2-2.2 (`label`,
`summary`, `kind`, `port`, `severity`, `evidence_ids`, **…**)" — a hard-reject rule with an
ellipsised set cannot be implemented or tested without judgment.
**Wording:** publish the closed list as a normative constant:
`ReservedAttrKeys = {id, engagement_id, seq, graph_seq, kind, label, summary, attrs, evidence_id,
evidence_ids, addresses, cidr, port, transport, protocol, sid, domain, credential_kind,
media_kind, size_bytes, severity, claim, basis, status, content_hash, quarantined,
quarantine_reason, report_excluded, supersedes_id, superseded_by_id, provenance, source_id,
target_id, source_kind, target_kind, retracted, principal_kind, run_id, job_id, task_id,
agent_node_id, operator_id, tool_id, tool_version, event_id, recorded_at, observed_claimed_at,
confidence}` and require a byte-exact membership test (no prefix or substring matching).

#### P-50 · SHOULD FIX · A2-6.4 (`AttrsTotalMaxBytes` measurement form undefined)
"≤ 4096 B for the serialized `attrs` object" — serialized how? A0-7.3's document rule counts the
transmitted bytes with the response encoder settings, but `attrs` is a sub-object and the node
read is not canonical.
**Wording:** "The 4096 B measurement is over the **A0-2 canonical form** of the `attrs` object —
the same bytes that enter `content_hash` (A2-4.6) — so the cap check and the fingerprint can never
disagree."

#### P-51 · SHOULD FIX · A2 §4 `AttrValue` (JSON form undefined)
`AttrValue` is a struct with `Type/Str/Num/Bool`; A2-6.1 and the §4.1 examples require a bare
scalar (`"cvss_v3_x10": 88`). Nothing says `Type` is not serialized.
**Wording:** "`AttrValue.MarshalJSON` emits the bare scalar of the live field (`\"s\"` / `1` /
`true`); `Type` is never serialized. `UnmarshalJSON` accepts a JSON string, integer or boolean
only and rejects `null`, floats, arrays and objects with `validation` (A2-6.1). Round-trip MUST
preserve `Type` (`TestAttrValueRoundTrip`)."

#### P-52 · SHOULD FIX · A2-3.5 vs A2-3.6 (`contradicts` symmetry is not expressible as a constraint)
The uniqueness constraint is on `(engagement_id, kind, source_id, target_id)`, so two concurrent
writes of the inverse pair create both directions; A2-3.5 claims the constraint makes races
impossible, and A2-3.6 leaves "which direction is stored" to platform choice at first write
(implementer judgment, non-convergent on replay).
**Wording (A2-3.6):** "For `contradicts` the platform normalizes the stored direction to
`source_id < target_id` byte-wise (A0-1.9) at composition, which makes A2-3.5's uniqueness
constraint sufficient and makes a replayed write converge; endpoint roles carry no meaning for
this kind."

#### P-53 · SHOULD FIX · A2-4.5 (`MaxSupersedeChain` has no write-side rule)
The 64-revision bound is stated only for reads ("past that the read paginates"), so nothing stops
the chain growing past it and the read bound is unenforceable where it matters.
**Wording:** "A `supersedes` write whose target chain already holds `MaxSupersedeChain` (64)
revisions is `conflict` (A0-3.1) naming the bound and the chain's first `gn_`; the bound is
enforced at write time, and A2-4.5's read bound is the consequence."

#### P-54 · SHOULD FIX · A2-4.2 / A2-8 / A2-10.6 (the three mutation paths are unnamed)
A2-1.3 declares content immutable, then A2-4.2 mutates `superseded_by_id` on an existing row and
A2-8 mutates `quarantined`, `quarantine_reason`, `report_excluded`, and A2-3.9 mutates
`retracted` — while A2-10.6 says the store accepts only constructor output. No mutation method is
named anywhere.
**Wording:** "The graph store seam exposes exactly four mutation methods, each taking an
engagement id (A2-11.1) and each returning `notfound`/`conflict` per A2-10.3:
`SetSupersededBy(ctx, engagementID, nodeID, newID)` · `SetQuarantine(ctx, engagementID, nodeID,
QuarantineState)` · `SetReportExcluded(ctx, engagementID, nodeID, bool)` ·
`SetEdgeRetracted(ctx, engagementID, edgeID, bool)`. No other update or delete method MAY exist
(A1-7.2's rule applied to the graph seam)."

#### P-55 · SHOULD FIX · A2-6.6 (an unowned MUST that contradicts the no-query posture)
"The platform MUST expose, per engagement and per node kind, the distinct `attrs` keys with their
occurrence counts" has no endpoint owner, no shape, no caller, and reads like the free-form query
Q1 excludes (A2 §2, A6 stub).
**Wording:** "The distinct `attrs` keys and their per-kind occurrence counts MUST be derivable
from stored rows (`attrs` stored per key, never as an opaque blob), ordered by key byte-wise
(A0-1.9). The operator-facing surface is the UI's (ADR-0011), not a `/api/v1` query endpoint
(Q1); A4 decides whether it exists in v1."

#### P-56 · SHOULD FIX · A2-2.3 / A2 §4 (`omitempty` conflates `size_bytes: 0` with absent)
A2-2.3 allows `size_bytes` int ≥ 0; A0-8.4 says 0 is a value, not "unset"; `omitempty` cannot
represent it.
**Wording:** "`size_bytes` MUST be ≥ 1 — a zero-byte artifact is not an artifact — so absence and
zero cannot be confused in the served representation (A0-8.3/8.4)."

#### P-57 · SHOULD FIX · A2-1.7 (the sentence says the opposite of A2-6.1)
"Floats, `null` and nested objects MUST NOT appear anywhere in a node or edge **except the flat
`attrs` map**" — but A2-6.1 forbids floats, `null` and nesting inside `attrs` too.
**Wording:** "Floats, `null` (A0-8.3) and nested objects MUST NOT appear in any node or edge
field. `attrs` (A2-6) is the only nested structure and is itself restricted to flat scalar
values: no float, no `null`, no array, no deeper object."

#### P-58 · SHOULD FIX · A2 §4 header comment + A2 header "Depends on"
The sketch calls `internal/graph` "Foundation layer" (DESIGN §1 puts it in Domain types; A1 §4.1
says "Domain layer"), and A2's header says "A0 (**frozen** for this document)" while A0's status
is `Draft` — per `contracts/README.md` no code may be written against a Draft.
**Wording:** "Domain layer (DESIGN §1); imports foundation only: `internal/ids`,
`internal/cjson`, `internal/errs`." and "Depends on A0 (Draft at the time of writing; every work
package below is gated on the A0/A1/A2 Freeze PR)."

#### P-59 · SHOULD FIX · A2-11.4 (`TestGraphNodeIDFromAIsNotFoundInB` has two oracles)
"whose `attrs.graph_node_id` is **either** absent **or** the requested id" is not a test
specification.
**Wording:** "…whose `attrs.graph_node_id` is the requested id (ids are not secrets, A0-1.7) and
whose `message` does not distinguish 'absent' from 'elsewhere' (A0-3.9)."

---

## 2. Cross-document consistency

#### P-60 · MUST FIX · A1-2.1 vs A2-5.3 (two actor vocabularies, one case missing)
A1 `actor.type` = `user | orchestrator | worker | platform | node`; A2 `principal_kind` =
`platform | orchestrator | worker | operator`. A2 has **no `node`** — a Q9/ADR-0013 remote-agent
observation cannot be attributed at all, although A2-5.3 provides `agent_node_id` for it — and
`operator` vs `user` are two names for one concept in two documents that describe the same write.
**Wording:** A2-5.3 adopts A1-2.1's list verbatim (`platform`, `orchestrator`, `worker`, `node`,
`user`), renames `operator` → `user` and `operator_id` → `user_id` (a `usr_` id per AM-1), and
publishes the identity mapping table A1 `actor.type` ↔ A2 `principal_kind`.

#### P-61 · SHOULD FIX · A2-5.2/A2-5.3 vs A1-3.3 (two attributions for one write, relationship unstated)
The anchoring event (`graph_node_written`) always has `actor = (platform, graph)` (A1-3.3), while
provenance must carry the originating worker/orchestrator. Nothing says so; the obvious
implementation copies `actor.type` into `principal_kind` and every agent attribution is lost.
**Wording:** A2-5.3 — "`principal_kind` is the principal whose **work produced the content**;
`provenance.event_id`'s event `actor` is the platform subsystem that wrote the row (A1-3.3). The
two are expected to differ, and copying one into the other is a defect." Mirror it in A1-3.3's
`graph_node_written` row.

#### P-62 · MUST FIX · A2-5.3/A2-7.1 `ToolVersionMaxBytes = 32` vs A1-4.5 `VersionMaxBytes = 64`
Same field name (`tool_version`), same registry value, two caps. A 40-character version string is
legal in an event and rejected by graph provenance, so ingest fails on data the audit log
accepted.
**Wording:** one constant (`ToolVersionMaxBytes = 64`) in the A0-7.1 registry (AM-2), cited by
A1-4.5 and A2-5.3; A2-7.1's 32 is a defect.

#### P-63 · MUST FIX · A2-8.1/A2-8.2/A2-2.8 vs A1-3.3 `graph_node_quarantined` (vocabularies and a missing operation)
A2 `quarantine_reason` = `out_of_scope | blacklisted`; A1 `quarantine_kind` =
`out_of_scope_discovery | blacklist_match | operator_quarantine | operator_release`, with no
mapping clause. Worse: A1-7.4 makes `graph_node_quarantined` user-composable with
`operator_quarantine`/`operator_release`, but A2-8.2/A2-2.8 forbid a caller setting quarantine and
A2 provides **no** operator quarantine or release operation; A2-8.5 makes `blacklisted`
unreleasable, so `operator_release` is only meaningful for `out_of_scope`.
**Wording:** (a) mapping table `out_of_scope ↔ out_of_scope_discovery`, `blacklisted ↔
blacklist_match`; (b) either add the operator operation to A2-8 ("an admin or assigned operator
MAY quarantine or release a node whose `quarantine_reason` is `out_of_scope`; releasing a
`blacklisted` node is `conflict` (A2-8.5); the write is `SetQuarantine` (P-54) and MUST be
accompanied by `graph_node_quarantined{operator_quarantine|operator_release}`") or delete the two
A1 enum values; (c) "a release is stored as `quarantined:false` with `quarantine_reason` absent;
the A1 event is the only record of the previous state."

#### P-64 · MUST FIX · A1-3.3 `graph_node_written` carries no `content_hash` (and cannot gain one after Freeze)
The A2-4.6 fingerprint is the only value that ties a graph row to a byte-exact content definition,
and the tamper-evident log never records it — so a silently rewritten graph row cannot be
cross-checked against the chain. A1-4.1/A1-4.10 make adding a field to an existing kind
**breaking** (`/api/v2` + ADR), so this must be decided now.
**Wording:** add `content_hash:64hex` to `graph_node_written` (edges have no fingerprint — say so
explicitly) and to A1-4.2: "`content_hash` MUST equal the A2-4.6 fingerprint of the written node."
No §4.3 vector change (row 2 is `graph_node_quarantined`).

#### P-65 · MUST FIX · A1-3.6 vs A2-1.6 / A2 §4 (two blessed spellings for one reference, no mapping)
A1 mandates `graph_node_id`, `graph_edge_id`, `from_graph_node_id`, `to_graph_node_id`,
`supersedes_graph_node_id` in event payloads; A2 mandates `id`, `source_id`, `target_id`,
`supersedes_id`, `superseded_by_id` in graph documents. The ingest path is the only consumer of
both and has no contract for the translation.
**Wording:** publish the mapping in A2-1.6 (and cite it from A1-3.6):
`graph_node_written.graph_node_id → Node.ID` · `graph_edge_written.graph_edge_id → Edge.ID` ·
`from_graph_node_id → Edge.SourceID (source_id)` · `to_graph_node_id → Edge.TargetID (target_id)` ·
`supersedes_graph_node_id → Node.SupersedesID (supersedes_id)`. Recommended (pre-Freeze, zero
cost): rename A2's served `id` to `graph_node_id`/`graph_edge_id` so one value has one name
platform-wide — A0-3.6 already reserves those spellings for errors and logs.

#### P-66 · MUST FIX · A0-2.17 V5 vs A1-1.1/A1-3.1 (a normative vector that is not a valid event)
V5's document has six keys and `kind:"tool_invoked"`, which is not one of A1-3.1's 39 kinds and not
the 17-key envelope — yet A2-5.4 cites V5 as the source of the **envelope field names**. Anyone
using V5 as an event fixture builds an invalid event.
**Wording:** annotate V5 in A0-2.17 — "a canonicalization vector with a synthetic key set:
`tool_invoked` is not an A1-3.1 kind and this is not a valid event document; the normative event
vector is A1 §4.3." Change A2-5.4's citation to A1-1.1 and delete A2 §6.9's "A1 was not on disk"
caveat (A1-3.8 has answered it).

#### P-67 · SHOULD FIX · A1-7.7 / A2-4.7 (which dedup protects a rewound watermark)
A1-3.8 claims A2-4.7's guarantee "follows from" the A1 dedup key, but A1's key only protects a
*client* retry; a rewound ingest watermark (A1-7.7 item 4) re-reads committed events and never
touches A1's dedup at all — only A2's content dedup stands between it and duplicate nodes.
**Wording (A2-4.7):** "Content dedup is the **only** protection against a rewound ingest
watermark (A1-7.7 item 4); `content_hash` MUST therefore be stable across platform releases
(A2-4.8, A0-2.16)."

#### P-68 · SHOULD FIX · A0-3.4 vs A2-9.5 / A1-4.9 (a lower document overrides a higher one, no tie-break rule)
A0-3.4 permits echoing untrusted request material in `message`; A2-9.5 forbids echoing a rejected
secret and notes it "overrides that here"; A1-4.9 agrees with A2. `contracts/README.md` fixes
ADR > SPEC > DESIGN > contract but says nothing about A0 vs A1–A8.
**Wording:** amend A0-3.4 — "…MAY echo untrusted request material (target, tool name, field name)
**except a value rejected by a secret-pattern rule, which MUST NOT be echoed in whole, in part or
as a digest (A2-9.5, A1-4.9)**" — and add to `contracts/README.md`: "A0 gates every later
contract; where A0 and A1–A8 disagree on a cross-cutting convention, A0 wins and the later
document is defective."

#### P-69 · SHOULD FIX · A1-4.9 vs A2-9.4 (one scan, two owners, two lists)
A1 delegates the pattern set to A2 and then states its own enumeration ("PEM blocks, NTLM/base64
hash shapes, cloud key prefixes, `krbtgt` material, high-entropy bearer strings") — a duplicate
source of truth for a safety rule.
**Wording:** A1-4.9 cites A2-9.4's rule table (P-46) and deletes its own list; rule ids live in
exactly one document.

#### P-70 · SHOULD FIX · A1-4.7 `EvidenceRefsMax` vs A2-7.1 `EvidenceIDsMax`; `evidence_refs` vs `evidence_ids`
Two constant names for the value 8 (A1-4.7 even says "matching A2-7.1's `EvidenceIDsMax`") and two
field names for one concept.
**Wording:** one constant (`EvidenceRefsMax = 8`) in the A0-7.1 registry, cited by both. Keep the
two field spellings only because A1's envelope key is digest-locked (A1-1.1, §4.3); record that
exception in A0-8.2 (see the AM-4 ruling).

#### P-71 · SHOULD FIX · A1-8.2 vs A2-11.4 (`TestCursorFromEngagementARejectedInB` has two oracles)
A1 requires `validation`; A2 accepts "an empty page or `validation`". One test name, two expected
outcomes.
**Wording:** once AM-4 lands, both MUST be `validation` (400) telling the client to restart from
the first page.

#### P-72 · SHOULD FIX · A0-7.1 vs A1-4.5 vs A2-7.1 (three cap tables, two const blocks, two constants in none)
A1's `EventMaxCanonicalBytes`, `ExitCodeMin/Max`, `IdempotencyKeyMaxBytes` and A2's
`MaxSupersedeChain` appear in no registry; A0-7.7 delegates mechanism assignment to "A2 and A3"
only, so A1-4.5's assignment is currently unsanctioned. Ruled in §3 (AM-2).

#### P-73 · SHOULD FIX · A2-7.1 last row pre-empts A3
A2 assigns A3's mechanism ("**A3** (T expected)") while A0-7.7 makes the assignment A3's and
A0-7.9/A2-12.4 leave the `next_cursor` question open.
**Wording:** "not assigned here — A3 per A0-7.7."

#### P-74 · SHOULD FIX · A1-3.6/A1-4.12 vs A2-2.1/A2-3.1 (`node_kind`/`edge_kind` validated by nobody on the event path)
A1 stores them as capped strings and explicitly excludes them from enum validation; A2 hard-rejects
unknown kinds. Since only the platform composes these kinds *after* a successful graph write, the
value is always valid — but as written no clause says who checks.
**Wording (A1-4.2, `graph_node_written`/`graph_edge_written` row):** "`node_kind` MUST be an
A2-2.1 kind and `edge_kind` an A2-3.1 kind; because only the platform composes these kinds after a
successful graph write, a violation is a platform defect → `internal` (A0-3.1)."

#### P-75 · NICE · A0-8.2 (no plural suffix rule)
A0-8.2 fixes `*_id` but A1 (`revert_event_ids`, `non_revertable_event_ids`) and A2
(`evidence_ids`, `addresses`) use plurals; A1's envelope uses `evidence_refs`.
**Wording:** add "`*_ids` (an array of identifiers, sorted and deduplicated per the owning
contract)" and record `evidence_refs` as the one approved exception (digest-locked, A1-1.1).

#### P-76 · NICE · A1 §4.1 vs A2 §4 (layer labels)
A1 says "Domain layer", A2 says "Foundation layer" for two packages DESIGN §1 puts in the same
layer. See P-58.

#### P-77 · NICE · A2-2.5 (`cvss_v3_x10` has no range)
The scale is fixed (×10, correctly, per A0-2.6) but the value range is not, and `attrs` ints are
otherwise unbounded.
**Wording:** "`cvss_v3_x10` is an integer in `[0, 100]`; a value outside it is `validation`."

#### P-78 · SHOULD FIX · A0-1.2 (no human-principal prefix) — the defect behind AM-1
A1-2.2 (`actor.principal_id` for `type:"user"`) and A2-5.3 (`operator_id`) both need a `usr_`
shape that A0-1.2 does not register, so A0-1.5 validation is impossible for every user-composed
event and every operator graph write. Ruled in §3: accept, and it blocks Freeze.

#### P-79 · NICE · A1-1.1 `node_id` vs A2-1.6 `agent_node_id` (correct, but untested)
The split is deliberate and correctly cross-cited (A0-3.6). No document requires a test that no
graph document ever emits `node_id`.
**Wording:** add `TestNoBareNodeIDInGraphDocuments` to A2-11.4's list.

---

## 3. Amendment requests — rulings

Eight requests (A1 AM-1…AM-4, A2 AM-1…AM-4; A1 AM-1 and A2 AM-1 are the same decision).
"Blocks Freeze" = the Freeze PR must not merge without it.

| AM | Doc · clause | Request | Ruling | One-line reason | Blocks Freeze |
|---|---|---|---|---|---|
| AM-1 | A1 §6.2 / A2 §6.2 · A0-1.2 | register a human-principal prefix (`usr_`) | **ACCEPT** (one decision, both cite it) | A0 owns id *shapes*; A5 owns tokens and scopes, so delegating the prefix to A5 would split A0-1.5 validation across two contracts, and without it every user-composed A1 kind (incl. the A1-6.5 override attribution) and every A2 operator write is unvalidatable | **YES** |
| AM-2 | A1 §6 / AM table · A0-7.7 | extend the mechanism-assignment duty from "A2 and A3" to every contract with capped fields | **ACCEPT** | A1-4.5 already discharges a duty A0-7.7 does not give it; one word ("A1, A2 and A3") makes the assignment sanctioned | No — but must land in the Freeze PR |
| AM-2 | A1 · A0-7.1 | adopt A1's ten local constants into the single A0 table/const block | **ACCEPT** | A1-4.7 already cross-references A2's constant by name, i.e. the drift has started; after Freeze a duplicated constant needs an ADR (A0-7.2) | No — same PR |
| AM-2 | A2 · A0-7.1 | adopt A2's nine local constants likewise | **ACCEPT** | same reasoning, and adoption *forces* the P-62 `tool_version` 32-vs-64 conflict into the open, which is the strongest argument for it | No — same PR |
| AM-3 | A1 §6 / AM table · A0-5.4 | record the `recorded_at` forward clamp in A0 | **ACCEPT** | monotone platform time is a time-authority rule (A0-5.6) that any later chained store (A3–A8) will need, not an event-local detail | No — but land it **with P-13**, or A0 records a rule no type can implement |
| AM-4 | A1 §6.11 · A0-4.8 | add "cursor `id` does not resolve in this collection" to the `validation` cases | **ACCEPT and strengthen to MUST** | A0-4.4 already promises "at worst an empty page or `validation`", and without the rule a foreign cursor is a *valid position* in the wrong chain (SPEC C8 / adversarial A12); it also collapses P-71's two oracles into one | No — but it gates WP-07's acceptance tests |
| AM-3 | A2 · A0-3.6 | extend the `node_id` reservation to payloads **and** bless A2's spellings (`source_id`, `target_id`, `supersedes_id`, `superseded_by_id`, `agent_node_id`) | **ACCEPT IN PART** — accept the reservation, **reject** the blessing | "`node_id` is the remote agent node in every JSON document, payload, error body and log attribute; a graph node is `graph_node_id`" is a genuine cross-cutting rule. Blessing A2's five field names in A0 creates a second source of truth for A2's own schema and freezes A2 vocabulary into the cross-cutting document — publish the A1↔A2 mapping instead (P-65) | No |
| AM-4 | A2 · A0-8.2 | confirm `*_ref` is not needed; `evidence_id`/`evidence_ids` for `evi_` references | **ACCEPT** | A0-8.2 is a "MUST NOT be used with another meaning" list, not a closed set of permitted suffixes, and `evidence_id(s)` already conforms to `*_id` | No — conditional on recording the two exceptions: the plural `*_ids` form (P-75) and A1's digest-locked `evidence_refs` (P-70) |

**On the double claim of `usr_`:** A1 §6.2 and A2 §6.2 both put it on their critical path and both
are right — it is one A0 table row plus one `ids.Kind` constant serving both, so it is a single PO
decision, not two. Land it as: `| user (human principal, SPEC §3) | usr_ |
usr_01m1y2whfhv3x6z9b2d5f8h1jk | ^usr_B{26}$ | 30 |`, `User Kind = "usr_"` in A0 §4, then cite
from A1-2.2 and A2-5.3. Do **not** defer the shape to A5: A5 would then own a value A0-1.5
validates.

**Freeze-PR items that are not amendment requests** (my own, from §1–§2): P-02, P-03, P-04, P-05,
P-06, P-11, P-19, P-20, P-21, P-22, P-23, P-24, P-34, P-35, P-36, P-46, P-66, P-68 and the
`contracts/README.md` tie-break sentence.

**PO decisions still open that are not confirmations** (each blocks a later contract, not A0–A2):
A0 §6.14 (500 × 512 B vs 64 KiB — A3's composition rule) · A1 §6.7 (head-hash anchoring over the
signed webhook — the only v1 mechanism that closes tail truncation) · A1 §6.9 (ownership of the
user/session audit gap; recommend A5 with its own platform-scoped store) · A2 §6.3 (is a
blacklisted discovery recorded at all) · **A2-2.7/A2-5.6 deviate from Q2's literal "Finding
carries confidence"** — the deviation is defensible and correctly flagged, but it is a change to a
locked PO decision and must be signed, not merely confirmed. Note also that A1 §6.4 and A2 §6.10
ask the *same* reject-vs-redact question twice: rule once (recommend reject, per P-46's rule
table) and cite from both.

---

## 4. Work-package split (WORKFLOW §5) and test coverage

Repo state: no `go.mod`, no `internal/`, no `.go` file — WP-01 is a real scaffold. Per
`contracts/README.md` **no code may be written against a Draft**, so WP-00 gates everything.
Every package below is one narrow concern, lists its exact contract excerpt, its files and its
named acceptance tests, and assumes nothing from a sibling except through the interfaces named in
its own excerpt. Parallelizable sets are marked ∥.

| WP | Concern | Contract excerpt (the only input) | Files | Acceptance tests (named) | Depends |
|---|---|---|---|---|---|
| 00 | Freeze PR: apply §1–§3 fixes, land AM-1…AM-4, PO signs the §6 checklists; flip A0/A1/A2 to `Frozen` | this report + A0 §6, A1 §6, A2 §6 | `contracts/A0-conventions.md`, `contracts/A1-events.md`, `contracts/A2-graph.md`, `contracts/README.md`, `sessions/…` | none (docs); `gofmt -l` n/a | — |
| 01 | Module scaffold + gate script | WORKFLOW §4, DESIGN §1 | `go.mod` (module `github.com/daten-krake/sleipnir`, go ≥1.26, **no** requires), `tools/gates.sh`, `.gitignore`, `README.md` pointer | `gates.sh` runs `gofmt -l`, `go vet ./...`, `go build ./...`, `go test ./...` green on the empty tree | 00 |
| 02 ∥ | `internal/errs` — kinds, envelope, ADR-0019 errors | A0-3.1…A0-3.10, A0 §4 `errs` sketch, ADR-0019 §2/§3/§5 | `internal/errs/{errs.go,kind.go,envelope.go,redact.go}` + colocated tests | `TestKindStatusTableAll13`, `TestEnvelopeAttrsOmittedNeverNull`, `TestRetryAfterPresence` (P-21), `TestErrorFormatComponentFunction`, `TestNoSecretInMessageOrAttrs`, `TestWrapCapturesOriginFunction` | 01 |
| 03 ∥ | `internal/logging` — slog subsystem loggers + correlation attrs | ADR-0019 §3, A0-3.6/3.8, DESIGN §1 | `internal/logging/*.go` | `TestSubsystemLoggerCarriesAttrs`, `TestNoSecretInLogRecords`, `TestLogOrReturnOnce` | 01 |
| 04 | `internal/ids` | A0-1.1…A0-1.10, A0 §4 `ids` sketch | `internal/ids/*.go` | `TestNewIDShapeAndAlphabet` (all 12 prefixes incl. `usr_`), `TestIDOrderIsChronological`, `TestValidRejectsNormalization` (case, `i`/`l`→`1`, `o`→`0`, wrong length, wrong prefix, empty), `TestValidIsByteExact`, `TestNewUsesCryptoRand` (no `math/rand` import; `go vet` + reflection audit), `TestUniquenessViolationIsInternal` | 02 |
| 05 ∥ | `internal/timex` | A0-5.1…A0-5.8, A0 §4 time block | `internal/timex/*.go` | `TestFormatTimeAlwaysThreeDigits` (whole second, ms), `TestParseTimeRejects` (table: 2 digits, 4 digits, `+02:00`, lowercase `t`/`z`, space, `:60`, year 2019/2100), `TestNowStripsMonotonic`, `TestInjectedClock` | 02 |
| 06 ∥ | `internal/caps` — every cap constant + mechanism T/R helpers | A0-7.1…A0-7.9 (+ the A1-4.5 and A2-7.1 constants per AM-2), A0 §4 caps block | `internal/caps/*.go` | `TestConstantsMatchA0Table` (one assertion per registry row), `TestTruncateRuneBoundary` (split multi-byte rune, cap == marker length, no-op under cap), `TestFitsMeasuresUTF8BytesNotRunes`, `TestTruncationMarkerIsTwelveBytes` | 02 |
| 07 | `internal/cjson` — **high review bar** | A0-2.1…A0-2.17 (all six vectors + V7 per P-04 + the rejection list), A0 §4 `cjson` sketch | `internal/cjson/*.go` | `TestVectorsV1ToV7ByteExact` (bytes, len, SHA-256), `TestRejections` (table: duplicate key, case-duplicate, `1.0`, `1e3`, `-0`, `01`, lone surrogate, trailing data, top-level array, BOM, 40-deep, `NaN`, raw control byte, invalid UTF-8, `null` via `CanonicalValue`, >1 MiB), `TestCanonicalIsStableAcrossRuns`, `TestCanonicalDecodesEscapes` (P-04), `TestExclusionListDropsTopLevelOnly`, `TestWithAddsKeys` (P-11), `TestSHA256HexLowercase64`, `TestDigestEqual` (equal, unequal, length mismatch, malformed hex) | 02 |
| 08 | `internal/paging` | A0-4.1…A0-4.8, A0 §4 paging block | `internal/paging/*.go` | `TestCursorRoundTripIsCanonicalBase64URL`, `TestCursorRejectsStandardAlphabetAndPadding` (A0-8.6), `TestDecodeCursorRejects` (bad base64, non-canonical JSON, wrong key set, wrong id kind per P-03), `TestHasMoreDetection` (last page exactly filling `limit` carries no `next_cursor`, A0-4.6), `TestLimitValidation` (P-22) | 04, 07 |
| 09 | `internal/events` — envelope, taxonomy, actor, payload structs | A1-1.1…A1-1.8, A1-2.1…A1-2.8, A1-3.1…A1-3.8 (all 39 tables), A1-4.1/4.8/4.12, A1 §4.1 | `internal/events/{kind.go,actor.go,event.go,payload_*.go}` | `TestEnvelopeKeySetIsSeventeen` (reflection: no `omitempty`, all keys present), `TestKindListIs39AndClosed`, `TestActorComponentIsEmptyForNonPlatform` (P-01), `TestPayloadStructsMatchA1Tables` (reflection: field names + `json` tags transcribed 1:1 per kind), `TestPayloadsAreFlat` (no nested object/array-of-object field), `TestNoTimeTimeInCanonicalizedTypes` (P-12), `TestUntrustedFieldsCoversEveryStarredField` (P-25) | 04, 05, 07 |
| 10 | `internal/events` — hash chain primitives | A1-5.1…A1-5.10, A1 §4.3 (normative vector) | `internal/events/{chain.go,served.go}` | `TestChainVectorDigests` (§4.3: preimage bytes, lens, hashes, served lens+digests, `prev_hash` links), `TestChainZeroIsSixtyFourZeros`, `TestChainSpecConstant`, `TestPreimageExcludesExactlyThreeNames`, `TestServedBytesReproducible`, `TestRecordedAtClampIsMonotone` (P-13, clock stepped backwards) | 09 |
| 11 | `internal/events` — pure validation pass | A1-7.10 steps 3/5/7/8/9/10/12/13, A1-4.2 (all obligations), A1-4.4…A1-4.7, A1-4.9 (delegating to WP-13), A1-7.3 derivation | `internal/events/{validate.go,derive.go}` | `TestValidationOrderIsDeterministic` (one error, first violated rule), `TestPerKindObligations` (table over A1-4.2, one row per kind), `TestCapsRejectWithSummaryTooLarge`, `TestArraysSortedDedupedAtComposition`, `TestEvidenceRefsDerivation` (P-07: `evidence_stored` included), `TestUntrustedFlagIsPlatformComputed`, `TestMaximalPayloadFitsCanonicalBound` (P-26), `TestFixedKeySetZeroValues` | 09, 06, 13 |
| 12 | `internal/events` — verification walk | A1-6.1…A1-6.9, A1-5.9 tamper matrix T1…T17 | `internal/events/{verify.go,chainstore.go}` (consumer-defined `ChainStore` interface + in-memory fake in tests) | `TestTamperMatrixT1ToT17` (one subtest per row, asserting detection **and** `break_kind`), `TestConcurrentAppendsProduceContiguousSeq`, `TestConcurrentAppendsAcrossEngagements`, `TestVerificationIsFullWalkNoSampling`, `TestVerificationIdempotent` (A1-6.8), `TestOneBreakEventPerRun` (P-32), `TestFailedChainStillAcceptsAppends` (A1-6.4) | 10, 11 |
| 13 ∥ | `internal/secretscan` — the shared rule table | A2-9.4 (+ the P-46 rule ids), A2-9.5, A1-4.9 | `internal/secretscan/*.go` | `TestEveryRuleIDMatchesItsCorpusValue`, `TestNoFalsePositiveOnBenignCorpus` (hostnames, CIDRs, argv, hex digests of artifacts), `TestErrorMessageNamesFieldAndRuleIDOnly` (never the value, a prefix or its digest), `TestEntropyRuleIsDeterministic` | 02 |
| 14 | `internal/graph` — kinds, `Node`/`Edge`/`Provenance`/`Attrs`, `contentDoc` | A2-1.1…A2-1.9, A2-2.1…A2-2.8, A2-3.1…A2-3.9, A2-4.6/4.8, A2-5.1…A2-5.8, A2-6.1…A2-6.5, A2 §4 sketches | `internal/graph/{kind.go,node.go,edge.go,provenance.go,attrs.go,content.go}` | `TestNodeKindListClosed` (10), `TestEdgeKindListClosed` (7), `TestContentDocKeySetIsFixedTwenty` (A2-4.6), `TestContentHashVector` (**P-82: needs a new normative vector**), `TestContentHashStableAcrossArrayOrder` (P-41), `TestAttrValueRoundTrip` (P-51), `TestAttrsRejectNestedFloatNull`, `TestProvenanceMandatory`, `TestNoBareNodeIDInGraphDocuments` (P-79) | 04, 05, 07, 13 |
| 15 | `internal/graph` — validation + pure write path | A2-10.2 steps 3…11, A2-10.3…A2-10.7, A2-7.1/7.2, A2-3.3…A2-3.7, A2-4.3/4.5/4.7 | `internal/graph/{validate.go,pending.go}` | `TestValidationOrderIsDeterministic` (A2-10.2), `TestEndpointMatrix` (one subtest per A2-3.2 row, allowed + disallowed), `TestNotApplicableFieldRejected`, `TestCIDRMustParse`, `TestCapsRejectWithSummaryTooLarge`, `TestReservedAttrKeysRejected` (P-49), `TestSelfEdgeRejected`, `TestSupersedeConstraints` (same kind, current target, no cycle, `MaxSupersedeChain` per P-53), `TestDuplicateEdgeCollapses`, `TestContradictsDirectionNormalized` (P-52), `TestNodeDedupIncludesConfidence` (P-42) | 14 |
| 16 | `internal/graph` — quarantine seam | A2-8.1…A2-8.10, A2-10.2 step 13, `QuarantineDecider` (P-48), A2-12.5 | `internal/graph/quarantine.go` + a stub decider in tests | `TestQuarantineIsNeverCallerSettable`, `TestBlacklistedNotReleasable`, `TestPerKindMatchedFields` (P-47 table), `TestOneHopPropagation`, `TestPlanningReadOmitsQuarantinedAndRetracted`, `TestReportingReadIncludesQuarantinedExcludesReportExcluded`, `TestQuarantineFlagFlipDoesNotSkipRows` (A2-8.9 mid-paging) | 15 |
| 17 | Store seams (interfaces only, no pgx) | A1-5.6/5.7, A1-7.2/7.8/7.9, A1-8.6, A2-1.4, A2-10.6, A2-11.1, P-45, P-54 | `internal/store/{events.go,graph.go}` | `TestNoUpdateOrDeleteOnEventSeam`, `TestNoBulkEventReadSpansEngagements`, `TestNoBulkReadSpansEngagements`, `TestEveryReadMethodTakesEngagementID` (reflection over both seams), `TestGraphSeamHasExactlyFourMutationMethods` (P-54) | 12, 16 |
| 18 | Ingest mapping (pure): A1 event → `NodeDraft`/`EdgeDraft` | A1-3.3 graph kinds, A1-3.6, A1-7.7, A2-5.3/5.4, the P-65 mapping table, P-61 | `internal/ingest/*.go` | `TestFieldNameMappingIsTotal` (every A1 graph-payload field maps to a graph field, P-65), `TestPrincipalKindIsNotCopiedFromActor` (P-61), `TestReplayIsIdempotent` (same event twice → one node, one edge), `TestRejectedWriteIsStillChained` (A2-10.8, via the events fake) | 12, 15, 17 |
| 19 | Shared contract suite: A0 owner package | `contracts/README.md` merge gate, A0-2.17, A0-6.1/6.2, A0-8.x | `internal/cjson/contract_test.go`, `internal/errs/contract_test.go`, `internal/paging/contract_test.go` | the six README gate items restricted to A0 types (P-83, P-87, P-88, P-89, P-90, P-93) | 07, 08 |
| 20 | Shared contract suite: A1 owner package | `contracts/README.md` merge gate, A1-4.9, A1-5.9, A1-8.6 | `internal/events/contract_test.go` | `TestEventSecretFreeSerialization`, `TestNoMaskingMappingInEvents`, `TestEventCarriesReferencesOnly`, `TestEventEngagementBReadNeverReturnsA`, `TestEventIDFromAIsNotFoundInB`, `TestCursorFromEngagementARejectedInB`, `TestFilterWithForeignIDReturnsNothing`, `TestSSEStreamNeverCarriesAnotherEngagement`, `TestServedEventIgnoresUnknownFields`, `TestAppendRejectsUnknownField` (incl. case variants, P-36), `TestAppendRejectsPlatformStampedFields` (A1-2.3), `TestKillPathDoesNotBlockOnEventStore` (A1-7.12) | 12, 13, 17 |
| 21 | Shared contract suite: A2 owner package | `contracts/README.md` merge gate, A2-9.7, A2-11.4 | `internal/graph/contract_test.go` | `TestGraphSecretFreeSerialization`, `TestCredentialNodeHoldsReferenceOnly`, `TestNoMaskingMappingInGraph`, `TestGraphEngagementBReadNeverReturnsA`, `TestGraphNodeIDFromAIsNotFoundInB`, `TestNoCrossEngagementEdge`, `TestCursorFromEngagementARejectedInB`, `TestNodeDedupDoesNotSpanEngagements` (P-85), `TestNodeWriteRejectsUnknownField`, `TestImmutableFieldUpdateIsConflict` (P-99), `TestDenormalizedEndpointKindsEqual` (P-100), `TestContentHashRecomputedFromStoredBytes` (P-101) | 16, 17 |
| 22 | `store/postgres` (pgx vendored, ADR-0010) — **integration, opt-in per DESIGN §8** | A1-5.4/5.6/5.7/7.6/7.9, A2-1.4/3.5/4.7/4.8, A0-1.9 | `internal/store/postgres/*.go`, `vendor/`, `deploy/…` | `TestAppendSerializesPerEngagement` (opt-in), `TestPreimageColumnRoundTrips`, `TestDedupConstraintBlocksRace`, `TestIDOrderingMatchesByteOrderCollateC` (P-105), `TestDBLevelAppendOnly` (A1-7.9 `REVOKE`) | 17 |

Fan-out rule for the implementers: WP-04/05/06 ∥ after WP-02; WP-09 needs 04+05+07; WP-13 ∥ with
04–08 (it imports only `errs`); WP-14 needs 13. No two packages in a ∥ set touch the same file.
Every package's brief quotes its excerpt verbatim plus DESIGN §2–§9 and AGENTS.md; no implementer
reads another contract document.

### 4.1 Merge-gate tests with no clause to point at

| `contracts/README.md` gate item | Gap | Finding |
|---|---|---|
| JSON round-trip for **every** contract type | No clause says how a *served event* is unmarshaled: `Event.Payload` is an interface and A1-7.3's two-pass decode is specified only for `AppendRequest`. `TestEventRoundTrip` cannot be written. | **P-80 · MUST FIX** — add `func UnmarshalEvent(b []byte) (Event, error)` to A1 §4.1: "two passes, both with `DisallowUnknownFields`: read the envelope keys to learn `kind`, select the concrete payload type (A1-4.1), decode `payload` into it; no `map[string]any` intermediate (A0-2.5)." |
| round-trip "including boundary sizes (A0 size caps)" | No boundary/maximal corpus is defined for A2 at all, and for A1 only by the P-26 wording I propose. | **P-81 · SHOULD FIX** — extend P-26's construction to A2: "a maximal node per kind (every field at its cap, `attrs` at 16 keys × 512 B, `addresses` at 16 × 64 B, `evidence_ids` at 8)" and name `TestMaximalNodeFitsCaps`. |
| canonical-JSON stability "matches the published test vectors" | A0 has six vectors and A1 has one; **A2 has none** for `content_hash`, although A2-4.6/4.8 make it an integrity and dedup value. Two implementations will diverge and every dedup will silently fail. | **P-82 · MUST FIX** — add a normative A2 vector in the style of A1 §4.3: the `contentDoc` canonical bytes for the §4.1 `finding` example, its length and its SHA-256, plus the `attrs`-only-differs and `addresses`-reordered variants proving P-41. |
| unknown-field tolerance | Clauses exist (A0-6.1, A1-4.10, A2-1.5/5.2) but no test is named anywhere, and the read-side decode path is P-80. | **P-83 · SHOULD FIX** — name `TestServedEventIgnoresUnknownFields`, `TestNodeReadIgnoresUnknownFields`, `TestAppendRejectsUnknownField`, `TestNodeWriteRejectsUnknownField` in A1-4.10 / A2-10.2. |
| hard-reject validation | A1-7.5 is a complete kind-by-kind table; A2-10.3 lists kinds but no test list, so the A2 negative half of the AGENTS.md rule ("a test proving the bypass attempt fails") is unnamed. | **P-84 · SHOULD FIX** — add to A2-10.3: "the shared suite proves each row: `TestHardRejectMatrix` with one subtest per rejection (unknown node kind, unknown edge kind, wrong endpoint kind, not-applicable field, unparseable `cidr`, over-cap field, over-cap `attrs`, reserved key, nested/float/null `attrs`, self-edge, supersede of non-current, supersede cycle, cross-engagement endpoint, missing provenance)." |
| cross-engagement negatives | Well covered (A1-8.6 ×6, A2-11.4 ×5). Missing: dedup across engagements. | **P-85 · SHOULD FIX** — add `TestNodeDedupDoesNotSpanEngagements` and `TestEdgeDedupDoesNotSpanEngagements` to A2-11.4: identical `content_hash` / identical endpoint tuple in A and B yields two distinct `gn_`/`ge_` ids. |
| secret-free serialization | Tests are named (A1-4.9 ×3, A2-9.7 ×3) but the **pattern corpus and rule ids do not exist**, so none of the six can be written. | **P-46** (MUST FIX, §1) — blocks this gate item outright. |

### 4.2 MUSTs with no test

| Clause | MUST | Named test needed | Finding |
|---|---|---|---|
| A0-1.4 | `crypto/rand` only; uniqueness violation → `internal`, never a retry loop | `TestIDGenerationUsesCryptoRand`, `TestUniquenessViolationIsInternal` | P-86 · SHOULD FIX |
| A0-1.5 | byte-exact validation, no Crockford normalization | `TestValidRejectsNormalization` (case fold, `i`/`l`→`1`, `o`→`0`) | P-86 |
| A0-2.5 | duplicate and case-duplicate keys rejected via a `Token()` walk | `TestDuplicateKeyRejected`, `TestCaseDuplicateKeyRejected` (the rejection list exists; no test id) | P-87 · SHOULD FIX |
| A0-2.11 | depth ≤ 32, size ≤ 1 MiB | `TestDepthLimit`, `TestSizeLimit` (only "40-deep nesting" is listed as a vector rejection) | P-87 |
| A0-2.15 | constant-time digest comparison | `TestDigestEqual` (equal, unequal, length mismatch, malformed hex, empty) | P-88 · SHOULD FIX |
| A0-3.10 | `Retry-After` == `ceil(retry_after_ms/1000)`, both present | `TestRetryAfterAgreement` | P-89 · SHOULD FIX |
| A0-4.6 | a last page that exactly fills `limit` carries no `next_cursor` | clause says "Contract test:" but names none → `TestHasMoreDetection` | P-90 · SHOULD FIX |
| A0-5.3 | six rejection classes, never normalize | `TestParseTimeRejects` (table) | P-91 · SHOULD FIX |
| A0-7.4/7.5 | rune-boundary truncation + marker counted against the cap | `TestTruncateRuneBoundary` (only A3 uses T, but `Truncate` ships in WP-06) | P-92 · SHOULD FIX |
| A0-8.6 | base64url unpadded; reject the standard alphabet and `=` | `TestCursorRejectsStandardAlphabetAndPadding` | P-93 · SHOULD FIX |
| A0-2.14 / A1-1.2 | fixed key set, zero values, no `omitempty` | `TestEnvelopeKeySetIsSeventeen` + `TestPayloadKeySetsAreClosed` (reflection over all 39 types) | P-94 · SHOULD FIX |
| A1-4.6 | `redacted` is platform-set, means redaction, never "truncated" | `TestRedactedIsPlatformSetOnly`, `TestRedactedNotUsedForTruncation` | P-95 · NICE |
| A1-5.4 | `recorded_at` clamped non-decreasing | `TestRecordedAtClampIsMonotone` (clock stepped backwards) | P-13 (test named in WP-10) |
| A1-6.8 | a verification run is idempotent; writes only the two integrity kinds | `TestVerificationIdempotent`, `TestVerificationWritesNoOtherKind` | P-96 · SHOULD FIX |
| A1-7.12 | kill paths MUST NOT block on the event store | `TestKillPathDoesNotBlockOnEventStore` (store that always fails; the kill still happens, the event is retried, the failure is logged and surfaced) — **safety-critical, currently untested** | P-97 · MUST FIX |
| A1-8.5 | emit after commit, in `seq` order, at-least-once | `TestSSEEmitAfterCommitInSeqOrder`, `TestSSEReplayDedupesByEventID` | P-98 · SHOULD FIX |
| A1-8.2 | a cursor must resolve in the requested engagement | `TestCursorFromEngagementARejectedInB` (named in A1-8.6 ✓, oracle conflict with A2 → P-71) | P-71 |
| A2-1.3 | any other field change → `conflict`, never partially applied | `TestImmutableFieldUpdateIsConflict` | P-99 · SHOULD FIX |
| A2-3.8 | denormalized `source_kind`/`target_kind` equal the nodes' kinds ("the contract test asserts equality" — unnamed) | `TestDenormalizedEndpointKindsEqual` | P-100 · SHOULD FIX |
| A2-4.8 | verification recomputes from stored canonical bytes | `TestContentHashRecomputedFromStoredBytes` (A2's analogue of A1's T16) | P-101 · SHOULD FIX |
| A2-6.3 | reserved-key rejection | `TestReservedAttrKeysRejected` | P-102 · SHOULD FIX |
| A2-8.9 | mutable flags never order, never skip or duplicate a row mid-paging | `TestQuarantineFlagFlipDoesNotSkipRows` | P-103 · SHOULD FIX |
| A2-10.8 | a rejected graph write is still chained | `TestRejectedWriteIsStillChained` (asserts an `action_blocked{graph_write_rejected}` event exists and the graph row does not) | P-104 · SHOULD FIX |
| A0-1.9 | `COLLATE "C"` ordering | `TestIDOrderingMatchesByteOrderCollateC` — integration, opt-in per DESIGN §8; no clause says so | P-105 · SHOULD FIX |

### 4.3 Coverage verdict

The three documents name **17** tests explicitly (A1: 10, A2: 8 minus one duplicate name,
A0: 1 "Contract test:" note). The merge gate needs roughly **60**. The gaps are not random: they
are concentrated exactly where a document describes a mechanism instead of a rule (P-46 rule ids,
P-82 A2 vector, P-80 event decode, P-84 A2 hard-reject matrix). Fixing those four closes about
half the gap; the rest is naming, which WP-00 can do in one pass over the three §4 sections by
adding a "Contract tests" subsection to each document listing the ids above against their clause.

## NOT COVERED by this run

- **ADR-by-ADR conformance** for ADR-0005, 0009, 0010, 0011, 0013, 0017, 0018, 0020: I verified
  ADR-0016 §1–§2 (node/edge kind lists match A2-2.1/A2-3.2 exactly, provenance fields match
  A2-5.3), ADR-0012 §2 (six notification kinds match A1-3.3), ADR-0019 §2 (message format) and
  SPEC C1–C11 / Q1–Q15 against the traceability tables, but I did not read the other eight ADRs
  line by line. Their traceability rows in A1 §5 / A2 §5 were spot-checked only.
- **A1 §4.3 vector arithmetic**: not re-done (parent-verified, per instruction). A0-2.17's six
  vectors: lengths re-derived by hand for V3 (57) and V6 (18) ✓, digests not recomputed.
- **A3–A8 forward compatibility**: only where A0/A1/A2 already delegate (A0-7.9/7.10, A1-8.9,
  A2-12). No opinion on whether the A0 cursor shape or the A0-3 kind list will survive A4.
- **`.pi/`, `docs/`, `blog/`, `next_steps.md`**: not reviewed (A0 §6.2 already flags
  `next_steps.md` as stale on `node_` vs `slp_node_`).
