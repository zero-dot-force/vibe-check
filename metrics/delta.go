package metrics

import (
	"sort"
	"strings"
)

// EntropyDirection classifies the net architectural-entropy movement of a PR
// relative to its base. It mirrors the [Status] typed-string const style. This
// field is reporting-only: it never affects the [Verdict] (both EntropyImproving
// and EntropyStable map to APPROVE).
type EntropyDirection string

const (
	// EntropyImproving indicates metrics net-improved: no gate was crossed and
	// at least one metric improved by at least the corresponding default gate
	// magnitude (or a cycle was resolved).
	EntropyImproving EntropyDirection = "improving"
	// EntropyStable indicates no material movement in either direction (or
	// offsetting shifts that cross no gate), and is also used whenever the
	// measurement is unreliable.
	EntropyStable EntropyDirection = "stable"
	// EntropyDegrading indicates at least one REQUEST_CHANGES gate was crossed
	// (a new cycle, or a rounded ΔInstability ≥ 0.15, ΔDistance ≥ 0.20, or
	// ΔLCOM ≥ 2 under the default protected thresholds).
	EntropyDegrading EntropyDirection = "degrading"
)

// Delta captures the per-module change (PR value minus base value) for a module
// present in BOTH the base and PR graphs, matched by [Module.Path]. Raw
// differences are stored unrounded; the 4-decimal rounding rule is applied only
// at comparison time in [DecideVerdict] and direction classification.
type Delta struct {
	// Path is the module import path this delta describes (present in both
	// graphs). It is the match key between the base and PR graphs.
	Path string `json:"path"`
	// Ca is the change in afferent coupling (PR Ca minus base Ca). Unit: signed
	// integer count of dependents.
	Ca int `json:"ca"`
	// Ce is the change in efferent coupling (PR Ce minus base Ce). Unit: signed
	// integer count of dependencies.
	Ce int `json:"ce"`
	// Instability is the change in instability (PR I minus base I). Unit: signed
	// ratio delta in [-1.0, 1.0]; positive means more unstable.
	Instability float64 `json:"instability"`
	// Abstractness is the change in abstractness (PR A minus base A). Unit:
	// signed ratio delta in [-1.0, 1.0]; positive means more abstract.
	Abstractness float64 `json:"abstractness"`
	// Distance is the change in distance from the main sequence (PR D minus base
	// D). Unit: signed ratio delta in [-1.0, 1.0]; positive means farther from
	// the main sequence.
	Distance float64 `json:"distance"`
	// LCOM is the change in LCOM4 cohesion (PR LCOM minus base LCOM). Unit:
	// signed integer; positive means less cohesive.
	LCOM int `json:"lcom"`
}

// GraphDelta is the complete, deterministic comparison between a base and a PR
// [ModuleGraph]. All slice fields are emitted in a stable, sorted order.
type GraphDelta struct {
	// Modules holds the per-module deltas for modules present in both graphs,
	// sorted ascending by Path.
	Modules []Delta `json:"modules"`
	// Added lists import paths present only in the PR graph, sorted ascending.
	// It is REPORT-ONLY and never feeds any gate or verdict; it is suppressed
	// (left empty) when Unreliable is true.
	Added []string `json:"added"`
	// Removed lists import paths present only in the base graph, sorted
	// ascending. It is REPORT-ONLY and never feeds any gate or verdict; it is
	// suppressed (left empty) when Unreliable is true.
	Removed []string `json:"removed"`
	// NewCycles lists cycles present in the PR graph but absent from the base
	// graph, sorted by their member-set key. Members within each cycle are
	// normalized (sorted lexicographically).
	NewCycles []Cycle `json:"newCycles"`
	// ResolvedCycles lists cycles present in the base graph but absent from the
	// PR graph, sorted by their member-set key. Members within each cycle are
	// normalized (sorted lexicographically).
	ResolvedCycles []Cycle `json:"resolvedCycles"`
	// Direction is the reporting-only net-entropy classification. It is
	// EntropyStable whenever Unreliable is true.
	Direction EntropyDirection `json:"direction"`
	// Unreliable is true when either input graph had a non-complete Status or a
	// load-error warning, marking the measurement as untrustworthy. When true,
	// [DecideVerdict] forces COMMENT.
	Unreliable bool `json:"unreliable"`
}

// ComputeDelta computes the per-module and structural delta between a base and a
// PR [ModuleGraph]. It is a PURE, DETERMINISTIC function: no I/O, no globals, no
// error return, and identical inputs always yield identical output.
//
// Behavior:
//   - If either input is nil, it returns a GraphDelta with Unreliable set,
//     Direction EntropyStable, and empty (non-nil) slices — never panicking.
//   - The measurement is marked Unreliable when either graph has
//     Status != StatusComplete or carries any Warning with Code "load-error".
//     When Unreliable, Added/Removed are suppressed and Direction is
//     EntropyStable.
//   - Modules are matched by [Module.Path]: modules in both graphs produce a
//     [Delta] (PR field minus base field); modules only in the PR graph are
//     appended to Added; modules only in the base graph are appended to Removed.
//     Added/Removed modules produce no Delta entry and no gate effect.
//   - Cycle identity is the sorted member set (rotation- and order-invariant).
//     NewCycles are in the PR but not the base; ResolvedCycles are in the base
//     but not the PR; cycles in both are pre-existing and excluded from both.
//
// All output slices are emitted in a stable, sorted order.
func ComputeDelta(base, pr *ModuleGraph) GraphDelta {
	if base == nil || pr == nil {
		return GraphDelta{
			Modules:        []Delta{},
			Added:          []string{},
			Removed:        []string{},
			NewCycles:      []Cycle{},
			ResolvedCycles: []Cycle{},
			Direction:      EntropyStable,
			Unreliable:     true,
		}
	}

	unreliable := graphUnreliable(base) || graphUnreliable(pr)

	baseByPath := indexModules(base.Modules)
	prByPath := indexModules(pr.Modules)

	modules := make([]Delta, 0, len(prByPath))
	added := make([]string, 0)
	for path, prm := range prByPath {
		bm, ok := baseByPath[path]
		if !ok {
			added = append(added, path)
			continue
		}
		modules = append(modules, Delta{
			Path:         path,
			Ca:           prm.Ca - bm.Ca,
			Ce:           prm.Ce - bm.Ce,
			Instability:  float64(prm.Instability) - float64(bm.Instability),
			Abstractness: float64(prm.Abstractness) - float64(bm.Abstractness),
			Distance:     float64(prm.Distance) - float64(bm.Distance),
			LCOM:         int(prm.LCOM) - int(bm.LCOM),
		})
	}

	removed := make([]string, 0)
	for path := range baseByPath {
		if _, ok := prByPath[path]; !ok {
			removed = append(removed, path)
		}
	}

	newCycles, resolvedCycles := diffCycles(base.Cycles, pr.Cycles)

	// Added/Removed are report-only; suppress them entirely when the
	// measurement is unreliable so no downstream consumer treats a
	// build-artifact difference as signal.
	if unreliable {
		added = []string{}
		removed = []string{}
	}

	sort.Slice(modules, func(i, j int) bool { return modules[i].Path < modules[j].Path })
	sort.Strings(added)
	sort.Strings(removed)

	direction := EntropyStable
	if !unreliable {
		direction = computeDirection(modules, newCycles, resolvedCycles)
	}

	return GraphDelta{
		Modules:        modules,
		Added:          added,
		Removed:        removed,
		NewCycles:      newCycles,
		ResolvedCycles: resolvedCycles,
		Direction:      direction,
		Unreliable:     unreliable,
	}
}

// graphUnreliable reports whether a graph indicates an unreliable measurement:
// a non-complete Status or any load-error warning.
func graphUnreliable(g *ModuleGraph) bool {
	if g.Status != StatusComplete {
		return true
	}
	for _, w := range g.Warnings {
		if w.Code == "load-error" {
			return true
		}
	}
	return false
}

// indexModules builds a path-keyed lookup of module results. On duplicate paths
// (which a well-formed graph does not contain) the last entry wins.
func indexModules(mods []ModuleResult) map[string]ModuleResult {
	byPath := make(map[string]ModuleResult, len(mods))
	for _, m := range mods {
		byPath[m.Path] = m
	}
	return byPath
}

// diffCycles classifies cycles into those new in the PR graph and those
// resolved relative to the base graph, comparing by rotation-invariant member
// set. Returned slices are non-nil, de-duplicated by key, carry
// member-normalized cycles, and are sorted by their member-set key.
func diffCycles(baseCycles, prCycles []Cycle) (newCycles, resolvedCycles []Cycle) {
	baseKeys := cycleKeySet(baseCycles)
	prKeys := cycleKeySet(prCycles)

	newCycles = make([]Cycle, 0)
	seenNew := make(map[string]bool)
	for _, c := range prCycles {
		k := cycleKey(c)
		if !baseKeys[k] && !seenNew[k] {
			seenNew[k] = true
			newCycles = append(newCycles, normalizedCycle(c))
		}
	}

	resolvedCycles = make([]Cycle, 0)
	seenResolved := make(map[string]bool)
	for _, c := range baseCycles {
		k := cycleKey(c)
		if !prKeys[k] && !seenResolved[k] {
			seenResolved[k] = true
			resolvedCycles = append(resolvedCycles, normalizedCycle(c))
		}
	}

	sort.Slice(newCycles, func(i, j int) bool { return cycleKey(newCycles[i]) < cycleKey(newCycles[j]) })
	sort.Slice(resolvedCycles, func(i, j int) bool { return cycleKey(resolvedCycles[i]) < cycleKey(resolvedCycles[j]) })
	return newCycles, resolvedCycles
}

// cycleKeySet returns the set of member-set keys for a slice of cycles.
func cycleKeySet(cycles []Cycle) map[string]bool {
	keys := make(map[string]bool, len(cycles))
	for _, c := range cycles {
		keys[cycleKey(c)] = true
	}
	return keys
}

// cycleKey returns the rotation- and order-invariant identity of a cycle: its
// members sorted lexicographically and joined with a NUL ("\x00") separator so
// that member strings cannot collide across boundaries.
func cycleKey(c Cycle) string {
	members := append([]string(nil), c...)
	sort.Strings(members)
	return strings.Join(members, "\x00")
}

// normalizedCycle returns a copy of the cycle with members sorted
// lexicographically, matching the [Cycle] contract of a sorted membership set.
func normalizedCycle(c Cycle) Cycle {
	members := append([]string(nil), c...)
	sort.Strings(members)
	return Cycle(members)
}

// computeDirection derives the reporting-only [EntropyDirection] from the
// per-module deltas and cycle changes using the default protected thresholds
// and the shared 4-decimal round-half-away-from-zero rule. A crossed gate wins
// (degrading); otherwise a sufficiently large improvement or a resolved cycle
// yields improving; otherwise stable.
func computeDirection(modules []Delta, newCycles, resolvedCycles []Cycle) EntropyDirection {
	t := DefaultVerdictThresholds()

	gateCrossed := len(newCycles) > 0
	improved := len(resolvedCycles) > 0

	for _, m := range modules {
		ri := roundTo4(m.Instability)
		rd := roundTo4(m.Distance)
		if ri >= t.MaxInstabilityDelta || rd >= t.MaxDistanceDelta || m.LCOM >= t.MaxLCOMDelta {
			gateCrossed = true
		}
		if ri <= -t.MaxInstabilityDelta || rd <= -t.MaxDistanceDelta || m.LCOM <= -t.MaxLCOMDelta {
			improved = true
		}
	}

	switch {
	case gateCrossed:
		return EntropyDegrading
	case improved:
		return EntropyImproving
	default:
		return EntropyStable
	}
}
