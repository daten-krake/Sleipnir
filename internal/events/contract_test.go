package events

// Contract-oracle parsing helpers shared by the seven WP-09 tests.
//
// Pinning rule (WP-09 brief, AGENTS.md): the closed kind list and the
// expected payload shapes asserted by these tests are parsed from the frozen
// contract text in contracts/A1-events.md — never derived from this
// package's own consts or structs, which would make the tests vacuous. The
// parsers fail loudly (t.Fatalf) on any contract shape they do not
// understand, so a contract amendment breaks the parse instead of silently
// weakening an assertion.

import (
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// contractRelPath locates the frozen A1 contract from this package directory.
const contractRelPath = "../../contracts/A1-events.md"

// kindNameRe is the A0-8.5 spelling rule for event kinds (A1-3.1).
var kindNameRe = regexp.MustCompile(`^[a-z][a-z0-9_]{0,31}$`)

// jsonKeyRe is the A0-8.1 spelling rule for JSON keys.
var jsonKeyRe = regexp.MustCompile(`^[a-z][a-z0-9_]{0,39}$`)

var (
	envelopeRowRe = regexp.MustCompile("^\\s*\\|\\s*(\\d+)\\s*\\|\\s*`([a-z0-9_]+)`\\s*\\|")
	actorPairRe   = regexp.MustCompile(`\(([^,()]+), "([^"]*)"\)`)
	backtickRe    = regexp.MustCompile("`([^`]+)`")
	fieldSpecRe   = regexp.MustCompile(`^([a-z][a-z0-9_]*):(.+)$`)
	otherKindRe   = regexp.MustCompile(`every other kind \((\d+) of 42\)`)
)

func readContract(t *testing.T) []string {
	t.Helper()
	b, err := os.ReadFile(contractRelPath)
	if err != nil {
		t.Fatalf("read frozen contract %s: %v", contractRelPath, err)
	}
	return strings.Split(string(b), "\n")
}

// section returns the contract lines strictly between the first line
// containing startMarker and the next line containing endMarker. Both
// markers must exist; a missing marker means the contract was restructured
// and the oracle must be updated deliberately, not silently.
func section(t *testing.T, lines []string, startMarker, endMarker string) []string {
	t.Helper()
	start := -1
	for i, l := range lines {
		if strings.Contains(l, startMarker) {
			start = i + 1
			break
		}
	}
	if start < 0 {
		t.Fatalf("contract section start %q not found in %s", startMarker, contractRelPath)
	}
	for i := start; i < len(lines); i++ {
		if strings.Contains(lines[i], endMarker) {
			return lines[start:i]
		}
	}
	t.Fatalf("contract section end %q not found after %q in %s", endMarker, startMarker, contractRelPath)
	return nil
}

// splitCells splits a markdown table row into its trimmed cells. Safe for
// A1-3.3/A1-4.4 rows: no cell of those tables contains a raw pipe.
func splitCells(line string) []string {
	trimmed := strings.TrimSpace(line)
	trimmed = strings.TrimPrefix(trimmed, "|")
	trimmed = strings.TrimSuffix(trimmed, "|")
	cells := strings.Split(trimmed, "|")
	for i := range cells {
		cells[i] = strings.TrimSpace(cells[i])
	}
	return cells
}

func isTableRow(line string) bool {
	return strings.HasPrefix(strings.TrimSpace(line), "|")
}

func isSeparatorRow(line string) bool {
	return strings.Contains(line, "---")
}

// parseEnvelopeKeys returns the top-level key names of the A1-1.1 envelope
// table, in table order. A1-1.1 closes the set at 17.
func parseEnvelopeKeys(t *testing.T) []string {
	t.Helper()
	lines := section(t, readContract(t), "### A1-1 · Event envelope", "### A1-2")
	var keys []string
	last := 0
	for _, l := range lines {
		if !isTableRow(l) || isSeparatorRow(l) {
			continue
		}
		m := envelopeRowRe.FindStringSubmatch(l)
		if m == nil {
			continue
		}
		num, err := strconv.Atoi(m[1])
		if err != nil {
			t.Fatalf("A1-1.1 row number %q is not an integer: %q", m[1], l)
		}
		if num != last+1 {
			t.Fatalf("A1-1.1 row numbering jumps from %d to %d: %q", last, num, l)
		}
		last = num
		keys = append(keys, m[2])
	}
	return keys
}

// contractActor is one (type, component) pair from an A1-3.3 actor cell.
// A pair may name a dynamic type ("requesting principal's type").
type contractActor struct {
	atype     string
	component string
}

// contractField is one payload field from an A1-3.3 payload cell, per the
// A1-3.2 notation: name:type with an optional (N) cap and * marker.
type contractField struct {
	name string
	typ  string // string | string(N) | int | bool | 64hex | timestamp | enum{…} | array[string]
	star bool   // the A1-4.4 untrusted-content marker
}

// contractKind is one parsed A1-3.3 table row.
type contractKind struct {
	name       string
	appendable bool // the **C** marker (A1-3.4)
	actors     []contractActor
	fields     []contractField
}

// parseTaxonomy returns every A1-3.3 kind row in table order. It fails the
// test on any row it cannot fully parse — a parser that silently skipped a
// reshaped row would weaken every test built on it.
func parseTaxonomy(t *testing.T) []contractKind {
	t.Helper()
	lines := section(t, readContract(t), "- **A1-3.3**", "- **A1-3.4**")
	var out []contractKind
	for _, l := range lines {
		if !isTableRow(l) || isSeparatorRow(l) {
			continue
		}
		cells := splitCells(l)
		if len(cells) != 4 {
			t.Fatalf("A1-3.3 row has %d cells, want 4: %q", len(cells), l)
		}
		km := backtickRe.FindStringSubmatch(cells[0])
		if km == nil {
			continue // the repeated header row ("Kind") carries no backticked name
		}
		name := km[1]
		if !kindNameRe.MatchString(name) {
			t.Fatalf("A1-3.3 kind %q violates the A0-8.5 spelling rule", name)
		}
		row := contractKind{
			name:       name,
			appendable: strings.Contains(cells[0], "**C**"),
		}
		for _, am := range actorPairRe.FindAllStringSubmatch(cells[1], -1) {
			row.actors = append(row.actors, contractActor{
				atype:     strings.TrimSpace(am[1]),
				component: am[2],
			})
		}
		if len(row.actors) == 0 {
			t.Fatalf("kind %s: no (type, component) pair parsed from %q", name, cells[1])
		}
		for _, fm := range backtickRe.FindAllStringSubmatch(cells[2], -1) {
			f, ok := parseFieldSpec(fm[1])
			if !ok {
				t.Fatalf("kind %s: field spec %q does not follow the A1-3.2 notation", name, fm[1])
			}
			row.fields = append(row.fields, f)
		}
		if len(row.fields) == 0 {
			t.Fatalf("kind %s: no payload fields parsed from %q", name, cells[2])
		}
		out = append(out, row)
	}
	if len(out) == 0 {
		t.Fatal("parsed zero A1-3.3 kind rows — the oracle found nothing to pin")
	}
	return out
}

// parseFieldSpec splits one A1-3.2 field spec ("reason:string(512)*") into
// its parts. ok is false when the spec does not follow the notation.
func parseFieldSpec(spec string) (contractField, bool) {
	var f contractField
	if strings.HasSuffix(spec, "*") {
		f.star = true
		spec = strings.TrimSuffix(spec, "*")
	}
	m := fieldSpecRe.FindStringSubmatch(spec)
	if m == nil {
		return f, false
	}
	f.name, f.typ = m[1], m[2]
	switch {
	case f.typ == "string", f.typ == "int", f.typ == "bool",
		f.typ == "64hex", f.typ == "timestamp", f.typ == "array[string]":
	case strings.HasPrefix(f.typ, "string(") && strings.HasSuffix(f.typ, ")"):
		if _, err := strconv.Atoi(f.typ[len("string(") : len(f.typ)-1]); err != nil {
			return f, false
		}
	case strings.HasPrefix(f.typ, "enum{") && strings.HasSuffix(f.typ, "}"):
	default:
		return f, false
	}
	return f, true
}

// parseStarTable returns the A1-4.4 "*-marked fields per kind" table
// (kind → starred field names, in table order) and the published count of
// kinds with no starred field ("every other kind (N of 42)").
func parseStarTable(t *testing.T) (starred map[string][]string, otherCount int) {
	t.Helper()
	lines := section(t, readContract(t), "The `*`-marked fields per kind", "Tests: TestUntrustedFlagMatchesStarredFields")
	starred = map[string][]string{}
	otherCount = -1
	for _, l := range lines {
		if !isTableRow(l) || isSeparatorRow(l) {
			continue
		}
		if m := otherKindRe.FindStringSubmatch(l); m != nil {
			if !strings.Contains(l, "| none |") {
				t.Fatalf("A1-4.4 'every other kind' row does not say none: %q", l)
			}
			otherCount, _ = strconv.Atoi(m[1])
			continue
		}
		cells := splitCells(l)
		if len(cells) != 2 {
			t.Fatalf("A1-4.4 star row has %d cells, want 2: %q", len(cells), l)
		}
		km := backtickRe.FindStringSubmatch(cells[0])
		if km == nil || !kindNameRe.MatchString(km[1]) {
			continue // the header row ("Kind")
		}
		var fields []string
		for _, fm := range backtickRe.FindAllStringSubmatch(cells[1], -1) {
			fields = append(fields, fm[1])
		}
		if len(fields) == 0 {
			t.Fatalf("A1-4.4 row for %s lists no fields: %q", km[1], l)
		}
		starred[km[1]] = fields
	}
	if otherCount < 0 {
		t.Fatal("A1-4.4 'every other kind (N of 42)' row not found")
	}
	return starred, otherCount
}

// parseComponentList returns the closed A1-2.4 platform-component enum in
// contract order.
func parseComponentList(t *testing.T) []string {
	t.Helper()
	lines := section(t, readContract(t), "- **A1-2.4**", "- **A1-2.5**")
	text := strings.Join(lines, " ")
	start := strings.Index(text, "`event_store`")
	if start < 0 {
		t.Fatal("A1-2.4 component list does not start with `event_store`")
	}
	rest := text[start:]
	end := strings.Index(rest, ". An unknown")
	if end < 0 {
		t.Fatal("A1-2.4 component list has no '. An unknown' terminator")
	}
	var out []string
	for _, m := range backtickRe.FindAllStringSubmatch(rest[:end], -1) {
		out = append(out, m[1])
	}
	if len(out) == 0 {
		t.Fatal("parsed zero A1-2.4 components")
	}
	return out
}

// parseActorTypes returns the closed A1-2.1 actor.type enum in contract
// order.
func parseActorTypes(t *testing.T) []string {
	t.Helper()
	lines := section(t, readContract(t), "- **A1-2.1**", "- **A1-2.2**")
	text := strings.Join(lines, " ")
	start := strings.Index(text, "closed enum:")
	if start < 0 {
		t.Fatal("A1-2.1 has no 'closed enum:' list")
	}
	rest := text[start+len("closed enum:"):]
	end := strings.Index(rest, ")")
	if end < 0 {
		t.Fatal("A1-2.1 closed enum has no ')' terminator")
	}
	var out []string
	for _, m := range backtickRe.FindAllStringSubmatch(rest[:end], -1) {
		out = append(out, m[1])
	}
	if len(out) == 0 {
		t.Fatal("parsed zero A1-2.1 actor types")
	}
	return out
}

// parseActorIDTable returns the A1-2.2 principal-id table as
// actor type → component cell text.
func parseActorIDTable(t *testing.T) map[string]string {
	t.Helper()
	types := parseActorTypes(t)
	known := make(map[string]bool, len(types))
	for _, a := range types {
		known[a] = true
	}
	lines := section(t, readContract(t), "- **A1-2.2**", "AM-1 — resolved and confirmed at the Freeze")
	out := map[string]string{}
	for _, l := range lines {
		if !isTableRow(l) || isSeparatorRow(l) {
			continue
		}
		cells := splitCells(l)
		if len(cells) != 3 {
			t.Fatalf("A1-2.2 row has %d cells, want 3: %q", len(cells), l)
		}
		tm := backtickRe.FindStringSubmatch(cells[0])
		if tm == nil || !known[tm[1]] {
			continue // the header row ("type")
		}
		out[tm[1]] = cells[2]
	}
	if len(out) != len(types) {
		t.Fatalf("A1-2.2 table parsed %d rows, want one per actor type (%d)", len(out), len(types))
	}
	return out
}
