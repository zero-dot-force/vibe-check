// Package pkga demonstrates efferent coupling.
//
// pkga imports pkgb (module-internal) and fmt (stdlib).
//
// Expected metrics:
//
//	Ca = 0  (no module-internal package imports pkga)
//	Ce = 2  (pkga imports fmt + pkgb)
package pkga

import (
	"fmt"

	"example.com/coupling/pkgb"
)

// Greet formats a greeting using pkgb's Name function.
func Greet() string {
	return fmt.Sprintf("Hello, %s!", pkgb.Name())
}
