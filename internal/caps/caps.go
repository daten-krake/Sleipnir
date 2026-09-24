package caps

import "unicode/utf8"

// Size caps (A0-7.1) — the one registry: Q4 contract constants plus every
// per-contract cap A1 and A2 declare. Changeable only via ADR (A0-7.2); an
// owning contract cites these names and never restates a value.
const (
	// Q4 (A0-7.1).
	StageViewMaxBytes      = 64 * 1024 // 65536 B serialized JSON
	StageViewMaxNodes      = 500       // count, not bytes (A0-7.9)
	NodeSummaryMaxBytes    = 512       // B, UTF-8 of the decoded value
	FindingSummaryMaxBytes = 2 * 1024  // 2048 B
	StageSummaryMaxBytes   = 2 * 1024  // 2048 B

	// A1 (A1-4.5/4.7) — mechanism R everywhere.
	EventMaxCanonicalBytes = 32768 // platform invariant on one canonical event
	ProseLongMaxBytes      = 2048  // command, task_description, result_summary, revert_action
	ProseMediumMaxBytes    = 512   // reason, detail, action_summary, message, entry
	TargetMaxBytes         = 256   // target, attempted_target, blacklist_entry
	LabelMaxBytes          = 128   // origin, container_ref, network_name, media_type, ...
	ToolVersionMaxBytes    = 64    // tool_version; one value for A1 and A2 (A2's 32 is a defect)
	KindNameMaxBytes       = 32    // node_kind, edge_kind, risk_tier (A2/A7 own the values)
	DigestMaxBytes         = 256   // image_digest (A0-8.7)
	EvidenceRefsMax        = 8     // A1 evidence_refs / A2 evidence_ids count (was EvidenceIDsMax)
	EventRefsMax           = 64    // revert_event_ids / non_revertable_event_ids count
	ExitCodeMin            = -1    // -1 = no exit status (A1-4.2)
	ExitCodeMax            = 255
	IdempotencyKeyMaxBytes = 64 // A1-7.6 client dedup key

	// A2 (A2-6.4/7.1) — mechanism R everywhere.
	NodeLabelMaxBytes       = 128
	HypothesisClaimMaxBytes = 512
	HypothesisBasisMaxBytes = 1024
	AttrValueMaxBytes       = 512
	AttrsTotalMaxBytes      = 4096
	AttrsMaxKeys            = 16
	AttrKeyMaxBytes         = 40 // = A0-8.1's key regex bound (A2-6.2)
	AddressesMax            = 16
	AddressMaxBytes         = 64
	MaxSupersedeChain       = 64 // A2-4.4/4.5: bound on one history walk
)

// TruncationMarker is the literal ASCII marker mechanism T appends to a
// shortened value (A0-7.5). It is 11 bytes and it is counted against the cap,
// so a caller must never pass a limit below len(TruncationMarker). A0-7.5
// originally said "12 B" while naming this literal; the erratum is A0 §6 item
// 16 (product owner, 2026-09-24) and no boundary here hardcodes either number.
const TruncationMarker = "[truncated]" // A0-7.5; 11 B; counted against the cap

// Truncate applies mechanism T: rune-boundary cut + marker (A0-7.4/A0-7.5).
// It reports whether anything was cut, for the sibling <field>_truncated bool
// (A0-7.5). A value that already fits is returned byte-identical with false.
//
// A limit below len(TruncationMarker) is a platform defect (A0-7.5): Truncate
// returns ("", true) — an empty value with the marker set beats a value that
// silently exceeds its cap — and the caller surfaces errs.Internal (A0-3.1).
// Truncate never returns more than limit bytes and never splits a rune;
// invalid UTF-8 in s is cut back to the last rune start (A0-2.3 rejects such
// values upstream, so the result of a malformed input is best-effort).
func Truncate(s string, limit int) (out string, truncated bool) {
	if limit < len(TruncationMarker) {
		return "", true
	}
	if len(s) <= limit {
		return s, false
	}

	budget := limit - len(TruncationMarker)
	cut := budget
	// Walk back off a continuation byte so the prefix holds whole runes only
	// (A0-7.4). budget < len(s), so s[budget] is inside the string.
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	return s[:cut] + TruncationMarker, true
}

// Fits applies the measurement rule of A0-7.3 — the UTF-8 byte count of the
// decoded string value (len(s) in Go), never a rune count, never a UTF-16
// length; surrounding quotes and escapes do not count — and reports whether s
// is within limit bytes.
//
// This is the check a caller performs before applying mechanism R (A0-7.6);
// the rejection itself, its summary_too_large kind and its message naming the
// field, the cap and the actual byte count belong to the caller, not here.
func Fits(s string, limit int) bool {
	return len(s) <= limit
}
