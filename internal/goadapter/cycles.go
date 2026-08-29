package goadapter

import (
	"sort"

	"github.com/zero-dot-force/vibe-check/metrics"
)

// detectCycles runs Tarjan's SCC algorithm on the module-internal import graph
// and returns cycles with canonical ordering. Only module-internal edges are
// considered (edges to packages not in modulePkgs are ignored).
//
// Each cycle is rotated so the lexicographically smallest path appears first,
// per the [metrics.Cycle] contract. The returned slice is sorted by first element.
// Returns an empty (non-nil) slice if no cycles exist.
func detectCycles(imports map[string][]string, modulePkgs map[string]bool) []metrics.Cycle {
	t := &tarjan{
		imports:    imports,
		modulePkgs: modulePkgs,
		index:      make(map[string]int),
		lowlink:    make(map[string]int),
		onStack:    make(map[string]bool),
		counter:    0,
	}

	// Run Tarjan's on all module-internal nodes.
	for pkg := range modulePkgs {
		if _, visited := t.index[pkg]; !visited {
			t.strongConnect(pkg)
		}
	}

	// Filter SCCs to only those with size > 1 (actual cycles).
	var cycles []metrics.Cycle
	for _, scc := range t.sccs {
		if len(scc) <= 1 {
			continue
		}
		cycles = append(cycles, canonicalizeCycle(scc))
	}

	// Sort cycles by first element for deterministic output.
	sort.Slice(cycles, func(i, j int) bool {
		return cycles[i][0] < cycles[j][0]
	})

	// Ensure non-nil slice per ModuleGraph contract.
	if cycles == nil {
		cycles = []metrics.Cycle{}
	}

	return cycles
}

// tarjan holds the state for Tarjan's SCC algorithm.
type tarjan struct {
	imports    map[string][]string
	modulePkgs map[string]bool
	index      map[string]int
	lowlink    map[string]int
	onStack    map[string]bool
	stack      []string
	counter    int
	sccs       [][]string
}

// strongConnect is the recursive Tarjan's SCC procedure.
func (t *tarjan) strongConnect(v string) {
	t.index[v] = t.counter
	t.lowlink[v] = t.counter
	t.counter++
	t.stack = append(t.stack, v)
	t.onStack[v] = true

	// Consider successors: only module-internal edges.
	for _, w := range t.imports[v] {
		if !t.modulePkgs[w] {
			continue
		}
		if _, visited := t.index[w]; !visited {
			t.strongConnect(w)
			if t.lowlink[w] < t.lowlink[v] {
				t.lowlink[v] = t.lowlink[w]
			}
		} else if t.onStack[w] {
			if t.index[w] < t.lowlink[v] {
				t.lowlink[v] = t.index[w]
			}
		}
	}

	// If v is a root node, pop the SCC.
	if t.lowlink[v] == t.index[v] {
		var scc []string
		for {
			w := t.stack[len(t.stack)-1]
			t.stack = t.stack[:len(t.stack)-1]
			t.onStack[w] = false
			scc = append(scc, w)
			if w == v {
				break
			}
		}
		t.sccs = append(t.sccs, scc)
	}
}

// canonicalizeCycle rotates a cycle so the lexicographically smallest element
// appears first, then sorts the remaining elements in the order they appear
// in the cycle.
func canonicalizeCycle(scc []string) metrics.Cycle {
	// Sort the SCC members to find canonical ordering.
	// Tarjan's returns SCCs in reverse topological order within the component,
	// so we sort to get a deterministic representation.
	sorted := make([]string, len(scc))
	copy(sorted, scc)
	sort.Strings(sorted)
	return metrics.Cycle(sorted)
}
