package graph

// This file holds the closed enum vocabularies of A2: the node kinds (A2-2.1),
// the edge kinds (A2-3.1) and the per-field value enums (A2-2.4, A2-2.5,
// A2-2.6, A2-5.3, A2-5.6, A2-8.1). Every enum is a string type plus exported
// constants, never an int enum, so values stay readable in logs, approval
// views and reports (A0-8.5). Comparison is byte-exact: no case folding, no
// synonyms, no alias lists.

// NodeKind is one of the closed node kinds of A2-2.1. An unknown kind on write
// is hard-rejected with validation (A0-6.3, Q3); on read a client preserves an
// unknown kind verbatim and skips the item.
type NodeKind string

// The closed A2-2.1 node-kind list, in contract table order (1–10).
const (
	KindHost        NodeKind = "host"
	KindNetwork     NodeKind = "network"
	KindService     NodeKind = "service"
	KindIdentity    NodeKind = "identity"
	KindGroup       NodeKind = "group"
	KindCredential  NodeKind = "credential"
	KindShare       NodeKind = "share"
	KindEvidenceRef NodeKind = "evidence_ref"
	KindFinding     NodeKind = "finding"
	KindHypothesis  NodeKind = "hypothesis"
)

// NodeKinds returns the closed A2-2.1 list in contract order. It returns a
// fresh slice on every call, so the closed list cannot be mutated through it
// (DESIGN §4: no package-level mutable state). The list exists so the closure
// of A2-2.1 is mechanically testable and enumerable for validation switches
// (A2-10.2 step 5, WP-15) rather than implied by scattered comparisons.
func NodeKinds() []NodeKind {
	return []NodeKind{
		KindHost, KindNetwork, KindService, KindIdentity, KindGroup,
		KindCredential, KindShare, KindEvidenceRef, KindFinding, KindHypothesis,
	}
}

// Valid reports whether k is one of the ten closed A2-2.1 kinds, byte-exactly
// (A0-8.5): no case folding, no synonyms.
func (k NodeKind) Valid() bool {
	switch k {
	case KindHost, KindNetwork, KindService, KindIdentity, KindGroup,
		KindCredential, KindShare, KindEvidenceRef, KindFinding, KindHypothesis:
		return true
	default:
		return false
	}
}

// EdgeKind is one of the closed edge kinds of A2-3.1. Edges are directed; an
// unknown edge kind on write is hard-rejected with validation (A0-6.3).
type EdgeKind string

// The closed A2-3.1 edge-kind list, in A2-3.2 endpoint-matrix row order.
const (
	EdgeReachable       EdgeKind = "reachable"
	EdgeAuthenticatesTo EdgeKind = "authenticates_to"
	EdgeMemberOf        EdgeKind = "member_of"
	EdgeGrantsAccess    EdgeKind = "grants_access"
	EdgeExploitedBy     EdgeKind = "exploited_by"
	EdgeContradicts     EdgeKind = "contradicts"
	EdgeSupersedes      EdgeKind = "supersedes"
)

// EdgeKinds returns the closed A2-3.1 list in contract order, fresh on every
// call (see NodeKinds).
func EdgeKinds() []EdgeKind {
	return []EdgeKind{
		EdgeReachable, EdgeAuthenticatesTo, EdgeMemberOf, EdgeGrantsAccess,
		EdgeExploitedBy, EdgeContradicts, EdgeSupersedes,
	}
}

// Valid reports whether k is one of the seven closed A2-3.1 kinds,
// byte-exactly (A0-8.5).
func (k EdgeKind) Valid() bool {
	switch k {
	case EdgeReachable, EdgeAuthenticatesTo, EdgeMemberOf, EdgeGrantsAccess,
		EdgeExploitedBy, EdgeContradicts, EdgeSupersedes:
		return true
	default:
		return false
	}
}

// Severity is the closed finding-severity scale of A2-2.5 (finding only).
// There is no CVSS field in v1; a numeric score may ride in attrs as the
// fixed-point integer cvss_v3_x10 (scale ×10, value range deliberately
// undeclared — declaring one later narrows an accepted set and needs an ADR).
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

// FindingStatus is the closed status enum of a finding node (A2-2.6). There is
// deliberately no superseded value: supersession is expressed by the supersedes
// edge and superseded_by_id only (A2-4), because a second encoding would be a
// second source of truth.
type FindingStatus string

const (
	FindingOpen       FindingStatus = "open"
	FindingConfirmed  FindingStatus = "confirmed"
	FindingRefuted    FindingStatus = "refuted"
	FindingRemediated FindingStatus = "remediated"
)

// HypothesisStatus is the closed status enum of a hypothesis node (A2-2.6).
type HypothesisStatus string

const (
	HypothesisOpen      HypothesisStatus = "open"
	HypothesisSupported HypothesisStatus = "supported"
	HypothesisRefuted   HypothesisStatus = "refuted"
)

// Confidence is the provenance evidence grade of A2-5.6 and ADR-0022: what
// kind of evidence backs a graph-content observation, tied to the A1 event
// that recorded it. It grades the evidence, not the author's optimism; there
// is no numeric form, no adjective certainty scale anywhere in A2, and no
// finding-level confidence field (A2-2.7, ADR-0022 Decision — the contract
// spelling of the grade is confidence, carried once per provenance entry).
type Confidence string

const (
	// ConfidenceObserved: directly present in captured tool output
	// referenced by event_id/evidence_ids (A2-5.6).
	ConfidenceObserved Confidence = "observed"
	// ConfidenceInferred: derived by reasoning from other graph content.
	ConfidenceInferred Confidence = "inferred"
	// ConfidenceVerified: reproduced by a second, independent observation.
	// Stored only through A2-4.7's independence rule (WP-15's dedup
	// collapse), so no single observation can assert it.
	ConfidenceVerified Confidence = "verified"
)

// known reports whether c is one of the three closed A2-5.6 grades,
// byte-exactly (A0-8.5).
func (c Confidence) known() bool {
	switch c {
	case ConfidenceObserved, ConfidenceInferred, ConfidenceVerified:
		return true
	default:
		return false
	}
}

// QuarantineReason is the closed A2-8.1 quarantine-state vocabulary
// (out_of_scope, blacklisted). The quarantine behaviour — derivation,
// propagation, recomputation and the QuarantineDecider seam — is WP-16's
// (A2-8); this package carries only the enum the Node field is typed with.
type QuarantineReason string

const (
	QuarantineOutOfScope QuarantineReason = "out_of_scope"
	QuarantineBlacklist  QuarantineReason = "blacklisted"
)

// PrincipalKind is the closed principal vocabulary of A2-5.3 — A1-2.1's list
// verbatim (one vocabulary, two documents): platform, orchestrator, worker,
// node, user. operator is renamed user and operator_id is renamed user_id
// platform-wide (AM-1); the prose word "operator" is unaffected. principal_kind
// is the principal whose work produced the content, never the A1 event actor
// (A2-5.3: copying one into the other is a defect).
type PrincipalKind string

const (
	PrincipalPlatform     PrincipalKind = "platform"
	PrincipalOrchestrator PrincipalKind = "orchestrator"
	PrincipalWorker       PrincipalKind = "worker"
	PrincipalNode         PrincipalKind = "node" // remote agent (Q9): required, A2-5.3
	PrincipalUser         PrincipalKind = "user" // was "operator"; prose "operator" unchanged
)

// known reports whether p is one of the five closed A2-5.3 principal kinds,
// byte-exactly (A0-8.5).
func (p PrincipalKind) known() bool {
	switch p {
	case PrincipalPlatform, PrincipalOrchestrator, PrincipalWorker, PrincipalNode, PrincipalUser:
		return true
	default:
		return false
	}
}

// CredentialKind is the closed A2-2.4 class-of-material enum of a credential
// node. It names the class of the captured material, never the material: the
// graph holds references, the bytes live only in the evidence store (A2-9.1,
// A2-9.2).
type CredentialKind string

const (
	CredentialPassword    CredentialKind = "password"
	CredentialHash        CredentialKind = "hash"
	CredentialTicket      CredentialKind = "ticket"
	CredentialKey         CredentialKind = "key"
	CredentialToken       CredentialKind = "token"
	CredentialCertificate CredentialKind = "certificate"
	CredentialOther       CredentialKind = "other"
)

// MediaKind is the closed A2-2.4 artifact-class enum of an evidence_ref node.
type MediaKind string

const (
	MediaFile       MediaKind = "file"
	MediaScreenshot MediaKind = "screenshot"
	MediaCapture    MediaKind = "capture"
	MediaDump       MediaKind = "dump"
	MediaLog        MediaKind = "log"
	MediaOther      MediaKind = "other"
)
