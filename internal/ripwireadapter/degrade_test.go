package ripwireadapter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zero-dot-force/vibe-check/metrics"
)

// Unparseable-file degradation tests — openspec/changes/ripwire-adapter
// task 5.3. An unreadable source file must not abort a scan: it produces
// a parse-error warning naming the path, its metrics are zeroed, and the
// any-warning rule forces partial status.

// skipIfRoot skips permission-based tests when running as root (root
// bypasses file modes, so the degradation path is unreachable).
func skipIfRoot(t *testing.T) {
	t.Helper()
	if os.Getuid() == 0 {
		t.Skip("running as root: permission-based degradation unreachable")
	}
}

func TestPythonAbstractnessScan_UnreadableFileDegrades(t *testing.T) {
	skipIfRoot(t)
	t.Parallel()

	root := t.TempDir()
	good := filepath.Join(root, "pkg", "good.py")
	bad := filepath.Join(root, "pkg", "bad.py")
	if err := os.MkdirAll(filepath.Dir(good), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(good, []byte("class A(abc.ABC):\n    pass\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bad, []byte("class B:\n    pass\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(bad, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(bad, 0o644) })

	results, warnings, err := pythonAbstractnessScan(root)
	if err != nil {
		t.Fatalf("pythonAbstractnessScan: %v", err)
	}

	// bad.py contributes zeroed metrics: only good.py's class counts.
	want := map[string]lightParseResult{"pkg": {AbstractTypes: 1, ExportedTypes: 1}}
	if got := results["pkg"]; got != want["pkg"] {
		t.Errorf("pkg: got %+v, want %+v", got, want["pkg"])
	}

	if len(warnings) != 1 {
		t.Fatalf("warnings: got %v, want exactly one", warnings)
	}
	if warnings[0].Code != warningParseError {
		t.Errorf("warning code: got %q, want %q", warnings[0].Code, warningParseError)
	}
	if !strings.Contains(warnings[0].Message, "pkg/bad.py") {
		t.Errorf("warning message %q does not name pkg/bad.py", warnings[0].Message)
	}
	if got := graphStatusForWarnings(warnings); got != metrics.StatusPartial {
		t.Errorf("status: got %q, want partial", got)
	}
}

func TestLCOM4Scan_UnreadableFileDegrades(t *testing.T) {
	skipIfRoot(t)
	t.Parallel()

	root := t.TempDir()
	good := filepath.Join(root, "pkg", "good.py")
	bad := filepath.Join(root, "pkg", "bad.py")
	if err := os.MkdirAll(filepath.Dir(good), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(good, []byte("class A:\n    def m(self):\n        return self.f\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bad, []byte("class B:\n    def m(self):\n        return self.g\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(bad, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(bad, 0o644) })

	results, warnings, err := lcom4Scan(root, languagePython)
	if err != nil {
		t.Fatalf("lcom4Scan: %v", err)
	}

	if got := results["pkg"]; got != 1 {
		t.Errorf("pkg LCOM: got %d, want 1 (bad.py zeroed)", got)
	}
	if len(warnings) != 1 || warnings[0].Code != warningParseError {
		t.Fatalf("warnings: got %v, want one parse-error", warnings)
	}
	if !strings.Contains(warnings[0].Message, "pkg/bad.py") {
		t.Errorf("warning message %q does not name pkg/bad.py", warnings[0].Message)
	}
	if got := graphStatusForWarnings(warnings); got != metrics.StatusPartial {
		t.Errorf("status: got %q, want partial", got)
	}
}
