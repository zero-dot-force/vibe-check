package goadapter

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/packages"

	"github.com/zero-dot-force/vibe-check/metrics"
)

// computeLCOM4 computes the LCOM4 metric for a package using the
// Hitz & Montazeri (1995) connected-component variant.
//
// Algorithm:
//  1. Identify exported methods (functions with a receiver of an exported type).
//  2. For each method, walk the AST body to find struct field accesses (s.field).
//  3. Build a graph where methods are nodes and edges connect methods that
//     access at least one common struct field.
//  4. Count connected components using union-find.
//
// Returns 0 if no exported methods exist (trivially cohesive).
// Package-level functions (no receiver) are excluded.
func computeLCOM4(pkg *packages.Package) metrics.LCOM {
	if pkg.TypesInfo == nil || len(pkg.Syntax) == 0 {
		return 0
	}

	// Phase 1: Collect exported methods and their field accesses.
	methods := collectExportedMethods(pkg)
	if len(methods) == 0 {
		return 0
	}

	// Phase 2: Build union-find and merge methods sharing fields.
	uf := newUnionFind(len(methods))

	// For each pair of methods, check if they share any field.
	// O(n^2 * f) where n = methods, f = avg fields per method.
	// Acceptable for typical package sizes.
	for i := 0; i < len(methods); i++ {
		for j := i + 1; j < len(methods); j++ {
			if sharesField(methods[i].fields, methods[j].fields) {
				uf.union(i, j)
			}
		}
	}

	// Phase 3: Count connected components.
	roots := make(map[int]bool)
	for i := range methods {
		roots[uf.find(i)] = true
	}

	return metrics.LCOM(len(roots))
}

// methodInfo captures an exported method's qualified name and the set of struct
// field keys it accesses. Each methodInfo is a node in the LCOM4 graph; nodes
// are connected when their field sets overlap.
type methodInfo struct {
	name   string
	fields map[string]bool
}

// collectExportedMethods walks the package AST and returns one methodInfo per
// exported method declared on an exported receiver type. It resolves the
// receiver's base type name (handling value, pointer, and generic receiver
// forms via resolveReceiverType) and records the struct fields each method
// body accesses. Package-level functions (no receiver) and methods on
// unexported types are excluded.
func collectExportedMethods(pkg *packages.Package) []methodInfo {
	var methods []methodInfo

	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || fn.Body == nil {
				continue
			}

			// Must have a receiver (i.e., be a method, not a function).
			if len(fn.Recv.List) == 0 {
				continue
			}

			// Method must be exported.
			if !fn.Name.IsExported() {
				continue
			}

			// Receiver type must be exported.
			recvType := resolveReceiverType(fn.Recv.List[0].Type)
			if recvType == "" || !ast.IsExported(recvType) {
				continue
			}

			// Walk the method body to find field accesses.
			fields := collectFieldAccesses(fn.Body, pkg.TypesInfo)
			methods = append(methods, methodInfo{
				name:   recvType + "." + fn.Name.Name,
				fields: fields,
			})
		}
	}

	return methods
}

// resolveReceiverType extracts the base type name from a method receiver
// expression. It supports every Go receiver form:
//
//	func (t T)          Ident                            → "T"
//	func (t *T)         StarExpr{Ident}                  → "T"
//	func (t T[P])       IndexExpr{X:Ident}               → "T"
//	func (t T[P1,P2])   IndexListExpr{X:Ident}           → "T"
//	func (t *T[P])      StarExpr{IndexExpr{X:Ident}}     → "T"
//	func (t *T[P1,P2])  StarExpr{IndexListExpr{X:Ident}} → "T"
//
// A single pointer indirection is unwrapped first, then the underlying value
// form is resolved. It returns "" only for genuinely unknown forms.
func resolveReceiverType(expr ast.Expr) string {
	// Unwrap a single pointer indirection: *T, *T[P], *T[P1, P2].
	if star, ok := expr.(*ast.StarExpr); ok {
		expr = star.X
	}
	return baseTypeName(expr)
}

// baseTypeName resolves a value-form receiver type expression to its base
// identifier name. It handles plain identifiers (T) and generic instantiations
// (T[P] and T[P1, P2]) by extracting the underlying generic type identifier.
// It returns "" for any other expression form.
func baseTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		// Value receiver: T.
		return t.Name
	case *ast.IndexExpr:
		// Single type parameter: T[P] — extract T.
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name
		}
	case *ast.IndexListExpr:
		// Multiple type parameters: T[P1, P2] — extract T.
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name
		}
	}
	return ""
}

// collectFieldAccesses walks an AST node and returns the set of struct field
// keys accessed within it. A field key is "TypeName.FieldName" to distinguish
// fields on different types.
func collectFieldAccesses(node ast.Node, info *types.Info) map[string]bool {
	fields := make(map[string]bool)

	ast.Inspect(node, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		// Resolve the selection to a types.Object.
		obj, ok := info.Selections[sel]
		if !ok {
			return true
		}

		// Only consider direct field accesses (not method calls).
		if obj.Kind() != types.FieldVal {
			return true
		}

		// Build a key from the receiver type and field name.
		recv := obj.Recv()
		if recv == nil {
			return true
		}

		// Dereference pointer types to get the underlying named type.
		underlying := recv
		if ptr, ok := underlying.(*types.Pointer); ok {
			underlying = ptr.Elem()
		}

		named, ok := underlying.(*types.Named)
		if !ok {
			return true
		}

		typeName := named.Obj().Name()
		fieldName := sel.Sel.Name
		fields[typeName+"."+fieldName] = true

		return true
	})

	return fields
}

// sharesField returns true if two field sets have at least one common element.
func sharesField(a, b map[string]bool) bool {
	// Iterate over the smaller set for efficiency.
	if len(a) > len(b) {
		a, b = b, a
	}
	for field := range a {
		if b[field] {
			return true
		}
	}
	return false
}

// unionFind implements a disjoint-set data structure with path compression
// and union by rank for efficient connected component tracking.
type unionFind struct {
	parent []int
	rank   []int
}

// newUnionFind creates a union-find with n elements, each in its own set.
func newUnionFind(n int) *unionFind {
	parent := make([]int, n)
	rank := make([]int, n)
	for i := range parent {
		parent[i] = i
	}
	return &unionFind{parent: parent, rank: rank}
}

// find returns the root representative of the set containing x,
// applying path compression.
func (uf *unionFind) find(x int) int {
	if uf.parent[x] != x {
		uf.parent[x] = uf.find(uf.parent[x])
	}
	return uf.parent[x]
}

// union merges the sets containing x and y using union by rank.
func (uf *unionFind) union(x, y int) {
	rx, ry := uf.find(x), uf.find(y)
	if rx == ry {
		return
	}
	if uf.rank[rx] < uf.rank[ry] {
		rx, ry = ry, rx
	}
	uf.parent[ry] = rx
	if uf.rank[rx] == uf.rank[ry] {
		uf.rank[rx]++
	}
}
