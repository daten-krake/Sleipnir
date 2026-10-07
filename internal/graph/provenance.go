package graph

import (
	"github.com/daten-krake/sleipnir/internal/errs"
	"github.com/daten-krake/sleipnir/internal/ids"
	"github.com/daten-krake/sleipnir/internal/timex"
)

// ProvenanceMaxEntries is the A2-local bound on one provenance list
// (A2-4.7, A2-5.1, A2-7.1 mechanism R). It is defined here and not in
// internal/caps because it is not a row of the A0-7.1 registry; the four
// other constants the A2 §4 sketch's const block lists (NodeLabelMaxBytes,
// HypothesisClaimMaxBytes, HypothesisBasisMaxBytes, AddressMaxBytes) ARE
// registry rows owned by internal/caps, and redefining them here would be a
// second definition of a registry value — a defect per A0-7.2.
const ProvenanceMaxEntries = 8

// Provenance is one entry of the mandatory provenance list (A2-5.1, A2-4.7:
// at most ProvenanceMaxEntries entries, ordered by the seq of event_id) and
// is platform-stamped only (A2-5.2): the write request types (WP-15) contain
// no provenance field, so a client body carrying one is an unknown field and
// is rejected. Absent optional fields are omitted, never null (A0-8.3).
//
// Timestamps are A0-5.1 strings, never time.Time. The slp_node_ field is
// agent_node_id — never node_id, which is reserved for the A1 envelope and
// error/log attrs spelling (A2-1.6, A0-3.6). Confidence is the A2-5.6
// evidence grade (ADR-0022), carried once per entry; a node's or edge's
// grade is the highest in its list, and verified is stored only through
// A2-4.7's independence rule — both are the dedup collapse's job (WP-15).
type Provenance struct {
	PrincipalKind     PrincipalKind `json:"principal_kind"`
	RunID             string        `json:"run_id"`
	JobID             string        `json:"job_id,omitempty"`
	TaskID            string        `json:"task_id,omitempty"`
	AgentNodeID       string        `json:"agent_node_id,omitempty"`       // slp_node_ (Q9), never "node_id" (A2-1.6)
	UserID            string        `json:"user_id,omitempty"`             // usr_ (A0-1.2, AM-1); was operator_id
	ToolID            string        `json:"tool_id,omitempty"`             // tool_ (A0-1.3)
	ToolVersion       string        `json:"tool_version,omitempty"`        // ≤ ToolVersionMaxBytes (64, A0-7.1; cap enforced by WP-15)
	EventID           string        `json:"event_id"`                      // originating A1 event (A2-5.4)
	RecordedAt        string        `json:"recorded_at"`                   // platform time, A0-5.1/5.4
	ObservedClaimedAt string        `json:"observed_claimed_at,omitempty"` // untrusted client-claimed time (A0-5.7)
	Confidence        Confidence    `json:"confidence"`
}

// ValidateProvenance checks one record's provenance list against the
// invariants this package owns: the list is non-empty (A2-5.1 — a record
// without provenance MUST NOT be stored; there is no "best-effort
// provenance" mode, A2-5.7) and holds at most ProvenanceMaxEntries entries
// (A2-4.7; the bound is an A2-7.1 mechanism-R cap, so exceeding it is
// summary_too_large, A0-7.6), and every entry passes Provenance.Validate.
//
// Two list-level rules are deliberately NOT checked here because a type
// cannot see the data they need: the ordering by the seq of event_id
// (A2-4.7 — the seq lives in the A1 event log) and the verified-independence
// rule (A2-5.6 — a property of the dedup collapse). Both belong to the
// ingest/write path (WP-15/WP-18).
func ValidateProvenance(list []Provenance) error {
	if len(list) == 0 {
		return errs.Newf(errs.Validation,
			"validating graph provenance: list is empty: every node and edge MUST carry at least one entry (A2-5.1)")
	}
	if len(list) > ProvenanceMaxEntries {
		return errs.Newf(errs.SummaryTooLarge,
			"validating graph provenance: field=provenance cap=%d actual=%d entries (A2-4.7, A2-7.1 mechanism R)",
			ProvenanceMaxEntries, len(list))
	}
	for i, p := range list {
		if err := p.Validate(); err != nil {
			return errs.Wrapf(err, "validating graph provenance entry %d of %d", i, len(list))
		}
	}
	return nil
}

// Validate checks one provenance entry against A2-5.3's field table, in the
// table's own field order — so the first violated rule is the one reported
// (A2-10.2's determinism, one entry deep): the closed principal_kind enum,
// required run_id, the conditional principal ids (job_id required for
// orchestrator and worker; task_id required for worker; agent_node_id
// required for node; user_id required for user), the tool pair (tool_version
// required with tool_id), required event_id, A0-5 timestamp shape for
// recorded_at and observed_claimed_at, and the closed confidence grade.
// Every id that is present is validated byte-exactly for its A0-1.2 kind
// (A0-1.5: no Crockford normalization, no case folding) — including a
// graph-node id smuggled into agent_node_id, the conflation A0-3.6 forbids.
// Every rejection is validation with a message naming the field (A2-5.7,
// A2-10.4).
//
// The conditionals are "required when", never "forbidden otherwise": §4.1's
// frozen example shows a platform entry carrying a job_id.
//
// Validate does NOT check: size caps (A2-7 — the validation pass, WP-15,
// which measures in A2-10.2's step order; that is why the 64-byte
// tool_version cap is not applied here), secret material (A2-5.8/A2-9.4 —
// the caller-side scan is WP-15's), that event_id resolves in this
// engagement (A2-5.4 — needs the store; a well-formed but unresolvable id is
// notfound, decided by the write path), and grade-raising rules (A2-5.6 —
// the dedup collapse).
func (p Provenance) Validate() error {
	if !p.PrincipalKind.known() {
		return errs.Newf(errs.Validation,
			"validating graph provenance entry: field=principal_kind: %d bytes: not one of the five closed A2-5.3 kinds (platform, orchestrator, worker, node, user) — a rejected value is never echoed (A2-5.8)",
			len(p.PrincipalKind))
	}
	if err := requireID("run_id", ids.Run, p.RunID); err != nil {
		return err
	}
	if err := p.validateConditionalIDs(); err != nil {
		return err
	}
	if err := requireID("event_id", ids.Event, p.EventID); err != nil {
		return err
	}
	return p.validateTimestampsAndGrade()
}

// validateConditionalIDs checks A2-5.3's conditional principal ids and the
// tool pair, in the field table's order (the first violated rule is the one
// reported — A2-10.2's determinism).
func (p Provenance) validateConditionalIDs() error {
	// job_id — required when principal_kind ∈ {orchestrator, worker}.
	if (p.PrincipalKind == PrincipalOrchestrator || p.PrincipalKind == PrincipalWorker) && p.JobID == "" {
		return missingField("job_id", "principal_kind="+string(p.PrincipalKind))
	}
	if err := optionalID("job_id", ids.Job, p.JobID); err != nil {
		return err
	}
	// task_id — required when principal_kind = worker.
	if p.PrincipalKind == PrincipalWorker && p.TaskID == "" {
		return missingField("task_id", "principal_kind=worker")
	}
	if err := optionalID("task_id", ids.Task, p.TaskID); err != nil {
		return err
	}
	// agent_node_id — required for a remote-agent principal; without it such
	// an observation has no attributable principal (A2-5.3, Q9).
	if p.PrincipalKind == PrincipalNode && p.AgentNodeID == "" {
		return missingField("agent_node_id", "principal_kind=node (a remote-agent observation MUST be attributable)")
	}
	if err := optionalID("agent_node_id", ids.AgentNode, p.AgentNodeID); err != nil {
		return err
	}
	// user_id — required when principal_kind = user (AM-1: usr_).
	if p.PrincipalKind == PrincipalUser && p.UserID == "" {
		return missingField("user_id", "principal_kind=user")
	}
	if err := optionalID("user_id", ids.KindUser, p.UserID); err != nil {
		return err
	}
	// tool_id / tool_version — the registry pair (A0-1.3: name and version
	// are never folded into the id). An orphan tool_version is rejected: the
	// Req column reads "with tool_id", and a version string without its
	// registry id attributes nothing (ruling recorded in doc.go).
	if err := optionalID("tool_id", ids.Tool, p.ToolID); err != nil {
		return err
	}
	if p.ToolID != "" && p.ToolVersion == "" {
		return missingField("tool_version", "tool_id is set (A2-5.3: the registry version string accompanies the id)")
	}
	if p.ToolID == "" && p.ToolVersion != "" {
		return errs.Newf(errs.Validation,
			"validating graph provenance entry: field=tool_version is set without field=tool_id: a version string without its registry id attributes nothing (A2-5.3, A0-1.3)")
	}
	return nil
}

// validateTimestampsAndGrade checks the trailing three A2-5.3 fields — the
// two timestamps (through timex, A0-5.3) and the closed confidence grade —
// in the field table's order.
func (p Provenance) validateTimestampsAndGrade() error {
	if err := checkTimestamp("recorded_at", p.RecordedAt, true); err != nil {
		return err
	}
	if err := checkTimestamp("observed_claimed_at", p.ObservedClaimedAt, false); err != nil {
		return err
	}
	if !p.Confidence.known() {
		return errs.Newf(errs.Validation,
			"validating graph provenance entry: field=confidence: %d bytes: not one of the three closed A2-5.6 evidence grades (observed, inferred, verified) — there is no adjective scale (ADR-0022); a rejected value is never echoed (A2-5.8)",
			len(p.Confidence))
	}
	return nil
}

// optionalID shape-validates an id field that is present; an absent optional
// id is valid here (the required-when conditions are checked by the caller,
// in A2-5.3's table order).
func optionalID(field string, kind ids.Kind, id string) error {
	if id != "" && !ids.Valid(kind, id) {
		return badID(field, kind, id)
	}
	return nil
}

// requireID rejects an absent or malformed required provenance id. A
// rejected value is never echoed: A0-1.7's "ids are not secrets" covers
// values *established* as ids, and a value that failed ids.Valid is not
// established — it may be secret material (A2-5.8: provenance is never
// echoed into an error message beyond ids and field names).
func requireID(field string, kind ids.Kind, id string) error {
	if id == "" {
		return missingField(field, "required (A2-5.3)")
	}
	if !ids.Valid(kind, id) {
		return badID(field, kind, id)
	}
	return nil
}

func missingField(field, why string) error {
	return errs.Newf(errs.Validation,
		"validating graph provenance entry: field=%s is missing: %s", field, why)
}

// badID rejects a value that failed byte-exact id validation for the field's
// declared A0-1.2 kind, naming the field, the expected kind and the value's
// byte length — never the value (see requireID; the same discipline
// checkTimestamp applies to oversized timestamps).
func badID(field string, kind ids.Kind, id string) error {
	return errs.Newf(errs.Validation,
		"validating graph provenance entry: field=%s: %d bytes: not a byte-exact %s id (A0-1.5: no normalization, no case folding)",
		field, len(id), string(kind))
}

// checkTimestamp validates one A0-5.1 timestamp field through timex
// (rejects, never normalizes — A0-5.3). The length pre-check bounds what
// timex's message can echo: the A0-5.1 form is exactly 24 bytes, so a value
// of any other length is rejected here naming the field and the length only,
// and only a 24-byte value reaches the parse whose message quotes it.
func checkTimestamp(field, s string, required bool) error {
	if s == "" {
		if required {
			return missingField(field, "required (A2-5.3)")
		}
		return nil
	}
	const a051Len = len("2006-01-02T15:04:05.000Z") // the A0-5.1 form is exactly 24 bytes
	if len(s) != a051Len {
		return errs.Newf(errs.Validation,
			"validating graph provenance entry: field=%s: length %d is not the 24-byte A0-5.1 form YYYY-MM-DDTHH:MM:SS.mmmZ",
			field, len(s))
	}
	if _, err := timex.ParseTime(s); err != nil {
		return errs.Wrapf(err, "validating graph provenance entry: field=%s", field)
	}
	return nil
}
