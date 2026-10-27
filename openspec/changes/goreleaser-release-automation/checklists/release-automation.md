## Release Automation Checklist

### Pre-Implementation
- [x] `.goreleaser.yaml` uses schema `version: 2`
- [x] Build targets all 6 platform/arch combinations
- [x] `CGO_ENABLED=0` for static linking
- [x] Ldflags: `main.version={{.Tag}}`, `main.commit={{.Commit}}`, `main.date={{.CommitDate}}` with `-s -w` stripping
- [x] Checksums: `checksums.txt` with SHA-256
- [x] Changelog: commit groups (Features, Bug Fixes, Documentation, Others) with `prerelease: auto`
- [x] Archives: `.tar.gz` on linux/darwin, `.zip` on windows

### Release Workflow
- [x] Header comment block (CI-011)
- [x] Trigger: `workflow_dispatch` with `tag` input only (no branch or tag-push trigger)
- [x] Reusable workflow refs SHA-pinned to complytime/org-infra (`reusable_release_preflight.yml`, `reusable_release_goreleaser.yml`) with only the declared inputs
- [x] Workflow-level `permissions: {}` (default-deny); job-level grants exactly mirror the callee-declared scopes at the pinned SHA (preflight: `contents: write` + `checks: read`; release: `contents: write` + `id-token: write`)
- [x] `secrets: inherit` on both jobs

### Verification (Local)
- [x] `goreleaser check` passes
- [x] `go build -ldflags` smoke test reports injected values
- [x] `goreleaser build --snapshot --clean` produces all 6 binaries, checksums, and changelog
- [x] Workflow YAML passes CI convention pack review

### Documentation
- [x] CHANGELOG.md entry with Spec paths
- [x] README.md Pre-built binaries subsection (platform listing, checksums, macOS quarantine)
- [x] AGENTS.md Project Structure updated (.goreleaser.yaml, .github/workflows/)
- [x] Website documentation sync issue filed in unbound-force/website (unbound-force/website#336)