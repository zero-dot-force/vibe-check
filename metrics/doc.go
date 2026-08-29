// Package metrics defines the universal coupling and cohesion metrics model
// for the Vibe-Check toolkit.
//
// # Architecture
//
// The package follows a two-layer architecture:
//
//   - Layer 1 (this package): A language-agnostic metrics model defining types
//     for afferent coupling (Ca), efferent coupling (Ce), instability (I),
//     abstractness (A), distance from main sequence (D), cohesion (LCOM4),
//     and circular dependency detection.
//
//   - Layer 2 (language adapters): Language-specific analyzers that implement
//     the [Adapter] interface and populate the universal model with data from
//     Go, Python, TypeScript, or other codebases.
//
// # Core Types
//
// [Module] represents the universal unit of analysis (a package in Go, a module
// in Python, a file/module in TypeScript). [ModuleResult] embeds Module and adds
// computed metrics. [ModuleGraph] contains the complete analysis result for a
// project, including all modules, detected cycles, warnings, and status.
//
// # Metrics
//
// Each metric is represented as a named type with documented value ranges:
//
//   - [Instability]: I = Ce / (Ca + Ce), range [0.0, 1.0]
//   - [Abstractness]: A = abstractTypes / exportedTypes, range [0.0, 1.0]
//   - [Distance]: D = |A + I - 1|, range [0.0, 1.0]
//   - [LCOM]: LCOM4 variant (Hitz & Montazeri, 1995), non-negative integer
//
// Metric computation functions ([ComputeInstability], [ComputeAbstractness],
// [ComputeDistance], [ComputeZone]) produce deterministic results for the same
// inputs.
//
// # Adapters
//
// Language-specific analyzers implement the [Adapter] interface and register
// with a [Registry]. The registry uses dependency injection (no global state).
// Adding a new language adapter requires only implementing [Adapter] and
// calling [Registry.Register] — no changes to this package are needed.
//
// External (out-of-process) adapters communicate via JSON-RPC 2.0 over
// stdin/stdout using the [ExternalAdapter] type. The protocol uses
// newline-delimited JSON framing.
//
// # JSON Schema
//
// The [ModuleGraph] type is serializable to JSON using the schema defined in
// modulegraph.schema.json (accessible via [SchemaJSON]). The [Validate]
// function checks JSON data against this schema. The schema includes a
// schemaVersion field for forward compatibility.
//
// Citations:
//   - Robert C. Martin, "Agile Software Development" (2003) — Ca, Ce, I, A, D
//   - Hitz & Montazeri, "Measuring Coupling and Cohesion in Object-Oriented Systems" (1995) — LCOM4
package metrics
