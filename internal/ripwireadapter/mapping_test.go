package ripwireadapter

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/zero-dot-force/vibe-check/metrics"
)

// Directory-module mapping tests — openspec/changes/ripwire-adapter
// task 4.3. They pin D4 identity, the authoritative Ca/Ce passthrough,
// the deduplicated directory-pair edge set, Ce-only externals, and the
// three mapping warnings.

func mapRecordedFixture(t *testing.T, language string) moduleMapping {
	t.Helper()
	recordings := filepath.Join("testdata", "recorded", language)
	archData := readRecording(t, filepath.Join(recordings, "arch.txt"))
	depsData := readRecording(t, filepath.Join(recordings, "deps.xml"))
	table, err := parseArchOutput(archData)
	if err != nil {
		t.Fatalf("parseArchOutput: %v", err)
	}
	deps, err := parseDepsOutput(depsData)
	if err != nil {
		t.Fatalf("parseDepsOutput: %v", err)
	}
	mapping, err := mapToModules(language, table, deps)
	if err != nil {
		t.Fatalf("mapToModules: %v", err)
	}
	return mapping
}

func warningCodes(ws []metrics.Warning) []string {
	codes := make([]string, 0, len(ws))
	for _, w := range ws {
		codes = append(codes, w.Code)
	}
	return codes
}

func assertHasCodes(t *testing.T, ws []metrics.Warning, wantPresent, wantAbsent []string) {
	t.Helper()
	codes := warningCodes(ws)
	for _, want := range wantPresent {
		found := false
		for _, c := range codes {
			if c == want {
				found = true
			}
		}
		if !found {
			t.Errorf("warnings %v missing code %q", codes, want)
		}
	}
	for _, unwanted := range wantAbsent {
		for _, c := range codes {
			if c == unwanted {
				t.Errorf("warnings %v unexpectedly contain code %q", codes, unwanted)
			}
		}
	}
}

func moduleByIdentity(t *testing.T, mapping moduleMapping, identity string) mappedModule {
	t.Helper()
	for _, m := range mapping.Modules {
		if m.Identity == identity {
			return m
		}
	}
	t.Fatalf("module %q not found in %v", identity, mapping.Modules)
	return mappedModule{}
}

func TestMapToModules_RecordedPython(t *testing.T) {
	t.Parallel()

	mapping := mapRecordedFixture(t, "python")

	want := map[string]mappedModule{
		"fixturepy":            {Identity: "fixturepy", Name: "fixturepy", Ca: 0, Ce: 0, Types: 5, Abstract: 0, ExternalOut: 1},
		"fixturepy.cycle_a":    {Identity: "fixturepy.cycle_a", Name: "cycle_a", Ca: 1, Ce: 1, InternalOut: 1, InternalIn: 1},
		"fixturepy.cycle_b":    {Identity: "fixturepy.cycle_b", Name: "cycle_b", Ca: 1, Ce: 1, InternalOut: 1, InternalIn: 1},
		"fixturepy.operations": {Identity: "fixturepy.operations", Name: "operations", Ca: 0, Ce: 0},
	}
	if len(mapping.Modules) != len(want) {
		t.Fatalf("modules: got %d, want %d", len(mapping.Modules), len(want))
	}
	for identity, expected := range want {
		got := moduleByIdentity(t, mapping, identity)
		if got.Ca != expected.Ca || got.Ce != expected.Ce {
			t.Errorf("%s: ca/ce: got %d/%d, want %d/%d", identity, got.Ca, got.Ce, expected.Ca, expected.Ce)
		}
		if got.Types != expected.Types || got.Abstract != expected.Abstract {
			t.Errorf("%s: types/abstract: got %d/%d, want %d/%d", identity, got.Types, got.Abstract, expected.Types, expected.Abstract)
		}
		if got.InternalOut != expected.InternalOut || got.InternalIn != expected.InternalIn {
			t.Errorf("%s: internal out/in: got %d/%d, want %d/%d", identity, got.InternalOut, got.InternalIn, expected.InternalOut, expected.InternalIn)
		}
		if got.ExternalOut != expected.ExternalOut {
			t.Errorf("%s: external out: got %d, want %d", identity, got.ExternalOut, expected.ExternalOut)
		}
		if got.Name != expected.Name {
			t.Errorf("%s: name: got %q, want %q", identity, got.Name, expected.Name)
		}
	}

	wantEdges := []moduleEdge{
		{From: "fixturepy.cycle_a", To: "fixturepy.cycle_b"},
		{From: "fixturepy.cycle_b", To: "fixturepy.cycle_a"},
	}
	if !reflect.DeepEqual(mapping.Edges, wantEdges) {
		t.Errorf("edges: got %v, want %v", mapping.Edges, wantEdges)
	}

	assertHasCodes(t, mapping.Warnings,
		[]string{"directory-as-module", "name-resolved-edges"},
		[]string{"ambiguity-rate"})
}

func TestMapToModules_RecordedTypeScript(t *testing.T) {
	t.Parallel()

	mapping := mapRecordedFixture(t, "typescript")

	if len(mapping.Modules) != 4 {
		t.Fatalf("modules: got %d, want 4", len(mapping.Modules))
	}
	src := moduleByIdentity(t, mapping, "src")
	if src.Ca != 0 || src.Ce != 0 || src.Types != 5 || src.Abstract != 4 {
		t.Errorf("src: ca/ce/types/abstract: got %d/%d/%d/%d, want 0/0/5/4", src.Ca, src.Ce, src.Types, src.Abstract)
	}
	if src.Name != "src" {
		t.Errorf("src name: got %q, want src", src.Name)
	}
	root := moduleByIdentity(t, mapping, ".")
	if root.Name != "." {
		t.Errorf("root name: got %q, want %q", root.Name, ".")
	}

	wantEdges := []moduleEdge{
		{From: "src/cycle-a", To: "src/cycle-b"},
		{From: "src/cycle-b", To: "src/cycle-a"},
	}
	if !reflect.DeepEqual(mapping.Edges, wantEdges) {
		t.Errorf("edges: got %v, want %v", mapping.Edges, wantEdges)
	}

	for _, m := range mapping.Modules {
		if m.ExternalOut != 0 {
			t.Errorf("%s: external out: got %d, want 0", m.Identity, m.ExternalOut)
		}
	}

	assertHasCodes(t, mapping.Warnings,
		[]string{"directory-as-module"},
		[]string{"name-resolved-edges", "ambiguity-rate"})
}

func TestMapToModules_Deterministic(t *testing.T) {
	t.Parallel()

	for _, language := range []string{"python", "typescript"} {
		t.Run(language, func(t *testing.T) {
			first := mapRecordedFixture(t, language)
			second := mapRecordedFixture(t, language)
			if !reflect.DeepEqual(first, second) {
				t.Errorf("mapping the same recorded output twice yielded different results")
			}
		})
	}
}

func syntheticArch(rows ...archModuleRow) archTable {
	table := archTable{
		Modules: len(rows),
		Rows:    rows,
	}
	for _, r := range rows {
		if r.Types > 0 {
			table.TypedModules++
		}
	}
	return table
}

func TestMapToModules_AmbiguityRate(t *testing.T) {
	t.Parallel()

	// A TypeScript target landing inside the crawl but matching no module
	// directory: the edge is dropped and ambiguity-rate fires.
	table := syntheticArch(
		archModuleRow{Path: "./src", Ca: 0, Ce: 1, Zone: "n/a"},
	)
	deps := depsGraph{
		Rows: []depsFile{
			{Path: "src/index.ts", IncTargets: []string{"./orphan-dir"}},
		},
	}
	mapping, err := mapToModules(languageTypeScript, table, deps)
	if err != nil {
		t.Fatalf("mapToModules: %v", err)
	}
	if len(mapping.Edges) != 0 {
		t.Errorf("edges: got %v, want none (unresolved dropped)", mapping.Edges)
	}
	assertHasCodes(t, mapping.Warnings, []string{"ambiguity-rate"}, nil)
}

func TestMapToModules_PythonAllExternalNoNameResolution(t *testing.T) {
	t.Parallel()

	// A tree with no internal edges at all: no name-resolved-edges
	// warning (nothing was resolved), externals counted in Ce only.
	table := syntheticArch(
		archModuleRow{Path: "./pkg", Ca: 0, Ce: 0, Zone: "n/a"},
	)
	deps := depsGraph{
		Rows: []depsFile{
			{Path: "pkg/mod.py", IncTargets: []string{"os", "sys"}},
		},
	}
	mapping, err := mapToModules(languagePython, table, deps)
	if err != nil {
		t.Fatalf("mapToModules: %v", err)
	}
	if got := moduleByIdentity(t, mapping, "pkg").ExternalOut; got != 2 {
		t.Errorf("external out: got %d, want 2", got)
	}
	assertHasCodes(t, mapping.Warnings,
		[]string{"directory-as-module"},
		[]string{"name-resolved-edges", "ambiguity-rate"})
}

func TestMapToModules_Errors(t *testing.T) {
	t.Parallel()

	t.Run("empty arch table", func(t *testing.T) {
		if _, err := mapToModules(languagePython, archTable{}, depsGraph{}); err == nil {
			t.Fatal("got no error, want error for empty arch table")
		}
	})

	t.Run("duplicate identity", func(t *testing.T) {
		table := syntheticArch(
			archModuleRow{Path: "./a/b", Zone: "n/a"},
			archModuleRow{Path: "./a.b", Zone: "n/a"},
		)
		if _, err := mapToModules(languagePython, table, depsGraph{}); err == nil {
			t.Fatal("got no error, want error for duplicate identity a.b")
		}
	})

	t.Run("deps file outside modules", func(t *testing.T) {
		table := syntheticArch(archModuleRow{Path: "./pkg", Zone: "n/a"})
		deps := depsGraph{Rows: []depsFile{{Path: "elsewhere/mod.py"}}}
		if _, err := mapToModules(languagePython, table, deps); err == nil {
			t.Fatal("got no error, want error for file outside module universe")
		}
	})
}

func TestModuleName(t *testing.T) {
	t.Parallel()

	cases := []struct{ identity, want string }{
		{"fixturepy.cycle_a", "cycle_a"},
		{"fixturepy", "fixturepy"},
		{"src/cycle-a", "cycle-a"},
		{"src", "src"},
		{".", "."},
	}
	for _, tc := range cases {
		if got := moduleName(tc.identity); got != tc.want {
			t.Errorf("moduleName(%q): got %q, want %q", tc.identity, got, tc.want)
		}
	}
}
