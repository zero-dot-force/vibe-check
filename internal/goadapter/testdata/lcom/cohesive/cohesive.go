// Package cohesive demonstrates a fully cohesive struct where all methods
// access the same field, forming a single connected component.
//
// Expected LCOM4 = 1 (one connected component).
//
// Method-field graph:
//
//	Get  → x
//	Set  → x
//	Inc  → x
//
// All three methods share field x, so they form one connected component.
package cohesive

// S is a struct with a single field accessed by all methods.
type S struct {
	x int //nolint:unused // accessed via methods
}

// Get returns the current value of x.
func (s *S) Get() int {
	return s.x
}

// Set assigns a new value to x.
func (s *S) Set(v int) {
	s.x = v
}

// Inc increments x by one.
func (s *S) Inc() {
	s.x++
}
