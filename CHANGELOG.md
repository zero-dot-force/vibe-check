# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
