## Context

Vibe-Check is a design quality and architectural metrics toolkit for Go codebases, with planned multi-language support. The project currently has no implementation — this is the foundational architecture decision that all subsequent work builds upon.

The RFC (unbound-force/discussions/483) defines a two-layer architecture inspired by Gaze's side-effect taxonomy pattern: a universal metrics model (Layer 1) that language-specific adapters (Layer 2) target. This mirrors how Gaze separates its side-effect classification model from language-specific AST walkers.

Key constraints from the project constitution:
- Metric computations MUST produce deterministic results for the same input
- Coupling analysis MUST handle circular dependencies without infinite loops or panics
- All metric values MUST have defined ranges and units documented in GoDoc
- Language adapters MUST implement a common interface; adding a new language MUST NOT require changes to the core analysis engine
- No global mutable state; dependency injection for all services
- Standard library `testing` package only

## Goals / Non-Goals

### Goals

- Define the universal metrics model as Go types that represent Ca, Ce, Instability, Abstractness, Distance from Main Sequence, LCOM cohesion, and circular dependency data
- Establish the `Adapter` interface that all language analyzers implement
- Specify the JSON interchange schema for metric results (used by external adapters and for serialization)
- Define the external analyzer protocol (JSON-RPC over stdin/stdout) for out-of-process language adapters
- Ensure the model is language-agnostic: no Go-specific assumptions leak into the universal layer

### Non-Goals

- Implementing any language-specific adapter (Go adapter is Group 1, Python is Group 2a)
- Building the CLI or output formatting (Group 3 and later)
- Implementing complex metric computation algorithms such as graph traversal, AST parsing, or cycle detection algorithms (simple formula applications like I=Ce/(Ca+Ce) are part of the model; the heavy analysis logic belongs in language adapters)
- Making zone classification thresholds user-configurable (default thresholds are part of the model; configurable overrides are a P2+ concern)
- Building the entropy sentinel or architectural drift tracking (P2+)

## Decisions

### D1: Package structure — flat `metrics` package for the universal model

**Decision**: Place all universal model types in a single `metrics` package at the module root. The `Adapter` interface lives in this same package.

**Rationale**: The model is cohesive — Ca, Ce, I, A, D are all properties of the same unit of analysis (a package/module). Splitting them across packages would create unnecessary coupling between packages and force consumers to import multiple packages for basic operations.

**Alternatives considered**:
- Separate `metrics/model` and `metrics/adapter` packages — rejected because the adapter interface depends on model types, creating a tight coupling that a package boundary would not meaningfully separate
- Nested `metrics/coupling`, `metrics/cohesion`, `metrics/circular` — rejected because the types are small and interrelated; premature decomposition would increase import surface

### D2: Unit of analysis — "Module" as the universal concept

**Decision**: Use `Module` as the universal term for the unit of analysis. In Go this maps to a package, in Python to a module/package, in TS/JS to a module/file. The universal model does not use language-specific terms.

**Rationale**: Every language has a concept of a module-level grouping that has imports (efferent coupling) and is imported by others (afferent coupling). Using a neutral term avoids baking Go's "package" terminology into the universal layer.

**Alternatives considered**:
- Use "Package" — rejected because Python and JS use "module" as the primary grouping, and "package" means something different in each language
- Use "Component" — rejected because it is too abstract and overloaded in software architecture terminology
- Use "Unit" — rejected because it conflicts with "unit test" terminology

### D3: Adapter interface — synchronous, single-module analysis

**Decision**: The `Adapter` interface has two core methods: `Analyze(ctx, path) -> ModuleGraph` for full-project analysis, and `Language() string` for capability identification. Adapters return a complete `ModuleGraph` containing all modules and their relationships.

**Rationale**: Full-project analysis is required to compute afferent coupling (you need to know all importers). Streaming or per-module analysis would require the caller to assemble the graph, pushing complexity to the wrong layer.

**Alternatives considered**:
- Per-module analysis with caller-side assembly — rejected because Ca computation requires global knowledge
- Async/channel-based results — rejected as premature optimization; adds concurrency complexity without demonstrated need

### D4: External adapter protocol — JSON-RPC 2.0 over stdin/stdout

**Decision**: External (out-of-process) language adapters communicate via JSON-RPC 2.0 over stdin/stdout. The host process spawns the adapter as a subprocess and exchanges JSON-RPC messages.

**Rationale**: JSON-RPC is a simple, well-specified protocol. stdin/stdout avoids port management and firewall complexity. This matches the pattern established by LSP and used successfully in the Gaze ecosystem for external analyzer integration (issue #95).

**Alternatives considered**:
- gRPC — rejected because it requires protobuf compilation and a heavier dependency chain, disproportionate for the data volume involved
- REST over HTTP — rejected because it requires port allocation and lifecycle management for the subprocess
- Plain JSON over stdin/stdout (no RPC framing) — rejected because JSON-RPC provides request/response correlation, error codes, and batch support for free

### D5: Metric value representation — concrete types with documented invariants

**Decision**: Each metric is represented as a named type (e.g., `Instability float64`) with documented value ranges in GoDoc. Metric values are plain numeric types, not wrapper structs.

**Rationale**: The metrics are simple numeric values with well-defined mathematical definitions. Wrapper structs would add allocation overhead and API complexity without benefit. Value ranges and units are documented via GoDoc comments and enforced by constructor/validation functions.

**Alternatives considered**:
- Wrapper structs with built-in validation (e.g., `type Instability struct { Value float64 }`) — rejected because it doubles the API surface for trivial values; validation belongs at the boundary where values are created
- Generic `Metric` type with name/value pairs — rejected because it loses type safety and makes the API stringly-typed

### D6: Circular dependency representation — cycle list with path information

**Decision**: Circular dependencies are represented as a slice of `Cycle` values, where each `Cycle` contains the ordered list of module identifiers forming the cycle. The shortest representation is used (no repeated start node).

**Rationale**: Consumers need to know which modules participate in each cycle and the dependency path. A simple boolean "has cycles" is insufficient for actionable diagnostics. The ordered path enables visualization and targeted refactoring advice.

**Alternatives considered**:
- Adjacency matrix with cycle detection delegated to consumers — rejected because cycle detection is a core responsibility of the analysis engine
- Strongly connected components (SCCs) — considered as the detection algorithm but the output representation is the cycle path, not the SCC grouping; SCCs may be used internally during detection

## Coverage Strategy

| Test Category | Scope | Target |
|---------------|-------|--------|
| **Unit tests** | Metric computation functions (ComputeInstability, ComputeAbstractness, ComputeDistance, ComputeZone), Cycle representation invariants | ≥90% line coverage |
| **Unit tests** | Registry operations (register, retrieve, duplicate, unknown) | ≥90% line coverage |
| **Integration tests** | JSON round-trip: construct ModuleGraph → marshal → validate → unmarshal → verify equality | ≥80% line coverage |
| **Integration tests** | ExternalAdapter subprocess protocol: mock subprocess ↔ JSON-RPC exchange, timeout, crash, shutdown | ≥80% line coverage |
| **Negative tests** | Schema validation with malformed/invalid JSON inputs | Part of integration target |

**Coverage ratchet**: Enforced via CI with `go test -coverprofile=coverage.out ./...`. Coverage MUST NOT decrease between commits. Minimum package-level target: 85% line coverage for the `metrics` package.

**Test parallelism**: Pure computation functions (metric formulas, zone classification) and Registry tests are safe for `t.Parallel()`. ExternalAdapter subprocess tests require sequential execution due to process lifecycle management.

## Risks / Trade-offs

- **[Risk] Model may not capture language-specific nuances** → Mitigation: The `ModuleGraph` includes a `Warnings` slice for language-specific caveats (e.g., Python dynamic imports, Go build tags). Adapters annotate results with language-specific context without polluting the universal model.

- **[Risk] "Module" abstraction may be too coarse for some languages** → Mitigation: The model supports hierarchical module paths (e.g., `github.com/foo/bar/baz`). If finer granularity is needed (e.g., class-level coupling), that is a P2+ concern and can be added as an optional detail level without breaking the module-level model.

- **[Risk] JSON-RPC protocol adds complexity for the first adapter (Go)** → Mitigation: The Go adapter runs in-process and implements the `Adapter` interface directly. JSON-RPC is only used for external (out-of-process) adapters. The protocol is designed but not required for the initial implementation.

- **[Risk] Metric definitions may diverge from academic sources** → Mitigation: Each metric type's GoDoc includes the mathematical formula and cites the source. Coupling metrics (Ca, Ce, I, A, D) cite Robert C. Martin, "Agile Software Development" (2003). Cohesion uses LCOM4 (Hitz & Montazeri, "Measuring Coupling and Cohesion in Object-Oriented Systems," 1995) — chosen for its connected-component semantics which map cleanly to Go's struct/method model. LCOM4 limitation: does not account for method call chains. Value ranges are explicitly documented and tested.

- **[Trade-off] Single `Analyze` call vs. incremental analysis** → We chose full-project analysis for simplicity. Incremental analysis (only re-analyzing changed modules) is a performance optimization for P2+ and can be added as an optional `Adapter` method without breaking the existing interface.

- **[Trade-off] Named types vs. plain float64** → Named types (e.g., `type Instability float64`) add a small ergonomic cost (explicit conversions) but provide self-documenting code and prevent metric value confusion (passing an Instability where Abstractness is expected).
