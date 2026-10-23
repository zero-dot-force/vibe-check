# Spec Delta

## Purpose

Markdown output for `vibe-check diff`, rendering the base-vs-PR structural-entropy verdict as a CI-publishable markdown report without changing verdict or gate behavior.

## ADDED Requirements

### Requirement: Markdown output mode
`vibe-check diff <base.json> <pr.json> --markdown` SHALL render the verdict, triggering reasons, entropy direction, per-module metric deltas, and new/resolved cycles as markdown to stdout instead of the default table.

#### Scenario: Request-changes verdict
- **WHEN** `diff --markdown` runs on snapshots whose delta triggers REQUEST_CHANGES
- **THEN** stdout contains the verdict, each triggering reason, the per-module deltas, and any new cycles

#### Scenario: Approve verdict
- **WHEN** `diff --markdown` runs on snapshots whose delta yields APPROVE
- **THEN** stdout contains the verdict and entropy direction, with no triggering reasons or new cycles listed

### Requirement: Flag mutual exclusion
`--markdown` and `--json` SHALL NOT be combinable.

#### Scenario: Both flags
- **WHEN** both `--markdown` and `--json` are passed
- **THEN** the command exits 2 with a usage error and writes no report to stdout

### Requirement: Default output unchanged
Without `--markdown` or `--json`, diff SHALL render the existing plain-text table exactly as before this change.

#### Scenario: Table default
- **WHEN** diff runs with no output flags
- **THEN** the existing tabwriter table renders, unchanged by this capability

### Requirement: Gate semantics preserved
`--markdown` SHALL NOT alter verdict computation, protected verdict thresholds, or exit semantics; `--gate` keeps exiting 1 on REQUEST_CHANGES. The markdown view is display-only and SHALL never be an input to verdict or gate computation.

#### Scenario: Gated markdown
- **WHEN** `diff --markdown --gate` runs on a REQUEST_CHANGES delta
- **THEN** the markdown report is written to stdout and the command exits 1

### Requirement: Byte-stable markdown
The markdown output SHALL be byte-identical for identical input files and contain no generation timestamps.

#### Scenario: Repeat run
- **WHEN** the same diff is rendered twice with `--markdown`
- **THEN** stdout is byte-identical between runs
