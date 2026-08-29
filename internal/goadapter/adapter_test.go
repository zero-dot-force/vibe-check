package goadapter

import (
	"context"
	"encoding/json"
	"path/filepath"
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
		{"pkga", 0, 1},
		{"pkgb", 2, 0},
		{"pkgc", 0, 1},
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
	// Ce should count pkgb only (Ce=1), not fmt.
	for _, m := range graph.Modules {
		if m.Name == "pkga" {
			if m.Ce != 1 {
				t.Errorf("pkga Ce: got %d, want 1 (should exclude stdlib)", m.Ce)
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
		if m.Path != "" && m.Path[:len("example.com/coupling")] != "example.com/coupling" {
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

	adapter := New()
	ctx := context.Background()
	dir := fixtureDir(t, "coupling")

	// Run analysis 10 times and verify byte-identical JSON output.
	var reference []byte
	for i := 0; i < 10; i++ {
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

	adapter := New()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately.

	_, err := adapter.Analyze(ctx, fixtureDir(t, "coupling"))
	if err == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}
}

func TestAdapter_ContextDeadline(t *testing.T) {
	t.Parallel()

	adapter := New()
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	_, err := adapter.Analyze(ctx, fixtureDir(t, "coupling"))
	if err == nil {
		t.Fatal("expected error for expired deadline, got nil")
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
		t.Error("expected warnings for partial build, got none")
	}

	// The good package should still be in the results.
	found := false
	for _, m := range graph.Modules {
		if m.Name == "good" {
			found = true
			break
		}
	}
	if !found {
		t.Error("good package not found in partial build results")
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
