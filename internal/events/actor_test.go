package events

import (
	"slices"
	"strings"
	"testing"
)

// TestActorComponentIsEmptyForNonPlatform pins A1-2.1/A1-3.3 (review finding
// P-01): for every kind row of the contract, every non-platform actor pair
// carries component "" — the subsystem named in the Source column appears in
// actor.component only when actor.type is "platform". The oracle is the
// parsed contract text; the package's ActorType/Component constants are the
// implementation side of the comparison.
func TestActorComponentIsEmptyForNonPlatform(t *testing.T) {
	rows := parseTaxonomy(t)

	// The dynamic actor cells of A1-3.3 name the principal's type indirectly;
	// they are non-platform by construction (scope_denied, blacklist_denied,
	// action_blocked, agent_error record a machine or user principal).
	dynamic := map[string]bool{
		"requesting principal's type": true,
		"failing principal's type":    true,
	}

	t.Run("NonPlatformPairsStampEmptyComponent", func(t *testing.T) {
		for _, row := range rows {
			for _, a := range row.actors {
				if a.atype == "platform" {
					continue // handled by the platform subtest
				}
				if a.component != "" {
					t.Errorf("kind %s: actor (%s, %q) — component must be \"\" for every non-platform actor (A1-2.1, A1-3.3)", row.name, a.atype, a.component)
				}
			}
		}
	})

	t.Run("PlatformPairsNameClosedComponent", func(t *testing.T) {
		compSet := map[string]bool{}
		for _, c := range parseComponentList(t) {
			compSet[c] = true
		}
		platformRows := 0
		for _, row := range rows {
			for _, a := range row.actors {
				if a.atype != "platform" {
					continue
				}
				platformRows++
				if a.component == "" {
					t.Errorf("kind %s: actor (platform, \"\") — a platform actor must name its subsystem (A1-2.1, A1-2.4)", row.name)
				}
				if !compSet[a.component] {
					t.Errorf("kind %s: component %q is not in the closed A1-2.4 enum", row.name, a.component)
				}
			}
		}
		if platformRows == 0 {
			t.Error("no platform-composed kind parsed — the oracle found nothing to pin")
		}
	})

	t.Run("ActorTypesAreTheClosedVocabulary", func(t *testing.T) {
		want := parseActorTypes(t)
		got := []string{
			string(ActorUser), string(ActorOrchestrator), string(ActorWorker),
			string(ActorPlatform), string(ActorNode),
		}
		if !slices.Equal(sortedCopy(want), sortedCopy(got)) {
			t.Errorf("package actor vocabulary %v, A1-2.1 closed enum %v", got, want)
		}
		for _, row := range rows {
			for _, a := range row.actors {
				if !slices.Contains(want, a.atype) && !dynamic[a.atype] {
					t.Errorf("kind %s: actor type %q is neither in the A1-2.1 enum nor a dynamic principal cell", row.name, a.atype)
				}
			}
		}
	})

	t.Run("ComponentConstantsMatchA1-2-4", func(t *testing.T) {
		want := parseComponentList(t)
		got := []string{
			string(CompEventStore), string(CompScope), string(CompApproval),
			string(CompSpawnBroker), string(CompLLMGateway), string(CompGraph),
			string(CompEvidence), string(CompCleanup), string(CompIntegrity),
			string(CompNotify), string(CompRuntime), string(CompAPI),
		}
		if len(got) != len(want) {
			t.Fatalf("package declares %d components, A1-2.4 lists %d", len(got), len(want))
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("component[%d] = %q, A1-2.4 says %q", i, got[i], want[i])
			}
		}
	})

	t.Run("A1-2-2TableAgrees", func(t *testing.T) {
		table := parseActorIDTable(t)
		for atype, componentCell := range table {
			if atype == "platform" {
				if !strings.Contains(componentCell, "A1-2.4") {
					t.Errorf("A1-2.2 platform row: component cell %q must point at the A1-2.4 closed list", componentCell)
				}
				continue
			}
			if componentCell != "`\"\"`" {
				t.Errorf("A1-2.2 row %s: component cell %q, want the literal \"\" (A1-2.1)", atype, componentCell)
			}
		}
	})
}

func sortedCopy(in []string) []string {
	out := slices.Clone(in)
	slices.Sort(out)
	return out
}
