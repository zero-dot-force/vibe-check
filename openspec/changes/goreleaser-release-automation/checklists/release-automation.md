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
- [ ] Header comment block (CI-011)
- [ ] Trigger: `on: push: tags: ['v*']` only (no branch trigger)
- [ ] `actions/checkout` SHA-pinned with `fetch-depth: 0`
- [ ] `actions/setup-go` SHA-pinned with `go-version-file: go.mod`
- [ ] `goreleaser/goreleaser-action` SHA-pinned with exact `version` pin
- [ ] `permissions.contents: write` on release job only
- [ ] `GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}` environment variable

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