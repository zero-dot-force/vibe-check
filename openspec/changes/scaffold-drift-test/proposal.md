## Why

The `internal/scaffold` package declares embedded assets as the "single source of truth" for agent and command definitions deployed into `.opencode/`, but there is no automated test verifying that the copies currently in `.opencode/` match what is embedded. Two of three assets (`vibe-check-reporter.md`, `vibe-check.md`) are already missing from `.opencode/`. Without a drift detection test, the deployed copies can silently diverge from the embedded source of truth over time.

## What Changes

- Add `assetPaths()` and `assetContent()` helper functions to `internal/scaffold/` that enumerate and read embedded assets, stripping the `assets/` prefix for comparison against `.opencode/` paths.
- Add `TestEmbeddedAssetsMatchSource` that byte-compares each embedded asset against its deployed copy in `.opencode/`, with a remediation hint directing users to run `vibe-check init --force .` when drift is detected.
- Deploy the two missing assets into `.opencode/` via `vibe-check init --force .`.
- Dogfood `vibe-check analyze` + `vibe-check diff` on the repo itself to confirm the scaffold changes do not introduce structural entropy regression.

## Capabilities

### New Capabilities
- `scaffold-drift-detection`: Automated test that detects divergence between embedded scaffold assets and their deployed copies in `.opencode/`, with a clear remediation path.

### Modified Capabilities
<!-- None: this change adds a test, no existing spec requirements change. -->

## Impact

- Affected code: `internal/scaffold/scaffold.go` (new helpers), `internal/scaffold/scaffold_test.go` (new test)
- Affected files: `.opencode/agents/vibe-check-reporter.md`, `.opencode/commands/vibe-check.md` (newly deployed)
- No API changes, no dependency additions, no breaking changes.