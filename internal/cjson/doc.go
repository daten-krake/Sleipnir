// Package cjson is the platform's one canonical JSON form (A0-2): the
// byte-exact serialization every hashed, signed or byte-compared value goes
// through — the event hash chain (A1) and the approval action fingerprints of
// ADR-0018 (A7) are digests of these bytes.
//
// Layer: foundation (DESIGN §1). Like the other foundation packages it
// imports exactly one internal package, internal/errs (A0 §4 preamble); it
// never imports internal/ids, internal/timex or any other sibling, because
// two implementations of the canonical form mean two different chains
// (A0-2.1). Nothing in the platform re-serializes a hashed value by hand.
//
// The exported surface is exactly A0 §4's sketch and nothing more: MaxDepth,
// MaxBytes, Canonical, CanonicalValue, With, SHA256Hex, DigestEqual. There is
// no Token type, no Encoder, no option struct and no second canonicalizer —
// A0-2.1 exists so there is one. cjson.go holds the canonicalizer, digest.go
// the digest helpers, and doc.go this comment.
//
// # Clauses implemented here
//
// A0-2.3 UTF-8 without BOM, with invalid UTF-8, a BOM, empty input and
// trailing data rejected, and a raw-input scan of every \u escape for
// unpaired surrogates. A0-2.4 keys sorted by unsigned UTF-8 byte value.
// A0-2.5 the json.Decoder.Token() walk that observes every key (duplicate
// keys rejected under one key equivalence, foldKey's SimpleFold orbit — see
// the rulings), UseNumber plus the literal-token regex and
// range check, Decoder.More() for trailing data, and a self-counted nesting
// depth. A0-2.6 integers only in [-(2^53-1), 2^53-1]. A0-2.7 the string
// emitter (only ", \ and U+0000-U+001F escaped, short escapes, lowercase
// \u00xx, everything else literal, no HTML escaping, U+2028/U+2029/U+007F
// raw). A0-2.8 array order preserved. A0-2.9 no insignificant whitespace.
// A0-2.10 top-level object required. A0-2.11 the MaxDepth/MaxBytes bounds.
// A0-2.12 top-level-only exclusion. A0-2.14 null rejected at any depth.
// A0-2.15 SHA256Hex and the constant-time DigestEqual.
//
// # Clauses deliberately NOT implemented here (who owns them)
//
// A0-2.13 (an exclusion list is part of the digest's definition and is frozen
// with its contract): the owning contract document declares the names; this
// package only accepts a caller-supplied list. A0-2.14's fixed-key-set half
// (every declared field present, no omitempty, no absent optional): the type
// owner in the domain package. A0-2.16 (persist the exact canonical bytes and
// recompute the digest from them, never from a re-serialization): the store;
// this package's part is that Canonical is deterministic, which
// TestCanonicalIsStableAcrossRuns pins. A0-2.2's identity with RFC 8785 for
// integer-valued, BMP-only documents is a property of the rules above, not
// extra code.
//
// # Rulings
//
//   - Input limits are checked over the raw bytes before the decoder runs, so
//     nothing proportional to an over-size document is ever allocated
//     (A0-2.11). Depth cannot be known without parsing, so it is enforced at
//     container entry — before descending — which bounds allocation the same
//     way.
//   - Every json.Decoder failure and every encoding/json Marshal failure in
//     CanonicalValue and With maps to errs.Validation, never Internal: A0-3.1
//     classifies "canonicalization failure (A0-2)" as validation, and Q3 says
//     hard reject rather than normalize. Internal is reserved for the two
//     decisions that really are platform defects: With's existing-key
//     collision, and canonical bytes this package just produced failing to
//     decode. Because errs.Wrap inherits the cause's kind, a decoder failure
//     is rendered into the message (bounded, never the document) rather than
//     wrapped, so the kind stays validation and no caller can act on it as if
//     it were an internal error.
//   - With's existing-key collision is errs.Internal (principal ruling):
//     A1-1.8/A1-5.7 make that composition platform-only, so a collision is a
//     platform defect, not client input. The comparison is case-insensitive,
//     because A0-2.5 already makes "hash" and "Hash" the same key.
//   - DigestEqual requires both sides to decode to exactly sha256.Size bytes:
//     A0-2.15 fixes a digest's width, so an empty string, a truncated value or
//     "00" never compares equal — not even to itself. The check is
//     load-bearing, not cosmetic: subtle.ConstantTimeCompare reports two empty
//     slices as equal, so DigestEqual("", "") would otherwise be true.
//   - Excluded top-level fields are still scanned and validated before being
//     dropped; the token stream has to be consumed either way, and validating
//     keeps "an excluded field must not influence the digest" (A0-2.12) from
//     masking a malformed document — or a duplicate key — behind the exclusion.
//   - Key identity is one thing in this package: foldKey, which maps each
//     rune to the minimum of its unicode.SimpleFold orbit with ASCII letters
//     normalized to lowercase first, so foldKey(a) == foldKey(b) exactly when
//     strings.EqualFold(a, b) does — the equivalence encoding/json uses for
//     field matching, which is A0-2.5's rationale. strings.ToLower was a
//     weaker equivalence: it accepted {"s":1,"\u017f":2} (U+017F LONG S) and
//     {"µ":1,"Μ":2}, yet json.Unmarshal of `{"ſ":7}` matches a `json:"s"`
//     field, and that producer/verifier disagreement is what the clause
//     exists to prevent. The fold is applied at all three key-identity
//     sites: the duplicate check in object, With's rejectCollisions, and the
//     A0-2.12 exclusion match. Pairwise strings.EqualFold was refused: it is
//     quadratic in an object's keys and a 1 MiB hostile document carries
//     ~40k of them (measured: 111 ms at 4k keys, ~11 s extrapolated at 40k —
//     adversarial finding A14); folding each key once into the existing map
//     stays linear (measured: 4.1 ms to 7.4 ms at 40k keys). Ceiling: simple
//     folding cannot equate the full-folding expansions (U+00DF "ß" vs "ss",
//     the U+FB01 ligature vs "fi", U+0130 "İ" vs "i" plus combining dot);
//     those need CaseFolding.txt tables the stdlib does not ship, and
//     strings.EqualFold does not equate them either — this is the achievable
//     ceiling and it matches encoding/json, not a halfway house. A0-8.1
//     declares every contract key ^[a-z][a-z0-9_]{0,39}$, so foldKey's
//     no-byte-≥0x80-and-no-ASCII-uppercase fast path — the same string value,
//     zero allocation — is the normal path for everything the platform
//     produces. No folding is applied to the emitted key text.
//   - Exclusion matching folds both sides (A0-2.12 read against A0-2.5): the
//     declared names are folded in excludeSet and each top-level member's key
//     is folded before the lookup, so there is no second notion of "the same
//     key". Either side alone would be wrong: folding only the declared names
//     leaves Canonical(`{"Hash":1,"a":2}`, "hash") returning the document
//     unchanged — an excluded field influencing the digest, which A0-2.12
//     forbids — and folding only the member keys would silently stop
//     excluding "hash" from spelling variants of A1's event documents and
//     break the chain. Vector V5 is unaffected by the widening: its excluded
//     names (seq, prev_hash, hash) are already lowercase ASCII. Matching
//     stays whole-name-only (A0-2.12): hash never drops hash_x, xhash or HA,
//     whatever the case.
//   - A null at any depth is rejected with validation, not internal, and the
//     two sentences of A0-2.14 that look contradictory are settled by
//     A0-8.3: "the platform MUST reject a null for a known field on a write
//     with validation; cjson.Canonical and cjson.CanonicalValue reject a null
//     at any depth". The "a canonical document containing null is a platform
//     defect → internal" sentence therefore governs a CONSUMER that finds
//     null inside stored canonical bytes (A0-2.16), not this producer's
//     reject path.
//   - CanonicalValue's input is a Go value, so invalid UTF-8 in a string field
//     is replaced by U+FFFD by encoding/json before this package sees any
//     bytes; A0-2.3's rejection of invalid UTF-8 therefore governs documents,
//     not Go values. Field-level validation (A0-8.1) is where that is caught.
package cjson
