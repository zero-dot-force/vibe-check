## ADDED Requirements

### Requirement: Adapter implements metrics.Adapter interface

The Go adapter MUST implement the `metrics.Adapter` interface from the `metrics`
package. `Language()` MUST return `"go"`. `Capabilities()` MUST return all seven
capabilities: `CapAfferentCoupling`, `CapEfferentCoupling`, `CapInstability`,
`CapAbstractness`, `CapDistance`, `CapLCOM`, `CapCircularDeps`.

#### Scenario: Language identifier

- **GIVEN** a Go adapter instance
- **WHEN** `Language()` is called
- **THEN** the return value MUST be `"go"`

#### Scenario: Full capability set

- **GIVEN** a Go adapter instance
- **WHEN** `Capabilities()` is called
- **THEN** the returned slice MUST contain exactly seven capabilities matching all `Cap*` constants defined in the `metrics` package

### Requirement: Adapter resolves Go package dependencies

The adapter MUST use `golang.org/x/tools/go/packages` to load packages from the
target project path. Each Go package within the analyzed module MUST be represented
as one `metrics.Module`. The adapter MUST resolve import relationships to compute
afferent coupling (Ca) and efferent coupling (Ce) for each package.

#### Scenario: Single module analysis

- **GIVEN** a Go module containing packages A, B, C where A imports B and C imports B
- **WHEN** `Analyze` is called with the module path
- **THEN** the result MUST contain three `ModuleResult` entries, and package B MUST have `Ca == 2` (A and C depend on it). Ce counts all distinct import targets including standard library and third-party packages.

#### Scenario: Standard library exclusion

- **GIVEN** a package that imports standard library packages (e.g., `fmt`, `os`)
- **WHEN** `Analyze` is called
- **THEN** standard library packages MUST NOT appear as modules in the output, but MUST be counted in the importing package's Ce

#### Scenario: Third-party dependency exclusion

- **GIVEN** a package that imports packages outside the target module
- **WHEN** `Analyze` is called
- **THEN** external packages MUST NOT appear as modules in the output, but MUST be counted in the importing package's Ce

### Requirement: Adapter counts exported and abstract types

The adapter MUST inspect the type-checker scope of each package to count exported type
declarations. An exported type MUST be counted as abstract when its underlying type is an
interface (an interface declaration, or a defined type whose underlying type is an
interface). All other exported types — structs, other named types, and type aliases
(including an alias to an interface) — MUST be counted as concrete.

#### Scenario: Interface counted as abstract

- **GIVEN** a package that declares `type Foo interface { Bar() }` as an exported type
- **WHEN** `Analyze` is called
- **THEN** the module's `AbstractTypes` MUST include this type in its count and `ExportedTypes` MUST include it as well

#### Scenario: Struct counted as concrete

- **GIVEN** a package that declares `type Baz struct { X int }` as an exported type
- **WHEN** `Analyze` is called
- **THEN** the module's `ExportedTypes` MUST include this type but `AbstractTypes` MUST NOT

#### Scenario: Unexported types excluded

- **GIVEN** a package that declares `type internal interface { foo() }`
- **WHEN** `Analyze` is called
- **THEN** neither `ExportedTypes` nor `AbstractTypes` MUST count this type

#### Scenario: Type alias counted as concrete

- **GIVEN** a package that declares `type Alias = SomeOtherType` as an exported type alias
- **WHEN** `Analyze` is called
- **THEN** `ExportedTypes` MUST include this type and `AbstractTypes` MUST NOT

#### Scenario: Empty package (no type declarations)

- **GIVEN** a package with no type declarations (only constants, variables, or functions)
- **WHEN** `Analyze` is called
- **THEN** the module's `ExportedTypes` MUST be 0 and `AbstractTypes` MUST be 0

### Requirement: Adapter computes derived metrics

The adapter MUST use `metrics.ComputeInstability`, `metrics.ComputeAbstractness`,
`metrics.ComputeDistance`, and `metrics.ComputeZone` to populate computed fields on
each `ModuleResult`. The adapter MUST NOT reimplement these computations.

#### Scenario: Derived metrics use canonical compute functions

- **GIVEN** a module with Ca=3, Ce=2, ExportedTypes=10, AbstractTypes=2
- **WHEN** `Analyze` is called
- **THEN** the `ModuleResult` MUST have Instability equal to `metrics.ComputeInstability(3, 2)`, Abstractness equal to `metrics.ComputeAbstractness(2, 10)`, Distance equal to `metrics.ComputeDistance(A, I)`, and Zone equal to `metrics.ComputeZone(A, I, D)`

### Requirement: Adapter computes LCOM4

The adapter MUST compute LCOM4 (Hitz & Montazeri 1995) for each package. LCOM4
MUST be calculated as the number of connected components in a graph where nodes are
exported methods (functions with a receiver of an exported type) and edges connect
methods that access at least one common struct field.

#### Scenario: Fully cohesive package

- **GIVEN** a package with a struct `type S struct { x int }` and three exported methods `func (s *S) A() { _ = s.x }`, `func (s *S) B() { _ = s.x }`, `func (s *S) C() { _ = s.x }` that all access field `x`
- **WHEN** `Analyze` is called
- **THEN** the module's LCOM MUST be 1

#### Scenario: Non-cohesive package

- **GIVEN** a package with `type S struct { x, y, z, w int }` and four exported methods where `A()` and `B()` access fields `x, y` and `C()` and `D()` access fields `z, w` with no overlap
- **WHEN** `Analyze` is called
- **THEN** the module's LCOM MUST be 2

#### Scenario: Package with no exported methods

- **GIVEN** a package that has no exported methods (only package-level functions or unexported methods)
- **WHEN** `Analyze` is called
- **THEN** the module's LCOM MUST be 0

#### Scenario: Package with only package-level functions (no receivers)

- **GIVEN** a package that has exported functions but no functions with a receiver
- **WHEN** `Analyze` is called
- **THEN** the module's LCOM MUST be 0 (package-level functions are excluded from the LCOM graph)

### Requirement: Adapter detects circular dependencies

The adapter MUST detect circular dependencies in the package import graph using
Tarjan's strongly connected components algorithm. Each SCC with more than one package
MUST be reported as a `metrics.Cycle` — a deterministic, lexicographically-sorted set of
the member package paths (the ordering carries no traversal meaning). The slice of
cycles is itself sorted by first element.

#### Scenario: No cycles in valid Go code

- **GIVEN** a Go module with no import cycles
- **WHEN** `Analyze` is called
- **THEN** the `Cycles` field of `ModuleGraph` MUST be an empty slice (not nil)

#### Scenario: Cycle detection in partial builds

- **GIVEN** analyzed code with partial build errors that prevent full resolution
- **WHEN** `Analyze` is called
- **THEN** the adapter MUST still attempt cycle detection on the resolvable portion and set `Status` to `StatusPartial`

### Requirement: Adapter produces schema-valid output

The `ModuleGraph` returned by `Analyze` MUST pass `metrics.Validate()` when
serialized to JSON. `SchemaVersion` MUST be `"1.1"` (reflecting the extensions addition).
`Language` MUST be `"go"`.

#### Scenario: Valid JSON output

- **GIVEN** a valid Go module
- **WHEN** `Analyze` completes successfully
- **THEN** marshaling the returned `ModuleGraph` to JSON and passing it to `metrics.Validate` MUST return nil error

#### Scenario: Complete status on success

- **GIVEN** all packages load without errors
- **WHEN** `Analyze` is called
- **THEN** `Status` MUST be `metrics.StatusComplete`

#### Scenario: Partial status on load errors

- **GIVEN** some packages fail to load due to build errors
- **WHEN** `Analyze` is called
- **THEN** `Status` MUST be `metrics.StatusPartial` and `Warnings` MUST contain entries with: (1) the affected package's import path in `Module`, (2) a machine-readable `Code` (e.g., `"load-error"`), and (3) a `Message` containing the underlying error description using relative paths

### Requirement: Adapter validates project path

The adapter MUST validate the project path before analysis using
`metrics.ValidateProjectPath`. Invalid paths (empty, path traversal, non-existent,
non-directory) MUST result in an error returned from `Analyze`.

#### Scenario: Path traversal rejected

- **GIVEN** a path containing `..` components
- **WHEN** `Analyze` is called with that path
- **THEN** the adapter MUST return an error without loading any packages

#### Scenario: Non-existent path rejected

- **GIVEN** a path that does not exist
- **WHEN** `Analyze` is called with that path
- **THEN** the adapter MUST return an error

### Requirement: Adapter uses sanitized environment

The adapter MUST set `packages.Config.Env` to a sanitized environment using
`metrics.SanitizeEnvironment` to prevent credential leakage to subprocesses
spawned by `go/packages` during package loading.

#### Scenario: Credential environment variables not leaked

- **GIVEN** the process environment contains `GITHUB_TOKEN=secret`
- **WHEN** `Analyze` is called
- **THEN** the `packages.Config.Env` MUST NOT contain `GITHUB_TOKEN`

### Requirement: Adapter returns deterministically ordered output

The adapter MUST sort the `Modules` slice in the returned `ModuleGraph` by
`Module.Path` in lexicographic order to ensure deterministic output across runs.

#### Scenario: Module ordering is deterministic

- **GIVEN** a Go module with packages A, B, C
- **WHEN** `Analyze` is called
- **THEN** the `Modules` slice MUST be ordered by `Path` lexicographically

### Requirement: Adapter propagates context for cancellation

The adapter MUST propagate the provided `context.Context` to `packages.Config.Context`.
When the context is cancelled or its deadline is exceeded, `Analyze` MUST return an
error wrapping the context error. Partial results MUST NOT be returned on cancellation.

#### Scenario: Context cancellation during analysis

- **GIVEN** a Go module with a non-trivial number of packages
- **WHEN** `Analyze` is called and the context is cancelled during `packages.Load`
- **THEN** the adapter MUST return an error wrapping `context.Canceled`

#### Scenario: Context deadline exceeded

- **GIVEN** a Go module
- **WHEN** `Analyze` is called with a context that has a deadline, and the deadline is exceeded
- **THEN** the adapter MUST return an error wrapping `context.DeadlineExceeded`

### Requirement: Adapter handles total load failure

When `packages.Load` returns zero packages, or returns packages where every package has
load/type errors OR nil type information (i.e., none can be type-checked), the adapter
MUST return an appropriate error rather than an empty or all-zeroed `ModuleGraph`.
In a partial build — where at least one package type-checks — individual packages that
have nil type information (`pkg.Types == nil`) MUST instead get ExportedTypes=0,
AbstractTypes=0, LCOM=0, and a warning. Note: when a single package with nil type
information is the only package, this yields the total-load-failure error rather than a
single zeroed-module graph.

#### Scenario: No Go files in target directory

- **GIVEN** a directory that exists but contains no Go files
- **WHEN** `Analyze` is called with that directory
- **THEN** the adapter MUST return an error

#### Scenario: Package with nil type information

- **GIVEN** a multi-package module in which one package fails to type-check (e.g., missing dependency) while at least one other package type-checks
- **WHEN** `Analyze` is called
- **THEN** the adapter MUST set ExportedTypes=0, AbstractTypes=0, LCOM=0 for the failing package, and add a warning with the package path

### Requirement: Adapter populates Go-specific extensions

The adapter MUST populate the `Extensions` field on each `ModuleResult` with
Go-specific metrics namespaced under the `go.` prefix. Extensions are language-specific
and do not modify the universal model. The adapter MUST declare the extension
capabilities as package-level constants — `CapInterfaceWidth` (`"go.interfaceWidth"`) and
`CapInterfaceProximity` (`"go.interfaceProximity"`) — and expose them via the
`ExtensionCapabilities() []string` accessor, separately from the universal
`Capabilities()` method. These constants live in the adapter package, not in the
universal `metrics` package.

#### Scenario: Extensions present in output

- **GIVEN** a Go module with packages that contain exported interfaces
- **WHEN** `Analyze` is called
- **THEN** each `ModuleResult` for packages with exported interfaces MUST have an `Extensions` map containing `go.interfaceWidth` and `go.interfaceProximity` keys

#### Scenario: No extensions for packages without interfaces

- **GIVEN** a package with no exported interfaces
- **WHEN** `Analyze` is called
- **THEN** the `Extensions` field for that module MAY be nil or omitted (no `go.interfaceWidth` or `go.interfaceProximity` keys)

### Requirement: Interface Width (Pike metric)

The adapter MUST compute interface width for all exported interfaces in each analyzed
package. Interface width is the method count of the interface. The result MUST be
stored in `Extensions["go.interfaceWidth"]` as a `map[string]int` mapping interface
name to method count.

#### Scenario: Single-method interface

- **GIVEN** a package with `type Reader interface { Read(p []byte) (n int, err error) }`
- **WHEN** `Analyze` is called
- **THEN** `Extensions["go.interfaceWidth"]` MUST contain `{"Reader": 1}`

#### Scenario: Multi-method interface

- **GIVEN** a package with `type ReadWriter interface { Read(p []byte) (n int, err error); Write(p []byte) (n int, err error) }`
- **WHEN** `Analyze` is called
- **THEN** `Extensions["go.interfaceWidth"]` MUST contain `{"ReadWriter": 2}`

#### Scenario: Embedded interface method counting

- **GIVEN** a package with `type Closer interface { Close() error }` and `type ReadCloser interface { Reader; Closer }`
- **WHEN** `Analyze` is called
- **THEN** `Extensions["go.interfaceWidth"]` MUST count the total flattened method set (e.g., `{"Closer": 1, "ReadCloser": 2}`)

#### Scenario: No exported interfaces

- **GIVEN** a package with no exported interfaces (only structs, functions, etc.)
- **WHEN** `Analyze` is called
- **THEN** `Extensions["go.interfaceWidth"]` MUST NOT be present (or the `Extensions` map itself may be nil)

### Requirement: Interface-to-Implementation Proximity

The adapter MUST compute interface proximity for all exported interfaces in each
analyzed package. Proximity indicates whether the interface is declared on the
consumer side (`"consumer"`) or producer side (`"producer"`). An interface is
`"producer"` if any type in the same package implements it; otherwise it is
`"consumer"`. The result MUST be stored in `Extensions["go.interfaceProximity"]`
as a `map[string]string`.

#### Scenario: Producer-side interface

- **GIVEN** a package with `type Saver interface { Save() error }` and `type FileSaver struct{}` with `func (f *FileSaver) Save() error { ... }` in the same package
- **WHEN** `Analyze` is called
- **THEN** `Extensions["go.interfaceProximity"]` MUST contain `{"Saver": "producer"}`

#### Scenario: Consumer-side interface

- **GIVEN** a package with `type Logger interface { Log(msg string) }` where no type in the same package implements `Logger`
- **WHEN** `Analyze` is called
- **THEN** `Extensions["go.interfaceProximity"]` MUST contain `{"Logger": "consumer"}`

### Requirement: Typed extension accessors for JSON round-trip safety

The adapter package MUST provide typed accessor functions for extracting extension
values after JSON round-trip. JSON unmarshaling converts `int` to `float64` and
`map[string]int` to `map[string]interface{}`. Consumers MUST NOT need unchecked
type assertions.

#### Scenario: Interface widths extracted after JSON round-trip

- **GIVEN** a `ModuleResult` with `Extensions["go.interfaceWidth"]` populated as `map[string]int`
- **WHEN** the `ModuleResult` is marshaled to JSON and unmarshaled back
- **THEN** calling the typed accessor function (e.g., `InterfaceWidths(extensions)`) MUST return the correct `map[string]int` values without panic

#### Scenario: Interface proximity extracted after JSON round-trip

- **GIVEN** a `ModuleResult` with `Extensions["go.interfaceProximity"]` populated as `map[string]string`
- **WHEN** the `ModuleResult` is marshaled to JSON and unmarshaled back
- **THEN** calling the typed accessor function (e.g., `InterfaceProximities(extensions)`) MUST return the correct `map[string]string` values without panic

#### Scenario: Accessor returns error on missing extension key

- **GIVEN** a `ModuleResult` with nil or empty `Extensions`
- **WHEN** the typed accessor function is called
- **THEN** it MUST return a zero-value map and an error (not panic)

#### Scenario: Mixed proximity

- **GIVEN** a package with two exported interfaces where one has an in-package implementation and the other does not
- **WHEN** `Analyze` is called
- **THEN** `Extensions["go.interfaceProximity"]` MUST contain entries with mixed `"producer"` and `"consumer"` values
