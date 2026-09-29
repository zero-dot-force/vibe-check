## ADDED Requirements

### Requirement: Provenance object on diff payload

The `diff` command's `--json` payload SHALL include an additive top-level
`provenance` object carrying producer name, producer version, and generation
timestamp. All existing diff payload keys SHALL remain unchanged. The diff
metric payload (verdict, reasons, entropyDirection, unreliable, modules, added,
removed, newCycles, resolvedCycles) SHALL be byte-identical regardless of
whether the `provenance` object is present, except for the `provenance` field
itself.

#### Scenario: Provenance included on diff by default
- **WHEN** `vibe-check diff base.json pr.json --json` is invoked without `--no-provenance`
- **THEN** the JSON output MUST contain a top-level `provenance` object

#### Scenario: Provenance omitted on diff with flag
- **WHEN** `vibe-check diff base.json pr.json --json --no-provenance` is invoked
- **THEN** the JSON output MUST NOT contain a `provenance` key

#### Scenario: Existing diff keys preserved
- **WHEN** `vibe-check diff base.json pr.json --json` is invoked
- **THEN** the JSON output MUST contain the same `verdict`, `reasons`,
  `entropyDirection`, and `modules` keys as before this change, with unchanged
  values

### Requirement: Provenance object on init payload

The `init` command's `--json` payload SHALL include an additive top-level
`provenance` object carrying producer name, producer version, and generation
timestamp. All existing init payload keys (`written`, `skipped`, `forced`)
SHALL remain unchanged.

#### Scenario: Provenance included on init by default
- **WHEN** `vibe-check init . --json` is invoked without `--no-provenance`
- **THEN** the JSON output MUST contain a top-level `provenance` object

#### Scenario: Provenance omitted on init with flag
- **WHEN** `vibe-check init . --json --no-provenance` is invoked
- **THEN** the JSON output MUST NOT contain a `provenance` key

#### Scenario: Existing init keys preserved
- **WHEN** `vibe-check init . --json` is invoked
- **THEN** the JSON output MUST contain the same `written`, `skipped`, and
  `forced` keys as before this change, with unchanged values

### Requirement: Consistent provenance envelope field shape

The `provenance` object emitted by `diff`, `init`, and `analyze` SHALL use a
consistent field shape. It SHALL contain:
- `producer` (string): the producing tool name, the constant `"vibe-check"`.
- `version` (string): the tool version from build info / ldflags.
- `generatedAt` (string): an RFC3339 UTC timestamp.

The producer constant, version resolution, and timestamp format SHALL be
provided by a single shared helper so the three commands cannot drift. The
`analyze` envelope additionally carries an `input` sub-object (analyzed path +
resolved module path); the `diff` and `init` envelopes carry only the three
shared fields above.

#### Scenario: Producer field is vibe-check
- **WHEN** `vibe-check diff` or `vibe-check init` emits a `provenance` object
- **THEN** `provenance.producer` MUST equal `"vibe-check"`

#### Scenario: GeneratedAt is RFC3339 UTC
- **WHEN** `vibe-check diff` or `vibe-check init` emits a `provenance` object
- **THEN** `provenance.generatedAt` MUST parse as an RFC3339 timestamp in UTC

#### Scenario: Version is non-empty
- **WHEN** `vibe-check diff` or `vibe-check init` emits a `provenance` object
- **THEN** `provenance.version` MUST be a non-empty string

### Requirement: Reproducible output flags

The `diff` and `init` commands SHALL each accept a `--no-provenance` boolean
flag. When set, the command SHALL omit the `provenance` object from output
entirely, so that the payload is byte-reproducible across runs. When unset
(default), the `provenance` object SHALL be included.

#### Scenario: Diff reproducible with flag
- **WHEN** `vibe-check diff base.json pr.json --json --no-provenance` is run twice
  against the same inputs
- **THEN** the JSON output MUST be byte-identical between runs

#### Scenario: Init reproducible with flag
- **WHEN** `vibe-check init . --json --no-provenance` is run twice against the
  same path
- **THEN** the JSON output MUST be byte-identical between runs

### Requirement: Provenance is metadata not metric

Provenance SHALL be treated as metadata, not a metric, and MUST NOT affect any
computed metric value or the diff verdict. The diff metric payload SHALL be
deterministic across two runs against the same inputs, with the sole exception
of the `provenance.generatedAt` field when provenance is enabled.

#### Scenario: Diff verdict unaffected by provenance
- **WHEN** `vibe-check diff` compares two graph snapshots, with and without
  `--no-provenance`
- **THEN** the `verdict`, `reasons`, `entropyDirection`, and `modules` values
  MUST be identical in both cases
