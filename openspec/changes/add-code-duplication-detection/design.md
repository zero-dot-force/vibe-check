## Context

vibe-check currently computes coupling (Ca/Ce), instability, abstractness, distance, LCOM4, zone classification, and circular dependency detection. It has no mechanism for detecting duplicated code, which is a structural quality concern orthogonal to coupling and cohesion. Issue #13 requests adding AST-based duplication detection as a configurable `vibe-check analyze` threshold gate. This aligns with AD-008 (DRY) from the agent-design convention pack, which specifies `vibe-check analyze --max-duplication=5` as enforcement.

The existing `ModuleResult` struct carries computed metrics and an `Extensions` map for language-specific metric values. The `ModuleGraph` schema is at version 1.2 and includes `Warnings` for analysis caveats. The Go adapter's `Analyze` method already iterates over packages loaded via `go/packages`, giving access to full `go/ast` syntax trees for each package.

## Goals / Non-Goals

**Goals:**
- Detect structurally similar Go function and method bodies using AST normalization, reporting exact file:line locations and similarity scores
- Add `--max-duplication` flag to `vibe-check analyze` (float, range [0.0, 100.0], default 5.0) that causes exit code 1 when a package's duplication percentage exceeds the threshold
- Add a language-agnostic `Duplication` type in `metrics/` and a `Duplications` field to `ModuleResult`
- Bump the output schema to version 1.3 (backward-compatible addition)
- Report violations to stderr in human-readable form alongside other threshold violations
- Follow existing patterns: `metrics.Adapter` extension via `Capabilities`, `validateFlags` for flag validation, `checkThresholds` for violation collection

**Non-Goals:**
- Token-based or text-line-based comparison (AST structural comparison only for Go, type-aware)
- Cross-package duplication detection (within-package only for MVP)
- Duplication detection for non-Go languages (the universal model supports it, but implementation is Go-only)
- Modifying the `diff` or verdict engine — duplication is an `analyze`-time gate only
- Normalization of comments (blank body comparisons strip comments; structural matches with differing comments are still matches)

## Decisions

### D1: AST-based structural similarity with identifier normalization

**Choice**: Compare function/method body AST subtrees after normalizing identifiers (all variable, type, and function names replaced with canonical placeholders) and normalizing basic literals (all integer, float, string, rune literals replaced with canonical placeholders). Two functions match if their normalized AST subtrees are deeply equal.

**Rationale**: This detects true structural duplicates while ignoring superficial differences in variable naming, literal values, and formatting. It avoids false positives from coincidentally similar lines of text. The `go/ast` package provides a robust AST representation; `go/ast.FilterDecl` and inline tree walking can handle normalization without third-party dependencies.

**Alternatives considered**:
- *Line-based hashing (e.g., Simian-style)*: Simpler but produces false positives on boilerplate (error checks, imports, struct tags) and misses structural duplicates with different line breaks or formatting.
- *Normalized token stream comparison*: More granular but loses structural context; two functions with the same control flow but different nesting order could match spuriously.
- *PMD CPD-style token-based approach*: Token-based comparison with an ignore list. More complex to tune; AST-based is a better fit for a Go-native tool.

### D2: Duplications reported per-module in ModuleResult

**Choice**: Add a `Duplications []Duplication` field to `ModuleResult` rather than emitting them as warnings or in the `Extensions` map.

**Rationale**: Duplications are structured data with file:line locations and scores — they are not caveats (warnings) and are not language-specific key-value pairs (extensions). A first-class field makes them discoverable in the JSON schema and allows consumers to parse them without special-case handling. The `Duplication` type is language-agnostic: it carries a `ModulePath`, file paths, line ranges, and similarity score.

**Alternatives considered**:
- *In Extensions map as `go.duplications`*: Would bury the data behind a map of `any`, losing type safety and schema visibility.
- *As ModuleGraph-level collection*: Requires joining back to modules manually; per-module is a better fit since duplication percentage is computed per-package.
- *As warnings*: Warnings are for analysis caveats (load errors, type-check failures), not for duplicated code that was successfully analyzed.

### D3: Duplication percentage computed from lines of duplicated code vs total lines

**Choice**: For each module, compute `DuplicationPct = (duplicatedLines / totalLines) * 100`, where `duplicatedLines` is the deduplicated (count each line at most once across all duplicate pairs) count of lines covered by any duplicate block. Two blocks "match" if their normalized AST is deeply equal and each block is at least 6 significant lines (excluding blank lines and single-brace lines).

**Rationale**: A percentage is the most intuitive threshold for CI gating. The 6-line minimum and deduplication prevent noise from tiny matches like `if err != nil { return err }` or counting the same code multiple times when it appears in 3+ places.

### D4: Bump schema to 1.3

**Choice**: Increment `SchemaVersionCurrent` from `"1.2"` to `"1.3"`.

**Rationale**: Adding a new field (`Duplications`) to `ModuleResult` is backward-compatible (existing consumers that ignore unknown fields are unaffected). A minor version bump communicates the addition transparently. The JSON schema should be updated to match.

## Coverage Strategy

Per Constitution IV (Testability), the following coverage strategy governs all new duplication detection code:

**Target Coverage**: 90% overall statement coverage for the `duplication/` package and any new code in `metrics/` and `internal/goadapter/`.

**Unit/Integration Breakdown**:
- **80% unit tests**: Tests that exercise AST normalization, tokenization, hashing, similarity computation, and pairwise comparison logic in isolation using mock or hand-crafted `go/ast` nodes. No `go/packages` dependency.
- **20% integration tests**: Tests that load real Go packages via the adapter, run `FindDuplications`, and verify the end-to-end pipeline including AST traversal from `go/packages` output.

**Contract Surface Coverage**: All exported functions and exported types in the `duplication/` package must achieve 100% coverage (every line and branch). Internal helpers (unexported normalization, tokenization, hashing) must meet the 90% overall target.

**CI Enforcement**: A package-level coverage check will be added to the CI workflow:
```bash
go test -cover -coverpkg=./duplication/... -count=1 -race ./duplication/...
```
The CI run will fail if `duplication/` coverage falls below 90%. The existing `-race` and `-count=1` flags are preserved.

## Risks / Trade-offs

- **[Performance] AST comparison is O(n²) per package**: For a module with N functions/methods, pairwise comparison of normalized ASTs is O(N²). Performance target: must complete within the existing `--timeout` (default 60s) for packages up to 500 functions. Mitigation: short-circuit by comparing token counts before bag-of-structural-tokens; skip modules with <2 functions. For typical Go packages (10-100 functions), this is negligible. Set a configurable timeout (existing `--timeout` flag covers this).
- **[False negatives] Duplicates with structural differences (loop vs recursion)**: AST normalization catches syntactic structure, not semantic equivalence. This is a known limitation documented in the type's GoDoc.
- **[False positives] Generated code**: Code generated by tools often contains repetitive structures. Mitigation: skip files matching `*.gen.go`, `*.pb.go`, and directories named `mock` or `generated` by default. Consider a future `--no-skip-generated` flag.
- **[Failure modes] AST parse errors and partial analysis**: If `go/packages` successfully loads a package but one or more files fail to parse, those files are skipped while other files in the package are still compared. A `Warning` is emitted for each skipped file. Empty packages (no Go files, or only excluded files) produce an empty `Duplications` array with a 0.0% duplication percentage — no error is raised.