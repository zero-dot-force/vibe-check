## Context

Vibe-Check's `metrics/` package provides a universal coupling metrics model (Layer 1)
with the `Adapter` interface, `Module`/`ModuleResult`/`ModuleGraph` types, and
deterministic compute functions. No language adapter or CLI exists yet. This design
covers the Go-native adapter (Layer 2) and the `vibe-check analyze` CLI command — the
first language adapter that makes the toolkit usable.

The existing `metrics.Adapter` interface requires:
- `Analyze(ctx context.Context, projectPath string) (*ModuleGraph, error)`
- `Language() string`
- `Capabilities() []Capability`

The adapter must populate `Module` fields (Path, Name, Ca, Ce, ExportedTypes,
AbstractTypes) and compute derived metrics using `metrics.Compute*` functions. The
result must pass `metrics.Validate()`.

## Goals / Non-Goals

**Goals:**

- Implement a Go-native adapter in `internal/goadapter/` that resolves package
  dependencies via `golang.org/x/tools/go/packages` and computes all seven metrics.
- Implement a `vibe-check analyze` CLI command using cobra with JSON output and CI
  gate threshold flags.
- Produce deterministic, schema-valid `ModuleGraph` output.
- Support analyzing arbitrary Go module paths (defaulting to `./...`).
- Propagate `context.Context` for cancellation and timeout support.
- Add an extensions mechanism to `ModuleResult` for language-specific metrics.
- Populate Go-specific extensions: interface width (Pike metric) and interface
  proximity (consumer vs. producer declaration).

**Non-Goals:**

- Multi-language support (Python, TS/JS adapters are P2/P3).
- Non-JSON output formats (table, SARIF — future work).
- Incremental/cached analysis (full analysis each invocation).
- IDE integration or watch mode.
- LCOM4 computation for non-Go languages.
- Provenance metadata in `ModuleGraph` schema (Constitution III gap — tracked as
  follow-up issue to add `producer`, `version`, `timestamp`, `commit` fields).
- CI gate flags for extension metrics (interface width, proximity — can be added later).
- Schema enforcement on extension contents (validated as proper JSON objects only).

## Decisions

### D1: Package-level analysis granularity

**Decision**: Analyze at the Go package level — each Go package is one `metrics.Module`.

**Rationale**: Go packages are the natural unit of encapsulation and dependency. Import
relationships are explicit and resolved by the compiler. Package-level analysis aligns
with how Go developers reason about coupling.

**Alternatives considered**:
- File-level: Too granular, misaligns with Go's package-centric model.
- Module-level: Too coarse, hides internal coupling problems.

### D2: `golang.org/x/tools/go/packages` for dependency resolution

**Decision**: Use `packages.Load` with `NeedName | NeedImports | NeedTypes | NeedSyntax |
NeedTypesInfo` mode flags to get type-aware dependency and type information. `NeedTypesInfo`
provides the `types.Info` mapping from AST nodes to resolved types, which is required for
accurate LCOM4 field-access resolution in method bodies.

**Rationale**: The `go/packages` API is the canonical way to load Go packages with full
type information. It handles build constraints, modules, vendoring, and CGo correctly.
Raw `go list` output lacks type information needed for abstractness and LCOM computation.

**Alternatives considered**:
- `go list -json`: Missing type info (exported/abstract type counts, method-field graphs).
- `go/build`: Deprecated in favor of `go/packages`.
- `go/ast` + manual resolution: Reimplements what `go/packages` already provides.

### D3: Abstractness via AST type classification

**Decision**: Walk the AST of each package to count exported types and classify
interfaces as abstract types. An exported type with at least one method in its method
set is counted. Interfaces are abstract; structs, named types, and type aliases
are concrete.

**Rationale**: Go has a single mechanism for abstraction — interfaces. Unlike Java/C#
there are no abstract classes. This makes classification unambiguous:
`ExportedTypes` = count of all exported type declarations,
`AbstractTypes` = count of exported interface declarations.

### D4: LCOM4 via method-field connected components

**Decision**: Compute LCOM4 per package using a graph where nodes are exported methods
and edges connect methods that share field access. LCOM4 = number of connected
components in this graph.

**Rationale**: LCOM4 (Hitz & Montazeri 1995) uses connected-component semantics which
are well-defined and deterministic. For Go, "fields" are package-level variables and
struct fields accessed by methods. Methods sharing no state form separate components,
indicating the package should potentially be split.

**Scope boundary**: LCOM is computed per package. Methods are functions with a receiver
belonging to an exported type. Package-level functions without receivers are excluded
from the LCOM graph (they don't indicate type cohesion).

### D5: Circular dependency detection via Tarjan's SCC

**Decision**: Use Tarjan's strongly connected components algorithm on the package
import graph to detect cycles. Each SCC with more than one node is a cycle.

**Rationale**: Tarjan's algorithm is O(V+E), well-understood, and produces canonical
cycle representations. Go technically prohibits import cycles at compile time, but
the adapter should still detect them to (a) report on near-cycles when analyzing
partial builds and (b) maintain consistency with multi-language adapters where cycles
are possible.

### D6: Scope filtering — module boundary

**Decision**: Only analyze packages within the target module. Standard library and
third-party packages contribute to Ca/Ce counts but are not themselves analyzed as
modules in the output.

**Rationale**: Users want to understand the coupling characteristics of _their_ code.
External dependencies are relevant as coupling sources but not as analysis targets.
This also bounds the analysis to a tractable scope. The adapter relies on `go/packages`
module resolution for scope filtering — packages resolved by `go/packages` that fall
within the target module's import path prefix are in scope regardless of their physical
location.

### D7: CLI structure with cobra

**Decision**: Use cobra for CLI with a root `vibe-check` command and an `analyze`
subcommand. Flags on the `analyze` command:
- `--max-instability <float>` — fail if any module exceeds this instability
- `--max-distance <float>` — fail if any module exceeds this distance
- `--no-circular-deps` — fail if any cycles detected
- `--max-lcom <int>` — fail if any module's LCOM exceeds this value
- `--timeout <duration>` — maximum analysis time (e.g., `5m`, `30s`); no timeout by default
- `--json` — JSON output (default and only format for P0; flag exists for future-proofing)

Exit codes:
- 0: success, no threshold violations
- 1: threshold violations detected (policy failure)
- 2: analysis error, invalid arguments, or missing Go toolchain (tool failure)

Violations printed to stderr, JSON output to stdout.

**Rationale**: cobra is adopted per convention CS-009 to maintain consistency with the
broader Go ecosystem and to support future subcommands (e.g., `vibe-check report`,
`vibe-check drift`) without refactoring the CLI layer. It provides subcommand routing,
automatic help/usage generation, shell completion, and flag validation that `flag` alone
does not offer. The `--max-lcom` flag name matches the metric name (LCOM) and follows
the `--max-*` pattern established by `--max-instability` and `--max-distance`, avoiding
semantic inversion.

**Testable CLI pattern**: The analyze command MUST follow AP-002/AP-003. A
`RunAnalyze(opts AnalyzeOptions) (*AnalyzeResult, error)` function handles all logic.
The cobra command's `RunE` delegates to `RunAnalyze` with `io.Writer` fields for
stdout/stderr, enabling unit testing without subprocess execution.

### D8: Package layout

**Decision**:
- `cmd/vibe-check/main.go` — entry point, minimal (cobra Execute)
- `cmd/vibe-check/analyze.go` — analyze command definition and flag handling
- `internal/goadapter/adapter.go` — `Adapter` struct implementing `metrics.Adapter`
- `internal/goadapter/resolve.go` — package dependency resolution via `go/packages`
- `internal/goadapter/types.go` — type counting (exported, abstract)
- `internal/goadapter/lcom.go` — LCOM4 computation
- `internal/goadapter/cycles.go` — Tarjan's SCC for cycle detection
- `internal/goadapter/extensions.go` — Go-specific extensions (interface width, proximity)
- `internal/goadapter/doc.go` — package-level GoDoc

**Rationale**: `internal/goadapter/` prevents external import (this is an implementation
detail) and avoids using `go` as a package name (which is a Go keyword). Separating
concerns across files keeps each file focused. The `cmd/` directory follows Go project
conventions.

### D9: Context propagation

**Decision**: The adapter MUST propagate the provided `context.Context` to
`packages.Config.Context` for cancellation support. The CLI MUST create a context with
signal handling (SIGINT/SIGTERM) and pass it through to the adapter. When SIGINT is
received during analysis, no partial JSON MUST be written to stdout — the command
MUST exit cleanly with a non-zero status and an error message to stderr.

**Rationale**: `go/packages.Load` can take minutes on large codebases (acknowledged in
R1). Without context propagation, analysis could hang indefinitely in CI environments.
The `packages.Config` struct accepts a `Context` field specifically for this purpose.

### D10: Registry integration

**Decision**: The CLI directly instantiates the Go adapter rather than going through the
`metrics.Registry`. The registry pattern is designed for multi-adapter scenarios where
adapter selection is dynamic. With a single adapter, direct instantiation is simpler and
avoids unnecessary indirection.

**Rationale**: The registry exists for future multi-language support where a dispatcher
would select adapters by language. The P0 CLI knows it wants the Go adapter and gains
nothing from registry lookup. When multi-language support is added (P2+), the CLI will
be refactored to use the registry.

### D11: Language-specific extensions mechanism

**Decision**: Add an optional `Extensions map[string]any` field to `ModuleResult`
(with `json:"extensions,omitempty"`) and a corresponding `extensions` object in the
JSON schema. Add `"extensions": { "type": "object" }` to the module item's `properties`
list while keeping `additionalProperties: false` — this preserves schema strictness
while allowing only the extensions field. Bump `SchemaVersion` from `"1.0"` to `"1.1"`
to signal the schema evolution (backward-compatible addition per semver). The core
model does not interpret extensions — adapters populate them, consumers opt-in to
reading them.

**Type safety**: Extension values undergo type coercion during JSON round-trip
(`int` → `float64`, `map[string]int` → `map[string]interface{}`). The adapter
package MUST provide typed accessor functions (e.g.,
`InterfaceWidths(extensions map[string]any) (map[string]int, error)`) that safely
extract and validate extension values after JSON unmarshaling. This prevents
consumers from needing unchecked type assertions.

**Rationale**: Go-specific metrics like interface width and interface proximity are
valuable for Go codebases but do not belong in the universal model. An extensions
mechanism allows language adapters to carry additional metrics without schema changes
for each new language. Keys are namespaced by language (e.g., `go.interfaceWidth`)
to prevent collisions between adapters.

**Alternatives considered**:
- Separate output field: Adds complexity to `ModuleGraph` and requires schema changes.
- Adapter-specific output structs: Breaks the unified `ModuleGraph` pipeline.

### D12: Interface Width (Go Pike Metric)

**Decision**: The Go adapter populates `go.interfaceWidth` in extensions as a
`map[string]int` mapping exported interface names to their method count. Go idiom
favors narrow interfaces (1-2 methods); wider interfaces signal abstraction problems.

**Rationale**: Rob Pike / Go Proverbs recommend small interfaces. Measuring interface
width surfaces packages with overly broad abstractions. This metric is language-specific
(Go's implicit interface satisfaction makes width semantics different from Java/C#).

**Scope**: Only exported interfaces in analyzed packages. Standard library and
third-party interfaces are excluded.

### D13: Interface-to-Implementation Proximity

**Decision**: The Go adapter populates `go.interfaceProximity` in extensions as a
`map[string]string` mapping exported interface names to `"consumer"` or `"producer"`.
An interface is `"consumer"` if it is declared in a different package from its primary
implementation; `"producer"` if declared in the same package.

**Rationale**: Consumer-side interface declaration is a Go best practice (per Go FAQ,
Effective Go). Measuring proximity surfaces packages where interfaces are declared
on the producer side, which can lead to unnecessary coupling.

**Heuristic**: For each exported interface, check whether any type in the same package
implements it. If yes → `"producer"`. If no implementation found in the same package →
`"consumer"`. This is a heuristic — the "primary" implementation is not always in the
same module.

## Coverage Strategy

**Unit tests**: Each `internal/goadapter/*.go` file has a corresponding `*_test.go`.
Tests use constructed adjacency maps, AST fixtures, and `testdata/` Go modules with
known characteristics. Coverage target: >= 80% line coverage for `internal/goadapter/`.

**Integration test**: Analyze a dedicated `testdata/` Go module (not self-analysis) with
known structure. Assert specific metric values and verify output passes
`metrics.Validate()`. Guard with `testing.Short()` per TC-011.

**CLI tests**: Use the testable CLI pattern (AP-003) — inject `bytes.Buffer` for
stdout/stderr. Test flag parsing, threshold violation logic, exit codes, and JSON output
validity. No subprocess execution in tests.

**Determinism verification**: Call `Analyze()` on the same testdata module multiple times
and assert JSON output is byte-identical, following the pattern in
`metrics/compute_test.go`.

**CLI tests coverage target**: >= 80% line coverage for `cmd/vibe-check/`.

**Coverage ratchet**: Enforced by CI via `go test -coverprofile`.

**Module ordering**: The adapter MUST sort `Modules` by `Module.Path` in lexicographic
order before returning the `ModuleGraph` to ensure deterministic output across runs.

## Operational Notes

**Memory consumption**: `go/packages` with `NeedSyntax | NeedTypes` loads full ASTs and
type information into memory for all requested packages simultaneously. For large
codebases (1000+ packages), expect approximately 1-2 MB per package. Users analyzing
large monorepos should scope analysis to specific packages rather than `./...`.

**Error messages**: Error messages MUST include the operation that failed, the underlying
cause, and a suggested remediation when determinable (e.g., "failed to load packages:
go.mod not found in /path — ensure the target directory contains a Go module").

**Version embedding**: The root command supports `--version` via cobra's built-in version
flag. Version, commit hash, and build date are embedded at build time via ldflags
(`-ldflags "-X main.version=... -X main.commit=... -X main.date=..."`). The version
output format MUST be: `vibe-check version <version> (commit <hash>, built <date>)`.

**Environment sanitization**: The adapter MUST set `packages.Config.Env` to a sanitized
environment using `metrics.SanitizeEnvironment` to prevent credential leakage to
subprocesses spawned by `go/packages` during package loading.

**Warning content**: Warning messages for partial builds MUST contain the affected
package's import path and the underlying error message using relative paths (relative
to the project root) rather than absolute paths. Warning code MUST be a machine-readable
identifier (e.g., `"load-error"`, `"type-check-error"`).

**Performance baseline**: Analysis of the vibe-check module itself (currently ~1 package)
MUST complete within 30 seconds on a standard CI runner. Large-codebase performance
optimization is deferred to P2.

## Risks / Trade-offs

**[R1] `go/packages` load time on large codebases** → The `NeedSyntax` mode flag
triggers full parsing. For very large monorepos (1000+ packages), initial load may be
slow. **Mitigation**: Accept for P0. Context cancellation provides timeout escape hatch.
Incremental analysis (P2+) will address performance.

**[R2] LCOM4 accuracy for Go idioms** → Go's implicit interface satisfaction and
package-level functions don't map perfectly to OOP LCOM models. **Mitigation**: Document
the adaptation in GoDoc. Emit a warning when a package has only package-level functions
(no receivers) since LCOM is not meaningful in that case.

**[R3] Build errors in analyzed code** → If the target codebase has compilation errors,
`go/packages` may return partial results. **Mitigation**: Check `packages.Package.Errors`
and set `Status: StatusPartial` with warnings listing the affected packages. When a
package's type information is unavailable (nil `pkg.Types`), set ExportedTypes and
AbstractTypes to 0, set LCOM to 0, and add a warning.

**[R4] New dependencies increase attack surface** → Adding cobra and x/tools introduces
transitive dependencies. `golang.org/x/tools` is large but only `go/packages` (and its
transitive deps within x/tools) is imported. cobra brings `pflag` and `mousetrap` as
transitive deps. **Mitigation**: Both are widely-used, well-maintained Go ecosystem
projects. Pin versions in `go.mod`. Run `go mod tidy` to include only necessary deps.
