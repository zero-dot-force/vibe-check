package ripwireadapter

import (
	"path/filepath"
	"testing"
)

// --arch metrics-table parser tests — openspec/changes/ripwire-adapter
// task 4.1.
//
// They pin the parser against the recorded ripwire 0.6.9 outputs and prove
// the 40-cap immunity: module counts come from the uncapped <metrics>
// table, never from per-file <inc> lists.

func TestParseArchOutput_RecordedPython(t *testing.T) {
	t.Parallel()

	data := readRecording(t, filepath.Join("testdata", "recorded", "python", "arch.txt"))
	table, err := parseArchOutput(data)
	if err != nil {
		t.Fatalf("parseArchOutput: %v", err)
	}

	if table.Modules != 4 || table.TypedModules != 1 {
		t.Errorf("aggregates: got modules=%d typed_modules=%d, want 4 and 1", table.Modules, table.TypedModules)
	}
	if table.ZonePain != 1 || table.ZoneUseless != 0 || table.ZoneOk != 0 || table.ZoneNa != 3 {
		t.Errorf("zone partition: got pain=%d useless=%d ok=%d na=%d, want 1/0/0/3",
			table.ZonePain, table.ZoneUseless, table.ZoneOk, table.ZoneNa)
	}
	if table.PropagationCost != 0.184 {
		t.Errorf("propagation_cost: got %v, want 0.184", table.PropagationCost)
	}

	want := map[string]archModuleRow{
		"./fixturepy": {
			Path: "./fixturepy", Ca: 0, Ce: 0, Types: 5, Abstract: 0,
			Instability: 0.00, Abstractness: 0.00, Distance: 1.00,
			Zone: "pain", Reachable: true, Isolated: true,
		},
		"./fixturepy/cycle_a": {
			Path: "./fixturepy/cycle_a", Ca: 1, Ce: 1, Types: 0, Abstract: 0,
			Instability: 0.50, Abstractness: 0.00, Distance: 0.50,
			Zone: "n/a", Reachable: false, Isolated: false,
		},
		"./fixturepy/cycle_b": {
			Path: "./fixturepy/cycle_b", Ca: 1, Ce: 1, Types: 0, Abstract: 0,
			Instability: 0.50, Abstractness: 0.00, Distance: 0.50,
			Zone: "n/a", Reachable: false, Isolated: false,
		},
		"./fixturepy/operations": {
			Path: "./fixturepy/operations", Ca: 0, Ce: 0, Types: 0, Abstract: 0,
			Instability: 0.00, Abstractness: 0.00, Distance: 1.00,
			Zone: "n/a", Reachable: true, Isolated: true,
		},
	}
	assertArchRows(t, table, want)
}

func TestParseArchOutput_RecordedTypeScript(t *testing.T) {
	t.Parallel()

	data := readRecording(t, filepath.Join("testdata", "recorded", "typescript", "arch.txt"))
	table, err := parseArchOutput(data)
	if err != nil {
		t.Fatalf("parseArchOutput: %v", err)
	}

	if table.Modules != 4 || table.TypedModules != 1 {
		t.Errorf("aggregates: got modules=%d typed_modules=%d, want 4 and 1", table.Modules, table.TypedModules)
	}
	if table.ZonePain != 0 || table.ZoneUseless != 0 || table.ZoneOk != 1 || table.ZoneNa != 3 {
		t.Errorf("zone partition: got pain=%d useless=%d ok=%d na=%d, want 0/0/1/3",
			table.ZonePain, table.ZoneUseless, table.ZoneOk, table.ZoneNa)
	}
	if table.PropagationCost != 0.556 {
		t.Errorf("propagation_cost: got %v, want 0.556", table.PropagationCost)
	}

	want := map[string]archModuleRow{
		".": {
			Path: ".", Ca: 0, Ce: 0, Types: 0, Abstract: 0,
			Instability: 0.00, Abstractness: 0.00, Distance: 1.00,
			Zone: "n/a", Reachable: true, Isolated: true,
		},
		"./src": {
			Path: "./src", Ca: 0, Ce: 0, Types: 5, Abstract: 4,
			Instability: 0.00, Abstractness: 0.80, Distance: 0.20,
			Zone: "ok", Reachable: true, Isolated: true,
		},
		"./src/cycle-a": {
			Path: "./src/cycle-a", Ca: 1, Ce: 1, Types: 0, Abstract: 0,
			Instability: 0.50, Abstractness: 0.00, Distance: 0.50,
			Zone: "n/a", Reachable: false, Isolated: false,
		},
		"./src/cycle-b": {
			Path: "./src/cycle-b", Ca: 1, Ce: 1, Types: 0, Abstract: 0,
			Instability: 0.50, Abstractness: 0.00, Distance: 0.50,
			Zone: "n/a", Reachable: false, Isolated: false,
		},
	}
	assertArchRows(t, table, want)
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
		if row.Ca != expected.Ca || row.Ce != expected.Ce {
			t.Errorf("%s: ca/ce: got %d/%d, want %d/%d", row.Path, row.Ca, row.Ce, expected.Ca, expected.Ce)
		}
		if row.Types != expected.Types || row.Abstract != expected.Abstract {
			t.Errorf("%s: types/abstract: got %d/%d, want %d/%d", row.Path, row.Types, row.Abstract, expected.Types, expected.Abstract)
		}
		if row.Instability != expected.Instability || row.Abstractness != expected.Abstractness || row.Distance != expected.Distance {
			t.Errorf("%s: I/A/D: got %v/%v/%v, want %v/%v/%v", row.Path,
				row.Instability, row.Abstractness, row.Distance,
				expected.Instability, expected.Abstractness, expected.Distance)
		}
		if row.Zone != expected.Zone {
			t.Errorf("%s: zone: got %q, want %q", row.Path, row.Zone, expected.Zone)
		}
		if row.Reachable != expected.Reachable || row.Isolated != expected.Isolated {
			t.Errorf("%s: reachable/isolated: got %v/%v, want %v/%v", row.Path,
				row.Reachable, row.Isolated, expected.Reachable, expected.Isolated)
		}
	}
}

func TestParseArchOutput_UncappedBeyondForty(t *testing.T) {
	t.Parallel()

	// Counts beyond ripwire's 40-per-file <inc> cap must survive: the
	// table is the aggregation, so 55/77 arrives intact. The synthetic
	// input also carries a long note attribute to prove prose is inert.
	arch := `<arch schema="ripwire.arch/v1" layers="1" rules="0"><metrics modules="1" typed_modules="0" zone_pain="1" zone_useless="0" zone_ok="0" zone_na="0" propagation_cost="9.5" note="long free-text note that must be ignored entirely by the parser"><m path="./big" ca="55" ce="77" types="0" abstract="0" I="0.58" A="0.00" D="0.42" zone="pain" reachable="1" isolated="0"/></metrics></arch>`

	table, err := parseArchOutput([]byte(arch))
	if err != nil {
		t.Fatalf("parseArchOutput: %v", err)
	}
	if len(table.Rows) != 1 {
		t.Fatalf("rows: got %d, want 1", len(table.Rows))
	}
	row := table.Rows[0]
	if row.Ca != 55 || row.Ce != 77 {
		t.Errorf("uncapped counts: got ca=%d ce=%d, want 55 and 77", row.Ca, row.Ce)
	}
	if row.Instability != 0.58 || row.Distance != 0.42 {
		t.Errorf("floats: got I=%v D=%v, want 0.58 and 0.42", row.Instability, row.Distance)
	}
}

func TestParseArchOutput_MalformedInputs(t *testing.T) {
	t.Parallel()

	validRow := `<m path="./x" ca="1" ce="1" types="0" abstract="0" I="0.5" A="0" D="0.5" zone="n/a" reachable="0" isolated="0"/>`
	metrics := func(rows string) string {
		return `<arch schema="ripwire.arch/v1"><metrics modules="1" typed_modules="0" zone_pain="0" zone_useless="0" zone_ok="0" zone_na="1" propagation_cost="0.1" note="n">` + rows + `</metrics></arch>`
	}

	cases := []struct {
		name string
		data string
	}{
		{"empty data", ""},
		{"non-XML garbage", "totally not xml at all"},
		{"wrong root element", `<deps schema="ripwire.deps/v1"/>`},
		{"missing metrics element", `<arch schema="ripwire.arch/v1"/>`},
		{"row missing ca", metrics(`<m path="./x" ce="1" types="0" abstract="0" I="0.5" A="0" D="0.5" zone="n/a"/>`)},
		{"row missing path", metrics(`<m ca="1" ce="1" types="0" abstract="0" I="0.5" A="0" D="0.5" zone="n/a"/>`)},
		{"row missing zone", metrics(`<m path="./x" ca="1" ce="1" types="0" abstract="0" I="0.5" A="0" D="0.5"/>`)},
		{"duplicate module paths", metrics(validRow + validRow)},
		{"non-numeric ca", metrics(`<m path="./x" ca="many" ce="1" types="0" abstract="0" I="0.5" A="0" D="0.5" zone="n/a"/>`)},
		{"metrics missing aggregates", `<arch schema="ripwire.arch/v1"><metrics>` + validRow + `</metrics></arch>`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			table, err := parseArchOutput([]byte(tc.data))
			if err == nil {
				t.Fatalf("got table %+v, want error", table)
			}
		})
	}
}

func TestParseArchOutput_EmptyTableIsValid(t *testing.T) {
	t.Parallel()

	// A present-but-rowless <metrics> element parses to an empty table:
	// an empty crawl is the pipeline's problem to classify (the
	// zero-coupling guard), not a parse failure.
	data := `<arch schema="ripwire.arch/v1"><metrics modules="0" typed_modules="0" zone_pain="0" zone_useless="0" zone_ok="0" zone_na="0" propagation_cost="0.0" note="empty"></metrics></arch>`

	table, err := parseArchOutput([]byte(data))
	if err != nil {
		t.Fatalf("parseArchOutput: %v", err)
	}
	if len(table.Rows) != 0 {
		t.Errorf("rows: got %d, want 0", len(table.Rows))
	}
	if table.Modules != 0 {
		t.Errorf("modules: got %d, want 0", table.Modules)
	}
}
