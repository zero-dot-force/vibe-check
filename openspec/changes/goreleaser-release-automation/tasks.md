<!-- Tasks marked [P] are eligible for parallel execution. Unmarked tasks
run sequentially. When parallel workers are not available, [P] tasks
execute sequentially but remain eligible for future parallelization. -->

## 1. GoReleaser Configuration

- [x] 1.1 Create `.goreleaser.yaml` at repository root with `version: 2` schema, `CGO_ENABLED=0` for static linking, and build targets `cmd/vibe-check` for linux/darwin/windows × amd64/arm64
- [x] 1.2 Configure ldflags injection: `main.version={{.Tag}}`, `main.commit={{.Commit}}`, `main.date={{.CommitDate}}` with `-s -w` stripping. Use `{{.Tag}}` to preserve the `v` prefix for `--version` output.
- [x] 1.3 Configure checksum generation, archive naming (`.tar.gz` on linux/darwin, `.zip` on windows), and changelog with commit groups (Features, Bug Fixes, Documentation, Others) and `prerelease: auto`

## 2. CI Workflow (prerequisite)

- [x] 2.0 Create `.github/workflows/ci.yml` with build, test, vet, and lint steps for PRs and pushes to `main` — establishes the SHA-pinning and least-privilege conventions that the release workflow follows.

## 3. Release Workflow

- [x] 3.1 Create `.github/workflows/release.yml` with a header comment block (CI-011), triggering on `workflow_dispatch` with a `tag` input (string, required). Top-level `permissions: {}`.
- [x] 3.2 Add `preflight` job calling `complytime/org-infra/.github/workflows/reusable_release_preflight.yml@main` with `tag: ${{ inputs.tag }}` and `secrets: inherit`.
- [x] 3.3 Add `release` job (needs `preflight`) calling `complytime/org-infra/.github/workflows/reusable_release_goreleaser.yml@v1` with `tag: ${{ needs.preflight.outputs.tag }}` and `secrets: inherit`.

## 4. Verification

- [x] 4.1 Run `goreleaser check` locally to validate `.goreleaser.yaml` syntax
- [x] 4.2 Verify `go build -ldflags "-X main.version=v0.2.0 -X main.commit=abc1234 -X main.date=2026-10-04T12:00:00Z" ./cmd/vibe-check && ./vibe-check --version` reports the injected values
- [x] 4.3 Review `.github/workflows/release.yml`: confirm `workflow_dispatch` trigger with `tag` input, correct org-infra reusable workflow references (`@main` for preflight, `@v1` for goreleaser), header comment (CI-011), and top-level `permissions: {}`
- [x] 4.4 Run `goreleaser build --snapshot --clean` locally to verify cross-platform builds produce correctly named binaries, `checksums.txt`, and changelog output with correct commit grouping

## 5. Documentation

- [x] 5.1 Add `CHANGELOG.md` entry under `[Unreleased]### Added` for release automation, including `Spec:` lines for both `openspec/changes/goreleaser-release-automation/specs/goreleaser-config/spec.md` and `openspec/changes/goreleaser-release-automation/specs/release-workflow/spec.md`
- [x] 5.2 Update `README.md` Install section: add a "Pre-built binaries" subsection documenting binaries for linux/darwin/windows × amd64/arm64 on GitHub Releases, with a link to the latest release, `checksums.txt` verification note, and macOS quarantine workaround (`xattr -d com.apple.quarantine` or right-click → Open)
- [x] 5.3 Update `AGENTS.md` Project Structure: add `.goreleaser.yaml` at the repository root level, and add a `.github/workflows/` entry (with `ci.yml`, `release.yml`, and `structural-gate.yml`)
- [x] 5.4 Create a GitHub issue in `unbound-force/website` documenting the new binary installation method for the installation page

<!-- spec-review: passed -->
<!-- code-review: passed -->