package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zero-dot-force/vibe-check/metrics"
	"github.com/zero-dot-force/vibe-check/metrics/report"
)

// --- Task 3.1: mode flag and RunReport wiring --------------------------------

// cleanGraph returns a schema-valid snapshot that fires no recommendation
// rules — every module metric sits strictly inside every rule bound — so the
// report renders the single green all-clear and exits 0.
func cleanGraph() metrics.ModuleGraph {
	return mkGraph(metrics.StatusComplete,
		[]metrics.ModuleResult{mkModule("ex/a", 1, 1, 0.50, 0.10, 0.20, 1)}, nil, nil)
}

// violationsGraph returns a schema-valid snapshot that fires exactly two red
// rules for ex/bad — distance 0.60 >= 0.5 (distance-refactor) and LCOM4 4
// >= 4 (lcom4-split) — so the report renders "Needs Attention" and exits 1.
func violationsGraph() metrics.ModuleGraph {
	return mkGraph(metrics.StatusComplete,
		[]metrics.ModuleResult{mkModule("ex/bad", 1, 2, 0.50, 0.10, 0.60, 4)}, nil, nil)
}

// cliRenderOptions returns the RenderOptions RunReport injects: the CLI
// producer name and semantic version as the renderer's provenance fallback.
// Tests compute the expected engine output with the same options, so the
// wiring assertion holds regardless of which version is embedded in the
// binary under test.
func cliRenderOptions() report.RenderOptions {
	return report.RenderOptions{
		FallbackProducer: producerName,
		FallbackVersion:  semanticVersion(),
	}
}

// TestRunReport_RendersEngineOutput pins the RunReport wiring: for a valid
// snapshot, stdout is byte-identical to the pure engine render for the
// selected mode, the provenance fallback carries the CLI identity, and the
// result model agrees with the rendered report.
func TestRunReport_RendersEngineOutput(t *testing.T) {
	t.Parallel()

	graph := cleanGraph()
	path := writeGraphFile(t, t.TempDir(), "clean.json", graph)

	tests := []struct {
		name string
		mode string
		want func(g *metrics.ModuleGraph, opts report.RenderOptions) string
	}{
		{name: "summary_mode", mode: ModeSummary, want: report.RenderSummary},
		{name: "detailed_mode", mode: ModeDetailed, want: report.RenderDetailed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var stdout, stderr bytes.Buffer
			result, err := RunReport(context.Background(), ReportOptions{
				Stdout:    &stdout,
				Stderr:    &stderr,
				GraphPath: path,
				Mode:      tt.mode,
			})
			if err != nil {
				t.Fatalf("RunReport returned error: %v", err)
			}
			want := tt.want(&graph, cliRenderOptions())
			if got := stdout.String(); got != want {
				t.Errorf("stdout: got %q, want %q", got, want)
			}
			if result.ExitCode != 0 {
				t.Errorf("ExitCode: got %d, want 0", result.ExitCode)
			}
			if result.Status != report.StatusAllClear {
				t.Errorf("Status: got %q, want %q", result.Status, report.StatusAllClear)
			}
			if len(result.Recommendations) != 1 {
				t.Errorf("Recommendations: got %d entries, want 1 (the all-clear)", len(result.Recommendations))
			}
			if result.Graph == nil || len(result.Graph.Modules) != 1 {
				t.Errorf("Graph: got %+v, want the parsed one-module snapshot", result.Graph)
			}
		})
	}
}

// TestRunReport_ModeValidation covers the --mode contract: the two defined
// modes are accepted, every other value (including the agent-owned
// "trending", the empty string, and wrong-case spellings) is rejected with
// exit code 2, no stdout, and an error naming the valid modes. The invalid
// cases point at a nonexistent path so a "read graph file" error would prove
// mode validation did not run before the snapshot was read.
func TestRunReport_ModeValidation(t *testing.T) {
	t.Parallel()

	cleanPath := writeGraphFile(t, t.TempDir(), "clean.json", cleanGraph())
	missingPath := filepath.Join(t.TempDir(), "missing.json")

	tests := []struct {
		name      string
		mode      string
		graphPath string
		wantErr   string // empty means expect success
	}{
		{name: "summary_explicit", mode: ModeSummary, graphPath: cleanPath},
		{name: "detailed_explicit", mode: ModeDetailed, graphPath: cleanPath},
		{name: "trending_rejected", mode: "trending", graphPath: missingPath, wantErr: `invalid --mode value "trending"`},
		{name: "empty_rejected", mode: "", graphPath: missingPath, wantErr: `invalid --mode value ""`},
		{name: "case_sensitive", mode: "Summary", graphPath: missingPath, wantErr: `invalid --mode value "Summary"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var stdout, stderr bytes.Buffer
			result, err := RunReport(context.Background(), ReportOptions{
				Stdout:    &stdout,
				Stderr:    &stderr,
				GraphPath: tt.graphPath,
				Mode:      tt.mode,
			})
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("RunReport returned error: %v", err)
				}
				if result.ExitCode != 0 {
					t.Errorf("ExitCode: got %d, want 0", result.ExitCode)
				}
				if stdout.Len() == 0 {
					t.Error("stdout must carry the rendered report")
				}
				return
			}
			if err == nil {
				t.Fatal("expected error for invalid mode, got nil")
			}
			if result.ExitCode != 2 {
				t.Errorf("ExitCode: got %d, want 2", result.ExitCode)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error %q missing %q", err.Error(), tt.wantErr)
			}
			if !strings.Contains(err.Error(), `must be "summary" or "detailed"`) {
				t.Errorf("error %q must name the valid modes", err.Error())
			}
			if stdout.Len() != 0 {
				t.Errorf("stdout must be empty on invalid mode, got %d bytes: %s", stdout.Len(), stdout.String())
			}
			if strings.Contains(err.Error(), "read graph file") {
				t.Errorf("mode validation must precede reading the snapshot: %v", err)
			}
		})
	}
}

// TestRunReport_ExitOneOnViolations covers the policy exit: a snapshot whose
// rule engine fires red recommendations renders the full report to stdout
// (byte-identical to the engine render) and returns exit code 1 with the
// Needs Attention status.
func TestRunReport_ExitOneOnViolations(t *testing.T) {
	t.Parallel()

	graph := violationsGraph()
	path := writeGraphFile(t, t.TempDir(), "violations.json", graph)

	var stdout, stderr bytes.Buffer
	result, err := RunReport(context.Background(), ReportOptions{
		Stdout:    &stdout,
		Stderr:    &stderr,
		GraphPath: path,
		Mode:      ModeSummary,
	})
	if err != nil {
		t.Fatalf("RunReport returned error: %v", err)
	}
	if result.ExitCode != 1 {
		t.Errorf("ExitCode: got %d, want 1", result.ExitCode)
	}
	if result.Status != report.StatusNeedsAttention {
		t.Errorf("Status: got %q, want %q", result.Status, report.StatusNeedsAttention)
	}
	if len(result.Recommendations) != 2 {
		t.Errorf("Recommendations: got %d entries, want 2 (distance-refactor and lcom4-split)", len(result.Recommendations))
	}
	want := report.RenderSummary(&graph, cliRenderOptions())
	if got := stdout.String(); got != want {
		t.Errorf("stdout: got %q, want %q", got, want)
	}
}

// --- Task 3.2: input handling and exit codes ----------------------------------

// TestRunReport_ExitTwoOnBadInput covers the input-handling failure contract:
// a missing, unreadable, or schema-invalid snapshot is a tool failure — exit
// code 2, an error naming the failing step, and nothing on stdout. The
// unsupported-version row pins the schema-version gate both ways: the 1.3
// document `vibe-check analyze` emits passes metrics.Validate (asserted as a
// precondition), while a mutated version string is rejected.
func TestRunReport_ExitTwoOnBadInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(t *testing.T) string
		wantSub string
	}{
		{
			name: "missing_file",
			setup: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "missing.json")
			},
			wantSub: "read graph file",
		},
		{
			name: "unreadable_directory",
			setup: func(t *testing.T) string {
				// A directory path is a deterministic unreadable input:
				// os.ReadFile fails with "is a directory" without relying on
				// file permissions, which a root user would bypass.
				return t.TempDir()
			},
			wantSub: "read graph file",
		},
		{
			name: "not_json",
			setup: func(t *testing.T) string {
				path := filepath.Join(t.TempDir(), "not-json.json")
				if err := os.WriteFile(path, []byte("not valid json"), 0o644); err != nil {
					t.Fatalf("write fixture: %v", err)
				}
				return path
			},
			wantSub: "validate graph file",
		},
		{
			name: "schema_invalid_missing_fields",
			setup: func(t *testing.T) string {
				path := filepath.Join(t.TempDir(), "missing-fields.json")
				// Readable JSON, but the required modules/cycles/warnings/status
				// fields are absent, so it is not a schema-valid ModuleGraph.
				if err := os.WriteFile(path, []byte(`{"schemaVersion":"1.3","language":"go"}`), 0o644); err != nil {
					t.Fatalf("write fixture: %v", err)
				}
				return path
			},
			wantSub: "validate graph file",
		},
		{
			name: "unsupported_schema_version",
			setup: func(t *testing.T) string {
				data, err := json.MarshalIndent(cleanGraph(), "", "  ")
				if err != nil {
					t.Fatalf("marshal fixture: %v", err)
				}
				// Precondition: the 1.3 snapshot `vibe-check analyze` emits
				// (metrics.SchemaVersionCurrent) must pass validation.
				if err := metrics.Validate(data); err != nil {
					t.Fatalf("precondition: 1.3 fixture must pass Validate, got: %v", err)
				}
				var raw map[string]interface{}
				if err := json.Unmarshal(data, &raw); err != nil {
					t.Fatalf("unmarshal fixture: %v", err)
				}
				raw["schemaVersion"] = "9.9"
				bad, err := json.Marshal(raw)
				if err != nil {
					t.Fatalf("marshal mutated fixture: %v", err)
				}
				path := filepath.Join(t.TempDir(), "bad-version.json")
				if err := os.WriteFile(path, bad, 0o644); err != nil {
					t.Fatalf("write fixture: %v", err)
				}
				return path
			},
			wantSub: "unsupported schema version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			path := tt.setup(t)

			var stdout, stderr bytes.Buffer
			result, err := RunReport(context.Background(), ReportOptions{
				Stdout:    &stdout,
				Stderr:    &stderr,
				GraphPath: path,
				Mode:      ModeSummary,
			})
			if err == nil {
				t.Fatal("expected error for bad input, got nil")
			}
			if result.ExitCode != 2 {
				t.Errorf("ExitCode: got %d, want 2", result.ExitCode)
			}
			if stdout.Len() != 0 {
				t.Errorf("stdout must be empty on bad input, got %d bytes: %s", stdout.Len(), stdout.String())
			}
			if !strings.Contains(err.Error(), tt.wantSub) {
				t.Errorf("error %q does not contain %q", err.Error(), tt.wantSub)
			}
		})
	}
}

// TestRunReport_ParseErrorAfterValidation covers the defensive unmarshal
// branch: a document that passes schema validation (ca is a non-negative
// number) but cannot be unmarshaled into the typed graph (ca is not an
// integer) still exits 2 with nothing on stdout.
func TestRunReport_ParseErrorAfterValidation(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// ca 1.5 is a valid schema number but not an integer, so typed unmarshal
	// fails after validation succeeds.
	badModule := `{"schemaVersion":"1.3","language":"go","modules":[{"path":"ex/a","name":"a","ca":1.5,"ce":0,"instability":0.5,"abstractness":0.1,"distance":0.4,"lcom":1,"exportedTypes":2,"abstractTypes":1,"zone":"normal"}],"cycles":[],"warnings":[],"status":"complete"}`
	if err := metrics.Validate([]byte(badModule)); err != nil {
		t.Fatalf("precondition: fixture must pass Validate, got: %v", err)
	}
	path := filepath.Join(dir, "bad-ca.json")
	if err := os.WriteFile(path, []byte(badModule), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	var stdout, stderr bytes.Buffer
	result, err := RunReport(context.Background(), ReportOptions{
		Stdout:    &stdout,
		Stderr:    &stderr,
		GraphPath: path,
		Mode:      ModeSummary,
	})
	if err == nil {
		t.Fatal("expected parse error, got nil")
	}
	if result.ExitCode != 2 {
		t.Errorf("ExitCode: got %d, want 2", result.ExitCode)
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout must be empty on parse error, got: %s", stdout.String())
	}
	if !strings.Contains(err.Error(), "parse graph file") {
		t.Errorf("error %q does not contain 'parse graph file'", err.Error())
	}
}

// TestReportCommand_ViolationsExitOne covers the CLI surface of the policy
// exit: a snapshot whose rule engine fires red recommendations still writes
// the full report to stdout, and the command layer signals exit code 1 through
// exitCodeError so main() maps it to the process exit code — the spec's
// exit-codes requirement that the command exits 1 while still writing the full
// report to stdout.
func TestReportCommand_ViolationsExitOne(t *testing.T) {
	t.Parallel()

	graph := violationsGraph()
	path := writeGraphFile(t, t.TempDir(), "violations.json", graph)

	cmd := rootCmd()
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs([]string{"report", path})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected exit-1 error for violation-severity recommendations, got nil")
	}
	var ece *exitCodeError
	if !errors.As(err, &ece) {
		t.Fatalf("error is not *exitCodeError: %T (%v)", err, err)
	}
	if ece.code != 1 {
		t.Errorf("exit code: got %d, want 1", ece.code)
	}
	// The full report must still reach stdout — byte-identical to the engine
	// render for the same snapshot and CLI render options.
	want := report.RenderSummary(&graph, cliRenderOptions())
	if got := out.String(); got != want {
		t.Errorf("stdout: got %q, want %q", got, want)
	}
	if !strings.Contains(errOut.String(), "violation-severity recommendations detected") {
		t.Errorf("stderr missing violation diagnostic: %s", errOut.String())
	}
}

// TestReportCommand_BadInputExitsTwo covers the CLI surface of input
// failures: unreadable or schema-invalid snapshots exit 2 via exitCodeError
// with the diagnostic on stderr and nothing on stdout. Together with
// TestReportCommand_DefaultModeSummary (exit 0), ViolationsExitOne (exit 1),
// and InvalidModeExitsTwo (exit 2), this completes the exit-code matrix over
// the rootCmd surface.
func TestReportCommand_BadInputExitsTwo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(t *testing.T) string
		wantSub string
	}{
		{
			name: "missing_file",
			setup: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "missing.json")
			},
			wantSub: "read graph file",
		},
		{
			name: "unreadable_directory",
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			wantSub: "read graph file",
		},
		{
			name: "not_json",
			setup: func(t *testing.T) string {
				path := filepath.Join(t.TempDir(), "not-json.json")
				if err := os.WriteFile(path, []byte("not valid json"), 0o644); err != nil {
					t.Fatalf("write fixture: %v", err)
				}
				return path
			},
			wantSub: "validate graph file",
		},
		{
			name: "schema_invalid",
			setup: func(t *testing.T) string {
				path := filepath.Join(t.TempDir(), "missing-fields.json")
				if err := os.WriteFile(path, []byte(`{"schemaVersion":"1.3","language":"go"}`), 0o644); err != nil {
					t.Fatalf("write fixture: %v", err)
				}
				return path
			},
			wantSub: "validate graph file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			path := tt.setup(t)

			cmd := rootCmd()
			var out, errOut bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetErr(&errOut)
			cmd.SetArgs([]string{"report", path})

			err := cmd.Execute()
			if err == nil {
				t.Fatal("expected exit-2 error for bad input, got nil")
			}
			var ece *exitCodeError
			if !errors.As(err, &ece) {
				t.Fatalf("error is not *exitCodeError: %T (%v)", err, err)
			}
			if ece.code != 2 {
				t.Errorf("exit code: got %d, want 2", ece.code)
			}
			if out.Len() != 0 {
				t.Errorf("stdout must be empty on bad input, got %d bytes: %s", out.Len(), out.String())
			}
			if !strings.Contains(errOut.String(), tt.wantSub) {
				t.Errorf("stderr missing %q: %s", tt.wantSub, errOut.String())
			}
		})
	}
}

// TestReportCommand_DefaultModeSummary covers the cobra wiring: the report
// subcommand is registered on the root command, and running it without
// --mode defaults to summary mode — stdout is the engine's summary render
// and carries none of the detailed-only sections.
func TestReportCommand_DefaultModeSummary(t *testing.T) {
	t.Parallel()

	graph := cleanGraph()
	path := writeGraphFile(t, t.TempDir(), "clean.json", graph)

	cmd := rootCmd()
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs([]string{"report", path})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: unexpected error: %v\nstderr: %s", err, errOut.String())
	}
	want := report.RenderSummary(&graph, cliRenderOptions())
	if got := out.String(); got != want {
		t.Errorf("stdout: got %q, want %q", got, want)
	}
	if strings.Contains(out.String(), "## 📊 Package Metrics") {
		t.Error("default mode must not render the detailed package metrics table")
	}
}

// TestReportCommand_InvalidModeExitsTwo covers the CLI surface of mode
// validation: the spec's trending scenario exits 2 via exitCodeError with
// the mode diagnostic on stderr and nothing on stdout.
func TestReportCommand_InvalidModeExitsTwo(t *testing.T) {
	t.Parallel()

	path := writeGraphFile(t, t.TempDir(), "clean.json", cleanGraph())

	cmd := rootCmd()
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs([]string{"report", "--mode", "trending", path})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected exit-2 error for invalid mode, got nil")
	}
	var ece *exitCodeError
	if !errors.As(err, &ece) {
		t.Fatalf("error is not *exitCodeError: %T (%v)", err, err)
	}
	if ece.code != 2 {
		t.Errorf("exit code: got %d, want 2", ece.code)
	}
	if out.Len() != 0 {
		t.Errorf("stdout must be empty on invalid mode, got %d bytes: %s", out.Len(), out.String())
	}
	if !strings.Contains(errOut.String(), `invalid --mode value "trending"`) {
		t.Errorf("stderr missing mode diagnostic: %s", errOut.String())
	}
}

func TestReportCommand_Help(t *testing.T) {
	t.Parallel()
	cmd := rootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"report", "--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := out.String()
	for _, want := range []string{"report", "--mode", "summary", "detailed", "--json"} {
		if !strings.Contains(output, want) {
			t.Errorf("help output missing %q", want)
		}
	}
}

// --- Task 3.3: golden-file tests ----------------------------------------------

// reportFixtureDir holds the committed ModuleGraph fixtures and their golden
// files. The fixtures mirror the metrics/render_test.go fixture shapes
// (cleanMods, violationMods, detailedFixture) so the CLI goldens and the
// renderer exact-output tests pin the same graphs through both the engine
// and the command surface.
const reportFixtureDir = "testdata/report"

// updateGoldens rewrites the committed golden files instead of comparing
// them — the standard Go golden-file regeneration switch:
//
//	go test ./cmd/vibe-check/ -update
//
// It governs every golden family in this package: the report goldens under
// testdata/report and the diff goldens under testdata/diff. Regeneration
// runs inside the test binary, so the no-provenance golden records exactly
// the fallback identity that binary renders.
var updateGoldens = flag.Bool("update", false, "rewrite golden files under testdata/report and testdata/diff instead of comparing")

// reportGoldenFixtures lists every committed fixture with the behavior the
// report command must produce for it:
//
//   - clean: no rule fires, so the report renders the single green
//     all-clear and the command exits 0 (the spec's Clean scenario).
//   - violations: pkg/monolith fires the red lcom4-split rule and pkg/hub
//     the yellow ce-decouple rule, so the command exits 1 while the full
//     report still reaches stdout (the spec's Violations scenario).
//   - example: the fictional example-report fixture (the metrics
//     detailedFixture shape), which fires the red cycle-break rule for both
//     cycle members plus the green duplication watch; task 5.4 regenerates
//     the committed example report from this fixture.
//   - no-provenance: the clean graph without a provenance object, so the
//     header falls back to the rendering binary's producer name and embedded
//     version (the spec's Report header fallback). Its golden therefore
//     carries the test binary's fallback identity, "vibe-check dev" —
//     regenerate with -update if the embedded default version ever changes.
//
// The three provenance-carrying fixtures pin producer and version so their
// goldens stay stable no matter which version is embedded in the binary
// under test.
var reportGoldenFixtures = []struct {
	name           string // fixture base name: testdata/report/<name>.json
	exitCode       int    // exit code the fixture's recommendations require
	wantProvenance bool   // whether the fixture pins a provenance object
}{
	{name: "clean", exitCode: 0, wantProvenance: true},
	{name: "violations", exitCode: 1, wantProvenance: true},
	{name: "example", exitCode: 1, wantProvenance: true},
	{name: "no-provenance", exitCode: 0, wantProvenance: false},
}

// reportGoldenModes lists the report detail levels that must render for
// every fixture; the golden for a fixture/mode pair lives at
// testdata/report/<fixture>.<mode>.golden.md.
var reportGoldenModes = []string{ModeSummary, ModeDetailed}

// reportGoldenPath returns the committed golden file for a fixture/mode pair.
func reportGoldenPath(fixture, mode string) string {
	return filepath.Join(reportFixtureDir, fixture+"."+mode+".golden.md")
}

// TestRunReport_GoldenFiles pins the report command's output contract on
// committed fixtures: for every fixture and both modes, RunReport writes
// stdout byte-identical to the committed golden and returns the fixture's
// exit code. Because the goldens are committed bytes, every test run is also
// a repeated-runs byte-stability check across machines and time — the golden
// was rendered once, and each later run must reproduce it exactly.
func TestRunReport_GoldenFiles(t *testing.T) {
	t.Parallel()

	for _, fx := range reportGoldenFixtures {
		t.Run(fx.name, func(t *testing.T) {
			t.Parallel()
			for _, mode := range reportGoldenModes {
				t.Run(mode, func(t *testing.T) {
					t.Parallel()

					var stdout, stderr bytes.Buffer
					result, err := RunReport(context.Background(), ReportOptions{
						Stdout:    &stdout,
						Stderr:    &stderr,
						GraphPath: filepath.Join(reportFixtureDir, fx.name+".json"),
						Mode:      mode,
					})
					if err != nil {
						t.Fatalf("RunReport returned error: %v", err)
					}
					if result.ExitCode != fx.exitCode {
						t.Errorf("ExitCode: got %d, want %d", result.ExitCode, fx.exitCode)
					}
					got := stdout.String()

					golden := reportGoldenPath(fx.name, mode)
					if *updateGoldens {
						if err := os.WriteFile(golden, []byte(got), 0o644); err != nil {
							t.Fatalf("update golden %s: %v", golden, err)
						}
						return
					}
					want, err := os.ReadFile(golden)
					if err != nil {
						t.Fatalf("read golden %s: %v (regenerate with: go test ./cmd/vibe-check/ -update)", golden, err)
					}
					if got != string(want) {
						t.Errorf("report output for %s in %s mode does not match the committed golden %s:\ngot:\n%s\nwant:\n%s\nIf the change is intentional, regenerate with: go test ./cmd/vibe-check/ -update",
							fx.name, mode, golden, got, want)
					}
				})
			}
		})
	}
}

// TestReportCommand_GoldenDetailed pins the cobra surface against the same
// committed goldens: the report command with --mode detailed on the example
// fixture routes the golden bytes through the command layer and exits 1 via
// exitCodeError while the full report still reaches stdout — closing the
// loop from flag parsing to rendered bytes on a committed file.
func TestReportCommand_GoldenDetailed(t *testing.T) {
	t.Parallel()

	// Under -update this test would race the golden-table test writing the
	// same file; regeneration is TestRunReport_GoldenFiles' job (the single
	// writer), so this comparison-only test stands down.
	if *updateGoldens {
		t.Skip("golden regeneration is handled by TestRunReport_GoldenFiles")
	}

	cmd := rootCmd()
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs([]string{"report", "--mode", ModeDetailed, filepath.Join(reportFixtureDir, "example.json")})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected exit-1 error for the example fixture's cycle violations, got nil")
	}
	var ece *exitCodeError
	if !errors.As(err, &ece) {
		t.Fatalf("error is not *exitCodeError: %T (%v)", err, err)
	}
	if ece.code != 1 {
		t.Errorf("exit code: got %d, want 1", ece.code)
	}
	want, err := os.ReadFile(reportGoldenPath("example", ModeDetailed))
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	if got := out.String(); got != string(want) {
		t.Errorf("command stdout does not match the committed golden:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

// TestReportGoldenFixtures_Contract pins the committed fixtures themselves:
// every fixture passes metrics.Validate — the same schema gate RunReport
// enforces — and the provenance discipline that keeps the goldens
// deterministic. The provenance-carrying fixtures pin producer "vibe-check"
// and version "1.2.3" so their headers are stable regardless of the version
// embedded in the binary under test, while the no-provenance fixture omits
// the object entirely so its golden exercises the header fallback branch.
func TestReportGoldenFixtures_Contract(t *testing.T) {
	t.Parallel()

	const pinnedProducer, pinnedVersion = "vibe-check", "1.2.3"
	for _, fx := range reportGoldenFixtures {
		t.Run(fx.name, func(t *testing.T) {
			t.Parallel()

			data, err := os.ReadFile(filepath.Join(reportFixtureDir, fx.name+".json"))
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}
			if err := metrics.Validate(data); err != nil {
				t.Fatalf("fixture is not schema-valid: %v", err)
			}
			var graph metrics.ModuleGraph
			if err := json.Unmarshal(data, &graph); err != nil {
				t.Fatalf("unmarshal fixture: %v", err)
			}

			if !fx.wantProvenance {
				if graph.Provenance != nil {
					t.Errorf("no-provenance fixture carries a provenance object %+v; the fallback branch requires it omitted", graph.Provenance)
				}
				return
			}
			if graph.Provenance == nil {
				t.Fatalf("%s fixture omits provenance; pin it so the golden header is version-independent", fx.name)
			}
			if got := graph.Provenance.Producer; got != pinnedProducer {
				t.Errorf("%s producer: got %q, want %q", fx.name, got, pinnedProducer)
			}
			if got := graph.Provenance.Version; got != pinnedVersion {
				t.Errorf("%s version: got %q, want %q", fx.name, got, pinnedVersion)
			}
		})
	}
}

// --- Task 3.4: --json output mode ----------------------------------------------

// TestRunReport_JSONModeExclusion covers the --json/--mode mutual exclusion
// at the RunReport level: --json together with an explicitly requested mode
// is a usage error — exit code 2, an error naming both flags, nothing on
// stdout — rejected before the snapshot is read (the rejected cases point
// at a nonexistent path, so a "read graph file" error would prove the flag
// check did not run first). --json without an explicit mode succeeds even
// though the mode flag's default value fills Mode, because only an explicit
// selection conflicts; an explicit mode without --json behaves exactly as
// before.
func TestRunReport_JSONModeExclusion(t *testing.T) {
	t.Parallel()

	cleanPath := writeGraphFile(t, t.TempDir(), "clean.json", cleanGraph())
	missingPath := filepath.Join(t.TempDir(), "missing.json")

	tests := []struct {
		name         string
		json         bool
		mode         string
		modeExplicit bool
		graphPath    string
		wantErr      string // empty means expect success
	}{
		{
			name: "json_with_explicit_detailed", json: true, mode: ModeDetailed, modeExplicit: true,
			graphPath: missingPath, wantErr: "--json and --mode are mutually exclusive",
		},
		{
			name: "json_with_explicit_default_summary", json: true, mode: ModeSummary, modeExplicit: true,
			graphPath: missingPath, wantErr: "--json and --mode are mutually exclusive",
		},
		{
			// The flag-conflict diagnosis takes precedence over the
			// invalid-mode diagnosis: --json --mode trending is a flag
			// conflict first, a bad mode value second.
			name: "json_with_explicit_invalid_mode", json: true, mode: "trending", modeExplicit: true,
			graphPath: missingPath, wantErr: "--json and --mode are mutually exclusive",
		},
		{
			name: "json_without_explicit_mode", json: true, mode: ModeSummary, modeExplicit: false,
			graphPath: cleanPath,
		},
		{
			name: "explicit_mode_without_json", json: false, mode: ModeDetailed, modeExplicit: true,
			graphPath: cleanPath,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var stdout, stderr bytes.Buffer
			result, err := RunReport(context.Background(), ReportOptions{
				Stdout:       &stdout,
				Stderr:       &stderr,
				GraphPath:    tt.graphPath,
				Mode:         tt.mode,
				ModeExplicit: tt.modeExplicit,
				JSON:         tt.json,
			})
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("RunReport returned error: %v", err)
				}
				if result.ExitCode != 0 {
					t.Errorf("ExitCode: got %d, want 0", result.ExitCode)
				}
				if stdout.Len() == 0 {
					t.Error("stdout must carry the rendered report")
				}
				return
			}
			if err == nil {
				t.Fatal("expected usage error for conflicting flags, got nil")
			}
			if result.ExitCode != 2 {
				t.Errorf("ExitCode: got %d, want 2", result.ExitCode)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error %q missing %q", err.Error(), tt.wantErr)
			}
			if stdout.Len() != 0 {
				t.Errorf("stdout must be empty on conflicting flags, got %d bytes: %s", stdout.Len(), stdout.String())
			}
			if strings.Contains(err.Error(), "read graph file") {
				t.Errorf("flag validation must precede reading the snapshot: %v", err)
			}
		})
	}
}

// TestRunReport_JSONOutput covers the --json output contract at the
// RunReport level (the spec's JSON invocation scenario): stdout is the
// marshaled report model — byte-identical to json.MarshalIndent of
// report.BuildReportModel over the same graph and recommendations, plus
// the single trailing newline — and the payload carries the model's
// content: aggregates, the exact per-dimension grades, the overall status
// word, and the fired recommendations with machine-readable severity
// strings. The exit code follows the recommendations exactly as in markdown
// mode: 0 for the clean graph, 1 for the violations graph with the payload
// still on stdout.
func TestRunReport_JSONOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		graph      metrics.ModuleGraph
		wantExit   int
		wantStatus string
		wantGrades report.ReportGrades
		wantRecs   int
	}{
		{
			name: "clean_graph_exits_zero", graph: cleanGraph(),
			wantExit: 0, wantStatus: report.StatusAllClear, wantRecs: 1,
			// ex/a: mean distance 0.20 (B), max LCOM4 1 (A), no cycles, no
			// duplication, no instability extremes.
			wantGrades: report.ReportGrades{
				AverageDistance:    report.GradeB,
				MaxLCOM4:           report.GradeA,
				CycleCount:         report.GradeA,
				DuplicationPercent: report.GradeA,
				InstabilitySpread:  report.GradeA,
			},
		},
		{
			name: "violations_graph_exits_one", graph: violationsGraph(),
			wantExit: 1, wantStatus: report.StatusNeedsAttention, wantRecs: 2,
			// ex/bad: mean distance 0.60 (C), max LCOM4 4 (C+), no cycles,
			// no duplication, no instability extremes.
			wantGrades: report.ReportGrades{
				AverageDistance:    report.GradeC,
				MaxLCOM4:           report.GradeCPlus,
				CycleCount:         report.GradeA,
				DuplicationPercent: report.GradeA,
				InstabilitySpread:  report.GradeA,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			path := writeGraphFile(t, t.TempDir(), "graph.json", tt.graph)

			var stdout, stderr bytes.Buffer
			result, err := RunReport(context.Background(), ReportOptions{
				Stdout:    &stdout,
				Stderr:    &stderr,
				GraphPath: path,
				Mode:      ModeSummary,
				JSON:      true,
			})
			if err != nil {
				t.Fatalf("RunReport returned error: %v", err)
			}
			if result.ExitCode != tt.wantExit {
				t.Errorf("ExitCode: got %d, want %d", result.ExitCode, tt.wantExit)
			}

			// Wiring: stdout is exactly the marshaled engine model.
			wantData, err := json.MarshalIndent(
				report.BuildReportModel(&tt.graph, report.ComputeRecommendations(&tt.graph)), "", "  ")
			if err != nil {
				t.Fatalf("marshal expected model: %v", err)
			}
			if got, want := stdout.String(), string(wantData)+"\n"; got != want {
				t.Errorf("stdout: got %q, want %q", got, want)
			}

			// Content: decode the payload and pin the model's fields.
			var model report.ReportModel
			if err := json.Unmarshal(stdout.Bytes(), &model); err != nil {
				t.Fatalf("unmarshal payload: %v", err)
			}
			if model.Status != tt.wantStatus {
				t.Errorf("Status: got %q, want %q", model.Status, tt.wantStatus)
			}
			if model.Grades != tt.wantGrades {
				t.Errorf("Grades: got %+v, want %+v", model.Grades, tt.wantGrades)
			}
			if len(model.Recommendations) != tt.wantRecs {
				t.Errorf("Recommendations: got %d entries, want %d", len(model.Recommendations), tt.wantRecs)
			}
			for i, r := range model.Recommendations {
				switch r.Severity {
				case report.SeverityRed, report.SeverityYellow, report.SeverityGreen:
				default:
					t.Errorf("Recommendations[%d].Severity %q is not a machine-readable severity string", i, r.Severity)
				}
			}
			if len(result.Recommendations) != len(model.Recommendations) {
				t.Errorf("result carries %d recommendations, payload carries %d: they must agree",
					len(result.Recommendations), len(model.Recommendations))
			}
		})
	}
}

// TestReportCommand_JSONFlags covers the cobra wiring of --json: the spec's
// flag-mutual-exclusion scenario (both --json and --mode passed exits 2 via
// exitCodeError with the usage error on stderr and nothing on stdout —
// including when the explicitly passed mode is the default summary), the
// JSON invocation scenario (--json on a snapshot with violations writes the
// JSON payload to stdout and exits 1), and the clean invocation (exit 0).
func TestReportCommand_JSONFlags(t *testing.T) {
	t.Parallel()

	cleanPath := writeGraphFile(t, t.TempDir(), "clean.json", cleanGraph())
	violationsPath := writeGraphFile(t, t.TempDir(), "violations.json", violationsGraph())

	tests := []struct {
		name       string
		args       []string
		wantExit   int
		graph      metrics.ModuleGraph // hasPayload: stdout must equal this graph's marshaled model
		hasPayload bool
		wantStderr string
	}{
		{
			name: "json_and_mode_rejected", wantExit: 2,
			wantStderr: "--json and --mode are mutually exclusive",
			args:       []string{"report", "--json", "--mode", ModeDetailed, cleanPath},
		},
		{
			name: "json_and_explicit_default_mode_rejected", wantExit: 2,
			wantStderr: "--json and --mode are mutually exclusive",
			args:       []string{"report", "--json", "--mode", ModeSummary, cleanPath},
		},
		{
			name: "json_clean_exits_zero", wantExit: 0, hasPayload: true, graph: cleanGraph(),
			args: []string{"report", "--json", cleanPath},
		},
		{
			name: "json_violations_exits_one", wantExit: 1, hasPayload: true, graph: violationsGraph(),
			wantStderr: "violation-severity recommendations detected",
			args:       []string{"report", "--json", violationsPath},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := rootCmd()
			var out, errOut bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetErr(&errOut)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if tt.wantExit == 0 {
				if err != nil {
					t.Fatalf("Execute: unexpected error: %v\nstderr: %s", err, errOut.String())
				}
			} else {
				var ece *exitCodeError
				if !errors.As(err, &ece) {
					t.Fatalf("error is not *exitCodeError: %T (%v)", err, err)
				}
				if ece.code != tt.wantExit {
					t.Errorf("exit code: got %d, want %d", ece.code, tt.wantExit)
				}
			}
			if tt.wantStderr != "" && !strings.Contains(errOut.String(), tt.wantStderr) {
				t.Errorf("stderr missing %q: %s", tt.wantStderr, errOut.String())
			}

			var wantOut string
			if tt.hasPayload {
				wantData, err := json.MarshalIndent(
					report.BuildReportModel(&tt.graph, report.ComputeRecommendations(&tt.graph)), "", "  ")
				if err != nil {
					t.Fatalf("marshal expected model: %v", err)
				}
				wantOut = string(wantData) + "\n"
			}
			if got := out.String(); got != wantOut {
				t.Errorf("stdout: got %q, want %q", got, wantOut)
			}
		})
	}
}

// reportGoldenJSONPath returns the committed golden file for a fixture's
// --json output: testdata/report/<fixture>.golden.json. The JSON model has
// no detail levels, so — unlike the markdown goldens — there is one golden
// per fixture, not per fixture/mode pair.
func reportGoldenJSONPath(fixture string) string {
	return filepath.Join(reportFixtureDir, fixture+".golden.json")
}

// TestRunReport_GoldenJSONFiles pins the --json output contract on the
// committed fixtures: for every fixture, RunReport with --json writes
// stdout byte-identical to the committed golden and returns the fixture's
// exit code — the JSON twin of TestRunReport_GoldenFiles, and like it every
// run doubles as a repeated-runs byte-stability check across machines and
// time. The no-provenance golden is byte-identical to the clean golden
// because the JSON model carries no producer identity; committing both
// keeps the table uniform and automatically exercises the fallback if the
// model ever gains a provenance field.
func TestRunReport_GoldenJSONFiles(t *testing.T) {
	t.Parallel()

	for _, fx := range reportGoldenFixtures {
		t.Run(fx.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer
			result, err := RunReport(context.Background(), ReportOptions{
				Stdout:    &stdout,
				Stderr:    &stderr,
				GraphPath: filepath.Join(reportFixtureDir, fx.name+".json"),
				Mode:      ModeSummary,
				JSON:      true,
			})
			if err != nil {
				t.Fatalf("RunReport returned error: %v", err)
			}
			if result.ExitCode != fx.exitCode {
				t.Errorf("ExitCode: got %d, want %d", result.ExitCode, fx.exitCode)
			}
			got := stdout.String()

			golden := reportGoldenJSONPath(fx.name)
			if *updateGoldens {
				if err := os.WriteFile(golden, []byte(got), 0o644); err != nil {
					t.Fatalf("update golden %s: %v", golden, err)
				}
				return
			}
			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("read golden %s: %v (regenerate with: go test ./cmd/vibe-check/ -update)", golden, err)
			}
			if got != string(want) {
				t.Errorf("report JSON for %s does not match the committed golden %s:\ngot:\n%s\nwant:\n%s\nIf the change is intentional, regenerate with: go test ./cmd/vibe-check/ -update",
					fx.name, golden, got, want)
			}
		})
	}
}

// TestRunReport_JSONByteStable pins the byte-stability requirement for
// --json: two renders of the same snapshot are byte-identical, and the
// payload's four top-level keys appear in the model's declaration order —
// aggregates, grades, status, recommendations — the stable field order the
// spec's JSON output requirement fixes (struct marshaling, never a map).
// This test reads and writes no golden files, so it is safe under -update.
func TestRunReport_JSONByteStable(t *testing.T) {
	t.Parallel()

	path := filepath.Join(reportFixtureDir, "example.json")
	render := func() string {
		var stdout, stderr bytes.Buffer
		if _, err := RunReport(context.Background(), ReportOptions{
			Stdout:    &stdout,
			Stderr:    &stderr,
			GraphPath: path,
			Mode:      ModeSummary,
			JSON:      true,
		}); err != nil {
			t.Fatalf("RunReport returned error: %v", err)
		}
		return stdout.String()
	}

	first, second := render(), render()
	if first != second {
		t.Errorf("two --json renders of the same snapshot differ:\nfirst:\n%s\nsecond:\n%s", first, second)
	}

	if !strings.HasPrefix(first, "{\n  \"aggregates\": {") {
		t.Errorf("payload must open with the aggregates object, got: %q", first[:min(len(first), 32)])
	}
	at := strings.Index(first, `"aggregates"`)
	for _, key := range []string{`"grades"`, `"status"`, `"recommendations"`} {
		next := strings.Index(first, key)
		if next < 0 {
			t.Fatalf("payload missing top-level key %s", key)
		}
		if next <= at {
			t.Errorf("top-level key %s at byte %d does not follow the preceding key at byte %d: field order is not stable", key, next, at)
		}
		at = next
	}
}
