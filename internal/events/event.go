package events

// Payload is the interface every per-kind payload struct satisfies: a flat
// object (A1-4.8) with a closed key set (A1-4.1) and the Kind it belongs to.
//
// Validate carries the A1-4.2 per-kind obligations and the A1-4.5 caps
// (errs kinds "validation" / "summary_too_large"). Per the §4.1 sketch the
// method is part of this interface, but its implementations belong to the
// composition/validation pass (A1-4.2–4.7, A1-4.9, A1-7.10): no payload
// struct implements Validate yet, so no struct in this package satisfies
// Payload yet. That is deliberate — a Validate that returns nil would be a
// stub of a safety path (AGENTS.md), and half of it (secret scan, reference
// resolution) cannot even be written inside a types-only package.
type Payload interface {
	Kind() Kind
	Validate() error
}

// ChainBlock groups the three chain fields. It is EMBEDDED in Event so the
// fields serialize at the envelope's top level (A1-1.3): A0-2.12 exclusion
// lists take plain top-level names only, so a nested "chain":{…} object
// would be unexcludable and is forbidden. All three are excluded from the
// digest (A1-5.2) and assigned by the store seam inside the append
// transaction (A1-5.4, A1-7.8) — never by a client (A1-2.3).
type ChainBlock struct {
	Seq      int64  `json:"seq"`       // per-engagement position (A1-5.4)
	PrevHash string `json:"prev_hash"` // hash of seq-1, or the zero constant at genesis (A1-5.3/5.5)
	Hash     string `json:"hash"`      // SHA-256 over the canonical preimage (A1-5.2)
}

// Event is the one envelope: 17 top-level keys, all always present with
// their zero value when unset (A1-1.1, A1-1.2, A0-2.14). No omitempty
// anywhere — a missing key changes the canonical bytes and therefore the
// digest. Timestamps are strings holding A0-5.1 values, never time.Time
// (A1-4.12). EvidenceRefs is sorted ascending by unsigned byte value and
// deduplicated at composition (A1-4.7) and initialized to empty, never nil,
// before marshaling (A1-1.2); Untrusted is platform-computed per A1-4.4 and
// MUST NOT be accepted from a client (A1-2.3).
//
// Event is a canonicalized type but is composed exclusively through the
// NewEvent constructor of the composition path (A1-7.10, §4.1 sketch), which
// owns stamping, validation and the non-nil initialization of EvidenceRefs;
// this package deliberately provides no partial constructor (DESIGN §4:
// never hand out half-built values).
type Event struct {
	EventID           string   `json:"event_id"`            // evt_ (A0-1.2)
	EngagementID      string   `json:"engagement_id"`       // eng_, chain scope (A1-5.1); from token binding, never from the body (A1-2.5)
	RunID             string   `json:"run_id"`              // "" when the event is not run-scoped
	JobID             string   `json:"job_id"`              // orchestrator container; "" when not applicable
	TaskID            string   `json:"task_id"`             // worker container; "" when not applicable
	NodeID            string   `json:"node_id"`             // remote agent node slp_node_ (Q9); "" for platform-local execution; never a graph node (A0-3.6)
	OccurredAt        string   `json:"occurred_at"`         // platform stamp (A1-1.4, A0-5.1)
	OccurredClaimedAt string   `json:"occurred_claimed_at"` // untrusted client claim (A0-5.7); "" when none supplied
	RecordedAt        string   `json:"recorded_at"`         // authoritative ingest time (A0-5.6)
	Actor             Actor    `json:"actor"`               // three keys, fixed (A1-2.1)
	Kind              Kind     `json:"kind"`                // closed taxonomy (A1-3.1)
	Payload           Payload  `json:"payload"`             // one flat type per kind (A1-4.1, A1-4.8)
	EvidenceRefs      []string `json:"evidence_refs"`       // evi_, sorted + deduped (A1-4.7); [] when none
	Untrusted         bool     `json:"untrusted"`           // platform-computed (A1-4.4)
	ChainBlock                 // seq, prev_hash, hash — top level (A1-1.3)
}
