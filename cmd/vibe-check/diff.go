package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/zero-dot-force/vibe-check/metrics"
)

// DiffOptions contains the configuration for the diff command.
// It follows the AP-001 Options struct pattern for testable CLI commands.
type DiffOptions struct {
	// Stdout is the writer for the delta report (table or JSON). Required.
	Stdout io.Writer
	// Stderr is the writer for diagnostics and errors. Required.
	Stderr io.Writer
	// BasePath is the filesystem path to the base ModuleGraph JSON document.
	BasePath string
	// PRPath is the filesystem path to the PR ModuleGraph JSON document.
	PRPath string
	// Thresholds holds the verdict gates applied by DecideVerdict. Callers build
	// this from metrics.DefaultVerdictThresholds, optionally tightened by the
	// command layer's tighten-only override flags.
	Thresholds metrics.VerdictThresholds
	// JSON selects machine-readable JSON output when true; otherwise a
	// human-readable table is written.
	JSON bool
}

// DiffResult contains the outcome of a diff comparison.
// It follows the AP-001 Result struct pattern.
type DiffResult struct {
	// Delta is the computed structural delta between the base and PR graphs.
	Delta metrics.GraphDelta
	// Verdict is the entropy verdict rendered from the delta and thresholds.
	Verdict metrics.Verdict
	// Reasons lists the machine-readable reasons each gate fired. It is empty for
	// an APPROVE verdict.
	Reasons []string
	// ExitCode is the process exit code: 0 when both inputs are valid and a
	// verdict was computed, 2 on a tool failure (unreadable or schema-invalid
	// input). A REQUEST_CHANGES verdict still exits 0 — the verdict travels in
	// the payload, not the exit code.
	ExitCode int
}

// diffJSON is the machine-readable diff payload emitted under --json. Field
// declaration order determines JSON key order and every key uses camelCase. The
// added/removed fields use omitempty so they are absent when there is nothing to
// report; because ComputeDelta suppresses them (leaves them empty) whenever the
// measurement is unreliable, an unreliable payload never carries added/removed
// signal.
type diffJSON struct {
	Verdict          metrics.Verdict          `json:"verdict"`
	Reasons          []string                 `json:"reasons"`
	EntropyDirection metrics.EntropyDirection `json:"entropyDirection"`
	Unreliable       bool                     `json:"unreliable"`
	Modules          []metrics.Delta          `json:"modules"`
	Added            []string                 `json:"added,omitempty"`
	Removed          []string                 `json:"removed,omitempty"`
	NewCycles        []metrics.Cycle          `json:"newCycles"`
	ResolvedCycles   []metrics.Cycle          `json:"resolvedCycles"`
}

// RunDiff reads two ModuleGraph JSON documents, computes their structural delta,
// renders an entropy verdict, and writes a report to opts.Stdout. It returns a
// non-nil *DiffResult and nil error on success. On any failure it returns a
// non-nil *DiffResult with ExitCode 2 and a wrapped error; nothing is written to
// opts.Stdout on error paths. This is the testable entry point per AP-002/AP-003:
// all business logic lives here, not in the cobra command layer.
//
// Exit code semantics (also mirrored in the returned DiffResult.ExitCode):
//   - 0: both inputs were read and validated, the delta and verdict were
//     computed, and the report was written to opts.Stdout. A REQUEST_CHANGES
//     verdict still returns 0 — diff is a reporting tool and conveys the
//     verdict in its payload, not in the process exit code.
//   - 2: either input is missing/unreadable or fails ModuleGraph schema
//     validation. In that case nothing is written to opts.Stdout and the
//     returned error describes the failure for the command layer to report.
//
// RunDiff never writes a partial payload to opts.Stdout on any error path.
func RunDiff(ctx context.Context, opts DiffOptions) (*DiffResult, error) {
	// Step 1: Read both inputs. A missing or unreadable file is a tool failure
	// (exit 2) with no stdout output.
	baseData, err := os.ReadFile(opts.BasePath)
	if err != nil {
		return &DiffResult{ExitCode: 2}, fmt.Errorf("read base file %s: %w", opts.BasePath, err)
	}
	prData, err := os.ReadFile(opts.PRPath)
	if err != nil {
		return &DiffResult{ExitCode: 2}, fmt.Errorf("read pr file %s: %w", opts.PRPath, err)
	}

	// Step 2: Validate both documents against the ModuleGraph schema before
	// computing anything. A schema-invalid input is a tool failure (exit 2).
	if err := metrics.Validate(baseData); err != nil {
		return &DiffResult{ExitCode: 2}, fmt.Errorf("validate base file %s: %w", opts.BasePath, err)
	}
	if err := metrics.Validate(prData); err != nil {
		return &DiffResult{ExitCode: 2}, fmt.Errorf("validate pr file %s: %w", opts.PRPath, err)
	}

	// Step 3: Unmarshal into typed graphs. Validation already parsed the JSON, so
	// a failure here indicates a value that cannot map onto the typed model.
	var base, pr metrics.ModuleGraph
	if err := json.Unmarshal(baseData, &base); err != nil {
		return &DiffResult{ExitCode: 2}, fmt.Errorf("parse base file %s: %w", opts.BasePath, err)
	}
	if err := json.Unmarshal(prData, &pr); err != nil {
		return &DiffResult{ExitCode: 2}, fmt.Errorf("parse pr file %s: %w", opts.PRPath, err)
	}

	// Step 4: Compute the delta and verdict. Both are pure and deterministic.
	delta := metrics.ComputeDelta(&base, &pr)
	verdict, reasons := metrics.DecideVerdict(delta, opts.Thresholds)

	// Step 5: Render the report into a buffer so a single write reaches Stdout.
	// Buffering keeps the output atomic (no partial payload on a write error) and
	// deterministic.
	var buf bytes.Buffer
	if opts.JSON {
		if err := writeDiffJSON(&buf, delta, verdict, reasons); err != nil {
			return &DiffResult{ExitCode: 2}, fmt.Errorf("encode diff json: %w", err)
		}
	} else {
		writeDiffTable(&buf, delta, verdict, reasons)
	}

	// Final context check before writing — prevent partial output on cancellation.
	if err := ctx.Err(); err != nil {
		return &DiffResult{ExitCode: 2}, fmt.Errorf("diff: %w", err)
	}

	if _, err := opts.Stdout.Write(buf.Bytes()); err != nil {
		return &DiffResult{ExitCode: 2}, fmt.Errorf("write diff output: %w", err)
	}

	return &DiffResult{
		Delta:    delta,
		Verdict:  verdict,
		Reasons:  reasons,
		ExitCode: 0,
	}, nil
}

// writeDiffJSON renders the diff as a single indented JSON object to w and
// returns nil on success. A nil reasons slice is normalized to an empty slice so
// the reasons key is always a JSON array, never null. Returns a wrapped error if
// marshaling or writing to w fails.
func writeDiffJSON(w io.Writer, delta metrics.GraphDelta, verdict metrics.Verdict, reasons []string) error {
	if reasons == nil {
		reasons = []string{}
	}
	payload := diffJSON{
		Verdict:          verdict,
		Reasons:          reasons,
		EntropyDirection: delta.Direction,
		Unreliable:       delta.Unreliable,
		Modules:          delta.Modules,
		Added:            delta.Added,
		Removed:          delta.Removed,
		NewCycles:        delta.NewCycles,
		ResolvedCycles:   delta.ResolvedCycles,
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal diff payload: %w", err)
	}
	if _, err := fmt.Fprintln(w, string(data)); err != nil {
		return fmt.Errorf("write diff payload: %w", err)
	}
	return nil
}

// writeDiffTable renders the diff as a human-readable report to w. When the
// measurement is unreliable it leads with a partial-build annotation and omits
// the added/removed sections whose signal ComputeDelta has suppressed. It then
// prints the per-module delta table, the new and resolved cycle lists, the
// entropy direction, the verdict, and the verdict reasons. All collections are
// consumed in the stable, sorted order that ComputeDelta and DecideVerdict
// guarantee, so output is byte-stable across runs. Writes target an in-memory
// buffer that cannot fail, so intermediate write errors are intentionally
// discarded.
func writeDiffTable(w io.Writer, delta metrics.GraphDelta, verdict metrics.Verdict, reasons []string) {
	if delta.Unreliable {
		_, _ = fmt.Fprintln(w, "partial build — measurement unreliable")
		_, _ = fmt.Fprintln(w)
	}

	_, _ = fmt.Fprintln(w, "Per-module deltas (PR minus base):")
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "PATH\tCa\tCe\tInstability\tAbstractness\tDistance\tLCOM")
	if len(delta.Modules) == 0 {
		_, _ = fmt.Fprintln(tw, "(no shared modules)")
	}
	for _, m := range delta.Modules {
		_, _ = fmt.Fprintf(tw, "%s\t%d\t%d\t%.4f\t%.4f\t%.4f\t%d\n",
			m.Path, m.Ca, m.Ce,
			normFloat(m.Instability), normFloat(m.Abstractness), normFloat(m.Distance),
			m.LCOM)
	}
	_ = tw.Flush()

	writeCycleSection(w, "New cycles:", delta.NewCycles)
	writeCycleSection(w, "Resolved cycles:", delta.ResolvedCycles)

	if !delta.Unreliable {
		writeListSection(w, "Added packages:", delta.Added)
		writeListSection(w, "Removed packages:", delta.Removed)
	}

	_, _ = fmt.Fprintf(w, "\nEntropy direction: %s\n", delta.Direction)
	_, _ = fmt.Fprintf(w, "Verdict: %s\n", verdict)
	if len(reasons) == 0 {
		_, _ = fmt.Fprintln(w, "Reasons: (none)")
		return
	}
	_, _ = fmt.Fprintln(w, "Reasons:")
	for _, r := range reasons {
		_, _ = fmt.Fprintf(w, "  - %s\n", r)
	}
}

// writeCycleSection writes a titled list of cycles to w, one member-joined cycle
// per line, or "(none)" when the list is empty.
func writeCycleSection(w io.Writer, title string, cycles []metrics.Cycle) {
	_, _ = fmt.Fprintln(w, title)
	if len(cycles) == 0 {
		_, _ = fmt.Fprintln(w, "  (none)")
		return
	}
	for _, c := range cycles {
		_, _ = fmt.Fprintf(w, "  - %s\n", strings.Join([]string(c), " "))
	}
}

// writeListSection writes a titled list of strings to w, one item per line, or
// "(none)" when the list is empty.
func writeListSection(w io.Writer, title string, items []string) {
	_, _ = fmt.Fprintln(w, title)
	if len(items) == 0 {
		_, _ = fmt.Fprintln(w, "  (none)")
		return
	}
	for _, it := range items {
		_, _ = fmt.Fprintf(w, "  - %s\n", it)
	}
}

// normFloat maps negative zero to positive zero so a delta of exactly zero
// always renders as "0.0000" rather than "-0.0000", keeping the table output
// stable. Returns the input unchanged for all other values.
func normFloat(f float64) float64 {
	if f == 0 {
		return 0
	}
	return f
}

// tightenThresholds applies tighten-only overrides to the protected default
// gates. Each override argument is nil when its flag was not set. An override is
// accepted only when it does not exceed the default: a smaller value tightens
// the gate and an equal value is a no-op, so neither weakens it. A looser
// (strictly greater) override returns an error naming the offending flag,
// enforcing the AGENTS.md gatekeeping mandate that protected quality gates may
// be tightened but never weakened.
func tightenThresholds(def metrics.VerdictThresholds, instability, distance *float64, lcom *int) (metrics.VerdictThresholds, error) {
	out := def
	if instability != nil {
		if *instability > def.MaxInstabilityDelta {
			return def, fmt.Errorf("--max-instability-delta %.4f is looser than the protected default %.4f: overrides may only tighten (lower) the gate", *instability, def.MaxInstabilityDelta)
		}
		out.MaxInstabilityDelta = *instability
	}
	if distance != nil {
		if *distance > def.MaxDistanceDelta {
			return def, fmt.Errorf("--max-distance-delta %.4f is looser than the protected default %.4f: overrides may only tighten (lower) the gate", *distance, def.MaxDistanceDelta)
		}
		out.MaxDistanceDelta = *distance
	}
	if lcom != nil {
		if *lcom > def.MaxLCOMDelta {
			return def, fmt.Errorf("--max-lcom-delta %d is looser than the protected default %d: overrides may only tighten (lower) the gate", *lcom, def.MaxLCOMDelta)
		}
		out.MaxLCOMDelta = *lcom
	}
	return out, nil
}

// diffCmd creates the cobra command for the diff subcommand. It wires flag
// parsing, tighten-only threshold validation, and signal handling, then
// delegates to RunDiff per AP-002 (no business logic in the command layer).
func diffCmd() *cobra.Command {
	var (
		jsonOut             bool
		maxInstabilityDelta float64
		maxDistanceDelta    float64
		maxLCOMDelta        int
	)

	defaults := metrics.DefaultVerdictThresholds()

	cmd := &cobra.Command{
		Use:   "diff <base.json> <pr.json>",
		Short: "Compare two ModuleGraph JSON files and report an entropy verdict",
		Long: `Diff compares a base ModuleGraph against a PR ModuleGraph, computes the
per-package structural-quality delta (Ca, Ce, instability, abstractness,
distance from main sequence, LCOM), classifies new and resolved circular
dependencies, and renders a deterministic entropy verdict (APPROVE, COMMENT, or
REQUEST_CHANGES).

Both inputs must be JSON documents conforming to the ModuleGraph schema, as
produced by 'vibe-check analyze'. Output is a human-readable table by default or
a JSON object with --json.

The verdict is reported in the output payload, not the exit code: diff exits 0
whenever both inputs are valid (even for a REQUEST_CHANGES verdict) and exits 2
only when an input is missing, unreadable, or schema-invalid.

The --max-*-delta override flags are TIGHTEN-ONLY: a value looser than the
protected default (instability 0.15, distance 0.20, LCOM 2) is rejected with
exit code 2.`,
		Args: cobra.ExactArgs(2),
		// SilenceUsage prevents cobra from printing usage on RunE errors; we
		// report errors ourselves.
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve tighten-only threshold overrides BEFORE reading any file or
			// computing a verdict, converting only explicitly-set flags to
			// pointers. A looser-than-default override is rejected with exit 2 and
			// no stdout payload.
			var (
				instOverride *float64
				distOverride *float64
				lcomOverride *int
			)
			if cmd.Flags().Changed("max-instability-delta") {
				instOverride = &maxInstabilityDelta
			}
			if cmd.Flags().Changed("max-distance-delta") {
				distOverride = &maxDistanceDelta
			}
			if cmd.Flags().Changed("max-lcom-delta") {
				lcomOverride = &maxLCOMDelta
			}

			thresholds, err := tightenThresholds(defaults, instOverride, distOverride, lcomOverride)
			if err != nil {
				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Error:", err)
				return &exitCodeError{code: 2, err: err}
			}

			// Signal handling: intercept SIGINT and SIGTERM.
			ctx, stop := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			opts := DiffOptions{
				Stdout:     cmd.OutOrStdout(),
				Stderr:     cmd.ErrOrStderr(),
				BasePath:   args[0],
				PRPath:     args[1],
				Thresholds: thresholds,
				JSON:       jsonOut,
			}

			result, err := RunDiff(ctx, opts)
			if err != nil {
				// Print error to stderr; cobra will not print usage because
				// SilenceUsage is true.
				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Error:", err)
				code := 2
				if result != nil && result.ExitCode != 0 {
					code = result.ExitCode
				}
				return &exitCodeError{code: code, err: err}
			}

			if result.ExitCode != 0 {
				// Defensive: RunDiff couples every non-zero exit code with a
				// non-nil error handled above, so this path is not expected.
				return &exitCodeError{
					code: result.ExitCode,
					err:  fmt.Errorf("diff failed with exit code %d", result.ExitCode),
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit a machine-readable JSON payload instead of a table")
	cmd.Flags().Float64Var(&maxInstabilityDelta, "max-instability-delta", defaults.MaxInstabilityDelta, "Tighten-only instability-increase gate (must be <= 0.15)")
	cmd.Flags().Float64Var(&maxDistanceDelta, "max-distance-delta", defaults.MaxDistanceDelta, "Tighten-only distance-increase gate (must be <= 0.20)")
	cmd.Flags().IntVar(&maxLCOMDelta, "max-lcom-delta", defaults.MaxLCOMDelta, "Tighten-only LCOM-increase gate (must be <= 2)")

	return cmd
}
