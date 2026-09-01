package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zero-dot-force/vibe-check/metrics"
)

// --- Fixture helpers ---------------------------------------------------------

// mkModule builds a ModuleResult with the given stored metric values. The raw
// exportedTypes/abstractTypes counts are fixed placeholders: these fixtures set
// the stored instability/abstractness/distance/lcom values directly to exercise
// specific deltas, and ComputeDelta operates on those stored values rather than
// recomputing the Martin formulas, so internal metric-formula consistency is not
// required here.
func mkModule(path string, ca, ce int, inst, abst, dist float64, lcom int) metrics.ModuleResult {
	return metrics.ModuleResult{
		Module: metrics.Module{
			Path:          path,
			Name:          path,
			Ca:            ca,
			Ce:            ce,
			ExportedTypes: 2,
			AbstractTypes: 1,
		},
		Instability:  metrics.Instability(inst),
		Abstractness: metrics.Abstractness(abst),
		Distance:     metrics.Distance(dist),
		LCOM:         metrics.LCOM(lcom),
		Zone:         metrics.ZoneNormal,
	}
}

// mkGraph builds a schema-valid ModuleGraph. Nil slices are normalized to empty
// (non-nil) slices so the marshaled document carries JSON arrays (never null),
// which metrics.Validate requires.
func mkGraph(status metrics.Status, mods []metrics.ModuleResult, cycles []metrics.Cycle, warnings []metrics.Warning) metrics.ModuleGraph {
	if mods == nil {
		mods = []metrics.ModuleResult{}
	}
	if cycles == nil {
		cycles = []metrics.Cycle{}
	}
	if warnings == nil {
		warnings = []metrics.Warning{}
	}
	return metrics.ModuleGraph{
		SchemaVersion: metrics.SchemaVersionCurrent,
		Language:      "go",
		Modules:       mods,
		Cycles:        cycles,
		Warnings:      warnings,
		Status:        status,
	}
}

// improvementFixtures returns a base/PR pair where ex/a improves (ΔCe -1,
// ΔInstability -0.20, ΔDistance -0.20, ΔLCOM -2) and a base cycle is resolved.
// Expected verdict APPROVE, direction improving.
func improvementFixtures() (base, pr metrics.ModuleGraph) {
	base = mkGraph(metrics.StatusComplete,
		[]metrics.ModuleResult{
			mkModule("ex/a", 1, 2, 0.50, 0.10, 0.40, 3),
			mkModule("ex/b", 2, 0, 0.00, 0.20, 0.80, 1),
		},
		[]metrics.Cycle{{"ex/b", "ex/a"}}, // deliberately unsorted; ComputeDelta normalizes
		nil)
	pr = mkGraph(metrics.StatusComplete,
		[]metrics.ModuleResult{
			mkModule("ex/a", 1, 1, 0.30, 0.10, 0.20, 1),
			mkModule("ex/b", 2, 0, 0.00, 0.20, 0.80, 1),
		},
		nil, nil)
	return base, pr
}

// degradeCycleFixtures returns a base/PR pair with identical module metrics but
// a cycle newly introduced in the PR. Expected verdict REQUEST_CHANGES,
// direction degrading, with a new-cycle reason.
func degradeCycleFixtures() (base, pr metrics.ModuleGraph) {
	mods := []metrics.ModuleResult{
		mkModule("ex/a", 1, 1, 0.50, 0.10, 0.40, 1),
		mkModule("ex/b", 2, 0, 0.00, 0.20, 0.80, 1),
	}
	base = mkGraph(metrics.StatusComplete, mods, nil, nil)
	pr = mkGraph(metrics.StatusComplete, mods, []metrics.Cycle{{"ex/a", "ex/b"}}, nil)
	return base, pr
}

// commentBandFixtures returns a base/PR pair where ex/a instability rises by
// 0.05 — a material shift below every REQUEST_CHANGES gate. Expected verdict
// COMMENT, direction stable.
func commentBandFixtures() (base, pr metrics.ModuleGraph) {
	base = mkGraph(metrics.StatusComplete,
		[]metrics.ModuleResult{mkModule("ex/a", 1, 1, 0.50, 0.10, 0.30, 1)}, nil, nil)
	pr = mkGraph(metrics.StatusComplete,
		[]metrics.ModuleResult{mkModule("ex/a", 1, 1, 0.55, 0.10, 0.30, 1)}, nil, nil)
	return base, pr
}

// partialBuildFixtures returns a base/PR pair where the PR is a partial build
// (Status partial plus a load-error warning) and differs structurally (ex/b
// removed, ex/c added). Expected verdict COMMENT, unreliable flag true, with the
// added/removed signal suppressed.
func partialBuildFixtures() (base, pr metrics.ModuleGraph) {
	base = mkGraph(metrics.StatusComplete,
		[]metrics.ModuleResult{
			mkModule("ex/a", 1, 1, 0.50, 0.10, 0.40, 1),
			mkModule("ex/b", 2, 0, 0.00, 0.20, 0.80, 1),
		}, nil, nil)
	pr = mkGraph(metrics.StatusPartial,
		[]metrics.ModuleResult{
			mkModule("ex/a", 1, 1, 0.50, 0.10, 0.40, 1),
			mkModule("ex/c", 0, 1, 1.00, 0.00, 0.00, 1),
		},
		nil,
		[]metrics.Warning{{Code: "load-error", Message: "failed to load ex/c"}})
	return base, pr
}

// structuralOnlyFixtures returns a base/PR pair with no shared modules (ex/a
// removed, ex/b added), no metric regression, and no cycle change. Expected
// verdict APPROVE, direction stable, with non-empty added/removed (reliable
// measurement).
func structuralOnlyFixtures() (base, pr metrics.ModuleGraph) {
	base = mkGraph(metrics.StatusComplete,
		[]metrics.ModuleResult{mkModule("ex/a", 1, 1, 0.50, 0.10, 0.40, 1)}, nil, nil)
	pr = mkGraph(metrics.StatusComplete,
		[]metrics.ModuleResult{mkModule("ex/b", 1, 1, 0.50, 0.10, 0.40, 1)}, nil, nil)
	return base, pr
}

// writeGraphFile marshals g to indented JSON, asserts it is schema-valid via
// metrics.Validate, writes it into dir under name, and returns the path.
func writeGraphFile(t *testing.T, dir, name string, g metrics.ModuleGraph) string {
	t.Helper()
	data, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		t.Fatalf("marshal fixture %s: %v", name, err)
	}
	if err := metrics.Validate(data); err != nil {
		t.Fatalf("fixture %s is not schema-valid: %v", name, err)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write fixture %s: %v", name, err)
	}
	return path
}

// writeGraphPair writes a base/PR fixture pair into a fresh temp dir and returns
// their paths.
func writeGraphPair(t *testing.T, base, pr metrics.ModuleGraph) (basePath, prPath string) {
	t.Helper()
	dir := t.TempDir()
	return writeGraphFile(t, dir, "base.json", base), writeGraphFile(t, dir, "pr.json", pr)
}

// decodedDiff mirrors the --json diff payload for test assertions.
type decodedDiff struct {
	Verdict          string          `json:"verdict"`
	Reasons          []string        `json:"reasons"`
	EntropyDirection string          `json:"entropyDirection"`
	Unreliable       bool            `json:"unreliable"`
	Modules          []metrics.Delta `json:"modules"`
	Added            []string        `json:"added"`
	Removed          []string        `json:"removed"`
	NewCycles        []metrics.Cycle `json:"newCycles"`
	ResolvedCycles   []metrics.Cycle `json:"resolvedCycles"`
}

// approxEqual reports whether two float64 deltas are equal within 1e-9,
// tolerating IEEE-754 subtraction noise in fixture deltas.
func approxEqual(got, want float64) bool {
	return math.Abs(got-want) <= 1e-9
}

// runDiffJSONViaCommand executes the diff subcommand through rootCmd and decodes
// the emitted JSON payload.
func runDiffJSONViaCommand(t *testing.T, args []string) decodedDiff {
	t.Helper()
	cmd := rootCmd()
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute(%v) error: %v\nstderr: %s", args, err, errOut.String())
	}
	var decoded decodedDiff
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("unmarshal diff json: %v\npayload: %s", err, out.String())
	}
	return decoded
}

// --- Task 3.5: JSON payload scenarios ---------------------------------------

func TestRunDiff_JSONScenarios(t *testing.T) {
	t.Parallel()

	type want struct {
		verdict        string
		direction      string
		unreliable     bool
		newCycles      int
		resolvedCycles int
		reasonSubstr   string // empty means expect zero reasons
	}
	tests := []struct {
		name  string
		build func() (metrics.ModuleGraph, metrics.ModuleGraph)
		want  want
	}{
		{
			name:  "improvement_approve",
			build: improvementFixtures,
			want: want{
				verdict: "APPROVE", direction: "improving",
				newCycles: 0, resolvedCycles: 1, reasonSubstr: "",
			},
		},
		{
			name:  "degradation_request_changes",
			build: degradeCycleFixtures,
			want: want{
				verdict: "REQUEST_CHANGES", direction: "degrading",
				newCycles: 1, resolvedCycles: 0, reasonSubstr: "new-cycle",
			},
		},
		{
			name:  "comment_band_stable",
			build: commentBandFixtures,
			want: want{
				verdict: "COMMENT", direction: "stable",
				newCycles: 0, resolvedCycles: 0, reasonSubstr: "materiality",
			},
		},
		{
			name:  "partial_build_unreliable",
			build: partialBuildFixtures,
			want: want{
				verdict: "COMMENT", direction: "stable", unreliable: true,
				newCycles: 0, resolvedCycles: 0, reasonSubstr: "partial-build",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			base, pr := tt.build()
			basePath, prPath := writeGraphPair(t, base, pr)

			var stdout, stderr bytes.Buffer
			result, err := RunDiff(context.Background(), DiffOptions{
				Stdout:     &stdout,
				Stderr:     &stderr,
				BasePath:   basePath,
				PRPath:     prPath,
				Thresholds: metrics.DefaultVerdictThresholds(),
				JSON:       true,
			})
			if err != nil {
				t.Fatalf("RunDiff returned error: %v", err)
			}
			if result.ExitCode != 0 {
				t.Errorf("ExitCode: got %d, want 0", result.ExitCode)
			}

			var out decodedDiff
			if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
				t.Fatalf("unmarshal diff json: %v\npayload: %s", err, stdout.String())
			}

			if out.Verdict != tt.want.verdict {
				t.Errorf("verdict: got %v, want %v", out.Verdict, tt.want.verdict)
			}
			if string(result.Verdict) != tt.want.verdict {
				t.Errorf("result.Verdict: got %v, want %v", result.Verdict, tt.want.verdict)
			}
			if out.EntropyDirection != tt.want.direction {
				t.Errorf("entropyDirection: got %v, want %v", out.EntropyDirection, tt.want.direction)
			}
			if out.Unreliable != tt.want.unreliable {
				t.Errorf("unreliable: got %v, want %v", out.Unreliable, tt.want.unreliable)
			}
			if len(out.NewCycles) != tt.want.newCycles {
				t.Errorf("newCycles count: got %v, want %v", len(out.NewCycles), tt.want.newCycles)
			}
			if len(out.ResolvedCycles) != tt.want.resolvedCycles {
				t.Errorf("resolvedCycles count: got %v, want %v", len(out.ResolvedCycles), tt.want.resolvedCycles)
			}

			if tt.want.reasonSubstr == "" {
				if len(out.Reasons) != 0 {
					t.Errorf("reasons: got %v, want none", out.Reasons)
				}
			} else {
				found := false
				for _, r := range out.Reasons {
					if strings.Contains(r, tt.want.reasonSubstr) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("reasons %v do not contain %q", out.Reasons, tt.want.reasonSubstr)
				}
			}

			// Per-module rows MUST be sorted ascending by Path.
			for i := 1; i < len(out.Modules); i++ {
				if out.Modules[i-1].Path > out.Modules[i].Path {
					t.Errorf("modules not sorted by path: %q before %q", out.Modules[i-1].Path, out.Modules[i].Path)
				}
			}
		})
	}
}

func TestRunDiff_JSONImprovementDeltaValues(t *testing.T) {
	t.Parallel()
	base, pr := improvementFixtures()
	basePath, prPath := writeGraphPair(t, base, pr)

	var stdout, stderr bytes.Buffer
	if _, err := RunDiff(context.Background(), DiffOptions{
		Stdout: &stdout, Stderr: &stderr,
		BasePath: basePath, PRPath: prPath,
		Thresholds: metrics.DefaultVerdictThresholds(), JSON: true,
	}); err != nil {
		t.Fatalf("RunDiff returned error: %v", err)
	}

	var out decodedDiff
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal diff json: %v", err)
	}

	if len(out.Modules) != 2 {
		t.Fatalf("modules count: got %d, want 2", len(out.Modules))
	}
	a := out.Modules[0]
	if a.Path != "ex/a" {
		t.Errorf("modules[0].path: got %q, want %q", a.Path, "ex/a")
	}
	if a.Ca != 0 {
		t.Errorf("ex/a ca delta: got %d, want 0", a.Ca)
	}
	if a.Ce != -1 {
		t.Errorf("ex/a ce delta: got %d, want -1", a.Ce)
	}
	if !approxEqual(a.Instability, -0.20) {
		t.Errorf("ex/a instability delta: got %v, want -0.20", a.Instability)
	}
	if !approxEqual(a.Distance, -0.20) {
		t.Errorf("ex/a distance delta: got %v, want -0.20", a.Distance)
	}
	if a.LCOM != -2 {
		t.Errorf("ex/a lcom delta: got %d, want -2", a.LCOM)
	}

	if len(out.ResolvedCycles) != 1 {
		t.Fatalf("resolvedCycles count: got %d, want 1", len(out.ResolvedCycles))
	}
	if got := strings.Join([]string(out.ResolvedCycles[0]), " "); got != "ex/a ex/b" {
		t.Errorf("resolved cycle members: got %q, want %q", got, "ex/a ex/b")
	}
	if len(out.Reasons) != 0 {
		t.Errorf("reasons: got %v, want none for APPROVE", out.Reasons)
	}
}

func TestRunDiff_JSONDegradationReasons(t *testing.T) {
	t.Parallel()
	base, pr := degradeCycleFixtures()
	basePath, prPath := writeGraphPair(t, base, pr)

	var stdout, stderr bytes.Buffer
	result, err := RunDiff(context.Background(), DiffOptions{
		Stdout: &stdout, Stderr: &stderr,
		BasePath: basePath, PRPath: prPath,
		Thresholds: metrics.DefaultVerdictThresholds(), JSON: true,
	})
	if err != nil {
		t.Fatalf("RunDiff returned error: %v", err)
	}
	if result.Verdict != metrics.VerdictRequestChanges {
		t.Errorf("verdict: got %v, want %v", result.Verdict, metrics.VerdictRequestChanges)
	}

	var out decodedDiff
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal diff json: %v", err)
	}
	if len(out.NewCycles) != 1 {
		t.Fatalf("newCycles count: got %d, want 1", len(out.NewCycles))
	}
	if got := strings.Join([]string(out.NewCycles[0]), " "); got != "ex/a ex/b" {
		t.Errorf("new cycle members: got %q, want %q", got, "ex/a ex/b")
	}
	found := false
	for _, r := range out.Reasons {
		if strings.Contains(r, "new-cycle: ex/a ex/b") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("reasons %v do not contain the new-cycle reason", out.Reasons)
	}
}

func TestRunDiff_PartialBuildSuppressesAddedRemoved(t *testing.T) {
	t.Parallel()
	base, pr := partialBuildFixtures()
	basePath, prPath := writeGraphPair(t, base, pr)

	var stdout, stderr bytes.Buffer
	result, err := RunDiff(context.Background(), DiffOptions{
		Stdout: &stdout, Stderr: &stderr,
		BasePath: basePath, PRPath: prPath,
		Thresholds: metrics.DefaultVerdictThresholds(), JSON: true,
	})
	if err != nil {
		t.Fatalf("RunDiff returned error: %v", err)
	}
	if result.Verdict != metrics.VerdictComment {
		t.Errorf("verdict: got %v, want %v", result.Verdict, metrics.VerdictComment)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(stdout.Bytes(), &raw); err != nil {
		t.Fatalf("unmarshal diff json: %v", err)
	}

	unreliableRaw, ok := raw["unreliable"]
	if !ok {
		t.Fatal("payload missing unreliable key")
	}
	var unreliable bool
	if err := json.Unmarshal(unreliableRaw, &unreliable); err != nil {
		t.Fatalf("unmarshal unreliable: %v", err)
	}
	if !unreliable {
		t.Error("unreliable: got false, want true")
	}
	if _, ok := raw["added"]; ok {
		t.Error("payload must not include added when unreliable")
	}
	if _, ok := raw["removed"]; ok {
		t.Error("payload must not include removed when unreliable")
	}
}

func TestRunDiff_PartialBuildTableAnnotation(t *testing.T) {
	t.Parallel()
	base, pr := partialBuildFixtures()
	basePath, prPath := writeGraphPair(t, base, pr)

	var stdout, stderr bytes.Buffer
	if _, err := RunDiff(context.Background(), DiffOptions{
		Stdout: &stdout, Stderr: &stderr,
		BasePath: basePath, PRPath: prPath,
		Thresholds: metrics.DefaultVerdictThresholds(), JSON: false,
	}); err != nil {
		t.Fatalf("RunDiff returned error: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "partial build") || !strings.Contains(out, "measurement unreliable") {
		t.Errorf("table missing partial-build annotation:\n%s", out)
	}
	if strings.Contains(out, "Added packages:") {
		t.Errorf("table must not include Added section when unreliable:\n%s", out)
	}
	if strings.Contains(out, "Removed packages:") {
		t.Errorf("table must not include Removed section when unreliable:\n%s", out)
	}
	if !strings.Contains(out, "Verdict: COMMENT") {
		t.Errorf("table missing COMMENT verdict:\n%s", out)
	}
}

func TestRunDiff_StructuralAddedRemovedReported(t *testing.T) {
	t.Parallel()
	base, pr := structuralOnlyFixtures()
	basePath, prPath := writeGraphPair(t, base, pr)

	// JSON: added/removed present and correct; verdict APPROVE; empty module set.
	var jsonOut, stderr bytes.Buffer
	if _, err := RunDiff(context.Background(), DiffOptions{
		Stdout: &jsonOut, Stderr: &stderr,
		BasePath: basePath, PRPath: prPath,
		Thresholds: metrics.DefaultVerdictThresholds(), JSON: true,
	}); err != nil {
		t.Fatalf("RunDiff(json) returned error: %v", err)
	}
	var out decodedDiff
	if err := json.Unmarshal(jsonOut.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal diff json: %v", err)
	}
	if out.Verdict != "APPROVE" {
		t.Errorf("verdict: got %v, want APPROVE", out.Verdict)
	}
	if len(out.Modules) != 0 {
		t.Errorf("modules: got %d, want 0 (no shared modules)", len(out.Modules))
	}
	if len(out.Added) != 1 || out.Added[0] != "ex/b" {
		t.Errorf("added: got %v, want [ex/b]", out.Added)
	}
	if len(out.Removed) != 1 || out.Removed[0] != "ex/a" {
		t.Errorf("removed: got %v, want [ex/a]", out.Removed)
	}

	// Table: shows the "(no shared modules)" row and both list sections.
	var tableOut, stderr2 bytes.Buffer
	if _, err := RunDiff(context.Background(), DiffOptions{
		Stdout: &tableOut, Stderr: &stderr2,
		BasePath: basePath, PRPath: prPath,
		Thresholds: metrics.DefaultVerdictThresholds(), JSON: false,
	}); err != nil {
		t.Fatalf("RunDiff(table) returned error: %v", err)
	}
	table := tableOut.String()
	for _, want := range []string{"(no shared modules)", "Added packages:", "ex/b", "Removed packages:", "ex/a"} {
		if !strings.Contains(table, want) {
			t.Errorf("table missing %q:\n%s", want, table)
		}
	}
}

// --- Task 3.5: exit codes ----------------------------------------------------

func TestRunDiff_ExitTwoOnBadInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(t *testing.T) (basePath, prPath string)
		wantSub string
	}{
		{
			name: "missing_base",
			setup: func(t *testing.T) (string, string) {
				base, pr := improvementFixtures()
				_, prPath := writeGraphPair(t, base, pr)
				return filepath.Join(t.TempDir(), "missing-base.json"), prPath
			},
			wantSub: "read base file",
		},
		{
			name: "missing_pr",
			setup: func(t *testing.T) (string, string) {
				base, pr := improvementFixtures()
				basePath, _ := writeGraphPair(t, base, pr)
				return basePath, filepath.Join(t.TempDir(), "missing-pr.json")
			},
			wantSub: "read pr file",
		},
		{
			name: "invalid_base",
			setup: func(t *testing.T) (string, string) {
				dir := t.TempDir()
				_, pr := improvementFixtures()
				prPath := writeGraphFile(t, dir, "pr.json", pr)
				badPath := filepath.Join(dir, "bad-base.json")
				// Missing required fields (modules/cycles/warnings/status).
				if err := os.WriteFile(badPath, []byte(`{"schemaVersion":"1.1","language":"go"}`), 0o644); err != nil {
					t.Fatalf("write bad base: %v", err)
				}
				return badPath, prPath
			},
			wantSub: "validate base file",
		},
		{
			name: "invalid_pr",
			setup: func(t *testing.T) (string, string) {
				dir := t.TempDir()
				base, _ := improvementFixtures()
				basePath := writeGraphFile(t, dir, "base.json", base)
				badPath := filepath.Join(dir, "bad-pr.json")
				if err := os.WriteFile(badPath, []byte(`not valid json`), 0o644); err != nil {
					t.Fatalf("write bad pr: %v", err)
				}
				return basePath, badPath
			},
			wantSub: "validate pr file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			basePath, prPath := tt.setup(t)

			var stdout, stderr bytes.Buffer
			result, err := RunDiff(context.Background(), DiffOptions{
				Stdout: &stdout, Stderr: &stderr,
				BasePath: basePath, PRPath: prPath,
				Thresholds: metrics.DefaultVerdictThresholds(), JSON: true,
			})
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if result.ExitCode != 2 {
				t.Errorf("ExitCode: got %d, want 2", result.ExitCode)
			}
			if stdout.Len() != 0 {
				t.Errorf("stdout must be empty on error (no payload), got %d bytes: %s", stdout.Len(), stdout.String())
			}
			if !strings.Contains(err.Error(), tt.wantSub) {
				t.Errorf("error %q does not contain %q", err.Error(), tt.wantSub)
			}
		})
	}
}

// TestRunDiff_ParseErrorAfterValidation covers the defensive unmarshal branch: a
// document that passes schema validation (ca is a non-negative number) but cannot
// be unmarshaled into the typed graph (ca is not an integer).
func TestRunDiff_ParseErrorAfterValidation(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	badModule := `{"schemaVersion":"1.1","language":"go","modules":[{"path":"ex/a","name":"a","ca":1.5,"ce":0,"instability":0.5,"abstractness":0.1,"distance":0.4,"lcom":1,"exportedTypes":2,"abstractTypes":1,"zone":"normal"}],"cycles":[],"warnings":[],"status":"complete"}`
	if err := metrics.Validate([]byte(badModule)); err != nil {
		t.Fatalf("precondition: base must pass Validate, got: %v", err)
	}
	basePath := filepath.Join(dir, "base.json")
	if err := os.WriteFile(basePath, []byte(badModule), 0o644); err != nil {
		t.Fatalf("write base: %v", err)
	}
	_, pr := improvementFixtures()
	prPath := writeGraphFile(t, dir, "pr.json", pr)

	var stdout, stderr bytes.Buffer
	result, err := RunDiff(context.Background(), DiffOptions{
		Stdout: &stdout, Stderr: &stderr,
		BasePath: basePath, PRPath: prPath,
		Thresholds: metrics.DefaultVerdictThresholds(), JSON: true,
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
	if !strings.Contains(err.Error(), "parse base file") {
		t.Errorf("error %q does not contain 'parse base file'", err.Error())
	}
}

func TestRunDiff_ContextCancelled(t *testing.T) {
	t.Parallel()
	base, pr := improvementFixtures()
	basePath, prPath := writeGraphPair(t, base, pr)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	var stdout, stderr bytes.Buffer
	result, err := RunDiff(ctx, DiffOptions{
		Stdout: &stdout, Stderr: &stderr,
		BasePath: basePath, PRPath: prPath,
		Thresholds: metrics.DefaultVerdictThresholds(), JSON: true,
	})
	if err == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}
	if result.ExitCode != 2 {
		t.Errorf("ExitCode: got %d, want 2", result.ExitCode)
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout must be empty on cancellation, got %d bytes", stdout.Len())
	}
}

// --- Task 3.5: default human table -------------------------------------------

func TestRunDiff_HumanTableDefault(t *testing.T) {
	t.Parallel()
	base, pr := improvementFixtures()
	basePath, prPath := writeGraphPair(t, base, pr)

	var stdout, stderr bytes.Buffer
	result, err := RunDiff(context.Background(), DiffOptions{
		Stdout: &stdout, Stderr: &stderr,
		BasePath: basePath, PRPath: prPath,
		Thresholds: metrics.DefaultVerdictThresholds(), JSON: false,
	})
	if err != nil {
		t.Fatalf("RunDiff returned error: %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("ExitCode: got %d, want 0", result.ExitCode)
	}
	out := stdout.String()
	if strings.HasPrefix(strings.TrimSpace(out), "{") {
		t.Errorf("default output looks like JSON, want a table:\n%s", out)
	}
	for _, want := range []string{
		"Per-module deltas", "Instability", "Distance", "LCOM",
		"ex/a", "Verdict: APPROVE", "Entropy direction: improving",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("table missing %q:\n%s", want, out)
		}
	}
}

// --- Task 3.5: determinism ---------------------------------------------------

func TestRunDiff_Deterministic(t *testing.T) {
	t.Parallel()

	for _, jsonMode := range []bool{true, false} {
		name := "table"
		if jsonMode {
			name = "json"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			base, pr := degradeCycleFixtures()
			basePath, prPath := writeGraphPair(t, base, pr)

			run := func() string {
				var stdout, stderr bytes.Buffer
				if _, err := RunDiff(context.Background(), DiffOptions{
					Stdout: &stdout, Stderr: &stderr,
					BasePath: basePath, PRPath: prPath,
					Thresholds: metrics.DefaultVerdictThresholds(), JSON: jsonMode,
				}); err != nil {
					t.Fatalf("RunDiff returned error: %v", err)
				}
				return stdout.String()
			}
			first := run()
			second := run()
			if first != second {
				t.Errorf("output not deterministic:\nfirst:\n%s\nsecond:\n%s", first, second)
			}
		})
	}
}

// --- Task 3.4: tighten-only override enforcement -----------------------------

func TestTightenThresholds(t *testing.T) {
	t.Parallel()
	def := metrics.DefaultVerdictThresholds()

	tests := []struct {
		name        string
		instability *float64
		distance    *float64
		lcom        *int
		wantErr     bool
		want        metrics.VerdictThresholds
	}{
		{name: "no_overrides", want: def},
		{name: "tighter_instability", instability: float64Ptr(0.05), want: metrics.VerdictThresholds{MaxInstabilityDelta: 0.05, MaxDistanceDelta: 0.20, MaxLCOMDelta: 2}},
		{name: "equal_instability_noop", instability: float64Ptr(0.15), want: def},
		{name: "looser_instability", instability: float64Ptr(0.16), wantErr: true},
		{name: "tighter_distance", distance: float64Ptr(0.10), want: metrics.VerdictThresholds{MaxInstabilityDelta: 0.15, MaxDistanceDelta: 0.10, MaxLCOMDelta: 2}},
		{name: "equal_distance_noop", distance: float64Ptr(0.20), want: def},
		{name: "looser_distance", distance: float64Ptr(0.21), wantErr: true},
		{name: "tighter_lcom", lcom: intPtr(1), want: metrics.VerdictThresholds{MaxInstabilityDelta: 0.15, MaxDistanceDelta: 0.20, MaxLCOMDelta: 1}},
		{name: "equal_lcom_noop", lcom: intPtr(2), want: def},
		{name: "looser_lcom", lcom: intPtr(3), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tightenThresholds(def, tt.instability, tt.distance, tt.lcom)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error for looser override, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("thresholds: got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestDiffCommand_TightenOnlyRejectsLooser(t *testing.T) {
	t.Parallel()
	base, pr := improvementFixtures()
	basePath, prPath := writeGraphPair(t, base, pr)

	tests := []struct {
		name string
		flag string
		val  string
	}{
		{"instability", "--max-instability-delta", "0.30"},
		{"distance", "--max-distance-delta", "0.50"},
		{"lcom", "--max-lcom-delta", "5"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := rootCmd()
			var out, errOut bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetErr(&errOut)
			cmd.SetArgs([]string{"diff", tt.flag, tt.val, basePath, prPath})

			err := cmd.Execute()
			if err == nil {
				t.Fatal("expected error for looser override, got nil")
			}
			var ece *exitCodeError
			if !errors.As(err, &ece) {
				t.Fatalf("error is not *exitCodeError: %T (%v)", err, err)
			}
			if ece.code != 2 {
				t.Errorf("exit code: got %d, want 2", ece.code)
			}
			// No verdict payload may be emitted when a looser override is rejected;
			// because the inputs are valid, an empty stdout proves the tighten check
			// ran BEFORE any file read or verdict computation.
			if out.Len() != 0 {
				t.Errorf("stdout must be empty (no payload), got %d bytes: %s", out.Len(), out.String())
			}
			if !strings.Contains(errOut.String(), "looser than the protected default") {
				t.Errorf("stderr missing tighten diagnostic: %s", errOut.String())
			}
		})
	}
}

func TestDiffCommand_TightenAppliesStricter(t *testing.T) {
	t.Parallel()
	base, pr := commentBandFixtures()
	basePath, prPath := writeGraphPair(t, base, pr)

	// Baseline: default thresholds yield COMMENT for the 0.05 instability shift.
	baseline := runDiffJSONViaCommand(t, []string{"diff", "--json", basePath, prPath})
	if baseline.Verdict != "COMMENT" {
		t.Fatalf("baseline verdict: got %v, want COMMENT", baseline.Verdict)
	}

	// Tightened: --max-instability-delta 0.05 (< 0.15) makes the 0.05 shift fire
	// the inclusive gate, flipping the verdict to REQUEST_CHANGES.
	tightened := runDiffJSONViaCommand(t, []string{"diff", "--json", "--max-instability-delta", "0.05", basePath, prPath})
	if tightened.Verdict != "REQUEST_CHANGES" {
		t.Errorf("tightened verdict: got %v, want REQUEST_CHANGES", tightened.Verdict)
	}
}

// --- Task 3.3/3.4: command-level exit codes ----------------------------------

func TestDiffCommand_ExitCodes(t *testing.T) {
	t.Parallel()
	base, pr := degradeCycleFixtures() // valid inputs; verdict REQUEST_CHANGES
	basePath, prPath := writeGraphPair(t, base, pr)

	t.Run("valid_inputs_exit_zero_even_for_request_changes", func(t *testing.T) {
		t.Parallel()
		cmd := rootCmd()
		var out, errOut bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&errOut)
		cmd.SetArgs([]string{"diff", "--json", basePath, prPath})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute: unexpected error: %v\nstderr: %s", err, errOut.String())
		}
		var decoded decodedDiff
		if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if decoded.Verdict != "REQUEST_CHANGES" {
			t.Errorf("verdict: got %v, want REQUEST_CHANGES (exit still 0)", decoded.Verdict)
		}
	})

	t.Run("missing_file_exit_two", func(t *testing.T) {
		t.Parallel()
		cmd := rootCmd()
		var out, errOut bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&errOut)
		cmd.SetArgs([]string{"diff", filepath.Join(t.TempDir(), "nope.json"), prPath})
		err := cmd.Execute()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var ece *exitCodeError
		if !errors.As(err, &ece) {
			t.Fatalf("error is not *exitCodeError: %T", err)
		}
		if ece.code != 2 {
			t.Errorf("exit code: got %d, want 2", ece.code)
		}
		if out.Len() != 0 {
			t.Errorf("stdout must be empty on tool failure, got: %s", out.String())
		}
	})
}

func TestDiffCommand_Help(t *testing.T) {
	t.Parallel()
	cmd := rootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"diff", "--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := out.String()
	for _, want := range []string{
		"diff", "--json", "--max-instability-delta", "--max-distance-delta", "--max-lcom-delta",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("help output missing %q", want)
		}
	}
}
