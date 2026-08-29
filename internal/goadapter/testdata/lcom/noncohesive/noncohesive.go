// Package noncohesive demonstrates a struct with two disconnected groups
// of methods, each accessing a disjoint set of fields.
//
// Expected LCOM4 = 2 (two connected components).
//
// Method-field graph:
//
//	MethodA → x, y
//	MethodB → x, y
//	MethodC → z, w
//	MethodD → z, w
//
// Component 1: {MethodA, MethodB} share fields {x, y}
// Component 2: {MethodC, MethodD} share fields {z, w}
package noncohesive

// S is a struct with four fields split across two method groups.
type S struct {
	x int //nolint:unused // accessed via MethodA, MethodB
	y int //nolint:unused // accessed via MethodA, MethodB
	z int //nolint:unused // accessed via MethodC, MethodD
	w int //nolint:unused // accessed via MethodC, MethodD
}

// MethodA accesses fields x and y.
func (s *S) MethodA() int {
	return s.x + s.y
}

// MethodB accesses fields x and y.
func (s *S) MethodB(v int) {
	s.x = v
	s.y = v
}

// MethodC accesses fields z and w.
func (s *S) MethodC() int {
	return s.z + s.w
}

// MethodD accesses fields z and w.
func (s *S) MethodD(v int) {
	s.z = v
	s.w = v
}
