## Why

Issue #8 (P1) requires CI gates that detect regression in package-level coupling
metrics between the base branch and a PR, failing the build when structural
quality degrades beyond tolerance. The `vibe-check diff` command already computes
the exact per-package delta and a protected verdict (`APPROVE`/`COMMENT`/
`REQUEST_CHANGES`) via `metrics.ComputeDelta` + `metrics.DecideVerdict` — but it
exits `0` for every valid input, encoding the verdict only in the payload. CI
therefore cannot fail on a `REQUEST_CHANGES` verdict. The missing pieces are a
gate-mode exit code and CI wiring; no new metric or verdict engine is required.

## What Changes

- Add a `--gate` flag to `vibe-check diff` that maps a `REQUEST_CHANGES` verdict
  to exit code `1` (policy failure), mirroring `analyze`'s established `0`/
  `1`/`2` taxonomy (`0` success, `1` policy failure, `2` tool failure).
  `APPROVE` and `COMMENT` verdicts continue to exit `0` under `--gate`. Without
  `--gate`, behavior is unchanged: all valid inputs exit `0`.
- Add a new CI workflow `.github/workflows/structural-gate.yml` that runs on
  `pull_request` into `main`: build the binary, `analyze` the PR head and the
  base SHA into snapshots, then run `diff --gate` and fail the check when the
  verdict is `REQUEST_CHANGES`.
- Document the `--gate` flag and the new workflow in `README.md` and
  `CHANGELOG.md`.

No breaking changes: `--gate` is opt-in and default `diff` behavior is preserved.

## Capabilities

### New Capabilities

<!-- None: the CI workflow is a build/deployment artifact, not a runtime
     capability with normative requirements and scenarios. Its behavior is
     captured in What Changes, design.md, and tasks.md. -->

### Modified Capabilities

- `diff-command`: the "Input Validation and Exit Codes" requirement gains
  gate-mode semantics — when `--gate` is set, a `REQUEST_CHANGES` verdict exits
  `1` instead of `0`. (The existing `diff-command` spec lives in the
  `add-divisor-entropy-agent` change and is not yet consolidated under
  `openspec/specs/`; this change produces the delta against it.)

### Removed Capabilities

<!-- None. -->

## Impact

- **Code**: `cmd/vibe-check/diff.go` (`DiffOptions.Gate`, `--gate` flag, exit-code
  branch in `RunDiff`); `cmd/vibe-check/diff_test.go` (gate exit-code and flag
  wiring tests). No changes to `metrics/` — `ComputeDelta`/`DecideVerdict` and
  their protected thresholds (`ΔI ≥ 0.15`, `ΔD ≥ 0.20`, `ΔLCOM ≥ 2`, new cycle)
  are reused as-is.
- **CI**: new `.github/workflows/structural-gate.yml` (pull_request only,
  `contents: read`, pinned actions).
- **APIs**: `diff` JSON/human output is unchanged; only the exit code changes,
  and only under `--gate`. The `--gate` flag is additive.
- **Docs**: `README.md` diff section (exit-code table/prose) and `CHANGELOG.md`.

## Constitution Alignment

| Principle | Assessment |
|-----------|------------|
| I. Autonomous Collaboration | N/A — no new agent artifacts; the gate signal is consumed by CI, not a new agent. |
| II. Composability First | N/A — no new dependencies or mandatory coupling between agents. |
| III. Observable Quality | PASS — structural regression becomes an automated, reproducible CI gate; the verdict is made machine-actionable via an exit code. |
| IV. Testability | PASS — gate exit-code behavior is unit-tested through the existing testable CLI pattern (`bytes.Buffer`); coverage strategy is defined in design.md. |
| V. Security by Default | PASS — workflow pins `actions/checkout` and `actions/setup-go` by commit SHA, uses least-privilege `contents: read`, holds no secrets, and skips forked PRs (which would otherwise let `analyze`'s compile step execute untrusted build-time code) per the `divisor-entropy` constraint. |
| VI. Metric Fidelity | PASS — no metric computation changes; the change reuses `ComputeDelta`/`DecideVerdict`, preserving determinism and protected thresholds. |
| VII. Language Agnosticism | N/A — CLI-layer change only; the core engine and adapter interface are unchanged. |
