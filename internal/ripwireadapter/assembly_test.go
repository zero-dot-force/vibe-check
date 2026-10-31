package ripwireadapter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/zero-dot-force/vibe-check/metrics"
)

// Analyze assembly tests — openspec/changes/ripwire-adapter task 6.2.
// The assembled graphs for both fixtures must match the hand-computed
// goldens (modules, cycles, warning codes, status), pass
// metrics.Validate, and carry the source-tool provenance fields.

// analyzeFixture runs the full Analyze pipeline for a fixture language
// through the fake ripwire harness.
func analyzeFixture(t *testing.T, language string) *metrics.ModuleGraph {
	t.Helper()
	adapter := NewPython()
	if language == languageTypeScript {
		adapter = NewTypeScript()
	}
	installFakeRipwireWrapper(t, filepath.Join("testdata", "recorded", language), "valid")

	project := filepath.Join("testdata", "python")
	if language == languageTypeScript {
		project = filepath.Join("testdata", "typescript")
	}
	graph, err := adapter.Analyze(t.Context(), project)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	return graph
}

// loadGolden unmarshals a fixture golden.json into a ModuleGraph.
func loadGolden(t *testing.T, language string) metrics.ModuleGraph {
	t.Helper()
	data := readRecording(t, filepath.Join("testdata", "recorded", language, "golden.json"))
	var golden metrics.ModuleGraph
	if err := json.Unmarshal(data, &golden); err != nil {
		t.Fatalf("unmarshal golden: %v", err)
	}
	return golden
}

func TestAnalyze_MatchesGolden(t *testing.T) {
	for _, language := range []string{languagePython, languageTypeScript} {
		t.Run(language, func(t *testing.T) {
			graph := analyzeFixture(t, language)
			golden := loadGolden(t, language)

			if graph.SchemaVersion != golden.SchemaVersion {
				t.Errorf("schemaVersion: got %q, want %q", graph.SchemaVersion, golden.SchemaVersion)
			}
			if graph.Language != golden.Language {
				t.Errorf("language: got %q, want %q", graph.Language, golden.Language)
			}
			if graph.Status != golden.Status {
				t.Errorf("status: got %q, want %q", graph.Status, golden.Status)
			}
			if !reflect.DeepEqual(graph.Modules, golden.Modules) {
				t.Errorf("modules:\ngot  %+v\nwant %+v", graph.Modules, golden.Modules)
			}
			if !reflect.DeepEqual(graph.Cycles, golden.Cycles) {
				t.Errorf("cycles: got %v, want %v", graph.Cycles, golden.Cycles)
			}

			gotCodes := make([]string, 0, len(graph.Warnings))
			for _, w := range graph.Warnings {
				gotCodes = append(gotCodes, w.Code)
			}
			wantCodes := make([]string, 0, len(golden.Warnings))
			for _, w := range golden.Warnings {
				wantCodes = append(wantCodes, w.Code)
			}
			if !reflect.DeepEqual(gotCodes, wantCodes) {
				t.Errorf("warning codes: got %v, want %v", gotCodes, wantCodes)
			}
		})
	}
}

func TestAnalyze_ValidatesAgainstSchema(t *testing.T) {
	for _, language := range []string{languagePython, languageTypeScript} {
		t.Run(language, func(t *testing.T) {
			graph := analyzeFixture(t, language)
			encoded, err := json.Marshal(graph)
			if err != nil {
				t.Fatalf("marshal graph: %v", err)
			}
			if err := metrics.Validate(encoded); err != nil {
				t.Errorf("metrics.Validate: %v", err)
			}
		})
	}
}

func TestAnalyze_ProvenanceSourceToolFields(t *testing.T) {
	for _, language := range []string{languagePython, languageTypeScript} {
		t.Run(language, func(t *testing.T) {
			graph := analyzeFixture(t, language)
			if graph.Provenance == nil {
				t.Fatal("provenance: got nil, want populated")
			}
			input := graph.Provenance.Input
			if input.SourceTool != "ripwire" {
				t.Errorf("sourceTool: got %q, want ripwire", input.SourceTool)
			}
			if input.SourceToolVersion != "0.6.9" {
				t.Errorf("sourceToolVersion: got %q, want 0.6.9", input.SourceToolVersion)
			}
			if !strings.HasPrefix(input.SourceToolCommit, "bedc0cd") {
				t.Errorf("sourceToolCommit: got %q, want prefix bedc0cd", input.SourceToolCommit)
			}
			if input.Path == "" {
				t.Error("provenance path: got empty, want the analyzed path")
			}
			if input.ModulePath == "" {
				t.Error("provenance modulePath: got empty, want the manifest identity")
			}
		})
	}
}

func TestAnalyze_StructuralInvariants(t *testing.T) {
	for _, language := range []string{languagePython, languageTypeScript} {
		t.Run(language, func(t *testing.T) {
			graph := analyzeFixture(t, language)

			if graph.Modules == nil {
				t.Error("modules slice: got nil, want non-nil")
			}
			if graph.Cycles == nil {
				t.Error("cycles slice: got nil, want non-nil")
			}
			if graph.Warnings == nil {
				t.Error("warnings slice: got nil, want non-nil")
			}
			for i, m := range graph.Modules {
				if m.Duplications == nil {
					t.Errorf("modules[%d].duplications: got nil, want non-nil", i)
				}
				if m.TotalLines != 0 {
					t.Errorf("modules[%d].totalLines: got %d, want 0 (duplication unsupported)", i, m.TotalLines)
				}
			}

			// Modules sorted by path.
			paths := make([]string, 0, len(graph.Modules))
			for _, m := range graph.Modules {
				paths = append(paths, m.Path)
			}
			if !sort.StringsAreSorted(paths) {
				t.Errorf("module paths not sorted: %v", paths)
			}

			// Warnings sorted by code.
			codes := make([]string, 0, len(graph.Warnings))
			for _, w := range graph.Warnings {
				codes = append(codes, w.Code)
			}
			if !sort.StringsAreSorted(codes) {
				t.Errorf("warning codes not sorted: %v", codes)
			}
		})
	}
}

func TestDepsWindowWarning(t *testing.T) {
	t.Parallel()

	if got := depsWindowWarning(depsGraph{HasMore: 0, Shown: 3, Total: 3}); got != nil {
		t.Errorf("complete window: got %v, want nil", got)
	}
	warnings := depsWindowWarning(depsGraph{HasMore: 1, Shown: 1000, Total: 2400})
	if len(warnings) != 1 || warnings[0].Code != warningDepsWindowTruncated {
		t.Fatalf("truncated window: got %v, want one deps-window-truncated", warnings)
	}
	if want := "1000 of 2400"; !strings.Contains(warnings[0].Message, want) {
		t.Errorf("message %q does not say %q", warnings[0].Message, want)
	}
}

func TestAnalyze_ModuleIdentityFromManifest(t *testing.T) {
	cases := []struct {
		language, wantModulePath string
	}{
		{languagePython, "fixturepy"},
		{languageTypeScript, "fixturets"},
	}
	for _, tc := range cases {
		t.Run(tc.language, func(t *testing.T) {
			graph := analyzeFixture(t, tc.language)
			if got := graph.Provenance.Input.ModulePath; got != tc.wantModulePath {
				t.Errorf("modulePath: got %q, want %q", got, tc.wantModulePath)
			}
		})
	}
}

func TestResolveModuleIdentity_Fallbacks(t *testing.T) {
	t.Parallel()

	// No manifest at all: the crawl root's base name.
	root := t.TempDir()
	pkg := filepath.Join(root, "widget")
	if err := os.MkdirAll(pkg, 0o755); err != nil {
		t.Fatal(err)
	}
	adapter := NewPython()
	got, err := adapter.resolveModuleIdentity(root, pkg)
	if err != nil {
		t.Fatalf("resolveModuleIdentity: %v", err)
	}
	if got != "widget" {
		t.Errorf("fallback identity: got %q, want widget", got)
	}

	// Manifest without a name: still the fallback.
	root2 := t.TempDir()
	if err := os.WriteFile(filepath.Join(root2, "pyproject.toml"), []byte("[project]\nversion = \"0.1.0\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err = adapter.resolveModuleIdentity(root2, root2)
	if err != nil {
		t.Fatalf("resolveModuleIdentity: %v", err)
	}
	if got == "" {
		t.Error("nameless manifest: got empty identity, want directory fallback")
	}
}

func TestPythonProjectName(t *testing.T) {
	t.Parallel()

	name, err := pythonProjectName(filepath.Join("testdata", "python"))
	if err != nil {
		t.Fatalf("pythonProjectName: %v", err)
	}
	if name != "fixturepy" {
		t.Errorf("name: got %q, want fixturepy", name)
	}

	// Legacy poetry layout.
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "pyproject.toml"),
		[]byte("[tool.poetry]\nname = \"legacy\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	name, err = pythonProjectName(dir)
	if err != nil {
		t.Fatalf("pythonProjectName: %v", err)
	}
	if name != "legacy" {
		t.Errorf("poetry name: got %q, want legacy", name)
	}

	// Absent manifest: empty, no error.
	name, err = pythonProjectName(t.TempDir())
	if err != nil || name != "" {
		t.Errorf("absent manifest: got (%q, %v), want empty and nil", name, err)
	}
}

func TestTypeScriptPackageName(t *testing.T) {
	t.Parallel()

	name, err := typeScriptPackageName(filepath.Join("testdata", "typescript"))
	if err != nil {
		t.Fatalf("typeScriptPackageName: %v", err)
	}
	if name != "fixturets" {
		t.Errorf("name: got %q, want fixturets", name)
	}

	// Malformed manifest fails closed.
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := typeScriptPackageName(dir); err == nil {
		t.Error("malformed manifest: got no error, want fail-closed error")
	}
}
