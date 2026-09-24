package ids

import (
	"crypto/rand"
	"io"
	"strings"
	"time"

	"github.com/daten-krake/sleipnir/internal/errs"
)

// BodyLen is the length of the identifier body of A0-1.1: 10 characters
// encoding the 48-bit millisecond Unix time, then 16 characters encoding 80
// bits of crypto/rand entropy. A full identifier is len(prefix)+BodyLen
// characters (the A0-1.2 table: 29, 30, 31 or 35).
const BodyLen = 26

// Alphabet is the 32-character lowercase Crockford base32 alphabet of A0-1.1:
// 0-9 plus a-z without i, l, o and u, in byte-lexicographic order. That order
// equals numeric value order, which is what makes equal-length bodies sort
// chronologically — a storage convenience, not an API guarantee (A0-1.6).
// Alphabet is exactly the [0-9a-hjkmnp-tv-z] class of A0-1.2.
const Alphabet = "0123456789abcdefghjkmnpqrstvwxyz"

// The byte layout of the 128-bit value a body encodes (A0-1.1): timeLen bytes
// of big-endian millisecond timestamp, then entropyLen bytes of random data.
const (
	timeLen    = 6
	entropyLen = 10
)

// Kind is an entity type, named by the closed set of 12 id prefixes of A0-1.2.
// A Kind's string value is its prefix, so a prefix is never renamed, reused for
// another type or dropped within /api/v1 (A0-1.10); a new entity type gets a
// new constant by contract amendment.
type Kind string

// The closed prefix set of A0-1.2. slp_node_ is fixed by Q9; A0-1.8 records
// that the node identifier is public and the node mesh token is secret, and
// that slp_node_ is the only prefix carrying the slp_ namespace.
const (
	Engagement Kind = "eng_"
	Run        Kind = "run_"
	Job        Kind = "job_"
	Task       Kind = "task_"
	Event      Kind = "evt_"
	AgentNode  Kind = "slp_node_" // Q9
	GraphNode  Kind = "gn_"
	GraphEdge  Kind = "ge_"
	Evidence   Kind = "evi_"
	Approval   Kind = "apr_"
	Tool       Kind = "tool_" // A0-1.3: name and version are registry fields, never part of the id
	KindUser   Kind = "usr_"  // human principal (SPEC §3)
)

// known reports whether k is one of the 12 prefixes of A0-1.2. The switch is
// the single place a new Kind constant has to be mentioned besides the
// declaration itself; both New and Valid refuse a Kind that is missing from it,
// so an unregistered kind can never mint or accept ids.
func (k Kind) known() bool {
	switch k {
	case Engagement, Run, Job, Task, Event, AgentNode, GraphNode, GraphEdge,
		Evidence, Approval, Tool, KindUser:
		return true
	default:
		return false
	}
}

// New returns a fresh identifier of kind k: k's prefix followed by a BodyLen
// body of A0-1.1, whose first character is always in 0-7 because only the top 3
// bits of the timestamp land in it.
//
// New errors only on entropy failure or an unknown Kind, both errs.Internal
// (A0-3.1: a platform defect, never validation — the caller passes a constant
// from this package, not client material). The entropy source is read exactly
// once; there is no retry loop (A0-1.4).
//
// A Kind is therefore never caller-supplied. A decoder holding client text MUST
// resolve it to a Kind itself and reject an unknown value as errs.Validation
// (A0-6.3) at that boundary: passing the text through as a Kind would turn a
// 400 into a 500 and put client bytes in the message.
//
// The result is opaque to clients (A0-1.6) and is not a secret (A0-1.7).
func New(k Kind) (string, error) {
	return newID(k, time.Now(), rand.Reader)
}

// newID is New with both of its inputs — the clock and the entropy source —
// injected as parameters, which is what makes it testable without package-level
// state (DESIGN §4). r is read once, for the entropyLen random bytes; a short
// read is an error, not a second attempt (A0-1.4).
//
// Timestamps outside the 48-bit window (before 1970-01-01 or after the year
// 10889) keep their low 48 bits, so the shape stays valid while ordering does
// not; A0-1.1 fixes the width and A0-1.6 disclaims ordering as an API promise.
func newID(k Kind, now time.Time, r io.Reader) (string, error) {
	if !k.known() {
		return "", errs.Newf(errs.Internal,
			"generating an identifier: kind %.16q is not one of the closed A0-1.2 prefixes", string(k))
	}

	var raw [timeLen + entropyLen]byte
	ms := uint64(now.UnixMilli())
	for i := 0; i < timeLen; i++ {
		raw[i] = byte(ms >> (8 * uint(timeLen-1-i)))
	}
	n, err := r.Read(raw[timeLen:])
	if err != nil {
		return "", errs.Wrapf(err, "reading %d entropy bytes: kind %q", entropyLen, string(k))
	}
	if n != entropyLen {
		return "", errs.Wrapf(io.ErrUnexpectedEOF,
			"reading %d entropy bytes: kind %q got %d bytes", entropyLen, string(k), n)
	}

	var body [BodyLen]byte
	encodeBody(body[:], raw[:])
	return string(k) + string(body[:]), nil
}

// Valid reports whether s is a byte-exact match for k's A0-1.2 regex,
// '^<prefix>[0-9a-hjkmnp-tv-z]{26}$' (A0-1.5). It is the validation that every
// trust boundary runs: no Crockford normalization, no case folding, no Unicode
// leniency, because ids enter canonical JSON and therefore digests (A0-2). An
// unknown Kind and the empty string are false.
//
// A body whose first character is 8-z is accepted (A0-1.1: validation is the
// regex only) but never generated by New.
func Valid(k Kind, s string) bool {
	if !k.known() {
		return false
	}
	if len(s) != len(k)+BodyLen {
		return false
	}
	if s[:len(k)] != string(k) {
		return false
	}
	for i := len(k); i < len(s); i++ {
		if strings.IndexByte(Alphabet, s[i]) < 0 {
			return false
		}
	}
	return true
}

// encodeBody writes the BodyLen base32 characters of raw (A0-1.1), most
// significant bit first: the 128 bits of raw left-aligned in a 130-bit frame
// whose first two bits are padding zeros. That is what restricts dst[0] to 0-7
// and leaves dst[10] onward to the entropy bytes.
func encodeBody(dst, raw []byte) {
	for i := range dst {
		var v byte
		for b := 0; b < 5; b++ {
			v = v<<1 | frameBit(raw, i*5+b)
		}
		dst[i] = Alphabet[v]
	}
}

// frameBit returns bit pos of the 130-bit encoding frame of raw, counted from
// the most significant end. Positions 0 and 1 are the padding bits; position 2
// is the most significant bit of raw[0].
func frameBit(raw []byte, pos int) byte {
	p := pos - 2
	if p < 0 {
		return 0
	}
	return raw[p/8] >> (7 - uint(p%8)) & 1
}
