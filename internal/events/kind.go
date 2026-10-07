package events

import "slices"

// Kind is the closed event taxonomy (A1-3.1, 42 kinds). A0-8.5 spelling
// (^[a-z][a-z0-9_]{0,31}$, <subject>_<past_participle>), byte-exact
// comparison, no synonyms. An unknown or malformed kind on a write is
// hard-rejected with validation (A0-6.3) by the composition path (A1-7.10);
// on a read a client preserves the raw string and skips what it cannot
// handle.
type Kind string

// The 42 kinds of A1-3.3, in contract-table order. The comments name the
// A1-3.3 group each constant belongs to; the spelling is byte-exact from the
// tables (TestKindListIs42AndClosed pins it against the contract text).
const (
	// Integrity (Q11, A1-5/A1-6).
	KindChainGenesis       Kind = "chain_genesis"
	KindChainVerified      Kind = "chain_verified"
	KindChainBreakDetected Kind = "chain_break_detected"
	KindIntegrityOverride  Kind = "integrity_override"
	KindArtifactReleased   Kind = "artifact_released" // A1-6.6, override attribution (A1-6.5, ADR-0021)

	// Engagement and run lifecycle (SPEC §5 steps 1–2, 8).
	KindEngagementCreated       Kind = "engagement_created" // seq 1 (A1-5.3 keeps genesis at seq 0)
	KindScopeChanged            Kind = "scope_changed"
	KindEngagementPolicyChanged Kind = "engagement_policy_changed"
	KindRunStarted              Kind = "run_started"
	KindRunEnded                Kind = "run_ended"
	KindModelConfigSnapshotted  Kind = "model_config_snapshotted"
	KindJobSpawned              Kind = "job_spawned"
	KindTaskSpawned             Kind = "task_spawned"
	KindContainerStarted        Kind = "container_started"
	KindContainerKilled         Kind = "container_killed"
	KindHardStopFired           Kind = "hard_stop_fired"
	KindEngagementClosed        Kind = "engagement_closed"

	// Spawn request (Q14, ADR-0017 §2).
	KindSpawnRequested Kind = "spawn_requested"

	// Command execution, evidence, results (ADR-0009 §1–§3) — the three
	// client-appendable C kinds live here (A1-3.4).
	KindCommandExecuted Kind = "command_executed" // C
	KindEvidenceStored  Kind = "evidence_stored"
	KindTaskResult      Kind = "task_result"     // C
	KindRevertRecorded  Kind = "revert_recorded" // C

	// Approvals (Q10, ADR-0018 §1–§4).
	KindApprovalRequested Kind = "approval_requested"
	KindApprovalGranted   Kind = "approval_granted"
	KindApprovalDenied    Kind = "approval_denied"
	KindApprovalExpired   Kind = "approval_expired"
	KindApprovalExecuted  Kind = "approval_executed"

	// LLM traffic (ADR-0020 §2–§5).
	KindLLMCall Kind = "llm_call"

	// Graph mutation (ADR-0016, A2 seam — A1-3.6).
	KindGraphNodeWritten       Kind = "graph_node_written"
	KindGraphEdgeWritten       Kind = "graph_edge_written"
	KindGraphNodeQuarantined   Kind = "graph_node_quarantined"
	KindQuarantineRecomputed   Kind = "quarantine_recomputed" // added for A2-8.5
	KindGraphEdgeRetracted     Kind = "graph_edge_retracted"  // added for A2-3.9
	KindReportInclusionChanged Kind = "report_inclusion_changed"

	// Cleanup / revert execution (ADR-0009 §4).
	KindCleanupPlanned  Kind = "cleanup_planned"
	KindCleanupExecuted Kind = "cleanup_executed"
	KindCleanupVerified Kind = "cleanup_verified"

	// Enforcement denials and agent errors (ADR-0005, C5/C9, ADR-0018 §2).
	KindScopeDenied     Kind = "scope_denied"
	KindBlacklistDenied Kind = "blacklist_denied"
	KindActionBlocked   Kind = "action_blocked"
	KindAgentError      Kind = "agent_error"

	// Notification delivery (ADR-0012 §2–§6).
	KindNotificationSent Kind = "notification_sent"
)

// kinds is the closed taxonomy in A1-3.3 table order. Read it through
// AllKinds; the slice is unexported so the closed list cannot be extended
// from outside this file (A1-3.1, A1-3.5: additions are contract amendments).
var kinds = []Kind{
	KindChainGenesis,
	KindChainVerified,
	KindChainBreakDetected,
	KindIntegrityOverride,
	KindArtifactReleased,

	KindEngagementCreated,
	KindScopeChanged,
	KindEngagementPolicyChanged,
	KindRunStarted,
	KindRunEnded,
	KindModelConfigSnapshotted,
	KindJobSpawned,
	KindTaskSpawned,
	KindContainerStarted,
	KindContainerKilled,
	KindHardStopFired,
	KindEngagementClosed,

	KindSpawnRequested,

	KindCommandExecuted,
	KindEvidenceStored,
	KindTaskResult,
	KindRevertRecorded,

	KindApprovalRequested,
	KindApprovalGranted,
	KindApprovalDenied,
	KindApprovalExpired,
	KindApprovalExecuted,

	KindLLMCall,

	KindGraphNodeWritten,
	KindGraphEdgeWritten,
	KindGraphNodeQuarantined,
	KindQuarantineRecomputed,
	KindGraphEdgeRetracted,
	KindReportInclusionChanged,

	KindCleanupPlanned,
	KindCleanupExecuted,
	KindCleanupVerified,

	KindScopeDenied,
	KindBlacklistDenied,
	KindActionBlocked,
	KindAgentError,

	KindNotificationSent,
}

// AllKinds returns the closed taxonomy (A1-3.1, 42 kinds) in A1-3.3 table
// order. The returned slice is a fresh copy; callers cannot extend or
// reorder the closed list.
func AllKinds() []Kind {
	return slices.Clone(kinds)
}

// isKnownKind reports whether k is in the closed taxonomy (A1-3.1).
func isKnownKind(k Kind) bool {
	return slices.Contains(kinds, k)
}

// ClientAppendable reports whether k is one of the three C kinds reachable
// through events:append (A1-3.4, A1-7.4): command_executed, task_result and
// revert_recorded. Every other kind — including an unknown one — is
// platform-composed and unreachable for any machine principal.
func ClientAppendable(k Kind) bool {
	switch k {
	case KindCommandExecuted, KindTaskResult, KindRevertRecorded:
		return true
	default:
		return false
	}
}
