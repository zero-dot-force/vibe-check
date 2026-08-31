## 1. Project Setup

- [x] 1.1 Add dependencies: `golang.org/x/tools` and `github.com/spf13/cobra` to `go.mod`
- [x] 1.2 Create package scaffolding: `internal/goadapter/doc.go`, `cmd/vibe-check/main.go` with package-level GoDoc
- [x] 1.3 Create `testdata/` fixture modules: (a) multi-package module with known import graph for Ca/Ce testing, (b) packages with specific type mixes (interfaces, structs, aliases, empty) for abstractness testing, (c) packages with specific method-field patterns for LCOM testing, (d) packages with exported interfaces of varying widths and consumer/producer proximity for extensions testing, (e) a fixture module with intentional compilation errors (missing import, syntax error in one package) for partial-build testing, (f) an empty directory (no `.go` files) for total-failure testing. Each fixture must have a valid `go.mod` and produce deterministic metric values.

## 2. Go Adapter — Dependency Resolution

- [x] 2.1 Implement `internal/goadapter/adapter.go`: `Adapter` struct with `Language()`, `Capabilities()`, and `Analyze()` method skeleton that validates project path via `metrics.ValidateProjectPath` and propagates `context.Context` to `packages.Config.Context`
- [x] 2.2 Implement `internal/goadapter/resolve.go`: load packages via `packages.Load` with `NeedName|NeedImports|NeedTypes|NeedSyntax|NeedTypesInfo`, filter to module-internal packages, build import adjacency map. Set `packages.Config.Env` to `metrics.SanitizeEnvironment(nil)` to prevent credential leakage.
- [x] 2.3 Compute Ca/Ce from the import adjacency map: Ce = number of distinct imports (module-internal + external), Ca = number of module-internal packages that import this package
- [x] 2.4 Write tests for dependency resolution: use `testdata/` fixture module with known import graph, verify Ca/Ce values, verify stdlib/external exclusion from module list

## 3. Go Adapter — Type Analysis

- [x] 3.1 Implement `internal/goadapter/types.go`: walk AST to count exported type declarations per package, classify interfaces as abstract. Handle edge cases: type aliases (concrete), empty packages (ExportedTypes=0), packages with nil type info (ExportedTypes=0, add warning)
- [x] 3.2 Write tests for type counting: packages with mixed exported/unexported types, interfaces vs structs, type aliases, empty packages (no type declarations)

## 4. Go Adapter — LCOM4 Cohesion

- [x] 4.1 Implement `internal/goadapter/lcom.go`: build method-field access graph for exported methods, compute connected components via union-find, return LCOM4 value. Package-level functions (no receiver) are excluded from the graph.
- [x] 4.2 Write tests for LCOM4: (a) fully cohesive package — struct with 3 methods all accessing same field (LCOM=1), (b) non-cohesive package — struct with 4 methods split into 2 groups accessing disjoint fields (LCOM=2), (c) package with no exported methods (LCOM=0), (d) package with only package-level functions (LCOM=0). Use concrete Go source in `testdata/` fixtures.

## 5. Go Adapter — Cycle Detection

- [x] 5.1 Implement `internal/goadapter/cycles.go`: Tarjan's SCC on the package import graph, convert SCCs with >1 node to `metrics.Cycle` with canonical ordering
- [x] 5.2 Write tests for cycle detection: (a) acyclic graph produces empty cycles slice (not nil), (b) test Tarjan's algorithm directly with constructed adjacency maps since Go import cycles are compile errors, (c) verify canonical ordering of cycle output

## 5a. Metrics Model — Extensions Mechanism

- [x] 5a.1 Add `Extensions map[string]any` field to `ModuleResult` in `metrics/graph.go` with `json:"extensions,omitempty"` tag and GoDoc comment explaining language-specific extensions namespaced by language
- [x] 5a.2 Update `metrics/modulegraph.schema.json`: add `"extensions": { "type": "object" }` to the module item's `properties` list while keeping `additionalProperties: false`. Bump `SchemaVersion` from `"1.0"` to `"1.1"` in `metrics/graph.go`.
- [x] 5a.3 Update `metrics/validate.go`: accept the `extensions` field as a valid JSON object on module items. If `extensions` is present, verify it is a JSON object (not a primitive or array). Accept both `SchemaVersion` `"1.0"` (no extensions) and `"1.1"` (with extensions) for backward compatibility.
- [x] 5a.4 Write tests for extensions: verify `ModuleResult` with extensions marshals/unmarshals correctly (including JSON round-trip type fidelity — `int` becomes `float64`), verify `metrics.Validate` passes with extensions present, verify `metrics.Validate` passes with extensions absent (omitempty), verify `metrics.Validate` rejects non-object extensions (e.g., string, array), verify existing tests still pass

## 5b. Go Adapter — Extensions (Interface Width and Proximity)

- [x] 5b.1 Implement `internal/goadapter/extensions.go`: compute interface width (method count per exported interface, including flattened embedded methods) and interface proximity (consumer vs. producer based on same-package implementation check). Populate `go.interfaceWidth` (map[string]int) and `go.interfaceProximity` (map[string]string) in `ModuleResult.Extensions`. Provide typed accessor functions (`InterfaceWidths(extensions map[string]any) (map[string]int, error)` and `InterfaceProximities(extensions map[string]any) (map[string]string, error)`) for safe extraction after JSON round-trip.
- [x] 5b.2 Write tests for extensions: (a) single-method interface (width=1), (b) multi-method interface (width=2), (c) embedded interface flattening, (d) no exported interfaces (no extensions), (e) producer-side interface (implementation in same package), (f) consumer-side interface (no implementation in same package), (g) mixed proximity, (h) typed accessor round-trip test (marshal → unmarshal → accessor → verify values), (i) accessor returns error on missing/nil extensions. Use `testdata/` fixtures.

## 6. Go Adapter — Assembly and Integration

- [x] 6.1 Complete `Analyze()`: wire resolve → types → LCOM → cycles → extensions → compute derived metrics via `metrics.Compute*` → assemble `ModuleGraph` with SchemaVersion `"1.1"`, Language, Status, Warnings. Sort `Modules` by `Module.Path` lexicographically for deterministic output. Populate `Extensions` on each `ModuleResult` via the extensions module. Return error on context cancellation. Handle total load failure (zero packages = error). Use relative paths in warning messages.
- [x] 6.2 Handle partial builds: check `packages.Package.Errors`, handle nil `pkg.Types` (set ExportedTypes=0, AbstractTypes=0, LCOM=0, add warning), set `StatusPartial` with warnings for failed packages
- [x] 6.3 Write integration test: analyze a dedicated `testdata/` fixture module with known structure, assert specific metric values, verify output passes `metrics.Validate()`. Guard with `if testing.Short() { t.Skip() }`.
- [x] 6.4 Write determinism test: call `Analyze()` on the same `testdata/` module 10 times, assert JSON output is byte-identical across runs
- [x] 6.5 Write context cancellation test: call `Analyze()` with an already-cancelled context, verify error wraps `context.Canceled`
- [x] 6.6 Write path validation tests: path traversal rejection (`../foo`), non-existent path, empty path — each MUST return error without loading packages
- [x] 6.7 Write context deadline test: call `Analyze()` with a context that has an expired deadline, verify error wraps `context.DeadlineExceeded`
- [x] 6.8 Write error-path tests: total load failure returns error (empty dir fixture), nil type info produces zero metrics with warning (partial-build fixture), partial build sets `StatusPartial` with warning containing package path and error description

## 7. CLI — Analyze Command

- [x] 7.1 Implement `cmd/vibe-check/main.go`: cobra root command with version info (version, commit, date embedded via ldflags), `--version` flag
- [x] 7.2 Implement `cmd/vibe-check/analyze.go`: `RunAnalyze(opts AnalyzeOptions) (*AnalyzeResult, error)` function per AP-002/AP-003 pattern. AnalyzeOptions includes `io.Writer` fields for stdout/stderr. Cobra `RunE` delegates to `RunAnalyze`. Create Go adapter, invoke `Analyze`, marshal result as indented JSON to stdout.
- [x] 7.3 Add threshold flags: `--max-instability`, `--max-distance`, `--no-circular-deps`, `--max-lcom` with validation logic. Reject `--max-instability` and `--max-distance` outside [0.0, 1.0]. Reject `--max-lcom` < 1. Invalid flags exit with status 2. Threshold comparison uses strict `>` (boundary value passes). Add `--timeout <duration>` flag — create `context.WithTimeout` when set, default no timeout.
- [x] 7.4 Implement threshold checking and signal handling: create context with `signal.NotifyContext` for SIGINT/SIGTERM, wrap with `context.WithTimeout` if `--timeout` is set. Iterate `ModuleGraph` results, collect ALL violations, print to stderr, exit 1 if any violations, always emit JSON to stdout. Use exit code 2 for tool errors (invalid args, adapter failure, timeout, signal). If signal received before JSON output, suppress partial JSON.
- [x] 7.5 Write CLI tests — flag parsing and help: test `--help` output lists analyze command, test `--version` output, test flag parsing for all threshold flags
- [x] 7.6 Write CLI tests — threshold violations: test each threshold flag individually (pass/fail), test multiple simultaneous violations reported, test JSON still emitted on violations. Use `bytes.Buffer` for stdout/stderr per AP-003.
- [x] 7.7 Write CLI tests — flag validation and exit codes: test invalid flag values (out of range, negative), test exit code 1 vs 2 distinction, test adapter error produces exit code 2, test `--timeout` creates deadline context, test boundary-value scenarios (metric exactly at threshold passes)
- [x] 7.8 Write CLI tests — JSON output validity: test stdout output passes `metrics.Validate`, test pretty-printed (indented) output

## 8. Validation and CI

- [x] 8.1 Create `.github/workflows/ci.yml` per CI convention pack:
  - Trigger: push to `main` and PRs targeting `main` (CI-014)
  - Go version: match `go.mod` (currently 1.25.7)
  - Steps: `go build ./...`, `go test -race -count=1 -coverprofile=coverage.out ./...`, `go vet ./...`, `golangci-lint run ./...`
  - Pin all actions by 40-character commit SHA (CI-001/CI-002): `actions/checkout`, `actions/setup-go`, `golangci/golangci-lint-action`
  - Add concurrency group per CI-012 to cancel stale runs
  - Set `permissions: contents: read` per CI-020/CI-021 (least privilege)
  - Descriptive workflow name per CI-010/CI-011
  - Add header comment with workflow purpose per CI-031
- [x] 8.2 Verify `go build ./...` passes
- [x] 8.3 Verify `go test -race -count=1 ./...` passes
- [x] 8.4 Verify `go vet ./...` passes
- [x] 8.5 Run `golangci-lint run ./...` if configured, fix any findings
- [x] 8.6 Verify all exported symbols have GoDoc comments

## 9. Documentation

- [x] 9.1 Update AGENTS.md: add `internal/goadapter/` and `cmd/vibe-check/` to Project Structure section, update Architecture section to reflect Go adapter implementation and extensions mechanism, add `go build ./cmd/vibe-check` to Build & Test Commands
- [x] 9.2 Add CHANGELOG.md entry for the go-analyze change

<!-- spec-review: passed -->
<!-- code-review: passed -->
