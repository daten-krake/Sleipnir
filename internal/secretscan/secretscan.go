package secretscan

import (
	"math"
	"regexp"

	"github.com/daten-krake/sleipnir/internal/errs"
)

// rule is one row of the normative, closed rule table of A2-9.4. The table
// order below IS the evaluation order (ruling S1) and is fixed by the source,
// never by map iteration. A rule id not present in this table MUST NOT be
// reported (A2-9.4).
type rule struct {
	id string
	// pattern is the A2-9.4 Go regexp, transcribed byte-exactly. It is nil
	// only for SEC-ENTROPY, which the table declares is not a regexp.
	pattern *regexp.Regexp
}

// rules is A2-9.4's table: normative, closed, additive-only (A0-6.5). All ten
// rules are implemented; the first to match in this order is the one reported
// and stops the scan (ruling S1). "An implementation MUST run every rule" is
// read as "no rule may be omitted from the set", not "no rule may be skipped
// after the decision is made": continuing would cost ten passes per field and
// could not change the reported id, since the table order fixes it.
var rules = []rule{
	{id: "SEC-PEM", pattern: regexp.MustCompile(`-----BEGIN [A-Z ]*PRIVATE KEY-----`)},
	{id: "SEC-NTLM", pattern: regexp.MustCompile(`(?i)\b[0-9a-f]{32}\b`)},
	{id: "SEC-KRB", pattern: regexp.MustCompile(`(?i)krbtgt[/@][A-Za-z0-9._-]{1,128}`)},
	{id: "SEC-AWSKEY", pattern: regexp.MustCompile(`(AKIA|ASIA)[0-9A-Z]{16}`)},
	{id: "SEC-GCPKEY", pattern: regexp.MustCompile(`AIza[0-9A-Za-z\-_]{35}`)},
	{id: "SEC-AZUREKEY", pattern: regexp.MustCompile(`(?i)(AccountKey|SharedAccessKey|sig)=[A-Za-z0-9+/=]{20,}`)},
	{id: "SEC-JWT", pattern: regexp.MustCompile(`eyJ[0-9A-Za-z_-]+\.[0-9A-Za-z_-]+\.[0-9A-Za-z_-]+`)},
	{id: "SEC-BEARER", pattern: regexp.MustCompile(`(?i)(bearer|token|api[_-]?key|password|passwd|secret)\s*[:=]\s*\S{8,}`)},
	{id: "SEC-URLCRED", pattern: regexp.MustCompile(`[a-z][a-z0-9+.-]*://[^/\s:@]{1,64}:[^/\s:@]{1,64}@`)},
	{id: "SEC-ENTROPY", pattern: nil},
}

// Scan reports whether value — the content of the caller-named field — matches
// the A2-9.4 rule table, and returns nil when it does not.
//
// A match is a rejection, never a redaction (A2-9.4, §6 item 10, ruling S5):
// the returned error has kind errs.Validation and its message names only the
// field, the rule id and the byte length of the value (ADR-0019 §2, A2-9.5,
// ruling S3). Scan itself puts the value nowhere: not in the message, not in an
// error attrs entry, not in a log record — in whole, in part or as a digest
// (A0-3.4's single exception to "a message may echo untrusted material"). A
// caller that logs the rejection must keep that true on the surfaces it owns.
//
// The field name is caller-supplied and is not a secret; it is echoed
// unescaped, so a caller MUST pass a schema field path ("node.label",
// "payload.summary"), not attacker-controlled text.
//
// Scan is a filter, not a guarantee (A2-9.3): the enforcement point for cloud
// egress is the gateway exclusion of ADR-0020 §4, applied independently. Scan
// never logs (ruling S4, ADR-0019 §3's last bullet, log-or-return); the caller
// that rejects the write logs the returned error once.
func Scan(field, value string) error {
	// Rules run in this slice's order and the first match is the one reported
	// (ruling S1); the order is a slice, never a map, so the reported id is
	// deterministic. "An implementation MUST run every rule" (A2-9.4) is read as
	// "no row may be omitted from the set", not "no row may be skipped once the
	// decision is made": the table order already fixes which id is reported, so
	// continuing past a match would cost ten passes per field and could not
	// change the outcome.
	for _, r := range rules {
		if r.matches(value) {
			return errs.Newf(errs.Validation,
				"scanning field %s for secret material: rule %s matched: %d bytes (A2-9.4)",
				field, r.id, len(value))
		}
	}
	return nil
}

// matches applies one rule to the whole value. Callers never hold a rule: the
// only way in is Scan.
func (r rule) matches(value string) bool {
	if r.pattern == nil {
		return highEntropyRun(value)
	}
	return r.pattern.MatchString(value)
}

// The SEC-ENTROPY thresholds, part of the rule per A2-9.4: entropy in bits per
// character over a window of at least this many base64/hex-alphabet characters.
const (
	entropyMinRunBits  = 4.5
	entropyWindowBytes = 32
)

// inEntropyAlphabet reports whether c is one of the base64/hex alphabet
// characters of A2-9.4's SEC-ENTROPY row: [A-Za-z0-9+/=_-].
func inEntropyAlphabet(c byte) bool {
	switch {
	case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9':
		return true
	case c == '+' || c == '/' || c == '=' || c == '_' || c == '-':
		return true
	}
	return false
}

// highEntropyRun implements SEC-ENTROPY: every maximal run of alphabet
// characters of length >= entropyWindowBytes is measured with
// -Σ p(c)·log2 p(c) over its bytes, and a run at or above entropyMinRunBits is
// a match. One linear pass, no sliding window — see doc.go for why the
// ">= 32 characters" wording does not define a window size.
func highEntropyRun(value string) bool {
	for i := 0; i < len(value); {
		if !inEntropyAlphabet(value[i]) {
			i++
			continue
		}
		end := i
		for end < len(value) && inEntropyAlphabet(value[end]) {
			end++
		}
		if end-i >= entropyWindowBytes && shannonBitsPerByte(value[i:end]) >= entropyMinRunBits {
			return true
		}
		i = end
	}
	return false
}

// shannonBitsPerByte returns -Σ p(c)·log2 p(c) over the bytes of run. Counting
// into a fixed 256-entry array and summing in index order keeps the result a
// pure function of the bytes: no map iteration, so the rule is deterministic
// (A2-9.4's SEC-ENTROPY row, TestEntropyRuleIsDeterministic).
//
// The [256]int is allocated per call on purpose (2 KB): the scan stays O(n)
// with no shared state, and callers apply their caps first (AttrValueMaxBytes
// 512, ProseLongMaxBytes 2048), so the worst-case churn per Scan is ~128 KB.
// Hoisting the array into a shared buffer would need a lock or a pool and buys
// nothing at those sizes.
func shannonBitsPerByte(run string) float64 {
	var counts [256]int
	for i := 0; i < len(run); i++ {
		counts[run[i]]++
	}
	total := float64(len(run))
	var bits float64
	for _, c := range counts {
		if c == 0 {
			continue
		}
		p := float64(c) / total
		bits -= p * math.Log2(p)
	}
	return bits
}
