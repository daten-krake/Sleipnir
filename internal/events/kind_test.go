package events

import (
	"slices"
	"strings"
	"testing"
)

// TestKindListIs42AndClosed pins the closed taxonomy (A1-3.1) against the
// frozen contract text: A1-3.3 lists exactly 42 unique kinds, AllKinds
// reproduces them byte-exactly in table order, every kind has a payload
// struct, the **C** markers equal ClientAppendable, and the spelling follows
// A0-8.5. The older "39" count is stale (A1-3.3 counted 2026-09-29, WP-09
// brief: settled, do not re-litigate).
func TestKindListIs42AndClosed(t *testing.T) {
	rows := parseTaxonomy(t)

	names := make([]string, 0, len(rows))
	for _, row := range rows {
		names = append(names, row.name)
	}

	t.Run("ContractLists42UniqueKinds", func(t *testing.T) {
		if len(rows) != 42 {
			t.Fatalf("A1-3.3 parsed %d kinds, the contract literal says 42 (A1-3.1, §6 item 14)", len(rows))
		}
		seen := map[string]bool{}
		for _, n := range names {
			if seen[n] {
				t.Errorf("kind %s appears twice in A1-3.3", n)
			}
			seen[n] = true
		}
	})

	t.Run("AllKindsMatchesContractByteExactInOrder", func(t *testing.T) {
		got := AllKinds()
		if len(got) != len(names) {
			t.Fatalf("AllKinds returns %d kinds, contract lists %d", len(got), len(names))
		}
		for i, k := range got {
			if string(k) != names[i] {
				t.Errorf("AllKinds[%d] = %q, A1-3.3 table order says %q", i, k, names[i])
			}
		}
	})

	t.Run("AllKindsReturnsACopy", func(t *testing.T) {
		a := AllKinds()
		a[0] = Kind("tampered_kind")
		if AllKinds()[0] != KindChainGenesis {
			t.Error("mutating the AllKinds result changed the closed taxonomy (A1-3.1)")
		}
	})

	t.Run("EveryKindHasExactlyOnePayloadStruct", func(t *testing.T) {
		if len(allPayloadTypes) != len(rows) {
			t.Errorf("%d payload structs registered for %d contract kinds", len(allPayloadTypes), len(rows))
		}
		for _, n := range names {
			if _, ok := allPayloadTypes[Kind(n)]; !ok {
				t.Errorf("kind %s has no payload struct (A1-4.1)", n)
			}
		}
	})

	t.Run("ClientAppendableMatchesCMarkers", func(t *testing.T) {
		cCount := 0
		for _, row := range rows {
			if got := ClientAppendable(Kind(row.name)); got != row.appendable {
				t.Errorf("ClientAppendable(%q) = %v, A1-3.3 C marker says %v (A1-3.4)", row.name, got, row.appendable)
			}
			if row.appendable {
				cCount++
			}
		}
		if cCount != 3 {
			t.Errorf("A1-3.3 marks %d client-appendable kinds, A1-3.4 names exactly 3", cCount)
		}
	})

	t.Run("SpellingFollowsA0-8-5", func(t *testing.T) {
		for _, n := range names {
			if !kindNameRe.MatchString(n) {
				t.Errorf("kind %q violates ^[a-z][a-z0-9_]{0,31}$ (A0-8.5, A1-3.1)", n)
			}
		}
	})

	t.Run("NearMissKindsAreNotAppendable", func(t *testing.T) {
		// The bypass attempt: a worker claiming a C kind under a near-miss
		// spelling must not be treated as client-appendable (A1-3.1 byte-exact
		// comparison, no synonyms).
		for _, bogus := range []Kind{
			"Command_Executed", "command_executed ", " command_executed",
			"command-executed", "command_executed\n", "task_results",
			"revert_recorded_", "", "chain_genesis",
		} {
			if ClientAppendable(bogus) {
				t.Errorf("ClientAppendable(%q) = true for a non-C spelling", bogus)
			}
		}
	})

	t.Run("KnownKindMembership", func(t *testing.T) {
		for _, n := range names {
			if !slices.Contains(AllKinds(), Kind(n)) {
				t.Errorf("contract kind %s is not a member of the closed list", n)
			}
		}
		if slices.Contains(AllKinds(), Kind("node_superseded")) {
			t.Error("node_superseded must not exist — A1-3.8 corrected that request into graph_node_written + graph_edge_written")
		}
	})

	t.Run("TypedEnumConstantsMatchContract", func(t *testing.T) {
		// The two payload enums the §4.1 sketch types as named Go types pin
		// their closed value lists against the A1-3.3 cells.
		want := enumValuesFor(t, rows, "chain_break_detected", "break_kind")
		got := []string{
			string(BreakPreimageMismatch), string(BreakLinkMismatch), string(BreakSeqGap),
			string(BreakSeqDisorder), string(BreakDuplicateEventID), string(BreakGenesisInvalid),
			string(BreakRowCountMismatch), string(BreakEngagementMismatch), string(BreakHeadRegression),
		}
		if !slices.Equal(got, want) {
			t.Errorf("BreakKind constants = %v, A1-3.3 break_kind enum = %v", got, want)
		}

		want = enumValuesFor(t, rows, "chain_verified", "trigger")
		got = []string{string(TriggerStartup), string(TriggerPreExport), string(TriggerOnDemand)}
		if !slices.Equal(got, want) {
			t.Errorf("VerifyTrigger constants = %v, A1-3.3 trigger enum = %v", got, want)
		}
	})
}

// enumValuesFor returns the parsed values of one enum{…} field of one kind.
func enumValuesFor(t *testing.T, rows []contractKind, kind, field string) []string {
	t.Helper()
	for _, row := range rows {
		if row.name != kind {
			continue
		}
		for _, f := range row.fields {
			if f.name != field {
				continue
			}
			if !strings.HasPrefix(f.typ, "enum{") || !strings.HasSuffix(f.typ, "}") {
				t.Fatalf("%s.%s is not an enum in A1-3.3: %q", kind, field, f.typ)
			}
			return strings.Split(f.typ[len("enum{"):len(f.typ)-1], ",")
		}
		t.Fatalf("kind %s has no field %s in A1-3.3", kind, field)
	}
	t.Fatalf("kind %s not found in the parsed taxonomy", kind)
	return nil
}
