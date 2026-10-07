package graph

import (
	"fmt"
	"slices"
	"unicode/utf8"

	"github.com/daten-krake/sleipnir/internal/cjson"
	"github.com/daten-krake/sleipnir/internal/errs"
)

// contentDoc is the purpose-built document content_hash is computed over
// (A2-4.6). Fixed key set of exactly 20 keys (A0-2.14), zero values for the
// fields that do not apply to the kind ("" / 0 / [] / {} — never absence,
// never null), keys in UTF-8 byte order (A0-2.4 — the field order below IS
// that byte order), integers only (A0-2.6), and an empty A0-2.12 exclusion
// list: ids, graph_seq, provenance, quarantine flags, report_excluded,
// supersedes_id, superseded_by_id and content_hash are not fields of this
// type at all, so none of them can influence the digest.
//
// Node MUST NOT be passed to cjson for fingerprinting (A2-4.6): Node carries
// omitempty tags and A0-8.3 absence semantics, contentDoc carries the fixed
// 20-key set. There is no time.Time and no float field here (A0-2.6;
// timestamps never enter the fingerprint — none of the 20 keys is a
// timestamp).
type contentDoc struct {
	Addresses      []string       `json:"addresses"`
	Attrs          Attrs          `json:"attrs"`
	Basis          string         `json:"basis"`
	CIDR           string         `json:"cidr"`
	Claim          string         `json:"claim"`
	CredentialKind CredentialKind `json:"credential_kind"`
	Domain         string         `json:"domain"`
	EvidenceID     string         `json:"evidence_id"`
	EvidenceIDs    []string       `json:"evidence_ids"`
	Kind           NodeKind       `json:"kind"`
	Label          string         `json:"label"`
	MediaKind      MediaKind      `json:"media_kind"`
	Port           int            `json:"port"`
	Protocol       string         `json:"protocol"`
	Severity       Severity       `json:"severity"`
	SID            string         `json:"sid"`
	SizeBytes      int64          `json:"size_bytes"`
	Status         string         `json:"status"`
	Summary        string         `json:"summary"`
	Transport      string         `json:"transport"`
}

// CanonicalContent returns the exact A0-2 canonical bytes of n's A2-4.6
// content document: the bytes a node's content_hash is computed over and the
// bytes that MUST be persisted with the node (A2-4.8, A0-2.16), because any
// later verification recomputes from these stored bytes and never from a
// re-serialization of decoded fields. The same bytes are the measurement
// input for A2-6.4's canonical-form attrs cap (WP-15), so the cap check and
// the fingerprint can never disagree.
//
// Building the document applies A2-4.6's composition rules: addresses and
// evidence_ids are copied, sorted ascending by unsigned byte value and
// deduplicated BEFORE canonicalization (A1-4.7's rule — two observations of
// the same content in a different input order produce the same bytes), attrs
// keys are ordered by the canonicalizer (A0-2.4), and every collection is
// non-nil in the result — an absent (nil) attrs, addresses or evidence_ids
// serializes as {} / [] (A0-2.14, §4.2 vector S1), never as null. n is not
// mutated: the sort operates on copies (DESIGN §3).
//
// Rejections: a string anywhere in the document that is not valid UTF-8 is
// validation (A0-2.3) naming the field and the byte length — encoding/json
// would replace the invalid bytes with U+FFFD before the canonicalizer could
// see them, so this check is the only place the rule is enforceable on a Go
// value (BACKLOG 2026-09-24 open question, ruled here: the domain type owns
// its field-level UTF-8 validation at the fingerprint boundary). A hand-built
// AttrValue with no live discriminator or an out-of-range integer is
// errs.Internal: a canonical content document that cannot be produced from
// decoded wire data is a platform defect (A2-4.6's defect rule). Everything
// else the canonicalizer itself rejects (through cjson) stays validation.
func CanonicalContent(n Node) ([]byte, error) {
	doc, err := contentDocOf(n)
	if err != nil {
		return nil, err
	}
	b, err := cjson.CanonicalValue(doc)
	if err != nil {
		return nil, errs.Wrapf(err, "canonicalizing graph node content: kind=%.32q (A2-4.6)", string(n.Kind))
	}
	return b, nil
}

// ContentHash returns n's content_hash: the 64-char lowercase hex SHA-256 of
// CanonicalContent's bytes (A2-4.6, A0-2.15). It is an integrity and dedup
// value, not a secret and not a substitute for the A1 event hash chain: it is
// not chained, and a mismatch against stored bytes is a platform defect
// (internal), never a customer-facing integrity_failed (A2-4.9).
func ContentHash(n Node) (string, error) {
	b, err := CanonicalContent(n)
	if err != nil {
		return "", err
	}
	return cjson.SHA256Hex(b), nil
}

// VerifyContentHash recomputes the digest from the canonical bytes that were
// persisted with the node (A2-4.8, A0-2.16: verification MUST recompute from
// the stored bytes, never from a re-serialization of decoded fields) and
// compares it with the persisted content_hash in constant time (A0-2.15,
// cjson.DigestEqual).
//
// Any mismatch — empty stored bytes, a malformed digest string, or different
// digest bytes — is errs.Internal, never integrity_failed: content_hash is
// not chained, so a disagreement here means the platform stored bytes and a
// digest that do not belong together (A2-4.9). Both digests are echoed:
// content_hash is explicitly not a secret (A2-4.9), and ADR-0019 §2 requires
// the message to be self-contained for troubleshooting from logs alone.
func VerifyContentHash(storedCanonical []byte, contentHash string) error {
	if len(storedCanonical) == 0 {
		return errs.Newf(errs.Internal,
			"verifying graph node content_hash: the stored canonical bytes are empty — A0-2.16 requires persisting them with the node")
	}
	got := cjson.SHA256Hex(storedCanonical)
	if !cjson.DigestEqual(got, contentHash) {
		return errs.Newf(errs.Internal,
			"verifying graph node content_hash: digest recomputed from the %d stored canonical bytes is %s, persisted content_hash is %.64q: a mismatch is a platform defect, never integrity_failed (A2-4.8, A2-4.9)",
			len(storedCanonical), got, contentHash)
	}
	return nil
}

// contentDocOf builds n's A2-4.6 content document: the twenty content fields,
// arrays sorted and deduplicated on copies, collections non-nil, and every
// string checked for UTF-8 validity in deterministic order.
func contentDocOf(n Node) (contentDoc, error) {
	doc := contentDoc{
		Addresses:      sortedDeduped(n.Addresses),
		Attrs:          n.Attrs,
		Basis:          n.Basis,
		CIDR:           n.CIDR,
		Claim:          n.Claim,
		CredentialKind: n.CredentialKind,
		Domain:         n.Domain,
		EvidenceID:     n.EvidenceID,
		EvidenceIDs:    sortedDeduped(n.EvidenceIDs),
		Kind:           n.Kind,
		Label:          n.Label,
		MediaKind:      n.MediaKind,
		Port:           n.Port,
		Protocol:       n.Protocol,
		Severity:       n.Severity,
		SID:            n.SID,
		SizeBytes:      n.SizeBytes,
		Status:         n.Status,
		Summary:        n.Summary,
		Transport:      n.Transport,
	}
	if doc.Attrs == nil {
		// A0-2.14: a nil map would marshal as null and change every digest;
		// an absent attrs serializes as {} (§4.2 vector S1).
		doc.Attrs = Attrs{}
	}
	if err := doc.checkAttrsLive(); err != nil {
		return contentDoc{}, err
	}
	if err := doc.checkUTF8(); err != nil {
		return contentDoc{}, err
	}
	return doc, nil
}

// checkAttrsLive rejects hand-built AttrValues that cannot enter a canonical
// document: an unknown Type discriminator or an integer outside A0-2.6's
// range. Decoded attrs never carry either (AttrValue.UnmarshalJSON sets a
// discriminator and enforces the range), so both are platform defects —
// errs.Internal, checked here rather than left to AttrValue.MarshalJSON
// because cjson.CanonicalValue would reclassify a marshaler failure as
// validation. Walked in sorted-key order so the first reported defect is
// deterministic (A2-10.2).
func (d contentDoc) checkAttrsLive() error {
	for _, k := range sortedAttrKeys(d.Attrs) {
		v := d.Attrs[k]
		switch v.Type {
		case AttrString, AttrBool:
		case AttrInt:
			if v.Num > maxSafeAttrInt || v.Num < -maxSafeAttrInt {
				return errs.Newf(errs.Internal,
					"canonicalizing graph node content: attrs key %q: integer %d is outside the canonical range [-(2^53-1), 2^53-1] (A0-2.6): a hand-built AttrValue is a platform defect",
					k, v.Num)
			}
		default:
			return errs.Newf(errs.Internal,
				"canonicalizing graph node content: attrs key %q: type discriminator %.16q is not one of string/int/bool (A2-6.1): a hand-built AttrValue is a platform defect",
				k, string(v.Type))
		}
	}
	return nil
}

// namedString is one string of the content document with the field name an
// error must use (A2-10.4: the message names the field).
type namedString struct {
	field string
	value string
}

// strings lists every string that enters the document, in contentDoc key
// order (which is A0-2.4 byte order) and, within arrays and attrs, in sorted
// order — so the first invalid field a rejection names is deterministic no
// matter how the input was built (A2-10.2: exactly one error, the first
// violated rule).
func (d contentDoc) strings() []namedString {
	out := make([]namedString, 0, 15+len(d.Addresses)+len(d.EvidenceIDs)+2*len(d.Attrs))
	for i, s := range d.Addresses {
		out = append(out, namedString{fmt.Sprintf("addresses[%d]", i), s})
	}
	for _, k := range sortedAttrKeys(d.Attrs) {
		out = append(out, namedString{fmt.Sprintf("attrs key %q", k), k})
		if v := d.Attrs[k]; v.Type == AttrString {
			out = append(out, namedString{fmt.Sprintf("attrs[%q]", k), v.Str})
		}
	}
	out = append(out,
		namedString{"basis", d.Basis},
		namedString{"cidr", d.CIDR},
		namedString{"claim", d.Claim},
		namedString{"credential_kind", string(d.CredentialKind)},
		namedString{"domain", d.Domain},
		namedString{"evidence_id", d.EvidenceID})
	for i, s := range d.EvidenceIDs {
		out = append(out, namedString{fmt.Sprintf("evidence_ids[%d]", i), s})
	}
	return append(out,
		namedString{"kind", string(d.Kind)},
		namedString{"label", d.Label},
		namedString{"media_kind", string(d.MediaKind)},
		namedString{"protocol", d.Protocol},
		namedString{"severity", string(d.Severity)},
		namedString{"sid", d.SID},
		namedString{"status", d.Status},
		namedString{"summary", d.Summary},
		namedString{"transport", d.Transport})
}

// checkUTF8 rejects the first string in the document that is not valid
// UTF-8 (A0-2.3) with validation naming the field and the byte length. The
// value itself is never echoed: the field is untrusted content that the
// secret scan has not seen yet at this point (A2-10.2 orders the fingerprint
// at step 13a, after the scan — but CanonicalContent must stay safe to call
// from anywhere), and A2-9.5's discipline is to name the field and the
// length, not the bytes. encoding/json replaces invalid UTF-8 with U+FFFD
// before cjson ever sees the value, so without this check two different
// invalid inputs could silently collapse into one digest (Q3: reject, never
// normalize; BACKLOG 2026-09-24: field-level UTF-8 validation is the domain
// type's).
func (d contentDoc) checkUTF8() error {
	for _, ns := range d.strings() {
		if !utf8.ValidString(ns.value) {
			return errs.Newf(errs.Validation,
				"canonicalizing graph node content: field=%s: %d bytes are not valid UTF-8 (A0-2.3: rejected, never U+FFFD-normalized)",
				ns.field, len(ns.value))
		}
	}
	return nil
}

// sortedDeduped returns a copy of in, sorted ascending by unsigned byte value
// and deduplicated (A2-4.6, A1-4.7's rule), never nil: an absent array
// serializes as [] in the canonical document (A0-2.14). The input slice is
// not mutated (DESIGN §3) — the caller's Node keeps its field order.
func sortedDeduped(in []string) []string {
	out := slices.Clone(in)
	if out == nil {
		out = []string{}
	}
	slices.Sort(out) // Go string comparison is unsigned-byte-wise (A0-1.9's order)
	return slices.Compact(out)
}
