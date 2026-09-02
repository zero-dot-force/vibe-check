# Spec: Detailed Mode

## ADDED Requirements

### Requirement: Detailed mode shows per-package metrics

The agent SHALL present a per-package breakdown of all Martin metrics
when invoked in detailed mode.

#### Scenario: Detailed mode invocation

- **GIVEN** the `vibe-check` binary is installed and in PATH
- **WHEN** user invokes `/vibe-check detailed`
- **THEN** the agent runs `vibe-check analyze --output <tempfile> ./...`
  and presents per-package results

#### Scenario: Detailed mode with package pattern

- **GIVEN** the `vibe-check` binary is installed and in PATH
- **WHEN** user invokes `/vibe-check detailed ./internal/...`
- **THEN** the agent validates the package pattern and analyzes only
  the specified packages, presenting per-package results for those
  packages

### Requirement: Per-package metric table

The detailed mode SHALL present each package's metrics in a structured
format including Ca, Ce, instability, abstractness, distance from main
sequence, LCOM4, and zone classification.

#### Scenario: Package in Zone of Pain

- **GIVEN** `vibe-check analyze` completes with exit code 0 or 1
- **WHEN** a package has high stability (low instability) and low
  abstractness (concrete + stable = Zone of Pain)
- **THEN** the detailed output identifies the package's zone as
  "Zone of Pain" and explains that changes to this package ripple
  widely because many packages depend on its concrete implementations

#### Scenario: Package in Zone of Uselessness

- **GIVEN** `vibe-check analyze` completes with exit code 0 or 1
- **WHEN** a package has high instability and high abstractness
  (Zone of Uselessness)
- **THEN** the detailed output identifies the zone and explains that
  the package defines abstractions nothing concrete depends on

### Requirement: Warnings and remediation guidance

The detailed mode SHALL include any warnings from the analysis and
provide actionable remediation guidance for packages with concerning
metrics.

#### Scenario: High LCOM4 package

- **GIVEN** `vibe-check analyze` completes with exit code 0 or 1
- **WHEN** a package has LCOM4 > 1
- **THEN** the detailed output explains what LCOM4 measures and
  suggests the package may benefit from being split into separate
  packages along its disconnected responsibility clusters

#### Scenario: Package with threshold violations

- **GIVEN** `vibe-check analyze` was invoked with `--max-*` threshold flags
- **WHEN** a package exceeds a configured threshold
- **THEN** the detailed output highlights the violation and provides
  specific guidance for reducing the metric value

#### Scenario: Analysis error in detailed mode

- **GIVEN** `vibe-check analyze` has been invoked
- **WHEN** the process exits with code 2 (error)
- **THEN** the agent displays the error and does not produce a
  per-package table
