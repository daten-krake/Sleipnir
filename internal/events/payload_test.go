package events

import (
	"reflect"
	"slices"
	"strings"
	"testing"
)

// allPayloadTypes enumerates the package's 42 payload structs. It is the
// tests' enumeration mechanism only: the oracle for field names, order,
// types and stars is the parsed A1-3.3 contract text, never this map (WP-09
// pinning rule). A struct registered here for a kind the contract does not
// list — or a contract kind missing here — fails
// TestPayloadStructsMatchA1Tables.
var allPayloadTypes = map[Kind]reflect.Type{
	KindChainGenesis:            reflect.TypeOf(ChainGenesisPayload{}),
	KindChainVerified:           reflect.TypeOf(ChainVerifiedPayload{}),
	KindChainBreakDetected:      reflect.TypeOf(ChainBreakDetectedPayload{}),
	KindIntegrityOverride:       reflect.TypeOf(IntegrityOverridePayload{}),
	KindArtifactReleased:        reflect.TypeOf(ArtifactReleasedPayload{}),
	KindEngagementCreated:       reflect.TypeOf(EngagementCreatedPayload{}),
	KindScopeChanged:            reflect.TypeOf(ScopeChangedPayload{}),
	KindEngagementPolicyChanged: reflect.TypeOf(EngagementPolicyChangedPayload{}),
	KindRunStarted:              reflect.TypeOf(RunStartedPayload{}),
	KindRunEnded:                reflect.TypeOf(RunEndedPayload{}),
	KindModelConfigSnapshotted:  reflect.TypeOf(ModelConfigSnapshottedPayload{}),
	KindJobSpawned:              reflect.TypeOf(JobSpawnedPayload{}),
	KindTaskSpawned:             reflect.TypeOf(TaskSpawnedPayload{}),
	KindContainerStarted:        reflect.TypeOf(ContainerStartedPayload{}),
	KindContainerKilled:         reflect.TypeOf(ContainerKilledPayload{}),
	KindHardStopFired:           reflect.TypeOf(HardStopFiredPayload{}),
	KindEngagementClosed:        reflect.TypeOf(EngagementClosedPayload{}),
	KindSpawnRequested:          reflect.TypeOf(SpawnRequestedPayload{}),
	KindCommandExecuted:         reflect.TypeOf(CommandExecutedPayload{}),
	KindEvidenceStored:          reflect.TypeOf(EvidenceStoredPayload{}),
	KindTaskResult:              reflect.TypeOf(TaskResultPayload{}),
	KindRevertRecorded:          reflect.TypeOf(RevertRecordedPayload{}),
	KindApprovalRequested:       reflect.TypeOf(ApprovalRequestedPayload{}),
	KindApprovalGranted:         reflect.TypeOf(ApprovalGrantedPayload{}),
	KindApprovalDenied:          reflect.TypeOf(ApprovalDeniedPayload{}),
	KindApprovalExpired:         reflect.TypeOf(ApprovalExpiredPayload{}),
	KindApprovalExecuted:        reflect.TypeOf(ApprovalExecutedPayload{}),
	KindLLMCall:                 reflect.TypeOf(LLMCallPayload{}),
	KindGraphNodeWritten:        reflect.TypeOf(GraphNodeWrittenPayload{}),
	KindGraphEdgeWritten:        reflect.TypeOf(GraphEdgeWrittenPayload{}),
	KindGraphNodeQuarantined:    reflect.TypeOf(GraphNodeQuarantinedPayload{}),
	KindQuarantineRecomputed:    reflect.TypeOf(QuarantineRecomputedPayload{}),
	KindGraphEdgeRetracted:      reflect.TypeOf(GraphEdgeRetractedPayload{}),
	KindReportInclusionChanged:  reflect.TypeOf(ReportInclusionChangedPayload{}),
	KindCleanupPlanned:          reflect.TypeOf(CleanupPlannedPayload{}),
	KindCleanupExecuted:         reflect.TypeOf(CleanupExecutedPayload{}),
	KindCleanupVerified:         reflect.TypeOf(CleanupVerifiedPayload{}),
	KindScopeDenied:             reflect.TypeOf(ScopeDeniedPayload{}),
	KindBlacklistDenied:         reflect.TypeOf(BlacklistDeniedPayload{}),
	KindActionBlocked:           reflect.TypeOf(ActionBlockedPayload{}),
	KindAgentError:              reflect.TypeOf(AgentErrorPayload{}),
	KindNotificationSent:        reflect.TypeOf(NotificationSentPayload{}),
}

// goKindForSpec transcribes an A1-3.2 type spec to the Go reflect.Kind the
// A1-4.12 rules mandate: string for string/string(N)/64hex/timestamp/enum,
// int64 for every :int except exit_code (the only field A1-4.2 ranges), bool
// for :bool, slice for array[string].
func goKindForSpec(f contractField) reflect.Kind {
	switch {
	case f.typ == "int":
		if f.name == "exit_code" {
			return reflect.Int
		}
		return reflect.Int64
	case f.typ == "bool":
		return reflect.Bool
	case f.typ == "array[string]":
		return reflect.Slice
	default: // string, string(N), 64hex, timestamp, enum{…}
		return reflect.String
	}
}

// TestPayloadStructsMatchA1Tables reflects every payload struct against its
// A1-3.3 table row parsed from the frozen contract: exactly the listed
// fields, in the listed order, with bare json tags (A1-4.1, A0-8.1), the
// A1-4.12 Go transcription, and a Kind() method reporting the right kind.
func TestPayloadStructsMatchA1Tables(t *testing.T) {
	rows := parseTaxonomy(t)
	if len(rows) != 42 {
		t.Fatalf("A1-3.3 parsed %d kinds, contract literal says 42 (A1-3.1)", len(rows))
	}

	t.Run("RegistryCoversExactlyTheTaxonomy", func(t *testing.T) {
		if len(allPayloadTypes) != len(rows) {
			t.Errorf("payload registry holds %d types, contract lists %d kinds", len(allPayloadTypes), len(rows))
		}
		for _, row := range rows {
			if _, ok := allPayloadTypes[Kind(row.name)]; !ok {
				t.Errorf("kind %s: contract lists it but no payload struct is registered (A1-4.1: exactly one payload type per kind)", row.name)
			}
		}
		for k := range allPayloadTypes {
			if !slices.ContainsFunc(rows, func(r contractKind) bool { return r.name == string(k) }) {
				t.Errorf("registered payload kind %s is not in the closed A1-3.3 taxonomy", k)
			}
		}
	})

	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			rt, ok := allPayloadTypes[Kind(row.name)]
			if !ok {
				t.Fatalf("no payload struct registered for kind %s", row.name)
			}
			if rt.Kind() != reflect.Struct {
				t.Fatalf("payload type %s is %s, A1-4.1 requires a flat struct", rt, rt.Kind())
			}
			if rt.NumField() != len(row.fields) {
				t.Fatalf("%s has %d fields, A1-3.3 lists %d — the key set is closed (A1-4.1)", rt, rt.NumField(), len(row.fields))
			}
			for i, cf := range row.fields {
				sf := rt.Field(i)
				if sf.Anonymous {
					t.Errorf("field %d: embedded struct %s — no shared payload base type (A1-4.1)", i, sf.Type)
					continue
				}
				if !sf.IsExported() {
					t.Errorf("field %d (%s) is unexported — every listed key must serialize (A0-2.14)", i, sf.Name)
					continue
				}
				tag, hasTag := sf.Tag.Lookup("json")
				if !hasTag {
					t.Errorf("field %s: no explicit json tag (A1-4.1, A0-8.1)", sf.Name)
					continue
				}
				name, opts, _ := strings.Cut(tag, ",")
				if opts != "" {
					t.Errorf("field %s: json tag %q carries options — canonicalized types use bare names, no omitempty (A0-2.14, A1-1.2)", sf.Name, tag)
				}
				if name != cf.name {
					t.Errorf("field %d: json name %q, A1-3.3 lists %q", i, name, cf.name)
				}
				if !jsonKeyRe.MatchString(name) {
					t.Errorf("field %s: json name %q violates A0-8.1 (^[a-z][a-z0-9_]{0,39}$)", sf.Name, name)
				}
				if want := goKindForSpec(cf); sf.Type.Kind() != want {
					t.Errorf("field %s (%s): Go kind %s, A1-4.12 transcribes %q as %s", sf.Name, name, sf.Type.Kind(), cf.typ, want)
				}
				if sf.Type.Kind() == reflect.Slice && sf.Type.Elem().Kind() != reflect.String {
					t.Errorf("field %s: array element kind %s, A1-3.2 array[string] requires string", sf.Name, sf.Type.Elem().Kind())
				}
			}
			p, ok := reflect.New(rt).Interface().(interface {
				Kind() Kind
			})
			if !ok {
				t.Fatalf("%s: no Kind() Kind method (the §4.1 Payload interface)", rt)
			}
			if got := p.Kind(); got != Kind(row.name) {
				t.Errorf("Kind() = %q, want %q", got, row.name)
			}
		})
	}
}

// TestPayloadsAreFlat proves A1-4.8 by reflection: every payload field is a
// string, integer, bool, or array of strings — no nested objects, no arrays
// of objects, no maps, no pointers, no floats (A0-2.6), and the envelope
// itself nests only actor and payload, so document depth stays ≤ 3.
func TestPayloadsAreFlat(t *testing.T) {
	for _, row := range parseTaxonomy(t) {
		rt, ok := allPayloadTypes[Kind(row.name)]
		if !ok {
			t.Fatalf("no payload struct registered for kind %s", row.name)
		}
		t.Run(row.name, func(t *testing.T) {
			for i := 0; i < rt.NumField(); i++ {
				f := rt.Field(i)
				if f.Anonymous {
					t.Errorf("field %d: embedded %s — a payload is one flat struct, no shared base (A1-4.1, A1-4.8)", i, f.Type)
					continue
				}
				switch f.Type.Kind() {
				case reflect.String, reflect.Bool, reflect.Int, reflect.Int64:
					// A1-4.8: strings, integers, booleans.
				case reflect.Slice:
					switch f.Type.Elem().Kind() {
					case reflect.String, reflect.Int, reflect.Int64:
						// A1-4.8: arrays of strings/integers.
					default:
						t.Errorf("field %s: array of %s — arrays of objects are forbidden (A1-4.8)", f.Name, f.Type.Elem())
					}
				case reflect.Float32, reflect.Float64:
					t.Errorf("field %s: float %s — a float in a payload is validation (A1-4.8, A0-2.6)", f.Name, f.Type)
				default:
					t.Errorf("field %s: %s (%s) — nested objects, maps and pointers are forbidden (A1-4.8)", f.Name, f.Type, f.Type.Kind())
				}
			}
		})
	}

	t.Run("EnvelopeNestsOnlyActorAndPayload", func(t *testing.T) {
		rt := reflect.TypeOf(Event{})
		actorType := reflect.TypeOf(Actor{})
		payloadType := reflect.TypeOf((*Payload)(nil)).Elem()
		for i := 0; i < rt.NumField(); i++ {
			f := rt.Field(i)
			switch f.Type {
			case actorType:
				for j := 0; j < f.Type.NumField(); j++ {
					af := f.Type.Field(j)
					if af.Type.Kind() != reflect.String {
						t.Errorf("actor.%s: %s — actor nests only strings (A1-2.1, depth ≤ 3)", af.Name, af.Type)
					}
				}
			case payloadType:
				// The payload interface: its concretes are the flat structs above.
			default:
				switch f.Type.Kind() {
				case reflect.String, reflect.Bool, reflect.Int64:
					// envelope scalars
				case reflect.Slice:
					if f.Type.Elem().Kind() != reflect.String {
						t.Errorf("envelope field %s: array of %s (A1-4.8)", f.Name, f.Type.Elem())
					}
				case reflect.Struct:
					// ChainBlock is embedded (A1-1.3); its fields are checked below.
					if !f.Anonymous {
						t.Errorf("envelope field %s: nested object %s — only actor and payload nest (A1-1.5)", f.Name, f.Type)
						continue
					}
					for j := 0; j < f.Type.NumField(); j++ {
						cf := f.Type.Field(j)
						if cf.Type.Kind() != reflect.String && cf.Type.Kind() != reflect.Int64 {
							t.Errorf("chain field %s: %s — chain fields are string/int64 (A1-1.3)", cf.Name, cf.Type)
						}
					}
				default:
					t.Errorf("envelope field %s: %s (%s) is not an envelope scalar, actor, payload or the embedded chain block (A1-1.1)", f.Name, f.Type, f.Type.Kind())
				}
			}
		}
	})
}
