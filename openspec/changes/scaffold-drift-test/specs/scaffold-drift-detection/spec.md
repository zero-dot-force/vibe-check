## ADDED Requirements

### Requirement: Embedded asset path enumeration
The scaffold package SHALL provide a function `assetPaths()` that returns the complete set of embedded asset paths from both `agentAssetsFS` and `commandAssetsFS`, with each path stripped of the `assets/` prefix so it matches the `.opencode/`-relative deployment layout (e.g., `agents/divisor-entropy.md`).

#### Scenario: Paths include both agent and command assets
- **GIVEN** the `agentAssetsFS` and `commandAssetsFS` embed directives are loaded
- **WHEN** `assetPaths()` is called
- **THEN** the returned slice SHALL contain at least `agents/divisor-entropy.md`, `agents/vibe-check-reporter.md`, and `commands/vibe-check.md`
- **AND** no path SHALL contain the `assets/` prefix

#### Scenario: Paths are stable across calls
- **GIVEN** the embed directives are loaded and no assets have changed
- **WHEN** `assetPaths()` is called twice
- **THEN** the returned slices SHALL be identical in content and order

### Requirement: Embedded asset content retrieval
The scaffold package SHALL provide a function `assetContent(relPath string)` that returns the full byte content of an embedded asset given its `.opencode/`-relative path, dispatching to the correct embed filesystem. `embed.FS` provides inherent path traversal protection, so no additional path sanitization is required.

#### Scenario: Agent asset content matches embedded source
- **GIVEN** the embedded asset `assets/agents/divisor-entropy.md` exists
- **WHEN** `assetContent("agents/divisor-entropy.md")` is called
- **THEN** the returned bytes SHALL equal the bytes read directly from `agentAssetsFS` at `assets/agents/divisor-entropy.md`

#### Scenario: Command asset content matches embedded source
- **GIVEN** the embedded asset `assets/commands/vibe-check.md` exists
- **WHEN** `assetContent("commands/vibe-check.md")` is called
- **THEN** the returned bytes SHALL equal the bytes read directly from `commandAssetsFS` at `assets/commands/vibe-check.md`

#### Scenario: Unknown path returns error
- **GIVEN** no embedded asset exists at the requested path
- **WHEN** `assetContent("agents/nonexistent.md")` is called
- **THEN** a non-nil error SHALL be returned

### Requirement: Project root discovery for tests
The scaffold package SHALL provide a function `findProjectRoot()` that walks up from the current working directory to locate the project root by finding `go.mod`. If found, the parent directory of `go.mod` is returned. If not found, an error is returned.

#### Scenario: Project root found
- **GIVEN** the test binary runs from within a Go module tree
- **WHEN** `findProjectRoot()` is called
- **THEN** the function SHALL return the directory containing `go.mod`
- **AND** no error SHALL be returned

#### Scenario: Test run outside a Go module
- **GIVEN** the test binary runs from a directory with no `go.mod` ancestor (e.g., unusual CI setup)
- **WHEN** `findProjectRoot()` is called
- **THEN** the function SHALL return a non-nil error
- **AND** `TestEmbeddedAssetsMatchSource` SHALL skip the test with an informative message

### Requirement: Drift detection test
The scaffold package SHALL include a test `TestEmbeddedAssetsMatchSource` that byte-compares every embedded asset against its deployed copy in `.opencode/<rel>` and fails with a remediation hint when drift is detected. The test is classified as an integration/contract test and SHALL use `testing.Short()` as a skip guard.

#### Scenario: Matching assets pass
- **GIVEN** the project root is found and all deployed `.opencode/` assets are byte-identical to their embedded counterparts
- **WHEN** `TestEmbeddedAssetsMatchSource` runs
- **THEN** the test SHALL pass

#### Scenario: Missing individual asset detected
- **GIVEN** the project root is found but a deployed asset file does not exist at `<projectRoot>/.opencode/<rel>`
- **WHEN** `TestEmbeddedAssetsMatchSource` runs
- **THEN** the test SHALL fail
- **AND** the failure message SHALL indicate the asset is not deployed and include a remediation hint directing the user to run `vibe-check init --force .`

#### Scenario: Content drift detected
- **GIVEN** the project root is found and a deployed asset's bytes differ from its embedded counterpart
- **WHEN** `TestEmbeddedAssetsMatchSource` runs
- **THEN** the test SHALL fail
- **AND** the failure message SHALL indicate content drift was detected and include a remediation hint directing the user to run `vibe-check init --force .`

#### Scenario: Missing .opencode directory
- **GIVEN** the project root is found but no `.opencode/` directory exists
- **WHEN** `TestEmbeddedAssetsMatchSource` runs
- **THEN** the test SHALL fail with a message indicating that scaffold assets are not deployed
- **AND** the failure message SHALL include a remediation hint directing the user to run `vibe-check init --force .`

#### Scenario: Unreadable deployed asset
- **GIVEN** the project root is found and a deployed asset exists but cannot be read (permission denied, I/O error)
- **WHEN** `TestEmbeddedAssetsMatchSource` runs
- **THEN** the test SHALL fail with a message that includes the asset path and the underlying error