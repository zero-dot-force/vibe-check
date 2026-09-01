package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/zero-dot-force/vibe-check/internal/goadapter"
	"github.com/zero-dot-force/vibe-check/metrics"
)

// AnalyzeOptions contains the configuration for the analyze command.
// It follows the AP-001 Options struct pattern for testable CLI commands.
type AnalyzeOptions struct {
	// Stdout is the writer for JSON output. Required.
	Stdout io.Writer
	// Stderr is the writer for violation reports and errors. Required.
	Stderr io.Writer
	// Path is the project directory to analyze.
	Path string
	// OutputPath, when non-empty, names a file to which the ModuleGraph JSON is
	// written (mode 0o644) instead of Stdout. Empty (the default) writes to Stdout.
	OutputPath string

	// MaxInstability is the threshold for instability violations.
	// nil means not set (no threshold check). Must be in [0.0, 1.0].
	MaxInstability *float64
	// MaxDistance is the threshold for distance-from-main-sequence violations.
	// nil means not set (no threshold check). Must be in [0.0, 1.0].
	MaxDistance *float64
	// NoCircularDeps when true treats any detected cycle as a violation.
	NoCircularDeps bool
	// MaxLCOM is the threshold for LCOM violations.
	// nil means not set (no threshold check). Must be >= 1.
	MaxLCOM *int

	// Timeout is the analysis timeout duration. Zero means no timeout.
	Timeout time.Duration
}

// AnalyzeResult contains the analysis outcome.
// It follows the AP-001 Result struct pattern.
type AnalyzeResult struct {
	// Graph is the computed module graph. May be nil if analysis failed.
	Graph *metrics.ModuleGraph
	// Violations is the list of threshold violation messages.
	Violations []string
	// ExitCode is the process exit code: 0 success, 1 policy failure, 2 tool failure.
	ExitCode int
}

// RunAnalyze executes the analysis and returns a result.
// This is the testable entry point per AP-002/AP-003: all business logic
// lives here, not in the cobra command layer.
//
// Exit code semantics:
//   - 0: success, no violations
//   - 1: analysis succeeded, threshold violations detected (policy failure)
//   - 2: tool failure (invalid args, adapter error, timeout, signal)
func RunAnalyze(ctx context.Context, opts AnalyzeOptions) (*AnalyzeResult, error) {
	// Step 1: Validate flags.
	if err := validateFlags(opts); err != nil {
		return &AnalyzeResult{ExitCode: 2}, err
	}

	// Step 2: Apply timeout if configured.
	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	// Step 3: Run analysis via the Go adapter.
	adapter := goadapter.New()
	graph, err := adapter.Analyze(ctx, opts.Path)
	if err != nil {
		return &AnalyzeResult{ExitCode: 2}, fmt.Errorf("analyze: %w", err)
	}

	// Step 4: Check context before writing output.
	// Spec: "no partial JSON MUST be written to stdout" on signal/timeout.
	if ctx.Err() != nil {
		return &AnalyzeResult{ExitCode: 2}, fmt.Errorf("analyze: %w", ctx.Err())
	}

	// Step 5: Marshal and write JSON to stdout (always, even when violations exist).
	data, err := json.MarshalIndent(graph, "", "  ")
	if err != nil {
		return &AnalyzeResult{ExitCode: 2}, fmt.Errorf("analyze: marshal: %w", err)
	}

	// Final context check before writing — prevent partial output.
	if ctx.Err() != nil {
		return &AnalyzeResult{ExitCode: 2}, fmt.Errorf("analyze: %w", ctx.Err())
	}

	// Step 5b: Emit the graph. With --output set, write to the file and NOTHING
	// to stdout (not even a partial write on error); otherwise preserve the
	// stdout default. Threshold checking still runs after a successful write.
	if opts.OutputPath != "" {
		if err := os.WriteFile(opts.OutputPath, data, 0o644); err != nil {
			return &AnalyzeResult{ExitCode: 2}, fmt.Errorf("write output file %s: %w", opts.OutputPath, err)
		}
	} else if _, err := fmt.Fprintln(opts.Stdout, string(data)); err != nil {
		return &AnalyzeResult{ExitCode: 2}, fmt.Errorf("write output: %w", err)
	}

	// Step 6: Check thresholds and collect violations.
	// Uses strict > comparison: metric == threshold passes.
	violations := checkThresholds(graph, opts)

	// Step 7: Report violations to stderr.
	for _, v := range violations {
		_, _ = fmt.Fprintln(opts.Stderr, v)
	}

	exitCode := 0
	if len(violations) > 0 {
		exitCode = 1
	}

	return &AnalyzeResult{
		Graph:      graph,
		Violations: violations,
		ExitCode:   exitCode,
	}, nil
}

// validateFlags checks that all threshold flag values are within valid ranges.
// Returns an error describing the first invalid flag found.
func validateFlags(opts AnalyzeOptions) error {
	if opts.MaxInstability != nil {
		v := *opts.MaxInstability
		if v < 0.0 || v > 1.0 {
			return fmt.Errorf("invalid --max-instability value %.2f: must be in [0.0, 1.0]", v)
		}
	}
	if opts.MaxDistance != nil {
		v := *opts.MaxDistance
		if v < 0.0 || v > 1.0 {
			return fmt.Errorf("invalid --max-distance value %.2f: must be in [0.0, 1.0]", v)
		}
	}
	if opts.MaxLCOM != nil {
		v := *opts.MaxLCOM
		if v < 1 {
			return fmt.Errorf("invalid --max-lcom value %d: must be >= 1", v)
		}
	}
	return nil
}

// checkThresholds compares module metrics against configured thresholds.
// Uses strict > (greater than) comparison: if metric == threshold, it passes.
// All violations are collected — does not short-circuit on first violation.
func checkThresholds(graph *metrics.ModuleGraph, opts AnalyzeOptions) []string {
	var violations []string

	for _, m := range graph.Modules {
		if opts.MaxInstability != nil {
			if float64(m.Instability) > *opts.MaxInstability {
				violations = append(violations, fmt.Sprintf(
					"VIOLATION: module %q instability %.2f exceeds threshold %.2f",
					m.Path, float64(m.Instability), *opts.MaxInstability,
				))
			}
		}
		if opts.MaxDistance != nil {
			if float64(m.Distance) > *opts.MaxDistance {
				violations = append(violations, fmt.Sprintf(
					"VIOLATION: module %q distance %.2f exceeds threshold %.2f",
					m.Path, float64(m.Distance), *opts.MaxDistance,
				))
			}
		}
		if opts.MaxLCOM != nil {
			if int(m.LCOM) > *opts.MaxLCOM {
				violations = append(violations, fmt.Sprintf(
					"VIOLATION: module %q lcom %d exceeds threshold %d",
					m.Path, int(m.LCOM), *opts.MaxLCOM,
				))
			}
		}
	}

	if opts.NoCircularDeps && len(graph.Cycles) > 0 {
		for _, cycle := range graph.Cycles {
			violations = append(violations, fmt.Sprintf(
				"VIOLATION: circular dependency detected: %v",
				[]string(cycle),
			))
		}
	}

	return violations
}

// analyzeCmd creates the cobra command for the analyze subcommand.
// It wires flag parsing and signal handling, then delegates to RunAnalyze
// per AP-002 (no business logic in the command layer).
func analyzeCmd() *cobra.Command {
	var (
		maxInstability float64
		maxDistance    float64
		maxLCOM        int
		noCircularDeps bool
		timeout        time.Duration
		output         string
	)

	cmd := &cobra.Command{
		Use:   "analyze [path]",
		Short: "Analyze Go packages and compute coupling metrics",
		Long: `Analyze computes package-level coupling metrics for a Go project:
afferent coupling (Ca), efferent coupling (Ce), instability, abstractness,
distance from main sequence, LCOM4 cohesion, and circular dependency detection.

Output is JSON conforming to the ModuleGraph schema (version 1.1).

Use threshold flags (--max-instability, --max-distance, --max-lcom,
--no-circular-deps) for CI gate enforcement. Violations cause exit code 1.
JSON output is always written to stdout, even when violations are detected.`,
		Args: cobra.MaximumNArgs(1),
		// SilenceUsage prevents cobra from printing usage on RunE errors.
		// We handle error reporting ourselves.
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Determine project path: argument or current directory.
			path := "."
			if len(args) > 0 {
				path = args[0]
			}

			// Signal handling: intercept SIGINT and SIGTERM.
			ctx, stop := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			// Build options, converting explicitly-set flags to pointers.
			opts := AnalyzeOptions{
				Stdout:         cmd.OutOrStdout(),
				Stderr:         cmd.ErrOrStderr(),
				Path:           path,
				OutputPath:     output,
				NoCircularDeps: noCircularDeps,
				Timeout:        timeout,
			}

			// Use cobra's Changed() to distinguish "not set" from "set to zero".
			if cmd.Flags().Changed("max-instability") {
				opts.MaxInstability = &maxInstability
			}
			if cmd.Flags().Changed("max-distance") {
				opts.MaxDistance = &maxDistance
			}
			if cmd.Flags().Changed("max-lcom") {
				opts.MaxLCOM = &maxLCOM
			}

			result, err := RunAnalyze(ctx, opts)
			if err != nil {
				// Print error to stderr; cobra will not print usage
				// because SilenceUsage is true.
				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Error:", err)
				return &exitCodeError{code: result.ExitCode, err: err}
			}

			if result.ExitCode != 0 {
				// Policy violations — error message already written to stderr
				// by RunAnalyze. Return an exitCodeError so main() can set
				// the correct exit code without bypassing deferred cleanup.
				return &exitCodeError{
					code: result.ExitCode,
					err:  fmt.Errorf("threshold violations detected"),
				}
			}

			return nil
		},
	}

	// Register flags with change-tracking callbacks.
	cmd.Flags().Float64Var(&maxInstability, "max-instability", 0, "Maximum allowed instability [0.0, 1.0]")
	cmd.Flags().Float64Var(&maxDistance, "max-distance", 0, "Maximum allowed distance from main sequence [0.0, 1.0]")
	cmd.Flags().IntVar(&maxLCOM, "max-lcom", 0, "Maximum allowed LCOM value (>= 1)")
	cmd.Flags().BoolVar(&noCircularDeps, "no-circular-deps", false, "Treat circular dependencies as violations")
	cmd.Flags().DurationVar(&timeout, "timeout", 0, "Analysis timeout (e.g., 30s, 2m)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Write ModuleGraph JSON to file instead of stdout")

	return cmd
}
