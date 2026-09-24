package caps

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"
)

// allASCII reports whether every byte of s is below utf8.RuneSelf, i.e. the
// value is pure ASCII and therefore one byte per rune.
func allASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			return false
		}
	}
	return true
}

// TestConstantsMatchA0Table pins every row of the A0-7.1 registry table to its
// literal contract value, so a silent edit of a cap fails here (A0-7.2: these
// values change only via ADR). The table is the test data: one row per
// constant, so a missing constant is a missing row. TruncationMarker is not a
// registry row — its literal, length and ASCII-ness are asserted once, in
// TestTruncationMarkerIsCountedAgainstTheCap.
func TestConstantsMatchA0Table(t *testing.T) {
	rows := []struct {
		name string
		got  int
		want int
	}{
		// Q4 (A0-7.1).
		{"StageViewMaxBytes", StageViewMaxBytes, 65536},
		{"StageViewMaxNodes", StageViewMaxNodes, 500},
		{"NodeSummaryMaxBytes", NodeSummaryMaxBytes, 512},
		{"FindingSummaryMaxBytes", FindingSummaryMaxBytes, 2048},
		{"StageSummaryMaxBytes", StageSummaryMaxBytes, 2048},

		// A1 (A1-4.5/4.7).
		{"EventMaxCanonicalBytes", EventMaxCanonicalBytes, 32768},
		{"ProseLongMaxBytes", ProseLongMaxBytes, 2048},
		{"ProseMediumMaxBytes", ProseMediumMaxBytes, 512},
		{"TargetMaxBytes", TargetMaxBytes, 256},
		{"LabelMaxBytes", LabelMaxBytes, 128},
		{"ToolVersionMaxBytes", ToolVersionMaxBytes, 64}, // A2's 32 is a defect (A0-7.1)
		{"KindNameMaxBytes", KindNameMaxBytes, 32},
		{"DigestMaxBytes", DigestMaxBytes, 256},
		{"EvidenceRefsMax", EvidenceRefsMax, 8}, // replaces A2's EvidenceIDsMax
		{"EventRefsMax", EventRefsMax, 64},
		{"ExitCodeMin", ExitCodeMin, -1},
		{"ExitCodeMax", ExitCodeMax, 255},
		{"IdempotencyKeyMaxBytes", IdempotencyKeyMaxBytes, 64},

		// A2 (A2-6.4/7.1).
		{"NodeLabelMaxBytes", NodeLabelMaxBytes, 128},
		{"HypothesisClaimMaxBytes", HypothesisClaimMaxBytes, 512},
		{"HypothesisBasisMaxBytes", HypothesisBasisMaxBytes, 1024},
		{"AttrValueMaxBytes", AttrValueMaxBytes, 512},
		{"AttrsTotalMaxBytes", AttrsTotalMaxBytes, 4096},
		{"AttrsMaxKeys", AttrsMaxKeys, 16},
		{"AttrKeyMaxBytes", AttrKeyMaxBytes, 40}, // A0-8.1's regex bound
		{"AddressesMax", AddressesMax, 16},
		{"AddressMaxBytes", AddressMaxBytes, 64},
		{"MaxSupersedeChain", MaxSupersedeChain, 64},
	}

	const registryRows = 28 // A0-7.1: 5 Q4 + 13 A1 + 10 A2 constants
	if len(rows) != registryRows {
		t.Fatalf("registry rows = %d, want %d (one row per A0-7.1 constant)", len(rows), registryRows)
	}

	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			if r.got != r.want {
				t.Errorf("%s = %d, want %d (A0-7.1 table; change only via ADR, A0-7.2)", r.name, r.got, r.want)
			}
		})
	}
}

// truncateCase is one mechanism-T input: a value, a cap, the expected output
// and whether anything was cut.
type truncateCase struct {
	name  string
	in    string
	limit int
	want  string
	cut   bool
}

// truncateCorpus is the mechanism-T corpus shared by TestTruncateRuneBoundary
// and TestNoSilentTruncationInHashedRecords, so the no-silent-truncation
// property is asserted over exactly the same values. Marker lengths are derived
// from len(TruncationMarker), never hardcoded (A0-7.5).
func truncateCorpus() []truncateCase {
	m := TruncationMarker
	// withRune puts a multi-byte rune at byte offset 10 of an otherwise-ASCII
	// value that is far longer than any limit used below, so the cut lands
	// inside that rune.
	withRune := func(r string) string {
		return strings.Repeat("a", 10) + r + strings.Repeat("b", 20)
	}
	prefix10 := strings.Repeat("a", 10)

	return []truncateCase{
		{
			name:  "empty string is unchanged",
			in:    "",
			limit: 20,
			want:  "",
			cut:   false,
		},
		{
			name:  "ascii under the cap is unchanged",
			in:    "hello",
			limit: 64,
			want:  "hello",
			cut:   false,
		},
		{
			name:  "ascii exactly at the cap is unchanged",
			in:    "abcdefghij0123456789AB",
			limit: 22,
			want:  "abcdefghij0123456789AB",
			cut:   false,
		},
		{
			name:  "ascii one byte over the cap",
			in:    strings.Repeat("a", 23),
			limit: 22,
			want:  strings.Repeat("a", 22-len(m)) + m,
			cut:   true,
		},
		{
			name:  "value exactly the marker length fits unchanged",
			in:    strings.Repeat("x", len(m)),
			limit: len(m),
			want:  strings.Repeat("x", len(m)),
			cut:   false,
		},
		{
			name:  "two-byte rune is not split",
			in:    withRune("é"), // é at bytes 10-11; budget 11 would split it
			limit: len(m) + 11,
			want:  prefix10 + m,
			cut:   true,
		},
		{
			name:  "three-byte rune is not split",
			in:    withRune("€"), // € at bytes 10-12; budget 12 would split it
			limit: len(m) + 12,
			want:  prefix10 + m,
			cut:   true,
		},
		{
			name:  "four-byte rune is not split",
			in:    withRune("😀"), // U+1F600 at bytes 10-13; budget 13 would split it
			limit: len(m) + 13,
			want:  prefix10 + m,
			cut:   true,
		},
		{
			// budget 1 and the first rune is four bytes wide, so the walk-back
			// reaches zero and only the marker survives (A0-7.4).
			name:  "first rune wider than the budget",
			in:    "\U0001F600" + strings.Repeat("b", 20),
			limit: len(m) + 1,
			want:  m,
			cut:   true,
		},
		{
			// budget 1 with an ASCII first byte: no walk-back, one prefix byte
			// plus the marker fills the limit exactly.
			name:  "ascii one byte over the marker budget",
			in:    strings.Repeat("c", 30),
			limit: len(m) + 1,
			want:  "c" + m,
			cut:   true,
		},
		{
			name:  "cut lands on a two-byte rune start",
			in:    strings.Repeat("é", 10), // 20 B, 10 runes
			limit: 15,
			want:  "éé" + m,
			cut:   true,
		},
		{
			name:  "only multi-byte runes, exact fill",
			in:    strings.Repeat("中", 10), // 30 B, 10 runes
			limit: 20,
			want:  "中中中" + m,
			cut:   true,
		},
		{
			name:  "multi-byte value, limit is the marker length",
			in:    strings.Repeat("é", 8), // 16 B, longer than the marker
			limit: len(m),
			want:  m,
			cut:   true,
		},
		{
			name:  "ascii value, limit is the marker length",
			in:    "far too long for this cap",
			limit: len(m),
			want:  m,
			cut:   true,
		},
	}
}

// defectLimits are the platform-defect inputs of A0-7.5: any limit below
// len(TruncationMarker) means Truncate returns ("", true) so the caller can
// surface errs.Internal.
func defectLimits() []int {
	return []int{0, 1, len(TruncationMarker) - 1, -1, -4096}
}

// TestTruncateRuneBoundary guards A0-7.4 (never split a rune) and A0-7.5
// (mechanism T, marker counted against the cap, cap-below-marker is a platform
// defect).
func TestTruncateRuneBoundary(t *testing.T) {
	m := TruncationMarker

	for _, tc := range truncateCorpus() {
		t.Run(tc.name, func(t *testing.T) {
			out, truncated := Truncate(tc.in, tc.limit)
			if out != tc.want {
				t.Errorf("Truncate(%q, %d) out = %q, want %q", tc.in, tc.limit, out, tc.want)
			}
			if truncated != tc.cut {
				t.Errorf("Truncate(%q, %d) truncated = %v, want %v", tc.in, tc.limit, truncated, tc.cut)
			}
			if len(out) > tc.limit {
				t.Errorf("len(out) = %d exceeds limit %d (A0-7.5)", len(out), tc.limit)
			}
			if !utf8.ValidString(out) {
				t.Errorf("out = %q is invalid UTF-8 (A0-7.4/A0-2.3)", out)
			}
			// No partial rune survived: the prefix is a whole-rune prefix of in.
			prefix := strings.TrimSuffix(out, m)
			if !strings.HasPrefix(tc.in, prefix) {
				t.Errorf("prefix %q is not a prefix of input %q", prefix, tc.in)
			}
			if rest := tc.in[len(prefix):]; rest != "" && !utf8.RuneStart(rest[0]) {
				t.Errorf("cut split a rune: remainder %q begins with a continuation byte", rest)
			}
			if truncated && len(prefix) > tc.limit-len(m) {
				t.Errorf("prefix is %d bytes, more than the %d bytes left by the marker", len(prefix), tc.limit-len(m))
			}
			// truncated is true exactly when the output differs from the input.
			if truncated != (out != tc.in) {
				t.Errorf("truncated = %v but out != in is %v", truncated, out != tc.in)
			}
			if truncated && !strings.HasSuffix(out, m) {
				t.Errorf("truncated output %q lacks the marker (A0-7.5)", out)
			}
		})
	}

	for _, limit := range defectLimits() {
		for _, in := range []struct {
			kind string
			s    string
		}{{"empty", ""}, {"ascii", "a value"}, {"multibyte", strings.Repeat("é", 4)}} {
			t.Run(fmt.Sprintf("limit_below_marker/%d/%s", limit, in.kind), func(t *testing.T) {
				out, truncated := Truncate(in.s, limit)
				if out != "" || !truncated {
					t.Errorf("Truncate(%q, %d) = (%q, %v), want (\"\", true) — a limit below len(TruncationMarker)=%d is a platform defect (A0-7.5)",
						in.s, limit, out, truncated, len(m))
				}
			})
		}
	}
}

// byteCap names one A0-7.1 registry entry that caps a byte length.
type byteCap struct {
	name  string
	limit int
}

// byteCaps lists every A0-7.1 byte cap, to prove each is at least
// len(TruncationMarker) so mechanism T is well-defined for it. This is NOT a
// per-field mechanism registry: which mechanism a field gets is the owning
// contract's assignment (the A0-7.1 Mechanism column, A0-7.7), most of these
// are mechanism R, and doc.go records the lookup map as deliberately absent.
func byteCaps() []byteCap {
	return []byteCap{
		{"NodeSummaryMaxBytes", NodeSummaryMaxBytes},
		{"FindingSummaryMaxBytes", FindingSummaryMaxBytes},
		{"StageSummaryMaxBytes", StageSummaryMaxBytes},
		{"ProseLongMaxBytes", ProseLongMaxBytes},
		{"ProseMediumMaxBytes", ProseMediumMaxBytes},
		{"TargetMaxBytes", TargetMaxBytes},
		{"LabelMaxBytes", LabelMaxBytes},
		{"ToolVersionMaxBytes", ToolVersionMaxBytes},
		{"KindNameMaxBytes", KindNameMaxBytes},
		{"DigestMaxBytes", DigestMaxBytes},
		{"NodeLabelMaxBytes", NodeLabelMaxBytes},
		{"HypothesisClaimMaxBytes", HypothesisClaimMaxBytes},
		{"HypothesisBasisMaxBytes", HypothesisBasisMaxBytes},
		{"AttrValueMaxBytes", AttrValueMaxBytes},
		{"AddressMaxBytes", AddressMaxBytes},
	}
}

// TestNoSilentTruncationInHashedRecords is the negative half of A0-7.4/7.5: a
// shortened value must always be distinguishable from one that was not,
// because a silently shortened value inside a hashed record (A0-2.15 digests,
// A1-4 canonical bytes) changes a digest with no machine-readable signal.
func TestNoSilentTruncationInHashedRecords(t *testing.T) {
	m := TruncationMarker

	t.Run("corpus", func(t *testing.T) {
		for _, tc := range truncateCorpus() {
			t.Run(tc.name, func(t *testing.T) {
				out, truncated := Truncate(tc.in, tc.limit)
				if truncated == (out == tc.in) {
					t.Errorf("Truncate(%q, %d) = (%q, %v): truncated must be true exactly when the value was shortened",
						tc.in, tc.limit, out, truncated)
				}
				if truncated != strings.HasSuffix(out, m) {
					t.Errorf("Truncate(%q, %d) out = %q, truncated = %v: marker and flag must agree (A0-7.5)",
						tc.in, tc.limit, out, truncated)
				}
			})
		}
	})

	for _, c := range byteCaps() {
		t.Run(c.name, func(t *testing.T) {
			if c.limit < len(m) {
				t.Fatalf("%s = %d is below len(TruncationMarker) = %d (A0-7.5 platform defect)", c.name, c.limit, len(m))
			}

			at := strings.Repeat("a", c.limit)
			if out, truncated := Truncate(at, c.limit); truncated || out != at {
				t.Errorf("value at the cap was altered: out = %q, truncated = %v", out, truncated)
			}

			over := strings.Repeat("b", c.limit+1)
			out, truncated := Truncate(over, c.limit)
			if !truncated {
				t.Fatalf("%s: value of %d bytes was truncated to %q without reporting truncated=true (silent truncation, A0-7.5)",
					c.name, len(over), out)
			}
			if !strings.HasSuffix(out, m) {
				t.Errorf("shortened %s value %q does not end in the marker", c.name, out)
			}
			if len(out) > c.limit {
				t.Errorf("shortened %s value is %d bytes, over the cap %d", c.name, len(out), c.limit)
			}
			if want := over[:c.limit-len(m)] + m; out != want {
				t.Errorf("shortened %s value = %q, want %q", c.name, out, want)
			}
		})
	}
}

// TestFitsMeasuresUTF8BytesNotRunes pins the A0-7.3 measurement rule: a field
// cap counts the UTF-8 bytes of the decoded value (len(s) in Go) — never a
// rune count, never a UTF-16 length.
func TestFitsMeasuresUTF8BytesNotRunes(t *testing.T) {
	rows := []struct {
		name  string
		s     string
		limit int
		want  bool
		// runesOver reports a row where a rune-count check would disagree with
		// the byte rule, proving the rule is not re-derived from runes.
		runesOver bool
		// decodedOnly reports a row whose JSON-encoded form is longer than the
		// limit while the decoded value fits: A0-7.3 counts the decoded value,
		// quotes and escapes do not count.
		decodedOnly bool
	}{
		// ASCII: byte count == rune count, so the two rules agree.
		{name: "ascii exactly at the cap", s: strings.Repeat("a", 8), limit: 8, want: true},
		{name: "ascii one byte over the cap", s: strings.Repeat("a", 9), limit: 8, want: false},

		// Multi-byte: under the cap in runes, over it in bytes.
		{name: "two-byte runes under in runes over in bytes", s: strings.Repeat("é", 5), limit: 8, want: false, runesOver: true},
		{name: "three-byte runes under in runes over in bytes", s: strings.Repeat("中", 3), limit: 8, want: false, runesOver: true},
		{name: "four-byte runes under in runes over in bytes", s: strings.Repeat("😀", 3), limit: 8, want: false, runesOver: true},
		{name: "four-byte runes exactly at the cap", s: strings.Repeat("😀", 2), limit: 8, want: true},

		// Decoded value, not the JSON form (A0-7.3: quotes and escapes do not count).
		{name: "decoded quote counts unescaped", s: "a\"b", limit: 3, want: true, decodedOnly: true},
		{name: "decoded backslash counts unescaped", s: `a\b` + strings.Repeat("c", 5), limit: 8, want: true, decodedOnly: true},

		// Degenerate caps.
		{name: "empty string at a zero cap", s: "", limit: 0, want: true},
		{name: "non-empty at a zero cap", s: "x", limit: 0, want: false},
		{name: "empty string against a negative cap", s: "", limit: -1, want: false},
		{name: "negative cap rejects a value", s: "x", limit: -8, want: false},
	}

	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			if got := Fits(r.s, r.limit); got != r.want {
				t.Errorf("Fits(%q, %d) = %v, want %v", r.s, r.limit, got, r.want)
			}
			// A0-7.3: the rule is len(s) in Go, for every input.
			if Fits(r.s, r.limit) != (len(r.s) <= r.limit) {
				t.Errorf("Fits(%q, %d) disagrees with len(s) <= limit", r.s, r.limit)
			}
			if r.decodedOnly {
				if enc := strconv.Quote(r.s); len(enc) <= r.limit {
					t.Fatalf("test row is not a valid discriminator: encoded form %q is %d bytes at limit %d",
						enc, len(enc), r.limit)
				}
				if !Fits(r.s, r.limit) {
					t.Errorf("the decoded value %q (%d B) must fit limit %d; escapes and quotes do not count (A0-7.3)",
						r.s, len(r.s), r.limit)
				}
			}
			if r.runesOver {
				if utf8.RuneCountInString(r.s) > r.limit {
					t.Fatalf("test row is not a valid discriminator: rune count %d exceeds limit %d",
						utf8.RuneCountInString(r.s), r.limit)
				}
				if Fits(r.s, r.limit) {
					t.Errorf("a rune-count check would have accepted %q at limit %d; the byte rule must reject it",
						r.s, r.limit)
				}
			}
		})
	}
}

// TestTruncationMarkerIsCountedAgainstTheCap guards A0-7.5's marker contract:
// the literal text, its length derived from that literal rather than a number
// (A0-7.5 reads 11 B since the erratum of A0 §6 item 16), that it is pure
// ASCII, and that it is paid for out of the cap rather than appended on top.
func TestTruncationMarkerIsCountedAgainstTheCap(t *testing.T) {
	const literal = "[truncated]"

	if TruncationMarker != literal {
		t.Errorf("TruncationMarker = %q, want %q (A0-7.5)", TruncationMarker, literal)
	}
	if len(TruncationMarker) != len(literal) {
		t.Errorf("len(TruncationMarker) = %d, want %d", len(TruncationMarker), len(literal))
	}
	if !allASCII(TruncationMarker) {
		t.Errorf("TruncationMarker = %q is not pure ASCII (A0-7.5)", TruncationMarker)
	}
	if utf8.RuneCountInString(TruncationMarker) != len(TruncationMarker) {
		t.Error("TruncationMarker has multi-byte runes; the cap counts bytes")
	}

	t.Run("limit_equal_to_marker_only_the_marker_survives", func(t *testing.T) {
		out, truncated := Truncate("0123456789abcdefghij", len(TruncationMarker))
		if out != TruncationMarker || !truncated {
			t.Errorf("Truncate over-long value with limit %d = (%q, %v), want (%q, true)",
				len(TruncationMarker), out, truncated, TruncationMarker)
		}
		if len(out) != len(TruncationMarker) {
			t.Errorf("len(out) = %d, want %d (the marker is the whole budget)", len(out), len(TruncationMarker))
		}
	})

	t.Run("prefix_pays_for_the_marker", func(t *testing.T) {
		limit := len(TruncationMarker) + 8
		in := strings.Repeat("z", limit+1)
		out, truncated := Truncate(in, limit)
		if !truncated {
			t.Fatalf("Truncate(%q, %d) reported truncated = false", in, limit)
		}
		if len(out) != limit {
			t.Errorf("len(out) = %d, want %d", len(out), limit)
		}
		if want := strings.Repeat("z", 8) + TruncationMarker; out != want {
			t.Errorf("out = %q, want %q (8 prefix bytes + marker)", out, want)
		}
		if !strings.HasPrefix(in, strings.TrimSuffix(out, TruncationMarker)) {
			t.Errorf("out %q is not in + marker truncated on a boundary", out)
		}
	})
}
