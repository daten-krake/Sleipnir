package graph

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/daten-krake/sleipnir/internal/errs"
	"github.com/daten-krake/sleipnir/internal/ids"
)

// Fixture ids are the frozen contract's own examples (A2 §4.1, A0-1.2's
// table): valid A0-1.1 bodies, byte-exact for their prefixes.
const (
	fixtureRunID       = "run_01m1y2whfhnjx2am9103w0pnqw"
	fixtureJobID       = "job_01m1y2whfhbt69j0h0fbxepw90"
	fixtureTaskID      = "task_01m1y2whfh1txm57x8dn41r9hg"
	fixtureAgentNodeID = "slp_node_01m1y2whfhxydsaem68cmazyc8"
	fixtureEventID     = "evt_01m1y2whfhp17g0avdqztd2p3x"
	fixtureToolID      = "tool_01m1y2whfhfjdvwqp9pfxqekmf"
	fixtureUserID      = "usr_01m1y2whfhv3x6z9b2d5f8h1jk"
	fixtureRecordedAt  = "2026-09-07T14:03:22.481Z"
)

// baseEntry is a minimally valid platform entry (§4.1's quarantined-host
// provenance minus the optional fields).
func baseEntry() Provenance {
	return Provenance{
		PrincipalKind: PrincipalPlatform,
		RunID:         fixtureRunID,
		EventID:       fixtureEventID,
		RecordedAt:    fixtureRecordedAt,
		Confidence:    ConfidenceObserved,
	}
}

func TestProvenanceMandatory(t *testing.T) {
	t.Run("fixture_ids_are_byte_exact_valid", func(t *testing.T) {
		// Guard the fixtures themselves: if one rots, every case below would
		// pass or fail for the wrong reason.
		for _, f := range []struct {
			kind ids.Kind
			id   string
		}{
			{ids.Run, fixtureRunID}, {ids.Job, fixtureJobID}, {ids.Task, fixtureTaskID},
			{ids.AgentNode, fixtureAgentNodeID}, {ids.Event, fixtureEventID},
			{ids.Tool, fixtureToolID}, {ids.KindUser, fixtureUserID},
		} {
			if !ids.Valid(f.kind, f.id) {
				t.Fatalf("fixture id %q is not a valid %s id — fix the fixture", f.id, f.kind)
			}
		}
	})

	t.Run("fields_match_the_A2-5.3_table", func(t *testing.T) {
		// The closed field set, the JSON spellings and the required/optional
		// split are parsed from A2-5.3's own table, not transcribed from the
		// struct (pinning rule).
		text := contractText(t, contractRelPath)
		rows := parseProvenanceFields(t, contractSlice(t, text, "### A2-5", "### A2-6"))
		if len(rows) != 12 {
			t.Fatalf("A2-5.3 parses to %d field rows, want 12", len(rows))
		}
		typ := reflect.TypeOf(Provenance{})
		if typ.NumField() != len(rows) {
			t.Fatalf("Provenance has %d fields, A2-5.3's table has %d — the field set is closed", typ.NumField(), len(rows))
		}
		tags := make(map[string]string, typ.NumField()) // json name → full tag
		for i := 0; i < typ.NumField(); i++ {
			tag := typ.Field(i).Tag.Get("json")
			name, _, _ := strings.Cut(tag, ",")
			tags[name] = tag
		}
		for _, row := range rows {
			tag, ok := tags[row.name]
			if !ok {
				t.Errorf("A2-5.3 field %q has no json tag on Provenance", row.name)
				continue
			}
			delete(tags, row.name)
			required := row.req == "yes"
			optional := strings.HasSuffix(tag, ",omitempty")
			if required && optional {
				t.Errorf("field %q is required (Req=%q) but carries omitempty — a required field is always serialized (A0-8.3)", row.name, row.req)
			}
			if !required && !optional {
				t.Errorf("field %q is conditional (Req=%q) but lacks omitempty — an unset optional field is absent, never null (A0-8.3)", row.name, row.req)
			}
		}
		for name := range tags {
			t.Errorf("Provenance carries json name %q, which A2-5.3's closed table does not define", name)
		}
	})

	t.Run("always_serialized_on_node_and_edge", func(t *testing.T) {
		// A2-5.1: every node and edge carries provenance — the key is never
		// omitted from a served document (no omitempty), unlike the optional
		// content fields.
		for _, typ := range []reflect.Type{reflect.TypeOf(Node{}), reflect.TypeOf(Edge{})} {
			f, ok := typ.FieldByName("Provenance")
			if !ok {
				t.Fatalf("%s has no Provenance field", typ)
			}
			if tag := f.Tag.Get("json"); tag != "provenance" {
				t.Errorf("%s.Provenance json tag = %q, want exactly \"provenance\" with no omitempty (A2-5.1)", typ, tag)
			}
		}
		// A zero Node still emits the key (a real record always carries at
		// least one entry — ValidateProvenance below is what enforces that
		// on the write path; the served key is never absent).
		b, err := json.Marshal(Node{})
		if err != nil {
			t.Fatalf("marshal zero Node: %v", err)
		}
		if !strings.Contains(string(b), `"provenance":`) {
			t.Errorf("serialized Node lacks the provenance key: %s", b)
		}
	})

	t.Run("empty_list_rejected", func(t *testing.T) {
		// A2-5.1: a record without provenance MUST NOT be stored; A2-5.7:
		// there is no "best-effort provenance" mode.
		for _, list := range [][]Provenance{nil, {}} {
			err := ValidateProvenance(list)
			if err == nil {
				t.Fatalf("ValidateProvenance(%v) = nil, want rejection (A2-5.1)", list)
			}
			if kind := errs.KindOf(err); kind != errs.Validation {
				t.Errorf("empty provenance: kind = %q, want validation (A2-5.7)", kind)
			}
		}
	})

	t.Run("list_bound_is_eight", func(t *testing.T) {
		// A2-4.7: at most 8 entries; A2-7.1 assigns mechanism R, so over the
		// bound is summary_too_large (A0-7.6), naming the cap and the count.
		eight := make([]Provenance, ProvenanceMaxEntries)
		for i := range eight {
			eight[i] = baseEntry()
		}
		if err := ValidateProvenance(eight); err != nil {
			t.Errorf("8 entries rejected: %v — the bound is inclusive (A2-4.7)", err)
		}
		nine := append(eight, baseEntry())
		err := ValidateProvenance(nine)
		if err == nil {
			t.Fatal("9 entries accepted — the list is closed at ProvenanceMaxEntries (A2-4.7)")
		}
		if kind := errs.KindOf(err); kind != errs.SummaryTooLarge {
			t.Errorf("9 entries: kind = %q, want summary_too_large (A2-7.1 mechanism R, A0-7.6)", kind)
		}
		if msg := err.Error(); !strings.Contains(msg, "cap=8") || !strings.Contains(msg, "actual=9") {
			t.Errorf("message must name the cap and the actual count (A0-7.6): %v", err)
		}
	})

	t.Run("entry_failure_names_its_position", func(t *testing.T) {
		bad := baseEntry()
		bad.RunID = ""
		err := ValidateProvenance([]Provenance{baseEntry(), baseEntry(), bad})
		if err == nil {
			t.Fatal("list with a broken entry accepted")
		}
		if kind := errs.KindOf(err); kind != errs.Validation {
			t.Errorf("kind = %q, want validation", kind)
		}
		if msg := err.Error(); !strings.Contains(msg, "entry 2 of 3") || !strings.Contains(msg, "field=run_id") {
			t.Errorf("message must name the entry position and the field (ADR-0019 §2): %v", err)
		}
	})

	t.Run("required_entry_fields", func(t *testing.T) {
		// A2-5.3's Req=yes column: principal_kind, run_id, event_id,
		// recorded_at, confidence. Each missing → validation naming it.
		mutations := map[string]func(*Provenance){
			"principal_kind": func(p *Provenance) { p.PrincipalKind = "" },
			"run_id":         func(p *Provenance) { p.RunID = "" },
			"event_id":       func(p *Provenance) { p.EventID = "" },
			"recorded_at":    func(p *Provenance) { p.RecordedAt = "" },
			"confidence":     func(p *Provenance) { p.Confidence = "" },
		}
		for field, mutate := range mutations {
			p := baseEntry()
			mutate(&p)
			err := p.Validate()
			if err == nil {
				t.Errorf("missing %s accepted — it is required (A2-5.3)", field)
				continue
			}
			if kind := errs.KindOf(err); kind != errs.Validation {
				t.Errorf("missing %s: kind = %q, want validation", field, kind)
			}
			if msg := err.Error(); !strings.Contains(msg, "field="+field) {
				t.Errorf("missing %s: message must name the field (A2-10.4): %v", field, err)
			}
		}
	})

	t.Run("principal_kind_is_closed_and_renamed", func(t *testing.T) {
		// A2-5.3: A1-2.1's list verbatim — operator is RENAMED user, so the
		// old spelling must be rejected, and comparison is byte-exact
		// (A0-8.5).
		rejects := []PrincipalKind{"", "operator", "Platform", "PLATFORM", "platform ", " platform", "orchestrators", "nodes", "agents", "user\n"}
		for _, pk := range rejects {
			p := baseEntry()
			p.PrincipalKind = pk
			if err := p.Validate(); err == nil {
				t.Errorf("principal_kind %q accepted — the enum is closed (A2-5.3)", pk)
			} else if kind := errs.KindOf(err); kind != errs.Validation {
				t.Errorf("principal_kind %q: kind = %q, want validation", pk, kind)
			}
		}
		// Each of the five valid kinds passes with its conditional ids set
		// (the conditional rules themselves are the next subtest).
		valid := map[PrincipalKind]func(*Provenance){
			PrincipalPlatform:     func(p *Provenance) {},
			PrincipalOrchestrator: func(p *Provenance) { p.JobID = fixtureJobID },
			PrincipalWorker:       func(p *Provenance) { p.JobID = fixtureJobID; p.TaskID = fixtureTaskID },
			PrincipalNode:         func(p *Provenance) { p.AgentNodeID = fixtureAgentNodeID },
			PrincipalUser:         func(p *Provenance) { p.UserID = fixtureUserID },
		}
		for pk, setup := range valid {
			p := baseEntry()
			p.PrincipalKind = pk
			setup(&p)
			if err := p.Validate(); err != nil {
				t.Errorf("principal_kind %q with its ids rejected: %v", pk, err)
			}
		}
	})

	t.Run("confidence_is_the_closed_evidence_grade", func(t *testing.T) {
		// A2-5.6 / ADR-0022: observed · inferred · verified — the rejected
		// adjective scale (low/medium/high) and every near miss are out.
		rejects := []Confidence{"", "high", "medium", "low", "Observed", "VERIFIED", "verified ", "verifed", "observed\n", "certain"}
		for _, c := range rejects {
			p := baseEntry()
			p.Confidence = c
			if err := p.Validate(); err == nil {
				t.Errorf("confidence %q accepted — the grade is a closed three-value scale (A2-5.6, ADR-0022)", c)
			} else if kind := errs.KindOf(err); kind != errs.Validation {
				t.Errorf("confidence %q: kind = %q, want validation", c, kind)
			}
		}
		for _, c := range []Confidence{ConfidenceObserved, ConfidenceInferred, ConfidenceVerified} {
			p := baseEntry()
			p.Confidence = c
			if err := p.Validate(); err != nil {
				t.Errorf("grade %q rejected: %v", c, err)
			}
		}
	})

	t.Run("conditional_ids_per_principal_kind", func(t *testing.T) {
		// A2-5.3's Req column and its actor/id-shape table. "Required when"
		// never means "forbidden otherwise": §4.1's platform entry carries a
		// job_id and must stay valid.
		tests := []struct {
			name    string
			mutate  func(*Provenance)
			wantErr string // "" = valid; otherwise a substring the message must carry
		}{
			{"platform_minimal", func(p *Provenance) {}, ""},
			{"platform_with_job_id_per_4_1_example", func(p *Provenance) { p.JobID = fixtureJobID }, ""},
			{"orchestrator_without_job_id", func(p *Provenance) { p.PrincipalKind = PrincipalOrchestrator }, "field=job_id"},
			{"orchestrator_with_job_id", func(p *Provenance) { p.PrincipalKind = PrincipalOrchestrator; p.JobID = fixtureJobID }, ""},
			{"worker_without_job_id", func(p *Provenance) { p.PrincipalKind = PrincipalWorker; p.TaskID = fixtureTaskID }, "field=job_id"},
			{"worker_without_task_id", func(p *Provenance) { p.PrincipalKind = PrincipalWorker; p.JobID = fixtureJobID }, "field=task_id"},
			{"worker_full_per_4_1_example", func(p *Provenance) {
				p.PrincipalKind = PrincipalWorker
				p.JobID = fixtureJobID
				p.TaskID = fixtureTaskID
				p.AgentNodeID = fixtureAgentNodeID
				p.ToolID = fixtureToolID
				p.ToolVersion = "1.4.2"
				p.ObservedClaimedAt = "2026-09-07T14:03:19.900Z"
			}, ""},
			{"node_without_agent_node_id", func(p *Provenance) { p.PrincipalKind = PrincipalNode }, "field=agent_node_id"},
			{"node_with_agent_node_id", func(p *Provenance) { p.PrincipalKind = PrincipalNode; p.AgentNodeID = fixtureAgentNodeID }, ""},
			{"user_without_user_id", func(p *Provenance) { p.PrincipalKind = PrincipalUser }, "field=user_id"},
			{"user_with_user_id", func(p *Provenance) { p.PrincipalKind = PrincipalUser; p.UserID = fixtureUserID }, ""},
			{"tool_id_without_tool_version", func(p *Provenance) { p.ToolID = fixtureToolID }, "field=tool_version"},
			{"tool_version_without_tool_id", func(p *Provenance) { p.ToolVersion = "1.4.2" }, "tool_version"},
			{"tool_pair_complete", func(p *Provenance) { p.ToolID = fixtureToolID; p.ToolVersion = "1.4.2" }, ""},
		}
		for _, tc := range tests {
			p := baseEntry()
			tc.mutate(&p)
			err := p.Validate()
			if tc.wantErr == "" {
				if err != nil {
					t.Errorf("%s: valid entry rejected: %v", tc.name, err)
				}
				continue
			}
			if err == nil {
				t.Errorf("%s: accepted, want rejection naming %s", tc.name, tc.wantErr)
				continue
			}
			if kind := errs.KindOf(err); kind != errs.Validation {
				t.Errorf("%s: kind = %q, want validation", tc.name, kind)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("%s: message %q does not name %s", tc.name, err, tc.wantErr)
			}
		}
	})

	t.Run("ids_are_byte_exact_per_A0-1.5", func(t *testing.T) {
		// No Crockford normalization, no case folding, no wrong-prefix
		// leniency — and no conflation of the graph-node id with the remote
		// agent node id (A0-3.6: agent_node_id is slp_node_, never gn_).
		tests := []struct {
			name   string
			mutate func(*Provenance)
		}{
			{"run_id_uppercase_body", func(p *Provenance) { p.RunID = "run_01m1y2whfhnjx2am9103w0pnqW" }},
			{"run_id_short_body", func(p *Provenance) { p.RunID = "run_01m1y2whfh" }},
			{"run_id_wrong_prefix", func(p *Provenance) { p.RunID = "eng_01m1y2whfhgbz06ays6dxnvyws" }},
			{"run_id_forbidden_alphabet_char", func(p *Provenance) { p.RunID = "run_01m1y2whfhnjx2am9103w0pnqi" }},
			{"event_id_wrong_prefix", func(p *Provenance) { p.EventID = "evtX01m1y2whfhp17g0avdqztd2p3" }},
			{"event_id_trailing_space", func(p *Provenance) { p.EventID = "evt_01m1y2whfhp17g0avdqztd2p3x " }},
			{"agent_node_id_is_graph_node_id", func(p *Provenance) {
				p.PrincipalKind = PrincipalNode
				p.AgentNodeID = "gn_01m1y2whfhh039ykj5x8mc5a0g"
			}},
			{"agent_node_id_missing_slp_prefix", func(p *Provenance) {
				p.PrincipalKind = PrincipalNode
				p.AgentNodeID = "node_01m1y2whfhxydsaem68cmazyc8xx"
			}},
			{"job_id_is_task_id", func(p *Provenance) {
				p.PrincipalKind = PrincipalWorker
				p.JobID = fixtureTaskID
				p.TaskID = fixtureTaskID
			}},
			{"user_id_old_operator_spelling", func(p *Provenance) {
				p.PrincipalKind = PrincipalUser
				p.UserID = "operator_01m1y2whfhv3x6z9b2d5f8h1"
			}},
			{"tool_id_slug_form", func(p *Provenance) {
				p.ToolID = "tool_nmap"
				p.ToolVersion = "1.4.2"
			}},
		}
		for _, tc := range tests {
			p := baseEntry()
			tc.mutate(&p)
			err := p.Validate()
			if err == nil {
				t.Errorf("%s: accepted — ids validate byte-exactly against their A0-1.2 regex (A0-1.5)", tc.name)
				continue
			}
			if kind := errs.KindOf(err); kind != errs.Validation {
				t.Errorf("%s: kind = %q, want validation", tc.name, kind)
			}
		}
	})

	t.Run("timestamps_reject_never_normalize", func(t *testing.T) {
		// A0-5.1/5.3 through timex; A2-5.5: observed_claimed_at is stored
		// next to recorded_at but never trusted — shape-checked like
		// everything else.
		rejects := []string{
			"2026-09-07T14:03:22.481+02:00", // numeric offset
			"2026-09-07T14:03:22Z",          // no fractional digits
			"2026-09-07T14:03:22.4810Z",     // four fractional digits
			"2026-09-07 14:03:22.481Z",      // space separator
			"2026-09-07t14:03:22.481z",      // lowercase t/z
			"2026-09-07T14:03:60.481Z",      // leap second
			"2019-09-07T14:03:22.481Z",      // below the year window
			"2100-09-07T14:03:22.481Z",      // at/over the year window
			"2026-13-45T99:99:99.999Z",      // calendar nonsense, right width
			"not a timestamp at all!!",      // garbage (24 bytes: passes the length gate; timex's A0-5.1 regex rejects it)
		}
		for _, ts := range rejects {
			p := baseEntry()
			p.RecordedAt = ts
			if err := p.Validate(); err == nil {
				t.Errorf("recorded_at %q accepted (A0-5.3 rejects, never normalizes)", ts)
			} else if kind := errs.KindOf(err); kind != errs.Validation {
				t.Errorf("recorded_at %q: kind = %q, want validation", ts, kind)
			}
		}
		p := baseEntry()
		p.ObservedClaimedAt = "2026-09-07T14:03:19.900Z"
		if err := p.Validate(); err != nil {
			t.Errorf("valid observed_claimed_at rejected: %v", err)
		}
		p = baseEntry()
		p.ObservedClaimedAt = "2026-09-07T14:03:19.900+02:00"
		if err := p.Validate(); err == nil {
			t.Error("observed_claimed_at with a numeric offset accepted")
		}
	})

	t.Run("oversized_timestamp_is_not_echoed", func(t *testing.T) {
		// The A0-5.1 form is exactly 24 bytes; a longer value is rejected on
		// length alone so an oversized garbage field cannot flood an error
		// message (the value is never echoed whole).
		sentinel := strings.Repeat("x", 200)
		p := baseEntry()
		p.RecordedAt = sentinel
		err := p.Validate()
		if err == nil {
			t.Fatal("oversized recorded_at accepted")
		}
		if strings.Contains(err.Error(), sentinel) {
			t.Errorf("message echoes the oversized value: %v", err)
		}
		if !strings.Contains(err.Error(), "field=recorded_at") {
			t.Errorf("message must name the field: %v", err)
		}
	})

	t.Run("rejected_values_are_never_echoed", func(t *testing.T) {
		// A2-5.8: provenance is never echoed into an error message beyond ids
		// and field names — and a value that FAILED id/enum validation is not
		// an id (A0-1.7 covers established ids only). Regression pin for the
		// review-demonstrated leak: a secret-shaped run_id must not appear in
		// the rejection, in any of the three value-checking sites.
		const sentinel = "sk-live-ABCDEF1234567890abcdef"
		mutations := []struct {
			name   string
			mutate func(*Provenance)
			field  string
		}{
			{"run_id_secret_shaped", func(p *Provenance) { p.RunID = sentinel }, "field=run_id"},
			{"principal_kind_secret_shaped", func(p *Provenance) { p.PrincipalKind = PrincipalKind(sentinel) }, "field=principal_kind"},
			{"confidence_secret_shaped", func(p *Provenance) { p.Confidence = Confidence(sentinel) }, "field=confidence"},
			{"event_id_secret_shaped", func(p *Provenance) { p.EventID = sentinel }, "field=event_id"},
		}
		for _, m := range mutations {
			p := baseEntry()
			m.mutate(&p)
			err := p.Validate()
			if err == nil {
				t.Fatalf("%s: sentinel value accepted", m.name)
			}
			if kind := errs.KindOf(err); kind != errs.Validation {
				t.Errorf("%s: kind = %q, want validation", m.name, kind)
			}
			if strings.Contains(err.Error(), sentinel) {
				t.Errorf("%s: message echoes the rejected value: %v", m.name, err)
			}
			if strings.Contains(err.Error(), "SENTINEL") || strings.Contains(err.Error(), "sk-live") {
				t.Errorf("%s: message echoes a fragment of the rejected value: %v", m.name, err)
			}
			if !strings.Contains(err.Error(), m.field) {
				t.Errorf("%s: message must name the field: %v", m.name, err)
			}
		}
	})

	t.Run("wire_shape_is_exact", func(t *testing.T) {
		// The §4.1 worker entry, byte-exact: required keys always present,
		// unset optionals absent (never null, A0-8.3), struct order.
		full := Provenance{
			PrincipalKind:     PrincipalWorker,
			RunID:             fixtureRunID,
			JobID:             fixtureJobID,
			TaskID:            fixtureTaskID,
			AgentNodeID:       fixtureAgentNodeID,
			ToolID:            fixtureToolID,
			ToolVersion:       "1.4.2",
			EventID:           fixtureEventID,
			RecordedAt:        fixtureRecordedAt,
			ObservedClaimedAt: "2026-09-07T14:03:19.900Z",
			Confidence:        ConfidenceObserved,
		}
		b, err := json.Marshal(full)
		if err != nil {
			t.Fatalf("marshal full entry: %v", err)
		}
		want := `{"principal_kind":"worker","run_id":"` + fixtureRunID +
			`","job_id":"` + fixtureJobID +
			`","task_id":"` + fixtureTaskID +
			`","agent_node_id":"` + fixtureAgentNodeID +
			`","tool_id":"` + fixtureToolID +
			`","tool_version":"1.4.2","event_id":"` + fixtureEventID +
			`","recorded_at":"` + fixtureRecordedAt +
			`","observed_claimed_at":"2026-09-07T14:03:19.900Z","confidence":"observed"}`
		if string(b) != want {
			t.Errorf("full entry wire bytes\n got: %s\nwant: %s", b, want)
		}
		minimal := baseEntry()
		b, err = json.Marshal(minimal)
		if err != nil {
			t.Fatalf("marshal minimal entry: %v", err)
		}
		want = `{"principal_kind":"platform","run_id":"` + fixtureRunID +
			`","event_id":"` + fixtureEventID +
			`","recorded_at":"` + fixtureRecordedAt +
			`","confidence":"observed"}`
		if string(b) != want {
			t.Errorf("minimal entry wire bytes\n got: %s\nwant: %s", b, want)
		}
		if strings.Contains(string(b), "null") {
			t.Errorf("absent optional fields must be omitted, never null (A0-8.3): %s", b)
		}
	})

	t.Run("no_time_time_in_types", func(t *testing.T) {
		// Timestamps are A0-5.1 strings everywhere in the served types; a
		// time.Time field would make canonical bytes encoder-dependent
		// (A0-5.2's reasoning) — reflection gate per the WP-14 brief.
		for _, typ := range []reflect.Type{
			reflect.TypeOf(Provenance{}), reflect.TypeOf(Node{}), reflect.TypeOf(Edge{}),
		} {
			for i := 0; i < typ.NumField(); i++ {
				if typ.Field(i).Type == reflect.TypeOf(time.Time{}) {
					t.Errorf("%s.%s is a time.Time — timestamps are timex strings (A0-5.1)", typ, typ.Field(i).Name)
				}
			}
		}
	})
}
