package graph

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// contractRelPath is the frozen A2 contract as seen from this package's
// directory. The closed-list tests parse it directly rather than trusting a
// transcription: the contract text is the pinning source (WP-14 pinning
// rule), and a test that cannot read it fails loudly instead of silently
// pinning stale values.
const contractRelPath = "../../contracts/A2-graph.md"

// contractText reads the frozen contract, or fails the test: a package whose
// closed lists cannot be checked against their source does not pass.
func contractText(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading the frozen contract at %s: %v", path, err)
	}
	return string(b)
}

// contractSlice returns the text between two literal markers, failing the
// test when either is absent or appears in the wrong order. Clause-group
// markers ("### A2-2", "### A2-3", ...) scope every parser to its section so
// a same-shaped table elsewhere in the document cannot leak into a pin.
func contractSlice(t *testing.T, text, from, to string) string {
	t.Helper()
	start := strings.Index(text, from)
	if start < 0 {
		t.Fatalf("contract marker %q not found", from)
	}
	rest := text[start+len(from):]
	end := strings.Index(rest, to)
	if end < 0 {
		t.Fatalf("contract marker %q not found after %q", to, from)
	}
	return rest[:end]
}

// nodeKindRowRE matches one row of A2-2.1's numbered table: | 1 | `host` |.
// The tables are indented inside their clause's bullet list, so leading
// whitespace is allowed. Only the numbered table matches — A2-2.3's per-kind
// field table starts its rows with a backticked kind and no number, so it
// cannot leak in.
var nodeKindRowRE = regexp.MustCompile("^\\s*\\|\\s*(\\d+)\\s*\\|\\s*`([a-z_]+)`\\s*\\|")

// parseNodeKinds extracts the closed A2-2.1 node-kind list from the contract
// section, in table order, checking the row numbers are the sequence 1..n.
func parseNodeKinds(t *testing.T, a22Section string) []NodeKind {
	t.Helper()
	var kinds []NodeKind
	for _, line := range strings.Split(a22Section, "\n") {
		m := nodeKindRowRE.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		n, err := strconv.Atoi(m[1])
		if err != nil || n != len(kinds)+1 {
			t.Fatalf("A2-2.1 row %q: number %q is not the next in sequence (%d)", line, m[1], len(kinds)+1)
		}
		kinds = append(kinds, NodeKind(m[2]))
	}
	if len(kinds) == 0 {
		t.Fatal("no A2-2.1 node-kind rows parsed — the table shape changed; fix the parser, do not weaken the pin")
	}
	return kinds
}

// edgeRowRE matches one row of A2-3.2's endpoint matrix (indented inside the
// clause's bullet list): the kind cell is backticked, the two endpoint cells
// are captured raw. The header row ("| kind | source kinds |") has no
// backticks and does not match.
var edgeRowRE = regexp.MustCompile("^\\s*\\|\\s*`([a-z_]+)`\\s*\\|([^|]*)\\|([^|]*)\\|")

// backtickTokenRE extracts every `token` from a matrix cell.
var backtickTokenRE = regexp.MustCompile("`([a-z_]+)`")

// edgeMatrixRow is one parsed A2-3.2 row: the allowed endpoint sets, or the
// same-kind rule of the supersedes row ("any kind *K*" → "same kind *K*").
type edgeMatrixRow struct {
	kind     EdgeKind
	sources  []NodeKind
	targets  []NodeKind
	sameKind bool
}

// allows is the parsed row's expectation for one endpoint pair. For the
// same-kind row the pairs tested are built from the parsed A2-2.1 list, so
// kind validity is covered by the iteration itself.
func (r edgeMatrixRow) allows(source, target NodeKind) bool {
	if r.sameKind {
		return source == target
	}
	return slices.Contains(r.sources, source) && slices.Contains(r.targets, target)
}

// parseEdgeMatrix extracts the closed A2-3.1 edge-kind list together with
// A2-3.2's endpoint matrix from the contract section, in table order.
func parseEdgeMatrix(t *testing.T, a23Section string) []edgeMatrixRow {
	t.Helper()
	var rows []edgeMatrixRow
	for _, line := range strings.Split(a23Section, "\n") {
		m := edgeRowRE.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		row := edgeMatrixRow{kind: EdgeKind(m[1])}
		if strings.Contains(m[2], "any kind") && strings.Contains(m[3], "same kind") {
			row.sameKind = true
		} else {
			for _, km := range backtickTokenRE.FindAllStringSubmatch(m[2], -1) {
				row.sources = append(row.sources, NodeKind(km[1]))
			}
			for _, km := range backtickTokenRE.FindAllStringSubmatch(m[3], -1) {
				row.targets = append(row.targets, NodeKind(km[1]))
			}
		}
		rows = append(rows, row)
	}
	if len(rows) == 0 {
		t.Fatal("no A2-3.2 matrix rows parsed — the table shape changed; fix the parser, do not weaken the pin")
	}
	return rows
}

// parseContentDocKeys extracts the fixed 20-key contentDoc key set from
// A2-4.6: the backticked list between "(A0-2.14): " (the first occurrence
// after the "**fixed key set**" anchor — A2-4.6 cites A0-2.14 a second time
// later in the clause) and the "— every field" terminator.
func parseContentDocKeys(t *testing.T, a24Section string) []string {
	t.Helper()
	anchor := strings.Index(a24Section, "**fixed key set**")
	if anchor < 0 {
		t.Fatal("A2-4.6's \"fixed key set\" anchor not found")
	}
	start := strings.Index(a24Section[anchor:], "(A0-2.14): ")
	if start < 0 {
		t.Fatal("A2-4.6's \"(A0-2.14): \" list opener not found")
	}
	span := a24Section[anchor+start:]
	end := strings.Index(span, "— every field")
	if end < 0 {
		t.Fatal("A2-4.6's \"— every field\" terminator not found")
	}
	var keys []string
	for _, m := range backtickTokenRE.FindAllStringSubmatch(span[:end], -1) {
		keys = append(keys, m[1])
	}
	if len(keys) == 0 {
		t.Fatal("no contentDoc keys parsed — the clause shape changed; fix the parser, do not weaken the pin")
	}
	return keys
}

// parseReservedAttrKeys extracts A2-6.3's normative closed reserved-key list
// from the "ReservedAttrKeys = {…}" literal (which the contract wraps across
// several lines; TrimSpace absorbs the breaks and indentation).
func parseReservedAttrKeys(t *testing.T, a26Section string) []string {
	t.Helper()
	open := strings.Index(a26Section, "ReservedAttrKeys = {")
	if open < 0 {
		t.Fatal("A2-6.3's ReservedAttrKeys literal not found")
	}
	span := a26Section[open+len("ReservedAttrKeys = {"):]
	end := strings.Index(span, "}")
	if end < 0 {
		t.Fatal("A2-6.3's ReservedAttrKeys literal is not closed")
	}
	var keys []string
	for _, part := range strings.Split(span[:end], ",") {
		keys = append(keys, strings.TrimSpace(part))
	}
	return keys
}

// provenanceFieldRow is one parsed row of A2-5.3's field table: the field
// name and the Req column verbatim ("yes", "no", or a condition).
type provenanceFieldRow struct {
	name string
	req  string
}

// parseProvenanceFields extracts A2-5.3's closed field table (the one with
// the "| Field | Type | Req | Meaning |" header) from the A2-5 section.
func parseProvenanceFields(t *testing.T, a25Section string) []provenanceFieldRow {
	t.Helper()
	lines := strings.Split(a25Section, "\n")
	start := -1
	for i, line := range lines {
		if strings.Contains(line, "| Field | Type | Req | Meaning |") {
			start = i + 1
			break
		}
	}
	if start < 0 {
		t.Fatal("A2-5.3's field-table header not found")
	}
	var rows []provenanceFieldRow
	for _, line := range lines[start:] {
		if !strings.HasPrefix(strings.TrimSpace(line), "|") {
			break
		}
		if strings.Contains(line, "|---") {
			continue // the header separator
		}
		cells := strings.Split(line, "|")
		if len(cells) < 5 {
			t.Fatalf("A2-5.3 table row %q has fewer than four cells", line)
		}
		m := backtickTokenRE.FindStringSubmatch(cells[1])
		if m == nil {
			t.Fatalf("A2-5.3 table row %q has no backticked field name", line)
		}
		rows = append(rows, provenanceFieldRow{name: m[1], req: strings.TrimSpace(cells[3])})
	}
	if len(rows) == 0 {
		t.Fatal("no A2-5.3 field rows parsed — the table shape changed; fix the parser, do not weaken the pin")
	}
	return rows
}

// mutateContract returns the contract text with old replaced by new, failing
// the test unless old occurs exactly once — a mutation whose anchor is
// ambiguous proves nothing.
func mutateContract(t *testing.T, text, old, new string) string {
	t.Helper()
	if n := strings.Count(text, old); n != 1 {
		t.Fatalf("mutation anchor %q occurs %d times in the contract, want exactly 1", old, n)
	}
	return strings.Replace(text, old, new, 1)
}

// writeTempContract writes a (mutated) contract copy OUTSIDE the repo —
// t.TempDir() is under the test's TMPDIR — and returns its path. The repo's
// frozen contract is never modified (AGENTS.md: mutate only copies outside
// the repo).
func writeTempContract(t *testing.T, text string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "A2-graph.md")
	if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
		t.Fatalf("writing the mutated contract copy: %v", err)
	}
	return path
}

func TestNodeKindListClosed(t *testing.T) {
	text := contractText(t, contractRelPath)
	parsed := parseNodeKinds(t, contractSlice(t, text, "### A2-2", "### A2-3"))

	// The pin: the contract's own A2-2.1 table, parsed — not a transcription
	// of this package's consts — is exactly the closed list, in order.
	if len(parsed) != 10 {
		t.Fatalf("A2-2.1 parses to %d kinds, want exactly 10: %v", len(parsed), parsed)
	}
	if !slices.Equal(parsed, NodeKinds()) {
		t.Fatalf("A2-2.1 contract list %v != package NodeKinds() %v", parsed, NodeKinds())
	}

	t.Run("Valid_accepts_exactly_the_contract_list", func(t *testing.T) {
		for _, k := range parsed {
			if !k.Valid() {
				t.Errorf("NodeKind(%q).Valid() = false, want true (contract row of A2-2.1)", string(k))
			}
		}
	})

	t.Run("Valid_rejects_unknown_and_near_misses", func(t *testing.T) {
		rejects := []NodeKind{
			"", "hosts", "Host", "HOST", "host ", " host", "evidence-ref",
			"evidenceref", "identity2", "finding ", "hypothesis\n", "node",
			"service\t", "credential_kind", "open", "reachable",
		}
		for _, k := range rejects {
			if k.Valid() {
				t.Errorf("NodeKind(%q).Valid() = true, want false — A2-2.1 is closed, comparison is byte-exact (A0-8.5)", string(k))
			}
		}
	})

	// Pinning rule: prove the pin fails on a mutated copy (outside the repo).
	t.Run("mutation_proof", func(t *testing.T) {
		mutated := mutateContract(t, text, "| 1 | `host` |", "| 1 | `hosts` |")
		path := writeTempContract(t, mutated)
		got := parseNodeKinds(t, contractSlice(t, contractText(t, path), "### A2-2", "### A2-3"))
		if slices.Equal(got, NodeKinds()) {
			t.Fatalf("the pin did not fail on the mutated copy: host→hosts still parsed as %v", got)
		}
		if got[0] != "hosts" {
			t.Fatalf("mutated copy parsed %v, want the mutation visible at row 1", got)
		}
	})
}

func TestEdgeKindListClosed(t *testing.T) {
	text := contractText(t, contractRelPath)
	rows := parseEdgeMatrix(t, contractSlice(t, text, "### A2-3", "### A2-4"))

	var parsed []EdgeKind
	for _, r := range rows {
		parsed = append(parsed, r.kind)
	}
	if len(parsed) != 7 {
		t.Fatalf("A2-3.2 parses to %d edge kinds, want exactly 7: %v", len(parsed), parsed)
	}
	if !slices.Equal(parsed, EdgeKinds()) {
		t.Fatalf("A2-3.1 contract list %v != package EdgeKinds() %v", parsed, EdgeKinds())
	}
	nodeKinds := parseNodeKinds(t, contractSlice(t, text, "### A2-2", "### A2-3"))

	t.Run("Valid_accepts_exactly_the_contract_list", func(t *testing.T) {
		for _, k := range parsed {
			if !k.Valid() {
				t.Errorf("EdgeKind(%q).Valid() = false, want true (contract row of A2-3.2)", string(k))
			}
		}
	})

	t.Run("Valid_rejects_unknown_and_near_misses", func(t *testing.T) {
		rejects := []EdgeKind{
			"", "Reachable", "REACHABLE", "reacheable", "reachable ", "member-of",
			"superseded", "supersede", "contradict", "host", "grants_access2",
		}
		for _, k := range rejects {
			if k.Valid() {
				t.Errorf("EdgeKind(%q).Valid() = true, want false — A2-3.1 is closed, byte-exact (A0-8.5)", string(k))
			}
		}
	})

	// The endpoint matrix AS DATA, parsed from A2-3.2, cross-checked against
	// AllowsEndpoints for every edge/node-kind triple the contract lists
	// (7 × 10 × 10 — the full closed universe of both lists).
	t.Run("endpoint_matrix_matches_contract", func(t *testing.T) {
		for _, row := range rows {
			for _, src := range nodeKinds {
				for _, tgt := range nodeKinds {
					want := row.allows(src, tgt)
					if got := row.kind.AllowsEndpoints(src, tgt); got != want {
						t.Errorf("%s.AllowsEndpoints(%s, %s) = %v, want %v (A2-3.2 row %q)",
							row.kind, src, tgt, got, want, row.kind)
					}
				}
			}
		}
	})

	t.Run("unknown_kinds_are_never_allowed", func(t *testing.T) {
		// A2-10.5: the closed-list rule covers endpoint kinds, not just the
		// edge kind itself.
		if (EdgeKind("reaches")).AllowsEndpoints(KindHost, KindHost) {
			t.Error("an unknown edge kind must allow nothing")
		}
		if EdgeReachable.AllowsEndpoints(NodeKind("hosts"), KindHost) {
			t.Error("an unknown source kind must allow nothing")
		}
		if EdgeReachable.AllowsEndpoints(KindHost, NodeKind("")) {
			t.Error("an unknown target kind must allow nothing")
		}
		// The supersedes same-kind rule is bounded by kind validity: a
		// matching pair of unknown kinds is still rejected.
		if EdgeSupersedes.AllowsEndpoints(NodeKind("hosts"), NodeKind("hosts")) {
			t.Error("supersedes between two unknown-but-equal kinds must allow nothing")
		}
		if !EdgeSupersedes.AllowsEndpoints(KindHost, KindHost) {
			t.Error("supersedes host→host must be allowed (A2-3.2: any kind K → same kind K)")
		}
		if EdgeSupersedes.AllowsEndpoints(KindHost, KindNetwork) {
			t.Error("supersedes across kinds must be rejected (A2-4.3)")
		}
	})

	t.Run("mutation_proof", func(t *testing.T) {
		mutated := mutateContract(t, text, "| `reachable` | `host`, `network` |", "| `reaches` | `host`, `network` |")
		path := writeTempContract(t, mutated)
		got := parseEdgeMatrix(t, contractSlice(t, contractText(t, path), "### A2-3", "### A2-4"))
		var gotKinds []EdgeKind
		for _, r := range got {
			gotKinds = append(gotKinds, r.kind)
		}
		if slices.Equal(gotKinds, EdgeKinds()) {
			t.Fatalf("the pin did not fail on the mutated copy: reachable→reaches still parsed as %v", gotKinds)
		}
	})
}
