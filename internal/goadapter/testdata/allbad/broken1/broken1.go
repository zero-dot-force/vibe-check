// Package broken1 deliberately imports a non-existent package so that it
// fails to type-check. Alongside broken2 it ensures every package in the
// allbad fixture module is un-analyzable, exercising the go-adapter
// total-load-failure scenario in which no package can be type-checked and
// Analyze must return an error rather than a graph of all-zeroed modules.
package broken1

import "example.com/doesnotexist" //nolint:all // deliberate missing import

// Broken references the non-existent package.
func Broken() string {
	return doesnotexist.Value() //nolint:all // deliberate reference to missing package
}
