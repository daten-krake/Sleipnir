package secretscan

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/daten-krake/sleipnir/internal/errs"
)

// ruleTableIDs is A2-9.4's rule-id column transcribed by hand from
// contracts/A2-graph.md, in table order. It is the authority the implementation
// is compared against: nothing here is derived from rules, so an added,
// renamed, dropped or reordered rule fails instead of agreeing with itself.
var ruleTableIDs = []string{
	"SEC-PEM",
	"SEC-NTLM",
	"SEC-KRB",
	"SEC-AWSKEY",
	"SEC-GCPKEY",
	"SEC-AZUREKEY",
	"SEC-JWT",
	"SEC-BEARER",
	"SEC-URLCRED",
	"SEC-ENTROPY",
}

// ruleTablePatterns is A2-9.4's Go regexp COLUMN for the nine rows that have
// one, transcribed by hand from contracts/A2-graph.md with the markdown-escaped
// alternation pipes (\|) un-escaped to the | the contract means. It exists
// because comparing ids alone pins nothing about the patterns: a NARROWING edit
// to any row still matches this package's planted corpus, so the whole suite
// stays green while real secrets start scanning clean.
var ruleTablePatterns = []struct{ ruleID, pattern string }{
	{"SEC-PEM", "-----BEGIN [A-Z ]*PRIVATE KEY-----"},
	{"SEC-NTLM", `(?i)\b[0-9a-f]{32}\b`},
	{"SEC-KRB", `(?i)krbtgt[/@][A-Za-z0-9._-]{1,128}`},
	{"SEC-AWSKEY", `(AKIA|ASIA)[0-9A-Z]{16}`},
	{"SEC-GCPKEY", `AIza[0-9A-Za-z\-_]{35}`},
	{"SEC-AZUREKEY", `(?i)(AccountKey|SharedAccessKey|sig)=[A-Za-z0-9+/=]{20,}`},
	{"SEC-JWT", `eyJ[0-9A-Za-z_-]+\.[0-9A-Za-z_-]+\.[0-9A-Za-z_-]+`},
	{"SEC-BEARER", `(?i)(bearer|token|api[_-]?key|password|passwd|secret)\s*[:=]\s*\S{8,}`},
	{"SEC-URLCRED", `[a-z][a-z0-9+.-]*://[^/\s:@]{1,64}:[^/\s:@]{1,64}@`},
}

// patternFor returns the implementation's compiled pattern for rule id, or nil
// if the id has no row or its row is the non-regexp SEC-ENTROPY row. Tests use
// it so they assert against the shipped table instead of a local copy.
func patternFor(id string) *regexp.Regexp {
	for _, r := range rules {
		if r.id == id {
			return r.pattern
		}
	}
	return nil
}

// secretCorpus is exactly one planted value per A2-9.4 rule id.
//
// A2-9.4 says this corpus "lives in the shared suite" (it is the corpus of
// TestEventSecretFreeSerialization and TestGraphSecretFreeSerialization).
// This is a LOCAL PROVISIONAL COPY owned by WP-13; WP-20/WP-21 must own the
// canonical one. Each value must be matched by its OWN rule, which — since the
// first match in table order wins — means it must not be matched by any earlier
// row. Values are chosen accordingly (no 32-hex bounded run before SEC-NTLM, no
// "eyJ" before SEC-JWT, and so on).
var secretCorpus = []struct {
	ruleID string
	field  string
	value  string
}{
	{
		ruleID: "SEC-PEM",
		field:  "node.summary",
		value:  "-----BEGIN RSA PRIVATE KEY-----\nMIIBOgIBAAJBAKj34GkxFhDDA7pJF7wFA8kT9w7Vf1lK\n-----END RSA PRIVATE KEY-----\n",
	},
	{
		ruleID: "SEC-NTLM",
		field:  "node.claim",
		value:  "captured LM response 209c617e7aae58d49a8d3f5c4b6a7e8d for user",
	},
	{
		ruleID: "SEC-KRB",
		field:  "node.basis",
		value:  "kerberoast target krbtgt/CORP.LOCAL@CORP.LOCAL",
	},
	{
		ruleID: "SEC-AWSKEY",
		field:  "edge.attrs",
		value:  "aws access key id AKIAIOSFODNN7EXAMPLE found in bundle",
	},
	{
		ruleID: "SEC-GCPKEY",
		field:  "node.label",
		value:  "AIzaSyA1234567890abcdefghijklmnopqrstuvwx",
	},
	{
		ruleID: "SEC-AZUREKEY",
		field:  "node.attrs",
		value:  "DefaultEndpointsProtocol=https;AccountKey=bbNpZx0eqvh1rQ3wLm6yPsXkJdGcFhTnRbUeZo=qA",
	},
	{
		ruleID: "SEC-JWT",
		field:  "payload.summary",
		value:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
	},
	{
		ruleID: "SEC-BEARER",
		field:  "attrs.transport",
		value:  "config dump: api_key=ukRyPmZsTgQxVbNmKpLwZcXeRtYuIo",
	},
	{
		ruleID: "SEC-URLCRED",
		field:  "edge.protocol",
		value:  "https://webhook:HookSk9sE0dZmvd2024xQ@ci.example.internal/push",
	},
	{
		ruleID: "SEC-ENTROPY",
		field:  "node.domain",
		value:  "Zx7Kq0WrT2sPbvLdNmH8cYf4gJaRoUeQIkX5VbN0mM",
	},
}

// TestEveryRuleIDMatchesItsCorpusValue is A2-9.4's TestEveryRuleIDMatchesItsCorpusValue.
func TestEveryRuleIDMatchesItsCorpusValue(t *testing.T) {
	t.Run("corpus_value_reports_its_own_rule_id", func(t *testing.T) {
		for _, c := range secretCorpus {
			t.Run(c.ruleID, func(t *testing.T) {
				err := Scan(c.field, c.value)
				if err == nil {
					t.Fatalf("Scan(%q, planted %s value) = nil, want a validation rejection", c.field, c.ruleID)
				}
				if got := errs.KindOf(err); got != errs.Validation {
					t.Errorf("kind = %q, want %q (A2-9.4 rejects with validation)", got, errs.Validation)
				}
				if id := reportedRuleID(t, err); id != c.ruleID {
					t.Errorf("reported rule id = %q, want %q (full message %q)", id, c.ruleID, err.Error())
				}
				// What would make this assertion pass while breaking the rule:
				// deleting the SEC-AWSKEY row from rules, or reordering the
				// slice so an earlier row claims the value — the first-match
				// order is part of the contract (ruling S1).
			})
		}
	})

	t.Run("reportable_rule_id_set_is_exactly_the_table", func(t *testing.T) {
		reportable := map[string]bool{}
		for _, c := range secretCorpus {
			err := Scan(c.field, c.value)
			if err == nil {
				t.Fatalf("planted %s value scanned clean", c.ruleID)
			}
			reportable[reportedRuleID(t, err)] = true
		}
		assertIDSet(t, ruleTableIDs, reportable, "ids Scan can report")

		// The implementation's own table, compared against the transcription.
		if len(rules) != len(ruleTableIDs) {
			t.Errorf("rule count = %d, want %d (A2-9.4 is closed and additive-only)", len(rules), len(ruleTableIDs))
		}
		for i, want := range ruleTableIDs {
			if i >= len(rules) {
				t.Fatalf("rule %d (%s) is missing from the implementation (A2-9.4)", i, want)
			}
			if rules[i].id != want {
				t.Errorf("rule %d = %q, want %q (table order is the report order, ruling S1)", i, rules[i].id, want)
			}
			if (rules[i].pattern == nil) != (want == "SEC-ENTROPY") {
				t.Errorf("rule %s: the regexp column disagrees with A2-9.4, where only SEC-ENTROPY is not a regexp", want)
			}
		}
		// The regexp column, row by row. What would make this fail while every
		// other assertion in this file still passes: narrowing a quantifier —
		// SEC-KRB's {1,128} to {1,8} or SEC-AZUREKEY's {20,} to {40,} both keep
		// the planted corpus matching (their corpus values are longer than either
		// bound) while dropping real ticket material and real 20–39 character SAS
		// signatures. Nothing else in the suite can see that; this can.
		if len(ruleTablePatterns) != 9 {
			t.Fatalf("transcribed %d patterns, want the 9 regexp rows of A2-9.4", len(ruleTablePatterns))
		}
		for _, want := range ruleTablePatterns {
			got := patternFor(want.ruleID)
			if got == nil {
				t.Fatalf("rule %s has no compiled pattern, but A2-9.4 gives it one", want.ruleID)
			}
			if got.String() != want.pattern {
				t.Errorf("rule %s pattern = %q, want %q (A2-9.4's regexp column)", want.ruleID, got.String(), want.pattern)
			}
		}
		// Nine regexp rows plus the one non-regexp row is the whole table: an
		// added tenth regexp would trip this and the id-count check above.
		if len(rules) != len(ruleTablePatterns)+1 {
			t.Errorf("%d rules for %d transcribed patterns, want one more rule than patterns (SEC-ENTROPY is not a regexp)", len(rules), len(ruleTablePatterns))
		}

		// What would make this assertion fail: adding an eleventh row (the count
		// and the set check trip), renaming or reordering a row (the order check
		// trips), or moving a compiled pattern into the SEC-ENTROPY row and vice
		// versa (the regexp-column check trips). It cannot pass vacuously:
		// nothing here derives ruleTableIDs from the implementation.
	})
}

func assertIDSet(t *testing.T, want []string, got map[string]bool, what string) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("%s = %d entries, want %d: %v", what, len(got), len(want), keys(got))
	}
	for _, id := range want {
		if !got[id] {
			t.Errorf("%s is missing %q (A2-9.4 closed table)", what, id)
		}
	}
	for id := range got {
		if !slices.Contains(want, id) {
			t.Errorf("%s reports %q, which is not in A2-9.4's table (ruling S2)", what, id)
		}
	}
}

func keys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	slices.Sort(out) // map order is random; a failure message must not be
	return out
}

// leakReport returns every labelled form of the value that appears in text.
// It is the detector TestErrorMessageNamesFieldAndRuleIDOnly runs, factored out
// so one subtest can prove it is not vacuous.
func leakReport(text string, forms map[string]string) map[string]string {
	found := map[string]string{}
	for label, form := range forms {
		if form != "" && strings.Contains(text, form) {
			found[label] = form
		}
	}
	return found
}

// reportedRuleID extracts the rule id segment of the A2-9.4/A2-9.5 message. It
// is a test helper only: the message is prose and MUST NOT be parsed by
// production code (A0-3.4).
func reportedRuleID(t *testing.T, err error) string {
	t.Helper()
	const prefix = "rule "
	msg := err.Error()
	i := strings.Index(msg, prefix)
	if i < 0 {
		t.Fatalf("message %q has no %q segment (A2-9.4 requires the rule id)", msg, prefix)
	}
	rest := msg[i+len(prefix):]
	// The id runs to the first space or colon: the shipped shape is
	// "rule <ID> matched: <n> bytes (A2-9.4)".
	j := strings.IndexAny(rest, " :")
	if j < 0 {
		t.Fatalf("message %q has an unterminated rule id segment", msg)
	}
	return rest[:j]
}

// benignCorpus is A2-9.4's no-false-positive set: values that are ordinary
// platform content and MUST NOT be rejected.
var benignCorpus = []string{
	"platform.internal.example.com",
	"dc01.corp.local",
	"10.0.0.0/8",
	"192.168.12.0/24",
	"2001:db8::/32",
	"nmap -sV -p- --script vuln 10.10.10.10",
	"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
	"sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
	"evt_01m1y2whfhp17g0avdqztd2p3x",
	"gn_01m1y2whfhh039ykj5x8mc5a0g",
	"2026-09-25T14:03:22Z",
	"The operator approved the scoped run and the worker uploaded the evidence artifact.",
	"https://ci.example.internal/push",
	"HTTP/2",
	"",
}

// TestNoFalsePositiveOnBenignCorpus is A2-9.4's TestNoFalsePositiveOnBenignCorpus.
func TestNoFalsePositiveOnBenignCorpus(t *testing.T) {
	for _, v := range benignCorpus {
		t.Run(benignCaseName(v), func(t *testing.T) {
			if err := Scan("node.label", v); err != nil {
				t.Errorf("Scan benigned value %q = %q, want nil", v, err)
			}
		})
	}

	t.Run("hex_digest_survives_sec_ntlm_because_no_word_boundary_at_32", func(t *testing.T) {
		// SEC-NTLM is `(?i)\b[0-9a-f]{32}\b`: a 32-hex run delimited by WORD
		// boundaries. Inside a 64-hex-character run there is no word boundary
		// at offset 32 (both neighbours are word characters), so the only
		// boundaries are at 0 and 64 and no 32-character window can be
		// delimited. The 64-hex artifact digest therefore survives, and it
		// also survives SEC-ENTROPY because a 16-symbol alphabet carries at
		// most log2(16) = 4.0 bits/char.
		// The implementation's own compiled pattern, not a local transcription:
		// a local regexp would test Go's regex engine here, and an edit to
		// rules[1] (say dropping the \b anchors) could not fail these two
		// assertions. Matching the shipped pattern makes both of them live.
		ntlm := patternFor("SEC-NTLM")
		if ntlm == nil {
			t.Fatal("SEC-NTLM has no compiled pattern in the implementation (A2-9.4 gives it one)")
		}
		const digest = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
		if ntlm.MatchString(digest) {
			t.Errorf("SEC-NTLM matched a 64-hex artifact digest; the word-boundary reasoning in this test is wrong")
		}
		if !ntlm.MatchString(digest[:32]) {
			t.Errorf("SEC-NTLM did not match the same value truncated to 32 hex characters, so the boundary argument above is not what is protecting the digest")
		}
		if err := Scan("node.label", digest); err != nil {
			t.Errorf("64-hex digest rejected: %q", err)
		}
		if err := Scan("node.attrs.image_digest", "sha256:"+digest); err != nil {
			t.Errorf("registry image digest rejected: %q (A0-8.7 shape)", err)
		}
	})
}

func sha256Bytes(b []byte) []byte { sum := sha256.Sum256(b); return sum[:] }

func sha256Hex(b []byte) string { return hex.EncodeToString(sha256Bytes(b)) }

// benignCaseName renders a benign corpus value as a readable subtest name.
func benignCaseName(s string) string {
	if s == "" {
		return "empty"
	}
	if len(s) > 24 {
		return fmt.Sprintf("%d_bytes", len(s))
	}
	return strings.NewReplacer(" ", "_", "/", "-", ".", "-").Replace(s)
}

// TestErrorMessageNamesFieldAndRuleIDOnly is A2-9.5's / A0-3.4's
// TestErrorMessageNamesFieldAndRuleIDOnly.
func TestErrorMessageNamesFieldAndRuleIDOnly(t *testing.T) {
	t.Run("exact_shape", func(t *testing.T) {
		err := Scan("node.summary", "AKIAIOSFODNN7EXAMPLE")
		want := "secretscan.Scan: scanning field node.summary for secret material: rule SEC-AWSKEY matched: 20 bytes (A2-9.4)"
		if got := err.Error(); got != want {
			t.Errorf("message =\n\t%q\nwant\n\t%q", got, want)
		}
	})

	for _, c := range secretCorpus {
		t.Run(c.ruleID, func(t *testing.T) {
			err := Scan(c.field, c.value)
			if err == nil {
				t.Fatalf("planted %s value scanned clean", c.ruleID)
			}
			msg := err.Error()

			if !strings.Contains(msg, c.field) {
				t.Errorf("message %q does not name the field %q (A2-9.4)", msg, c.field)
			}
			if id := reportedRuleID(t, err); id != c.ruleID {
				t.Errorf("rule id = %q, want %q", id, c.ruleID)
			}
			if want := fmt.Sprintf(": %d bytes", len(c.value)); !strings.Contains(msg, want) {
				t.Errorf("message %q does not carry the byte length %q (A2-9.5 allows field, rule id and length only)", msg, want)
			}

			// The leak surface: whole value, prefix, suffix, hex, SHA-256 and
			// base64 forms. A2-9.5 forbids the value in whole, in part or as a
			// digest — in the message, in an error attrs entry, or in any slog
			// record.
			forms := leakForms(c.value)
			for _, surface := range surfaces(t, err, c.field, c.value) {
				if found := leakReport(surface.text, forms); len(found) > 0 {
					t.Errorf("%s leaks the rejected value as %v: %s", surface.name, found, surface.text)
				}
			}
			// What would make this assertion pass while breaking the rule:
			// changing the errs.Newf call to include %q of the value, or
			// passing value through errs.Redact into attrs — or, the sneaky
			// one, handing the value to a String-only redaction helper and
			// letting the slog JSON handler marshal it by reflection.
		})
	}

	t.Run("the_leak_detector_is_not_vacuous", func(t *testing.T) {
		// What would make every assertion in this Test function pass while
		// breaking the rule is a detector that never fires — an empty forms
		// map, or a substring check against a constant. This subtest feeds the
		// same detector one deliberately leaky message per form and requires
		// each to fire, so the clean results above are evidence and not an
		// artefact of a check that cannot fail.
		c := secretCorpus[0]
		forms := leakForms(c.value)
		for label, form := range forms {
			t.Run(label, func(t *testing.T) {
				leaky := "secretscan.Scan: scanning field " + c.field + " for secret material: rule SEC-PEM matched: " + form
				if _, ok := leakReport(leaky, forms)[label]; !ok {
					t.Errorf("detector did not flag %q in a message that embeds it verbatim", label)
				}
			})
		}
		if got := leakReport(Scan(c.field, c.value).Error(), forms); len(got) != 0 {
			t.Errorf("the detector flags the real message (%v), which contradicts the leaky cases above", got)
		}
	})

	t.Run("field_name_is_the_only_caller_text_echoed", func(t *testing.T) {
		// A field name is not a secret (A0-3.4 exception permits naming the
		// field); the value never is. Scan must not be the place that decides
		// otherwise.
		err := Scan("attrs.evidence_blob", "aad3b435b51404eeaad3b435b51404ee")
		if got := err.Error(); strings.Contains(got, "aad3b435") {
			t.Errorf("message carries a prefix of the rejected value: %q", got)
		}
	})
}

type surface struct {
	name string
	text string
}

func surfaces(t *testing.T, err error, field, value string) []surface {
	t.Helper()
	out := []surface{{"error message", err.Error()}}

	env := errs.NewEnvelope(err, errs.Attrs{EngagementID: "eng_01m1y2whfhgbz06ays6dxnvyws"}, 0)
	body, merr := json.Marshal(env)
	if merr != nil {
		t.Fatalf("marshaling the error envelope: %v", merr)
	}
	out = append(out, surface{"error envelope message", env.Error.Message})
	// The decoded message is grepped as its own surface, not folded into the
	// JSON body: the body carries the same prose ESCAPED (a value's newline is
	// two bytes there), so an escaped-only grep cannot see a newline-bearing
	// leak. Today errs sets Message from err.Error() verbatim, so this surface
	// also pins that the envelope performs no transform of its own.
	out = append(out, surface{"error envelope JSON", string(body)})

	var logbuf strings.Builder
	lg := slog.New(slog.NewJSONHandler(&logbuf, nil))
	lg.Error("graph write rejected",
		"component", "graph",
		"field", field,
		"secret", errs.NewSecret(value),
		"err", err,
	)
	out = append(out, surface{"slog JSON record", logbuf.String()})
	return out
}

func leakForms(value string) map[string]string {
	forms := map[string]string{"whole": value}
	if n := len(value); n > 0 {
		if n >= 8 {
			forms["first 8"] = value[:8]
			forms["last 8"] = value[n-8:]
			forms["middle 8"] = value[n/2-4:][:8]
		}
	}
	// The JSON-escaped variants of the whole value and its last 8 bytes. The
	// envelope body and the slog record are escaped text: a raw newline in the
	// SEC-PEM value renders there as \n (two bytes), so a whole-value or
	// last-8 leak into an escaped surface is invisible to a raw-bytes grep.
	// For this corpus (ASCII plus newlines) strconv.Quote escapes exactly as
	// encoding/json does. What would make this useless: dropping it and leaving
	// the newline-bearing forms greppable only on the unescaped surfaces.
	if escaped := escapeForms(value); escaped != value {
		forms["whole (escaped)"] = escaped
		if len(value) >= 8 {
			forms["last 8 (escaped)"] = escapeForms(value[len(value)-8:])
		}
	}
	forms["hex"] = hex.EncodeToString([]byte(value))
	forms["base64"] = base64.StdEncoding.EncodeToString([]byte(value))
	forms["sha256 hex"] = sha256Hex([]byte(value))
	forms["sha256 base64"] = base64.StdEncoding.EncodeToString(sha256Bytes([]byte(value)))
	return forms
}

func escapeForms(s string) string {
	return strings.Trim(strconv.Quote(s), "\"")
}

// TestEntropyRuleIsDeterministic is A2-9.4's TestEntropyRuleIsDeterministic.
func TestEntropyRuleIsDeterministic(t *testing.T) {
	hostile := "Zx7Kq0WrT2sPbvLdNmH8cYf4gJaRoUeQIkX5VbN0mM" +
		"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789+/=" +
		strings.Repeat("Aa1!-", 200)

	t.Run("same_input_1000_times", func(t *testing.T) {
		first := fmt.Sprint(Scan("node.label", hostile))
		for i := 0; i < 1000; i++ {
			if got := fmt.Sprint(Scan("node.label", hostile)); got != first {
				t.Fatalf("iteration %d = %q, want %q (SEC-ENTROPY MUST be deterministic)", i, got, first)
			}
		}
		if id := reportedRuleID(t, Scan("node.label", hostile)); id != "SEC-ENTROPY" {
			t.Errorf("reported %q, want SEC-ENTROPY", id)
		}
	})

	t.Run("concurrent_scans_agree", func(t *testing.T) {
		// The callers are concurrent (an ingest path per request, a job worker
		// per task) and the rule table is a package-level slice, so "the same
		// result" has to hold across goroutines too, not just across repeats.
		// Run with -race: a mutated shared rule slice or a lazily built pattern
		// cache shows up here.
		const writers = 8
		want := fmt.Sprint(Scan("node.label", hostile))
		got := make([]string, writers)
		var wg sync.WaitGroup
		for i := 0; i < writers; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				for n := 0; n < 125; n++ {
					got[i] = fmt.Sprint(Scan("node.label", hostile))
				}
			}(i)
		}
		wg.Wait()
		for i, s := range got {
			if s != want {
				t.Errorf("goroutine %d = %q, want %q", i, s, want)
			}
		}
	})

	t.Run("window_boundary", func(t *testing.T) {
		// 32 distinct alphabet characters: H = log2(32) = 5.0 >= 4.5 -> match.
		// 31 distinct alphabet characters: H = log2(31) = 4.95, which is over
		// the bits/char bound but UNDER the 32-character window, so the rule
		// MUST NOT fire: A2-9.4 requires a window of >= 32 characters.
		const high32 = "Bzk9X4QwR2T7pLm1NfYh6VsD0cGjA5uE"
		const high31 = "Bzk9X4QwR2T7pLm1NfYh6VsD0cGjA5u"
		if len(high32) != 32 || len(high31) != 31 {
			t.Fatalf("corpus lengths drifted: %d and %d", len(high32), len(high31))
		}
		if err := Scan("node.label", high31); err != nil {
			t.Errorf("31-character high-entropy run rejected (%q), want nil (window bound is >= 32)", err)
		}
		err := Scan("node.label", high32)
		if err == nil {
			t.Fatalf("32-character high-entropy run accepted, want SEC-ENTROPY")
		}
		if id := reportedRuleID(t, err); id != "SEC-ENTROPY" {
			t.Errorf("reported %q, want SEC-ENTROPY", id)
		}
	})

	t.Run("threshold_boundary", func(t *testing.T) {
		// Both runs are 32 alphabet characters, so only the bits/char
		// distribution differs. All probabilities here are exact powers of
		// two, so the float comparison at >= 4.5 has no rounding slack:
		//   below: 12 singletons + 10 pairs -> H = 12*(1/32)*5 + 10*(2/32)*4
		//                                   = 1.875 + 2.5 = 4.375 <  4.5
		//   at:    16 singletons +  8 pairs -> H = 2.5 + 2.0 = 4.5   >= 4.5
		// So "below" MUST be accepted and "exactly at the bound" MUST be
		// rejected: the clause says >= 4.5, not > 4.5.
		below := buildRun(12, 10)
		at := buildRun(16, 8)
		if len(below) != 32 || len(at) != 32 {
			t.Fatalf("built runs are not 32 characters: %d, %d", len(below), len(at))
		}
		if err := Scan("node.label", below); err != nil {
			t.Errorf("run at 4.375 bits/char rejected (%q), want nil", err)
		}
		err := Scan("node.label", at)
		if err == nil {
			t.Fatalf("run at exactly 4.5 bits/char accepted, want SEC-ENTROPY (the bound is >=)")
		}
		if id := reportedRuleID(t, err); id != "SEC-ENTROPY" {
			t.Errorf("reported %q, want SEC-ENTROPY", id)
		}
		// A third point inside the gap between "below" and the bound, to pin the
		// threshold from below. The two runs above only prove the bound is
		// somewhere in (4.375, 4.5]: nothing in this file (and nothing in
		// benignCorpus, whose maximum is H = 3.671) lands in that interval, so
		// entropyMinRunBits 4.5 -> 4.4 passes every other assertion here. This one
		// cannot:
		//   14 singletons + 9 pairs, 14 + 2*9 = 32 characters:
		//   H = 14*(1/32)*log2(32) + 9*(2/32)*log2(16)
		//     = 14*5/32 + 9*4/16 = 70/32 + 72/32 = 142/32 = 4.4375 < 4.5
		// so the run MUST be accepted as clean. Every probability here is an exact
		// power of two, so the arithmetic holds bit-for-bit; the tolerance is
		// slack for the summation order, not for the value.
		mid := buildRun(14, 9)
		if len(mid) != 32 {
			t.Fatalf("built run is %d characters, want 32", len(mid))
		}
		if got := shannonBitsPerByte(mid); math.Abs(got-4.4375) > 1e-9 {
			t.Errorf("4.4375 run measures %.6f bits/char, want 4.4375 (the arithmetic above is what the next assertion rests on)", got)
		}
		if err := Scan("node.label", mid); err != nil {
			t.Errorf("run at 4.4375 bits/char rejected (%q), want nil: the bound is >= 4.5", err)
		}
	})

	t.Run("documented_consequences_of_the_reading", func(t *testing.T) {
		// doc.go states two consequences of the maximal-run reading, and both
		// are checkable, so they are asserted rather than claimed.
		//
		// 1. A 64-character lowercase hex digest can carry at most
		//    log2(16) = 4.0 bits/char, so it never trips SEC-ENTROPY and
		//    artifact digests (A0-8.7) survive.
		// 2. A random 64-character base64 token sits near 5.2 bits/char, well
		//    over the 4.5 bound, so it always trips.
		//
		// The seed is fixed: rand without a fixed seed would make this a
		// flaky test, and the point of the subtest is the deterministic
		// classification, not the sampling.
		rng := rand.New(rand.NewSource(20260925))
		var trips, total int
		minBits, maxBits := 8.0, 0.0
		for i := 0; i < 500; i++ {
			buf := make([]byte, 48)
			for j := range buf {
				buf[j] = byte(rng.Intn(256))
			}
			token := base64.StdEncoding.EncodeToString(buf) // 64 chars
			if len(token) != 64 {
				t.Fatalf("base64 of 48 bytes is %d characters, want 64", len(token))
			}
			total++
			bits := shannonBitsPerByte(token)
			minBits, maxBits = math.Min(minBits, bits), math.Max(maxBits, bits)
			if Scan("node.attrs.token", token) != nil {
				trips++
			}
		}
		if trips != total {
			t.Errorf("%d of %d random 64-character base64 tokens tripped SEC-ENTROPY, want all of them", trips, total)
		}
		if minBits < 4.5 {
			t.Errorf("sampled base64 tokens fell to %.3f bits/char, below the 4.5 bound; the doc.go consequence statement is wrong", minBits)
		}
		for _, digest := range []string{
			"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			hex.EncodeToString(make([]byte, 32)),
			strings.Repeat("0123456789abcdef", 4),
		} {
			// The ceiling the doc.go claim rests on: 16 symbols cannot exceed
			// log2(16) = 4.0 bits/char, no matter how the digest is chosen. An
			// off-by-one in entropyWindowBytes or a widened alphabet makes this
			// fire; a 4.0 bound that a real digest exceeds means the measurement
			// is not Shannon entropy over the run.
			if got := shannonBitsPerByte(digest); got > 4.0+1e-9 {
				t.Errorf("lowercase hex digest %q measures %.4f bits/char, above the log2(16)=4.0 ceiling stated in doc.go", digest, got)
			}
			if err := Scan("node.attrs.image_digest", digest); err != nil {
				t.Errorf("lowercase hex digest %q rejected: %q", digest, err)
			}
		}
		t.Logf("sampled base64 entropy in [%.3f, %.3f] bits/char over %d tokens", minBits, maxBits, total)
	})

	t.Run("run_splitting_is_maximal_not_arbitrary", func(t *testing.T) {
		// The same 32 characters embedded in prose must be found as one
		// maximal run; separated by a non-alphabet character it must not be.
		// '.' is NOT in the SEC-ENTROPY alphabet, so it breaks runs.
		const high32 = "Bzk9X4QwR2T7pLm1NfYh6VsD0cGjA5uE"
		if err := Scan("node.summary", "observed "+high32+" in output"); err == nil {
			t.Errorf("run delimited by spaces scanned clean, want SEC-ENTROPY")
		}
		half := len(high32) / 2
		if err := Scan("node.summary", high32[:half]+"."+high32[half:]); err != nil {
			t.Errorf("two 16-character runs rejected (%q), want nil", err)
		}
		// The dot case above is what catches '.' being ADDED to the alphabet.
		// Nothing caught a space: with ' ' inside [A-Za-z0-9+/=_-] this string is
		// ONE run of 33 distinct characters, H = log2(33) = 5.0444 >= 4.5, so it
		// would be rejected; as written the space splits it into two 16-character
		// runs, both under the 32-character window, so Scan MUST return nil.
		if err := Scan("node.summary", high32[:half]+" "+high32[half:]); err != nil {
			t.Errorf("two 16-character runs split by a space rejected (%q), want nil: a space is not in the SEC-ENTROPY alphabet", err)
		}
	})
}

// buildRun returns a 32-character run of distinct alphabet characters with
// singleCount characters appearing once and pairCount characters appearing
// twice (singleCount + 2*pairCount == 32).
func buildRun(singleCount, pairCount int) string {
	const alphabet = "zyxwvutsrqponmlkjihgfeDCBA0987654321MLKJIHGF"
	var b strings.Builder
	for i := 0; i < singleCount+pairCount; i++ {
		b.WriteByte(alphabet[i])
	}
	b.WriteString(alphabet[:pairCount])
	return b.String()
}
