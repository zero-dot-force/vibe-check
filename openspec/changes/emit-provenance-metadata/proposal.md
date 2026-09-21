## Why

Constitution III (Observable Quality) requires every artifact to carry provenance —
producer, version, timestamp, and input identity. The `ModuleGraph` JSON emitted by
`vibe-check analyze` currently carries none of these, so a CI system consuming a
metrics report cannot verify which tool version produced it or when. This was
explicitly deferred as a Non-Goal in `openspec/changes/go-analyze/design.md` and
recorded as **PARTIAL** in that proposal's Constitution Alignment table; this change
closes the gap (issue #24).

## What Changes

- Add an optional `provenance` object to `ModuleGraph` (`metrics/graph.go`) via new
  `Provenance` and `ProvenanceInput` types, tagged `json:"provenance,omitempty"` so
  the field is absent when nil — fully backward compatible with existing consumers.
- Populate `provenance` on `analyze` output with four fields: `producer`
  (constant `"vibe-check"`), `version` (build info / ldflags), `generatedAt`
  (RFC3339 UTC), and `input` (analyzed path + resolved module path).
- Add a `--no-provenance` flag to `vibe-check analyze` that omits the `provenance`
  object entirely, preserving byte-reproducible output for the deterministic metric
  payload.
- Additive schema bump `1.1` → `1.2`; update `metrics/modulegraph.schema.json` and
  `metrics.Validate` to accept the optional `provenance` object. Existing `1.0`/`1.1`
  consumers remain valid.

## Capabilities

### New Capabilities

- `provenance-metadata`: optional provenance object on `ModuleGraph` — producer,
  version, generation timestamp, and input identity (analyzed path + resolved module
  path) — with a reproducible-output escape hatch and additive schema/validator
  support. Covers the `metrics.Provenance`/`ProvenanceInput` types, the `1.2` schema
  bump, the validator update, and the `--no-provenance` CLI flag.

### Modified Capabilities

<!-- None: there are no live specs under openspec/specs/ to modify; schema and model
     changes are folded into the provenance-metadata capability above. -->

### Removed Capabilities

<!-- None. -->

## Impact

- **Code**: `metrics/graph.go` (new `Provenance`, `ProvenanceInput` types + field,
  `SchemaVersionCurrent` → `"1.2"`), `metrics/modulegraph.schema.json`,
  `metrics/validate.go`, `cmd/vibe-check/analyze.go` (+`--no-provenance` flag and
  provenance assembly), `internal/goadapter/` (expose resolved module path + analyzed
  path for input identity).
- **APIs**: `ModuleGraph` gains an optional `Provenance *Provenance` field. Existing
  consumers are unaffected (field is omitted when nil). `metrics.Adapter` interface is
  unchanged — new languages need no core-engine changes.
- **Schema**: additive `1.2` bump; `1.0`/`1.1` inputs still validate.
- **Diff**: `ComputeDelta` ignores provenance (it reads only modules/cycles/warnings/
  status), so `vibe-check diff` is unaffected and remains deterministic.
- **Determinism**: provenance is treated as metadata, not a metric; `generatedAt` is
  the only non-deterministic field and is removable via `--no-provenance`.

## Constitution Alignment

| Principle | Assessment |
|-----------|------------|
| I. Autonomous Collaboration | N/A — no new agent artifacts or inter-agent exchange formats introduced. |
| II. Composability First | N/A — no new dependencies or mandatory coupling between agents. |
| III. Observable Quality | PASS — `ModuleGraph` output now carries producer, version, timestamp, and input identity. (Previously PARTIAL.) |
| IV. Testability | PARTIAL — coverage strategy (unit/integration/e2e split, >= 80% line coverage target) is now defined in design.md; coverage-ratchet enforcement (Constitution IV) remains deferred as a follow-up. |
| V. Security by Default | N/A — no new dependencies, secrets, or permission changes. |
| VI. Metric Fidelity | PASS — provenance is metadata, not a metric; metric-payload determinism is preserved via `--no-provenance`. |
| VII. Language Agnosticism | PASS — `metrics.Adapter` interface unchanged; input identity is populated by each adapter with no core-engine change. |
