package errs_test

import (
	"bytes"
	"crypto/sha256"
	"encoding"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/daten-krake/sleipnir/internal/errs"
)

// The expectations in this file are transcribed from contracts/A0-conventions.md
// A0-3.1/A0-3.6/A0-3.10 and adr/ADR-0019 §1–§2, not from internal/errs. Where a
// row looks like the implementation it is coincidence of correctness.

// --- A0-3.1 kind table (contract oracle) ------------------------------------

// a0KindRow is one row of the A0-3.1 table: kind, HTTP status, retryability.
type a0KindRow struct {
	kind       errs.Kind
	httpStatus int
	retryable  bool
	why        string
}

// a0KindTable is the A0-3.1 table as written in the contract. It MUST be kept
// byte-identical to the contract document; changing code here does not change
// it, only editing this list does.
var a0KindTable = []a0KindRow{
	{"validation", 400, false, "malformed input, bad id, unknown field on write, bad enum, bad timestamp, canonicalization failure"},
	{"auth", 401, false, "credential missing, expired, malformed or unverifiable"},
	{"forbidden", 403, false, "principal-level denial naming no object"},
	{"notfound", 404, false, "object absent or outside the caller's engagement scope"},
	{"conflict", 409, false, "state prevents the request"},
	{"approval_required", 409, false, "execution or spawn with no valid approval for the fingerprint"},
	{"approval_expired", 409, false, "approval past its window"},
	{"integrity_failed", 409, false, "hash-chain verification failed"},
	{"summary_too_large", 413, false, "a capped field or count exceeded its budget under mechanism R"},
	{"timeout", 504, true, "a platform-side deadline expired"},
	{"upstream", 502, true, "an upstream answered with an error or unusable bytes"},
	{"rate_limited", 429, true, "platform rate limiter"},
	{"internal", 500, false, "platform defect or unclassified failure"},
}

func TestKindStatusMatchesA0Table(t *testing.T) {
	if len(a0KindTable) != 13 {
		t.Fatalf("A0-3.1 closes the vocabulary at 13 kinds, table has %d", len(a0KindTable))
	}
	seen := map[errs.Kind]bool{}
	for _, row := range a0KindTable {
		if seen[row.kind] {
			t.Fatalf("duplicate row for kind %q in the contract table", row.kind)
		}
		seen[row.kind] = true
		t.Run(string(row.kind), func(t *testing.T) {
			if got := row.kind.Status(); got != row.httpStatus {
				t.Errorf("Status() = %d, want %d (A0-3.1: %s)", got, row.httpStatus, row.why)
			}
		})
	}
}

func TestKindRetryableMatchesA0Table(t *testing.T) {
	for _, row := range a0KindTable {
		t.Run(string(row.kind), func(t *testing.T) {
			if got := row.kind.Retryable(); got != row.retryable {
				t.Errorf("Retryable() = %v, want %v (A0-3.11)", got, row.retryable)
			}
		})
	}
}

// TestUnknownKindIsInternal covers A0-3.3 (a kind the receiver cannot classify
// is internal and terminal) and A0-3.2 (these kinds were deliberately dropped
// and must never be honoured).
func TestUnknownKindIsInternal(t *testing.T) {
	tests := []struct {
		name string
		kind errs.Kind
	}{
		{"empty", ""},
		{"uppercase", "VALIDATION"},
		{"hyphen", "not-found"},
		{"camel", "notFound"},
		{"dropped-quota", "quota_exceeded"},
		{"dropped-unavailable", "unavailable"},
		{"dropped-not-implemented", "not_implemented"},
		{"dropped-precondition", "precondition_failed"},
		{"dropped-scope-denied", "scope_denied"},
		{"future-addition", "stream_interrupted"},
		{"trailing-space", "internal "},
		{"long", errs.Kind(strings.Repeat("k", 512))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.kind.Status(); got != 500 {
				t.Errorf("Status() = %d, want 500", got)
			}
			// KindOf normalizes a hand-built error carrying the unknown kind.
			if got := errs.KindOf(&errs.Error{Kind: tt.kind, Op: "x.Fn", Msg: "y"}); got != errs.Internal {
				t.Errorf("KindOf() = %q, want internal", got)
			}
			// Creation normalizes, so no unknown kind can reach a client.
			if got := errs.KindOf(errs.New(tt.kind, "msg")); got != errs.Internal {
				t.Errorf("KindOf(New(%q)) = %q, want internal", tt.kind, got)
			}
			env := errs.NewEnvelope(errs.New(tt.kind, "msg"), errs.Attrs{}, 0)
			if env.Error.Kind != errs.Internal || env.Error.Message == "" {
				t.Errorf("NewEnvelope() = %+v, want internal kind and non-empty message", env.Error)
			}
			if got := errs.New(tt.kind, "msg").Error(); strings.Contains(got, string(tt.kind)) && tt.kind != "" {
				t.Errorf("rendered message %q still carries the unknown kind %q", got, tt.kind)
			}
		})
	}
}

// --- ADR-0019 §1 origin capture ---------------------------------------------

// The helpers below are each exactly one frame away from the constructor, so
// the captured op MUST name the helper and never the function that called it.
// With an off-by-one skip these return "errs_test.TestNewCapturesOriginFunction".

func originViaNew() error {
	return errs.New(errs.Validation, "check blacklist entry")
}

func originViaNewf() error {
	return errs.Newf(errs.Validation, "check %s entry", "blacklist")
}

func originViaWrap() error {
	return errs.Wrap(errTestSentinel, "check blacklist entry")
}

func originViaWrapf() error {
	return errs.Wrapf(errTestSentinel, "check %s entry", "blacklist")
}

var errTestSentinel = errors.New("sentinel root cause")

func TestNewCapturesOriginFunction(t *testing.T) {
	tests := []struct {
		name string
		call func() error
		op   string
	}{
		{"New", originViaNew, "errs_test.originViaNew"},
		{"Newf", originViaNewf, "errs_test.originViaNewf"},
		{"Wrap", originViaWrap, "errs_test.originViaWrap"},
		{"Wrapf", originViaWrapf, "errs_test.originViaWrapf"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.call()
			var e *errs.Error
			if !errors.As(err, &e) {
				t.Fatalf("errors.As(%v) found no *errs.Error", err)
			}
			if e.Op != tt.op {
				t.Errorf("captured op = %q, want %q (ADR-0019 §1: the error carries the function it originated in)", e.Op, tt.op)
			}
			if got := errs.OpOf(err); got != tt.op {
				t.Errorf("OpOf() = %q, want %q", got, tt.op)
			}
			if !strings.HasPrefix(err.Error(), tt.op+": ") {
				t.Errorf("Error() = %q, want prefix %q", err.Error(), tt.op+": ")
			}
		})
	}
}

func TestNewfFormatsMessage(t *testing.T) {
	tests := []struct {
		name   string
		format string
		args   []any
		want   string
	}{
		{"no args", "read node", nil, "read node"},
		{"string", "read node %s", []any{"gn_01"}, "read node gn_01"},
		{"multiple", "%s=%s cap=%d", []any{"field", "summary", 512}, "field=summary cap=512"},
		{"percent literal", "100%% sure", nil, "100% sure"},
		{"extra verb arg", "one %s", []any{"two", "ignored"}, "one two%!(EXTRA string=ignored)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errs.Newf(errs.Validation, tt.format, tt.args...)
			var e *errs.Error
			if !errors.As(err, &e) {
				t.Fatalf("no *errs.Error in %v", err)
			}
			if e.Msg != tt.want {
				t.Errorf("Msg = %q, want %q", e.Msg, tt.want)
			}
		})
	}
}

// --- ADR-0019 §2 message format ---------------------------------------------

type errTestForeign struct{ text string }

func (e errTestForeign) Error() string { return e.text }

func adrLeaf() error {
	return errs.Wrap(errTestForeign{"store driver: pq: duplicate key value violates unique constraint"},
		"insert event for engagement=eng_01m1 run=run_01m1")
}

func adrMiddle() error { return errs.Wrap(adrLeaf(), "append batch of 3 events") }

func adrTop() error { return errs.Wrap(adrMiddle(), "handle events:append") }

func TestMessageFormatMatchesADR0019(t *testing.T) {
	const want = "errs_test.adrTop: handle events:append: " +
		"errs_test.adrMiddle: append batch of 3 events: " +
		"errs_test.adrLeaf: insert event for engagement=eng_01m1 run=run_01m1: " +
		"store driver: pq: duplicate key value violates unique constraint"

	got := adrTop().Error()
	if got != want {
		t.Errorf("Error() mismatch\n got: %s\nwant: %s", got, want)
	}
	// The shape is component.Function: attempt: ids: cause, one segment per
	// layer, so the string doubles as the execution trace (ADR-0019 §2).
	segments := strings.Split(got, ": ")
	if len(segments) < 4 {
		t.Fatalf("rendered %q has %d colon-separated segments, want >= 4", got, len(segments))
	}

	// An empty segment contributes nothing: a hand-built error without an op or
	// a message must not render ": " noise around the real content.
	empty := []struct {
		err  error
		want string
	}{
		{&errs.Error{Kind: errs.Internal, Msg: "no op captured"}, "no op captured"},
		{&errs.Error{Kind: errs.Internal, Op: "store.InsertEvent"}, "store.InsertEvent: "},
		{&errs.Error{Kind: errs.Internal, Op: "api.Handler", Err: errTestForeign{"cause text"}}, "api.Handler: cause text"},
		{&errs.Error{Kind: errs.Internal, Msg: "attempt", Err: errTestForeign{"cause text"}}, "attempt: cause text"},
	}
	for _, tt := range empty {
		if got := tt.err.Error(); got != tt.want {
			t.Errorf("Error() = %q, want %q", got, tt.want)
		}
	}
}

// TestMessageHasNoStackTraceOrSQL guards A0-3.5: the package must never put a
// stack trace, file:line pair or multi-line body into a message. It is a
// regression net for someone "improving" origin capture to include file and
// line, which is the natural next step and is forbidden.
func TestMessageHasNoStackTraceOrSQL(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"leaf", originViaNew()},
		{"formatted leaf", originViaNewf()},
		{"foreign cause", adrLeaf()},
		{"three layers", adrTop()},
		{"wrapped sentinel", originViaWrap()},
	}
	forbidden := []string{"goroutine", ".go:", "\n\t", "\n", "0x", "runtime.Stack", "SELECT ", "INSERT INTO "}
	for _, tt := range tests {
		got := tt.err.Error()
		for _, bad := range forbidden {
			if strings.Contains(got, bad) {
				t.Errorf("%s: message contains %q (A0-3.5): %q", tt.name, bad, got)
			}
		}
	}
}

// --- bounded walk over the cause chain (review ruling E-a) ------------------

// wantChainDepth is A0-2.11's MaxDepth, the platform's single nesting bound.
// It is repeated here on purpose: the test states the contract value, so
// changing the constant in errs.go without changing A0-2.11 turns this red.
const wantChainDepth = 32

// truncationSegment is the final segment Error() MUST append when it stops at
// the depth cap, so the rendered string stays diagnosable instead of ending
// silently.
const truncationSegment = "chain truncated at 32 layers"

// chainWithLeaf builds depth *Error layers with leaf as the innermost cause.
// Layer i renders as "layer<i>.Run: attempt <i>", so the number of rendered
// layers can be counted in the output.
func chainWithLeaf(depth int, leaf error) *errs.Error {
	cur := leaf
	for i := depth; i >= 1; i-- {
		cur = &errs.Error{
			Kind: errs.Internal,
			Op:   fmt.Sprintf("layer%d.Run", i),
			Msg:  fmt.Sprintf("attempt %d", i),
			Err:  cur,
		}
	}
	return cur.(*errs.Error)
}

// returnsWithin runs f on a private goroutine and returns its result, failing
// the test if f has not returned within budget. The select-on-timer is a
// deadlock guard, not a synchronisation primitive — there is no ordering being
// waited for, which is what DESIGN §8 forbids.
func returnsWithin(t *testing.T, budget time.Duration, what string, f func() string) string {
	t.Helper()
	done := make(chan string, 1)
	go func() { done <- f() }()
	select {
	case got := <-done:
		return got
	case <-time.After(budget):
		t.Fatalf("%s did not return within %s: the cause-chain walk is unbounded", what, budget)
		return "" // unreachable; t.Fatalf stops this test goroutine
	}
}

func TestErrorChainWalkIsBounded(t *testing.T) {
	// A hand-built cycle: e.Err = e. Unreachable through New/Wrap, reachable
	// through the exported fields — which is the whole point of the cap.
	selfCycle := &errs.Error{Kind: errs.Internal, Op: "store.AppendEvent", Msg: "append event for engagement=eng_01m1"}
	selfCycle.Err = selfCycle

	// A two-node cycle, to show the cap counts layers rather than distinct
	// errors, and that it holds for any shape of loop.
	cycleA := &errs.Error{Kind: errs.Internal, Op: "nodeA.Run", Msg: "attempt A"}
	cycleB := &errs.Error{Kind: errs.Internal, Op: "nodeB.Run", Msg: "attempt B", Err: cycleA}
	cycleA.Err = cycleB

	// A cycle whose layers carry no op: the cap must stop the walk here too, and
	// the notice must join the rendered segments with the ordinary separator.
	blankCycle := &errs.Error{Kind: errs.Internal, Op: "no-op layer"}
	blankCycle.Err = blankCycle

	// The pathological corner: nothing to render at all, still bounded.
	emptyCycle := &errs.Error{Kind: errs.Internal}
	emptyCycle.Err = emptyCycle

	leaf := errTestForeign{"store driver: duplicate key"}

	// wantTruncations is 1 for a chain the cap stopped and 0 for a chain that
	// ended on its own; A0-2.11 accepts depth 32 and rejects 33, so the boundary
	// rows below pin both sides.
	tests := []struct {
		name            string
		err             error
		layerMarker     string
		wantLayerCopes  int
		wantTruncations int
		mustContain     []string
		mustNotContain  []string
		wantExact       string
	}{
		{
			name:            "self cycle stops at the cap",
			err:             selfCycle,
			layerMarker:     "store.AppendEvent: append",
			wantLayerCopes:  wantChainDepth,
			wantTruncations: 1,
			mustContain:     []string{"store.AppendEvent: append event for engagement=eng_01m1"},
		},
		{
			name:            "two-node cycle stops at the cap",
			err:             cycleA,
			layerMarker:     ".Run: attempt ",
			wantLayerCopes:  wantChainDepth,
			wantTruncations: 1,
			mustContain:     []string{"nodeA.Run: attempt A", "nodeB.Run: attempt B"},
		},
		{
			name:            "cycle with one message and no op stops at the cap",
			err:             blankCycle,
			layerMarker:     "no-op layer: ",
			wantLayerCopes:  wantChainDepth,
			wantTruncations: 1,
			mustContain:     []string{"no-op layer: "},
		},
		{
			name:            "cycle with nothing to render still terminates",
			err:             emptyCycle,
			layerMarker:     ".Run: attempt ",
			wantLayerCopes:  0,
			wantTruncations: 1,
			mustContain:     []string{truncationSegment},
		},
		{
			name:            "chain at the boundary renders in full",
			err:             chainWithLeaf(wantChainDepth, leaf),
			layerMarker:     ".Run: attempt ",
			wantLayerCopes:  wantChainDepth,
			wantTruncations: 0,
			mustContain:     []string{"layer1.Run: attempt 1", "layer32.Run: attempt 32", "store driver: duplicate key"},
			mustNotContain:  []string{truncationSegment},
		},
		{
			name:            "chain one past the boundary is truncated",
			err:             chainWithLeaf(wantChainDepth+1, leaf),
			layerMarker:     ".Run: attempt ",
			wantLayerCopes:  wantChainDepth,
			wantTruncations: 1,
			mustContain:     []string{"layer32.Run: attempt 32"},
			mustNotContain:  []string{"layer33.Run", "store driver: duplicate key"},
		},
		{
			name:            "thousand-layer chain still terminates",
			err:             chainWithLeaf(1000, leaf),
			layerMarker:     ".Run: attempt ",
			wantLayerCopes:  wantChainDepth,
			wantTruncations: 1,
			mustNotContain:  []string{"layer33.Run", "layer1000.Run"},
		},
		{
			name:           "short chain is unaffected",
			err:            chainWithLeaf(3, leaf),
			layerMarker:    ".Run: attempt ",
			wantLayerCopes: 3,
			wantExact: "layer1.Run: attempt 1: layer2.Run: attempt 2: layer3.Run: attempt 3: " +
				"store driver: duplicate key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 10s is generous: a capped walk of 32 layers takes microseconds.
			got := returnsWithin(t, 10*time.Second, "Error()", tt.err.Error)

			if tt.wantExact != "" && got != tt.wantExact {
				t.Errorf("Error() =\n %q\nwant\n %q", got, tt.wantExact)
			}
			if n := strings.Count(got, truncationSegment); n != tt.wantTruncations {
				t.Errorf("rendered string contains %d copies of %q, want %d: %q",
					n, truncationSegment, tt.wantTruncations, got)
			}
			if n := strings.Count(got, tt.layerMarker); n != tt.wantLayerCopes {
				t.Errorf("%d layers rendered (%d copies of %q), want %d: %q",
					n, n, tt.layerMarker, tt.wantLayerCopes, got)
			}
			if tt.wantTruncations == 1 && !strings.HasSuffix(got, truncationSegment) {
				t.Errorf("truncation notice is not the final segment: %q", got)
			}
			for _, want := range tt.mustContain {
				if !strings.Contains(got, want) {
					t.Errorf("Error() missing %q: %q", want, got)
				}
			}
			for _, bad := range tt.mustNotContain {
				if strings.Contains(got, bad) {
					t.Errorf("Error() must not contain %q: %q", bad, got)
				}
			}
		})
	}

	// The availability argument behind the ruling is the log and envelope path:
	// Error() rides into every slog record and every /api/v1 body, so pin that
	// both conversions terminate on a cyclic error.
	envelope := returnsWithin(t, 10*time.Second, "NewEnvelope()", func() string {
		b, err := json.Marshal(errs.NewEnvelope(selfCycle, errs.Attrs{}, 0))
		if err != nil {
			return "marshal error: " + err.Error()
		}
		return string(b)
	})
	if !strings.Contains(envelope, truncationSegment) {
		t.Errorf("envelope message does not name the truncation: %s", envelope)
	}

	// The same chain through slog's text handler, which is the hot path the
	// ruling cites: an error value in attrs is rendered by its Error method.
	logLine := returnsWithin(t, 10*time.Second, "slog handler", func() string {
		var buf bytes.Buffer
		slog.New(slog.NewTextHandler(&buf, nil)).Error("append batch failed", "err", selfCycle)
		return buf.String()
	})
	if !strings.Contains(logLine, truncationSegment) {
		t.Errorf("log record does not name the truncation: %s", logLine)
	}
}

// --- cause chain ------------------------------------------------------------

func TestWrapPreservesCauseChain(t *testing.T) {
	root := errTestSentinel
	mid := errTestForeign{"foreign middle"}
	outer := errs.Wrap(errs.Wrap(errors.Join(root, mid), "join layer"), "outer layer")

	if !errors.Is(outer, root) {
		t.Error("errors.Is could not reach the sentinel root through two errs layers")
	}
	var target errTestForeign
	if !errors.As(outer, &target) {
		t.Fatal("errors.As could not reach the typed foreign error through two errs layers")
	}
	if target.text != "foreign middle" {
		t.Errorf("errors.As reached %+v, want the foreign middle", target)
	}
	// The chain is walkable one link at a time as well.
	l1 := errors.Unwrap(outer)
	if l1 == nil {
		t.Fatal("Unwrap(outer) = nil, want the inner errs layer")
	}
	if errors.Unwrap(l1) == nil {
		t.Fatal("Unwrap(inner) = nil, want the joined pair")
	}
	var e *errs.Error
	if !errors.As(outer, &e) || e.Op != "errs_test.TestWrapPreservesCauseChain" {
		t.Errorf("errors.As gave %+v, want the outermost errs layer of this test function", e)
	}
}

func TestWrapInheritsKindFromCause(t *testing.T) {
	for _, row := range a0KindTable {
		t.Run(string(row.kind), func(t *testing.T) {
			leaf := errs.New(row.kind, "leaf attempt")
			once := errs.Wrap(leaf, "middle attempt")
			twice := errs.Wrapf(once, "outer %s", "attempt")
			for name, err := range map[string]error{"once": once, "twice": twice} {
				if got := errs.KindOf(err); got != row.kind {
					t.Errorf("%s: KindOf = %q, want inherited %q", name, got, row.kind)
				}
				if got := errs.NewEnvelope(err, errs.Attrs{}, 0).Error.Kind; got != row.kind {
					t.Errorf("%s: envelope kind = %q, want %q", name, got, row.kind)
				}
			}
		})
	}
}

// TestWrapCannotReclassifyKind is the negative test for ADR-0019 §1 and A0-3.12:
// no wrapping layer may change the kind an agent is keyed to.
func TestWrapCannotReclassifyKind(t *testing.T) {
	tests := []struct {
		name string
		leaf error
		want errs.Kind
	}{
		{"validation stays validation", errs.New(errs.Validation, "bad id"), errs.Validation},
		{"approval stays approval", errs.New(errs.ApprovalRequired, "no approval"), errs.ApprovalRequired},
		{"retryable stays retryable", errs.New(errs.Timeout, "deadline"), errs.Timeout},
		{"foreign stays internal", errTestForeign{"driver said no"}, errs.Internal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.leaf
			for i := range 6 {
				err = errs.Wrap(err, fmt.Sprintf("layer %d", i))
				if got := errs.KindOf(err); got != tt.want {
					t.Fatalf("after %d wraps KindOf = %q, want %q (wrapping must never reclassify)", i+1, got, tt.want)
				}
				if got := errs.NewEnvelope(err, errs.Attrs{}, 0).Error.Kind; got != tt.want {
					t.Fatalf("envelope kind = %q, want %q", got, tt.want)
				}
			}
			// A validation failure must not surface as a retryable server error.
			if tt.want == errs.Validation {
				got := errs.KindOf(err)
				if got.Retryable() || got.Status() != 400 {
					t.Fatalf("validation surfaced as %q (retryable=%v status=%d)", got, got.Retryable(), got.Status())
				}
			}
		})
	}
}

func TestWrapNilReturnsNil(t *testing.T) {
	tests := []struct {
		name string
		call func() error
	}{
		{"Wrap", func() error { return errs.Wrap(nil, "nothing to wrap") }},
		{"Wrapf", func() error { return errs.Wrapf(nil, "nothing to %s", "wrap") }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.call()
			if got != nil {
				t.Fatalf("%s(nil) = %v (%T), want an untyped nil error", tt.name, got, got)
			}
			if v := reflect.ValueOf(got); v.IsValid() {
				t.Fatalf("%s(nil) returned a non-empty interface holding %s", tt.name, v.Kind())
			}
			// The idiomatic guard must be reachable without a panic.
			if errors.Is(got, errTestSentinel) || errs.KindOf(got) != errs.Internal || errs.OpOf(got) != "" {
				t.Fatalf("%s(nil) is not inert: KindOf=%q OpOf=%q", tt.name, errs.KindOf(got), errs.OpOf(got))
			}
			if errs.NewEnvelope(got, errs.Attrs{}, 0).Error.Kind != errs.Internal {
				t.Fatalf("%s(nil) envelope must be internal", tt.name)
			}
		})
	}
}

// --- OpOf -------------------------------------------------------------------

func opLeaf() error  { return errs.New(errs.NotFound, "read approval") }
func opOuter() error { return errs.Wrap(opLeaf(), "authorize action") }

func TestOpOfReturnsOutermostOp(t *testing.T) {
	if got := errs.OpOf(opOuter()); got != "errs_test.opOuter" {
		t.Errorf("OpOf() = %q, want the outermost op errs_test.opOuter", got)
	}
	if got := errs.OpOf(opLeaf()); got != "errs_test.opLeaf" {
		t.Errorf("OpOf() = %q, want errs_test.opLeaf", got)
	}
	if got := errs.OpOf(errs.Wrap(errs.Wrap(opOuter(), "extra"), "outermost")); got != "errs_test.TestOpOfReturnsOutermostOp" {
		t.Errorf("OpOf() = %q, want the outermost wrap site", got)
	}
}

func TestOpOfForeignErrorIsEmpty(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"errors.New", errors.New("plain")},
		{"custom type", errTestForeign{"plain"}},
		{"nil", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := errs.OpOf(tt.err); got != "" {
				t.Errorf("OpOf(%v) = %q, want empty", tt.err, got)
			}
		})
	}
}

// --- A0-3.6 envelope --------------------------------------------------------

func TestEnvelopeOmitsAbsentAttrs(t *testing.T) {
	tests := []struct {
		name     string
		attrs    errs.Attrs
		wantKeys []string
		reject   []string
	}{
		{
			name:     "all absent",
			attrs:    errs.Attrs{},
			wantKeys: nil,
			reject:   []string{"engagement_id", "run_id", "job_id", "node_id", "graph_node_id", "retry_after_ms"},
		},
		{
			name:     "partial",
			attrs:    errs.Attrs{EngagementID: "eng_01m1y2whfhgbz06ays6dxnvyws", JobID: "job_01m1y2whfhbt69j0h0fbxepw90"},
			wantKeys: []string{"engagement_id", "job_id"},
			reject:   []string{"run_id", "node_id", "graph_node_id", "retry_after_ms"},
		},
		{
			name: "graph node is never node id",
			attrs: errs.Attrs{
				EngagementID: "eng_01m1y2whfhgbz06ays6dxnvyws",
				RunID:        "run_01m1y2whfhnjx2am9103w0pnqw",
				JobID:        "job_01m1y2whfhbt69j0h0fbxepw90",
				GraphNodeID:  "gn_01m1y2whfhh039ykj5x8mc5a0g",
			},
			wantKeys: []string{"engagement_id", "run_id", "job_id", "graph_node_id"},
			reject:   []string{"node_id", "retry_after_ms", "node"},
		},
		{
			name:     "remote agent node is node id",
			attrs:    errs.Attrs{NodeID: "slp_node_01m1y2whfhxydsaem68cmazyc8"},
			wantKeys: []string{"node_id"},
			reject:   []string{"graph_node_id", "gn_id", "agent_node_id"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(errs.NewEnvelope(opOuter(), tt.attrs, 0))
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if strings.Contains(string(body), "null") {
				t.Fatalf("A0-8.3 forbids null in contract JSON, got %s", body)
			}
			inner := decodeErrorObject(t, body)
			for _, key := range tt.wantKeys {
				if _, ok := inner["attrs"].(map[string]any)[key]; !ok {
					t.Errorf("attrs missing %q in %s", key, body)
				}
			}
			for _, key := range tt.reject {
				if _, ok := inner["attrs"].(map[string]any)[key]; ok {
					t.Errorf("attrs carries absent key %q in %s", key, body)
				}
				if _, ok := inner[key]; ok && key != "attrs" {
					t.Errorf("error object carries absent key %q in %s", key, body)
				}
			}
			for _, key := range []string{"kind", "message", "attrs"} {
				if _, ok := inner[key]; !ok {
					t.Errorf("error object missing required key %q in %s", key, body)
				}
			}
		})
	}
}

func TestEnvelopeAlwaysCarriesAttrsObject(t *testing.T) {
	tests := []struct {
		name  string
		err   error
		attrs errs.Attrs
	}{
		{"platform error", opOuter(), errs.Attrs{}},
		{"foreign error", errTestForeign{"no ids here"}, errs.Attrs{}},
		{"nil error", nil, errs.Attrs{}},
		{"partial", originViaNew(), errs.Attrs{RunID: "run_01m1y2whfhnjx2am9103w0pnqw"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(errs.NewEnvelope(tt.err, tt.attrs, 0))
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if !strings.Contains(string(body), `"attrs":{`) {
				t.Fatalf(`A0-3.6 requires an always-present attrs object, got %s`, body)
			}
			if inner := decodeErrorObject(t, body); inner["attrs"] == nil {
				t.Fatalf("attrs decoded as nil: %s", body)
			}
		})
	}
}

// TestRetryAfterPresence is the contract-suite test of A0-3.10: retry_after_ms
// belongs to rate_limited only, and the >= 1 floor is its presence rule.
func TestRetryAfterPresence(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		retryAfter  int64
		wantPresent bool
		wantValue   int64
	}{
		{"rate limited with budget", errs.New(errs.RateLimited, "too many appends"), 5000, true, 5000},
		{"rate limited one ms", errs.New(errs.RateLimited, "too many appends"), 1, true, 1},
		{"rate limited zero", errs.New(errs.RateLimited, "too many appends"), 0, false, 0},
		{"rate limited negative", errs.New(errs.RateLimited, "too many appends"), -1000, false, 0},
		{"timeout is terminal for the field", errs.New(errs.Timeout, "deadline"), 5000, false, 0},
		{"upstream never carries it", errs.New(errs.Upstream, "bad gateway"), 5000, false, 0},
		{"validation never carries it", errs.New(errs.Validation, "bad id"), 5000, false, 0},
		{"internal never carries it", errs.New(errs.Internal, "defect"), 5000, false, 0},
		{"conflict never carries it", errs.New(errs.Conflict, "already consumed"), 5000, false, 0},
		{"foreign rate limit text never carries it", errTestForeign{"rate_limited?"}, 5000, false, 0},
		{"nil error never carries it", nil, 5000, false, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := errs.NewEnvelope(tt.err, errs.Attrs{}, tt.retryAfter)
			body, err := json.Marshal(env)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			inner := decodeErrorObject(t, body)
			raw, present := inner["retry_after_ms"]
			if present != tt.wantPresent {
				t.Fatalf("retry_after_ms present = %v, want %v (%s)", present, tt.wantPresent, body)
			}
			if !present {
				return
			}
			got, ok := raw.(float64)
			if !ok {
				t.Fatalf("retry_after_ms is %T, want a JSON number", raw)
			}
			if int64(got) != tt.wantValue {
				t.Errorf("retry_after_ms = %v, want %d", got, tt.wantValue)
			}
			if got < 1 {
				t.Errorf("A0-3.10 requires retry_after_ms >= 1 when present, got %v", got)
			}
			if env.Error.RetryAfterMS != tt.wantValue {
				t.Errorf("RetryAfterMS field = %d, want %d", env.Error.RetryAfterMS, tt.wantValue)
			}
		})
	}
}

func TestEnvelopeNilErrorIsInternal(t *testing.T) {
	env := errs.NewEnvelope(nil, errs.Attrs{EngagementID: "eng_01"}, 9999)
	if env.Error.Kind != errs.Internal {
		t.Errorf("kind = %q, want internal", env.Error.Kind)
	}
	if env.Error.Message == "" {
		t.Error("nil error must still produce a self-contained message")
	}
	if env.Error.RetryAfterMS != 0 {
		t.Errorf("retry_after_ms = %d, want 0 for a non-rate-limited envelope", env.Error.RetryAfterMS)
	}
	if env.Error.Kind.Status() != 500 {
		t.Errorf("status = %d, want 500", env.Error.Kind.Status())
	}
}

// TestEnvelopeKindAndMessageRidesTheError checks the envelope against the two
// worked examples in A0 §4.
func TestEnvelopeKindAndMessageRidesTheError(t *testing.T) {
	err := errs.New(errs.SummaryTooLarge, "graph.WriteNode: node summary exceeds cap for engagement=eng_01 graph_node=gn_01: field=summary cap=512 actual=613 bytes")
	env := errs.NewEnvelope(err, errs.Attrs{
		EngagementID: "eng_01m1y2whfhgbz06ays6dxnvyws",
		RunID:        "run_01m1y2whfhnjx2am9103w0pnqw",
		JobID:        "job_01m1y2whfhbt69j0h0fbxepw90",
		GraphNodeID:  "gn_01m1y2whfhh039ykj5x8mc5a0g",
	}, 0)
	body, marshalErr := json.Marshal(env)
	if marshalErr != nil {
		t.Fatalf("marshal: %v", marshalErr)
	}
	inner := decodeErrorObject(t, body)
	if inner["kind"] != "summary_too_large" {
		t.Errorf("kind = %v, want summary_too_large", inner["kind"])
	}
	if inner["message"] != err.Error() {
		t.Errorf("message = %v, want the rendered error chain", inner["message"])
	}
	if env.Error.Kind.Status() != 413 {
		t.Errorf("status = %d, want 413", env.Error.Kind.Status())
	}
}

func decodeErrorObject(t *testing.T, body []byte) map[string]any {
	t.Helper()
	var doc map[string]any
	if err := json.Unmarshal(body, &doc); err != nil {
		t.Fatalf("envelope is not valid JSON: %v (%s)", err, body)
	}
	inner, ok := doc["error"].(map[string]any)
	if !ok {
		t.Fatalf(`envelope has no top-level "error" object: %s`, body)
	}
	if _, ok := inner["attrs"].(map[string]any); !ok {
		t.Fatalf(`envelope has no "attrs" object: %s`, body)
	}
	return inner
}

// --- A0-3.7 / ADR-0019 §5 redaction ----------------------------------------

// secretValue is deliberately long, mixed-case and free of any substring that
// occurs in the literal "[redacted]" or in the fixed prose of these tests, so a
// prefix/suffix scan cannot false-positive.
const secretValue = "Zk9xWvR7tLpMbNqHgJdFfAzcEuiO-DZG-kn7r"

func carriers(t *testing.T) map[string]string {
	t.Helper()
	secret := errs.NewSecret(secretValue)
	rendered := map[string]string{
		"Newf %v":  errs.Newf(errs.Validation, "bind token %v", secret).Error(),
		"Newf %s":  errs.Newf(errs.Auth, "bind token %s", secret).Error(),
		"Newf %q":  errs.Newf(errs.Auth, "bind token %q", secret).Error(),
		"Newf %x":  errs.Newf(errs.Auth, "bind token %x", secret).Error(),
		"Newf %+v": errs.Newf(errs.Auth, "bind token %+v", secret).Error(),
		"Wrapf":    errs.Wrapf(errs.New(errs.Upstream, "upstream said "+errs.Redact(secretValue)), "relay %v", secret).Error(),
		"Redact":   errs.Newf(errs.Auth, "bind token %s", errs.Redact(secretValue)).Error(),
		"struct":   fmt.Sprintf("%+v", struct{ Token errs.Secret }{secret}),
	}
	for n, env := range []errs.Envelope{
		errs.NewEnvelope(errs.Wrapf(env0(errTestSentinel), "relay %v", secret), errs.Attrs{}, 0),
		errs.NewEnvelope(errs.Newf(errs.Auth, "bind token %v", secret), errs.Attrs{EngagementID: "eng_01"}, 5000),
	} {
		body, err := json.Marshal(env)
		if err != nil {
			t.Fatalf("marshal envelope: %v", err)
		}
		rendered[fmt.Sprintf("envelope %d", n)] = string(body)
	}
	return rendered
}

// env0 exists only so the envelope carriers hold a *errs.Error created one
// frame away from Wrapf.
func env0(err error) error { return errs.Wrap(err, "deliver webhook") }

// formatSecretInError renders an error that was built with a Secret as a
// formatter argument, from a named frame so the expected op is stable.
func formatSecretInError() string {
	return fmt.Sprintf("%s", errs.Newf(errs.Auth, "token %s", errs.NewSecret(secretValue)))
}

// TestNoSecretValueOrDigestInError is the A0 §4.1 contract-suite test for
// A0-3.4/A0-3.7: a rejected secret value must not appear in whole, in part or
// as a digest.
func TestNoSecretValueOrDigestInError(t *testing.T) {
	encodings := secretEncodings(secretValue)
	for name, text := range carriers(t) {
		assertNoSecretTrace(t, name, text, encodings)
	}
	// The scan itself must be able to fail: a control that does leak the value
	// has to be caught, or the test proves nothing.
	control := fmt.Sprintf("leaked: %s", secretValue)
	if leaked := detectSecretTrace(control, secretEncodings(secretValue)); len(leaked) == 0 {
		t.Fatal("secret-trace detector did not catch an obvious leak; it is vacuous")
	}
}

func assertNoSecretTrace(t *testing.T, name, text string, encodings []secretEncoding) {
	t.Helper()
	if bad := detectSecretTrace(text, encodings); len(bad) > 0 {
		t.Errorf("%s: carries secret trace %v: %q", name, bad, text)
	}
}

type secretEncoding struct {
	label string
	value string
}

// secretEncodings returns every form of value that must never reach an error:
// the value, its >= 6-char prefixes and suffixes, hex/base64 digests of it, and
// hex/base64 renderings of it.
func secretEncodings(value string) []secretEncoding {
	out := []secretEncoding{{"value", value}}
	for n := 6; n <= len(value); n++ {
		out = append(out,
			secretEncoding{"prefix", value[:n]},
			secretEncoding{"suffix", value[len(value)-n:]},
		)
	}
	sum := sha256.Sum256([]byte(value))
	h := hex.EncodeToString(sum[:])
	out = append(out,
		secretEncoding{"sha256-hex", h},
		secretEncoding{"sha256-hex-upper", strings.ToUpper(h)},
		secretEncoding{"sha256-b64-std", base64.StdEncoding.EncodeToString(sum[:])},
		secretEncoding{"sha256-b64-url", base64.RawURLEncoding.EncodeToString(sum[:])},
		secretEncoding{"value-hex", hex.EncodeToString([]byte(value))},
		secretEncoding{"value-b64", base64.StdEncoding.EncodeToString([]byte(value))},
		secretEncoding{"value-b64url", base64.RawURLEncoding.EncodeToString([]byte(value))},
	)
	for n := 8; n <= len(h); n += 8 {
		out = append(out, secretEncoding{"digest-prefix", h[:n]})
	}
	return out
}

func detectSecretTrace(text string, encodings []secretEncoding) []string {
	var bad []string
	for _, enc := range encodings {
		if strings.Contains(text, enc.value) {
			bad = append(bad, enc.label+"="+enc.value)
		}
	}
	return bad
}

func TestSecretNeverLeaksUnderAnyVerb(t *testing.T) {
	secret := errs.NewSecret(secretValue)
	other := errs.NewSecret("")
	tests := []struct {
		name string
		call func() string
		want string
	}{
		{"%v", func() string { return fmt.Sprintf("%v", secret) }, errs.Redacted},
		{"%s", func() string { return fmt.Sprintf("%s", secret) }, errs.Redacted},
		{"%q", func() string { return fmt.Sprintf("%q", secret) }, errs.Redacted},
		{"%x", func() string { return fmt.Sprintf("%x", secret) }, errs.Redacted},
		{"%X", func() string { return fmt.Sprintf("%X", secret) }, errs.Redacted},
		{"%+v", func() string { return fmt.Sprintf("%+v", secret) }, errs.Redacted},
		{"%#v", func() string { return fmt.Sprintf("%#v", secret) }, errs.Redacted},
		{"%d", func() string { return fmt.Sprintf("%d", secret) }, errs.Redacted},
		{"%T", func() string { return fmt.Sprintf("%T", secret) }, "errs.Secret"},
		{"ptr %v", func() string { return fmt.Sprintf("%v", &secret) }, errs.Redacted},
		{"ptr %s", func() string { return fmt.Sprintf("%s", &secret) }, errs.Redacted},
		{"nested %+v", func() string { return fmt.Sprintf("%+v", struct{ Token errs.Secret }{secret}) }, "{Token:" + errs.Redacted + "}"},
		{"nested %v", func() string { return fmt.Sprintf("%v", struct{ Token errs.Secret }{secret}) }, "{" + errs.Redacted + "}"},
		{"slice %v", func() string { return fmt.Sprintf("%v", []errs.Secret{secret, other}) }, "[" + errs.Redacted + " " + errs.Redacted + "]"},
		{"zero value %v", func() string { return fmt.Sprintf("%v", errs.Secret{}) }, errs.Redacted},
		{"in error %s", formatSecretInError, "errs_test.formatSecretInError: token " + errs.Redacted},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.call()
			if got != tt.want {
				t.Errorf("rendered %q, want %q", got, tt.want)
			}
			if tt.name != "%T" && strings.Contains(got, secretValue) {
				t.Errorf("verb %s leaked the secret: %q", tt.name, got)
			}
		})
	}
	// %p is printed by fmt before Formatter is consulted (badVerb path), so it
	// cannot be made to print the placeholder; it must at least carry no trace of
	// the value, which is why Secret stores it behind a closure.
	ptr := fmt.Sprintf("%p", secret)
	if bad := detectSecretTrace(ptr, secretEncodings(secretValue)); len(bad) > 0 {
		t.Errorf("%%p leaked the secret (%v): %q", bad, ptr)
	}
	t.Logf("%%p renders as %q", ptr)

	jsonBody, err := json.Marshal(secret)
	if err != nil {
		t.Fatalf("json.Marshal(Secret): %v", err)
	}
	if string(jsonBody) != `"`+errs.Redacted+`"` {
		t.Errorf("json.Marshal = %s, want %q", jsonBody, errs.Redacted)
	}
	tm, ok := any(secret).(encoding.TextMarshaler)
	if !ok {
		t.Fatal("Secret does not implement encoding.TextMarshaler")
	}
	text, err := tm.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText: %v", err)
	}
	if string(text) != errs.Redacted {
		t.Errorf("MarshalText = %q, want %q", text, errs.Redacted)
	}
	if gs := fmt.Sprintf("%v", secret.String()); gs != errs.Redacted {
		t.Errorf("String() = %q, want %q", gs, errs.Redacted)
	}
	// Each defence must hold on its own, not only while fmt prefers Formatter.
	if secret.GoString() != errs.Redacted {
		t.Errorf("GoString() = %q, want %q", secret.GoString(), errs.Redacted)
	}
	if zero := (errs.Secret{}); zero.Reveal() != "" || zero.GoString() != errs.Redacted || zero.String() != errs.Redacted {
		t.Errorf("zero Secret is not inert: Reveal=%q String=%q GoString=%q", zero.Reveal(), zero.String(), zero.GoString())
	}
	if _, err := secret.MarshalJSON(); err != nil {
		t.Errorf("MarshalJSON: %v", err)
	}
	if secret.Reveal() != secretValue {
		t.Errorf("Reveal() = %q, want the value passed to NewSecret", secret.Reveal())
	}
}

func TestRedactIsConstant(t *testing.T) {
	if errs.Redacted != "[redacted]" {
		t.Errorf(`Redacted = %q, want "[redacted]"`, errs.Redacted)
	}
	tests := []struct {
		name  string
		value string
	}{
		{"empty", ""},
		{"one byte", "a"},
		{"credential", "ghp_" + strings.Repeat("A", 36)},
		{"jwt-like", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U"},
		{"invalid utf8", "\xff\xfe\x00binary"},
		{"already redacted", errs.Redacted},
		{"long", strings.Repeat("0123456789", 1024)},
		{"newline", "line1\nline2\tsecret"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := errs.Redact(tt.value); got != errs.Redacted {
				t.Errorf("Redact() = %q, want the constant %q", got, errs.Redacted)
			}
			if len(errs.Redact(tt.value)) != len(errs.Redacted) {
				t.Error("Redact output length varies with input, which discloses length")
			}
		})
	}
}
