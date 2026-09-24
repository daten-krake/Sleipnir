package errs

// Envelope is the single top-level error object of the /api/v1 JSON face
// (A0-3.6). The HTMX face MUST NOT emit it: it renders Message through
// html/template and branches on Kind server-side (A0-3.10, ADR-0011).
type Envelope struct {
	Error ErrorBody `json:"error"`
}

// ErrorBody is the content of an Envelope. Message is human prose and MUST NOT
// be parsed programmatically, matched by substring, or used as a control input
// (A0-3.4); Kind is the contract.
type ErrorBody struct {
	Kind    Kind   `json:"kind"`
	Message string `json:"message"`
	// Attrs is always present as an object, empty when no correlation id is
	// known at the failure point (A0-3.6).
	Attrs Attrs `json:"attrs"`
	// RetryAfterMS is present for RateLimited only. A0-3.10's "≥ 1" floor
	// is what makes omitempty safe here: a rate_limited envelope without it is a
	// platform defect, never a legitimate absent value.
	RetryAfterMS int64 `json:"retry_after_ms,omitempty"`
}

// Attrs are the correlation ids known at the failure point (ADR-0019 §3). Each
// is absent when unknown, never null (A0-8.3), so every field carries
// omitempty.
//
// NodeID denotes the remote agent node (slp_node_, Q9); a graph node is
// GraphNodeID (gn_). The two MUST NOT be conflated (A0-3.6) — ADR-0019 §3 lists
// the attribute as "node", which predates the graph vocabulary of ADR-0016.
type Attrs struct {
	EngagementID string `json:"engagement_id,omitempty"`
	RunID        string `json:"run_id,omitempty"`
	JobID        string `json:"job_id,omitempty"`
	NodeID       string `json:"node_id,omitempty"`
	GraphNodeID  string `json:"graph_node_id,omitempty"`
}

// NewEnvelope builds the A0-3.6 response body for err, which the calling
// handler is about to log exactly once (A0-3.8: log-or-return, never both — the
// handler that converts an error into a response is the single place that logs
// it).
//
// The kind is normalized by KindOf, so an unrecognized kind reaches a client as
// internal (A0-3.3). retryAfterMS is honoured only for RateLimited and only
// when it is at least 1 (A0-3.10); for any other kind it is dropped, because
// retry_after_ms on a terminal error would invite a client to retry a request
// that A0-3.11 declares terminal. A nil err yields an internal envelope rather
// than a panic: reaching this point with no error is a platform defect, and a
// defect in the error path must not take the request handler with it.
func NewEnvelope(err error, attrs Attrs, retryAfterMS int64) Envelope {
	if err == nil {
		return Envelope{Error: ErrorBody{
			Kind:    Internal,
			Message: "errs.NewEnvelope: no error supplied: platform defect in the error path",
			Attrs:   attrs,
		}}
	}
	body := ErrorBody{Kind: KindOf(err), Message: err.Error(), Attrs: attrs}
	if body.Kind == RateLimited && retryAfterMS >= 1 {
		body.RetryAfterMS = retryAfterMS
	}
	return Envelope{Error: body}
}
