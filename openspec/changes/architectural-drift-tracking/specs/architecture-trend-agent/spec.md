# Spec: Architecture Trend Agent

## ADDED Requirements

### Requirement: Agent stores and retrieves snapshots via Dewey MCP

The `mx-f-architecture-trend.md` agent SHALL store `vibe-check analyze --store` output as Dewey learnings using `dewey_store_learning` with the tag `vibe-check-snapshot`, and retrieve stored snapshots via `dewey_semantic_search` filtered by module path.

#### Scenario: Store a snapshot

- **GIVEN** the `vibe-check` binary produced `snap.json` via `analyze --store`
- **WHEN** the agent is invoked with snapshot storage
- **THEN** the agent calls `dewey_store_learning` with the snapshot content, tag `vibe-check-snapshot`, and category `metric-snapshot`

#### Scenario: Retrieve snapshots by module path

- **GIVEN** 10 snapshots exist in Dewey for module `github.com/example/project`
- **WHEN** the agent retrieves snapshots for that module
- **THEN** all 10 snapshots are returned, ordered by `generatedAt`

### Requirement: Agent produces an architectural health report

The agent SHALL produce a structured report containing: overall health classification (improving/stable/degrading), per-package trend table with sparklines, drift alerts, and projected threshold crossings.

#### Scenario: Full health report

- **GIVEN** 30 daily snapshots exist for the module
- **WHEN** the agent generates a health report
- **THEN** the report includes overall classification, a per-package table with 7/30/90-day trends and sparklines, any drift alerts, and any projected crossings

### Requirement: Agent degrades gracefully when Dewey is unavailable

The agent SHALL report a clear limitation when Dewey MCP tools are not available and suggest alternative modes (summary, detailed), matching the existing trending-mode graceful degradation spec.

#### Scenario: Dewey unavailable

- **GIVEN** Dewey MCP tools are not available
- **WHEN** the agent is invoked
- **THEN** the agent reports that architectural trend analysis requires Dewey and suggests using summary or detailed mode via `/vibe-check`

### Requirement: Agent enforces snapshot retention policy

The agent SHALL enforce the 90-day daily / 1-year weekly retention policy during each run by identifying and removing stale snapshots beyond their retention window.

#### Scenario: Agent prunes stale snapshots

- **GIVEN** 150 daily snapshots exist for the module, spanning 150 days
- **WHEN** the agent runs retention enforcement
- **THEN** snapshots older than 90 days (excluding weekly-Sunday snapshots) are removed, and weekly snapshots older than 1 year are removed