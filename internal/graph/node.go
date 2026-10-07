package graph

// Node is one graph node (A2-2). Fields are immutable except the quarantine
// and report flags (A2-1.3, A2-8); the only mutation surface is the store
// seam's four methods (A2-10.6, WP-15/WP-17). Fields are EXPORTED with the
// A2 §4 sketch's json tags: encoding/json cannot marshal unexported fields,
// and NewNode (WP-15) is the only documented construction path, with the seam
// re-validating every value it is given (A2-10.6).
//
// Field naming follows A2-1.6's published mapping: the served id is
// graph_node_id, never a bare id and never node_id — node_id is reserved
// platform-wide for the remote agent node (slp_node_, A0-3.6, Q9). A bare
// node_id MUST NOT appear in any graph document (TestNoBareNodeIDInGraphDocuments).
//
// JSON presence semantics (A0-8.3): optional content fields are absent when
// unset (omitempty), never null; quarantined, report_excluded and provenance
// are always present because false/[] are values, not absence (A2-2.2,
// A2-8.1, A2-8.7). The API representation keeps these absence semantics; only
// the fingerprint document (content.go) serializes a fixed key set with zero
// values (A0-2.14, A2-4.6).
//
// There is no confidence field on Node: Q2's "Finding carries confidence" is
// discharged by the mandatory provenance grade (A2-2.7, ADR-0022). There is
// no time.Time anywhere in this type: timestamps are A0-5.1 strings.
type Node struct {
	ID           string   `json:"graph_node_id"` // gn_ (A0-1.2, A2-1.6), platform-generated (A0-1.4)
	EngagementID string   `json:"engagement_id"` // platform-stamped from the request binding (A2-1.5)
	GraphSeq     int64    `json:"graph_seq"`     // immutable per-engagement order key (A2-1.4, A2-1.4a)
	Kind         NodeKind `json:"kind"`

	Label       string   `json:"label"`                  // required on all kinds (A2-2.2)
	Summary     string   `json:"summary,omitempty"`      // required on finding (A2-2.4 semantics; WP-15 enforces)
	Attrs       Attrs    `json:"attrs,omitempty"`        // bounded flat escape hatch (A2-6)
	EvidenceIDs []string `json:"evidence_ids,omitempty"` // evi_ (A2-9.2)

	// Kind-specific fields (A2-2.3); a field that does not apply to Kind MUST
	// be absent — the write validation (WP-15, A2-10.2 step 6) hard-rejects a
	// not-applicable field.
	Addresses      []string       `json:"addresses,omitempty"`       // host, network
	CIDR           string         `json:"cidr,omitempty"`            // network (net.ParseCIDR — WP-15)
	Port           int            `json:"port,omitempty"`            // service, 1..65535
	Transport      string         `json:"transport,omitempty"`       // service: tcp|udp
	Protocol       string         `json:"protocol,omitempty"`        // service
	SID            string         `json:"sid,omitempty"`             // identity, group
	Domain         string         `json:"domain,omitempty"`          // identity, group, credential, share
	CredentialKind CredentialKind `json:"credential_kind,omitempty"` // credential
	EvidenceID     string         `json:"evidence_id,omitempty"`     // credential, evidence_ref
	MediaKind      MediaKind      `json:"media_kind,omitempty"`      // evidence_ref
	SizeBytes      int64          `json:"size_bytes,omitempty"`      // evidence_ref, ≥ 1 (A0-8.2)
	Severity       Severity       `json:"severity,omitempty"`        // finding
	Claim          string         `json:"claim,omitempty"`           // hypothesis
	Basis          string         `json:"basis,omitempty"`           // hypothesis

	// Status is FindingStatus or HypothesisStatus per Kind (A2-2.6): one JSON
	// field, two closed enums, validated by the kind switch (WP-15).
	Status string `json:"status,omitempty"`

	ContentHash      string           `json:"content_hash"`                // A2-4.6, 64-char lowercase hex (A0-8.7)
	Quarantined      bool             `json:"quarantined"`                 // A2-8.1, always present
	QuarantineReason QuarantineReason `json:"quarantine_reason,omitempty"` // absent when not quarantined (A2-8.1, A0-8.3)
	ReportExcluded   bool             `json:"report_excluded"`             // A2-8.7, always present
	SupersedesID     string           `json:"supersedes_id,omitempty"`     // A2-4.2: set by the revising write
	SupersededByID   string           `json:"superseded_by_id,omitempty"`  // A2-4.2: platform-set derived pointer
	Provenance       []Provenance     `json:"provenance"`                  // ≤ 8 entries, by A1 event seq (A2-4.7, A2-5.1)
}

// Current reports whether n is the current revision (A2-1.8): a node is
// current iff superseded_by_id is absent. "Current" is a defined term derived
// from that one field, never a stored flag — one source of truth.
func (n Node) Current() bool { return n.SupersededByID == "" }
