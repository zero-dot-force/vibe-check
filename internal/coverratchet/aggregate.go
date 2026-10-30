package coverratchet

import (
	"golang.org/x/tools/cover"
)

// Coverage reports a package's statement-weighted coverage aggregated from
// coverage profile blocks.
type Coverage struct {
	// Covered is the sum of NumStmt over blocks with a non-zero Count. In
	// mode: atomic profiles (produced under -race) any non-zero count means
	// executed.
	Covered int

	// Total is the sum of NumStmt over all blocks.
	Total int
}

// Percent returns the package's coverage percentage (covered/total × 100)
// truncated (rounded down) to one decimal place. The truncation uses exact
// integer arithmetic, so a floor recorded from a measured percentage can
// never exceed what the same run measured. A zero Total reports 0 — below
// every valid floor, so the gate fails closed rather than panicking;
// Coverage values produced by [Aggregate] always have a positive Total
// because packages without coverable statements are excluded.
func (c Coverage) Percent() float64 {
	if c.Total == 0 {
		return 0
	}
	return float64(c.Covered*1000/c.Total) / 10
}

// Aggregate computes statement-weighted per-package coverage from profiles:
// for each package, covered is the sum of NumStmt over blocks with a
// non-zero Count and total is the sum of NumStmt over all blocks. Packages
// whose blocks sum to zero statements have no coverable statements and are
// excluded from the result, so the returned map is exactly the set of
// packages the baseline must register. A profile file outside the module is
// a tool error.
func Aggregate(profiles []*cover.Profile) (map[string]Coverage, error) {
	agg := make(map[string]Coverage)
	for _, p := range profiles {
		key, err := PackageKey(p.FileName)
		if err != nil {
			return nil, err
		}
		c := agg[key]
		for _, b := range p.Blocks {
			c.Total += b.NumStmt
			if b.Count > 0 {
				c.Covered += b.NumStmt
			}
		}
		agg[key] = c
	}
	for key, c := range agg {
		if c.Total == 0 {
			delete(agg, key)
		}
	}
	return agg, nil
}
