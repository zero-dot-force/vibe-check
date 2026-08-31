// Package broken2 deliberately imports a non-existent package so that it
// fails to type-check. Together with broken1 it guarantees the allbad
// fixture module contains no analyzable package, so Analyze exercises the
// total-load-failure path.
package broken2

import "example.com/alsomissing" //nolint:all // deliberate missing import

// AlsoBroken references the non-existent package.
func AlsoBroken() string {
	return alsomissing.Value() //nolint:all // deliberate reference to missing package
}
