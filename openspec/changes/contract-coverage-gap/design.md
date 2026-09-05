## Context

Gaze analysis reveals a split between two contract coverage signals:

- **Per-function** (`gaze crap`): 94.4% average across 36 mapped
  functions. Only 2 have 0%: `goadapter.New` (return value unasserted
  in 5 tests) and `metrics.(*limitedBuffer).String`.
- **Per-test** (`gaze quality`): 38.9% average across 506 tests. Many
  test functions exercise code without asserting on contractual returns.

The dominant gap is that **77 of 113 functions have no quality mapping**
— gaze cannot trace any test to them. The `unmapped_reason` field shows
`helper_param` for most: assertions exist but flow through helper
function parameters that gaze's mechanical tracer cannot follow. These
77 functions are out-of-scope; improving their mapping is a gaze
upstream concern.

The addressable scope is the 2 functions at 0% plus assertion
strengthening in tests gaze can already trace.

## Goals / Non-Goals

**Goals:**
- Bring all per-function contract coverage to 100% (currently 94.4%
  avg; 2 functions at 0%: `goadapter.New`, `(*limitedBuffer).String`)
- Add nil-guard assertions for `goadapter.New()` in 5 existing tests
- Add explicit return-value assertions for `goadapter.Analyze` in tests
  that currently rely on downstream field access
- Fix `(*limitedBuffer).String` contract coverage in external_test.go
- Promote ambiguous quality-level side effect classifications where
  GoDoc annotations can push confidence above the 80 threshold
- Maintain or increase line coverage (currently 91.7%)

**Non-Goals:**
- Refactoring production code — test files and GoDoc comments only
- Achieving 100% per-test quality coverage (38.9% → depends on gaze
  upstream tracing improvements for the 77 unmapped functions)
- Adding new test functions — this change strengthens existing tests
- Fixing gaze's `helper_param` tracing limit — that's gaze upstream
- Modifying table-driven patterns in delta/verdict tests — they already
  have 100% per-function contract coverage

## Decisions

### 1. Assertion style: nil-guard then field assertion

**Decision**: Use `if x == nil { t.Fatal("...") }` followed by field-level
assertions, matching the existing codebase pattern.

**Rationale**: This is already the pattern used in `TestAdapter_Integration`
(adapter_test.go:116–162). Consistency over novelty. The `t.Fatal` nil guard
prevents nil dereference panics in subsequent assertions.

**Alternative considered**: `reflect.DeepEqual` on full structs — rejected
because it's brittle (breaks on any field addition) and the existing tests
already use targeted field assertions.

### 2. Scope: modify existing tests, don't add new ones

**Decision**: Add assertions to the 5 identified tests rather than creating
new test functions.

**Rationale**: The tests already exercise the right code paths. The problem
is missing assertions, not missing coverage. Adding new tests would increase
the test count without addressing the contract gap.

### 3. GoDoc annotations: narrowly scoped

**Decision**: Add GoDoc annotations only to functions where gaze's
*quality pipeline* reports ambiguous side effects with confidence ≥ 58.
The `gaze analyze --classify` output already shows 0 ambiguous / 0
incidental across all exported functions — the ambiguity is at the
test-level quality pipeline, not the function-level classify pipeline.

**Rationale**: GoDoc adds a `godoc` signal (weight ~15) to gaze's
confidence scoring. This only helps functions whose confidence is
between 58–79 (adding ~15 pushes them above the 80 contractual
threshold). Functions already at 80+ get no benefit. Unexported
functions get lower `visibility` weight, so GoDoc on them has less
total impact.

**Alternative considered**: Blanket GoDoc on all 77 unmapped functions
— rejected because the unmapped issue is `helper_param` tracing, not
classification confidence. GoDoc wouldn't change the mapping.

### 4. Delta and verdict tests: minimal changes

**Decision**: Make minimal additions to `delta_test.go` and
`verdict_test.go` — only where specific return fields are genuinely
unasserted. Do not restructure the table-driven patterns.

**Rationale**: Review of the existing tests shows they already assert on
`GraphDelta` fields (Ca, Ce, Instability, Abstractness, Distance, LCOM,
Direction, Unreliable, Added, Removed, NewCycles, ResolvedCycles) and on
`Verdict` + `[]string` reasons. The 0% contract coverage is a gaze
mechanical-tracing limitation, not an assertion gap. Restructuring to
help gaze's tracer would harm readability for negligible benefit.

## Risks / Trade-offs

- **Risk**: GoDoc annotations may need to follow a specific format that
  gaze recognizes. → Mitigation: verify with `gaze quality` after adding
  annotations; adjust format if needed.
- **Risk**: Adding assertions to existing tests could make them fail if
  assumptions about return values are wrong. → Mitigation: run the full
  test suite after each file is modified.
- **Risk**: The `helper_param` tracing limit means per-test coverage
  (38.9%) may not significantly improve even after all assertion
  additions. The 77 unmapped functions will remain unmapped. →
  Accepted: this change targets the *real* assertion gaps (2 functions
  at 0% per-function cc), not the measurement artifact. The per-test
  number is tracked as a gaze upstream concern.
- **Trade-off**: The per-function metric (`summary.avg_contract_coverage`
  in CRAP JSON) is already at 94.4%. This change will push it toward
  100% but the headline improvement is modest in that metric. The value
  is in closing genuine gaps, not moving a number.
