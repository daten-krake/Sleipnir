package logging

import (
	"context"
	"io"
	"log/slog"
)

// Keys of the attributes this package emits, and the message a record carries
// when the caller supplies none. The correlation spellings are A0-3.6's; see the
// package comment for why they differ from ADR-0019 §3's wording.
const (
	fallbackMessage = "operation failed"

	keySubsystem  = "subsystem"
	keyOp         = "op"
	keyError      = "error"
	keyMsg        = "msg"
	keyEngagement = "engagement_id"
	keyRun        = "run_id"
	keyJob        = "job_id"
	keyNode       = "node_id"
)

// New returns a logger that writes one JSON object per record to w, emitting
// records at or above level. It is the only logger constructor: there is no
// global logger and no package-level state, so every logger is created by a
// binary's main and threaded downward through its dependencies (DESIGN §4).
//
// Source location is never added (AddSource: false). The origin function belongs
// in the error message, which ADR-0019 §2 already carries as
// component.Function, and A0-3.5 forbids stack-trace material in a record;
// duplicating it as a source field would only invite drift between the two.
func New(w io.Writer, level slog.Level) *slog.Logger {
	return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level:     level,
		AddSource: false,
	}))
}

// Subsystem returns a logger that stamps every record with the subsystem
// attribute set to name (ADR-0019 §3). name is the full dotted subsystem name
// as ADR-0019 spells it — "platform.policy", "broker", "api", "notify" — and
// Subsystem does not compose names: a caller that wants a nested name passes the
// whole thing. An empty name adds no attribute and base is returned unchanged,
// so an optional subsystem label cannot produce subsystem="".
func Subsystem(base *slog.Logger, name string) *slog.Logger {
	if name == "" {
		return base
	}
	return base.With(slog.String(keySubsystem, name))
}

// Correlation returns the correlation attributes for the ids known at a call
// site, as key-value pairs suitable for a logger's variadic arguments or for
// ErrorRecord. Only the non-empty ids are returned, in the fixed order
// engagement_id, run_id, job_id, node_id; if all four are empty the result is
// nil (A0-8.3 — an unknown id is absent from the record, never null and never
// the empty string).
//
// nodeID is the remote agent node (slp_node_, Q9). There is no graph-node
// parameter: a graph node is graph_node_id (A0-3.6) and the handler that owns
// graph content adds that attribute itself, so the two ids cannot be conflated
// through this function.
func Correlation(engagementID, runID, jobID, nodeID string) []any {
	ids := [...]struct {
		key string
		val string
	}{
		{keyEngagement, engagementID},
		{keyRun, runID},
		{keyJob, jobID},
		{keyNode, nodeID},
	}
	var attrs []any
	for _, id := range ids {
		if id.val == "" {
			continue
		}
		attrs = append(attrs, slog.String(id.key, id.val))
	}
	return attrs
}

// ErrorRecord emits the one error-level record shape of ADR-0019 §3: the
// operation, the full error chain, and whatever attributes the caller supplies —
// normally Correlation plus the ADR-0019 §4 request context of a failed external
// call (endpoint, status, duration in ms, retry count; never a request or
// response body, which may carry secret material — A0-3.7).
//
// The record message is the caller's msg attribute when it passed one as
// slog.String("msg", …) — that attribute becomes the message and is not repeated
// as a second key — and otherwise the fixed text "operation failed", which keeps
// the message field free of the variable content a reader finds in op and error.
// op is omitted when empty. error is err.Error(), the rendered chain, so a single
// log line carries the execution trace ADR-0019 §2 promises.
//
// Nothing happens when err is nil: a nil error is not a failure, and an
// unconditional call site ("log what you decided, even when it went well") is a
// caller bug this function must not turn into a panic (ADR-0019 §6). ctx is
// passed to the handler so it can honour cancellation and tracing values
// (DESIGN §3).
//
// When lg is nil the record goes to slog.Default(), which the stdlib guarantees
// is never nil. ErrorRecord runs on request and job paths, where a nil logger
// can only mean that wiring was missed somewhere, and ADR-0019 §6 reserves
// panics for programmer-invariant violations at startup — not for error
// transport on a live path; dropping the record would break the audit premise
// the platform is built on (ADR-0019 §3). So it is neither: the record is
// written, and it appears on the default handler's output instead of the
// configured sink, which is how the wiring defect becomes visible. The fallback
// is a safety net, not a supported way to obtain a logger — every logger comes
// from New and is threaded downward by the caller.
//
// ErrorRecord is the logging half of log-or-return (ADR-0019 §3, A0-3.8). It
// returns nothing, so it cannot be mistaken for a wrap: call it only at the
// layer that handles or decides, and return the error onward from there. Inner
// layers wrap with errs and return without logging.
func ErrorRecord(ctx context.Context, lg *slog.Logger, op string, err error, attrs ...any) {
	if err == nil {
		return
	}
	if lg == nil {
		// Wiring defect, not a mode: see the doc comment. Never panic, never drop.
		lg = slog.Default()
	}
	msg, rest := recordMessage(attrs)
	args := make([]any, 0, len(rest)+2)
	if op != "" {
		args = append(args, slog.String(keyOp, op))
	}
	args = append(args, slog.String(keyError, err.Error()))
	args = append(args, rest...)
	lg.ErrorContext(ctx, msg, args...)
}

// recordMessage returns the message the caller asked for, or fallbackMessage, and
// the remaining attributes. The input slice is not modified.
func recordMessage(attrs []any) (string, []any) {
	msg := fallbackMessage
	rest := make([]any, 0, len(attrs))
	for _, a := range attrs {
		if attr, ok := a.(slog.Attr); ok && attr.Key == keyMsg {
			msg = attr.Value.String()
			continue
		}
		rest = append(rest, a)
	}
	return msg, rest
}
