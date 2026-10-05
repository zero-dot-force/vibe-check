package goadapter

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"

	"github.com/zero-dot-force/vibe-check/metrics"
)

// detectDuplications finds structurally similar function bodies within each
// package by normalizing ASTs (replacing identifiers and literals with
// canonical placeholders) and pairwise-comparing the resulting canonical
// strings. Only functions with at least 6 significant lines are considered.
// Generated and excluded files (*.gen.go, *.pb.go, mock/, generated/, mocks/,
// testdata/) are skipped.
//
// Each Duplication groups two or more blocks that are structurally identical
// after normalization (Similarity = 1.0). For a 3-way match, two Duplication
// entries are produced (one per pair).
func detectDuplications(pkgs []*packages.Package) []metrics.Duplication {
	var duplications []metrics.Duplication

	for _, pkg := range pkgs {
		dups := detectPackageDuplications(pkg)
		duplications = append(duplications, dups...)
	}

	return duplications
}

// funcInfo holds the pre-computed data for a single function during
// duplication detection.
type funcInfo struct {
	fn       *ast.FuncDecl
	file     string
	fset     *token.FileSet
	normBody string
	sigLines int
}

// detectPackageDuplications finds duplicate function bodies within a single
// package.
func detectPackageDuplications(pkg *packages.Package) []metrics.Duplication {
	var funcs []funcInfo

	for _, file := range pkg.Syntax {
		filename := pkg.Fset.Position(file.Pos()).Filename
		if isExcludedFile(filename) {
			continue
		}

		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}

			sigLines := countSignificantLines(fn, pkg.Fset)
			if sigLines < 6 {
				continue
			}

			normBody := normalizeFuncBody(fn, pkg.Fset)
			if normBody == "" {
				continue
			}

			funcs = append(funcs, funcInfo{
				fn:       fn,
				file:     filename,
				fset:     pkg.Fset,
				normBody: normBody,
				sigLines: sigLines,
			})
		}
	}

	// Pairwise comparison: for each pair of functions with identical
	// normalized bodies, produce a Duplication entry.
	var duplications []metrics.Duplication
	for i := 0; i < len(funcs); i++ {
		for j := i + 1; j < len(funcs); j++ {
			if funcs[i].normBody == funcs[j].normBody {
				duplications = append(duplications, metrics.Duplication{
					ModulePath: pkg.PkgPath,
					Blocks: []metrics.DuplicateBlock{
						makeBlock(funcs[i]),
						makeBlock(funcs[j]),
					},
					Similarity: 1.0,
				})
			}
		}
	}

	return duplications
}

// normalizeFuncBody produces a canonical string representation of a function
// body suitable for structural comparison. It normalizes the AST in-place by
// replacing identifier names with "_id" and basic literal values with "_lit",
// strips comments, then formats the body with go/format.Node.
//
// The normalization is destructive to the AST — callers that need the
// original AST should copy it first. In practice, the goadapter does not
// use the AST after duplication detection, so in-place modification is safe.
func normalizeFuncBody(fn *ast.FuncDecl, fset *token.FileSet) string {
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.Ident:
			n.Name = "_id"
		case *ast.BasicLit:
			n.Value = "_lit"
		}
		return true
	})

	stripComments(fn.Body)

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, fn.Body); err != nil {
		return ""
	}
	return buf.String()
}

// stripComments removes all comment groups from an AST subtree by setting
// Doc, Comment, and Comments fields to nil on all node types that carry them.
// This is done in-place via ast.Inspect.
func stripComments(node ast.Node) {
	ast.Inspect(node, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.File:
			n.Comments = nil
		case *ast.FuncDecl:
			n.Doc = nil
		case *ast.GenDecl:
			n.Doc = nil
		case *ast.Field:
			n.Doc = nil
			n.Comment = nil
		case *ast.ValueSpec:
			n.Doc = nil
			n.Comment = nil
		case *ast.TypeSpec:
			n.Doc = nil
			n.Comment = nil
		case *ast.ImportSpec:
			n.Doc = nil
			n.Comment = nil
		}
		return true
	})
}

// countSignificantLines returns the number of non-blank, non-single-brace
// lines in a function body. It reads the source file from disk to count
// lines. Returns 0 if the file cannot be read.
func countSignificantLines(fn *ast.FuncDecl, fset *token.FileSet) int {
	filename := fset.Position(fn.Pos()).Filename
	startLine := fset.Position(fn.Pos()).Line
	endLine := fset.Position(fn.End()).Line

	src, err := os.ReadFile(filename)
	if err != nil {
		return 0
	}

	lines := strings.Split(string(src), "\n")
	count := 0
	for i := startLine - 1; i < endLine && i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" || trimmed == "{" || trimmed == "}" {
			continue
		}
		count++
	}
	return count
}

// isExcludedFile reports whether a file should be skipped during duplication
// detection. Files matching *.gen.go or *.pb.go are excluded, as are files
// in directories named mock, generated, mocks, or testdata (checked against
// every component of the file path).
func isExcludedFile(filename string) bool {
	base := filepath.Base(filename)

	// Exclude generated protobuf and code-gen files.
	if strings.HasSuffix(base, ".gen.go") || strings.HasSuffix(base, ".pb.go") {
		return true
	}

	// Exclude files in mock/generated/mocks/testdata directories.
	// Check every directory component of the path.
	excludedDirs := map[string]bool{
		"mock":      true,
		"generated": true,
		"mocks":     true,
		"testdata":  true,
	}

	dir := filepath.Dir(filename)
	for {
		component := filepath.Base(dir)
		if component == "." || component == "" {
			break
		}
		if excludedDirs[component] {
			return true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return false
}

// makeBlock creates a DuplicateBlock from a funcInfo, capturing the file
// path, line range, and significant line count.
func makeBlock(fi funcInfo) metrics.DuplicateBlock {
	startLine := fi.fset.Position(fi.fn.Pos()).Line
	endLine := fi.fset.Position(fi.fn.End()).Line
	return metrics.DuplicateBlock{
		File:      fi.file,
		StartLine: startLine,
		EndLine:   endLine,
		LineCount: fi.sigLines,
	}
}
