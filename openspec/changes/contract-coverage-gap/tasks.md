## 1. goadapter New() nil-guard assertions

- [x] 1.1 [P] Add `if adapter == nil { t.Fatal("New returned nil") }` to `TestAdapter_Capabilities` (adapter_test.go:27)
- [x] 1.2 [P] Add `if adapter == nil { t.Fatal("New returned nil") }` to `TestAdapter_ContextCancellation` (adapter_test.go)
- [x] 1.3 [P] Add `if adapter == nil { t.Fatal("New returned nil") }` to `TestAdapter_ContextDeadline` (adapter_test.go)
- [x] 1.4 [P] Add `if adapter == nil { t.Fatal("New returned nil") }` to `TestAdapter_CouplingMetrics` (adapter_test.go:21)
- [x] 1.5 [P] Add `if adapter == nil { t.Fatal("New returned nil") }` to `TestAdapter_Determinism` (adapter_test.go:165)

## 2. goadapter Analyze() return assertions

- [x] 2.1 Add `if graph == nil { t.Fatal("Analyze returned nil graph") }` and `len(graph.Modules) > 0` assertion to `TestAdapter_Determinism`
- [x] 2.2 [P] Add graph non-nil assertion to `TestAdapter_ExternalExclusion` (adapter_test.go:96)
- [x] 2.3 [P] Verify `TestAdapter_CouplingMetrics` asserts on graph fields — add `if graph == nil` guard if missing

## 3. (*limitedBuffer).String contract coverage

- [x] 3.1 Locate tests exercising `(*limitedBuffer).String` in metrics/external_test.go
- [x] 3.2 Add explicit assertion on the `String()` return value (expected content, not just non-empty)

## 4. GoDoc annotations for quality-level ambiguous side effects

- [x] 4.1 Run `gaze quality --format=json ./...` and extract side effects with `classification.label == "ambiguous"` and `confidence >= 58`
- [x] 4.2 [P] For each identified function, add GoDoc annotations describing return values or mutations
- [x] 4.3 [P] Verify with `gaze quality` that annotations promoted the classification to contractual

## 5. metrics/ test assertion review

- [x] 5.1 Review `validate_test.go` — ensure error return is explicitly asserted (not just panic-or-not)

## 6. Verification

- [x] 6.1 Run `go test -race -count=1 ./...` — all tests pass
- [x] 6.2 Run `go vet ./...` — no issues
- [x] 6.3 Run `gaze crap --format=json ./...` — verify `summary.avg_contract_coverage` is 100% and no function has `contract_coverage: 0`
- [x] 6.4 Run `gaze crap --format=json ./...` — verify GazeCRAPload remains 0
- [x] 6.5 Update CHANGELOG.md with test improvement entry
- [x] 6.6 Document the 77 unmapped functions as a known limitation for issue #30 (comment on the issue noting the `helper_param` tracing limit)

<!-- spec-review: passed -->
<!-- code-review: passed -->
