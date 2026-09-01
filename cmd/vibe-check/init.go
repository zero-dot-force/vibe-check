package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/zero-dot-force/vibe-check/internal/scaffold"
)

// InitOptions contains the configuration for the init command.
// It follows the AP-001 Options struct pattern for testable CLI commands.
type InitOptions struct {
	// Stdout is the writer for the deployment summary (table or JSON). Required.
	Stdout io.Writer
	// Stderr is the writer for diagnostics and errors. Required.
	Stderr io.Writer
	// Path is the target project directory into which agent assets are deployed.
	// Empty defaults to the current directory (".").
	Path string
	// Force overwrites existing agent asset files instead of skipping them.
	Force bool
	// JSON selects machine-readable JSON output when true; otherwise a
	// human-readable summary is written.
	JSON bool

	// writeFile is an optional filesystem seam forwarded to scaffold.Run so the
	// I/O-failure exit path can be exercised deterministically in tests. When
	// nil, scaffold.Run defaults to os.WriteFile. It is unexported because it is
	// a test seam, not part of the public command contract.
	writeFile func(path string, data []byte, perm fs.FileMode) error
}

// InitResult contains the outcome of an init deployment.
// It follows the AP-001 Result struct pattern.
type InitResult struct {
	// Written lists the asset filenames newly created.
	Written []string
	// Skipped lists the asset filenames left untouched because they already
	// existed and Force was not set.
	Skipped []string
	// Forced lists the asset filenames overwritten because Force was set.
	Forced []string
	// ExitCode is the process exit code: 0 on success (including an all-skipped
	// run), 2 on an invalid target path or an I/O failure.
	ExitCode int
}

// initJSON is the machine-readable init payload emitted under --json. Field
// declaration order determines JSON key order and every key uses lowercase. All
// three slices are normalized to empty (never null) so each key is always a JSON
// array.
type initJSON struct {
	Written []string `json:"written"`
	Skipped []string `json:"skipped"`
	Forced  []string `json:"forced"`
}

// RunInit deploys the embedded Review Council agent assets into the target
// project's .opencode/agents/ directory and writes a summary to opts.Stdout. It
// is the testable entry point per AP-002/AP-003: all business logic lives here,
// not in the cobra command layer.
//
// Exit code semantics (also mirrored in the returned InitResult.ExitCode):
//   - 0: assets were deployed (or all skipped) and the summary was written.
//   - 2: the target path is invalid (missing, not a directory, or traversal) or
//     an I/O failure occurred while writing an asset. In that case nothing is
//     written to opts.Stdout and the returned error describes the failure.
//
// RunInit never writes a partial summary to opts.Stdout on any error path.
func RunInit(ctx context.Context, opts InitOptions) (*InitResult, error) {
	// Default the target path to the current directory.
	path := opts.Path
	if path == "" {
		path = "."
	}

	// Honor cancellation before doing any filesystem work.
	if err := ctx.Err(); err != nil {
		return &InitResult{ExitCode: 2}, fmt.Errorf("init: %w", err)
	}

	// Deploy the embedded assets. scaffold.Run validates the target path and
	// performs the writes; any error (invalid path or I/O failure) is a tool
	// failure (exit 2) with no stdout output.
	res, err := scaffold.Run(scaffold.Options{
		TargetDir: path,
		Force:     opts.Force,
		WriteFile: opts.writeFile,
	})
	if err != nil {
		return &InitResult{ExitCode: 2}, fmt.Errorf("init: %w", err)
	}

	// Honor cancellation after the writes but before emitting output.
	if err := ctx.Err(); err != nil {
		return &InitResult{ExitCode: 2}, fmt.Errorf("init: %w", err)
	}

	// Render the summary into a buffer so a single write reaches Stdout, keeping
	// output atomic (no partial payload on a write error) and deterministic.
	var buf bytes.Buffer
	if opts.JSON {
		if err := writeInitJSON(&buf, res); err != nil {
			return &InitResult{ExitCode: 2}, fmt.Errorf("encode init json: %w", err)
		}
	} else {
		writeInitSummary(&buf, path, res)
	}

	if _, err := opts.Stdout.Write(buf.Bytes()); err != nil {
		return &InitResult{ExitCode: 2}, fmt.Errorf("write init output: %w", err)
	}

	return &InitResult{
		Written:  res.Written,
		Skipped:  res.Skipped,
		Forced:   res.Forced,
		ExitCode: 0,
	}, nil
}

// writeInitJSON renders the deployment result as a single indented JSON object
// to w. Each slice is normalized so its key is always a JSON array, never null.
func writeInitJSON(w io.Writer, res *scaffold.Result) error {
	payload := initJSON{
		Written: normStringSlice(res.Written),
		Skipped: normStringSlice(res.Skipped),
		Forced:  normStringSlice(res.Forced),
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal init payload: %w", err)
	}
	if _, err := fmt.Fprintln(w, string(data)); err != nil {
		return fmt.Errorf("write init payload: %w", err)
	}
	return nil
}

// writeInitSummary renders the deployment result as a human-readable summary to
// w. It names the target agents directory, then lists the written, skipped, and
// forced assets. The lists are consumed in the stable, sorted order scaffold.Run
// guarantees, so output is byte-stable across runs.
func writeInitSummary(w io.Writer, targetDir string, res *scaffold.Result) {
	agentsDir := filepath.Join(targetDir, ".opencode", "agents")
	_, _ = fmt.Fprintf(w, "Deployed agent assets under %s:\n", agentsDir)
	writeListSection(w, "Written:", res.Written)
	writeListSection(w, "Skipped:", res.Skipped)
	writeListSection(w, "Forced:", res.Forced)
}

// normStringSlice returns s unchanged, or an empty (non-nil) slice when s is
// nil, so JSON marshaling always yields an array rather than null.
func normStringSlice(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

// initCmd creates the cobra command for the init subcommand. It wires flag
// parsing and signal handling, then delegates to RunInit per AP-002 (no
// business logic in the command layer).
func initCmd() *cobra.Command {
	var (
		force   bool
		jsonOut bool
	)

	cmd := &cobra.Command{
		Use:   "init [path]",
		Short: "Deploy vibe-check Review Council agent assets into a project",
		Long: `Init deploys the embedded vibe-check Review Council agent assets (such as the
divisor-entropy structural-entropy reviewer) into a target project's
.opencode/agents/ directory.

Existing files are skipped by default; use --force to overwrite them. The target
path defaults to the current directory. Output is a human-readable summary by
default or a JSON object with --json.

Exit code is 0 on success (including an all-skipped run) and 2 when the target
path is invalid or an asset cannot be written.`,
		Args: cobra.MaximumNArgs(1),
		// SilenceUsage prevents cobra from printing usage on RunE errors; we
		// report errors ourselves.
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Determine target path: argument or current directory.
			path := "."
			if len(args) > 0 {
				path = args[0]
			}

			// Signal handling: intercept SIGINT and SIGTERM.
			ctx, stop := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			opts := InitOptions{
				Stdout: cmd.OutOrStdout(),
				Stderr: cmd.ErrOrStderr(),
				Path:   path,
				Force:  force,
				JSON:   jsonOut,
			}

			result, err := RunInit(ctx, opts)
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
				// Defensive: RunInit couples every non-zero exit code with a
				// non-nil error handled above, so this path is not expected.
				return &exitCodeError{
					code: result.ExitCode,
					err:  fmt.Errorf("init failed with exit code %d", result.ExitCode),
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing agent asset files instead of skipping them")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit a machine-readable JSON payload instead of a summary")

	return cmd
}
