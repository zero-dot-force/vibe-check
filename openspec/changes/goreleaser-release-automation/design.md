## Context

vibe-check is a Go CLI tool (`github.com/zero-dot-force/vibe-check`). `v0.1.0` was tagged but binaries were not published to GitHub Releases. The `cmd/vibe-check/main.go` declares three ldflags variables (`version`, `commit`, `date`) that default to `"dev"`, `"none"`, `"unknown"`. The `root.go:versionString()` function already handles ldflags → `runtime/debug.ReadBuildInfo()` fallback. GoReleaser fills the gap by injecting ldflags and publishing cross-platform artifacts on tag push.

The release workflow delegates to `complytime/org-infra` reusable workflows (`reusable_release_preflight.yml` and `reusable_release_goreleaser.yml`), which handle tag validation, CI check verification, GoReleaser execution, cosign signing, and SBOM generation. Action pinning and permissions are centralized in the reusable workflows. `ci.yml` is a prerequisite for CI check validation in preflight.

## Goals / Non-Goals

**Goals:**
- Build cross-platform binaries for `cmd/vibe-check` (linux/darwin/windows × amd64/arm64) on tag push
- Inject `main.version`, `main.commit`, `main.date` via ldflags so `--version` reports real values
- Publish GitHub Release with generated release notes and `checksums.txt`
- SHA-pinned actions and least-privilege `permissions` throughout

**Non-Goals:**
- Homebrew tap, Scoop bucket, Nix flake, or Docker image publishing
- Nightly / snapshot builds from `main`
- macOS notarization (signing is handled by cosign via the reusable release workflow)
- Changing how `versionString()` or ldflags variables are declared

## Decisions

### D1: GoReleaser configuration in `.goreleaser.yaml`

**Choice:** Single GoReleaser config at repository root.

**Alternatives considered:**
- **Multiple configs (`.goreleaser/darwin.yaml`, etc.)** — overkill for a single binary. One config is simpler to maintain.
- **CI-only build via `go build` matrix** — GoReleaser handles changelog generation, checksums, and GitHub Release creation out of the box. Rebuilding these features manually would be wasted effort.

### D2: Release trigger via `workflow_dispatch`

**Choice:** Workflow triggers `on: workflow_dispatch` with a `tag` input (e.g., `v0.2.0`). The org-infra `reusable_release_preflight.yml` validates the tag, verifies CI checks, and creates the annotated tag via the GitHub API.

**Rationale:** This is the pattern used by `unbound-force/gaze` and other Unbound Force repos. It centralizes tag validation, semver ordering checks, and CI gate enforcement in the preflight workflow rather than relying on manual `git tag && git push`. The maintainer triggers the release from the GitHub Actions UI or via `gh workflow run` rather than pushing a tag directly.

### D3: Ldflags injection in GoReleaser, not in CI matrix

**Choice:** Let GoReleaser's `builds[].ldflags` set the variables using its built-in templates (`{{.Tag}}`, `{{.Commit}}`, `{{.CommitDate}}`).

**Rationale:** GoReleaser knows the tag version, commit SHA, and commit date. Injecting these via its template system is the canonical approach. `{{.Tag}}` is used instead of `{{.Version}}` to preserve the `v` prefix (producing `v0.2.0`), consistent with `versionString()` expectations in `cmd/vibe-check/root.go` and the existing `v0.1.0` tag convention.

### D4: Delegate to org-infra reusable workflows

**Choice:** The release workflow has two jobs: `preflight` (calling `complytime/org-infra/.github/workflows/reusable_release_preflight.yml@main`) and `release` (calling `complytime/org-infra/.github/workflows/reusable_release_goreleaser.yml@main`). Top-level `permissions: {}`; each reusable workflow declares its own permissions.

**Rationale:** This is the pattern used by other Unbound Force repos (`gaze`, etc.). The reusable workflows handle tag validation, CI check verification, GoReleaser execution, cosign signing, SBOM generation, and action pinning. vibe-check only provides the tag input and `.goreleaser.yaml` config.

### D5: SHA pinning handled by org-infra reusable workflows

**Choice:** Action SHAs are centralized in the org-infra reusable workflows. vibe-check does not pin actions directly — it references the reusable workflows by branch (`@main` for preflight, `@v1` for release).

**Rationale:** Centralized pinning reduces maintenance burden and ensures all Unbound Force repos stay in sync with security updates. The reusable workflows are maintained by the org-infra team.

### D6: Build patterns adopted from `unbound-force/gaze`

**Choice:** Adopt proven GoReleaser patterns from the `gaze` project (same ecosystem, same author): schema `version: 2`, `CGO_ENABLED=0` for static linking, ldflags stripping (`-s -w`) for smaller binaries, changelog commit groups (Features, Bug Fixes, Documentation, Others), and `prerelease: auto` for automatic pre-release detection.

**Rationale:** `gaze` is a Go CLI tool in the Unbound Force ecosystem with a mature GoReleaser pipeline. Its patterns are battle-tested. `CGO_ENABLED=0` ensures pure Go statically linked binaries compatible with all platforms. `-s -w` stripping reduces binary size without affecting functionality.

**Alternatives considered:**
- **Standard defaults** — GoReleaser defaults without CGO_ENABLED, stripping, or changelog groups produce larger binaries and less useful release notes.
- **Org-infra reusable workflows** — `gaze` uses complytime/org-infra reusable workflows (`reusable_release_preflight.yml`, `reusable_release_goreleaser.yml`). These are not applicable to vibe-check (different GitHub org). vibe-check's release workflow is self-contained.

## Risks / Trade-offs

- **[workflow_dispatch requires manual trigger]** → A maintainer must go to the GitHub Actions UI or use `gh workflow run` to trigger a release. → Mitigation: This is intentional — the preflight workflow validates the tag and runs CI checks before creating the tag, which tag-push triggers cannot do before the push.
- **[No notarization on macOS]** → macOS users see "unidentified developer" warnings. → Mitigation: Expected for open-source CLI tools. Binaries are cosign-signed by the reusable release workflow. Notarization can be added later.
- **[GoReleaser version in reusable workflow]** → The reusable workflow defaults to `~> v2`. → Mitigation: vibe-check can override with a `goreleaser_version` input if needed. The reusable workflow also pins the `goreleaser/goreleaser-action` SHA.
- **[org-infra dependency]** → A breaking change in the org-infra reusable workflows could break vibe-check releases. → Mitigation: The reusable workflows are versioned (`@main` for preflight, `@v1` for release). vibe-check can pin to specific SHAs if needed.

## Coverage Strategy

This change introduces no new Go source code. Configuration validation is covered by:
- `goreleaser check` — YAML syntax validation (task 3.1)
- `go build -ldflags ...` — ldflags injection smoke test (task 3.2)
- `goreleaser build --snapshot --clean` — cross-platform build + artifact verification (task 3.4)
- Manual workflow pattern review (task 3.3)
- The release workflow itself serves as the integration test on first tag push.

## Documentation Impact

User-facing infrastructure change — requires:
- `CHANGELOG.md` entry for release automation
- `README.md` Install section updated with pre-built binary download instructions
- `AGENTS.md` Project Structure updated with new files
- Website documentation sync issue in `unbound-force/website`