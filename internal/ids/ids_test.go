package ids

import (
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/daten-krake/sleipnir/internal/errs"
)

// exampleIDs are the twelve example ids of the A0-1.2 table, in table order.
// They are byte-exact bodies and therefore the positive fixtures of A0-1.5.
var exampleIDs = []struct {
	k  Kind
	id string
}{
	{Engagement, "eng_01m1y2whfhgbz06ays6dxnvyws"},
	{Run, "run_01m1y2whfhnjx2am9103w0pnqw"},
	{Job, "job_01m1y2whfhbt69j0h0fbxepw90"},
	{Task, "task_01m1y2whfh1txm57x8dn41r9hg"},
	{Event, "evt_01m1y2whfhp17g0avdqztd2p3x"},
	{AgentNode, "slp_node_01m1y2whfhxydsaem68cmazyc8"},
	{GraphNode, "gn_01m1y2whfhh039ykj5x8mc5a0g"},
	{GraphEdge, "ge_01m1y2whfh62ej11jf4x5gjzv4"},
	{Evidence, "evi_01m1y2whfh3ca875z2x8v8h7qt"},
	{Approval, "apr_01m1y2whfh0asxstccc64q4cfx"},
	{Tool, "tool_01m1y2whfhfjdvwqp9pfxqekmf"},
	{KindUser, "usr_01m1y2whfhv3x6z9b2d5f8h1jk"},
}

// kinds is the closed A0-1.2 set with the total length each row of that table
// fixes (len(prefix)+26).
var kinds = []struct {
	k     Kind
	total int
}{
	{Engagement, 30},
	{Run, 30},
	{Job, 30},
	{Task, 31},
	{Event, 30},
	{AgentNode, 35},
	{GraphNode, 29},
	{GraphEdge, 29},
	{Evidence, 30},
	{Approval, 30},
	{Tool, 31},
	{KindUser, 30},
}

// entropyPattern is the fixed 80-bit entropy most tests generate ids with.
var entropyPattern = []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a}

// patternReader answers one Read with exactly len(want) bytes of want, which is
// what newID asks the entropy source for.
type patternReader struct{ want []byte }

func (r patternReader) Read(p []byte) (int, error) {
	return copy(p, r.want), nil
}

// failReader always fails and counts the attempts made on it, which is how the
// no-retry rule of A0-1.4 is observed.
type failReader struct {
	err   error
	calls int
}

func (r *failReader) Read(p []byte) (int, error) {
	r.calls++
	return 0, r.err
}

// shortReader returns fewer bytes than asked for and no error: an entropy
// source that under-delivers, which must be an error and must not be retried.
type shortReader struct{ calls int }

func (r *shortReader) Read(p []byte) (int, error) {
	r.calls++
	return 3, nil
}

func TestNewIDShapeAndAlphabet(t *testing.T) {
	t.Run("alphabet", func(t *testing.T) {
		if len(Alphabet) != 32 {
			t.Errorf("len(Alphabet) = %d, want 32 (A0-1.1)", len(Alphabet))
		}
		for _, c := range "ilou" {
			if strings.IndexByte(Alphabet, byte(c)) >= 0 {
				t.Errorf("Alphabet contains the excluded character %q (A0-1.1)", c)
			}
		}
		if Alphabet != sortedRunes(Alphabet) {
			t.Errorf("Alphabet is not in byte-lexicographic order, so bodies would not sort chronologically")
		}
	})

	for _, tt := range kinds {
		t.Run(string(tt.k), func(t *testing.T) {
			// New uses the real crypto/rand.Reader; every assertion here is
			// entropy-independent on purpose.
			id, err := New(tt.k)
			if err != nil {
				t.Fatalf("New(%q): %v", tt.k, err)
			}
			if len(id) != tt.total {
				t.Errorf("len(%q) = %d, want %d per the A0-1.2 table", id, len(id), tt.total)
			}
			if !strings.HasPrefix(id, string(tt.k)) {
				t.Errorf("id %q does not start with the byte-exact prefix %q", id, tt.k)
			}
			body := id[len(tt.k):]
			if len(body) != BodyLen {
				t.Fatalf("body %q of %q has length %d, want BodyLen %d", body, id, len(body), BodyLen)
			}
			for i := 0; i < len(body); i++ {
				if strings.IndexByte(Alphabet, body[i]) < 0 {
					t.Errorf("body character %d of %q is %q, not a member of Alphabet", i, body, body[i])
				}
			}
			if body[0] > '7' {
				t.Errorf("first body character of %q is %q, want 0-7 (A0-1.1: only 3 timestamp bits land there)", body, body[0])
			}
			if !Valid(tt.k, id) {
				t.Errorf("Valid(%q, %q) = false for a freshly generated id", tt.k, id)
			}
		})
	}

	t.Run("unknown kind is internal", func(t *testing.T) {
		// A0-3.1: the caller passed a constant of this package, so a Kind
		// outside the closed set is a platform defect, never validation.
		id, err := New(Kind("bog_"))
		if err == nil {
			t.Fatalf("New(%q) = %q, want an error", "bog_", id)
		}
		if got := errs.KindOf(err); got != errs.Internal {
			t.Errorf("errs.KindOf(err) = %q, want %q: %v", got, errs.Internal, err)
		}
	})
}

func TestIDOrderIsChronological(t *testing.T) {
	stamps := []int64{0, 1, 1 << 20, 1 << 32, 1_700_000_000_000, 1_700_000_000_001, 1 << 47, 1<<48 - 1}
	prev := ""
	for _, ms := range stamps {
		id, err := newID(Run, time.UnixMilli(ms), patternReader{want: entropyPattern})
		if err != nil {
			t.Fatalf("newID(Run, %d): %v", ms, err)
		}
		body := id[len(Run):]
		if want := wantBody(t, ms, entropyPattern); body != want {
			t.Errorf("body of ms %d = %q, want %q", ms, body, want)
		}
		if body[0] > '7' {
			t.Errorf("first body character of ms %d is %q, want 0-7", ms, body[0])
		}
		if prev != "" && body <= prev {
			t.Errorf("body %q of ms %d does not sort after the previous body %q", body, ms, prev)
		}
		prev = body
	}

	t.Run("largest 48 bit timestamp", func(t *testing.T) {
		id, err := newID(Run, time.UnixMilli(1<<48-1), patternReader{want: entropyPattern})
		if err != nil {
			t.Fatalf("newID(Run, max48): %v", err)
		}
		if body := id[len(Run):]; body[0] != '7' {
			t.Errorf("first body character = %q, want '7' (all three timestamp bits set)", body[0])
		}
		if !Valid(Run, id) {
			t.Errorf("Valid(Run, %q) = false for the largest representable timestamp", id)
		}
	})

	t.Run("timestamp beyond 48 bits keeps the shape", func(t *testing.T) {
		// The documented ruling: the width is fixed by A0-1.1, so an
		// out-of-range clock keeps its low 48 bits — valid shape, no panic,
		// no error, and A0-1.6 disclaims ordering as an API promise.
		id, err := newID(Run, time.UnixMilli(1<<48), patternReader{want: entropyPattern})
		if err != nil {
			t.Fatalf("newID(Run, 1<<48): %v", err)
		}
		if !Valid(Run, id) {
			t.Errorf("Valid(Run, %q) = false for a truncated timestamp", id)
		}
	})
}

func TestValidRejectsNormalization(t *testing.T) {
	const (
		body = "01m1y2whfhgbz06ays6dxnvyws" // the engagement example of A0-1.2
		eng  = "eng_" + body
	)
	cases := []struct {
		name string
		k    Kind
		s    string
	}{
		{"uppercase", Engagement, strings.ToUpper(eng)},
		{"mixed case", Engagement, "eng_" + strings.ToUpper(body[:13]) + body[13:]},
		{"uppercase prefix only", Engagement, "ENG_" + body},
		{"i mapped to 1", Engagement, "eng_" + strings.Replace(body, "1", "i", 1)},
		{"l mapped to 1", Engagement, "eng_" + strings.Replace(body, "1", "l", 1)},
		{"o mapped to 0", Engagement, "eng_" + strings.Replace(body, "0", "o", 1)},
		{"u substituted", Engagement, "eng_" + strings.Replace(body, "y", "u", 1)},
		{"body of 25", Engagement, "eng_" + body[:25]},
		{"body of 27", Engagement, "eng_" + body + "0"},
		{"empty body", Engagement, "eng_"},
		{"wrong prefix for the type", Run, eng},
		{"prefix of a longer kind", Engagement, "slp_node_" + body},
		{"prefix of a shorter kind", Engagement, "gn_" + body},
		{"empty string", Engagement, ""},
		{"body only, no prefix", Engagement, body},
		{"trailing newline", Engagement, eng + "\n"},
		{"leading space", Engagement, " " + eng},
		{"non-ASCII fullwidth zero", Engagement, "eng_０" + body[1:]},
		{"non-ASCII roman numeral m", Engagement, "eng_" + "ⅿ" + body[1:]},
		// The two rows above are also wrong in length, so they die in the
		// length check and never reach the alphabet scan. These two keep the
		// body at exactly BodyLen bytes while embedding a multi-byte rune, so
		// only the per-byte scan of Alphabet can reject them (A0-1.5).
		{"two-byte rune, body byte length preserved", Engagement, "eng_" + body[:20] + "é" + body[22:]},
		{"four-byte rune, body byte length preserved", Engagement, "eng_" + body[:20] + "\U0001F600" + body[24:]},
		{"unknown kind", Kind("bog_"), eng},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if Valid(tt.k, tt.s) {
				t.Errorf("Valid(%q, %q) = true, want false (A0-1.5 is byte-exact)", tt.k, tt.s)
			}
		})
	}

	t.Run("the multi-byte rows reach the alphabet scan", func(t *testing.T) {
		// A row that is also the wrong length proves nothing about the scan, so
		// pin the length these two rows exist to satisfy.
		for _, s := range []string{
			"eng_" + body[:20] + "é" + body[22:],
			"eng_" + body[:20] + "\U0001F600" + body[24:],
		} {
			if len(s) != len(Engagement)+BodyLen {
				t.Fatalf("case %q is %d bytes, want %d: it would die in the length check and never reach the alphabet scan",
					s, len(s), len(Engagement)+BodyLen)
			}
			if Valid(Engagement, s) {
				t.Errorf("Valid(%q, %q) = true, want false: the per-byte scan of Alphabet must reject a multi-byte rune (A0-1.5)",
					Engagement, s)
			}
		}
	})
}

func TestValidIsByteExact(t *testing.T) {
	for _, tt := range kinds {
		t.Run("generated/"+string(tt.k), func(t *testing.T) {
			id, err := newID(tt.k, time.UnixMilli(1_700_000_000_000), patternReader{want: entropyPattern})
			if err != nil {
				t.Fatalf("newID(%q): %v", tt.k, err)
			}
			if !Valid(tt.k, id) {
				t.Errorf("Valid(%q, %q) = false, want true", tt.k, id)
			}
		})
	}

	for _, tt := range exampleIDs {
		t.Run("example/"+string(tt.k), func(t *testing.T) {
			if !Valid(tt.k, tt.id) {
				t.Errorf("Valid(%q, %q) = false, want true for the A0-1.2 example", tt.k, tt.id)
			}
			for _, other := range exampleIDs {
				if other.k == tt.k {
					continue
				}
				if Valid(other.k, tt.id) {
					t.Errorf("Valid(%q, %q) = true, want false: a prefix identifies exactly one type", other.k, tt.id)
				}
			}
		})
	}
}

func TestIDGenerationUsesCryptoRand(t *testing.T) {
	t.Run("imports", func(t *testing.T) {
		allowed := map[string]bool{
			"crypto/rand": true,
			"io":          true,
			"strings":     true,
			"time":        true,
			"github.com/daten-krake/sleipnir/internal/errs": true, // the one internal import A0 §4 allows
		}
		// An allow-list alone is only the negative half of A0-1.4: dropping
		// crypto/rand for a deterministic reader would drop the import too and
		// stay green. Require the positive half as well.
		var sawCryptoRand bool
		for _, f := range packageFiles(t, ".") {
			for _, imp := range f.imports {
				if imp == "crypto/rand" {
					sawCryptoRand = true
				}
				if imp == "math/rand" || imp == "math/rand/v2" {
					t.Errorf("%s imports %q: A0-1.4 forbids math/rand", f.path, imp)
					continue
				}
				if !allowed[imp] {
					t.Errorf("%s imports %q, which is outside the allowed set of A0 §4 (stdlib plus internal/errs)", f.path, imp)
				}
			}
		}
		if !sawCryptoRand {
			t.Error("no file in this package imports crypto/rand: A0-1.4 requires it as the only entropy source")
		}
	})

	t.Run("body follows the entropy source", func(t *testing.T) {
		const ms = 1_700_000_000_000
		zeros := make([]byte, entropyLen)
		pattern := entropyPattern

		zeroID, err := newID(Engagement, time.UnixMilli(ms), patternReader{want: zeros})
		if err != nil {
			t.Fatalf("newID with zero entropy: %v", err)
		}
		wantID, err := newID(Engagement, time.UnixMilli(ms), patternReader{want: pattern})
		if err != nil {
			t.Fatalf("newID with pattern entropy: %v", err)
		}
		if want := wantBody(t, ms, pattern); want != wantID[len(Engagement):] {
			t.Errorf("body = %q, want %q: the 80 entropy bits must be the reader's bytes encoded most significant bit first",
				wantID[len(Engagement):], want)
		}
		if want := wantBody(t, ms, zeros); want != zeroID[len(Engagement):] {
			t.Errorf("body = %q, want %q for zero entropy", zeroID[len(Engagement):], want)
		}
		// Same clock, different entropy: the time prefix of the body is
		// identical and the 16 entropy characters differ, so they come from
		// the reader — not from a counter, a name or a hash.
		if got := zeroID[len(Engagement) : len(Engagement)+10]; got != wantID[len(Engagement):len(Engagement)+10] {
			t.Errorf("time characters %q and %q disagree for the same timestamp", got, wantID[len(Engagement):len(Engagement)+10])
		}
		if zeroID[len(Engagement)+10:] == wantID[len(Engagement)+10:] {
			t.Errorf("entropy characters of two different entropy sources are equal: %q vs %q", zeroID, wantID)
		}
	})
}

func TestUniquenessViolationIsInternal(t *testing.T) {
	// A0-1.4 owns two properties here: a generation failure is internal, and
	// the entropy source is attempted exactly once. The store-side half — a
	// uniqueness violation at insert surfaces as internal and is never
	// silently retried — belongs to the store package (WP-22); see doc.go.
	t.Run("entropy failure is internal and attempted once", func(t *testing.T) {
		cause := errors.New("entropy exhausted")
		r := &failReader{err: cause}
		id, err := newID(Run, time.UnixMilli(1_700_000_000_000), r)
		if err == nil {
			t.Fatalf("newID with a failing reader = %q, want an error", id)
		}
		if r.calls != 1 {
			t.Errorf("the entropy reader was attempted %d times, want exactly 1 (A0-1.4: never a retry loop)", r.calls)
		}
		if got := errs.KindOf(err); got != errs.Internal {
			t.Errorf("errs.KindOf(err) = %q, want %q", got, errs.Internal)
		}
		if !errors.Is(err, cause) {
			t.Errorf("errors.Is(err, cause) = false, want the entropy failure reachable as the cause: %v", err)
		}

		t.Run("message shape", func(t *testing.T) {
			assertShape(t, err, "ids.newID", ": "+cause.Error())
		})
	})

	t.Run("short read is internal and attempted once", func(t *testing.T) {
		r := &shortReader{}
		id, err := newID(Run, time.UnixMilli(1_700_000_000_000), r)
		if err == nil {
			t.Fatalf("newID with a short read = %q, want an error", id)
		}
		if r.calls != 1 {
			t.Errorf("the entropy reader was attempted %d times, want exactly 1 (A0-1.4: never a retry loop)", r.calls)
		}
		if got := errs.KindOf(err); got != errs.Internal {
			t.Errorf("errs.KindOf(err) = %q, want %q", got, errs.Internal)
		}
		if !errors.Is(err, io.ErrUnexpectedEOF) {
			t.Errorf("errors.Is(err, io.ErrUnexpectedEOF) = false: %v", err)
		}
	})

	t.Run("unknown kind is internal", func(t *testing.T) {
		_, err := newID(Kind("nope_"), time.UnixMilli(1_700_000_000_000), patternReader{want: entropyPattern})
		if err == nil {
			t.Fatal("newID with an unknown kind = nil error, want internal")
		}
		if got := errs.KindOf(err); got != errs.Internal {
			t.Errorf("errs.KindOf(err) = %q, want %q", got, errs.Internal)
		}
		t.Run("message shape", func(t *testing.T) {
			assertShape(t, err, "ids.newID", "")
		})
	})

	t.Run("New reports the defect of its own table", func(t *testing.T) {
		// New must not lose the classification on the way to newID.
		_, err := New(Kind("bog_"))
		if err == nil {
			t.Fatal("New with an unknown kind = nil error, want internal")
		}
		if errs.KindOf(err) != errs.Internal {
			t.Errorf("errs.KindOf(New(bog_)) = %q, want %q", errs.KindOf(err), errs.Internal)
		}
		if op := errs.OpOf(err); op != "ids.newID" {
			t.Errorf("errs.OpOf(err) = %q, want %q, the function that attempted the generation", op, "ids.newID")
		}
	})
}

// assertShape checks the ADR-0019 §2 rendering
// component.Function: what was attempted: key identifiers: cause, where cause
// is wantSuffix (empty for a leaf error).
func assertShape(t *testing.T, err error, wantOp, wantSuffix string) {
	t.Helper()
	msg := err.Error()
	if got := errs.OpOf(err); got != wantOp {
		t.Errorf("errs.OpOf(err) = %q, want %q", got, wantOp)
	}
	segments := strings.Split(msg, ": ")
	if len(segments) < 3 {
		t.Fatalf("err.Error() = %q, want at least op, what-was-attempted and identifier segments", msg)
	}
	if segments[0] != wantOp {
		t.Errorf("first segment = %q, want %q", segments[0], wantOp)
	}
	if wantSuffix == "" {
		return
	}
	if !strings.HasSuffix(msg, wantSuffix) {
		t.Errorf("err.Error() = %q, want it to end with the cause %q", msg, wantSuffix)
	}
}

// wantBody is the reference encoder of A0-1.1, written from the definition
// rather than from ids.go's shifts: it spells the 130-bit frame (two padding
// bits, then the 48 timestamp bits, then the 80 entropy bits) as a string of
// ones and zeros and reads it five bits at a time.
func wantBody(t *testing.T, ms int64, entropy []byte) string {
	t.Helper()
	if len(entropy) != entropyLen {
		t.Fatalf("entropy is %d bytes, want %d", len(entropy), entropyLen)
	}
	var bits strings.Builder
	bits.WriteString("00")
	for _, b := range append(be48(ms), entropy...) {
		bits.WriteString(fmt.Sprintf("%08b", b))
	}
	if bits.Len() != 130 {
		t.Fatalf("reference frame is %d bits, want 130", bits.Len())
	}
	var body strings.Builder
	for i := 0; i < BodyLen; i++ {
		v, err := strconv.ParseUint(bits.String()[i*5:(i+1)*5], 2, 8)
		if err != nil {
			t.Fatalf("parsing reference bits %q as a binary value: %v", bits.String()[i*5:(i+1)*5], err)
		}
		body.WriteByte(Alphabet[v])
	}
	return body.String()
}

// be48 returns the 48-bit big-endian form of ms.
func be48(ms int64) []byte {
	u := uint64(ms)
	return []byte{byte(u >> 40), byte(u >> 32), byte(u >> 24), byte(u >> 16), byte(u >> 8), byte(u)}
}

// sortedRunes returns s with its bytes in ascending order.
func sortedRunes(s string) string {
	b := []byte(s)
	for i := 1; i < len(b); i++ {
		for j := i; j > 0 && b[j-1] > b[j]; j-- {
			b[j-1], b[j] = b[j], b[j-1]
		}
	}
	return string(b)
}

// goFile holds the facts the import-boundary scan needs from one source file.
type goFile struct {
	path    string
	imports []string
}

// packageFiles parses the non-test Go files of dir in name order.
func packageFiles(t *testing.T, dir string) []goFile {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("reading package directory %s: %v", dir, err)
	}
	fset := token.NewFileSet()
	var files []goFile
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			t.Fatalf("parsing %s: %v", path, err)
		}
		var imports []string
		for _, spec := range file.Imports {
			p, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				t.Fatalf("unquoting import %s in %s: %v", spec.Path.Value, path, err)
			}
			imports = append(imports, p)
		}
		files = append(files, goFile{path: path, imports: imports})
	}
	if len(files) == 0 {
		t.Fatalf("directory %s holds no non-test Go files", dir)
	}
	return files
}
