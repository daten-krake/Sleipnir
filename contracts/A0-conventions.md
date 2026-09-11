# A0 — Cross-cutting conventions

## 1. Header

| | |
|---|---|
| **Contract id** | A0 |
| **Status** | `Draft` (`contracts/README.md` lifecycle: Draft → Frozen → Implemented) |
| **Owner** | architect |
| **Gates** | A1 (events), A2 (graph) and every later contract (A3–A8); every package that serializes JSON, hashes a value, returns an error, or lists rows |
| **Implements** | ADR-0010 §2–§3 · ADR-0011 · ADR-0012 §2/§7 · ADR-0016 §1/§4 · ADR-0018 §1–§3 · ADR-0019 §1–§5 · ADR-0020 §3–§4 · SPEC §2 C1/C9/C11, §6, §8, §9 · Q3, Q4, Q6, Q9, Q10, Q11, Q13, Q15 |
| **External refs** | RFC 2119 (MUST) · RFC 8785 JCS · RFC 3339 · RFC 4648 §5 · RFC 9110 · ULID layout |
| **Authority** | ADR > SPEC > DESIGN > contract (`contracts/README.md`). A clause here that contradicts an Accepted ADR is a defect in this document. |

## 2. Scope

**Fixed here:** identifier shape and prefixes; the one canonical JSON
serialization used by the event hash chain (Q11) and approval fingerprints
(ADR-0018); error kinds → HTTP mapping and the error envelope; pagination
cursors; timestamp format and authority; the unknown-field/versioning
asymmetry (Q13 vs Q3); the Q4 size caps and the two enforcement mechanisms;
field, enum, hex and base64 conventions.

**Not fixed here:** event taxonomy and chain fields (A1) · node/edge kinds,
provenance, quarantine (A2) · stage-view composition (A3) · endpoints,
scopes, SSE framing (A4) · token formats and the Q6 exclusion list (A5) ·
fingerprint *content* (A7) · config (A8) · DDL/persistence schema (backlog 6)
· rate-limiter values (A4) · UI rendering of errors and time (UI session).

## 3. Normative clauses

### A0-1 · Identifiers

- **A0-1.1** Every entity identifier is `<prefix><body>`. The body is **26
  characters** of lowercase Crockford base32 (`0123456789abcdefghjkmnpqrstvwxyz`
  — no `i l o u`) encoding 128 bits in ULID layout: 48-bit unsigned millisecond
  Unix time (10 chars) followed by 80 bits from `crypto/rand` (16 chars). The
  48-bit millisecond value is encoded big-endian in 10 characters, whose first
  is therefore in `0`–`7`; the 80 random bits are encoded in 16 characters with
  no restriction. Validation (A0-1.5) is the regex only: a body whose first
  character is `8`–`z` is accepted but never generated.
  _One generator, no coordination, lexicographic order = chronological order,
  stdlib-only (`crypto/rand` + a 32-char alphabet table) — **PO confirm**._
- **A0-1.2** Prefixes are closed and per type; the set is closed at **12**
  prefixes. `slp_node_` is fixed by Q9; the rest are recommendations (**PO
  confirm**). `B` below is the body of A0-1.1 and the regex character class
  `[0-9a-hjkmnp-tv-z]` is exactly that alphabet.

  | Entity | Prefix | Example | Regex | Total len |
  |---|---|---|---|---|
  | engagement | `eng_` | `eng_01m1y2whfhgbz06ays6dxnvyws` | `^eng_B{26}$` | 30 |
  | run | `run_` | `run_01m1y2whfhnjx2am9103w0pnqw` | `^run_B{26}$` | 30 |
  | job (orchestrator container) | `job_` | `job_01m1y2whfhbt69j0h0fbxepw90` | `^job_B{26}$` | 30 |
  | task (worker container) | `task_` | `task_01m1y2whfh1txm57x8dn41r9hg` | `^task_B{26}$` | 31 |
  | event | `evt_` | `evt_01m1y2whfhp17g0avdqztd2p3x` | `^evt_B{26}$` | 30 |
  | remote agent node (Q9) | `slp_node_` | `slp_node_01m1y2whfhxydsaem68cmazyc8` | `^slp_node_B{26}$` | 35 |
  | graph node (ADR-0016) | `gn_` | `gn_01m1y2whfhh039ykj5x8mc5a0g` | `^gn_B{26}$` | 29 |
  | graph edge (ADR-0016) | `ge_` | `ge_01m1y2whfh62ej11jf4x5gjzv4` | `^ge_B{26}$` | 29 |
  | evidence (ADR-0009) | `evi_` | `evi_01m1y2whfh3ca875z2x8v8h7qt` | `^evi_B{26}$` | 30 |
  | approval (ADR-0018) | `apr_` | `apr_01m1y2whfh0asxstccc64q4cfx` | `^apr_B{26}$` | 30 |
  | tool (registry entry, ADR-0008) | `tool_` | `tool_01m1y2whfhfjdvwqp9pfxqekmf` | `^tool_B{26}$` | 31 |
  | user (human principal, SPEC §3) | `usr_` | `usr_01m1y2whfhv3x6z9b2d5f8h1jk` | `^usr_B{26}$` | 30 |

- **A0-1.3** Tool identifiers follow A0-1.1 like every other entity; the
  human-readable `name` and `version` are separate registry fields (ADR-0008,
  Q14) and MUST NOT be embedded in the id. _A slug id (`tool_nmap`) would have
  to change on rename/re-version, and ADR-0018 fingerprints must stay stable —
  **PO confirm**._
- **A0-1.4** Generation MUST use `crypto/rand`; MUST NOT use `math/rand`, a
  counter, a name, a hash of user input, or any client-supplied material. A
  uniqueness violation at insert MUST surface as `internal` (A0-3) and MUST NOT
  be silently retried in a loop.
  Tests: TestIDGenerationUsesCryptoRand, TestUniquenessViolationIsInternal.
- **A0-1.5** Validation happens at every trust boundary, anchored and
  byte-exact against the type's regex. Decoders MUST NOT apply Crockford
  normalization (case folding, `i`/`l` → `1`, `o` → `0`): an id that is not a
  byte-exact match is invalid → `validation`. _Byte-exactness matters because
  ids enter canonical JSON and therefore digests (A0-2)._
  Tests: TestValidIsByteExact, TestValidRejectsNormalization.
- **A0-1.6** Identifiers are **opaque**. A client MUST NOT parse the prefix or
  body, derive creation time, ordering, sharding or type from an id, or
  construct an id. The time-sortability of A0-1.1 is a platform storage
  convenience, **not** an API guarantee.
- **A0-1.7** Identifiers are **not secrets**: they appear in URLs, logs, error
  `attrs` (A0-3.6) and customer reports. An id MUST NOT be used as a credential
  (credentials are A5), and authorization MUST NOT rely on id unguessability
  (ADR-0005 §5). Approximate creation time leaked by the body is accepted.
- **A0-1.8** Whether the node *identifier* and the node *mesh token* of Q9 are
  the same value is decided by A5. A0 assumes they are distinct: the identifier
  is public (A0-1.7), the token is secret. `slp_node_` is the only prefix
  carrying the `slp_` namespace — Q9 is binding, and a distinctive prefix makes
  a leaked token detectable by secret scanners.
- **A0-1.9** Ordering by identifier text in PostgreSQL MUST use `COLLATE "C"`
  so database order equals byte order (A0-4.3). The `COLLATE "C"` ordering is
  declared **on the column** (DDL); queries MUST NOT re-specify it (a per-query
  `COLLATE` silently disables index use on every paginated read, A0-4.6).
  Tests: TestIDOrderingMatchesByteOrderCollateC — integration, opt-in per
  DESIGN §8. _Second-order consequence of
  A0-1.1: the default collation is locale-aware and would not sort `0` < `B` <
  `_` < `a`._
- **A0-1.10** A prefix MUST NOT be renamed, reused for another type, or dropped
  within `/api/v1` (A0-6.5). A new entity type gets a new prefix by contract
  amendment (additive).

### A0-2 · Canonical JSON

- **A0-2.1** "Canonical JSON" is the single byte-exact serialization defined
  here. It is used by the event hash chain (Q11, A1), approval action
  fingerprints (ADR-0018 §1, A7), and any other value that is hashed, signed,
  or compared byte-wise. It MUST be produced by exactly one implementation
  (§4) and MUST NOT be re-implemented, hand-formatted, or approximated per
  package. _Two independently written canonicalizers = two different chains._
- **A0-2.2** The form is RFC 8785 (JCS) with **two declared restrictions**:
  integers only (A0-2.6) and UTF-8 byte key order (A0-2.4). For any document
  whose numbers are integers and whose keys are BMP-only, the output is
  byte-identical to RFC 8785.
- **A0-2.3** Output MUST be UTF-8 without BOM. Input MUST be UTF-8; invalid
  UTF-8, a BOM, empty input, or trailing data after the top-level value MUST be
  rejected → `validation`. Non-ASCII characters are emitted **literally** (no
  `\uXXXX` re-encoding) except control characters (A0-2.7). Implementations MUST
  NOT rely on `encoding/json` for UTF-8 validation — it substitutes U+FFFD. The
  canonicalizer MUST (a) reject the whole document when `utf8.Valid(doc)` is
  false and (b) scan every `\u` escape and reject a surrogate code point
  (`D800`–`DFFF`) not followed by a complementary surrogate forming a valid
  pair; both `validation`.
  Tests: TestInvalidUTF8Rejected, TestLoneSurrogateRejected.
- **A0-2.4** Object keys MUST be sorted ascending by the **unsigned byte value
  of their UTF-8 encoding**. This deviates from RFC 8785 §3.2.3 (UTF-16 code
  units) only for keys containing code points ≥ U+10000, where byte order puts
  U+FFFD *before* U+1F600 while JCS does the reverse; vector V6 locks our
  behaviour. _Byte order is what `sort.Strings` and `COLLATE "C"` already give
  us; UTF-16 order would be hand-rolled conversion code for keys contract
  types never produce (A0-8.1)._
- **A0-2.5** A document MUST NOT contain duplicate keys in one object; two keys
  differing only by case count as duplicates. Violation → `validation`.
  Implementation MUST drive `json.Decoder.Token()` (or an equivalent stream
  walk) that observes every key: decoding into `map[string]any` silently keeps
  the last duplicate, and `encoding/json` matches field names
  case-insensitively, so `{"a":1,"A":2}` would otherwise resolve by map
  iteration order. _A nondeterministic digest is the failure mode this clause
  exists to prevent._ The `Token()` walk MUST call `Decoder.UseNumber()` and MUST
  re-validate every number's **literal token text** against
  `^-?(0|[1-9][0-9]{0,15})$` before range-checking with `strconv.ParseInt`
  against A0-2.6's `[-(2^53-1), 2^53-1]`; a token whose text is not already
  canonical (`.`, `e`/`E`, `+`, leading zeros, `-0`) is `validation`. The
  canonicalizer emits the validated literal text verbatim. The walk MUST call
  `Decoder.More()` after the top-level value (A0-2.3 no trailing data) and MUST
  count nesting depth itself (`encoding/json` has none).
  Tests: TestDuplicateKeyRejected, TestCaseDuplicateKeyRejected,
  TestCanonicalEmitsNumberLiteralText, TestRejections.
- **A0-2.6** Numbers MUST be integers in `[-(2^53-1), 2^53-1]`, written with no
  leading zeros (except a bare `0`), no `+`, no fraction, no exponent. The
  literal `-0`, any `.`, any `e`/`E`, `NaN` and `Infinity` MUST be rejected →
  `validation`. A type that is ever canonicalized MUST NOT declare a float
  field: use integers, or fixed-point with the scale fixed by the owning
  contract. The literal text of an accepted number matches
  `^-?(0|[1-9][0-9]{0,15})$` (16 digits is the width of `2^53-1`) and is
  re-emitted verbatim (A0-2.5). _RFC 8785 requires ECMAScript number
  serialization, which is
  hand-rolled high-review-bar code (`AGENTS.md`) for a benefit nothing in the
  hashed payloads needs — **PO confirm**._
- **A0-2.7** Strings: escape only `"`, `\`, and U+0000–U+001F. Use the short
  escapes `\b \f \n \r \t` for those five and `\u00xx` (**lowercase** hex) for
  the other control characters; every other code point is literal. HTML
  escaping MUST NOT be applied (`<`, `>`, `&` stay literal — i.e.
  `SetEscapeHTML(false)` semantics); raw control bytes inside a string and lone
  surrogates (`\ud800` unpaired) MUST be rejected → `validation` (A0-2.3 (b)).
  `Canonical` MUST decode every JSON string escape in the input (`\uXXXX` incl.
  well-formed surrogate pairs, `\n`, `\"`, `\\`) and re-emit per this clause; it
  MUST NOT pass an input escape through. **U+2028 and U+2029 are emitted
  literally (raw UTF-8, 3 B each), not as `\u2028`/`\u2029`; U+007F is literal
  (1 B).** `encoding/json` escapes U+2028/9 unconditionally and HTML-escapes
  `<>&`, so it MUST NOT be the canonical emitter — it MAY produce the
  intermediate bytes that the canonicalizer re-parses (`CanonicalValue`), never
  the final ones.
- **A0-2.8** Array element order MUST be preserved exactly; arrays are never
  sorted. `true`, `false`, `null` are lowercase. `null` is recognized when
  scanning input and rejected by `Canonical` (A0-2.14); no contract document
  contains it (A0-8.3).
- **A0-2.9** Output MUST contain no insignificant whitespace: no spaces, tabs,
  newlines or indentation; `,` and `:` are bare separators.
- **A0-2.10** The top-level value MUST be a JSON object. A top-level array,
  string, number or literal MUST be rejected → `validation`.
- **A0-2.11** Input limits: nesting depth ≤ **32** levels, size ≤ **1 MiB**;
  either exceeded → `validation`. Depth counts container boundaries — the
  top-level object is level 1, each nested object or array adds 1, scalars are
  not levels; a document at depth 32 MUST be accepted, at 33 rejected
  (`validation`); size is `len(doc)` of the input; both are checked before
  canonicalization. This walk is AGENTS.md high-review untrusted-input parsing.
  _Untrusted-input bound (`AGENTS.md`
  high-review list); contract types nest ≤ 6 — **PO confirm**._
  Tests: TestDepthLimit, TestSizeLimit.
- **A0-2.12** A canonicalized type declares its **exclusion list**: a fixed set
  of *top-level JSON field names* removed before canonicalization (event chain
  fields `hash`, `prev_hash`, `seq`; approval decision fields). Exclusion lists
  are declared in the owning contract document (A1, A7) and MUST use plain
  top-level names — no path syntax, no wildcards, no nested exclusion. An
  excluded field MUST NOT influence the digest.
- **A0-2.13** An exclusion list is part of the digest's definition. After a
  contract is Frozen it MUST NOT change except by a new ADR: for A1 a change
  invalidates verification of the entire historical chain (Q11).
- **A0-2.14** A canonicalized type MUST serialize a **fixed key set**: every
  field declared in its contract document appears on every instance, with the
  zero value (`""`, `0`, `false`, `[]`, `{}`) when unset. No `omitempty`, no
  absent optional fields, no `null` (A0-8.3). The rule applies per concrete Go
  type; a kind-specific payload type (A1) has its own fixed key set.
  Constructors MUST initialize every slice and map field of a canonicalized type
  to a non-nil empty value (`events.NewEvent`, `graph.NewNode`,
  `cjson.CanonicalValue`); `json.Marshal` emits `null` for a nil
  slice/map/pointer, which changes every digest. **`cjson.Canonical` and
  `cjson.CanonicalValue` MUST reject a `null` at any depth with `validation`.**
  A canonical document containing `null` is a platform defect → `internal`.
  _Producer and verifier must agree byte-for-byte or ADR-0018 §2 re-validation
  aborts every action._
  Tests: TestNilCollectionNeverSerializesAsNull, TestRejections.
- **A0-2.15** A digest is SHA-256 over the canonical bytes, written as **64
  lowercase hex** characters (`crypto/sha256` + `encoding/hex`). A digest
  comparison that gates an action (chain verification, fingerprint match,
  webhook MAC) MUST use `crypto/subtle.ConstantTimeCompare` on decoded bytes.
  Tests: TestDigestEqual, TestGatingComparisonsAreConstantTime.
- **A0-2.16** Every record whose digest is persisted MUST also persist the
  **exact canonical bytes** the digest was computed over, and verification MUST
  recompute from those stored bytes — never from a re-serialization of decoded
  fields. _Re-marshaling a struct later (new field, new Go version, different
  encoder settings) would silently break every historical hash. This is the
  most consequential implementation rule in A0; A1 must provide the column._
- **A0-2.17** Test vectors below are normative: the shared contract-test suite
  (`contracts/README.md`) MUST assert canonical bytes and digests byte-exactly.
  `SHA-256` is over the canonical bytes column.

  | # | Input (as received) | Canonical bytes (exact) | len | SHA-256 |
  |---|---|---|---|---|
  | V1 | `{"b": 1, "a": 2}` | `{"a":2,"b":1}` | 13 | `d3626ac30a87e6f7a6428233b3c68299976865fa5508e4267c5415c76af7a772` |
  | V2 | `{"a":1,"B":2,"_c":3,"0":4}` | `{"0":4,"B":2,"_c":3,"a":1}` | 26 | `735f90d32fc3437e6c52f922158bc966ab66a8ba875ccc60693c3c93d513935a` |
  | V3 | `{"s":"a\"b\\c\nd\te\u0001f<>&\u00e9\u4e2d\ud83d\ude00","n":-42,"t":true}` | `{"n":-42,"s":"a\"b\\c\nd\te\u0001f<>&é中😀","t":true}` | 57 | `63fc2ef866b5542bfcd5d8c943f71a92b1b6a8a2403704f0416822a18aefcd6c` |
  | V4 | `{"arr":[3,1,2,{"y":1,"x":2}],"obj":{},"e":[]}` | `{"arr":[3,1,2,{"x":2,"y":1}],"e":[],"obj":{}}` | 45 | `d66e67ee4f3bef3250a4b86aa3ea680d7c9a5545424cf4316a9cf917e39db521` |
  | V5 | `{"engagement_id":"eng_01m1y2whfhgbz06ays6dxnvyws","kind":"tool_invoked","recorded_at":"2026-09-07T14:03:22.481Z","seq":4711,"prev_hash":"9b74c9897bac770ffc029102a200c5de","hash":"03c8a7d2b9e4f1a6d5c0b8e7f2a1d4c3b6a9e8f7d0c1b2a3e4f5061728394a5b"}`<br>exclusions: `seq`, `prev_hash`, `hash` | `{"engagement_id":"eng_01m1y2whfhgbz06ays6dxnvyws","kind":"tool_invoked","recorded_at":"2026-09-07T14:03:22.481Z"}` | 113 | `3695e846ce9e48d6577c0a7ef48ace7b9d34e4ca498dc91aac4309b40c4e10cb` |
  | V6 | `{"😀":1,"\uFFFD":2}` (keys are U+1F600 and U+FFFD) | `{"�":2,"😀":1}` (the U+FFFD key is the three bytes `EF BF BD`; the escape appears only in the input column; len 18 and the published digest are correct) | 18 | `9fbfff35f05fb9c72de19e392d0a1b848acb4d6c42489709d709982d10dd883c` |
  | V7 | `{"s":"a\u2028b\u2029c\u007fd"}` | `{"s":"a<2028>b<2029>c<7F>d"}` (placeholders for the literal bytes `E2 80 A8`, `E2 80 A9`, `7F` - not escapes) | 19 | `aedd6df88cc462fdbdc5788549d753c9b8c21ac8b9e16c51c72016cf564c3e85` |
  | V8 | `{"s":"<a href=\"x\">&é\u2028"}` | `{"s":"<a href=\"x\">&é<2028>"}` (`<2028>` = the literal bytes `E2 80 A8`; `<`, `>`, `&` and é stay literal; `\"` stays escaped per A0-2.7) | 28 | `63995ca86de5cce6f4d74df8e1d90aed78918aad912cc7c7feaff4d89fb15b2d` |

  In V3 the canonical bytes contain literal UTF-8 `é` (2 B), `中` (3 B), `😀`
  (4 B), literal `<>&`, short escapes for newline/tab, `\u0001` in lowercase
  hex, and keys reordered `n`,`s`,`t`. In V6 the U+FFFD key is the three bytes
  `EF BF BD` \u2014 the escape appears only in the input column; len 18 and the
  published digest are correct. In V7 the canonical bytes carry a literal U+2028
  (`E2 80 A8`), a literal U+2029 (`E2 80 A9`) and a literal U+007F (`7F`),
  printed as `<2028>`, `<2029>`, `<7F>`: all three are invisible in a terminal,
  so verify by len 19 and the digest, not by eye. In V8 `<`, `>`, `&` stay literal, `é` is `C3 A9`, U+2028 is printed
  as `<2028>` and is the bytes `E2 80 A8`, and
  `\"` stays escaped because A0-2.7 requires it.

  **Notation, canonical-bytes column.** A `\uXXXX` sequence in that column is
  **literal text** when it is a required JSON escape under A0-2.7 (V3's
  `\u0001`; V8's `\"`, `\\`, and the `\n`/`\t` short escapes) and is otherwise
  printed as **the character itself** (V6's U+FFFD key; V3's `é中😀`) -
  **except** that a code point which is a line break or a C0/C1 control
  (U+2028, U+2029, U+007F) is printed as an angle-bracket placeholder naming its
  bytes: `<2028>` = `E2 80 A8`, `<2029>` = `E2 80 A9`, `<7F>` = `7F` (V7, V8).
  _Reason: a raw U+2028/U+2029 inside a Markdown table row makes that row's line
  count reader-dependent - Python's `str.splitlines` and several editors split on
  them, `grep` and `awk` do not - so a normative vector would parse differently
  per tool. The placeholder is a display form only: the canonical bytes carry the
  real code points, and the `len` and `SHA-256` columns are the authority._
  Where a cell could be read both ways the `len` column
  decides: V3 is 57 bytes only if `\u0001` is the six characters
  `\`,`u`,`0`,`0`,`0`,`1`, and V6 is 18 bytes only if its first key is the three
  bytes `EF BF BD` (the six-character reading would give 21 and a different
  digest). The input column always shows the bytes as received, escapes included.

  **V5 annotation.** V5 is a canonicalization vector with a **synthetic key
  set**: `tool_invoked` is not an A1-3.1 kind and this is not a valid event
  document (six keys, not the 17-key envelope of A1-1.1). The normative event
  vector is A1 §4.3; the envelope field names are A1-1.1.

  Rejections (each → `validation`, no digest): `{"a":1,"a":2}` (duplicate) ·
  `{"a":1,"A":2}` (case-duplicate) · `{"a":1.0}` · `{"a":1e3}` · `{"a":1E3}`
  (uppercase exponent) · `{"a":-0}` · `{"a":-0.0}` · `{"a":01}` ·
  `{"a":9007199254740993}` (2^53+1, outside A0-2.6) ·
  `{"a":10000000000000000000}` (20 digits, outside A0-2.5's token regex) ·
  `{"a":"\ud800"}` (lone surrogate, A0-2.3) · `{"a":"\xff"}` (invalid UTF-8,
  A0-2.3) · `{"a":null}` (A0-2.14) · `{"a":1},` (trailing) · `[1,2]` (top
  level) · BOM-prefixed · 33-deep nesting (A0-2.11 boundary + 1) · 40-deep
  nesting · a document larger than 1 MiB · `{"a":NaN}` · raw control byte in a
  string.

  Accepts (normative boundary cases): a **32-deep** document (A0-2.11 accepts
  its own boundary) · `{"a":"\ud83d\ude00"}` (a well-formed surrogate pair,
  emitted as the literal 😀, 4 B).

### A0-3 · Error kinds and HTTP mapping

- **A0-3.1** The machine-readable `kind` is the contract. The v1 list is closed
  (13 kinds, A0-8.4 spelling). Clients MUST switch on `kind` — never on the
  HTTP status (the mapping is many-to-one) and never on `message`.

  | kind | HTTP | Retryable | Why it exists |
  |---|---|---|---|
  | `validation` | 400 | no | malformed input, bad id (A0-1.5), unknown field on write (A0-6.2), bad enum (A0-6.3), bad timestamp (A0-5.3), canonicalization failure (A0-2) — Q3 hard reject |
  | `auth` | 401 | no | credential missing, expired, malformed, or unverifiable (A5) |
  | `forbidden` | 403 | no | principal-level denial naming no object: a machine principal on a Q6-excluded endpoint, or a missing verb scope. Distinct from `auth` so an agent stops instead of re-authenticating |
  | `notfound` | 404 | no | object absent **or outside the caller's engagement scope** (A0-3.9) |
  | `conflict` | 409 | no | state prevents the request: single-use approval already consumed (Q10), duplicate write, immutable field, spawn-quota rejection (§6.12) |
  | `approval_required` | 409 | no | execution/spawn attempted with no valid approval for the fingerprint (ADR-0005 §1, ADR-0018). 409 keeps 403 for principal-level denial so an agent can tell "never allowed" from "not yet allowed" |
  | `approval_expired` | 409 | no | approval past its 2 h window (ADR-0012 §7, SPEC §5.6): orchestrator replans instead of waiting |
  | `integrity_failed` | 409 | no | hash-chain verification failed; export blocked until an operator override is logged as an event (Q11). Not 5xx: 5xx would be retried and alerted as a platform bug |
  | `summary_too_large` | 413 | no | a capped field **or count** exceeded a budget declared by its owning contract under mechanism R (A0-7.6) — the Q4 constants of A0-7.1 and the per-contract caps of A1-4.5 / A2-7.1. Named by Q3/Q4; distinguishable from `validation` because the remedy is "send fewer bytes", not "fix a malformed field" |
  | `timeout` | 504 | yes | a platform-side deadline expired (own DB, LLM gateway, broker) |
  | `upstream` | 502 | yes | an upstream (LLM endpoint, webhook target, container runtime) answered with an error or unusable bytes |
  | `rate_limited` | 429 | yes | platform rate limiter (ADR-0011 names rate limiting as part of the API attack surface). Reserved now so a later limiter does not have to reuse the non-retryable `conflict` |
  | `internal` | 500 | no | platform defect or unclassified failure |

- **A0-3.2** Kinds considered and dropped: `quota_exceeded` (use `conflict` +
  prose in v1, add additively in A7 if needed — §6.12), `unavailable`
  (→ `internal`/`upstream`), `not_implemented` (v1 has no stub endpoints,
  Q12), `precondition_failed` (→ `conflict`), `scope_denied` (folded into
  `forbidden`).
- **A0-3.3** A client receiving a kind it does not know MUST treat it as
  `internal` (terminal) and MUST NOT fail to parse the body (Q13 additive-only:
  new kinds may appear in `/api/v1`).
- **A0-3.4** `message` is human prose: it MUST follow ADR-0019 §2
  (`component.Function: what was attempted: key identifiers: cause`), MUST be
  self-contained for troubleshooting, and MUST NOT be parsed programmatically,
  matched by substring, or used as a control input. A message MAY echo untrusted
  request material (target, tool name, field name) **except a value rejected by
  a secret-pattern rule, which MUST NOT be echoed in whole, in part or as a
  digest (A2-9.5, A1-4.9)**; consumers MUST treat it as
  untrusted content when rendering or feeding it to a model (SPEC §6, ADR-0018
  §4).
  Tests: TestSecretScanNamesFieldNotValue, TestNoSecretValueOrDigestInError.
- **A0-3.5** For `internal` and `upstream` the cause segment MAY be generalized
  to the subsystem (`postgres: statement failed`, `llm: endpoint returned 500`)
  when the underlying text is not platform-controlled; the full cause chain
  stays in the log (ADR-0019 §2–§3). A message MUST NOT contain a stack trace,
  SQL text, a request body, or an upstream response body.
- **A0-3.6** Envelope: one top-level `error` object (§4). `attrs` carries the
  correlation ids known at the failure point (ADR-0019 §3) — `engagement_id`,
  `run_id`, `job_id`, `node_id` — each absent when unknown (A0-8.3). In error
  and `slog` attributes `node_id` denotes the **remote agent node**
  (`slp_node_`, Q9); a graph node MUST be `graph_node_id` (`gn_`). The two MUST
  NOT be conflated. _ADR-0019 §3 lists the attribute as "node", which predates
  the graph vocabulary of ADR-0016._
- **A0-3.7** No secret material in an error or its log record: no credential,
  token, session cookie, captured hash, per-run masking-mapping value
  (ADR-0020 §3), pseudonymization table entry, or raw model/tool payload
  (ADR-0019 §5, ADR-0020 §4). The contract suite asserts this
  (`contracts/README.md`: secret-free serialization).
- **A0-3.8** Log-or-return, once (ADR-0019 §3): the handler that converts an
  error into an HTTP response is the single place that logs it, at error level,
  with the correlation attrs and the error chain. Inner layers wrap and return.
- **A0-3.9** Object-access outcomes: a syntactically invalid id → `validation`
  (400); a well-formed id that does not exist → `notfound` (404); a well-formed
  id that exists but lies outside the caller's engagement scope → **`notfound`
  (404), never `forbidden`** — the existence of another engagement's objects
  MUST NOT be disclosed (SPEC C8, adversarial A12). `forbidden` is reserved for
  principal-level denial that names no object (A0-3.1).
- **A0-3.10** Transport: `Content-Type: application/json; charset=utf-8`,
  status per table. For `rate_limited` the envelope MUST carry `retry_after_ms`
  and the response MUST carry `Retry-After` in whole seconds equal to
  `ceil(retry_after_ms/1000)`; the two MUST agree. A rate-limit response MUST
  carry `retry_after_ms ≥ 1` and `Retry-After ≥ 1` (`ceil(retry_after_ms/1000)`):
  the `≥ 1` floor **is** the field's presence rule, so the `omitempty` tag on
  `RetryAfterMS` in the §4 sketch can never drop a rate-limit value, and a
  `rate_limited` envelope without it is a platform defect → `internal`. The HTMX
  face (ADR-0011)
  MUST NOT emit the envelope: it renders `message` through `html/template`
  (auto-escaped) and branches on `kind` server-side, never in client JS.
  Tests: TestRetryAfterPresence, TestRetryAfterAgreement.
- **A0-3.11** Retryability is per the table: `timeout`, `upstream`,
  `rate_limited` are retried with bounded exponential backoff plus jitter,
  honouring `retry_after_ms`; every other kind is terminal for the same
  request. Retrying a non-idempotent write is only safe where the owning
  contract defines a deduplication key — A1 MUST define one for `events:append`
  (Q6 worker principal, offline buffering per ADR-0013). An owning contract MAY
  declare one specific write retryable under `internal` where that contract
  defines a deduplication key (A1-7.6); the retry MUST reuse that key. Every
  other kind remains terminal.
- **A0-3.12** An error kind MUST NOT be added, renamed, or repurposed inside
  `/api/v1` except additively (A0-6.5); a kind MUST NOT be reused for a
  different remedy, because agent behaviour is keyed to it.

### A0-4 · Pagination

- **A0-4.1** Every list in `/api/v1` uses cursor paging with one envelope:
  `{"items":[…],"next_cursor":"…"}`. There is no offset paging and **no total
  count** in v1. _Counting an append-only table is an unbounded query, and Q1
  deliberately ships no free-form query endpoint._
- **A0-4.2** `next_cursor` is **absent** when the collection is exhausted
  (A0-8.3). A client MUST stop when it is absent and MUST NOT infer
  completeness from `len(items) < limit`.
- **A0-4.3** Each paginated collection MUST declare a total, stable order in
  its own contract (A1, A4) built from an immutable unique key (append sequence
  or id) — never from a mutable field (`status`, `updated_at`, the quarantine
  or report-inclusion flag an operator changes per Q5). Text keys MUST compare
  byte-wise (`COLLATE "C"`, A0-1.9). Every paginated collection MUST have an
  integer primary ordering key so A0-4.4's `{id,k}` cursor is expressible; a
  text key is only the tie-breaker carried in `id`. A collection whose order key
  is text MUST declare its own cursor shape in its own contract; A0-4.4's
  two-key set is closed.
- **A0-4.4** A cursor is base64url-unpadded (A0-8.5) of the **canonical JSON**
  (A0-2) object `{"id":"<id of last row>","k":<ordering value of last row>}`
  (canonical order puts `id` first, A0-2.4) — the complete ordering tuple of the
  last returned row. Cursors are opaque: replay byte-for-byte, never decode,
  construct or modify. No MAC is applied.

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
- **A0-4.5** `limit` is a query parameter: absent → **100** (the default), hard
  maximum **1000**. A `limit` that is present but unparseable, ≤ 0,
  non-integer, or > 1000 MUST be rejected with `validation` — never silently
  clamped. _Silent clamping lets an agent believe it saw the whole collection
  (Q3: never trust client discipline) — **PO confirm**._
  Tests: TestLimitValidation, TestLimitAboveMaxRejectedNotClamped.
- **A0-4.6** Has-more detection: read `limit+1` rows, return the first `limit`,
  set `next_cursor` **iff** row `limit+1` existed. Contract test: a last page
  that exactly fills `limit` MUST NOT carry `next_cursor`.
  Tests: TestHasMoreDetection.
- **A0-4.7** Mid-paging semantics: the append-only collections (event log, Q11;
  graph, which revises by adding a new node with a `supersedes` edge rather
  than overwriting — ADR-0016 §4, Q2) give a cursor the semantics of a **stable
  watermark**: no row is skipped or duplicated, and rows appended after the
  cursor appear on later pages. This is the easy case and it is the reason the
  platform has no snapshot/isolation knob. For any collection with mutable
  rows, no snapshot is guaranteed: a client MUST deduplicate by id and MUST NOT
  assume a paged set reflects one point in time.
- **A0-4.8** Cursor decode failure (bad base64 variant, non-canonical JSON,
  wrong key set, id of the wrong type per A0-1, the cursor's `id` does not
  resolve in this collection, or its `k` disagrees with the ordering value of
  the row that `id` resolves to — A0-4.4) → `validation` telling the client to
  restart from the first page. Cursors are not versioned and MUST NOT be cached
  across platform releases.

### A0-5 · Time

- **A0-5.1** Every timestamp in a payload, cursor, log attribute, and report is
  RFC 3339, UTC, trailing `Z`, exactly three fractional digits:
  `2026-09-07T14:03:22.481Z`, matching
  `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z$`.
- **A0-5.2** One precision for the whole platform: **milliseconds**. _ms is the
  resolution of the A0-1.1 body timestamp and of every clock we read; µs/ns
  would make canonical bytes encoder-dependent and burn the 512 B summary
  budget (A0-7) — **PO confirm**._
- **A0-5.3** Parsing MUST reject, not normalize (Q3 philosophy): any other
  precision, a numeric offset (`+02:00`), lowercase `t`/`z`, a space separator,
  a leap second (`:60`), or a year outside `[2020, 2100)` → `validation`.
  Implementation: parsing is three checks in order — (1) the byte-exact A0-5.1
  regex `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z$` with the seconds group
  additionally range-checked `00`–`59` (this is what forces `Z`, exactly three
  digits, and rejects `:60`); (2) `time.Parse(TimeLayout, s)` for calendar
  validity; (3) the year window `[MinYear, MaxYear)`. `time.Parse` alone MUST NOT
  be used: `Z07:00` accepts numeric offsets and Go normalizes leap seconds. A
  parsed value MUST round-trip: `FormatTime(ParseTime(s)) == s` byte-exactly,
  else `validation`. `time.RFC3339Nano` MUST NOT be used — it drops trailing
  zeros.
  Tests: TestParseTimeRejects (table over the six rejection classes, incl. a
  round-trip subtest).
- **A0-5.4** Platform timestamps come from one injected clock (DESIGN §4, §8)
  as `time.Now().UTC().Truncate(time.Millisecond)`. The monotonic clock reading
  MUST be stripped before a timestamp is stored, hashed, or serialized
  (`Truncate` strips it). Durations MAY be measured with the monotonic clock
  but are serialized as integer `*_ms` fields (A0-8.2) — never as timestamps
  and never as Go duration strings.

  `recorded_at` is platform-stamped and MUST be non-decreasing per chain: the
  writer applies the forward clamp `max(clock_now, prev_recorded_at)`. **The
  clamp MUST be bounded to 1000 ms**: beyond that the platform MUST use the true
  clock reading, MUST log at error level with the engagement/run correlation
  attributes (ADR-0019 §3), and `recorded_at` MAY then be non-monotone — A1-8.1's
  rule that only `seq` orders anything is the compensating control. Any
  client-claimed time (`*_claimed_at`) is clamped by the same bound so the two
  never diverge by more than it.
  Tests: TestRecordedAtClampIsMonotone, TestRecordedAtClampBypassRejected.
- **A0-5.5** No local time zones anywhere: no offset other than `Z`, no
  time-zone name field, no per-user conversion in v1. UI rendering MUST display
  the zone explicitly (rendering choice belongs to the UI session).
- **A0-5.6** Q11 authority rule: platform-recorded time is authoritative.
  Ordering, chain verification (Q11), approval expiry (ADR-0012 §7) and
  single-use approval checks (Q10) MUST use platform-recorded timestamps only.
  v1 has **no external anchoring** — no NTP attestation, no timestamp
  authority, no third-party anchor (Q11). The A0-5.4 forward clamp is bounded to
  1000 ms, so `recorded_at` MAY be non-monotone past that bound and `seq` stays
  the only ordering authority (A1-8.1).
- **A0-5.7** A client-supplied timestamp is untrusted input: it MUST live in a
  field whose contract marks it as such (recommended suffix `_claimed_at`),
  MUST NOT drive ordering, expiry, or any digest, and MUST be stored next to
  the platform-recorded `recorded_at` so divergence is visible in the audit
  trail (workers buffer offline — ADR-0013, Q9). _An agent could otherwise
  extend its own approval window._
- **A0-5.8** Declared consequence, not a defect: since the platform clock is
  the only anchor, whoever controls the platform host controls recorded time.
  Hash chaining (Q11) proves **order and immutability within a chain**, not
  absolute wall-clock truth. Changing that requires a new ADR.

### A0-6 · Unknown fields and versioning

Both rules, side by side — neither generalizes to the other:

| Direction | Unknown **field** | Unknown **enum value** |
|---|---|---|
| platform → client (read) | MUST ignore (A0-6.1, Q13) | MUST carry through unchanged, MUST NOT coerce, skip what cannot be handled (A0-6.3) |
| client → platform (write) | MUST reject: `validation` (A0-6.2) | MUST reject: `validation` (A0-6.3, Q3) |

- **A0-6.1** Reads: a client decoding a platform response MUST ignore unknown
  fields and MUST NOT enable strict decoding (`DisallowUnknownFields`) on
  responses. Contract test: injected unknown fields decode successfully and are
  dropped (`contracts/README.md` merge gate, Q13).
- **A0-6.2** Writes: the platform decoding a client request MUST reject unknown
  fields with `validation`, naming the offending field
  (`json.Decoder.DisallowUnknownFields`). Write-side key matching is byte-exact:
  a key differing from the declared `json` tag only by case is an unknown field
  and MUST be rejected with `validation` naming it. Because `encoding/json`
  matches case-insensitively, the decoder MUST verify observed key spelling (the
  same `Token()` walk A0-2.5 requires) rather than rely on
  `DisallowUnknownFields` alone.
  Tests: TestAppendRejectsUnknownField (incl. case variants),
  TestCaseDuplicateKeyRejected. _An LLM-driven client that
  hallucinates a field name must fail loudly; a silently dropped
  safety-relevant field is exactly the drift ADR-0018 exists to prevent.
  Consequence: the platform is upgraded **before** agent images — a newer agent
  image against an older platform fails fast instead of losing data (C7 pins
  agent images per run). Q3 mandates hard reject for graph writes; extending it
  to all writes is a recommendation — **PO confirm**._
- **A0-6.3** Enums: an unknown or malformed value of a closed list (node kind,
  edge kind, event kind, error kind, any status) on a **write** is hard-rejected
  with `validation` (Q3). On a **read**, a client MUST NOT coerce an unknown
  value to a known one, MUST preserve the raw string when echoing it, and MUST
  skip the item it cannot handle without aborting the page.
- **A0-6.4** There is **no** per-object `schema_version` field in v1.
  _Additive-only evolution plus the `/api/v1` path prefix already identifies
  the shape; a per-object version counter inside a hashed document would make
  every digest depend on a mutable field and would have to be excluded from the
  canonical form (A0-2.12) — a second source of truth for no gain — **PO
  confirm**._
- **A0-6.5** Additive-only inside `/api/v1` (Q13). **Additive:** a new response
  field, a new enum value, a new event/node/edge kind, a new error kind, a new
  endpoint, a new id prefix. **Breaking (→ `/api/v2` + new ADR):** removing,
  renaming, or repurposing any of those; changing a field's type, unit,
  precision, or default; changing an id prefix or body format; changing the
  canonical form (A0-2) or an exclusion list (A0-2.13). No formal deprecation
  window in v1 (Q13).
- **A0-6.6** For a type whose value is canonicalized (A1 events, A7
  fingerprints), evolution inside `/api/v1` happens by **adding kinds**, not by
  reshaping an existing kind's fields: a written record's canonical bytes are
  immutable (A0-2.16).
- **A0-6.7** The path prefix is the only version indicator: no version media
  type, no `Accept` negotiation, no `X-API-Version` header.

### A0-7 · Size caps

- **A0-7.1** The one cap registry: the Q4 constants plus every per-contract cap
  constant A1 and A2 declare — one table, one const block (§4). `KiB` = 1024
  bytes. **One constant per value:** where two contracts capped the same thing
  under different names or different numbers, the registry holds exactly one
  name and one value (`ToolVersionMaxBytes = 64` — A2-7.1's 32 is a defect;
  `EvidenceRefsMax = 8` replaces A2's `EvidenceIDsMax`), and both documents cite
  the A0 name. An owning contract MUST NOT restate or shadow a registry value.

  | Constant | Cap | Applies to | Mechanism | Source |
  |---|---|---|---|---|
  | `StageViewMaxBytes` | 65536 B (64 KiB) serialized JSON | one stage view document (A3) | A3 assigns (A0-7.7) | Q4 |
  | `StageViewMaxNodes` | 500 nodes | one stage view document (A3) | A3 assigns (A0-7.9) | Q4 |
  | `NodeSummaryMaxBytes` | 512 B | graph node `summary` (A2) | **R** (A2-7.1) | Q4 |
  | `FindingSummaryMaxBytes` | 2048 B (2 KiB) | `finding` node `summary` (A2, Q2) | **R** (A2-7.1) | Q4 |
  | `StageSummaryMaxBytes` | 2048 B (2 KiB) | per-stage summary in a view (A3) | A3 assigns (A0-7.7) | Q4 |
  | `EventMaxCanonicalBytes` | 32768 B | canonical bytes of one A1 event | platform invariant, not a client cap (A1-4.5) | A1 |
  | `ProseLongMaxBytes` | 2048 B | A1 long untrusted prose: `command`, `task_description`, `result_summary`, `revert_action` | **R** (A1-4.5) | A1 |
  | `ProseMediumMaxBytes` | 512 B | A1 medium prose: `reason`, `detail`, `action_summary`, `message`, `entry` | **R** (A1-4.5) | A1 |
  | `TargetMaxBytes` | 256 B | A1 target text: `target`, `attempted_target`, `blacklist_entry` | **R** (A1-4.5) | A1 |
  | `LabelMaxBytes` | 128 B | A1 platform/config labels: `origin`, `container_ref`, `network_name`, `media_type`, `model_name`, `endpoint_name`, `target_name`, `old_value`, `new_value` | **R** (A1-4.5) | A1 |
  | `ToolVersionMaxBytes` | 64 B | `tool_version` (A1-4.5, A2-7.1) — one value for both | **R** | A1 + A2 |
  | `KindNameMaxBytes` | 32 B | A1 `node_kind`, `edge_kind`, `risk_tier` (A2/A7 own the values) | **R** (A1-4.5) | A1 |
  | `DigestMaxBytes` | 256 B | `image_digest` (A0-8.7) | **R** (A1-4.5) | A1 |
  | `EvidenceRefsMax` | 8 | A1 `evidence_refs` count (A1-4.7) and A2 `evidence_ids` count — one value, A2's `EvidenceIDsMax` name is dropped | **R** | A1 + A2 |
  | `EventRefsMax` | 64 | A1 `revert_event_ids` / `non_revertable_event_ids` count | **R** (A1-4.7) | A1 |
  | `ExitCodeMin` / `ExitCodeMax` | -1 / 255 | A1 `exit_code` range (`-1` = no exit status) | reject → `validation` (A1-4.2) | A1 |
  | `IdempotencyKeyMaxBytes` | 64 B | A1-7.6 client deduplication key | reject → `validation` (A1-7.6) | A1 |
  | `NodeLabelMaxBytes` | 128 B | A2 node `label` | **R** (A2-7.1) | A2 |
  | `HypothesisClaimMaxBytes` | 512 B | A2 `hypothesis.claim` | **R** (A2-7.1) | A2 |
  | `HypothesisBasisMaxBytes` | 1024 B | A2 `hypothesis.basis` | **R** (A2-7.1) | A2 |
  | `AttrValueMaxBytes` | 512 B | A2 `attrs` string value | **R** (A2-6.4) | A2 |
  | `AttrsTotalMaxBytes` | 4096 B | A2 serialized `attrs` object | **R** (A2-6.4) | A2 |
  | `AttrsMaxKeys` | 16 | A2 `attrs` key count | **R** (A2-6.4) | A2 |
  | `AttrKeyMaxBytes` | 40 B | A2 `attrs` key length — the bound A0-8.1's `^[a-z][a-z0-9_]{0,39}$` already fixes; A2-6.2 declares no separate A2 constant | reject → `validation` (A2-6.2) | A0-8.1 |
  | `AddressesMax` | 16 | A2 `addresses` count | **R** (A2-7.1) | A2 |
  | `AddressMaxBytes` | 64 B | A2 `addresses` entry | **R** (A2-7.1) | A2 |
  | `MaxSupersedeChain` | 64 | A2 `supersedes` history length (A2-4.4/4.5) | reject at write → `conflict` (A2-4.5) | A2 |

  Tests: TestConstantsMatchA0Table (one assertion per registry row),
  TestCapsRejectWithSummaryTooLarge.

- **A0-7.2** These values — the Q4 constants and every adopted A1/A2 cap — are
  contract constants: they MUST NOT change except by a new ADR (Q4). Lowering a
  cap additionally requires a migration plan for already-stored over-cap rows.
  An owning contract cites the A0-7.1 name; a second definition of the same
  value anywhere in the platform is a defect.
- **A0-7.3** Measurement: a **field** cap counts the UTF-8 bytes of the decoded
  string value (`len(s)` in Go) — surrounding quotes and escapes do not count.
  A **document** cap counts the UTF-8 bytes of the serialized JSON document as
  transmitted; implementations MUST measure after serialization, with the same
  encoder settings used for the response. Never rune counts, never UTF-16
  lengths. _A3 SHOULD transmit stage views in canonical form (A0-2) so the
  measurement is reproducible across implementations — §6.13._
- **A0-7.4** Truncation MUST cut on a UTF-8 rune boundary: a split rune yields
  invalid UTF-8 and breaks A0-2.3.
- **A0-7.5** **Mechanism T — truncate + marker.** The platform shortens the
  value so that `len(prefix) + len("[truncated]") ≤ cap`, appends the literal
  ASCII marker `[truncated]` (12 B, counted against the cap), and sets the
  sibling boolean field `<field>_truncated` to `true`. A cap smaller than
  `len(TruncationMarker)` is a platform defect: `caps.Truncate` returns
  `("", true)` (§4) and the caller surfaces `internal` (A0-3.1) — an empty
  value with the marker set is preferable to a value that silently exceeds its
  cap. The boolean is the machine-readable signal; the marker is prose for
  humans and reports and MUST NOT be parsed (A0-3.4). Used where a shortened
  value is still useful (handover views, model context hygiene).
  Tests: TestTruncateRuneBoundary, TestNoSilentTruncationInHashedRecords.
- **A0-7.6** **Mechanism R — reject.** The platform refuses the write with
  `summary_too_large` (413), stores nothing, and the message names the field,
  the cap, and the actual byte count (ADR-0019 §2). Used where a silently
  shortened value would mislead an approver (ADR-0018 §1: operators approve the
  concrete action) or corrupt evidence (ADR-0016 §1: graph content is evidence).
- **A0-7.7** **A1, A2 and A3** — every contract with capped fields — MUST
  assign exactly one mechanism to every capped field class and record the
  assignment in their own contract. A0 defines the mechanisms and holds the
  registry (A0-7.1); the mechanism a class gets stays the owning contract's
  assignment. Recommended default: client-submitted summaries → R;
  platform-computed view fields → T (**PO confirm**).
- **A0-7.8** The enforcement point is the platform ingest/render path, never
  the producer (Q3: never trust worker discipline). A client-side pre-check MAY
  exist and is not enforcement.
- **A0-7.9** `StageViewMaxNodes` is a count, not bytes: a view builder MUST
  stop adding nodes at 500 in its declared deterministic order (A0-4.3) and set
  `nodes_truncated` (mechanism T). Whether the remaining nodes are reachable
  through `next_cursor` (A0-4) is A3's decision.
- **A0-7.10** Declared conflict, **not** resolved here: the Q4 caps are
  independent and not jointly satisfiable for a maximal view — 500 nodes × 512 B
  ≈ 250 KiB ≫ 64 KiB, i.e. the 64 KiB budget allows ~131 B per node. A3 MUST
  define the composition rule (e.g. compact node references in the view, full
  summaries only in the capped 1-hop drill-down of Q1). Escalated to the
  product owner (§6.14).

  Interim composition rule for the Freeze (escalated, §6.14): where two A0-7.1 caps
  apply to one composed document and cannot both hold, the **smaller** governs; the
  builder MUST stop at the first cap it reaches, in its declared deterministic order
  (A0-4.3), and set the truncation marker (mechanism T). A3 owns the composition rule
  and MAY raise it only via ADR (A0-7.2).

### A0-8 · Field and enum conventions

- **A0-8.1** JSON keys are `snake_case` ASCII matching `^[a-z][a-z0-9_]{0,39}$`;
  Go fields are MixedCaps with acronyms upper-case (`ID`, `URL`, `HTTP`, `TOTP`
  — DESIGN §9). Every contract field carries an explicit `json:"…"` tag; no key
  is inferred from the Go identifier. Key spelling on a write is byte-exact
  (A0-6.2): a key that differs from the declared tag only by case is an unknown
  field, not a match.
- **A0-8.2** Suffixes are fixed and MUST NOT be used with another meaning:
  `*_id` (identifier, A0-1) · `*_ids` (array of identifiers, sorted and
  deduplicated per the owning contract) · `*_at` (timestamp, A0-5) · `*_ms`
  (duration in milliseconds, integer) · `*_bytes` (byte count, integer) ·
  `*_hash` (lowercase hex digest, A0-2.15) · `sha256` (artifact-integrity
  digest, ADR-0009 §2) · `*_truncated` (bool, A0-7.5) ·
  `*_claimed_at` (untrusted client-supplied timestamp, A0-5.7). `evidence_refs`
  (A1-1.1) is the one approved exception to the suffix rule: it is an array of
  `evi_` ids whose name does not end in `_ids`, and it cannot be renamed
  because it is inside the digest (A0-2.14, A1-5.2).
- **A0-8.3** Absent vs null: `null` MUST NOT appear in contract JSON in v1. An
  unset optional field is **absent**; an empty collection is `[]` or `{}`. A
  client MUST treat a received `null` as absent; the platform MUST reject a
  `null` for a known field on a write with `validation`. `cjson.Canonical` and
  `cjson.CanonicalValue` reject a `null` at **any** depth, not only on known
  fields (A0-2.14). Canonicalized types are
  stricter — fixed key set, zero values instead of absence (A0-2.14).
- **A0-8.4** `""` is a value, not "unset": an optional string field is absent
  when unset (A0-8.3). _The two have different canonical bytes and therefore
  different digests (A0-2.14)._
- **A0-8.5** Closed enums are lowercase `snake_case` ASCII strings matching
  `^[a-z][a-z0-9_]{0,31}$` — no dots, no hyphens, no `SCREAMING_CASE`
  (ADR-0012 §2 already ships `approval_required`, `hard_stop_fired`). Go
  representation: `type NodeKind string` plus exported constants — never an int
  enum, because the value must stay readable in logs, approval views and
  reports (ADR-0019). Comparison is byte-exact: no case folding, no synonyms,
  no alias lists.
- **A0-8.6** Binary inside a JSON string is base64url **without padding**
  (RFC 4648 §5, `base64.RawURLEncoding`, alphabet `-`/`_`). Decoders MUST
  reject the standard alphabet (`+`/`/`) and MUST reject `=` padding. Used for
  cursors (A0-4.4) and opaque binary values. Evidence *files* are not JSON and
  are out of scope (ADR-0009).
  Tests: TestCursorRejectsStandardAlphabetAndPadding, TestEncodeCursorRoundTrip.
- **A0-8.7** Digests and MACs are lowercase hex (`encoding/hex`), 64 characters
  for SHA-256 — never base64, never uppercase. Container image digests are the
  exception: `image_digest` carries the registry's `<algorithm>:<hex>` form,
  validated as `^[a-z0-9]+(?:[._-][a-z0-9]+)*:[0-9a-f]{64}$` and capped by the
  owning contract (A1-4.5: 256 B, `DigestMaxBytes`). _Greppable and
  eyeball-comparable in logs, approval views and customer reports._
- **A0-8.8** Booleans are JSON `true`/`false` named as adjectives or
  participles (`quarantined`, `truncated`) — no `is_` prefix in JSON.
- **A0-8.9** All JSON bodies are `application/json; charset=utf-8`. Request
  bodies MUST be size-bounded before parsing (`http.MaxBytesReader`; A4 sets the
  per-endpoint numbers) and MUST be UTF-8 (A0-2.3). The handler MUST call
  `utf8.Valid` on the raw body before decoding and reject with `validation`;
  `cjson.Canonical` MUST call `utf8.Valid(doc)` (A0-2.3). Relying on the decoder
  is a defect.

## 4. Types

A0 requires six foundation packages; none imports another internal package except
`errs` (DESIGN §1 layering): `internal/ids` (A0-1), `internal/cjson` (A0-2),
`internal/errs` (A0-3), `internal/paging` (A0-4: `Page`, `Cursor`, `EncodeCursor`,
`DecodeCursor`), `internal/timex` (A0-5: `Clock`, `Now`, `FormatTime`, `ParseTime` —
the name avoids shadowing stdlib `time`, the same reasoning DESIGN §1 gives for
`logging`), `internal/caps` (A0-7: every cap constant of the A0-7.1 registry,
`TruncationMarker`, `Truncate`, `Fits`). The sketches below are illustrative and not
compiled (contracts/README.md §4); each carries its `package` line.

```go
// internal/ids — foundation: identifier generation and validation (A0-1).
package ids

const BodyLen = 26 // A0-1.1: 10 chars time + 16 chars crypto/rand

// Alphabet is lowercase Crockford base32 without i, l, o, u (A0-1.1).
// Byte-lexicographic order of the alphabet equals numeric value order, so
// equal-length bodies sort chronologically (A0-1.6: not an API guarantee).
const Alphabet = "0123456789abcdefghjkmnpqrstvwxyz"

type Kind string // engagement, run, job, task, event, node, graph node, ...

const (
	Engagement Kind = "eng_"
	Run        Kind = "run_"
	Job        Kind = "job_"
	Task       Kind = "task_"
	Event      Kind = "evt_"
	AgentNode  Kind = "slp_node_" // Q9
	GraphNode  Kind = "gn_"
	GraphEdge  Kind = "ge_"
	Evidence   Kind = "evi_"
	Approval   Kind = "apr_"
	Tool       Kind = "tool_"
	KindUser   Kind = "usr_" // human principal (SPEC §3, AM-1 — §6 item 15)
)

// New returns a fresh id of kind k. Errors only on entropy failure.
func New(k Kind) (string, error)

// Valid reports whether s is a byte-exact match for k's regex (A0-1.5):
// anchored, no Crockford normalization, no case folding.
func Valid(k Kind, s string) bool
```

```go
// internal/cjson — foundation: the one canonical JSON form (A0-2).
package cjson

const (
	MaxDepth = 32            // A0-2.11
	MaxBytes = 1 << 20       // A0-2.11
)

// Canonical re-serializes doc: keys in UTF-8 byte order, no whitespace,
// integers only, literal UTF-8 strings, top-level object required, duplicate
// and case-duplicate keys rejected, fields named in exclude dropped from the
// top level (A0-2.12). Any violation returns an errs kind "validation".
// Drives json.Decoder.Token() so every key is observed (A0-2.5), with
// UseNumber() and each number's literal token text re-validated and re-emitted
// verbatim (A0-2.5), Decoder.More() for trailing data (A0-2.3), its own nesting
// counter (A0-2.11), utf8.Valid(doc) plus a lone-surrogate scan (A0-2.3), and a
// rejection of null at any depth (A0-2.14).
func Canonical(doc []byte, exclude ...string) ([]byte, error)

// CanonicalValue marshals v with encoding/json and canonicalizes the result.
// encoding/json only ever produces the intermediate bytes: Canonical re-parses
// and re-emits them, so its U+2028/9 escaping and HTML escaping never reach the
// final bytes (A0-2.7). Every slice and map field of v MUST be non-nil
// (A0-2.14).
func CanonicalValue(v any, exclude ...string) ([]byte, error)

// With adds top-level fields to an already-canonical document and re-canonicalizes; it never
// decodes into a typed struct. A key already present in doc is an error (A1-1.8, A1-5.7).
func With(doc []byte, add map[string]any) ([]byte, error)

// SHA256Hex returns the 64-char lowercase hex digest (A0-2.15).
func SHA256Hex(b []byte) string

// DigestEqual compares two digests in constant time (A0-2.15).
func DigestEqual(a, b string) bool
```

```go
// internal/errs — the A0-3 client-facing envelope (ADR-0019).
package errs

type Kind string

const (
	Validation       Kind = "validation"
	Auth             Kind = "auth"
	Forbidden        Kind = "forbidden"
	NotFound         Kind = "notfound"
	Conflict         Kind = "conflict"
	ApprovalRequired Kind = "approval_required"
	ApprovalExpired  Kind = "approval_expired"
	IntegrityFailed  Kind = "integrity_failed"
	SummaryTooLarge  Kind = "summary_too_large"
	Timeout          Kind = "timeout"
	Upstream         Kind = "upstream"
	RateLimited      Kind = "rate_limited"
	Internal         Kind = "internal"
)

// Envelope is the /api/v1 error body (A0-3.6). Absent attrs are omitted,
// never null (A0-8.3). RetryAfterMS is present for rate_limited only, and the
// ≥ 1 floor of A0-3.10 is what makes omitempty safe for it.
type Envelope struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Kind         Kind       `json:"kind"`
	Message      string     `json:"message"`      // prose, never parsed (A0-3.4)
	Attrs        Attrs      `json:"attrs"`
	RetryAfterMS int64      `json:"retry_after_ms,omitempty"` // ≥ 1 whenever present (A0-3.10)
}

// Attrs are the ADR-0019 §3 correlation ids. NodeID is the remote agent node
// (slp_node_, Q9); a graph node is GraphNodeID (gn_) and is set only by
// graph-facing handlers (A0-3.6).
type Attrs struct {
	EngagementID string `json:"engagement_id,omitempty"`
	RunID        string `json:"run_id,omitempty"`
	JobID        string `json:"job_id,omitempty"`
	NodeID       string `json:"node_id,omitempty"`
	GraphNodeID  string `json:"graph_node_id,omitempty"`
}

// Status maps a kind to its HTTP status (A0-3.1 table) — many-to-one.
func (k Kind) Status() int
```

```json
{
  "error": {
    "kind": "summary_too_large",
    "message": "graph.WriteNode: node summary exceeds cap for engagement=eng_01m1y2whfhgbz06ays6dxnvyws graph_node=gn_01m1y2whfhh039ykj5x8mc5a0g: field=summary cap=512 actual=613 bytes",
    "attrs": {
      "engagement_id": "eng_01m1y2whfhgbz06ays6dxnvyws",
      "run_id": "run_01m1y2whfhnjx2am9103w0pnqw",
      "job_id": "job_01m1y2whfhbt69j0h0fbxepw90",
      "graph_node_id": "gn_01m1y2whfhh039ykj5x8mc5a0g"
    }
  }
}
```

```json
{
  "error": {
    "kind": "rate_limited",
    "message": "api.EventsAppend: event append rate limit exceeded for job=job_01m1y2whfhbt69j0h0fbxepw90 engagement=eng_01m1y2whfhgbz06ays6dxnvyws: retry in 5000 ms",
    "attrs": {
      "engagement_id": "eng_01m1y2whfhgbz06ays6dxnvyws",
      "job_id": "job_01m1y2whfhbt69j0h0fbxepw90"
    },
    "retry_after_ms": 5000
  }
}
```

```go
// internal/paging — foundation: the one list envelope and its cursors (A0-4).
package paging

// Page is the one list envelope (A0-4.1). NextCursor absent = exhausted.
type Page[T any] struct {
	Items      []T    `json:"items"`
	NextCursor string `json:"next_cursor,omitempty"`
}

// Cursor is the canonicalized (A0-2) then base64url-unpadded payload (A0-4.4).
type Cursor struct {
	K  int64  `json:"k"`  // ordering value of the last returned row (A0-4.3)
	ID string `json:"id"` // id of the last returned row (A0-1)
}

func EncodeCursor(c Cursor) (string, error) // cjson.CanonicalValue + RawURLEncoding

// DecodeCursor validates the decoded cursor and validates its id against k's
// regex (A0-1.5); any deviation is errs.Validation (A0-4.8).
func DecodeCursor(s string, k ids.Kind) (Cursor, error)
```

```json
{
  "items": [
    {
      "event_id": "evt_01m1y2whfhp17g0avdqztd2p3x",
      "kind": "tool_invoked",
      "recorded_at": "2026-09-07T14:03:22.481Z"
    }
  ],
  "next_cursor": "eyJpZCI6ImV2dF8wMW0xeTJ3aGZocDE3ZzBhdmRxenRkMnAzIiwiayI6NDcxMX0"
}
```

```go
// internal/timex — foundation: the one layout, clock and precision (A0-5).
// The name avoids shadowing stdlib `time` (same reasoning as `logging`).
package timex

// Time (A0-5). One layout, one clock, one precision.
const (
	TimeLayout = "2006-01-02T15:04:05.000Z07:00" // A0-5.1; prints Z only for UTC
	MinYear    = 2020                            // A0-5.3
	MaxYear    = 2100                            // A0-5.3 (exclusive)
)

type Clock interface{ Now() time.Time } // injected (DESIGN §4, §8)

// Now returns UTC, millisecond-truncated, monotonic reading stripped (A0-5.4).
func Now(c Clock) time.Time { return c.Now().UTC().Truncate(time.Millisecond) }

func FormatTime(t time.Time) string          // A0-5.1
func ParseTime(s string) (time.Time, error)  // A0-5.3: reject, never normalize
```

```go
// internal/caps — foundation: the A0-7.1 registry, mechanism T and mechanism R.
package caps

// Size caps (A0-7.1) — the one registry: Q4 contract constants plus every
// per-contract cap A1 and A2 declare. Changeable only via ADR (A0-7.2); an
// owning contract cites these names and never restates a value.
const (
	// Q4 (A0-7.1).
	StageViewMaxBytes      = 64 * 1024 // 65536 B serialized JSON
	StageViewMaxNodes      = 500       // count, not bytes (A0-7.9)
	NodeSummaryMaxBytes    = 512       // B, UTF-8 of the decoded value
	FindingSummaryMaxBytes = 2 * 1024  // 2048 B
	StageSummaryMaxBytes   = 2 * 1024  // 2048 B

	// A1 (A1-4.5/4.7) — mechanism R everywhere.
	EventMaxCanonicalBytes = 32768 // platform invariant on one canonical event
	ProseLongMaxBytes      = 2048  // command, task_description, result_summary, revert_action
	ProseMediumMaxBytes    = 512   // reason, detail, action_summary, message, entry
	TargetMaxBytes         = 256   // target, attempted_target, blacklist_entry
	LabelMaxBytes          = 128   // origin, container_ref, network_name, media_type, ...
	ToolVersionMaxBytes    = 64    // tool_version; one value for A1 and A2 (A2's 32 is a defect)
	KindNameMaxBytes       = 32    // node_kind, edge_kind, risk_tier (A2/A7 own the values)
	DigestMaxBytes         = 256   // image_digest (A0-8.7)
	EvidenceRefsMax        = 8     // A1 evidence_refs / A2 evidence_ids count (was EvidenceIDsMax)
	EventRefsMax           = 64    // revert_event_ids / non_revertable_event_ids count
	ExitCodeMin            = -1    // -1 = no exit status (A1-4.2)
	ExitCodeMax            = 255
	IdempotencyKeyMaxBytes = 64 // A1-7.6 client dedup key

	// A2 (A2-6.4/7.1) — mechanism R everywhere.
	NodeLabelMaxBytes       = 128
	HypothesisClaimMaxBytes = 512
	HypothesisBasisMaxBytes = 1024
	AttrValueMaxBytes       = 512
	AttrsTotalMaxBytes      = 4096
	AttrsMaxKeys            = 16
	AttrKeyMaxBytes         = 40 // = A0-8.1's key regex bound (A2-6.2)
	AddressesMax            = 16
	AddressMaxBytes         = 64
	MaxSupersedeChain       = 64 // A2-4.4/4.5: bound on one history walk
)

const TruncationMarker = "[truncated]" // A0-7.5, 12 B, counted against the cap

// Truncate applies mechanism T: rune-boundary cut + marker (A0-7.4/7.5).
// Reports whether anything was cut, for the sibling <field>_truncated bool.
// `limit` (not `cap`, which shadows the builtin) < len(TruncationMarker) is a
// platform defect: Truncate returns ("", true) and the caller surfaces
// errs.Internal (A0-7.5).
func Truncate(s string, limit int) (out string, truncated bool)

// Fits applies the measurement rule of A0-7.3 before mechanism R (A0-7.6).
func Fits(s string, cap int) bool
```

### 4.1 Contract tests (A0)

The shared contract-test suite (`contracts/README.md`) MUST implement these ids
for A0; the name is the identifier, one name per test, and the clause named is
the rule the test guards. A0-2.17's vectors are data, not a test id: the suite
asserts their canonical bytes, lengths and digests byte-exactly.

| Clause | Test ids |
|---|---|
| A0-1.4 | `TestIDGenerationUsesCryptoRand`, `TestUniquenessViolationIsInternal` |
| A0-1.5 | `TestValidRejectsNormalization`, `TestValidIsByteExact` |
| A0-1.9 | `TestIDOrderingMatchesByteOrderCollateC` (integration, opt-in per DESIGN §8) |
| A0-2.3 | `TestInvalidUTF8Rejected`, `TestLoneSurrogateRejected` |
| A0-2.5 | `TestDuplicateKeyRejected`, `TestCaseDuplicateKeyRejected`, `TestCanonicalEmitsNumberLiteralText`, `TestRejections` |
| A0-2.11 | `TestDepthLimit`, `TestSizeLimit` |
| A0-2.14 | `TestNilCollectionNeverSerializesAsNull` |
| A0-2.15 | `TestDigestEqual`, `TestGatingComparisonsAreConstantTime` |
| A0-3.4 | `TestSecretScanNamesFieldNotValue`, `TestNoSecretValueOrDigestInError` |
| A0-3.10 | `TestRetryAfterPresence`, `TestRetryAfterAgreement` |
| A0-4.4 | `TestDecodeCursorRejects`, `TestCursorWithInconsistentKAndIDRejected` |
| A0-4.5 | `TestLimitValidation`, `TestLimitAboveMaxRejectedNotClamped` |
| A0-4.6 | `TestHasMoreDetection` |
| A0-5.3 | `TestParseTimeRejects` (incl. a round-trip subtest) |
| A0-5.4 | `TestRecordedAtClampIsMonotone`, `TestRecordedAtClampBypassRejected` |
| A0-6.2 | `TestAppendRejectsUnknownField` (incl. case variants) |
| A0-7.1 | `TestConstantsMatchA0Table` (one assertion per registry row) |
| A0-7.4 / A0-7.5 | `TestTruncateRuneBoundary`, `TestNoSilentTruncationInHashedRecords` |
| A0-8.6 | `TestCursorRejectsStandardAlphabetAndPadding` |
| §4 `cjson.With` | `TestWithAddsKeys`, `TestWithRejectsExistingKey` |

Naming rulings carried by this table: `TestExactFullPageHasNoNextCursor` is the
same test as `TestHasMoreDetection` and is **not** a second id; the round-trip
check of A0-5.3 is a subtest of `TestParseTimeRejects`, not `TestTimeParse…`.
`TestCursorWithInconsistentKAndIDRejected` has exactly one oracle across A0, A1
and A2: `validation` (400), never "empty page or `validation`" (A0-4.4, A0-4.8).
The positive counterparts of the canonicalization rules are the A0-2.17 accept
vectors (V1–V8, the 32-deep document, the surrogate-pair string); the negative
counterparts are its rejection entries.

## 5. Traceability

| Source | Decision | Clauses |
|---|---|---|
| **Q3** | graph writes hard-reject schema violations; platform enforces caps (truncate + marker / `summary_too_large`), never trusts worker discipline | A0-6.3, A0-6.2, A0-2.5/2.6/2.7 (reject not normalize), A0-4.5 (no silent clamp), A0-5.3, A0-7.5, A0-7.6, A0-7.8, A0-3.1 (`validation`, `summary_too_large`) |
| **Q4** | size budgets as contract constants, changeable only via ADR | A0-7.1, A0-7.2, A0-7.3, A0-7.9, A0-7.10, A0-3.1 (`summary_too_large` → 413), §4 const block |
| **Q6** | machine-token exclusion; three enforcement layers incl. per-request engagement/run binding | A0-3.1 (`forbidden` = principal-level denial), A0-3.9 (cross-engagement → `notfound`, no existence disclosure), A0-4.4 (authz re-derived per page, never from the cursor), A0-1.7 (ids are not credentials), A0-3.11 (safe retries need an A1 dedup key for the report-only worker) |
| **Q11** | day-one hash chaining, per-engagement chains, verification at startup and before export, operator override logged, no external timestamp anchoring | A0-2.1, A0-2.12–A0-2.17 (canonical form, exclusions, stored canonical bytes, vectors), A0-3.1 (`integrity_failed` → 409), A0-5.6, A0-5.8, A0-6.6 |
| **Q13** | additive-only `/api/v1`, clients ignore unknown fields, breaking changes → `/api/v2` | A0-6.1, A0-6.3, A0-6.5, A0-6.6, A0-6.7, A0-3.3 (unknown kind tolerated), A0-1.10 (prefix stability) |
| **ADR-0018** | approvals bind to an action fingerprint; execution-time re-validation; fingerprint + action recorded together | A0-2.1, A0-2.12, A0-2.13, A0-2.14 (fixed key set: no digest drift), A0-2.15, A0-2.16, A0-1.3 (stable tool ids), A0-3.1 (`approval_required`, `approval_expired`), A0-7.6 (no silent truncation of what an operator approves) |
| **ADR-0019** | error kinds; self-contained greppable messages; slog with correlation attrs; log-or-return; mandatory redaction | A0-3.1, A0-3.4, A0-3.5, A0-3.6, A0-3.7, A0-3.8, A0-3.10, A0-8.5 (readable enum values), §4 `errs` sketch |
| Q9 | `slp_node_…` node identity over mesh; node theft risk accepted | A0-1.2, A0-1.8, A0-3.6 (`node_id` = agent node), A0-5.7 (offline buffering → untrusted claimed time) |
| Q10 / Q2 | single-use approvals; revision by new node, never overwrite | A0-3.1 (`conflict`), A0-5.6, A0-4.7 (graph is append-only for ordering) |
| Q12 / Q15 | full `/api/v1` from day one; A0 frozen first | A0-4.1, A0-6.7, A0-3.2 (no `not_implemented`) |
| ADR-0011 | one core, two faces; SSE + JSON API; rate limiting is part of the attack surface | A0-3.10 (UI face renders prose, never the envelope), A0-3.1 (`rate_limited`), A0-6.7 |
| ADR-0016 | graph is per-engagement evidence with provenance; no cross-engagement flow | A0-1.2 (`gn_`/`ge_`), A0-3.9, A0-7.6, A0-4.7 |
| ADR-0020 | pseudonymization per run; captured secrets never leave | A0-3.7 (mapping values are secret material), A0-8.2 |
| ADR-0010 / SPEC C1 | stdlib only; pgx is the sole exception behind the store seam | A0-2.5/2.6 (`json.Decoder.Token()`, integers only), A0-1.9 (`COLLATE "C"`), A0-5.3, §4 (no third-party module anywhere in A0) |
| SPEC §6 / C9 | untrusted content is never configuration; errors are machine-traceable | A0-3.4, A0-3.7, A0-5.7, A0-6.2 |

## 6. Open for product owner

Recommendations that are genuinely product-owner calls (naming, alphabets,
precision, algorithms). Each is marked **PO confirm** at the clause too; none is
decided silently.

1. **A0-1.1** Id body: lowercase Crockford base32, 26 chars (48-bit ms + 80-bit
   `crypto/rand`), ULID layout. _Sortable, stdlib-only, no coordination._
2. **A0-1.2** Prefix set. Only `eng_`, `run_`, `evt_` (next_steps §2.1) and
   `slp_node_` (Q9) pre-exist; `job_`, `task_`, `gn_`, `ge_`, `evi_`, `apr_`,
   `tool_` are new. Note `next_steps.md` says `node_` where Q9 says
   `slp_node_` — A0 follows Q9 (binding) and treats next_steps as stale.
3. **A0-1.3** Tool ids are ULID-style, not `tool_nmap` slugs; name/version stay
   separate registry fields. _A slug changes on rename; fingerprints must not._
4. **A0-2.4** Key order = UTF-8 byte order, deviating from RFC 8785 for keys
   with code points ≥ U+10000 (locked by vector V6). _Strict JCS would need
   hand-rolled UTF-16 conversion for keys we never produce._
5. **A0-2.6** Canonical numbers are **integers only**; floats and exponents are
   rejected platform-wide in canonicalized documents. _ES6 number serialization
   is high-review-bar hand-rolled code (`AGENTS.md`) for zero benefit in v1._
6. **A0-2.11** Canonicalization input limits: depth 32, size 1 MiB.
7. **A0-5.2 / A0-5.3** Timestamp precision **milliseconds**, fixed at three
   digits; validity window `[2020, 2100)`.
8. **A0-6.2** Unknown fields on **writes** are rejected (`validation`), not
   ignored — broader than Q3's literal graph-write scope — with the deploy
   consequence "platform before agent images".
9. **A0-6.4** No per-object `schema_version` in v1.
10. **A0-4.1 / A0-4.5** No total count; `limit` default 100, hard max 1000,
    hard-rejected rather than clamped.
11. **A0-3.1** HTTP mapping of the new kinds: `approval_required`,
    `approval_expired`, `integrity_failed` → 409; `summary_too_large` → 413;
    `rate_limited` → 429. Also the spelling `rate_limited` (not
    `too_many_requests`) and folding `scope_denied` into `forbidden`.
12. **A0-3.2** Spawn-quota rejection (ADR-0017 §2) uses `conflict` + prose in
    v1; a dedicated `quota_exceeded` kind would be added additively by A7.
13. **§4** New foundation packages `internal/ids` and `internal/cjson`
    (DESIGN §1 allows adding packages within the layer rules); A0-7.3's
    recommendation that A3 transmit stage views in canonical form.
14. **A0-7.10 — needs a decision, not a confirmation.** The Q4 caps cannot all
    be satisfied by a maximal stage view (500 × 512 B ≈ 250 KiB > 64 KiB;
    ~131 B per node). A3 needs a composition rule. **Interim rule applied for the
    Freeze (BLOCK-PO2, A0-7.10):** where two A0-7.1 caps cannot both hold for one
    composed document, the smaller governs and the builder truncates with mechanism
    T. The product owner MUST confirm this fail-safe default or replace it with A3's
    composition rule before A3 is drafted; ~131 B per node is not a usable view
    budget.
15. **AM-1 — resolved by default for the Freeze (PO confirm):** A0-1.2 registers
    the human-principal prefix `usr_` (`^usr_B{26}$`, 30 B) and A0 §4 adds
    `KindUser`. A0 owns id *shapes*; delegating the spelling to A5 would split
    A0-1.5 validation across two contracts. Every user-composed A1 kind
    (`actor.principal_id`, A1-2.2) and every A2 operator write (`user_id`,
    A2-5.3) validates against it. The product owner MUST confirm the prefix
    spelling before Frozen; it is additive-only afterwards (A0-1.10).
