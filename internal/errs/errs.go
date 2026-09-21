package errs

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

// Error is the platform error type. It carries the kind that makes handling
// programmatic (A0-3.1), the operation that produced it, captured automatically
// (ADR-0019 §1), a human-readable message that names what was attempted and the
// key identifiers (ADR-0019 §2), and the cause, if any.
//
// Error values are written once: New and Newf attach the kind at construction,
// Wrap and Wrapf inherit it from the cause, and nothing in this package mutates
// a field afterwards, so no layer can silently reclassify a failure (A0-3.12).
// The fields are exported because a handler reaches an error through errors.As
// and needs the kind, the op and the message from it.
//
// That discipline is not enforceable from inside the package: an exported field
// can be written by any caller, so a hand-built cycle (e.Err = e) is reachable.
// Error() therefore bounds its walk over the cause chain (maxChainDepth, the
// platform's A0-2.11 depth bound) and says so when it stops early, instead of
// spinning on a path that runs for every log record and every error envelope.
type Error struct {
	// Kind is the A0-3.1 classification, always a member of the closed
	// vocabulary.
	Kind Kind
	// Op is the originating operation as component.Function, captured via
	// runtime.Caller at construction. It is never a stack trace (A0-3.5).
	Op string
	// Msg is what was attempted plus the key identifiers. It is prose for
	// humans and MUST NOT be parsed or matched by substring (A0-3.4).
	Msg string
	// Err is the wrapped cause, or nil for a leaf error. errors.Is and
	// errors.As see through it.
	Err error
}

// Error renders the ADR-0019 §2 shape by walking the cause chain:
//
//	component.Function: what was attempted: key identifiers: cause
//
// Each wrapping layer contributes its own "Op: Msg" segment, so the rendered
// string doubles as an execution trace from the point of failure outward. The
// final segment is the cause's own message when the chain ends in an error this
// package did not create. An empty segment contributes nothing, so an error
// built by hand without an op never renders a leading ": ".
//
// The walk covers at most maxChainDepth layers of *Error, so a cause chain that
// is longer or cyclic — reachable, because the fields are exported and any
// caller can write Err — stops at the cap and ends with a segment naming the
// truncation instead of silently running out of stack or never returning. A
// chain that ends within the cap, including one whose last cause is a foreign
// error, renders in full with no notice.
func (e *Error) Error() string {
	var b strings.Builder
	for cur, depth := error(e), 0; cur != nil; {
		ee, ok := cur.(*Error)
		if !ok {
			b.WriteString(cur.Error())
			break
		}
		if depth == maxChainDepth {
			// Every iteration that continues has either written nothing or ended
			// with a separator, so the notice needs no ": " of its own.
			b.WriteString(truncationNotice())
			break
		}
		depth++
		if ee.Op != "" {
			b.WriteString(ee.Op)
			b.WriteString(": ")
		}
		b.WriteString(ee.Msg)
		if ee.Err == nil {
			break
		}
		if ee.Msg != "" {
			b.WriteString(": ")
		}
		cur = ee.Err
	}
	return b.String()
}

// maxChainDepth is the bound on Error()'s walk over a cause chain, set to the
// platform's single nesting-depth constant, A0-2.11 MaxDepth (32), so one number
// means the same thing everywhere it is used as an untrusted-structure bound.
// The cause chain is not untrusted input in the A0-2.11 sense — it is built by
// this package — but its fields are exported and mutable, so a cyclic chain is
// constructible by any caller, and Error() sits on the hot path of every log
// record and every /api/v1 error envelope.
const maxChainDepth = 32

// truncationNotice is the final segment Error() appends when it stops at the
// depth cap. The count is derived from maxChainDepth so the rendered number can
// never disagree with the cap actually applied.
func truncationNotice() string {
	return fmt.Sprintf("chain truncated at %d layers", maxChainDepth)
}

// Unwrap returns the cause, keeping errors.Is and errors.As intact through
// every layer (ADR-0019 §1).
func (e *Error) Unwrap() error { return e.Err }

// New returns an error of the given kind, tagged with the calling function.
// The kind is attached at creation (ADR-0019 §1) and cannot be changed later.
//
// msg is prose for humans: name what was attempted and the key identifiers
// (engagement, run, job, node, target, tool). It MUST NOT contain secret
// material (A0-3.7) — use Secret or Redact for any value that might be a
// credential — and MUST NOT contain a stack trace, SQL text, a request body or
// an upstream response body (A0-3.5).
func New(kind Kind, msg string) error {
	return &Error{Kind: kind.known(), Op: callerOp(), Msg: msg}
}

// Newf is New with a formatted message. The same rules about secret material
// and message content apply; a formatted value is still a value, so passing a
// secret as an argument leaks it just as directly.
func Newf(kind Kind, format string, args ...any) error {
	return &Error{Kind: kind.known(), Op: callerOp(), Msg: fmt.Sprintf(format, args...)}
}

// Wrap adds the calling function and what it was doing to err, preserving the
// cause chain. The kind is inherited from err, so wrapping never reclassifies a
// failure; a cause that carries no platform kind yields Internal.
//
// Wrap of a nil error returns nil, which makes the idiomatic
// "if err != nil { return errs.Wrap(err, ...) }" safe at every layer.
func Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}
	return &Error{Kind: KindOf(err), Op: callerOp(), Msg: msg, Err: err}
}

// Wrapf is Wrap with a formatted message.
func Wrapf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return &Error{Kind: KindOf(err), Op: callerOp(), Msg: fmt.Sprintf(format, args...), Err: err}
}

// KindOf returns the kind a handler must act on: the outermost platform error's
// kind, normalized to the closed A0-3.1 vocabulary. A foreign error, or a nil
// error, is Internal — which is also what A0-3.3 requires a client to assume
// for a kind it does not recognize.
func KindOf(err error) Kind {
	var e *Error
	if errors.As(err, &e) {
		return e.Kind.known()
	}
	return Internal
}

// OpOf returns the operation that produced err — the outermost platform error's
// component.Function, or "" when err carries none. It exists so a layer that
// logs an error record can name the origin without importing this package's
// error type into the logging package (DESIGN §1: logging has zero internal
// imports).
func OpOf(err error) string {
	var e *Error
	if errors.As(err, &e) {
		return e.Op
	}
	return ""
}

// callerOp returns the component.Function of the function that called an errs
// constructor. Frames count from here: 0 is callerOp, 1 is New/Newf/Wrap/Wrapf,
// 2 is the caller whose name ADR-0019 §1 requires. It therefore has no skip
// parameter and MUST be called directly from a constructor: an intermediate
// helper shifts the frame by one and mislabels the origin of every error in the
// platform.
func callerOp() string {
	pc, _, _, ok := runtime.Caller(2)
	if !ok {
		return "unknown"
	}
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown"
	}
	name := fn.Name()
	// Keep component.Function: drop the import path, which ends at the last '/'.
	if i := strings.LastIndexByte(name, '/'); i >= 0 {
		name = name[i+1:]
	}
	return name
}
