package metrics

// Cycle represents a circular dependency between modules as the set of module
// paths that participate in the cycle. The members are reported as a
// deterministic, lexicographically-sorted list of package paths. The ordering
// carries no traversal meaning — it is the sorted membership set — and each
// member appears exactly once.
//
// Example: modules that form a cycle among A, B, and C are represented as
// ["A", "B", "C"] regardless of the direction in which the cycle was traversed.
//
// Multiple cycles in a result set are themselves sorted lexicographically by
// their first (smallest) element.
type Cycle []string
