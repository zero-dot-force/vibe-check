## Release Automation Checklist

### Pre-Implementation
- [ ] `.goreleaser.yaml` uses schema `version: 2`
- [ ] Build targets all 6 platform/arch combinations
- [ ] `CGO_ENABLED=0` for static linking
- [ ] Ldflags: `main.version={{.Tag}}`, `main.commit={{.Commit}}`, `main.date={{.CommitDate}}` with `-s -w` stripping
- [ ] Checksums: `checksums.txt` with SHA-256
- [ ] Changelog: commit groups (Features, Bug Fixes, Documentation, Others) with `prerelease: auto`
- [ ] Archives: `.tar.gz` on linux/darwin, `.zip` on windows

### Release Workflow
- [x] Header comment block (CI-011)
- [x] Trigger: `workflow_dispatch` with `tag` input only (no branch or tag-push trigger)
- [x] Reusable workflow refs SHA-pinned to complytime/org-infra (`reusable_release_preflight.yml`, `reusable_release_goreleaser.yml`) with only the declared inputs
- [x] Workflow-level `permissions: {}` (default-deny); job-level grants exactly mirror the callee-declared scopes at the pinned SHA (preflight: `contents: write` + `checks: read`; release: `contents: write` + `id-token: write`)
- [x] `secrets: inherit` on both jobs

### Verification (Local)
- [ ] `goreleaser check` passes
- [ ] `go build -ldflags` smoke test reports injected values
- [ ] `goreleaser build --snapshot --clean` produces all 6 binaries, checksums, and changelog
- [ ] Workflow YAML passes CI convention pack review

### Documentation
- [ ] CHANGELOG.md entry with Spec paths
- [ ] README.md Pre-built binaries subsection (platform listing, checksums, macOS quarantine)
- [ ] AGENTS.md Project Structure updated (.goreleaser.yaml, .github/workflows/)
- [ ] Website documentation sync issue filed in unbound-force/website