// Package pkgc demonstrates efferent coupling to pkgb.
//
// pkgc imports pkgb (module-internal) and no stdlib packages.
//
// Expected metrics:
//
//	Ca = 0  (no module-internal package imports pkgc)
//	Ce = 1  (pkgc imports pkgb)
package pkgc

import "example.com/coupling/pkgb"

// DoubleValue returns twice pkgb's Value.
func DoubleValue() int {
	return pkgb.Value() * 2
}
