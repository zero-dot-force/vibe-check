// Package generic demonstrates a generic type whose methods use a pointer
// receiver and share a field. It is a regression fixture for the receiver
// resolution bug: before the fix, a generic pointer receiver *Box[T]
// (StarExpr wrapping an IndexExpr) resolved to "" and its methods were
// silently dropped from the LCOM4 graph, corrupting the metric.
//
// Expected LCOM4 = 1: both Get and Set access field val, forming a single
// connected component. If the generic pointer receiver were dropped, LCOM4
// would incorrectly be 0.
package generic

// Box is a generic container with methods declared on a pointer receiver.
type Box[T any] struct {
	val T //nolint:unused // accessed via Get and Set
}

// Get returns the stored value.
func (b *Box[T]) Get() T {
	return b.val
}

// Set stores a new value.
func (b *Box[T]) Set(v T) {
	b.val = v
}
