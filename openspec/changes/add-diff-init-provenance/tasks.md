## 1. Shared Provenance Helper

- [x] 1.1 Create `cmd/vibe-check/provenance.go` with a `producerName = "vibe-check"` constant, a `provenanceEnvelope` struct (`Producer`, `Version`, `GeneratedAt` with `json:"producer"`, `json:"version"`, `json:"generatedAt"` tags), and a `newProvenanceEnvelope()` constructor returning a populated envelope using `semanticVersion()` and `time.Now().UTC().Format(time.RFC3339)`. Add GoDoc comments to every exported symbol.

## 2. CLI — diff Provenance and Flag

- [x] 2.1 Add a `Provenance *provenanceEnvelope` field to the `diffJSON` struct in `cmd/vibe-check/diff.go`, tagged `json:"provenance,omitempty"`.
- [x] 2.2 Add a `NoProvenance bool` field to `DiffOptions`; extend `writeDiffJSON` to accept the provenance flag (or the assembled envelope) so that, when false (default), it populates `payload.Provenance` via `newProvenanceEnvelope()`, and when true, leaves it nil.
- [x] 2.3 Add the `--no-provenance` boolean flag to `diffCmd()` and wire it into `DiffOptions.NoProvenance`.

## 3. CLI — init Provenance and Flag

- [x] 3.1 Add a `Provenance *provenanceEnvelope` field to the `initJSON` struct in `cmd/vibe-check/init.go`, tagged `json:"provenance,omitempty"`.
- [x] 3.2 Add a `NoProvenance bool` field to `InitOptions`; extend `writeInitJSON` to accept the provenance flag (or the assembled envelope) so that, when false (default), it populates `payload.Provenance` via `newProvenanceEnvelope()`, and when true, leaves it nil.
- [x] 3.3 Add the `--no-provenance` boolean flag to `initCmd()` and wire it into `InitOptions.NoProvenance`.

## 4. CLI — analyze Refactor to Shared Helper

- [x] 4.1 Update `RunAnalyze` in `cmd/vibe-check/analyze.go` to source `Producer`, `Version`, and `GeneratedAt` from the shared `producerName` constant and `newProvenanceEnvelope()` (replacing the inlined literal and inline `semanticVersion()`/`time.Now()` calls). Preserve the existing `input` population (adapter-driven) and the `--no-provenance` → `graph.Provenance = nil` behavior unchanged.

## 5. Tests

- [x] 5.1 Write unit tests for `newProvenanceEnvelope()`: `Producer` equals `"vibe-check"`, `Version` is non-empty, `GeneratedAt` parses as RFC3339 UTC.
- [x] 5.2 Write CLI tests for `diff` JSON output: default output contains a top-level `provenance` object with correct `producer`; `--no-provenance` output has no `provenance` key; existing `verdict`/`reasons`/`modules` keys are preserved.
- [x] 5.3 Write CLI tests for `init` JSON output: default output contains a top-level `provenance` object with correct `producer`; `--no-provenance` output has no `provenance` key; existing `written`/`skipped`/`forced` keys are preserved.
- [x] 5.4 Write determinism tests: `writeDiffJSON` run twice with the same inputs and `NoProvenance` set is byte-identical; `writeInitJSON` run twice with the same inputs and `NoProvenance` set is byte-identical.
- [x] 5.5 Write a cross-flag equivalence test: `writeDiffJSON` with `NoProvenance` false vs true produces identical `verdict`, `reasons`, `entropyDirection`, and `modules` values (only the `provenance` key differs).

## 6. Validation and CI

- [x] 6.1 Verify `go build ./...` passes.
- [x] 6.2 Verify `go test -race -count=1 ./...` passes.
- [x] 6.3 Verify `go vet ./...` passes.
- [x] 6.4 Run `golangci-lint run ./...` if configured, fix any findings.
- [x] 6.5 Verify all exported symbols have GoDoc comments.

## 7. Documentation

- [x] 7.1 Update `AGENTS.md`: the CLI command docs for `diff` and `init` (mention the new `provenance` object and `--no-provenance` flags), and the Project Structure annotations for `cmd/vibe-check/` (mention `provenance.go`).
- [x] 7.2 Add a `CHANGELOG.md` entry for the add-diff-init-provenance change.
- [x] 7.3 File a documentation issue against this repository for the user-facing `--no-provenance` flags and the new `provenance` output fields (AGENTS.md Documentation gate).
- [x] 7.4 File a website documentation-sync issue in `unbound-force/website` for the new `--no-provenance` flags and the `provenance` output change (Constitution Website Documentation Sync).

<!-- spec-review: passed -->
<!-- code-review: passed -->
