## 1. Metrics Model — Provenance Types

- [x] 1.1 Add `Provenance` and `ProvenanceInput` types to `metrics/graph.go` with GoDoc comments: `Provenance` (Producer, Version, GeneratedAt, Input) and `ProvenanceInput` (Path, ModulePath). Add `Provenance *Provenance` field to `ModuleGraph` with `json:"provenance,omitempty"` tag.
- [x] 1.2 Bump `SchemaVersionCurrent` from `"1.1"` to `"1.2"` in `metrics/graph.go` and update the `SchemaVersion` field's GoDoc example.

## 2. Schema and Validator

- [x] 2.1 Update `metrics/modulegraph.schema.json`: add an optional `provenance` object property (`producer`, `version`, `generatedAt`, `input{path,modulePath}`), extend the `schemaVersion` enum to `["1.0", "1.1", "1.2"]`, keep `additionalProperties: false` at top level.
- [x] 2.2 Update `metrics/validate.go`: accept `schemaVersion` `"1.2"` in `validateTopLevel`; add a `validateProvenance` check (when present: `producer`/`version` strings, `input` object with string `path`/`modulePath`); wire it into `Validate`.
- [x] 2.3 Write validator tests: `1.2` accepted, `1.0`/`1.1` accepted, well-formed `provenance` accepted, malformed `provenance` (string/array/missing `input`) rejected, `provenance` absent accepted (omitempty).

## 3. Go Adapter — Input Identity

- [x] 3.1 Extend `resolvePackages` in `internal/goadapter/resolve.go` to also return the resolved module path (from `packages.Package.Module` / module resolution already used for scope filtering).
- [x] 3.2 Update `Analyze` in `internal/goadapter/adapter.go` to populate `graph.Provenance` with `Input{Path, ModulePath}` (normalized project path + resolved module path). Leave `Producer`, `Version`, `GeneratedAt` empty — the CLI owns them.
- [x] 3.3 Write adapter tests: `Input.Path` equals the analyzed path and `Input.ModulePath` equals the resolved module path for a `testdata/` fixture.

## 4. CLI — Provenance Assembly and Flag

- [x] 4.1 Extract a `semanticVersion()` helper from `versionString()` in `cmd/vibe-check/root.go` (returns version only, same ldflags→build-info fallback); keep `--version` output unchanged.
- [x] 4.2 Update `RunAnalyze` in `cmd/vibe-check/analyze.go`: after `Analyze` returns, populate `graph.Provenance.Producer = "vibe-check"`, `Version = semanticVersion()`, `GeneratedAt = time.Now().UTC().Format(time.RFC3339)` (preserving the adapter's `Input`). Add a `NoProvenance bool` field to `AnalyzeOptions`; when true, set `graph.Provenance = nil` before marshaling.
- [x] 4.3 Add the `--no-provenance` boolean flag to `analyzeCmd()` and wire it into `AnalyzeOptions.NoProvenance` (respecting `cmd.Flags().Changed` semantics is unnecessary — plain default-false bool suffices).
- [x] 4.4 Write CLI tests: default output contains `provenance` with correct `producer`/`input`; `--no-provenance` output has no `provenance` key; `generatedAt` parses as RFC3339 UTC; `semanticVersion()` matches `--version`'s version component.

## 5. Determinism and Diff

- [x] 5.1 Write determinism test: run `analyze` logic twice with `NoProvenance` set and assert byte-identical JSON output.
- [x] 5.2 Write diff-ignores-provenance test: `ComputeDelta` on two graphs differing only in `Provenance` yields the same delta/verdict as without provenance.

## 6. Validation and CI

- [x] 6.1 Verify `go build ./...` passes.
- [x] 6.2 Verify `go test -race -count=1 ./...` passes.
- [x] 6.3 Verify `go vet ./...` passes.
- [x] 6.4 Run `golangci-lint run ./...` if configured, fix any findings.
- [x] 6.5 Verify all exported symbols have GoDoc comments.

## 7. Documentation

- [x] 7.1 Update `AGENTS.md`: the Architecture/Layer 1 description (mention the `Provenance`/`ProvenanceInput` types and schema `1.2`), the Project Structure annotations (`modulegraph.schema.json` `(v1.1)` → `(v1.2)`; `validate.go` `(accepts v1.0 and v1.1)` → `(accepts v1.0, v1.1, and v1.2)`), and the analyze command docs (mention `--no-provenance`).
- [x] 7.2 Update stale schema-version references in code GoDoc and CLI help: `cmd/vibe-check/main.go` (`"version 1.1"`), `cmd/vibe-check/analyze.go` Long help (`"version 1.1"` → `"1.2"`, and mention `--no-provenance`), `internal/goadapter/doc.go` (`"schema version \"1.1\""`), `internal/goadapter/adapter.go` (`"schema version \"1.1\""`).
- [x] 7.3 Update `README.md` schema examples (`schemaVersion` `"1.1"` → `"1.2"`) and mention the `provenance` object and the `--no-provenance` flag.
- [x] 7.4 Add a `CHANGELOG.md` entry for the emit-provenance-metadata change.
- [x] 7.5 File a documentation issue against this repository for the user-facing `--no-provenance` flag and the new `provenance` output field (AGENTS.md Documentation gate).
- [x] 7.6 File a website documentation-sync issue in `unbound-force/website` for the new `--no-provenance` flag and the `provenance` output schema change (Constitution Website Documentation Sync).

<!-- spec-review: passed -->
<!-- code-review: passed -->
