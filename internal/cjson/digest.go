package cjson

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
)

// SHA256Hex returns the digest of b as 64 lowercase hex characters (A0-2.15).
// Callers pass the canonical bytes produced by Canonical: a digest of anything
// else is not reproducible (A0-2.16).
func SHA256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// DigestEqual reports whether two hex digests denote the same bytes, comparing
// the decoded bytes in constant time (A0-2.15). It is the comparison for every
// gating decision: hash-chain verification, approval fingerprint match,
// webhook MAC check.
//
// A digest is 64 hex characters of SHA-256 (A0-2.15), so a malformed hex
// string, an odd length, an empty string or any other width decodes to
// something that is not a digest and returns false; nothing here panics.
//
// The width check is a format test, not a content test: it reveals only that
// one of the two strings is not 64 hex characters, which is what makes
// DigestEqual("", "") false (subtle.ConstantTimeCompare reports two empty
// slices as equal). Beyond it the only length-dependent branch is the one
// crypto/subtle already has — ConstantTimeCompare returns 0 immediately when
// the decoded lengths differ, and after the width check that cannot happen —
// because a plain `a == b` on the hex strings would leak the position of the
// first differing byte, and this function gates the hash chain and the
// approval fingerprint.
func DigestEqual(a, b string) bool {
	decodedA, err := hex.DecodeString(a)
	if err != nil {
		return false
	}
	decodedB, err := hex.DecodeString(b)
	if err != nil {
		return false
	}
	if len(decodedA) != sha256.Size || len(decodedB) != sha256.Size {
		return false
	}
	return subtle.ConstantTimeCompare(decodedA, decodedB) == 1
}
