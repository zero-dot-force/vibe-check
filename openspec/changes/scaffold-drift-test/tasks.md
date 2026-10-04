## 1. Asset enumeration and content helpers

- [x] 1.1 [P] Add `assetPaths()` to `internal/scaffold/scaffold.go` that walks both `agentAssetsFS` and `commandAssetsFS` via `fs.WalkDir`, strips the `assets/` prefix from each path, and returns a sorted slice of relative paths.
- [x] 1.2 [P] Add `assetContent(relPath string)` to `internal/scaffold/scaffold.go` that prepends `assets/` to the relative path, tries reading from `agentAssetsFS` first then `commandAssetsFS`, and returns the bytes or an error.

## 2. Drift detection test

- [x] 2.1 Add `findProjectRoot()` to `internal/scaffold/scaffold_test.go` that walks up from the current working directory until it finds `go.mod`, returning the project root directory. If no `go.mod` ancestor is found, return an error (the test skips with an informative message).
- [x] 2.2 Add `TestEmbeddedAssetsMatchSource` to `internal/scaffold/scaffold_test.go` that: calls `assetPaths()`, reads each embedded asset via `assetContent()`, reads the corresponding deployed file at `<projectRoot>/.opencode/<rel>`, and byte-compares them. On missing file or content mismatch, fail with distinct messages that include the remediation hint `run 'vibe-check init --force .' from the repo root`. Use `testing.Short()` as a skip guard (integration test classified per design D6). On unreadable files, fail with the path and underlying error.

## 3. Deploy missing assets and dogfood

- [x] 3.1 Run `go run ./cmd/vibe-check init --force .` from the repo root to deploy missing assets (`vibe-check-reporter.md`, `vibe-check.md`) into `.opencode/`.
- [x] 3.2 Generate a before-change `ModuleGraph` snapshot: `go run ./cmd/vibe-check analyze -o /tmp/vibe-check-base.json ./...`.
- [x] 3.3 After all implementation is complete, generate an after-change snapshot (`/tmp/vibe-check-pr.json`), then run `go run ./cmd/vibe-check diff /tmp/vibe-check-base.json /tmp/vibe-check-pr.json` to verify the scaffold changes produce no structural entropy regression (APPROVE verdict expected).

## 4. CI parity

- [x] 4.1 Run `go build ./...` and confirm clean build.
- [x] 4.2 Run `go vet ./...` and confirm no issues.
- [x] 4.3 Run `go test -race -count=1 -cover ./...` and confirm all tests pass (including `TestEmbeddedAssetsMatchSource`) with ≥80% statement coverage in `internal/scaffold/`.
- [x] 4.4 Run `golangci-lint run ./...` and confirm no new issues.

## 5. Documentation gate

- [x] 5.1 Assess documentation impact: add a changelog entry under `[Unreleased]` for the new drift detection test and deployed missing assets; confirm `AGENTS.md` and `README.md` require no updates (internal change, no structural or description changes).

<!-- spec-review: passed -->
<!-- code-review: passed -->