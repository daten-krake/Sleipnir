package events

// VerifyTrigger is the closed trigger enum of chain_verified (A1-3.3,
// A1-6.2): startup | pre_export | on_demand.
type VerifyTrigger string

const (
	TriggerStartup   VerifyTrigger = "startup"
	TriggerPreExport VerifyTrigger = "pre_export"
	TriggerOnDemand  VerifyTrigger = "on_demand"
)

// BreakKind is the closed break_kind enum of chain_break_detected (A1-3.3,
// A1-6.3), including head_regression from the A1-5.8 out-of-band head trail.
type BreakKind string

const (
	BreakPreimageMismatch   BreakKind = "preimage_mismatch"
	BreakLinkMismatch       BreakKind = "link_mismatch"
	BreakSeqGap             BreakKind = "seq_gap"
	BreakSeqDisorder        BreakKind = "seq_disorder"
	BreakDuplicateEventID   BreakKind = "duplicate_event_id"
	BreakGenesisInvalid     BreakKind = "genesis_invalid"
	BreakRowCountMismatch   BreakKind = "row_count_mismatch"
	BreakEngagementMismatch BreakKind = "engagement_mismatch"
	BreakHeadRegression     BreakKind = "head_regression" // A1-5.8, A1-6.3
)

// ChainGenesisPayload is the payload of kind chain_genesis (A1-3.3,
// A1-5.3): the integrity anchor at seq 0 of every engagement chain.
// ChainSpec must equal the A1-5.3 constant byte-exactly (checked by the
// composition path, A1-4.2).
type ChainGenesisPayload struct {
	ChainSpec string `json:"chain_spec"`
}

// Kind returns KindChainGenesis.
func (ChainGenesisPayload) Kind() Kind { return KindChainGenesis }

// ChainVerifiedPayload is the payload of kind chain_verified (A1-3.3,
// A1-6.2): the record of one successful verification walk. HeadSeq and
// HeadHash are the chain-head state at the end of the walk; VerifiedCount is
// the number of events walked.
type ChainVerifiedPayload struct {
	Trigger       VerifyTrigger `json:"trigger"`
	HeadSeq       int64         `json:"head_seq"`
	HeadHash      string        `json:"head_hash"` // 64-hex (A0-2.15)
	VerifiedCount int64         `json:"verified_count"`
	DurationMS    int64         `json:"duration_ms"`
}

// Kind returns KindChainVerified.
func (ChainVerifiedPayload) Kind() Kind { return KindChainVerified }

// ChainBreakDetectedPayload is the payload of kind chain_break_detected
// (A1-3.3, A1-6.3): the first break in seq order. ExpectedPrevHash and
// ActualPrevHash are both always present and equal when the break is not a
// link break (A1-4.2). BreakEventID is "" when the row is missing.
type ChainBreakDetectedPayload struct {
	BreakKind        BreakKind `json:"break_kind"`
	BreakSeq         int64     `json:"break_seq"`
	BreakEventID     string    `json:"break_event_id"`
	ExpectedPrevHash string    `json:"expected_prev_hash"` // 64-hex (A0-2.15)
	ActualPrevHash   string    `json:"actual_prev_hash"`   // 64-hex (A0-2.15)
	VerifiedCount    int64     `json:"verified_count"`
}

// Kind returns KindChainBreakDetected.
func (ChainBreakDetectedPayload) Kind() Kind { return KindChainBreakDetected }

// IntegrityOverridePayload is the payload of kind integrity_override
// (A1-3.3, A1-6.5): the admin-only, reasoned human decision that released an
// export or an internal view despite a known break. Reason must be non-empty
// (A1-4.2: never silent) and is untrusted-for-rendering human prose
// (A1-4.4). Scope is the closed enum export | internal_view; the two are not
// interchangeable (A1-6.4/6.5).
type IntegrityOverridePayload struct {
	Reason       string `json:"reason"` // ≤ 512 B, untrusted (A1-4.4/4.5)
	BreakEventID string `json:"break_event_id"`
	Scope        string `json:"scope"` // enum export|internal_view
}

// Kind returns KindIntegrityOverride.
func (IntegrityOverridePayload) Kind() Kind { return KindIntegrityOverride }

// ArtifactReleasedPayload is the payload of kind artifact_released
// (A1-3.3, A1-6.6, ADR-0021): the export stamp as a chained event, composed
// for every released customer-facing artifact. IntegrityState is the A1-6.6
// export enum (verified | failed_overridden) — deliberately NOT the A1-6.4
// chain-state vocabulary (TestExportIntegrityStateIsNotTheChainStateEnum).
// OverrideEventID must be non-empty iff IntegrityState is
// "failed_overridden" (A1-4.2).
type ArtifactReleasedPayload struct {
	ArtifactKind       string `json:"artifact_kind"` // enum report_html|report_pdf|findings_json|evidence_bundle|verification_bundle
	ArtifactEvidenceID string `json:"artifact_evidence_id"`
	HeadSeq            int64  `json:"head_seq"`
	HeadHash           string `json:"head_hash"`       // 64-hex (A0-2.15)
	IntegrityState     string `json:"integrity_state"` // enum verified|failed_overridden (A1-6.6)
	OverrideEventID    string `json:"override_event_id"`
	RecipientRef       string `json:"recipient_ref"` // ≤ 128 B label, never a person's name (A0-3.7)
}

// Kind returns KindArtifactReleased.
func (ArtifactReleasedPayload) Kind() Kind { return KindArtifactReleased }
