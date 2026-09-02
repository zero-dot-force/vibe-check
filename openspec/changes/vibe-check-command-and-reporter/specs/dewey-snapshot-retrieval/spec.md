# Spec: Dewey Snapshot Retrieval

## ADDED Requirements

### Requirement: Agent retrieves previous snapshots from Dewey

The agent SHALL retrieve the most recent snapshot for the current
project from Dewey when operating in trending mode.

#### Scenario: Retrieving latest snapshot

- **GIVEN** the agent operates in trending mode and Dewey is available
- **WHEN** the agent queries for previous snapshots
- **THEN** it calls `dewey_semantic_search` with a query containing
  `vibe-check-snapshot` and the current Go module path (from `go.mod`),
  filters results to match the current module path in the snapshot
  content, parses the ISO 8601 timestamp from each matching result,
  and selects the most recent result

#### Scenario: Multiple snapshots available

- **GIVEN** Dewey contains multiple snapshots for the project
- **WHEN** the agent retrieves snapshots
- **THEN** the agent selects the most recent snapshot (by parsed
  timestamp) as the comparison baseline

#### Scenario: Snapshot from wrong project returned

- **GIVEN** Dewey returns a snapshot whose module path does not match
  the current project
- **WHEN** the agent filters retrieval results
- **THEN** the agent skips the mismatched snapshot and selects the next
  matching result, or falls through to the "no snapshots found" case

#### Scenario: No snapshots found

- **GIVEN** the agent operates in trending mode
- **WHEN** Dewey returns no matching snapshots for the current module path
- **THEN** the agent reports that no historical baseline exists and
  stores the current analysis as the first snapshot

### Requirement: Snapshot comparison produces deltas

The agent SHALL compute the difference between the current analysis
and the retrieved snapshot to produce per-package metric deltas.

#### Scenario: Package exists in both snapshots

- **GIVEN** a valid baseline snapshot has been retrieved
- **WHEN** a package exists in both the current analysis and the
  retrieved snapshot
- **THEN** the agent computes the delta for each metric (current − baseline)
  and classifies the direction as improving (delta ≤ −0.01 for
  instability/distance, delta ≤ −1 for LCOM4), degrading (delta ≥
  0.01 for instability/distance, delta ≥ 1 for LCOM4), or stable
  (delta within tolerance). For integer LCOM4, "decreased by 1 or
  more" is improving, consistent with the trending mode spec

#### Scenario: New package not in baseline

- **GIVEN** a valid baseline snapshot has been retrieved
- **WHEN** a package exists in the current analysis but not in the
  retrieved snapshot
- **THEN** the agent reports the package as "new" with its current
  metric values and no delta

#### Scenario: Removed package in baseline only

- **GIVEN** a valid baseline snapshot has been retrieved
- **WHEN** a package exists in the retrieved snapshot but not in the
  current analysis
- **THEN** the agent reports the package as "removed"

#### Scenario: Corrupted or malformed snapshot

- **GIVEN** a snapshot is retrieved from Dewey
- **WHEN** the snapshot content cannot be parsed (missing required
  fields, non-numeric metric values, or truncated content)
- **THEN** the agent skips the corrupted snapshot, logs a warning,
  and falls through to the next most recent snapshot or the "no
  snapshots found" case
