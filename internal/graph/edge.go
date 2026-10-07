package graph

// Edge is one directed relationship (A2-3). Content is immutable; Retracted
// is the only mutable field (A2-3.9), changed solely through the store seam's
// SetEdgeRetracted (A2-10.6, WP-15/WP-17). source_id/target_id are gn_ ids in
// the same engagement (A2-3.1, A2-11.3); source_kind/target_kind are
// denormalized onto the edge by the platform at ingest and MUST equal the
// referenced nodes' kinds (A2-3.8 — the contract suite asserts the equality,
// so the duplication cannot drift silently).
//
// An edge carries no content_hash: A2-4.6 fingerprints nodes only; the edge
// dedup key is A2-3.5's (engagement_id, kind, source_id, target_id).
// Field naming follows A2-1.6: graph_edge_id, source_id, target_id — never a
// bare id and never node_id (A0-3.6).
type Edge struct {
	ID           string       `json:"graph_edge_id"` // ge_ (A0-1.2, A2-1.6), platform-generated (A0-1.4)
	EngagementID string       `json:"engagement_id"` // platform-stamped (A2-1.5)
	GraphSeq     int64        `json:"graph_seq"`     // immutable order key (A2-1.4a)
	Kind         EdgeKind     `json:"kind"`
	SourceID     string       `json:"source_id"` // gn_
	SourceKind   NodeKind     `json:"source_kind"`
	TargetID     string       `json:"target_id"` // gn_
	TargetKind   NodeKind     `json:"target_kind"`
	Retracted    bool         `json:"retracted"` // A2-3.9: the only mutable edge field, always present
	Provenance   []Provenance `json:"provenance"`
}

// AllowsEndpoints reports whether an edge of kind k may run from a source of
// kind source to a target of kind target, per the closed A2-3.2 endpoint
// matrix. It is the matrix as data; the rejection of a disallowed pair —
// validation naming the edge kind, the offending endpoint kind and the
// allowed set (A2-10.5) — and the endpoint resolution inside the engagement
// (A2-11.3, notfound) belong to the write validation (WP-15).
//
// The supersedes row is not a static pair of sets: it allows any kind K to
// the same kind K (A2-3.2, A2-4.3), so it is decided by equality plus kind
// validity. An unknown edge kind, or an unknown endpoint kind anywhere else
// in the matrix, is never allowed (A0-6.3: the closed-list rule covers
// endpoint kinds too, A2-10.5).
func (k EdgeKind) AllowsEndpoints(source, target NodeKind) bool {
	switch k {
	case EdgeReachable:
		return (source == KindHost || source == KindNetwork) &&
			(target == KindHost || target == KindNetwork || target == KindService)
	case EdgeAuthenticatesTo:
		return (source == KindCredential || source == KindIdentity) &&
			(target == KindHost || target == KindService || target == KindShare)
	case EdgeMemberOf:
		return (source == KindIdentity || source == KindGroup) &&
			target == KindGroup
	case EdgeGrantsAccess:
		return (source == KindGroup || source == KindIdentity) &&
			(target == KindShare || target == KindHost || target == KindService)
	case EdgeExploitedBy:
		// Source: every kind except evidence_ref, finding and hypothesis
		// (A2-3.2's row lists the seven); target: the claim (§6 item 6).
		return (source == KindHost || source == KindNetwork || source == KindService ||
			source == KindIdentity || source == KindGroup || source == KindCredential ||
			source == KindShare) &&
			(target == KindFinding || target == KindHypothesis)
	case EdgeContradicts:
		return (source == KindFinding || source == KindHypothesis) &&
			(target == KindFinding || target == KindHypothesis)
	case EdgeSupersedes:
		return source == target && source.Valid()
	default:
		return false
	}
}
