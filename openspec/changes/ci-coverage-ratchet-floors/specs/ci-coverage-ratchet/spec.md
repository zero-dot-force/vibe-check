# Spec Delta

## Purpose

Enforces per-package Go statement-coverage floors in CI via a committed
baseline file and a deterministic gate tool, so coverage can ratchet upward
but never silently regress — implementing Constitution Article IV's mandate
that coverage regressions block the build.

## ADDED Requirements

### Requirement: Per-package coverage floor evaluation
The CI test job SHALL evaluate per-package statement coverage from the
collected coverage profile against the recorded floors, and SHALL fail when
any package's measured coverage is below its floor. Evaluation at exactly the
floor value SHALL pass.

#### Scenario: Package below its floor fails the build
- **WHEN** the coverage profile shows a package whose measured statement coverage is below its recorded floor
- **THEN** the coverage gate exits non-zero with a policy-failure exit code and the CI job fails

#### Scenario: All packages at or above their floors pass
- **WHEN** every package's measured statement coverage is at or above its recorded floor
- **THEN** the coverage gate exits 0 and the CI job continues

#### Scenario: Measurement exactly at the floor passes
- **WHEN** a package's measured statement coverage equals its recorded floor exactly
- **THEN** the coverage gate treats the package as passing

### Requirement: Coverage floors recorded in a committed baseline
Coverage floors SHALL be recorded in a version-controlled JSON baseline file
at the repository root, keyed by module-relative package path, with floor
values expressed as percentages of one-decimal precision. The file SHALL be
serialized deterministically (lexicographically sorted package keys) so
baseline updates produce minimal, reviewable diffs.

#### Scenario: Baseline is committed and consumed by CI
- **WHEN** the CI coverage gate runs
- **THEN** it reads the floors from the committed baseline file in the repository, not from workflow configuration

#### Scenario: Deterministic serialization on update
- **WHEN** the baseline file is rewritten by the gate tool's update mode
- **THEN** package keys are emitted in lexicographic order with stable formatting, so repeated updates with unchanged coverage produce no diff

#### Scenario: One-decimal floor precision
- **WHEN** a floor is recorded or updated
- **THEN** it is stored as a percentage rounded down to one decimal place, matching the precision `go test` reports

### Requirement: Complete package registration
The baseline SHALL register a floor for every package in the module that
contains coverable statements, and SHALL NOT carry an entry for a package
that has none. The coverage gate SHALL fail when a package with coverable
statements is missing from the baseline, forcing explicit registration of
new packages rather than silent exemption, and SHALL likewise fail when a
baseline entry names a package with no coverable statements, so deleted
or renamed packages cannot leave a stale floor behind.

#### Scenario: Unregistered package fails the gate
- **WHEN** the module contains a package with coverable statements that has no entry in the baseline
- **THEN** the coverage gate exits non-zero, reporting the unregistered package

#### Scenario: Stale floor fails the gate
- **WHEN** the baseline contains an entry for a package that has no coverable statements in the module
- **THEN** the coverage gate exits non-zero, reporting the stale floor, and update mode refuses to rewrite the baseline until the entry is removed by a direct, human-reviewed edit

#### Scenario: All registered packages pass completeness
- **WHEN** every package with coverable statements has a baseline entry and every baseline entry names such a package
- **THEN** the coverage gate reports no registration violation

### Requirement: Tighten-only ratchet updates
The gate tool's baseline update mode SHALL only tighten floors: it SHALL
raise an existing floor to a higher measured value and SHALL NOT lower any
existing floor under any tool invocation. A newly registered package SHALL
enter the baseline only at a measured coverage of at least 80 percent.
Lowering a floor SHALL require an explicit, human-reviewed edit of the
baseline file; no tool path may do it.

#### Scenario: Update raises a floor to measured coverage
- **WHEN** update mode runs with a package measured above its recorded floor
- **THEN** the baseline entry for that package is raised to the measured value and no other floor changes

#### Scenario: Update never lowers a floor
- **WHEN** update mode runs with a package measured below its recorded floor
- **THEN** the baseline entry is left unchanged and the regression is reported

#### Scenario: New package below the entry bar is refused
- **WHEN** update mode runs with an unregistered package measured below 80 percent
- **THEN** the update fails and the package is not added to the baseline

#### Scenario: No tool path loosens a floor
- **WHEN** any gate tool invocation is attempted with the intent of lowering a floor
- **THEN** no flag or mode produces a lowered floor; only a direct, reviewed edit of the baseline file can change a floor downward

### Requirement: Deterministic gate reporting
The coverage gate SHALL produce a deterministic per-package report — a
human-readable table and a machine-readable JSON mode carrying a provenance
envelope (producer, version, generation timestamp) — with exit codes
mirroring the project CLI convention: 0 for success, 1 for policy failure
(floor or registration violation), 2 for tool failure (unreadable inputs,
malformed profile or baseline).

#### Scenario: Human-readable per-package report
- **WHEN** the coverage gate runs in default mode
- **THEN** it prints one line per package with the measured coverage, the floor, and pass/fail status, followed by a summary

#### Scenario: Machine-readable JSON output
- **WHEN** the coverage gate runs with the JSON flag
- **THEN** it emits a JSON object containing per-package results, the overall verdict, and a provenance envelope with producer, version, and RFC 3339 UTC generation timestamp

#### Scenario: Tool failure on malformed inputs
- **WHEN** the coverage profile or baseline file is missing or unparseable
- **THEN** the coverage gate exits 2 with a diagnostic, without modifying the baseline

#### Scenario: Identical inputs produce identical output
- **WHEN** the coverage gate runs twice against the same coverage profile and baseline
- **THEN** the report content (excluding the provenance generation timestamp) is byte-identical
