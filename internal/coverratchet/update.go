package coverratchet

// EntryFloor is the minimum truncated measured coverage at which Update
// registers a previously unregistered package — the project's established
// bar for new code. Registering a package below it would silently weaken
// the gate relative to the coverage bar every prior change followed.
const EntryFloor = 80.0

// Update computes the tightened baseline implied by measured coverage: it
// raises an existing floor to a higher truncated measured value, adds an
// unregistered package at its truncated measured value when that value is
// at or above [EntryFloor], and leaves every other floor untouched. No
// invocation path lowers a floor — a package measured below its floor is
// reported as a below-floor violation with its entry kept, an unregistered
// package below EntryFloor is refused, and a baseline entry whose package
// has no coverable statements is reported as a stale-floor violation.
// Loosening or removing a floor requires a direct, human-reviewed edit of
// the baseline file.
//
// When any violation exists, Update returns the input baseline unchanged
// and the caller must not write it: a failed policy state never mutates
// the baseline file, so updates are all-or-nothing. Only when the result
// reports no violation does the returned baseline carry the raises and
// additions. A raised or added row reports the post-update floor — the
// value this update assigns the package — rather than the pre-update one,
// so the report reflects the ratchet's effect; when a violation refuses
// the update, the caller's "baseline not modified" line states that no
// floor was written.
func Update(measured map[string]Coverage, baseline Baseline) (Baseline, Result) {
	floors := make(map[string]float64, len(baseline.Floors))
	for pkg, floor := range baseline.Floors {
		floors[pkg] = floor
	}
	result := Result{Packages: make([]PackageResult, 0, len(measured)+len(floors))}
	for pkg, cov := range measured {
		percent := cov.Percent()
		floor, registered := floors[pkg]
		p := PackageResult{Package: pkg, Measured: percent, Floor: floor}
		switch {
		case !registered && percent < EntryFloor:
			p.Status = StatusRefused
		case !registered:
			p.Status = StatusAdded
			floors[pkg] = percent
			p.Floor = percent
		case percent > floor:
			p.Status = StatusRaised
			floors[pkg] = percent
			p.Floor = percent
		case percent < floor:
			p.Status = StatusBelowFloor
		default:
			p.Status = StatusUnchanged
		}
		result.Packages = append(result.Packages, p)
	}
	result.finalize(measured, baseline)
	if !result.OK() {
		return baseline, result
	}
	return Baseline{Version: BaselineVersion, Floors: floors}, result
}
