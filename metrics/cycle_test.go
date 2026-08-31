package metrics

import "testing"

func TestCycle_Construction(t *testing.T) {
	t.Parallel()

	// Cycle is a named type over []string. Verify basic construction
	// and element access work as expected.
	cycle := Cycle{"A", "B", "C"}

	if len(cycle) != 3 {
		t.Fatalf("len(cycle) = %d, want 3", len(cycle))
	}
	if cycle[0] != "A" {
		t.Errorf("cycle[0] = %q, want %q", cycle[0], "A")
	}
	if cycle[1] != "B" {
		t.Errorf("cycle[1] = %q, want %q", cycle[1], "B")
	}
	if cycle[2] != "C" {
		t.Errorf("cycle[2] = %q, want %q", cycle[2], "C")
	}
}

func TestCycle_CanonicalOrdering(t *testing.T) {
	t.Parallel()

	// Per the Cycle contract, a cycle's members are reported as a
	// deterministic, lexicographically-sorted set of package paths:
	// ["A", "B", "C"]. The ordering carries no traversal meaning.
	// This test verifies that a correctly constructed Cycle follows
	// the sorted-set convention (smallest element first, no repeated
	// start node).
	//
	// Note: producing the sorted set from detected cycles is the
	// responsibility of the cycle detection algorithm in language
	// adapters, not the Cycle type itself. This test validates the
	// convention on a pre-constructed cycle.
	cycle := Cycle{"A", "B", "C"}

	// The first element must be the lexicographically smallest.
	for i := 1; i < len(cycle); i++ {
		if cycle[i] < cycle[0] {
			t.Errorf("cycle[%d] = %q is lexicographically smaller than cycle[0] = %q; "+
				"canonical ordering requires the smallest path first", i, cycle[i], cycle[0])
		}
	}

	// The start node must not be repeated at the end.
	if len(cycle) > 1 && cycle[len(cycle)-1] == cycle[0] {
		t.Errorf("cycle repeats start node %q at the end; canonical representation omits it", cycle[0])
	}
}

func TestCycle_EmptyCycle(t *testing.T) {
	t.Parallel()

	// An empty Cycle is valid — it represents "no cycle".
	var cycle Cycle
	if len(cycle) != 0 {
		t.Errorf("len(empty cycle) = %d, want 0", len(cycle))
	}
}

func TestCycle_TwoNodeCycle(t *testing.T) {
	t.Parallel()

	// A two-node cycle A→B→A is represented as ["A", "B"].
	cycle := Cycle{"A", "B"}

	if len(cycle) != 2 {
		t.Fatalf("len(cycle) = %d, want 2", len(cycle))
	}
	if cycle[0] != "A" {
		t.Errorf("cycle[0] = %q, want %q", cycle[0], "A")
	}
	if cycle[1] != "B" {
		t.Errorf("cycle[1] = %q, want %q", cycle[1], "B")
	}
}
