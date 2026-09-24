package timex

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/daten-krake/sleipnir/internal/errs"
)

// a0_5_1 is the byte-exact shape of A0-5.1, restated in the tests so an
// implementation that drifts from the contract is caught rather than
// self-certified by the package's own layout constant.
var a0_5_1 = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z$`)

func TestParseTimeRejects(t *testing.T) {
	// want is the rejection class each message must name.
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"two fractional digits", "2026-09-07T14:03:22.48Z", "A0-5.1 form"},
		{"four fractional digits", "2026-09-07T14:03:22.4814Z", "A0-5.1 form"},
		{"one fractional digit", "2026-09-07T14:03:22.4Z", "A0-5.1 form"},
		{"nine fractional digits", "2026-09-07T14:03:22.481000000Z", "A0-5.1 form"},
		{"no fraction", "2026-09-07T14:03:22Z", "A0-5.1 form"},
		{"numeric offset plus", "2026-09-07T14:03:22.481+02:00", "A0-5.1 form"},
		{"numeric offset minus", "2026-09-07T14:03:22.481-00:00", "A0-5.1 form"},
		{"lowercase t", "2026-09-07t14:03:22.481Z", "A0-5.1 form"},
		{"lowercase z", "2026-09-07T14:03:22.481z", "A0-5.1 form"},
		{"space separator", "2026-09-07 14:03:22.481Z", "A0-5.1 form"},
		{"short month and seconds", "2026-9-07T14:03:2.481Z", "A0-5.1 form"},
		{"plus four year form", "+2026-09-07T14:03:22.481Z", "A0-5.1 form"},
		{"trailing space", "2026-09-07T14:03:22.481Z ", "A0-5.1 form"},
		{"leap second 60", "2026-09-07T14:03:60.481Z", "seconds field"},
		{"seconds 61", "2026-09-07T14:03:61.481Z", "seconds field"},
		{"seconds 99", "2026-09-07T14:03:99.481Z", "seconds field"},
		{"year below MinYear", "2019-12-31T23:59:59.999Z", "year"},
		{"year at MaxYear", "2100-01-01T00:00:00.000Z", "year"},
		{"year before 1970", "1969-12-31T23:59:59.000Z", "year"},
		{"year far future", "9999-01-01T00:00:00.000Z", "year"},
		{"month 13", "2026-13-07T14:03:22.481Z", "calendar fields"},
		{"day 32", "2026-09-32T14:03:22.481Z", "calendar fields"},
		{"february 30", "2026-02-30T00:00:00.000Z", "calendar fields"},
		{"february 29 non leap", "2026-02-29T00:00:00.000Z", "calendar fields"},
		{"empty string", "", "A0-5.1 form"},
		{"bare date", "2026-09-07", "A0-5.1 form"},
		{"bare time", "14:03:22.481Z", "A0-5.1 form"},
		{"trailing newline", "2026-09-07T14:03:22.481Z\n", "A0-5.1 form"},
		{"leading space", " 2026-09-07T14:03:22.481Z", "A0-5.1 form"},
		{"id shaped string", "evt_01m1x2whfhp17g0avdqztd2p3", "A0-5.1 form"},
		{"null byte inside", "2026-09-07T14:03:22.\x000Z", "A0-5.1 form"},
		{"unicode digit", "2026-09-07T14:03:2٢.481Z", "A0-5.1 form"},
		{"truncated by a cap", "2026-09-07T14:03:2", "A0-5.1 form"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseTime(tc.in)
			if err == nil {
				t.Fatalf("ParseTime(%q) = %s, want a rejection", tc.in, FormatTime(got))
			}
			if k := errs.KindOf(err); k != errs.Validation {
				t.Errorf("KindOf(%v) = %q, want %q (A0-5.3 is a hard reject)", err, k, errs.Validation)
			}
			msg := err.Error()
			if !strings.Contains(msg, tc.want) {
				t.Errorf("message %q does not name rejection class %q", msg, tc.want)
			}
			if !strings.Contains(msg, strconv.Quote(tc.in)) {
				t.Errorf("message %q does not echo %q", msg, tc.in)
			}
			if !got.IsZero() {
				t.Errorf("ParseTime(%q) returned a usable time %s, want the zero value (never normalize)", tc.in, FormatTime(got))
			}
		})
	}

	// A0-5.3 lists six rejection classes; the accept side of each, so a
	// guard that over-rejects cannot hide behind the table above.
	t.Run("accepts", func(t *testing.T) {
		tests := []struct {
			name string
			in   string
		}{
			{"A0-5.1 example", "2026-09-07T14:03:22.481Z"},
			{"whole second", "2026-09-07T14:03:22.000Z"},
			{"last millisecond of a minute", "2026-09-07T14:03:59.999Z"},
			{"MinYear boundary", "2020-01-01T00:00:00.000Z"},
			{"MaxYear minus one", "2099-12-31T23:59:59.999Z"},
			{"leap day", "2024-02-29T12:00:00.500Z"},
			{"post 2038", "2039-05-06T07:08:09.010Z"},
			{"first millisecond of MinYear", "2020-01-01T00:00:00.001Z"},
		}
		for _, tc := range tests {
			got, err := ParseTime(tc.in)
			if err != nil {
				t.Errorf("ParseTime(%q) = %v, want acceptance", tc.in, err)
				continue
			}
			if got.Location() != time.UTC {
				t.Errorf("ParseTime(%q) location = %v, want UTC", tc.in, got.Location())
			}
			if got.Nanosecond()%int(time.Millisecond) != 0 {
				t.Errorf("ParseTime(%q) keeps sub-millisecond precision %d ns (A0-5.2)", tc.in, got.Nanosecond())
			}
		}
	})

	// The clause's own requirement, named by A0 §4.1 as a subtest of this id.
	t.Run("round trip", func(t *testing.T) {
		for _, s := range []string{
			"2020-01-01T00:00:00.000Z",
			"2026-09-07T14:03:22.481Z",
			"2026-09-07T14:03:22.000Z",
			"2099-12-31T23:59:59.999Z",
		} {
			parsed, err := ParseTime(s)
			if err != nil {
				t.Fatalf("ParseTime(%q) = %v", s, err)
			}
			if got := FormatTime(parsed); got != s {
				t.Errorf("FormatTime(ParseTime(%q)) = %q, want the input byte-exact (A0-5.3)", s, got)
			}
		}
		// The other direction: a canonical platform value survives format →
		// parse unchanged, including the .000 that RFC3339Nano would drop.
		for _, want := range []time.Time{
			time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2026, 9, 7, 14, 3, 22, 481_000_000, time.UTC),
			time.Date(2099, 12, 31, 23, 59, 59, 999_000_000, time.UTC),
		} {
			s := FormatTime(want)
			got, err := ParseTime(s)
			if err != nil {
				t.Fatalf("ParseTime(%q) = %v", s, err)
			}
			if !got.Equal(want) {
				t.Errorf("ParseTime(FormatTime(%v)) = %v, want the same instant", want, got)
			}
			if got.Nanosecond() != want.Nanosecond() {
				t.Errorf("ParseTime(FormatTime(%v)) nanoseconds = %d, want %d", want, got.Nanosecond(), want.Nanosecond())
			}
		}
	})

	// A0-3.4 / A0-3.5: the echoed value is quoted, so hostile bytes cannot
	// inject a second log line or forge the message tail.
	t.Run("injected newline stays quoted", func(t *testing.T) {
		in := "2026-09-07T14:03:22.481Z\nrate_limited forgery"
		_, err := ParseTime(in)
		if err == nil {
			t.Fatal("ParseTime accepted a value with an injected newline")
		}
		if k := errs.KindOf(err); k != errs.Validation {
			t.Errorf("KindOf = %q, want %q", k, errs.Validation)
		}
		msg := err.Error()
		if strings.Contains(msg, "\n") {
			t.Errorf("message contains a raw newline, which lets the value inject a log line: %q", msg)
		}
		if !strings.Contains(msg, `\nrate_limited forgery`) {
			t.Errorf("message %q does not echo the value quoted with %%q", msg)
		}
	})

	t.Run("bypass via time.Parse alone is closed", func(t *testing.T) {
		// The two parsers the clause forbids as the implementation: prove the
		// normalizing behaviour exists in them and that ParseTime rejects the
		// same inputs (test that the bypass attempt fails, AGENTS.md).
		if _, err := time.Parse(TimeLayout, "2026-09-07T14:03:22.481+02:00"); err != nil {
			t.Fatalf("premise wrong: time.Parse accepted no offset: %v", err)
		}
		if _, err := ParseTime("2026-09-07T14:03:22.481+02:00"); errs.KindOf(err) != errs.Validation {
			t.Errorf("offset accepted by ParseTime: %v", err)
		}
		if dropped := FormatTime(time.Date(2026, 9, 7, 14, 3, 22, 0, time.UTC)); dropped != "2026-09-07T14:03:22.000Z" {
			t.Errorf("FormatTime of a whole second = %q, want the .000 RFC3339Nano would drop", dropped)
		}
		if !strings.HasSuffix(time.Date(2026, 9, 7, 14, 3, 22, 0, time.UTC).Format(time.RFC3339Nano), "22Z") {
			t.Fatal("premise wrong: RFC3339Nano no longer drops trailing zeros")
		}
	})
}

func TestFormatTimeAlwaysThreeDigits(t *testing.T) {
	tests := []struct {
		name string
		in   time.Time
		want string
	}{
		{"whole second", time.Date(2026, 9, 7, 14, 3, 22, 0, time.UTC), "2026-09-07T14:03:22.000Z"},
		{"one millisecond", time.Date(2026, 9, 7, 14, 3, 22, 1_000_000, time.UTC), "2026-09-07T14:03:22.001Z"},
		{"exact millisecond", time.Date(2026, 9, 7, 14, 3, 22, 481_000_000, time.UTC), "2026-09-07T14:03:22.481Z"},
		{"microsecond input", time.Date(2026, 9, 7, 14, 3, 22, 481_900_000, time.UTC), "2026-09-07T14:03:22.481Z"},
		{"nanosecond input", time.Date(2026, 9, 7, 14, 3, 22, 481_999_999, time.UTC), "2026-09-07T14:03:22.481Z"},
		{"max nanoseconds do not carry", time.Date(2026, 9, 7, 14, 3, 59, 999_999_999, time.UTC), "2026-09-07T14:03:59.999Z"},
		{"single nanosecond", time.Date(2026, 9, 7, 14, 3, 22, 1, time.UTC), "2026-09-07T14:03:22.000Z"},
		// FormatTime applies no year window; ParseTime's [2020, 2100) does.
		{"pre 1970", time.Date(1969, 12, 31, 23, 59, 59, 500_000_000, time.UTC), "1969-12-31T23:59:59.500Z"},
		{"post 2038", time.Date(2039, 5, 6, 7, 8, 9, 10_000_000, time.UTC), "2039-05-06T07:08:09.010Z"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := FormatTime(tc.in)
			if got != tc.want {
				t.Errorf("FormatTime(%v) = %q, want %q", tc.in, got, tc.want)
			}
			if !a0_5_1.MatchString(got) {
				t.Errorf("FormatTime(%v) = %q, does not match the A0-5.1 shape", tc.in, got)
			}
			if !strings.HasSuffix(got, "Z") {
				t.Errorf("FormatTime(%v) = %q, want a trailing Z (A0-5.5)", tc.in, got)
			}
		})
	}

	t.Run("non UTC input yields the same instant in Z", func(t *testing.T) {
		utc := time.Date(2026, 9, 7, 14, 3, 22, 481_000_000, time.UTC)
		offset := utc.In(time.FixedZone("CEST", 2*60*60))
		got := FormatTime(offset)
		if got != "2026-09-07T14:03:22.481Z" {
			t.Errorf("FormatTime(%v) = %q, want the UTC rendering, never an offset", offset, got)
		}
		back, err := time.Parse(time.RFC3339Nano, got)
		if err != nil {
			t.Fatalf("premise wrong: %v", err)
		}
		if !back.Equal(offset) {
			t.Errorf("FormatTime changed the instant: %v -> %v", offset, back)
		}
	})
}

func TestNowStripsMonotonic(t *testing.T) {
	// A clock that hands back what the platform would really read: time.Now()
	// carries a monotonic reading.
	// Captured once so the assertions below are on a fixed instant: a clock
	// read twice can straddle a millisecond boundary.
	injected := realtimeClock{}.Now()
	got := Now(&fixedClock{at: injected})

	if got.Round(0) != got {
		t.Errorf("Now kept a monotonic reading: round-trip %v differs from %v (A0-5.4)", got.Round(0), got)
	}
	if got.Location() != time.UTC {
		t.Errorf("Now location = %v, want UTC", got.Location())
	}
	if got.Nanosecond()%int(time.Millisecond) != 0 {
		t.Errorf("Now nanoseconds = %d, want a whole millisecond (A0-5.2)", got.Nanosecond())
	}
	if want := FormatTime(injected.UTC().Truncate(time.Millisecond)); FormatTime(got) != want {
		t.Errorf("FormatTime(Now(c)) = %q, want %q", FormatTime(got), want)
	}
	if !got.Equal(injected.Truncate(time.Millisecond)) {
		t.Errorf("Now(c) = %v, want the injected instant truncated to milliseconds (%v)", got, injected.Truncate(time.Millisecond))
	}
}

func TestInjectedClock(t *testing.T) {
	t.Run("fake instant is honoured exactly", func(t *testing.T) {
		fake := &fixedClock{at: time.Date(2021, 3, 4, 5, 6, 7, 891_000_000, time.UTC)}
		got := Now(fake)
		if want := fake.at; !got.Equal(want) {
			t.Errorf("Now = %v, want the injected %v", got, want)
		}
		if got.Format(TimeLayout) != "2021-03-04T05:06:07.891Z" {
			t.Errorf("Now formatted = %q, want the injected instant, never the real clock", got.Format(TimeLayout))
		}
		// The real clock is elsewhere; prove Now never read it.
		if !time.Now().Truncate(time.Millisecond).After(got.Add(time.Hour)) {
			t.Errorf("real clock is not ahead of the injected 2021 instant: test premise wrong")
		}
	})

	t.Run("sub-millisecond injected value is truncated", func(t *testing.T) {
		fake := &fixedClock{at: time.Date(2026, 9, 7, 14, 3, 22, 481_999_999, time.UTC)}
		if got, want := Now(fake).Nanosecond(), 481_000_000; got != want {
			t.Errorf("Now nanoseconds = %d, want %d (A0-5.4 truncate)", got, want)
		}
	})

	t.Run("stepped clock", func(t *testing.T) {
		c := &steppedClock{
			at:   time.Date(2026, 9, 7, 14, 3, 22, 481_000_000, time.UTC),
			step: 15 * time.Minute,
		}
		first, second := Now(c), Now(c)
		if !first.Equal(c.inst(0)) {
			t.Errorf("first Now = %v, want %v", first, c.inst(0))
		}
		if !second.Equal(c.inst(1)) {
			t.Errorf("second Now = %v, want %v", second, c.inst(1))
		}
		if d := second.Sub(first); d != c.step {
			t.Errorf("delta between calls = %v, want the stepped %v", d, c.step)
		}
	})

	t.Run("non UTC fake is normalized", func(t *testing.T) {
		at := time.Date(2026, 9, 7, 16, 3, 22, 481_000_000, time.FixedZone("CEST", 2*60*60))
		got := Now(&fixedClock{at: at})
		if got.Location() != time.UTC {
			t.Errorf("Now location = %v, want UTC (A0-5.5)", got.Location())
		}
		if !got.Equal(at.Truncate(time.Millisecond)) {
			t.Errorf("Now = %v, want the same instant as %v", got, at)
		}
	})

	// Ruling (doc.go): a nil Clock is a wiring defect and panics naturally
	// rather than being replaced by the real clock. Asserted, not relied on at
	// runtime: the clock is wired once at the cmd/ edge (DESIGN §7).
	t.Run("nil clock", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Error("Now(nil) returned without panicking; a nil Clock must not silently read the real clock")
			}
		}()
		_ = Now(nil)
	})
}

// realtimeClock is the production shape: it hands back time.Now(), which
// carries a monotonic reading.
type realtimeClock struct{}

func (realtimeClock) Now() time.Time { return time.Now() }

// fixedClock returns one instant for every call.
type fixedClock struct {
	at time.Time
}

func (c *fixedClock) Now() time.Time { return c.at }

// steppedClock advances by a fixed step on every call; deterministic, no
// sleeping (DESIGN §8).
type steppedClock struct {
	at   time.Time
	step time.Duration
	n    int
}

// inst returns the instant the nth call yields.
func (c *steppedClock) inst(n int) time.Time {
	return c.at.Add(time.Duration(n) * c.step)
}

func (c *steppedClock) Now() time.Time {
	defer func() { c.n++ }()
	return c.inst(c.n)
}
