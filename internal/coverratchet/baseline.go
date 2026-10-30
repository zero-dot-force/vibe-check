package coverratchet

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
)

// BaselineVersion is the only baseline schema version this package reads
// and writes.
const BaselineVersion = 1

// Baseline is the parsed representation of a coverage-floors.json file: a
// schema version and a map of module-relative package path to its recorded
// coverage floor percentage.
type Baseline struct {
	Version int                `json:"version"`
	Floors  map[string]float64 `json:"floors"`
}

// ReadBaseline reads and validates the coverage floors baseline at path.
// A missing file, unparseable JSON, an unknown schema version, or a floor
// outside the value domain — (0, 100] with at most one decimal place — is
// a tool error.
func ReadBaseline(path string) (Baseline, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Baseline{}, fmt.Errorf("coverratchet: read coverage floors %s: %w", path, err)
	}
	var b Baseline
	if err := json.Unmarshal(data, &b); err != nil {
		return Baseline{}, fmt.Errorf("coverratchet: parse coverage floors %s: %w", path, err)
	}
	if b.Version != BaselineVersion {
		return Baseline{}, fmt.Errorf("coverratchet: coverage floors %s: unsupported version %d (want %d)", path, b.Version, BaselineVersion)
	}
	if err := validateFloors(b.Floors); err != nil {
		return Baseline{}, fmt.Errorf("coverratchet: coverage floors %s: %w", path, err)
	}
	return b, nil
}

// WriteBaseline writes the baseline to path as deterministic JSON:
// lexicographically sorted package keys, two-space indentation, and a
// trailing newline, so repeated writes of unchanged floors produce a
// byte-identical file and baseline updates yield minimal, reviewable
// diffs. The write is atomic — the file is staged next to its target and
// renamed into place — so an interrupted update cannot leave a truncated
// baseline behind; a failed rename removes the staged file, leaving no
// .tmp residue.
func WriteBaseline(path string, b Baseline) error {
	if b.Version != BaselineVersion {
		return fmt.Errorf("coverratchet: write coverage floors %s: unsupported version %d (want %d)", path, b.Version, BaselineVersion)
	}
	if err := validateFloors(b.Floors); err != nil {
		return fmt.Errorf("coverratchet: write coverage floors %s: %w", path, err)
	}
	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return fmt.Errorf("coverratchet: encode coverage floors %s: %w", path, err)
	}
	data = append(data, '\n')
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("coverratchet: stage coverage floors %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("coverratchet: commit coverage floors %s: %w", path, err)
	}
	return nil
}

// validateFloors enforces the floor value domain: every floor must lie in
// (0, 100] and carry at most one decimal place. A floor outside the domain
// would either disable the gate for its package (at or below zero) or
// permanently fail it (above 100), so both are rejected as malformed input.
func validateFloors(floors map[string]float64) error {
	for pkg, floor := range floors {
		if floor <= 0 || floor > 100 {
			return fmt.Errorf("package %q floor %v out of range (0, 100]", pkg, floor)
		}
		if math.Round(floor*10)/10 != floor {
			return fmt.Errorf("package %q floor %v exceeds one-decimal precision", pkg, floor)
		}
	}
	return nil
}
