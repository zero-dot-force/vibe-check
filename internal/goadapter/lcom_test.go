package goadapter

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/zero-dot-force/vibe-check/metrics"
)

func TestLCOM4_Cohesive(t *testing.T) {
	t.Parallel()

	pkg := loadTestPackage(t, "lcom", "cohesive")
	got := computeLCOM4(pkg)

	if got != metrics.LCOM(1) {
		t.Errorf("LCOM4: got %d, want %d", got, 1)
	}
}

func TestLCOM4_NonCohesive(t *testing.T) {
	t.Parallel()

	pkg := loadTestPackage(t, "lcom", "noncohesive")
	got := computeLCOM4(pkg)

	if got != metrics.LCOM(2) {
		t.Errorf("LCOM4: got %d, want %d", got, 2)
	}
}

func TestLCOM4_NoMethods(t *testing.T) {
	t.Parallel()

	pkg := loadTestPackage(t, "lcom", "nomethods")
	got := computeLCOM4(pkg)

	if got != metrics.LCOM(0) {
		t.Errorf("LCOM4: got %d, want %d", got, 0)
	}
}

// TestLCOM4_GenericPointerReceiver proves that methods on a generic pointer
// receiver (*Box[T]) are counted, not dropped. Both Get and Set share field
// val, so LCOM4 must be 1. Before the receiver-resolution fix the generic
// pointer receiver resolved to "", the methods were dropped, and LCOM4 was 0.
func TestLCOM4_GenericPointerReceiver(t *testing.T) {
	t.Parallel()

	pkg := loadTestPackage(t, "lcom", "generic")
	got := computeLCOM4(pkg)

	if got != metrics.LCOM(1) {
		t.Errorf("LCOM4: got %d, want %d (generic pointer-receiver methods must not be dropped)", got, 1)
	}
}

// parseReceiverType parses a single method declaration and returns its
// receiver type expression for direct testing of resolveReceiverType.
func parseReceiverType(t *testing.T, methodSrc string) ast.Expr {
	t.Helper()

	src := "package p\n" + methodSrc + "\n"
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "recv.go", src, 0)
	if err != nil {
		t.Fatalf("parse %q: %v", methodSrc, err)
	}
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv == nil || len(fn.Recv.List) == 0 {
			continue
		}
		return fn.Recv.List[0].Type
	}
	t.Fatalf("no method receiver found in %q", methodSrc)
	return nil
}

// TestResolveReceiverType_AllForms verifies that every supported Go receiver
// form resolves to the base type name "T".
func TestResolveReceiverType_AllForms(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		src  string
		want string
	}{
		{"value", "func (t T) M() {}", "T"},
		{"pointer", "func (t *T) M() {}", "T"},
		{"generic_value_single", "func (t T[P]) M() {}", "T"},
		{"generic_value_multi", "func (t T[P1, P2]) M() {}", "T"},
		{"generic_pointer_single", "func (t *T[P]) M() {}", "T"},
		{"generic_pointer_multi", "func (t *T[P1, P2]) M() {}", "T"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			recv := parseReceiverType(t, tt.src)
			if got := resolveReceiverType(recv); got != tt.want {
				t.Errorf("resolveReceiverType(%s): got %q, want %q", tt.src, got, tt.want)
			}
		})
	}
}

// TestResolveReceiverType_UnknownForm verifies that unrecognized receiver
// expressions resolve to "" rather than guessing a name.
func TestResolveReceiverType_UnknownForm(t *testing.T) {
	t.Parallel()

	// A selector expression (pkg.T) is not a valid receiver form.
	unknown := &ast.SelectorExpr{X: ast.NewIdent("pkg"), Sel: ast.NewIdent("T")}
	if got := resolveReceiverType(unknown); got != "" {
		t.Errorf("resolveReceiverType(SelectorExpr): got %q, want empty string", got)
	}

	// A pointer to an unknown form must also resolve to "".
	ptrUnknown := &ast.StarExpr{X: unknown}
	if got := resolveReceiverType(ptrUnknown); got != "" {
		t.Errorf("resolveReceiverType(*SelectorExpr): got %q, want empty string", got)
	}
}
