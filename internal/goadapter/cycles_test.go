package goadapter

import (
	"encoding/json"
	"testing"
)

func TestDetectCycles_Acyclic(t *testing.T) {
	t.Parallel()

	imports := map[string][]string{
		"example.com/foo/a": {"example.com/foo/b"},
		"example.com/foo/b": {"example.com/foo/c"},
		"example.com/foo/c": {},
	}
	modulePkgs := map[string]bool{
		"example.com/foo/a": true,
		"example.com/foo/b": true,
		"example.com/foo/c": true,
	}

	cycles := detectCycles(imports, modulePkgs)

	if cycles == nil {
		t.Fatal("cycles slice is nil, want non-nil empty slice")
	}
	if len(cycles) != 0 {
		t.Errorf("got %d cycles, want 0", len(cycles))
	}
}

func TestDetectCycles_ConstructedCycle(t *testing.T) {
	t.Parallel()

	// A→B→C→A forms a cycle.
	imports := map[string][]string{
		"example.com/foo/a": {"example.com/foo/b"},
		"example.com/foo/b": {"example.com/foo/c"},
		"example.com/foo/c": {"example.com/foo/a"},
	}
	modulePkgs := map[string]bool{
		"example.com/foo/a": true,
		"example.com/foo/b": true,
		"example.com/foo/c": true,
	}

	cycles := detectCycles(imports, modulePkgs)

	if len(cycles) != 1 {
		t.Fatalf("got %d cycles, want 1", len(cycles))
	}
	if len(cycles[0]) != 3 {
		t.Fatalf("cycle length: got %d, want 3", len(cycles[0]))
	}
}

func TestDetectCycles_CanonicalOrdering(t *testing.T) {
	t.Parallel()

	// Cycle: C→A→B→C. Canonical ordering should start with A.
	imports := map[string][]string{
		"example.com/foo/c": {"example.com/foo/a"},
		"example.com/foo/a": {"example.com/foo/b"},
		"example.com/foo/b": {"example.com/foo/c"},
	}
	modulePkgs := map[string]bool{
		"example.com/foo/a": true,
		"example.com/foo/b": true,
		"example.com/foo/c": true,
	}

	cycles := detectCycles(imports, modulePkgs)

	if len(cycles) != 1 {
		t.Fatalf("got %d cycles, want 1", len(cycles))
	}

	cycle := cycles[0]
	if cycle[0] != "example.com/foo/a" {
		t.Errorf("first element: got %q, want %q", cycle[0], "example.com/foo/a")
	}
	if cycle[1] != "example.com/foo/b" {
		t.Errorf("second element: got %q, want %q", cycle[1], "example.com/foo/b")
	}
	if cycle[2] != "example.com/foo/c" {
		t.Errorf("third element: got %q, want %q", cycle[2], "example.com/foo/c")
	}
}

// TestDetectCycles_SortedSetSemantics pins the [metrics.Cycle] contract:
// cycles are reported as the fully lexicographically-sorted member set, not a
// rotated traversal path. The import graph B→A→C→B has a traversal order that
// differs from the sorted set: a rotation-to-smallest-first scheme would yield
// [A, C, B], whereas sorted-set semantics yield [A, B, C]. Asserting the fully
// sorted order distinguishes the two schemes non-coincidentally.
func TestDetectCycles_SortedSetSemantics(t *testing.T) {
	t.Parallel()

	// Cycle: B→A→C→B.
	imports := map[string][]string{
		"example.com/foo/b": {"example.com/foo/a"},
		"example.com/foo/a": {"example.com/foo/c"},
		"example.com/foo/c": {"example.com/foo/b"},
	}
	modulePkgs := map[string]bool{
		"example.com/foo/a": true,
		"example.com/foo/b": true,
		"example.com/foo/c": true,
	}

	cycles := detectCycles(imports, modulePkgs)

	if len(cycles) != 1 {
		t.Fatalf("got %d cycles, want 1", len(cycles))
	}

	want := []string{"example.com/foo/a", "example.com/foo/b", "example.com/foo/c"}
	got := []string(cycles[0])
	if len(got) != len(want) {
		t.Fatalf("cycle length: got %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("cycle[%d]: got %q, want %q (canonical form is the fully-sorted member set)",
				i, got[i], want[i])
		}
	}
}

// TestDetectCycles_Determinism verifies that cycle detection produces
// byte-identical, stably-ordered output across repeated runs. detectCycles
// seeds Tarjan's algorithm by ranging over a map (whose iteration order is
// randomized by the runtime), so this test guards against nondeterministic
// output leaking through.
func TestDetectCycles_Determinism(t *testing.T) {
	t.Parallel()

	// Two independent cycles (a→b→c→a and x→y→x) plus an acyclic node z.
	imports := map[string][]string{
		"example.com/m/a": {"example.com/m/b"},
		"example.com/m/b": {"example.com/m/c"},
		"example.com/m/c": {"example.com/m/a"},
		"example.com/m/x": {"example.com/m/y"},
		"example.com/m/y": {"example.com/m/x"},
		"example.com/m/z": {"example.com/m/a"},
	}
	modulePkgs := map[string]bool{
		"example.com/m/a": true,
		"example.com/m/b": true,
		"example.com/m/c": true,
		"example.com/m/x": true,
		"example.com/m/y": true,
		"example.com/m/z": true,
	}

	var reference string
	for i := 0; i < 20; i++ {
		cycles := detectCycles(imports, modulePkgs)
		data, err := json.Marshal(cycles)
		if err != nil {
			t.Fatalf("run %d: marshal: %v", i, err)
		}
		if i == 0 {
			reference = string(data)
			continue
		}
		if string(data) != reference {
			t.Errorf("run %d: cycle output differs from reference\ngot:  %s\nwant: %s", i, data, reference)
		}
	}

	// Sanity check: exactly two cycles are detected.
	if got := detectCycles(imports, modulePkgs); len(got) != 2 {
		t.Fatalf("got %d cycles, want 2", len(got))
	}
}
