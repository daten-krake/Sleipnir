// Package ids implements A0-1: generation and validation of entity
// identifiers.
//
// Layer: foundation (DESIGN §1). Like the other five A0 §4 foundation
// packages it imports no internal package except errs, and nothing above the
// foundation layer may be assumed about it.
//
// # What this package implements
//
//   - A0-1.1 — the body: 26 lowercase Crockford base32 characters encoding 128
//     bits (48-bit big-endian millisecond Unix time, then 80 bits of
//     crypto/rand entropy), most significant bit first, so the first body
//     character is always in 0-7.
//   - A0-1.2 — the closed set of 12 prefixes as the Kind type, with the total
//     lengths 29/30/31/35 that the table lists.
//   - A0-1.4 — generation uses crypto/rand and nothing else: no math/rand, no
//     counter, no name, no hash of user input. The entropy source is read
//     exactly once; there is no retry loop anywhere in this package.
//   - A0-1.5 — Valid, the anchored byte-exact check against a kind's regex,
//     with no Crockford normalization, no case folding and no Unicode
//     leniency. A bad id is a validation error at the boundary that checked
//     it; Valid itself returns a bool.
//
// # What this package does not implement, and why
//
//   - A0-1.3 (no slug in tool ids) and A0-1.10 (prefixes are frozen) are
//     carried by structure: Tool's id is generated like every other kind and
//     the prefixes are the Kind constants themselves, so there is no code path
//     that could rename or reuse one. The registry fields A0-1.3 mentions
//     live with the tool registry (ADR-0008).
//   - A0-1.6 (opacity) is honoured by omission: there is no body parser, no
//     creation-time accessor, no ordering helper and no prefix-registry API
//     exported here, because no clause of A0-1 asks for one (DESIGN §2). The
//     chronological order of A0-1.1 stays an internal storage property.
//   - A0-1.7 and A0-1.8 are policy for the layers that use ids: nothing here
//     can turn an id into a credential, and the node mesh token of Q9 is
//     decided by A5, not by this package.
//   - A0-1.4's store-side half — a uniqueness violation at insert surfaces as
//     internal and is never silently retried — belongs to the store package
//     (WP-22). ids supplies the two properties that make it checkable from
//     here: the errs.Internal classification of a generation failure, and the
//     single-attempt rule. TestUniquenessViolationIsInternal asserts those.
//   - A0-1.9, TestIDOrderingMatchesByteOrderCollateC, is a PostgreSQL
//     integration test (COLLATE "C" declared on the column) and is out of
//     scope for a package with no database; per DESIGN §8 it belongs to
//     store/postgres, where it is opt-in. It is deliberately absent here
//     rather than present and skipped.
//
// # Rulings made for this contract
//
//   - Kind values are the prefixes themselves ("eng_", "slp_node_", ...), per
//     the A0 §4 sketch, so no second prefix table can drift out of step with
//     the constants (A0-1.2, A0-1.10). The closedness check is one unexported
//     switch used by both New and Valid, so an unregistered Kind can neither
//     mint nor validate an id.
//   - The constant for the user prefix is KindUser, the sketch's name, kept
//     verbatim even though it stutters next to Tool (DESIGN §9) — A0 §4 says
//     the sketch's names are the source of truth for this package's surface.
//   - New classifies an unknown Kind as errs.Internal, not errs.Validation
//     (A0-3.1): a Kind is a compile-time constant of this package, never
//     caller-supplied material, so reaching that branch is a platform defect.
//     Validation stays a boundary concern and lives with the decoders — and
//     New's doc comment puts the obligation there explicitly, because Kind is
//     an open string type and nothing else stops a future decoder forwarding
//     client text. That message also echoes the offending value bounded
//     (%.16q; the longest real prefix is 9 bytes), so the same mistake cannot
//     put an arbitrary client string into a log record.
//   - A timestamp outside the 48-bit window keeps its low 48 bits (see
//     newID). A0-1.1 fixes the width and says nothing about the range; the
//     encoding stays well-formed under truncation, which is all the clause
//     requires, and rejecting 1970-era clocks would add a failure mode no
//     clause asks for.
//   - Valid is a length check, a prefix comparison and a per-byte scan of
//     Alphabet rather than a compiled regexp: the same language as the A0-1.2
//     regex with less machinery (DESIGN §2). A package-level regexp value would
//     be immutable, so DESIGN §4's rule against mutable state does not decide
//     it either way.
//   - A generation failure reports errs.OpOf == "ids.newID", the unexported
//     function that actually attempted the entropy read, because ADR-0019 §1
//     captures the originating function and New is a one-line delegator to it.
package ids
