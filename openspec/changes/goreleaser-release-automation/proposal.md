## Why

v0.1.0 was tagged and released manually. There is no `.goreleaser.yaml` and no GitHub Release, so users must install with a Go toolchain (`go install .../cmd/vibe-check@v0.1.0`) rather than downloading a pre-built binary. Additionally, `--version` on installed binaries reports `(commit none, built unknown)` because ldflags embedding is only applied during GoReleaser-managed builds.

## What Changes

- Add `.goreleaser.yaml` building `cmd/vibe-check` across linux/darwin/windows × amd64/arm64 with ldflags for `main.version`, `main.commit`, and `main.date`.
- Add a release workflow (`.github/workflows/release.yml`) triggered via `workflow_dispatch` that delegates to `complytime/org-infra` reusable workflows (`reusable_release_preflight.yml` for tag validation+creation, `reusable_release_goreleaser.yml` for GoReleaser execution with cosign signing and SBOMs). Action pinning and permissions are handled by the reusable workflows.

## Capabilities

### New Capabilities

- `goreleaser-config`: GoReleaser configuration building cross-platform binaries with ldflags version embedding
- `release-workflow`: CI workflow that triggers on tag push, runs GoReleaser, and publishes a GitHub Release with checksums

### Modified Capabilities

<!-- No existing specs to modify -->

### Removed Capabilities

<!-- None. -->

## Impact

- New files: `.goreleaser.yaml`, `.github/workflows/release.yml`
- Existing `cmd/vibe-check/main.go` ldflags variables (`version`, `commit`, `date`) are consumed by GoReleaser — no source changes required
- No API, dependency, or breaking changes

## Constitution Alignment

| Principle | Assessment |
|-----------|------------|
| **I. Autonomous Collaboration** | N/A — No new agent artifacts or inter-agent communication. |
| **II. Composability First** | N/A — No new installable agent components. |
| **III. Observable Quality** | PASS — `checksums.txt` provides artifact provenance; `--version` ldflags provide build provenance via `main.version`, `main.commit`, `main.date`. |
| **IV. Testability** | PASS — Configuration-only change. `goreleaser check` validates YAML syntax; `goreleaser build --snapshot --clean` verifies cross-platform builds; ldflags smoke test verifies version embedding. Release workflow serves as integration test on first tag push. |
| **V. Security by Default** | PASS — SHA-pinned actions (CI-001/CI-002/CI-003); least-privilege `permissions` (`contents: write` only on release job per CI-021); automatic `secrets.GITHUB_TOKEN` (no PAT); no new dependencies. |
| **VI. Metric Fidelity** | N/A — No metric computation changes. |
| **VII. Language Agnosticism** | N/A — No adapter, parser, or language-specific component. CI configuration only. |