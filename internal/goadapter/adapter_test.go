package goadapter

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/zero-dot-force/vibe-check/metrics"
)

// fixtureDir returns the absolute path to a test fixture directory.
func fixtureDir(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join(testdataDir(t), name)
}

func TestAdapter_CouplingMetrics(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	adapter := New()
	graph, err := adapter.Analyze(context.Background(), fixtureDir(t, "coupling"))
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}

	// Build lookup by package name for easier assertions.
	byName := make(map[string]metrics.ModuleResult)
	for _, m := range graph.Modules {
		byName[m.Name] = m
	}

	tests := []struct {
		name   string
		wantCa int
		wantCe int
	}{
		{"pkga", 0, 2}, // Ce=2: imports fmt (stdlib) + pkgb (internal)
		{"pkgb", 2, 0}, // Ce=0: no imports
		{"pkgc", 0, 1}, // Ce=1: imports pkgb (internal only)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m, ok := byName[tt.name]
			if !ok {
				t.Fatalf("package %q not found in results", tt.name)
			}
			if m.Ca != tt.wantCa {
				t.Errorf("Ca: got %d, want %d", m.Ca, tt.wantCa)
			}
			if m.Ce != tt.wantCe {
				t.Errorf("Ce: got %d, want %d", m.Ce, tt.wantCe)
			}
		})
	}
}

func TestAdapter_StdlibExclusion(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	adapter := New()
	graph, err := adapter.Analyze(context.Background(), fixtureDir(t, "coupling"))
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}

	// Verify no stdlib packages appear in the module list.
	for _, m := range graph.Modules {
		if m.Path == "fmt" || m.Path == "io" || m.Path == "strings" {
			t.Errorf("stdlib package %q should not appear in modules", m.Path)
		}
	}

	// pkga imports fmt (stdlib) + pkgb (internal).
	// Ce counts all imports including stdlib per Martin's definition.
	for _, m := range graph.Modules {
		if m.Name == "pkga" {
			if m.Ce != 2 {
				t.Errorf("pkga Ce: got %d, want 2 (should count fmt + pkgb)", m.Ce)
			}
		}
	}
}

func TestAdapter_ExternalExclusion(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	adapter := New()
	graph, err := adapter.Analyze(context.Background(), fixtureDir(t, "coupling"))
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}

	// All modules should have paths starting with the module prefix.
	for _, m := range graph.Modules {
		if !strings.HasPrefix(m.Path, "example.com/coupling") {
			t.Errorf("external package %q should not appear in modules", m.Path)
		}
	}
}

func TestAdapter_Integration(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	adapter := New()
	graph, err := adapter.Analyze(context.Background(), fixtureDir(t, "coupling"))
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}

	// Verify schema version.
	if graph.SchemaVersion != "1.1" {
		t.Errorf("SchemaVersion: got %q, want %q", graph.SchemaVersion, "1.1")
	}

	// Verify language.
	if graph.Language != "go" {
		t.Errorf("Language: got %q, want %q", graph.Language, "go")
	}

	// Verify status.
	if graph.Status != metrics.StatusComplete {
		t.Errorf("Status: got %q, want %q", graph.Status, metrics.StatusComplete)
	}

	// Verify non-nil slices.
	if graph.Modules == nil {
		t.Error("Modules is nil, want non-nil")
	}
	if graph.Cycles == nil {
		t.Error("Cycles is nil, want non-nil")
	}
	if graph.Warnings == nil {
		t.Error("Warnings is nil, want non-nil")
	}

	// Verify output passes schema validation via JSON round-trip.
	data, err := json.Marshal(graph)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if err := metrics.Validate(data); err != nil {
		t.Errorf("Validate: %v", err)
	}
}

func TestAdapter_Determinism(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Verify byte-identical JSON output across repeated runs over multiple
	// fixtures. The coupling fixture exercises module ordering; the partial
	// fixture additionally exercises stable warning ordering (it yields
	// warnings for its errored package).
	for _, fixture := range []string{"coupling", "partial"} {
		t.Run(fixture, func(t *testing.T) {
			t.Parallel()
			assertDeterministicAnalyze(t, fixture, 10)
		})
	}
}

// assertDeterministicAnalyze runs Analyze over the named fixture n times and
// fails if any run's marshaled JSON differs from the first run's output.
func assertDeterministicAnalyze(t *testing.T, fixture string, n int) {
	t.Helper()

	adapter := New()
	ctx := context.Background()
	dir := fixtureDir(t, fixture)

	var reference []byte
	for i := 0; i < n; i++ {
		graph, err := adapter.Analyze(ctx, dir)
		if err != nil {
			t.Fatalf("run %d: Analyze: %v", i, err)
		}

		data, err := json.Marshal(graph)
		if err != nil {
			t.Fatalf("run %d: marshal: %v", i, err)
		}

		if i == 0 {
			reference = data
			continue
		}

		if string(data) != string(reference) {
			t.Errorf("run %d: output differs from reference\ngot:  %s\nwant: %s", i, data, reference)
		}
	}
}

func TestAdapter_ContextCancellation(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	adapter := New()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately.

	_, err := adapter.Analyze(ctx, fixtureDir(t, "coupling"))
	if err == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected error wrapping context.Canceled, got: %v", err)
	}
}

func TestAdapter_ContextDeadline(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	adapter := New()
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	_, err := adapter.Analyze(ctx, fixtureDir(t, "coupling"))
	if err == nil {
		t.Fatal("expected error for expired deadline, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected error wrapping context.DeadlineExceeded, got: %v", err)
	}
}

func TestAdapter_PathValidation(t *testing.T) {
	t.Parallel()

	adapter := New()
	ctx := context.Background()

	tests := []struct {
		name string
		path string
	}{
		{"traversal", "/tmp/../etc/passwd"},
		{"non-existent", "/nonexistent/path/that/does/not/exist"},
		{"empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := adapter.Analyze(ctx, tt.path)
			if err == nil {
				t.Errorf("expected error for path %q, got nil", tt.path)
			}
		})
	}
}

func TestAdapter_EmptyDirectory(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	adapter := New()

	// The empty fixture has only .gitkeep, no Go files.
	_, err := adapter.Analyze(context.Background(), fixtureDir(t, "empty"))
	if err == nil {
		t.Fatal("expected error for empty directory, got nil")
	}
}

func TestAdapter_PartialBuild(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	adapter := New()
	graph, err := adapter.Analyze(context.Background(), fixtureDir(t, "partial"))
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}

	// Should have partial status due to the bad package.
	if graph.Status != metrics.StatusPartial {
		t.Errorf("Status: got %q, want %q", graph.Status, metrics.StatusPartial)
	}

	// Should have at least one warning about the bad package.
	if len(graph.Warnings) == 0 {
		t.Fatal("expected warnings for partial build, got none")
	}

	// Verify warning format per spec: Code, Module, and Message fields.
	w := graph.Warnings[0]
	if w.Code != "load-error" {
		t.Errorf("Warning.Code: got %q, want %q", w.Code, "load-error")
	}
	if w.Module == "" {
		t.Error("Warning.Module is empty, want affected package path")
	}
	if w.Message == "" {
		t.Error("Warning.Message is empty, want error description")
	}
	// Warning.Module should contain the partial package path.
	if !strings.Contains(w.Module, "bad") {
		t.Errorf("Warning.Module %q should reference the bad package", w.Module)
	}

	byName := make(map[string]metrics.ModuleResult)
	for _, m := range graph.Modules {
		byName[m.Name] = m
	}

	// The good package should still be in the results.
	if _, ok := byName["good"]; !ok {
		t.Error("good package not found in partial build results")
	}

	// Per the go-adapter partial-build scenario, the errored package MUST appear
	// as a zeroed ModuleResult rather than being silently dropped.
	bad, ok := byName["bad"]
	if !ok {
		t.Fatal(`errored package "bad" not found in results; it must appear as a zeroed ModuleResult`)
	}
	if bad.ExportedTypes != 0 {
		t.Errorf("bad.ExportedTypes: got %d, want 0", bad.ExportedTypes)
	}
	if bad.AbstractTypes != 0 {
		t.Errorf("bad.AbstractTypes: got %d, want 0", bad.AbstractTypes)
	}
	if bad.LCOM != 0 {
		t.Errorf("bad.LCOM: got %d, want 0", bad.LCOM)
	}

	// A warning MUST exist naming the errored package.
	foundWarning := false
	for _, w := range graph.Warnings {
		if w.Module == bad.Path {
			foundWarning = true
			break
		}
	}
	if !foundWarning {
		t.Errorf("no warning found for errored package %q", bad.Path)
	}
}

// TestAdapter_TotalLoadFailure verifies the go-adapter total-load-failure
// scenario: when packages are found but NONE can be type-checked, Analyze must
// return an error rather than a graph of all-zeroed modules. Emitting all-zero
// metrics as if real, when nothing could be analyzed, would violate Metric
// Fidelity. This is distinct from the partial build (TestAdapter_PartialBuild),
// where at least one package type-checks and the graph is returned with
// warnings for the errored packages.
func TestAdapter_TotalLoadFailure(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	adapter := New()
	// The allbad fixture contains only packages that import non-existent
	// packages, so every package fails to type-check.
	graph, err := adapter.Analyze(context.Background(), fixtureDir(t, "allbad"))
	if err == nil {
		t.Fatal("expected error when no package can be type-checked, got nil")
	}
	if !errors.Is(err, errTotalLoadFailure) {
		t.Errorf("expected error wrapping errTotalLoadFailure, got: %v", err)
	}
	// A total load failure MUST NOT yield a graph of all-zeroed modules.
	if graph != nil {
		t.Errorf("expected nil graph on total load failure, got %d modules", len(graph.Modules))
	}
}

func TestAdapter_Language(t *testing.T) {
	t.Parallel()

	adapter := New()
	if got := adapter.Language(); got != "go" {
		t.Errorf("Language: got %q, want %q", got, "go")
	}
}

func TestAdapter_Capabilities(t *testing.T) {
	t.Parallel()

	adapter := New()
	caps := adapter.Capabilities()

	if len(caps) != 7 {
		t.Fatalf("Capabilities count: got %d, want 7", len(caps))
	}

	expected := map[metrics.Capability]bool{
		metrics.CapAfferentCoupling: true,
		metrics.CapEfferentCoupling: true,
		metrics.CapInstability:      true,
		metrics.CapAbstractness:     true,
		metrics.CapDistance:         true,
		metrics.CapLCOM:             true,
		metrics.CapCircularDeps:     true,
	}

	for _, cap := range caps {
		if !expected[cap] {
			t.Errorf("unexpected capability: %q", cap)
		}
	}
}
