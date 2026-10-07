package graph

import (
	"bytes"
	"encoding/json"
	"regexp"
	"slices"
	"strconv"

	"github.com/daten-krake/sleipnir/internal/errs"
)

// Attrs is the bounded flat escape hatch (A2-6): map[string]AttrValue with a
// depth of exactly one. Nested objects, arrays, null and floats are rejected
// with validation (A2-6.1, A2-1.7, A0-2.6); keys must match A0-8.1's
// ^[a-z][a-z0-9_]{0,39}$ and must not be a reserved field name (A2-6.2,
// A2-6.3). The wire and canonical form is {"<key>": <JSON string | integer |
// boolean>} — depth exactly one, no wrapper object.
//
// UnmarshalJSON is hand-written (a json.Decoder token walk over the object
// structure and keys, then a map decode whose values pass through
// AttrValue.UnmarshalJSON's raw-byte validation). MarshalJSON is deliberately
// NOT hand-written: encoding/json's map encoding already emits exactly the
// wire form — keys sorted, each value the bare scalar of
// AttrValue.MarshalJSON — and a hand-written marshaler that emitted {} for a
// nil Attrs would paper over the A0-2.14/A2-4.6 non-nil requirement, which
// must surface as a defect instead (ruling recorded in doc.go).
//
// The size caps of A2-6.4 (AttrsMaxKeys, AttrValueMaxBytes,
// AttrsTotalMaxBytes — all internal/caps registry rows, measured on the A0-2
// canonical form) are enforced by the write validation pass (WP-15, A2-10.2
// step 8), not here: this package does not import internal/caps, and a cap
// value restated here would be a second definition (A0-7.2).
type Attrs map[string]AttrValue

// AttrType is the Go-only discriminator of AttrValue: which of Str/Num/Bool
// is live. It is never serialized (A2-6.1).
type AttrType string

const (
	AttrString AttrType = "string"
	AttrInt    AttrType = "int"
	AttrBool   AttrType = "bool"
)

// AttrValue serializes as the BARE scalar of its live field (A2-6.1): the
// wire and canonical form of attrs is {"<key>": <string | integer | boolean>},
// depth exactly one, no wrapper object. Type is a Go-only discriminator and
// is never serialized, hence json:"-" on every field below and hand-written
// methods (P-51).
type AttrValue struct {
	Type AttrType `json:"-"` // which of Str/Num/Bool is live
	Str  string   `json:"-"` // ≤ AttrValueMaxBytes (A2-7.1; cap enforced by WP-15)
	Num  int64    `json:"-"` // within A0-2.6's [-(2^53-1), 2^53-1]
	Bool bool     `json:"-"`
}

// maxSafeAttrInt is A0-2.6's integer bound: canonical numbers live in
// [-(2^53-1), 2^53-1]. internal/cjson keeps the same bound unexported; this
// package needs it at the attrs decode/encode boundary, where the canonical
// document has not been built yet.
const maxSafeAttrInt = int64(1<<53 - 1)

// attrKeyRE is A0-8.1's key shape, which is also the A2-6.2 key length cap
// (AttrKeyMaxBytes = 40 is the bound this regex expresses — no separate
// length check, no second definition of the cap).
var attrKeyRE = regexp.MustCompile(`^[a-z][a-z0-9_]{0,39}$`)

// attrIntRE is A0-2.5/A0-2.6's canonical integer literal text: no leading
// zeros (except a bare 0), no '+', no fraction, no exponent, at most 16
// digits (the width of 2^53-1). A2-6.1 accepts "a JSON integer" — the
// canonical form is what enters content_hash, so the decoder accepts exactly
// the literals the canonicalizer would re-emit verbatim.
var attrIntRE = regexp.MustCompile(`^-?(0|[1-9][0-9]{0,15})$`)

// reservedAttrKeys is the normative, closed, additive-only reserved set of
// A2-6.3, transcribed byte-exactly from the contract (51 keys). Membership
// is byte-exact — no prefix, suffix or substring matching. seq stays reserved
// even though the graph field is graph_seq (A2-1.4a); node_id likewise — no
// A2 field bears the name, but A0-3.6 reserves it platform-wide for the
// remote agent node (A2 §6 item 19); operator_id is gone
// with the A2-5.3 rename to user_id; confidence is the A2-5.6 provenance
// grade's field name.
//
// The A2-6.3 sketch spells this as an exported var map. It is unexported
// here because an exported package-level map is mutable state any other
// package could write (DESIGN §4 outranks the illustrative sketch —
// authority order ADR > SPEC > DESIGN > contract, and the sketch is
// "not compiled" by contracts/README.md). Every consumer that needs it
// (WP-15's validation, WP-21's contract suite) lives in this package.
var reservedAttrKeys = map[string]bool{
	"id": true, "engagement_id": true, "seq": true, "graph_seq": true,
	"kind": true, "label": true, "summary": true, "attrs": true,
	"evidence_id": true, "evidence_ids": true, "addresses": true, "cidr": true,
	"port": true, "transport": true, "protocol": true, "sid": true,
	"domain": true, "credential_kind": true, "media_kind": true, "size_bytes": true,
	"severity": true, "claim": true, "basis": true, "status": true,
	"content_hash": true, "quarantined": true, "quarantine_reason": true,
	"report_excluded": true, "supersedes_id": true, "superseded_by_id": true,
	"provenance": true, "source_id": true, "target_id": true,
	"source_kind": true, "target_kind": true, "retracted": true,
	"principal_kind": true, "run_id": true, "job_id": true, "task_id": true,
	"agent_node_id": true, "user_id": true, "tool_id": true, "tool_version": true,
	"event_id": true, "recorded_at": true, "observed_claimed_at": true,
	"confidence": true, "graph_node_id": true, "graph_edge_id": true,
	"node_id": true,
}

// MarshalJSON emits the bare scalar of the live field (A2-6.1, P-51). A zero
// or unknown Type, and an out-of-range Num, are hand-built-value defects —
// a decoded AttrValue always carries one of the three discriminators and a
// Num inside A0-2.6 — so both are errs.Internal, never a silent null or a
// non-canonical number (A0-2.14's defect rule by analogy; A0-2.6).
func (v AttrValue) MarshalJSON() ([]byte, error) {
	switch v.Type {
	case AttrString:
		// encoding/json is the intermediate emitter only: on the fingerprint
		// path cjson re-parses and re-emits these bytes per A0-2.7, so its
		// HTML escaping never reaches canonical output.
		return json.Marshal(v.Str)
	case AttrInt:
		if v.Num > maxSafeAttrInt || v.Num < -maxSafeAttrInt {
			return nil, errs.Newf(errs.Internal,
				"marshaling attrs value: integer %d is outside the canonical range [-(2^53-1), 2^53-1] (A0-2.6): a hand-built AttrValue is a platform defect", v.Num)
		}
		return []byte(strconv.FormatInt(v.Num, 10)), nil
	case AttrBool:
		if v.Bool {
			return []byte("true"), nil
		}
		return []byte("false"), nil
	default:
		return nil, errs.Newf(errs.Internal,
			"marshaling attrs value: type discriminator %.16q is not one of string/int/bool (A2-6.1): a hand-built AttrValue with no live field is a platform defect",
			string(v.Type))
	}
}

// UnmarshalJSON accepts a JSON string, integer or boolean only and rejects
// null, floats, arrays and objects with validation (A2-6.1). Round-trip
// preserves Type (P-51).
//
// Strings are scanned for lone surrogates before decoding (A0-2.3 (b)):
// encoding/json replaces an unpaired \uD800-\uDFFF escape with U+FFFD
// instead of rejecting, and a silently normalized value would enter
// content_hash as something the writer never sent (Q3: reject, never
// normalize). Integers must arrive in A0-2.5's canonical literal text —
// "1.5", "1e3", "-0", "01", "+1" and out-of-range values are rejected, not
// converted. The error messages name the rejected class and the byte length,
// never the value of a string or the content of a container: attrs is
// untrusted input that may carry secret material, and the secret-scan
// rejection that follows at WP-15's step 10 is not the only reason to keep
// rejected bytes out of an error (A2-9.5's discipline applied early).
func (v *AttrValue) UnmarshalJSON(b []byte) error {
	s := bytes.TrimSpace(b)
	if len(s) == 0 {
		return errs.Newf(errs.Validation, "decoding attrs value: input is empty (A2-6.1)")
	}
	switch {
	case s[0] == '"':
		if err := rejectLoneSurrogates(s); err != nil {
			return err
		}
		var str string
		if err := json.Unmarshal(s, &str); err != nil {
			return errs.Newf(errs.Validation,
				"decoding attrs string value: %d input bytes are not a valid JSON string (A2-6.1)", len(s))
		}
		*v = AttrValue{Type: AttrString, Str: str}
		return nil
	case s[0] == 't' || s[0] == 'f':
		switch string(s) {
		case "true":
			*v = AttrValue{Type: AttrBool, Bool: true}
			return nil
		case "false":
			*v = AttrValue{Type: AttrBool, Bool: false}
			return nil
		}
		return errs.Newf(errs.Validation,
			"decoding attrs value %.24q: booleans are the exact literals true and false (A2-6.1)", string(s))
	case s[0] == '-' || (s[0] >= '0' && s[0] <= '9'):
		// "-0" matches the literal regex but is rejected outright (A0-2.6:
		// the literal -0 MUST be rejected) — the same explicit check
		// internal/cjson runs.
		if !attrIntRE.Match(s) || string(s) == "-0" {
			return errs.Newf(errs.Validation,
				"decoding attrs value %.24q: floats, exponents, leading zeros, '+', '-0' and non-integer forms are rejected; attrs integers match the canonical A0-2.6 literal form",
				string(s))
		}
		n, err := strconv.ParseInt(string(s), 10, 64)
		if err != nil || n > maxSafeAttrInt || n < -maxSafeAttrInt {
			return errs.Newf(errs.Validation,
				"decoding attrs value %.24q: integer is outside the canonical range [-(2^53-1), 2^53-1] (A0-2.6)", string(s))
		}
		*v = AttrValue{Type: AttrInt, Num: n}
		return nil
	case s[0] == 'n':
		return errs.Newf(errs.Validation,
			"decoding attrs value: null is rejected — an unset attrs entry is absent, never null (A2-6.1, A0-8.3)")
	case s[0] == '[':
		return errs.Newf(errs.Validation,
			"decoding attrs value: arrays are rejected — attrs is flat, depth exactly one (A2-6.1)")
	case s[0] == '{':
		return errs.Newf(errs.Validation,
			"decoding attrs value: objects are rejected — attrs is flat, depth exactly one (A2-6.1)")
	default:
		return errs.Newf(errs.Validation,
			"decoding attrs value: want a JSON string, integer or boolean (A2-6.1)")
	}
}

// UnmarshalJSON decodes the attrs object: a flat map whose keys match A0-8.1
// and are not reserved (A2-6.2, A2-6.3) and whose values are bare scalars
// (A2-6.1). It runs in two passes over the same input bytes:
//
//  1. a json.Decoder token walk that observes the object structure and every
//     key — this is what makes duplicate-key rejection possible, since
//     decoding into a map silently keeps the last duplicate (A0-2.5's
//     reasoning) — and rejects nested containers and null values before any
//     of their content is decoded;
//  2. a map decode in which every value passes through
//     AttrValue.UnmarshalJSON with its raw bytes, so string escapes and
//     number literals are validated on the wire form, not on a
//     decoder-normalized form.
//
// Pass 1 has already rejected everything that could make pass 2 lossy, so
// the two passes cannot disagree. Every rejection is errs.Validation naming
// the offending key (bounded; keys are field names, which A0-3.4 permits
// echoing) and the rule — never the content of a rejected value (A2-9.5's
// discipline applied early; the value may be secret material).
func (a *Attrs) UnmarshalJSON(b []byte) error {
	if err := walkAttrsStructure(b); err != nil {
		return err
	}
	var m map[string]AttrValue
	if err := json.Unmarshal(b, &m); err != nil {
		// Pass 1 rejected every structural defect, so what reaches here is an
		// AttrValue value-form rejection (float literal, out-of-range integer,
		// lone surrogate). Wrap it — never flatten it into a new error
		// (ADR-0019 §2: the chain must survive errors.Unwrap) — and describe
		// it accurately: the input IS a flat object; a *value's form* is what
		// violates A2-6.1.
		return errs.Wrapf(err,
			"decoding graph attrs: %d input bytes carry a value that is not a bare scalar in A2-6.1 form", len(b))
	}
	if m == nil {
		// Unreachable after walkAttrsStructure rejected a null top level;
		// guarded so *a can never be set to a nil map from decoded input.
		return errs.Newf(errs.Validation, "decoding graph attrs: input is null (A0-8.3)")
	}
	*a = Attrs(m)
	return nil
}

// walkAttrsStructure is pass 1 of Attrs.UnmarshalJSON: the token walk that
// observes every key (A0-2.5's rule) and rejects structure violations before
// any value content is decoded. AGENTS.md lists untrusted-input parsing as
// high-review-bar code; the walk aborts at the first violation, so its work
// is bounded by the input size, which the API edge bounds (A0-8.9).
func walkAttrsStructure(b []byte) error {
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber() // A0-2.5: numbers keep their literal token text
	tok, err := dec.Token()
	if err != nil {
		return errs.Newf(errs.Validation,
			"decoding graph attrs: input is not valid JSON: %v (A2-6.1)", err)
	}
	if d, ok := tok.(json.Delim); !ok || d != '{' {
		return errs.Newf(errs.Validation,
			"decoding graph attrs: top-level value is not an object — attrs is a flat map, depth exactly one, no wrapper (A2-6.1)")
	}
	seen := make(map[string]bool)
	for dec.More() {
		if err := walkAttrEntry(dec, seen); err != nil {
			return err
		}
	}
	if tok, err := dec.Token(); err != nil || tok != json.Delim('}') {
		return errs.Newf(errs.Validation, "decoding graph attrs: object is not properly closed (A2-6.1)")
	}
	if dec.More() {
		return errs.Newf(errs.Validation, "decoding graph attrs: trailing data after the attrs object (A0-2.3)")
	}
	return nil
}

// walkAttrEntry consumes one key/value pair of the token walk, rejecting
// key-shape (A2-6.2), reserved-name (A2-6.3), duplicate-key (A0-2.5) and
// nested-container/null-value (A2-6.1) violations before any value content is
// decoded. A string/number/bool value's *form* is validated by
// AttrValue.UnmarshalJSON in pass 2, on the raw bytes.
func walkAttrEntry(dec *json.Decoder, seen map[string]bool) error {
	keyTok, err := dec.Token()
	if err != nil {
		return errs.Newf(errs.Validation, "decoding graph attrs: reading a key: %v (A2-6.1)", err)
	}
	key, ok := keyTok.(string)
	if !ok {
		// json.Decoder guarantees a string (or the closing delimiter,
		// which More() excluded) at key position.
		return errs.Newf(errs.Internal, "decoding graph attrs: decoder returned a non-string key token %T", keyTok)
	}
	if !attrKeyRE.MatchString(key) {
		return errs.Newf(errs.Validation,
			"decoding graph attrs: key %.64q does not match the A0-8.1 form ^[a-z][a-z0-9_]{0,39}$ (A2-6.2)", key)
	}
	if reservedAttrKeys[key] {
		return errs.Newf(errs.Validation,
			"decoding graph attrs: key %q is a reserved field name (A2-6.3, byte-exact membership)", key)
	}
	if seen[key] {
		return errs.Newf(errs.Validation,
			"decoding graph attrs: duplicate key %q — a map decode would silently keep the last (A0-2.5)", key)
	}
	seen[key] = true
	valTok, err := dec.Token()
	if err != nil {
		return errs.Newf(errs.Validation,
			"decoding graph attrs: reading the value of key %q: %v (A2-6.1)", key, err)
	}
	switch t := valTok.(type) {
	case json.Delim:
		return errs.Newf(errs.Validation,
			"decoding graph attrs: key %q: nested %q — objects and arrays are rejected, attrs is flat (A2-6.1)", key, string(t))
	case nil:
		return errs.Newf(errs.Validation,
			"decoding graph attrs: key %q: null is rejected — an unset attrs entry is absent (A2-6.1, A0-8.3)", key)
	}
	// string/number/bool: the content is validated by
	// AttrValue.UnmarshalJSON in pass 2, on the raw bytes.
	return nil
}

// rejectLoneSurrogates scans a JSON string literal (including its quotes) for
// \u escapes carrying a surrogate code point that is not half of a
// well-formed complementary pair, and rejects with validation (A0-2.3 (b)).
// It mirrors the scan internal/cjson runs over canonical documents; the
// attrs decoder needs its own because encoding/json replaces lone surrogates
// with U+FFFD before any caller sees them, and a replaced value would enter
// content_hash as bytes the writer never sent.
//
// The walk consumes every backslash escape in pairs, so an escaped backslash
// followed by the literal text "u..." ("\\ud800" on the wire) is not
// mistaken for an escape. Malformed hex or a truncated escape is left to
// json.Unmarshal, which rejects it; this scan decides only the surrogate
// rule.
func rejectLoneSurrogates(s []byte) error {
	for i := 0; i < len(s); {
		if s[i] != '\\' {
			i++
			continue
		}
		if i+1 >= len(s) {
			return nil // trailing backslash: json.Unmarshal rejects
		}
		if s[i+1] != 'u' {
			i += 2 // a short escape (\", \\, \n, ...): consume both bytes
			continue
		}
		if i+6 > len(s) {
			return nil // truncated escape: json.Unmarshal rejects
		}
		cp, ok := parseHex4(s[i+2 : i+6])
		if !ok {
			return nil // malformed hex: json.Unmarshal rejects
		}
		switch {
		case cp >= 0xD800 && cp <= 0xDBFF: // high surrogate: needs \uDC00-\uDFFF right after
			if i+12 <= len(s) && s[i+6] == '\\' && s[i+7] == 'u' {
				if lo, ok := parseHex4(s[i+8 : i+12]); ok && lo >= 0xDC00 && lo <= 0xDFFF {
					i += 12 // a well-formed pair: json.Unmarshal decodes it
					continue
				}
			}
			return errs.Newf(errs.Validation,
				"decoding attrs string value (%d bytes): lone high surrogate \\u%04x is rejected — encoding/json would silently replace it with U+FFFD (A0-2.3)",
				len(s), cp)
		case cp >= 0xDC00 && cp <= 0xDFFF:
			return errs.Newf(errs.Validation,
				"decoding attrs string value (%d bytes): unpaired low surrogate \\u%04x is rejected (A0-2.3)",
				len(s), cp)
		}
		i += 6
	}
	return nil
}

// parseHex4 decodes exactly four ASCII hex digits (upper or lower case, as
// JSON allows) and reports whether they were well-formed.
func parseHex4(b []byte) (rune, bool) {
	var cp rune
	for _, c := range b {
		switch {
		case c >= '0' && c <= '9':
			cp = cp<<4 | rune(c-'0')
		case c >= 'a' && c <= 'f':
			cp = cp<<4 | rune(c-'a'+10)
		case c >= 'A' && c <= 'F':
			cp = cp<<4 | rune(c-'A'+10)
		default:
			return 0, false
		}
	}
	return cp, true
}

// sortedAttrKeys returns a's keys in ascending byte order (A0-2.4's order,
// which sort gives directly for Go strings). Callers use it wherever attrs
// must be walked deterministically — an error that names "the first invalid
// field" must not depend on Go's map iteration order (A2-10.2: one error,
// the first violated rule).
func sortedAttrKeys(a Attrs) []string {
	keys := make([]string, 0, len(a))
	for k := range a {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}
