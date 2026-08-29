## Why

No single OSS tool computes the full Martin metrics suite (Ca, Ce, Instability, Abstractness, Distance from Main Sequence) for Go — or any other language — through a unified model. Vibe-Check needs a universal, language-agnostic metrics model so that the core analysis engine remains stable while language-specific adapters handle extraction. Without this shared model, each language adapter would define its own metric representations, making cross-language comparison impossible and coupling the analysis engine to language-specific details. This is a P1 prerequisite: the universal model must exist before any language-specific implementation (Group 1 Go adapter, Group 2a Python adapter) can begin.

## What Changes

- Define a universal metrics model covering seven coupling and cohesion metrics: Afferent Coupling (Ca), Efferent Coupling (Ce), Instability (I), Abstractness (A), Distance from Main Sequence (D), Cohesion (LCOM), and Circular Dependency detection
- Establish a two-layer architecture: Layer 1 is the universal metrics model (language-agnostic types, interfaces, and JSON schema); Layer 2 is language-specific adapters that extract raw data and produce Layer 1 structures
- Define the `Adapter` interface that all language analyzers must implement, ensuring new languages can be added without modifying the core analysis engine
- Specify the external analyzer protocol (JSON-RPC over stdin/stdout) for out-of-process language adapters
- Define the JSON interchange schema with `language` field, `warnings` array, zone classification, and status metadata
- Document metric value ranges, units, and determinism guarantees for all seven metrics

## Capabilities

### New Capabilities

- `universal-metrics-model`: Core metric types (Ca, Ce, I, A, D, LCOM, circular deps), value ranges, units, and determinism contracts for all computed metrics
- `adapter-interface`: The `Adapter` interface contract that language-specific analyzers implement, including the registration mechanism and capability discovery
- `analyzer-protocol`: External analyzer JSON-RPC protocol for out-of-process language adapters communicating over stdin/stdout
- `metrics-schema`: JSON interchange schema for metric results including `language` field, `warnings` array, zone classification, and status metadata

### Modified Capabilities

<!-- None — greenfield project -->

### Removed Capabilities

<!-- None — greenfield project -->

## Constitution Alignment

| Principle | Assessment |
|-----------|------------|
| I. Autonomous Collaboration | PASS — Adapter interface enables independent language analyzer development |
| II. Composability First | PASS — Two-layer architecture separates universal model from language-specific adapters |
| III. Observable Quality | PASS — All metric values have defined ranges, units, and determinism guarantees |
| IV. Testability | PASS — All metrics have concrete formulas enabling table-driven tests |
| V. Security by Default | PASS — External analyzer protocol defines lifecycle management and error handling |
| VI. Metric Fidelity | PASS — Each metric cites its formula and source (Robert C. Martin); LCOM variant to be specified |
| VII. Language Agnosticism | PASS — Universal "Module" abstraction; no language-specific assumptions in Layer 1 |

Addresses: https://github.com/zero-dot-force/vibe-check/issues/1

## Impact

- **Code**: Establishes the foundational types and interfaces in a new `metrics` package (or similar) that all subsequent Groups (1, 2a, 3) will depend on
- **APIs**: Defines the `Adapter` interface and JSON-RPC protocol that every language analyzer must conform to — changes to these after initial release are **BREAKING**
- **Dependencies**: No new external dependencies for the model itself; Go adapter (future) will require `golang.org/x/tools/go/packages`
- **Systems**: Sets the contract for the Unbound Force ecosystem's entropy sentinel and architectural drift tracking capabilities
- **Sequencing**: Blocks Group 1 (Go coupling engine), Group 2a (Python adapter), and Group 3 (circular dependency detection) — all depend on this universal model
