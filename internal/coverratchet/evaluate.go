package coverratchet

import (
	"slices"
	"strings"
)

// Status classifies the evaluation of a single package against the
// baseline.
type Status string

const (
	// StatusPass marks a package whose truncated measured coverage equals
	// or exceeds its recorded floor.
	StatusPass Status = "pass"
	// StatusBelowFloor marks a package whose truncated measured coverage
	// is below its recorded floor.
	StatusBelowFloor Status = "below-floor"
	// StatusUnregistered marks a package with coverable statements that
	// has no baseline entry.
	StatusUnregistered Status = "unregistered"
	// StatusStaleFloor marks a baseline entry whose package has no
	// coverable statements in the coverage profile.
	StatusStaleFloor Status = "stale-floor"
	// StatusRaised marks a package whose floor an update raises to its
	// higher truncated measured coverage.
	StatusRaised Status = "raised"
	// StatusAdded marks a previously unregistered package that an update
	// registers at its truncated measured coverage at or above EntryFloor.
	StatusAdded Status = "added"
	// StatusUnchanged marks a package whose truncated measured coverage
	// equals its recorded floor, leaving the entry as-is on update.
	StatusUnchanged Status = "unchanged"
	// StatusRefused marks an unregistered package whose truncated measured
	// coverage is below EntryFloor, which an update refuses to register.
	StatusRefused Status = "refused"
)

// PackageResult reports the floor evaluation of a single package.
type PackageResult struct {
	// Package is the module-relative package path.
	Package string
	// Measured is the package's truncated statement coverage percentage.
	// It is zero when the package was not present in the coverage profile.
	Measured float64
	// Floor is the package's recorded coverage floor. It is zero when the
	// package has no baseline entry.
	Floor float64
	// Status classifies the evaluation outcome.
	Status Status
}

// Result is the outcome of evaluating measured per-package coverage
// against a baseline. Packages holds one entry per package in the union of
// the measured packages and the baseline floors, sorted by package path.
type Result struct {
	Packages []PackageResult
}

// OK reports whether the evaluation found no violation: every measured
// package is registered and at or above its floor, and every baseline
// entry corresponds to a measured package. Update results additionally
// treat raised, added, and unchanged packages as non-violating.
func (r Result) OK() bool {
	for _, p := range r.Packages {
		switch p.Status {
		case StatusPass, StatusRaised, StatusAdded, StatusUnchanged:
		default:
			return false
		}
	}
	return true
}

// finalize completes a result under assembly with the steps Evaluate and
// Update share: it appends a StatusStaleFloor row for every baseline entry
// whose package is absent from measured — the baseline must mirror the
// module's coverable packages exactly — and sorts the rows by package path,
// yielding the canonical deterministic result ordering.
func (r *Result) finalize(measured map[string]Coverage, baseline Baseline) {
	for pkg, floor := range baseline.Floors {
		if _, ok := measured[pkg]; ok {
			continue
		}
		r.Packages = append(r.Packages, PackageResult{
			Package: pkg,
			Floor:   floor,
			Status:  StatusStaleFloor,
		})
	}
	slices.SortFunc(r.Packages, func(a, b PackageResult) int {
		return strings.Compare(a.Package, b.Package)
	})
}

// Evaluate compares measured per-package coverage against the baseline
// floors. A package whose truncated measured coverage is below its floor
// is a below-floor violation; measured == floor passes. A package with
// coverable statements that has no baseline entry is an unregistered
// violation, forcing explicit registration of new packages. A baseline
// entry whose package has no coverable statements in the profile is a
// stale-floor violation: the baseline must mirror the module's coverable
// packages exactly, so a deleted package's floor must be removed by a
// reviewed baseline edit and an emptied or narrowed profile cannot pass
// vacuously.
func Evaluate(measured map[string]Coverage, baseline Baseline) Result {
	result := Result{Packages: make([]PackageResult, 0, len(measured)+len(baseline.Floors))}
	for pkg, cov := range measured {
		floor, registered := baseline.Floors[pkg]
		p := PackageResult{Package: pkg, Measured: cov.Percent(), Floor: floor}
		switch {
		case !registered:
			p.Status = StatusUnregistered
		case p.Measured < floor:
			p.Status = StatusBelowFloor
		default:
			p.Status = StatusPass
		}
		result.Packages = append(result.Packages, p)
	}
	result.finalize(measured, baseline)
	return result
}
