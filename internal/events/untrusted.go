package events

import "slices"

// untrustedFields is the A1-4.4 registry: for every kind that A1-3.3 marks
// with `*`, the JSON names of its untrusted-content fields, in A1-3.3 field
// order. The 23 kinds not listed have no starred field ("every other kind
// (23 of 42) | none", A1-4.4). The envelope's untrusted boolean must be true
// iff at least one of these fields is non-empty on the instance (A1-4.4);
// the computation itself is the composition path's job (A1-7.10).
var untrustedFields = map[Kind][]string{
	KindIntegrityOverride:      {"reason"},
	KindRunEnded:               {"detail"},
	KindHardStopFired:          {"reason"},
	KindSpawnRequested:         {"task_description"},
	KindCommandExecuted:        {"command", "target"},
	KindTaskResult:             {"result_summary"},
	KindRevertRecorded:         {"target", "revert_action"},
	KindApprovalRequested:      {"target", "action_summary"},
	KindApprovalDenied:         {"reason"},
	KindApprovalExecuted:       {"target", "action_summary"},
	KindGraphNodeQuarantined:   {"reason"},
	KindGraphEdgeRetracted:     {"reason"},
	KindReportInclusionChanged: {"reason"},
	KindCleanupExecuted:        {"detail"},
	KindCleanupVerified:        {"detail"},
	KindScopeDenied:            {"attempted_target"},
	KindBlacklistDenied:        {"attempted_target"},
	KindActionBlocked:          {"attempted_target", "detail"},
	KindAgentError:             {"message"},
}

// UntrustedFields returns the JSON names of kind k's `*`-marked payload
// fields — the content that originates outside the platform's own vocabulary
// (tool output, target-supplied data, model prose, client prose, human-typed
// free text) per A1-4.4 — in A1-3.3 field order. The returned slice is a
// fresh copy the caller may freely modify.
//
// fields is empty for a kind with no starred field (23 of 42). ok is false
// when k is not in the closed taxonomy (A1-3.1); callers must reject an
// unknown kind with validation before ever asking for its fields, so ok
// false indicates a platform defect at the call site, not a client error.
func UntrustedFields(k Kind) (fields []string, ok bool) {
	if !isKnownKind(k) {
		return nil, false
	}
	starred := untrustedFields[k]
	if starred == nil {
		return []string{}, true
	}
	return slices.Clone(starred), true
}
