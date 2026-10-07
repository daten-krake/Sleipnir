package events

// GraphNodeWrittenPayload is the payload of kind graph_node_written
// (A1-3.3, ADR-0016 §1/§4, A2-4.6/4.7): the event-side half of the A2
// provenance seam. The field names A2 must reference and must not rename are
// fixed by A1-3.6. SourceEventID must be non-empty — provenance is mandatory
// on the graph side. NodeKind is A2-2.1 vocabulary stored as a capped string
// so an A2 addition needs no A1 change. A dedup collapse must still emit the
// event with DedupHit true (A1-3.3). The originating worker/orchestrator is
// NOT the actor — the actor is (platform, "graph") and the producer lives in
// the node's provenance.principal_kind (A1-3.6, A2-5.3).
type GraphNodeWrittenPayload struct {
	GraphNodeID           string `json:"graph_node_id"`
	NodeKind              string `json:"node_kind"` // ≤ 32 B; A2 owns the value list (A1-3.6)
	SourceEventID         string `json:"source_event_id"`
	SupersedesGraphNodeID string `json:"supersedes_graph_node_id"`
	Quarantined           bool   `json:"quarantined"`
	ContentHash           string `json:"content_hash"` // 64-hex, the A2-4.6 fingerprint
	DedupHit              bool   `json:"dedup_hit"`    // A2-4.7 collapse still emits the event
	SummaryBytes          int64  `json:"summary_bytes"`
	AttrsCount            int64  `json:"attrs_count"`
}

// Kind returns KindGraphNodeWritten.
func (GraphNodeWrittenPayload) Kind() Kind { return KindGraphNodeWritten }

// GraphEdgeWrittenPayload is the payload of kind graph_edge_written
// (A1-3.3, ADR-0016 §1, A2-4.7). Edges have no fingerprint, hence no
// content_hash; a dedup collapse must still emit the event with DedupHit
// true.
type GraphEdgeWrittenPayload struct {
	GraphEdgeID     string `json:"graph_edge_id"`
	EdgeKind        string `json:"edge_kind"` // ≤ 32 B; A2 owns the value list (A1-3.6)
	FromGraphNodeID string `json:"from_graph_node_id"`
	ToGraphNodeID   string `json:"to_graph_node_id"`
	SourceEventID   string `json:"source_event_id"`
	Quarantined     bool   `json:"quarantined"`
	DedupHit        bool   `json:"dedup_hit"`
}

// Kind returns KindGraphEdgeWritten.
func (GraphEdgeWrittenPayload) Kind() Kind { return KindGraphEdgeWritten }

// GraphNodeQuarantinedPayload is the payload of kind graph_node_quarantined
// (A1-3.3, ADR-0016 §2, C5, Q5). QuarantineKind is the closed A1 occurrence
// enum out_of_scope_discovery | blacklist_match | operator_quarantine — A2
// owns the state vocabulary and derives it from this mapping; there is no
// operator_release value (A1-3.6, §6 item 14).
type GraphNodeQuarantinedPayload struct {
	GraphNodeID    string `json:"graph_node_id"`
	QuarantineKind string `json:"quarantine_kind"` // enum out_of_scope_discovery|blacklist_match|operator_quarantine
	Reason         string `json:"reason"`          // ≤ 512 B, untrusted (A1-4.4/4.5)
	SourceEventID  string `json:"source_event_id"`
}

// Kind returns KindGraphNodeQuarantined.
func (GraphNodeQuarantinedPayload) Kind() Kind { return KindGraphNodeQuarantined }

// QuarantineRecomputedPayload is the payload of kind quarantine_recomputed
// (A1-3.3, added for A2-8.5): the batch recomputation a scope/blacklist
// change or a node/edge write touching a quarantined node causes. Trigger is
// the closed enum scope_changed | blacklist_changed | node_written |
// edge_written (A1-3.3). TriggerEventID must resolve, in this engagement, to
// the event that caused the recomputation (A1-4.2). Per-node effects are
// separate graph_node_quarantined events; this event records the batch and
// its trigger.
type QuarantineRecomputedPayload struct {
	Trigger          string `json:"trigger"` // enum scope_changed|blacklist_changed|node_written|edge_written
	TriggerEventID   string `json:"trigger_event_id"`
	NodesEvaluated   int64  `json:"nodes_evaluated"`
	NodesQuarantined int64  `json:"nodes_quarantined"`
	NodesReleased    int64  `json:"nodes_released"`
	DurationMS       int64  `json:"duration_ms"`
}

// Kind returns KindQuarantineRecomputed.
func (QuarantineRecomputedPayload) Kind() Kind { return KindQuarantineRecomputed }

// GraphEdgeRetractedPayload is the payload of kind graph_edge_retracted
// (A1-3.3, added for A2-3.9, ADR-0016 §4): the self-correction event that
// always accompanies the one mutable edge field. SourceEventID is the
// observation that contradicted the edge, "" when operator-initiated; Reason
// must be non-empty when the retraction is operator-initiated (A1-4.2).
type GraphEdgeRetractedPayload struct {
	GraphEdgeID     string `json:"graph_edge_id"`
	EdgeKind        string `json:"edge_kind"` // ≤ 32 B; A2 owns the value list (A1-3.6)
	FromGraphNodeID string `json:"from_graph_node_id"`
	ToGraphNodeID   string `json:"to_graph_node_id"`
	Reason          string `json:"reason"` // ≤ 512 B, untrusted (A1-4.4/4.5)
	SourceEventID   string `json:"source_event_id"`
}

// Kind returns KindGraphEdgeRetracted.
func (GraphEdgeRetractedPayload) Kind() Kind { return KindGraphEdgeRetracted }

// ReportInclusionChangedPayload is the payload of kind
// report_inclusion_changed (A1-3.3, Q5): the operator removes or restores a
// (typically quarantined) discovery in the report. The flag itself is mutable
// graph state (A2) and must never be an order key (A0-4.3).
type ReportInclusionChangedPayload struct {
	GraphNodeID string `json:"graph_node_id"`
	Included    bool   `json:"included"`
	Reason      string `json:"reason"` // ≤ 512 B, untrusted (A1-4.4/4.5)
}

// Kind returns KindReportInclusionChanged.
func (ReportInclusionChangedPayload) Kind() Kind { return KindReportInclusionChanged }
