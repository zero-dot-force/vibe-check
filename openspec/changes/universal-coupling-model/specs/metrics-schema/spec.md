## ADDED Requirements

### Requirement: JSON schema for ModuleGraph

The system SHALL define a JSON schema for the `ModuleGraph` structure used as the interchange format between adapters and the core engine. The schema SHALL be the authoritative definition for serialization and deserialization of analysis results.

#### Scenario: Valid ModuleGraph serialization
- **GIVEN** a ModuleGraph with all fields populated
- **WHEN** it is serialized to JSON
- **THEN** the output SHALL conform to the defined JSON schema

#### Scenario: Schema validation of external input
- **GIVEN** an external analyzer returns a JSON result
- **WHEN** the host receives the result
- **THEN** the host SHALL validate the result against the JSON schema before processing

### Requirement: Schema version

The JSON schema SHALL include a required `schemaVersion` field (string) at the top level of the ModuleGraph. The initial schema version SHALL be `"1.0"`. The version MUST follow semantic versioning. The host MUST check the schema version before processing the result and return a descriptive error if the version is unsupported.

#### Scenario: Schema version present
- **GIVEN** a ModuleGraph is constructed
- **WHEN** it is serialized to JSON
- **THEN** the JSON SHALL contain a `schemaVersion` field with value `"1.0"`

#### Scenario: Unsupported schema version
- **GIVEN** an external analyzer returns a result with `schemaVersion: "2.0"`
- **WHEN** the host processes the result
- **THEN** the host SHALL return a descriptive error indicating the schema version is not supported

### Requirement: Language field

The JSON schema SHALL include a required `language` field (string) at the top level of the ModuleGraph. The value MUST be a lowercase language identifier (e.g., "go", "python", "typescript").

#### Scenario: Language field present
- **GIVEN** a ModuleGraph is constructed
- **WHEN** it is serialized to JSON
- **THEN** the JSON SHALL contain a `language` field with a non-empty string value

#### Scenario: Language field validation
- **GIVEN** an external analyzer returns a result without a `language` field
- **WHEN** schema validation runs
- **THEN** schema validation SHALL fail

### Requirement: Warnings array

The JSON schema SHALL include a `warnings` array at the top level of the ModuleGraph. Each warning SHALL be an object with `code` (string), `message` (string), and optional `module` (string, the affected module path) fields. The array MUST be present even when empty (not omitted or null).

#### Scenario: Warnings present
- **GIVEN** an analysis produces warnings
- **WHEN** the results are serialized to JSON
- **THEN** the JSON SHALL contain a `warnings` array with warning objects

#### Scenario: No warnings
- **GIVEN** an analysis produces no warnings
- **WHEN** the results are serialized to JSON
- **THEN** the JSON SHALL contain an empty `warnings` array (`[]`)

#### Scenario: Warning structure
- **GIVEN** a warning is included in the analysis results
- **WHEN** it is serialized to JSON
- **THEN** it SHALL have at minimum `code` and `message` string fields

### Requirement: Module metrics structure

The JSON schema SHALL represent each module's metrics as an object with fields: `path` (string), `name` (string), `ca` (integer), `ce` (integer), `instability` (number), `abstractness` (number), `distance` (number), `lcom` (integer), `exportedTypes` (integer), and `abstractTypes` (integer). All numeric metric fields MUST be present (no omission for zero values).

#### Scenario: Complete module metrics
- **GIVEN** a module has all metrics computed
- **WHEN** the module's metrics are serialized to JSON
- **THEN** all metric fields SHALL be present with their computed values, including both `exportedTypes` and `abstractTypes`

#### Scenario: Zero values not omitted
- **GIVEN** a module has Ca = 0 and Ce = 0
- **WHEN** the module's metrics are serialized to JSON
- **THEN** the JSON SHALL include `"ca": 0` and `"ce": 0` (not omitted)

### Requirement: Cycles representation

The JSON schema SHALL represent circular dependencies as a `cycles` array. Each cycle SHALL be an array of module path strings representing the ordered dependency chain, starting from the lexicographically smallest module path. The `cycles` array SHALL be sorted lexicographically by the first element of each cycle. The array MUST be present even when empty.

#### Scenario: Cycles present
- **GIVEN** a circular dependency A → B → C → A exists
- **WHEN** the results are serialized to JSON
- **THEN** the `cycles` array SHALL contain `[["A", "B", "C"]]`

#### Scenario: No cycles
- **GIVEN** no circular dependencies exist
- **WHEN** the results are serialized to JSON
- **THEN** the `cycles` array SHALL be empty (`[]`)

### Requirement: Zone classification

The JSON schema SHALL include a `zone` field for each module indicating its position relative to the main sequence. Valid zone values SHALL be: `main-sequence` (D < 0.2), `zone-of-pain` (A < 0.2 and I < 0.2), `zone-of-uselessness` (A > 0.8 and I > 0.8), and `normal` (all other cases). Zone classification precedence: `main-sequence` is evaluated first, then `zone-of-pain`, then `zone-of-uselessness`, then `normal`. The zone thresholds are defined in the schema as default constants.

#### Scenario: Module on main sequence
- **GIVEN** a module has D = 0.1
- **WHEN** zone classification is applied
- **THEN** its zone SHALL be `main-sequence`

#### Scenario: Module in zone of pain
- **GIVEN** a module has A = 0.0, I = 0.0, and D = 1.0
- **WHEN** zone classification is applied
- **THEN** its zone SHALL be `zone-of-pain`

#### Scenario: Module in zone of uselessness
- **GIVEN** a module has A = 1.0 and I = 1.0
- **WHEN** zone classification is applied
- **THEN** its zone SHALL be `zone-of-uselessness`

#### Scenario: Normal module
- **GIVEN** a module has D = 0.5 and does not qualify for any special zone
- **WHEN** zone classification is applied
- **THEN** its zone SHALL be `normal`

#### Scenario: Overlapping zone criteria — main-sequence takes precedence
- **GIVEN** a module has A = 0.1, I = 0.1, and D = 0.2 (qualifies for both main-sequence boundary and zone-of-pain)
- **WHEN** zone classification is applied
- **THEN** D is NOT < 0.2, so it SHALL be classified as `zone-of-pain`

### Requirement: Status metadata

The JSON schema SHALL include a top-level `status` field indicating the overall analysis outcome. Valid values SHALL be: `complete` (all metrics computed successfully), `partial` (some metrics unavailable due to adapter limitations), and `error` (analysis failed with partial or no results). When status is `partial`, the `warnings` array MUST contain entries explaining which metrics are unavailable and why.

#### Scenario: Complete analysis
- **GIVEN** all metrics are computed successfully
- **WHEN** the results are serialized
- **THEN** status SHALL be `complete`

#### Scenario: Partial analysis
- **GIVEN** an adapter does not support LCOM computation
- **WHEN** the results are serialized
- **THEN** status SHALL be `partial` and warnings SHALL explain that LCOM is unavailable

#### Scenario: Error status
- **GIVEN** analysis encounters a fatal error but produces partial results
- **WHEN** the results are serialized
- **THEN** status SHALL be `error` with warnings describing the failure

## MODIFIED Requirements

<!-- None — greenfield project -->

## REMOVED Requirements

<!-- None — greenfield project -->
