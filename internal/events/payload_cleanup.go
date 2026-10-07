package events

// CleanupPlannedPayload is the payload of kind cleanup_planned (A1-3.3,
// ADR-0009 §4): the plan assembled from revert records. ApprovalID must be
// non-empty — a cleanup plan executes only after human approval (A1-4.2).
// NonRevertableEventIDs documents the effects that cannot be reverted; they
// are documented, never dropped (ADR-0009 §4, A1-4.2 revert_recorded row).
// Both arrays are sorted and deduplicated at composition (A1-4.7).
type CleanupPlannedPayload struct {
	PlanEvidenceID        string   `json:"plan_evidence_id"`
	RevertEventIDs        []string `json:"revert_event_ids"`         // ≤ EventRefsMax, sorted (A1-4.7)
	NonRevertableEventIDs []string `json:"non_revertable_event_ids"` // ≤ EventRefsMax, sorted (A1-4.7)
	PlannedActionCount    int64    `json:"planned_action_count"`
	ApprovalID            string   `json:"approval_id"`
}

// Kind returns KindCleanupPlanned.
func (CleanupPlannedPayload) Kind() Kind { return KindCleanupPlanned }

// CleanupExecutedPayload is the payload of kind cleanup_executed (A1-3.3,
// ADR-0009 §4). Status is the closed enum reverted | failed | skipped; a
// failed execution carries a non-empty Detail (A1-4.2). RevertEventID must
// resolve to a revert_recorded event.
type CleanupExecutedPayload struct {
	RevertEventID  string `json:"revert_event_id"`
	Status         string `json:"status"` // enum reverted|failed|skipped
	CommandEventID string `json:"command_event_id"`
	Detail         string `json:"detail"` // ≤ 512 B, untrusted (A1-4.4/4.5)
}

// Kind returns KindCleanupExecuted.
func (CleanupExecutedPayload) Kind() Kind { return KindCleanupExecuted }

// CleanupVerifiedPayload is the payload of kind cleanup_verified (A1-3.3,
// ADR-0009 §4: execute then verify). RevertEventID must resolve to a
// revert_recorded event.
type CleanupVerifiedPayload struct {
	RevertEventID          string `json:"revert_event_id"`
	Verified               bool   `json:"verified"`
	VerificationEvidenceID string `json:"verification_evidence_id"`
	Detail                 string `json:"detail"` // ≤ 512 B, untrusted (A1-4.4/4.5)
}

// Kind returns KindCleanupVerified.
func (CleanupVerifiedPayload) Kind() Kind { return KindCleanupVerified }
