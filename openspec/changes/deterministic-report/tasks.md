# Tasks

## 1. Report Engine (metrics/)

- [x] 1.1 Add `Grade` type, grade-to-emoji mapping (A/B+ 🟢, B/C+ 🟡, C/F 🔴), and the five A–F threshold tables as named constants in a new `metrics/report.go` with GoDoc'd ranges and units; verify `go build ./...` passes and exported symbols have GoDoc
- [x] 1.2 Implement grade classification functions for the five scorecard dimensions (avg distance, max LCOM4, cycles, duplication %, instability spread); verify table-driven boundary tests in `metrics/report_test.go` pass with `go test -race -count=1 ./metrics/` covering exact thresholds (distance 0.1 → B+, LCOM4 4 → C+, duplication 5% → C+, spread 1 → B, one cycle → F)
- [x] 1.3 Implement aggregate computations (average instability, average distance, max LCOM4, cycle count, instability spread with ≤ 0.001 / ≥ 0.999 bounds) and overall duplication percentage (deduplicated by file+range, divided by summed total lines); verify table tests covering the empty module set, 0.0/1.0 instability extremes, and multi-block dedup pass
- [x] 1.4 Implement the recommendation rule engine (10 rules: LCOM4 ≥ 4, distance ≥ 0.5, any cycle, duplication ≥ 5% as 🔴; instability ≥ 0.95, instability ≤ 0.05, Ce > 20, abstractness 0 with > 10 exported types as 🟡; 0 < duplication < 5% as 🟢 watch; single 🟢 all-clear fallback, emitted only when no rule of any severity fires) with ordering (severity → rule-specific metric extremity → path ascending) and cap of 5; verify table-driven tests: each rule fires and does not fire, ordering tiebreaks, cap, and all-clear fallback

## 2. Markdown Rendering (metrics/)

- [x] 2.1 Implement the summary-mode renderer (🏗️ title header with producer version, package count, total lines, no timestamp; 📖 fixed legend linking the metrics guide by absolute URL (https://github.com/zero-dot-force/vibe-check/blob/main/docs/metrics-guide.md); "All Clear"/"Needs Attention" status summary line; templated recommendations); verify unit tests pin formatting (%.2f metrics, %.1f percentage, integer LCOM4/ca/ce)
- [x] 2.2 Implement the detailed-mode renderer sections (📊 package metrics table with full-word headers and zone column, 🔗 coupling summary with max Ca/max Ce/leaves/heavily-depended-on, 🧩 cohesion summary with best/worst LCOM4 and ≥ 4 list, 🔄 cycle chains joined with →, 📋 duplication table, 🏥 health scorecard with overall status, ⚠️ warnings including Zone of Uselessness packages); verify section-renderer unit tests pass
- [x] 2.3 Guarantee byte-stability: modules rendered in path order regardless of input array order, no generation timestamp; verify a unit test renders the same graph twice and a path-reordered copy, asserting byte-identical output

## 3. Report Subcommand (cmd/vibe-check/)

- [x] 3.1 Create `cmd/vibe-check/report.go` with `ReportOptions`/`ReportResult`/`RunReport` per the AP-001..003 conventions (no logic in the cobra layer), `--mode summary|detailed` defaulting to summary and rejecting other values; verify `go build ./cmd/vibe-check` succeeds and `vibe-check report --help` renders
- [x] 3.2 Implement input handling: read the graph file, validate via `metrics.Validate`, and enforce exit codes 0 (clean) / 1 (any 🔴 or 🟡 recommendation, full report still on stdout) / 2 (unreadable or schema-invalid input, error on stderr, nothing on stdout); verify CLI tests cover all three exits
- [x] 3.3 Add golden-file tests with committed `cmd/vibe-check/testdata/` fixtures (a clean graph, a violations graph, and the fictional example-report fixture) for both modes; verify `go test -race -count=1 ./cmd/vibe-check/` passes
- [x] 3.4 Add `--json` to `report` (mutually exclusive with `--mode`; both passed → exit 2 usage error, no stdout), emitting the computed report model (aggregates, per-dimension grades, overall status, fired recommendations with severity and rule) as a single JSON object with the exit-code contract unchanged; verify flag tests and a JSON golden test in `cmd/vibe-check/report_test.go` pass

## 4. Diff Markdown Flag (cmd/vibe-check/)

- [x] 4.1 Add `--markdown` to `diff` with mutual exclusion against `--json` (both → exit 2 usage error, no stdout), reusing the metrics markdown helpers for verdict/reasons/entropy direction/per-module deltas/new and resolved cycles; verify flag tests and a regression test that the default table output is unchanged
- [x] 4.2 Verify diff exit semantics under `--markdown`: `--gate` still exits 1 on REQUEST_CHANGES, markdown output is byte-stable across runs; verify golden-file test and repeat-run test in `diff_test.go` pass

## 5. Agent Asset Rewrite (internal/scaffold/)

- [x] 5.1 Rewrite `internal/scaffold/assets/agents/vibe-check-reporter.md`: summary/detailed flow becomes `vibe-check analyze --output <tmp.json> <dir>` → `vibe-check report --mode <mode> <tmp.json>` → emit markdown verbatim and remove the temp snapshot after rendering; on exit 1 append per-violation explanations with remediation suggestions grounded in code and AGENTS.md; fix the pattern bug (directory-only analyze invocation); trending mode and graceful degradation (binary missing → report failure, no hand-rendered numbers) unchanged; verify the embedded build compiles and the asset review checklist in the change discussion is satisfied
- [x] 5.2 Update the bash allowlist to exactly `vibe-check analyze *`, `vibe-check report *`, `git rev-parse *` plus deny-all, and update `contract_test.go` expectations (required sections for the new flow, allowlist count 2 → 3); verify `go test -race -count=1 ./internal/scaffold/` passes
- [x] 5.3 Sync the contract-protected pair `.opencode/agents/vibe-check-reporter.md` with the rewritten embedded asset; verify the scaffold sync check passes
- [x] 5.4 Regenerate `.opencode/references/vibe-check-example-report.md` from the committed fictional fixture using the new renderer (keeping the fictional-data disclaimer); verify the output matches the renderer golden test format
- [x] 5.5 Rewrite the `/vibe-check` command asset pair (`internal/scaffold/assets/commands/vibe-check.md` and `.opencode/commands/vibe-check.md`): mode table and examples describe the new flow (directory-only invocation routed through `vibe-check report`, mode descriptions matching the report-command spec), removing the package-pattern example that exits 2; verify the pair stays byte-identical and every documented example runs as written

## 6. CI Step Summary (.github/workflows/)

- [x] 6.1 Add a render step to `structural-gate.yml` after the gate step: `if: always()` and `continue-on-error: true`, appending `bin/vibe-check diff --markdown base.json head.json` to `$GITHUB_STEP_SUMMARY`, inside the existing fork-guarded job with the gate step untouched; verify YAML validity (actionlint or `python3 -c` yaml parse) and a review confirming the fork guard and gate step are unmodified
- [x] 6.2 Run CI parity locally using the exact commands from `.github/workflows/ci.yml` (`go build ./...`, `go test -race -count=1 ./...`, `go vet ./...`, plus golangci-lint if configured); verify all pass on the feature branch

## 7. Documentation and Integration

- [x] 7.1 Update README with the `report` subcommand and `diff --markdown` usage; verify every documented command runs as written
- [x] 7.2 Update AGENTS.md project structure (metrics/report.go, cmd/vibe-check/report.go) and build-and-test command examples; verify structure matches the actual tree
- [x] 7.3 Add a CHANGELOG entry leading with the shared byte-stable report (CLI, CI step summary, agent) and explicitly mentioning the new CI step-summary rendering; verify format matches existing entries and spec is referenced
- [x] 7.4 File a documentation issue with the `docs:` title prefix for the user-facing report format change, searching by title first to avoid duplicates (labels are not created in this repo), and file the constitution-required website-sync issue in the `unbound-force/website` repository documenting the change; verify both issues exist and are open before PR
- [x] 7.5 Final integration check: run `go run ./cmd/vibe-check analyze -o /tmp/vibe-check-self.json .` then `go run ./cmd/vibe-check report --mode detailed /tmp/vibe-check-self.json`; verify the report renders, exits 1 (known `internal/goadapter` LCOM4=4 violation), and the recommendations match the engine output

## Workflow follow-up

- Run `/uf.review-council` and resolve all REQUEST CHANGES findings before PR submission.
- Ensure the docs issue from 7.4 stays open until PR merge (documentation gate).
- Archive the change (`/opsx-archive`) only after implementation completes, review requirements are satisfied, and both predecessor changes (`vibe-check-command-and-reporter`, `revamp-vibe-check-reporter`) have archived — reconciling the reporter-agent delta against the archived capabilities via MODIFIED/REMOVED so main specs stay non-contradictory.

<!-- spec-review: passed -->
<!-- code-review: passed -->
