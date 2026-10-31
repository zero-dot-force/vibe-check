package ripwireadapter

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/zero-dot-force/vibe-check/metrics"
)

// Full-pipeline integration tests — openspec/changes/ripwire-adapter
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
	if fixturepy.Ca != 0 || fixturepy.Ce != 0 {
		t.Errorf("fixturepy ca/ce: got %d/%d, want 0/0", fixturepy.Ca, fixturepy.Ce)
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
	if cycleA.Ca != 1 || cycleA.Ce != 1 || cycleA.Instability != 0.5 {
		t.Errorf("cycle_a ca/ce/I: got %d/%d/%v, want 1/1/0.5", cycleA.Ca, cycleA.Ce, float64(cycleA.Instability))
	}
	cycleB := module("fixturepy.cycle_b")
	if cycleB.Ca != 1 || cycleB.Ce != 1 {
		t.Errorf("cycle_b ca/ce: got %d/%d, want 1/1", cycleB.Ca, cycleB.Ce)
	}

	wantCycle := metrics.Cycle{"fixturepy.cycle_a", "fixturepy.cycle_b"}
	if len(graph.Cycles) != 1 || !reflect.DeepEqual(graph.Cycles[0], wantCycle) {
		t.Errorf("cycles: got %v, want %v", graph.Cycles, wantCycle)
	}

	wantCodes := []string{
		"cycle-approximation", "denylist-pruned", "directory-as-module",
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
// fixtures against a real ripwire 0.6.9 install (task 8.3). It skips
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
