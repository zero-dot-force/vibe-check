## 1. CLI — `diff --gate` Flag and Exit Code

- [x] 1.1 Add a `Gate bool` field to `DiffOptions` in `cmd/vibe-check/diff.go` with a GoDoc comment describing gate-mode exit semantics.
- [x] 1.2 Update `RunDiff`'s final return in `cmd/vibe-check/diff.go` to set `ExitCode: 1` when `opts.Gate && verdict == metrics.VerdictRequestChanges`, else `0` (tool-failure exit 2 paths unchanged).
- [x] 1.3 Add the `--gate` boolean flag to `diffCmd()` and wire it into `DiffOptions.Gate`.
- [x] 1.4 Update the `DiffResult.ExitCode` GoDoc (diff.go ~54-58), the `RunDiff` doc comment (~87-94), and the `diffCmd` `Long` help (~326-346) to document the `0`/`1`/`2` exit taxonomy under `--gate`.

## 2. Tests

- [x] 2.1 Add `TestRunDiff_GateExitCodes`: table-driven, driving `RunDiff` with `bytes.Buffer`; assert `Gate: true` maps `REQUEST_CHANGES`→`1`, `APPROVE`→`0`, `COMMENT`→`0`, and `Gate: false` always `0`.
- [x] 2.2 Add `TestDiffCommand_GateFlagWiring`: invoke `diffCmd` with `--gate` and a REQUEST_CHANGES fixture; assert `*exitCodeError{code: 1}`; assert without `--gate` the same fixture exits `0`.
- [x] 2.3 Extend/confirm `TestDiffCommand_ExitCodes` (diff_test.go ~871) still asserts default (no `--gate`) REQUEST_CHANGES exits `0`.

## 3. CI Workflow

- [x] 3.1 Add `.github/workflows/structural-gate.yml`: `on: pull_request` → `main`; `permissions: contents: read`; concurrency group `structural-gate-${{ github.ref }}` with `cancel-in-progress: true`.
- [x] 3.2 Add a fork guard: skip the gate (no `analyze`/`diff` step) when `github.event.pull_request.head.repo.full_name != github.repository`, so forked PRs are not analyzed (per the `divisor-entropy` constraint — `analyze` compiles the module and would execute untrusted build-time code).
- [x] 3.3 In the job (non-fork path): `actions/checkout` (head, pinned commit SHA), second `actions/checkout` of `${{ github.event.pull_request.base.sha }}` into `./base`, `actions/setup-go` (pinned, `go-version-file: go.mod`), `go build -o bin/vibe-check ./cmd/vibe-check`.
- [x] 3.4 Run `bin/vibe-check analyze --output head.json .` and `bin/vibe-check analyze --output base.json ./base`, then `bin/vibe-check diff --gate base.json head.json`.
- [x] 3.5 Local dry-run validation of the workflow steps (build, two analyze snapshots, `diff --gate`) to confirm the base-path analysis mechanics before relying on CI.

## 4. Validation (CI parity)

- [x] 4.1 Verify `go build ./...` passes.
- [x] 4.2 Verify `go test -race -count=1 ./...` passes.
- [x] 4.3 Verify `go vet ./...` passes.
- [x] 4.4 Run `golangci-lint run ./...` locally (the local CLI equivalent of `ci.yml`'s `golangci/golangci-lint-action` at v2.12.2) and fix any findings.
- [x] 4.5 Verify all exported symbols have GoDoc comments.

## 5. Documentation

- [x] 5.1 Update `README.md` diff section: document the `--gate` flag and the `0`/`1`/`2` exit-code semantics, and reference `.github/workflows/structural-gate.yml`.
- [x] 5.2 Add a `CHANGELOG.md` `[Unreleased] → Added` entry with `Spec: openspec/changes/ci-regression-gate/`.
- [x] 5.3 File a documentation issue against this repository for the user-facing `--gate` flag and new CI workflow (AGENTS.md Documentation gate).
- [x] 5.4 File a website documentation-sync issue in `unbound-force/website` for the `--gate` flag (Constitution Website Documentation Sync).
- [x] 5.5 Document that the gate is advisory: merges are only blocked once a repo admin enables branch protection and marks `structural-gate` required (design.md R3 follow-up; reflected in the README and the workflow header comment).

<!-- spec-review: passed -->
<!-- code-review: passed -->
