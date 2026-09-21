## ADDED Requirements

### Requirement: Provenance object on ModuleGraph

The `ModuleGraph` SHALL support an optional top-level `provenance` object carrying
producer name, producer version, generation timestamp, and input identity. The
`metrics` package SHALL expose a `Provenance` type and a `ProvenanceInput` type, and
`ModuleGraph` SHALL have a `Provenance *Provenance` field tagged
`json:"provenance,omitempty"` so the field is absent when nil. The metric payload
(modules, cycles, warnings, status, and schema version) SHALL be byte-identical
regardless of whether provenance is present, except for the `provenance` field
itself.

#### Scenario: Provenance omitted when nil
- **WHEN** a `ModuleGraph` is marshaled with a nil `Provenance` field
- **THEN** the JSON output MUST NOT contain a `provenance` key

#### Scenario: Provenance present when populated
- **WHEN** a `ModuleGraph` is marshaled with a populated `Provenance` field
- **THEN** the JSON output MUST contain a `provenance` object with `producer`,
  `version`, `generatedAt`, and `input` fields

### Requirement: Provenance field shape

The `provenance` object SHALL contain the following fields:

- `producer` (string): the producing tool name, the constant `"vibe-check"`.
- `version` (string): the tool version from build info / ldflags.
- `generatedAt` (string): an RFC3339 UTC timestamp.
- `input` (object): input identity with `path` (string, the analyzed path) and
  `modulePath` (string, the resolved module path).

#### Scenario: Producer field is vibe-check
- **WHEN** `vibe-check analyze` emits a `provenance` object
- **THEN** `provenance.producer` MUST equal `"vibe-check"`

#### Scenario: GeneratedAt is RFC3339 UTC
- **WHEN** `vibe-check analyze` emits a `provenance` object
- **THEN** `provenance.generatedAt` MUST parse as an RFC3339 timestamp in UTC

#### Scenario: Input identity is populated
- **WHEN** `vibe-check analyze /path/to/project` is invoked against a Go module
  resolving to module path `example.com/mymod`
- **THEN** `provenance.input.path` MUST be the analyzed path and
  `provenance.input.modulePath` MUST be `example.com/mymod`

### Requirement: Schema version 1.2

The `ModuleGraph` schema version SHALL be bumped from `"1.1"` to `"1.2"` as a
backward-compatible additive change. The `schemaVersion` enum SHALL accept
`"1.0"`, `"1.1"`, and `"1.2"`. The schema SHALL declare an optional `provenance`
object property and continue to reject unknown properties (`additionalProperties:
false`).

#### Scenario: Current schema version is 1.2
- **WHEN** `vibe-check analyze` emits a `ModuleGraph`
- **THEN** `schemaVersion` MUST equal `"1.2"`

#### Scenario: Older versions still validate
- **WHEN** a `ModuleGraph` with `schemaVersion` `"1.0"` or `"1.1"` and no
  `provenance` field is validated
- **THEN** validation MUST succeed

### Requirement: Provenance validation

`metrics.Validate` SHALL accept an optional `provenance` object when present. If
`provenance` is present, the validator SHALL require `producer` and `version` to be
strings and `input` to be an object with string `path` and `modulePath` fields. A
malformed `provenance` value SHALL fail validation.

#### Scenario: Valid provenance accepted
- **WHEN** a `ModuleGraph` with a well-formed `provenance` object is validated
- **THEN** validation MUST succeed

#### Scenario: Malformed provenance rejected
- **WHEN** a `ModuleGraph` whose `provenance` is a string or array is validated
- **THEN** validation MUST fail

### Requirement: Reproducible output flag

The `analyze` command SHALL accept a `--no-provenance` boolean flag. When set, the
command SHALL omit the `provenance` object from output entirely, so that the
metric payload is byte-reproducible across runs. When unset (default), the
`provenance` object SHALL be included.

#### Scenario: Provenance omitted with flag
- **WHEN** `vibe-check analyze --no-provenance` is invoked
- **THEN** the JSON output MUST NOT contain a `provenance` key

#### Scenario: Provenance included by default
- **WHEN** `vibe-check analyze` is invoked without `--no-provenance`
- **THEN** the JSON output MUST contain a `provenance` object

### Requirement: Determinism of metric payload

The metric payload (everything except the `provenance` field) SHALL be
byte-identical across two runs against the same input, regardless of generation
time. Provenance is treated as metadata, not a metric, and MUST NOT affect any
computed metric value.

#### Scenario: Metric payload deterministic with provenance suppressed
- **WHEN** `vibe-check analyze --no-provenance` is run twice against the same module
- **THEN** the JSON output MUST be byte-identical between runs

#### Scenario: Diff ignores provenance
- **WHEN** `vibe-check diff` compares two graphs that differ only in their
  `provenance` fields
- **THEN** the computed delta and verdict MUST be identical to comparing the same
  graphs without provenance
