package ripwireadapter

import (
	"fmt"
	"sort"
	"strings"

	"github.com/zero-dot-force/vibe-check/metrics"
)

// warningCycleApproximation is emitted when the adapter derives cycles
// itself: the derivation is a directory-grain SCC approximation of
// ripwire's capped, file-level include-cycle listing.
const warningCycleApproximation = "cycle-approximation"

// ripwireCaps are the listing caps ripwire applies to its own <cycles>
// report (validated against the bundled 0.6.9 reference): at most 20
// SCCs, at most 12 members each, with no total count. The adapter's own
// derivation is uncapped, which is one reason it is an approximation.
const (
	ripwireCycleMaxSCCs    = 20
	ripwireCycleMaxMembers = 12
)

// deriveCycles runs Tarjan's strongly-connected-components algorithm over
// the directory-module edge graph and returns every non-trivial SCC
// (two or more members) as a [metrics.Cycle], with the canonical
// ordering the universal model requires: members lexicographically
// sorted within a cycle, cycles sorted by their first member.
//
// Corroboration: ripwire's own <cycles> listing (file-grain, capped at
// 20 SCCs × 12 members) is mapped to directory grain and compared against
// the derivation. The corroboration result feeds the cycle-approximation
// warning, which is emitted whenever the adapter derived at least one
// cycle — the derivation is an approximation even when ripwire's listing
// happens to agree, and ripwire's caps mean its listing can never serve
// as the complete truth.
func deriveCycles(mapping moduleMapping, deps depsGraph) ([]metrics.Cycle, []metrics.Warning) {
	// Tarjan over the identity graph.
	identities := make([]string, 0, len(mapping.Modules))
	adjacency := make(map[string][]string, len(mapping.Modules))
	for _, m := range mapping.Modules {
		identities = append(identities, m.Identity)
		adjacency[m.Identity] = nil
	}
	for _, edge := range mapping.Edges {
		adjacency[edge.From] = append(adjacency[edge.From], edge.To)
	}
	sort.Strings(identities)

	sccs := tarjanSCCs(identities, adjacency)

	cycles := make([]metrics.Cycle, 0, len(sccs))
	for _, scc := range sccs {
		if len(scc) < 2 {
			continue
		}
		members := append([]string(nil), scc...)
		sort.Strings(members)
		cycles = append(cycles, metrics.Cycle(members))
	}
	sort.Slice(cycles, func(i, j int) bool {
		return cycles[i][0] < cycles[j][0]
	})

	if len(cycles) == 0 {
		return cycles, nil
	}

	confirmed := corroborateCycles(cycles, deps)
	warning := metrics.Warning{
		Code: warningCycleApproximation,
		Message: fmt.Sprintf(
			"cycles are the adapter's own directory-grain SCC derivation; "+
				"ripwire's capped file-level listing (%d SCCs × %d members) ",
			ripwireCycleMaxSCCs, ripwireCycleMaxMembers,
		) + corroborationVerdict(confirmed),
	}
	return cycles, []metrics.Warning{warning}
}

// corroborationVerdict renders the corroboration outcome for the warning
// message.
func corroborationVerdict(confirmed bool) string {
	if confirmed {
		return "corroborates the derived cycles"
	}
	return "does not corroborate them (its caps and file grain make it advisory only)"
}

// corroborateCycles maps ripwire's file-level cycle listing to directory
// grain and reports whether every derived cycle appears there. The Python
// fixtures routinely fail this check by construction: a two-directory
// import cycle whose file-level edges are acyclic produces no ripwire
// listing at all.
func corroborateCycles(cycles []metrics.Cycle, deps depsGraph) bool {
	if len(deps.Cycles) == 0 {
		return false
	}
	listed := make(map[string]bool, len(deps.Cycles))
	for _, cycle := range deps.Cycles {
		dirs := make([]string, 0, len(cycle.Members))
		for _, member := range cycle.Members {
			dir := member
			if i := strings.LastIndex(dir, "/"); i >= 0 {
				dir = dir[:i]
			} else {
				dir = "."
			}
			dirs = append(dirs, dir)
		}
		sort.Strings(dirs)
		listed[strings.Join(dirs, "\x00")] = true
	}
	for _, cycle := range cycles {
		if !listed[strings.Join(cycle, "\x00")] {
			return false
		}
	}
	return true
}

// tarjanSCCs computes the strongly connected components of the graph over
// nodes with the given adjacency. Deterministic: nodes are visited in
// sorted order and adjacency lists are pre-sorted, so the component
// partition (though not the emission order) is stable for identical
// input; callers sort the final output canonically.
func tarjanSCCs(nodes []string, adjacency map[string][]string) [][]string {
	index := make(map[string]int, len(nodes))
	low := make(map[string]int, len(nodes))
	onStack := make(map[string]bool, len(nodes))
	var stack []string
	var sccs [][]string
	nextIndex := 0

	for _, node := range nodes {
		adjacency[node] = sortedCopy(adjacency[node])
	}

	var strongConnect func(v string)
	strongConnect = func(v string) {
		index[v] = nextIndex
		low[v] = nextIndex
		nextIndex++
		stack = append(stack, v)
		onStack[v] = true

		for _, w := range adjacency[v] {
			if _, visited := index[w]; !visited {
				strongConnect(w)
				if low[w] < low[v] {
					low[v] = low[w]
				}
			} else if onStack[w] {
				if index[w] < low[v] {
					low[v] = index[w]
				}
			}
		}

		if low[v] == index[v] {
			var scc []string
			for {
				w := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				onStack[w] = false
				scc = append(scc, w)
				if w == v {
					break
				}
			}
			sccs = append(sccs, scc)
		}
	}

	for _, node := range nodes {
		if _, visited := index[node]; !visited {
			strongConnect(node)
		}
	}
	return sccs
}

// sortedCopy returns a sorted copy of s.
func sortedCopy(s []string) []string {
	out := append([]string(nil), s...)
	sort.Strings(out)
	return out
}
