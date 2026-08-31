// Package goadapter implements a Go-native language adapter for the vibe-check
// metrics toolkit. It analyzes Go packages using [golang.org/x/tools/go/packages]
// to compute coupling metrics (Ca, Ce), abstractness, instability, distance from
// main sequence, LCOM4 cohesion, and circular dependency detection.
//
// The adapter implements the [metrics.Adapter] interface and produces a
// [metrics.ModuleGraph] that conforms to schema version "1.1". Each Go package
// within the target module is treated as one [metrics.Module].
//
// # Type Classification
//
// An exported type is classified as abstract when its underlying type is an
// interface — this covers interface declarations and defined types whose
// underlying type is an interface (e.g., type Z SomeInterface). All other
// exported types (structs, other named types, and type aliases, including an
// alias to an interface) are classified as concrete. Unexported types are
// excluded from both counts.
//
// # LCOM4 Computation
//
// LCOM4 uses the Hitz & Montazeri (1995) connected-component variant.
// Nodes are exported methods (functions with a receiver of an exported type).
// Edges connect methods that access at least one common struct field.
// Package-level functions (no receiver) are excluded from the graph.
//
// # Extensions
//
// The adapter populates language-specific extensions under the "go." namespace:
//   - go.interfaceWidth: method count per exported interface (map[string]int)
//   - go.interfaceProximity: "consumer" or "producer" per interface (map[string]string)
//
// Use [InterfaceWidths] and [InterfaceProximities] to safely extract typed
// extension values after JSON round-trip.
//
// # Security
//
// Analysis loads packages with [golang.org/x/tools/go/packages], which invokes
// the Go toolchain and may execute code (e.g., cgo preprocessing) from the
// analyzed module. Run vibe-check only on trusted, self-owned code. The adapter
// sanitizes the subprocess environment via [metrics.SanitizeEnvironment] and
// excludes GOFLAGS to remove the -toolexec command-execution vector.
package goadapter
