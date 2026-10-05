## Context

`vibe-check analyze` produces point-in-time ModuleGraph snapshots with coupling metrics (Instability, Abstractness, Distance, LCOM4, cycles). `vibe-check diff` detects single-commit regressions. Neither captures gradual, multi-commit degradation. The RFC (Group 7) specifies architectural drift tracking via time-series storage in Dewey, trend detection, and drift alerting. The existing `/vibe-check trending` mode spec compares only the most recent snapshot — it needs extension to multi-snapshot time-series baselines.

**Constraints:** Dewey is the storage backend (MCP tools: `dewey_store_learning`, `dewey_semantic_search`, `dewey_get_page`). The Go binary does not embed Dewey SDK logic — all MCP interaction lives in agent prompts. The `vibe-check analyze` binary must produce Dewey-ingestible output via a new `--store` flag without adding heavyweight dependencies.

## Goals / Non-Goals

**Goals:**
- `vibe-check analyze --store` enriches ModuleGraph output with snapshot metadata (timestamp, commit SHA, branch) for CI-driven archival
- `mx-f-architecture-trend.md` agent retrieves stored snapshots via Dewey MCP, computes trends across time windows, and generates drift alerts
- `/vibe-check trending` mode extends from single-baseline to multi-snapshot time-series with 7/30/90-day baselines and trend classification
- Drift alerts with configurable thresholds for sustained degradation and projected threshold crossings

**Non-Goals:**
- Built-in scheduling or cron — CI and agent invocation are external concerns
- Dewey schema changes or Dewey-side code — the agent uses existing MCP tools
- A full time-series database or charting library — text-based sparklines in the agent output suffice
- Changes to the ModuleGraph JSON schema for snapshot storage — metadata enrichment via CLI flags, not schema evolution
- Breaking changes to existing `analyze`, `diff`, or `init` commands — the `--store` flag is additive only

## Decisions

### D1: Snapshot Format — Provenance-Rich ModuleGraph

`vibe-check analyze --store` augments the existing `provenance` envelope (producer, version, generatedAt) with snapshot-specific fields: `commitSHA` (string, 0 or 40 hex chars; empty when not in a git repo), `branch` (string, 0+ chars; empty when not in a git repo), `modulePath` (string, non-empty Go module path). The output remains standard ModuleGraph JSON — Dewey storage keys on modulePath + generatedAt. Snapshot integrity is validated by JSON schema validation against the ModuleGraph schema as a pre-storage check; Dewey's storage layer handles its own persistence integrity.

**Rationale:** Avoids schema churn. The provenance object already carries generator identity; snapshot fields are transport metadata the agent uses for Dewey storage keys and time-series alignment.

**Alternative considered:** New top-level `snapshot` wrapper type. Rejected — adds a schema variant that complicates validation without benefit.

### D2: Storage via Agent, Not Binary

The `mx-f-architecture-trend.md` agent (not the Go binary) calls `dewey_store_learning` to persist snapshots. The binary's `--store` flag enriches output; the agent captures it in CI and stores it.

**Rationale:** Dewey MCP tools are agent-native. Adding an HTTP client or JSON-RPC dependency to the Go binary for Dewey storage violates Layer-1 purity and creates a circular build dependency. The agent is the integration point.

### D3: Trend Detection — Windowed Linear Regression

For each tracked metric (Instability, Distance, LCOM4) per package, the agent computes linear regression over 7-day, 30-day, and 90-day windows. A package is "degrading" if slope exceeds a configurable threshold and R² ≥ 0.5 (non-noise). "Improving" mirrors with negative slope.

**Rationale:** Simple linear regression is deterministic, explainable, and sufficient for coarse trend classification. R² gate filters noisy flat lines from false-positive degradation alerts.

**Alternative considered:** Exponential moving average (EMA). Rejected — EMA weights recent snapshots too heavily for 90-day baselines and lacks a natural confidence metric.

### D4: Drift Alert Thresholds — Separate from Entropy Gates

Drift alerts use their own threshold bands, configured in the agent prompt, not the protected entropy gate thresholds in `metrics/verdict.go`. Defaults: ΔInstability ≥ 0.05 per-snapshot vs 0.15 for entropy gates; ΔDistance ≥ 0.10 vs 0.20.

**Rationale:** Drift is cumulative and lower-magnitude than a single-commit regression. Conflating the two would either miss subtle drift or produce noise at PR time. Separate thresholds per AGENTS.md gatekeeping rule.

### D5: Trending Mode Extension — Additive, Not Replacement

`/vibe-check trending` retains existing single-baseline behavior. New behavior layers on: when ≥2 snapshots exist for the module, the agent adds a time-series section with sparklines and trend classification, using the 7/30/90-day windows. When only 1 snapshot exists, behavior matches the current spec (compare, classify direction).

**Rationale:** Backward compatible. Existing trending-mode spec scenarios pass unchanged. New scenarios are additive.

### D6: Sparkline — ASCII Block Characters

Trend visualization uses Unicode block characters (▁▂▃▄▅▆▇█) normalized to the metric's observed range per package. Max width: 30 characters. Rendered inline in the agent's markdown table output.

**Rationale:** No charting library needed. Text-based renderer works in any terminal, PR comment, or agent output. Fits the "zero new dependencies" constraint.

**Edge case — single-value range**: When all observed values for a metric are identical (range = 0), the sparkline renders all characters at the mid-block (▄) to indicate a flat trend rather than attempting division by zero. When the range is near-zero (floating-point noise), values are clamped to the nearest block character without amplification.

## Verification Strategy

Spec acceptance scenarios serve as the verification test cases for agent-level logic (trend detection, drift alerting, sparkline rendering). The acceptance scenarios in `specs/trend-detection/spec.md` (relative to this change directory) provide explicit data-point inputs that test slope calculation bounds and R² gating. Agent invocation is tested end-to-end via scenario GIVEN/WHEN/THEN patterns rather than Go-level unit tests, consistent with the agent-as-integration-point architecture (D2).

## Risks / Trade-offs

- **[R] Dewey MCP unavailability blocks trending mode**: Agent degrades gracefully per existing spec — reports limitation, suggests summary/detailed mode. No Go-level failure.
- **[R] Linear regression on sparse data produces false trends**: Mitigated by R² ≥ 0.5 gate and minimum 3 snapshots per window before classification.
- **[R] Snapshot metadata (commit SHA, branch) may be stale in detached-HEAD CI**: Agent uses git to resolve context when `--store` metadata is empty; falls back to generatedAt timestamp for ordering.
- **[R] No built-in retention monitoring**: The agent enforces retention during each run, but if the agent stops being invoked (e.g., CI pipeline removed), snapshots accumulate silently. Operators should monitor agent invocation frequency or set a Dewey-side alert for snapshot count growth. This is an operational concern, not a code defect — the agent reports pruned counts in its output for external monitoring.
- **[R] Retention policy enforcement is agent-only**: No Go-side enforcement. If the agent is not invoked, old snapshots accumulate. Acceptable — Dewey pages are low-overhead and can be cleaned manually or via a future Dewey-side retention feature. Individual bad snapshots can be removed via Dewey learning removal; bulk cleanup is a future Dewey-side concern.
- **[R] Retention policy timezone ambiguity**: The `generatedAt` timestamp uses the local system timezone of the CI runner. A snapshot taken at 11 PM UTC-5 on a Tuesday is already Wednesday in UTC. The agent uses the timestamp as-is for day-of-week calculations (Sunday detection for weekly retention). This may cause edge-case misclassification for snapshots near midnight, but the impact is limited to retention (not metric computation) and self-corrects within 24 hours.