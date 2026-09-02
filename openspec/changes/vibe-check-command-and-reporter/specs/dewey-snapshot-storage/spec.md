# Spec: Dewey Snapshot Storage

## ADDED Requirements

### Requirement: Agent stores metric snapshots in Dewey

The agent SHALL store a summary snapshot of the current analysis in
Dewey after each analysis run when Dewey is available. Snapshots are
stored in all modes (not just trending) so that historical baselines
accumulate for future trending comparisons.

#### Scenario: Successful analysis with Dewey available

- **GIVEN** `vibe-check analyze` completes successfully
- **AND** Dewey MCP tools are available
- **WHEN** the agent prepares to store a snapshot
- **THEN** the agent stores a snapshot via `dewey_store_learning` with
  the tag `vibe-check-snapshot`, the Go module path (from `go.mod`),
  the current commit SHA, a timestamp, and per-package metric summaries
  (instability, abstractness, distance, LCOM4, zone)

#### Scenario: Analysis with Dewey unavailable

- **GIVEN** `vibe-check analyze` completes successfully
- **WHEN** Dewey MCP tools are not available
- **THEN** the agent skips snapshot storage without error and proceeds
  with result presentation

#### Scenario: Duplicate snapshot for same commit

- **GIVEN** Dewey is available and contains a snapshot for the current
  commit SHA and module path
- **WHEN** the agent completes analysis on the same commit
- **THEN** the agent SHOULD skip storage to avoid redundant snapshots
  and note that a snapshot for this commit already exists

### Requirement: Snapshot content is compact

The snapshot stored in Dewey SHALL contain only the per-package metric
summary values, not the full ModuleGraph JSON with type details.

#### Scenario: Snapshot size for large codebase

- **GIVEN** a codebase has 50+ packages
- **WHEN** the agent stores a snapshot
- **THEN** the stored snapshot contains one record per package with
  the six core metric values (Ca, Ce, instability, abstractness,
  distance, LCOM4), zone classification, and module-level metadata
  (module path, commit SHA, timestamp, package count, cycle count).
  Estimated size: ~50 bytes per package (e.g., 200 packages ≈ 10KB),
  well within Dewey learning size limits.

### Requirement: Snapshot includes commit metadata

Each stored snapshot SHALL include the current commit SHA and timestamp
to enable chronological ordering and commit-level traceability.

#### Scenario: Snapshot metadata

- **GIVEN** a snapshot is being stored
- **WHEN** the agent constructs the learning content
- **THEN** the learning content includes the Go module path (from
  `go.mod`), the output of `git rev-parse HEAD` as the commit SHA,
  and the current ISO 8601 timestamp. The module path enables
  disambiguation when multiple projects share the same Dewey instance.
