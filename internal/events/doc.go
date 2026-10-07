// Package events is the A1 event-log contract's type surface: the one
// 17-key envelope (A1-1), typed actors and the closed platform-component
// vocabulary (A1-2), the closed 42-kind taxonomy (A1-3), the per-kind
// untrusted-field registry (A1-4.4), and exactly one flat payload struct per
// kind (A1-4.1, A1-4.8, A1-4.12).
//
// Layer: domain types (DESIGN §1) — types + invariants, no I/O, no store, no
// clock, no HTTP.
//
// # Imports
//
// The package imports nothing from internal/ — `go list -deps
// ./internal/events | grep daten-krake` lists only this package — and from
// the standard library only slices, for the defensive copies AllKinds and
// UntrustedFields return. Each foundation package the A1 §4 sketch allows a
// domain type to import is unnecessary for a types-only surface, and
// importing one "for later" would be a speculative dependency (DESIGN §2):
//
//   - errs — nothing here returns an error yet: the composition/validation
//     pass that produces errs kinds (A1-4.2–4.7, A1-4.9, A1-7.10) is WP-11's
//     and imports errs where it lands.
//   - ids — A0-1.5 id validation happens at composition (WP-11); the
//     envelope and payloads carry ids as plain strings (A1-4.11).
//   - timex — every timestamp of a canonicalized type is a Go string holding
//     an A0-5.1 value (A1-4.12); nothing here formats or parses time.
//   - cjson — canonicalization (Preimage, Served) is the hash-chain work
//     package's (WP-10, A1-5).
//
// # Clauses implemented here
//
//   - A1-1.1/1.2/1.3 — Event: one envelope shape, 17 closed top-level keys,
//     every key always present with its zero value when unset, no omitempty
//     anywhere. ChainBlock is embedded so seq, prev_hash and hash serialize
//     at the envelope's top level: A0-2.12 exclusion lists take plain
//     top-level names only, and a nested "chain":{…} object would be
//     unexcludable and is forbidden.
//   - A1-1.5 / A1-4.8 — flatness: payload fields are strings, integers,
//     booleans or arrays of strings; no nested objects, no arrays of
//     objects, no maps, no pointers, no floats (A0-2.6). TestPayloadsAreFlat
//     is the reflection merge-gate sessions/BACKLOG.md (2026-09-24) asked for
//     when the domain types landed.
//   - A1-2.1 — Actor: exactly three keys, all always present; ActorType is
//     the closed five-value enum A2-5.3 adopts verbatim.
//   - A1-2.2 — shape only: PrincipalID is the A0-1.2 identifier selected by
//     Type, held as a string; byte-exact validation is the composition
//     path's job (WP-11).
//   - A1-2.4 — Component: the closed twelve-value platform-subsystem enum.
//   - A1-3.1/3.3 — Kind, the 42 constants, AllKinds (the closed list in
//     table order), and the 42 payload structs transcribed 1:1 from the
//     A1-3.3 tables: field names, field order and json tags byte-exact.
//   - A1-3.4 — ClientAppendable: the three C kinds (command_executed,
//     task_result, revert_recorded) and nothing else, byte-exact.
//   - A1-4.1 — exactly one payload type per kind, a flat struct with an
//     explicit json tag on every field; no shared payload base type, no
//     per-kind map[string]any, no field optional in the sense of absent.
//   - A1-4.4 — UntrustedFields: the registry of `*`-marked fields per kind
//     ("normative: this table is what UntrustedFields returns"). The
//     untrusted-flag computation itself is the composition path's (WP-11).
//   - A1-4.12 — timestamps are strings, never time.Time, in every
//     canonicalized type (TestNoTimeTimeInCanonicalizedTypes); every :int is
//     int64 except exit_code, the only field A1-4.2 ranges; enums are
//     byte-exact closed-list strings.
//
// # Disposition of the row's remaining clauses (by number)
//
//   - A1-1.4 — time semantics: the envelope carries recorded_at, occurred_at
//     and occurred_claimed_at as strings; stamping, clamping (A1-5.4 →
//     WP-10) and the untrusted marking of occurred_claimed_at (A0-5.7,
//     A1-4.4) are composition behaviour → WP-11.
//   - A1-1.6 — no post-commit-mutable fields: satisfied structurally — no
//     method here mutates an Event, and every annotation is its own kind
//     (chain_break_detected, report_inclusion_changed, …). Append-only
//     enforcement on the write path → WP-11 and the store seam (WP-17).
//   - A1-1.7 — evidence references: evidence_refs and every *_evidence_id
//     field are transcribed as strings; evi_ kind validation (A0-1.5) and the
//     A1-4.7/7.3 evidence_refs derivation are composition → WP-11.
//   - A1-1.8 — served representation: the Served type and its canonical
//     17-key bytes are A1-5.7 → WP-10.
//   - A1-2.3 — platform-only composition, caller-supplied platform-stamped
//     fields rejected: write path → WP-11 (negatives
//     TestClientCannotSupplyEnvelopeFields, TestUntrustedFlagCannotBeSupplied
//     → WP-20).
//   - A1-2.5 — engagement_id derived from the token binding, body-supplied
//     rejected, cross-engagement → notfound: request handling → WP-11/WP-18
//     and the api edge; nothing here types it beyond the envelope field.
//   - A1-2.6/2.7 — machine-principal append limits and the unreachability of
//     actor.type="platform": write-path enforcement → WP-11 (negatives →
//     WP-20).
//   - A1-2.8 — the A2 provenance pair (source_event_id + graph ids): field
//     names implemented byte-exact in payload_graph.go; consumption is A2's
//     (internal/graph, WP-18 ingest).
//   - A1-3.2 — the payload-column notation (name:type, (N) caps, * stars,
//     enum{…}, array[string]): the transcription rule this package follows;
//     contract_test.go parses exactly this notation.
//   - A1-3.5 — kinds are added only by contract revision: pinned by
//     TestKindListIs42AndClosed (42, unique, table order) including the
//     absent-node_superseded subtest.
//   - A1-3.6 — A2-coordinated field names: implemented in payload_graph.go's
//     tags, byte-exact.
//   - A1-3.7/A1-3.8 — SPEC §5 coverage table and the A2 cross-contract
//     answers: completeness/governance text, no code obligation here.
//
// # Clauses deliberately not implemented here, and who owns them
//
//   - A1-4.2–4.7, A1-4.9, A1-7 — the composition/validation pass: NewEvent,
//     Draft, Binding, AppendRequest, DedupRecord, UnmarshalEvent, cap
//     enforcement, array sorting/dedup, the secret scan (with the
//     platform-minted exemption of the 2026-09-29 erratum), reference
//     resolution. Owner: WP-11. The Payload interface already declares
//     Validate() per the §4.1 sketch; its implementations land with WP-11,
//     so no payload struct in this package satisfies Payload yet. That is
//     deliberate: a Validate returning nil would stub a safety path
//     (AGENTS.md), and half of the pass cannot be written inside a
//     types-only package.
//   - A1-5 — the hash-chain primitives: Preimage, HashEvent, Served,
//     ChainSpecV1, ChainZero, ChainExclude, HeadLogIntervalSeq — and
//     ChainHead, which A1-5.6 (inside WP-10's A1-5.1…5.10 row) defines as
//     the write-side chain-head state; WP-12 consumes it. Owner: WP-10.
//   - A1-6 — the verification walk and its bookkeeping types: VerifyResult,
//     IntegrityState, ExportIntegrity. Owner: WP-12.
//     VerifyTrigger and BreakKind DO live here because A1-3.3's payload
//     structs are typed with them (chain_verified, chain_break_detected).
//   - A1-4.5 cap constants — not redeclared here: the A0-7.1 registry exists
//     as internal/caps (ProseLongMaxBytes … IdempotencyKeyMaxBytes), A1-4.5
//     itself says A1 "declares no constant of its own", and the §4.1
//     sketch's "A1-local cap constants" block predates that package. WP-11
//     imports internal/caps.
//   - A0-2.3 field-level UTF-8 validation — no owner exists
//     (sessions/BACKLOG.md, 2026-09-24) and §4.1's sketch assigns it to
//     nobody, so no mechanism is invented here. Flagged for the principal in
//     the WP-09 report.
//   - The remaining §4.4 contract-suite test ids — owners: WP-10/WP-11/WP-12
//     per clause, suite assembly WP-20.
//
// # Rulings made for this contract
//
//   - UntrustedFields is (Kind) → ([]string, bool): A1-4.4 names the
//     function but §4.1 never sketches it (a contract gap, reported), so the
//     signature is ruled here. The bool distinguishes "kind has no starred
//     field" (23 of 42) from "not a kind at all"; the slice is a fresh copy,
//     so callers can never mutate the registry (DESIGN §4: no package-level
//     mutable state — kinds and untrustedFields are read-only data handed
//     out as copies).
//   - AllKinds() is exported even though the sketch shows no enumerator:
//     closedness is untestable and downstream obligations uniterable without
//     one ("all 42 kinds", A1-4.5's TestMaximalPayloadFitsCanonicalBound).
//     It returns a copy of the unexported table.
//   - chain_break_detected.break_kind is typed BreakKind and
//     chain_verified.trigger VerifyTrigger, exactly as the §4.1 sketch
//     shows; every other enum field is a plain string, following the
//     sketch's shown payloads (TaskResultPayload.Status,
//     IntegrityOverridePayload.Scope, QuarantineRecomputedPayload.Trigger).
//     Both are named string types, so A1-4.12's string transcription holds.
//   - The sketch's QuarantineRecomputedPayload comment ("policy_changed") is
//     stale against normative A1-3.3 (scope_changed | blacklist_changed |
//     node_written | edge_written); the struct documents the normative list
//     and the defect is reported, per "published literals win".
//   - exit_code is Go int; every other :int — including attempt,
//     http_status, containers_killed and all counts — is int64 (A1-4.12's
//     transcription rule).
//   - evidence_stored.sha256 is Go field SHA256 (acronym casing, DESIGN §9);
//     the json tag stays "sha256".
//   - No constructor in this package: DESIGN §4's "construct via NewX that
//     validates" is discharged by NewEvent (§4.1 sketch), which the
//     composition pass (WP-11) owns. A partial constructor here would hand
//     out half-built canonicalized values (A1-1.2's non-nil slice rule
//     needs the full composition context).
//   - Kind() methods exist on all 42 payload structs now: the kind↔payload
//     binding is taxonomy, not validation, and
//     TestPayloadStructsMatchA1Tables calls Kind() on every struct.
//
// # Tests
//
// The seven WP-09 test ids (TestEnvelopeKeySetIsSeventeen,
// TestKindListIs42AndClosed, TestActorComponentIsEmptyForNonPlatform,
// TestPayloadStructsMatchA1Tables, TestPayloadsAreFlat,
// TestNoTimeTimeInCanonicalizedTypes,
// TestUntrustedFieldsCoversEveryStarredField) pin this package against the
// frozen contract TEXT: contract_test.go parses ../../contracts/A1-events.md
// (A1-1.1's envelope table, A1-3.3's taxonomy rows, A1-4.4's star table,
// A1-2.1/2.2/2.4's vocabulary lists) and the assertions compare that parse
// against the implementation — never the implementation against itself.
// Every parse failure is a t.Fatalf, so a contract restructure breaks the
// oracle loudly instead of silently weakening a pin. The mutation proof
// (each test fails on a mutated copy of the contract or implementation) was
// executed on copies outside the repository; see the WP-09 report.
package events
