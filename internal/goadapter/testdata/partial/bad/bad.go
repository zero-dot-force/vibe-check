// Package bad deliberately imports a non-existent package to test
// partial-build error handling. The adapter should report warnings
// for this package while still analyzing valid sibling packages.
package bad

import "example.com/doesnotexist" //nolint:all // deliberate missing import

// Broken attempts to use the non-existent package.
func Broken() string {
	return doesnotexist.Value() //nolint:all // deliberate reference to missing package
}
