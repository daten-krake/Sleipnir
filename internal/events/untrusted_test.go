package events

import (
	"slices"
	"testing"
)

// TestUntrustedFieldsCoversEveryStarredField pins the A1-4.4 registry
// against the frozen contract text, from two independent oracles: the `*`
// markers inside the A1-3.3 payload cells and the dedicated A1-4.4 table
// ("normative: this table is what UntrustedFields returns"). Both must agree
// with each other and with UntrustedFields, every starred name must exist as
// a json field of that kind's payload struct, and an unknown kind must not
// silently return an empty list.
func TestUntrustedFieldsCoversEveryStarredField(t *testing.T) {
	rows := parseTaxonomy(t)
	starTable, otherCount := parseStarTable(t)

	// Stars derived from the A1-3.3 payload cells, in field order.
	starsFromTaxonomy := map[string][]string{}
	for _, row := range rows {
		var starred []string
		for _, f := range row.fields {
			if f.star {
				starred = append(starred, f.name)
			}
		}
		if len(starred) > 0 {
			starsFromTaxonomy[row.name] = starred
		}
	}

	t.Run("ContractTablesAgree", func(t *testing.T) {
		if len(starTable) != len(starsFromTaxonomy) {
			t.Fatalf("A1-4.4 table lists %d starred kinds, A1-3.3 markers give %d", len(starTable), len(starsFromTaxonomy))
		}
		for kind, fields := range starsFromTaxonomy {
			other, ok := starTable[kind]
			if !ok {
				t.Errorf("kind %s: starred in A1-3.3 (%v) but absent from the A1-4.4 table", kind, fields)
				continue
			}
			if !slices.Equal(fields, other) {
				t.Errorf("kind %s: A1-3.3 stars %v, A1-4.4 table says %v", kind, fields, other)
			}
		}
		// The published literals: 19 starred kinds + "every other kind
		// (23 of 42)" = 42 (A1-4.4). Published literals win.
		if otherCount != 23 {
			t.Errorf("A1-4.4 'every other kind' count = %d, the published literal is 23", otherCount)
		}
		if len(rows)-len(starsFromTaxonomy) != otherCount {
			t.Errorf("%d parsed kinds − %d starred kinds = %d, A1-4.4 publishes %d", len(rows), len(starsFromTaxonomy), len(rows)-len(starsFromTaxonomy), otherCount)
		}
	})

	t.Run("RegistryMatchesContractStars", func(t *testing.T) {
		for _, row := range rows {
			got, ok := UntrustedFields(Kind(row.name))
			if !ok {
				t.Errorf("kind %s: UntrustedFields reports an unknown kind", row.name)
				continue
			}
			want := starsFromTaxonomy[row.name]
			if want == nil {
				want = []string{}
			}
			if !slices.Equal(got, want) {
				t.Errorf("kind %s: UntrustedFields = %v, A1-3.3 stars say %v", row.name, got, want)
			}
		}
	})

	t.Run("StarredNamesArePayloadFields", func(t *testing.T) {
		for _, row := range rows {
			rt := allPayloadTypes[Kind(row.name)]
			if rt == nil {
				t.Fatalf("no payload struct registered for kind %s", row.name)
			}
			starred, ok := UntrustedFields(Kind(row.name))
			if !ok {
				t.Fatalf("kind %s: UntrustedFields reports an unknown kind", row.name)
			}
			for _, name := range starred {
				if _, found := fieldByJSONName(rt, name); !found {
					t.Errorf("kind %s: starred field %s has no json field in %s", row.name, name, rt)
				}
			}
		}
	})

	t.Run("UnknownKindIsNotOK", func(t *testing.T) {
		for _, bogus := range []Kind{"", "command_executed ", "command_executed_", "node_superseded", "chain_genesis\n", "Chain_Genesis"} {
			fields, ok := UntrustedFields(bogus)
			if ok {
				t.Errorf("UntrustedFields(%q) = (%v, true) — an unknown kind must not be answered (A1-3.1)", bogus, fields)
			}
			if fields != nil {
				t.Errorf("UntrustedFields(%q) returned non-nil fields with ok=false", bogus)
			}
		}
	})

	t.Run("ReturnedSliceIsACopy", func(t *testing.T) {
		fields, ok := UntrustedFields(KindCommandExecuted)
		if !ok || len(fields) != 2 {
			t.Fatalf("UntrustedFields(command_executed) = (%v, %v), want the two A1-3.3 stars", fields, ok)
		}
		fields[0] = "tampered"
		again, _ := UntrustedFields(KindCommandExecuted)
		if again[0] != "command" {
			t.Errorf("mutating the returned slice corrupted the registry: second call = %v", again)
		}
	})
}
