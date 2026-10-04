## ADDED Requirements

### Requirement: GoReleaser schema version
The `.goreleaser.yaml` SHALL use GoReleaser schema `version: 2`.

### Requirement: Cross-platform binary builds
The `.goreleaser.yaml` config SHALL build `cmd/vibe-check` binaries for linux (amd64, arm64), darwin (amd64, arm64), and windows (amd64, arm64). Builds SHALL use `CGO_ENABLED=0` for static linking.

#### Scenario: Build on tag push
- **WHEN** a tag matching `v*` is pushed
- **THEN** GoReleaser SHALL produce statically linked binaries for all platform/arch combinations in the build matrix

### Requirement: Ldflags version embedding
The `.goreleaser.yaml` config SHALL inject ldflags for `main.version={{.Tag}}`, `main.commit={{.Commit}}`, and `main.date={{.CommitDate}}`. Additionally, the build SHALL strip debug symbols with ldflags `-s -w`.

#### Scenario: Version is tag-derived
- **WHEN** a tag `v0.2.0` is pushed
- **THEN** the built binary SHALL report `v0.2.0` via `--version` using GoReleaser's `{{.Tag}}` template (which preserves the `v` prefix)

#### Scenario: Commit is the git SHA
- **WHEN** a build runs from a git commit
- **THEN** the built binary SHALL report the full commit SHA via `--version`

#### Scenario: Date is the commit timestamp
- **WHEN** a build runs from a git commit
- **THEN** the built binary SHALL report the commit date via `--version`

### Requirement: Checksums file generation
The `.goreleaser.yaml` config SHALL generate a `checksums.txt` containing SHA-256 hashes of all release artifacts.

#### Scenario: Checksums included in release
- **WHEN** a tag push triggers a release
- **THEN** a `checksums.txt` file SHALL be uploaded alongside the binaries

### Requirement: Archive format
Release artifacts SHALL be packaged as `.tar.gz` archives on linux and darwin, and `.zip` archives on windows.

### Requirement: Changelog / release notes
The `.goreleaser.yaml` config SHALL include a `changelog` section with commit groups (Features, Bug Fixes, Documentation, Others) and `prerelease: auto` for automatic pre-release detection.

#### Scenario: Release notes are generated with commit groups
- **WHEN** a tag push triggers a release
- **THEN** GoReleaser SHALL group release notes by commit type (Features, Bug Fixes, Documentation, Others)

### Requirement: Archive naming convention
Release artifact archives SHALL follow the naming convention `<binary>_<GOOS>_<GOARCH>.<ext>` where `<ext>` is `.tar.gz` on linux/darwin and `.zip` on windows. The binary inside the archive is named `vibe-check` (or `vibe-check.exe` on windows).

#### Scenario: Linux amd64 binary
- **WHEN** a release is published
- **THEN** the artifact for linux/amd64 SHALL be named `vibe-check_linux_amd64` (no extension)

#### Scenario: Windows amd64 archive
- **WHEN** a release is published
- **THEN** the archive for windows/amd64 SHALL be named `vibe-check_windows_amd64.zip`

#### Scenario: Validating binary is statically linked
- **WHEN** a build completes
- **THEN** the resulting binary SHALL have no dynamic library dependencies, verifiable via `ldd vibe-check` (linux) returning "not a dynamic executable" or `otool -L vibe-check` (darwin) showing no shared library references

#### Scenario: GoReleaser failure
- **WHEN** `goreleaser` exits with a non-zero status
- **THEN** the build SHALL fail with a descriptive error message

### Requirement: End-to-end user accessibility
A user SHALL be able to download a pre-built `vibe-check` binary for their platform from the GitHub Releases page without requiring a Go toolchain.

#### Scenario: User downloads binary without Go
- **WHEN** a release is published
- **THEN** the GitHub Release page SHALL list downloadable archives for all supported platform/arch combinations with checksums

> **Verification**: See design.md § Coverage Strategy and tasks 4.1–4.5.