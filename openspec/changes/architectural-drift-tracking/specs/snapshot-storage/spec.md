# Spec: Snapshot Storage

## ADDED Requirements

### Requirement: analyze --store enriches output with snapshot metadata

When invoked with the `--store` flag, `vibe-check analyze` SHALL include snapshot-specific metadata fields in the `provenance` envelope of the ModuleGraph JSON output: `commitSHA`, `branch`, and `modulePath`. The standard `producer`, `version`, and `generatedAt` fields remain present.

#### Scenario: analyze --store includes commit and branch

- **GIVEN** the `vibe-check` binary is installed
- **AND** the current directory is a git repository on branch `main` with HEAD at `abc123`
- **WHEN** user invokes `vibe-check analyze --store -o snap.json ./...`
- **THEN** the output `snap.json` contains a `provenance` object with `commitSHA` equal to `abc123`, `branch` equal to `main`, and `modulePath` equal to the Go module path

#### Scenario: analyze --store outside a git repo

- **GIVEN** the `vibe-check` binary is installed
- **AND** the current directory is not a git repository
- **WHEN** user invokes `vibe-check analyze --store -o snap.json ./...`
- **THEN** the output `snap.json` contains a `provenance` object with `commitSHA` and `branch` set to empty strings, and `modulePath` equal to the Go module path

#### Scenario: analyze --store with missing go.mod
- **GIVEN** the `vibe-check` binary is installed
- **AND** the current directory is a git repository but contains no `go.mod` file
- **WHEN** user invokes `vibe-check analyze --store -o snap.json ./...`
- **THEN** the output `snap.json` contains a `provenance` object with `modulePath` set to an empty string
- **AND** a warning is emitted indicating that the Go module path could not be resolved

#### Scenario: --store flag is additive

- **GIVEN** the `vibe-check` binary is installed
- **WHEN** user invokes `vibe-check analyze --no-provenance --store ./...`
- **THEN** the `--no-provenance` flag takes precedence; no provenance object is included (matching existing behavior)

### Requirement: Snapshot format is Dewey-compatible ModuleGraph

The `--store` output SHALL be valid ModuleGraph JSON per schema v1.2, so the `mx-f-architecture-trend.md` agent can store it in Dewey via `dewey_store_learning` using `modulePath` and `generatedAt` as the storage key.

#### Scenario: --store output validates against ModuleGraph schema

- **GIVEN** the `vibe-check` binary is installed
- **WHEN** user invokes `vibe-check analyze --store -o snap.json ./...`
- **THEN** the output `snap.json` validates against `metrics/modulegraph.schema.json` v1.2

### Requirement: Snapshot retention policy is agent-enforced

The `mx-f-architecture-trend.md` agent SHALL enforce a retention policy: keep daily snapshots for 90 days, keep weekly snapshots (Sunday) for 1 year. Snapshots outside these windows SHALL be removed during each agent run.

#### Scenario: Daily snapshot within 90 days is retained

- **GIVEN** a snapshot exists with `generatedAt` within the last 90 days
- **AND** no other snapshot exists for the same day
- **WHEN** the agent enforces retention
- **THEN** the snapshot is retained

#### Scenario: Weekly snapshot older than 90 days but within 1 year is retained

- **GIVEN** a snapshot exists with `generatedAt` on a Sunday, now - 120 days
- **AND** the snapshot is the only one for that week
- **WHEN** the agent enforces retention
- **THEN** the snapshot is retained

#### Scenario: Sunday snapshot within 90 days is retained under both policies

- **GIVEN** a snapshot exists with `generatedAt` on a Sunday within the last 90 days
- **AND** no other snapshot exists for that day
- **WHEN** the agent enforces retention
- **THEN** the snapshot is retained (qualifies under both daily and weekly policies; the daily policy takes precedence for retention within 90 days)

#### Scenario: Daily snapshot older than 90 days with no weekly role is pruned

- **GIVEN** a snapshot exists with `generatedAt` on a Wednesday, now - 100 days
- **WHEN** the agent enforces retention
- **THEN** the snapshot is pruned