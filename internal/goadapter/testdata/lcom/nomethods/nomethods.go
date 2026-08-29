// Package nomethods contains only package-level functions with no receivers.
//
// Expected LCOM4 = 0 (no methods, trivially cohesive).
//
// LCOM4 only considers exported methods on exported types. Package-level
// functions are excluded from the method-field graph entirely.
package nomethods

// Add returns the sum of two integers.
func Add(a, b int) int {
	return a + b
}

// Multiply returns the product of two integers.
func Multiply(a, b int) int {
	return a * b
}

// Negate returns the negation of an integer.
func Negate(a int) int {
	return -a
}
