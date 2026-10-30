// Package coverratchet enforces per-package Go statement-coverage floors
// against a committed baseline, implementing the repository's coverage
// ratchet gate: coverage may ratchet upward, but a regression below a
// recorded floor is a policy failure that blocks the build.
//
// # Role
//
// coverratchet is the engine behind cmd/coverage-ratchet. It parses a Go
// coverage profile produced by `go test -coverprofile`, aggregates
// statement-weighted per-package coverage, compares it against the floors
// recorded in the committed coverage-floors.json baseline, and reports
// floor and registration violations. Floors are tighten-only: the update
// mode raises existing floors to higher measured values and registers new
// packages only at or above 80.0 percent coverage. No code path lowers a
// floor — loosening requires a direct, human-reviewed edit of the baseline
// file, because coverage thresholds are protected gatekeeping values.
//
// # Package Keys
//
// Profile file names are recorded as module paths plus a file name (e.g.
// "github.com/zero-dot-force/vibe-check/metrics/values.go"). The package
// universe is keyed by the module-relative package path between the module
// prefix and the file name ("metrics" above). A profile file that does not
// carry the module prefix is a tool error.
//
// # Measurement
//
// Per-package coverage is statement-weighted: covered is the sum of NumStmt
// over blocks with a non-zero Count, total is the sum of NumStmt over all
// blocks, and the percentage (covered/total × 100) is truncated (rounded
// down) to one decimal place. Truncation is applied identically to
// measurements and stored floors, so a floor recorded by an update can
// never exceed what the same run measured.
package coverratchet
