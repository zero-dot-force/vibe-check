package main

import (
	"bytes"
	"context"
	"encoding/json"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/zero-dot-force/vibe-check/internal/scaffold"
	"github.com/zero-dot-force/vibe-check/metrics"
)

// --- Task 5.1: shared helper -------------------------------------------------

func TestNewProvenanceEnvelope(t *testing.T) {
	t.Parallel()

	env := newProvenanceEnvelope()
	if env == nil {
		t.Fatal("newProvenanceEnvelope returned nil")
	}
	if env.Producer != producerName {
		t.Errorf("Producer: got %q, want %q", env.Producer, producerName)
	}
	if env.Version == "" {
		t.Error("Version is empty, want non-empty")
	}
	if _, err := time.Parse(time.RFC3339, env.GeneratedAt); err != nil {
		t.Errorf("GeneratedAt is not RFC3339: %v", err)
	}
	if !strings.HasSuffix(env.GeneratedAt, "Z") {
		t.Errorf("GeneratedAt is not UTC (want Z suffix): %q", env.GeneratedAt)
	}
}

// --- Task 5.2: diff JSON provenance ------------------------------------------

// diffProvenancePayload mirrors the --json diff payload including the
// provenance envelope for assertion purposes.
type diffProvenancePayload struct {
	Provenance       *provenanceEnvelope `json:"provenance"`
	Verdict          string              `json:"verdict"`
	Reasons          []string            `json:"reasons"`
	EntropyDirection string              `json:"entropyDirection"`
	Unreliable       bool                `json:"unreliable"`
	Modules          []metrics.Delta     `json:"modules"`
}

func TestRunDiff_JSONProvenance(t *testing.T) {
	t.Parallel()

	base, pr := improvementFixtures()
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
		t.Fatalf("ExitCode: got %d, want 0", result.ExitCode)
	}

	var out diffProvenancePayload
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal diff json: %v", err)
	}
	if out.Provenance == nil {
		t.Fatal("provenance key missing, want present by default")
	}
	if out.Provenance.Producer != producerName {
		t.Errorf("provenance.producer: got %q, want %q", out.Provenance.Producer, producerName)
	}
	if out.Provenance.Version == "" {
		t.Error("provenance.version is empty, want non-empty")
	}
	if _, err := time.Parse(time.RFC3339, out.Provenance.GeneratedAt); err != nil {
		t.Errorf("provenance.generatedAt is not RFC3339: %v", err)
	}

	// Existing result keys are preserved.
	if out.Verdict != "APPROVE" {
		t.Errorf("verdict: got %q, want APPROVE", out.Verdict)
	}
	if out.EntropyDirection != "improving" {
		t.Errorf("entropyDirection: got %q, want improving", out.EntropyDirection)
	}
	if len(out.Modules) != 2 {
		t.Errorf("modules count: got %d, want 2", len(out.Modules))
	}
}

func TestRunDiff_JSONNoProvenance(t *testing.T) {
	t.Parallel()

	base, pr := improvementFixtures()
	basePath, prPath := writeGraphPair(t, base, pr)

	var stdout, stderr bytes.Buffer
	if _, err := RunDiff(context.Background(), DiffOptions{
		Stdout:       &stdout,
		Stderr:       &stderr,
		BasePath:     basePath,
		PRPath:       prPath,
		Thresholds:   metrics.DefaultVerdictThresholds(),
		JSON:         true,
		NoProvenance: true,
	}); err != nil {
		t.Fatalf("RunDiff returned error: %v", err)
	}
	if strings.Contains(stdout.String(), `"provenance"`) {
		t.Error("output contains provenance key, want omitted with --no-provenance")
	}
}

// --- Task 5.5: cross-flag equivalence ----------------------------------------

func TestRunDiff_CrossFlagEquivalence(t *testing.T) {
	t.Parallel()

	base, pr := degradeCycleFixtures()
	basePath, prPath := writeGraphPair(t, base, pr)

	run := func(noProvenance bool) (decodedDiff, *provenanceEnvelope) {
		t.Helper()
		var stdout, stderr bytes.Buffer
		if _, err := RunDiff(context.Background(), DiffOptions{
			Stdout:       &stdout,
			Stderr:       &stderr,
			BasePath:     basePath,
			PRPath:       prPath,
			Thresholds:   metrics.DefaultVerdictThresholds(),
			JSON:         true,
			NoProvenance: noProvenance,
		}); err != nil {
			t.Fatalf("RunDiff returned error: %v", err)
		}
		var d decodedDiff
		if err := json.Unmarshal(stdout.Bytes(), &d); err != nil {
			t.Fatalf("unmarshal diff json: %v", err)
		}
		var raw struct {
			Provenance *provenanceEnvelope `json:"provenance"`
		}
		_ = json.Unmarshal(stdout.Bytes(), &raw)
		return d, raw.Provenance
	}

	withProv, _ := run(false)
	noProv, noProvEnvelope := run(true)

	// Verdict, reasons, entropy direction, and modules MUST be identical.
	if withProv.Verdict != noProv.Verdict {
		t.Errorf("verdict differs: %q vs %q", withProv.Verdict, noProv.Verdict)
	}
	if !slices.Equal(withProv.Reasons, noProv.Reasons) {
		t.Errorf("reasons differ: %v vs %v", withProv.Reasons, noProv.Reasons)
	}
	if withProv.EntropyDirection != noProv.EntropyDirection {
		t.Errorf("entropyDirection differs: %q vs %q", withProv.EntropyDirection, noProv.EntropyDirection)
	}
	if withProv.Unreliable != noProv.Unreliable {
		t.Errorf("unreliable differs: %v vs %v", withProv.Unreliable, noProv.Unreliable)
	}
	if len(withProv.Modules) != len(noProv.Modules) {
		t.Fatalf("modules count differs: %d vs %d", len(withProv.Modules), len(noProv.Modules))
	}
	for i := range withProv.Modules {
		if withProv.Modules[i] != noProv.Modules[i] {
			t.Errorf("modules[%d] differs: %+v vs %+v", i, withProv.Modules[i], noProv.Modules[i])
		}
	}
	if noProvEnvelope != nil {
		t.Error("--no-provenance output still carries a provenance envelope")
	}
}

// --- Task 5.4: determinism ---------------------------------------------------

func TestRunDiff_NoProvenanceDeterministic(t *testing.T) {
	t.Parallel()

	base, pr := degradeCycleFixtures()
	basePath, prPath := writeGraphPair(t, base, pr)

	run := func() string {
		t.Helper()
		var stdout, stderr bytes.Buffer
		if _, err := RunDiff(context.Background(), DiffOptions{
			Stdout:       &stdout,
			Stderr:       &stderr,
			BasePath:     basePath,
			PRPath:       prPath,
			Thresholds:   metrics.DefaultVerdictThresholds(),
			JSON:         true,
			NoProvenance: true,
		}); err != nil {
			t.Fatalf("RunDiff returned error: %v", err)
		}
		return stdout.String()
	}

	if first, second := run(), run(); first != second {
		t.Errorf("output not byte-identical across runs with --no-provenance:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}

func TestRunInit_NoProvenanceDeterministic(t *testing.T) {
	t.Parallel()

	res := &scaffold.Result{Written: []string{"agents/a.md"}, Skipped: []string{}, Forced: []string{}}

	run := func() string {
		t.Helper()
		var buf bytes.Buffer
		if err := writeInitJSON(&buf, res, true); err != nil {
			t.Fatalf("writeInitJSON returned error: %v", err)
		}
		return buf.String()
	}

	if first, second := run(), run(); first != second {
		t.Errorf("output not byte-identical across runs with --no-provenance:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}

// --- Task 5.3: init JSON provenance ------------------------------------------

type initProvenancePayload struct {
	Provenance *provenanceEnvelope `json:"provenance"`
	Written    []string            `json:"written"`
	Skipped    []string            `json:"skipped"`
	Forced     []string            `json:"forced"`
}

func TestRunInit_JSONProvenance(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	res, err := RunInit(context.Background(), InitOptions{
		Stdout: &stdout,
		Stderr: &stderr,
		Path:   t.TempDir(),
		JSON:   true,
	})
	if err != nil {
		t.Fatalf("RunInit returned error: %v", err)
	}
	if res.ExitCode != 0 {
		t.Fatalf("ExitCode: got %d, want 0", res.ExitCode)
	}

	var out initProvenancePayload
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal init json: %v", err)
	}
	if out.Provenance == nil {
		t.Fatal("provenance key missing, want present by default")
	}
	if out.Provenance.Producer != producerName {
		t.Errorf("provenance.producer: got %q, want %q", out.Provenance.Producer, producerName)
	}
	if out.Provenance.Version == "" {
		t.Error("provenance.version is empty, want non-empty")
	}
	if _, err := time.Parse(time.RFC3339, out.Provenance.GeneratedAt); err != nil {
		t.Errorf("provenance.generatedAt is not RFC3339: %v", err)
	}
}

func TestRunInit_JSONNoProvenance(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	if _, err := RunInit(context.Background(), InitOptions{
		Stdout:       &stdout,
		Stderr:       &stderr,
		Path:         t.TempDir(),
		JSON:         true,
		NoProvenance: true,
	}); err != nil {
		t.Fatalf("RunInit returned error: %v", err)
	}
	if strings.Contains(stdout.String(), `"provenance"`) {
		t.Error("output contains provenance key, want omitted with --no-provenance")
	}
}

// --- Command-level flag wiring ------------------------------------------------

func TestDiffCommand_NoProvenanceFlag(t *testing.T) {
	t.Parallel()

	base, pr := improvementFixtures()
	basePath, prPath := writeGraphPair(t, base, pr)

	cmd := rootCmd()
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs([]string{"diff", "--json", "--no-provenance", basePath, prPath})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v\nstderr: %s", err, errOut.String())
	}
	if strings.Contains(out.String(), `"provenance"`) {
		t.Error("command output contains provenance key, want omitted with --no-provenance flag")
	}
}

func TestInitCmd_NoProvenanceFlag(t *testing.T) {
	t.Parallel()

	cmd := initCmd()
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs([]string{"--json", "--no-provenance", t.TempDir()})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v\nstderr: %s", err, errOut.String())
	}
	if strings.Contains(out.String(), `"provenance"`) {
		t.Error("command output contains provenance key, want omitted with --no-provenance flag")
	}
}
