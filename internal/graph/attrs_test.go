package graph

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/daten-krake/sleipnir/internal/errs"
)

func TestAttrValueRoundTrip(t *testing.T) {
	// P-51 / A2-6.1: MarshalJSON emits the BARE scalar of the live field, the
	// Type discriminator is never serialized, and the round trip preserves
	// Type. AttrValue is comparable, so round-trip equality is ==.
	tests := []struct {
		name     string
		value    AttrValue
		wantWire string
	}{
		{"string", AttrValue{Type: AttrString, Str: "ntlmrelayx"}, `"ntlmrelayx"`},
		{"empty_string", AttrValue{Type: AttrString, Str: ""}, `""`},
		{"string_with_escapes", AttrValue{Type: AttrString, Str: "a\"b\\c\nd\te"}, `"a\"b\\c\nd\te"`},
		{"string_unicode_literal", AttrValue{Type: AttrString, Str: "é中😀"}, `"é中😀"`},
		// encoding/json escapes U+2028 in the intermediate wire form; the
		// round trip must still restore the rune. (Canonical output re-emits
		// it literally per A0-2.7 — proven by cjson's own V7, and §4.2's
		// vectors contain no U+2028.)
		{"string_u2028", AttrValue{Type: AttrString, Str: "a\u2028b"}, `"a\u2028b"`},
		{"int_zero", AttrValue{Type: AttrInt, Num: 0}, `0`},
		{"int_positive", AttrValue{Type: AttrInt, Num: 88}, `88`},
		{"int_negative", AttrValue{Type: AttrInt, Num: -42}, `-42`},
		{"int_max_safe", AttrValue{Type: AttrInt, Num: 9007199254740991}, `9007199254740991`},
		{"int_min_safe", AttrValue{Type: AttrInt, Num: -9007199254740991}, `-9007199254740991`},
		{"bool_true", AttrValue{Type: AttrBool, Bool: true}, `true`},
		{"bool_false", AttrValue{Type: AttrBool, Bool: false}, `false`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wire, err := json.Marshal(tc.value)
			if err != nil {
				t.Fatalf("MarshalJSON: %v", err)
			}
			if string(wire) != tc.wantWire {
				t.Fatalf("wire bytes = %s, want the bare scalar %s", wire, tc.wantWire)
			}
			var got AttrValue
			if err := json.Unmarshal(wire, &got); err != nil {
				t.Fatalf("UnmarshalJSON(%s): %v", wire, err)
			}
			if got != tc.value {
				t.Errorf("round trip = %+v, want %+v — Type must be preserved (A2-6.1)", got, tc.value)
			}
		})
	}

	t.Run("attrs_map_round_trip", func(t *testing.T) {
		in := Attrs{
			"cvss_v3_x10": {Type: AttrInt, Num: 88},
			"relay_tool":  {Type: AttrString, Str: "ntlmrelayx"},
			"flag":        {Type: AttrBool, Bool: true},
		}
		wire, err := json.Marshal(in)
		if err != nil {
			t.Fatalf("marshal Attrs: %v", err)
		}
		// The wire form of attrs is {"<key>": <bare scalar>}, depth exactly
		// one, no wrapper object, keys sorted (A2-6.1).
		want := `{"cvss_v3_x10":88,"flag":true,"relay_tool":"ntlmrelayx"}`
		if string(wire) != want {
			t.Fatalf("Attrs wire bytes = %s, want %s", wire, want)
		}
		var got Attrs
		if err := json.Unmarshal(wire, &got); err != nil {
			t.Fatalf("unmarshal Attrs: %v", err)
		}
		if len(got) != len(in) {
			t.Fatalf("round trip lost entries: %v", got)
		}
		for k, wantV := range in {
			if got[k] != wantV {
				t.Errorf("attrs[%q] round trip = %+v, want %+v", k, got[k], wantV)
			}
		}
	})

	t.Run("surrogate_handling", func(t *testing.T) {
		// A0-2.3 (b) / A0-2.17: a well-formed pair is accepted and decodes
		// to its rune; a lone surrogate is REJECTED, not U+FFFD-replaced —
		// a normalized value would enter content_hash as bytes the writer
		// never sent.
		var got AttrValue
		if err := json.Unmarshal([]byte(`"\ud83d\ude00"`), &got); err != nil {
			t.Fatalf("well-formed surrogate pair rejected: %v", err)
		}
		if got != (AttrValue{Type: AttrString, Str: "😀"}) {
			t.Errorf("surrogate pair decoded to %+v, want the 😀 string value", got)
		}
		for _, in := range []string{`"\ud800"`, `"\udc00"`, `"\ud83d"`, `"\ud83dx"`, `"a\ud800b"`} {
			var v AttrValue
			err := json.Unmarshal([]byte(in), &v)
			if err == nil {
				t.Errorf("lone surrogate %s accepted as %+v — must be rejected (A0-2.3)", in, v)
				continue
			}
			if kind := errs.KindOf(err); kind != errs.Validation {
				t.Errorf("lone surrogate %s: kind = %q, want validation", in, kind)
			}
		}
		// An escaped backslash followed by the literal text "ud800" is NOT
		// an escape and must be accepted verbatim.
		var lit AttrValue
		if err := json.Unmarshal([]byte(`"\\ud800"`), &lit); err != nil {
			t.Fatalf(`escaped backslash + literal text rejected: %v`, err)
		}
		if lit != (AttrValue{Type: AttrString, Str: `\ud800`}) {
			t.Errorf(`escaped-backslash value = %+v, want Str %q`, lit, `\ud800`)
		}
	})

	t.Run("integer_literal_form_is_canonical_only", func(t *testing.T) {
		// A0-2.5/A0-2.6: the decoder accepts exactly the literal text the
		// canonicalizer re-emits verbatim. Direct calls: several of these
		// forms are not syntactically valid JSON, so encoding/json's scanner
		// would reject them before AttrValue.UnmarshalJSON ever saw them
		// (that outer mapping is WP-15's request-decode job).
		rejects := []string{`1.5`, `1e3`, `1E3`, `-0`, `-0.0`, `01`, `+1`,
			`9007199254740993`, `10000000000000000000`, `.5`, `-.5`, `0x10`}
		for _, in := range rejects {
			var v AttrValue
			err := v.UnmarshalJSON([]byte(in))
			if err == nil {
				t.Errorf("integer literal %s accepted as %+v — floats, exponents, leading zeros, '+', '-0' and out-of-range values are rejected (A2-6.1, A0-2.6)", in, v)
				continue
			}
			if kind := errs.KindOf(err); kind != errs.Validation {
				t.Errorf("integer literal %s: kind = %q, want validation", in, kind)
			}
		}
		accepts := map[string]int64{`0`: 0, `7`: 7, `-7`: -7,
			`9007199254740991`: 9007199254740991, `-9007199254740991`: -9007199254740991}
		for in, want := range accepts {
			// Through json.Unmarshal, so the integration path (a bare
			// scalar value inside a decoded document) is what is pinned.
			var v AttrValue
			if err := json.Unmarshal([]byte(in), &v); err != nil {
				t.Errorf("canonical integer %s rejected: %v", in, err)
				continue
			}
			if v != (AttrValue{Type: AttrInt, Num: want}) {
				t.Errorf("canonical integer %s decoded to %+v, want Num %d", in, v, want)
			}
		}
	})

	t.Run("hand_built_defects_are_internal", func(t *testing.T) {
		// A decoded AttrValue always carries one of the three discriminators
		// and an in-range Num; anything else is a hand-built platform defect
		// and must fail loudly, never marshal as null or a float-shaped
		// number (A2-6.1, A0-2.14's defect rule by analogy).
		defects := []AttrValue{
			{},                                  // no live field
			{Type: AttrType("bogus"), Str: "x"}, // unknown discriminator
			{Type: AttrInt, Num: 1 << 62},       // outside A0-2.6
			{Type: AttrInt, Num: -(1 << 62)},    // outside A0-2.6
		}
		for _, v := range defects {
			if _, err := json.Marshal(v); err == nil {
				t.Errorf("hand-built %+v marshaled without error — want internal", v)
			} else if kind := errs.KindOf(err); kind != errs.Internal {
				t.Errorf("hand-built %+v: kind = %q, want internal", v, kind)
			}
		}
	})
}

func TestAttrsRejectNestedFloatNull(t *testing.T) {
	// A2-6.1: UnmarshalJSON accepts a flat object of bare scalars only and
	// rejects nested objects, arrays, floats and null with errs kind
	// "validation". The rejects table also carries the key-shape (A2-6.2),
	// duplicate-key (A0-2.5) and structural cases the same walk owns.
	//
	// Every input in this table is SYNTACTICALLY VALID JSON on purpose:
	// json.Unmarshal runs its own scanner over the whole input first, so a
	// syntactically invalid body (01, +1, trailing data, truncated input)
	// fails as a foreign *json.SyntaxError before any UnmarshalJSON runs —
	// mapping that foreign error onto the A0-3 vocabulary is the
	// request-decode path's job (WP-15, following internal/paging's ruling).
	// The type-level guards for those forms are proven through direct
	// UnmarshalJSON calls in the subtest below.
	rejects := []struct {
		name  string
		input string
	}{
		{"nested_object", `{"k":{"a":1}}`},
		{"nested_array", `{"k":[1,2]}`},
		{"nested_array_of_objects", `{"k":[{"a":1}]}`},
		{"deeply_nested", `{"k":{"d":{"e":{"f":1}}}}`},
		{"null_value", `{"k":null}`},
		{"top_level_null", `null`},
		{"top_level_array", `[{"a":1}]`},
		{"top_level_scalar_string", `"x"`},
		{"top_level_scalar_number", `123`},
		{"float_value", `{"k":1.5}`},
		{"float_exponent", `{"k":1e3}`},
		{"float_upper_exponent", `{"k":1E3}`},
		{"negative_zero", `{"k":-0}`},
		{"float_negative_zero", `{"k":-0.0}`},
		{"int_above_safe_range", `{"k":9007199254740993}`},
		{"int_twenty_digits", `{"k":10000000000000000000}`},
		{"duplicate_key", `{"a":1,"a":2}`},
		{"case_duplicate_key", `{"a":1,"A":2}`},
		{"key_uppercase", `{"A":1}`},
		{"key_leading_digit", `{"1a":1}`},
		{"key_hyphen", `{"a-b":1}`},
		{"key_space", `{"a b":1}`},
		{"key_empty", `{"":1}`},
		{"key_41_chars", `{"` + "a" + strings.Repeat("b", 40) + `":1}`},
		{"key_lone_surrogate_string", `{"k":"\ud800"}`},
	}
	for _, tc := range rejects {
		t.Run("reject_"+tc.name, func(t *testing.T) {
			var a Attrs
			err := json.Unmarshal([]byte(tc.input), &a)
			if err == nil {
				t.Fatalf("input %s accepted as %v — must be rejected (A2-6.1)", tc.input, a)
			}
			if kind := errs.KindOf(err); kind != errs.Validation {
				t.Errorf("input %s: kind = %q, want validation", tc.input, kind)
			}
			if a != nil {
				t.Errorf("input %s: rejected decode left partial state %v", tc.input, a)
			}
		})
	}

	t.Run("direct_call_structural_guards", func(t *testing.T) {
		// Called directly — as WP-15's request-decode Token() walk will do
		// with raw value spans — the type guards the forms encoding/json's
		// own scanner would otherwise absorb before UnmarshalJSON ever runs.
		for _, in := range []string{
			`{"k":01}`, `{"k":+1}`, `{"a":1} {"b":2}`, ``, `{`, `{"a":1`,
		} {
			var a Attrs
			err := a.UnmarshalJSON([]byte(in))
			if err == nil {
				t.Errorf("direct UnmarshalJSON(%q) accepted %v — must be rejected", in, a)
				continue
			}
			if kind := errs.KindOf(err); kind != errs.Validation {
				t.Errorf("direct UnmarshalJSON(%q): kind = %q, want validation", in, kind)
			}
		}
	})

	t.Run("accepts_flat_scalars", func(t *testing.T) {
		var a Attrs
		in := `{"cvss_v3_x10":88,"first_seen_task":"task_01m1y2whfh1txm57x8dn41r9hg","relay_tool":"ntlmrelayx","seen":true,"unseen":false,"zero":0}`
		if err := json.Unmarshal([]byte(in), &a); err != nil {
			t.Fatalf("flat scalar object rejected: %v", err)
		}
		want := Attrs{
			"cvss_v3_x10":     {Type: AttrInt, Num: 88},
			"first_seen_task": {Type: AttrString, Str: "task_01m1y2whfh1txm57x8dn41r9hg"},
			"relay_tool":      {Type: AttrString, Str: "ntlmrelayx"},
			"seen":            {Type: AttrBool, Bool: true},
			"unseen":          {Type: AttrBool, Bool: false},
			"zero":            {Type: AttrInt, Num: 0},
		}
		for k, w := range want {
			if a[k] != w {
				t.Errorf("attrs[%q] = %+v, want %+v", k, a[k], w)
			}
		}
	})

	t.Run("accepts_empty_object_as_non_nil_map", func(t *testing.T) {
		var a Attrs
		if err := json.Unmarshal([]byte(`{}`), &a); err != nil {
			t.Fatalf("empty attrs object rejected: %v", err)
		}
		if a == nil {
			t.Fatal("decoded {} must be a non-nil empty map: a nil Attrs marshals back as null (A0-2.14)")
		}
		if len(a) != 0 {
			t.Errorf("decoded {} = %v, want empty", a)
		}
	})

	t.Run("rejected_value_content_is_never_echoed", func(t *testing.T) {
		// A2-9.5's discipline applied early: the error names the key and the
		// rejection class, never the content of a rejected value — attrs is
		// untrusted input that may carry secret material, and shape
		// rejection happens BEFORE the secret scan (A2-10.2 step 9 < 10).
		var a Attrs
		err := json.Unmarshal([]byte(`{"k":{"deep":"SENTINEL-SECRET-MATERIAL"}}`), &a)
		if err == nil {
			t.Fatal("nested object accepted")
		}
		if strings.Contains(err.Error(), "SENTINEL") {
			t.Errorf("error echoes rejected value content: %v", err)
		}
		if !strings.Contains(err.Error(), `"k"`) {
			t.Errorf("error must name the offending key (A0-3.4): %v", err)
		}
	})

	// A2-6.3: the reserved set is normative, closed and byte-exact. The
	// expected list is parsed from the contract's own ReservedAttrKeys
	// literal (pinning rule), never transcribed from the package's map.
	text := contractText(t, contractRelPath)
	parsed := parseReservedAttrKeys(t, contractSlice(t, text, "### A2-6", "### A2-7"))

	t.Run("reserved_set_matches_contract_byte_exactly", func(t *testing.T) {
		if len(parsed) != 51 {
			t.Fatalf("A2-6.3 parses to %d reserved keys, want exactly 51: %v", len(parsed), parsed)
		}
		if len(reservedAttrKeys) != len(parsed) {
			t.Fatalf("package reserved set has %d keys, contract has %d", len(reservedAttrKeys), len(parsed))
		}
		for _, k := range parsed {
			if !reservedAttrKeys[k] {
				t.Errorf("contract reserved key %q is missing from the package set", k)
			}
		}
		for k := range reservedAttrKeys {
			if !slices.Contains(parsed, k) {
				t.Errorf("package reserved key %q is not in the contract list", k)
			}
		}
	})

	t.Run("every_reserved_key_is_rejected", func(t *testing.T) {
		for _, k := range parsed {
			var a Attrs
			err := json.Unmarshal([]byte(`{"`+k+`":"v"}`), &a)
			if err == nil {
				t.Errorf("reserved key %q accepted as %v (A2-6.3)", k, a)
				continue
			}
			if kind := errs.KindOf(err); kind != errs.Validation {
				t.Errorf("reserved key %q: kind = %q, want validation", k, kind)
			}
		}
	})

	t.Run("membership_is_byte_exact_not_substring", func(t *testing.T) {
		// A2-6.3: "Membership MUST be tested byte-exactly — a prefix or
		// substring match MUST NOT be used."
		accepts := []string{"ids", "id2", "my_label", "labels", "graph_node_ids",
			"sequence", "confidences", "attacker_status", "kind_of", "provenance_note"}
		for _, k := range accepts {
			var a Attrs
			if err := json.Unmarshal([]byte(`{"`+k+`":"v"}`), &a); err != nil {
				t.Errorf("near-miss key %q rejected: %v — membership is byte-exact, no prefix/substring matching", k, err)
			}
		}
		rejectsExact := []string{"seq", "confidence", "graph_node_id", "label", "status", "provenance"}
		for _, k := range rejectsExact {
			var a Attrs
			if err := json.Unmarshal([]byte(`{"`+k+`":"v"}`), &a); err == nil {
				t.Errorf("byte-exact reserved key %q accepted (A2-6.3)", k)
			}
		}
	})

	t.Run("mutation_proof", func(t *testing.T) {
		mutated := mutateContract(t, text, "ReservedAttrKeys = {id, engagement_id, seq", "ReservedAttrKeys = {idd, engagement_id, seq")
		path := writeTempContract(t, mutated)
		got := parseReservedAttrKeys(t, contractSlice(t, contractText(t, path), "### A2-6", "### A2-7"))
		same := len(got) == len(reservedAttrKeys)
		for _, k := range got {
			if !reservedAttrKeys[k] {
				same = false
				break
			}
		}
		if same {
			t.Fatalf("the pin did not fail on the mutated copy: id→idd still parsed as the package's reserved set")
		}
	})
}
