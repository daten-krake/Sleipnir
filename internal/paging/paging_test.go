package paging

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/daten-krake/sleipnir/internal/cjson"
	"github.com/daten-krake/sleipnir/internal/errs"
	"github.com/daten-krake/sleipnir/internal/ids"
)

// The fixtures below are the A0 §4 paging example, transcribed rather than
// computed: publishedCursor is the literal printed in the contract's JSON
// example and publishedPayload is what it decodes to, both checked by hand
// against the base64url alphabet and recomputed independently in Python by the
// principal. A0-4.4 cited this encoding rule as A0-8.5, which is the
// closed-enum clause; the rule is A0-8.6, ruled an erratum 2026-09-29 (A0 §6
// item 20).
//
// publishedCursor is the CORRECTED literal of A0 §6 item 19 (erratum, product
// owner 2026-09-29). §4's example originally published preErratumCursor, whose
// id body is 25 characters while items[0].event_id in the same example carries
// the 26-character body A0-1.1 requires — so the frozen contract's own example
// cursor was not a valid A0-1.2 id. The example is fixed; the defective literal
// stays pinned below as the rejection fixture, because a well-formed cursor no
// collection can resolve is exactly A0-4.8's case.
const (
	publishedCursor  = "eyJpZCI6ImV2dF8wMW0xeTJ3aGZocDE3ZzBhdmRxenRkMnAzeCIsImsiOjQ3MTF9"
	publishedPayload = `{"id":"evt_01m1y2whfhp17g0avdqztd2p3x","k":4711}`
	publishedID      = "evt_01m1y2whfhp17g0avdqztd2p3x"
	publishedK       = int64(4711)

	// preErratumCursor is the literal A0 §4 published before item 19, and
	// preErratumPayload what it decodes to: canonical JSON, valid unpadded
	// base64url, 25-character id body.
	preErratumCursor  = "eyJpZCI6ImV2dF8wMW0xeTJ3aGZocDE3ZzBhdmRxenRkMnAzIiwiayI6NDcxMX0"
	preErratumPayload = `{"id":"evt_01m1y2whfhp17g0avdqztd2p3","k":4711}`

	// validID is the event id of §4's example items[], and is a byte-exact
	// A0-1.2 evt_ id: prefix plus the 26-character body of A0-1.1. Since item 19
	// it is also the id publishedCursor encodes.
	validID = "evt_01m1y2whfhp17g0avdqztd2p3x"
	// otherKindID is the same body under the gn_ prefix: valid for
	// ids.GraphNode, invalid for ids.Event.
	otherKindID = "gn_01m1y2whfhp17g0avdqztd2p3x"
)

// reencodeStage is the prose of DecodeCursor's last step, the one that rejects
// a payload whose re-encoding does not reproduce the input bytes. Tests assert
// on it rather than on "canonical", which the earlier cjson wrapper also
// contains, so a row cannot silently be attributed to the wrong step.
const reencodeStage = "did not reproduce its input bytes"

// encodeTestCursor builds a cursor string from raw payload text with the
// standard library only. Reject-table fixtures MUST NOT be built with
// EncodeCursor: a cursor this package produced could then never be the input
// that proves this package rejects it.
func encodeTestCursor(t *testing.T, payload string) string {
	t.Helper()
	return base64.RawURLEncoding.EncodeToString([]byte(payload))
}

// task is the row type the NewPage tests page over. Its json tags exist so
// cjson.CanonicalValue can render a whole Page for the byte-exact envelope
// assertions.
type task struct {
	ID string `json:"id"`
	K  int64  `json:"k"`
}

// cursorOfTask satisfies A0-4.3's requirement that the cursor carry the row's
// immutable id and its integer ordering key.
func cursorOfTask(r task) Cursor { return Cursor{K: r.K, ID: r.ID} }

func TestCursorRoundTripIsCanonicalBase64URL(t *testing.T) {
	t.Run("published_literal_is_reproduced_byte_for_byte", func(t *testing.T) {
		// The encoder must agree with the contract's own bytes. Editing
		// EncodeCursor to sort the keys by hand, to emit {"k":…,"id":…}, or to
		// use StdEncoding makes this fail while breaking A0-2.4 and A0-8.6.
		got, err := EncodeCursor(Cursor{K: publishedK, ID: publishedID})
		if err != nil {
			t.Fatalf("EncodeCursor: %v", err)
		}
		if got != publishedCursor {
			t.Errorf("EncodeCursor = %q, want the A0 §4 literal %q", got, publishedCursor)
		}
	})

	t.Run("payload_is_the_canonical_two_key_object", func(t *testing.T) {
		// Independent of the encoder: decode the published literal with the
		// stdlib and check the bytes A0-4.4 describes — canonical order, "id"
		// first, no whitespace.
		raw, err := base64.RawURLEncoding.DecodeString(publishedCursor)
		if err != nil {
			t.Fatalf("decoding the published literal: %v", err)
		}
		if string(raw) != publishedPayload {
			t.Errorf("published literal decodes to %s, want %s", raw, publishedPayload)
		}
	})

	t.Run("decode_returns_the_encoded_cursor", func(t *testing.T) {
		s, err := EncodeCursor(Cursor{K: publishedK, ID: validID})
		if err != nil {
			t.Fatalf("EncodeCursor: %v", err)
		}
		want, err := DecodeCursor(s, ids.Event)
		if err != nil {
			t.Fatalf("DecodeCursor(%q, evt_): %v", s, err)
		}
		if want != (Cursor{K: publishedK, ID: validID}) {
			t.Errorf("round trip = %+v, want %+v", want, Cursor{K: publishedK, ID: validID})
		}
	})

	t.Run("pre_erratum_literal_is_rejected_as_an_invalid_id", func(t *testing.T) {
		// A0 §6 item 19. The defective literal is still valid unpadded base64url
		// of canonical {"id","k"} JSON, so it reaches DecodeCursor's last
		// substantive check and fails there on A0-1.5, not earlier. Attributing
		// it to the id shape is the assertion that matters: an error blaming
		// base64 or canonicality would mean a step runs out of order.
		// Non-vacuity: dropping DecodeCursor's ids.Valid step makes this subtest
		// fail with err == nil.
		raw, err := base64.RawURLEncoding.DecodeString(preErratumCursor)
		if err != nil {
			t.Fatalf("decoding the pre-erratum literal: %v", err)
		}
		if string(raw) != preErratumPayload {
			t.Errorf("pre-erratum literal decodes to %s, want %s", raw, preErratumPayload)
		}
		if _, err := DecodeCursor(preErratumCursor, ids.Event); err == nil {
			t.Error("DecodeCursor accepted a 25-character id body; A0-1.1 requires 26")
		} else if got := errs.KindOf(err); got != errs.Validation {
			t.Errorf("kind = %q, want %q", got, errs.Validation)
		} else if !strings.Contains(err.Error(), "A0-1.5") {
			t.Errorf("error %q does not reject the literal as an invalid A0-1 id", err)
		}
	})

	t.Run("alphabet_and_stability", func(t *testing.T) {
		// A0-8.6: the emitted value is base64url WITHOUT padding, so it
		// can carry no '='; and for the payload of a real cursor neither can the
		// two alphabet-specific characters: a six-bit group of 62 or 63 needs a
		// payload byte of '>' , '?', '~' or DEL, none of which a canonical id or
		// an integer can be. So a valid cursor's base64url is [A-Za-z0-9]+ —
		// '-' and '_' appearing would mean the payload stopped being canonical
		// JSON of {"id","k"}. Dropping RawURLEncoding for StdEncoding fails the
		// '=' half; emitting hand-built JSON with a "_" key fails the '-'/'_' half.
		seen := map[string]bool{}
		for _, k := range []int64{0, 1, 4711, -4711, 9007199254740991, -9007199254740991} {
			s, err := EncodeCursor(Cursor{K: k, ID: validID})
			if err != nil {
				t.Fatalf("EncodeCursor(k=%d): %v", k, err)
			}
			if strings.ContainsAny(s, "+/=_-") {
				t.Errorf("EncodeCursor(k=%d) = %q, contains a padding, standard-alphabet or url-alphabet-exclusive character", k, s)
			}
			seen[s] = true
		}
		if len(seen) != 6 {
			t.Errorf("6 distinct ordering values produced %d distinct cursors", len(seen))
		}

		// Deterministic: a cursor is a digest-adjacent opaque value, so two runs
		// must not disagree (A0-2.1). Map iteration order or a hand-built JSON
		// string is what this pins.
		first, err := EncodeCursor(Cursor{K: publishedK, ID: validID})
		if err != nil {
			t.Fatalf("EncodeCursor: %v", err)
		}
		for range 100 {
			again, err := EncodeCursor(Cursor{K: publishedK, ID: validID})
			if err != nil {
				t.Fatalf("EncodeCursor: %v", err)
			}
			if again != first {
				t.Fatalf("EncodeCursor is not stable: %q then %q", first, again)
			}
		}
	})

	t.Run("ordering_value_outside_the_canonical_range_is_rejected", func(t *testing.T) {
		// A0-2.6 bounds canonical integers to [-(2^53-1), 2^53-1]; cjson decides
		// it and paging inherits it. An implementation that hand-built JSON with
		// fmt.Sprintf would silently emit an uncanonical cursor here instead of
		// failing.
		tests := []struct {
			name string
			k    int64
		}{
			{"above the upper bound", 9007199254740992},
			{"below the lower bound", -9007199254740992},
			{"int64 maximum", int64(1 << 62)},
		}
		for _, tc := range tests {
			got, err := EncodeCursor(Cursor{K: tc.k, ID: validID})
			if err == nil {
				t.Errorf("%s: EncodeCursor = %q, want an error (A0-2.6)", tc.name, got)
				continue
			}
			if kind := errs.KindOf(err); kind != errs.Validation {
				t.Errorf("%s: kind = %q, want %q", tc.name, kind, errs.Validation)
			}
			if got != "" {
				t.Errorf("%s: EncodeCursor returned %q alongside an error", tc.name, got)
			}
		}
	})

	t.Run("boundary_values_round_trip", func(t *testing.T) {
		tests := []struct {
			name string
			c    Cursor
			kind ids.Kind
		}{
			{"zero ordering value", Cursor{K: 0, ID: validID}, ids.Event},
			{"negative ordering value", Cursor{K: -1, ID: validID}, ids.Event},
			{"most negative canonical value", Cursor{K: -9007199254740991, ID: validID}, ids.Event},
			{"largest canonical value", Cursor{K: 9007199254740991, ID: validID}, ids.Event},
			{"longest prefix", Cursor{K: 7, ID: "slp_node_01m1y2whfhp17g0avdqztd2p3x"}, ids.AgentNode},
			{"task kind", Cursor{K: 7, ID: "task_01m1y2whfhp17g0avdqztd2p3x"}, ids.Task},
			{"empty id rejected rather than round-tripped", Cursor{K: 7, ID: ""}, ids.Event},
		}
		for _, tc := range tests {
			s, err := EncodeCursor(tc.c)
			if err != nil {
				t.Fatalf("%s: EncodeCursor: %v", tc.name, err)
			}
			if tc.c.ID == "" {
				// EncodeCursor does not validate ids (it has no Kind to
				// validate against); the decode half must refuse it.
				if _, err := DecodeCursor(s, tc.kind); err == nil {
					t.Errorf("%s: DecodeCursor accepted an empty id", tc.name)
				}
				continue
			}
			got, err := DecodeCursor(s, tc.kind)
			if err != nil {
				t.Fatalf("%s: DecodeCursor(%q): %v", tc.name, s, err)
			}
			if got != tc.c {
				t.Errorf("%s: round trip = %+v, want %+v", tc.name, got, tc.c)
			}
		}
	})
}

func TestCursorRejectsStandardAlphabetAndPadding(t *testing.T) {
	// A0-8.6 puts the whole rejection in the choice of decoder:
	// base64.RawURLEncoding refuses '+' and '/' (the standard alphabet) and
	// refuses '=' padding. No character scan is added to DecodeCursor, so these
	// subtests are the proof that the stdlib covers the clause. If
	// RawURLEncoding were swapped for StdEncoding the padding row below fails;
	// if it were swapped for a permissive decoder the '+' and '/' rows fail.
	tildePayload := `{"id":"evt_~01m1y2whfhp17g0avdqztd2p3x","k":4711}`
	questPayload := `{"id":"evt_?01m1y2whfhp17g0avdqztd2p3x","k":4711}`

	tests := []struct {
		name  string
		in    string
		stage string // the failure this row must be attributed to
	}{
		{
			name:  "equals padding appended to the published literal",
			in:    publishedCursor + "=",
			stage: "base64url",
		},
		{
			name:  "standard alphabet plus",
			in:    base64.StdEncoding.EncodeToString([]byte(tildePayload)),
			stage: "base64url",
		},
		{
			name:  "standard alphabet slash",
			in:    base64.StdEncoding.EncodeToString([]byte(questPayload)),
			stage: "base64url",
		},
		{
			name:  "not base64 at all",
			in:    "!!!not base64!!!",
			stage: "base64url",
		},
		{
			name:  "one leftover character cannot form a quantum",
			in:    "A",
			stage: "base64url",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := DecodeCursor(tc.in, ids.Event)
			if err == nil {
				t.Fatalf("DecodeCursor(%q) accepted the value, want rejection (A0-8.6)", tc.in)
			}
			if kind := errs.KindOf(err); kind != errs.Validation {
				t.Errorf("kind = %q, want %q (A0-4.8)", kind, errs.Validation)
			}
			// Stage attribution: A0-3.4 forbids a CLIENT from parsing messages;
			// a test naming the clause the failure cites is how "the alphabet is
			// what was rejected" is proved rather than assumed.
			if !strings.Contains(err.Error(), tc.stage) {
				t.Errorf("error %q does not name the %q stage", err, tc.stage)
			}
		})
	}

	t.Run("non_canonical_trailing_bits_of_the_final_character", func(t *testing.T) {
		// The canonical-ENCODING half of A0-8.6, which the re-encode comparison
		// at paging.go's last step is the only check that can catch: the unused
		// low bits of the final base64 quantum may differ while every decoded
		// byte is identical. This fixture is 46 canonical bytes (46 mod 3 == 1),
		// so its final character carries 4 unused bits and 16 spellings decode to
		// the same payload. Go's RawURLEncoding is lenient about them, so such a
		// cursor passes the alphabet check, cjson and ids.Valid alike. Deleting
		// the re-encode comparison accepts all sixteen. If a future stdlib starts
		// refusing the non-canonical spellings at decode time, the Fatalf below
		// says so instead of letting the loop pin nothing.
		payload := `{"id":"` + validID + `","k":47}`
		canonical := encodeTestCursor(t, payload)
		class := base64Class(t, canonical, payload)
		if len(class) != 15 {
			t.Fatalf("final-character equivalence class has %d non-canonical members, want 15 of 16 (stdlib leniency changed?): %q", len(class), class)
		}
		for _, mutated := range class {
			_, err := DecodeCursor(mutated, ids.Event)
			if err == nil {
				t.Errorf("DecodeCursor accepted %q, which is not the canonical base64url of the payload (A0-8.6, A0-4.4)", mutated)
				continue
			}
			if got := errs.KindOf(err); got != errs.Validation {
				t.Errorf("kind = %q, want %q", got, errs.Validation)
			}
			if !strings.Contains(err.Error(), reencodeStage) {
				t.Errorf("error %q is not attributed to the re-encode step", err)
			}
		}
	})

	t.Run("padding_on_an_otherwise_fully_valid_cursor", func(t *testing.T) {
		// The cleanest form of A0-8.6's padding rule: one payload, valid in
		// every respect (kind-correct id, canonical bytes), offered in both
		// encodings. 47 canonical bytes need one '=' of padding in the
		// standard form, so the two strings differ by that character alone:
		// the unpadded form must decode and the padded form must not. If
		// RawURLEncoding were swapped for StdEncoding, the padded row is
		// accepted; swapping in a padding-stripping decoder breaks the
		// unpadded row's twin assertion below instead.
		payload := `{"id":"` + validID + `","k":471}`
		if len(payload)%3 == 0 {
			t.Fatalf("fixture is %d bytes, so no standard-alphabet padding exists to reject", len(payload))
		}
		unpadded := encodeTestCursor(t, payload)
		padded := base64.StdEncoding.EncodeToString([]byte(payload))
		if !strings.HasSuffix(padded, "=") {
			t.Fatalf("standard encoding %q carries no padding", padded)
		}
		if padded != unpadded+"=" {
			t.Fatalf("fixtures differ by more than padding: %q vs %q", padded, unpadded)
		}
		if got, err := DecodeCursor(unpadded, ids.Event); err != nil {
			t.Errorf("DecodeCursor(%q): %v, want it accepted", unpadded, err)
		} else if got.ID != validID || got.K != 471 {
			t.Errorf("DecodeCursor(%q) = %+v", unpadded, got)
		}
		if _, err := DecodeCursor(padded, ids.Event); err == nil {
			t.Error("DecodeCursor accepted a padded cursor (A0-8.6)")
		} else if !strings.Contains(err.Error(), "base64url") {
			t.Errorf("error %q does not blame the base64 stage", err)
		}
	})

	t.Run("an_unpadded_literal_is_not_rejected_for_padding", func(t *testing.T) {
		// The pre-erratum literal of A0 §6 item 19: valid unpadded base64url of
		// canonical JSON whose ONLY defect is a 25-character id body. It must
		// fail on the id, never on the base64 stage — otherwise a padding check
		// would be firing on a cursor A0-8.6 accepts.
		// Non-vacuity: making DecodeCursor reject unpadded input fails the
		// "blames base64" assertion below.
		got, err := DecodeCursor(preErratumCursor, ids.Event)
		if err == nil {
			t.Fatalf("unexpected accept of %+v: the literal carries a 25-character id body, so it must fail on its id", got)
		}
		if strings.Contains(err.Error(), "base64url") {
			t.Errorf("error %q blames base64 for a cursor whose only defect is its id body", err)
		}
		// The positive control with the example's real event id.
		s := encodeTestCursor(t, `{"id":"`+validID+`","k":4711}`)
		if _, err := DecodeCursor(s, ids.Event); err != nil {
			t.Errorf("DecodeCursor rejected its unpadded form: %v", err)
		}
	})

	t.Run("url_alphabet_form_of_the_tilde_payload_fails_on_the_id_instead", func(t *testing.T) {
		// Distinguishes the alphabet rejection from the id rejection: identical
		// decoded bytes, only the alphabet differs, so the two failures must be
		// attributed to different stages. Both fixtures are stdlib-encoded.
		std := base64.StdEncoding.EncodeToString([]byte(tildePayload))
		stdErr := errOr(DecodeCursor(std, ids.Event))
		url := strings.NewReplacer("+", "-", "/", "_", "=", "").Replace(std)
		urlErr := errOr(DecodeCursor(url, ids.Event))
		if stdErr == nil || urlErr == nil {
			t.Fatalf("both forms accepted: std %v url %v", stdErr, urlErr)
		}
		if !strings.Contains(stdErr.Error(), "base64url") {
			t.Errorf("standard-alphabet error %q does not blame the alphabet", stdErr)
		}
		if strings.Contains(urlErr.Error(), "base64url") {
			t.Errorf("url-alphabet error %q should have reached the id check", urlErr)
		}
		if !strings.Contains(urlErr.Error(), "A0-1.5") {
			t.Errorf("url-alphabet error %q does not name A0-1.5", urlErr)
		}
	})

	t.Run("embedded_newline_is_rejected_though_the_stdlib_ignores_it", func(t *testing.T) {
		// Go's base64 decoder skips a trailing newline; A0-4.4 requires
		// byte-for-byte replay, so the re-encode comparison is what catches it.
		// Removing that comparison makes this fail while every other subtest of
		// this function still passes.
		s := encodeTestCursor(t, `{"id":"`+validID+`","k":4711}`)
		if _, err := DecodeCursor(s+"\n", ids.Event); err == nil {
			t.Error("DecodeCursor accepted a cursor with a trailing newline (A0-4.4, A0-8.6)")
		}
	})
}

// errOr returns the error half of a (value, error) pair, so a table of
// rejection attributions reads as a table of errors.
func errOr(_ Cursor, err error) error { return err }

// base64Class returns the encodings of payload that differ from the canonical
// one only in the final character yet still decode to payload's bytes — the
// non-canonical spelling stdlib decoders accept.
func base64Class(t *testing.T, canonical, payload string) []string {
	t.Helper()
	var out []string
	for _, r := range "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_" {
		mutated := canonical[:len(canonical)-1] + string(r)
		if mutated == canonical {
			continue
		}
		raw, err := base64.RawURLEncoding.DecodeString(mutated)
		if err == nil && string(raw) == payload {
			out = append(out, mutated)
		}
	}
	return out
}

func TestDecodeCursorRejects(t *testing.T) {
	long := strings.Repeat("a", 500)
	tests := []struct {
		name    string
		payload string // raw JSON text, base64url-encoded by the test only
		stage   string // the clause text the error must cite
	}{
		// A0-4.8 "non-canonical JSON": cjson.Canonical normalizes key order and
		// whitespace instead of rejecting them, so these rows are what proves
		// DecodeCursor's re-encode comparison exists. Deleting it passes every
		// of them while breaking A0-4.4's "replay byte-for-byte". The stage they
		// assert on is the re-encode step's own prose, not "canonical": the
		// cjson wrapper at paging.go's canonicalizing message also contains
		// "canonical", so that substring would match whichever step fired.
		{"unsorted keys", `{"k":4711,"id":"` + validID + `"}`, reencodeStage},
		{"whitespace after the colon", `{"id": "evt_01m1y2whfhp17g0avdqztd2p3x", "k":4711}`, reencodeStage},
		{"leading whitespace", ` {"id":"` + validID + `","k":4711}`, reencodeStage},
		{"trailing whitespace", `{"id":"` + validID + `","k":4711} `, reencodeStage},
		{"uppercase key id", `{"ID":"` + validID + `","k":4711}`, reencodeStage},
		{"uppercase key k", `{"id":"` + validID + `","K":4711}`, reencodeStage},
		{"number literal with a leading zero", `{"id":"` + validID + `","k":04711}`, "A0-2.5"},

		// Rejected inside cjson itself (A0-2.5, A0-2.10, A0-2.14, A0-2.3).
		{"duplicate key", `{"id":"` + validID + `","id":"` + otherKindID + `","k":4711}`, "duplicate"},
		{"case-duplicate key", `{"id":"` + validID + `","ID":"` + validID + `","k":4711}`, "duplicate"},
		{"trailing data", `{"id":"` + validID + `","k":4711}{"id":"` + validID + `","k":1}`, "top-level value"},
		{"top-level array", `[{"id":"` + validID + `","k":4711}]`, "JSON object"},
		{"top-level string", `"just a string"`, "JSON object"},
		{"top-level number", `4711`, "JSON object"},
		{"top-level null", `null`, "null"},
		{"null id", `{"id":null,"k":4711}`, "null"},
		{"null k", `{"id":"` + validID + `","k":null}`, "null"},
		{"float ordering value", `{"id":"` + validID + `","k":4711.0}`, "integer"},
		{"exponent ordering value", `{"id":"` + validID + `","k":4.711e3}`, "integer"},
		{"string ordering value", `{"id":"` + validID + `","k":"4711"}`, "key set"},
		{"boolean id", `{"id":true,"k":4711}`, "key set"},
		{"ordering value above 2^53-1", `{"id":"` + validID + `","k":9007199254740992}`, "integers in"},
		{"malformed json", `{"id":`, "token stream"},
		{"empty payload", ``, "empty"},

		// A0-4.8 "wrong key set": the two-key set of A0-4.4 is closed.
		{"missing id", `{"k":4711}`, `"id" key`},
		{"missing k", `{"id":"` + validID + `"}`, `"k" key`},
		{"extra key", `{"extra":true,"id":"` + validID + `","k":4711}`, "extra"},
		{"empty object", `{}`, "neither key"},
		{"nested object in k", `{"id":"` + validID + `","k":{"a":1}}`, "key set"},

		// A0-4.4/A0-1.5: the id must be byte-exact for the declared kind.
		{"id of another registered kind", `{"id":"` + otherKindID + `","k":4711}`, "A0-1.5"},
		{"bare body with no prefix", `{"id":"01m1y2whfhp17g0avdqztd2p3x","k":4711}`, "A0-1.5"},
		{"valid prefix with a short body", `{"id":"evt_01m1y2whfhp17g0avdqztd2p3","k":4711}`, "A0-1.5"},
		{"valid prefix with a long body", `{"id":"evt_01m1y2whfhp17g0avdqztd2p3xx","k":4711}`, "A0-1.5"},
		{"uppercase prefix", `{"id":"EVT_01m1y2whfhp17g0avdqztd2p3x","k":4711}`, "A0-1.5"},
		{"crockford i in the body", `{"id":"evt_01m1y2whfhp17g0avdqztd2p3i","k":4711}`, "A0-1.5"},
		{"crockford l in the body", `{"id":"evt_01m1y2whfhp17g0avdqztd2p3l","k":4711}`, "A0-1.5"},
		{"crockford o in the body", `{"id":"evt_01m1y2whfhp17g0avdqztd2p3o","k":4711}`, "A0-1.5"},
		{"crockford u in the body", `{"id":"evt_01m1y2whfhp17g0avdqztd2p3u","k":4711}`, "A0-1.5"},
		{"dashes in the body", `{"id":"evt_-1m1y2whfhp17g0avdqztd2p3x","k":4711}`, "A0-1.5"},
		{"unicode digit in the body", `{"id":"evt_01m1y2whfhp17g0avdqztd2٢3x","k":4711}`, "A0-1.5"},
		{"empty id", `{"id":"","k":4711}`, "A0-1.5"},
		{"over-long id", `{"id":"` + long + `","k":4711}`, "A0-1.5"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := encodeTestCursor(t, tc.payload)
			got, err := DecodeCursor(s, ids.Event)
			if err == nil {
				t.Fatalf("DecodeCursor accepted %q and returned %+v (A0-4.8)", tc.payload, got)
			}
			if kind := errs.KindOf(err); kind != errs.Validation {
				t.Errorf("kind = %q, want %q (A0-4.8, A0-3.1)", kind, errs.Validation)
			}
			if got != (Cursor{}) {
				t.Errorf("returned %+v alongside an error, want the zero Cursor", got)
			}
			msg := err.Error()
			if !strings.Contains(msg, tc.stage) {
				t.Errorf("error %q does not name %q", msg, tc.stage)
			}
			// A0-3.5, and the WP-08 ruling that a rejected cursor's value never
			// reaches the log: a cursor is opaque client material. Skipped for
			// fixtures too short to be distinctive (the empty payload encodes to
			// "", which every string "contains").
			if len(s) >= 12 && strings.Contains(msg, s) {
				t.Errorf("error %q echoes the rejected cursor %q", msg, s)
			}
			// The over-long id must not be echoed whole (A0-3.5: no request
			// bodies in messages).
			if strings.Contains(msg, long) {
				t.Errorf("error %q echoes a 500-character id in full", msg)
			}
		})
	}

	t.Run("kind_is_not_one_of_the_twelve_registered_prefixes", func(t *testing.T) {
		// A0-1.2's prefix set is closed: an unregistered ids.Kind accepts
		// nothing, so a handler that maps client text straight to a Kind cannot
		// page a collection that does not exist.
		s := encodeTestCursor(t, `{"id":"`+validID+`","k":4711}`)
		for _, k := range []ids.Kind{"", "evt", "zzz_", "Engagement", "node_"} {
			got, err := DecodeCursor(s, k)
			if err == nil {
				t.Errorf("DecodeCursor accepted kind %q and returned %+v", k, got)
				continue
			}
			if kind := errs.KindOf(err); kind != errs.Validation {
				t.Errorf("kind %q: error kind = %q, want %q", k, kind, errs.Validation)
			}
		}
	})

	t.Run("a_valid_cursor_still_decodes_after_every_rejection_row", func(t *testing.T) {
		// Guards the table above from passing because the package rejects
		// everything: the same call with the same arguments must succeed.
		s := encodeTestCursor(t, `{"id":"`+validID+`","k":4711}`)
		got, err := DecodeCursor(s, ids.Event)
		if err != nil {
			t.Fatalf("DecodeCursor: %v", err)
		}
		if want := (Cursor{K: 4711, ID: validID}); got != want {
			t.Errorf("got %+v, want %+v", got, want)
		}
	})

	t.Run("cjson_normalizes_key_order_so_the_reencode_check_is_load_bearing", func(t *testing.T) {
		// Why the "unsorted keys", "whitespace" and "uppercase key" rows above
		// can only be rejected by DecodeCursor's re-encode comparison: cjson
		// normalizes key order and whitespace instead of rejecting them, and
		// encoding/json matches a struct key case-insensitively, so
		// DisallowUnknownFields does not see "ID" as unknown. Deleting the
		// re-encode step from DecodeCursor lets these three through while every
		// other row still fails — that is the edit that would make the table's
		// assertions pass while breaking A0-4.4's "replay byte-for-byte" and
		// A0-6.2's "verify observed key spelling".
		for _, payload := range []string{
			`{"k":4711,"id":"` + validID + `"}`,
			`{"ID":"` + validID + `","k":4711}`,
			`{ "id" : "` + validID + `" , "k" : 4711 }`,
		} {
			if _, err := cjson.Canonical([]byte(payload)); err != nil {
				t.Fatalf("cjson.Canonical(%s) = %v; if cjson already rejects this, the row above does not prove the re-encode step", payload, err)
			}
			if _, err := DecodeCursor(encodeTestCursor(t, payload), ids.Event); err == nil {
				t.Errorf("DecodeCursor accepted %s, so nothing but the re-encode step stands between a client and a cursor A0-4.4 does not describe", payload)
			}
		}
	})

	t.Run("pointer_fields_are_what_make_an_absent_key_observable", func(t *testing.T) {
		// The guard this pins is DecodeCursor's *string/*int64 payload struct.
		// Changing either field to a value type makes a missing key decode to
		// the zero value, so {"k":4711} would hand back Cursor{ID:""} — a
		// half-built cursor (DESIGN §4) that A0-4.8 requires rejecting. The
		// re-encode step would still catch it, but the error would blame the
		// canonical form instead of naming the missing key, which is what the
		// two rows above assert on.
		s := encodeTestCursor(t, `{"k":4711}`)
		_, err := DecodeCursor(s, ids.Event)
		if err == nil {
			t.Fatal("a cursor with no id decoded successfully")
		}
		if !strings.Contains(err.Error(), `"id"`) {
			t.Errorf("error %q does not name the missing key", err)
		}
	})
}

func TestCursorWithInconsistentKAndIDRejected(t *testing.T) {
	// The guard this table pins is DecodeCursor's ids.Valid(k, c.ID) call: the
	// edit that would make every row below pass while breaking A0-4.4 is
	// deleting that call (or replacing it with a prefix-only comparison, which
	// accepts a 25-character body), and ids is then never consulted with the
	// collection's kind. The positive rows are part of the table for exactly
	// that reason: an unconditional rejection passes the negative rows.
	// A0-4.4 and A0-4.8 pair two requirements. Only one of them is checkable in
	// this package:
	//
	//  1. the cursor's id is not of the kind the collection declared → rejected
	//     here, by DecodeCursor's ids.Kind parameter (A0-4.4's last sentence,
	//     A0-1.5). Covered by the table below.
	//  2. the cursor's k disagrees with the ordering value of the row its id
	//     resolves to → NOT checkable here: A0 §4 gives this package no store,
	//     no collection and no row lookup, so there is nothing to compare k
	//     against. It belongs to whoever resolves a cursor against a collection
	//     (WP-19/WP-20/WP-22), and is reported as a contract-ownership gap.
	//
	// §4.1 fixes the oracle for the whole clause as validation (400), never an
	// empty page — so the subtests below assert the kind on every rejection, and
	// the last subtest asserts that this package has not grown a row lookup it
	// is not allowed to have.
	tests := []struct {
		name   string
		id     string
		kind   ids.Kind
		wantOK bool
	}{
		{"event cursor against the run collection", validID, ids.Run, false},
		{"event cursor against the graph-node collection", validID, ids.GraphNode, false},
		{"graph-node cursor against the event collection", otherKindID, ids.Event, false},
		{"engagement cursor against the event collection", "eng_01m1y2whfhp17g0avdqztd2p3x", ids.Event, false},
		{"evidence cursor against the approval collection", "evi_01m1y2whfhp17g0avdqztd2p3x", ids.Approval, false},
		{"event cursor against the event collection", validID, ids.Event, true},
		{"graph-node cursor against the graph-node collection", otherKindID, ids.GraphNode, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s, err := EncodeCursor(Cursor{K: publishedK, ID: tc.id})
			if err != nil {
				t.Fatalf("EncodeCursor: %v", err)
			}
			got, err := DecodeCursor(s, tc.kind)
			if tc.wantOK {
				if err != nil {
					t.Fatalf("DecodeCursor: %v", err)
				}
				if got.ID != tc.id || got.K != publishedK {
					t.Errorf("got %+v, want K=%d ID=%q", got, publishedK, tc.id)
				}
				return
			}
			if err == nil {
				t.Fatalf("DecodeCursor accepted %+v (A0-4.4)", got)
			}
			// The single oracle of §4.1: validation, 400 — never notfound,
			// never an empty page.
			if kind := errs.KindOf(err); kind != errs.Validation {
				t.Errorf("kind = %q, want %q", kind, errs.Validation)
			}
			if status := errs.KindOf(err).Status(); status != 400 {
				t.Errorf("status = %d, want 400 (A0-3.1)", status)
			}
			if got != (Cursor{}) {
				t.Errorf("returned %+v alongside an error", got)
			}
		})
	}

	t.Run("k_is_returned_verbatim_because_no_row_lookup_exists_here", func(t *testing.T) {
		// A0-4.4: the platform MUST ignore k for seeking and derive the position
		// from the resolved row. So DecodeCursor hands k back untouched for a
		// kind-valid id — including values no ordering key could ever take, which
		// a resolver is expected to catch against its own rows. This assertion
		// fails the moment anyone adds a store-shaped parameter or a row
		// callback to DecodeCursor, which is the A0 §4 signature this package
		// must not widen (DESIGN §2).
		s := encodeTestCursor(t, `{"id":"`+validID+`","k":-9007199254740991}`)
		got, err := DecodeCursor(s, ids.Event)
		if err != nil {
			t.Fatalf("DecodeCursor: %v", err)
		}
		if got.K != -9007199254740991 || got.ID != validID {
			t.Errorf("got %+v, want the tuple echoed back unchanged", got)
		}
	})
}

func TestHasMoreDetection(t *testing.T) {
	page := func(rows []task, limit int) Page[task] {
		t.Helper()
		p, err := NewPage(rows, limit, cursorOfTask)
		if err != nil {
			t.Fatalf("NewPage(limit=%d, rows=%d): %v", limit, len(rows), err)
		}
		return p
	}

	// A0-4.6's named contract test: a last page that exactly fills limit MUST
	// NOT carry next_cursor. §4.1 rules that this subtest is also
	// TestExactFullPageHasNoNextCursor and is not a second test id.
	t.Run("exactly_limit_rows_has_no_next_cursor", func(t *testing.T) {
		rows := []task{{ID: validID, K: 1}, {ID: otherKindID, K: 2}}
		p := page(rows, 2)
		if len(p.Items) != 2 {
			t.Fatalf("Items = %d rows, want 2", len(p.Items))
		}
		if p.NextCursor != "" {
			t.Errorf("NextCursor = %q, want absent (A0-4.2, A0-4.6)", p.NextCursor)
		}
		// Absent, not empty: A0-8.3. The key must not appear at all.
		b, err := json.Marshal(p)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		if want := `{"items":[{"id":"` + validID + `","k":1},{"id":"` + otherKindID + `","k":2}]}`; string(b) != want {
			t.Errorf("envelope = %s, want %s", b, want)
		}
	})

	t.Run("limit_plus_one_rows_has_the_last_returned_row_cursor", func(t *testing.T) {
		rows := []task{{ID: validID, K: 1}, {ID: otherKindID, K: 2}, {ID: "run_01m1y2whfhp17g0avdqztd2p3x", K: 3}}
		p := page(rows, 2)
		if len(p.Items) != 2 {
			t.Fatalf("Items = %d rows, want the first 2 (A0-4.6)", len(p.Items))
		}
		want, err := EncodeCursor(cursorOfTask(rows[1]))
		if err != nil {
			t.Fatalf("EncodeCursor: %v", err)
		}
		if p.NextCursor != want {
			t.Errorf("NextCursor = %q, want %q — the cursor of rows[limit-1], the LAST RETURNED row", p.NextCursor, want)
		}
		// The unreleased row must not be reachable through the cursor: its id is
		// not in the payload.
		if strings.Contains(p.NextCursor, encodeTestCursor(t, `{"id":"run_01m1y2whfhp17g0avdqztd2p3x","k":3}`)) {
			t.Errorf("NextCursor %q carries the row past the page boundary", p.NextCursor)
		}
	})

	t.Run("zero_rows", func(t *testing.T) {
		// A0-2.14: an empty page is [], never null. cjson.CanonicalValue refuses
		// a null at any depth, so a nil Items is a serialization failure at the
		// one place the platform cannot tolerate it.
		for _, rows := range [][]task{nil, {}} {
			p := page(rows, 5)
			if p.Items == nil {
				t.Fatal("Items is nil, want an empty non-nil slice (A0-2.14)")
			}
			if p.NextCursor != "" {
				t.Errorf("NextCursor = %q on an empty page", p.NextCursor)
			}
			doc, err := cjson.CanonicalValue(p)
			if err != nil {
				t.Fatalf("CanonicalValue(Page): %v", err)
			}
			if want := `{"items":[]}`; string(doc) != want {
				t.Errorf("envelope = %s, want %s", doc, want)
			}
		}
	})

	t.Run("one_and_two_rows_with_limit_one", func(t *testing.T) {
		one := []task{{ID: validID, K: 1}}
		p := page(one, 1)
		if len(p.Items) != 1 || p.NextCursor != "" {
			t.Errorf("one row at limit 1: items=%d next=%q, want 1 and absent", len(p.Items), p.NextCursor)
		}

		two := []task{{ID: validID, K: 1}, {ID: otherKindID, K: 2}}
		p = page(two, 1)
		if len(p.Items) != 1 {
			t.Fatalf("two rows at limit 1: items=%d, want 1", len(p.Items))
		}
		want, err := EncodeCursor(cursorOfTask(two[0]))
		if err != nil {
			t.Fatalf("EncodeCursor: %v", err)
		}
		if p.NextCursor != want {
			t.Errorf("NextCursor = %q, want the cursor of the returned row %q", p.NextCursor, want)
		}
		// A client decoding the cursor of a full page must get the row it came
		// from, or the seek lands on the wrong row (A0-4.4).
		got, err := DecodeCursor(p.NextCursor, ids.Event)
		if err != nil {
			t.Fatalf("DecodeCursor: %v", err)
		}
		if got != cursorOfTask(two[0]) {
			t.Errorf("decoded %+v, want %+v", got, cursorOfTask(two[0]))
		}
	})

	t.Run("more_than_limit_plus_one_rows", func(t *testing.T) {
		// A caller that read further than limit+1 still gets exactly limit rows:
		// the envelope cannot leak a row the page did not authorize.
		rows := make([]task, 5)
		for i := range rows {
			rows[i] = task{ID: validID, K: int64(i + 1)}
		}
		p := page(rows, 2)
		if len(p.Items) != 2 {
			t.Fatalf("items = %d, want 2", len(p.Items))
		}
		want, err := EncodeCursor(cursorOfTask(rows[1]))
		if err != nil {
			t.Fatalf("EncodeCursor: %v", err)
		}
		if p.NextCursor != want {
			t.Errorf("NextCursor = %q, want %q", p.NextCursor, want)
		}
	})

	t.Run("cursor_of_a_full_page_is_built_from_the_last_row_not_from_the_row_after_it", func(t *testing.T) {
		// The off-by-one this clause is about: rows[limit] is the row that does
		// NOT belong to the client.
		rows := []task{{ID: validID, K: 10}, {ID: otherKindID, K: 20}, {ID: "run_01m1y2whfhp17g0avdqztd2p3x", K: 30}}
		p := page(rows, 2)
		after, err := EncodeCursor(cursorOfTask(rows[2]))
		if err != nil {
			t.Fatalf("EncodeCursor: %v", err)
		}
		if p.NextCursor == after {
			t.Errorf("NextCursor is the cursor of the unreleased row (A0-4.4)")
		}
	})

	t.Run("encode_failure_from_a_row_is_not_swallowed", func(t *testing.T) {
		// cursorOf handed NewPage a row whose ordering value is outside A0-2.6's
		// canonical integer range. Returning a page with no cursor here would
		// tell the client the collection ended (A0-4.2), so the failure must
		// travel outward with the validation kind cjson gave it. Replacing the
		// wrap with `return Page[T]{Items: rows[:limit]}, nil` is the edit this
		// row catches.
		bad := []task{{ID: validID, K: 9007199254740992}, {ID: otherKindID, K: 1}}
		p, err := NewPage(bad, 1, func(r task) Cursor { return Cursor{K: r.K, ID: r.ID} })
		if err == nil {
			t.Fatalf("NewPage accepted an unrepresentable ordering value and returned %+v", p)
		}
		if kind := errs.KindOf(err); kind != errs.Validation {
			t.Errorf("kind = %q, want %q (A0-2.6, A0-3.1)", kind, errs.Validation)
		}
		if len(p.Items) != 0 || p.NextCursor != "" {
			t.Errorf("returned %+v alongside an error", p)
		}
	})

	t.Run("nil_cursor_of_with_more_rows_is_internal_not_a_panic", func(t *testing.T) {
		// ADR-0019 §6 forbids a panic in a request path and A0-3.1 makes a
		// wiring defect internal. Returning an exhausted page instead would
		// silently drop the rest of the collection.
		rows := []task{{ID: validID, K: 1}, {ID: otherKindID, K: 2}}
		p, err := NewPage(rows, 1, nil)
		if err == nil {
			t.Fatalf("NewPage accepted a nil cursorOf and returned %+v", p)
		}
		if kind := errs.KindOf(err); kind != errs.Internal {
			t.Errorf("kind = %q, want %q", kind, errs.Internal)
		}
		if len(p.Items) != 0 || p.NextCursor != "" {
			t.Errorf("returned %+v alongside an error, want the zero Page", p)
		}
		// An exhausted page needs no cursor function at all, so a nil one is
		// not a defect there (DESIGN §2: no check without a reachable failure).
		if _, err := NewPage([]task{{ID: validID, K: 1}}, 1, nil); err != nil {
			t.Errorf("NewPage on a full-but-last page: %v", err)
		}
	})
}

func TestLimitValidation(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		present bool
		want    int
		reject  bool
	}{
		{name: "absent means the default", present: false, want: DefaultLimit},
		{name: "absent ignores the stale raw value", raw: "", present: false, want: DefaultLimit},
		{name: "one", raw: "1", present: true, want: 1},
		{name: "nine hundred ninety nine", raw: "999", present: true, want: 999},
		{name: "at the maximum", raw: "1000", present: true, want: MaxLimit},
		{name: "the default spelled out", raw: "100", present: true, want: 100},

		{name: "present and empty", raw: "", present: true, reject: true},
		{name: "zero", raw: "0", present: true, reject: true},
		{name: "negative", raw: "-1", present: true, reject: true},
		{name: "negative one written with a plus", raw: "+1", present: true, reject: true},
		{name: "above the maximum", raw: "1001", present: true, reject: true},
		{name: "far above the maximum", raw: "100000", present: true, reject: true},
		{name: "float", raw: "1.0", present: true, reject: true},
		{name: "exponent", raw: "1e3", present: true, reject: true},
		{name: "hexadecimal", raw: "0x10", present: true, reject: true},
		{name: "binary", raw: "0b11", present: true, reject: true},
		{name: "underscore separator", raw: "1_000", present: true, reject: true},
		{name: "leading space", raw: " 1", present: true, reject: true},
		{name: "trailing space", raw: "1 ", present: true, reject: true},
		{name: "trailing newline", raw: "1\n", present: true, reject: true},
		{name: "unicode digits", raw: "١٠", present: true, reject: true},
		{name: "word", raw: "ten", present: true, reject: true},
		{name: "overflow", raw: "99999999999999999999999", present: true, reject: true},
		{name: "just a sign", raw: "-", present: true, reject: true},
		{name: "percent encoded space survived decoding", raw: "%20", present: true, reject: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseLimit(tc.raw, tc.present)
			if !tc.reject {
				if err != nil {
					t.Fatalf("ParseLimit(%q, %v): %v", tc.raw, tc.present, err)
				}
				if got != tc.want {
					t.Errorf("ParseLimit(%q, %v) = %d, want %d", tc.raw, tc.present, got, tc.want)
				}
				return
			}
			if err == nil {
				t.Fatalf("ParseLimit(%q, %v) = %d, want a rejection (A0-4.5)", tc.raw, tc.present, got)
			}
			if kind := errs.KindOf(err); kind != errs.Validation {
				t.Errorf("kind = %q, want %q", kind, errs.Validation)
			}
			// Never a usable value alongside a rejection: this is the clamp
			// trap in its plainest form.
			if got != 0 {
				t.Errorf("returned %d alongside an error, want 0", got)
			}
			if !strings.Contains(err.Error(), "A0-4.5") {
				t.Errorf("error %q does not cite A0-4.5", err)
			}
		})
	}

	t.Run("newpage_rejects_the_same_range", func(t *testing.T) {
		// A0-4.5 bites twice: the query parameter and the row-slice truncation.
		// A NewPage that clamped instead of rejecting would hand the caller a
		// 1000-row page for a limit-1001 read, i.e. a silent truncation of rows
		// the caller already had.
		rows := make([]task, 1002)
		for i := range rows {
			rows[i] = task{ID: validID, K: int64(i + 1)}
		}
		for _, limit := range []int{0, -1, -1000, MaxLimit + 1, MaxLimit * 10} {
			p, err := NewPage(rows, limit, cursorOfTask)
			if err == nil {
				t.Fatalf("NewPage(limit=%d) accepted, returned %d items", limit, len(p.Items))
			}
			if kind := errs.KindOf(err); kind != errs.Validation {
				t.Errorf("limit %d: kind = %q, want %q", limit, kind, errs.Validation)
			}
			// ADR-0019 §2 wants the origin as component.Function. The generic
			// instantiation is named "paging.NewPage[...]" by runtime.Caller, so
			// the assertion fails if internal/errs ever stops stripping the
			// type-argument list — the defect is invisible in the message, which
			// reads plausibly either way, so nothing else here would notice.
			if op := errs.OpOf(err); op != "paging.NewPage" {
				t.Errorf("limit %d: errs.OpOf(err) = %q, want %q", limit, op, "paging.NewPage")
			}
			if len(p.Items) != 0 {
				t.Errorf("limit %d: returned %d items alongside an error", limit, len(p.Items))
			}
			if !strings.Contains(err.Error(), "A0-4.5") {
				t.Errorf("limit %d: error %q does not cite A0-4.5", limit, err)
			}
		}
	})

	t.Run("newpage_accepts_the_boundaries", func(t *testing.T) {
		for _, limit := range []int{1, DefaultLimit, MaxLimit} {
			if _, err := NewPage([]task{{ID: validID, K: 1}}, limit, cursorOfTask); err != nil {
				t.Errorf("NewPage(limit=%d): %v", limit, err)
			}
		}
	})

	t.Run("the_default_and_the_maximum_are_the_a0_4_5_numbers", func(t *testing.T) {
		// Restated from A0-4.5 and §6 item 10 rather than from the package's own
		// constants, so a drift in the constants is caught here.
		if DefaultLimit != 100 {
			t.Errorf("DefaultLimit = %d, want 100 (A0-4.5)", DefaultLimit)
		}
		if MaxLimit != 1000 {
			t.Errorf("MaxLimit = %d, want 1000 (A0-4.5)", MaxLimit)
		}
	})
}

func TestLimitAboveMaxRejectedNotClamped(t *testing.T) {
	// A0-4.5's decided rule (product owner, 2026-09-21): a client asking for
	// more than the maximum is refused. Clamping is the defect this test exists
	// to catch — an agent that received 1000 items would believe it had seen
	// the whole collection.
	for _, raw := range []string{"1001", "1002", "2000", "100000", "99999999999999999999999"} {
		got, err := ParseLimit(raw, true)
		if err == nil {
			t.Fatalf("ParseLimit(%q) = %d, want rejection", raw, got)
		}
		if kind := errs.KindOf(err); kind != errs.Validation {
			t.Errorf("%q: kind = %q, want %q", raw, kind, errs.Validation)
		}
		if got == MaxLimit {
			t.Errorf("%q: returned the clamped maximum — A0-4.5 forbids it", raw)
		}
		if got != 0 {
			t.Errorf("%q: returned %d alongside an error, want 0", raw, got)
		}
		if !strings.Contains(err.Error(), "A0-4.5") {
			t.Errorf("%q: error %q does not cite A0-4.5", raw, err)
		}
	}

	rows := make([]task, MaxLimit+2)
	for i := range rows {
		rows[i] = task{ID: validID, K: int64(i + 1)}
	}
	for _, limit := range []int{MaxLimit + 1, MaxLimit + 500, 1 << 20} {
		p, err := NewPage(rows, limit, cursorOfTask)
		if err == nil {
			t.Fatalf("NewPage(limit=%d) accepted and returned %d items", limit, len(p.Items))
		}
		if kind := errs.KindOf(err); kind != errs.Validation {
			t.Errorf("limit %d: kind = %q, want %q", limit, kind, errs.Validation)
		}
		// Not clamped to MaxLimit either: the page is empty, not truncated.
		if len(p.Items) == MaxLimit || len(p.Items) == limit {
			t.Errorf("limit %d: returned %d items alongside an error, want none", limit, len(p.Items))
		}
		if len(p.Items) != 0 {
			t.Errorf("limit %d: returned %d items alongside an error", limit, len(p.Items))
		}
		if p.NextCursor != "" {
			t.Errorf("limit %d: returned cursor %q alongside an error", limit, p.NextCursor)
		}
	}

	t.Run("the_maximum_itself_is_not_rejected", func(t *testing.T) {
		// The boundary must not be read as exclusive: off-by-one here turns a
		// legal request into a 400.
		if got, err := ParseLimit("1000", true); err != nil || got != 1000 {
			t.Errorf("ParseLimit(\"1000\") = %d, %v, want 1000, nil", got, err)
		}
		p, err := NewPage(rows[:MaxLimit+1], MaxLimit, cursorOfTask)
		if err != nil {
			t.Fatalf("NewPage at MaxLimit: %v", err)
		}
		if len(p.Items) != MaxLimit {
			t.Errorf("items = %d, want %d", len(p.Items), MaxLimit)
		}
		want, err := EncodeCursor(cursorOfTask(rows[MaxLimit-1]))
		if err != nil {
			t.Fatalf("EncodeCursor: %v", err)
		}
		if p.NextCursor != want {
			t.Errorf("NextCursor = %q, want %q", p.NextCursor, want)
		}
	})
}
