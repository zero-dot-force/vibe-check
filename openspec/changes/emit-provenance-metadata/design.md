## Context

Vibe-Check's `metrics/` package (Layer 1) provides the universal `ModuleGraph` model
and a hand-rolled JSON validator (`metrics.Validate`) plus an embedded JSON Schema
(`metrics/modulegraph.schema.json`, currently `1.1`). The Go adapter (Layer 2,
`internal/goadapter`) computes metrics and returns a `ModuleGraph`; the CLI (Layer 3,
`cmd/vibe-check`) marshals that graph to JSON via `RunAnalyze` in `analyze.go`.
Version/commit/date are already embedded at build time in `package main`
(`version`, `commit`, `date` vars) and resolved via `versionString()` in `root.go`
with a `debug.ReadBuildInfo()` fallback.

The `Adapter` interface is `Analyze(ctx, projectPath) (*ModuleGraph, error)` and is
implemented by direct instantiation in the CLI (go-analyze design D10). The adapter
already resolves the target module's import path internally in `resolvePackages` for
scope filtering (go-analyze design D6) but does not expose it.

Provenance (producer, version, timestamp, input) is the Constitution III gap left as
a Non-Goal in go-analyze and tracked as issue #24.

## Goals / Non-Goals

**Goals:**

- Add an optional, backward-compatible `provenance` object to `ModuleGraph` with
  `producer`, `version`, `generatedAt` (RFC3339 UTC), and `input` (`path`,
  `modulePath`).
- Populate provenance on `vibe-check analyze` output by default.
- Provide `--no-provenance` for byte-reproducible output so the determinism guarantee
  for the metric payload holds.
- Bump the schema additively (`1.1` → `1.2`) and update both the embedded JSON Schema
  and the hand-rolled validator, keeping `1.0`/`1.1` inputs valid.
- Treat provenance as metadata, never as a metric; keep `ComputeDelta`/`diff`
  unaffected.

**Non-Goals:**

- A top-level envelope wrapping the graph (breaking change — rejected).
- Adding `commit`/`date` as provenance fields (issue scope is producer, version,
  timestamp, input).
- Signing or SLSA attestation of the artifact (future work).
- Changing the `metrics.Adapter` interface or requiring core-engine changes for new
  languages.
- Schema enforcement on `input`/`provenance` contents beyond the declared field
  shapes.

## Decisions

### D1: Optional top-level `provenance` field (not an envelope)

**Decision**: Add `Provenance *Provenance` to `ModuleGraph` with
`json:"provenance,omitempty"`, where `Provenance` holds `Producer`, `Version`,
`GeneratedAt`, and `Input ProvenanceInput` (`Path`, `ModulePath`).

**Rationale**: A nil pointer is omitted entirely, so existing `1.0`/`1.1` consumers
parsing the output are unaffected. This mirrors the existing `Extensions` mechanism
in spirit (additive, opt-in).

**Alternatives considered**:
- Top-level envelope `{ "graph": {...}, "provenance": {...} }` — breaks every
  consumer; rejected as a MAJOR breaking change.
- Flat top-level scalar fields (`producer`, `version`, `generatedAt`) — clutters the
  schema with many optional fields; a single nested object is tidier and groups the
  concern.

### D2: Provenance split across adapter and CLI

**Decision**: The Go adapter populates the `Input` sub-object (the identity it
uniquely knows — the analyzed path and the resolved module path), and the CLI
populates `Producer`, `Version`, and `GeneratedAt` (the build identity it uniquely
owns). `resolvePackages` is extended to also return the resolved module path, which
`Analyze` places on `graph.Provenance.Input` alongside the (normalized) project path.

**Rationale**: Input identity is resolved by Layer 2 (it already loads the module and
knows the module path); build identity (version from ldflags/`--version` logic,
timestamp) is a Layer 3 concern. Keeping the `metrics.Adapter` interface unchanged
(`Analyze` still returns `(*ModuleGraph, error)`) preserves the language-agnostic
contract: a future Python adapter would populate `Input` the same way without any
core-engine change.

**Alternatives considered**:
- CLI resolves module path separately (e.g., parsing `go.mod`) — duplicates work the
  adapter already does and risks drift; rejected.
- Adapter populates everything — it cannot see the ldflags-injected `main.version`
  (would regress to build-info-only version); rejected.

### D3: Reuse existing version resolution for `version`

**Decision**: Extract the semantic-version portion of the existing `versionString()`
logic into a helper (e.g., `semanticVersion()`) that returns just the version string
(no commit/date), with the same `debug.ReadBuildInfo()` fallback. `Provenance.Version`
uses this value; `--version` output is unchanged.

**Rationale**: Avoids duplicating the ldflags→build-info fallback and keeps the
reported provenance version consistent with `--version`.

**Alternatives considered**:
- Full `versionString()` (with commit/date) in `version` — pollutes the `version`
  field with display formatting; rejected.

### D4: `--no-provenance` flag for reproducibility

**Decision**: Add a `--no-provenance` boolean flag to `analyze`. When set, the CLI
sets `graph.Provenance = nil` before marshaling, omitting the object entirely.

**Rationale**: A wall-clock timestamp is the only non-deterministic field. Removing
the entire provenance object (rather than just the timestamp) is simpler, easier to
reason about, and gives a single documented reproducible mode.

**Alternatives considered**:
- `--reproducible` that only omits `generatedAt` — leaves partial provenance that is
  still non-deterministic-free but more confusing; rejected.
- Timestamp injection via an option — adds API surface without clear value; rejected.

### D5: Schema bump and validation

**Decision**: Bump `SchemaVersionCurrent` to `"1.2"`. Add a `provenance` property to
`modulegraph.schema.json` (object with `producer`, `version`, `generatedAt`,
`input{path,modulePath}`), extend the `schemaVersion` enum to `["1.0","1.1","1.2"]`,
and keep `additionalProperties: false`. Update `metrics.Validate` to accept
`"1.2"`, accept `1.0`/`1.1`, and validate `provenance` when present (producer/version
strings, `input` object with string `path`/`modulePath`).

**Rationale**: Additive minor bump per the existing semver convention in `graph.go`.
Both the embedded JSON Schema and the hand-rolled validator are sources of truth and
must stay in lockstep (the validator is the runtime trust-boundary check used by
`diff` and external analyzers).

**Alternatives considered**:
- MAJOR bump — provenance is additive, so minor is correct; rejected.
- JSON Schema only — the hand-rolled validator is what `diff` actually calls at
  runtime; leaving it stale would allow invalid graphs through; rejected.

## Risks / Trade-offs

- **[R1] Non-deterministic timestamp breaks byte-reproducibility expectations** →
  Mitigation: `--no-provenance` flag; provenance documented as metadata, not a metric;
  determinism tests assert byte-identical output under `--no-provenance`.
- **[R2] Schema drift between embedded JSON Schema and hand-rolled validator** →
  Mitigation: update both in the same task and add validator tests covering the new
  `provenance` shapes.
- **[R3] Adapter/CLI provenance split could desynchronize** → Mitigation: the adapter
  owns only `Input`; the CLI owns only build identity; integration tests assert the
  full assembled object is correct end-to-end.
- **[R4] External analyzers that reject unknown top-level fields** → Mitigation: the
  field is optional (`omitempty`); only newly-minted `1.2` outputs carry it, and those
  consumers opt in by accepting `1.2`.

## Coverage Strategy

**Unit tests**: validator unit tests in `metrics/` cover the new `provenance` shapes
(task 2.3); adapter unit tests in `internal/goadapter/` cover input-identity
population (task 3.3); CLI unit tests in `cmd/vibe-check/` cover the
`--no-provenance` flag and `semanticVersion()` (task 4.4). Coverage target: >= 80%
line coverage for `metrics/`, `internal/goadapter/`, and `cmd/vibe-check/`.

**Integration test**: analyze a dedicated `testdata/` Go module (not self-analysis)
with known structure, asserting `provenance.input.path` equals the analyzed path and
`provenance.input.modulePath` equals the resolved module path. Guard with
`testing.Short()` per TC-011.

**CLI tests**: use the testable CLI pattern — inject `bytes.Buffer` for stdout/stderr.
Test flag parsing (`--no-provenance`), JSON output validity, `generatedAt` RFC3339 UTC
parsing, and the `producer`/`version` values. No subprocess execution in tests. Together
with the integration test (adapter-level `testdata` analysis), these CLI tests serve as
the **end-to-end tier**, exercising the full adapter → CLI → JSON-marshaling path.

**Determinism verification**: run the `analyze` logic twice with `NoProvenance` set and
assert byte-identical JSON output (task 5.1); assert `ComputeDelta` on two graphs
differing only in `Provenance` yields the same delta/verdict as without provenance
(task 5.2).

**Coverage profile**: CI generates a coverage profile (`go test -race -count=1
-coverprofile=coverage.out ./...`) but does not yet enforce a coverage ratchet or a
minimum-threshold gate. Ratchet/threshold enforcement is a follow-up.
