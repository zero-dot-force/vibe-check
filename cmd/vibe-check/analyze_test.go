package main

import (
	"bytes"
	"context"
	"encoding/json"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/zero-dot-force/vibe-check/metrics"
)

// fixtureDir returns the absolute path to a named test fixture under
// internal/goadapter/testdata/. Uses runtime.Caller to locate the source
// tree, then navigates to the goadapter testdata directory.
func fixtureDir(t *testing.T, name string) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to determine test file location")
	}
	// Navigate from cmd/vibe-check/ up to repo root, then into internal/goadapter/testdata/<name>.
	repoRoot := filepath.Join(filepath.Dir(filename), "..", "..")
	return filepath.Join(repoRoot, "internal", "goadapter", "testdata", name)
}

// couplingFixtureDir returns the absolute path to the coupling test fixture.
func couplingFixtureDir(t *testing.T) string {
	t.Helper()
	return fixtureDir(t, "coupling")
}

// float64Ptr returns a pointer to the given float64 value.
func float64Ptr(v float64) *float64 {
	t := v
	return &t
}

// intPtr returns a pointer to the given int value.
func intPtr(v int) *int {
	t := v
	return &t
}

// --- Task 7.5: Flag parsing and help ---

func TestAnalyzeHelp(t *testing.T) {
	t.Parallel()

	cmd := rootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"analyze", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "analyze") {
		t.Error("help output does not mention 'analyze'")
	}
	if !strings.Contains(output, "--max-instability") {
		t.Error("help output does not mention '--max-instability'")
	}
	if !strings.Contains(output, "--max-distance") {
		t.Error("help output does not mention '--max-distance'")
	}
	if !strings.Contains(output, "--max-lcom") {
		t.Error("help output does not mention '--max-lcom'")
	}
	if !strings.Contains(output, "--no-circular-deps") {
		t.Error("help output does not mention '--no-circular-deps'")
	}
	if !strings.Contains(output, "--timeout") {
		t.Error("help output does not mention '--timeout'")
	}
}

func TestVersion(t *testing.T) {
	t.Parallel()

	cmd := rootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--version"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	// Version format: "vibe-check version dev (commit none, built unknown)"
	if !strings.Contains(output, "vibe-check") {
		t.Errorf("version output missing 'vibe-check': %q", output)
	}
	if !strings.Contains(output, "dev") {
		t.Errorf("version output missing default version 'dev': %q", output)
	}
}

func TestAnalyzeFlagParsing(t *testing.T) {
	t.Parallel()

	// Verify the command structure accepts all expected flags without error.
	cmd := analyzeCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	// Parse flags without executing (just verify they're registered).
	err := cmd.Flags().Parse([]string{
		"--max-instability", "0.5",
		"--max-distance", "0.3",
		"--max-lcom", "3",
		"--no-circular-deps",
		"--timeout", "30s",
	})
	if err != nil {
		t.Fatalf("flag parsing failed: %v", err)
	}

	// Verify parsed values.
	instability, err := cmd.Flags().GetFloat64("max-instability")
	if err != nil {
		t.Fatalf("GetFloat64 max-instability: %v", err)
	}
	if instability != 0.5 {
		t.Errorf("max-instability: got %f, want 0.5", instability)
	}

	distance, err := cmd.Flags().GetFloat64("max-distance")
	if err != nil {
		t.Fatalf("GetFloat64 max-distance: %v", err)
	}
	if distance != 0.3 {
		t.Errorf("max-distance: got %f, want 0.3", distance)
	}

	lcom, err := cmd.Flags().GetInt("max-lcom")
	if err != nil {
		t.Fatalf("GetInt max-lcom: %v", err)
	}
	if lcom != 3 {
		t.Errorf("max-lcom: got %d, want 3", lcom)
	}

	noCycles, err := cmd.Flags().GetBool("no-circular-deps")
	if err != nil {
		t.Fatalf("GetBool no-circular-deps: %v", err)
	}
	if !noCycles {
		t.Error("no-circular-deps: got false, want true")
	}

	timeout, err := cmd.Flags().GetDuration("timeout")
	if err != nil {
		t.Fatalf("GetDuration timeout: %v", err)
	}
	if timeout != 30*time.Second {
		t.Errorf("timeout: got %v, want 30s", timeout)
	}
}

// --- Task 7.6: Threshold violations ---

func TestRunAnalyze_ThresholdViolations(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	fixtureDir := couplingFixtureDir(t)

	tests := []struct {
		name           string
		opts           AnalyzeOptions
		wantExitCode   int
		wantViolations int
		wantContains   string
	}{
		{
			name: "instability_violation",
			opts: AnalyzeOptions{
				Path:           fixtureDir,
				MaxInstability: float64Ptr(0.0), // pkga and pkgc have I=1.0, exceeds 0.0
			},
			wantExitCode:   1,
			wantViolations: 2, // pkga (I=1.0) and pkgc (I=1.0)
			wantContains:   "instability",
		},
		{
			name: "instability_passes",
			opts: AnalyzeOptions{
				Path:           fixtureDir,
				MaxInstability: float64Ptr(1.0), // all modules have I <= 1.0
			},
			wantExitCode:   0,
			wantViolations: 0,
		},
		{
			name: "distance_violation",
			opts: AnalyzeOptions{
				Path:        fixtureDir,
				MaxDistance: float64Ptr(0.0), // most modules will exceed 0.0
			},
			wantExitCode: 1,
			wantContains: "distance",
		},
		{
			name: "distance_passes",
			opts: AnalyzeOptions{
				Path:        fixtureDir,
				MaxDistance: float64Ptr(1.0), // all distances are <= 1.0
			},
			wantExitCode:   0,
			wantViolations: 0,
		},
		{
			name: "lcom_passes",
			opts: AnalyzeOptions{
				Path:    fixtureDir,
				MaxLCOM: intPtr(100), // generous threshold
			},
			wantExitCode:   0,
			wantViolations: 0,
		},
		{
			name: "no_circular_deps_passes",
			opts: AnalyzeOptions{
				Path:           fixtureDir,
				NoCircularDeps: true, // coupling fixture has no cycles
			},
			wantExitCode:   0,
			wantViolations: 0,
		},
		{
			name: "multiple_violations",
			opts: AnalyzeOptions{
				Path:           fixtureDir,
				MaxInstability: float64Ptr(0.0),
				MaxDistance:    float64Ptr(0.0),
			},
			wantExitCode: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer
			tt.opts.Stdout = &stdout
			tt.opts.Stderr = &stderr

			result, err := RunAnalyze(context.Background(), tt.opts)
			if err != nil {
				t.Fatalf("RunAnalyze returned error: %v", err)
			}

			if result.ExitCode != tt.wantExitCode {
				t.Errorf("ExitCode: got %d, want %d\nstderr: %s", result.ExitCode, tt.wantExitCode, stderr.String())
			}

			if tt.wantViolations > 0 && len(result.Violations) != tt.wantViolations {
				t.Errorf("Violations count: got %d, want %d\nviolations: %v", len(result.Violations), tt.wantViolations, result.Violations)
			}

			if tt.wantContains != "" {
				found := false
				for _, v := range result.Violations {
					if strings.Contains(v, tt.wantContains) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("no violation contains %q\nviolations: %v", tt.wantContains, result.Violations)
				}
			}

			// JSON is always written to stdout, even when violations exist.
			if stdout.Len() == 0 {
				t.Error("stdout is empty, expected JSON output")
			}
		})
	}
}

func TestRunAnalyze_LCOMViolation(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// The lcom fixture has noncohesive package with LCOM=2.
	// Setting threshold to 1 should trigger a violation for noncohesive.
	var stdout, stderr bytes.Buffer
	opts := AnalyzeOptions{
		Stdout:  &stdout,
		Stderr:  &stderr,
		Path:    fixtureDir(t, "lcom"),
		MaxLCOM: intPtr(1),
	}

	result, err := RunAnalyze(context.Background(), opts)
	if err != nil {
		t.Fatalf("RunAnalyze returned error: %v", err)
	}

	if result.ExitCode != 1 {
		t.Errorf("ExitCode: got %d, want 1", result.ExitCode)
	}

	if len(result.Violations) == 0 {
		t.Fatal("expected at least one LCOM violation")
	}

	found := false
	for _, v := range result.Violations {
		if strings.Contains(v, "lcom") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("no violation contains 'lcom'\nviolations: %v", result.Violations)
	}

	// JSON should still be written to stdout.
	if stdout.Len() == 0 {
		t.Error("stdout is empty, expected JSON output")
	}
}

func TestRunAnalyze_JSONWrittenOnViolation(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	var stdout, stderr bytes.Buffer
	opts := AnalyzeOptions{
		Stdout:         &stdout,
		Stderr:         &stderr,
		Path:           couplingFixtureDir(t),
		MaxInstability: float64Ptr(0.0), // Will cause violations
	}

	result, err := RunAnalyze(context.Background(), opts)
	if err != nil {
		t.Fatalf("RunAnalyze returned error: %v", err)
	}

	// Verify exit code 1 (violations detected).
	if result.ExitCode != 1 {
		t.Errorf("ExitCode: got %d, want 1", result.ExitCode)
	}

	// Verify JSON was still written to stdout.
	if stdout.Len() == 0 {
		t.Fatal("stdout is empty, expected JSON output even with violations")
	}

	// Verify the JSON is valid and passes schema validation.
	if err := metrics.Validate(stdout.Bytes()); err != nil {
		t.Errorf("JSON output failed validation: %v", err)
	}

	// Verify violations were written to stderr.
	if stderr.Len() == 0 {
		t.Error("stderr is empty, expected violation messages")
	}
}

// --- Task 7.7: Flag validation and exit codes ---

func TestRunAnalyze_FlagValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		opts    AnalyzeOptions
		wantErr string
	}{
		{
			name:    "max_instability_too_high",
			opts:    AnalyzeOptions{MaxInstability: float64Ptr(1.5)},
			wantErr: "--max-instability",
		},
		{
			name:    "max_instability_negative",
			opts:    AnalyzeOptions{MaxInstability: float64Ptr(-0.1)},
			wantErr: "--max-instability",
		},
		{
			name:    "max_distance_too_high",
			opts:    AnalyzeOptions{MaxDistance: float64Ptr(1.5)},
			wantErr: "--max-distance",
		},
		{
			name:    "max_distance_negative",
			opts:    AnalyzeOptions{MaxDistance: float64Ptr(-0.5)},
			wantErr: "--max-distance",
		},
		{
			name:    "max_lcom_zero",
			opts:    AnalyzeOptions{MaxLCOM: intPtr(0)},
			wantErr: "--max-lcom",
		},
		{
			name:    "max_lcom_negative",
			opts:    AnalyzeOptions{MaxLCOM: intPtr(-1)},
			wantErr: "--max-lcom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer
			tt.opts.Stdout = &stdout
			tt.opts.Stderr = &stderr
			tt.opts.Path = "/nonexistent" // Won't reach adapter

			result, err := RunAnalyze(context.Background(), tt.opts)
			if err == nil {
				t.Fatal("expected error for invalid flag, got nil")
			}
			if result.ExitCode != 2 {
				t.Errorf("ExitCode: got %d, want 2", result.ExitCode)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error %q does not contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestRunAnalyze_AdapterError(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	opts := AnalyzeOptions{
		Stdout: &stdout,
		Stderr: &stderr,
		Path:   "/nonexistent/path/that/does/not/exist",
	}

	result, err := RunAnalyze(context.Background(), opts)
	if err == nil {
		t.Fatal("expected error for nonexistent path, got nil")
	}
	if result.ExitCode != 2 {
		t.Errorf("ExitCode: got %d, want 2", result.ExitCode)
	}
}

func TestRunAnalyze_ContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately.

	var stdout, stderr bytes.Buffer
	opts := AnalyzeOptions{
		Stdout: &stdout,
		Stderr: &stderr,
		Path:   couplingFixtureDir(t),
	}

	result, err := RunAnalyze(ctx, opts)
	if err == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}
	if result.ExitCode != 2 {
		t.Errorf("ExitCode: got %d, want 2", result.ExitCode)
	}

	// No partial JSON should be written.
	if stdout.Len() > 0 {
		t.Errorf("stdout should be empty on cancellation, got %d bytes", stdout.Len())
	}
}

func TestRunAnalyze_TimeoutCreatesDeadline(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Use a generous timeout that should not expire.
	var stdout, stderr bytes.Buffer
	opts := AnalyzeOptions{
		Stdout:  &stdout,
		Stderr:  &stderr,
		Path:    couplingFixtureDir(t),
		Timeout: 60 * time.Second,
	}

	result, err := RunAnalyze(context.Background(), opts)
	if err != nil {
		t.Fatalf("RunAnalyze returned error: %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("ExitCode: got %d, want 0", result.ExitCode)
	}

	// Verify JSON was produced.
	if stdout.Len() == 0 {
		t.Error("stdout is empty, expected JSON output")
	}
}

func TestRunAnalyze_BoundaryValues(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// In the coupling fixture:
	// pkga: I=1.0 (Ce=1, Ca=0), pkgb: I=0.0 (Ce=0, Ca=2), pkgc: I=1.0 (Ce=1, Ca=0)
	// Setting threshold to exactly the metric value should PASS (strict >).

	var stdout, stderr bytes.Buffer
	opts := AnalyzeOptions{
		Stdout:         &stdout,
		Stderr:         &stderr,
		Path:           couplingFixtureDir(t),
		MaxInstability: float64Ptr(1.0), // pkga and pkgc have I=1.0, which equals threshold → passes
	}

	result, err := RunAnalyze(context.Background(), opts)
	if err != nil {
		t.Fatalf("RunAnalyze returned error: %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("ExitCode: got %d, want 0 (metric == threshold should pass)\nviolations: %v", result.ExitCode, result.Violations)
	}
}

func TestRunAnalyze_ExitCodeDistinction(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Exit 0: no violations.
	t.Run("exit_0_no_violations", func(t *testing.T) {
		t.Parallel()
		var stdout, stderr bytes.Buffer
		opts := AnalyzeOptions{
			Stdout: &stdout,
			Stderr: &stderr,
			Path:   couplingFixtureDir(t),
		}
		result, err := RunAnalyze(context.Background(), opts)
		if err != nil {
			t.Fatalf("RunAnalyze returned error: %v", err)
		}
		if result.ExitCode != 0 {
			t.Errorf("ExitCode: got %d, want 0", result.ExitCode)
		}
	})

	// Exit 1: threshold violations.
	t.Run("exit_1_violations", func(t *testing.T) {
		t.Parallel()
		var stdout, stderr bytes.Buffer
		opts := AnalyzeOptions{
			Stdout:         &stdout,
			Stderr:         &stderr,
			Path:           couplingFixtureDir(t),
			MaxInstability: float64Ptr(0.0),
		}
		result, err := RunAnalyze(context.Background(), opts)
		if err != nil {
			t.Fatalf("RunAnalyze returned error: %v", err)
		}
		if result.ExitCode != 1 {
			t.Errorf("ExitCode: got %d, want 1", result.ExitCode)
		}
	})

	// Exit 2: tool failure (invalid flag).
	t.Run("exit_2_invalid_flag", func(t *testing.T) {
		t.Parallel()
		var stdout, stderr bytes.Buffer
		opts := AnalyzeOptions{
			Stdout:         &stdout,
			Stderr:         &stderr,
			Path:           "/nonexistent",
			MaxInstability: float64Ptr(2.0),
		}
		result, err := RunAnalyze(context.Background(), opts)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if result.ExitCode != 2 {
			t.Errorf("ExitCode: got %d, want 2", result.ExitCode)
		}
	})

	// Exit 2: adapter error.
	t.Run("exit_2_adapter_error", func(t *testing.T) {
		t.Parallel()
		var stdout, stderr bytes.Buffer
		opts := AnalyzeOptions{
			Stdout: &stdout,
			Stderr: &stderr,
			Path:   "/nonexistent/path",
		}
		result, err := RunAnalyze(context.Background(), opts)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if result.ExitCode != 2 {
			t.Errorf("ExitCode: got %d, want 2", result.ExitCode)
		}
	})
}

// --- Task 7.8: JSON output validation ---

func TestRunAnalyze_JSONOutputValidation(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	var stdout, stderr bytes.Buffer
	opts := AnalyzeOptions{
		Stdout: &stdout,
		Stderr: &stderr,
		Path:   couplingFixtureDir(t),
	}

	_, err := RunAnalyze(context.Background(), opts)
	if err != nil {
		t.Fatalf("RunAnalyze returned error: %v", err)
	}

	// Verify JSON passes schema validation.
	data := stdout.Bytes()
	if err := metrics.Validate(data); err != nil {
		t.Errorf("JSON output failed schema validation: %v", err)
	}
}

func TestRunAnalyze_JSONOutputPrettyPrinted(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	var stdout, stderr bytes.Buffer
	opts := AnalyzeOptions{
		Stdout: &stdout,
		Stderr: &stderr,
		Path:   couplingFixtureDir(t),
	}

	_, err := RunAnalyze(context.Background(), opts)
	if err != nil {
		t.Fatalf("RunAnalyze returned error: %v", err)
	}

	output := stdout.String()

	// Pretty-printed JSON should contain newlines and indentation.
	if !strings.Contains(output, "\n") {
		t.Error("JSON output is not pretty-printed: no newlines found")
	}
	if !strings.Contains(output, "  ") {
		t.Error("JSON output is not pretty-printed: no indentation found")
	}

	// Verify it's valid JSON by unmarshaling.
	var graph metrics.ModuleGraph
	if err := json.Unmarshal([]byte(output), &graph); err != nil {
		t.Errorf("JSON output is not valid JSON: %v", err)
	}

	// Verify expected fields are present.
	if graph.SchemaVersion != "1.1" {
		t.Errorf("SchemaVersion: got %q, want %q", graph.SchemaVersion, "1.1")
	}
	if graph.Language != "go" {
		t.Errorf("Language: got %q, want %q", graph.Language, "go")
	}
	if len(graph.Modules) == 0 {
		t.Error("Modules is empty, expected at least one module")
	}
}

func TestRunAnalyze_JSONOutputContainsModules(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	var stdout, stderr bytes.Buffer
	opts := AnalyzeOptions{
		Stdout: &stdout,
		Stderr: &stderr,
		Path:   couplingFixtureDir(t),
	}

	result, err := RunAnalyze(context.Background(), opts)
	if err != nil {
		t.Fatalf("RunAnalyze returned error: %v", err)
	}

	// Verify the graph contains the expected coupling fixture packages.
	if result.Graph == nil {
		t.Fatal("Graph is nil")
	}

	byName := make(map[string]metrics.ModuleResult)
	for _, m := range result.Graph.Modules {
		byName[m.Name] = m
	}

	expectedPkgs := []string{"pkga", "pkgb", "pkgc"}
	for _, name := range expectedPkgs {
		if _, ok := byName[name]; !ok {
			t.Errorf("expected package %q not found in results", name)
		}
	}
}

// --- validateFlags unit tests (pure function, no integration) ---

func TestValidateFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		opts    AnalyzeOptions
		wantErr bool
	}{
		{
			name: "all_nil_passes",
			opts: AnalyzeOptions{},
		},
		{
			name: "valid_instability",
			opts: AnalyzeOptions{MaxInstability: float64Ptr(0.5)},
		},
		{
			name: "valid_instability_zero",
			opts: AnalyzeOptions{MaxInstability: float64Ptr(0.0)},
		},
		{
			name: "valid_instability_one",
			opts: AnalyzeOptions{MaxInstability: float64Ptr(1.0)},
		},
		{
			name:    "invalid_instability_above",
			opts:    AnalyzeOptions{MaxInstability: float64Ptr(1.1)},
			wantErr: true,
		},
		{
			name:    "invalid_instability_below",
			opts:    AnalyzeOptions{MaxInstability: float64Ptr(-0.1)},
			wantErr: true,
		},
		{
			name: "valid_distance",
			opts: AnalyzeOptions{MaxDistance: float64Ptr(0.5)},
		},
		{
			name:    "invalid_distance_above",
			opts:    AnalyzeOptions{MaxDistance: float64Ptr(1.5)},
			wantErr: true,
		},
		{
			name:    "invalid_distance_below",
			opts:    AnalyzeOptions{MaxDistance: float64Ptr(-0.5)},
			wantErr: true,
		},
		{
			name: "valid_lcom",
			opts: AnalyzeOptions{MaxLCOM: intPtr(1)},
		},
		{
			name: "valid_lcom_large",
			opts: AnalyzeOptions{MaxLCOM: intPtr(100)},
		},
		{
			name:    "invalid_lcom_zero",
			opts:    AnalyzeOptions{MaxLCOM: intPtr(0)},
			wantErr: true,
		},
		{
			name:    "invalid_lcom_negative",
			opts:    AnalyzeOptions{MaxLCOM: intPtr(-1)},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateFlags(tt.opts)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// --- checkThresholds unit tests (pure function, no integration) ---

func TestCheckThresholds(t *testing.T) {
	t.Parallel()

	// Build a synthetic graph for threshold testing.
	graph := &metrics.ModuleGraph{
		Modules: []metrics.ModuleResult{
			{
				Module:       metrics.Module{Path: "example.com/foo", Name: "foo"},
				Instability:  0.75,
				Distance:     0.50,
				LCOM:         3,
				Abstractness: 0.0,
				Zone:         metrics.ZoneNormal,
			},
			{
				Module:       metrics.Module{Path: "example.com/bar", Name: "bar"},
				Instability:  0.25,
				Distance:     0.10,
				LCOM:         1,
				Abstractness: 0.0,
				Zone:         metrics.ZoneNormal,
			},
		},
		Cycles: []metrics.Cycle{
			{"example.com/a", "example.com/b"},
		},
	}

	tests := []struct {
		name       string
		opts       AnalyzeOptions
		wantCount  int
		wantSubstr string
	}{
		{
			name:      "no_thresholds",
			opts:      AnalyzeOptions{},
			wantCount: 0,
		},
		{
			name:       "instability_one_violation",
			opts:       AnalyzeOptions{MaxInstability: float64Ptr(0.50)},
			wantCount:  1, // foo (0.75 > 0.50), bar passes (0.25 <= 0.50)
			wantSubstr: "foo",
		},
		{
			name:      "instability_boundary_passes",
			opts:      AnalyzeOptions{MaxInstability: float64Ptr(0.75)},
			wantCount: 0, // foo (0.75 == 0.75) passes with strict >
		},
		{
			name:       "distance_one_violation",
			opts:       AnalyzeOptions{MaxDistance: float64Ptr(0.30)},
			wantCount:  1, // foo (0.50 > 0.30)
			wantSubstr: "distance",
		},
		{
			name:       "lcom_one_violation",
			opts:       AnalyzeOptions{MaxLCOM: intPtr(2)},
			wantCount:  1, // foo (3 > 2)
			wantSubstr: "lcom",
		},
		{
			name:      "lcom_boundary_passes",
			opts:      AnalyzeOptions{MaxLCOM: intPtr(3)},
			wantCount: 0, // foo (3 == 3) passes with strict >
		},
		{
			name:       "circular_deps_violation",
			opts:       AnalyzeOptions{NoCircularDeps: true},
			wantCount:  1,
			wantSubstr: "circular dependency",
		},
		{
			name:      "circular_deps_not_checked",
			opts:      AnalyzeOptions{NoCircularDeps: false},
			wantCount: 0,
		},
		{
			name: "multiple_thresholds",
			opts: AnalyzeOptions{
				MaxInstability: float64Ptr(0.50),
				MaxDistance:    float64Ptr(0.30),
				MaxLCOM:        intPtr(2),
				NoCircularDeps: true,
			},
			wantCount: 4, // foo: instability + distance + lcom, plus 1 cycle
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			violations := checkThresholds(graph, tt.opts)
			if len(violations) != tt.wantCount {
				t.Errorf("violation count: got %d, want %d\nviolations: %v", len(violations), tt.wantCount, violations)
			}
			if tt.wantSubstr != "" && len(violations) > 0 {
				found := false
				for _, v := range violations {
					if strings.Contains(strings.ToLower(v), tt.wantSubstr) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("no violation contains %q\nviolations: %v", tt.wantSubstr, violations)
				}
			}
		})
	}
}
