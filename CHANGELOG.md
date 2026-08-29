# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `vibe-check analyze` CLI command for Go package-level coupling
  metrics analysis. Computes afferent coupling (Ca), efferent coupling
  (Ce), instability, abstractness, distance from main sequence, LCOM4
  cohesion, and circular dependency detection.
- Go language adapter (`internal/goadapter/`) implementing the
  `metrics.Adapter` interface. Uses `golang.org/x/tools/go/packages`
  for type-aware dependency resolution.
- CI gate threshold flags: `--max-instability`, `--max-distance`,
  `--max-lcom`, `--no-circular-deps`. Non-zero exit code on threshold
  violations.
- `--timeout` flag for analysis time bounds.
- `--version` flag with build metadata (version, commit, date).
- Go-specific extensions: `go.interfaceWidth` (method count per
  exported interface) and `go.interfaceProximity` (producer/consumer
  classification).
- `Extensions map[string]any` field on `ModuleResult` for
  language-specific extension metrics.
- JSON schema version bumped from 1.0 to 1.1 to accommodate the
  extensions field.
- `metrics.Validate()` now accepts both schema versions 1.0 and 1.1.
- CI workflow (`.github/workflows/ci.yml`) with build, test, vet,
  and lint steps using SHA-pinned GitHub Actions.
