# Spec Delta

## Purpose

Lets `vibe-check init --ci` scaffold a correctly pinned, advisory structural-gate caller workflow into any repository's `.github/workflows/`, so adopting the reusable structural gate takes one command instead of a hand-edited copy-paste with a version-pinning footgun.

## ADDED Requirements

### Requirement: Additive CI scaffolding flag

`vibe-check init` SHALL accept a `--ci` flag that additionally emits the structural-gate caller workflow; without `--ci`, the command's writes, output, and exit codes SHALL be identical to the existing behavior (`.opencode/` assets and UF registration only, nothing under `.github/`).

#### Scenario: Plain init is unchanged

- **WHEN** `vibe-check init` runs without `--ci`
- **THEN** nothing is created under the target's `.github/` directory and the deployment summary is unchanged

#### Scenario: CI scaffolding is emitted alongside assets

- **WHEN** `vibe-check init --ci` runs on a project
- **THEN** the `.opencode/` assets and UF registration behave exactly as without `--ci` and the caller workflow is additionally emitted

### Requirement: Fail-closed release pin resolution

The caller's pins SHALL be resolved from the running binary's own release identity (or an explicit `--vibe-version` override): the release tag, validated as a `v`-prefixed semantic version `>= v0.3.0` (the first release containing the reusable workflow), and the release commit's full 40-character SHA — taken from the binary's embedded release commit when running a release build, or resolved from the tag at init time otherwise. An unresolvable, unparseable, or below-floor value for either pin SHALL exit 2 with a message naming the offending value and the remediation, before any filesystem writes occur.

#### Scenario: Valid release pins the caller

- **WHEN** the binary reports a valid release tag `>= v0.3.0` (e.g. `v0.3.0`, `v0.4.1`) and the release commit SHA is available (embedded in a release binary, or resolved from the tag)
- **THEN** the caller workflow is emitted with its `uses:` reference pinned to that commit SHA and its `version` input pinned to that tag

#### Scenario: Dev or unparseable version fails closed

- **WHEN** the binary's version resolves to `dev`, empty, a non-`v`-prefixed, or otherwise unparseable value
- **THEN** the command exits 2 with an error naming the resolved value and suggesting installing from a release tag or passing `--vibe-version`, and no files are written

#### Scenario: Below-floor version fails closed

- **WHEN** the resolved or overridden version is a valid semver below `v0.3.0` (e.g. `v0.2.9`)
- **THEN** the command exits 2 before any filesystem writes

#### Scenario: Unresolvable commit SHA fails closed

- **WHEN** the tag is valid but the release commit SHA cannot be resolved (no embedded commit, and tag resolution unavailable, e.g. offline)
- **THEN** the command exits 2 with an error naming the tag and the remediation (install from a release binary, or rerun with network access), and no files are written

#### Scenario: Version override honored

- **WHEN** `--vibe-version` is passed with `--ci` and a valid `>= v0.3.0` tag
- **THEN** the caller workflow is emitted pinned to the overridden tag's commit SHA (`uses:`) and tag (`version`) instead of the binary's own release

#### Scenario: Override without the CI flag is a usage error

- **WHEN** `--vibe-version` is passed without `--ci`
- **THEN** the command fails with a usage error and performs no filesystem writes

### Requirement: Self-contained pinned caller workflow

The emitted workflow SHALL be a single job that calls the vibe-check reusable structural-gate workflow, with the `uses:` reference pinned to the resolved release commit's full 40-character SHA and the `with: version:` input pinned to the resolved release tag, written to `.github/workflows/vibe-check-structural-gate.yaml`.

#### Scenario: Single reusable-workflow job

- **WHEN** the emitted workflow is inspected
- **THEN** it contains exactly one job, which calls the reusable workflow and declares no local steps, test jobs, artifacts, or secrets

#### Scenario: Both pins name the same release

- **WHEN** the emitted workflow is inspected
- **THEN** the `uses:` reference carries the release commit's full 40-character SHA and the `with: version:` input carries the release tag, so workflow logic and analyzer binary move together and the reference is immutable (SHA-pinned, not tag-pinned)

#### Scenario: Advisory defaults with documented tighten-only overrides

- **WHEN** the emitted workflow is inspected
- **THEN** it sets no threshold inputs, and a commented-out block documents the three tighten-only overrides (`max-instability-delta`, `max-distance-delta`, `max-lcom-delta`) and the protected defaults they tighten beyond

### Requirement: Caller trigger, concurrency, and permissions

The emitted workflow SHALL trigger only on `pull_request` targeting `main` (a hardcoded `branches: [main]` filter — repositories whose default branch differs must edit the filter; a documented limitation surfaced in the command's next-steps output and the README's known-limitations entry), SHALL own a workflow-scoped concurrency group that cancels superseded runs, and SHALL grant exactly the reusable workflow's declared `contents: read` permission — no broader.

#### Scenario: Pull-request-only trigger

- **WHEN** the emitted workflow is inspected
- **THEN** its `on:` block contains only `pull_request` with a `branches: [main]` filter, so the base-revision context is always present without an event-shape guard

#### Scenario: Caller-owned concurrency group

- **WHEN** the emitted workflow is inspected
- **THEN** it declares a concurrency group keyed on the workflow ref with `cancel-in-progress: true`

#### Scenario: Least-privilege permissions

- **WHEN** the emitted workflow is inspected
- **THEN** its permissions block grants only `contents: read`, mirroring the callee's declaration

### Requirement: Skip-if-present caller protection

An existing caller workflow file SHALL never be overwritten — a rerun with the file present SHALL report it as skipped and leave its contents unchanged, regardless of `--force`; `--force` SHALL continue to apply only to `.opencode/` asset overwrites.

#### Scenario: Rerun reports the caller as skipped

- **WHEN** `vibe-check init --ci` runs in a project that already contains the caller workflow
- **THEN** the file's contents are unchanged and it is reported as skipped

#### Scenario: Skip survives force

- **WHEN** `vibe-check init --ci --force` runs in a project that already contains the caller workflow
- **THEN** the caller workflow is still skipped unchanged while `.opencode/` assets are overwritten per the existing `--force` semantics

### Requirement: Language-neutral manifest notice

The emitted caller SHALL be language-neutral (passing no `go-version-file` input, so the reusable workflow's default applies), and when no language manifest is detected at the target root, the command SHALL emit a non-blocking notice that the gate currently analyzes Go modules — and still emit the caller workflow.

#### Scenario: No manifest at target root

- **WHEN** `vibe-check init --ci` runs against a directory containing no `go.mod`
- **THEN** a notice explains that the reusable gate currently analyzes Go modules and that future language adapters will reuse the same caller workflow, and the caller workflow is still emitted

#### Scenario: Language-neutral caller

- **WHEN** the emitted workflow is inspected
- **THEN** it passes no `go-version-file` input to the reusable workflow

### Requirement: CI scaffolding surfaced in command output

The init command's output SHALL surface the caller workflow: the human summary SHALL include a CI next-steps section (commit the workflow; the gate is advisory until branch protection marks the check required; uncomment thresholds to tighten; edit `branches:` when the repository's default branch is not `main`; bump pins by editing the file, since reruns skip), and the `--json` payload SHALL list the workflow path (prefixed `workflows/`) in `written` or `skipped`.

#### Scenario: JSON payload includes the workflow

- **WHEN** `vibe-check init --ci --json` runs and the caller is written (or skipped on rerun)
- **THEN** the JSON payload's `written` (or `skipped`) array contains `workflows/vibe-check-structural-gate.yaml`

#### Scenario: Summary includes CI next steps

- **WHEN** `vibe-check init --ci` runs with human-readable output
- **THEN** the summary includes a CI next-steps section covering committing the workflow, the advisory-until-required posture, tightening thresholds, editing `branches:` for non-`main` default branches, and the manual pin-bump upgrade path
