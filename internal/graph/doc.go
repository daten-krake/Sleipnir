// Package graph holds the A2 contract types: the closed node and edge kind
// lists, Node, Edge, Provenance and the bounded flat Attrs escape hatch, and
// the A2-4.6 content document with its fingerprint. It is the graph's
// types-and-invariants layer — no I/O, no store, no policy, no HTTP.
//
// Layer: domain types (DESIGN §1). The package imports exactly four internal
// packages, each with one job:
//
//   - internal/errs — every returned error is an errs value with an A0-3.1
//     kind and the ADR-0019 §2 message shape (component.Function is captured
//     by errs itself).
//   - internal/ids — A0-1.5 byte-exact id validation inside
//     Provenance.Validate; no graph id shape is re-implemented here.
//   - internal/timex — A0-5.1/5.3 timestamp rejection inside
//     Provenance.Validate; the A0-5 parsing rules are never re-implemented
//     here. (The A2 §4 sketch header lists only ids/cjson/errs; the WP-14
//     brief admits timex, and A2-5.3's timestamp fields make it necessary —
//     re-deriving A0-5.3 locally would violate A0-2.1's one-implementation
//     principle. Ruling recorded below.)
//   - internal/cjson — the one canonical JSON form (A0-2.1): CanonicalValue
//     for the content document, SHA256Hex and DigestEqual for the digest
//     (A0-2.15).
//
// Deliberately NOT imported: internal/caps (no cap is enforced here — cap
// enforcement is the write validation pass, WP-15; the one A2-local constant
// that is not an A0-7.1 registry row, ProvenanceMaxEntries, is defined in
// provenance.go) · internal/secretscan (the A2-9.4 ingest scan is A2-10.2
// step 10 and belongs to the write path, WP-15, including the caller-side
// platform-minted exemption of the 2026-09-29 erratum — no clause in
// A2-1…A2-6 mandates Scan inside the types package) · internal/policy
// (A2-8.2: graph never imports policy) · stdlib time (timestamps are A0-5.1
// strings; no time.Time appears in any served type). Proved mechanically by
// `go list -deps ./internal/graph | grep daten-krake`.
//
// # Clauses implemented here
//
//   - A2-1.2/1.4/1.4a/1.5/1.6 — Node/Edge field set and JSON spellings:
//     graph_node_id / graph_edge_id / source_id / target_id / supersedes_id /
//     superseded_by_id / agent_node_id, graph_seq, engagement_id. A bare
//     node_id or id key exists nowhere in this package (TestNoBareNodeIDInGraphDocuments).
//   - A2-1.7 — closed value types: no float, no null, no nested structure
//     except flat attrs; reflection gates in the tests.
//   - A2-1.8 — Node.Current(): current iff superseded_by_id is absent.
//   - A2-2.1…A2-2.7 — the ten closed node kinds, Severity, FindingStatus,
//     HypothesisStatus, CredentialKind, MediaKind as exported constants
//     (A0-8.5), and the Node field set of A2-2.2/A2-2.3. There is NO
//     confidence field on Node or on any kind: the grade lives once per
//     provenance entry (A2-2.7, ADR-0022).
//   - A2-3.1/3.2 — the seven closed edge kinds and the endpoint matrix as
//     data: EdgeKind.AllowsEndpoints, including supersedes' same-kind rule.
//   - A2-3.8/3.9 — Edge.SourceKind/TargetKind denormalization fields and the
//     mutable Retracted flag (the mutation itself is the seam's, WP-15/17).
//   - A2-4.6 — contentDoc: the fixed 20-key content document, zero values
//     not absence (A0-2.14), empty exclusion list, addresses/evidence_ids
//     sorted ascending by unsigned byte value and deduplicated on copies
//     BEFORE canonicalization, non-nil collections, UTF-8 gate, canonical
//     bytes via cjson. CanonicalContent returns the exact bytes; ContentHash
//     their SHA-256 (A0-2.15). §4.2's vectors F1/F1-R/F3/S1 are reproduced
//     byte-exactly (TestContentHashVector), and the published
//     defective-implementation marker 4a017e71…ce9cae is proven to be the
//     unsorted canonicalization and never produced by this implementation
//     (TestContentHashStableAcrossArrayOrder).
//   - A2-4.8/4.9 — VerifyContentHash recomputes from STORED bytes (A0-2.16),
//     compares in constant time (A0-2.15) and classifies every mismatch as
//     internal, never integrity_failed.
//   - A2-5.1/5.3/5.6/5.7 (type-level half) — Provenance, ValidateProvenance
//     (non-empty → validation, over ProvenanceMaxEntries → summary_too_large
//     per A2-7.1 mechanism R) and Provenance.Validate (closed enums,
//     required and conditional fields per A2-5.3's table, byte-exact ids per
//     A0-1.5, A0-5 timestamps through timex).
//   - A2-6.1…A2-6.3 — Attrs/AttrValue/AttrType with hand-written codecs:
//     bare-scalar wire form, Type discriminator never serialized, round-trip
//     preserves Type (P-51); UnmarshalJSON rejects nested values, floats,
//     null, lone surrogates, malformed and reserved keys with validation.
//   - A2-6.5 — attrs is part of the fingerprint (vectors F1 vs F3 lock it).
//
// # Disposition of the row's remaining clauses (by number)
//
//   - A2-1.1 — per-engagement property graph in PostgreSQL behind the store
//     seam: this package is its type layer (no I/O, no store, no query
//     language, no traversal API — Q1); persistence → WP-17/WP-22.
//   - A2-1.3 — content immutability: no content setter exists on Node or
//     Edge by construction; the four mutable fields (quarantined,
//     quarantine_reason, report_excluded, retracted) change only through the
//     seam mutators, and the conflict enforcement is the write path's →
//     WP-15/WP-17.
//   - A2-1.9 — no merge operation: discharged by absence — no merge kind,
//     method or field exists anywhere in this package; adding one needs an
//     ADR.
//   - A2-5.5 — observed_claimed_at drives no ordering, supersession,
//     quarantine, expiry or digest: the digest half is implemented here (none
//     of the 20 contentDoc keys is a timestamp; the field rides only inside
//     Provenance and never enters CanonicalContent); the ordering half
//     (graph_seq/recorded_at) belongs to the write path → WP-15/WP-18.
//   - A2-6.4 — attrs caps (AttrsMaxKeys/AttrValueMaxBytes/AttrsTotalMaxBytes,
//     mechanism R, measured on the A0-2 canonical form): registry rows with
//     their single Go definition in internal/caps; enforcement at A2-10.2
//     step 8 → WP-15, which is why this package imports no caps.
//
// # Clauses deliberately NOT implemented here (who owns them)
//
//   - A2-2.3 per-kind required/not-applicable enforcement, A2-2.8's setter
//     matrix, A2-3.3…A2-3.7 (cardinality, collapse, contradicts
//     normalization, self-edge), A2-4.1…A2-4.5/4.7 (supersession, dedup,
//     provenance append and the verified-independence rule), A2-7 (caps),
//     A2-9.4 (secret scan incl. the platform-minted exemption), A2-10
//     (NodeDraft/PendingNode/NewNode/WriteNode/EdgeDraft/PendingEdge/
//     NewEdge/WriteEdge and the validation order), A2-11 enforcement,
//     A2-5.2's "write types carry no provenance field" (the draft types are
//     the proof): the validation + write path, WP-15, in this package's
//     later files (validate.go/pending.go per the plan).
//   - A2-8 quarantine in full, including QuarantineState and the
//     QuarantineDecider interface: WP-16 (quarantine.go). This package
//     carries only the QuarantineReason enum the Node field needs.
//   - A2-10.6's four seam mutation methods and A2-11.1's engagement-scoped
//     reads: the store seam, WP-17 (internal/store).
//   - A2-5.4 (event_id resolves in this engagement), A2-4.7's ordering by
//     the A1 event seq, TestPrincipalKindIsNotCopiedFromActor and
//     TestFieldNameMappingIsTotal: the ingest mapping, WP-18 — they need
//     the event log, which a types package does not have.
//   - The shared contract-suite ids (TestContentHashRecomputedFromStoredBytes,
//     TestReservedAttrKeysRejected, TestHardRejectMatrix, …): WP-21, in this
//     package's contract_test.go. Where the WP-14 nine-id rule covers the
//     same oracle, it ships as a subtest of the nine (e.g. A2-4.8's
//     recompute oracle under TestContentHashVector, A2-6.3's reserved-key
//     oracle under TestAttrsRejectNestedFloatNull).
//   - A2-12 (the seam to A3): nothing to implement until A3 exists.
//
// # Rulings
//
//   - The evidence grade IS spelled confidence. The WP-14 brief says "there
//     is no confidence field anywhere"; the frozen contract says otherwise:
//     A2 §4 declares type Confidence and Provenance.Confidence
//     `json:"confidence"`, A2-5.3's table requires the field, A2-6.3's
//     reserved list contains it, and ADR-0022 itself rules that
//     "'Confidence' in the Sleipnir graph means the provenance evidence
//     grade". The brief's own conflict rule (contract wins, surface it) is
//     applied: Confidence is implemented exactly per the sketch. What does
//     NOT exist anywhere, per ADR-0022 and A2-2.7, is a finding-level
//     confidence field or an adjective (low/medium/high) scale —
//     TestProvenanceMandatory rejects that scale byte-exactly. The stale
//     name TestNodeDedupIncludesConfidence on the WP-15 review row is
//     surfaced for the principal, not implemented.
//   - The grade is NOT inside the content fingerprint, contra ADR-0022's
//     Decision bullet ("inside the content fingerprint (A2-4.8)"). Frozen
//     A2-4.6 excludes provenance from the 20-key document entirely, and
//     §4.2's normative vectors — independently recomputed before being
//     encoded — contain no confidence key. A2-4.8, which ADR-0022 cites, is
//     the recompute-from-stored-bytes rule and says nothing about
//     provenance. An Accepted ADR outranks a contract, but this ADR bullet
//     is factually inconsistent with the frozen A2 it cites and with the
//     vectors it would invalidate; the vectors are the authority the WP-14
//     brief pins. RULED (product owner, 2026-10-07): an erratum note inside
//     ADR-0022 records that the grade lives in the append-only provenance
//     entry, OUTSIDE the digest — immutability comes from provenance being
//     append-only (A2-5.6, A2-2.8), not from fingerprint inclusion. The
//     shipped fingerprint is A2-4.6's (and a grade change therefore does NOT
//     force a revision — provenance growth never changes content_hash, which
//     is exactly what A2-4.7's dedup collapse relies on).
//   - The four "A2-local" cap constants of the A2 §4 sketch const block
//     (NodeLabelMaxBytes, HypothesisClaimMaxBytes, HypothesisBasisMaxBytes,
//     AddressMaxBytes) are NOT redefined here: they are A0-7.1 registry rows
//     with a single Go definition in internal/caps, and a second definition
//     would be an A0-7.2 defect. A2-7.1's "(A2-local)" markers for those
//     four contradict the A0-7.1 registry as adopted by AM-2 — ruled a
//     stale-marker erratum (product owner, 2026-10-07, A2 §6 item 16);
//     internal/caps governs. Only
//     ProvenanceMaxEntries (no registry row) is defined here, per the sketch.
//   - ReservedAttrKeys is unexported. A2-6.3's sketch shows an exported
//     `var ReservedAttrKeys = map[string]bool`; an exported package-level map
//     is mutable state any package can write (DESIGN §4 forbids it, and
//     DESIGN outranks the illustrative, "not compiled" sketch). The set
//     itself is transcribed byte-exactly from A2-6.3 (51 keys, pinned
//     against the contract text in tests with a mutation proof) and every
//     consumer — WP-15's validation, WP-21's contract suite — lives in this
//     package.
//   - Attrs has no hand-written MarshalJSON. encoding/json's map encoding
//     already emits exactly A2-6.1's wire form (sorted keys, bare scalars
//     via AttrValue.MarshalJSON). A hand-written marshaler returning {} for
//     a nil Attrs would paper over the A0-2.14/A2-4.6 non-nil requirement,
//     which must surface as a defect; the sketch's "custom
//     MarshalJSON/UnmarshalJSON" is discharged by AttrValue's pair plus
//     Attrs.UnmarshalJSON.
//   - UnmarshalJSON error mapping stops at the type boundary. json.Unmarshal
//     runs its own scanner first, so a syntactically invalid document (01,
//     +1, trailing data) fails as a foreign *json.SyntaxError before any
//     codec here runs; rendering foreign decode errors into the A0-3
//     vocabulary is the request-decode path's job (WP-15, following
//     internal/paging's ruling). The type-level guards for those forms exist
//     and are tested through direct UnmarshalJSON calls.
//   - tool_version without tool_id is rejected. A2-5.3's Req column reads
//     "with tool_id"; the strict reading is applied because an orphan
//     version string has no registry entry to attribute it to (A0-1.3 keeps
//     name/version out of the id). The principal-id conditionals (job_id,
//     task_id, agent_node_id, user_id) are enforced as "required when",
//     never "forbidden otherwise": §4.1's frozen example shows a platform
//     entry carrying a job_id.
//   - Field-level UTF-8 validation is owned here, at the fingerprint
//     boundary (BACKLOG 2026-09-24's open question, answered for the graph):
//     encoding/json replaces invalid UTF-8 with U+FFFD before cjson can see
//     it, so CanonicalContent checks every string of the document itself and
//     rejects with validation naming the field and the byte length — never
//     echoing the value (the secret scan has not run yet at A2-10.2 step
//     13a's position, and A2-9.5's discipline applies early). Checks walk in
//     contentDoc key order so the first reported field is deterministic
//     (A2-10.2: exactly one error).
//   - Hand-built AttrValue defects (unknown Type discriminator, integer
//     outside A0-2.6) are errs.Internal and are pre-checked in
//     contentDocOf: cjson.CanonicalValue would reclassify a marshaler
//     failure as validation, and A2-4.6's defect rule says a canonical
//     document that cannot come from decoded wire data is a platform defect.
//   - Nil collections on a Node normalize to empty in the content document
//     ({} / [] per vector S1), they are not an error: absence is the legal
//     API representation (A0-8.3) and A0-2.14's non-nil duty falls on the
//     document builder, which is contentDocOf.
//   - Provenance.Validate does not check size caps (A2-7, WP-15's pass,
//     which is why the 64-byte tool_version cap is not applied here), secret
//     material (A2-5.8/A2-9.4, WP-15), event resolution (A2-5.4, needs the
//     store) or list ordering by event seq (A2-4.7, needs the event log).
//     The over-bound list is summary_too_large because A2-7.1 assigns
//     mechanism R to the ProvenanceMaxEntries class; the empty list is
//     validation per A2-5.1/A2-5.7. A2-5.7's caller-dependent split —
//     internal when the PLATFORM ingest path failed to produce a run_id or
//     event_id, validation when a supplied id is malformed — is not decidable
//     at type level (a type cannot see its caller), so the kinds returned
//     here are the type-level defaults: WP-15 owns the reclassification at
//     A2-10.2 step 12, and since A2-5.2 makes provenance non-suppliable by
//     clients, an empty list or a missing run_id/event_id on the ingest path
//     is A2-5.7's internal bullet. A silent WP-15 pass-through would ship
//     the wrong kind; TestProvenanceMandatory pins the type-level defaults,
//     not the ingest classification (owner recorded 2026-10-07 after the
//     independent review flagged the gap).
//   - Closed-list pinning: TestNodeKindListClosed, TestEdgeKindListClosed
//     and the 20-key/reserved-set pins parse contracts/A2-graph.md directly
//     (relative path from this package), never a transcription of this
//     package's consts, and each carries a mutation proof on a copy outside
//     the repo (t.TempDir). If the contract's table shape changes, the
//     parsers fail loudly — fix the parser, never weaken the pin.
//   - Exactly nine top-level test ids exist (the WP-14 review row's list);
//     every other oracle is a subtest of one of them.
package graph
