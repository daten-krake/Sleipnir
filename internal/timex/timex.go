package timex

import (
	"regexp"
	"strconv"
	"time"

	"github.com/daten-krake/sleipnir/internal/errs"
)

// TimeLayout is the one timestamp layout of the platform (A0-5.1): RFC 3339 in
// UTC with a trailing "Z" and exactly three fractional digits, as in
// 2026-09-07T14:03:22.481Z. The ".000" directive forces the three digits
// (A0-5.2's millisecond precision) and the "Z07:00" directive prints "Z" only
// for a UTC value, which is why FormatTime converts to UTC before formatting.
//
// The layout is for formatting. Parsing MUST NOT rely on it alone: "Z07:00"
// also accepts numeric offsets such as "+02:00", which A0-5.3 rejects, so
// ParseTime checks the byte-exact A0-5.1 shape first.
const TimeLayout = "2006-01-02T15:04:05.000Z07:00"

// Parsing bounds of the accepted year window (A0-5.3). MinYear is inclusive,
// MaxYear is exclusive: a timestamp is only accepted for 2020 <= year < 2100.
const (
	MinYear = 2020
	MaxYear = 2100
)

// timestampRE is the byte-exact shape of A0-5.1, with the fields captured so
// the seconds group can be range-checked separately (A0-5.3 check 1). Go's \d
// is the ASCII class [0-9] — not Unicode digits — and the anchors plus the
// fixed-width groups are what make a numeric offset, a missing or over-long
// fraction, a lowercase "t"/"z", a space separator and any trailing byte fail.
var timestampRE = regexp.MustCompile(`^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2}):(\d{2})\.(\d{3})Z$`)

// Clock is the platform's single injected time source (A0-5.4; DESIGN §4, §8):
// anything that stamps time takes a Clock as a parameter instead of calling
// time.Now, so tests inject a deterministic value rather than sleeping.
type Clock interface {
	// Now returns the current instant. The reading may carry a monotonic
	// clock; Now strips it.
	Now() time.Time
}

// Now returns the clock's instant as the platform stores it: UTC, truncated to
// millisecond precision, monotonic reading stripped (A0-5.2, A0-5.4).
// Truncate is what strips the monotonic reading.
//
// A nil Clock is a wiring defect — the dependency was never injected — and is
// left to panic rather than be papered over with a time.Now fallback: a
// timestamp taken from an uninjectable clock cannot be reproduced by a test or
// a replay, which on an audit chain is worse than the crash. This is not the
// request-path panic ADR-0019 §6 forbids: no client input can make a Clock
// nil, the value arrives through a constructor, and a nil one means a caller
// struct was handed out half-built, which DESIGN §4 forbids outright. The
// panic is a deliberate fail-loud at that defect, not error transport, and the
// fallback it refuses to install is what TestInjectedClock/nil_clock pins.
func Now(c Clock) time.Time {
	return c.Now().UTC().Truncate(time.Millisecond)
}

// FormatTime renders t in the one layout of the platform (A0-5.1): always
// exactly three fractional digits, always a trailing "Z". It converts to UTC
// first, so a value in any other zone can never emit an offset string
// (A0-5.5: no local time zones anywhere).
//
// Platform values arrive here already truncated by Now. When a caller passes
// finer precision, Go's ".000" directive drops the sub-millisecond digits
// (it truncates; it does not round, so 14:03:59.999999999 cannot carry into
// the next minute); the output stays A0-5.1-shaped either way. Test
// FormatTimeAlwaysThreeDigits pins that behaviour.
//
// No year window is applied here: [MinYear, MaxYear) is A0-5.3's parsing
// bound, not a rendering rule, and adding one would change the §4 sketch's
// signature. A host clock outside that window — or outside the years
// 1000-9999, where Go emits a five-character year — renders a string
// ParseTime then rejects; that asymmetry is a deployment fault rather than a
// formatting one, and A0-5.6 keeps platform-recorded time authoritative.
func FormatTime(t time.Time) string {
	return t.UTC().Format(TimeLayout)
}

// ParseTime parses a timestamp string accepted by the platform and rejects,
// never normalizes, anything else with errs.Validation (A0-5.3, in the
// clause's order):
//
//  1. the byte-exact A0-5.1 shape, with the seconds group additionally
//     range-checked 00-59 — this is what forces "Z", exactly three fractional
//     digits and a capital "T", and rejects a leap second (":60");
//  2. time.Parse in TimeLayout, for calendar validity (month 13, day 32);
//  3. the year window [MinYear, MaxYear);
//  4. the clause's round-trip requirement FormatTime(ParseTime(s)) == s,
//     byte-exact.
//
// time.Parse alone is not usable here — "Z07:00" accepts a numeric offset —
// and time.RFC3339Nano is forbidden outright (it drops trailing zeros, so a
// whole second would parse as "2026-09-07T14:03:22Z").
//
// Each rejection message names the rejected class and echoes the value with
// %q, so a newline or quote inside it cannot inject into a log record
// (ADR-0019 §2, A0-3.4). The value is bounded by the caller's field cap, so
// no length constant is defined here (A0-7 owns caps; a second definition of
// a cap value would be a defect per A0-7.2).
func ParseTime(s string) (time.Time, error) {
	m := timestampRE.FindStringSubmatch(s)
	if m == nil {
		return time.Time{}, errs.Newf(errs.Validation,
			"parse timestamp: value %q: not in A0-5.1 form, want "+
				"YYYY-MM-DDTHH:MM:SS.mmmZ (UTC, capital T, trailing Z, exactly three fractional digits)", s)
	}
	if sec, err := strconv.Atoi(m[6]); err != nil || sec > 59 {
		return time.Time{}, errs.Newf(errs.Validation,
			"parse timestamp: value %q: seconds field %q outside 00-59, leap seconds are rejected (A0-5.3)", s, m[6])
	}
	t, err := time.Parse(TimeLayout, s)
	if err != nil {
		// %v of a *time.ParseError re-embeds the raw value unquoted, so this
		// branch relies on check 1 having already restricted s to [0-9T:.Z-]
		// for its injection safety. Reordering the checks would silently
		// remove that guard; no test can cover it, because no hostile byte
		// reaches this line.
		return time.Time{}, errs.Newf(errs.Validation,
			"parse timestamp: value %q: calendar fields are not a valid date or time: %v", s, err)
	}
	if t.Year() < MinYear || t.Year() >= MaxYear {
		return time.Time{}, errs.Newf(errs.Validation,
			"parse timestamp: value %q: year %d outside the accepted window [%d, %d) (A0-5.3)", s, t.Year(), MinYear, MaxYear)
	}
	if got := FormatTime(t); got != s {
		return time.Time{}, errs.Newf(errs.Validation,
			"parse timestamp: value %q: does not round-trip, formatted back as %q (A0-5.3)", s, got)
	}
	return t, nil
}
