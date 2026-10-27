# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.0] - 2026-10-27

### Changed

- One byte-stable architectural health report now shared across the CLI, the CI
  step summary, and the `vibe-check-reporter` agent: report formatting, A–F
  health grades, and prioritized recommendations moved out of agent prompt text
  into pure, table-tested Go functions in the `metrics/report/` package — a
  grade engine with fixed threshold tables, aggregate computations, and a
  deterministic recommendation rule engine (ordered by severity, metric
  extremity, then package path; capped at 5). Identical input renders
  byte-identical output: no timestamps, modules in path order.
- New `vibe-check report <graph.json>` subcommand renders the markdown report
  from a previously emitted ModuleGraph snapshot: `--mode summary|detailed`
  (default summary) and `--json` (mutually exclusive with `--mode`) emitting
  the computed report model — aggregates, per-dimension grades, overall
  status, fired recommendations — as a single JSON object. Exit codes: 0
  clean, 1 any red/yellow recommendation (the full report is still written to
  stdout), 2 unreadable or schema-invalid input. `vibe-check diff --markdown`
  renders the verdict, triggering reasons, entropy direction, per-module
  deltas, and new/resolved cycles as a byte-stable markdown report instead of
  the default table — mutually exclusive with `--json` and display-only, never
  a gate input. The structural-gate workflow's new Report step appends the
  `diff --markdown` report to the PR step summary (`if: always()`,
  `continue-on-error`, inside the existing fork-guarded job; the gate step is
  untouched).
- The `vibe-check-reporter` agent asset is rewritten to `vibe-check analyze
  --output <tmp.json>` → `vibe-check report --mode <mode>` → emit the
  Go-rendered markdown verbatim, appending per-violation explanations and
  remediation suggestions only on exit 1; its bash allowlist grows from 2 to
  3 entries (`vibe-check report *`), and the package-pattern invocation bug
  is fixed (`analyze` takes a directory, not package patterns).
  `.opencode/references/vibe-check-example-report.md` is regenerated from the
  committed fictional fixture with the new renderer.
  Spec: `openspec/changes/deterministic-report/`
- Revamped `vibe-check-reporter` agent output format: emoji-structured reports
  with closed vocabulary (🏗️📊🔗🧩🔄📋📖🏥🟢🟡🔴⚠️), full metric names
  (Instability, Abstractness, Distance, LCOM4), legend block with metrics guide
  link, health scorecard with letter grades (A–F), prioritized recommendations,
  and duplication analysis section. Summary mode uses traffic-light indicators.
  Trending mode updated with emoji and legend support.
- New `docs/metrics-guide.md` — plain-language ELI5 reference for all
  architectural metrics with emoji severity ranges and further reading links.
- New `.opencode/references/vibe-check-example-report.md` — formatting
  reference for the agent output contract.

### Added

- `vibe-check analyze --max-duplication <pct>` flag: fails the build (exit code 1)
  when any module's duplicated-line percentage exceeds the given threshold.
  Duplication detection identifies structurally similar code blocks within a
  module after identifier and literal normalization. Results are reported in the
  new `duplications` array on each `ModuleResult`, with each entry carrying a
  `modulePath`, `blocks` (file, startLine, endLine, lineCount), and `similarity`
  score in [0, 1]. A new `totalLines` field on `ModuleResult` records the total
  source lines per module for percentage computation. The ModuleGraph schema is
  bumped to `1.3` (additive; `1.0`, `1.1`, and `1.2` remain valid).
  Spec: `openspec/changes/add-code-duplication-detection/`

- Release automation via GoReleaser: cross-platform pre-built binaries for
  linux/darwin/windows × amd64/arm64 are published to GitHub Releases via a
  `workflow_dispatch`-triggered pipeline using org-infra reusable workflows.
  Ldflags embed `main.version` (`{{.Tag}}`), `main.commit`, and `main.date`
  so `--version` reports the release version. Binaries are statically linked
  (`CGO_ENABLED=0`) with stripped debug symbols (`-s -w`). Release notes are
  auto-generated with commit groups (Features, Bug Fixes, Documentation,
  Others).
  Spec: `openspec/changes/goreleaser-release-automation/specs/goreleaser-config/`
  Spec: `openspec/changes/goreleaser-release-automation/specs/release-workflow/`

- `vibe-check diff --gate` adds a CI-regression gate: when `--gate` is set, a
  `REQUEST_CHANGES` verdict exits `1` (instead of `0`) so CI can block a PR on
  structural regression; `APPROVE` and `COMMENT` still exit `0`, and tool failures
  still exit `2`. The verdict and reasons remain in the payload. A new
  `.github/workflows/structural-gate.yml` workflow runs `analyze` on both the PR
  head and base (skipping forked PRs) and then `diff --gate` on the
  `pull_request` event.
  Spec: `openspec/changes/ci-regression-gate/`
- `vibe-check analyze` now emits a top-level `provenance` object in its JSON
  output recording how the graph was produced: `producer` (`"vibe-check"`),
  `version` (the tool's semantic version), `generatedAt` (an RFC 3339 UTC
  timestamp), and `input` (`path` and `modulePath`). A new `--no-provenance`
  flag omits it for byte-reproducible output. The ModuleGraph schema is bumped
  to `1.2` (additive; `1.0` and `1.1` remain valid), and `ComputeDelta`/
  `vibe-check diff` ignore provenance entirely.
  Spec: `openspec/changes/emit-provenance-metadata/`
- `vibe-check diff --json` and `vibe-check init --json` now emit a top-level
  `provenance` object (`producer` `"vibe-check"`, `version`, and `generatedAt`)
  mirroring `analyze`, sourced from a shared `cmd/vibe-check/provenance.go`
  helper (`producerName` constant + `newProvenanceEnvelope()`). Each command
  gains a `--no-provenance` flag to omit the envelope for byte-reproducible
  output. The `analyze` command was refactored to source producer/version/
  generatedAt from the same shared helper (its `input` sub-object is unchanged).
  Spec: `openspec/changes/add-diff-init-provenance/`
- Agent-design convention pack (`.opencode/uf/packs/agent-design.md`) defining
  10 structural quality rules (AD-001 through AD-010) covering coupling,
  cohesion, complexity, naming, file size, duplication, contract coverage, and
  test assertion depth. Enforcement mapped to vibe-check, gaze, golangci-lint,
  and review agents. Includes `agent-design-custom.md` placeholder for
  project-level overrides.
- `docs/agent-design-pack.md` — standalone human-readable reference for the
  10 agent-design rules (AD-001–AD-010) covering rule reference table,
  enforcement tool matrix, override mechanism, forward references, and LCOM4
  semantics.
  Spec: `openspec/changes/doc-agent-design-pack/specs/agent-design-reference/`
- `vibe-check diff <base.json> <pr.json>` compares two ModuleGraph JSON
  snapshots and reports the structural-entropy delta (per-module Ca, Ce,
  instability, abstractness, distance, and LCOM deltas), new and resolved
  circular dependencies, an entropy direction (improving/stable/
  degrading), and a verdict (APPROVE/COMMENT/REQUEST_CHANGES). Protected
  gate thresholds — ΔInstability ≥ 0.15, ΔDistance ≥ 0.20, ΔLCOM ≥ 2, or a
  new circular dependency — yield REQUEST_CHANGES; smaller non-zero shifts
  yield COMMENT; improving or stable yields APPROVE. Exit code is 0
  whenever both inputs are valid (the verdict travels in the payload, not
  the exit code); exit code 2 is reserved for missing, unreadable, or
  schema-invalid input and for a looser-than-default (non-tightening)
  threshold override. `--json` emits a machine-readable payload; the
  `--max-instability-delta`, `--max-distance-delta`, and
  `--max-lcom-delta` overrides are tighten-only.
- `vibe-check init [path]` deploys the embedded Review Council agent
  assets into `.opencode/agents/` — skipping existing files by default,
  overwriting with `--force`, and printing a machine-readable summary of
  written/skipped/forced files with `--json`. Path defaults to `.`.
- `divisor-entropy` Review Council agent (embedded source of truth,
  deployed by `vibe-check init`): measures the base↔PR design-quality
  delta via `vibe-check analyze` and `vibe-check diff` inside an isolated
  git worktree and reports a verdict. Runs only on trusted refs.
- `/vibe-check` slash command asset (deployed by `vibe-check init` to
  `.opencode/commands/vibe-check.md`): delegates to the `vibe-check-reporter`
  agent for conversational architectural analysis with three modes —
  summary (traffic-light health indicator), detailed (per-package
  breakdown), and trending (longitudinal metric comparison via Dewey
  snapshots).
  Spec: `openspec/changes/vibe-check-command-and-reporter/`
- `vibe-check-reporter` agent asset (deployed by `vibe-check init` to
  `.opencode/agents/vibe-check-reporter.md`): interprets Martin coupling
  metrics in natural language, runs `vibe-check analyze` to gather data,
  and stores metric snapshots in Dewey for trend tracking.
  Spec: `openspec/changes/vibe-check-command-and-reporter/`
- `vibe-check init` now deploys command assets to `.opencode/commands/`
  alongside agent assets in `.opencode/agents/`. The scaffold system
  uses a `deployCategory` helper to iterate both asset categories with
  the same symlink-safe, containment-checked pattern.
  Spec: `openspec/changes/vibe-check-command-and-reporter/`
- `vibe-check analyze --output <file>` (`-o`) writes the ModuleGraph JSON
  to a file instead of stdout (stdout remains the default; a failed write
  exits with code 2 and a stderr diagnostic without emitting partial
  stdout).

- Drift detection test (`TestEmbeddedAssetsMatchSource`) in `internal/scaffold/`
  verifies that every asset embedded via `//go:embed` matches its deployed copy
  in `.opencode/`. Two new helpers, `assetPaths()` (enumerates embedded asset
  paths) and `assetContent()` (retrieves embedded content by relative path),
  support the test. The test is gated behind `testing.Short()` and provides a
  remediation hint (`vibe-check init --force .`) on failure.
- Architectural drift tracking: `vibe-check analyze --store` enriches
  provenance with git metadata (`commitSHA`, `branch`, `modulePath`) for
  CI-driven snapshot archival. A new `mx-f-architecture-trend` agent
  stores/retrieves snapshots via Dewey MCP, computes windowed linear
  regression trends (7/30/90-day) with R² confidence gating, renders
  Unicode block-character sparklines, and generates drift alerts for
  sustained degradation and projected threshold crossings. Retention
  policies prune daily snapshots > 90 days and weekly-Sunday snapshots
  > 1 year. The `/vibe-check trending` mode is extended from single-
  baseline to multi-snapshot time-series when ≥2 snapshots exist.
  Spec: `openspec/changes/architectural-drift-tracking/`

### Changed

- `vibe-check analyze` now forces `GOTOOLCHAIN=local` for its
  `go/packages` subprocess, so analysis never downloads a toolchain named
  by a target module's `go.mod`. Trade-off: a trusted module whose
  `go.mod` `toolchain` directive requires a newer-than-local Go must be
  built/analyzed manually.
- Strengthened test assertions across `internal/goadapter` and `metrics/`
  packages: added nil-guard assertions for `goadapter.New()` in 5 tests,
  added return-value assertions for `goadapter.Analyze()`, and added
  explicit `(*limitedBuffer).String()` content assertions. Per-function
  contract coverage rose from 94.4% to 98.2% (no function at 0%).
  Spec: `openspec/changes/contract-coverage-gap/`
- Enhanced GoDoc annotations on `ExternalAdapter.Analyze`,
  `ExternalAdapter.Capabilities`, `Registry.Register`, `RunAnalyze`,
  `RunDiff`, `RunInit`, and `(*limitedBuffer).Write` to explicitly
  document return-value semantics and mutation contracts.

### Fixed

- Release workflow dispatch: the `release.yml` org-infra reusable workflow
  references used hyphenated filenames that never existed upstream, so
  `workflow_dispatch` failed with HTTP 422 ("workflow was not found"), and an
  undeclared `go-version-file` input was passed to the preflight workflow.
  Both jobs now call the underscore filenames verified at the pinned
  org-infra SHA (`reusable_release_preflight.yml`,
  `reusable_release_goreleaser.yml`) with only the declared `tag` input,
  restoring the ability to cut releases.
  Spec: `openspec/changes/goreleaser-release-automation/specs/release-workflow/`

## [0.1.0] - 2026-08-31

Initial release: package-level design-quality and architectural metrics
for Go.

### Added

- `vibe-check analyze [path]` computes the Martin coupling metrics suite
  for a Go module — afferent coupling (Ca), efferent coupling (Ce),
  instability, abstractness, distance from the main sequence, LCOM4
  cohesion, and circular-dependency detection — and prints the results
  as JSON on stdout (path defaults to `.`).
- Go language adapter (`internal/goadapter/`) implementing the
  `metrics.Adapter` interface, using `golang.org/x/tools/go/packages`
  for type-aware dependency resolution.
- CI gate threshold flags — `--max-instability`, `--max-distance`,
  `--max-lcom`, and `--no-circular-deps` — so pipelines can fail the
  build (exit code 1) when a module breaches an architectural budget.
  All violations are reported before exit, and JSON is still written to
  stdout.
- `--timeout` flag to bound analysis time (no timeout by default).
- `--version` reporting version, commit, and build date. Binaries
  installed via `go install ...@v0.1.0` (built without ldflags) fall
  back to `runtime/debug` build info, so the reported version stays
  meaningful.
- Go-specific metric extensions surfaced under the JSON `extensions`
  object: `go.interfaceWidth` (method count per exported interface) and
  `go.interfaceProximity` (producer/consumer classification).
- `Extensions map[string]any` field on `ModuleResult` for
  language-specific extension metrics.
- CI workflow (`.github/workflows/ci.yml`) running build, vet, test
  (with `-race` and a coverage profile), and lint, using SHA-pinned
  GitHub Actions.

### Changed

- Output JSON schema version bumped from `1.0` to `1.1` to carry the
  optional `extensions` object. Backward-compatible; existing 1.0
  consumers need no migration.
- `metrics.Validate()` accepts both schema versions `1.0` and `1.1`,
  and now enforces numeric ranges (instability, abstractness, and
  distance in [0, 1]; counts non-negative).
- The external-analyzer trust boundary (`ExternalAdapter.Analyze`) now
  rejects unknown fields via a strict JSON decoder, hardening it against
  malformed subprocess output.

[Unreleased]: https://github.com/zero-dot-force/vibe-check/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/zero-dot-force/vibe-check/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/zero-dot-force/vibe-check/releases/tag/v0.1.0
