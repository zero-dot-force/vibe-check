## ADDED Requirements

### Requirement: README documents the provenance object

`README.md` SHALL document the top-level `provenance` object in its Output section,
naming three core fields — `producer`, `version`, `generatedAt` — plus an `input`
sub-object (with
`path` and `modulePath`). The description SHALL note that `producer` is the constant
`"vibe-check"`, `version` is the tool version, `generatedAt` is an RFC 3339 UTC
timestamp, and `input` records the analyzed path and resolved module path.

#### Scenario: Output section describes provenance fields
- **WHEN** the Output section of `README.md` is read
- **THEN** it names `producer`, `version`, `generatedAt`, and `input`, states
  that `producer` is the constant `"vibe-check"`, that `generatedAt` is an RFC
  3339 UTC timestamp, and that `input` contains `path` and `modulePath`

### Requirement: README flag table lists the no-provenance flag

`README.md`'s flag table SHALL list `--no-provenance` and state its purpose: to omit
the `provenance` object from output for byte-reproducible JSON.

#### Scenario: Flag table documents byte reproducibility
- **WHEN** the flag table in `README.md` is read
- **THEN** it contains a `--no-provenance` row whose description mentions
  byte-reproducible output or omitting the `provenance` object

### Requirement: README JSON example reflects schema 1.2 and provenance

`README.md`'s JSON example SHALL set `schemaVersion` to `"1.2"` and SHALL include an
example `provenance` object with `producer`, `version`, `generatedAt`, and `input`
(`path` and `modulePath`) fields.

#### Scenario: Example shows schema 1.2
- **WHEN** the JSON example in `README.md` is read
- **THEN** its `schemaVersion` value is `"1.2"`

#### Scenario: Example includes a provenance object
- **WHEN** the JSON example in `README.md` is read
- **THEN** it includes a `provenance` object containing `producer`, `version`,
  `generatedAt`, and `input` (with `path` and `modulePath`)

### Requirement: README documents provenance for diff and init

`README.md`'s `vibe-check diff` section SHALL state that `diff --json` emits a
top-level `provenance` object carrying `producer`, `version`, and `generatedAt`
(no `input` sub-object) and accepts a `--no-provenance` flag. `README.md`'s
`vibe-check init` section SHALL state that `init --json` emits the same
three-field `provenance` object and accepts a `--no-provenance` flag. Neither
section SHALL describe an `input` field, which is `analyze`-only.

#### Scenario: Diff section mentions provenance
- **WHEN** the `vibe-check diff` section of `README.md` is read
- **THEN** it mentions the `provenance` object (with `producer` as the constant
  `"vibe-check"`, `version`, and RFC 3339 UTC `generatedAt`) and the
  `--no-provenance` flag, and does not describe an `input` field

#### Scenario: Init section mentions provenance
- **WHEN** the `vibe-check init` section of `README.md` is read
- **THEN** it mentions the `provenance` object (with `producer` as the constant
  `"vibe-check"`, `version`, and RFC 3339 UTC `generatedAt`) and the `--no-provenance`
  flag, and does not describe an `input` field

### Requirement: CHANGELOG records the provenance changes

`CHANGELOG.md` SHALL contain, under `[Unreleased]`, an entry for the `analyze`
provenance change (`openspec/changes/emit-provenance-metadata/`) and an entry for
the `diff`/`init` provenance change (`openspec/changes/add-diff-init-provenance/`).

#### Scenario: Unreleased entries present for both changes
- **WHEN** the `[Unreleased]` section of `CHANGELOG.md` is read
- **THEN** it references both `emit-provenance-metadata` and
  `add-diff-init-provenance` in connection with the provenance feature
