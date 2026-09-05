## Why

Gaze analysis (2026-09-02, updated 2026-09-04) reveals two distinct
contract coverage signals that the original issue conflated:

- **Per-test average** (`gaze quality`): **38.9%** across 506 tests.
  This measures whether each test function asserts on the contractual
  side effects of its target. Many tests exercise code without
  asserting on observable returns.
- **Per-function average** (`gaze crap`): **94.4%** across the 36
  functions gaze could map to tests. Only 2 functions have 0%:
  `goadapter.New` and `metrics.(*limitedBuffer).String`.

The real gap is that **77 of 113 functions have no quality mapping at
all** — gaze cannot trace any test to them, mostly because they are
unexported helpers tested indirectly through exported callers. The
`helper_param` unmapped reason in gaze quality data confirms this:
assertions exist but flow through helper function parameters that
gaze's mechanical tracer cannot follow.

Closing the addressable gaps raises confidence that the test suite
verifies observable behavior, not just reachability.

## What Changes

- **Add return-value assertions** to existing `goadapter` tests — 5 tests
  call `goadapter.New()` without asserting on the returned `*Adapter`:
  `TestAdapter_Capabilities`, `TestAdapter_ContextCancellation`,
  `TestAdapter_ContextDeadline`, `TestAdapter_CouplingMetrics`,
  `TestAdapter_Determinism`.
- **Add nil-guard and field assertions** for `goadapter.Analyze` — tests
  that invoke `Analyze` should assert on both the `*ModuleGraph` return
  (non-nil, expected modules populated) and the `error` return.
- **Fix `(*limitedBuffer).String` contract coverage** — the only other
  function at 0% contract coverage. Add or strengthen assertions in
  `metrics/external_test.go` for the `String()` return value.
- **Assert on `Validate` error returns** — verify error vs. nil for both
  valid and invalid inputs where currently only the side effect (panic or
  not) is tested.
- **Promote ambiguous side effect classifications** — add GoDoc annotations
  where gaze's quality pipeline classifies test-level side effects as
  ambiguous. Scope limited to functions where the quality JSON shows
  ambiguous classifications with confidence ≥ 58 (close enough to push
  above the 80 contractual threshold with a `godoc` signal boost).
- **Document unmapped function limitation** — 77 of 113 functions have no
  quality mapping due to gaze's `helper_param` tracing limit. These are
  out-of-scope for this change; improving their mapping requires gaze
  upstream work.

**Note**: `ComputeDelta` and `DecideVerdict` already have 100% per-function
contract coverage in the CRAP data. The 0% reported by the original issue
was a per-test measurement artifact — the tests thoroughly assert on return
values but gaze's quality pipeline cannot trace assertions through
table-driven subtest helpers.

## Capabilities

### New Capabilities

_(none — no new user-facing capabilities)_

### Modified Capabilities

_(none — no spec-level requirement changes; this is a test-quality
improvement targeting the existing test suite)_

## Impact

- **Files modified**: `internal/goadapter/adapter_test.go`,
  `metrics/delta_test.go`, `metrics/verdict_test.go`,
  `metrics/validate_test.go`. Possibly GoDoc comments in
  `internal/goadapter/adapter.go`, `metrics/delta.go`,
  `metrics/verdict.go`, `metrics/validate.go`.
- **No production logic changes** — only test files and documentation
  comments.
- **No dependency changes** — uses only the standard `testing` package.
- **Risk**: Low. All changes are additive assertions in existing tests.
  No behavioral changes, no API changes, no schema changes.
