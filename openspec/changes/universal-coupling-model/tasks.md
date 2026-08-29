## 1. Project Initialization

- [x] 1.1 Initialize Go module (`go mod init github.com/zero-dot-force/vibe-check`) and create the `metrics` package directory
- [x] 1.2 Configure golangci-lint with project conventions (gofmt, GoDoc enforcement, error wrapping checks)

## 2. Core Metric Types

- [x] [P] 2.1 Define `Module` type with `Path` (string), `Name` (string), and metric fields: `Ca` (int), `Ce` (int), `ExportedTypes` (int), `AbstractTypes` (int)
- [x] [P] 2.2 Define named metric types: `Instability float64`, `Abstractness float64`, `Distance float64` with GoDoc documenting formulas, value ranges [0.0, 1.0], and units
- [x] [P] 2.3 Define `LCOM int` named type with GoDoc documenting the LCOM4 variant (Hitz & Montazeri, 1995): connected components in method-field graph; 0 = no methods/fields, 1 = fully cohesive, >1 = number of independent components
- [x] 2.4 Implement metric computation functions: `ComputeInstability(ca, ce int) Instability`, `ComputeAbstractness(abstractTypes, totalExported int) Abstractness`, `ComputeDistance(a Abstractness, i Instability) Distance`
- [x] 2.5 Write table-driven tests for all metric computation functions covering: (a) happy-path with known values, (b) zero-denominator edge cases, (c) boundary values (0.0, 1.0), (d) determinism verification (same inputs produce same outputs across calls)

## 3. Circular Dependency Types

- [x] 3.1 Define `Cycle` type as `[]string` (ordered module paths) and document the representation (no repeated start node)
- [x] 3.2 Write tests verifying Cycle representation invariants

## 4. ModuleGraph Structure

- [x] [P] 4.1 Define `Warning` type with `Code` (string), `Message` (string), and `Module` (string, optional) fields
- [x] [P] 4.2 Define `Zone` string type with constants: `ZoneMainSequence`, `ZoneOfPain`, `ZoneOfUselessness`, `ZoneNormal` and GoDoc for each (JSON values: `main-sequence`, `zone-of-pain`, `zone-of-uselessness`, `normal`)
- [x] [P] 4.3 Define `Status` string type with constants: `StatusComplete`, `StatusPartial`, `StatusError` and GoDoc for each
- [x] 4.4 Define `ModuleGraph` struct with fields: `Language` (string), `Modules` ([]ModuleResult), `Cycles` ([]Cycle), `Warnings` ([]Warning), `Status` (Status)
- [x] 4.5 Define `ModuleResult` struct embedding `Module` and adding computed metrics (Instability, Abstractness, Distance, LCOM) and Zone classification
- [x] 4.6 Implement `ComputeZone(a Abstractness, i Instability, d Distance) Zone` function with threshold constants and precedence: main-sequence first, then zone-of-pain, then zone-of-uselessness, then normal
- [x] 4.7 Write tests for zone classification covering all four zones, boundary conditions (D=0.2 exactly), and overlapping criteria precedence

## 5. Adapter Interface

- [x] 5.1 Define `Adapter` interface with `Analyze(ctx context.Context, projectPath string) (*ModuleGraph, error)` and `Language() string` methods
- [x] 5.2 Define `Capability` type and `Capabilities() []Capability` method on the Adapter interface for metric capability discovery
- [x] 5.3 Implement `Registry` struct (not global — injectable) with `Register(Adapter) error` and `Get(language string) (Adapter, error)` methods
- [x] 5.4 Write tests for Registry: register/retrieve, duplicate registration error, unknown language error
- [x] 5.5 Write compile-time interface satisfaction checks (`var _ Adapter = (*GoAdapter)(nil)` pattern) and a test verifying a mock adapter satisfies the interface contract

## 6. JSON Schema and Serialization

- [x] 6.1 Add JSON struct tags to all model types (`ModuleGraph`, `ModuleResult`, `Warning`, `Cycle`) ensuring zero values are serialized (no `omitempty` on metric fields); include `schemaVersion` field (initial value `"1.0"`) on `ModuleGraph`
- [x] 6.2 Define the JSON schema document (as a Go embed or standalone `.json` file) for ModuleGraph validation
- [x] 6.3 Implement `Validate(data []byte) error` function to validate JSON against the schema
- [x] 6.4 Write round-trip tests: construct ModuleGraph → marshal to JSON → validate against schema → unmarshal back → verify equality
- [x] 6.5 Write table-driven negative tests for `Validate()`: missing `language` field, null `warnings`, omitted metric fields, invalid `status` value, malformed JSON, empty input, extra unknown fields

## 7. External Analyzer Protocol

- [x] 7.1 Define JSON-RPC 2.0 request/response types for the `analyze`, `capabilities`, and `shutdown` methods (include `protocolVersion` in capabilities response)
- [x] 7.2 Implement `ExternalAdapter` struct wrapping a subprocess (exec.Cmd) that satisfies the `Adapter` interface, communicating via JSON-RPC over stdin/stdout with newline-delimited framing
- [x] 7.3 Implement subprocess lifecycle management: spawn, timeout (defaults: analyze=300s, capabilities=10s, shutdown=5s), clean shutdown via `shutdown` notification, SIGKILL after grace period, stderr capture (max 1 MB)
- [x] 7.5 Implement input validation (projectPath: no `..` traversal, must exist), environment sanitization (minimal allowlist), and response size limits (default: 100 MB)
- [x] 7.4 Write tests for ExternalAdapter using the `TestHelperProcess` pattern (`os.Args[0]` with `-test.run=TestHelperProcess`): (a) successful analysis round-trip, (b) timeout simulation (helper process sleeps past deadline), (c) crash simulation (helper exits non-zero), (d) shutdown lifecycle, (e) stderr capture, (f) response size limit enforcement, (g) environment sanitization verification

## 8. Documentation

- [x] [P] 8.1 Write package-level GoDoc for the `metrics` package explaining the universal model, two-layer architecture, and relationship to language adapters
- [x] [P] 8.2 Update AGENTS.md project structure section to reflect the new `metrics` package

<!-- spec-review: passed -->
