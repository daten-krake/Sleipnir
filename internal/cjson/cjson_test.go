package cjson

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/daten-krake/sleipnir/internal/errs"
)

// The A0-2.17 vectors' canonical-bytes column uses angle-bracket placeholders
// for code points that are invisible in a terminal (<2028>, <2029>, <7F>); the
// test writes them as Go escapes instead, so this file contains no invisible
// bytes. Everything else in that column is literal text: a raw string keeps
// \", \\, \n, \t and \u00xx as the two- and six-character escapes A0-2.7
// requires in the output.

// mustCanonical canonicalizes doc or fails the test.
func mustCanonical(t *testing.T, doc []byte, exclude ...string) []byte {
	t.Helper()
	out, err := Canonical(doc, exclude...)
	if err != nil {
		t.Fatalf("Canonical(%q) = error %v, want bytes", doc, err)
	}
	return out
}

// nest returns a document whose innermost container sits at the given depth:
// the top-level object is level 1 (A0-2.11).
func nest(depth int) []byte {
	return []byte(strings.Repeat(`{"a":`, depth) + "1" + strings.Repeat("}", depth))
}

// padded returns a document of exactly n bytes.
func padded(t *testing.T, n int) []byte {
	t.Helper()
	doc := []byte(`{"a":"` + strings.Repeat("x", n-len(`{"a":""}`)) + `"}`)
	if len(doc) != n {
		t.Fatalf("fixture is %d bytes, want %d", len(doc), n)
	}
	return doc
}

// A0-2.17: the eight normative vectors, asserted byte-exactly against the
// canonical-bytes, len and SHA-256 columns.
func TestVectorsV1ToV8ByteExact(t *testing.T) {
	tests := []struct {
		name      string
		doc       string
		exclude   []string
		canonical string
		length    int
		digest    string
	}{
		{
			name:      "V1",
			doc:       `{"b": 1, "a": 2}`,
			canonical: `{"a":2,"b":1}`,
			length:    13,
			digest:    "d3626ac30a87e6f7a6428233b3c68299976865fa5508e4267c5415c76af7a772",
		},
		{
			name:      "V2",
			doc:       `{"a":1,"B":2,"_c":3,"0":4}`,
			canonical: `{"0":4,"B":2,"_c":3,"a":1}`,
			length:    26,
			digest:    "735f90d32fc3437e6c52f922158bc966ab66a8ba875ccc60693c3c93d513935a",
		},
		{
			name:      "V3",
			doc:       `{"s":"a\"b\\c\nd\te\u0001f<>&\u00e9\u4e2d\ud83d\ude00","n":-42,"t":true}`,
			canonical: `{"n":-42,"s":"a\"b\\c\nd\te\u0001f<>&é中😀","t":true}`,
			length:    57,
			digest:    "63fc2ef866b5542bfcd5d8c943f71a92b1b6a8a2403704f0416822a18aefcd6c",
		},
		{
			name:      "V4",
			doc:       `{"arr":[3,1,2,{"y":1,"x":2}],"obj":{},"e":[]}`,
			canonical: `{"arr":[3,1,2,{"x":2,"y":1}],"e":[],"obj":{}}`,
			length:    45,
			digest:    "d66e67ee4f3bef3250a4b86aa3ea680d7c9a5545424cf4316a9cf917e39db521",
		},
		{
			name: "V5",
			doc: `{"engagement_id":"eng_01m1y2whfhgbz06ays6dxnvyws","kind":"tool_invoked",` +
				`"recorded_at":"2026-09-07T14:03:22.481Z","seq":4711,"prev_hash":"9b74c9897bac770ffc029102a200c5de",` +
				`"hash":"03c8a7d2b9e4f1a6d5c0b8e7f2a1d4c3b6a9e8f7d0c1b2a3e4f5061728394a5b"}`,
			exclude:   []string{"seq", "prev_hash", "hash"},
			canonical: `{"engagement_id":"eng_01m1y2whfhgbz06ays6dxnvyws","kind":"tool_invoked","recorded_at":"2026-09-07T14:03:22.481Z"}`,
			length:    113,
			digest:    "3695e846ce9e48d6577c0a7ef48ace7b9d34e4ca498dc91aac4309b40c4e10cb",
		},
		{
			// A0-2.4: UTF-8 byte order puts U+FFFD (EF BF BD) before U+1F600,
			// the reverse of RFC 8785's UTF-16 code-unit order. The input's
			// \uFFFD is the six characters the contract's input column shows;
			// the output key is the three bytes EF BF BD.
			name:      "V6",
			doc:       `{"😀":1,"\uFFFD":2}`,
			canonical: "{\"\uFFFD\":2,\"😀\":1}",
			length:    18,
			digest:    "9fbfff35f05fb9c72de19e392d0a1b848acb4d6c42489709d709982d10dd883c",
		},
		{
			name:      "V7",
			doc:       `{"s":"a\u2028b\u2029c\u007fd"}`,
			canonical: "{\"s\":\"a\u2028b\u2029c\u007fd\"}",
			length:    19,
			digest:    "aedd6df88cc462fdbdc5788549d753c9b8c21ac8b9e16c51c72016cf564c3e85",
		},
		{
			name:      "V8",
			doc:       `{"s":"<a href=\"x\">&é\u2028"}`,
			canonical: "{\"s\":\"<a href=\\\"x\\\">&é\u2028\"}",
			length:    28,
			digest:    "63995ca86de5cce6f4d74df8e1d90aed78918aad912cc7c7feaff4d89fb15b2d",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mustCanonical(t, []byte(tt.doc), tt.exclude...)
			if !bytes.Equal(got, []byte(tt.canonical)) {
				t.Errorf("Canonical bytes =\n% x\n%q\nwant\n% x\n%q", got, got, []byte(tt.canonical), tt.canonical)
			}
			if len(got) != tt.length {
				t.Errorf("len = %d, want %d (the len column is the authority)", len(got), tt.length)
			}
			if d := SHA256Hex(got); d != tt.digest {
				t.Errorf("SHA-256 = %s, want %s", d, tt.digest)
			}
		})
	}
}

// A0-2.3: encoding/json substitutes U+FFFD for invalid UTF-8 instead of
// failing, so utf8.Valid on the raw bytes must run first.
func TestInvalidUTF8Rejected(t *testing.T) {
	tests := []struct {
		name string
		doc  []byte
	}{
		{"two-byte 0xff inside a string", []byte("{\"a\":\"\xff\"}")},
		{"truncated three-byte sequence", []byte("{\"a\":\"\xe4\xb8\")}")},
		{"overlong encoding of U+002F", []byte("{\"a\":\"\xc0\xaf\"}")},
		{"stray byte after the value", append([]byte(`{"a":1}`), 0xff)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if utf8.Valid(tt.doc) {
				t.Fatal("fixture is valid UTF-8, so it proves nothing")
			}
			got, err := Canonical(tt.doc)
			if err == nil {
				t.Fatalf("Canonical(%q) = %q, want a validation error", tt.doc, got)
			}
			if errs.KindOf(err) != errs.Validation {
				t.Errorf("kind = %v, want validation", errs.KindOf(err))
			}
			if got != nil {
				t.Errorf("returned %q alongside the error, want no bytes", got)
			}
		})
	}
}

// A0-2.3 (b): the decoder accepts a lone surrogate escape and hands back
// U+FFFD, so the scan must run over the raw input bytes.
func TestLoneSurrogateRejected(t *testing.T) {
	tests := []struct {
		name string
		doc  string
	}{
		{"high surrogate alone", `{"a":"\ud800"}`},
		{"upper-case escape", `{"a":"\uD83D"}`},
		{"high surrogate then an ordinary char", `{"a":"\ud800x"}`},
		{"high surrogate then an escaped backslash", `{"a":"\ud83d\\x"}`},
		{"high surrogate then a surrogate-pair-looking plain text", `{"a":"\ud83dude00"}`},
		{"low surrogate without a high one", `{"a":"\udc00"}`},
		{"two high surrogates", `{"a":"\ud800\ud800"}`},
		{"lone surrogate in a key", `{"\ud800":1}`},
		{"high surrogate then a BMP escape", `{"a":"\ud800\u0041"}`},
		{"high surrogate then text that looks like a low escape", `{"a":"\ud800xxdc00"}`},
		{"high surrogate then bare hex digits, no escape", `{"a":"\ud800dc00"}`},
		{"high surrogate then a non-surrogate escape", `{"a":"\ud83d\ufdff"}`},
		{"truncated escape at end of document", `{"a":"\ud`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Canonical([]byte(tt.doc))
			if err == nil {
				t.Fatalf("Canonical(%q) = %q, want a validation error", tt.doc, got)
			}
			if errs.KindOf(err) != errs.Validation {
				t.Errorf("kind = %v, want validation", errs.KindOf(err))
			}
			if got != nil {
				t.Errorf("returned %q alongside the error, want no bytes", got)
			}
		})
	}

	t.Run("well-formed pair is accepted and emitted literally", func(t *testing.T) {
		got := mustCanonical(t, []byte(`{"a":"\ud83d\ude00"}`))
		if want := `{"a":"😀"}`; !bytes.Equal(got, []byte(want)) {
			t.Errorf("Canonical = %q, want %q", got, want)
		}
	})

	t.Run("a literal backslash before u is not an escape", func(t *testing.T) {
		// The text \uD800 in a string preceded by an escaped backslash is
		// ordinary characters and must not be read as a surrogate escape.
		doc := []byte(`{"a":"\\uD800"}`)
		got := mustCanonical(t, doc)
		if !bytes.Equal(got, doc) {
			t.Errorf("Canonical = %q, want %q", got, doc)
		}
	})
}

// A0-2.5: an exact duplicate key is rejected; decoding into a map would
// silently keep the last value and produce a different digest.
func TestDuplicateKeyRejected(t *testing.T) {
	tests := []string{
		`{"a":1,"a":2}`,
		`{"a":1,"a":1}`,
		`{"\u0061":1,"a":2}`,
		`{"o":{"b":1,"b":2},"a":0}`,
		`{"arr":[{"x":1,"x":2}]}`,
		`{"":1,"":2}`,
		`{"a":{"deep":{"z":1,"Z":2}},"b":0}`,
	}
	for _, doc := range tests {
		t.Run(doc, func(t *testing.T) {
			_, err := Canonical([]byte(doc))
			if err == nil {
				t.Fatalf("Canonical(%q) succeeded, want a validation error", doc)
			}
			if errs.KindOf(err) != errs.Validation {
				t.Errorf("kind = %v, want validation", errs.KindOf(err))
			}
			if !strings.Contains(err.Error(), "duplicate key") {
				t.Errorf("error %q does not name the rule", err)
			}
		})
	}
}

// A0-2.5: two keys differing only by case are duplicates. The Token() walk
// observes both keys; only map or struct decoding would let one win.
func TestCaseDuplicateKeyRejected(t *testing.T) {
	tests := []string{
		`{"a":1,"A":2}`,
		`{"A":1,"a":2}`,
		`{"é":1,"É":2}`,
		`{"outer":{"k":1,"K":2}}`,
		`{"ToolName":1,"toolname":2,"toolNAME":3}`,
		// The pairs strings.ToLower misses but unicode.SimpleFold (and
		// strings.EqualFold, and encoding/json's field matching) catches.
		// ("\u0130" is not one: U+0130 has no simple fold — only the full
		// expansion to i plus combining dot — that is the ceiling.)
		`{"s":1,"ſ":2}`,
		`{"k":1,"\u212a":2}`,
		`{"µ":1,"Μ":2}`,
		`{"Μ":1,"μ":2}`,
	}
	for _, doc := range tests {
		t.Run(doc, func(t *testing.T) {
			_, err := Canonical([]byte(doc))
			if err == nil {
				t.Fatalf("Canonical(%q) succeeded, want a validation error", doc)
			}
			if errs.KindOf(err) != errs.Validation {
				t.Errorf("kind = %v, want validation", errs.KindOf(err))
			}
		})
	}

	t.Run("keys that merely share a prefix are not duplicates", func(t *testing.T) {
		got := mustCanonical(t, []byte(`{"a":1,"ab":2,"B":3}`))
		if want := `{"B":3,"a":1,"ab":2}`; !bytes.Equal(got, []byte(want)) {
			t.Errorf("Canonical = %q, want %q", got, want)
		}
	})

	t.Run("the fold is exactly strings.EqualFold's equivalence", func(t *testing.T) {
		// A0-2.5 adopts encoding/json's field-matching equivalence, so
		// foldKey(a) == foldKey(b) must hold exactly when
		// strings.EqualFold(a, b) does — including at the documented
		// simple-folding ceiling, where neither equates a pair (the
		// CaseFolding.txt full foldings strings.EqualFold does not do
		// either). want is that shared verdict, asserted both ways.
		tests := []struct {
			name     string
			a, b     string
			wantSame bool
		}{
			{"long s folds with s", "s", "\u017f", true},
			{"kelvin sign folds with k", "k", "\u212a", true},
			{"micro sign folds with greek capital mu", "\u00b5", "\u039c", true},
			{"greek capital mu folds with greek small mu", "\u039c", "\u03bc", true},
			{"ascii case pair across a name", "ToolName", "toolname", true},
			{"an already-folded key", "hash", "hash", true},
			{"a prefix is not a fold", "hash", "hash_x", false},
			{"ceiling: sharp s does not fold with ss", "\u00df", "ss", false},
			{"ceiling: fi ligature does not fold with fi", "\ufb01", "fi", false},
			{"ceiling: dotted capital i has no simple fold of i", "i", "\u0130", false},
			{"ceiling: dotted capital i does not fold with i plus combining dot", "\u0130", "i\u0307", false},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := strings.EqualFold(tt.a, tt.b); got != tt.wantSame {
					t.Fatalf("strings.EqualFold(%q, %q) = %v, want %v (fixture)", tt.a, tt.b, got, tt.wantSame)
				}
				if got := foldKey(tt.a) == foldKey(tt.b); got != tt.wantSame {
					t.Errorf("foldKey(%q) == foldKey(%q): got %v, want %v", tt.a, tt.b, got, tt.wantSame)
				}
			})
		}
	})

	t.Run("the fast path is the A0-8.1 normal path", func(t *testing.T) {
		for _, k := range []string{"a", "engagement_id", "hash", "prev_hash", "seq", "0", "_c7"} {
			if got := foldKey(k); got != k {
				t.Errorf("foldKey(%q) = %q, want it unchanged", k, got)
			}
		}
		// The linear-not-quadratic ruling in doc.go is only true because
		// contract keys (A0-8.1: ^[a-z][a-z0-9_]{0,39}$) never build a
		// folded copy at all. AllocsPerRun is deterministic, not timing.
		var sink string
		if n := testing.AllocsPerRun(50, func() { sink = foldKey("engagement_id") }); n != 0 {
			t.Errorf("foldKey on an A0-8.1 key allocated %v times per run, want 0", n)
		}
		if sink != "engagement_id" {
			t.Errorf("the allocating run produced %q", sink)
		}
	})
}

// A0-2.5, A0-2.6: an accepted integer's literal characters are reproduced
// verbatim.
func TestCanonicalEmitsNumberLiteralText(t *testing.T) {
	tests := []struct {
		doc  string
		want string
	}{
		{`{"a":0}`, `{"a":0}`},
		{`{"a":-42}`, `{"a":-42}`},
		{`{"a":9007199254740991}`, `{"a":9007199254740991}`},
		{`{"a":-9007199254740991}`, `{"a":-9007199254740991}`},
		{`{"a":  7  }`, `{"a":7}`},
		{`{"a":1000000000000000}`, `{"a":1000000000000000}`},
		{`{"a":1e3,"b":2}`, ""},
		{`{"a":-0}`, ""},
		{`{"a":9007199254740993}`, ""},
	}
	for _, tt := range tests {
		t.Run(tt.doc, func(t *testing.T) {
			got, err := Canonical([]byte(tt.doc))
			if tt.want == "" {
				if errs.KindOf(err) != errs.Validation {
					t.Fatalf("kind = %v (%v), want validation", errs.KindOf(err), err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Canonical(%q) = error %v, want bytes", tt.doc, err)
			}
			if !bytes.Equal(got, []byte(tt.want)) {
				t.Errorf("Canonical = %q, want %q", got, tt.want)
			}
		})
	}
}

// A0-2.17's rejection list, one subtest per entry, plus the accept list. Every
// entry must be validation with no bytes returned: a decoder failure is never
// a platform defect (A0-2.5).
func TestRejections(t *testing.T) {
	tests := []struct {
		name string
		doc  []byte
	}{
		{"duplicate", []byte(`{"a":1,"a":2}`)},
		{"case-duplicate", []byte(`{"a":1,"A":2}`)},
		{"fraction", []byte(`{"a":1.0}`)},
		{"lowercase exponent", []byte(`{"a":1e3}`)},
		{"uppercase exponent", []byte(`{"a":1E3}`)},
		{"negative zero", []byte(`{"a":-0}`)},
		{"negative zero float", []byte(`{"a":-0.0}`)},
		{"leading zero", []byte(`{"a":01}`)},
		{"2^53+1 matches the regex but is out of range", []byte(`{"a":9007199254740993}`)},
		{"20 digits fails the regex", []byte(`{"a":10000000000000000000}`)},
		{"lone surrogate", []byte(`{"a":"\ud800"}`)},
		{"invalid utf-8", []byte("{\"a\":\"\xff\"}")},
		{"null", []byte(`{"a":null}`)},
		{"trailing", []byte(`{"a":1},`)},
		{"top-level array", []byte(`[1,2]`)},
		{"bom-prefixed", append([]byte{0xEF, 0xBB, 0xBF}, []byte(`{"a":1}`)...)},
		{"33-deep nesting", nest(33)},
		{"40-deep nesting", nest(40)},
		{"over 1 MiB", padded(t, MaxBytes+1)},
		{"NaN", []byte(`{"a":NaN}`)},
		{"raw control byte in a string", []byte("{\"a\":\"a\x01b\"}")},
		{"raw newline in a string", []byte("{\"a\":\"a\nb\"}")},
		{"Infinity", []byte(`{"a":Infinity}`)},
		{"plus sign", []byte(`{"a":+1}`)},
		{"empty input", []byte("")},
		{"top-level string", []byte(`"a"`)},
		{"top-level number", []byte(`1`)},
		{"top-level true", []byte(`true`)},
		{"top-level null", []byte(`null`)},
		{"trailing second document", []byte(`{"a":1}{"b":2}`)},
		{"trailing close brace", []byte(`{"a":1}}`)},
		{"null in a nested array", []byte(`{"a":[1,null,3]}`)},
		{"null nested in an object", []byte(`{"a":{"b":{"c":null}}}`)},
		{"unterminated object", []byte(`{"a":1`)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Canonical(tt.doc)
			if err == nil {
				t.Fatalf("Canonical = %q, want a validation error", got)
			}
			if errs.KindOf(err) != errs.Validation {
				t.Errorf("kind = %v, want validation: no decoder failure may leak as internal (%v)",
					errs.KindOf(err), err)
			}
			if got != nil {
				t.Errorf("returned %q alongside the error, want no bytes", got)
			}
		})
	}

	t.Run("malformed streams are validation, never a panic", func(t *testing.T) {
		docs := []string{
			`{`, `}`, `]`, `[,2]`, `{"a"`, `{"a":`, `{"a":[}`, `{"a":[1,]}`,
			`{"a":1,,}`, `{"a":1"b":2}`, `{"a"1}`, `{"a":tru}`, `{"a"true}`, `{"a" 1}`,
			`{"a":}` + strings.Repeat(",1", 200), `{"\u` + strings.Repeat("0", 4096) + `}`,
			strings.Repeat("[", 200), strings.Repeat(`{"a":`, 200),
		}
		for _, doc := range docs {
			got, err := Canonical([]byte(doc))
			if err == nil {
				t.Errorf("Canonical(%q) = %q, want a validation error", doc, got)
				continue
			}
			if errs.KindOf(err) != errs.Validation {
				t.Errorf("%q: kind = %v, want validation", doc, errs.KindOf(err))
			}
		}
	})

	t.Run("mutated documents never panic and never reach internal", func(t *testing.T) {
		// AGENTS.md high-review bar: untrusted-input parsing. A fixed seed keeps
		// this deterministic (DESIGN §8). Every failure must be validation, and
		// every accepted output must be a fixed point of Canonical, which is
		// what A0-2.16 asks the store to rely on.
		seeds := []string{
			`{"b": 1, "a": 2}`,
			`{"s":"a\"b\\c\nd\te\u0001f<>&\u00e9\u4e2d\ud83d\ude00","n":-42,"t":true}`,
			`{"arr":[3,1,2,{"y":1,"x":2}],"obj":{},"e":[]}`,
			`{"engagement_id":"eng_01","seq":4711,"hash":"03c8"}`,
			`{"a":{"b":{"c":[1,2,{"d":null}]}}}`,
			`{"\u00e9":"\ud83d\ude00","x":[[[]]]}`,
		}
		rng := rand.New(rand.NewSource(1))
		for i := 0; i < 3000; i++ {
			doc := []byte(seeds[rng.Intn(len(seeds))])
			for m := rng.Intn(4) + 1; m > 0; m-- {
				switch rng.Intn(3) {
				case 0: // corrupt one byte
					doc[rng.Intn(len(doc))] = byte(rng.Intn(256))
				case 1: // append text
					doc = append(doc, seeds[rng.Intn(len(seeds))][:3]...)
				case 2: // truncate, keeping at least one byte
					doc = doc[:rng.Intn(len(doc))+1]
				}
			}
			got, err := Canonical(doc)
			if err != nil {
				if errs.KindOf(err) != errs.Validation {
					t.Fatalf("mutated %q: kind = %v, want validation (no input may reach internal): %v", doc, errs.KindOf(err), err)
				}
				if got != nil {
					t.Fatalf("mutated %q returned %q with an error", doc, got)
				}
				continue
			}
			again, err := Canonical(got)
			if err != nil {
				t.Fatalf("canonical bytes %q do not re-canonicalize: %v", got, err)
			}
			if !bytes.Equal(again, got) || SHA256Hex(again) != SHA256Hex(got) {
				t.Fatalf("%q is not a fixed point: %q", doc, again)
			}
		}
	})

	t.Run("a huge number token stays bounded in the message", func(t *testing.T) {
		huge := "1" + strings.Repeat("0", 60000)
		_, err := Canonical([]byte(`{"a":` + huge + `}`))
		if errs.KindOf(err) != errs.Validation {
			t.Fatalf("kind = %v (%v), want validation", errs.KindOf(err), err)
		}
		if strings.Contains(err.Error(), huge) || len(err.Error()) > 400 {
			t.Errorf("message is %d bytes and echoes the token: %q", len(err.Error()), err)
		}
	})

	t.Run("an error names the rule and the offset, never the document", func(t *testing.T) {
		// A0-3.5: a message MUST NOT contain a request body. The input here is
		// a rejected document carrying a long secret-like value.
		secret := "hunter2-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		doc := []byte(`{"a":"` + secret + `","a":"x"}`)
		_, err := Canonical(doc)
		if errs.KindOf(err) != errs.Validation {
			t.Fatalf("kind = %v (%v), want validation", errs.KindOf(err), err)
		}
		if strings.Contains(err.Error(), secret) {
			t.Errorf("the error echoes the document: %v", err)
		}
		if !strings.Contains(err.Error(), "A0-2.5") {
			t.Errorf("the error does not name the rule it enforces: %v", err)
		}
	})

	t.Run("a long rejected token stays out of the message", func(t *testing.T) {
		// A0-3.5: no request body in a message. The offending key here is far
		// longer than the bound the error echoes.
		long := strings.Repeat("k", 4096)
		_, err := Canonical([]byte(`{"` + long + `":1,"` + long + `":2}`))
		if errs.KindOf(err) != errs.Validation {
			t.Fatalf("kind = %v (%v), want validation", errs.KindOf(err), err)
		}
		if strings.Contains(err.Error(), long) {
			t.Errorf("the error echoes the whole key: %d bytes of message", len(err.Error()))
		}
		if n := len(err.Error()); n > 400 {
			t.Errorf("message is %d bytes, want a bounded one: %q", n, err)
		}
	})

	t.Run("accepts the A0-2.17 accept list", func(t *testing.T) {
		accepts := []struct {
			name string
			doc  []byte
			want string
		}{
			{"32-deep document", nest(32), string(nest(32))},
			{"well-formed surrogate pair", []byte(`{"a":"\ud83d\ude00"}`), `{"a":"😀"}`},
			{"2^53-1 boundary", []byte(`{"a":9007199254740991}`), `{"a":9007199254740991}`},
			{"trailing whitespace", []byte("{\"a\":1}\n "), `{"a":1}`},
			{"whitespace around members", []byte("{ \"a\" : 1 , \"b\" : [ 1 , 2 ] }"), `{"a":1,"b":[1,2]}`},
		}
		for _, tt := range accepts {
			t.Run(tt.name, func(t *testing.T) {
				got := mustCanonical(t, tt.doc)
				if !bytes.Equal(got, []byte(tt.want)) {
					t.Errorf("Canonical = %q, want %q", got, tt.want)
				}
			})
		}
	})
}

// A0-2.11: the bound's effect is asserted at the boundary and one past it. The
// bound is never disabled to see what would happen.
func TestDepthLimit(t *testing.T) {
	for _, depth := range []int{1, 2, 31, MaxDepth} {
		doc := nest(depth)
		got := mustCanonical(t, doc)
		if !bytes.Equal(got, doc) {
			t.Errorf("depth %d: Canonical = %q, want the input unchanged", depth, got)
		}
	}
	for _, depth := range []int{MaxDepth + 1, MaxDepth + 8, 4096} {
		_, err := Canonical(nest(depth))
		if err == nil {
			t.Fatalf("a %d-deep document was accepted, want rejection", depth)
		}
		if errs.KindOf(err) != errs.Validation {
			t.Errorf("kind = %v, want validation", errs.KindOf(err))
		}
	}
	t.Run("arrays count as levels too", func(t *testing.T) {
		// A well-formed depth-33 document whose innermost container is an
		// array, so only the depth rule can reject it. The previous fixture
		// was delimiter-malformed — it also failed in close with the rule
		// removed, and so could not detect removal of the rule it names.
		// The bound itself is never disabled (AGENTS.md).
		deep := strings.Repeat(`{"a":`, MaxDepth) + "[1]" + strings.Repeat("}", MaxDepth)
		_, err := Canonical([]byte(deep))
		if err == nil {
			t.Fatal("a document with an array at depth 33 was accepted, want rejection")
		}
		if errs.KindOf(err) != errs.Validation {
			t.Errorf("kind = %v, want validation", errs.KindOf(err))
		}
		if !strings.Contains(err.Error(), "MaxDepth") {
			t.Errorf("the error does not name the bound: %v", err)
		}
		sibling := strings.Repeat(`{"a":`, MaxDepth-1) + "[1]" + strings.Repeat("}", MaxDepth-1)
		if got, err := Canonical([]byte(sibling)); err != nil || !bytes.Equal(got, []byte(sibling)) {
			t.Errorf("depth-32 array document: got %q (err %v), want it accepted unchanged", got, err)
		}
	})
	t.Run("a deep array inside the bound is accepted", func(t *testing.T) {
		doc := []byte(`{"a":` + strings.Repeat("[", MaxDepth-1) + "1" + strings.Repeat("]", MaxDepth-1) + `}`)
		if got, err := Canonical(doc); err != nil {
			t.Fatalf("Canonical(%q) = error %v, want bytes", doc, err)
		} else if !bytes.Equal(got, doc) {
			t.Errorf("Canonical = %q, want %q", got, doc)
		}
	})
}

// A0-2.11: size is len(doc) of the input, at most 1 MiB, and it is checked
// before the canonicalizer produces anything.
func TestSizeLimit(t *testing.T) {
	t.Run("exactly MaxBytes accepted", func(t *testing.T) {
		doc := padded(t, MaxBytes)
		got := mustCanonical(t, doc)
		if !bytes.Equal(got, doc) {
			t.Errorf("a document at the bound changed: got %d bytes, want %d", len(got), len(doc))
		}
	})
	t.Run("one byte past MaxBytes rejected", func(t *testing.T) {
		doc := padded(t, MaxBytes+1)
		got, err := Canonical(doc)
		if err == nil {
			t.Fatalf("a document of %d bytes was accepted, want rejection", len(doc))
		}
		if errs.KindOf(err) != errs.Validation {
			t.Errorf("kind = %v, want validation", errs.KindOf(err))
		}
		if got != nil {
			t.Errorf("canonicalizer produced %d bytes for an over-size document, want none", len(got))
		}
		if strings.Contains(err.Error(), "xxxx") {
			t.Errorf("the error echoes the document body (A0-3.5): %v", err)
		}
	})
	t.Run("CanonicalValue is bounded too", func(t *testing.T) {
		_, err := CanonicalValue(map[string]string{"a": strings.Repeat("x", MaxBytes)})
		if errs.KindOf(err) != errs.Validation {
			t.Errorf("kind = %v (%v), want validation", errs.KindOf(err), err)
		}
	})
}

// A0-2.14: encoding/json emits null for a nil slice, map or pointer, so a
// constructor that forgot to initialize a collection is caught here rather
// than silently changing every digest.
func TestNilCollectionNeverSerializesAsNull(t *testing.T) {
	type payload struct {
		Kind   string            `json:"kind"`
		Items  []string          `json:"items"`
		Labels map[string]string `json:"labels"`
		Ref    *string           `json:"ref"`
	}

	_, err := CanonicalValue(&payload{Kind: "tool_invoked"})
	if err == nil {
		t.Fatal("a struct with nil slice, map and pointer fields was canonicalized, want rejection")
	}
	if errs.KindOf(err) != errs.Validation {
		t.Errorf("kind = %v, want validation", errs.KindOf(err))
	}
	if !strings.Contains(err.Error(), "null") {
		t.Errorf("error %q does not name null as the violation", err)
	}

	ref := "evi_01m1y2whfhgbz06ays6dxnvyws"
	got, err := CanonicalValue(&payload{
		Kind:   "tool_invoked",
		Items:  []string{},
		Labels: map[string]string{},
		Ref:    &ref,
	})
	if err != nil {
		t.Fatalf("CanonicalValue with initialized collections = error %v, want bytes", err)
	}
	want := `{"items":[],"kind":"tool_invoked","labels":{},"ref":"evi_01m1y2whfhgbz06ays6dxnvyws"}`
	if string(got) != want {
		t.Errorf("CanonicalValue =\n%s\nwant\n%s", got, want)
	}

	for _, v := range []any{
		nil,
		map[string]any{"a": nil},
		&struct {
			Inner *payload `json:"inner"`
		}{},
		&struct {
			Score float64 `json:"score"`
		}{Score: 1.5},
	} {
		if _, err := CanonicalValue(v); errs.KindOf(err) != errs.Validation {
			t.Errorf("CanonicalValue(%T): kind = %v (%v), want validation", v, errs.KindOf(err), err)
		}
	}
}

// A0-2.15: a digest is 64 lowercase hex characters over the canonical bytes.
func TestSHA256HexLowercase64(t *testing.T) {
	input := []byte(`{"a":2,"b":1}`)
	got := SHA256Hex(input)
	if len(got) != 64 {
		t.Fatalf("len = %d, want 64 (%q)", len(got), got)
	}
	for _, c := range got {
		if !strings.ContainsRune("0123456789abcdef", c) {
			t.Fatalf("digest %q contains %q, which is not a lowercase hex digit", got, c)
		}
	}
	// The same digest computed independently in the test, without this package.
	sum := sha256.Sum256(input)
	if want := hex.EncodeToString(sum[:]); got != want {
		t.Errorf("SHA256Hex = %s, want %s", got, want)
	}
	if SHA256Hex(nil) != SHA256Hex([]byte{}) {
		t.Error("the empty input has two digests")
	}
}

// A0-2.15: the digest comparison used by gating decisions.
func TestDigestEqual(t *testing.T) {
	sum := SHA256Hex([]byte(`{"a":2,"b":1}`))
	other := SHA256Hex([]byte(`{"a":2,"b":2}`))
	tests := []struct {
		name string
		a, b string
		want bool
	}{
		{"equal", sum, sum, true},
		{"differing, equal length", sum, other, false},
		{"length mismatch", sum, sum[:62], false},
		{"malformed hex", "zz" + sum[2:], sum, false},
		{"odd length", sum[:63], sum[:63], false},
		{"uppercase hex of the same digest", strings.ToUpper(sum), sum, true},
		{"empty strings", "", "", false},
		{"one empty", "", sum, false},
		{"short but valid hex", "00", "00", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DigestEqual(tt.a, tt.b); got != tt.want {
				t.Errorf("DigestEqual(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// A0-2.15: a comparison that gates an action uses crypto/subtle. This asserts
// the mechanism by reading this package's own source (house precedent:
// internal/logging's boundary tests, internal/ids's import scan) rather than
// by timing, which DESIGN §8 rules out as flaky.
func TestGatingComparisonsAreConstantTime(t *testing.T) {
	usesSubtle, directCompare := scanDigestEqual(t)
	if !usesSubtle {
		t.Error("DigestEqual does not call crypto/subtle.ConstantTimeCompare")
	}
	if directCompare {
		t.Error("DigestEqual compares the digest strings directly, which leaks the first differing byte")
	}

	sum := SHA256Hex([]byte(`{"a":1}`))
	for _, pair := range []struct {
		a, b string
		want bool
	}{
		{sum, sum, true},
		{sum, SHA256Hex([]byte(`{"a":2}`)), false},
		{sum, sum[:20], false},
		{"nope!", sum, false},
		{"", "", false},
		{strings.ToUpper(sum), sum, true}, // the same digest, spelled in hex caps
	} {
		if got := DigestEqual(pair.a, pair.b); got != pair.want {
			t.Errorf("DigestEqual(%q, %q) = %v, want %v", pair.a, pair.b, got, pair.want)
		}
	}
}

// scanDigestEqual parses the package's non-test sources and reports whether
// DigestEqual's body calls subtle.ConstantTimeCompare and whether it compares
// one of its parameters with ==. While parsing it also enforces the A0 §4
// preamble for this package: the only internal import allowed is errs.
func scanDigestEqual(t *testing.T) (usesSubtle, directStringCompare bool) {
	t.Helper()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("reading the package directory: %v", err)
	}
	fset := token.NewFileSet()
	found := false
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, filepath.Join(".", e.Name()), nil, parser.SkipObjectResolution)
		if err != nil {
			t.Fatalf("parsing %s: %v", e.Name(), err)
		}
		for _, spec := range file.Imports {
			path := strings.Trim(spec.Path.Value, `"`)
			if strings.Contains(path, "/internal/") && !strings.HasSuffix(path, "/errs") {
				t.Errorf("%s imports %s; a foundation package may import only internal/errs (A0 §4)", e.Name(), path)
			}
		}
		ast.Inspect(file, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}
			if fn.Name.Name != "DigestEqual" {
				return true
			}
			found = true
			params := digestEqualParamNames(t, fn)
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.CallExpr:
					sel, ok := x.Fun.(*ast.SelectorExpr)
					if !ok {
						return true
					}
					if id, ok := sel.X.(*ast.Ident); ok && id.Name == "subtle" && sel.Sel.Name == "ConstantTimeCompare" {
						usesSubtle = true
					}
				case *ast.BinaryExpr:
					if x.Op != token.EQL {
						return true
					}
					if isParam(x.X, params) || isParam(x.Y, params) {
						directStringCompare = true
					}
				}
				return true
			})
			return true
		})
	}
	if !found {
		t.Fatal("no DigestEqual function found in the package sources")
	}
	return usesSubtle, directStringCompare
}

// digestEqualParamNames verifies the signature the == scan relies on —
// func(string, string) bool, both parameters named — and returns their names.
// Taking the names from the declaration keeps the "no direct == on the
// digests" assertion following the code: with names hardcoded in the test,
// renaming DigestEqual's parameters would silently make the scan vacuous
// (DESIGN §8: a test must be able to detect removal of the rule it names).
func digestEqualParamNames(t *testing.T, fn *ast.FuncDecl) []string {
	t.Helper()
	if fn.Recv != nil {
		t.Fatal("DigestEqual is a method, want a function")
	}
	var names []string
	for _, field := range fn.Type.Params.List {
		id, ok := field.Type.(*ast.Ident)
		if !ok || id.Name != "string" {
			t.Fatalf("DigestEqual has a non-string parameter (%T), want func(string, string) bool", field.Type)
		}
		if len(field.Names) == 0 {
			t.Fatal("DigestEqual has an unnamed parameter; a name-matching == scan could never fire")
		}
		for _, n := range field.Names {
			if n.Name == "_" {
				t.Fatal("DigestEqual has a blank parameter; a name-matching == scan could never fire")
			}
			names = append(names, n.Name)
		}
	}
	if len(names) != 2 {
		t.Fatalf("DigestEqual has %d parameters, want func(string, string) bool", len(names))
	}
	return names
}

func isParam(expr ast.Expr, names []string) bool {
	id, ok := expr.(*ast.Ident)
	if !ok {
		return false
	}
	for _, n := range names {
		if id.Name == n {
			return true
		}
	}
	return false
}

// A0-2.16's precondition, and the reason a chain written by one process
// verifies in another: Canonical is a pure function of its input bytes.
func TestCanonicalIsStableAcrossRuns(t *testing.T) {
	const runs = 16
	tests := []struct {
		name    string
		doc     string
		exclude []string
	}{
		{"nested", `{"z":1,"a":{"y":[{"b":1,"a":2},true],"x":"é"},"m":"{\"q\":1}"}`, nil},
		{
			"with exclusions",
			`{"hash":"03c8a7d2","seq":4711,"engagement_id":"eng_01m1y2whfhgbz06ays6dxnvyws",` +
				`"kind":"tool_invoked","recorded_at":"2026-09-07T14:03:22.481Z","prev_hash":"9b74c989"}`,
			[]string{"seq", "prev_hash", "hash"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			first := mustCanonical(t, []byte(tt.doc), tt.exclude...)
			firstDigest := SHA256Hex(first)
			for i := 1; i <= runs; i++ {
				got := mustCanonical(t, []byte(tt.doc), tt.exclude...)
				if !bytes.Equal(got, first) {
					t.Fatalf("run %d produced %q, want %q", i, got, first)
				}
				if d := SHA256Hex(got); d != firstDigest {
					t.Fatalf("run %d digest = %s, want %s", i, d, firstDigest)
				}
			}
		})
	}

	t.Run("input key order cannot leak", func(t *testing.T) {
		// Keys that are pairwise distinct even case-folded: A0-2.5 would
		// reject a and A in one object, so the ordering fixture cannot use them.
		const want = `{"0":4,"B":2,"_c":3,"a":1,"aa":5}`
		for _, in := range []string{
			`{"a":1,"aa":5,"0":4,"B":2,"_c":3}`,
			`{"_c":3,"B":2,"0":4,"aa":5,"a":1}`,
			`{"aa":5,"_c":3,"a":1,"B":2,"0":4}`,
		} {
			got := mustCanonical(t, []byte(in))
			if string(got) != want {
				t.Errorf("Canonical(%q) = %q, want %q", in, got, want)
			}
		}
	})

	t.Run("map iteration order cannot leak through CanonicalValue", func(t *testing.T) {
		value := map[string]any{"b": 1, "a": 2, "Z": 3, "0": 4, "_": 5}
		const want = `{"0":4,"Z":3,"_":5,"a":2,"b":1}`
		for i := 1; i <= runs; i++ {
			got, err := CanonicalValue(value)
			if err != nil {
				t.Fatalf("CanonicalValue = error %v, want bytes", err)
			}
			if string(got) != want {
				t.Fatalf("run %d produced %s, want %s", i, got, want)
			}
		}
	})

	t.Run("case-duplicate keys fail the same way every run", func(t *testing.T) {
		for i := 1; i <= runs; i++ {
			_, err := Canonical([]byte(`{"a":1,"A":2}`))
			if errs.KindOf(err) != errs.Validation {
				t.Fatalf("run %d: kind = %v (%v), want validation", i, errs.KindOf(err), err)
			}
		}
	})
}

// A0-2.7 (P-04): every escape class in the clause is decoded and re-emitted
// per the clause, never passed through. An input written as a Go raw string
// carries the literal six characters of a \uXXXX escape, exactly as the wire
// does; an expected value written as an interpreted string carries the real
// code point.
func TestCanonicalDecodesEscapes(t *testing.T) {
	tests := []struct {
		name string
		doc  string
		want string
	}{
		{"well-formed surrogate pair becomes the literal emoji", `{"s":"\ud83d\ude00"}`, `{"s":"😀"}`},
		{"a literal emoji passes through as the same bytes", `{"s":"😀"}`, `{"s":"😀"}`},
		{"U+2028 escape becomes three raw bytes", `{"s":"\u2028"}`, "{\"s\":\"\u2028\"}"},
		{"U+2029 escape becomes three raw bytes", `{"s":"a\u2029b"}`, "{\"s\":\"a\u2029b\"}"},
		{"U+007F escape becomes one raw byte", `{"s":"\u007f"}`, "{\"s\":\"\u007f\"}"},
		{"a BMP escape decodes", `{"s":"\u0041"}`, `{"s":"A"}`},
		{"upper-case hex escape of a newline becomes the short escape", `{"s":"\u000A"}`, `{"s":"\n"}`},
		{"upper-case hex escape of U+001F becomes lowercase", `{"s":"\u001F"}`, `{"s":"\u001f"}`},
		{"NUL keeps a lowercase u escape", `{"s":"\u0000"}`, `{"s":"\u0000"}`},
		{"escaped quote stays escaped", `{"s":"a\"b"}`, `{"s":"a\"b"}`},
		{"u escape of a quote becomes the escaped quote", `{"s":"\u0022"}`, `{"s":"\""}`},
		{"escaped backslash stays escaped", `{"s":"a\\b"}`, `{"s":"a\\b"}`},
		{"u escape of a backslash stays escaped", `{"s":"\u005c"}`, `{"s":"\\"}`},
		{"the five short escapes are preserved", `{"s":"\b\f\n\r\t"}`, `{"s":"\b\f\n\r\t"}`},
		{"solidus is decoded, not kept escaped", `{"s":"a\/b"}`, `{"s":"a/b"}`},
		{"angle brackets and ampersand are never HTML-escaped", `{"s":"<a href=\"x\">&y"}`, `{"s":"<a href=\"x\">&y"}`},
		{"non-ASCII is literal, never re-encoded", `{"s":"caf\u00e9 \u4e2d"}`, `{"s":"café 中"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Canonical([]byte(tt.doc))
			if err != nil {
				t.Fatalf("Canonical(%q) = error %v, want bytes", tt.doc, err)
			}
			if string(got) != tt.want {
				t.Errorf("Canonical(%q) =\n% x\n%q\nwant\n% x\n%q",
					tt.doc, got, got, []byte(tt.want), tt.want)
			}
		})
	}
	t.Run("CanonicalValue's intermediate bytes are re-emitted, not passed through", func(t *testing.T) {
		// encoding/json escapes U+2028/U+2029 and HTML-escapes < and > when it
		// marshals; A0-2.7 allows it only as the producer of the intermediate
		// bytes, so none of that escaping may reach the canonical output.
		const (
			lineSep = "\u2028"
			paraSep = "\u2029"
			delChar = "\u007f"
			emoji   = "\U0001F600"
		)
		value := map[string]string{
			"sep":   lineSep + paraSep,
			"html":  "<a href=\"x\">&y",
			"del":   delChar,
			"emoji": emoji,
		}
		intermediate, err := json.Marshal(value)
		if err != nil {
			t.Fatalf("marshaling the fixture: %v", err)
		}
		for _, escape := range []string{`\u2028`, `\u003c`} {
			if !bytes.Contains(intermediate, []byte(escape)) {
				t.Fatalf("the fixture never shows encoding/json escaping %q: %s", escape, intermediate)
			}
		}
		got, err := CanonicalValue(value)
		if err != nil {
			t.Fatalf("CanonicalValue = error %v, want bytes", err)
		}
		want := `{"del":"` + delChar + `","emoji":"` + emoji +
			`","html":"<a href=\"x\">&y","sep":"` + lineSep + paraSep + `"}`
		if string(got) != want {
			t.Errorf("CanonicalValue =\n% x\n%q\nwant\n% x\n%q", got, got, []byte(want), want)
		}
		for _, leaked := range []string{`\u2028`, `\u2029`, `\u003c`, `\u003e`} {
			if bytes.Contains(got, []byte(leaked)) {
				t.Errorf("the canonical bytes still carry %q from encoding/json", leaked)
			}
		}
	})
}

// A0-2.12: the exclusion list is a fixed set of top-level names.
func TestExclusionListDropsTopLevelOnly(t *testing.T) {
	t.Run("V5 exclusions", func(t *testing.T) {
		const v5 = `{"engagement_id":"eng_01m1y2whfhgbz06ays6dxnvyws","kind":"tool_invoked",` +
			`"recorded_at":"2026-09-07T14:03:22.481Z","seq":4711,"prev_hash":"9b74c9897bac770ffc029102a200c5de",` +
			`"hash":"03c8a7d2b9e4f1a6d5c0b8e7f2a1d4c3b6a9e8f7d0c1b2a3e4f5061728394a5b"}`
		got := mustCanonical(t, []byte(v5), "seq", "prev_hash", "hash")
		want := `{"engagement_id":"eng_01m1y2whfhgbz06ays6dxnvyws","kind":"tool_invoked","recorded_at":"2026-09-07T14:03:22.481Z"}`
		if string(got) != want {
			t.Errorf("Canonical = %s, want %s", got, want)
		}
		if d := SHA256Hex(got); d != "3695e846ce9e48d6577c0a7ef48ace7b9d34e4ca498dc91aac4309b40c4e10cb" {
			t.Errorf("digest = %s, want V5's published digest", d)
		}
	})

	t.Run("a nested field with an excluded name is kept", func(t *testing.T) {
		doc := `{"seq":1,"outer":{"seq":2,"hash":"x"},"hash":3}`
		got := mustCanonical(t, []byte(doc), "seq", "hash")
		if want := `{"outer":{"hash":"x","seq":2}}`; string(got) != want {
			t.Errorf("Canonical = %s, want %s", got, want)
		}
	})

	t.Run("an array element with an excluded name is kept", func(t *testing.T) {
		got := mustCanonical(t, []byte(`{"a":[{"hash":1}],"hash":2}`), "hash")
		if want := `{"a":[{"hash":1}]}`; string(got) != want {
			t.Errorf("Canonical = %s, want %s", got, want)
		}
	})

	t.Run("excluding an absent name is not an error", func(t *testing.T) {
		got := mustCanonical(t, []byte(`{"a":1}`), "hash", "prev_hash", "seq")
		if want := `{"a":1}`; string(got) != want {
			t.Errorf("Canonical = %s, want %s", got, want)
		}
	})

	t.Run("an excluded field cannot influence the digest", func(t *testing.T) {
		const exclude = "hash"
		first := mustCanonical(t, []byte(`{"a":1,"hash":"0000"}`), exclude)
		second := mustCanonical(t, []byte(`{"a":1,"hash":"ffff"}`), exclude)
		if !bytes.Equal(first, second) {
			t.Fatalf("bytes differ: %q vs %q", first, second)
		}
		if SHA256Hex(first) != SHA256Hex(second) {
			t.Error("digests differ for a document that differs only inside an excluded field")
		}
	})

	t.Run("an excluded field is still validated", func(t *testing.T) {
		// Ruling in doc.go: the token stream must be consumed either way, and
		// dropping without validating would let a malformed document — or a
		// duplicate key — through the exclusion.
		for _, doc := range []string{`{"a":1,"hash":null}`, `{"hash":1,"hash":2}`} {
			if _, err := Canonical([]byte(doc), "hash"); errs.KindOf(err) != errs.Validation {
				t.Errorf("Canonical(%q, \"hash\"): kind = %v (%v), want validation",
					doc, errs.KindOf(err), err)
			}
		}
	})

	t.Run("exclusion matching folds both sides", func(t *testing.T) {
		// A0-2.12 read against A0-2.5: the package holds one notion of "the
		// same key", so a field the caller asked to exclude must not
		// influence the digest whatever its spelling (doc.go's ruling).
		got := mustCanonical(t, []byte(`{"Hash":1,"a":2}`), "hash")
		if want := `{"a":2}`; string(got) != want {
			t.Errorf(`Canonical({"Hash":1,"a":2}, "hash") = %s, want %s`, got, want)
		}
		got = mustCanonical(t, []byte(`{"hash":1,"a":2}`), "HASH")
		if want := `{"a":2}`; string(got) != want {
			t.Errorf(`Canonical({"hash":1,"a":2}, "HASH") = %s, want %s`, got, want)
		}
		// Whole-name matching survives the fold: no case variant of
		// hash_x, xhash or HA is dropped by the declared name hash.
		got = mustCanonical(t, []byte(`{"Hash":1,"HASH_X":2,"xHash":3,"HA":4,"a":5}`), "hash")
		if want := `{"HA":4,"HASH_X":2,"a":5,"xHash":3}`; string(got) != want {
			t.Errorf("Canonical = %s, want %s", got, want)
		}
		first := mustCanonical(t, []byte(`{"Hash":1,"a":2}`), "hash")
		second := mustCanonical(t, []byte(`{"Hash":99,"a":2}`), "hash")
		if !bytes.Equal(first, second) {
			t.Errorf("a folded excluded field still influenced the bytes: %s vs %s", first, second)
		}
	})

	t.Run("exclusion matches the whole name only", func(t *testing.T) {
		got := mustCanonical(t, []byte(`{"hash":1,"hash_x":2,"xhash":3,"HA":4}`), "hash")
		if want := `{"HA":4,"hash_x":2,"xhash":3}`; string(got) != want {
			t.Errorf("Canonical = %s, want %s", got, want)
		}
	})
}

// The §4 sketch's With: add top-level fields to canonical bytes and
// re-canonicalize.
func TestWithAddsKeys(t *testing.T) {
	tests := []struct {
		name string
		doc  string
		add  map[string]any
		want string
	}{
		{
			name: "adds and re-sorts",
			doc:  `{"b":1,"m":{"y":1,"x":2}}`,
			add:  map[string]any{"a": true, "z": []any{}},
			want: `{"a":true,"b":1,"m":{"x":2,"y":1},"z":[]}`,
		},
		{
			name: "an empty document",
			doc:  `{}`,
			add:  map[string]any{"seq": 4711},
			want: `{"seq":4711}`,
		},
		{
			name: "no additions is the canonical document",
			doc:  `{"b": 1, "a": 2}`,
			add:  map[string]any{},
			want: `{"a":2,"b":1}`,
		},
		{
			name: "values go through the same string rules",
			doc:  `{"a":1}`,
			add:  map[string]any{"s": "<&>é", "t": "a\nb"},
			want: "{\"a\":1,\"s\":\"<&>é\",\"t\":\"a\\nb\"}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := With([]byte(tt.doc), tt.add)
			if err != nil {
				t.Fatalf("With(%q, %v) = error %v, want bytes", tt.doc, tt.add, err)
			}
			if string(got) != tt.want {
				t.Errorf("With = %s, want %s", got, tt.want)
			}
			// The same bytes as canonicalizing the equivalent whole document.
			whole, err := CanonicalValue(wholeDocument(t, tt.doc, tt.add))
			if err != nil {
				t.Fatalf("CanonicalValue of the equivalent whole document = error %v", err)
			}
			if !bytes.Equal(got, whole) {
				t.Errorf("With produced %s but the whole document canonicalizes to %s", got, whole)
			}
		})
	}
}

// wholeDocument decodes doc into a plain map and overlays add, so the test can
// compare With's output against canonicalizing the equivalent value. The
// numbers become float64, which is exactly the point: the two paths must agree
// on the bytes regardless of how the value arrived.
func wholeDocument(t *testing.T, doc string, add map[string]any) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(doc), &m); err != nil {
		t.Fatalf("fixture document %q does not decode: %v", doc, err)
	}
	for k, v := range add {
		m[k] = v
	}
	return m
}

// A key already present in doc is a platform defect, not client input.
func TestWithRejectsExistingKey(t *testing.T) {
	_, err := With([]byte(`{"a":1,"b":"x"}`), map[string]any{"b": 2})
	if err == nil {
		t.Fatal("With overwrote an existing key, want rejection")
	}
	if errs.KindOf(err) != errs.Internal {
		t.Errorf("kind = %v, want internal: A1-1.8/A1-5.7 make With platform-only", errs.KindOf(err))
	}
	if !strings.Contains(err.Error(), "b") {
		t.Errorf("error %q does not name the colliding key", err)
	}

	t.Run("a case-duplicate key collides too", func(t *testing.T) {
		// A0-2.5 makes "hash" and "Hash" the same key, so adding one to a
		// document that carries the other is the same platform defect.
		_, err := With([]byte(`{"Hash":1}`), map[string]any{"hash": 2})
		if errs.KindOf(err) != errs.Internal {
			t.Errorf("kind = %v (%v), want internal", errs.KindOf(err), err)
		}
	})

	t.Run("added values meet the ordinary rules", func(t *testing.T) {
		for name, add := range map[string]map[string]any{
			"a float":          {"x": 1.5},
			"a null":           {"x": nil},
			"a nested null":    {"x": map[string]any{"y": nil}},
			"a case-duplicate": {"a": 1, "A": 2},
		} {
			if _, err := With([]byte(`{"a":1}`), add); errs.KindOf(err) != errs.Validation {
				t.Errorf("With with %v = kind %v (%v), want validation", name, errs.KindOf(err), err)
			}
		}
		for name, add := range map[string]map[string]any{
			"an integer string": {"x": "1.0"},
			"an empty object":   {"x": map[string]any{}},
		} {
			if _, err := With([]byte(`{"a":1}`), add); err != nil {
				t.Errorf("With with %v = error %v, want it accepted", name, err)
			}
		}
		big := map[string]any{"x": map[string]any{"y": []any{map[string]any{"z": strings.Repeat("q", MaxBytes)}}}}
		if _, err := With([]byte(`{"a":1}`), big); errs.KindOf(err) != errs.Validation {
			t.Errorf("over-size addition: kind = %v (%v), want validation", errs.KindOf(err), err)
		}
		deep := map[string]any{"x": nested(MaxDepth + 2)}
		if _, err := With([]byte(`{"a":1}`), deep); errs.KindOf(err) != errs.Validation {
			t.Errorf("over-deep addition: kind = %v (%v), want validation", errs.KindOf(err), err)
		}
	})

	t.Run("a malformed document is rejected as validation", func(t *testing.T) {
		_, err := With([]byte(`{"a":01}`), map[string]any{"b": 1})
		if errs.KindOf(err) != errs.Validation {
			t.Errorf("kind = %v (%v), want validation", errs.KindOf(err), err)
		}
	})
}

// nested returns a map nested depth levels deep.
func nested(depth int) any {
	if depth <= 0 {
		return 1
	}
	return map[string]any{"a": nested(depth - 1)}
}
