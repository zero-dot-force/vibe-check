# Proposal: /vibe-check Command and vibe-check-reporter Agent

## Why

Vibe-Check produces rich structural metrics (instability, abstractness,
distance from main sequence, LCOM4, circular dependencies) but
currently requires manual CLI invocation and JSON interpretation. There
is no way for developers to get architectural feedback within their
normal AI-assisted workflow, and no mechanism to track how metrics
evolve across commits.

This change closes that gap by creating an OpenCode slash command and
companion agent that bring vibe-check analysis directly into the
developer conversation. It addresses GitHub issue #5 and is scoped to
Phase P2 of the RFC.

## What Changes

Two new embedded assets are added to the scaffold system and deployed
via `vibe-check init`:

1. **`/vibe-check` slash command** (`internal/scaffold/assets/commands/vibe-check.md`)
   — An OpenCode command definition that accepts a mode parameter and
   delegates to the vibe-check-reporter agent.

2. **`vibe-check-reporter` agent** (`internal/scaffold/assets/agents/vibe-check-reporter.md`)
   — An OpenCode agent that invokes `vibe-check analyze`, interprets the
   JSON output, and presents results in natural language with actionable
   guidance.

3. **Scaffold system extension** — The embed directives and deployment
   logic in `internal/scaffold/` are extended to support a `commands/`
   asset directory alongside the existing `agents/` directory. The
   `vibe-check init` command deploys both asset types.

## Capabilities

### New Capabilities

| Capability | Description |
|---|---|
| Summary mode | Traffic-light overview (green/yellow/red) of module health with top-level instability, distance, and LCOM4 summaries. Default mode. |
| Detailed mode | Per-package breakdown showing all Martin metrics, zone classification, and specific warnings with remediation guidance. |
| Trending mode | Compares current analysis against stored baselines using Dewey. Shows metric direction (improving/degrading/stable) per package. |
| Dewey snapshot storage | Agent stores ModuleGraph snapshots in Dewey with timestamp and commit metadata for longitudinal tracking. |
| Dewey snapshot retrieval | Agent retrieves previous snapshots from Dewey to compute trends and show historical context. |
| Natural language interpretation | Agent translates raw metric values into plain-English assessments (e.g., "Package X has high instability (0.89) — it depends on many packages but nothing depends on it, making it fragile to upstream changes"). |

### Modified Capabilities

| Capability | Change |
|---|---|
| `internal/scaffold/embed.go` | Add `//go:embed` directive for `assets/commands/*.md`. |
| `internal/scaffold/scaffold.go` | Extend `Run()` to deploy command assets to `.opencode/commands/` in addition to agent assets to `.opencode/agents/`. |
| `internal/scaffold/scaffold_test.go` | Add test coverage for command asset deployment. |

## Impact

- **Scaffold contract**: The `Result` struct gains command entries in
  Written/Skipped/Forced slices. Existing agent-only consumers see no
  behavioral change — new command assets are additive.
- **Dependencies**: No new Go module dependencies. The agent and command
  are markdown assets embedded at compile time. The agent relies on
  `vibe-check analyze` being available in PATH (already a prerequisite
  for any project using vibe-check).
- **Dewey dependency**: Trending mode requires Dewey MCP tools
  (`dewey_store_learning`, `dewey_semantic_search`). The agent degrades
  gracefully when Dewey is unavailable — trending mode reports that
  historical data is not available instead of failing.
- **Issue #2 dependency**: The agent invokes `vibe-check analyze` which
  is already implemented (issue #2 is complete).
- **Documentation**: README.md, AGENTS.md, and CHANGELOG.md require
  updates to reflect the new command, agent, and modified init behavior.
  A website documentation sync issue is required per constitution.

## Constitution Alignment

| Principle | Assessment |
|---|---|
| I. Autonomous Collaboration | **Aligned.** The agent and command are self-describing markdown artifacts with metadata (frontmatter). The agent consumes `vibe-check analyze` output (a well-defined JSON schema) and Dewey learnings (tagged, timestamped artifacts) — no synchronous agent-to-agent coupling. |
| II. Composability First | **Aligned.** The agent degrades gracefully when Dewey is unavailable (trending mode reports limitation instead of failing). The command and agent are independently deployable via `vibe-check init`. The agent composes with the existing `vibe-check analyze` CLI without modification. |
| III. Observable Quality | **Partial.** The agent produces conversational markdown, which is human-readable but not machine-parseable. This is acceptable because the agent is a developer-facing interpretation layer, not a CI gate (that role belongs to `divisor-entropy`). The underlying metrics remain machine-parseable via `vibe-check analyze --output`. |
| IV. Testability | **Aligned.** The scaffold Go code (embed, deploy) is tested via unit tests following the existing pattern. Agent and command markdown assets are validated via embedded asset contract tests (frontmatter structure, required sections). Agent behavioral correctness is validated at integration time. Coverage strategy is defined in design.md. |
| V. Security by Default | **Aligned.** The agent's bash allowlist follows least-privilege (only `vibe-check analyze *` and `git rev-parse *`). The scaffold reuses the existing symlink-safe, containment-checked deployment pattern. User-supplied package patterns are validated before shell interpolation. |
| VI. Metric Fidelity | **Aligned.** The agent does not compute metrics — it delegates to `vibe-check analyze` which produces deterministic results. The agent interprets and presents metrics faithfully using the zone classification and threshold definitions from the `metrics/` package. Trending mode uses defined tolerance thresholds for stable/improving/degrading classification. |
| VII. Language Agnosticism | **Not applicable.** This change adds Go-specific agent and command assets. The underlying metrics model remains language-agnostic; the agent invokes the CLI which routes through the adapter registry. |
