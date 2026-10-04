## Context

The `internal/scaffold` package embeds agent and command assets via two separate `//go:embed` directives (`agentAssetsFS` for `assets/agents/*.md`, `commandAssetsFS` for `assets/commands/*.md`). These are the canonical source of truth — the copies deployed into `.opencode/` are generated from them by `vibe-check init`.

Currently, two assets are missing from `.opencode/` (`vibe-check-reporter.md`, `vibe-check.md`), and there is no automated test that detects byte-level drift between embedded assets and deployed copies. The gaze project established a proven pattern for this: `TestEmbeddedAssetsMatchSource` with `assetPaths()` and `assetContent()` helpers.

## Goals / Non-Goals

**Goals:**
- Add `assetPaths()` that enumerates all embedded asset paths (stripped of the `assets/` prefix) from both `agentAssetsFS` and `commandAssetsFS`.
- Add `assetContent(relPath)` that reads an embedded asset's bytes given its relative path (without `assets/` prefix).
- Add `TestEmbeddedAssetsMatchSource` that byte-compares each embedded asset against its deployed copy in `.opencode/<rel>`, using `findProjectRoot()` to locate the repo root from the test binary's working directory.
- Deploy missing assets into `.opencode/` via `vibe-check init --force .`.
- Dogfood `vibe-check analyze` + `vibe-check diff` on the repo to confirm no structural entropy regression.

**Non-Goals:**
- Changing the embed directives or asset directory structure.
- Adding a CI gate that runs the drift test (CI already treats test failures as blocking).
- Adding drift detection for non-scaffold `.opencode/` files (only scaffold-managed assets).

## Decisions

### D1: `assetPaths()` returns paths stripped of `assets/` prefix

**Rationale**: The `.opencode/` tree mirrors the embedded `assets/` tree with the `assets/` prefix removed (`assets/agents/divisor-entropy.md` → `agents/divisor-entropy.md`). Stripping the prefix once in `assetPaths()` means all downstream code (tests, remediation hints) works with `.opencode/`-relative paths.

**Alternative considered**: Keeping the `assets/` prefix and mapping in the test. Rejected — adds noise to every comparison and error message.

### D2: `assetPaths()` walks both `agentAssetsFS` and `commandAssetsFS`

**Rationale**: The two embed directives mean we have two separate `fs.FS` values. `assetPaths()` must union them to return the complete set of embedded assets. The implementation uses `fs.WalkDir` on each FS and merges the results.

**Alternative considered**: Refactoring to a single `//go:embed assets` directive. Rejected — the two-directive pattern is intentional for category disambiguation in `Run()`, and changing it is out of scope for this change.

**Maintenance note**: The FS list in `assetPaths()` must stay in sync with the `categories` slice in `Run()` (scaffold.go). If a third asset category is added, both locations need updating.

### D3: `assetContent(relPath)` prepends `assets/` and dispatches to the correct FS

**Rationale**: The function receives a `.opencode/`-relative path (e.g., `agents/divisor-entropy.md`) and needs to read from the embedded filesystem which uses `assets/`-prefixed paths. It prepends `assets/` and tries both `agentAssetsFS` and `commandAssetsFS` via `fs.ReadFile`, returning the first successful read or an error if neither contains the path.

### D4: `findProjectRoot()` walks up from working directory to find `go.mod`

**Rationale**: Tests run with the working directory set to the package directory (`internal/scaffold/`). The function walks parent directories until it finds `go.mod`, then returns that directory as the project root. This matches the gaze pattern. If no `go.mod` ancestor is found (e.g., the test binary runs from an unusual CI directory), the function returns an error and the test skips with a clear message. This two-phase lookup is distinct from the missing `.opencode/` check: `findProjectRoot()` handles the case where the project cannot be located at all, while `TestEmbeddedAssetsMatchSource` handles the case where the project root is found but `.opencode/` is absent.

**Alternative considered**: Using a fixed relative path (`../../.opencode/`). Rejected — fragile if tests are run from a different working directory.

### D6: `TestEmbeddedAssetsMatchSource` is classified as an integration test

**Rationale**: Unlike existing scaffold tests that deploy to `t.TempDir()`, the drift detection test reads the real `.opencode/` directory to verify that deployed assets match the embedded source. This is a feature, not a bug — using a temp dir would defeat the purpose of detecting actual drift. The test is classified as an integration/contract test, uses `testing.Short()` as a skip guard (so `go test -short` bypasses it), and is documented in the Test Strategy section below as intentionally departing from `t.TempDir()` isolation.

### D5: Remediation hint uses `vibe-check init --force .`

**Rationale**: The embedded assets are the source of truth. The remediation path is to regenerate deployed copies from the embedded source, not the reverse. The most common scenario (missing files) only requires `vibe-check init .` (without `--force`), but the hint uses `--force` to also cover content-drift cases.

## Risks / Trade-offs

- **Test depends on `.opencode/` being present**: If `.opencode/` is deleted, the test fails. This is intentional — the test is a guard against accidental deletion or drift. The remediation hint tells users exactly how to restore.
- **Two-embed-FS complexity**: `assetPaths()` and `assetContent()` must handle two FS values, which adds some branching. This is acceptable because the two-directive pattern is stable and well-understood.
- **findProjectRoot could fail in unusual CI setups**: If the test binary runs from a directory with no `go.mod` ancestor, the test will skip with a clear message. This matches the gaze pattern's graceful degradation.

## Test Strategy

**Coverage target**: ≥80% statement coverage for all new functions (`assetPaths`, `assetContent`, `findProjectRoot`). Coverage is enforced via `go test -cover ./internal/scaffold/` and the CI parity gate in `tasks.md` section 4.

**Test classification**:

| Test | Type | Isolation | Skip Guard |
|------|------|-----------|------------|
| Unit tests for `assetPaths()`, `assetContent()` | Unit | No filesystem | — |
| `TestEmbeddedAssetsMatchSource` | Integration/Contract | Reads real `.opencode/` | `testing.Short()` |

`TestEmbeddedAssetsMatchSource` is classified as an integration test because it reads the actual `.opencode/` directory to detect byte-level drift between embedded assets and deployed copies. Using `t.TempDir()` would defeat this purpose — the test must verify the real deployment, not a temporary copy. The test uses `testing.Short()` as a skip guard so `go test -short ./...` bypasses the real-filesystem check. This is an intentional departure from the project's `t.TempDir()` isolation convention (TC-004), justified by the test's purpose as a drift detection guard.

**Unit test coverage**: `assetPaths()` is tested for path enumeration correctness (agent + command assets, no `assets/` prefix, stable ordering). `assetContent()` is tested via table-driven tests for agent asset retrieval, command asset retrieval, and unknown-path error returns. `findProjectRoot()` is tested for both the success path (finds `go.mod`) and the failure path (no `go.mod` ancestor → error).