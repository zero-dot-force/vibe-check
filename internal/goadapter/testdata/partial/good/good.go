// Package good is a valid package for testing partial-build scenarios.
// When analyzed alongside a broken sibling package, this package should
// still produce valid metrics.
package good

// Hello returns a greeting string.
func Hello() string {
	return "hello"
}
