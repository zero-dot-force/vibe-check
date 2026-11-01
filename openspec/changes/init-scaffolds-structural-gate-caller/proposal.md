# Proposal

## Why

The reusable structural gate (`reusable_structural_gate.yml`, shipped in v0.3.0) is the
product — any Go repo on GitHub can get an advisory design-quality regression gate on
pull requests with zero secrets and zero configuration — but adopting it today means
copy-pasting the consumer example out of the README "CI integration" section, and the
biggest footgun in that copy-paste is version pinning: the `uses:` ref must point at a
release tag that actually contains the workflow (the reusable workflow first ships in
v0.3.0), and a consumer pinned to an older tag fails only at run time with a confusing
"workflow not found" error. A deterministic installer that knows its own version
eliminates this failure class entirely. Precedent: unbound-force/gaze#306 shipped the
identical feature for gaze's AI-report workflow (`gaze init --ci`); this change ports
the pattern to vibe-check's structural gate.

## What Changes

- Add an additive `--ci` flag to `vibe-check init` (design D1): plain `vibe-check init`
  behavior stays byte-identical (`.opencode/` assets + UF registration, nothing under
  `.github/`); `--ci` additionally emits the caller workflow. No `--ci-only` variant.
- The generated workflow is a self-contained starter (D2): a single job calling the
  reusable workflow `zero-dot-force/vibe-check/.github/workflows/reusable_structural_gate.yml`,
  with the `uses:` ref pinned to the release commit's full 40-character SHA and the
  `with: version:` input pinned to the release tag (D3, as amended by maintainer
  input in the spec-review decision gate — see design.md's decision log), advisory
  by default with a commented-out tighten-only threshold `TODO` block (D4), a
  `pull_request`-only trigger with a caller-owned concurrency group (D9), and
  exactly the callee's declared `contents: read` permission.
- Resolve both pins from the binary's own release identity, fail closed (D6, as
  amended): the tag via `semanticVersion()` in `cmd/vibe-check/root.go` (ldflags →
  `debug.ReadBuildInfo` fallback) and the release commit SHA via the
  ldflags-injected `main.commit` (release-binary path, fully offline) or
  `git ls-remote refs/tags/<tag>^{}` at init time (`go install` / `--vibe-version`
  paths), validated — tag as a `v`-prefixed semver `>= v0.3.0`, SHA as a full
  40-character hex string — **before any filesystem writes**; dev builds,
  unparseable values, below-floor versions, or an unresolvable (offline) SHA
  exit 2. A `--vibe-version` override flag carries the same validation; passing
  it without `--ci` is a usage error.
- The generated file is skip-if-present even under `--force` (D7) — users edit
  thresholds in this file, so it must never be clobbered; `--force` keeps its existing
  `.opencode/`-only semantics. A rerun with the file present reports it as skipped.
- Keep the caller forward-compatible with future language adapters (D8):
  language-neutral template (no `go-version-file` pin), a small manifest-detection seam
  (`go.mod` → go today; future adapters extend a table), and a notice + still-emit when
  no manifest is found at the target root.
- Name both the generated caller and this repo's own self-gate
  `vibe-check-structural-gate.yaml` (D5): the self-gate rename
  (`.github/workflows/structural-gate.yml` →
  `.github/workflows/vibe-check-structural-gate.yaml`, via `git mv`, contents
  byte-identical) leaves its `name:` field (`Structural Gate — Design Quality
  Regression`) and job names unchanged, so branch protection (check names come from
  `name:`/job names, not the file) and Actions run history are unaffected. Update path
  references at `AGENTS.md` (project-structure listing), `README.md` (link), `ci.yml`
  (comments), and `reusable_structural_gate.yml` (comment); historical artifacts
  (archived/unarchived `openspec/changes/*`, old CHANGELOG entries,
  `.uf/dewey/learnings/`) are deliberately untouched.
- Embed the template at
  `internal/scaffold/assets/ci/vibe-check-structural-gate.yaml` (rendered by simple
  substitution of two distinct placeholders — `__VIBE_CHECK_VERSION_SHA__` for the
  `uses:` ref and `__VIBE_CHECK_VERSION_TAG__` for the `with: version:` pin, each
  appearing exactly once), exposed through a new `//go:embed assets/ci/*.yaml`
  filesystem that is excluded from the existing categories table, `assetPaths()`,
  `assetContent()`, and the embedded-asset mirror test (CI templates have no
  `.opencode/` mirror semantics).
- Wire `cmd/vibe-check/init.go`: `--ci` and `--vibe-version` flags, tag + commit-SHA
  resolution before any writes, a new CI next-steps section in the summary output,
  and a `--json` payload that includes the workflow path (under a `workflows/`
  prefix) in `written`/`skipped`.
- Promote `golang.org/x/mod` (semver validation) from indirect to direct and add
  `go.yaml.in/yaml/v3` (test-only YAML parse of the rendered output) as a direct
  require in `go.mod`; no version changes — x/mod's zip is already in `go.sum`,
  yaml/v3's is fetched once at first build (its `h1:` line is new; see design D13's
  corrected dependency facts).
- Document and gate: README gains an "Install the caller workflow" subsection at the
  top of "## CI integration" — its two-pin explanation reconciled with the existing
  "### Version pinning" section (SHA for the `uses:` ref, tag for the `version`
  input; the emitted artifact now matches the published consumer guidance) — plus
  known-limitations entries for the hardcoded `branches: [main]` filter (edit when
  the default branch differs) and the root-only manifest assumption, and
  `--ci`/`--vibe-version` in the init section; `AGENTS.md` project structure gains
  the `assets/ci/` line and the renamed workflow (its build-commands init comment
  and Architecture Layer-3 init description gain the new flags); `CHANGELOG.md`
  gains `Added` (feature) and `Changed` (rename) entries; a user-facing
  documentation issue is filed against `unbound-force/website` before merge (new
  user-facing flag — no exemption applies), alongside the repo-local
  documentation-gate issue AGENTS.md requires for user-facing changes.

## Capabilities

### New Capabilities

- `init-ci-scaffolding`: `vibe-check init --ci` scaffolds the structural-gate caller
  workflow — resolving and validating its own release tag fail-closed before any
  writes, rendering a pinned, advisory, self-contained caller from an embedded
  template, skipping an existing (possibly hand-edited) caller even under `--force`,
  reporting a language-neutral manifest notice, and surfacing the workflow in the
  command's human and JSON output.

### Modified Capabilities

<!-- No existing main specs to modify (openspec list --specs is empty); plain
     `vibe-check init` behavior is unchanged (D1: additive flag, byte-identical
     default path). -->

## Impact

- **New files**: `internal/scaffold/assets/ci/vibe-check-structural-gate.yaml`
  (embedded template).
- **Modified files**: `internal/scaffold/scaffold.go` (`Options.CI`,
  `Options.VibeVersion`, `Options.VibeCommit`, pin validation,
  `deployCIWorkflow`, manifest seam), `internal/scaffold/embed.go`
  (`ciAssetsFS`), `cmd/vibe-check/init.go` (flags, tag + SHA resolution wiring,
  next-steps section, JSON payload), `go.mod`/`go.sum` (x/mod promoted
  indirect → direct; yaml/v3 added direct with its first-fetched zip hash),
  `.github/workflows/structural-gate.yml` (renamed to `vibe-check-structural-gate.yaml`,
  contents byte-identical), `AGENTS.md`, `README.md`, `CHANGELOG.md`, plus
  comment-level path-reference updates in `.github/workflows/ci.yml` and
  `reusable_structural_gate.yml`.
- **Dependencies**: `golang.org/x/mod v0.39.0` promoted from indirect to direct
  (its module zip hash is already in `go.sum` — no download);
  `go.yaml.in/yaml/v3 v3.0.4` added as a direct require (in the module graph
  via `golang.org/x/tools` but never fetched — the test-only import performs
  one module-zip download and adds go.sum's `h1:` line); no version changes, no
  other new modules.
- **Behavioral surface**: plain `vibe-check init` output and exit codes stay
  byte-identical; `--ci` is additive. Exit 2 gains a new cause
  (unresolvable/invalid tag or commit SHA) that fires before any filesystem
  writes — atomic, no partial state.
- **Security posture**: the generated caller grants exactly the callee's declared
  `contents: read`, pins its `uses:` reference to an immutable full-length
  commit SHA (Constitution Principle V / CI-001, per the amended D3), uses no
  secrets, and is fork-safe by construction (the reusable workflow's design);
  pin validation prevents emitting a caller pinned to a tag that predates the
  reusable workflow.
- **Coverage**: all touched packages already have tests and committed floors; floors
  must not drop (coverage-ratchet is tighten-only).
- **External**: documentation issues — `unbound-force/website` (per the originating
  issue's documentation gate) and the repo-local documentation-gate issue AGENTS.md
  requires for user-facing changes.
- **Related**: #77 (originating issue), `openspec/changes/reusable-structural-gate`
  (the callee this change scaffolds callers for), RFC discussion #483; precedent
  unbound-force/gaze#306.

## Constitution Alignment

Assessed against the vibe-check constitution (parent: unbound-force org
constitution v1.2.0); added during the spec-review decision gate (the section
was missing from the initial proposal).

| Principle | Assessment |
|-----------|------------|
| I. Autonomous Collaboration | PASS — the change is artifact-based end to end: the decision record lives in `openspec/changes/`, the template is a versioned embedded asset, and the `init --json` payload (workflow path in `written`/`skipped`, provenance envelope) keeps the scaffold's output self-describing for downstream consumers. |
| II. Composability First | PASS — `--ci` is additive and composes with `--force`/`--json` exactly as today; the manifest-detection seam extends a table rather than coupling the caller to the Go adapter; plain `init` stays byte-identical, so no consumer is forced onto the new path. |
| III. Observable Quality | PASS — the emitted caller is deterministic (golden-file tested), pin resolution is fail-closed before any writes, and the generated gate itself makes design-quality regressions observable in consumer CI. |
| IV. Testability | PASS — pin-resolution matrix, golden-file, YAML-parse, skip-if-present, atomicity, and plain-init regression tests are tasked; the tag→SHA resolution is an injectable seam so `RunInit` stays unit-testable offline. |
| V. Security by Default | PASS (as amended) — the original D3 pinned the `uses:` ref to a mutable tag, deviating from this principle and CI-001; the maintainer-approved amendment pins `uses:` to the full 40-character release commit SHA (ldflags-embedded and offline on the release-binary path, `git ls-remote`-resolved otherwise) with `with: version:` on the tag, matching the README's published consumer guidance. Least privilege (`contents: read` only), no secrets, fork-safe. |
| VI. Metric Fidelity | N/A — scaffolding emits no metrics; the gate's protected thresholds are untouched (tighten-only inputs stay commented out per D4). |
| VII. Language Agnosticism | PASS — the caller is language-neutral (no `go-version-file` pin) and manifest detection is a seam designed for the planned Python/TypeScript adapters (D8). |
