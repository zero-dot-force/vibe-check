## 1. Metrics model — Duplication type and schema update

- [x] [P] 1.1 Define `Duplication` and `DuplicateBlock` types in `metrics/duplication.go` with fields: `ModulePath`, `Blocks` ([]DuplicateBlock), `Similarity`; `DuplicateBlock` with `File`, `StartLine`, `EndLine`, `LineCount`. Add full GoDoc comments.
- [x] [P] 1.2 Add `Duplications []Duplication` field to `ModuleResult` in `metrics/graph.go`
- [x] [P] 1.3 Add `CapDuplication` to `metrics/adapter.go` capability constants
- [x] 1.4 Bump `SchemaVersionCurrent` from `"1.2"` to `"1.3"` in `metrics/graph.go`
- [x] 1.5 Add `target` capability to `adapter.go`

## 2. Go adapter — AST-based duplication detection

- [x] 2.1 Create `internal/goadapter/duplication.go` with `detectDuplications(pkgs []*packages.Package) []metrics.Duplication` function
- [x] 2.2 Implement AST normalization: `normalizeFuncBody` that walks a `*ast.FuncDecl` body, replaces identifiers with canonical placeholders, replaces basic literals with canonical placeholders, strips comments
- [x] 2.3 Implement duplicate pairing: for each package, pairwise-compare normalized ASTs of all functions/methods with ≥6 significant lines, building `[]metrics.Duplication` for matching pairs
- [x] 2.4 Implement generated/excluded file detection: skip `*.gen.go`, `*.pb.go`, and files under directories named `mock`, `generated`, `mocks`, `testdata`
- [x] 2.5 Integrate `detectDuplications` into `Adapter.Analyze` in `adapter.go` after type classification (after line 141 in current code), populating `ModuleResult.Duplications` per module
- [x] 2.6 Add `CapDuplication` to `Adapter.Capabilities()` return list in `adapter.go`

## 3. CLI — `--max-duplication` flag and threshold check

- [x] 3.1 Add `MaxDuplication *float64` field to `AnalyzeOptions` struct in `cmd/vibe-check/analyze.go`
- [x] 3.2 Add `--max-duplication` flag to the `analyzeCmd` cobra command with float64VarP, default `5.0`, description
- [x] 3.3 Add `MaxDuplication` validation to `validateFlags`: must be in [0.0, 100.0]
- [x] 3.4 Add duplication threshold logic to `checkThresholds`: for each module, compute `duplicationPct = duplicatedLines / totalLines * 100` with safe divide-by-zero handling, report violation if exceeds `MaxDuplication`

## 4. Test fixtures

- [x] 4.1 Create `internal/goadapter/testdata/duplication/` directory with a test package containing known structural duplicates (two identical functions with different variable names, two different functions, a sub-6-line duplicate)
- [x] 4.2 Create an excluded-file fixture: add `*.gen.go` file with duplicates that should be skipped

## 5. Tests

- [x] [P] [unit] 5.1 Write `TestDetectDuplications_IdenticalStructure` in `internal/goadapter/duplication_test.go`: verifies two structurally identical functions with different identifiers are detected
- [x] [P] [unit] 5.2 Write `TestDetectDuplications_DifferentStructure` in `internal/goadapter/duplication_test.go`: verifies structurally different functions are not detected
- [x] [P] [unit] 5.3 Write `TestDetectDuplications_SkipShortFunctions` in `internal/goadapter/duplication_test.go`: verifies functions under 6 significant lines are excluded
- [x] [P] [unit] 5.4 Write `TestDetectDuplications_SkipGenerated` in `internal/goadapter/duplication_test.go`: verifies `*.gen.go` and excluded directories are skipped
- [x] [P] [unit] 5.5 Write `TestDetectDuplications_NoDuplicates` in `internal/goadapter/duplication_test.go`: verifies a clean package reports no duplications
- [x] [P] [unit] 5.6 Write `TestDetectDuplications_MultipleMatches` in `internal/goadapter/duplication_test.go`: verifies a 3-way duplicate reports 2 pairs
- [x] [integration] 5.7 Write `TestDuplicationThreshold_Violation` in `cmd/vibe-check/analyze_test.go`: integration test verifying exit code 1 when duplication exceeds `--max-duplication`
- [x] [integration] 5.8 Write `TestDuplicationThreshold_Pass` in `cmd/vibe-check/analyze_test.go`: integration test verifying exit code 0 when duplication is within threshold
- [x] [integration] 5.9 Write `TestDuplicationFlag_Validation` in `cmd/vibe-check/analyze_test.go`: verifies invalid flag values produce exit code 2

## 6. Schema and documentation

- [x] [P] 6.1 Update `metrics/modulegraph.schema.json` to include `Duplications` array in `ModuleResult` definition
- [x] 6.2 Run existing test suite: `go test -race -count=1 ./...` and fix any regressions
- [x] [P] 6.3 Run `go vet ./...` and fix any issues
- [x] 6.4 Assess documentation impact per AGENTS.md gate: update `CHANGELOG.md` with change entry, file website documentation issue in `unbound-force/website` repo

<!-- spec-review: passed -->
<!-- code-review: passed -->