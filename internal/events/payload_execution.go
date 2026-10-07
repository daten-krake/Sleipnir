package events

// CommandExecutedPayload is the payload of kind command_executed (C —
// client-appendable, A1-3.4; ADR-0009 §1). Command is the exact argv as
// executed, never a paraphrase (A1-4.2). OutputEvidenceID must be non-empty
// when OutputBytes > 0; the output itself is an evi_ reference, never bytes
// in the payload (A1-4.3). ExitCode is [-1,255] with -1 = no exit status; it
// is the only field A1-4.12 transcribes as Go int instead of int64.
type CommandExecutedPayload struct {
	Command          string `json:"command"` // ≤ 2048 B, untrusted (A1-4.4/4.5)
	Target           string `json:"target"`  // ≤ 256 B, untrusted
	ToolID           string `json:"tool_id"`
	ToolVersion      string `json:"tool_version"` // ≤ 64 B, registry vocabulary (A1-3.6)
	ExitCode         int    `json:"exit_code"`    // [-1,255] (A1-4.2); -1 = no exit status
	DurationMS       int64  `json:"duration_ms"`
	OutputBytes      int64  `json:"output_bytes"`
	OutputEvidenceID string `json:"output_evidence_id"`
	Redacted         bool   `json:"redacted"` // A1-4.6: platform-set redaction marker, never "truncated"
}

// Kind returns KindCommandExecuted.
func (CommandExecutedPayload) Kind() Kind { return KindCommandExecuted }

// EvidenceStoredPayload is the payload of kind evidence_stored (A1-3.3,
// ADR-0009 §2). EvidenceID must be the id returned by evidence:upload;
// SHA256 is the artifact-integrity digest (64-hex, A0-2.15) and must be
// non-empty — it is not "a captured hash" in A0-3.7's sense (A1-4.9).
// EvidenceKind and Source are closed enums (A1-3.3).
type EvidenceStoredPayload struct {
	EvidenceID   string `json:"evidence_id"`
	EvidenceKind string `json:"evidence_kind"` // enum command_output|screenshot|file_capture|memory_dump|packet_capture|browser_session|config_snapshot|report_artifact|other
	MediaType    string `json:"media_type"`    // ≤ 128 B
	SizeBytes    int64  `json:"size_bytes"`
	SHA256       string `json:"sha256"`   // 64-hex (A0-2.15)
	Source       string `json:"source"`   // enum worker|node|orchestrator|browser|platform
	Redacted     bool   `json:"redacted"` // A1-4.6
}

// Kind returns KindEvidenceStored.
func (EvidenceStoredPayload) Kind() Kind { return KindEvidenceStored }

// TaskResultPayload is the payload of kind task_result (C —
// client-appendable, A1-3.4; Q6 worker scope task:result). Status is the
// closed enum succeeded | failed | partial; a failed result carries a
// non-empty ErrorKind equal to an A0-3.1 kind (A1-4.2, A1-4.12).
// RevertEventIDs lists the revert_recorded events this task produced and may
// be empty; it is sorted and deduplicated at composition (A1-4.7).
// CommandCount is the worker's own claim — a divergence from the platform's
// count is reportable, not a rejection (A1-4.2).
type TaskResultPayload struct {
	Status         string   `json:"status"`         // enum succeeded|failed|partial
	ResultSummary  string   `json:"result_summary"` // ≤ 2048 B, untrusted (A1-4.4/4.5)
	DurationMS     int64    `json:"duration_ms"`
	CommandCount   int64    `json:"command_count"`
	RevertEventIDs []string `json:"revert_event_ids"` // ≤ EventRefsMax, sorted (A1-4.7)
	ErrorKind      string   `json:"error_kind"`       // A0-3.1 kind; non-empty when Status == "failed"
}

// Kind returns KindTaskResult.
func (TaskResultPayload) Kind() Kind { return KindTaskResult }

// RevertRecordedPayload is the payload of kind revert_recorded (C —
// client-appendable, A1-3.4; ADR-0009 §3). The revert record IS the event:
// it is identified by its event_id (A1-1.6). EffectKind is the closed enum
// of A1-3.3. With Revertable false, RevertAction may be empty and the effect
// must be carried forward into cleanup_planned.non_revertable_event_ids
// (A1-4.2).
type RevertRecordedPayload struct {
	EffectKind            string `json:"effect_kind"`   // enum account_created|account_modified|acl_changed|file_dropped|scheduled_task|service_installed|registry_edit|config_changed|credential_changed|persistence_added|other
	Target                string `json:"target"`        // ≤ 256 B, untrusted (A1-4.4/4.5)
	RevertAction          string `json:"revert_action"` // ≤ 2048 B, untrusted
	Revertable            bool   `json:"revertable"`
	ToolID                string `json:"tool_id"`
	StateChangeEvidenceID string `json:"state_change_evidence_id"`
}

// Kind returns KindRevertRecorded.
func (RevertRecordedPayload) Kind() Kind { return KindRevertRecorded }
