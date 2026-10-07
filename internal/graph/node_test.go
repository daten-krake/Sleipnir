package graph

import (
	"bytes"
	"encoding/json"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/daten-krake/sleipnir/internal/ids"
)

// sampleEdge is §4.1's exploited_by edge: the finding example's target link,
// with a platform provenance entry.
func sampleEdge() Edge {
	return Edge{
		ID:           "ge_01m1y2whfh62ej11jf4x5gjzv4",
		EngagementID: "eng_01m1y2whfhgbz06ays6dxnvyws",
		GraphSeq:     413,
		Kind:         EdgeExploitedBy,
		SourceID:     "gn_01m1y2whfhdc01srv4x8mc5a0g",
		SourceKind:   KindHost,
		TargetID:     "gn_01m1y2whfhh039ykj5x8mc5a0g",
		TargetKind:   KindFinding,
		Retracted:    false,
		Provenance: []Provenance{{
			PrincipalKind: PrincipalPlatform,
			RunID:         fixtureRunID,
			JobID:         fixtureJobID,
			EventID:       fixtureEventID,
			RecordedAt:    "2026-09-07T14:03:22.502Z",
			Confidence:    ConfidenceInferred,
		}},
	}
}

func TestNoBareNodeIDInGraphDocuments(t *testing.T) {
	// A2-1.6 / A0-3.6 / P-79: one value, one name platform-wide — a graph
	// node is graph_node_id and a graph edge is graph_edge_id in every JSON
	// document; node_id is RESERVED for the remote agent node (slp_node_,
	// Q9) and is spelled agent_node_id in A2's payloads (A2-5.3). No graph
	// document carries a bare node_id — or a bare id — key.
	documentTypes := []reflect.Type{
		reflect.TypeOf(Node{}),
		reflect.TypeOf(Edge{}),
		reflect.TypeOf(Provenance{}),
		reflect.TypeOf(contentDoc{}),
		reflect.TypeOf(AttrValue{}),
	}

	t.Run("no_json_key_is_bare_node_id_or_bare_id", func(t *testing.T) {
		for _, typ := range documentTypes {
			for i := 0; i < typ.NumField(); i++ {
				f := typ.Field(i)
				tag := f.Tag.Get("json")
				name, _, _ := strings.Cut(tag, ",")
				if name == "node_id" {
					t.Errorf("%s.%s carries json key %q — node_id is the remote agent node, reserved platform-wide (A2-1.6, A0-3.6)", typ, f.Name, name)
				}
				if name == "id" {
					t.Errorf("%s.%s carries a bare json key %q — the served names are graph_node_id / graph_edge_id (A2-1.2, A2-1.6)", typ, f.Name, name)
				}
				if f.Name == "NodeID" {
					t.Errorf("%s has a Go field NodeID — a graph node is GraphNodeID/SourceID/TargetID, the agent node is AgentNodeID; the two must never be conflated (A0-3.6)", typ)
				}
			}
		}
	})

	t.Run("A2-1.6_field_mapping_is_exact", func(t *testing.T) {
		// The published A1↔A2 mapping table, pinned field by field.
		mapping := []struct {
			typ     reflect.Type
			field   string
			wantTag string
		}{
			{reflect.TypeOf(Node{}), "ID", "graph_node_id"},
			{reflect.TypeOf(Edge{}), "ID", "graph_edge_id"},
			{reflect.TypeOf(Edge{}), "SourceID", "source_id"},
			{reflect.TypeOf(Edge{}), "TargetID", "target_id"},
			{reflect.TypeOf(Node{}), "SupersedesID", "supersedes_id"},
			{reflect.TypeOf(Node{}), "SupersededByID", "superseded_by_id"},
			{reflect.TypeOf(Provenance{}), "AgentNodeID", "agent_node_id"},
		}
		for _, m := range mapping {
			f, ok := m.typ.FieldByName(m.field)
			if !ok {
				t.Errorf("%s has no field %s (A2-1.6 mapping)", m.typ, m.field)
				continue
			}
			tag := f.Tag.Get("json")
			got, _, _ := strings.Cut(tag, ",") // the KEY is the mapping; presence is the next subtest's audit
			if got != m.wantTag {
				t.Errorf("%s.%s json key = %q (tag %q), want %q (A2-1.6)", m.typ, m.field, got, tag, m.wantTag)
			}
		}
	})

	t.Run("serialized_documents_carry_no_bare_node_id_key", func(t *testing.T) {
		nodeBytes, err := json.Marshal(f1Node([]string{eviA, eviB}, 88))
		if err != nil {
			t.Fatalf("marshal sample node: %v", err)
		}
		edgeBytes, err := json.Marshal(sampleEdge())
		if err != nil {
			t.Fatalf("marshal sample edge: %v", err)
		}
		docBytes, err := CanonicalContent(f1Node([]string{eviA, eviB}, 88))
		if err != nil {
			t.Fatalf("CanonicalContent: %v", err)
		}
		// The quoted search form: "agent_node_id" contains node_id as a
		// SUBSTRING but never as the quoted key "node_id".
		for _, doc := range []struct {
			name string
			b    []byte
		}{
			{"served node", nodeBytes},
			{"served edge", edgeBytes},
			{"canonical content document", docBytes},
		} {
			if bytes.Contains(doc.b, []byte(`"node_id"`)) {
				t.Errorf("%s carries a bare \"node_id\" key: %s", doc.name, doc.b)
			}
			if bytes.Contains(doc.b, []byte(`"id":`)) {
				t.Errorf("%s carries a bare \"id\" key: %s", doc.name, doc.b)
			}
		}
		// The sample documents DO carry ids, in their reserved spellings —
		// the scan above cannot pass by vacuity.
		if !bytes.Contains(nodeBytes, []byte(`"graph_node_id":"gn_`)) {
			t.Errorf("served node lacks its graph_node_id key: %s", nodeBytes)
		}
		if !bytes.Contains(nodeBytes, []byte(`"agent_node_id":"slp_node_`)) {
			t.Errorf("served node provenance lacks the agent_node_id spelling: %s", nodeBytes)
		}
		if !bytes.Contains(edgeBytes, []byte(`"graph_edge_id":"ge_`)) {
			t.Errorf("served edge lacks its graph_edge_id key: %s", edgeBytes)
		}
		// The canonical content document carries NO id keys at all — that
		// is A2-4.6's empty exclusion list expressed as a key set: the only
		// *_id-shaped keys are the two content fields evidence_id and
		// evidence_ids.
		for _, absent := range []string{`"graph_node_id"`, `"engagement_id"`, `"supersedes_id"`, `"superseded_by_id"`, `"content_hash"`, `"provenance"`, `"graph_seq"`, `"quarantined"`} {
			if bytes.Contains(docBytes, []byte(absent)) {
				t.Errorf("canonical content document carries excluded key %s: %s", absent, docBytes)
			}
		}
		if !bytes.Contains(docBytes, []byte(`"evidence_id":""`)) || !bytes.Contains(docBytes, []byte(`"evidence_ids":[`)) {
			t.Errorf("canonical content document lacks its evidence content keys: %s", docBytes)
		}
	})

	t.Run("sample_ids_are_valid_A0-1.2", func(t *testing.T) {
		// Guard the samples: a rotten fixture would make the byte scans above
		// meaningless.
		e := sampleEdge()
		for _, f := range []struct {
			kind ids.Kind
			id   string
		}{
			{ids.GraphEdge, e.ID}, {ids.Engagement, e.EngagementID},
			{ids.GraphNode, e.SourceID}, {ids.GraphNode, e.TargetID},
			{ids.GraphNode, f1Node(nil, 88).ID}, {ids.GraphNode, s1Node().ID},
			{ids.GraphNode, f1Node(nil, 88).SupersedesID},
			{ids.GraphNode, f1Node(nil, 88).SupersededByID},
			{ids.Engagement, f1Node(nil, 88).EngagementID},
		} {
			if !ids.Valid(f.kind, f.id) {
				t.Errorf("sample id %q is not a valid %s id — fix the fixture", f.id, f.kind)
			}
		}
	})

	t.Run("platform_keys_are_always_present", func(t *testing.T) {
		// A2-2.2/A2-8.1/A2-8.7: quarantined and report_excluded always
		// serialize (false is a value, A0-8.3/8.8), as do the platform-set
		// identity fields. Only the optional CONTENT fields carry omitempty.
		nodeAlways := []string{"graph_node_id", "engagement_id", "graph_seq", "kind",
			"label", "content_hash", "quarantined", "report_excluded", "provenance"}
		assertNoOmitEmpty(t, reflect.TypeOf(Node{}), nodeAlways)
		// Every Edge field is always present (the sketch has no omitempty).
		edgeAlways := []string{"graph_edge_id", "engagement_id", "graph_seq", "kind",
			"source_id", "source_kind", "target_id", "target_kind", "retracted", "provenance"}
		assertNoOmitEmpty(t, reflect.TypeOf(Edge{}), edgeAlways)
		if got := jsonFieldCount(reflect.TypeOf(Edge{})); got != len(edgeAlways) {
			t.Errorf("Edge serializes %d keys, want exactly %d (A2 §4 sketch)", got, len(edgeAlways))
		}

		b, err := json.Marshal(Node{})
		if err != nil {
			t.Fatalf("marshal zero Node: %v", err)
		}
		for _, want := range []string{`"quarantined":false`, `"report_excluded":false`,
			`"content_hash":""`, `"graph_seq":0`, `"kind":""`, `"label":""`} {
			if !bytes.Contains(b, []byte(want)) {
				t.Errorf("zero Node serialization lacks %s: %s", want, b)
			}
		}
		// And the optional content fields are absent on a zero Node, never
		// null (A0-8.3).
		if bytes.Contains(b, []byte("null")) {
			// provenance is the one declared exception on a ZERO value: the
			// no-omitempty tag means the key is present; a served record
			// always carries ≥ 1 entry (A2-5.1, enforced by
			// ValidateProvenance on the write path).
			withoutProv := bytes.Replace(b, []byte(`"provenance":null`), nil, 1)
			if bytes.Contains(withoutProv, []byte("null")) {
				t.Errorf("zero Node serializes null outside the provenance placeholder: %s", b)
			}
		}
	})

	t.Run("A2-1.8_current_is_superseded_by_absence", func(t *testing.T) {
		// "Current" is a defined term derived from the one platform-set
		// pointer, never a stored field.
		if !(Node{}).Current() {
			t.Error("a node without superseded_by_id must be current")
		}
		superseded := Node{SupersededByID: "gn_01m1y2whfjk5t8nq2z7x1vb3rt"}
		if superseded.Current() {
			t.Error("a node with superseded_by_id must not be current")
		}
		revising := Node{SupersedesID: "gn_01m1y2whfh9x2b4c7d1e8f0a3b"}
		if !revising.Current() {
			t.Error("supersedes_id (the backward pointer) does not affect current-ness (A2-4.2)")
		}
	})
}

// assertNoOmitEmpty fails the test if any of the named json keys on typ
// carries the omitempty option.
func assertNoOmitEmpty(t *testing.T, typ reflect.Type, jsonNames []string) {
	t.Helper()
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		tag := f.Tag.Get("json")
		name, opts, _ := strings.Cut(tag, ",")
		if slices.Contains(jsonNames, name) && strings.Contains(opts, "omitempty") {
			t.Errorf("%s.%s json tag %q carries omitempty — the key must always be present (A2-2.2, A0-8.3)", typ, f.Name, tag)
		}
	}
}

// jsonFieldCount counts the fields of typ that serialize (tag not "-").
func jsonFieldCount(typ reflect.Type) int {
	n := 0
	for i := 0; i < typ.NumField(); i++ {
		if tag := typ.Field(i).Tag.Get("json"); tag != "-" {
			n++
		}
	}
	return n
}
