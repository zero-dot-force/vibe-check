package ripwireadapter

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// --arch metrics-table parser tests — openspec/changes/archive/2026-11-24-ripwire-adapter
// task 4.1, revised by openspec/changes/archive/2026-11-25-bump-ripwire-pin task 2.1.
//
// They pin the parser against the recorded outputs of the pinned ripwire
// release, prove the 40-cap immunity (module counts come from the uncapped
// <metrics> table, never from per-file <inc> lists), and pin the
// consumed-attribute contract of design D2: unconsumed attributes may be
// absent, consumed ones fail closed, and zone="unknown" rows carry floors.

func TestParseArchOutput_RecordedPython(t *testing.T) {
	t.Parallel()

	data := readRecording(t, filepath.Join("testdata", "recorded", "python", "arch.txt"))
	table, err := parseArchOutput(data)
	if err != nil {
		t.Fatalf("parseArchOutput: %v", err)
	}

	want := map[string]archModuleRow{
		"fixturepy":                  {Path: "fixturepy", Ca: 3, Ce: 0, Types: 5, Abstract: 0},
		"fixturepy/alias":            {Path: "fixturepy/alias", Ca: 0, Ce: 1},
		"fixturepy/collide":          {Path: "fixturepy/collide", Ca: 0, Ce: 0},
		"fixturepy/cycle_a":          {Path: "fixturepy/cycle_a", Ca: 2, Ce: 1},
		"fixturepy/cycle_b":          {Path: "fixturepy/cycle_b", Ca: 2, Ce: 1},
		"fixturepy/operations":       {Path: "fixturepy/operations", Ca: 0, Ce: 0},
		"fixturepy/operations/build": {Path: "fixturepy/operations/build", Ca: 0, Ce: 0},
		"fixturepy/rel_a":            {Path: "fixturepy/rel_a", Ca: 1, Ce: 1},
		"fixturepy/rel_b":            {Path: "fixturepy/rel_b", Ca: 1, Ce: 2},
		"fixturepy/ring_a":           {Path: "fixturepy/ring_a", Ca: 1, Ce: 2},
		"fixturepy/ring_b":           {Path: "fixturepy/ring_b", Ca: 1, Ce: 1},
		"fixturepy/wide":             {Path: "fixturepy/wide", Ca: 0, Ce: 2},
	}
	assertArchRows(t, table, want)

	// The pinned release indexes a build/ directory whose files git
	// tracks, and the recording is captured inside this repository's
	// checkout, so nothing is pruned.
	if len(table.PrunedPackages) != 0 || table.PrunedPackageCount != 0 {
		t.Errorf("pruned packages: got %v (count %d), want none (tracked build/ is indexed)",
			table.PrunedPackages, table.PrunedPackageCount)
	}
}

func TestParseArchOutput_RecordedTypeScript(t *testing.T) {
	t.Parallel()

	data := readRecording(t, filepath.Join("testdata", "recorded", "typescript", "arch.txt"))
	table, err := parseArchOutput(data)
	if err != nil {
		t.Fatalf("parseArchOutput: %v", err)
	}

	want := map[string]archModuleRow{
		".":           {Path: "."},
		"src":         {Path: "src", Types: 5, Abstract: 4},
		"src/cycle-a": {Path: "src/cycle-a", Ca: 1, Ce: 1},
		"src/cycle-b": {Path: "src/cycle-b", Ca: 1, Ce: 1},
	}
	assertArchRows(t, table, want)

	if len(table.PrunedPackages) != 0 || table.PrunedPackageCount != 0 {
		t.Errorf("pruned packages: got %v (count %d), want none", table.PrunedPackages, table.PrunedPackageCount)
	}
}

// TestParseArchOutput_RecordedMixedLanguage pins the ARCH-HONEST surface
// (design D2/D3): the module holding a Go file reads zone="unknown"
// no_file_edges="go" with no I/D/reachable, and must parse as a flagged
// floor row instead of failing.
func TestParseArchOutput_RecordedMixedLanguage(t *testing.T) {
	t.Parallel()

	data := readRecording(t, filepath.Join("testdata", "recorded", "python-mixed", "arch.txt"))
	table, err := parseArchOutput(data)
	if err != nil {
		t.Fatalf("parseArchOutput: %v", err)
	}

	var flagged []archModuleRow
	for _, row := range table.Rows {
		if row.CouplingFloor {
			flagged = append(flagged, row)
		}
	}
	want := []archModuleRow{{
		Path: "fixturepy/cycle_a", Ca: 2, Ce: 1,
		CouplingFloor: true, NoFileEdges: []string{"go"},
	}}
	if !reflect.DeepEqual(flagged, want) {
		t.Errorf("flagged rows: got %+v, want %+v", flagged, want)
	}

	// Captured from a temporary copy outside git, so ripwire prunes the
	// build/ package and discloses it.
	if !reflect.DeepEqual(table.PrunedPackages, []string{"fixturepy/operations/build"}) || table.PrunedPackageCount != 1 {
		t.Errorf("pruned packages: got %v (count %d), want [fixturepy/operations/build] (count 1)",
			table.PrunedPackages, table.PrunedPackageCount)
	}
}

func assertArchRows(t *testing.T, table archTable, want map[string]archModuleRow) {
	t.Helper()
	if len(table.Rows) != len(want) {
		t.Fatalf("rows: got %d (%v), want %d", len(table.Rows), table.Rows, len(want))
	}
	for _, row := range table.Rows {
		expected, ok := want[row.Path]
		if !ok {
			t.Errorf("unexpected row path %q", row.Path)
			continue
		}
		if !reflect.DeepEqual(row, expected) {
			t.Errorf("%s: got %+v, want %+v", row.Path, row, expected)
		}
	}
}

func TestParseArchOutput_UncappedBeyondForty(t *testing.T) {
	t.Parallel()

	// Counts beyond ripwire's 40-per-file <inc> cap must survive: the
	// table is the aggregation, so 55/77 arrives intact. The synthetic
	// input also carries a long note attribute to prove prose is inert.
	arch := `<arch schema="ripwire.arch/v1" layers="1" rules="0"><metrics modules="1" note="long free-text note that must be ignored entirely by the parser"><m path="big" ca="55" ce="77" types="0" abstract="0" I="0.58" A="0.00" D="0.42" zone="pain" reachable="1"/></metrics></arch>`

	table, err := parseArchOutput([]byte(arch))
	if err != nil {
		t.Fatalf("parseArchOutput: %v", err)
	}
	if len(table.Rows) != 1 {
		t.Fatalf("rows: got %d, want 1", len(table.Rows))
	}
	if row := table.Rows[0]; row.Ca != 55 || row.Ce != 77 || row.CouplingFloor {
		t.Errorf("uncapped counts: got ca=%d ce=%d floor=%v, want 55, 77, false", row.Ca, row.Ce, row.CouplingFloor)
	}
}

// TestParseArchOutput_ConsumedAttributeContract pins design D2: only the
// consumed attributes are required, unknown rows carry floors, and
// unconsumed attributes or aggregates never cause a failure.
func TestParseArchOutput_ConsumedAttributeContract(t *testing.T) {
	t.Parallel()

	wrap := func(rows string) string {
		return `<arch schema="ripwire.arch/v1"><metrics>` + rows + `</metrics></arch>`
	}
	cases := []struct {
		name string
		data string
		want archModuleRow
	}{
		{
			name: "unconsumed attributes and aggregates absent",
			data: wrap(`<m path="x" ca="1" ce="2" types="3" abstract="1"/>`),
			want: archModuleRow{Path: "x", Ca: 1, Ce: 2, Types: 3, Abstract: 1},
		},
		{
			name: "unknown row without ca/ce is a zero floor",
			data: wrap(`<m path="x" types="0" abstract="0" A="0.00" zone="unknown" no_file_edges="go"/>`),
			want: archModuleRow{Path: "x", CouplingFloor: true, NoFileEdges: []string{"go"}},
		},
		{
			name: "unknown row with floors keeps them",
			data: wrap(`<m path="x" ca="4" ce="5" types="0" abstract="0" zone="unknown" no_file_edges="kt,go,go"/>`),
			want: archModuleRow{Path: "x", Ca: 4, Ce: 5, CouplingFloor: true, NoFileEdges: []string{"go", "kt"}},
		},
		{
			name: "missing zone with ca/ce present",
			data: wrap(`<m path="./x" ca="1" ce="1" types="0" abstract="0"/>`),
			want: archModuleRow{Path: "./x", Ca: 1, Ce: 1},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			table, err := parseArchOutput([]byte(tc.data))
			if err != nil {
				t.Fatalf("parseArchOutput: %v", err)
			}
			if len(table.Rows) != 1 || !reflect.DeepEqual(table.Rows[0], tc.want) {
				t.Errorf("rows: got %+v, want [%+v]", table.Rows, tc.want)
			}
		})
	}
}

func TestParseArchOutput_PrunedPackagesPartialListing(t *testing.T) {
	t.Parallel()

	data := `<arch schema="ripwire.arch/v1"><pruned_packages count="12" shown="2" next="--skipped"><p path="z/build"/><p path="a/dist"/><p path="a/dist"/></pruned_packages><metrics></metrics></arch>`
	table, err := parseArchOutput([]byte(data))
	if err != nil {
		t.Fatalf("parseArchOutput: %v", err)
	}
	if !reflect.DeepEqual(table.PrunedPackages, []string{"a/dist", "z/build"}) || table.PrunedPackageCount != 12 {
		t.Errorf("pruned packages: got %v (count %d), want [a/dist z/build] (count 12)",
			table.PrunedPackages, table.PrunedPackageCount)
	}
}

func TestParseArchOutput_MalformedInputs(t *testing.T) {
	t.Parallel()

	validRow := `<m path="x" ca="1" ce="1" types="0" abstract="0" zone="n/a"/>`
	metrics := func(rows string) string {
		return `<arch schema="ripwire.arch/v1"><metrics>` + rows + `</metrics></arch>`
	}

	cases := []struct {
		name    string
		data    string
		wantSub string
	}{
		{"empty data", "", "parse arch output"},
		{"non-XML garbage", "totally not xml at all", "parse arch output"},
		{"wrong root element", `<deps schema="ripwire.deps/v1"/>`, "parse arch output"},
		{"missing metrics element", `<arch schema="ripwire.arch/v1"/>`, "no <metrics>"},
		{"row missing ca", metrics(`<m path="x" ce="1" types="0" abstract="0" zone="n/a"/>`), `"x" missing ca`},
		{"row missing ce", metrics(`<m path="x" ca="1" types="0" abstract="0" zone="pain"/>`), `"x" missing ce`},
		{"row missing ca and zone", metrics(`<m path="x" ce="1" types="0" abstract="0"/>`), `"x" missing ca`},
		{"row missing path", metrics(`<m ca="1" ce="1" types="0" abstract="0" zone="n/a"/>`), "missing path"},
		{"row missing types", metrics(`<m path="x" ca="1" ce="1" abstract="0" zone="n/a"/>`), `"x" missing types`},
		{"unknown row missing abstract", metrics(`<m path="x" types="0" zone="unknown"/>`), `"x" missing abstract`},
		{"duplicate module paths", metrics(validRow + validRow), "duplicate module path"},
		{"non-numeric ca", metrics(`<m path="x" ca="many" ce="1" types="0" abstract="0" zone="n/a"/>`), "parse arch output"},
		{"pruned packages missing count", `<arch><pruned_packages><p path="a/build"/></pruned_packages><metrics/></arch>`, "missing count"},
		{"pruned row missing path", `<arch><pruned_packages count="1"><p/></pruned_packages><metrics/></arch>`, "row missing path"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			table, err := parseArchOutput([]byte(tc.data))
			if err == nil {
				t.Fatalf("got table %+v, want error", table)
			}
			if !strings.Contains(err.Error(), tc.wantSub) {
				t.Errorf("error: got %q, want it to contain %q", err, tc.wantSub)
			}
		})
	}
}

func TestParseArchOutput_EmptyTableIsValid(t *testing.T) {
	t.Parallel()

	// A present-but-rowless <metrics> element parses to an empty table:
	// an empty crawl is the pipeline's problem to classify (the
	// zero-coupling guard), not a parse failure.
	data := `<arch schema="ripwire.arch/v1"><metrics modules="0" note="empty"></metrics></arch>`

	table, err := parseArchOutput([]byte(data))
	if err != nil {
		t.Fatalf("parseArchOutput: %v", err)
	}
	if len(table.Rows) != 0 {
		t.Errorf("rows: got %d, want 0", len(table.Rows))
	}
}
