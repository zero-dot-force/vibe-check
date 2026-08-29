package metrics

// Cycle represents a circular dependency between modules as an ordered list
// of module paths. The cycle starts from the lexicographically smallest module
// path and does not repeat the start node.
//
// Example: if modules A→B→C→A form a cycle, it is represented as ["A", "B", "C"].
//
// Canonical ordering: when a cycle is detected (e.g., C→A→B→C), it is rotated
// so that the lexicographically smallest module path appears first, yielding
// ["A", "B", "C"]. Multiple cycles in a result set are sorted lexicographically
// by their first element.
type Cycle []string
