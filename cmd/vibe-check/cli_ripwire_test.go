package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/zero-dot-force/vibe-check/internal/goadapter"
	"github.com/zero-dot-force/vibe-check/internal/ripwireadapter"
	"github.com/zero-dot-force/vibe-check/metrics"
)

// CLI ripwire wiring tests — openspec/changes/ripwire-adapter tasks 7.2,
// 7.3, and 7.4. They prove adapter selection through the registry, the
// absent-binary exit path, and that the Go analysis path is byte-stable
// through the registry change without ripwire anywhere on PATH.

// repoRoot returns the repository root computed from this test file's
// location.
func repoRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to determine test file location")
	}
	return filepath.Join(filepath.Dir(filename), "..", "..")
}

// ripwireFixtureDir returns the ripwire-adapter fixture project path for
// a language.
func ripwireFixtureDir(t *testing.T, language string) string {
	t.Helper()
	return filepath.Join(repoRoot(t), "internal", "ripwireadapter", "testdata", language)
}

// prependPath returns a PATH value with dir prepended.
func prependPath(dir string) string {
	return dir + string(os.PathListSeparator) + os.Getenv("PATH")
}

// installPoisonRipwire puts a ripwire binary on PATH that fails loudly.
// Any invocation proves the code under test called ripwire when it must
// not have.
func installPoisonRipwire(t *testing.T) {
	t.Helper()
	bin := t.TempDir()
	script := filepath.Join(bin, "ripwire")
	if err := os.WriteFile(script, []byte("#!/bin/sh\necho 'poison ripwire invoked' >&2\nexit 3\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", prependPath(bin))
}

// installFakeRipwireWrapper puts a wrapper for the committed fake
// ripwire on PATH, baking the recordings directory into the wrapper so
// the production adapter's sanitized environment cannot strip it.
func installFakeRipwireWrapper(t *testing.T, language string) {
	t.Helper()
	bin := t.TempDir()
	recordings := filepath.Join(repoRoot(t), "internal", "ripwireadapter", "testdata", "recorded", language)
	fake := filepath.Join(repoRoot(t), "internal", "ripwireadapter", "testdata", "fake-ripwire", "ripwire")
	wrapper := filepath.Join(bin, "ripwire")
	script := fmt.Sprintf("#!/bin/sh\nFAKE_RIPWIRE_RECORDINGS=%q FAKE_RIPWIRE_VARIANT=valid exec %q \"$@\"\n", recordings, fake)
	if err := os.WriteFile(wrapper, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", prependPath(bin))
}

// TestRunAnalyze_AdapterSelection proves registry selection (task 7.2):
// Go targets select the native adapter — a poison ripwire on PATH never
// runs — while Python and TypeScript targets select the ripwire-backed
// adapters through the fake harness.
func TestRunAnalyze_AdapterSelection(t *testing.T) {
	t.Run("go target selects native adapter without ripwire", func(t *testing.T) {
		installPoisonRipwire(t)

		var stdout, stderr bytes.Buffer
		result, err := RunAnalyze(t.Context(), AnalyzeOptions{
			Stdout: &stdout,
			Stderr: &stderr,
			Path:   couplingFixtureDir(t),
		})
		if err != nil {
			t.Fatalf("RunAnalyze: %v (stderr: %s)", err, stderr.String())
		}
		if result.ExitCode != 0 {
			t.Errorf("exit code: got %d, want 0", result.ExitCode)
		}
		if result.Graph == nil || result.Graph.Language != "go" {
			t.Fatalf("graph language: got %+v, want go", result.Graph)
		}
		if strings.Contains(stderr.String(), "poison ripwire invoked") {
			t.Fatal("Go analysis invoked ripwire")
		}
	})

	t.Run("python target selects ripwire adapter", func(t *testing.T) {
		installFakeRipwireWrapper(t, "python")

		var stdout, stderr bytes.Buffer
		result, err := RunAnalyze(t.Context(), AnalyzeOptions{
			Stdout: &stdout,
			Stderr: &stderr,
			Path:   ripwireFixtureDir(t, "python"),
		})
		if err != nil {
			t.Fatalf("RunAnalyze: %v (stderr: %s)", err, stderr.String())
		}
		if result.Graph == nil || result.Graph.Language != "python" {
			t.Fatalf("graph language: got %+v, want python", result.Graph)
		}
		if len(result.Graph.Modules) != 4 {
			t.Errorf("modules: got %d, want 4 (fixture graph)", len(result.Graph.Modules))
		}
	})

	t.Run("typescript target selects ripwire adapter", func(t *testing.T) {
		installFakeRipwireWrapper(t, "typescript")

		var stdout, stderr bytes.Buffer
		result, err := RunAnalyze(t.Context(), AnalyzeOptions{
			Stdout: &stdout,
			Stderr: &stderr,
			Path:   ripwireFixtureDir(t, "typescript"),
		})
		if err != nil {
			t.Fatalf("RunAnalyze: %v (stderr: %s)", err, stderr.String())
		}
		if result.Graph == nil || result.Graph.Language != "typescript" {
			t.Fatalf("graph language: got %+v, want typescript", result.Graph)
		}
		if len(result.Graph.Modules) != 4 {
			t.Errorf("modules: got %d, want 4 (fixture graph)", len(result.Graph.Modules))
		}
	})
}

// TestRunAnalyze_AbsentBinaryExitsWithGuidance proves the absent-binary
// CLI path (task 7.3): a Python target without ripwire on PATH exits
// with code 2 and an error naming ripwire, its install location, and the
// validated version — and writes no ModuleGraph to stdout.
func TestRunAnalyze_AbsentBinaryExitsWithGuidance(t *testing.T) {
	// A PATH containing nothing: detection is pure filesystem work and
	// the ripwire adapter invokes no other binary.
	t.Setenv("PATH", t.TempDir())

	var stdout, stderr bytes.Buffer
	result, err := RunAnalyze(t.Context(), AnalyzeOptions{
		Stdout: &stdout,
		Stderr: &stderr,
		Path:   ripwireFixtureDir(t, "python"),
	})
	if err == nil {
		t.Fatal("got no error, want absent-binary error")
	}
	if result == nil || result.ExitCode != 2 {
		t.Fatalf("exit code: got %+v, want 2", result)
	}
	if !errors.Is(err, ripwireadapter.ErrRipwireAbsent) {
		t.Errorf("error %v does not wrap ErrRipwireAbsent", err)
	}
	for _, want := range []string{
		"ripwire",
		"https://github.com/redhat-et/ripwire",
		"0.6.9",
		"bedc0cd",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not carry guidance %q", err.Error(), want)
		}
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout: got %d bytes, want no ModuleGraph output", stdout.Len())
	}
}

// TestRunAnalyze_GoPathByteRegression proves the Go path is unchanged by
// the registry wiring (task 7.4): the CLI's output for a Go fixture is
// byte-identical to the native adapter's direct serialization, matches
// the recorded golden, and is produced with a poison ripwire on PATH.
func TestRunAnalyze_GoPathByteRegression(t *testing.T) {
	installPoisonRipwire(t)

	var stdout, stderr bytes.Buffer
	result, err := RunAnalyze(t.Context(), AnalyzeOptions{
		Stdout:       &stdout,
		Stderr:       &stderr,
		Path:         couplingFixtureDir(t),
		NoProvenance: true,
	})
	if err != nil {
		t.Fatalf("RunAnalyze: %v (stderr: %s)", err, stderr.String())
	}
	if result.ExitCode != 0 {
		t.Fatalf("exit code: got %d, want 0", result.ExitCode)
	}
	got := stdout.Bytes()

	// The "before" reference: the native adapter, serialized exactly as
	// the CLI serializes (MarshalIndent + trailing newline). The Go
	// adapter is untouched by the ripwire-adapter change, so this is the
	// pre-change output by construction.
	direct, err := goadapter.New().Analyze(t.Context(), couplingFixtureDir(t))
	if err != nil {
		t.Fatalf("direct goadapter Analyze: %v", err)
	}
	direct.Provenance = nil
	want, err := marshalCLIGraph(direct)
	if err != nil {
		t.Fatalf("marshal direct graph: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Error("registry-path output differs from direct native-adapter output")
	}

	// The recorded golden: stability across branches, not just this run.
	goldenPath := filepath.Join(repoRoot(t), "cmd", "vibe-check", "testdata", "analyze", "go-regression.json")
	if os.Getenv("WRITE_GOLDEN") == "1" {
		if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(goldenPath, got, 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	golden, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden (generate with WRITE_GOLDEN=1): %v", err)
	}
	if !bytes.Equal(got, golden) {
		t.Error("output differs from the recorded golden; if intentional, regenerate with WRITE_GOLDEN=1")
	}
}

// marshalCLIGraph serializes a graph exactly as RunAnalyze writes it:
// two-space MarshalIndent plus a trailing newline.
func marshalCLIGraph(graph *metrics.ModuleGraph) ([]byte, error) {
	data, err := json.MarshalIndent(graph, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}
