# Release Workflow

[See `openspec/changes/goreleaser-release-automation/specs/goreleaser-config/spec.md` for the GoReleaser config specification.]

## ADDED Requirements

### Requirement: Release workflow trigger
The release workflow SHALL trigger via `workflow_dispatch` with a `tag` input string. The workflow file SHALL include a header comment block describing its purpose (CI-011). Top-level `permissions` SHALL be `{}` (permissions are scoped within each reusable workflow).

#### Scenario: Release is manually dispatched
- **WHEN** a maintainer triggers the workflow via `workflow_dispatch` with a tag input (e.g., `v0.2.0`)
- **THEN** the preflight job SHALL call `complytime/org-infra/.github/workflows/reusable_release_preflight.yml@<commit-sha>` with the tag input

#### Scenario: Push to main does not trigger
- **WHEN** a commit is pushed to `main`
- **THEN** the release workflow SHALL NOT execute

### Requirement: Preflight job
The workflow SHALL include a `preflight` job calling `complytime/org-infra/.github/workflows/reusable_release_preflight.yml@<commit-sha>` with `tag: ${{ inputs.tag }}`. The preflight workflow validates semver format, checks tag uniqueness, verifies CI checks on HEAD, validates semver ordering against the latest existing tag, verifies unreleased commits exist, and creates an annotated tag via the GitHub API. The job SHALL pass `secrets: inherit` to forward `GITHUB_TOKEN` and any other secrets to the reusable workflow.

#### Scenario: Preflight succeeds
- **WHEN** the tag is a valid semver (`vX.Y.Z`) that is greater than the latest existing tag and unreleased commits exist
- **THEN** the preflight job SHALL output `tag` (the validated tag string) and `tag_created` (whether a new tag was created)

#### Scenario: Preflight fails due to invalid tag
- **WHEN** the `tag` input is not valid semver (e.g., `not-a-tag`)
- **THEN** the preflight job SHALL fail with a descriptive error, and the release job SHALL NOT execute

#### Scenario: Preflight fails due to CI check failure
- **WHEN** required CI checks have not passed on HEAD
- **THEN** the preflight job SHALL fail, and no tag SHALL be created

### Requirement: GoReleaser release job
The workflow SHALL include a `release` job that depends on `preflight` and calls `complytime/org-infra/.github/workflows/reusable_release_goreleaser.yml@<commit-sha>` with `tag: ${{ needs.preflight.outputs.tag }}`. The reusable workflow handles: checkout of the tag ref, Go setup from `go.mod`, cosign signing, syft SBOM generation, and GoReleaser execution with `release --clean --verbose`. The job SHALL pass `secrets: inherit` to forward `GITHUB_TOKEN` for release creation and artifact upload.

#### Scenario: Release publishes successfully
- **WHEN** the preflight job succeeds and the `release` job executes
- **THEN** GoReleaser SHALL build cross-platform binaries per `.goreleaser.yaml`, generate a `checksums.txt`, sign artifacts with cosign, generate an SBOM, and publish a GitHub Release with generated release notes

#### Scenario: GoReleaser failure
- **WHEN** GoReleaser fails (config error, network error, build failure, permission denied)
- **THEN** the workflow SHALL exit with a non-zero status and the release SHALL NOT be created

### Requirement: End-to-end user accessibility
This requirement is defined in the `goreleaser-config` specification.

### Requirement: Cross-reference
This workflow SHALL be consistent with the release workflow pattern used by other Unbound Force repos (`unbound-force/gaze`), delegating to org-infra reusable workflows for preflight validation and GoReleaser execution.