// Package pkgb is a shared dependency with high afferent coupling.
//
// pkgb is imported by both pkga and pkgc but imports no module-internal packages.
//
// Expected metrics:
//
//	Ca = 2  (pkga and pkgc both import pkgb)
//	Ce = 0  (pkgb imports no module-internal packages)
package pkgb

// Name returns a fixed name string.
func Name() string {
	return "world"
}

// Value returns a fixed integer value.
func Value() int {
	return 42
}
