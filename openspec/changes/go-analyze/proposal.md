## Why

Vibe-Check has a universal coupling metrics model (`metrics/` package) but no way to
actually analyze Go code. Without a Go-native adapter and CLI entry point, the toolkit
cannot compute any metrics. This is the first language adapter (P1 per the RFC phasing)
and the highest-priority remaining work — it makes the toolkit usable and unblocks all
downstream features (CI gating, multi-language support, architectural drift tracking).

Note: While AGENTS.md classifies language adapters as P1, the Go adapter is effectively
the minimum viable product. The P0 universal model (`metrics/`) is complete but produces
no value without at least one adapter. AGENTS.md will be updated to reflect this.

## What Changes

- Add a Go-native adapter (`internal/goadapter/`) that implements `metrics.Adapter` using
  `golang.org/x/tools/go/packages` for type-aware dependency resolution.
- Add a `vibe-check analyze [path]` CLI command (`cmd/vibe-check/`) using cobra
  that invokes the Go adapter, produces JSON output conforming to `metrics.ModuleGraph`,
  and supports CI gate flags for threshold enforcement.
- Add new external dependencies: `golang.org/x/tools/go/packages`, `github.com/spf13/cobra`.

## Capabilities

### New Capabilities

- `go-adapter`: Go-native language adapter implementing `metrics.Adapter` — resolves
  package dependencies via `golang.org/x/tools/go/packages`, counts exported/abstract
  types, computes LCOM4 cohesion, and detects circular dependencies.
- `analyze-command`: CLI command `vibe-check analyze` — orchestrates adapter invocation,
  formats JSON output, validates against schema, and enforces CI gate thresholds
  (`--max-instability`, `--max-distance`, `--no-circular-deps`, `--max-lcom`).
- `extensions-mechanism`: Language-specific extensions field on `ModuleResult` —
  enables adapters to carry metrics beyond the universal model (e.g., Go interface
  width, interface proximity) without modifying the core schema. Keys are namespaced
  by language (e.g., `go.interfaceWidth`).

### Modified Capabilities

- `metrics-schema`: Add optional `extensions` object field to module items in
  `modulegraph.schema.json` while keeping `additionalProperties: false` (extensions
  is an explicitly allowed property). Bump `SchemaVersion` from `"1.0"` to `"1.1"`.
  Top-level `ModuleGraph` and `Warning` schemas remain strict.
- `metrics-model`: Add `Extensions map[string]any` field to `ModuleResult` with
  `json:"extensions,omitempty"` tag. Update `metrics.Validate` to accept the
  extensions field.

## Impact

- **New packages**: `internal/goadapter/` (adapter), `cmd/vibe-check/` (CLI entry point)
- **Dependencies**: `golang.org/x/tools` (Go analysis), `github.com/spf13/cobra` (CLI)
- **Build artifacts**: `vibe-check` binary
- **APIs**: `ModuleResult` gains an `Extensions` field (`map[string]any`, omitempty).
  JSON schema updated to allow `extensions` on module items. Existing consumers are
  unaffected (field is optional and omitted when empty).
- **CI**: New binary build target; threshold flags enable CI gating in downstream pipelines
- **Distribution**: `go install github.com/zero-dot-force/vibe-check/cmd/vibe-check@v0.1.0`
  for initial distribution. An initial `v0.1.0` release tag is required for `go install`
  to resolve.
- **Follow-up obligations** (issues to file or refresh before PR merge):
  - Documentation: refresh the scope of the existing docs issue (#18) to cover the new
    `vibe-check analyze` CLI (usage, flags, exit codes) — it predates this command.
  - Content: existing blog (#19) and tutorial (#22) issues track ecosystem write-ups.
  - GoReleaser / release-automation: not yet tracked — file an issue before PR merge.
  - Website documentation sync: not yet tracked — file an issue before PR merge.
  - Provenance metadata (Constitution III gap): tracked as a follow-up to emit
    `producer`, `version`, `timestamp`, and input fields in the `ModuleGraph`.

## Constitution Alignment

| Principle | Assessment |
|-----------|------------|
| I. Autonomous Collaboration | PASS — Adapter produces self-describing JSON output |
| II. Composability First | PASS — Implements existing `metrics.Adapter` interface |
| III. Observable Quality | PARTIAL — JSON output with schema validation and `--version` embedding are in place, but provenance metadata (producer, version, timestamp, input) is NOT yet emitted in the `ModuleGraph`. It is deferred to a tracked follow-up issue. |
| IV. Testability | PASS — Coverage strategy defined in design.md |
| V. Security by Default | PASS — Path validation via `metrics.ValidateProjectPath`; context cancellation support |
| VI. Metric Fidelity | PASS — Uses canonical `metrics.Compute*` functions; LCOM4 (Hitz & Montazeri) variant cited. LCOM4 generic pointer-receiver handling was corrected (e.g., `func (t *T[P]) M()`). |
| VII. Language Agnosticism | PASS — Adapter pattern; extensions mechanism is language-agnostic (any adapter can use it) |
