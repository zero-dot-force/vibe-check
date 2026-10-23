package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/zero-dot-force/vibe-check/metrics"
	"github.com/zero-dot-force/vibe-check/metrics/report"
)

// Report detail levels selectable via the report command's --mode flag. The
// strings are the user-facing flag values and are contract with the
// report-command spec's Report modes requirement: summary is the default and
// any other value is rejected with exit code 2 before the snapshot is read.
const (
	// ModeSummary renders the compact report: the header, metric legend,
	// status summary line, and prioritized recommendations.
	ModeSummary = "summary"
	// ModeDetailed renders the full report: everything in summary mode plus
	// the package metrics table, coupling and cohesion summaries, circular
	// dependency chains, duplication table, health scorecard, and warnings.
	ModeDetailed = "detailed"
)

// ReportOptions contains the configuration for the report command.
// It follows the AP-001 Options struct pattern for testable CLI commands.
type ReportOptions struct {
	// Stdout is the writer for the rendered report: the markdown report
	// or the JSON payload. Required.
	Stdout io.Writer
	// Stderr is the writer for diagnostics and errors. Required.
	Stderr io.Writer
	// GraphPath is the filesystem path to the ModuleGraph JSON document.
	GraphPath string
	// Mode selects the report detail level: ModeSummary or ModeDetailed.
	// Any other value is rejected with exit code 2 before the snapshot is
	// read. Ignored when JSON is set, because the JSON payload has no
	// detail levels.
	Mode string
	// ModeExplicit reports whether a report mode was explicitly requested
	// (the --mode flag on the command surface). It exists to enforce the
	// mutual exclusion between --json and --mode: the mode flag defaults to
	// summary, so the Mode value alone cannot distinguish an explicit
	// --mode summary from the default. The command layer sets it from the
	// flag's changed state; direct callers set it when they mean a mode to
	// have been chosen explicitly.
	ModeExplicit bool
	// JSON selects machine-readable output: the computed report model —
	// aggregates, per-dimension grades, overall status, and fired
	// recommendations with severity and rule — serialized as a single JSON
	// object instead of a markdown report. The exit-code contract is
	// unchanged under JSON. Mutually exclusive with an explicitly
	// requested mode: JSON together with ModeExplicit is rejected with
	// exit code 2 before the snapshot is read.
	JSON bool
}

// ReportResult contains the outcome of a report render.
// It follows the AP-001 Result struct pattern.
type ReportResult struct {
	// Graph is the parsed ModuleGraph snapshot the report was rendered
	// from. It is nil on error paths.
	Graph *metrics.ModuleGraph
	// Recommendations lists the fired recommendations, already ordered and
	// capped by report.ComputeRecommendations.
	Recommendations []report.Recommendation
	// Status is the overall status word: report.StatusAllClear when no
	// violation-severity recommendation fires, otherwise
	// report.StatusNeedsAttention.
	Status string
	// ExitCode is the process exit code: 0 when no red or yellow
	// recommendation fires, 1 when any fires (the full report — markdown
	// or JSON — is still written to opts.Stdout), and 2 on a tool failure
	// (conflicting flags, invalid mode, or unreadable or schema-invalid
	// input) — in which case nothing is written to opts.Stdout.
	ExitCode int
}

// RunReport reads a ModuleGraph JSON snapshot, validates it against the
// embedded ModuleGraph schema, renders the output selected by opts — the
// markdown report for opts.Mode, or the computed report model as a single
// JSON object for opts.JSON — and writes it to opts.Stdout in a single
// write. It returns a non-nil *ReportResult and nil error on success. On
// any failure it returns a non-nil *ReportResult with ExitCode 2 and a
// wrapped error; nothing is written to opts.Stdout on error paths. This is
// the testable entry point per AP-002/AP-003: all business logic lives
// here, not in the cobra command layer.
//
// Rendering is delegated to the pure report package renderers and model builder,
// so output is byte-stable for identical input files. The markdown
// renderer's provenance fallback carries this binary's producer name and
// semantic version, so a snapshot that omits provenance renders the same
// identity `vibe-check analyze` stamps into the snapshots it emits; the
// JSON model carries no producer identity at all, so its payload is
// byte-identical with or without snapshot provenance.
//
// Exit code semantics (also mirrored in the returned ReportResult.ExitCode,
// and unchanged under --json):
//   - 0: the snapshot was read, validated, and rendered, and no red or
//     yellow recommendation fired.
//   - 1: the snapshot was rendered and at least one red or yellow
//     recommendation fired (a policy violation). The full report — markdown
//     or JSON — is still written to opts.Stdout; only the exit code differs
//     from 0.
//   - 2: tool failure — the flags are invalid (an unknown mode, or --json
//     together with an explicit --mode), or the input is missing,
//     unreadable, or schema-invalid. Nothing is written to opts.Stdout and
//     the returned error describes the failure for the command layer to
//     report.
func RunReport(ctx context.Context, opts ReportOptions) (*ReportResult, error) {
	// Step 1: Validate flags. Flag validation runs BEFORE the snapshot is
	// read so an invalid flag combination is rejected with no filesystem
	// access, mirroring how RunAnalyze validates threshold flags before
	// analysis. --json and an explicit --mode are mutually exclusive per
	// the report spec's JSON output requirement: --json emits the computed
	// report model, --mode selects a markdown detail level. The
	// combination check precedes mode validation so a flag-conflict is
	// always reported as such, and mode validation applies only to the
	// markdown path because the JSON payload has no detail levels.
	if opts.JSON && opts.ModeExplicit {
		return &ReportResult{ExitCode: 2}, fmt.Errorf(
			"--json and --mode are mutually exclusive: pass --json for the JSON report model or --mode for a markdown detail level, not both")
	}
	if !opts.JSON {
		if err := validateReportMode(opts.Mode); err != nil {
			return &ReportResult{ExitCode: 2}, err
		}
	}

	// Step 2: Read the snapshot. A missing or unreadable file is a tool
	// failure (exit 2) with no stdout output.
	data, err := os.ReadFile(opts.GraphPath)
	if err != nil {
		return &ReportResult{ExitCode: 2}, fmt.Errorf("read graph file %s: %w", opts.GraphPath, err)
	}

	// Step 3: Validate the document against the ModuleGraph schema before
	// rendering anything. A schema-invalid input is a tool failure (exit 2).
	if err := metrics.Validate(data); err != nil {
		return &ReportResult{ExitCode: 2}, fmt.Errorf("validate graph file %s: %w", opts.GraphPath, err)
	}

	// Step 4: Unmarshal into the typed graph. Validation already parsed the
	// JSON, so a failure here indicates a value that cannot map onto the
	// typed model.
	var graph metrics.ModuleGraph
	if err := json.Unmarshal(data, &graph); err != nil {
		return &ReportResult{ExitCode: 2}, fmt.Errorf("parse graph file %s: %w", opts.GraphPath, err)
	}

	// Step 5: Compute the recommendations and render the selected output.
	// The recommendations are computed once so the exit code, the result
	// model, and the rendered report always agree. Under --json the output
	// is the computed report model itself — aggregates, per-dimension
	// grades, overall status, fired recommendations — assembled by the
	// pure engine builder and marshaled directly from its struct, so the
	// field order is the struct's declaration order and the payload is
	// byte-stable. Under --mode it is the markdown render.
	recs := report.ComputeRecommendations(&graph)
	renderOpts := report.RenderOptions{
		FallbackProducer: producerName,
		FallbackVersion:  semanticVersion(),
	}
	var out string
	switch {
	case opts.JSON:
		data, err := json.MarshalIndent(report.BuildReportModel(&graph, recs), "", "  ")
		if err != nil {
			return &ReportResult{ExitCode: 2}, fmt.Errorf("encode report model: %w", err)
		}
		// The trailing newline matches the markdown report's single final
		// newline, keeping both outputs POSIX line-friendly.
		out = string(data) + "\n"
	case opts.Mode == ModeDetailed:
		out = report.RenderDetailed(&graph, renderOpts)
	default: // ModeSummary — the only other validated value
		out = report.RenderSummary(&graph, renderOpts)
	}

	// Final context check before writing — prevent partial output on
	// cancellation.
	if err := ctx.Err(); err != nil {
		return &ReportResult{ExitCode: 2}, fmt.Errorf("report: %w", err)
	}

	// Step 6: Write the report in a single write so a write error never
	// emits a partial report.
	if _, err := io.WriteString(opts.Stdout, out); err != nil {
		return &ReportResult{ExitCode: 2}, fmt.Errorf("write report output: %w", err)
	}

	exitCode := 0
	if report.HasViolations(recs) {
		exitCode = 1
	}

	return &ReportResult{
		Graph:           &graph,
		Recommendations: recs,
		Status:          report.ReportStatus(recs),
		ExitCode:        exitCode,
	}, nil
}

// validateReportMode checks that mode is one of the defined report detail
// levels. Per the report-command spec, trending is agent-owned and not a
// report mode, so it — and every other value — is rejected with an error
// naming the valid modes.
func validateReportMode(mode string) error {
	switch mode {
	case ModeSummary, ModeDetailed:
		return nil
	default:
		return fmt.Errorf("invalid --mode value %q: must be %q or %q", mode, ModeSummary, ModeDetailed)
	}
}

// reportCmd creates the cobra command for the report subcommand. It wires
// flag parsing and signal handling, then delegates to RunReport per AP-002
// (no business logic in the command layer).
func reportCmd() *cobra.Command {
	var (
		mode    string
		jsonOut bool
	)

	cmd := &cobra.Command{
		Use:   "report <graph.json>",
		Short: "Render a markdown architectural health report from a ModuleGraph snapshot",
		Long: `Report renders a deterministic architectural health report from a
ModuleGraph JSON snapshot previously emitted by 'vibe-check analyze'. It never
analyzes code: reading the snapshot, validating it against the embedded
ModuleGraph schema, and rendering are pure, so identical input files render
byte-identical reports regardless of run time.

--mode selects the detail level:
  summary  (default) report header, metric legend, status summary line, and
           prioritized recommendations
  detailed everything in summary mode plus the package metrics table, coupling
           and cohesion summaries, circular dependency chains, duplication
           table, health scorecard, and warnings

--json emits the computed report model — aggregates, per-dimension grades,
overall status, and fired recommendations with severity and rule — as a
single JSON object with a stable field order, for machine consumers that
should not parse markdown. It is mutually exclusive with --mode.

Exit codes: 0 when no red or yellow recommendation fires, 1 when any fires
(the full report — markdown or JSON — is still written to stdout), and 2
when the flags are invalid (an unknown --mode value, or --json and --mode
passed together) or the input is missing, unreadable, or schema-invalid
(error on stderr, nothing on stdout).`,
		Args: cobra.ExactArgs(1),
		// SilenceUsage prevents cobra from printing usage on RunE errors; we
		// report errors ourselves.
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Signal handling: intercept SIGINT and SIGTERM.
			ctx, stop := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			opts := ReportOptions{
				Stdout:    cmd.OutOrStdout(),
				Stderr:    cmd.ErrOrStderr(),
				GraphPath: args[0],
				Mode:      mode,
				// Only an explicitly passed --mode conflicts with --json;
				// the flag's default value must not, so the changed state —
				// not the mode value — carries the mutual-exclusion signal.
				ModeExplicit: cmd.Flags().Changed("mode"),
				JSON:         jsonOut,
			}

			result, err := RunReport(ctx, opts)
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
				// Violation exit — the full report is already written to
				// stdout. Return an exitCodeError with a descriptive
				// message: cobra prints it to stderr (SilenceErrors is not
				// set) and main() then sets the correct exit code (1)
				// without re-printing it.
				return &exitCodeError{
					code: result.ExitCode,
					err:  fmt.Errorf("violation-severity recommendations detected"),
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&mode, "mode", ModeSummary, "Report detail level: summary or detailed")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit the computed report model (aggregates, grades, status, recommendations) as a JSON object instead of markdown")

	return cmd
}
