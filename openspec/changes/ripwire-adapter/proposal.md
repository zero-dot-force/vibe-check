# Proposal

## Why

Vibe-Check computes the full Martin metrics suite for Go only. The universal model
(Layer 1) and the `metrics.Adapter` interface were built for multi-language analysis,
but no Python or TypeScript adapter exists — the two languages the RFC phases as P2/P3
and the two most common in the Unbound Force ecosystem. Building native,
type-accurate adapters per language is high-cost; ripwire
(https://github.com/redhat-et/ripwire) already observes multi-language structural
data, and the Phase 0 spike (#67, GO 2026-10-30) validated its Python and TypeScript
surfaces as GOOD-ENOUGH for coupling metrics. ripwire observes; vibe-check computes
and judges. Because the adapter is optional and PATH-discovered, vibe-check releases
stay zero-dependency and Go analysis stays native and type-accurate.

## What Changes

- Add an optional Layer-2 adapter package `internal/ripwireadapter/` implementing
  `metrics.Adapter` for the languages `python` and `typescript`, sourcing
  multi-language structural data from a PATH-discovered `ripwire` binary —
  **alongside, never replacing** `internal/goadapter`.
- Hybrid data sourcing per the spike's capability matrix: Ca/Ce/Instability/cycles
  (both languages) and TypeScript Abstractness come from ripwire (`--arch` metrics
  table, `--deps` XML, `--export=cc.json`); LCOM4 field usage (both languages) and
  Python Abstractness come from vibe-check's own light language parse — mirroring
  `internal/goadapter`, which already owns LCOM4/type classification via its own AST.
- CLI: `vibe-check analyze` detects the project language (go.mod / Python markers /
  TypeScript markers) and selects the adapter. Python/TypeScript analysis requires
  `ripwire` on PATH; an absent binary yields a clear error with install guidance
  (exit 2). Go-project behavior is unchanged.
- Version pinning: the adapter validates ripwire's version against the
  spike-validated surface (0.6.9 @ `bedc0cd`) and fails closed on mismatch (upstream
  is 0.x with no semver policy). Provenance records the ripwire version and commit.
- Deterministic mapping rules to the universal `ModuleGraph` with mandatory
  `Warning` emissions for approximations: crawl-denylist pre-check, cycle-listing
  caps, measured error budgets (Python import graph 1.9% lost / 2.2% spurious; TS
  0.16% unresolved), name-resolved edges, ambiguity rates, and the
  directory-as-module unit.
- Operational guards from the spike: crawl from the package root (never the repo
  root — src-layout repo-root crawls silently yield `ca=ce=0`); read Ca/Ce/I from
  the `--arch` metrics-table aggregates, not the 40-capped per-file `<inc>` lists;
  `--deps` XML parsing is adapter-internal; `--export=cc.json` supplies per-file
  fan-in/out.
- Golden fixtures per language under `testdata/` plus two-run byte-equal determinism
  tests; the test suite passes without ripwire installed (fixture replay), so CI
  stays zero-dependency.
- No new Go module dependencies (stdlib `encoding/xml`, `encoding/json` only).
- Protected values untouched: verdict thresholds (ΔI ≥ 0.15, ΔD ≥ 0.20, ΔLCOM ≥ 2,
  new-cycle detection), the `modulegraph.schema.json` schema file, and the report
  engine all stay in Layer 1, unchanged. Provenance input gains optional
  `omitempty` fields recording the external source tool (following the existing
  `commitSHA`/`branch` precedent from architectural-drift-tracking).

## Capabilities

### New Capabilities

- `ripwire-adapter`: Optional, PATH-discovered ripwire-backed Layer-2 adapter for
  Python and TypeScript — activation and absent-binary behavior in `vibe-check
  analyze`, per-language capability gating, hybrid metric sourcing (ripwire
  coupling data + vibe-check's own light parse), approximation `Warning` emissions,
  version-pinned fail-closed provenance, and deterministic mapping to the universal
  `ModuleGraph`.

### Modified Capabilities

(None — `openspec/specs/` has no main specs to modify. The `analyze` command's new
language-detection and adapter-selection behaviors are specified as ADDED
requirements inside `ripwire-adapter`, following the repo convention used by
`add-diff-init-provenance`.)

## Impact

- **New code**: `internal/ripwireadapter/` (adapter, subprocess invocation,
  `--arch`/`--deps`/cc.json parsing, light Python/TypeScript parse, mapping);
  `internal/ripwireadapter/testdata/` golden fixtures per language.
- **Modified code**: `cmd/vibe-check/analyze.go` (language detection, registry
  wiring, absent-binary error path); `metrics/graph.go` (optional provenance-input
  fields only — no schema, threshold, or report-engine changes).
- **APIs**: `ProvenanceInput` gains optional `sourceTool`, `sourceToolVersion`,
  `sourceToolCommit` fields (`omitempty`); `metrics.Validate` continues to pass.
  Registry gains `python` and `typescript` entries in CLI wiring.
- **Dependencies**: none (Go stdlib only; ripwire is optional and never vendored).
- **CI**: no workflow changes; `go build ./...`, `go test -race -count=1 ./...`,
  `go vet ./...` must pass without ripwire installed.
- **Distribution**: releases remain zero-dependency; Python/TypeScript analysis is
  opt-in by installing ripwire separately.
- **Never** (hard constraints from issue #69): read ripwire's internal cache;
  consume ranked agent-facing answers as metrics input; vendor the ripwire binary.
- **Follow-up obligations**: documentation issue #70 must be resolved before PR
  merge; upstream ask #68 (non-blocking — ship on the current surface with
  `Warning`s). Depends on #67 (GO, met 2026-10-30); part of epic #66.

## Constitution Alignment

| Principle | Assessment |
|-----------|------------|
| I. Autonomous Collaboration | PASS — self-describing ModuleGraph output; absent-binary error is actionable (install guidance) |
| II. Composability First | PASS — implements the existing `metrics.Adapter` interface; no core-engine changes for the new languages |
| III. Observable Quality | PASS — provenance carries the ripwire version/commit; every approximation emits a `Warning` |
| IV. Testability | PASS — golden fixtures per language; fixture replay keeps tests hermetic without ripwire |
| V. Security by Default | PASS — sanitized subprocess environment, bounded output capture, path validation |
| VI. Metric Fidelity | PARTIAL by design — spike-validated GOOD-ENOUGH coupling data with measured error budgets surfaced as `Warning`s; LCOM4 and Python Abstractness owned by vibe-check's own parse |
| VII. Language Agnosticism | PASS — universal model untouched; adapter maps Python/TS onto it; Go stays native |
