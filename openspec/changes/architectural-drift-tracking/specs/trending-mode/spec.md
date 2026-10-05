# Spec: Trending Mode (Delta)

## MODIFIED Requirements

### Requirement: Trending mode compares current against historical snapshots

The agent SHALL compare the current analysis results against previously stored snapshots when invoked in trending mode. When 2 or more snapshots exist, the agent SHALL display multi-snapshot time-series with 7-day, 30-day, and 90-day baselines, per-package sparklines, and trend classifications alongside single-baseline comparison. When only 1 snapshot exists, behavior matches the current single-baseline comparison spec.

#### Scenario: Trending with available history

- **GIVEN** the `vibe-check` binary is installed and in PATH
- **WHEN** user invokes `/vibe-check trending`
- **AND** Dewey contains previous snapshots for this project's module path
- **THEN** the agent retrieves the most recent snapshot and shows per-package metric direction (improving/degrading/stable) with delta values
- **AND** if 2 or more snapshots exist, the agent adds a time-series section with sparklines and trend classifications over 7/30/90-day windows

#### Scenario: Trending with no history

- **GIVEN** the `vibe-check` binary is installed and in PATH
- **WHEN** user invokes `/vibe-check trending`
- **AND** Dewey contains no previous snapshots for this project's module path
- **THEN** the agent reports that no historical data is available, stores the current analysis as the first baseline, and suggests running trending mode again after future changes

#### Scenario: Trending with package pattern

- **GIVEN** the `vibe-check` binary is installed and in PATH
- **WHEN** user invokes `/vibe-check trending ./internal/...`
- **THEN** the agent validates the package pattern and compares only the specified packages against their historical values

### Requirement: Trending output includes time-series visualization

When 2 or more snapshots exist, the trending output SHALL include per-package metric sparklines using Unicode block characters (▁▂▃▄▅▆▇█), normalized to the metric's observed range, max width 30 characters.

#### Scenario: Sparkline rendered for multi-snapshot history

- **GIVEN** 5 snapshots exist for package `pkg/foo` with Instability values [0.30, 0.35, 0.38, 0.42, 0.45]
- **WHEN** user invokes `/vibe-check trending`
- **THEN** the output includes a 5-character sparkline for `pkg/foo` Instability with trend classification
- **AND** the sparkline represents increasing values with progressively taller block characters

#### Scenario: No sparkline for single snapshot

- **GIVEN** only 1 snapshot exists for the module
- **WHEN** user invokes `/vibe-check trending`
- **THEN** no sparkline or time-series section is rendered (single-baseline comparison only)

#### Scenario: Sparkline with flat (single-value) range
- **GIVEN** 5 snapshots exist for package `pkg/foo` with Instability values [0.50, 0.50, 0.50, 0.50, 0.50]
- **WHEN** user invokes `/vibe-check trending`
- **THEN** the sparkline renders all characters at the mid-block (▄) to indicate a flat trend
- **AND** the trend is classified as "stable"

### Requirement: Trending output shows multi-window trend classification

The trending output SHALL classify per-package per-metric trends as improving, degrading, or stable for each of the 7-day, 30-day, and 90-day windows.

#### Scenario: Multi-window classification

- **GIVEN** 90 daily snapshots exist for package `pkg/bar`
- **WHEN** user invokes `/vibe-check trending`
- **THEN** the output shows trend classifications for 7-day, 30-day, and 90-day windows independently