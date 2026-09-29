## Why

The `emit-provenance-metadata` change (issue #24) added a `provenance` object to
`vibe-check analyze` output, but the machine-readable payloads emitted by
`vibe-check diff --json` and `vibe-check init --json` still carry only result
data — deltas/verdict/reasons for `diff`, and written/skipped/forced for
`init`. Downstream consumers (CI, the divisor-entropy agent, dashboards) cannot
attribute or audit these two outputs. This was explicitly deferred from the
initial divisor-entropy work to keep that change scoped; the JSON schemas were
designed to allow additive metadata. This change closes the gap (issue #25).

## What Changes

- Add an additive top-level `provenance` object to the `diff --json` payload
  with `producer`, `version`, and `generatedAt` fields. Existing keys are
  unchanged.
- Add the same `provenance` object to the `init --json` payload.
- Extract a shared CLI helper (`producerName` constant, `provenanceEnvelope`
  struct, `newProvenanceEnvelope()` constructor) so `analyze`, `diff`, and
  `init` all emit a consistent envelope,
  with `version` sourced from the same ldflags/build-info path as
  `vibe-check --version`.
- Add a `--no-provenance` flag to both `diff` and `init`, mirroring `analyze`'s
  existing flag, so each payload remains byte-reproducible when the
  non-deterministic `generatedAt` field must be excluded.

## Capabilities

### New Capabilities

- `diff-init-provenance`: a common provenance envelope (`producer`, `version`,
  `generatedAt`) on the `diff --json` and `init --json` payloads, backed by a
  shared CLI helper and `--no-provenance` flags so all three commands
  (`analyze`, `diff`, `init`) emit consistent, reproducible provenance.

### Modified Capabilities

<!-- None: there are no live specs under openspec/specs/ to modify; the
     analyze-side provenance lives in the emit-provenance-metadata change and
     is unchanged by this change. -->

### Removed Capabilities

<!-- None. -->

## Impact

- **Code**: new `cmd/vibe-check/provenance.go` (shared `producerName` constant,
  `provenanceEnvelope` struct, `newProvenanceEnvelope()` constructor);
  `cmd/vibe-check/diff.go` (`provenance` field on `diffJSON`, `--no-provenance`
  flag); `cmd/vibe-check/init.go` (`provenance` field on `initJSON`,
  `--no-provenance` flag); `cmd/vibe-check/analyze.go` (refactor producer/
  version/generatedAt assembly to use the shared helper; `input` population
  unchanged).
- **APIs**: `diff` and `init` `--json` payloads gain an additive top-level
  `provenance` key; existing keys are unchanged, so existing consumers are
  unaffected. No changes to `metrics` package types or the `ModuleGraph`
  schema.
- **Determinism**: `generatedAt` is the only non-deterministic field;
  `--no-provenance` restores byte-reproducible output for both commands,
  matching `analyze`. `diff` metric payloads otherwise remain deterministic.

## Constitution Alignment

| Principle | Assessment |
|-----------|------------|
| I. Autonomous Collaboration | PARTIAL — diff/init payloads become self-describing (producer identity, version, timestamp) for downstream consumers; the divisor-entropy agent can now attribute its input. |
| II. Composability First | N/A — no new dependencies or mandatory coupling between agents. |
| III. Observable Quality | PASS — `diff --json` and `init --json` now carry producer, version, and generation timestamp, completing the provenance requirement across all JSON-emitting commands. |
| IV. Testability | PARTIAL — coverage strategy (unit/integration split) is defined in design.md; coverage-ratchet enforcement (Constitution IV) remains a project-wide follow-up. |
| V. Security by Default | N/A — no new dependencies, secrets, or permission changes. |
| VI. Metric Fidelity | PASS — provenance is metadata, not a metric; metric-payload determinism is preserved via `--no-provenance`. |
| VII. Language Agnosticism | N/A — CLI-level change only; no core-engine or adapter interface changes. |
