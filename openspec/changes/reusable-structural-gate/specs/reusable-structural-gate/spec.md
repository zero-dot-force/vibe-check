# Spec Delta

## Purpose

A `workflow_call`-triggered reusable CI workflow that any repository can consume via `uses:` to run the vibe-check structural-regression gate using the released binary — installing it version-pinned and checksum-verified, analyzing the caller's PR head and base, enforcing the protected thresholds, and publishing the markdown report as a step summary, with no secrets and a documented trust model.

## ADDED Requirements

### Requirement: Reusable invocation via workflow_call
The workflow SHALL be invocable via the `workflow_call` trigger from any repository through a `uses:` reference, and SHALL NOT build vibe-check from source.

#### Scenario: Cross-repository consumption
- **WHEN** a caller workflow references the reusable workflow via a `uses:` reference pinned to a commit SHA
- **THEN** the gate job runs in the caller's repository context against the caller's pull request

#### Scenario: No build-from-source path
- **WHEN** the gate job runs
- **THEN** vibe-check is obtained exclusively from a published GitHub Release; no step checks out or builds vibe-check's own source

### Requirement: Released binary installation with checksum verification
The workflow SHALL install the released vibe-check binary for the runner platform, selected by a `version` input (a release tag; default latest), and SHALL verify the downloaded archive against the release's `checksums.txt` before any analysis runs.

#### Scenario: Pinned version
- **WHEN** the `version` input is set to an existing release tag
- **THEN** the binary published under that tag is installed

#### Scenario: Latest by default
- **WHEN** the `version` input is left at its default
- **THEN** the binary from the latest published release is installed

#### Scenario: Checksum mismatch fails the job
- **WHEN** the downloaded archive's digest does not match the release's `checksums.txt`
- **THEN** the job fails before any analysis step runs

### Requirement: Snapshot acquisition via caller revision checkout
The workflow SHALL check out the caller's pull-request head revision and base revision (the base into a separate path) and run `vibe-check analyze --output` on each, producing two ModuleGraph snapshots; a `path` input (default `.`) SHALL select the caller's module directory within the repository, and a `timeout` input SHALL bound each analysis.

#### Scenario: Both revisions analyzed
- **WHEN** the gate job runs on a pull request
- **THEN** base and head ModuleGraph snapshots exist as inputs to the gate step

#### Scenario: Module at a subpath
- **WHEN** the caller sets `path` to a subdirectory of the repository
- **THEN** the head analysis runs against that subdirectory and the base analysis against the same subdirectory under the base checkout

#### Scenario: Analysis failure
- **WHEN** `analyze` fails on either revision
- **THEN** the gate step does not run and the job fails

### Requirement: Gate enforcement with protected thresholds
The workflow SHALL run `vibe-check diff --gate` on the base and head snapshots with the protected default thresholds; a `REQUEST_CHANGES` verdict SHALL fail the job (exit 1) while `APPROVE` and `COMMENT` pass.

#### Scenario: Structural regression
- **WHEN** the diff verdict is `REQUEST_CHANGES`
- **THEN** the gate step exits 1 and the job fails

#### Scenario: Clean or comment-worthy change
- **WHEN** the diff verdict is `APPROVE` or `COMMENT`
- **THEN** the gate step exits 0

#### Scenario: Tool failure
- **WHEN** either snapshot is missing, unreadable, or schema-invalid
- **THEN** the diff exits 2 and the job fails as a tool failure rather than a gate verdict

### Requirement: Tighten-only threshold override inputs
The workflow SHALL expose optional threshold-override inputs forwarded to `diff --gate`; values looser than the protected defaults SHALL be rejected by the tool's tighten-only enforcement, and the workflow SHALL NOT weaken or bypass that enforcement.

#### Scenario: Tighter thresholds forwarded
- **WHEN** a caller supplies threshold overrides tighter than the protected defaults
- **THEN** the overrides are forwarded to the gate and the gate applies them

#### Scenario: Looser override rejected
- **WHEN** a caller supplies a threshold override looser than the protected default
- **THEN** the diff exits 2 and the job fails without the gate being weakened

### Requirement: Step-summary report
The workflow SHALL append the `diff --markdown` report for the base-vs-head snapshots to `$GITHUB_STEP_SUMMARY` in a step that runs even when the gate step fails, with render failures non-blocking.

#### Scenario: Failing gate
- **WHEN** the gate step exits non-zero
- **THEN** the step summary still contains the markdown diff report

#### Scenario: Passing gate
- **WHEN** the gate step passes
- **THEN** the step summary contains the markdown diff report

#### Scenario: Analysis failure
- **WHEN** the analysis step failed and a snapshot is missing
- **THEN** nothing is appended to the step summary and the gate job's failure remains the authoritative conclusion

### Requirement: Go toolchain provisioning
The workflow SHALL provision a Go toolchain satisfying the caller module's `go.mod` before analysis, via a `go-version-file` input (default `go.mod`), because `analyze` forces `GOTOOLCHAIN=local` and never downloads a toolchain itself.

#### Scenario: Caller go.mod satisfied
- **WHEN** the caller module's `go.mod` declares a Go version
- **THEN** that version is installed before the analysis steps run

### Requirement: Minimal permissions and no secrets
The workflow SHALL declare only the `contents: read` permission, SHALL require no secrets, and SHALL document its trust model in the workflow header: `analyze` compiles the caller's target module — the same trust posture as running Go tests on a fork PR in the consumer's CI.

#### Scenario: No secrets required
- **WHEN** a caller invokes the workflow without passing any secrets
- **THEN** the gate runs to completion

#### Scenario: Fork pull request
- **WHEN** the gate runs on a pull request whose head comes from a fork
- **THEN** it runs with the same read-only, no-secrets posture — the workflow itself imposes no fork guard, and callers MAY add their own skip conditions

### Requirement: Consumer contract documentation
The README SHALL document the consumer contract: a usage snippet with a SHA-pinned `uses:` reference, the input table (`version`, `path`, `timeout`, threshold overrides, `go-version-file`), the required caller permissions (including the job-level `contents: read` grant needed by default-deny callers), the trust model, and version-pinning guidance; a documentation issue SHALL be filed in `unbound-force/website` for the CI-integration guide.

#### Scenario: Inputs documented
- **WHEN** a reader consults the README section
- **THEN** every workflow input is listed with its type, default, and effect

#### Scenario: Default-deny caller guidance
- **WHEN** a caller's workflow declares a top-level `permissions: {}`
- **THEN** the documentation states the job-level `contents: read` grant required for the called workflow's jobs to start

#### Scenario: Website documentation issue
- **WHEN** the change is implemented
- **THEN** an issue exists in `unbound-force/website` covering the reusable-workflow CI-integration guide section

### Requirement: Dogfood consumer in vibe-check CI
vibe-check's own CI SHALL consume the reusable workflow on pull requests as an end-to-end smoke test, additive to the existing build-from-source self-gate, which SHALL remain unchanged and authoritative.

#### Scenario: Dogfood job on pull requests
- **WHEN** a pull request opens against vibe-check
- **THEN** a CI job consumes the reusable workflow end-to-end via a same-repo `uses:` reference

#### Scenario: Self-gate preserved
- **WHEN** the dogfood job is added
- **THEN** the existing `structural-gate.yml` build-from-source gate continues to run with its fork guard, thresholds, and report step unchanged
