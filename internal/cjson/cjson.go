package cjson

import (
	"bytes"
	"encoding/json"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/daten-krake/sleipnir/internal/errs"
)

const (
	// MaxDepth is the maximum nesting depth of a canonical document (A0-2.11):
	// the top-level object is level 1, each nested object or array adds 1, and
	// scalars are not levels, so a document with 32 nested containers is
	// accepted and one with 33 is rejected. It is the same number internal/errs
	// uses to bound its error-chain walk, deliberately: one constant means the
	// same thing everywhere it bounds an untrusted structure.
	MaxDepth = 32

	// MaxBytes is the maximum input size accepted by Canonical: len(doc) of the
	// raw input, 1 MiB (A0-2.11).
	MaxBytes = 1 << 20
)

// maxSafeInt is the largest magnitude a canonical number may have (A0-2.6).
const maxSafeInt = int64(1<<53 - 1)

// numberLiteral is A0-2.5's token-text regex. Sixteen digits is the width of
// 2^53-1, so an over-long literal fails here; the range check in
// walker.number catches what the regex lets through (9007199254740993) and
// this regex catches what the range check would happily accept (1.0, 1e3, a
// leading zero, a plus sign).
var numberLiteral = regexp.MustCompile(`^-?(0|[1-9][0-9]{0,15})$`)

// member is one object entry: the key as decoded, and the canonical bytes of
// its value. Members are buffered so the keys can be sorted (A0-2.4) before
// the object is emitted.
type member struct {
	key   string
	value []byte
}

// walker drives one canonicalization of one document.
type walker struct {
	dec     *json.Decoder
	doc     []byte
	exclude map[string]bool
}

// Canonical re-serializes doc into the one canonical JSON form (A0-2): keys in
// UTF-8 byte order, no whitespace, integers only, literal UTF-8 strings, a
// top-level object required, duplicate and case-duplicate keys rejected, and
// the named top-level fields dropped before anything is emitted (A0-2.12).
// Every violation returns an error of errs.Kind Validation and no bytes:
// A0-3.1 classifies a canonicalization failure as validation, and Q3 says
// hard reject rather than normalize.
//
// The output is a pure function of the input, which is what makes A0-2.16
// (persist these exact bytes, recompute the digest from them) possible; see
// doc.go for what this package deliberately does not implement.
func Canonical(doc []byte, exclude ...string) ([]byte, error) {
	if err := preflight(doc); err != nil {
		return nil, err
	}
	dec := json.NewDecoder(bytes.NewReader(doc))
	dec.UseNumber() // A0-2.5: numbers keep their literal token text
	w := &walker{
		dec:     dec,
		doc:     doc,
		exclude: excludeSet(exclude),
	}
	if err := w.requireObject(); err != nil {
		return nil, err
	}
	out, err := w.object(nil, 1)
	if err != nil {
		return nil, err
	}
	if err := w.requireEnd(); err != nil {
		return nil, err
	}
	return out, nil
}

// CanonicalValue marshals v with encoding/json and canonicalizes the result.
// encoding/json only ever produces the intermediate bytes: Canonical re-parses
// and re-emits them, so its U+2028/U+2029 escaping and HTML escaping never
// reach the final bytes (A0-2.7).
//
// Every slice and map field of v MUST be non-nil (A0-2.14): json.Marshal emits
// null for a nil collection, and a null at any depth is rejected here with
// errs.Kind Validation. That rejection is a constructor defect surfacing as
// validation — which is why the clause requires constructors to initialize
// every collection field.
func CanonicalValue(v any, exclude ...string) ([]byte, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		// Reachable causes: a NaN/+Inf float (A0-2.6), a type encoding/json
		// cannot represent at all (a chan, a func), a cyclic structure, or a
		// MarshalJSON error from v's own type. A0-3.1 classifies every
		// canonicalization failure as validation, not internal, so the
		// annotation cites the general rule and not the integers clause.
		return nil, errs.Newf(errs.Validation, "marshaling %T for canonicalization: %.200s (A0-2, A0-3.1)", v, err)
	}
	return Canonical(raw, exclude...)
}

// With adds top-level fields to an already-canonical document and
// re-canonicalizes the result, so every addition meets the same rules as the
// document it joins. It never decodes into a typed struct: both sides are
// canonical bytes, merged and re-parsed.
//
// A key that doc already carries is errs.Kind Internal: A1-1.8 and A1-5.7 make
// this composition platform-only, so a collision is a platform defect, not
// client input. It is matched case-insensitively, because A0-2.5 already makes
// "hash" and "Hash" the same key. Everything else about add — a float, a null,
// a duplicate within add itself, an over-deep or over-size value — is rejected
// by the ordinary rules, with errs.Kind Validation.
func With(doc []byte, add map[string]any) ([]byte, error) {
	canonicalDoc, err := Canonical(doc)
	if err != nil {
		return nil, err
	}
	if len(add) == 0 {
		return canonicalDoc, nil
	}
	canonicalAdd, err := CanonicalValue(add)
	if err != nil {
		return nil, err
	}
	if err := rejectCollisions(canonicalDoc, canonicalAdd); err != nil {
		return nil, err
	}
	merged, err := Canonical(mergeObjects(canonicalDoc, canonicalAdd))
	if err != nil {
		return nil, err
	}
	return merged, nil
}

// preflight applies the checks that must happen before any work proportional
// to the document, and before the decoder, which sanitizes instead of
// rejecting (A0-2.3): empty input, the size bound, a BOM, invalid UTF-8, and a
// surrogate escape that is not part of a well-formed pair.
func preflight(doc []byte) error {
	switch {
	case len(doc) == 0:
		return errs.New(errs.Validation, "canonicalizing empty input: a document is a JSON object (A0-2.3)")
	case len(doc) > MaxBytes:
		return errs.Newf(errs.Validation, "canonicalizing %d bytes: exceeds MaxBytes %d (A0-2.11)", len(doc), MaxBytes)
	case bytes.HasPrefix(doc, []byte{0xEF, 0xBB, 0xBF}):
		return errs.New(errs.Validation, "canonicalizing a document with a UTF-8 BOM: output is UTF-8 without BOM (A0-2.3)")
	case !utf8.Valid(doc):
		return errs.New(errs.Validation, "canonicalizing non-UTF-8 input: encoding/json substitutes U+FFFD rather than failing, so the document must be rejected here (A0-2.3)")
	}
	return checkSurrogates(doc)
}

// requireObject consumes the opening delimiter of the top-level object: a
// top-level array, string, number or literal is rejected (A0-2.10). A null is
// named as such (A0-2.14).
func (w *walker) requireObject() error {
	tok, err := w.dec.Token()
	if err != nil {
		return w.decodeError(err)
	}
	if d, ok := tok.(json.Delim); ok && d == '{' {
		return nil
	}
	const rule = "the document must be a JSON object (A0-2.10)"
	switch t := tok.(type) {
	case nil:
		return errs.Newf(errs.Validation, "canonicalizing a top-level null at offset %d: %s, and null appears in no contract document (A0-2.14)", w.dec.InputOffset(), rule)
	case json.Delim:
		return errs.Newf(errs.Validation, "canonicalizing a top-level %c at offset %d: %s", t, w.dec.InputOffset(), rule)
	default:
		return errs.Newf(errs.Validation, "canonicalizing a top-level %T at offset %d: %s", t, w.dec.InputOffset(), rule)
	}
}

// requireEnd rejects anything after the top-level value (A0-2.3). A0-2.5
// requires the walk to call Decoder.More(); More() alone is not enough, because
// it reports false when the next byte is a stray closing delimiter, so the
// unread bytes are checked too.
func (w *walker) requireEnd() error {
	off := w.dec.InputOffset()
	if w.dec.More() || len(bytes.TrimSpace(w.doc[off:])) > 0 {
		return errs.Newf(errs.Validation, "canonicalizing a document with data after its top-level value at offset %d of %d bytes (A0-2.3)", off, len(w.doc))
	}
	return nil
}

// object emits the object whose '{' is consumed. depth is this object's level;
// the top-level object is level 1 (A0-2.11).
func (w *walker) object(dst []byte, depth int) ([]byte, error) {
	if err := w.checkDepth(depth); err != nil {
		return nil, err
	}
	var members []member
	seen := make(map[string]bool)
	for w.dec.More() {
		key, err := w.key(depth)
		if err != nil {
			return nil, err
		}
		// A0-2.5: two keys that strings.EqualFold calls equal are duplicates,
		// and driving Token() is what makes the second one observable at all.
		// One fold per key: the same folded name drives the duplicate check
		// and the exclusion match, so the package holds exactly one notion of
		// "the same key" (see foldKey and doc.go's rulings).
		folded := foldKey(key)
		if seen[folded] {
			return nil, errs.Newf(errs.Validation, "canonicalizing an object at depth %d: duplicate key %.64q at offset %d, compared case-insensitively (A0-2.5)", depth, key, w.dec.InputOffset())
		}
		seen[folded] = true

		value, err := w.value(nil, depth+1)
		if err != nil {
			return nil, err
		}
		if depth == 1 && w.exclude[folded] {
			continue
		}
		members = append(members, member{key: key, value: value})
	}
	if err := w.close('}', depth); err != nil {
		return nil, err
	}

	// A0-2.4: ascending by the unsigned byte value of the UTF-8 encoding, which
	// is Go's string order (and not UTF-16 code-unit order: see vector V6).
	sort.Slice(members, func(i, j int) bool { return members[i].key < members[j].key })

	dst = append(dst, '{')
	for i, m := range members {
		if i > 0 {
			dst = append(dst, ',')
		}
		dst = appendQuote(dst, m.key)
		dst = append(dst, ':')
		dst = append(dst, m.value...)
	}
	return append(dst, '}'), nil
}

// array emits the array whose '[' is consumed. Element order is preserved
// exactly; arrays are never sorted (A0-2.8).
func (w *walker) array(dst []byte, depth int) ([]byte, error) {
	if err := w.checkDepth(depth); err != nil {
		return nil, err
	}
	dst = append(dst, '[')
	for i := 0; w.dec.More(); i++ {
		if i > 0 {
			dst = append(dst, ',')
		}
		var err error
		if dst, err = w.value(dst, depth+1); err != nil {
			return nil, err
		}
	}
	if err := w.close(']', depth); err != nil {
		return nil, err
	}
	return append(dst, ']'), nil
}

// value emits the next value of the stream, which sits one level below the
// container that holds it.
func (w *walker) value(dst []byte, depth int) ([]byte, error) {
	tok, err := w.dec.Token()
	if err != nil {
		return nil, w.decodeError(err)
	}
	switch t := tok.(type) {
	case nil:
		return nil, errs.Newf(errs.Validation, "canonicalizing a value at depth %d, offset %d: null is never canonical (A0-2.14); a nil slice or map field marshals to null, so constructors MUST initialize every collection", depth, w.dec.InputOffset())
	case bool:
		if t {
			return append(dst, "true"...), nil
		}
		return append(dst, "false"...), nil
	case string:
		return appendQuote(dst, t), nil
	case json.Number:
		return w.number(dst, string(t), depth)
	case json.Delim:
		switch t {
		case '{':
			return w.object(dst, depth)
		case '[':
			return w.array(dst, depth)
		}
	}
	return nil, errs.Newf(errs.Validation, "canonicalizing a value at depth %d: unexpected token %.64v (A0-2.5)", depth, tok)
}

// foldKey returns the key-identity class of s used at every key comparison
// in this package (A0-2.5, A0-2.12): foldKey(a) == foldKey(b) exactly when
// strings.EqualFold(a, b) says the keys are equal, which is the equivalence
// encoding/json uses when it matches field names — the agreement A0-2.5's
// rationale is about. Each rune maps to the minimum of its unicode.SimpleFold
// orbit with ASCII letters normalized to lowercase first, so a lowercase
// ASCII key is already folded and an orbit containing one (S, s, U+017F LONG S)
// all land on the lowercase letter; pure orbit minima would map "a" to "A"
// and split the very pairs A0-2.5 names. strings.ToLower is weaker than this
// and accepts {"s":1,"\u017f":2}, which a downstream encoding/json decode
// reads as one field.
//
// Fast path: a key with no byte >= utf8.RuneSelf and no ASCII uppercase
// letter is returned unchanged, same string value, no Builder, no
// allocation. A0-8.1 declares every contract key ^[a-z][a-z0-9_]{0,39}$, so
// that is the normal path for every document the platform produces.
//
// Simple folding cannot equate the full-folding expansions (\u00df vs "ss",
// the \ufb01 ligature vs "fi"); those need CaseFolding.txt tables the stdlib
// does not ship, and strings.EqualFold does not equate them either. That is
// the achievable ceiling, documented in doc.go.
func foldKey(s string) string {
	for i := 0; i < len(s); i++ {
		if c := s[i]; c >= utf8.RuneSelf || ('A' <= c && c <= 'Z') {
			return foldKeySlow(s)
		}
	}
	return s
}

// foldKeySlow folds a key that contains a non-ASCII byte or an ASCII
// uppercase letter. Cost is linear in the key: folding each key once into a
// map is what keeps the duplicate check linear overall (doc.go rules out the
// quadratic pairwise EqualFold scan).
func foldKeySlow(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		b.WriteRune(foldRune(r))
	}
	return b.String()
}

// foldRune returns the canonical representative of r's SimpleFold orbit: the
// minimum orbit member after ASCII uppercase is normalized to lowercase.
// Two runes get the same representative exactly when they are in the same
// orbit, i.e. exactly when strings.EqualFold calls them equal (pinned by a
// test over random key pairs).
func foldRune(r rune) rune {
	m := asciiLower(r)
	for f := unicode.SimpleFold(r); f != r; f = unicode.SimpleFold(f) {
		if c := asciiLower(f); c < m {
			m = c
		}
	}
	return m
}

// asciiLower maps an ASCII uppercase letter to its lowercase partner, so an
// orbit containing an ASCII letter pair represents by the lowercase letter;
// every other rune is itself.
func asciiLower(r rune) rune {
	if 'A' <= r && r <= 'Z' {
		return r + ('a' - 'A')
	}
	return r
}

// key returns the next object key.
func (w *walker) key(depth int) (string, error) {
	tok, err := w.dec.Token()
	if err != nil {
		return "", w.decodeError(err)
	}
	key, ok := tok.(string)
	if !ok {
		return "", errs.Newf(errs.Validation, "canonicalizing an object key at depth %d: got %.64v, want a string (A0-2.5)", depth, tok)
	}
	return key, nil
}

// number appends a number's literal token text verbatim after validating it
// (A0-2.5). The regex and the range check are separate steps because each
// catches what the other does not.
func (w *walker) number(dst []byte, text string, depth int) ([]byte, error) {
	const rangeRule = "only integers in [-(2^53-1), 2^53-1], written without a plus sign, fraction, exponent, leading zero or negative zero (A0-2.6)"
	if text == "-0" || !numberLiteral.MatchString(text) {
		return nil, errs.Newf(errs.Validation, "canonicalizing the number %.64q at depth %d, offset %d: %s", text, depth, w.dec.InputOffset(), rangeRule)
	}
	n, err := strconv.ParseInt(text, 10, 64)
	if err != nil || n < -maxSafeInt || n > maxSafeInt {
		return nil, errs.Newf(errs.Validation, "canonicalizing the number %.64q at depth %d, offset %d: %s", text, depth, w.dec.InputOffset(), rangeRule)
	}
	return append(dst, text...), nil
}

// checkDepth enforces the nesting bound at a container boundary (A0-2.11).
func (w *walker) checkDepth(depth int) error {
	if depth > MaxDepth {
		return errs.Newf(errs.Validation, "canonicalizing a container at depth %d, offset %d: exceeds MaxDepth %d (A0-2.11)", depth, w.dec.InputOffset(), MaxDepth)
	}
	return nil
}

// close consumes the delimiter that ends a container, naming the rule rather
// than repeating the decoder.
func (w *walker) close(want json.Delim, depth int) error {
	tok, err := w.dec.Token()
	if err != nil {
		return w.decodeError(err)
	}
	if d, ok := tok.(json.Delim); ok && d == want {
		return nil
	}
	return errs.Newf(errs.Validation, "closing the container at depth %d: expected %q, got %.64v (A0-2.5)", depth, string(want), tok)
}

// decodeError maps every json.Decoder failure to validation: a leading zero,
// NaN, a raw control byte in a string and a malformed document are all
// malformed input, none is a platform defect, and none may leak as internal
// (A0-2.5, A0-3.1).
func (w *walker) decodeError(err error) error {
	return errs.Newf(errs.Validation, "reading the JSON token stream at offset %d: %.200s (A0-2.5)", w.dec.InputOffset(), err)
}

// excludeSet returns the exclusion list as a lookup set keyed by foldKey, so
// a declared name matches document keys under the same A0-2.5 equivalence the
// duplicate check uses (A0-2.12 ruling in doc.go). It is a fixed set of
// top-level names: no path syntax, no wildcards, no nested exclusion.
func excludeSet(names []string) map[string]bool {
	if len(names) == 0 {
		return nil
	}
	set := make(map[string]bool, len(names))
	for _, name := range names {
		set[foldKey(name)] = true
	}
	return set
}

// checkSurrogates scans the RAW input bytes for \uXXXX escapes and rejects a
// surrogate code point that does not form a well-formed pair (A0-2.3 (b)). It
// must read the raw bytes and not decoded strings: the decoder accepts
// "\ud800" and hands back U+FFFD, which would then canonicalize silently and
// change the digest with no error.
func checkSurrogates(doc []byte) error {
	inString := false
	for i := 0; i < len(doc); i++ {
		switch {
		case !inString:
			inString = doc[i] == '"'
		case doc[i] == '"':
			inString = false
		case doc[i] != '\\':
			// An ordinary character, including every byte of a multi-byte
			// rune: no JSON string escape starts at a byte >= 0x80.
		case i+1 >= len(doc) || doc[i+1] != 'u':
			i++ // a two-byte escape such as \" or \\
		default:
			cp, ok := hexEscape(doc, i+2)
			if !ok {
				i += 5 // not a code point: let the decoder report the syntax error
				continue
			}
			span, err := escapeSpan(doc, i, cp)
			if err != nil {
				return err
			}
			i += span
		}
	}
	return nil
}

// escapeSpan returns how many bytes past the backslash the escape at index i
// occupies: 5 for one \uXXXX escape, 11 for a well-formed surrogate pair.
func escapeSpan(doc []byte, i int, cp rune) (int, error) {
	if !isHighSurrogate(cp) && !isLowSurrogate(cp) {
		return 5, nil
	}
	// A pair is exactly two adjacent escapes: the second one starts six bytes
	// after this one's backslash. The \u prefix must be checked as well —
	// without it, the text "\ud800xxdc00" would read as a pair and the lone
	// surrogate would slip through as the decoder's U+FFFD.
	if isHighSurrogate(cp) && lowSurrogateEscape(doc, i+6) {
		return 11, nil
	}
	return 0, errs.Newf(errs.Validation, "canonicalizing the lone surrogate escape %.6q at offset %d: a surrogate must be paired (A0-2.3)", doc[i:min(i+6, len(doc))], i)
}

func isHighSurrogate(r rune) bool { return r >= 0xD800 && r <= 0xDBFF }

func isLowSurrogate(r rune) bool { return r >= 0xDC00 && r <= 0xDFFF }

// lowSurrogateEscape reports whether a \uXXXX escape starting at index i holds
// a low surrogate — the second half of a well-formed pair (A0-2.3).
func lowSurrogateEscape(doc []byte, i int) bool {
	if i+6 > len(doc) || doc[i] != '\\' || doc[i+1] != 'u' {
		return false
	}
	cp, ok := hexEscape(doc, i+2)
	return ok && isLowSurrogate(cp)
}

// hexEscape decodes the four hex digits of a \u escape starting at index i.
func hexEscape(doc []byte, i int) (rune, bool) {
	if i+4 > len(doc) {
		return 0, false
	}
	var r rune
	for _, c := range doc[i : i+4] {
		switch {
		case '0' <= c && c <= '9':
			r = r*16 + rune(c-'0')
		case 'a' <= c && c <= 'f':
			r = r*16 + rune(c-'a'+10)
		case 'A' <= c && c <= 'F':
			r = r*16 + rune(c-'A'+10)
		default:
			return 0, false
		}
	}
	return r, true
}

// appendQuote emits s as a canonical JSON string (A0-2.7): only the double
// quote, the backslash and U+0000-U+001F are escaped; those five get their
// short escapes and the other control characters \u00xx with lowercase hex.
// Every other code point is literal — non-ASCII, U+007F, U+2028 and U+2029
// included — and no HTML escaping is applied, which is exactly why
// encoding/json is not the emitter. The escapes in the input are decoded by
// the token walk before s arrives here, so no input escape is passed through.
//
// Copying every byte >= 0x80 raw keeps A0-2.3's "output MUST be UTF-8" true
// on one invariant: s is already valid UTF-8. Two guarantors, one per input
// path — preflight's utf8.Valid on the raw bytes for Canonical, and
// encoding/json's U+FFFD substitution for Go strings on the CanonicalValue
// path (the decoder never emits an unpaired surrogate). Re-validating per
// string here would be a second pass for nothing.
func appendQuote(dst []byte, s string) []byte {
	dst = append(dst, '"')
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case c == '"':
			dst = append(dst, `\"`...)
		case c == '\\':
			dst = append(dst, `\\`...)
		case c == '\b':
			dst = append(dst, `\b`...)
		case c == '\f':
			dst = append(dst, `\f`...)
		case c == '\n':
			dst = append(dst, `\n`...)
		case c == '\r':
			dst = append(dst, `\r`...)
		case c == '\t':
			dst = append(dst, `\t`...)
		case c < 0x20:
			const hexDigits = "0123456789abcdef"
			dst = append(dst, `\u00`...)
			dst = append(dst, hexDigits[c>>4], hexDigits[c&0xf])
		default:
			dst = append(dst, c)
		}
	}
	return append(dst, '"')
}

// rejectCollisions reports the platform defect of With adding a field the
// document already carries. Both sides are canonical bytes this package just
// produced, so their keys cannot fail to decode in any other way.
func rejectCollisions(doc, add []byte) error {
	held, err := objectKeys(doc)
	if err != nil {
		return err
	}
	existing := make(map[string]bool, len(held))
	for _, k := range held {
		existing[foldKey(k)] = true
	}
	added, err := objectKeys(add)
	if err != nil {
		return err
	}
	for _, k := range added {
		// Under the A0-2.5 key equivalence: A0-2.5 makes "hash" and "Hash"
		// the same key, so a collision under either spelling is the same
		// platform defect.
		if existing[foldKey(k)] {
			return errs.Newf(errs.Internal, "adding the field %.64q to a document that already carries it: With is platform-only composition (A1-1.8, A1-5.7)", k)
		}
	}
	return nil
}

// objectKeys returns the sorted top-level keys of canonical object bytes.
// Decoding into a map is safe here, and only here, because the input is this
// package's own output: canonical bytes carry no duplicate keys and never have
// a top level other than an object.
func objectKeys(doc []byte) ([]string, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(doc, &fields); err != nil {
		return nil, errs.Wrapf(err, "reading the keys of a canonical object of %d bytes", len(doc))
	}
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys, nil
}

// mergeObjects joins two canonical objects. The result is not canonical — it
// is handed back to Canonical, which sorts and re-validates it.
func mergeObjects(doc, add []byte) []byte {
	if len(doc) == 2 { // doc is "{}"
		return add
	}
	out := make([]byte, 0, len(doc)+len(add))
	out = append(out, doc[:len(doc)-1]...)
	out = append(out, ',')
	out = append(out, add[1:]...)
	return out
}
