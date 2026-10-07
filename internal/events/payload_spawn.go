package events

// SpawnRequestedPayload is the payload of kind spawn_requested (A1-3.3,
// Q14, ADR-0017 §2): the orchestrator's request for a job or task
// container. Subject is the closed enum job | task; Network is the closed
// enum run_isolated | target_only | none. ImageDigest and RiskTier are
// platform-derived from the registry and are "" only when validation failed
// before derivation (A1-4.2). TaskDescription is untrusted model prose,
// capped (A1-4.4/4.5). TargetGraphNodeID is "" or a gn_ id resolving in this
// engagement — the target of a spawn is never taken from a graph field, only
// cited by id (A2-8.3).
type SpawnRequestedPayload struct {
	Subject           string `json:"subject"` // enum job|task
	ToolID            string `json:"tool_id"`
	ToolVersion       string `json:"tool_version"`     // ≤ 64 B, registry vocabulary (A1-3.6)
	TaskDescription   string `json:"task_description"` // ≤ 2048 B, untrusted (A1-4.4/4.5)
	CPUMillicores     int64  `json:"cpu_millicores"`
	MemoryBytes       int64  `json:"memory_bytes"`
	Network           string `json:"network"`      // enum run_isolated|target_only|none
	RiskTier          string `json:"risk_tier"`    // ≤ 32 B; A7 owns the vocabulary (A1-3.6)
	ImageDigest       string `json:"image_digest"` // ≤ 256 B; platform-derived (A1-4.2)
	ApprovalID        string `json:"approval_id"`
	TargetGraphNodeID string `json:"target_graph_node_id"`
}

// Kind returns KindSpawnRequested.
func (SpawnRequestedPayload) Kind() Kind { return KindSpawnRequested }
