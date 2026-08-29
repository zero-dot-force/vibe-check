## ADDED Requirements

### Requirement: Adapter interface definition

The system SHALL define an `Adapter` interface with the following methods:
- `Analyze(ctx context.Context, projectPath string) (*ModuleGraph, error)` — analyzes the project at the given path and returns a complete module graph
- `Language() string` — returns the lowercase language identifier (e.g., "go", "python", "typescript")

All language adapters MUST implement this interface.

#### Scenario: Go adapter implements interface
- **GIVEN** a Go adapter type is defined
- **WHEN** it is compiled
- **THEN** it SHALL satisfy the `Adapter` interface at compile time (verified via `var _ Adapter = (*GoAdapter)(nil)`)

#### Scenario: Analyze returns complete graph
- **GIVEN** a valid project path exists
- **WHEN** `Analyze` is called with the project path
- **THEN** it SHALL return a `ModuleGraph` with all modules and their metrics computed

#### Scenario: Analyze with context cancellation
- **GIVEN** an adapter is analyzing a project
- **WHEN** the context is cancelled during analysis
- **THEN** the adapter SHALL return a context cancellation error without panic

#### Scenario: Analyze with invalid path
- **GIVEN** a project path that does not exist
- **WHEN** `Analyze` is called with that path
- **THEN** the adapter SHALL return a descriptive error wrapping the underlying cause

### Requirement: Adapter registration

The system SHALL provide a registration mechanism for adapters. Registration MUST associate a language identifier (string) with an `Adapter` implementation. The registry MUST NOT use global mutable state — it SHALL be a value that is passed via dependency injection.

#### Scenario: Register and retrieve adapter
- **GIVEN** an empty registry
- **WHEN** a Go adapter is registered with language "go"
- **THEN** the registry SHALL return that adapter when queried for language "go"

#### Scenario: Duplicate registration
- **GIVEN** a registry with a Go adapter registered for language "go"
- **WHEN** a second adapter is registered for language "go"
- **THEN** the registry SHALL return an error on the second registration

#### Scenario: Unknown language lookup
- **GIVEN** a registry with no adapter registered for language "rust"
- **WHEN** the registry is queried for language "rust"
- **THEN** it SHALL return a descriptive error (not nil)

### Requirement: Adapter capability discovery

The system SHALL allow adapters to declare their capabilities via a `Capabilities() []Capability` method on the `Adapter` interface. At minimum, each adapter MUST declare which metrics it can compute. When an adapter does not support a metric, the corresponding metric value in the `ModuleResult` SHALL use its zero value and the `ModuleGraph.Status` SHALL be `partial` with a warning explaining which metrics are unavailable.

#### Scenario: Full capability adapter
- **GIVEN** an adapter declares support for all metrics (Ca, Ce, I, A, D, LCOM, circular deps)
- **WHEN** it is queried for capabilities
- **THEN** it SHALL return a list containing all metric identifiers

#### Scenario: Partial capability adapter
- **GIVEN** an adapter declares support for coupling metrics (Ca, Ce, I) but not cohesion (LCOM)
- **WHEN** the adapter produces a ModuleGraph
- **THEN** LCOM fields SHALL use zero values, Status SHALL be `partial`, and Warnings SHALL explain that LCOM is unavailable

#### Scenario: Capability query
- **GIVEN** an adapter is registered
- **WHEN** the system queries the adapter's capabilities
- **THEN** the adapter SHALL return a list of supported metric identifiers

### Requirement: Adding a new language adapter

Adding a new language adapter MUST NOT require changes to the core `metrics` package or to any existing adapter. The new adapter MUST only need to implement the `Adapter` interface and register itself with the registry.

#### Scenario: New adapter integration
- **GIVEN** an existing system with a Go adapter registered
- **WHEN** a developer creates a new adapter for language "rust" implementing the `Adapter` interface
- **THEN** registering it with the registry SHALL make it available for analysis without modifying any existing code

#### Scenario: No core package changes
- **GIVEN** a new language adapter is added
- **WHEN** the `metrics` package source is compared before and after
- **THEN** the `metrics` package SHALL have zero diff (no modifications required)

## MODIFIED Requirements

<!-- None — greenfield project -->

## REMOVED Requirements

<!-- None — greenfield project -->
