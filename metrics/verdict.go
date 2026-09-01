package metrics

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// Verdict is the outcome of applying the entropy-regression gates to a
// [GraphDelta]. Exactly one verdict is rendered per comparison.
type Verdict string

const (
	// VerdictApprove indicates the PR improves or holds metrics steady: no gate
	// fired and no material worsening was detected.
	VerdictApprove Verdict = "APPROVE"
	// VerdictComment indicates a non-blocking regression: metrics shifted for
	// the worse but stayed below every REQUEST_CHANGES threshold, or the
	// measurement was unreliable (partial build) and was downgraded.
	VerdictComment Verdict = "COMMENT"
	// VerdictRequestChanges indicates a blocking regression: at least one gate
	// threshold was met or exceeded, or the PR introduced a new cycle.
	VerdictRequestChanges Verdict = "REQUEST_CHANGES"
)

// VerdictThresholds holds the per-metric delta gates that drive [DecideVerdict].
// These are PROTECTED quality gates: overrides may only ever TIGHTEN them (make
// them smaller). They MUST NOT be weakened (loosened) to let a larger regression
// pass, per the AGENTS.md gatekeeping mandate.
type VerdictThresholds struct {
	// MaxInstabilityDelta is the inclusive instability-increase gate. A rounded
	// per-module ΔInstability greater than or equal to this value triggers
	// REQUEST_CHANGES. Unit: dimensionless ratio in [0.0, 1.0]. Default: 0.15.
	MaxInstabilityDelta float64
	// MaxDistanceDelta is the inclusive distance-increase gate. A rounded
	// per-module ΔDistance greater than or equal to this value triggers
	// REQUEST_CHANGES. Unit: dimensionless ratio in [0.0, 1.0]. Default: 0.20.
	MaxDistanceDelta float64
	// MaxLCOMDelta is the inclusive LCOM-increase gate. A per-module ΔLCOM
	// greater than or equal to this value triggers REQUEST_CHANGES. Unit:
	// non-negative integer (connected components). Default: 2.
	MaxLCOMDelta int
}

// DefaultVerdictThresholds returns the protected default gate values:
// ΔInstability ≥ 0.15, ΔDistance ≥ 0.20, and ΔLCOM ≥ 2. These defaults are
// quality gates and MUST NOT be loosened; callers may only tighten them.
func DefaultVerdictThresholds() VerdictThresholds {
	return VerdictThresholds{
		MaxInstabilityDelta: 0.15,
		MaxDistanceDelta:    0.20,
		MaxLCOMDelta:        2,
	}
}

// roundTo4 rounds x to 4 decimal places (precision 1e-4) using
// round-half-away-from-zero semantics. Go's [math.Round] rounds halves away
// from zero, so 0.14999 → 0.1500 and -0.14999 → -0.1500.
func roundTo4(x float64) float64 {
	return math.Round(x*1e4) / 1e4
}

// DecideVerdict renders a deterministic [Verdict] for a computed [GraphDelta]
// against a set of [VerdictThresholds], returning the verdict and a
// stable-sorted, machine-readable list of the reasons each gate fired.
//
// DecideVerdict is self-contained: it recomputes the outcome from d.Modules,
// d.NewCycles, and d.Unreliable and does NOT consult d.Direction (which is a
// reporting-only field).
//
// Rounding: every float delta is rounded to 4 decimal places (precision 1e-4)
// using round-half-away-from-zero (math.Round(x*1e4)/1e4) BEFORE it is
// compared. Consequently a raw ΔInstability of 0.14999 rounds to 0.1500 and
// meets the 0.15 gate, whereas 0.14994 rounds to 0.1499 and does not.
//
// Comparison direction: the gates are INCLUSIVE (≥). A regression of exactly
// the threshold amount fires the gate. This is deliberately stricter than
// analyze.go's checkThresholds, which uses a strict > comparison for its
// --max-* flags — a regression equal to an entropy gate is already too much.
//
// Precedence:
//
//  1. If d.Unreliable (partial build), the verdict is forced to COMMENT — never
//     APPROVE — with a single partial-build reason.
//  2. Otherwise, if any gate fires (a new cycle, or a rounded ΔInstability ≥
//     MaxInstabilityDelta, ΔDistance ≥ MaxDistanceDelta, or ΔLCOM ≥
//     MaxLCOMDelta), the verdict is REQUEST_CHANGES with one reason per gate.
//  3. Otherwise, if any sub-gate worsening exists (a rounded ΔInstability > 0,
//     ΔDistance > 0, or ΔLCOM > 0, or a new cycle), the verdict is COMMENT.
//  4. Otherwise the verdict is APPROVE with no reasons.
//
// Added and removed packages never contribute to any gate or sub-gate.
func DecideVerdict(d GraphDelta, t VerdictThresholds) (Verdict, []string) {
	if d.Unreliable {
		return VerdictComment, []string{"partial-build: measurement unreliable, verdict downgraded to COMMENT"}
	}

	reasons := gateReasons(d, t)
	if len(reasons) > 0 {
		return VerdictRequestChanges, reasons
	}

	if hasMaterialShift(d) {
		return VerdictComment, []string{"materiality: non-zero metric shifts below REQUEST_CHANGES thresholds"}
	}

	return VerdictApprove, nil
}

// gateReasons collects the REQUEST_CHANGES reasons for a delta in a
// deterministic, stable-sorted order: new-cycle reasons first (sorted by their
// member-set key), then per-module reasons (sorted by module path, and within a
// module ordered instability, distance, LCOM).
func gateReasons(d GraphDelta, t VerdictThresholds) []string {
	var reasons []string

	newCycles := append([]Cycle(nil), d.NewCycles...)
	sort.Slice(newCycles, func(i, j int) bool {
		return cycleKey(newCycles[i]) < cycleKey(newCycles[j])
	})
	for _, c := range newCycles {
		members := append([]string(nil), c...)
		sort.Strings(members)
		reasons = append(reasons, fmt.Sprintf("new-cycle: %s", strings.Join(members, " ")))
	}

	mods := append([]Delta(nil), d.Modules...)
	sort.Slice(mods, func(i, j int) bool { return mods[i].Path < mods[j].Path })
	for _, m := range mods {
		if ri := roundTo4(m.Instability); ri >= t.MaxInstabilityDelta {
			reasons = append(reasons, fmt.Sprintf(
				"instability: module %q delta %.4f >= threshold %.4f",
				m.Path, ri, t.MaxInstabilityDelta,
			))
		}
		if rd := roundTo4(m.Distance); rd >= t.MaxDistanceDelta {
			reasons = append(reasons, fmt.Sprintf(
				"distance: module %q delta %.4f >= threshold %.4f",
				m.Path, rd, t.MaxDistanceDelta,
			))
		}
		if m.LCOM >= t.MaxLCOMDelta {
			reasons = append(reasons, fmt.Sprintf(
				"lcom: module %q delta %d >= threshold %d",
				m.Path, m.LCOM, t.MaxLCOMDelta,
			))
		}
	}

	return reasons
}

// hasMaterialShift reports whether any module worsened by a rounded, non-zero
// amount, or a new cycle exists. Improvements (negative deltas) and added or
// removed packages never count as a material shift.
func hasMaterialShift(d GraphDelta) bool {
	if len(d.NewCycles) > 0 {
		return true
	}
	for _, m := range d.Modules {
		if roundTo4(m.Instability) > 0 || roundTo4(m.Distance) > 0 || m.LCOM > 0 {
			return true
		}
	}
	return false
}
