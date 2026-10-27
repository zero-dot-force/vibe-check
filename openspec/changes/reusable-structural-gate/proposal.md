# Proposal

## Why

`.github/workflows/structural-gate.yml` is vibe-check's own self-gate: it has no
`workflow_call` trigger and builds vibe-check from the PR's own source, so it can
only dogfood the gate on vibe-check itself. Other repositories that want the
structural gate today must copy ~60 lines of workflow YAML and build vibe-check
from source on every run. The distribution channel already exists — the
GoReleaser release automation (#23) publishes binaries plus a `checksums.txt` to
GitHub Releases (v0.2.0 is current) — and the gate semantics are stable and
protected (ΔInstability ≥ 0.15, ΔDistance ≥ 0.20, ΔLCOM ≥ 2, or any new
circular dependency → `REQUEST_CHANGES`). This change ships the vibe-check
equivalent of the reusable-workflow treatment gaze already shipped
(unbound-force/gaze#300): a reusable `workflow_call` workflow any repo can
consume via `uses:`, running the released binary instead of building from
source.

## What Changes

- Add a reusable workflow (`.github/workflows/reusable_structural_gate.yml` —
  underscore prefix per convention CI-010; the issue's suggested hyphenated
  name is superseded by that MUST rule) triggered via `workflow_call` that:
  - Installs the **released** vibe-check binary selected by a `version` input
    (release tag; default `latest`), verified against the release
    `checksums.txt`. There is no build-from-source path.
  - Checks out the caller's PR head and base revisions (base into a separate
    path, mirroring `structural-gate.yml`) and runs
    `vibe-check analyze --output` on each.
  - Gates with `vibe-check diff --gate base.json head.json`, preserving the
    protected thresholds and the `--gate` exit semantics; optional tighten-only
    threshold override inputs are forwarded to `diff`.
  - Reports by appending `diff --markdown` to `$GITHUB_STEP_SUMMARY`
    (`if: always()`, `continue-on-error: true`).
  - Provisions the Go toolchain via `setup-go` with a `go-version-file` input,
    because `analyze` forces `GOTOOLCHAIN=local`.
- Consume the reusable workflow from vibe-check's own CI as a dogfood smoke
  test (additive job; the build-from-source self-gate remains vibe-check's
  authoritative gate so PR-source analyzer changes stay dogfooded).
- Document the consumer contract (usage snippet, inputs, permissions, trust
  model, version pinning) in `README.md`, and add a `CHANGELOG.md` entry.
- File a documentation issue in `unbound-force/website` for a CI-integration
  guide section (gaze `docs/guides/ci-integration.md` reusable-workflow
  pattern; cf. website#296, website#336, website#311).
- No production Go source changes: the gate semantics in `metrics/` and
  `cmd/vibe-check` are consumed as-is.

## Capabilities

### New Capabilities

- `reusable-structural-gate`: A `workflow_call`-triggered reusable CI workflow
  that any repository can consume via `uses:` to install the released
  vibe-check binary (version-pinned, checksum-verified), analyze the caller's
  PR head and base revisions, enforce the structural-regression gate with the
  protected thresholds, and publish the markdown diff report as a step summary
  — with no secrets required and a documented trust model.

### Modified Capabilities

<!-- No existing main specs to modify (openspec list --specs is empty); the
     self-gate's own behavior is unchanged. -->

## Impact

- **New files**: `.github/workflows/reusable_structural_gate.yml`.
- **Modified files**: `.github/workflows/ci.yml` (dogfood consumer job),
  `README.md` (CI-integration consumer-contract section),
  `CHANGELOG.md` (change entry).
- **No production source changes**: `metrics.ComputeDelta`/`DecideVerdict`,
  the protected thresholds, and the `diff`/`analyze` command surfaces are
  reused unmodified; no API, schema, or dependency changes.
- **External**: documentation issue filed against `unbound-force/website`;
  follow-up consumer is vibe-check's own CI (dogfood), with org-repo adoption
  (gaze's `unbound-force/unbound-force#752` pattern) as the follow-on.
- **Security posture**: the reusable workflow declares `contents: read` only,
  requires no secrets, and is fork-PR safe; `vibe-check analyze` compiles the
  caller's target module, the same trust posture as running Go tests on a fork
  PR in the consumer's CI. The trust model is documented in the workflow
  header and README.
- **Related**: #56 (originating issue), #8 (gate engine), #23 (release
  automation), PR #40 (self-gate), RFC discussion #483; precedent
  unbound-force/gaze#300 and consumer unbound-force/unbound-force#752.
