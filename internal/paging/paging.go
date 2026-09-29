package paging

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"strconv"

	"github.com/daten-krake/sleipnir/internal/cjson"
	"github.com/daten-krake/sleipnir/internal/errs"
	"github.com/daten-krake/sleipnir/internal/ids"
)

// DefaultLimit is the page size when a list request carries no limit
// parameter (A0-4.5), and MaxLimit is the hard maximum above which a limit is
// rejected rather than clamped. They are the one Go definition of the two
// numbers of A0-4.5 and of §6 item 10 (A0-7.2 forbids a second definition
// anywhere), so an owning contract cites these names and never restates a
// value. They are NOT in A0-7.1's registry table, which carries no page-limit
// row — although ExitCodeMin/ExitCodeMax (-1/255, reject → validation) shows a
// numeric range is a shape the registry does take; see doc.go for the erratum
// candidate.
const (
	DefaultLimit = 100
	MaxLimit     = 1000
)

// Page is the one list envelope of every /api/v1 collection (A0-4.1). There is
// no total count field and there will not be one additively: A0-4.1 forbids
// counting an append-only table.
//
// An empty page serializes as {"items":[]}, never {"items":null} — A0-2.14 and
// cjson reject a null at any depth, so NewPage never leaves Items nil on the
// success path. NextCursor is absent exactly when the collection is exhausted
// (A0-4.2, A0-8.3); a client MUST stop on absence and MUST NOT infer
// completeness from len(Items) < limit.
type Page[T any] struct {
	Items      []T    `json:"items"`
	NextCursor string `json:"next_cursor,omitempty"`
}

// Cursor is the complete ordering tuple of the last returned row (A0-4.4): the
// row's id and its immutable ordering value (A0-4.3), never a mutable field.
//
// K is bounded by cjson to [-(2^53-1), 2^53-1] (A0-2.6); EncodeCursor and
// DecodeCursor inherit that bound instead of re-stating it (A0-7.2).
type Cursor struct {
	K  int64  `json:"k"`  // ordering value of the last returned row (A0-4.3)
	ID string `json:"id"` // id of the last returned row (A0-1)
}

// EncodeCursor returns base64url-unpadded (A0-8.6) of the canonical
// JSON (A0-2) object {"id":"…","k":…} — canonical key order puts "id" first
// (A0-2.4), which cjson.CanonicalValue applies; the two keys are never sorted
// or assembled by hand here.
//
// No MAC is applied (A0-4.4): a cursor is a position hint, not a grant.
// Authorization is re-derived per page by the handler; nothing in this package
// can or should substitute for that.
//
// An error is errs.Validation, the kind A0-3.1 gives a canonicalization
// failure: the only reachable causes are a K outside A0-2.6's integer range and
// an ID long enough that the two-key document passes cjson's A0-2.11 1 MiB
// bound, both of which the caller's own row data must be fixed. The ID's shape
// is not validated here — only DecodeCursor takes an ids.Kind to validate it
// against (A0-4.4).
func EncodeCursor(c Cursor) (string, error) {
	doc, err := cjson.CanonicalValue(c)
	if err != nil {
		return "", errs.Wrapf(err, "canonicalizing the cursor of id %.64q and ordering value %d (A0-4.4)", c.ID, c.K)
	}
	return base64.RawURLEncoding.EncodeToString(doc), nil
}

// DecodeCursor decodes a cursor of the collection whose id kind is k, and
// returns errs.Validation naming the violated clause for every deviation
// (A0-4.8): the checks run in the order A0-4.4 lists them —
//
//  1. base64url without padding (A0-8.6). base64.RawURLEncoding is the
//     whole rejection: it rejects the standard alphabet's '+' and '/' and any
//     '=' padding, so no character scan of our own is added (DESIGN §2).
//  2. cjson.Canonical on the decoded bytes (A0-2, A0-4.8's "non-canonical
//     JSON"): duplicate and case-duplicate keys, trailing data, a non-object
//     top level, a null at any depth, non-integer numbers, the A0-2.6 integer
//     range, depth and size bounds.
//  3. a strict decode of the canonical bytes with DisallowUnknownFields into
//     pointer fields (A0-4.8's "wrong key set"): an extra key and a missing key
//     are both rejections, because A0-4.4's two-key set is closed.
//  4. ids.Valid(k, id) (A0-4.4, A0-1.5): byte-exact, no normalization.
//  5. re-encoding the decoded cursor must reproduce the input bytes exactly.
//     This is A0-4.4's "replay byte-for-byte" and the guard step 2 cannot be:
//     cjson.Canonical normalizes key order and whitespace rather than rejecting
//     them, and encoding/json matches keys case-insensitively, so
//     {"ID":"evt_…","k":1} passes steps 2 and 3 and is still not A0-4.4's
//     cursor. A0-6.2 requires verifying observed key spelling rather than
//     relying on DisallowUnknownFields alone; one re-encode comparison does
//     that, the key-order and whitespace rejection, and the base64url
//     canonical-encoding check in a single line.
//
// k is returned verbatim. Its agreement with the ordering value of the row the
// cursor's id resolves to is NOT checked here and cannot be: this package has
// no collection and no row lookup (A0 §4 gives it no store-shaped dependency).
// A0-4.4 assigns that half of A0-4.8 to whoever resolves a cursor against a
// collection, and must ignore k for seeking in any case. See doc.go.
//
// A rejected cursor is never echoed into the returned error (A0-3.5); the
// decoded id is echoed bounded, because ids are not secrets (A0-1.7) and
// A0-6.2 requires a validation error to name the offending value.
func DecodeCursor(s string, k ids.Kind) (Cursor, error) {
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		// Rendered, not wrapped: a base64.CorruptInputError is a foreign error,
		// and errs.Wrap inherits the cause's kind, which for a foreign error is
		// Internal (A0-3.3). A client's malformed cursor must stay validation
		// (A0-4.8). The rendered text is "illegal base64 data at input byte N" —
		// an offset, never the cursor bytes (A0-3.5).
		return Cursor{}, errs.Newf(errs.Validation,
			"decoding a %d-byte base64url cursor for collection kind %.16q: the value must be base64url with no padding and no +, / or =: %.200s (A0-8.6, A0-4.4)", len(s), string(k), err)
	}

	doc, err := cjson.Canonical(raw)
	if err != nil {
		return Cursor{}, errs.Wrapf(err, "canonicalizing the payload of a cursor for collection kind %.16q (A0-4.8, A0-2)", string(k))
	}

	// Pointer fields are what make an absent key observable: a value type would
	// silently decode a missing "k" to 0 and hand out a half-built cursor
	// (DESIGN §4).
	var payload struct {
		ID *string `json:"id"`
		K  *int64  `json:"k"`
	}
	dec := json.NewDecoder(bytes.NewReader(doc))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&payload); err != nil {
		// Rendered, not wrapped, for the same reason as the base64 failure: a
		// json error is foreign and would reclassify the client's cursor as an
		// internal platform defect (A0-3.1, A0-3.3).
		return Cursor{}, errs.Newf(errs.Validation,
			"decoding the payload of a cursor for collection kind %.16q: the key set is exactly {\"id\",\"k\"} with a string and an integer value: %.200s (A0-4.4, A0-4.8, A0-6.2)", string(k), err)
	}

	// The closed two-key set of A0-4.4. Inlined rather than factored out:
	// errs captures the calling function as the error's origin (ADR-0019 §1),
	// and a helper would label every one of these rejections with the helper's
	// name instead of paging.DecodeCursor.
	switch {
	case payload.ID == nil && payload.K == nil:
		return Cursor{}, errs.Newf(errs.Validation,
			"decoding a cursor with neither key for collection kind %.16q: the cursor key set is exactly {\"id\",\"k\"} (A0-4.4, A0-4.8)", string(k))
	case payload.ID == nil:
		return Cursor{}, errs.Newf(errs.Validation,
			"decoding a cursor without its \"id\" key for collection kind %.16q: the cursor key set is exactly {\"id\",\"k\"} (A0-4.4, A0-4.8)", string(k))
	case payload.K == nil:
		return Cursor{}, errs.Newf(errs.Validation,
			"decoding a cursor without its \"k\" key for collection kind %.16q: the cursor key set is exactly {\"id\",\"k\"} (A0-4.4, A0-4.8)", string(k))
	}

	c := Cursor{K: *payload.K, ID: *payload.ID}

	if !ids.Valid(k, c.ID) {
		return Cursor{}, errs.Newf(errs.Validation,
			"validating cursor id %.64q against collection kind %.16q: the id does not have that kind's A0-1.2 shape (A0-4.4, A0-1.5)", c.ID, string(k))
	}

	reencoded, err := EncodeCursor(c)
	if err != nil {
		return Cursor{}, errs.Wrapf(err, "re-encoding the decoded cursor of collection kind %.16q (A0-4.4)", string(k))
	}
	if reencoded != s {
		return Cursor{}, errs.Newf(errs.Validation,
			"re-encoding the decoded cursor of collection kind %.16q did not reproduce its input bytes: the payload is not the canonical JSON form of the cursor's two-key object, its key spelling differs, or its base64url is not the exact encoding (A0-4.4, A0-4.8, A0-6.2)", string(k))
	}
	return c, nil
}

// ParseLimit reads the limit query parameter (A0-4.5). present is whether the
// parameter appeared at all — "?limit=" is present and empty, and therefore
// rejected rather than defaulted; only an absent parameter yields DefaultLimit.
//
// A present limit must be a base-10 integer in [1, MaxLimit]. Anything else is
// errs.Validation with a returned int of 0, and no value is ever clamped
// (A0-4.5, decided by the product owner 2026-09-21): a silent clamp would let
// an agent believe it saw the whole collection. Unparseable covers "0", "-1",
// "+1", "1e3", "1.0", "0x10", " 1", "1 ", "" and non-ASCII digits, so a leading
// sign is refused even though strconv.Atoi accepts one.
func ParseLimit(raw string, present bool) (int, error) {
	if !present {
		return DefaultLimit, nil
	}
	if raw == "" || raw[0] == '+' || raw[0] == '-' {
		return 0, errs.Newf(errs.Validation,
			"parsing limit %.16q: the limit must be a base-10 integer in [1, %d] (A0-4.5)", raw, MaxLimit)
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		// Rendered, not wrapped: strconv's NumError is foreign, and wrapping it
		// would inherit Internal for what is malformed client input (A0-3.1).
		return 0, errs.Newf(errs.Validation,
			"parsing limit %.16q as a base-10 integer in [1, %d]: %.200s (A0-4.5)", raw, MaxLimit, err)
	}
	if n < 1 || n > MaxLimit {
		return 0, errs.Newf(errs.Validation,
			"parsing limit %d: out of [1, %d]; a limit is rejected, never clamped (A0-4.5)", n, MaxLimit)
	}
	return n, nil
}

// NewPage builds the envelope for a page of at most limit rows (A0-4.6). The
// caller has already read limit+1 rows: the presence of row limit+1 is the only
// has-more signal, so a last page that exactly fills limit carries no
// next_cursor. cursorOf returns the ordering tuple of a row, and is called only
// for the last returned row, rows[limit-1].
//
// A limit outside [1, MaxLimit] is errs.Validation and a zero Page, never a
// clamp: this is the second place A0-4.5 bites, and clamping here would
// silently truncate the caller's own row slice.
//
// A nil cursorOf is a wiring defect rather than client input (errs.Internal,
// A0-3.1) and is reported instead of being turned into a nil-func panic in a
// request path (ADR-0019 §6).
//
// Items shares the caller's backing array when the page is full; NewPage does
// not mutate rows (DESIGN §3).
func NewPage[T any](rows []T, limit int, cursorOf func(T) Cursor) (Page[T], error) {
	if limit < 1 || limit > MaxLimit {
		return Page[T]{}, errs.Newf(errs.Validation,
			"building a page of %d rows with limit %d: the limit must be in [1, %d] and is never clamped (A0-4.5, A0-4.6)", len(rows), limit, MaxLimit)
	}
	if len(rows) <= limit {
		items := rows
		if items == nil {
			items = []T{}
		}
		return Page[T]{Items: items}, nil
	}
	if cursorOf == nil {
		return Page[T]{}, errs.Newf(errs.Internal,
			"building a page of %d rows with limit %d: a further row exists but cursorOf is nil (A0-4.6)", len(rows), limit)
	}
	next, err := EncodeCursor(cursorOf(rows[limit-1]))
	if err != nil {
		return Page[T]{}, errs.Wrapf(err, "building a page of %d rows with limit %d from its last returned row (A0-4.6)", len(rows), limit)
	}
	return Page[T]{Items: rows[:limit], NextCursor: next}, nil
}
