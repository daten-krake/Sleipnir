// Package logging_test exercises the foundation logging primitives: the JSON
// setup, subsystem loggers, the A0-3.6 correlation attributes and the single
// error-record shape of ADR-0019 §3. Two tests additionally police the package
// boundary itself (no internal imports, no package-level state) by parsing the
// package directory, so a future violation fails a build rather than waiting for
// a reviewer.
package logging_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/daten-krake/sleipnir/internal/logging"
)

const (
	// pkgImportPath is the import path of the package under test. It is the one
	// internal path this directory may import: an external test package has to
	// import the package it tests.
	pkgImportPath = "github.com/daten-krake/sleipnir/internal/logging"
	// internalPrefix is the forbidden prefix for the package under test.
	internalPrefix = "github.com/daten-krake/sleipnir/internal/"
	// pkgDir is the package directory; a test binary runs with its own package
	// directory as the working directory.
	pkgDir = "."
)

func TestNewEmitsJSON(t *testing.T) {
	var buf bytes.Buffer
	logging.New(&buf, slog.LevelInfo).Info("gateway reachable")

	rec := decodeRecord(t, &buf)
	for _, key := range []string{"time", "level", "msg"} {
		got, ok := rec[key]
		if !ok {
			t.Errorf("record %s has no %q key", buf.String(), key)
			continue
		}
		if s, ok := got.(string); !ok || s == "" {
			t.Errorf("key %q = %v, want a non-empty string", key, got)
		}
	}
	if got, want := rec["level"], "INFO"; got != want {
		t.Errorf("level = %v, want %v", got, want)
	}
	// AddSource is false by contract: the origin function travels in the error
	// message (ADR-0019 §2) and A0-3.5 forbids stack-trace material.
	for _, forbidden := range []string{"source", "pc", "stack", "trace"} {
		if v, ok := rec[forbidden]; ok {
			t.Errorf("record carries %s=%v; New must not add source information", forbidden, v)
		}
	}
	if lines := strings.Count(buf.String(), "\n"); lines != 1 {
		t.Errorf("one log call wrote %d lines, want 1: %q", lines, buf.String())
	}
}

func TestNewHonoursLevel(t *testing.T) {
	tests := []struct {
		name    string
		handler slog.Level
		record  slog.Level
		want    bool
	}{
		{"debug suppressed by info", slog.LevelInfo, slog.LevelDebug, false},
		{"info emitted at info", slog.LevelInfo, slog.LevelInfo, true},
		{"info suppressed by warn", slog.LevelWarn, slog.LevelInfo, false},
		{"warn emitted at warn", slog.LevelWarn, slog.LevelWarn, true},
		{"warn suppressed by error", slog.LevelError, slog.LevelWarn, false},
		{"error emitted at error", slog.LevelError, slog.LevelError, true},
		{"everything emitted at debug", slog.LevelDebug, slog.LevelInfo, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			logging.New(&buf, tc.handler).Log(context.Background(), tc.record, "probe")
			if got := buf.Len() > 0; got != tc.want {
				t.Fatalf("emitted %v (%q), want %v", got, buf.String(), tc.want)
			}
		})
	}
}

func TestSubsystemCarriesName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"dotted platform subsystem", "platform.policy"},
		{"plain subsystem", "broker"},
		{"api edge", "api"},
		{"notification service", "notify"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			lg := logging.Subsystem(logging.New(&buf, slog.LevelInfo), tc.want)
			lg.Info("checked scope")

			if got := decodeRecord(t, &buf)["subsystem"]; got != tc.want {
				t.Errorf("subsystem = %v, want %q", got, tc.want)
			}
		})
	}
}

func TestSubsystemEmptyNameReturnsBase(t *testing.T) {
	var buf bytes.Buffer
	base := logging.New(&buf, slog.LevelInfo)

	got := logging.Subsystem(base, "")
	if got != base {
		t.Errorf("Subsystem(base, \"\") returned a new logger %p, want base %p", got, base)
	}
	got.Info("no label")
	if _, ok := decodeRecord(t, &buf)["subsystem"]; ok {
		t.Error(`record carries a "subsystem" key; an empty name must add no attribute`)
	}
}

func TestCorrelationOmitsUnknown(t *testing.T) {
	tests := []struct {
		name         string
		engagementID string
		runID        string
		jobID        string
		nodeID       string
		wantKeys     []string
	}{
		{name: "all unknown returns nil"},
		{
			name:         "engagement only",
			engagementID: "eng_01m1y2whfhgbz06ays6dxnvyws",
			wantKeys:     []string{"engagement_id"},
		},
		{
			name:  "run and job",
			runID: "run_01m1y2whfh0123456789abcdef",
			jobID: "job_01m1y2whfhbt69j0h0fbxepw90",
			wantKeys: []string{
				"run_id",
				"job_id",
			},
		},
		{
			name:     "agent node only",
			nodeID:   "slp_node_01m1y2whfh9x2b4c7d1e8f0a3b",
			wantKeys: []string{"node_id"},
		},
		{
			name:         "all four known",
			engagementID: "eng_a",
			runID:        "run_b",
			jobID:        "job_c",
			nodeID:       "slp_node_d",
			wantKeys:     []string{"engagement_id", "run_id", "job_id", "node_id"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := logging.Correlation(tc.engagementID, tc.runID, tc.jobID, tc.nodeID)
			if tc.wantKeys == nil {
				if got != nil {
					t.Fatalf("Correlation with no known ids = %#v, want nil", got)
				}
				return
			}
			if keys := attrKeys(t, got); !reflect.DeepEqual(keys, tc.wantKeys) {
				t.Errorf("keys = %v, want %v", keys, tc.wantKeys)
			}
		})
	}
}

func TestCorrelationKeySpelling(t *testing.T) {
	// A0-3.6 fixes these spellings: node_id is the remote agent node and a graph
	// node is graph_node_id, which this package must never emit.
	attrs := logging.Correlation("eng_a", "run_b", "job_c", "slp_node_d")

	want := []string{"engagement_id", "run_id", "job_id", "node_id"}
	keys := attrKeys(t, attrs)
	if !reflect.DeepEqual(keys, want) {
		t.Fatalf("keys = %v, want exactly %v in that order", keys, want)
	}
	wantValues := []string{"eng_a", "run_b", "job_c", "slp_node_d"}
	if got := attrValues(t, attrs); !reflect.DeepEqual(got, wantValues) {
		t.Errorf("values = %v, want %v", got, wantValues)
	}
	for _, forbidden := range []string{"node", "engagement", "run", "job", "gn", "graph_node", "graph_node_id"} {
		if slices.Contains(keys, forbidden) {
			t.Errorf("correlation key %q appears exactly (ADR-0019 §3's pre-graph wording): %v", forbidden, keys)
		}
	}
	for _, key := range keys {
		if strings.HasPrefix(key, "graph") {
			t.Errorf("correlation key %q names a graph node; A0-3.6 reserves node_id for the remote agent node", key)
		}
	}
}

func TestErrorRecordCarriesOpAndChain(t *testing.T) {
	// A two-layer chain built locally: logging cannot import internal/errs, so
	// the caller's rendered text is what the record must carry verbatim.
	cause := errors.New("pq: duplicate key value violates unique constraint")
	middle := fmt.Errorf("store.InsertEvent: insert event for engagement=eng_a run=run_b: %w", cause)
	outer := fmt.Errorf("policy.Check: evaluate scope decision: %w", middle)

	tests := []struct {
		name      string
		op        string
		err       error
		attrs     []any
		wantMsg   string
		wantNoKey []string
	}{
		{
			name:    "op, chain and correlation ids",
			op:      "platform.policy.Check",
			err:     outer,
			attrs:   logging.Correlation("eng_a", "run_b", "", ""),
			wantMsg: "operation failed",
		},
		{
			name:      "empty op is omitted",
			err:       outer,
			attrs:     logging.Correlation("eng_a", "", "", ""),
			wantMsg:   "operation failed",
			wantNoKey: []string{"op", "run_id", "job_id", "node_id"},
		},
		{
			name:    "caller msg attribute becomes the message",
			op:      "broker.Spawn",
			err:     outer,
			attrs:   append([]any{slog.String("msg", "spawn refused")}, logging.Correlation("", "", "job_c", "")...),
			wantMsg: "spawn refused",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			lg := logging.Subsystem(logging.New(&buf, slog.LevelInfo), "broker")
			logging.ErrorRecord(context.Background(), lg, tc.op, tc.err, tc.attrs...)

			out := buf.String()
			rec := decodeRecord(t, &buf)
			if got, want := rec["level"], "ERROR"; got != want {
				t.Errorf("level = %v, want %v", got, want)
			}
			if got := rec["msg"]; got != tc.wantMsg {
				t.Errorf("msg = %v, want %q", got, tc.wantMsg)
			}
			if strings.Count(out, `"msg"`) != 1 {
				t.Errorf(`record has more than one "msg" key: %s`, out)
			}
			text, ok := rec["error"].(string)
			if !ok {
				t.Fatalf(`error attribute = %v (%T), want the rendered chain as a string`, rec["error"], rec["error"])
			}
			for _, segment := range []string{
				"policy.Check: evaluate scope decision",
				"store.InsertEvent: insert event for engagement=eng_a run=run_b",
				"pq: duplicate key value violates unique constraint",
			} {
				if !strings.Contains(text, segment) {
					t.Errorf("error %q missing chain segment %q", text, segment)
				}
			}
			gotOp, hasOp := rec["op"]
			if hasOp != (tc.op != "") {
				t.Fatalf("record has op=%v for wanted op %q: %s", gotOp, tc.op, out)
			}
			if hasOp && gotOp != tc.op {
				t.Errorf("op = %v, want %q", gotOp, tc.op)
			}
			for _, key := range tc.wantNoKey {
				if _, ok := rec[key]; ok {
					t.Errorf("record carries %s=%v; %q must be absent (A0-8.3): %s", key, rec[key], key, out)
				}
			}
			for _, key := range []string{"engagement_id", "run_id", "job_id", "node_id"} {
				if v, ok := rec[key]; ok && v == nil {
					t.Errorf("%s is null; A0-8.3 requires absence, never null", key)
				}
			}
			if got, want := rec["subsystem"], "broker"; got != want {
				t.Errorf("subsystem = %v, want %q", got, want)
			}
			// op, then error, then the caller's attributes (the subsystem attribute
			// comes from the logger and precedes them all).
			assertKeyOrder(t, out, "op", "error", "engagement_id", "run_id")
		})
	}
}

func TestErrorRecordNilErrorLogsNothing(t *testing.T) {
	tests := []struct {
		name  string
		op    string
		attrs []any
	}{
		{"bare call", "", nil},
		{"with op and message", "api.GetRun", []any{slog.String("msg", "not an error")}},
		{"with correlation ids", "store.InsertEvent", logging.Correlation("eng_a", "run_b", "job_c", "slp_node_d")},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			lg := logging.Subsystem(logging.New(&buf, slog.LevelDebug), "api")

			logging.ErrorRecord(context.Background(), lg, tc.op, nil, tc.attrs...)

			if buf.Len() != 0 {
				t.Errorf("writer holds %q after a nil error, want nothing", buf.String())
			}
		})
	}
}

// TestErrorRecordNilLoggerFallsBackToDefault pins the nil-logger contract:
// ErrorRecord must neither panic nor drop the record. A panic would be error
// transport on a request or job path, which ADR-0019 §6 reserves for
// programmer-invariant violations at startup; dropping the record would break
// the audit premise of ADR-0019 §3. So the record goes to slog.Default(), which
// the stdlib guarantees is never nil. The default handler is process-global state
// this package does not own, so the test swaps it for its duration and restores
// it through t.Cleanup: no global state leaks between cases.
func TestErrorRecordNilLoggerFallsBackToDefault(t *testing.T) {
	original := slog.Default()
	t.Cleanup(func() { slog.SetDefault(original) })

	tests := []struct {
		name     string
		op       string
		err      error
		attrs    []any
		wantMsg  string
		wantKeys []string
		wantAny  bool
	}{
		{
			name:     "nil logger still writes the record",
			op:       "api.GetRun",
			err:      errors.New("api.handleGetRun: load run: id=run_b: context deadline exceeded"),
			attrs:    logging.Correlation("eng_a", "run_b", "job_c", ""),
			wantMsg:  "operation failed",
			wantKeys: []string{"engagement_id", "run_id", "job_id"},
			wantAny:  true,
		},
		{
			name: "caller msg attribute survives the fallback",
			op:   "broker.Spawn",
			err:  errors.New("broker.Spawn: create container: node=slp_node_d: runtime refused"),
			attrs: append([]any{slog.String("msg", "spawn refused")},
				logging.Correlation("", "", "", "slp_node_d")...),
			wantMsg:  "spawn refused",
			wantKeys: []string{"node_id"},
			wantAny:  true,
		},
		{
			name: "nil error with a nil logger still writes nothing",
			op:   "api.GetRun",
			err:  nil,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			slog.SetDefault(logging.New(&buf, slog.LevelInfo))

			// A panic is the bug under test; turn it into a failure report instead
			// of letting it take down the test binary.
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Fatalf("ErrorRecord panicked on a nil logger (ADR-0019 §6): %v", r)
					}
				}()
				logging.ErrorRecord(context.Background(), nil, tc.op, tc.err, tc.attrs...)
			}()

			if !tc.wantAny {
				if buf.Len() != 0 {
					t.Errorf("writer holds %q after a nil error, want nothing", buf.String())
				}
				return
			}

			out := buf.String()
			rec := decodeRecord(t, &buf)
			if got, want := rec["level"], "ERROR"; got != want {
				t.Errorf("level = %v, want %v", got, want)
			}
			if got := rec["msg"]; got != tc.wantMsg {
				t.Errorf("msg = %v, want %q at the default handler", got, tc.wantMsg)
			}
			if strings.Count(out, `"msg"`) != 1 {
				t.Errorf(`record has more than one "msg" key: %s`, out)
			}
			if got := rec["op"]; got != tc.op {
				t.Errorf("op = %v, want %q", got, tc.op)
			}
			if got := rec["error"]; got != tc.err.Error() {
				t.Errorf("error = %v, want the rendered chain %q", got, tc.err.Error())
			}
			for _, key := range tc.wantKeys {
				if _, ok := rec[key]; !ok {
					t.Errorf("record written via the fallback is missing %s: %s", key, out)
				}
			}
			for _, key := range []string{"run_id", "job_id", "node_id"} {
				if !slices.Contains(tc.wantKeys, key) {
					if _, ok := rec[key]; ok {
						t.Errorf("record carries an unexpected %s: %s", key, out)
					}
				}
			}
		})
	}
}

// secretValue stands in for a value whose text form must never reach a record
// (ADR-0019 §5, A0-3.7). It is defined here rather than imported from
// internal/errs, which this package must not depend on. The value is held behind
// a closure and String(), MarshalText() and MarshalJSON() all answer
// redactedText, because slog's JSONHandler serialises a KindAny value with
// encoding/json, which ignores fmt.Stringer: a redaction helper that implements
// only String() is not redacted in a JSON record.
type secretValue struct {
	reveal func() string
}

// redactedText is the placeholder this test expects in place of a secret.
const redactedText = "[redacted]"

// String implements [fmt.Stringer].
func (s secretValue) String() string { return redactedText }

// MarshalText implements [encoding.TextMarshaler].
func (s secretValue) MarshalText() ([]byte, error) { return []byte(redactedText), nil }

// MarshalJSON implements [encoding/json.Marshaler].
func (s secretValue) MarshalJSON() ([]byte, error) { return []byte(`"` + redactedText + `"`), nil }

func TestErrorRecordDoesNotStringifyAttrsLeakily(t *testing.T) {
	const secret = "hunter2-super-secret-token"
	secretAttr := secretValue{reveal: func() string { return secret }}
	if got := secretAttr.reveal(); got != secret {
		t.Fatalf("fixture is broken: reveal() = %q, want %q", got, secret)
	}
	if got := secretAttr.String(); got != redactedText {
		t.Fatalf("fixture is broken: String() = %q, want %q", got, redactedText)
	}
	attrTests := []struct {
		name   string
		attr   any
		want   any
		absent string
	}{
		{name: "string", attr: slog.String("endpoint", "https://llm.example/v1"), want: "https://llm.example/v1"},
		{name: "int stays a number", attr: slog.Int("status", 502), want: float64(502)},
		{name: "bool", attr: slog.Bool("retry", true), want: true},
		{name: "duration stays a number of nanoseconds", attr: slog.Duration("duration", 450*time.Millisecond), want: float64(450_000_000)},
		{
			name:   "stringer value is redacted by its own marshaler",
			attr:   slog.Any("target", secretAttr),
			want:   redactedText,
			absent: secret,
		},
	}
	for _, tc := range attrTests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			logging.ErrorRecord(context.Background(), logging.New(&buf, slog.LevelInfo), "llm.Call", errors.New("llm: endpoint returned 500"), tc.attr)

			out := buf.String()
			rec := decodeRecord(t, &buf)
			key := tc.attr.(slog.Attr).Key
			if got := rec[key]; got != tc.want {
				t.Errorf("attribute %s = %#v, want %#v (ErrorRecord must pass attributes through to slog untouched)", key, got, tc.want)
			}
			if tc.absent != "" && strings.Contains(out, tc.absent) {
				t.Errorf("record %q contains secret material %q", out, tc.absent)
			}
		})
	}
}

func TestLoggingHasNoInternalImports(t *testing.T) {
	tests := []struct {
		name  string
		files map[string]string // fixture sources; nil means the package directory
		want  []string          // offending import paths
	}{
		{
			name:  "the package under test imports nothing internal",
			files: nil,
			want:  nil,
		},
		{
			name: "a stdlib-only file is clean",
			files: map[string]string{"f.go": `package logging

import (
	"io"
	"log/slog"
)

func use(w io.Writer, l *slog.Logger) {}
`},
			want: nil,
		},
		{
			name: "importing a sibling internal package is a violation",
			files: map[string]string{"f.go": `package logging

import "github.com/daten-krake/sleipnir/internal/errs"

func origin(err error) string { return errs.OpOf(err) }
`},
			want: []string{"github.com/daten-krake/sleipnir/internal/errs"},
		},
		{
			name: "a violation hidden in an in-package test file is caught",
			files: map[string]string{"f_test.go": `package logging

import (
	"testing"

	"github.com/daten-krake/sleipnir/internal/store"
)

func TestUsesStore(t *testing.T) { _ = store.Dialed }
`},
			want: []string{"github.com/daten-krake/sleipnir/internal/store"},
		},
		{
			name: "importing another internal package from the external test package is a violation",
			files: map[string]string{"f_test.go": `package logging_test

import (
	"testing"

	"github.com/daten-krake/sleipnir/internal/errs"
)

func TestUsesErrs(t *testing.T) { _ = errs.Redacted }
`},
			want: []string{"github.com/daten-krake/sleipnir/internal/errs"},
		},
		{
			name: "the external test package may import the package it tests",
			files: map[string]string{"f_test.go": `package logging_test

import (
	"testing"

	"github.com/daten-krake/sleipnir/internal/logging"
)

func TestNewReturnsLogger(t *testing.T) { _ = logging.New(io.Discard) }
`},
			want: nil,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := pkgDir
			if tc.files != nil {
				dir = writeFixturePkg(t, tc.files)
			}
			if got := internalImportViolations(t, dir); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("internal import violations = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestNoPackageLevelMutableState(t *testing.T) {
	tests := []struct {
		name  string
		files map[string]string // fixture sources; nil means the package directory
		want  []string          // wanted file-scope var names
	}{
		{
			name:  "the package under test declares no file-scope var",
			files: nil,
			want:  nil,
		},
		{
			name: "const, type and func are fine",
			files: map[string]string{"f.go": `package logging

import "log/slog"

const keyOp = "op"

type label string

func use(l *slog.Logger) label { return "" }
`},
			want: nil,
		},
		{
			name: "a single file-scope var is caught",
			files: map[string]string{"f.go": `package logging

var defaultLogger *slog.Logger
`},
			want: []string{"defaultLogger"},
		},
		{
			name: "every name in a var block is caught",
			files: map[string]string{"f.go": `package logging

var (
	cache   = map[string]int{}
	warnings int
)
`},
			want: []string{"cache", "warnings"},
		},
		{
			name: "a var inside a function is not state",
			files: map[string]string{"f.go": `package logging

func count() int {
	n := 0
	var total int
	return n + total
}
`},
			want: nil,
		},
		{
			name: "a var in a test file is not production state but the file is still parsed",
			files: map[string]string{
				"f.go": `package logging

const keyOp = "op"
`,
				"f_test.go": `package logging_test

var fixtureName = "x"
`,
			},
			want: nil,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := pkgDir
			if tc.files != nil {
				dir = writeFixturePkg(t, tc.files)
			}
			if got := fileScopeVars(t, dir); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("file-scope vars = %v, want %v", got, tc.want)
			}
		})
	}
}

// internalImportViolations parses every Go file in dir, test files included, and
// returns the offending import paths in file order. pkgImportPath is not an
// offense: the external test package in the directory has to import the package
// it tests. Any other path under internalPrefix breaks the foundation rule
// (DESIGN §1: internal/logging imports nothing from internal/).
func internalImportViolations(t *testing.T, dir string) []string {
	t.Helper()
	var violations []string
	for _, f := range parsePackage(t, dir, true) {
		for _, path := range f.imports {
			if strings.HasPrefix(path, internalPrefix) && path != pkgImportPath {
				t.Logf("%s imports %s", f.path, path)
				violations = append(violations, path)
			}
		}
	}
	return violations
}

// fileScopeVars parses the non-test Go files in dir and returns the names of
// every file-scope var declaration. Package-level mutable state is forbidden
// (DESIGN §4): loggers come from New and are threaded by the caller.
func fileScopeVars(t *testing.T, dir string) []string {
	t.Helper()
	var names []string
	for _, f := range parsePackage(t, dir, false) {
		for _, name := range f.vars {
			t.Logf("%s declares file-scope var %s", f.path, name)
			names = append(names, name)
		}
	}
	return names
}

// goFile holds the facts the boundary tests need from one parsed source file.
type goFile struct {
	path    string
	imports []string
	vars    []string
}

// parsePackage parses the Go files of dir in name order, with or without test
// files, and reports the file's imports and file-scope var names.
func parsePackage(t *testing.T, dir string, withTests bool) []goFile {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("reading package directory %s: %v", dir, err)
	}
	fset := token.NewFileSet()
	var files []goFile
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		if !withTests && strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			t.Fatalf("parsing %s: %v", path, err)
		}
		files = append(files, goFile{path: path, imports: importPaths(file), vars: declVars(file)})
	}
	if len(files) == 0 {
		t.Fatalf("directory %s holds no Go files", dir)
	}
	return files
}

// importPaths returns the unquoted import paths of file, in source order.
func importPaths(file *ast.File) []string {
	var paths []string
	for _, spec := range file.Imports {
		path, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			continue
		}
		paths = append(paths, path)
	}
	return paths
}

// declVars returns the names declared by file-scope var declarations, including
// the blank identifier, in source order.
func declVars(file *ast.File) []string {
	var names []string
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.VAR {
			continue
		}
		for _, spec := range gen.Specs {
			for _, name := range spec.(*ast.ValueSpec).Names {
				names = append(names, name.Name)
			}
		}
	}
	return names
}

// writeFixturePkg writes files into a fresh directory under t.TempDir() and
// returns that directory.
func writeFixturePkg(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, source := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(source), 0o600); err != nil {
			t.Fatalf("writing fixture %s: %v", name, err)
		}
	}
	return dir
}

// decodeRecord decodes the single record written to buf.
func decodeRecord(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()
	if buf.Len() == 0 {
		t.Fatal("no record was emitted, want one")
	}
	var rec map[string]any
	if err := json.Unmarshal(buf.Bytes(), &rec); err != nil {
		t.Fatalf("decoding record %q as JSON: %v", buf.String(), err)
	}
	return rec
}

// attrKeys returns the keys of the values Correlation produced, in order.
func attrKeys(t *testing.T, attrs []any) []string {
	t.Helper()
	keys := make([]string, 0, len(attrs))
	for _, a := range attrs {
		attr, ok := a.(slog.Attr)
		if !ok {
			t.Fatalf("Correlation produced %T, want slog.Attr values", a)
		}
		keys = append(keys, attr.Key)
	}
	return keys
}

// attrValues returns the string values of Correlation's attributes, in order.
func attrValues(t *testing.T, attrs []any) []string {
	t.Helper()
	values := make([]string, 0, len(attrs))
	for _, a := range attrs {
		attr, ok := a.(slog.Attr)
		if !ok {
			t.Fatalf("Correlation produced %T, want slog.Attr values", a)
		}
		values = append(values, attr.Value.String())
	}
	return values
}

// assertKeyOrder checks that keys appear in a record in the given order, so the
// fixed attribute layout of ADR-0019 §3 (op, then error, then caller attributes)
// is a tested property rather than an accident of the handler.
func assertKeyOrder(t *testing.T, record string, keys ...string) {
	t.Helper()
	prev := -1
	for _, key := range keys {
		idx := strings.Index(record, `"`+key+`":`)
		if idx < 0 {
			continue
		}
		if idx < prev {
			t.Errorf("key %q precedes an earlier key in %s", key, record)
		}
		prev = idx
	}
}
