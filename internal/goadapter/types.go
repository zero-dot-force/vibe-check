package goadapter

import (
	"go/types"

	"golang.org/x/tools/go/packages"
)

// countTypes counts exported type declarations in a package, classifying
// each as abstract (interface) or concrete (struct, named type, alias).
// Unexported types are excluded from both counts.
//
// Returns (0, 0) if pkg.Types is nil (type-checking failed for this package).
func countTypes(pkg *packages.Package) (exportedTypes, abstractTypes int) {
	if pkg.Types == nil {
		return 0, 0
	}

	scope := pkg.Types.Scope()
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)

		// Only count type names (not vars, funcs, consts).
		tn, ok := obj.(*types.TypeName)
		if !ok {
			continue
		}

		// Only count exported types.
		if !tn.Exported() {
			continue
		}

		exportedTypes++

		// Type aliases: check the underlying type of the RHS.
		// A type alias like `type Foo = SomeInterface` is classified
		// based on the aliased type. However, most aliases (e.g.,
		// `type Alias = string`) are concrete.
		if tn.IsAlias() {
			if types.IsInterface(tn.Type()) {
				abstractTypes++
			}
			continue
		}

		// Named types: check if the underlying type is an interface.
		named, ok := tn.Type().(*types.Named)
		if !ok {
			continue
		}
		if types.IsInterface(named.Underlying()) {
			abstractTypes++
		}
	}

	return exportedTypes, abstractTypes
}
