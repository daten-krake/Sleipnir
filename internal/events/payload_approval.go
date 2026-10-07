package events

// ApprovalRequestedPayload is the payload of kind approval_requested
// (A1-3.3, ADR-0018 §1/§4, ADR-0012 §7). FingerprintHash is the A7 digest
// of the action; ActionSummary is prose in addition to the fingerprint,
// never instead of it, and must be composed only from platform vocabulary
// and the A7 action spec's own fields — copying model or tool prose into it
// is a laundering violation (A1-4.2, A1-4.4). ExpiresAt and
// UntrustedContext are platform-computed and must not be caller values;
// ExpiresAt is a timestamp string (A0-5.1, A1-4.12).
type ApprovalRequestedPayload struct {
	ApprovalID           string `json:"approval_id"`
	FingerprintHash      string `json:"fingerprint_hash"` // 64-hex (A0-2.15); A7 owns the content (A1-3.6)
	ActionSpecEvidenceID string `json:"action_spec_evidence_id"`
	TargetGraphNodeID    string `json:"target_graph_node_id"`
	ArgvHash             string `json:"argv_hash"` // 64-hex digest of the exact argv (A1-4.2)
	ToolID               string `json:"tool_id"`
	ToolVersion          string `json:"tool_version"`      // ≤ 64 B, registry vocabulary
	Target               string `json:"target"`            // ≤ 256 B, untrusted (A1-4.4/4.5)
	ActionSummary        string `json:"action_summary"`    // ≤ 512 B, untrusted
	RiskTier             string `json:"risk_tier"`         // ≤ 32 B; A7 owns the vocabulary (A1-3.6)
	ExpiresAt            string `json:"expires_at"`        // timestamp string, platform-computed (A1-4.2/4.12)
	UntrustedContext     bool   `json:"untrusted_context"` // platform-computed (A1-4.4)
	RequestEventID       string `json:"request_event_id"`
}

// Kind returns KindApprovalRequested.
func (ApprovalRequestedPayload) Kind() Kind { return KindApprovalRequested }

// ApprovalGrantedPayload is the payload of kind approval_granted (A1-3.3,
// ADR-0012 §1, Q10). The approver's identity is the envelope actor and the
// decision time is platform time. FingerprintHash must equal the one on the
// approval_requested event with the same ApprovalID (A1-4.2), and ExpiresAt
// must be byte-equal to the requested value (A1-4.2). SingleUse is true for
// every v1 approval (Q10).
type ApprovalGrantedPayload struct {
	ApprovalID      string `json:"approval_id"`
	FingerprintHash string `json:"fingerprint_hash"` // 64-hex (A0-2.15)
	ExpiresAt       string `json:"expires_at"`       // timestamp string (A1-4.12)
	SingleUse       bool   `json:"single_use"`
	QueueWaitMS     int64  `json:"queue_wait_ms"`
}

// Kind returns KindApprovalGranted.
func (ApprovalGrantedPayload) Kind() Kind { return KindApprovalGranted }

// ApprovalDeniedPayload is the payload of kind approval_denied (A1-3.3,
// ADR-0012 §1).
type ApprovalDeniedPayload struct {
	ApprovalID      string `json:"approval_id"`
	FingerprintHash string `json:"fingerprint_hash"` // 64-hex (A0-2.15)
	Reason          string `json:"reason"`           // ≤ 512 B, untrusted (A1-4.4/4.5)
}

// Kind returns KindApprovalDenied.
func (ApprovalDeniedPayload) Kind() Kind { return KindApprovalDenied }

// ApprovalExpiredPayload is the payload of kind approval_expired (A1-3.3,
// ADR-0012 §7): the platform let an approval time out; the orchestrator
// replans on it (A0-3.1).
type ApprovalExpiredPayload struct {
	ApprovalID      string `json:"approval_id"`
	FingerprintHash string `json:"fingerprint_hash"` // 64-hex (A0-2.15)
	ExpiresAt       string `json:"expires_at"`       // timestamp string (A1-4.12)
	QueueWaitMS     int64  `json:"queue_wait_ms"`
}

// Kind returns KindApprovalExpired.
func (ApprovalExpiredPayload) Kind() Kind { return KindApprovalExpired }

// ApprovalExecutedPayload is the payload of kind approval_executed
// (A1-3.3, ADR-0018 §2–§3, Q10): action and fingerprint recorded together at
// execution. Revalidated records the execution-time re-validation outcome
// and must be true for a successful execution; SingleUseConsumed is the Q10
// consumption record. ActionSpecEvidenceID and FingerprintHash must equal
// the approval_requested values for the same ApprovalID (A1-4.2).
type ApprovalExecutedPayload struct {
	ApprovalID           string `json:"approval_id"`
	FingerprintHash      string `json:"fingerprint_hash"` // 64-hex (A0-2.15)
	ActionSpecEvidenceID string `json:"action_spec_evidence_id"`
	TargetGraphNodeID    string `json:"target_graph_node_id"`
	ArgvHash             string `json:"argv_hash"` // 64-hex (A0-2.15)
	Revalidated          bool   `json:"revalidated"`
	SingleUseConsumed    bool   `json:"single_use_consumed"`
	ExpiresAt            string `json:"expires_at"` // timestamp string (A1-4.12)
	ToolID               string `json:"tool_id"`
	ToolVersion          string `json:"tool_version"`   // ≤ 64 B, registry vocabulary
	Target               string `json:"target"`         // ≤ 256 B, untrusted (A1-4.4/4.5)
	ActionSummary        string `json:"action_summary"` // ≤ 512 B, untrusted
}

// Kind returns KindApprovalExecuted.
func (ApprovalExecutedPayload) Kind() Kind { return KindApprovalExecuted }
