package errs

import (
	"fmt"
	"io"
)

// Redacted is the single literal that replaces any secret material that reaches
// a message, an error string, a log record or a response body (ADR-0019 §5,
// A0-3.7). It carries no length, prefix, suffix or digest information about the
// value it stands for.
const Redacted = "[redacted]"

// Redact returns the [Redacted] placeholder for any input. It never inspects,
// measures, prefixes, suffixes, hashes or truncates its argument, so its output
// leaks nothing about the value — not its length, not its first characters, not
// a digest an attacker could verify offline. The constant behaviour is the
// contract; callers may pass a secret unconditionally.
func Redact(value string) string {
	return Redacted
}

// Secret holds a credential, token, session cookie, captured hash or any other
// value that A0-3.7 forbids in an error or log record, and makes it safe to
// carry that value through code that formats it.
//
// Secret is a struct with an unexported field, deliberately: a named string
// type (type Secret string) prints its value under %q, %x, %X and %d, which
// bypass String() entirely. On top of that, the value is held behind a closure
// rather than a string field, because fmt's badVerb path — reached by %p and by
// every verb that is invalid for a struct — prints the operand by reflection
// with Stringer and Formatter suppressed, and would therefore render a string
// field verbatim. A func field prints as an address instead.
//
// The zero Secret holds no value; Reveal on it returns "". Secret is not
// comparable and must not be used as a map key: comparing two secrets for
// equality is a timing side channel and is never what the caller means.
type Secret struct {
	revealValue func() string
}

// NewSecret wraps a raw credential value. The value is never copied into a
// message, an error or a log record by anything in this package; the only way
// back to it is [Secret.Reveal].
func NewSecret(value string) Secret {
	return Secret{revealValue: func() string { return value }}
}

// Reveal returns the underlying value.
//
// The result MUST NOT be passed to any fmt or slog formatter, an error message,
// a log record or a response body — pass the [Secret] itself, which prints as
// [Redacted] under every verb, or [Redact] its value at the boundary where it
// stops being needed. Reveal exists for exactly one purpose: handing the
// credential to the protocol code that has to send it on the wire (an
// Authorization header, a token exchange).
func (s Secret) Reveal() string {
	if s.revealValue == nil {
		return ""
	}
	return s.revealValue()
}

// Format implements [fmt.Formatter] so that every verb — %v, %+v, %#v, %s, %q,
// %x, %X and any invalid one — writes [Redacted]. Width and flag are ignored on
// purpose: padding would disclose the length of the value the placeholder
// stands for.
//
// %T and %p at the top level are printed by fmt before Formatter is consulted;
// %T therefore reveals the type name (which is why the type is called Secret
// and not something value-carrying), and %p reveals a func address, never the
// secret.
func (s Secret) Format(f fmt.State, verb rune) {
	io.WriteString(f, Redacted)
}

// String implements [fmt.Stringer]. [Secret.Format] is what fmt actually calls;
// String exists for the callers that use it directly.
func (s Secret) String() string { return Redacted }

// GoString implements [fmt.GoStringer] so the %#v form of the value cannot
// reach the field.
func (s Secret) GoString() string { return Redacted }

// MarshalText implements [encoding.TextMarshaler] so a Secret in any
// text-marshalable position serializes as [Redacted].
func (s Secret) MarshalText() ([]byte, error) { return []byte(Redacted), nil }

// MarshalJSON implements [encoding/json.Marshaler]. A Secret never marshals to
// null (A0-8.3) and never to its value; it marshals to the placeholder string.
func (s Secret) MarshalJSON() ([]byte, error) { return []byte(`"` + Redacted + `"`), nil }
