package goadapter

import (
	"go/ast"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"
)

// dupModule creates a temp directory with a Go module containing the given
// source files for duplication detection testing. Files are written outside
// of testdata/ so isExcludedFile does not skip them.
func dupModule(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()

	goMod := "module example.com/duptest\n\ngo 1.25\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	return dir
}

// loadDupPackage loads the package from a temp module directory.
func loadDupPackage(t *testing.T, dir string) *packages.Package {
	t.Helper()
	cfg := &packages.Config{
		Mode: loadFlags,
		Dir:  dir,
	}
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		t.Fatalf("load package: %v", err)
	}
	if len(pkgs) == 0 {
		t.Fatal("no packages loaded")
	}
	return pkgs[0]
}

// identicalSrc returns source for two structurally identical functions
// that differ only in identifier names.
func identicalSrc() string {
	return `package p

func ProcessItems(items []string) []string {
	var result []string
	for _, item := range items {
		if len(item) > 0 {
			result = append(result, item)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func FilterValues(values []string) []string {
	var output []string
	for _, val := range values {
		if len(val) > 0 {
			output = append(output, val)
		}
	}
	if len(output) == 0 {
		return nil
	}
	return output
}

func SelectEntries(entries []string) []string {
	var selected []string
	for _, entry := range entries {
		if len(entry) > 0 {
			selected = append(selected, entry)
		}
	}
	if len(selected) == 0 {
		return nil
	}
	return selected
}

func ShortFunc(x int) int {
	return x * 2
}
`
}

// differentSrc returns source for two structurally different functions.
func differentSrc() string {
	return `package p

func ComputeSum(nums []int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

func FindMax(nums []int) int {
	if len(nums) == 0 {
		return 0
	}
	max := nums[0]
	for _, n := range nums {
		if n > max {
			max = n
		}
	}
	return max
}
`
}

// TestDetectDuplications_IdenticalStructure verifies that functions with
// structurally identical bodies (differing only in identifier names) are
// detected as duplicates with Similarity == 1.0 and populated block metadata.
func TestDetectDuplications_IdenticalStructure(t *testing.T) {
	t.Parallel()

	dir := dupModule(t, map[string]string{
		"main.go": identicalSrc(),
	})
	pkg := loadDupPackage(t, dir)
	dups := detectDuplications([]*packages.Package{pkg})

	if len(dups) == 0 {
		t.Fatal("expected at least one duplication, got none")
	}

	for _, d := range dups {
		if d.Similarity != 1.0 {
			t.Errorf("Similarity: got %.2f, want 1.0", d.Similarity)
		}
		if d.ModulePath == "" {
			t.Error("ModulePath is empty")
		}
		if len(d.Blocks) != 2 {
			t.Errorf("Blocks count: got %d, want 2", len(d.Blocks))
		}
		for _, b := range d.Blocks {
			if b.File == "" {
				t.Error("DuplicateBlock.File is empty")
			}
			if b.StartLine < 1 {
				t.Errorf("DuplicateBlock.StartLine: got %d, want >= 1", b.StartLine)
			}
			if b.EndLine < b.StartLine {
				t.Errorf("DuplicateBlock.EndLine (%d) < StartLine (%d)", b.EndLine, b.StartLine)
			}
			if b.LineCount < 6 {
				t.Errorf("DuplicateBlock.LineCount: got %d, want >= 6", b.LineCount)
			}
		}
	}
}

// TestDetectDuplications_DifferentStructure verifies that functions with
// structurally different bodies are NOT reported as duplicates.
func TestDetectDuplications_DifferentStructure(t *testing.T) {
	t.Parallel()

	dir := dupModule(t, map[string]string{
		"main.go": differentSrc(),
	})
	pkg := loadDupPackage(t, dir)
	dups := detectDuplications([]*packages.Package{pkg})

	if len(dups) != 0 {
		t.Errorf("expected 0 duplications for structurally different functions, got %d", len(dups))
	}
}

// TestDetectDuplications_SkipShortFunctions verifies that functions with
// fewer than 6 significant lines are excluded from duplication detection.
func TestDetectDuplications_SkipShortFunctions(t *testing.T) {
	t.Parallel()

	dir := dupModule(t, map[string]string{
		"main.go": identicalSrc(),
	})
	pkg := loadDupPackage(t, dir)
	dups := detectDuplications([]*packages.Package{pkg})

	// ShortFunc has only 1 significant line and should never appear in results.
	for _, d := range dups {
		for _, b := range d.Blocks {
			if b.LineCount < 6 {
				t.Errorf("block with LineCount %d should not appear in duplications (minimum is 6)", b.LineCount)
			}
		}
	}
}

// TestDetectDuplications_SkipGenerated verifies that *.gen.go files are
// excluded from duplication detection.
func TestDetectDuplications_SkipGenerated(t *testing.T) {
	t.Parallel()

	// Create a module with a normal file and a .gen.go file, both containing
	// structurally identical functions.
	dir := dupModule(t, map[string]string{
		"main.go": `package p

func ProcessItems(items []string) []string {
	var result []string
	for _, item := range items {
		if len(item) > 0 {
			result = append(result, item)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}
`,
		"generated.gen.go": `// Code generated by tool. DO NOT EDIT.

package p

func GeneratedFilter(items []string) []string {
	var filtered []string
	for _, item := range items {
		if len(item) > 0 {
			filtered = append(filtered, item)
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	return filtered
}
`,
	})
	pkg := loadDupPackage(t, dir)
	dups := detectDuplications([]*packages.Package{pkg})

	for _, d := range dups {
		for _, b := range d.Blocks {
			if strings.Contains(b.File, ".gen.go") {
				t.Errorf("block from generated file %q should be excluded", b.File)
			}
		}
	}
}

// TestDetectDuplications_NoDuplicates verifies that a package with no
// structural duplicates produces zero duplications.
func TestDetectDuplications_NoDuplicates(t *testing.T) {
	t.Parallel()

	dir := dupModule(t, map[string]string{
		"main.go": differentSrc(),
	})
	pkg := loadDupPackage(t, dir)
	dups := detectDuplications([]*packages.Package{pkg})

	if len(dups) != 0 {
		t.Errorf("expected 0 duplications, got %d", len(dups))
	}
}

// TestDetectDuplications_MultipleMatches verifies that when three functions
// share identical structure, all pairwise combinations produce Duplication
// entries (3 entries for 3 functions: A-B, A-C, B-C).
func TestDetectDuplications_MultipleMatches(t *testing.T) {
	t.Parallel()

	dir := dupModule(t, map[string]string{
		"main.go": identicalSrc(),
	})
	pkg := loadDupPackage(t, dir)
	dups := detectDuplications([]*packages.Package{pkg})

	// With 3 identical functions (ProcessItems, FilterValues, SelectEntries),
	// we expect 3 pairwise matches.
	if len(dups) != 3 {
		t.Errorf("pairwise duplicate count for 3 identical functions: got %d, want 3", len(dups))
	}
}

// TestDetectDuplications_ModulePath verifies that the ModulePath field is
// correctly populated from the package path.
func TestDetectDuplications_ModulePath(t *testing.T) {
	t.Parallel()

	dir := dupModule(t, map[string]string{
		"main.go": identicalSrc(),
	})
	pkg := loadDupPackage(t, dir)
	dups := detectDuplications([]*packages.Package{pkg})

	for _, d := range dups {
		if d.ModulePath != "example.com/duptest" {
			t.Errorf("ModulePath: got %q, want %q", d.ModulePath, "example.com/duptest")
		}
	}
}

// TestIsExcludedFile verifies the file exclusion logic for various path
// patterns including *.gen.go, *.pb.go, and excluded directory names.
func TestIsExcludedFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{"normal_go_file", "/path/to/pkg/file.go", false},
		{"gen_go_file", "/path/to/pkg/file.gen.go", true},
		{"pb_go_file", "/path/to/pkg/file.pb.go", true},
		{"gen_go_no_path", "file.gen.go", true},
		{"mock_directory", "/path/to/mock/file.go", true},
		{"generated_directory", "/path/to/generated/file.go", true},
		{"mocks_directory", "/path/to/mocks/file.go", true},
		{"testdata_directory", "/path/to/testdata/file.go", true},
		{"deep_mock", "/a/b/mock/c/d/file.go", true},
		{"deep_testdata", "/a/b/testdata/c/d/file.go", true},
		{"not_excluded_dir", "/path/to/pkg/file.go", false},
		{"mock_in_filename", "/path/to/mockery.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := isExcludedFile(tt.filename)
			if got != tt.want {
				t.Errorf("isExcludedFile(%q): got %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

// TestCountSignificantLines verifies line counting for various function
// bodies including edge cases.
func TestCountSignificantLines(t *testing.T) {
	t.Parallel()

	dir := dupModule(t, map[string]string{
		"main.go": identicalSrc(),
	})
	pkg := loadDupPackage(t, dir)

	// Find ShortFunc and verify its significant line count.
	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			if fn.Name.Name == "ShortFunc" {
				count := countSignificantLines(fn, pkg.Fset)
				if count >= 6 {
					t.Errorf("ShortFunc significant lines: got %d, want < 6", count)
				}
				return
			}
		}
	}
	t.Fatal("ShortFunc not found in package")
}

// TestNormalizeFuncBody_Deterministic verifies that normalizeFuncBody
// produces the same output for structurally identical functions.
func TestNormalizeFuncBody_Deterministic(t *testing.T) {
	t.Parallel()

	dir := dupModule(t, map[string]string{
		"main.go": identicalSrc(),
	})
	pkg := loadDupPackage(t, dir)

	var processBody, filterBody string
	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			switch fn.Name.Name {
			case "ProcessItems":
				processBody = normalizeFuncBody(fn, pkg.Fset)
			case "FilterValues":
				filterBody = normalizeFuncBody(fn, pkg.Fset)
			}
		}
	}

	if processBody == "" {
		t.Fatal("ProcessItems normalized body is empty")
	}
	if filterBody == "" {
		t.Fatal("FilterValues normalized body is empty")
	}
	if processBody != filterBody {
		t.Errorf("normalized bodies differ for structurally identical functions:\nProcessItems: %s\nFilterValues: %s", processBody, filterBody)
	}
}

// TestNormalizeFuncBody_Different verifies that normalizeFuncBody produces
// different output for structurally different functions.
func TestNormalizeFuncBody_Different(t *testing.T) {
	t.Parallel()

	dir := dupModule(t, map[string]string{
		"main.go": differentSrc(),
	})
	pkg := loadDupPackage(t, dir)

	var sumBody, maxBody string
	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			switch fn.Name.Name {
			case "ComputeSum":
				sumBody = normalizeFuncBody(fn, pkg.Fset)
			case "FindMax":
				maxBody = normalizeFuncBody(fn, pkg.Fset)
			}
		}
	}

	if sumBody == "" {
		t.Fatal("ComputeSum normalized body is empty")
	}
	if maxBody == "" {
		t.Fatal("FindMax normalized body is empty")
	}
	if sumBody == maxBody {
		t.Error("normalized bodies should differ for structurally different functions")
	}
}
