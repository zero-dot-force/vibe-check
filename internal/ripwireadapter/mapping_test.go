package ripwireadapter

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/zero-dot-force/vibe-check/metrics"
)

// Directory-module mapping tests — openspec/changes/archive/2026-11-24-ripwire-adapter
// task 4.3, extended by openspec/changes/archive/2026-11-25-bump-ripwire-pin tasks 2.2 and
// 3.1–3.2. They pin D4 identity, the authoritative Ca/Ce passthrough,
// the deduplicated directory-pair edge set, Ce-only externals, the
// mapping warnings, both arch path forms, and coupling floors.

func mapRecordedFixture(t *testing.T, language string) moduleMapping {
	t.Helper()
	return mapRecording(t, language, language)
}

// mapRecording maps the recorded corpus (a directory under
// testdata/recorded) with the given language's identity rules.
func mapRecording(t *testing.T, corpus, language string) moduleMapping {
	t.Helper()
	recordings := filepath.Join("testdata", "recorded", corpus)
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
	mapping, err := mapToModules(language, table, deps, fixtureMemberImports(t, language))
	if err != nil {
		t.Fatalf("mapToModules: %v", err)
	}
	return mapping
}

// fixtureMemberImports returns the member-import light parse of the
// committed Python fixture tree (the tree every Python recording was
// captured from or replays over), or nil for other languages.
func fixtureMemberImports(t *testing.T, language string) map[string][]memberImport {
	t.Helper()
	if language != languagePython {
		return nil
	}
	members, warnings, err := pythonMemberImportScan(filepath.Join("testdata", "python", "src"))
	if err != nil {
		t.Fatalf("pythonMemberImportScan: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("pythonMemberImportScan warnings: %v", warnings)
	}
	return members
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

	// Ca/Ce come from the --arch rows (ripwire's resolved edges); the
	// Internal*/External* counters come from the adapter's own edge set:
	// name-resolved --deps <inc t=> targets plus the member edges of
	// wide-form imports. --deps records `from fixturepy import cycle_b` as
	// "fixturepy" alone; the member-import light parse recovers cycle_b
	// (anchored to that include), so wide's InternalOut equals its Ce=2 and
	// cycle_b's InternalIn equals its Ca=2 (openspec/changes/
	// wide-form-cycle-edges). Relative targets (rel_a's "..rel_b.api",
	// rel_b's "..") resolve against the importing package and are not
	// external; rel_b's member edge to rel_a comes from `from .. import
	// rel_a` (openspec/changes/archive/2026-11-24-relative-import-edges). On this fixture
	// every module's internal counters agree with its arch row.
	want := map[string]mappedModule{
		"fixturepy":                  {Identity: "fixturepy", Name: "fixturepy", Ca: 3, Ce: 0, Types: 5, Abstract: 0, InternalIn: 3, ExternalOut: 1},
		"fixturepy.alias":            {Identity: "fixturepy.alias", Name: "alias", Ca: 0, Ce: 1, InternalOut: 1},
		"fixturepy.collide":          {Identity: "fixturepy.collide", Name: "collide", Ca: 0, Ce: 0, ExternalOut: 1},
		"fixturepy.cycle_a":          {Identity: "fixturepy.cycle_a", Name: "cycle_a", Ca: 2, Ce: 1, InternalOut: 1, InternalIn: 2},
		"fixturepy.cycle_b":          {Identity: "fixturepy.cycle_b", Name: "cycle_b", Ca: 2, Ce: 1, InternalOut: 1, InternalIn: 2},
		"fixturepy.operations":       {Identity: "fixturepy.operations", Name: "operations", Ca: 0, Ce: 0},
		"fixturepy.operations.build": {Identity: "fixturepy.operations.build", Name: "build", Ca: 0, Ce: 0},
		"fixturepy.rel_a":            {Identity: "fixturepy.rel_a", Name: "rel_a", Ca: 1, Ce: 1, InternalOut: 1, InternalIn: 1},
		"fixturepy.rel_b":            {Identity: "fixturepy.rel_b", Name: "rel_b", Ca: 1, Ce: 2, InternalOut: 2, InternalIn: 1},
		"fixturepy.ring_a":           {Identity: "fixturepy.ring_a", Name: "ring_a", Ca: 1, Ce: 2, InternalOut: 2, InternalIn: 1},
		"fixturepy.ring_b":           {Identity: "fixturepy.ring_b", Name: "ring_b", Ca: 1, Ce: 1, InternalOut: 1, InternalIn: 1},
		"fixturepy.wide":             {Identity: "fixturepy.wide", Name: "wide", Ca: 0, Ce: 2, InternalOut: 2},
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
		{From: "fixturepy.alias", To: "fixturepy.cycle_a"},
		{From: "fixturepy.cycle_a", To: "fixturepy.cycle_b"},
		{From: "fixturepy.cycle_b", To: "fixturepy.cycle_a"},
		{From: "fixturepy.rel_a", To: "fixturepy.rel_b"},
		{From: "fixturepy.rel_b", To: "fixturepy"},
		{From: "fixturepy.rel_b", To: "fixturepy.rel_a"},
		{From: "fixturepy.ring_a", To: "fixturepy"},
		{From: "fixturepy.ring_a", To: "fixturepy.ring_b"},
		{From: "fixturepy.ring_b", To: "fixturepy.ring_a"},
		{From: "fixturepy.wide", To: "fixturepy"},
		{From: "fixturepy.wide", To: "fixturepy.cycle_b"},
	}
	if !reflect.DeepEqual(mapping.Edges, wantEdges) {
		t.Errorf("edges: got %v, want %v", mapping.Edges, wantEdges)
	}

	assertHasCodes(t, mapping.Warnings,
		[]string{"directory-as-module", "name-resolved-edges"},
		[]string{"ambiguity-rate", warningCouplingFloor})
}

// TestMapToModules_RecordedMixedLanguage pins design D3 on the ARCH-HONEST
// recording: only the module holding the Go file is flagged, its floors
// pass through, and one coupling-floor warning names it and the language.
func TestMapToModules_RecordedMixedLanguage(t *testing.T) {
	t.Parallel()

	mapping := mapRecording(t, "python-mixed", languagePython)
	for _, m := range mapping.Modules {
		wantFlag := m.Identity == "fixturepy.cycle_a"
		if m.CouplingFloor != wantFlag {
			t.Errorf("%s: coupling floor: got %v, want %v", m.Identity, m.CouplingFloor, wantFlag)
		}
	}
	cycleA := moduleByIdentity(t, mapping, "fixturepy.cycle_a")
	if cycleA.Ca != 2 || cycleA.Ce != 1 || !reflect.DeepEqual(cycleA.NoFileEdges, []string{"go"}) {
		t.Errorf("cycle_a: got ca=%d ce=%d langs=%v, want 2, 1, [go]", cycleA.Ca, cycleA.Ce, cycleA.NoFileEdges)
	}

	var floors []string
	for _, w := range mapping.Warnings {
		if w.Code == warningCouplingFloor {
			floors = append(floors, w.Message)
		}
	}
	want := []string{"ripwire resolves almost no imports for go; coupling for 1 module(s) is a floor, not an exact count: fixturepy.cycle_a"}
	if !reflect.DeepEqual(floors, want) {
		t.Errorf("coupling-floor warnings: got %q, want %q", floors, want)
	}
}

func TestCouplingFloorWarning(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		modules []mappedModule
		want    string // "" means no warning
	}{
		{
			name:    "no flagged module",
			modules: []mappedModule{{Identity: "a"}, {Identity: "b"}},
		},
		{
			name: "identities and languages sorted and de-duplicated",
			modules: []mappedModule{
				{Identity: "z", CouplingFloor: true, NoFileEdges: []string{"kt"}},
				{Identity: "m"},
				{Identity: "a", CouplingFloor: true, NoFileEdges: []string{"go", "kt"}},
			},
			want: "ripwire resolves almost no imports for go, kt; coupling for 2 module(s) is a floor, not an exact count: a, z",
		},
		{
			name:    "flagged without listed languages",
			modules: []mappedModule{{Identity: "x", CouplingFloor: true}},
			want:    "ripwire resolves almost no imports for unlisted languages; coupling for 1 module(s) is a floor, not an exact count: x",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := couplingFloorWarning(tc.modules)
			if tc.want == "" {
				if ok {
					t.Fatalf("got warning %+v, want none", got)
				}
				return
			}
			if !ok || got.Code != warningCouplingFloor || got.Message != tc.want {
				t.Errorf("got %+v (ok=%v), want %s %q", got, ok, warningCouplingFloor, tc.want)
			}
		})
	}
}

// TestMapToModules_ArchPathForms pins task 2.2: ripwire prints row paths
// with a leading "./" up to the previous pin and without it since; both forms (and
// the root row ".") must map to the same identities.
func TestMapToModules_ArchPathForms(t *testing.T) {
	t.Parallel()

	cases := []struct {
		language string
		rows     []string
		want     []string
	}{
		{languagePython, []string{".", "./fixturepy/cycle_a", "fixturepy/cycle_b"}, []string{".", "fixturepy.cycle_a", "fixturepy.cycle_b"}},
		{languageTypeScript, []string{".", "./src/cycle-a", "src/cycle-b"}, []string{".", "src/cycle-a", "src/cycle-b"}},
	}
	for _, tc := range cases {
		t.Run(tc.language, func(t *testing.T) {
			for _, prefixed := range []bool{false, true} {
				rows := make([]archModuleRow, 0, len(tc.rows))
				for _, p := range tc.rows {
					if prefixed && p != "." && !strings.HasPrefix(p, "./") {
						p = "./" + p
					}
					rows = append(rows, archModuleRow{Path: p})
				}
				mapping, err := mapToModules(tc.language, syntheticArch(rows...), depsGraph{}, nil)
				if err != nil {
					t.Fatalf("mapToModules: %v", err)
				}
				got := make([]string, 0, len(mapping.Modules))
				for _, m := range mapping.Modules {
					got = append(got, m.Identity)
				}
				if !reflect.DeepEqual(got, tc.want) {
					t.Errorf("prefixed=%v: identities got %v, want %v", prefixed, got, tc.want)
				}
			}
		})
	}
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
	return archTable{Rows: rows}
}

func TestMapToModules_AmbiguityRate(t *testing.T) {
	t.Parallel()

	// A TypeScript target landing inside the crawl but matching no module
	// directory: the edge is dropped and ambiguity-rate fires.
	table := syntheticArch(
		archModuleRow{Path: "./src", Ca: 0, Ce: 1},
	)
	deps := depsGraph{
		Rows: []depsFile{
			{Path: "src/index.ts", IncTargets: []string{"./orphan-dir"}},
		},
	}
	mapping, err := mapToModules(languageTypeScript, table, deps, nil)
	if err != nil {
		t.Fatalf("mapToModules: %v", err)
	}
	if len(mapping.Edges) != 0 {
		t.Fatalf("got %d resolved edges, want 0 (the only include target is ambiguous)", len(mapping.Edges))
	}
	assertHasCodes(t, mapping.Warnings, []string{warningAmbiguityRate}, nil)
}

// TestMapToModules_WarningProse pins the disclosure prose of the three
// mapping warnings. The code-level tests above prove which warnings fire;
// this test proves the messages still carry their load-bearing disclosure
// clauses, so a rewording that drops the disclosure cannot pass silently.
func TestMapToModules_WarningProse(t *testing.T) {
	findMessage := func(t *testing.T, ws []metrics.Warning, code string) string {
		t.Helper()
		for _, w := range ws {
			if w.Code == code {
				return w.Message
			}
		}
		t.Fatalf("warning %q not present in %v", code, warningCodes(ws))
		return ""
	}

	t.Run("directory as module", func(t *testing.T) {
		mapping := mapRecordedFixture(t, languagePython)
		msg := findMessage(t, mapping.Warnings, warningDirectoryAsModule)
		if !strings.Contains(msg, "invisible to the graph") {
			t.Errorf("directory-as-module message = %q, want disclosure that intra-directory coupling is invisible to the graph", msg)
		}
	})

	t.Run("name resolved edges", func(t *testing.T) {
		mapping := mapRecordedFixture(t, languagePython)
		msg := findMessage(t, mapping.Warnings, warningNameResolvedEdges)
		if !strings.Contains(msg, "misattribute") {
			t.Errorf("name-resolved-edges message = %q, want disclosure that name-based resolution can misattribute", msg)
		}
	})

	t.Run("ambiguity rate", func(t *testing.T) {
		table := syntheticArch(archModuleRow{Path: "./src", Ca: 0, Ce: 1})
		deps := depsGraph{Rows: []depsFile{{Path: "src/index.ts", IncTargets: []string{"./orphan-dir"}}}}
		mapping, err := mapToModules(languageTypeScript, table, deps, nil)
		if err != nil {
			t.Fatalf("mapToModules: %v", err)
		}
		msg := findMessage(t, mapping.Warnings, warningAmbiguityRate)
		if !strings.Contains(msg, "edges were dropped") {
			t.Errorf("ambiguity-rate message = %q, want disclosure that unmatched in-crawl edges were dropped", msg)
		}
	})
}

func TestMapToModules_PythonAllExternalNoNameResolution(t *testing.T) {
	t.Parallel()

	// A tree with no internal edges at all: no name-resolved-edges
	// warning (nothing was resolved), externals counted in Ce only.
	table := syntheticArch(
		archModuleRow{Path: "./pkg", Ca: 0, Ce: 0},
	)
	deps := depsGraph{
		Rows: []depsFile{
			{Path: "pkg/mod.py", IncTargets: []string{"os", "sys"}},
		},
	}
	mapping, err := mapToModules(languagePython, table, deps, nil)
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
		if _, err := mapToModules(languagePython, archTable{}, depsGraph{}, nil); err == nil {
			t.Fatal("got no error, want error for empty arch table")
		}
	})

	t.Run("duplicate identity", func(t *testing.T) {
		table := syntheticArch(
			archModuleRow{Path: "./a/b"},
			archModuleRow{Path: "./a.b"},
		)
		if _, err := mapToModules(languagePython, table, depsGraph{}, nil); err == nil {
			t.Fatal("got no error, want error for duplicate identity a.b")
		}
	})

	t.Run("deps file outside modules", func(t *testing.T) {
		table := syntheticArch(archModuleRow{Path: "./pkg"})
		deps := depsGraph{Rows: []depsFile{{Path: "elsewhere/mod.py"}}}
		if _, err := mapToModules(languagePython, table, deps, nil); err == nil {
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

// memberEdgeArch is the module universe shared by the member-edge tests:
// a package root with two subpackages.
func memberEdgeArch() archTable {
	return syntheticArch(
		archModuleRow{Path: "pkg", Ca: 1, Ce: 0},
		archModuleRow{Path: "pkg/a", Ca: 0, Ce: 2},
		archModuleRow{Path: "pkg/b", Ca: 1, Ce: 0},
	)
}

func TestMapToModules_MemberEdgeAnchored(t *testing.T) {
	t.Parallel()

	deps := depsGraph{Rows: []depsFile{{Path: "pkg/a/mod.py", IncTargets: []string{"pkg"}}}}
	members := map[string][]memberImport{"pkg/a/mod.py": {{Base: "pkg", Name: "b"}}}
	mapping, err := mapToModules(languagePython, memberEdgeArch(), deps, members)
	if err != nil {
		t.Fatalf("mapToModules: %v", err)
	}
	wantEdges := []moduleEdge{{From: "pkg.a", To: "pkg"}, {From: "pkg.a", To: "pkg.b"}}
	if !reflect.DeepEqual(mapping.Edges, wantEdges) {
		t.Errorf("edges: got %v, want %v", mapping.Edges, wantEdges)
	}
	if a := moduleByIdentity(t, mapping, "pkg.a"); a.InternalOut != 2 {
		t.Errorf("pkg.a internal out: got %d, want 2", a.InternalOut)
	}
	if b := moduleByIdentity(t, mapping, "pkg.b"); b.InternalIn != 1 {
		t.Errorf("pkg.b internal in: got %d, want 1", b.InternalIn)
	}
}

func TestMapToModules_MemberEdgeUnanchored(t *testing.T) {
	t.Parallel()

	// The light parse found `from pkg import b`, but ripwire recorded no
	// include target "pkg" for the file (e.g. the text sat somewhere the
	// parse could not exclude): no member edge.
	deps := depsGraph{Rows: []depsFile{{Path: "pkg/a/mod.py", IncTargets: []string{"os"}}}}
	members := map[string][]memberImport{"pkg/a/mod.py": {{Base: "pkg", Name: "b"}}}
	mapping, err := mapToModules(languagePython, memberEdgeArch(), deps, members)
	if err != nil {
		t.Fatalf("mapToModules: %v", err)
	}
	if len(mapping.Edges) != 0 {
		t.Errorf("edges: got %v, want none", mapping.Edges)
	}
}

func TestMapToModules_MemberEdgeFileNotInDeps(t *testing.T) {
	t.Parallel()

	// A scanned file absent from the --deps window (capped or pruned)
	// contributes nothing.
	deps := depsGraph{Rows: []depsFile{{Path: "pkg/a/other.py", IncTargets: []string{"pkg"}}}}
	members := map[string][]memberImport{"pkg/a/mod.py": {{Base: "pkg", Name: "b"}}}
	mapping, err := mapToModules(languagePython, memberEdgeArch(), deps, members)
	if err != nil {
		t.Fatalf("mapToModules: %v", err)
	}
	wantEdges := []moduleEdge{{From: "pkg.a", To: "pkg"}}
	if !reflect.DeepEqual(mapping.Edges, wantEdges) {
		t.Errorf("edges: got %v, want %v", mapping.Edges, wantEdges)
	}
}

func TestMapToModules_MemberEdgeSymbolOrFile(t *testing.T) {
	t.Parallel()

	// `from pkg import helper` (a symbol or pkg/helper.py) and
	// `from pkg.b import thing` resolve to no module directory: only the
	// name-resolved base edges remain.
	deps := depsGraph{Rows: []depsFile{{Path: "pkg/a/mod.py", IncTargets: []string{"pkg", "pkg.b"}}}}
	members := map[string][]memberImport{"pkg/a/mod.py": {
		{Base: "pkg", Name: "helper"},
		{Base: "pkg.b", Name: "thing"},
	}}
	mapping, err := mapToModules(languagePython, memberEdgeArch(), deps, members)
	if err != nil {
		t.Fatalf("mapToModules: %v", err)
	}
	wantEdges := []moduleEdge{{From: "pkg.a", To: "pkg"}, {From: "pkg.a", To: "pkg.b"}}
	if !reflect.DeepEqual(mapping.Edges, wantEdges) {
		t.Errorf("edges: got %v, want %v", mapping.Edges, wantEdges)
	}
}

func TestMapToModules_MemberEdgeDedup(t *testing.T) {
	t.Parallel()

	// The member edge pkg.a → pkg.b duplicates a name-resolved edge from
	// a sibling file: it is counted once.
	deps := depsGraph{Rows: []depsFile{
		{Path: "pkg/a/first.py", IncTargets: []string{"pkg.b.api"}},
		{Path: "pkg/a/second.py", IncTargets: []string{"pkg"}},
	}}
	members := map[string][]memberImport{"pkg/a/second.py": {{Base: "pkg", Name: "b"}}}
	mapping, err := mapToModules(languagePython, memberEdgeArch(), deps, members)
	if err != nil {
		t.Fatalf("mapToModules: %v", err)
	}
	wantEdges := []moduleEdge{{From: "pkg.a", To: "pkg"}, {From: "pkg.a", To: "pkg.b"}}
	if !reflect.DeepEqual(mapping.Edges, wantEdges) {
		t.Errorf("edges: got %v, want %v", mapping.Edges, wantEdges)
	}
	if a := moduleByIdentity(t, mapping, "pkg.a"); a.InternalOut != 2 {
		t.Errorf("pkg.a internal out: got %d, want 2", a.InternalOut)
	}
	if b := moduleByIdentity(t, mapping, "pkg.b"); b.InternalIn != 1 {
		t.Errorf("pkg.b internal in: got %d, want 1", b.InternalIn)
	}
}

func TestMapToModules_MemberEdgeCaCeUnchanged(t *testing.T) {
	t.Parallel()

	// Member edges never touch the authoritative arch values, even when
	// the arch row disagrees with the edge set.
	table := syntheticArch(
		archModuleRow{Path: "pkg", Ca: 7, Ce: 3},
		archModuleRow{Path: "pkg/a", Ca: 5, Ce: 0},
		archModuleRow{Path: "pkg/b", Ca: 0, Ce: 9},
	)
	deps := depsGraph{Rows: []depsFile{{Path: "pkg/a/mod.py", IncTargets: []string{"pkg"}}}}
	members := map[string][]memberImport{"pkg/a/mod.py": {{Base: "pkg", Name: "b"}}}
	mapping, err := mapToModules(languagePython, table, deps, members)
	if err != nil {
		t.Fatalf("mapToModules: %v", err)
	}
	want := map[string][2]int{"pkg": {7, 3}, "pkg.a": {5, 0}, "pkg.b": {0, 9}}
	for identity, caCe := range want {
		m := moduleByIdentity(t, mapping, identity)
		if m.Ca != caCe[0] || m.Ce != caCe[1] {
			t.Errorf("%s ca/ce: got %d/%d, want %d/%d", identity, m.Ca, m.Ce, caCe[0], caCe[1])
		}
	}
}

func TestMapToModules_MemberEdgeDeterministic(t *testing.T) {
	t.Parallel()

	first := mapRecordedFixture(t, "python")
	second := mapRecordedFixture(t, "python")
	if !reflect.DeepEqual(first, second) {
		t.Errorf("mapping differs across runs:\nfirst  %+v\nsecond %+v", first, second)
	}
}
