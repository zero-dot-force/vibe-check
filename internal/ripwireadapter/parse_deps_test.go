package ripwireadapter

import (
	"path/filepath"
	"testing"
)

// --deps and cc.json parser tests — openspec/changes/ripwire-adapter
// task 4.2. They pin the parsers against the recorded ripwire 0.6.9
// outputs: exact edge sets, fan-in/out values, cycles listing, and paging
// attributes.

func TestParseDepsOutput_RecordedPython(t *testing.T) {
	t.Parallel()

	data := readRecording(t, filepath.Join("testdata", "recorded", "python", "deps.xml"))
	graph, err := parseDepsOutput(data)
	if err != nil {
		t.Fatalf("parseDepsOutput: %v", err)
	}

	if graph.Files != 3 || graph.Shown != 3 || graph.Total != 3 || graph.Capped != 0 || graph.HasMore != 0 {
		t.Errorf("paging: got files=%d shown=%d total=%d capped=%d has_more=%d, want 3/3/3/0/0",
			graph.Files, graph.Shown, graph.Total, graph.Capped, graph.HasMore)
	}
	if graph.Limit != 1000 || graph.Offset != 0 || graph.NextOffset != 3 {
		t.Errorf("paging window: got limit=%d offset=%d next_offset=%d, want 1000/0/3",
			graph.Limit, graph.Offset, graph.NextOffset)
	}
	if graph.Root != "." {
		t.Errorf("root: got %q, want %q", graph.Root, ".")
	}

	if graph.Health.Files != 7 || graph.Health.DepFiles != 7 {
		t.Errorf("health: got files=%d dep_files=%d, want 7/7", graph.Health.Files, graph.Health.DepFiles)
	}
	if graph.Health.CCD != 9 || graph.Health.ACD != 1.3 || graph.Health.NCCD != 0.53 {
		t.Errorf("health: got ccd=%d acd=%v nccd=%v, want 9/1.3/0.53", graph.Health.CCD, graph.Health.ACD, graph.Health.NCCD)
	}
	if graph.Health.Shape != "horizontal" {
		t.Errorf("health shape: got %q, want horizontal", graph.Health.Shape)
	}

	if len(graph.Godfiles) != 2 {
		t.Fatalf("godfiles: got %d, want 2", len(graph.Godfiles))
	}
	wantGod := map[string]int{
		"fixturepy/cycle_a/api.py": 1,
		"fixturepy/cycle_b/api.py": 1,
	}
	for _, g := range graph.Godfiles {
		if want, ok := wantGod[g.Path]; !ok || g.Afferent != want {
			t.Errorf("godfile: got %q afferent=%d, want present with %d", g.Path, g.Afferent, want)
		}
	}

	if len(graph.Cycles) != 0 {
		t.Errorf("cycles: got %d (%v), want 0 (Python recording has none)", len(graph.Cycles), graph.Cycles)
	}

	wantFiles := map[string][]string{
		"fixturepy/cycle_a/__init__.py": {"fixturepy.cycle_b.api"},
		"fixturepy/cycle_b/__init__.py": {"fixturepy.cycle_a.api"},
		"fixturepy/models.py":           {"abc"},
	}
	if len(graph.Rows) != len(wantFiles) {
		t.Fatalf("files: got %d, want %d", len(graph.Rows), len(wantFiles))
	}
	for _, f := range graph.Rows {
		want, ok := wantFiles[f.Path]
		if !ok {
			t.Errorf("unexpected file %q", f.Path)
			continue
		}
		if len(f.IncTargets) != len(want) {
			t.Fatalf("%s: inc targets: got %v, want %v", f.Path, f.IncTargets, want)
		}
		for i, target := range want {
			if f.IncTargets[i] != target {
				t.Errorf("%s: inc[%d]: got %q, want %q (verbatim)", f.Path, i, f.IncTargets[i], target)
			}
		}
		if f.Afferent != 0 {
			t.Errorf("%s: afferent: got %d, want 0", f.Path, f.Afferent)
		}
	}
}

func TestParseDepsOutput_RecordedTypeScript(t *testing.T) {
	t.Parallel()

	data := readRecording(t, filepath.Join("testdata", "recorded", "typescript", "deps.xml"))
	graph, err := parseDepsOutput(data)
	if err != nil {
		t.Fatalf("parseDepsOutput: %v", err)
	}

	if graph.Health.Files != 5 || graph.Health.DepFiles != 3 {
		t.Errorf("health: got files=%d dep_files=%d, want 5/3", graph.Health.Files, graph.Health.DepFiles)
	}
	if graph.Health.Shape != "vertical" {
		t.Errorf("health shape: got %q, want vertical", graph.Health.Shape)
	}

	if len(graph.Cycles) != 1 {
		t.Fatalf("cycles: got %d, want 1", len(graph.Cycles))
	}
	cycle := graph.Cycles[0]
	if cycle.Size != 2 || cycle.Cost != 4 {
		t.Errorf("cycle: got size=%d cost=%d, want 2/4", cycle.Size, cycle.Cost)
	}
	wantMembers := map[string]bool{
		"src/cycle-a/index.ts": true,
		"src/cycle-b/index.ts": true,
	}
	if len(cycle.Members) != 2 {
		t.Fatalf("cycle members: got %v, want 2", cycle.Members)
	}
	for _, m := range cycle.Members {
		if !wantMembers[m] {
			t.Errorf("unexpected cycle member %q", m)
		}
	}

	wantFiles := map[string]string{
		"src/cycle-a/index.ts": "../cycle-b",
		"src/cycle-b/index.ts": "../cycle-a",
	}
	if len(graph.Rows) != len(wantFiles) {
		t.Fatalf("files: got %d, want %d", len(graph.Rows), len(wantFiles))
	}
	for _, f := range graph.Rows {
		want, ok := wantFiles[f.Path]
		if !ok {
			t.Errorf("unexpected file %q", f.Path)
			continue
		}
		if len(f.IncTargets) != 1 || f.IncTargets[0] != want {
			t.Errorf("%s: inc targets: got %v, want [%s] verbatim", f.Path, f.IncTargets, want)
		}
		if f.Afferent != 1 || f.Instability != 0.50 {
			t.Errorf("%s: afferent/instab: got %d/%v, want 1/0.50", f.Path, f.Afferent, f.Instability)
		}
	}
}

func TestParseDepsOutput_MalformedInputs(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		data string
	}{
		{"empty data", ""},
		{"garbage", "not xml"},
		{"wrong root", `<arch schema="ripwire.arch/v1"/>`},
		{"missing root attrs", `<deps schema="ripwire.deps/v1"><health files="1" dep_files="1" ccd="1" acd="1" nccd="1" shape="x" dep_langs="py"/></deps>`},
		{"health missing attrs", `<deps schema="ripwire.deps/v1" files="0" shown="0" capped="0" total="0" has_more="0" next_offset="0" offset="0" limit="1" root="."><health dep_files="1" ccd="1" acd="1" nccd="1" shape="x" dep_langs="py"/></deps>`},
		{"f row missing p", `<deps schema="ripwire.deps/v1" files="1" shown="1" capped="0" total="1" has_more="0" next_offset="1" offset="0" limit="1" root="."><health files="1" dep_files="1" ccd="1" acd="1" nccd="1" shape="x" dep_langs="py"/><godfiles total="0" shown="0" capped="0"/><f includes="0" afferent="0" instab="0" transitive="1"/></deps>`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			graph, err := parseDepsOutput([]byte(tc.data))
			if err == nil {
				t.Fatalf("got graph %+v, want error", graph)
			}
		})
	}
}

func TestParseCCOutput_RecordedPython(t *testing.T) {
	t.Parallel()

	data := readRecording(t, filepath.Join("testdata", "recorded", "python", "cc.json"))
	files, err := parseCCOutput(data)
	if err != nil {
		t.Fatalf("parseCCOutput: %v", err)
	}

	if len(files) != 7 {
		t.Fatalf("files: got %d (%v), want 7 (operations/build is pruned)", len(files), ccPaths(files))
	}
	if n := countCCFilesWithPrefix(files, "fixturepy/operations/build"); n != 0 {
		t.Errorf("pruned operations/build: got %d files under it, want 0", n)
	}

	byPath := ccFileByPath(files)
	models, ok := byPath["fixturepy/models.py"]
	if !ok {
		t.Fatalf("models.py missing from %v", ccPaths(files))
	}
	if models.Loc != 94 || models.Symbols != 16 || models.Cx != 11 {
		t.Errorf("models.py: got loc=%d symbols=%d cx=%d, want 94/16/11", models.Loc, models.Symbols, models.Cx)
	}
	if init, ok := byPath["fixturepy/cycle_a/__init__.py"]; !ok || init.FanOut != 1 || init.FanIn != 0 {
		got := ccFile{}
		if ok {
			got = init
		}
		t.Errorf("cycle_a/__init__.py: got fan_in=%d fan_out=%d, want 0/1", got.FanIn, got.FanOut)
	}
	if api, ok := byPath["fixturepy/cycle_a/api.py"]; !ok || api.FanIn != 1 || api.FanOut != 0 {
		got := ccFile{}
		if ok {
			got = api
		}
		t.Errorf("cycle_a/api.py: got fan_in=%d fan_out=%d, want 1/0", got.FanIn, got.FanOut)
	}
}

func TestParseCCOutput_RecordedTypeScript(t *testing.T) {
	t.Parallel()

	data := readRecording(t, filepath.Join("testdata", "recorded", "typescript", "cc.json"))
	files, err := parseCCOutput(data)
	if err != nil {
		t.Fatalf("parseCCOutput: %v", err)
	}

	if len(files) != 5 {
		t.Fatalf("files: got %d (%v), want 5 (config files are indexed)", len(files), ccPaths(files))
	}
	byPath := ccFileByPath(files)
	for _, config := range []string{"tsconfig.json", "package.json"} {
		if _, ok := byPath[config]; !ok {
			t.Errorf("config file %q missing from cc.json tree: %v", config, ccPaths(files))
		}
	}
	if _, ok := byPath["src/models.ts"]; !ok {
		t.Errorf("src/models.ts missing from %v", ccPaths(files))
	}
}

func TestParseCCOutput_MalformedInputs(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		data string
	}{
		{"empty data", ""},
		{"garbage", "not json"},
		{"array top level", `[]`},
		{"wrong shape", `{"projectName":"x"}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			files, err := parseCCOutput([]byte(tc.data))
			if err == nil {
				t.Fatalf("got files %v, want error", files)
			}
		})
	}
}

func TestParseCCOutput_EmptyTreeIsValid(t *testing.T) {
	t.Parallel()

	data := `{"projectName":"project","apiVersion":"1.3","nodes":[{"name":"root","type":"Folder","attributes":{},"children":[]}]}`
	files, err := parseCCOutput([]byte(data))
	if err != nil {
		t.Fatalf("parseCCOutput: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("files: got %d, want 0", len(files))
	}
}

func ccPaths(files []ccFile) []string {
	paths := make([]string, 0, len(files))
	for _, f := range files {
		paths = append(paths, f.Path)
	}
	return paths
}
