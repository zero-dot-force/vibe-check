package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
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
	// Path is the target project directory into which agent and command assets
	// are deployed. Empty defaults to the current directory (".").
	Path string
	// Force overwrites existing asset files instead of skipping them.
	Force bool
	// JSON selects machine-readable JSON output when true; otherwise a
	// human-readable summary is written.
	JSON bool
	// NoProvenance when true omits the provenance envelope from the JSON
	// payload, yielding byte-reproducible output across runs. Provenance is
	// emitted by default.
	NoProvenance bool
	// CI additionally scaffolds the structural-gate caller workflow into the
	// target's .github/workflows/ directory, pinned to a vibe-check release
	// (tag plus commit SHA). The emitted file is user property: it is
	// skipped on rerun even under Force (design D1/D7).
	CI bool
	// VibeVersion carries the --vibe-version override: the release tag the
	// scaffolded caller workflow is pinned to. Empty means "this binary's
	// own release". Passing it without CI is a usage error (design D12).
	VibeVersion string

	// writeFile is an optional filesystem seam forwarded to scaffold.Run so the
	// I/O-failure exit path can be exercised deterministically in tests. When
	// nil, scaffold.Run defaults to os.WriteFile. It is unexported because it is
	// a test seam, not part of the public command contract.
	writeFile func(path string, data []byte, perm fs.FileMode) error

	// resolveTagCommit is an optional seam resolving a release tag to its
	// peeled commit SHA, so RunInit stays unit-testable offline. When nil,
	// the real git ls-remote resolver runs. It is unexported because it is
	// a test seam, not part of the public command contract.
	resolveTagCommit func(ctx context.Context, tag string) (string, error)
}

// InitResult contains the outcome of an init deployment.
// It follows the AP-001 Result struct pattern.
type InitResult struct {
	// Written lists the category-prefixed asset paths newly created (e.g.
	// "agents/divisor-entropy.md", "commands/vibe-check.md").
	Written []string
	// Skipped lists the category-prefixed asset paths left untouched because
	// they already existed and Force was not set.
	Skipped []string
	// Forced lists the category-prefixed asset paths overwritten because Force
	// was set.
	Forced []string
	// Notices lists non-blocking registration caveats from the UF policy
	// pass: absent policy files, unexpected file shapes, and advisory gaps
	// in entries that are already present.
	Notices []string
	// ExitCode is the process exit code: 0 on success (including an
	// all-skipped run), 2 on a usage error (--vibe-version without --ci),
	// an invalid target path, an unsafe-to-rewrite policy file (symlinked or
	// non-regular, including path components), an I/O failure, or — with
	// --ci — an unresolvable or invalid release tag or commit SHA.
	ExitCode int
}

// initJSON is the machine-readable init payload emitted under --json. Field
// declaration order determines JSON key order and every key uses lowercase. All
// four slices are normalized to empty (never null) so each key is always a JSON
// array.
type initJSON struct {
	Provenance *provenanceEnvelope `json:"provenance,omitempty"`
	Written    []string            `json:"written"`
	Skipped    []string            `json:"skipped"`
	Forced     []string            `json:"forced"`
	Notices    []string            `json:"notices"`
}

// RunInit deploys the embedded Review Council agent and command assets into the
// project's .opencode/agents/ and .opencode/commands/ directories, registers
// the divisor-entropy reviewer into the UF policy files (.uf/) when they are
// present, optionally scaffolds the structural-gate caller workflow into
// .github/workflows/ (Options.CI), and writes a deployment summary (table or
// JSON) to opts.Stdout. It returns a non-nil *InitResult and nil error on
// success. On any failure it returns a non-nil *InitResult with ExitCode 2 and
// a wrapped error; nothing is written to opts.Stdout on error paths. This is
// the testable entry point per AP-002/AP-003: all business logic lives here,
// not in the cobra command layer.
//
// Exit code semantics (also mirrored in the returned InitResult.ExitCode):
//   - 0: assets were deployed (or all skipped), the UF policy registration
//     pass completed (absent or shape-failed policy files only produce
//     notices), and the summary was written to opts.Stdout.
//   - 2: --vibe-version was passed without --ci (usage error), the target
//     path is invalid (missing, not a directory, or traversal), a policy file
//     cannot be safely rewritten (symlinked or non-regular), an I/O failure
//     occurred while writing an asset, or — with --ci — the release tag or
//     commit SHA could not be resolved or validated. In that case nothing is
//     written to opts.Stdout and the returned error describes the failure.
//
// RunInit never writes a partial summary to opts.Stdout on any error path.
func RunInit(ctx context.Context, opts InitOptions) (*InitResult, error) {
	// Default the target path to the current directory.
	path := opts.Path
	if path == "" {
		path = "."
	}

	// D12: reject the flag-combination misuse before any work — the version
	// override pins the caller workflow that --ci emits, so it is meaningless
	// without --ci.
	if opts.VibeVersion != "" && !opts.CI {
		return &InitResult{ExitCode: 2}, fmt.Errorf("init: --vibe-version requires --ci (the override pins the caller workflow that --ci emits)")
	}

	// Honor cancellation before doing any filesystem work.
	if err := ctx.Err(); err != nil {
		return &InitResult{ExitCode: 2}, fmt.Errorf("init: %w", err)
	}

	// Resolve both release pins when CI scaffolding is requested (design
	// D12). The tag is the --vibe-version override when set, else the
	// binary's own semanticVersion(); the commit SHA follows D3's channel
	// table — the ldflags-injected main.commit when it is a full SHA and no
	// override is present (the offline release-binary path), else tag→SHA
	// resolution at init time. All resolution failures fail closed with exit
	// 2 before any files are written.
	scaffoldOpts := scaffold.Options{
		TargetDir: path,
		Force:     opts.Force,
		WriteFile: opts.writeFile,
		CI:        opts.CI,
	}
	if opts.CI {
		tag := opts.VibeVersion
		if tag == "" {
			tag = semanticVersion()
		}
		// Fail fast on a tag that can never resolve (dev build,
		// unparseable override, or below-floor release) before spending a
		// network round trip on it.
		if err := scaffold.ValidateReleaseTag(tag); err != nil {
			return &InitResult{ExitCode: 2}, fmt.Errorf("init: %w", err)
		}
		commitSHA, err := resolveReleaseCommit(ctx, opts, tag)
		if err != nil {
			return &InitResult{ExitCode: 2}, err
		}
		scaffoldOpts.VibeVersion = tag
		scaffoldOpts.VibeCommit = commitSHA
	}

	// Deploy the embedded assets. scaffold.Run validates the target path and
	// the release pins, performs the writes, and reports the caller workflow
	// under the workflows/ prefix; any error (invalid path, invalid pin, or
	// I/O failure) is a tool failure (exit 2) with no stdout output.
	res, err := scaffold.Run(scaffoldOpts)
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
		if err := writeInitJSON(&buf, res, opts.NoProvenance); err != nil {
			return &InitResult{ExitCode: 2}, fmt.Errorf("encode init json: %w", err)
		}
	} else {
		writeInitSummary(&buf, path, res, opts.CI)
	}

	if _, err := opts.Stdout.Write(buf.Bytes()); err != nil {
		return &InitResult{ExitCode: 2}, fmt.Errorf("write init output: %w", err)
	}

	return &InitResult{
		Written:  res.Written,
		Skipped:  res.Skipped,
		Forced:   res.Forced,
		Notices:  res.Notices,
		ExitCode: 0,
	}, nil
}

// vibeCheckRepoURL is the canonical repository the tag→SHA resolver queries
// (design D3). The caller workflow's `uses:` reference pins a workflow file
// in this repository, so the release commit must be resolved here even when
// init runs from a fork's build.
const vibeCheckRepoURL = "https://github.com/zero-dot-force/vibe-check.git"

// resolveReleaseCommit resolves the release commit SHA for an already
// validated tag per design D3's channel table: the ldflags-injected main.commit
// when it is a full 40-hex SHA and no --vibe-version override is present (the
// release-binary path, fully offline — the binary knows the exact commit it
// was built from), else tag→SHA resolution at init time. Resolution failures
// are wrapped with the fail-closed remediation from design D6.
func resolveReleaseCommit(ctx context.Context, opts InitOptions, tag string) (string, error) {
	if opts.VibeVersion == "" && scaffold.IsCommitSHA(commit) {
		return commit, nil
	}
	resolver := opts.resolveTagCommit
	if resolver == nil {
		resolver = lsRemoteTagCommit
	}
	sha, err := resolver(ctx, tag)
	if err != nil {
		return "", fmt.Errorf("init: cannot resolve the release commit SHA for tag %q (%v) — install vibe-check from a release binary or rerun with network access", tag, err)
	}
	return sha, nil
}

// lsRemoteTagCommit resolves a release tag to its peeled commit SHA in the
// canonical vibe-check repository (see vibeCheckRepoURL).
func lsRemoteTagCommit(ctx context.Context, tag string) (string, error) {
	return lsRemoteTagCommitAt(ctx, vibeCheckRepoURL, tag)
}

// lsRemoteTagCommitAt resolves a release tag to its peeled commit SHA in the
// repository at repoURL via `git ls-remote <repoURL> refs/tags/<tag>^{}`. The
// ^{} suffix peels an annotated tag to the commit it names, so the `uses:`
// pin carries the exact commit the release was built from rather than the
// tag object. The command honors ctx for cancellation. Output shape:
// `<sha>\trefs/tags/<tag>^{}`; an empty match (tag absent remotely) is an
// error, not a zero value, so a partially resolved pin can never reach the
// scaffolded file. The repoURL parameter exists so tests exercise this exact
// code path against a local repository instead of the network.
func lsRemoteTagCommitAt(ctx context.Context, repoURL, tag string) (string, error) {
	ref := "refs/tags/" + tag + "^{}"
	out, err := exec.CommandContext(ctx, "git", "ls-remote", repoURL, ref).Output()
	if err != nil {
		return "", fmt.Errorf("git ls-remote %s %s: %w", repoURL, ref, err)
	}
	fields := strings.Fields(string(out))
	if len(fields) == 0 {
		return "", fmt.Errorf("git ls-remote %s returned no peeled reference for %s", repoURL, ref)
	}
	return fields[0], nil
}

// writeInitJSON renders the deployment result as a single indented JSON object
// to w and returns nil on success. Each slice is normalized so its key is always
// a JSON array, never null. Unless noProvenance is true, a provenance envelope
// is attached so consumers can attribute and audit the output. Returns a wrapped
// error if marshaling or writing to w fails.
func writeInitJSON(w io.Writer, res *scaffold.Result, noProvenance bool) error {
	payload := initJSON{
		Written: normStringSlice(res.Written),
		Skipped: normStringSlice(res.Skipped),
		Forced:  normStringSlice(res.Forced),
		Notices: normStringSlice(res.Notices),
	}
	if !noProvenance {
		payload.Provenance = newProvenanceEnvelope()
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
// w. It names the target .opencode directory, then lists the written, skipped,
// and forced assets, followed by a Notices section when the UF policy pass
// reported any caveats. When ci is true (init ran with --ci), a CI next-steps
// section closes the summary, mirroring the notice section's conditional
// rendering. The lists contain category-prefixed paths (e.g.
// "agents/divisor-entropy.md", "commands/vibe-check.md", and — with --ci —
// "workflows/vibe-check-structural-gate.yaml") in stable sorted order.
func writeInitSummary(w io.Writer, targetDir string, res *scaffold.Result, ci bool) {
	openCodeDir := filepath.Join(targetDir, ".opencode")
	_, _ = fmt.Fprintf(w, "Deployed agent and command assets under %s:\n", openCodeDir)
	writeListSection(w, "Written:", res.Written)
	writeListSection(w, "Skipped:", res.Skipped)
	writeListSection(w, "Forced:", res.Forced)
	if len(res.Notices) > 0 {
		writeListSection(w, "Notices:", res.Notices)
	}
	if ci {
		writeCINextSteps(w, targetDir)
	}
}

// writeCINextSteps renders the CI scaffolding next-steps section closing the
// human summary after a --ci run. The workflow file is user property once
// written (design D7), so the summary tells the user the follow-up actions
// init deliberately does not take on their behalf: committing the file,
// wiring branch protection, and any tightening or pin bumps they want.
func writeCINextSteps(w io.Writer, targetDir string) {
	workflowPath := filepath.Join(targetDir, ".github", "workflows", scaffold.CIWorkflowName)
	_, _ = fmt.Fprintf(w, "CI workflow next steps (%s):\n", workflowPath)
	_, _ = fmt.Fprintln(w, "  - commit the workflow file to your repository")
	_, _ = fmt.Fprintln(w, "  - the gate is advisory until branch protection marks its check required")
	_, _ = fmt.Fprintln(w, "  - uncomment the tighten-only max-*-delta inputs to tighten thresholds beyond the protected defaults")
	_, _ = fmt.Fprintln(w, "  - edit `branches:` when your default branch is not `main` — the gate never triggers otherwise")
	_, _ = fmt.Fprintln(w, "  - bump the pinned release by editing the file; reruns of `init --ci` skip it")
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
		force        bool
		jsonOut      bool
		noProvenance bool
		ci           bool
		vibeVersion  string
	)

	cmd := &cobra.Command{
		Use:   "init [path]",
		Short: "Deploy vibe-check agent and command assets and register the entropy reviewer into a project",
		Long: `Init deploys the embedded vibe-check agent assets (such as the
divisor-entropy structural-entropy reviewer and the vibe-check-reporter) and
command assets (such as the /vibe-check slash command) into a target project's
.opencode/agents/ and .opencode/commands/ directories. It also registers the
divisor-entropy reviewer into the project's Unbound Force policy files
(.uf/reviewer-capabilities.yaml and .uf/review-matrix.yaml) when those files
are present; absent or unexpected policy files are reported as notices and
never created, and --force does not affect registration.

With --ci, init additionally scaffolds the structural-gate caller workflow
into .github/workflows/vibe-check-structural-gate.yaml, pinned to a
vibe-check release with both a tag (with: version:) and the release's commit
SHA (uses:) — the caller invokes the reusable workflow this repository
publishes. By default the pins come from this binary's own release; pass
--vibe-version to pin a different release (requires --ci). Pin resolution is
fail-closed: a dev build, an unparseable or below-v0.3.0 tag, or an
unresolvable release commit exits 2 before any files are written. The
emitted workflow is user property — skipped on rerun even under --force,
which keeps its .opencode/-only overwrite semantics — and the run summary
ends with next steps for committing and wiring it.

Existing files are skipped by default; use --force to overwrite them. The target
path defaults to the current directory. Output is a human-readable summary by
default or a JSON object with --json.

With --json, a provenance envelope (producer, version, generatedAt) is included
by default so downstream consumers can attribute and audit the output. Use
--no-provenance to omit it for byte-reproducible output.

Exit code is 0 on success (including an all-skipped run and notice-only policy
gaps) and 2 when --vibe-version is passed without --ci, the target path is
invalid, a policy file cannot be safely rewritten (symlinked or non-regular),
an asset cannot be written, or --ci cannot resolve or validate the release
tag and commit SHA it pins to.`,
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
				Stdout:       cmd.OutOrStdout(),
				Stderr:       cmd.ErrOrStderr(),
				Path:         path,
				Force:        force,
				JSON:         jsonOut,
				NoProvenance: noProvenance,
				CI:           ci,
				VibeVersion:  vibeVersion,
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

	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing asset files instead of skipping them")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit a machine-readable JSON payload instead of a summary")
	cmd.Flags().BoolVar(&noProvenance, "no-provenance", false, "Omit provenance metadata for byte-reproducible output")
	cmd.Flags().BoolVar(&ci, "ci", false, "Additionally scaffold the structural-gate caller workflow into .github/workflows/ (pinned to a vibe-check release; skipped on rerun)")
	cmd.Flags().StringVar(&vibeVersion, "vibe-version", "", "Release tag the scaffolded caller workflow pins (default: this binary's own release; requires --ci)")

	return cmd
}
