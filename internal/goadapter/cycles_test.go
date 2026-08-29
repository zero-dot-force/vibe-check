package goadapter

import (
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
