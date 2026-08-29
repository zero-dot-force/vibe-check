## ADDED Requirements

### Requirement: Module identity

The system SHALL represent each unit of analysis as a `Module` with a unique string identifier (the module path) and a human-readable name. Module paths MUST be unique within a single analysis result. The `Module` type SHALL contain raw metric input data: `Path` (string), `Name` (string), `Ca` (int), `Ce` (int), `ExportedTypes` (int), `AbstractTypes` (int).

#### Scenario: Distinct modules have distinct paths
- **GIVEN** an adapter is analyzing a project containing two distinct packages
- **WHEN** the analysis completes
- **THEN** each module in the result SHALL have a unique `Path` value

#### Scenario: Module path preserves language convention
- **GIVEN** a Go adapter is analyzing a project with package `github.com/foo/bar`
- **WHEN** the analysis completes
- **THEN** the module path SHALL be `github.com/foo/bar`

### Requirement: ModuleResult structure

The system SHALL define a `ModuleResult` type that combines `Module` identity data with computed metrics (Instability, Abstractness, Distance, LCOM) and Zone classification. `ModuleResult` SHALL embed `Module` (providing raw data) and add computed metric fields. The `ModuleGraph.Modules` slice SHALL contain `ModuleResult` entries, ensuring each entry carries both raw input data and computed output values.

#### Scenario: ModuleResult contains both raw and computed data
- **GIVEN** a module with Ca=3, Ce=7, ExportedTypes=5, AbstractTypes=2
- **WHEN** a ModuleResult is constructed for this module
- **THEN** the ModuleResult SHALL contain the raw Module fields AND computed Instability=0.7, Abstractness=0.4, Distance=0.1

### Requirement: Afferent coupling metric

The system SHALL compute Afferent Coupling (Ca) for each module as the count of other modules that depend on it. Ca MUST be a non-negative integer. A module with no dependents SHALL have Ca = 0.

#### Scenario: Module with no dependents
- **GIVEN** a project is analyzed
- **WHEN** a module is not imported by any other module in the project
- **THEN** Ca SHALL equal 0

#### Scenario: Module with multiple dependents
- **GIVEN** a project is analyzed
- **WHEN** module A is imported by modules B, C, and D
- **THEN** Ca for module A SHALL equal 3

#### Scenario: Deterministic Ca computation
- **GIVEN** a project is analyzed twice with no changes between runs
- **WHEN** the results are compared
- **THEN** Ca values for all modules SHALL be identical in both results

### Requirement: Efferent coupling metric

The system SHALL compute Efferent Coupling (Ce) for each module as the count of other modules it depends on. Ce MUST be a non-negative integer. A module with no dependencies SHALL have Ce = 0.

#### Scenario: Module with no dependencies
- **GIVEN** a project is analyzed
- **WHEN** a module does not import any other module in the project
- **THEN** Ce SHALL equal 0

#### Scenario: Module with multiple dependencies
- **GIVEN** a project is analyzed
- **WHEN** module A imports modules B, C, and D
- **THEN** Ce for module A SHALL equal 3

#### Scenario: Deterministic Ce computation
- **GIVEN** a project is analyzed twice with no changes between runs
- **WHEN** the results are compared
- **THEN** Ce values for all modules SHALL be identical in both results

### Requirement: Instability metric

The system SHALL compute Instability (I) for each module using the formula I = Ce / (Ca + Ce). I MUST be a float64 in the range [0.0, 1.0]. When both Ca and Ce are 0, I SHALL be 0.0 (maximally stable by convention).

#### Scenario: Maximally stable module
- **GIVEN** a project is analyzed
- **WHEN** a module has Ca = 5 and Ce = 0
- **THEN** I SHALL equal 0.0

#### Scenario: Maximally unstable module
- **GIVEN** a project is analyzed
- **WHEN** a module has Ca = 0 and Ce = 5
- **THEN** I SHALL equal 1.0

#### Scenario: Mixed coupling
- **GIVEN** a project is analyzed
- **WHEN** a module has Ca = 3 and Ce = 7
- **THEN** I SHALL equal 0.7

#### Scenario: Isolated module
- **GIVEN** a project is analyzed
- **WHEN** a module has Ca = 0 and Ce = 0
- **THEN** I SHALL equal 0.0

#### Scenario: Deterministic Instability computation
- **GIVEN** a project is analyzed twice with no changes between runs
- **WHEN** the results are compared
- **THEN** Instability values for all modules SHALL be identical in both results

### Requirement: Abstractness metric

The system SHALL compute Abstractness (A) for each module as the ratio of abstract types to total exported types. A MUST be a float64 in the range [0.0, 1.0]. When a module has no exported types, A SHALL be 0.0. An abstract type is a type that cannot be directly instantiated and serves as a contract for implementations. Each language adapter MUST document its mapping from language-specific constructs to the abstract/concrete classification (e.g., Go interfaces, Python ABCs, TypeScript abstract classes/interfaces).

#### Scenario: Fully abstract module
- **GIVEN** a project is analyzed
- **WHEN** a module exports 3 abstract types and 0 concrete types (abstractTypes=3, exportedTypes=3)
- **THEN** A SHALL equal 1.0

#### Scenario: Fully concrete module
- **GIVEN** a project is analyzed
- **WHEN** a module exports 0 abstract types and 5 concrete types (abstractTypes=0, exportedTypes=5)
- **THEN** A SHALL equal 0.0

#### Scenario: Mixed module
- **GIVEN** a project is analyzed
- **WHEN** a module exports 2 abstract types and 3 concrete types (abstractTypes=2, exportedTypes=5)
- **THEN** A SHALL equal 0.4

#### Scenario: No exported types
- **GIVEN** a project is analyzed
- **WHEN** a module exports no types (exportedTypes=0)
- **THEN** A SHALL equal 0.0

#### Scenario: Deterministic Abstractness computation
- **GIVEN** a project is analyzed twice with no changes between runs
- **WHEN** the results are compared
- **THEN** Abstractness values for all modules SHALL be identical in both results

### Requirement: Distance from main sequence metric

The system SHALL compute Distance from Main Sequence (D) for each module using the formula D = |A + I - 1|. D MUST be a float64 in the range [0.0, 1.0]. A value of 0.0 indicates the module lies on the main sequence.

#### Scenario: Module on the main sequence
- **GIVEN** a project is analyzed
- **WHEN** a module has A = 0.5 and I = 0.5
- **THEN** D SHALL equal 0.0

#### Scenario: Zone of pain
- **GIVEN** a project is analyzed
- **WHEN** a module has A = 0.0 and I = 0.0
- **THEN** D SHALL equal 1.0

#### Scenario: Zone of uselessness
- **GIVEN** a project is analyzed
- **WHEN** a module has A = 1.0 and I = 1.0
- **THEN** D SHALL equal 1.0

#### Scenario: Deterministic Distance computation
- **GIVEN** a project is analyzed twice with no changes between runs
- **WHEN** the results are compared
- **THEN** Distance values for all modules SHALL be identical in both results

### Requirement: Cohesion metric

The system SHALL compute Lack of Cohesion of Methods using the LCOM4 variant (Hitz & Montazeri, 1995). LCOM4 counts the number of connected components in the method-field graph, where methods are connected if they access at least one common field. LCOM MUST be a non-negative integer. LCOM = 1 indicates a fully cohesive module (all methods form a single connected component). LCOM = 0 indicates a module with no methods or fields (trivially cohesive). LCOM > 1 indicates the module could be split into LCOM independent classes. The "fields" concept maps to language-specific shared state: struct fields in Go, instance attributes in Python, class properties in TypeScript. Each adapter MUST document its mapping. Limitation: LCOM4 does not account for method call chains — two methods that share no fields but call each other are treated as disconnected.

#### Scenario: Fully cohesive module
- **GIVEN** a project is analyzed
- **WHEN** a module has 5 methods and all 5 access at least one common field (forming a single connected component)
- **THEN** LCOM SHALL be 1

#### Scenario: Trivially cohesive module
- **GIVEN** a project is analyzed
- **WHEN** a module has no methods or no fields
- **THEN** LCOM SHALL be 0

#### Scenario: Disjoint groups
- **GIVEN** a project is analyzed
- **WHEN** a module contains two groups of functions that share no common types or state
- **THEN** LCOM SHALL be 2

#### Scenario: Deterministic LCOM computation
- **GIVEN** a project is analyzed twice with no changes between runs
- **WHEN** the results are compared
- **THEN** LCOM values for all modules SHALL be identical in both results

### Requirement: Circular dependency detection

The system SHALL detect circular dependencies between modules. Each cycle SHALL be represented as an ordered list of module paths forming the cycle, starting from the lexicographically smallest module path. The `Cycles` slice SHALL be sorted lexicographically by the first element of each cycle. The system MUST handle circular dependencies without infinite loops or panics.

#### Scenario: No circular dependencies
- **GIVEN** a project with no circular dependencies is analyzed
- **WHEN** the analysis completes
- **THEN** the cycles list SHALL be empty

#### Scenario: Simple circular dependency
- **GIVEN** a project is analyzed
- **WHEN** module A depends on module B and module B depends on module A
- **THEN** the cycles list SHALL contain one cycle with path [A, B]

#### Scenario: Complex circular dependency
- **GIVEN** a project is analyzed
- **WHEN** module A depends on B, B depends on C, and C depends on A
- **THEN** the cycles list SHALL contain one cycle with path [A, B, C]

#### Scenario: Canonical cycle ordering
- **GIVEN** a project is analyzed
- **WHEN** a cycle exists between modules C, A, and B (C→A→B→C)
- **THEN** the cycle SHALL be represented as [A, B, C] (starting from the lexicographically smallest path)

#### Scenario: Multiple cycles sorted
- **GIVEN** a project is analyzed
- **WHEN** two independent cycles exist: D→E→D and A→B→A
- **THEN** the cycles list SHALL be [[A, B], [D, E]] (sorted by first element)

#### Scenario: Termination guarantee
- **GIVEN** a project contains circular dependencies of any depth
- **WHEN** the analysis runs
- **THEN** the cycle detection algorithm SHALL terminate without panic and produce a result

#### Scenario: Deterministic cycle detection
- **GIVEN** a project with circular dependencies is analyzed twice with no changes between runs
- **WHEN** the results are compared
- **THEN** the cycles list SHALL be identical in both results (same cycles, same order)

### Requirement: Module graph structure

The system SHALL represent the complete analysis result as a `ModuleGraph` containing all modules (as `ModuleResult` entries), their computed metrics, and detected cycles. The graph MUST include a `Language` field identifying which language adapter produced the result.

#### Scenario: Complete graph output
- **GIVEN** an adapter analyzes a project with 5 modules
- **WHEN** the analysis completes
- **THEN** the ModuleGraph SHALL contain exactly 5 ModuleResult entries with all metrics computed

#### Scenario: Language identification
- **GIVEN** a Go adapter produces a ModuleGraph
- **WHEN** the analysis completes
- **THEN** the Language field SHALL equal "go"

### Requirement: Warnings in analysis results

The system SHALL support a `Warnings` slice in the `ModuleGraph` for language-specific caveats that may affect metric accuracy. Warnings MUST NOT prevent metric computation — they annotate results with context.

#### Scenario: Warning for dynamic imports
- **GIVEN** a Python adapter encounters dynamic imports that cannot be statically analyzed
- **WHEN** the analysis completes
- **THEN** the ModuleGraph SHALL include a warning describing the limitation

#### Scenario: No warnings
- **GIVEN** analysis completes without caveats
- **WHEN** the results are inspected
- **THEN** the Warnings slice SHALL be empty (not nil)

## MODIFIED Requirements

<!-- None — greenfield project -->

## REMOVED Requirements

<!-- None — greenfield project -->
