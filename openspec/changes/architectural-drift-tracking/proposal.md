## Why

Teams need visibility into how structural quality evolves over time. Point-in-time metrics from `vibe-check analyze` and `vibe-check diff` catch single-commit regressions, but gradual degradation across many small changes goes undetected until it becomes critical. This change introduces time-series metric storage in Dewey and trend-detection capabilities — surfacing drift before it harms the codebase.

## What Changes

- **`vibe-check analyze --store`**: Outputs metric snapshots in a Dewey-compatible format for CI-driven archival
- **Snapshot storage schema**: Module-level and package-level metrics as versioned time-series records in Dewey, with retention policy (90-day daily, 1-year weekly)
- **`mx-f-architecture-trend.md` agent**: Periodically analyzes stored snapshots to detect degradation trends, projects threshold violations, and generates natural-language drift summaries
- **`/vibe-check trending` mode enhancement**: Evolves from single-snapshot comparison to multi-snapshot time-series with 7-day, 30-day, and 90-day baselines, text-based sparklines, and drift highlighting
- **Drift alerts**: Configurable thresholds for sustained degradation (N consecutive snapshots) and projected threshold crossings (M days until violation)

## Capabilities

### New Capabilities

- `snapshot-storage`: CI-driven archival of `vibe-check analyze` output as Dewey time-series records with retention policies
- `trend-detection`: Statistical analysis of metric snapshots to identify degradation trends across time windows
- `drift-alerts`: Configurable alerting when metrics degrade across consecutive snapshots or project threshold violations
- `architecture-trend-agent`: The `mx-f-architecture-trend.md` agent that monitors, projects, and reports architectural health trends

### Modified Capabilities

- `trending-mode`: Extends `/vibe-check trending` from single-baseline comparison to multi-snapshot time-series with 7/30/90-day baselines, sparkline visualization, and drift highlighting

## Spec Dependencies

- `snapshot-storage` ← (prerequisite for) → `architecture-trend-agent` ← → `trending-mode`
- `trend-detection` → `architecture-trend-agent`
- `drift-alerts` → `architecture-trend-agent`

## Impact

- **New code**: `cmd/vibe-check/` — `--store` flag on analyze subcommand; `metrics/` — snapshot data types, serialization, and provenance field extensions (`commitSHA`, `branch`, `modulePath`); `internal/scaffold/` — embedded `mx-f-architecture-trend.md` agent asset
- **External dependency**: Dewey (`dewey_store_learning` / `dewey_semantic_search` / `dewey_get_page`) for snapshot storage and retrieval; `vibe-check analyze` for metric snapshots
- **No breaking changes**: Existing `analyze`, `diff`, and `init` commands unchanged; new `--store` flag is additive