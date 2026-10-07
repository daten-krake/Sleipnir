package events

// ScopeDeniedPayload is the payload of kind scope_denied (A1-3.3, ADR-0005,
// C5): a denied action is auditable. ActionKind is the closed enum
// tool_exec | spawn | graph_read | llm_call | api. AttemptedTarget is the
// target as requested — untrusted, never normalized into scope vocabulary
// (A1-4.2).
type ScopeDeniedPayload struct {
	ActionKind      string `json:"action_kind"`      // enum tool_exec|spawn|graph_read|llm_call|api
	AttemptedTarget string `json:"attempted_target"` // ≤ 256 B, untrusted (A1-4.4/4.5)
	ToolID          string `json:"tool_id"`
	RequestEventID  string `json:"request_event_id"`
}

// Kind returns KindScopeDenied.
func (ScopeDeniedPayload) Kind() Kind { return KindScopeDenied }

// BlacklistDeniedPayload is the payload of kind blacklist_denied (A1-3.3,
// C5): blacklist beats allowlist beats approval, separately filterable by
// design (A1-8.3). BlacklistEntry must be the matched entry.
type BlacklistDeniedPayload struct {
	ActionKind      string `json:"action_kind"`      // enum tool_exec|spawn|graph_read|llm_call|api
	AttemptedTarget string `json:"attempted_target"` // ≤ 256 B, untrusted (A1-4.4/4.5)
	BlacklistEntry  string `json:"blacklist_entry"`  // ≤ 256 B, the matched entry (A1-4.2)
	ToolID          string `json:"tool_id"`
	RequestEventID  string `json:"request_event_id"`
}

// Kind returns KindBlacklistDenied.
func (BlacklistDeniedPayload) Kind() Kind { return KindBlacklistDenied }

// ActionBlockedPayload is the payload of kind action_blocked (A1-3.3,
// A1-3.5): the designated home for enforcement denials that do not deserve
// their own filter. Reason and ActionKind must both be set and are closed
// enums (A1-4.2); Detail carries human prose and is never parsed (A0-3.4).
type ActionBlockedPayload struct {
	Reason          string `json:"reason"`           // enum approval_consumed|approval_expired_at_exec|approval_metadata_mismatch|fingerprint_mismatch|target_quarantined|image_not_allowed|quota_exceeded|hard_stop_active|node_not_paired|llm_egress_blocked|secret_excluded|graph_write_rejected|append_not_permitted|append_rejected|unknown_event_kind|token_revoked
	ActionKind      string `json:"action_kind"`      // enum tool_exec|spawn|graph_read|graph_write|llm_call|api|events_append
	AttemptedTarget string `json:"attempted_target"` // ≤ 256 B, untrusted (A1-4.4/4.5)
	Detail          string `json:"detail"`           // ≤ 512 B, untrusted; human prose, never parsed
	ToolID          string `json:"tool_id"`
	ApprovalID      string `json:"approval_id"`
	RequestEventID  string `json:"request_event_id"`
}

// Kind returns KindActionBlocked.
func (ActionBlockedPayload) Kind() Kind { return KindActionBlocked }

// AgentErrorPayload is the payload of kind agent_error (A1-3.3, ADR-0012
// §2, ADR-0019 §2–§3, C9). ErrorKind is an A0-3.1 kind; Origin is
// component.Function; Message is already redacted (A0-3.7, A1-4.9) and is
// untrusted prose (A1-4.4).
type AgentErrorPayload struct {
	ErrorKind string `json:"error_kind"` // A0-3.1 kind
	Origin    string `json:"origin"`     // ≤ 128 B, component.Function (ADR-0019 §2–§3)
	Message   string `json:"message"`    // ≤ 512 B, untrusted, already redacted (A1-4.4/4.9)
	Retryable bool   `json:"retryable"`
}

// Kind returns KindAgentError.
func (AgentErrorPayload) Kind() Kind { return KindAgentError }
