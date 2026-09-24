// Package timex is the platform's single place where time is formatted,
// parsed and read: the one layout, the one injected clock, the one precision,
// and reject-never-normalize parsing (A0-5).
//
// Layer: foundation (DESIGN §1). It imports only the standard library and
// internal/errs, the sole internal import A0 §4 allows a foundation package
// (A0 §4 preamble: `internal/ids`, `internal/cjson`, `internal/errs`,
// `internal/paging`, `internal/timex`, `internal/caps` import nothing from
// internal/ except errs; they never import each other). The name avoids
// shadowing the standard library's time package, for the reason DESIGN §1
// gives for internal/logging (`log`/`log/slog`).
//
// # Clauses implemented here
//
//   - A0-5.1 — TimeLayout and FormatTime: RFC 3339, UTC, trailing "Z", exactly
//     three fractional digits, matching ^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z$.
//   - A0-5.2 — one precision for the whole platform: milliseconds. It is
//     carried by the ".000" directive and by Now's Truncate(time.Millisecond);
//     there is no µs/ns path in this package.
//   - A0-5.3 — ParseTime: the clause's four checks in its order (byte-exact
//     regex with the seconds group range-checked 00-59, time.Parse for
//     calendar validity, the [MinYear, MaxYear) window, then the byte-exact
//     round-trip), each rejection an errs.Validation. time.Parse alone and
//     time.RFC3339Nano are both deliberately not used as the parser.
//   - A0-5.4, first half — Clock and Now: the one injected clock, UTC,
//     millisecond-truncated, monotonic reading stripped by Truncate before a
//     value is stored, hashed or serialized.
//
// # Clauses deliberately not implemented here
//
//   - A0-5.4's recorded_at forward clamp max(clock_now, prev_recorded_at), its
//     1000 ms bound and its error-level log with the engagement/run correlation
//     attributes, plus the TestRecordedAtClampIsMonotone and
//     TestRecordedAtClampBypassRejected ids: the clause assigns the clamp to
//     the event writer, the test ids belong to the event-chain work package
//     (WP-10), and the log needs internal/logging, which a foundation package
//     MUST NOT import (A0 §4 preamble, DESIGN §1). There is no Clamp helper and
//     no clamp constant here because the consumer does not exist yet —
//     DESIGN §2 forbids code for a hypothetical caller.
//   - A0-5.5's UI rule (render the zone explicitly) is an obligation on the
//     rendering layer. What this package does for it is the format half: no
//     output ever carries an offset, because FormatTime converts to UTC first.
//   - A0-5.6 (platform-recorded time is authoritative, no external anchoring),
//     A0-5.7 (a client-supplied timestamp lives in a *_claimed_at field, never
//     drives ordering/expiry/digest) and A0-5.8 (whoever controls the host
//     controls recorded time) are rules about how other packages use these
//     values. They exist because this package is the only clock and the only
//     parser; nothing here enforces them.
//
// # Rulings made for this package
//
//   - FormatTime converts to UTC before formatting rather than emitting the
//     input zone. "Z07:00" would print "+02:00" for a non-UTC value, which
//     violates A0-5.1's byte-exact shape and A0-5.5's ban on any offset other
//     than "Z" — so a caller that forgot to convert cannot produce a
//     non-conforming timestamp.
//   - Sub-millisecond input is truncated by Go's ".000" directive, not
//     rounded (verified: 14:03:22.481999999 formats as .481, and
//     14:03:59.999999999 does not carry into the next minute). Platform values
//     arrive already truncated via Now, so this is belt-and-braces; the
//     behaviour is pinned by a subtest of TestFormatTimeAlwaysThreeDigits
//     instead of being reimplemented.
//   - A nil Clock passed to Now is a caller bug and is left to panic
//     naturally; see the doc comment on Now for the ADR-0019 §6 reasoning, and
//     TestInjectedClock/nil_clock, which asserts the panic rather than
//     installing a fallback.
//   - A rejection message echoes the rejected value with %q (A0-3.4 permits
//     echoing untrusted request material that no secret-pattern rule rejects,
//     and a timestamp field is not one; A0-3.5 forbids a request body, and a
//     single field value is not a body). %q is what keeps an embedded newline
//     or quote from injecting into the log record. No length cap is defined
//     for the echo: A0-7 owns cap values and A0-7.2 makes a second definition
//     of one a defect — the caller's field cap bounds the input.
//   - ParseTime checks the seconds group (`if sec > 59`) separately from the
//     regex, as A0-5.3 words it, so the regex stays byte-identical to the one
//     printed in the clause.
package timex
