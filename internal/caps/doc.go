// Package caps is the foundation-layer home of the A0-7 size caps: the one
// cap registry (A0-7.1/A0-7.2), the measurement rule (A0-7.3) and mechanism T,
// truncate + marker (A0-7.4/A0-7.5). Per DESIGN §1 it is a foundation package:
// it imports nothing from this module — not even internal/errs — because
// nothing it returns is an error (Truncate and Fits report a bool), and it has
// no package-level state.
//
// # The registry is the one source
//
// Every cap constant of the A0-7.1 table lives in this package exactly once
// (A0-7.1: "one table, one const block"). An owning contract cites these
// names; a second definition of the same value anywhere in the platform is a
// defect (A0-7.2), which is why the values are also pinned literally, row by
// row, in TestConstantsMatchA0Table — the constants are contract constants and
// change only via ADR (A0-7.2), never as a side effect of an edit here.
//
// Where two contracts capped the same thing, the registry holds one name and
// one value: ToolVersionMaxBytes = 64 (A2-7.1's 32 is a defect, A0-7.1) and
// EvidenceRefsMax = 8, which replaces A2's EvidenceIDsMax name.
//
// # Clauses implemented here
//
//   - A0-7.1 — the registry itself (all 28 constants) plus TruncationMarker.
//   - A0-7.3 — the measurement rule for a field cap: UTF-8 bytes of the
//     decoded string value, len(s) in Go, never rune count, never UTF-16.
//     Fits is that rule; document caps (A0-7.3's second sentence) are measured
//     by their serializers, see "Deliberately absent" below.
//   - A0-7.4 — a truncation cuts on a UTF-8 rune boundary.
//   - A0-7.5 — mechanism T: prefix + marker, marker counted against the cap,
//     capped output never exceeds the limit, and a limit below
//     len(TruncationMarker) is a platform defect reported as ("", true).
//
// # Clauses this package does not implement, and where they live
//
//   - A0-7.6 — mechanism R. Fits is only the pre-check; the rejection (store
//     nothing, 413 summary_too_large, a message naming field, cap and actual
//     per ADR-0019 §2) is the ingest path's decision, so no Reject or
//     summary_too_large helper exists here.
//   - A0-7.2 — the ADR requirement and the migration plan for a lowered cap
//     are process rules, not code.
//   - A0-7.7 — assigning one mechanism per capped field class is the owning
//     contract's job; the "Mechanism" column of A0-7.1 records those
//     assignments, and this package deliberately exposes no per-field registry
//     map that would restate them (a second source of truth would be an
//     A0-7.2 defect, and there is no consumer that needs a lookup at runtime).
//   - A0-7.8 — the enforcement point is the platform ingest/render path; a
//     client-side pre-check calling Fits is not enforcement.
//   - A0-7.9 — the node-count stop and nodes_truncated are a stage-view
//     builder's behaviour in a declared deterministic order; this package only
//     holds the StageViewMaxNodes value.
//   - A0-7.10 — the composition rule for jointly unsatisfiable caps is A3's
//     (interim: the smaller cap governs, escalated to the product owner).
//
// # Deliberately absent
//
// Four things a reader might expect are intentionally not here, each because
// another clause owns it and no consumer exists yet (DESIGN §2: no abstraction
// without a concrete second consumer):
//
//  1. no Reject / mechanism-R error helper (A0-7.6 assigns the rejection to
//     the caller),
//  2. no composition rule over two caps in one document (A0-7.10 assigns it to
//     A3),
//  3. no document-measuring helper (A0-7.3 requires measuring the serialized
//     JSON with the encoder settings of the response, which only the
//     serializer owns),
//  4. no per-field registry map from field name to cap (A0-7.7 assigns the
//     mechanism to the owning contract, and a map here would be a second
//     source of truth about which cap guards which field).
//
// Follow-up, not this package's work: once a third caller needs the A0-7.6
// rejection message, the helper does belong here — one line over errs.Newf
// carrying kind summary_too_large and naming field, cap and actual byte count
// (DESIGN §2: extract when the same pattern appears the third time, not the
// first).
//
// # Rulings made while writing this package
//
//   - Fits' second parameter is named limit, not cap as the illustrative
//     §4 sketch spells it. cap is a Go builtin (since 1.21) and shadowing it
//     is contrary to idiomatic Go, which AGENTS.md makes a merge gate; the
//     same sketch already uses limit for Truncate. The signature types are
//     otherwise exactly the sketch's.
//   - The constants stay untyped numerics, exactly as the §4 sketch declares
//     them, so they convert freely to int, int64, string lengths and slice
//     bounds without conversion noise at every call site. No contract requires
//     a named type, and a dedicated type would force conversions in callers
//     this package cannot predict (A0-7.1 fixes names and values, not types).
//   - Truncate takes the limit as the caller's cap value and knows only
//     len(TruncationMarker); it never looks a cap up, because field names are
//     the caller's context (A0-7.6) and a lookup map would be a second source
//     (A0-7.1).
//   - TruncationMarker is 11 bytes. A0-7.5 and the §4 sketch comment said
//     "12 B" while naming the literal "[truncated]"; the literal governs, and
//     the erratum has landed in contracts/A0-conventions.md (A0-7.5 and §6
//     item 16, product owner decision 2026-09-24). No cap value changed, so
//     it needed no ADR. Nothing in this package hardcodes either number:
//     caps.go compares against len(TruncationMarker) and the tests derive
//     their boundaries from it, so a literal change by ADR cannot leave a
//     stale constant behind.
//   - A0-7.1's own "Tests:" line also names TestCapsRejectWithSummaryTooLarge.
//     That id needs mechanism R (a rejection with field, cap and actual in the
//     message), which no code here can produce; it belongs to the first
//     ingest-path caller of Fits, and is therefore not implemented in this
//     package.
package caps
