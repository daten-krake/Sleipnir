package graph

import (
	"bytes"
	"encoding/json"
	"reflect"
	"slices"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/daten-krake/sleipnir/internal/cjson"
	"github.com/daten-krake/sleipnir/internal/errs"
)

// The A2 §4.2 normative content fingerprint vector, transcribed byte-exactly
// from the frozen contract's published canonical byte strings (lines 1579,
// 1582, 1585) and its len/SHA-256 table. Independently verified before being
// encoded here (scratch python3+hashlib outside the repo, 2026-10-07):
// F1 656 B → ad8f188e…076e, F3 656 B → 1748b813…b983, S1 304 B →
// 3955d821…2b67, and the F1-R defective marker is the SHA-256 of F1's bytes
// with the evidence_ids array left in the reversed input order. Published
// literals win (WP-14 brief): these constants are the contract's, not
// recomputed values.
const (
	f1CanonicalBytes = `{"addresses":[],"attrs":{"cvss_v3_x10":88,"first_seen_task":"task_01m1y2whfh1txm57x8dn41r9hg","relay_tool":"ntlmrelayx"},"basis":"","cidr":"","claim":"","credential_kind":"","domain":"","evidence_id":"","evidence_ids":["evi_01m1y2whfh3ca875z2x8v8h7qt","evi_01m1y2whfh7kq2m4c8x1z9vb3n"],"kind":"finding","label":"SMB relay to SYSVOL on dc01","media_kind":"","port":0,"protocol":"","severity":"high","sid":"","size_bytes":0,"status":"confirmed","summary":"Captured NTLM authentication from 10.20.0.14 was relayed to the SYSVOL share on dc01, yielding read access to group policy preferences. Secret material is referenced, not stored (evi_).","transport":""}`
	f3CanonicalBytes = `{"addresses":[],"attrs":{"cvss_v3_x10":87,"first_seen_task":"task_01m1y2whfh1txm57x8dn41r9hg","relay_tool":"ntlmrelayx"},"basis":"","cidr":"","claim":"","credential_kind":"","domain":"","evidence_id":"","evidence_ids":["evi_01m1y2whfh3ca875z2x8v8h7qt","evi_01m1y2whfh7kq2m4c8x1z9vb3n"],"kind":"finding","label":"SMB relay to SYSVOL on dc01","media_kind":"","port":0,"protocol":"","severity":"high","sid":"","size_bytes":0,"status":"confirmed","summary":"Captured NTLM authentication from 10.20.0.14 was relayed to the SYSVOL share on dc01, yielding read access to group policy preferences. Secret material is referenced, not stored (evi_).","transport":""}`
	s1CanonicalBytes = `{"addresses":["10.20.0.14"],"attrs":{},"basis":"","cidr":"","claim":"","credential_kind":"","domain":"","evidence_id":"","evidence_ids":[],"kind":"service","label":"microsoft-ds","media_kind":"","port":445,"protocol":"smb","severity":"","sid":"","size_bytes":0,"status":"","summary":"","transport":"tcp"}`

	f1Digest = "ad8f188e63b2563f9adad88df585d94df9c662e4082895e974b0b31e8a65076e"
	f3Digest = "1748b813a0d28d9f17f4d89dff2a08536741bf45e8fe0defbb3c4363d3f9b983"
	s1Digest = "3955d82160284d3e76c9be1b06981530df50e1635a1bad7ceb7b50c5f73c2b67"

	// defectiveF1RDigest is A2 §4.2's published defective-implementation
	// marker: the digest an implementation produces for the F1-R input when
	// it canonicalizes WITHOUT A2-4.6's sort/dedup. An implementation that
	// produces it MUST fail TestContentHashStableAcrossArrayOrder and MUST
	// NOT ship.
	defectiveF1RDigest = "4a017e71658de7ebb2d3a5429f90818b8a69302ff1909badbc782d9049ce9cae"
)

// The two evidence ids of the §4.1 finding example, in ascending byte order
// ('3' < '7' after the common prefix): F1 shows the sorted order; F1-R is
// "the same node with evidence_ids GIVEN in the reverse order".
const (
	eviA = "evi_01m1y2whfh3ca875z2x8v8h7qt"
	eviB = "evi_01m1y2whfh7kq2m4c8x1z9vb3n"
)

// f1Node is the §4.1 finding example whose content_hash is §4.2 vector F1.
// Every platform-set field that A2-4.6 EXCLUDES from the document is
// deliberately populated — including content_hash itself, both supersession
// pointers, the quarantine flags and a two-entry provenance list — so the
// byte-exact match below proves the empty exclusion list: none of them can
// influence the digest. evidenceIDs is the caller's input order.
func f1Node(evidenceIDs []string, cvss int64) Node {
	return Node{
		ID:           "gn_01m1y2whfhh039ykj5x8mc5a0g",
		EngagementID: "eng_01m1y2whfhgbz06ays6dxnvyws",
		GraphSeq:     412,
		Kind:         KindFinding,

		Label: "SMB relay to SYSVOL on dc01",
		Summary: "Captured NTLM authentication from 10.20.0.14 was relayed to the SYSVOL " +
			"share on dc01, yielding read access to group policy preferences. " +
			"Secret material is referenced, not stored (evi_).",
		EvidenceIDs: evidenceIDs,
		Attrs: Attrs{
			"cvss_v3_x10":     {Type: AttrInt, Num: cvss},
			"first_seen_task": {Type: AttrString, Str: "task_01m1y2whfh1txm57x8dn41r9hg"},
			"relay_tool":      {Type: AttrString, Str: "ntlmrelayx"},
		},
		Severity: SeverityHigh,
		Status:   string(FindingConfirmed),

		// Excluded platform-set fields, all populated on purpose (A2-4.6).
		ContentHash:      f1Digest,
		Quarantined:      true,
		QuarantineReason: QuarantineBlacklist,
		ReportExcluded:   true,
		SupersedesID:     "gn_01m1y2whfh9x2b4c7d1e8f0a3b",
		SupersededByID:   "gn_01m1y2whfjk5t8nq2z7x1vb3rt",
		Provenance: []Provenance{
			{
				PrincipalKind:     PrincipalWorker,
				RunID:             "run_01m1y2whfhnjx2am9103w0pnqw",
				JobID:             "job_01m1y2whfhbt69j0h0fbxepw90",
				TaskID:            "task_01m1y2whfh1txm57x8dn41r9hg",
				AgentNodeID:       "slp_node_01m1y2whfhxydsaem68cmazyc8",
				ToolID:            "tool_01m1y2whfhfjdvwqp9pfxqekmf",
				ToolVersion:       "1.4.2",
				EventID:           "evt_01m1y2whfhp17g0avdqztd2p3x",
				RecordedAt:        "2026-09-07T14:03:22.481Z",
				ObservedClaimedAt: "2026-09-07T14:03:19.900Z",
				Confidence:        ConfidenceObserved,
			},
			{
				PrincipalKind:     PrincipalWorker,
				RunID:             "run_01m1y2whfhnjx2am9103w0pnqw",
				JobID:             "job_01m1y2whfhvk83rt5x1z7c4m2b",
				TaskID:            "task_01m1y2whfh4kd8nq2z7x1vb3rt",
				AgentNodeID:       "slp_node_01m1y2whfhq3vb8nrt5x1z7c4m",
				ToolID:            "tool_01m1y2whfhfjdvwqp9pfxqekmf",
				ToolVersion:       "1.4.2",
				EventID:           "evt_01m1y2whfhr8t2nb5x9qz4vb7m",
				RecordedAt:        "2026-09-07T14:07:55.140Z",
				ObservedClaimedAt: "2026-09-07T14:07:52.880Z",
				Confidence:        ConfidenceVerified,
			},
		},
	}
}

// s1Node is §4.2's vector S1: a service node whose inapplicable keys sit at
// their zero values, with NO attrs and NO evidence_ids — nil collections on
// the Node must serialize as {} and [] in the document (A0-2.14), never as
// null. Platform-set fields are populated for the same exclusion proof.
func s1Node() Node {
	return Node{
		ID:           "gn_01m1y2whfhdc01srv4x8mc5a0g",
		EngagementID: "eng_01m1y2whfhgbz06ays6dxnvyws",
		GraphSeq:     777,
		Kind:         KindService,
		Label:        "microsoft-ds",
		Addresses:    []string{"10.20.0.14"},
		Port:         445,
		Transport:    "tcp",
		Protocol:     "smb",

		Attrs:       nil, // → "attrs":{} (vector S1)
		EvidenceIDs: nil, // → "evidence_ids":[] (vector S1)
		ContentHash: s1Digest,
		Quarantined: false,
		Provenance:  []Provenance{{PrincipalKind: PrincipalPlatform, RunID: "run_01m1y2whfhnjx2am9103w0pnqw", EventID: "evt_01m1y2whfhp17g0avdqztd2p3x", RecordedAt: "2026-09-07T14:03:22.502Z", Confidence: ConfidenceObserved}},
	}
}

func TestContentHashVector(t *testing.T) {
	tests := []struct {
		name       string
		node       Node
		wantBytes  string
		wantLen    int
		wantDigest string
	}{
		{"F1", f1Node([]string{eviA, eviB}, 88), f1CanonicalBytes, 656, f1Digest},
		{"F1-R", f1Node([]string{eviB, eviA}, 88), f1CanonicalBytes, 656, f1Digest},
		{"F3", f1Node([]string{eviA, eviB}, 87), f3CanonicalBytes, 656, f3Digest},
		{"S1", s1Node(), s1CanonicalBytes, 304, s1Digest},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotBytes, err := CanonicalContent(tc.node)
			if err != nil {
				t.Fatalf("CanonicalContent: %v", err)
			}
			if string(gotBytes) != tc.wantBytes {
				t.Errorf("canonical bytes mismatch (A2 §4.2 is normative)\n got: %s\nwant: %s", gotBytes, tc.wantBytes)
			}
			if len(gotBytes) != tc.wantLen {
				t.Errorf("canonical length = %d, want the published %d", len(gotBytes), tc.wantLen)
			}
			gotDigest, err := ContentHash(tc.node)
			if err != nil {
				t.Fatalf("ContentHash: %v", err)
			}
			if gotDigest != tc.wantDigest {
				t.Errorf("content_hash = %s, want the published %s", gotDigest, tc.wantDigest)
			}
			// A0-2.15: the digest IS the SHA-256 of exactly these bytes.
			if sum := cjson.SHA256Hex(gotBytes); sum != tc.wantDigest {
				t.Errorf("SHA256Hex(canonical bytes) = %s, want %s", sum, tc.wantDigest)
			}
		})
	}

	t.Run("F1-R_is_byte_identical_to_F1", func(t *testing.T) {
		// The contract row: F1-R produces bytes and digest IDENTICAL to F1 —
		// that is what A2-4.6's "sorted and deduplicated before
		// canonicalization" buys (A1-4.7's rule).
		f1Bytes, err := CanonicalContent(f1Node([]string{eviA, eviB}, 88))
		if err != nil {
			t.Fatalf("F1: %v", err)
		}
		f1rBytes, err := CanonicalContent(f1Node([]string{eviB, eviA}, 88))
		if err != nil {
			t.Fatalf("F1-R: %v", err)
		}
		if !bytes.Equal(f1Bytes, f1rBytes) {
			t.Errorf("F1-R canonical bytes differ from F1:\n F1: %s\nF1-R: %s", f1Bytes, f1rBytes)
		}
	})

	t.Run("F1_and_F3_differ_in_exactly_one_attrs_value", func(t *testing.T) {
		// §4.2's own claim about the pair (A2-6.5 + A2-4.7's dedup key):
		// one attrs value apart, same length, different digest.
		if len(f1CanonicalBytes) != len(f3CanonicalBytes) {
			t.Fatalf("published F1/F3 lengths differ: %d vs %d", len(f1CanonicalBytes), len(f3CanonicalBytes))
		}
		var diffs int
		for i := 0; i < len(f1CanonicalBytes); i++ {
			if f1CanonicalBytes[i] != f3CanonicalBytes[i] {
				diffs++
			}
		}
		if diffs != 1 {
			t.Errorf("F1 and F3 differ in %d byte positions, want exactly 1", diffs)
		}
		if f1Digest == f3Digest {
			t.Error("F1 and F3 must not share a digest")
		}
	})

	t.Run("A2-4.8_verify_recomputes_from_stored_bytes", func(t *testing.T) {
		if err := VerifyContentHash([]byte(f1CanonicalBytes), f1Digest); err != nil {
			t.Errorf("VerifyContentHash(stored F1 bytes, F1 digest) = %v, want nil", err)
		}
		tampered := []byte(f1CanonicalBytes)
		tampered[100] ^= 0xff // flip a byte of the stored canonical form
		if err := VerifyContentHash(tampered, f1Digest); err == nil {
			t.Error("tampered stored bytes must not verify")
		} else if got := errs.KindOf(err); got != errs.Internal {
			t.Errorf("tampered stored bytes: kind = %q, want internal — a content_hash mismatch is a platform defect, never integrity_failed (A2-4.9)", got)
		}
		if err := VerifyContentHash(nil, f1Digest); err == nil || errs.KindOf(err) != errs.Internal {
			t.Errorf("empty stored bytes: err = %v (kind %q), want internal (A0-2.16 requires the bytes to be persisted)", err, errs.KindOf(err))
		}
		if err := VerifyContentHash([]byte(f1CanonicalBytes), "not-a-digest"); err == nil || errs.KindOf(err) != errs.Internal {
			t.Errorf("malformed persisted digest: err = %v (kind %q), want internal", err, errs.KindOf(err))
		}
		if err := VerifyContentHash([]byte(f1CanonicalBytes), f3Digest); err == nil || errs.KindOf(err) != errs.Internal {
			t.Errorf("bytes and digest of different vectors: err = %v (kind %q), want internal", err, errs.KindOf(err))
		}
	})
}

func TestContentHashStableAcrossArrayOrder(t *testing.T) {
	t.Run("F1-R_input_never_yields_the_defective_marker", func(t *testing.T) {
		// §4.2: "an implementation that produces [the marker] for this input
		// fails TestContentHashStableAcrossArrayOrder and MUST NOT ship."
		got, err := ContentHash(f1Node([]string{eviB, eviA}, 88))
		if err != nil {
			t.Fatalf("ContentHash: %v", err)
		}
		if got == defectiveF1RDigest {
			t.Fatalf("implementation produces the published defective-implementation marker %s for the F1-R input — it canonicalized without A2-4.6's sort and MUST NOT ship", defectiveF1RDigest)
		}
		if got != f1Digest {
			t.Errorf("F1-R input digest = %s, want F1's %s — sorting is what makes F1-R == F1", got, f1Digest)
		}
	})

	t.Run("marker_is_exactly_the_unsorted_canonicalization", func(t *testing.T) {
		// Prove the published marker corresponds to skipping the sort (and
		// nothing else): F1's canonical bytes with the evidence_ids array in
		// the reversed input order digest to the marker.
		sortedArr := `["` + eviA + `","` + eviB + `"]`
		reversedArr := `["` + eviB + `","` + eviA + `"]`
		if n := strings.Count(f1CanonicalBytes, sortedArr); n != 1 {
			t.Fatalf("F1 bytes contain the sorted evidence_ids literal %d times, want 1", n)
		}
		unsorted := strings.Replace(f1CanonicalBytes, sortedArr, reversedArr, 1)
		if got := cjson.SHA256Hex([]byte(unsorted)); got != defectiveF1RDigest {
			t.Errorf("SHA-256 of the unsorted canonicalization = %s, want the published marker %s", got, defectiveF1RDigest)
		}
	})

	t.Run("duplicates_collapse_before_canonicalization", func(t *testing.T) {
		// A2-4.6: deduplicated BEFORE the document is canonicalized — the
		// input [B, A, B, A] produces F1's exact bytes.
		gotBytes, err := CanonicalContent(f1Node([]string{eviB, eviA, eviB, eviA}, 88))
		if err != nil {
			t.Fatalf("CanonicalContent: %v", err)
		}
		if string(gotBytes) != f1CanonicalBytes {
			t.Errorf("deduplicated canonical bytes mismatch\n got: %s\nwant: %s", gotBytes, f1CanonicalBytes)
		}
	})

	t.Run("addresses_sorted_and_deduped_too", func(t *testing.T) {
		messy := Node{Kind: KindHost, Label: "dc01", Addresses: []string{"10.0.0.2", "10.0.0.1", "10.0.0.2", "10.0.0.1"}}
		clean := Node{Kind: KindHost, Label: "dc01", Addresses: []string{"10.0.0.1", "10.0.0.2"}}
		messyHash, err := ContentHash(messy)
		if err != nil {
			t.Fatalf("messy: %v", err)
		}
		cleanHash, err := ContentHash(clean)
		if err != nil {
			t.Fatalf("clean: %v", err)
		}
		if messyHash != cleanHash {
			t.Errorf("addresses: messy input digest %s != sorted/deduped digest %s (A2-4.6)", messyHash, cleanHash)
		}
	})

	t.Run("input_slices_are_not_mutated", func(t *testing.T) {
		// DESIGN §3: the sort operates on copies; the caller's Node fields
		// keep their order.
		evidence := []string{eviB, eviA}
		addresses := []string{"10.0.0.2", "10.0.0.1"}
		n := Node{Kind: KindHost, Label: "dc01", EvidenceIDs: evidence, Addresses: addresses}
		if _, err := CanonicalContent(n); err != nil {
			t.Fatalf("CanonicalContent: %v", err)
		}
		if !slices.Equal(evidence, []string{eviB, eviA}) {
			t.Errorf("input EvidenceIDs was mutated to %v", evidence)
		}
		if !slices.Equal(addresses, []string{"10.0.0.2", "10.0.0.1"}) {
			t.Errorf("input Addresses was mutated to %v", addresses)
		}
	})

	t.Run("invalid_utf8_is_rejected_not_normalized", func(t *testing.T) {
		// encoding/json would replace invalid bytes with U+FFFD before cjson
		// ever sees them (BACKLOG 2026-09-24): CanonicalContent owns the
		// field-level A0-2.3 check at the fingerprint boundary.
		n := Node{Kind: KindHost, Label: "dc01", Summary: "valid\xfe\xffinvalid"}
		_, err := CanonicalContent(n)
		if err == nil {
			t.Fatal("invalid UTF-8 in summary must be rejected, never U+FFFD-normalized into a digest")
		}
		if got := errs.KindOf(err); got != errs.Validation {
			t.Errorf("kind = %q, want validation (A0-2.3)", got)
		}
		if !strings.Contains(err.Error(), "field=summary") {
			t.Errorf("message must name the field (A2-10.4): %v", err)
		}
		// attrs string values are covered by the same walk.
		n2 := Node{Kind: KindHost, Label: "dc01", Attrs: Attrs{"k": {Type: AttrString, Str: "\xff"}}}
		if _, err := CanonicalContent(n2); err == nil || errs.KindOf(err) != errs.Validation {
			t.Errorf("invalid UTF-8 in an attrs value: err = %v, want validation", err)
		}
	})

	t.Run("hand_built_attr_defects_are_internal", func(t *testing.T) {
		// A2-4.6's defect rule: a canonical content document that cannot
		// come from decoded wire data is a platform defect → internal.
		zeroType := Node{Kind: KindHost, Label: "dc01", Attrs: Attrs{"k": {}}}
		if _, err := CanonicalContent(zeroType); err == nil || errs.KindOf(err) != errs.Internal {
			t.Errorf("zero AttrValue: err = %v (kind %q), want internal", err, errs.KindOf(err))
		}
		bigInt := Node{Kind: KindHost, Label: "dc01", Attrs: Attrs{"k": {Type: AttrInt, Num: 1 << 62}}}
		if _, err := CanonicalContent(bigInt); err == nil || errs.KindOf(err) != errs.Internal {
			t.Errorf("out-of-range hand-built AttrInt: err = %v (kind %q), want internal (A0-2.6)", err, errs.KindOf(err))
		}
	})
}

func TestContentDocKeySetIsFixedTwenty(t *testing.T) {
	// The pin: the key list is parsed from A2-4.6's contract text, never
	// transcribed from the struct (WP-14 pinning rule).
	text := contractText(t, contractRelPath)
	wantKeys := parseContentDocKeys(t, contractSlice(t, text, "### A2-4", "### A2-5"))
	if len(wantKeys) != 20 {
		t.Fatalf("A2-4.6 parses to %d keys, want exactly 20: %v", len(wantKeys), wantKeys)
	}
	if !sort.StringsAreSorted(wantKeys) {
		t.Fatalf("A2-4.6's published list is not in UTF-8 byte order: %v", wantKeys)
	}

	t.Run("struct_tags_are_the_contract_key_set_in_byte_order", func(t *testing.T) {
		got := contentDocJSONKeys(t)
		if !slices.Equal(got, wantKeys) {
			t.Fatalf("contentDoc json tags %v != A2-4.6 contract keys %v", got, wantKeys)
		}
	})

	t.Run("zero_value_canonical_bytes_carry_exactly_the_20_keys", func(t *testing.T) {
		// A2-4.6's own reflection oracle (TestContentDocFixedKeySet): the
		// canonical bytes of a zero-valued contentDoc contain exactly the 20
		// keys, in byte order, with zero values — [] and {} not null
		// (A0-2.14).
		zero := contentDoc{Addresses: []string{}, Attrs: Attrs{}, EvidenceIDs: []string{}}
		b, err := cjson.CanonicalValue(zero)
		if err != nil {
			t.Fatalf("CanonicalValue(zero contentDoc): %v", err)
		}
		got := canonicalTopLevelKeys(t, b)
		if !slices.Equal(got, wantKeys) {
			t.Fatalf("canonical zero-doc keys %v != A2-4.6 keys %v", got, wantKeys)
		}
		if bytes.Contains(b, []byte("null")) {
			t.Errorf("canonical zero-doc contains null (A0-2.14 forbids it): %s", b)
		}
		if !bytes.Contains(b, []byte(`"addresses":[]`)) || !bytes.Contains(b, []byte(`"attrs":{}`)) {
			t.Errorf("zero collections must serialize as [] and {}: %s", b)
		}
	})

	t.Run("canonicalization_goes_through_the_node_content_path", func(t *testing.T) {
		// CanonicalContent of a Node with no content produces the same bytes
		// as the zero document: proof the builder normalizes nil
		// collections (A0-2.14) and uses no exclusion list (A2-4.6: the
		// A0-2.12 exclusion list is EMPTY — every one of the 20 keys is
		// always present).
		got, err := CanonicalContent(Node{})
		if err != nil {
			t.Fatalf("CanonicalContent(Node{}): %v", err)
		}
		zero := contentDoc{Addresses: []string{}, Attrs: Attrs{}, EvidenceIDs: []string{}}
		want, err := cjson.CanonicalValue(zero)
		if err != nil {
			t.Fatalf("CanonicalValue: %v", err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("CanonicalContent(Node{}) = %s, want the zero document %s", got, want)
		}
	})

	t.Run("no_float_and_no_time_fields", func(t *testing.T) {
		// A0-2.6: a type that is ever canonicalized MUST NOT declare a float
		// field (BACKLOG 2026-09-24: enforceable only by review — this is
		// the review, mechanized). And no time.Time anywhere in the
		// canonicalized types: timestamps are timex strings (A0-5.1).
		forbidFloatAndTime(t, reflect.TypeOf(contentDoc{}), "contentDoc")
	})

	t.Run("mutation_proof", func(t *testing.T) {
		mutated := mutateContract(t, text, "(A0-2.14): `addresses`, `attrs`, `basis`", "(A0-2.14): `addresses`, `attrs`, `bases`")
		path := writeTempContract(t, mutated)
		got := parseContentDocKeys(t, contractSlice(t, contractText(t, path), "### A2-4", "### A2-5"))
		if slices.Equal(got, contentDocJSONKeys(t)) {
			t.Fatalf("the pin did not fail on the mutated copy: basis→bases still parsed as the struct's key set %v", got)
		}
	})
}

// contentDocJSONKeys returns contentDoc's json tag names in struct field
// order (which must be A0-2.4 byte order), without any tag options.
func contentDocJSONKeys(t *testing.T) []string {
	t.Helper()
	typ := reflect.TypeOf(contentDoc{})
	keys := make([]string, 0, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		tag := typ.Field(i).Tag.Get("json")
		name, opts, _ := strings.Cut(tag, ",")
		if name == "" || name == "-" || opts != "" {
			t.Fatalf("contentDoc field %s: json tag %q — the fixed key set has no omitempty and no unnamed field (A0-2.14)", typ.Field(i).Name, tag)
		}
		keys = append(keys, name)
	}
	return keys
}

// canonicalTopLevelKeys walks the top-level keys of a canonical document in
// emission order.
func canonicalTopLevelKeys(t *testing.T, b []byte) []string {
	t.Helper()
	dec := json.NewDecoder(bytes.NewReader(b))
	tok, err := dec.Token()
	if err != nil || tok != json.Delim('{') {
		t.Fatalf("canonical bytes do not start with an object: %v (%v)", tok, err)
	}
	var keys []string
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			t.Fatalf("reading key: %v", err)
		}
		key, ok := tok.(string)
		if !ok {
			t.Fatalf("non-string key token %T", tok)
		}
		keys = append(keys, key)
		if err := skipValue(dec); err != nil {
			t.Fatalf("skipping value of %q: %v", key, err)
		}
	}
	return keys
}

// skipValue consumes the next value from the decoder whole: a scalar in one
// token, a container by counting delimiter depth (encoding/json has no
// exported value-skip on this toolchain).
func skipValue(dec *json.Decoder) error {
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	if _, ok := tok.(json.Delim); !ok {
		return nil // a scalar value is one token
	}
	depth := 1
	for depth > 0 {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		if d, ok := tok.(json.Delim); ok {
			switch d {
			case '{', '[':
				depth++
			case '}', ']':
				depth--
			}
		}
	}
	return nil
}

// forbidFloatAndTime fails the test if typ (or a struct/map/slice element
// type reachable from it, one level of named graph types deep) declares a
// float or time.Time field.
func forbidFloatAndTime(t *testing.T, typ reflect.Type, name string) {
	t.Helper()
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		switch f.Type.Kind() {
		case reflect.Float32, reflect.Float64:
			t.Errorf("%s.%s is a float field — forbidden in a canonicalized type (A0-2.6)", name, f.Name)
		case reflect.Struct:
			if f.Type == reflect.TypeOf(time.Time{}) {
				t.Errorf("%s.%s is a time.Time — timestamps are A0-5.1 strings", name, f.Name)
			}
			forbidFloatAndTime(t, f.Type, name+"."+f.Name)
		case reflect.Slice, reflect.Array:
			checkElem(t, f.Type.Elem(), name+"."+f.Name)
		case reflect.Map:
			checkElem(t, f.Type.Elem(), name+"."+f.Name)
		}
	}
}

func checkElem(t *testing.T, elem reflect.Type, name string) {
	t.Helper()
	switch elem.Kind() {
	case reflect.Float32, reflect.Float64:
		t.Errorf("%s holds floats — forbidden in a canonicalized type (A0-2.6)", name)
	case reflect.Struct:
		if elem == reflect.TypeOf(time.Time{}) {
			t.Errorf("%s holds time.Time — timestamps are A0-5.1 strings", name)
		}
		forbidFloatAndTime(t, elem, name)
	}
}
