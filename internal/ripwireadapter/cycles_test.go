package ripwireadapter

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/zero-dot-force/vibe-check/metrics"
)

// Cycle derivation tests — openspec/changes/ripwire-adapter task 4.4.
//
// They pin the Tarjan SCC derivation over directory-module edges, the
// canonical lexicographic ordering, the corroboration against ripwire's
// capped file-level listing, and the cycle-approximation warning.

func deriveForRecorded(t *testing.T, language string) ([]metrics.Cycle, moduleMapping, depsGraph) {
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
	cycles, _ := deriveCycles(mapping, deps)
	return cycles, mapping, deps
}

func TestDeriveCycles_RecordedPython(t *testing.T) {
	t.Parallel()

	cycles, _, _ := deriveForRecorded(t, "python")
	want := []metrics.Cycle{{"fixturepy.cycle_a", "fixturepy.cycle_b"}}
	if !reflect.DeepEqual(cycles, want) {
		t.Errorf("cycles: got %v, want %v", cycles, want)
	}
}

func TestDeriveCycles_RecordedTypeScript(t *testing.T) {
	t.Parallel()

	cycles, _, _ := deriveForRecorded(t, "typescript")
	want := []metrics.Cycle{{"src/cycle-a", "src/cycle-b"}}
	if !reflect.DeepEqual(cycles, want) {
		t.Errorf("cycles: got %v, want %v", cycles, want)
	}
}

func TestDeriveCycles_WarningPresent(t *testing.T) {
	t.Parallel()

	for _, language := range []string{"python", "typescript"} {
		t.Run(language, func(t *testing.T) {
			_, mapping, deps := deriveForRecorded(t, language)
			cycles, warnings := deriveCycles(mapping, deps)
			if len(cycles) == 0 {
				t.Fatal("no cycles derived for a fixture that has one")
			}
			if len(warnings) != 1 {
				t.Fatalf("warnings: got %d (%v), want exactly 1", len(warnings), warnings)
			}
			if warnings[0].Code != "cycle-approximation" {
				t.Errorf("warning code: got %q, want cycle-approximation", warnings[0].Code)
			}
		})
	}
}

func TestDeriveCycles_NoCyclesNoWarning(t *testing.T) {
	t.Parallel()

	mapping := moduleMapping{
		Modules: []mappedModule{
			{Identity: "a"}, {Identity: "b"},
		},
		Edges: []moduleEdge{{From: "a", To: "b"}}, // acyclic
	}
	cycles, warnings := deriveCycles(mapping, depsGraph{})
	if len(cycles) != 0 {
		t.Errorf("cycles: got %v, want none", cycles)
	}
	if len(warnings) != 0 {
		t.Errorf("warnings: got %v, want none for an acyclic graph", warnings)
	}
}

func TestDeriveCycles_MultiSCCCanonicalOrdering(t *testing.T) {
	t.Parallel()

	// Three SCCs of different shapes plus isolated nodes, with adjacency
	// deliberately listed out of order to prove the canonical sort:
	//   - {z, y, x}: three-member cycle
	//   - {b, a}: two-member cycle
	//   - {m}: self-referential node ignored (SCC of one)
	//   - {n}: isolated
	mapping := moduleMapping{
		Modules: []mappedModule{
			{Identity: "x"}, {Identity: "y"}, {Identity: "z"},
			{Identity: "a"}, {Identity: "b"},
			{Identity: "m"}, {Identity: "n"},
		},
		Edges: []moduleEdge{
			{From: "z", To: "y"}, {From: "y", To: "x"}, {From: "x", To: "z"},
			{From: "b", To: "a"}, {From: "a", To: "b"},
			{From: "m", To: "m"}, // self-loop: not a cycle member
		},
	}
	cycles, warnings := deriveCycles(mapping, depsGraph{})

	want := []metrics.Cycle{
		{"a", "b"},
		{"x", "y", "z"},
	}
	if !reflect.DeepEqual(cycles, want) {
		t.Errorf("cycles: got %v, want %v", cycles, want)
	}
	if len(warnings) != 1 || warnings[0].Code != "cycle-approximation" {
		t.Errorf("warnings: got %v, want one cycle-approximation", warnings)
	}
}

func TestDeriveCycles_Corroboration(t *testing.T) {
	t.Parallel()

	mapping := moduleMapping{
		Modules: []mappedModule{
			{Identity: "src/cycle-a"}, {Identity: "src/cycle-b"},
		},
		Edges: []moduleEdge{
			{From: "src/cycle-a", To: "src/cycle-b"},
			{From: "src/cycle-b", To: "src/cycle-a"},
		},
	}

	t.Run("ripwire listing confirms", func(t *testing.T) {
		deps := depsGraph{Cycles: []depsCycle{{
			Size:    2,
			Members: []string{"src/cycle-b/index.ts", "src/cycle-a/index.ts"},
		}}}
		_, warnings := deriveCycles(mapping, deps)
		if len(warnings) != 1 {
			t.Fatalf("warnings: got %v, want one", warnings)
		}
		want := "corroborates the derived cycles"
		if !strings.Contains(warnings[0].Message, want) {
			t.Errorf("message %q does not say %q", warnings[0].Message, want)
		}
	})

	t.Run("ripwire listing empty (python case)", func(t *testing.T) {
		_, warnings := deriveCycles(mapping, depsGraph{})
		if len(warnings) != 1 {
			t.Fatalf("warnings: got %v, want one", warnings)
		}
		want := "does not corroborate"
		if !strings.Contains(warnings[0].Message, want) {
			t.Errorf("message %q does not say %q", warnings[0].Message, want)
		}
	})
}

func TestDeriveCycles_Deterministic(t *testing.T) {
	t.Parallel()

	for _, language := range []string{"python", "typescript"} {
		first, _, _ := deriveForRecorded(t, language)
		second, _, _ := deriveForRecorded(t, language)
		if !reflect.DeepEqual(first, second) {
			t.Errorf("%s: deriving twice yielded different cycles", language)
		}
	}
}
