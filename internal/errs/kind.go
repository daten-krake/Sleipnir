package errs

import "net/http"

// Kind is the machine-readable error classification of A0-3.1. Clients MUST
// switch on Kind, never on the HTTP status (the mapping is many-to-one) and
// never on the message (A0-3.4). The v1 list is closed; adding a kind is
// additive-only and a kind MUST NOT be renamed or repurposed (A0-3.12), because
// agent behaviour is keyed to it.
type Kind string

// The closed A0-3.1 kind vocabulary. Spelling follows A0-8.5 (closed enum:
// lowercase snake_case); the HTTP status and retryability of each are in Status
// and Retryable.
const (
	// Validation is malformed input: a bad id (A0-1.5), an unknown field on
	// write (A0-6.2), a bad enum (A0-6.3), a bad timestamp (A0-5.3) or a
	// canonicalization failure (A0-2). Hard reject per Q3.
	Validation Kind = "validation"
	// Auth is a missing, expired, malformed or unverifiable credential (A5).
	Auth Kind = "auth"
	// Forbidden is a principal-level denial that names no object: a machine
	// principal on a Q6-excluded endpoint, or a missing verb scope. It is
	// distinct from Auth so an agent stops instead of re-authenticating.
	Forbidden Kind = "forbidden"
	// NotFound is an absent object, or one outside the caller's engagement
	// scope: a foreign-engagement id resolves here, never to Forbidden, so the
	// existence of another engagement's objects is not disclosed (A0-3.9,
	// SPEC C8).
	NotFound Kind = "notfound"
	// Conflict is a state that prevents the request: a single-use approval
	// already consumed (Q10), a duplicate write, an immutable field, or a
	// spawn-quota rejection.
	Conflict Kind = "conflict"
	// ApprovalRequired is an execution or spawn attempted with no valid
	// approval for the fingerprint (ADR-0005 §1, ADR-0018). It is 409 rather
	// than 403 so an agent can tell "never allowed" from "not yet allowed".
	ApprovalRequired Kind = "approval_required"
	// ApprovalExpired is an approval past its window (ADR-0012 §7, SPEC §5.6);
	// the orchestrator replans instead of waiting.
	ApprovalExpired Kind = "approval_expired"
	// IntegrityFailed is a hash-chain verification failure; exports stay blocked
	// until an operator override is logged as an event (Q11, ADR-0021). It is not
	// 5xx, because 5xx would be retried and alerted as a platform defect.
	IntegrityFailed Kind = "integrity_failed"
	// SummaryTooLarge is a capped field or count that exceeded a budget declared
	// by its owning contract under mechanism R (A0-7.6). It is distinguishable
	// from Validation because the remedy is "send fewer bytes", not "fix a
	// malformed field".
	SummaryTooLarge Kind = "summary_too_large"
	// Timeout is a platform-side deadline that expired (own database, LLM
	// gateway, spawn broker). Retryable.
	Timeout Kind = "timeout"
	// Upstream is an upstream (LLM endpoint, webhook target, container runtime)
	// that answered with an error or unusable bytes. Retryable.
	Upstream Kind = "upstream"
	// RateLimited is the platform rate limiter (ADR-0011). Retryable, and the
	// envelope MUST carry retry_after_ms (A0-3.10).
	RateLimited Kind = "rate_limited"
	// Internal is a platform defect or an unclassified failure. A client that
	// receives a kind it does not know MUST treat it as Internal (A0-3.3).
	Internal Kind = "internal"
)

// Status maps a Kind to its HTTP status code, per the A0-3.1 table. The mapping
// is many-to-one, so Status MUST NOT be used to recover the kind. An unknown
// kind maps to 500, which is what A0-3.3 requires of a client that cannot
// classify what it received.
func (k Kind) Status() int {
	switch k {
	case Validation:
		return http.StatusBadRequest
	case Auth:
		return http.StatusUnauthorized
	case Forbidden:
		return http.StatusForbidden
	case NotFound:
		return http.StatusNotFound
	case Conflict, ApprovalRequired, ApprovalExpired, IntegrityFailed:
		return http.StatusConflict
	case SummaryTooLarge:
		return http.StatusRequestEntityTooLarge
	case RateLimited:
		return http.StatusTooManyRequests
	case Timeout:
		return http.StatusGatewayTimeout
	case Upstream:
		return http.StatusBadGateway
	case Internal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// Retryable reports whether a request that failed with k may be retried, per
// A0-3.11: timeout, upstream and rate_limited are retried with bounded
// exponential backoff plus jitter, honouring retry_after_ms; every other kind
// is terminal for the same request. Retrying a non-idempotent write is only safe
// where the owning contract defines a deduplication key (A1-7.6).
func (k Kind) Retryable() bool {
	switch k {
	case Timeout, Upstream, RateLimited:
		return true
	default:
		return false
	}
}

// known normalizes k to a member of the closed A0-3.1 vocabulary, mapping
// anything else to Internal (A0-3.3). It keeps a misspelled or future kind from
// reaching a client as an unclassifiable value.
func (k Kind) known() Kind {
	switch k {
	case Validation, Auth, Forbidden, NotFound, Conflict, ApprovalRequired,
		ApprovalExpired, IntegrityFailed, SummaryTooLarge, Timeout, Upstream,
		RateLimited, Internal:
		return k
	default:
		return Internal
	}
}
