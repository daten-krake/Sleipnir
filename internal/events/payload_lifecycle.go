package events

// EngagementCreatedPayload is the payload of kind engagement_created
// (A1-3.3, SPEC §5.1, ADR-0005), composed at seq 1 of the engagement chain
// — chain_genesis stays the integrity anchor at seq 0 and carries no
// engagement metadata (A1-5.3). RoeEvidenceID and PolicyEvidenceID must be
// non-empty: an engagement whose rules of engagement are not an artifact is
// not reproducible (A1-4.2, SPEC §7). ClientRef is a customer-chosen label,
// never a person's name (A0-3.7).
type EngagementCreatedPayload struct {
	ClientRef        string `json:"client_ref"` // ≤ 128 B
	RoeEvidenceID    string `json:"roe_evidence_id"`
	PolicyEvidenceID string `json:"policy_evidence_id"`
	OperatorCount    int64  `json:"operator_count"`
}

// Kind returns KindEngagementCreated.
func (EngagementCreatedPayload) Kind() Kind { return KindEngagementCreated }

// ScopeChangedPayload is the payload of kind scope_changed (A1-3.3, C5,
// ADR-0005). ChangeKind is the closed enum allowlist_added |
// allowlist_removed | blacklist_added | blacklist_removed | roe_changed.
// EntryHash is the A0-2.15 digest of Entry, so a scope entry is comparable
// without echoing it into every report (A1-4.2). A global-blacklist change
// is composed into every affected engagement chain (adversarial C-04).
type ScopeChangedPayload struct {
	ChangeKind string `json:"change_kind"` // enum allowlist_added|allowlist_removed|blacklist_added|blacklist_removed|roe_changed
	Entry      string `json:"entry"`       // ≤ 512 B operator-typed value (A1-4.5); platform vocabulary, not starred (A1-4.4)
	EntryHash  string `json:"entry_hash"`  // 64-hex (A0-2.15)
}

// Kind returns KindScopeChanged.
func (ScopeChangedPayload) Kind() Kind { return KindScopeChanged }

// EngagementPolicyChangedPayload is the payload of kind
// engagement_policy_changed (A1-3.3, ADR-0020 §1, ADR-0012 §7, ADR-0008).
// PolicyKind is the closed enum llm_data_policy | approval_timeout |
// model_role_matrix | risk_tier | notification_channel.
type EngagementPolicyChangedPayload struct {
	PolicyKind     string `json:"policy_kind"` // enum llm_data_policy|approval_timeout|model_role_matrix|risk_tier|notification_channel
	OldValue       string `json:"old_value"`   // ≤ 128 B
	NewValue       string `json:"new_value"`   // ≤ 128 B
	AffectedToolID string `json:"affected_tool_id"`
}

// Kind returns KindEngagementPolicyChanged.
func (EngagementPolicyChangedPayload) Kind() Kind { return KindEngagementPolicyChanged }

// RunStartedPayload is the payload of kind run_started (A1-3.3, SPEC §5.2,
// ADR-0020 §1). LLMDataPolicy must equal the engagement policy in force at
// start and is the closed enum local_only | cloud_masked | cloud_raw;
// ScopeSnapshotEvidenceID must be non-empty — a run without a scope snapshot
// is not reproducible (A1-4.2, SPEC §7).
type RunStartedPayload struct {
	LLMDataPolicy           string `json:"llm_data_policy"` // enum local_only|cloud_masked|cloud_raw (ADR-0020 §1)
	ApprovalTimeoutMS       int64  `json:"approval_timeout_ms"`
	ScopeSnapshotEvidenceID string `json:"scope_snapshot_evidence_id"`
}

// Kind returns KindRunStarted.
func (RunStartedPayload) Kind() Kind { return KindRunStarted }

// RunEndedPayload is the payload of kind run_ended (A1-3.3, SPEC §5).
// EndReason is the closed enum completed | failed | cancelled | hard_stop.
type RunEndedPayload struct {
	EndReason string `json:"end_reason"` // enum completed|failed|cancelled|hard_stop
	Detail    string `json:"detail"`     // ≤ 512 B, untrusted (A1-4.4/4.5)
}

// Kind returns KindRunEnded.
func (RunEndedPayload) Kind() Kind { return KindRunEnded }

// ModelConfigSnapshottedPayload is the payload of kind
// model_config_snapshotted (A1-3.3, SPEC §7): the reproducibility snapshot
// of the exact model configuration in force. MatrixHash is the 64-hex
// A0-2.15 digest of the model role matrix.
type ModelConfigSnapshottedPayload struct {
	ConfigEvidenceID string `json:"config_evidence_id"`
	MatrixHash       string `json:"matrix_hash"` // 64-hex (A0-2.15)
	RoleCount        int64  `json:"role_count"`
}

// Kind returns KindModelConfigSnapshotted.
func (ModelConfigSnapshottedPayload) Kind() Kind { return KindModelConfigSnapshotted }

// JobSpawnedPayload is the payload of kind job_spawned (A1-3.3, ADR-0017
// §3): the orchestrator container the spawn broker created. ImageDigest is
// the registry-derived digest, never an orchestrator-supplied string
// (A1-4.2, Q14).
type JobSpawnedPayload struct {
	SpawnRequestEventID string `json:"spawn_request_event_id"`
	ImageDigest         string `json:"image_digest"` // ≤ 256 B, registry-derived (A1-4.5)
	NetworkName         string `json:"network_name"` // ≤ 128 B
}

// Kind returns KindJobSpawned.
func (JobSpawnedPayload) Kind() Kind { return KindJobSpawned }

// TaskSpawnedPayload is the payload of kind task_spawned (A1-3.3, Q14,
// ADR-0017 §2–§3): the worker container the spawn broker created.
// TargetGraphNodeID must equal the spawn_requested value for the same spawn
// — the target is cited by gn_ id, never by string (A1-4.2, C-02/T-03,
// A2-8.3).
type TaskSpawnedPayload struct {
	SpawnRequestEventID string `json:"spawn_request_event_id"`
	ToolID              string `json:"tool_id"`
	ToolVersion         string `json:"tool_version"` // ≤ 64 B, registry vocabulary (A1-3.6)
	RiskTier            string `json:"risk_tier"`    // ≤ 32 B, A7 owns the vocabulary (A1-3.6)
	ImageDigest         string `json:"image_digest"` // ≤ 256 B, registry-derived
	ApprovalID          string `json:"approval_id"`
	NetworkName         string `json:"network_name"` // ≤ 128 B
	TargetGraphNodeID   string `json:"target_graph_node_id"`
}

// Kind returns KindTaskSpawned.
func (TaskSpawnedPayload) Kind() Kind { return KindTaskSpawned }

// ContainerStartedPayload is the payload of kind container_started
// (A1-3.3, ADR-0007/0017). Subject is the closed enum job | task.
type ContainerStartedPayload struct {
	Subject      string `json:"subject"`       // enum job|task
	ContainerRef string `json:"container_ref"` // ≤ 128 B platform label
	ImageDigest  string `json:"image_digest"`  // ≤ 256 B, registry-derived
}

// Kind returns KindContainerStarted.
func (ContainerStartedPayload) Kind() Kind { return KindContainerStarted }

// ContainerKilledPayload is the payload of kind container_killed (A1-3.3,
// ADR-0005, ADR-0017 §3). KillReason is the closed enum completed | timeout
// | hard_stop | quota | node_lost | error | operator. ExitCode is -1 when no
// exit status exists (A1-4.2) — the only field A1-4.12 transcribes as Go int
// instead of int64. A hard_stop kill carries StopEventID, the hard_stop_fired
// event it answers (A1-4.2, A1-7.12).
type ContainerKilledPayload struct {
	Subject      string `json:"subject"`       // enum job|task
	ContainerRef string `json:"container_ref"` // ≤ 128 B platform label
	KillReason   string `json:"kill_reason"`   // enum completed|timeout|hard_stop|quota|node_lost|error|operator
	ExitCode     int    `json:"exit_code"`     // [-1,255] (A1-4.2); -1 = no exit status
	DurationMS   int64  `json:"duration_ms"`
	StopEventID  string `json:"stop_event_id"`
}

// Kind returns KindContainerKilled.
func (ContainerKilledPayload) Kind() Kind { return KindContainerKilled }

// HardStopFiredPayload is the payload of kind hard_stop_fired (A1-3.3,
// ADR-0005, ADR-0012 §2): the operator's emergency stop. StopScope is the
// closed enum run | engagement. The platform commits this event before it
// issues any kill (A1-7.12).
type HardStopFiredPayload struct {
	StopScope        string `json:"stop_scope"` // enum run|engagement
	Reason           string `json:"reason"`     // ≤ 512 B, untrusted (A1-4.4/4.5)
	ContainersKilled int64  `json:"containers_killed"`
	RespawnBlocked   bool   `json:"respawn_blocked"`
}

// Kind returns KindHardStopFired.
func (HardStopFiredPayload) Kind() Kind { return KindHardStopFired }

// EngagementClosedPayload is the payload of kind engagement_closed
// (A1-3.3, SPEC §5 step 10): the last user-composed fact of a chain, after
// cleanup is verified. CloseReason is the closed enum completed | cancelled
// | abandoned; ReportEvidenceID must be non-empty when CloseReason is
// "completed" (A1-4.2).
type EngagementClosedPayload struct {
	CloseReason      string `json:"close_reason"` // enum completed|cancelled|abandoned
	ReportEvidenceID string `json:"report_evidence_id"`
}

// Kind returns KindEngagementClosed.
func (EngagementClosedPayload) Kind() Kind { return KindEngagementClosed }
