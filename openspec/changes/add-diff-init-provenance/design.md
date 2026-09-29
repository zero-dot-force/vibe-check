## Context

The `emit-provenance-metadata` change (issue #24) added a `provenance` object to
`vibe-check analyze` output via `metrics.Provenance`/`ProvenanceInput` and a
`--no-provenance` flag. Two other JSON-emitting commands remain without
provenance: `vibe-check diff` and `vibe-check init`.

- `diff.go` marshals a `diffJSON` struct (verdict, reasons, entropyDirection,
  unreliable, modules, added, removed, newCycles, resolvedCycles) via
  `writeDiffJSON`, using `json.MarshalIndent`. Its metric payload is
  deliberately deterministic.
- `init.go` marshals an `initJSON` struct (written, skipped, forced) via
  `writeInitJSON`.

Neither payload has an associated JSON Schema (only `ModuleGraph` does), so this
change needs no schema bump or validator change. Version/commit/date are already
embedded at build time (`package main` `version`/`commit`/`date`) and resolved
via `versionString()` and `semanticVersion()` in `root.go`, the latter already
returning the bare semantic version with a `debug.ReadBuildInfo()` fallback.

The `analyze.go` provenance assembly currently inlines the producer literal
`"vibe-check"`, `semanticVersion()`, and `time.Now().UTC().Format(time.RFC3339)`
— this is the pattern to generalize (issue #25, "consistent envelope" note).

## Goals / Non-Goals

**Goals:**

- Add an additive top-level `provenance` object to the `diff --json` payload
  with `producer`, `version`, and `generatedAt` (RFC3339 UTC).
- Add the same `provenance` object to the `init --json` payload.
- Extract a shared CLI helper so `analyze`, `diff`, and `init` emit a consistent
  envelope (same producer constant, version source, and timestamp format).
- Add `--no-provenance` flags to `diff` and `init`, mirroring `analyze`, so each
  payload remains byte-reproducible.
- Treat provenance as metadata, never as a metric; keep the `diff` metric
  payload deterministic.

**Non-Goals:**

- A top-level envelope wrapping the diff/init payloads (breaking change —
  rejected).
- Adding `commit`/`date` as provenance fields (issue scope is producer,
  version, timestamp).
- Changing `metrics.Provenance`/`ProvenanceInput` or the `ModuleGraph` schema.
- Signing or SLSA attestation of the artifacts (future work).
- Adding `input` identity to diff/init provenance (these commands have no single
  analyzed module path; analyze's `input` population is unchanged).

## Decisions

### D1: Additive `provenance` object with `generatedAt` (not `timestamp`)

**Decision**: Add a top-level `provenance` object to `diffJSON` and `initJSON`
with fields `producer`, `version`, and `generatedAt`. Use the field name
`generatedAt` — not the issue's illustrative `timestamp` — to match analyze's
existing `metrics.Provenance.GeneratedAt` and honor the issue's "consistent
envelope" note.

**Rationale**: A consistent envelope across `analyze`, `diff`, and `init` lets
downstream consumers parse provenance identically. Reusing `generatedAt` avoids
a naming fork where three commands disagree on the timestamp key.

**Alternatives considered**:
- Use `timestamp` per the issue title — creates an inconsistency with analyze's
  already-shipped `generatedAt`; rejected.
- Flat top-level scalar fields (`producer`, `version`, `generatedAt` directly on
  the payload) — clutters the payload and diverges from analyze's nested-object
  shape; rejected.

### D2: Shared CLI helper for a consistent envelope

**Decision**: Add `cmd/vibe-check/provenance.go` with:
- a `producerName = "vibe-check"` constant (extracting the literal currently
  inlined in `analyze.go`);
- a `provenanceEnvelope` struct (`Producer`, `Version`, `GeneratedAt` with
  `json:"producer"`, `json:"version"`, `json:"generatedAt"`);
- a `newProvenanceEnvelope()` constructor returning a populated envelope using
  `semanticVersion()` and `time.Now().UTC().Format(time.RFC3339)`.

`diff` and `init` embed a `provenanceEnvelope` field; `analyze` refactors its
producer/version/generatedAt assembly to consume the same constant and
constructor (its `input` population stays adapter-driven and unchanged).

**Rationale**: A single source of truth for producer identity, version
resolution, and timestamp format prevents drift across the three commands and
satisfies the issue's "shared helper" note.

**Alternatives considered**:
- Reuse `metrics.Provenance` directly for diff/init — it carries a required
  `input` sub-object (`json:"input"`, no `omitempty`) that is meaningless for
  diff/init and would emit an empty `input` object; rejected.
- Duplicate the three-line assembly in each command — triplicates the
  producer/version/timestamp logic and invites drift; rejected.

### D3: Reuse `semanticVersion()` for `version`

**Decision**: `provenanceEnvelope.Version` is populated from the existing
`semanticVersion()` helper (bare version, no commit/date, ldflags →
`debug.ReadBuildInfo()` fallback).

**Rationale**: Keeps the reported provenance version consistent with
`vibe-check --version` and avoids reimplementing the fallback.

**Alternatives considered**:
- Full `versionString()` (with commit/date) — pollutes the `version` field with
  display formatting; rejected.
- Re-resolve build info independently — duplicates existing logic; rejected.

### D4: `--no-provenance` flags on `diff` and `init`

**Decision**: Add a `--no-provenance` boolean flag to both `diff` and `init`,
defaulting to provenance ON. When set, the `provenance` field is set to nil /
omitted from the marshaled payload.

**Rationale**: A wall-clock timestamp is the only non-deterministic field.
`diff` output is currently byte-reproducible by design, and Constitution VI
treats non-deterministic output as a P0 bug. An opt-out flag (mirroring
`analyze`) restores byte-reproducibility without sacrificing the default
attribution.

**Alternatives considered**:
- Provenance opt-in (`--provenance`) — the issue wants attribution by default
  for downstream consumers; rejected.
- Omit only `generatedAt` under a reproducibility flag — leaves partial
  provenance that is harder to reason about; rejected.

### D5: No schema or validator change

**Decision**: No changes to `metrics/modulegraph.schema.json`, `metrics.Validate`,
`SchemaVersionCurrent`, or `metrics.Provenance`.

**Rationale**: `diffJSON` and `initJSON` are ad-hoc payloads without an
associated JSON Schema; only `ModuleGraph` carries a schema. Provenance on
diff/init is a CLI-layer concern with no validator to update.

**Alternatives considered**:
- Introduce schemas for diff/init payloads — out of scope and would be net-new
  governance surface; rejected.

## Risks / Trade-offs

- **[R1] Non-deterministic timestamp breaks diff byte-reproducibility** →
  Mitigation: `--no-provenance` flag on `diff`; determinism tests assert
  byte-identical output under `--no-provenance`; provenance documented as
  metadata, not a metric.
- **[R2] Divergence between the shared envelope and analyze's `input`-carrying
  `metrics.Provenance`** → Mitigation: the shared helper is the single source of
  producer/version/timestamp; analyze's `input` remains a separate,
  adapter-owned concern documented in D2.
- **[R3] Downstream consumers that parse diff/init JSON strictly and reject
  unknown top-level keys** → Mitigation: the new `provenance` key is additive
  and only emitted when provenance is enabled; consumers can opt out via
  `--no-provenance` or ignore the key.
- **[R4] `init` may run before a Go module context exists, so build-info
  resolution could behave differently** → Mitigation: `semanticVersion()` already
  falls back to `"dev"` (or `debug.ReadBuildInfo()`) in exactly this scenario;
  tests assert a non-empty `version` string rather than a specific value.

## Coverage Strategy

**Unit tests** (`cmd/vibe-check/`): cover `newProvenanceEnvelope()` field
population (producer constant, version from `semanticVersion()`, `generatedAt`
RFC3339 UTC parseability); cover `diffJSON`/`initJSON` marshaling with and
without `--no-provenance`; cover flag wiring in `diffCmd()`/`initCmd()`. Use the
testable CLI pattern — inject `bytes.Buffer` for stdout/stderr, no subprocess
execution.

**Determinism verification**: run `writeDiffJSON` twice with the same inputs and
`NoProvenance` set, asserting byte-identical output; run `writeInitJSON` twice
with the same inputs and `NoProvenance` set, asserting byte-identical output.

**Coverage target**: >= 80% line coverage for the new/modified
`cmd/vibe-check/` files. CI generates a coverage profile
(`go test -race -count=1 -coverprofile=coverage.out ./...`) but does not yet
enforce a coverage ratchet; ratchet/threshold enforcement remains a follow-up.
