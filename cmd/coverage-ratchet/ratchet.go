package main

import (
	"errors"
	"fmt"
	"io"
	"runtime/debug"

	"github.com/spf13/cobra"

	"github.com/zero-dot-force/vibe-check/internal/coverratchet"
)

// versionString returns the --version output string in the form
// "<version> (commit <hash>, built <date>)". When the ldflags-injected version
// is empty or the default "dev" (e.g., the tool runs via `go run`), it falls
// back to build information embedded by the Go toolchain: the main module
// version and the vcs.revision / vcs.time build settings. This ensures
// --version reports meaningful data even without explicit ldflags.
func versionString() string {
	v, c, d := version, commit, date
	if v == "" || v == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok {
			if mv := info.Main.Version; mv != "" && mv != "(devel)" {
				v = mv
			}
			for _, s := range info.Settings {
				switch s.Key {
				case "vcs.revision":
					if s.Value != "" {
						c = s.Value
					}
				case "vcs.time":
					if s.Value != "" {
						d = s.Value
					}
				}
			}
		}
	}
	return fmt.Sprintf("%s (commit %s, built %s)", v, c, d)
}

// semanticVersion returns the semantic version only (no commit/date
// decoration). It applies the same ldflags → build-info fallback as
// versionString: when the ldflags-injected version is empty or the default
// "dev", it falls back to the main module version embedded by the Go
// toolchain.
func semanticVersion() string {
	v := version
	if v == "" || v == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok {
			if mv := info.Main.Version; mv != "" && mv != "(devel)" {
				v = mv
			}
		}
	}
	return v
}

// RatchetOptions configures a gate run.
type RatchetOptions struct {
	// Profile is the path to the go test coverage profile.
	Profile string
	// Floors is the path to the committed coverage floors baseline.
	Floors string
	// Update tightens the baseline (raise floors, register new packages
	// at or above the entry bar) instead of only checking it.
	Update bool
	// JSON emits a machine-readable JSON report instead of the table.
	JSON bool
	// Stdout receives the report.
	Stdout io.Writer
}

// RatchetResult contains the gate run outcome.
type RatchetResult struct {
	// Result is the per-package evaluation produced by the engine.
	Result coverratchet.Result
	// Updated reports whether the baseline file was rewritten (update
	// mode with no violations).
	Updated bool
	// ExitCode is the process exit code: 0 success, 1 policy failure,
	// 2 tool failure.
	ExitCode int
}

// RunRatchet executes the coverage gate and returns the outcome. It is the
// testable entry point: all business logic lives here, not in the cobra
// command layer.
//
// The run loads and aggregates the coverage profile, reads the baseline,
// then either evaluates measured coverage against the floors or — with
// Update set — computes the tightened baseline and writes it back when no
// violation exists. The report is written to opts.Stdout as a table (or a
// JSON object with a provenance envelope when JSON is set).
//
// Exit code semantics:
//   - 0: success, no violations
//   - 1: policy failure (package below floor, unregistered package, stale
//     floor, or a refused registration in update mode)
//   - 2: tool failure (unreadable or malformed profile or baseline)
func RunRatchet(opts RatchetOptions) (*RatchetResult, error) {
	profiles, err := coverratchet.LoadProfiles(opts.Profile)
	if err != nil {
		return &RatchetResult{ExitCode: 2}, err
	}
	measured, err := coverratchet.Aggregate(profiles)
	if err != nil {
		return &RatchetResult{ExitCode: 2}, err
	}
	baseline, err := coverratchet.ReadBaseline(opts.Floors)
	if err != nil {
		return &RatchetResult{ExitCode: 2}, err
	}
	result := &RatchetResult{ExitCode: 0}
	if opts.Update {
		updated, report := coverratchet.Update(measured, baseline)
		result.Result = report
		if report.OK() {
			if err := coverratchet.WriteBaseline(opts.Floors, updated); err != nil {
				return &RatchetResult{ExitCode: 2}, err
			}
			result.Updated = true
		}
	} else {
		result.Result = coverratchet.Evaluate(measured, baseline)
	}
	if err := writeReport(opts.Stdout, opts, result); err != nil {
		return &RatchetResult{ExitCode: 2}, err
	}
	if !result.Result.OK() {
		result.ExitCode = 1
	}
	return result, nil
}

// rootCmd creates the root cobra command for coverage-ratchet.
func rootCmd() *cobra.Command {
	opts := RatchetOptions{}
	cmd := &cobra.Command{
		Use:   "coverage-ratchet",
		Short: "Enforce per-package Go statement-coverage floors against a committed baseline",
		Long: `coverage-ratchet parses a go test coverage profile, computes statement-weighted
per-package coverage, and compares it against the floors recorded in a committed
coverage-floors.json baseline. Floors are tighten-only: --update raises existing
floors and registers new packages at or above 80%; no invocation lowers a floor.`,
		Version:      versionString(),
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Stdout = cmd.OutOrStdout()
			result, err := RunRatchet(opts)
			if err != nil {
				// Cobra prints the error to stderr (SilenceErrors is not
				// set); the carried code reaches main() via exitCodeError.
				return &exitCodeError{code: result.ExitCode, err: err}
			}
			if result.ExitCode != 0 {
				// Violations are reported in the table on stdout. Return an
				// exitCodeError so main() sets exit code 1; cobra prints the
				// summary message to stderr.
				msg := "coverage floor violations detected"
				if opts.Update {
					msg = "coverage floor update refused: violations detected"
				}
				return &exitCodeError{
					code: result.ExitCode,
					err:  errors.New(msg),
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&opts.Profile, "profile", "coverage.out", "path to the go test coverage profile")
	cmd.Flags().StringVar(&opts.Floors, "floors", "coverage-floors.json", "path to the committed coverage floors baseline")
	cmd.Flags().BoolVar(&opts.Update, "update", false, "tighten the baseline to measured coverage (never lowers a floor)")
	cmd.Flags().BoolVar(&opts.JSON, "json", false, "emit a machine-readable JSON report with a provenance envelope")

	return cmd
}
