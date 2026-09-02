# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Agent-design convention pack (`.opencode/uf/packs/agent-design.md`) defining
  10 structural quality rules (AD-001 through AD-010) covering coupling,
  cohesion, complexity, naming, file size, duplication, contract coverage, and
  test assertion depth. Enforcement mapped to vibe-check, gaze, golangci-lint,
  and review agents. Includes `agent-design-custom.md` placeholder for
  project-level overrides.
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

### Changed

- `vibe-check analyze` now forces `GOTOOLCHAIN=local` for its
  `go/packages` subprocess, so analysis never downloads a toolchain named
  by a target module's `go.mod`. Trade-off: a trusted module whose
  `go.mod` `toolchain` directive requires a newer-than-local Go must be
  built/analyzed manually.

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

[Unreleased]: https://github.com/zero-dot-force/vibe-check/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/zero-dot-force/vibe-check/releases/tag/v0.1.0
