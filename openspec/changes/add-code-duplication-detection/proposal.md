## Why

Duplicated code is a significant maintainability risk, especially with AI-assisted coding where agents may generate code without awareness of existing implementations. No automated duplication detection currently exists in vibe-check's structural quality gates. Adding AST-based duplication detection as a configurable gate gives developers and agents visibility into redundancy, directly supporting AD-008 (DRY) from the convention pack.

## What Changes

- Add AST-based structural similarity detection for Go code, leveraging `go/ast` to compare function bodies and top-level declarations
- Add a `--max-duplication` flag to `vibe-check analyze` (default: 5.0) that sets the maximum allowed duplication percentage per package
- Include a `Duplications` field in `ModuleResult` reporting duplicated code blocks with file locations, line ranges, and similarity scores
- Exit code 1 when any package's duplication percentage exceeds `--max-duplication`
- Report duplication information in both JSON output and human-readable stderr violation messages
- Add a `metrics.Duplication` type in the universal model for language-agnostic duplication reporting

## Capabilities

### New Capabilities
- `duplication-detection`: AST-based code duplication detection for Go, surfaced as a configurable `vibe-check analyze` threshold gate with structured JSON output

### Modified Capabilities
<!-- No existing specs to modify -->

## Impact

- **`metrics/`**: New `Duplication` type in `module.go` or a new file; `ModuleResult` gains a `Duplications` field; `ModuleGraph` schema may need a minor version bump
- **`internal/goadapter/`**: New duplication detection logic using `go/ast` for structural comparison of function bodies; integrated into the adapter's `Analyze` pipeline
- **`cmd/vibe-check/analyze.go`**: New `--max-duplication` flag and `MaxDuplication` threshold field in `AnalyzeOptions`; `validateFlags` and `checkThresholds` gain duplication checks
- **`.github/workflows/`**: The structural-gate workflow may include `--max-duplication` in its `analyze` invocation
- **`CHANGELOG.md`**: Add change entry for duplication detection feature
- **`README.md`**: May reference new `--max-duplication` flag in CLI examples
- **`.opencode/`**: Agent and skill assets may reference new duplication gate
- **Blog opportunity**: Duplication detection is a significant new capability suitable for a feature announcement