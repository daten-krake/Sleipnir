// Package paging is the platform's one list envelope and its cursors (A0-4):
// every /api/v1 collection serializes {"items":[…],"next_cursor":"…"} and every
// one of them reads its limit parameter through this package, so "no offset
// paging, no total count, no silent clamp" is true in one place rather than
// per endpoint.
//
// Layer: foundation (DESIGN §1). A0 §4's preamble says the six foundation
// packages import no internal package except errs; that sentence governs the
// other five, because §4's own paging sketch declares ids.Kind in
// DecodeCursor's signature and cjson.CanonicalValue in EncodeCursor's. So this
// package imports exactly three: internal/errs, internal/ids and internal/cjson
// — no store-shaped, http-shaped or clock dependency, and nothing else (proved
// mechanically by `go list -deps ./internal/paging`).
//
// The exported surface is exactly A0 §4's sketch plus what A0-4.5 and A0-4.6
// need and nothing more: DefaultLimit, MaxLimit, Page, Cursor, EncodeCursor,
// DecodeCursor, ParseLimit, NewPage. There is no Page type per collection, no
// cursor option struct, no offset path (A0-4.1) and no total-count field
// (A0-4.1).
//
// # Clauses implemented here
//
//   - A0-4.1 — Page: one envelope, items plus an optional next_cursor. No
//     offset, no total count.
//   - A0-4.2 — NextCursor's omitempty: absent exactly when the collection is
//     exhausted (A0-8.3), never an empty string in the JSON.
//   - A0-4.4 — Cursor, EncodeCursor and DecodeCursor: base64url-unpadded of the
//     canonical JSON {"id":…,"k":…}, "id" first because canonical key order
//     decides it (A0-2.4), never hand-sorted; no MAC applied; ids validated
//     against the collection's kind (A0-1.5).
//   - A0-4.5 — DefaultLimit, MaxLimit, ParseLimit and NewPage's limit check,
//     hard-rejecting and never clamping (Decided, product owner 2026-09-21).
//   - A0-4.6 — NewPage: the caller reads limit+1 rows; row limit+1's existence
//     is the only has-more signal, so an exactly-full last page carries no
//     cursor.
//   - A0-4.8 — every cursor deviation above is errs.Validation.
//   - A0-8.6 — base64url without padding, standard alphabet and "="
//     padding rejected by using base64.RawURLEncoding and nothing else. (A0-4.4
//     cited this rule as A0-8.5, which is the closed-enum clause; ruled an
//     erratum 2026-09-29, A0 §6 item 20 — A0-4.4 now reads A0-8.6.)
//   - A0-2.14's items half — NewPage never returns nil Items, so an exhausted
//     page serializes as [] and cjson never sees a null.
//
// # Clauses deliberately NOT implemented here (who owns them)
//
//   - A0-4.3's ordering declaration (which immutable key orders which
//     collection, and the COLLATE "C" rule) and A0-4.7's watermark semantics:
//     the owning collection contract (A1 event log, A2/A4 graph) and the store
//     query that seeks to the cursor. This package supplies the tuple, not the
//     order.
//   - A0-4.4's second paragraph: deriving the seek position from the row the
//     cursor's id resolves to, ignoring k for seeking, and rejecting when k
//     disagrees with that row's ordering value (A0-4.8). That needs a
//     collection and a row lookup, which a foundation package does not have, so
//     it belongs to whoever resolves a cursor against a collection (WP-19,
//     WP-20, WP-22). DecodeCursor's only detectable inconsistency is the
//     id-vs-kind one its own signature makes possible. Ruled by the product
//     owner 2026-09-29 and recorded as A0 §6 item 22, which also amends §4.1's
//     naming rulings: the decode-time half — the cursor's id is not of the
//     declared ids.Kind — is WP-08 and is what
//     TestCursorWithInconsistentKAndIDRejected covers here; the row-resolution
//     half of A0-4.8 belongs to WP-19/WP-20 and is exercised against
//     PostgreSQL at WP-22. §4.1 registers that id under A0-4.4 as the ONLY
//     oracle for the whole clause across A0, A1 and A2, so it now says the two
//     halves apart; without that a traceability audit reads the row-resolution
//     half as covered here.
//   - A0-4.4's "authorization is re-derived per page": the api handler's scope
//     check (A0-3.9, internal/policy). A cursor is a position hint, not a
//     grant, and this package grants nothing.
//   - A0-4.8's "restart from the first page" wording and the 400 mapping: the
//     errs envelope (A0-3, internal/errs). A0-4.8's "cursors MUST NOT be
//     cached across releases": the caching layer; nothing here caches.
//   - A0-2's canonical form itself: internal/cjson. EncodeCursor does not
//     implement key order or integer bounds; it hands the struct to
//     cjson.CanonicalValue, so A0-2.6's integer range is what bounds Cursor.K
//     and no pre-check of that range is added here (A0-7.2: one definition of a
//     value).
//
// # Rulings
//
//   - Never clamp, in either direction. ParseLimit rejects a present limit
//     outside [1, MaxLimit] and NewPage rejects the same range a second time,
//     because a caller that clamped its own read to MaxLimit would otherwise
//     silently truncate: A0-4.5's reasoning ("an agent believes it saw the
//     whole collection") applies to the row slice too, not only to the query
//     parameter. On rejection both return the zero value alongside a non-nil
//     error, so no clamped number can escape.
//   - "?limit=" is present-and-empty and is rejected, not defaulted: only an
//     absent parameter reaches DefaultLimit (A0-4.5, A0-8.4 — "" is a value,
//     not "unset").
//   - A leading '+' or '-' is rejected before strconv.Atoi even though Atoi
//     accepts "+1": A0-4.5 says a non-integer limit is rejected, and a sign is
//     not part of the digits a limit query parameter spells. Leading zeros are
//     NOT rejected ("0007" is 7): A0-4.5's four failure classes are
//     unparseable, ≤ 0, non-integer and > 1000, and a leading zero is none of
//     them. A0-2.6's canonical-number rule does forbid them, but a query
//     parameter is not a canonical document. The principal may still want the
//     stricter reading; it is one clause away, not a silent choice here.
//   - A0 §4's example is internally inconsistent and this package follows the
//     bytes rather than the prose. The published next_cursor literal,
//     "eyJpZCI6ImV2dF8wMW0xeTJ3aGZocDE3ZzBhdmRxenRkMnAzIiwiayI6NDcxMX0" (63
//     base64url characters, 47 decoded bytes), is the canonical object
//     {"id":"evt_01m1y2whfhp17g0avdqztd2p3","k":4711} — an id body of 25
//     characters — while the same example's items[0].event_id,
//     "evt_01m1y2whfhp17g0avdqztd2p3x", has the 26-character body A0-1.1
//     requires. So the published cursor of the frozen contract is not a valid
//     A0-1.2 id and DecodeCursor must reject it. EncodeCursor reproduces the
//     published literal byte-for-byte for the payload it does encode, and
//     TestCursorRoundTripIsCanonicalBase64URL pins both halves — the byte-exact
//     encoding and the rejection — because it is the contract's own bytes; the
//     fixture is NOT silently swapped to the corrected one. Ruled an erratum
//     2026-09-29, A0 §6 item 19: the example's next_cursor now reads
//     "eyJpZCI6ImV2dF8wMW0xeTJ3aGZocDE3ZzBhdmRxenRkMnAzeCIsImsiOjQ3MTF9" (64
//     characters, 48 decoded bytes, the 26-character body), so the literal this
//     package pins is the corrected one and the 25-character form survives only
//     as the rejection subtest.
//   - A0-4.4's own text was the source of this package's mis-citation, so it
//     was an erratum and not a code defect: A0-4.4 wrote
//     "base64url-unpadded (A0-8.5)", but A0-8.5 is the closed-enum snake_case
//     rule; the base64url-without-padding rule is A0-8.6. Ruled 2026-09-29,
//     A0 §6 item 20; the parenthetical now reads A0-8.6 and this package's five
//     inherited citations were corrected with it.
//   - DecodeCursor compares the re-encoded cursor with the input string as its
//     last step. cjson.Canonical normalizes key order and whitespace rather
//     than rejecting them, and encoding/json matches struct keys
//     case-insensitively, so a payload of {"k":…,"id":…} or {"ID":…,"k":…}
//     would otherwise decode as a valid cursor for a document A0-4.4 does not
//     describe. One comparison covers all of it and is what makes A0-4.8's
//     "non-canonical JSON" and A0-6.2's "verify observed key spelling, not
//     DisallowUnknownFields alone" true here. It also carries a base64 property
//     the stdlib does not: Go 1.27's RawURLEncoding is LENIENT about the final
//     quantum's unused low bits, so for a payload whose length is 1 mod 3 there
//     are 16 final-character spellings that all decode to byte-identical bytes
//     (4 ways when the length is 2 mod 3). Fifteen of them are not the canonical
//     encoding A0-4.4's "replay byte-for-byte" requires, and only this
//     comparison rejects them. TestCursorRoundTripIsCanonicalBase64URL walks the
//     whole equivalence class and fails if its size ever changes.
//   - stdlib decode and json errors on the cursor path are rendered into an
//     errs.Validation message rather than wrapped. errs.Wrap inherits the
//     cause's kind, and a foreign error's kind is Internal (A0-3.3): wrapping a
//     base64 or json failure would turn malformed client input into a 500. Only
//     the failure the platform itself decides is created with errs.Newf.
//   - A rejected cursor's bytes never appear in the error message (A0-3.5):
//     the message names the clause, the offending id (bounded, %.64q — ids are
//     not secrets, A0-1.7, and A0-6.2 requires naming the offending value) and
//     the collection kind (bounded, %.16q, as internal/ids spells a Kind), but
//     not the payload. TestDecodeCursorRejects pins it.
//   - NewPage's limit is an int parameter, not re-read from a query string, so
//     a caller that already parsed with ParseLimit still cannot smuggle a
//     larger limit past the row-slice truncation.
//   - A nil cursorOf with more rows present is errs.Internal, not a panic and
//     not a silent "no more rows": A0-4.6's contract test would otherwise pass
//     an exhausted-looking page for a collection that has one. It is a wiring
//     defect (ADR-0019 §6 forbids panics in a request path; DESIGN §4 forbids
//     half-built inputs), which is exactly what Internal means.
//   - DefaultLimit and MaxLimit live here. A0-4.5 has fixed both numbers since
//     the freeze but A0-7.1's registry carried no row for them, while A0-7.2
//     forbids a second definition of a registry value anywhere — so this was
//     their only Go home. Ruled an erratum 2026-09-29, A0 §6 item 21: the
//     registry now carries DefaultPageLimit = 100 and MaxPageLimit = 1000 with
//     mechanism "reject → validation", following the precedent of its only
//     other numeric range (ExitCodeMin/ExitCodeMax), A0-4.5 cites both names,
//     and internal/paging stays their single Go definition. Both stay exported
//     because a consumer outside this package is already planned: the walking-
//     skeleton plan (docs/2026-09-29-walking-skeleton-plan.md, decisions D10 and
//     D11) freezes an A4-min /api/v1 surface in which every list endpoint is
//     paged under A0-4 with a hard-capped limit, so the cap is read through
//     these names rather than restated per handler.
//   - A0-8.6's "Tests:" line named TestEncodeCursorRoundTrip, which §4.1's
//     table did not register. The six ids §4.1's table registers across its
//     A0-4.4, A0-4.5, A0-4.6 and A0-8.6 rows
//     (TestDecodeCursorRejects, TestCursorWithInconsistentKAndIDRejected,
//     TestLimitValidation, TestLimitAboveMaxRejectedNotClamped,
//     TestHasMoreDetection, TestCursorRejectsStandardAlphabetAndPadding) are
//     the contract suite's names, so the round-trip check is a subtest of
//     TestCursorRoundTripIsCanonicalBase64URL rather than a seventh top-level
//     id here. That seventh name, TestCursorRoundTripIsCanonicalBase64URL, is
//     registered nowhere in A0: it comes from the WP-08 review row, and it is
//     what carries A0-8.6's clause-level TestEncodeCursorRoundTrip. Ruled
//     2026-09-29, A0 §6 item 20: one test, two names, A0-8.6 now says so and
//     the shipped id is the WP-08 one, which keeps §4.1 the single registry.
package paging
