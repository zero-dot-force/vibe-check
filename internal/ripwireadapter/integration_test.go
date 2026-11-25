package ripwireadapter

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/zero-dot-force/vibe-check/metrics"
)

// Full-pipeline integration tests — openspec/changes/archive/2026-11-24-ripwire-adapter
// tasks 8.1, 8.2, and 8.3. Task 8.1 runs the entire Analyze pipeline
// through the fake ripwire on PATH and asserts specific metric values
// against the hand-computed goldens. Task 8.2 proves two-run byte
// equality with provenance omitted. Task 8.3 is the optional
// real-binary corroboration, skipped cleanly when ripwire is not
// installed.

// TestIntegration_PythonFixturePipeline runs the full pipeline for the
// Python fixture through the fake ripwire and asserts the golden values
// metric by metric.
func TestIntegration_PythonFixturePipeline(t *testing.T) {
	installFakeRipwireWrapper(t, filepath.Join("testdata", "recorded", "python"), "valid")

	graph, err := NewPython().Analyze(t.Context(), filepath.Join("testdata", "python"))
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}

	if graph.Language != "python" || graph.Status != metrics.StatusPartial {
		t.Errorf("language/status: got %q/%q, want python/partial", graph.Language, graph.Status)
	}

	module := func(identity string) metrics.ModuleResult {
		t.Helper()
		for _, m := range graph.Modules {
			if m.Path == identity {
				return m
			}
		}
		t.Fatalf("module %q not found", identity)
		return metrics.ModuleResult{}
	}

	fixturepy := module("fixturepy")
	if fixturepy.Ca != 3 || fixturepy.Ce != 0 {
		// wide and ring_a both import `from fixturepy import ...`; rel_b
		// imports `from .. import rel_a`, which ripwire resolves to
		// fixturepy/__init__.py.
		t.Errorf("fixturepy ca/ce: got %d/%d, want 3/0", fixturepy.Ca, fixturepy.Ce)
	}
	if fixturepy.AbstractTypes != 1 || fixturepy.ExportedTypes != 4 {
		t.Errorf("fixturepy abstract/exported: got %d/%d, want 1/4", fixturepy.AbstractTypes, fixturepy.ExportedTypes)
	}
	if fixturepy.Abstractness != 0.25 {
		t.Errorf("fixturepy abstractness: got %v, want 0.25", float64(fixturepy.Abstractness))
	}
	if fixturepy.LCOM != 6 {
		t.Errorf("fixturepy lcom: got %d, want 6", fixturepy.LCOM)
	}

	cycleA := module("fixturepy.cycle_a")
	if cycleA.Ca != 2 || cycleA.Ce != 1 || cycleA.Instability != metrics.ComputeInstability(2, 1) {
		t.Errorf("cycle_a ca/ce/I: got %d/%d/%v, want 2/1/%v", cycleA.Ca, cycleA.Ce,
			float64(cycleA.Instability), float64(metrics.ComputeInstability(2, 1)))
	}
	cycleB := module("fixturepy.cycle_b")
	if cycleB.Ca != 2 || cycleB.Ce != 1 {
		t.Errorf("cycle_b ca/ce: got %d/%d, want 2/1", cycleB.Ca, cycleB.Ce)
	}

	// openspec/changes/archive/2026-11-25-bump-ripwire-pin spec "Fixtures exercise the pinned
	// import fixes": each upstream fix shows at directory grain.
	if wide := module("fixturepy.wide"); wide.Ce != 2 {
		// gap 1: `from fixturepy import cycle_b` reaches the cycle_b
		// subpackage as well as fixturepy/__init__.py.
		t.Errorf("wide ce: got %d, want 2 (fixturepy and fixturepy.cycle_b)", wide.Ce)
	}
	if alias := module("fixturepy.alias"); alias.Ce != 1 {
		// gap 2: `import fixturepy.cycle_a.api as ca` resolves.
		t.Errorf("alias ce: got %d, want 1 (fixturepy.cycle_a)", alias.Ce)
	}
	if collide := module("fixturepy.collide"); collide.Ce != 0 {
		// gap 4: in-package `import logging` binds the stdlib, not the
		// sibling logging.py.
		t.Errorf("collide ce: got %d, want 0 (no false sibling edge)", collide.Ce)
	}

	// The ring pair closes only through the wide-form member edge
	// ring_a → ring_b (openspec/changes/archive/2026-11-24-wide-form-cycle-edges); the rel
	// pair only through relative imports (openspec/changes/
	// relative-import-edges).
	wantCycles := []metrics.Cycle{
		{"fixturepy.cycle_a", "fixturepy.cycle_b"},
		{"fixturepy.rel_a", "fixturepy.rel_b"},
		{"fixturepy.ring_a", "fixturepy.ring_b"},
	}
	if !reflect.DeepEqual(graph.Cycles, wantCycles) {
		t.Errorf("cycles: got %v, want %v", graph.Cycles, wantCycles)
	}

	wantCodes := []string{
		"cycle-approximation", "directory-as-module",
		"error-budget", "name-resolved-edges", "unsupported-metric",
	}
	gotCodes := make([]string, 0, len(graph.Warnings))
	for _, w := range graph.Warnings {
		gotCodes = append(gotCodes, w.Code)
	}
	if !reflect.DeepEqual(gotCodes, wantCodes) {
		t.Errorf("warning codes: got %v, want %v", gotCodes, wantCodes)
	}

	encoded, err := json.Marshal(graph)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := metrics.Validate(encoded); err != nil {
		t.Errorf("metrics.Validate: %v", err)
	}
}

// TestIntegration_MixedLanguagePipeline runs the full pipeline over the
// ARCH-HONEST recording (a Go file inside fixturepy/cycle_a; openspec/
// changes/bump-ripwire-pin spec "Mixed-language Python tree analyzes"):
// analysis must succeed, validate, be partial, and carry one
// coupling-floor warning naming the module and the language.
func TestIntegration_MixedLanguagePipeline(t *testing.T) {
	installFakeRipwireWrapper(t, filepath.Join("testdata", "recorded", "python-mixed"), "valid")

	graph, err := NewPython().Analyze(t.Context(), filepath.Join("testdata", "python"))
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if graph.Status != metrics.StatusPartial {
		t.Errorf("status: got %q, want partial", graph.Status)
	}

	var floors []metrics.Warning
	for _, w := range graph.Warnings {
		if w.Code == warningCouplingFloor {
			floors = append(floors, w)
		}
	}
	if len(floors) != 1 || !strings.Contains(floors[0].Message, "fixturepy.cycle_a") || !strings.Contains(floors[0].Message, "go") {
		t.Errorf("coupling-floor warnings: got %+v, want one naming fixturepy.cycle_a and go", floors)
	}

	// This recording was captured outside git, where ripwire prunes the
	// build/ package: both sources name it and the warning lists it once.
	var denylist []metrics.Warning
	for _, w := range graph.Warnings {
		if w.Code == warningDenylistPruned {
			denylist = append(denylist, w)
		}
	}
	wantDenylist := "ripwire prunes source-bearing denylist directories from the crawl: fixturepy/operations/build"
	if len(denylist) != 1 || denylist[0].Message != wantDenylist {
		t.Errorf("denylist-pruned warnings: got %+v, want one with message %q", denylist, wantDenylist)
	}

	for _, m := range graph.Modules {
		if m.Path == "fixturepy.cycle_a" && (m.Ca != 2 || m.Ce != 1) {
			t.Errorf("cycle_a floors: got ca/ce %d/%d, want 2/1", m.Ca, m.Ce)
		}
	}

	encoded, err := json.Marshal(graph)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := metrics.Validate(encoded); err != nil {
		t.Errorf("metrics.Validate: %v", err)
	}
}

// TestIntegration_TypeScriptFixturePipeline runs the full pipeline for
// the TypeScript fixture through the fake ripwire and asserts the golden
// values metric by metric.
func TestIntegration_TypeScriptFixturePipeline(t *testing.T) {
	installFakeRipwireWrapper(t, filepath.Join("testdata", "recorded", "typescript"), "valid")

	graph, err := NewTypeScript().Analyze(t.Context(), filepath.Join("testdata", "typescript"))
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}

	if graph.Language != "typescript" || graph.Status != metrics.StatusPartial {
		t.Errorf("language/status: got %q/%q, want typescript/partial", graph.Language, graph.Status)
	}

	module := func(identity string) metrics.ModuleResult {
		t.Helper()
		for _, m := range graph.Modules {
			if m.Path == identity {
				return m
			}
		}
		t.Fatalf("module %q not found", identity)
		return metrics.ModuleResult{}
	}

	src := module("src")
	if src.AbstractTypes != 4 || src.ExportedTypes != 5 {
		t.Errorf("src abstract/exported: got %d/%d, want 4/5", src.AbstractTypes, src.ExportedTypes)
	}
	if src.Abstractness != 0.8 {
		t.Errorf("src abstractness: got %v, want 0.8", float64(src.Abstractness))
	}
	if src.LCOM != 2 {
		t.Errorf("src lcom: got %d, want 2", src.LCOM)
	}
	if src.Zone != metrics.ZoneMainSequence {
		t.Errorf("src zone: got %q, want main-sequence", src.Zone)
	}

	cycleA := module("src/cycle-a")
	if cycleA.Ca != 1 || cycleA.Ce != 1 || cycleA.Instability != 0.5 {
		t.Errorf("cycle-a ca/ce/I: got %d/%d/%v, want 1/1/0.5", cycleA.Ca, cycleA.Ce, float64(cycleA.Instability))
	}

	wantCycle := metrics.Cycle{"src/cycle-a", "src/cycle-b"}
	if len(graph.Cycles) != 1 || !reflect.DeepEqual(graph.Cycles[0], wantCycle) {
		t.Errorf("cycles: got %v, want %v", graph.Cycles, wantCycle)
	}

	wantCodes := []string{
		"cycle-approximation", "directory-as-module", "error-budget",
		"ts-abstractness-bias", "unsupported-metric",
	}
	gotCodes := make([]string, 0, len(graph.Warnings))
	for _, w := range graph.Warnings {
		gotCodes = append(gotCodes, w.Code)
	}
	if !reflect.DeepEqual(gotCodes, wantCodes) {
		t.Errorf("warning codes: got %v, want %v", gotCodes, wantCodes)
	}

	encoded, err := json.Marshal(graph)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := metrics.Validate(encoded); err != nil {
		t.Errorf("metrics.Validate: %v", err)
	}
}

// TestIntegration_Determinism proves two full Analyze runs per fixture
// produce byte-identical ModuleGraph JSON with provenance omitted
// (task 8.2).
func TestIntegration_Determinism(t *testing.T) {
	for _, language := range []string{languagePython, languageTypeScript} {
		t.Run(language, func(t *testing.T) {
			adapter := NewPython()
			if language == languageTypeScript {
				adapter = NewTypeScript()
			}
			installFakeRipwireWrapper(t, filepath.Join("testdata", "recorded", language), "valid")

			project := filepath.Join("testdata", "python")
			if language == languageTypeScript {
				project = filepath.Join("testdata", "typescript")
			}

			first, err := adapter.Analyze(t.Context(), project)
			if err != nil {
				t.Fatalf("first Analyze: %v", err)
			}
			second, err := adapter.Analyze(t.Context(), project)
			if err != nil {
				t.Fatalf("second Analyze: %v", err)
			}

			first.Provenance = nil
			second.Provenance = nil
			firstJSON, err := json.Marshal(first)
			if err != nil {
				t.Fatalf("marshal first: %v", err)
			}
			secondJSON, err := json.Marshal(second)
			if err != nil {
				t.Fatalf("marshal second: %v", err)
			}
			if !bytes.Equal(firstJSON, secondJSON) {
				t.Error("two runs produced different ModuleGraph JSON")
			}
		})
	}
}

// TestIntegration_RealBinaryCorroboration optionally corroborates the
// fixtures against a real install of the pinned ripwire release (task 8.3). It skips
// cleanly when ripwire is not on PATH; where installed, the real
// binary's output must reproduce the golden module metrics.
func TestIntegration_RealBinaryCorroboration(t *testing.T) {
	if _, err := exec.LookPath("ripwire"); err != nil {
		t.Skip("ripwire not on PATH: real-binary corroboration skipped")
	}

	for _, language := range []string{languagePython, languageTypeScript} {
		t.Run(language, func(t *testing.T) {
			adapter := NewPython()
			if language == languageTypeScript {
				adapter = NewTypeScript()
			}
			project := filepath.Join("testdata", "python")
			if language == languageTypeScript {
				project = filepath.Join("testdata", "typescript")
			}
			graph, err := adapter.Analyze(t.Context(), project)
			if err != nil {
				t.Fatalf("Analyze with real ripwire: %v", err)
			}

			// The real binary must reproduce the recorded goldens:
			// module identity, coupling, and types are pinned; the
			// light-parse metrics (abstractness for Python, LCOM)
			// are adapter-owned and therefore identical by
			// construction.
			golden := loadGolden(t, language)
			if len(graph.Modules) != len(golden.Modules) {
				t.Fatalf("modules: got %d, want %d", len(graph.Modules), len(golden.Modules))
			}
			for i, m := range graph.Modules {
				want := golden.Modules[i]
				if m.Path != want.Path || m.Ca != want.Ca || m.Ce != want.Ce {
					t.Errorf("module %d (%s): ca/ce got %d/%d, want %d/%d",
						i, m.Path, m.Ca, m.Ce, want.Ca, want.Ce)
				}
				if m.ExportedTypes != want.ExportedTypes || m.AbstractTypes != want.AbstractTypes {
					t.Errorf("module %s types: got %d/%d, want %d/%d",
						m.Path, m.AbstractTypes, m.ExportedTypes, want.AbstractTypes, want.ExportedTypes)
				}
			}
			if !reflect.DeepEqual(graph.Cycles, golden.Cycles) {
				t.Errorf("cycles: got %v, want %v", graph.Cycles, golden.Cycles)
			}
		})
	}
}
