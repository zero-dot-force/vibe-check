package goadapter

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/packages"
)

// loadTestPackage loads a single package from the testdata directory for
// unit-level testing of internal functions. It uses the same load flags
// as the production code.
func loadTestPackage(t *testing.T, fixture, pkg string) *packages.Package {
	t.Helper()

	dir := filepath.Join(testdataDir(t), fixture)
	cfg := &packages.Config{
		Mode: loadFlags,
		Dir:  dir,
	}

	pattern := "./" + pkg
	pkgs, err := packages.Load(cfg, pattern)
	if err != nil {
		t.Fatalf("load test package %s/%s: %v", fixture, pkg, err)
	}
	if len(pkgs) == 0 {
		t.Fatalf("load test package %s/%s: no packages loaded", fixture, pkg)
	}

	return pkgs[0]
}

// testdataDir returns the absolute path to the testdata directory.
func testdataDir(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to determine test file location")
	}
	return filepath.Join(filepath.Dir(filename), "testdata")
}

func TestCountTypes_Mixed(t *testing.T) {
	t.Parallel()

	pkg := loadTestPackage(t, "types", "mixed")
	exported, abstract := countTypes(pkg)

	// The mixed fixture declares 7 exported types (Reader, Writer, Point,
	// Config, Pair, Alias, IReader). Only Reader and Writer are abstract.
	// IReader aliases an interface but MUST be counted concrete per spec.
	if exported != 7 {
		t.Errorf("ExportedTypes: got %d, want %d", exported, 7)
	}
	if abstract != 2 {
		t.Errorf("AbstractTypes: got %d, want %d (alias-to-interface must NOT be abstract)", abstract, 2)
	}
}

func TestCountTypes_Empty(t *testing.T) {
	t.Parallel()

	pkg := loadTestPackage(t, "types", "empty")
	exported, abstract := countTypes(pkg)

	if exported != 0 {
		t.Errorf("ExportedTypes: got %d, want %d", exported, 0)
	}
	if abstract != 0 {
		t.Errorf("AbstractTypes: got %d, want %d", abstract, 0)
	}
}
