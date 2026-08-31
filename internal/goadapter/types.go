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

		// Type aliases are always classified as concrete per the go-adapter
		// spec, even when they alias an interface (e.g.,
		// `type IReader = SomeInterface`). An alias introduces no new abstract
		// type — it is merely a concrete name for an existing type — so it
		// counts toward exportedTypes (incremented above) but never
		// abstractTypes.
		if tn.IsAlias() {
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
