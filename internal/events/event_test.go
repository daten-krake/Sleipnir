package events

import (
	"reflect"
	"slices"
	"strings"
	"testing"
)

// jsonKeys walks a canonicalized struct type and returns its JSON key names
// in field order, promoting the fields of embedded (tagless anonymous)
// structs to the top level exactly as encoding/json — and therefore the
// canonical form — does (A1-1.3). It fails the test on a missing tag or any
// tag option: canonicalized types carry bare names, no omitempty (A1-1.2,
// A0-2.14).
func jsonKeys(t *testing.T, rt reflect.Type) []string {
	t.Helper()
	var keys []string
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		tag, ok := f.Tag.Lookup("json")
		if f.Anonymous && !ok {
			keys = append(keys, jsonKeys(t, f.Type)...)
			continue
		}
		if !ok {
			t.Fatalf("%s.%s: canonicalized field without an explicit json tag (A1-4.1, A0-8.1)", rt, f.Name)
		}
		name, opts, _ := strings.Cut(tag, ",")
		if opts != "" {
			t.Fatalf("%s.%s: json tag %q carries options — no omitempty on a canonicalized type (A1-1.2, A0-2.14)", rt, f.Name, tag)
		}
		if name == "-" {
			t.Fatalf("%s.%s: json tag \"-\" hides a field of a fixed key set (A0-2.14)", rt, f.Name)
		}
		keys = append(keys, name)
	}
	return keys
}

// TestEnvelopeKeySetIsSeventeen pins the one envelope shape (A1-1.1): the
// reflection over Event yields exactly the 17 keys of the A1-1.1 table, in
// table order, all tags bare (no omitempty, A1-1.2/A0-2.14), and the actor
// object is exactly the three keys of A1-2.1. The oracle is the parsed
// contract table; the literal count 17 is A1-1.1's own published number.
func TestEnvelopeKeySetIsSeventeen(t *testing.T) {
	want := parseEnvelopeKeys(t)

	t.Run("ContractClosesTheSetAtSeventeen", func(t *testing.T) {
		if len(want) != 17 {
			t.Fatalf("A1-1.1 table parsed %d keys; the clause closes the set at 17", len(want))
		}
		if slices.Contains(want, "chain") {
			t.Fatal("a nested \"chain\" object key would be unexcludable (A1-1.3)")
		}
	})

	got := jsonKeys(t, reflect.TypeOf(Event{}))

	t.Run("EventSerializesTheSeventeenKeysInOrder", func(t *testing.T) {
		if !slices.Equal(got, want) {
			t.Errorf("Event json keys =\n\t%v\nA1-1.1 table says =\n\t%v", got, want)
		}
	})

	t.Run("NoDuplicateKeys", func(t *testing.T) {
		seen := map[string]bool{}
		for _, k := range got {
			if seen[k] {
				t.Errorf("envelope key %q appears twice", k)
			}
			seen[k] = true
		}
	})

	t.Run("ChainFieldsAreTopLevel", func(t *testing.T) {
		// A1-1.3: seq, prev_hash and hash are top-level keys via the embedded
		// ChainBlock, never fields of a nested object.
		for _, k := range []string{"seq", "prev_hash", "hash"} {
			if !slices.Contains(got, k) {
				t.Errorf("envelope lacks top-level %q (A1-1.3)", k)
			}
		}
	})

	t.Run("KeysAreA0-8-1Spelled", func(t *testing.T) {
		for _, k := range got {
			if !jsonKeyRe.MatchString(k) {
				t.Errorf("envelope key %q violates A0-8.1 (^[a-z][a-z0-9_]{0,39}$)", k)
			}
		}
	})

	t.Run("ActorKeySetIsThree", func(t *testing.T) {
		actorKeys := jsonKeys(t, reflect.TypeOf(Actor{}))
		if !slices.Equal(actorKeys, []string{"type", "principal_id", "component"}) {
			t.Errorf("Actor json keys = %v, A1-2.1 fixes exactly type, principal_id, component", actorKeys)
		}
	})
}

// checkNoTimeTypes fails the test when rt or anything reachable from it is a
// type of the standard time package. The walk is bounded by the seen set and
// by A1-4.8's flatness, so it cannot diverge (AGENTS.md: assert the bound's
// effect, never remove it).
func checkNoTimeTypes(t *testing.T, rt reflect.Type, path string, seen map[reflect.Type]bool) {
	t.Helper()
	if seen[rt] {
		return
	}
	seen[rt] = true
	if rt.PkgPath() == "time" {
		t.Errorf("%s: canonicalized type reaches %s — time.Time must not appear in any canonicalized type; timestamps are strings (A1-4.12, A0-5.1)", path, rt)
		return
	}
	switch rt.Kind() {
	case reflect.Struct:
		for i := 0; i < rt.NumField(); i++ {
			f := rt.Field(i)
			checkNoTimeTypes(t, f.Type, path+"."+f.Name, seen)
		}
	case reflect.Slice, reflect.Array, reflect.Pointer:
		checkNoTimeTypes(t, rt.Elem(), path+"[]", seen)
	case reflect.Map:
		checkNoTimeTypes(t, rt.Key(), path+"[key]", seen)
		checkNoTimeTypes(t, rt.Elem(), path+"[value]", seen)
	}
}

// canonicalizedTypes returns every type of this package whose values enter
// the canonical event document: the envelope, its nested actor and chain
// block, and all 42 payload structs (A0-2.14, A1-5.2).
func canonicalizedTypes() []reflect.Type {
	types := []reflect.Type{
		reflect.TypeOf(Event{}),
		reflect.TypeOf(Actor{}),
		reflect.TypeOf(ChainBlock{}),
	}
	for _, rt := range allPayloadTypes {
		types = append(types, rt)
	}
	return types
}

// TestNoTimeTimeInCanonicalizedTypes pins A1-4.12 (review finding P-12):
// time.Time must not appear in any canonicalized type — its MarshalJSON drops
// trailing zero fractional digits and silently changes digests — and every
// timestamp field of the contract is a Go string.
func TestNoTimeTimeInCanonicalizedTypes(t *testing.T) {
	t.Run("NoTimePackageTypeIsReachable", func(t *testing.T) {
		seen := map[reflect.Type]bool{}
		for _, rt := range canonicalizedTypes() {
			checkNoTimeTypes(t, rt, rt.Name(), seen)
		}
		// The walk must actually have visited the canonicalized types: an
		// empty traversal would make the assertion above vacuous.
		for _, rt := range canonicalizedTypes() {
			if !seen[rt] {
				t.Errorf("type walk never visited %s", rt)
			}
		}
	})

	t.Run("EnvelopeTimestampsAreStrings", func(t *testing.T) {
		rt := reflect.TypeOf(Event{})
		for _, name := range []string{"OccurredAt", "OccurredClaimedAt", "RecordedAt"} {
			f, ok := rt.FieldByName(name)
			if !ok {
				t.Fatalf("Event has no %s field (A1-1.1)", name)
			}
			if f.Type.Kind() != reflect.String {
				t.Errorf("Event.%s is %s — a timestamp field is a Go string holding an A0-5.1 value (A1-4.12)", name, f.Type)
			}
		}
	})

	t.Run("PayloadTimestampFieldsAreStrings", func(t *testing.T) {
		// Every field the contract types as `timestamp` (A1-3.2 notation) —
		// currently expires_at on the four approval kinds — must be a Go
		// string. No payload timestamp is ever a *_claimed_at field (A1-4.12).
		count := 0
		for _, row := range parseTaxonomy(t) {
			rt, ok := allPayloadTypes[Kind(row.name)]
			if !ok {
				t.Fatalf("no payload struct registered for kind %s", row.name)
			}
			for _, cf := range row.fields {
				if strings.HasSuffix(cf.name, "_claimed_at") {
					t.Errorf("kind %s: payload field %s is a *_claimed_at field — the only client-supplied time is the envelope's occurred_claimed_at (A1-4.12, A1-1.4)", row.name, cf.name)
				}
				if cf.typ != "timestamp" {
					continue
				}
				count++
				sf, ok := fieldByJSONName(rt, cf.name)
				if !ok {
					t.Errorf("kind %s: contract timestamp field %s missing from the struct", row.name, cf.name)
					continue
				}
				if sf.Type.Kind() != reflect.String {
					t.Errorf("kind %s field %s: %s — a timestamp field is a Go string holding an A0-5.1 value (A1-4.12)", row.name, cf.name, sf.Type)
				}
			}
		}
		if count == 0 {
			t.Error("parsed zero timestamp fields from A1-3.3 — the oracle found nothing to pin (expires_at exists on four approval kinds)")
		}
	})
}

// fieldByJSONName returns the struct field whose bare json tag equals name.
func fieldByJSONName(rt reflect.Type, name string) (reflect.StructField, bool) {
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		tag, ok := f.Tag.Lookup("json")
		if !ok {
			continue
		}
		if n, _, _ := strings.Cut(tag, ","); n == name {
			return f, true
		}
	}
	return reflect.StructField{}, false
}
