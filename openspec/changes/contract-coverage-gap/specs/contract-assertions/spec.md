## ADDED Requirements

### Requirement: goadapter.New return assertion

Every test function that calls `goadapter.New()` SHALL assert that the
returned `*Adapter` is non-nil before proceeding with further operations.

#### Scenario: nil guard on New
- **WHEN** a test calls `goadapter.New()`
- **THEN** the test asserts `adapter != nil` using `t.Fatal`

### Requirement: goadapter.Analyze return assertions

Every test function that calls `(*Adapter).Analyze` SHALL assert on both
return values: the `*ModuleGraph` (non-nil and structurally valid) and
the `error` (nil for success paths, non-nil for error paths).

#### Scenario: success path assertions
- **WHEN** a test calls `Analyze` on a valid fixture
- **THEN** the test asserts `graph != nil` and `err == nil`
- **THEN** the test asserts at least one structural field of the returned
  `ModuleGraph` (e.g., `len(graph.Modules) > 0`)

#### Scenario: error path assertions
- **WHEN** a test calls `Analyze` with a context expected to fail
- **THEN** the test asserts `err != nil`

### Requirement: (*limitedBuffer).String return assertion

Tests exercising `(*limitedBuffer).String` SHALL assert on the returned
string value rather than relying on indirect observation.

#### Scenario: String return value asserted
- **WHEN** a test calls `(*limitedBuffer).String()`
- **THEN** the test asserts the returned string matches the expected content

### Requirement: GoDoc side effect annotations

Functions with ambiguous side effect classifications in the gaze quality
pipeline SHALL have GoDoc comments updated when their classification
confidence is ≥ 58, so the added `godoc` signal pushes them above the
80 contractual threshold.

#### Scenario: ambiguous promotion via GoDoc
- **WHEN** a function's quality-level side effect has confidence ≥ 58
  and classification `ambiguous`
- **THEN** a GoDoc annotation describing the return value or mutation is
  added to the function's documentation comment

#### Scenario: already-contractual functions unchanged
- **WHEN** a function's classify-level side effects are all `contractual`
  (confidence ≥ 80)
- **THEN** no GoDoc changes are made for classification purposes

### Requirement: per-function contract coverage target

After all assertion additions, the per-function average contract coverage
as reported by `summary.avg_contract_coverage` in `gaze crap --format=json
./...` SHALL be 100% (all mapped functions at full coverage), and the
number of functions at 0% SHALL be 0.

#### Scenario: CRAP contract coverage measurement
- **WHEN** `gaze crap --format=json ./...` is executed
- **THEN** `summary.avg_contract_coverage` is 100
- **THEN** no function in the `scores` array has `contract_coverage: 0`

