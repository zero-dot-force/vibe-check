# Design

## Context

The architectural health report currently exists only as formatting rules inside the `vibe-check-reporter` agent prompt (567-line embedded asset). The CLI emits JSON (`analyze`, `diff`) and plain tables (`diff`), CI renders only raw log output, and the manually-maintained example report (`.opencode/references/vibe-check-example-report.md`) has already drifted from the prompt contract once. See `proposal.md` for motivation.

Constraints that shape this design:

- Three-layer architecture: `metrics/` (pure, language-agnostic), `internal/goadapter/` (Go adapter), `cmd/vibe-check/` (CLI). The report engine belongs in Layer 1 because every adapter produces the same `ModuleGraph`.
- CLI conventions: cobra commands delegate to `Run<Cmd>` functions with `Options`/`Result` structs and no logic in the cobra layer (agent-design pack AP-001..003); exit codes 0 clean / 1 policy violation / 2 tool failure (established by `analyze`).
- Protected gates: `diff` verdict thresholds (ΔI ≥ 0.15, ΔD ≥ 0.20, ΔLCOM ≥ 2, new cycle) are tighten-only protected values. Report grades are informational and MUST NOT become gates.
- Determinism mandate: identical input MUST produce identical output (AGENTS.md).
- Security: `analyze` compiles the target module, so it must never run on untrusted fork PRs; `structural-gate.yml` is fork-guarded at job level.
- The in-flight `revamp-vibe-check-reporter` change (implementation complete, manual verification remaining) owns the current agent contract this change supersedes.

## Goals / Non-Goals

**Goals:**

- One deterministic markdown renderer shared by CLI, CI, and the agent.
- Recommendation and grading logic as pure, table-tested Go functions in `metrics/`.
- The PR Checks UI shows the structural-entropy report via step summary.
- The agent's LLM budget goes to violation explanation and remediation, not arithmetic.

**Non-Goals:**

- Trending math (stays with the agent and `mx-f-architecture-trend`).
- Making report grades enforcement gates; the `diff` verdict remains the only gate.
- Changing `analyze` behavior or adding package-pattern support to it.
- PR comment posting (step summary only).
- Fixing the stale schemaVersion 1.2/1.3 documentation mismatch.

## Decisions

### D1: Report engine lives in `metrics/report/` as pure functions

Grades, aggregates, spread, duplication percentage, and recommendation rules become pure functions over `ModuleGraph` — no I/O, no CLI dependencies. Rendering (string building) also lives there, returning markdown strings.

- **Rationale**: any adapter's `ModuleGraph` can be reported; the engine is table-testable without process spawns; keeps the cobra layer thin per AP-001..003.
- **Alternatives**: rendering in `cmd/vibe-check/` (couples report to CLI, blocks future non-Go adapters from reusing it); keeping rules in the agent prompt (status quo — untestable, non-deterministic).
- **Amendment (2026-10-23, post-review restructure)**: the engine originally landed as four files directly in `metrics/` (`report.go`, `render.go`, `render_diff.go`, `reportmodel.go`). That raised the `metrics` package's LCOM4 from 3 to 5 — tripping this change's own protected ΔLCOM ≥ 2 entropy gate — because the report engine's method-field cluster is disjoint from the universal model's. The human chose restructure over accept-and-document at code review: the four files and their tests moved, byte-identical in behavior, into the Layer-1 subpackage `metrics/report/`, which imports the parent `metrics` types but adds no adapter or CLI dependencies. D1's intent — pure, adapter-agnostic, table-testable, thin cobra layer — is unchanged; only the import path moved. Exported identifier names were deliberately kept stable so callers change only their import path; revive's stutter warnings on `ReportStatus`/`ReportGrades`/`ReportModel` are waived per-site with documented `//nolint:revive` directives rather than by weakening the lint gate.

### D2: `report` reads a snapshot; it never analyzes

`vibe-check report <graph.json>` consumes a previously emitted ModuleGraph file. It does not invoke the adapter or `go/packages`.

- **Rationale**: analysis is the expensive, compiling, security-sensitive step; rendering is pure. Splitting them lets CI render from artifacts the gate job already produces, enables golden-file tests on committed fixtures, and keeps report usable on snapshots from any source.
- **Alternatives**: `analyze --markdown` flag (recompiles per render, harder to test, couples output mode to analysis).

### D3: CLI surface is a new subcommand plus a diff flag (user decision)

`vibe-check report <graph.json> --mode summary|detailed` and `vibe-check diff ... --markdown`. Table default of `diff` is untouched.

- **Rationale**: matches the approved product direction; `--markdown` on `diff` is where the CI consumer lives; a subcommand fits the existing analyze/diff/init command family.
- **Alternatives**: single `--report` flag on analyze (rejected in planning); separate binary (unnecessary surface).

### D4: Byte-stable output rules

- No generation timestamp anywhere in markdown (the JSON snapshot's `generatedAt` is not rendered).
- Modules rendered in path-sorted order regardless of array order.
- Floats formatted with fixed precision: instability/abstractness/distance `%.2f`, duplication percentage `%.1f`, LCOM4/ca/ce as integers.
- Recommendations ordered severity (🔴 > 🟡 > 🟢) → rule-specific metric extremity → path ascending, capped at 5.

- **Rationale**: golden tests, CI diffability, and the determinism mandate. Fixed precision mirrors the agent template's existing 2-decimal contract.
- **Alternatives**: rendering timestamps (breaks byte-stability); `%v` float formatting (Go's shortest-representation floats vary in width and hurt table stability).

### D5: Grades and thresholds are named constants, informational only

The five A–F threshold tables (avg distance, max LCOM4, cycles, duplication %, spread) are expressed as named constants and pure classification functions with table-driven boundary tests. The `diff` verdict thresholds are not touched.

- **Rationale**: AGENTS.md forbids weakening gates; adding new gates is out of scope. Grades advise; the verdict gates enforce.
- **Alternatives**: making any grade a failure gate (scope creep, changes CI semantics).

### D6: Markdown built with stdlib string building, no new dependencies

Rendering uses `strings.Builder` and small helper functions in the style of the existing `writeDiffTable` — no `text/template`, no third-party markdown libraries.

- **Rationale**: zero new dependencies (project rule: justify every dependency); the output is simple tables and headings; golden tests pin exact bytes.
- **Alternatives**: `text/template` (indirection without benefit at this size); goldmark or similar (heavyweight for static structure).

### D7: Exit codes mirror `analyze`

`report` exits 0 when no 🔴/🟡 recommendation fires, 1 when any fires, 2 on tool failure (unreadable/schema-invalid input, via `metrics.Validate`). The full report is still written to stdout on exit 1.

- **Rationale**: uniform, scriptable contract across subcommands; lets the agent branch on exit status instead of parsing output.
- **Alternatives**: always exit 0 (forces output parsing); separate `--fail-on-violation` flag (unnecessary; the violation set is fixed by the rule table).

### D8: `diff --markdown` reuses the report helpers, mutual exclusion with `--json`

The markdown diff renderer shares section helpers (headings, tables, cycle chains) with `report`. Passing both `--markdown` and `--json` exits 2 with a usage error. Verdict computation, protected thresholds, and gate exit semantics are untouched.

- **Rationale**: one formatting vocabulary across both reports; `--json` consumers (divisor-entropy agent) must not silently lose their format.
- **Alternatives**: separate renderer for diff (duplicates table/heading code).

### D9: Agent becomes a thin consumer plus explainer

The rewritten agent asset: run `analyze --output <tmp.json> <dir>`, run `report --mode <mode> <tmp.json>`, emit its markdown verbatim as the skeleton, then — when the report exits 1 — append per-violation explanations and remediation suggestions grounded in code and AGENTS.md. Trending mode is unchanged. The pattern bug (asset instructs `./internal/...` package patterns, which `metrics/security.go` path validation rejects) is fixed by specifying directory-only invocation. Bash allowlist gains `vibe-check report *` (2 → 3 entries). The `.opencode/agents/` synced copy is updated in the same commit (contract-test enforced).

- **Rationale**: keeps the LLM where it adds value; the skeleton is byte-identical to CLI output, making agent output partially golden-verifiable.
- **Alternatives**: agent fully hand-renders (status quo); agent renders and the CLI explains (nonsensical).

### D10: CI render step is `if: always()` + `continue-on-error`, inside the fork-guarded job

`structural-gate.yml` gains a step after the gate that appends `bin/vibe-check diff --markdown base.json head.json` to `$GITHUB_STEP_SUMMARY`. It runs regardless of the gate step's outcome (`if: always()`), is `continue-on-error: true`, and lives inside the existing fork-guarded job. The gate step is not modified.

- **Rationale**: the report must appear on failing gates (that is when it matters most); a rendering failure must not block merges (the gate step remains the authoritative conclusion); the fork guard already encodes the trusted-execution boundary and must not be weakened. `diff` is a pure function of two JSON files, so re-running it is cheap.
- **Alternatives**: gating the render step on success (report invisible exactly when needed); no `continue-on-error` (a cosmetic failure would fail a required check).

### D11: Example report reference is regenerated to the new format

`.opencode/references/vibe-check-example-report.md` is rewritten to match the new renderer's output for a committed fictional fixture; the renderer's golden tests (in package `testdata/`) machine-pin the format the reference documents. The fictional-data disclaimer stays.

- **Rationale**: the reference previously drifted from the template (4 review findings); pinning the format in golden tests closes the class of drift.
- **Alternatives**: deleting the reference (loses onboarding value); leaving it (known drift risk).

### D12: `report --json` emits the computed model (Constitution III)

`--json` serializes the report model — aggregates, per-dimension grades, overall status, fired recommendations with severity and rule — as a single JSON object, mutually exclusive with `--mode`. The exit-code contract is unchanged.

- **Rationale**: Constitution III requires that human-readable output not be the only format; the engine computes the model as data before rendering, so JSON is a thin serialization. Gives drift-tracking consumers (entropy sentinel, trend agents) a stable machine contract that does not depend on parsing markdown.
- **Alternatives**: documenting a rationale instead (compliance stays an argued position and non-Go consumers get no machine surface); deferring to a follow-up change (retrofitting the JSON shape after release risks a breaking contract change).

## Test Strategy

Constitution Principle IV coverage, three categories:

1. **Pure engine tests (`metrics/report/*_test.go`)** — table-driven: grade boundaries for all five dimensions (exact threshold values, e.g. avg distance 0.1 → B+, LCOM4 4 → C+, duplication 5% → C+); grade-to-emoji mapping; aggregate and spread computation (including 0.0/1.0 extremes); duplication percentage (single block, multiple blocks dedup, zero total lines guard); every recommendation rule fires/not-fires; ordering (severity, extremity, path tiebreak) and the cap of 5; empty module set.
2. **CLI tests (`cmd/vibe-check/report_test.go`, `diff_test.go` additions)** — golden-file tests for summary and detailed markdown (committed `testdata/` fixtures, including a fixture matching the example report); exit codes 0/1/2; `--mode` validation (rejects `trending`); `--json` mutual exclusion with `--mode` plus a JSON golden test; `--markdown`/`--json` mutual exclusion; diff default table unchanged; byte-stability (render twice, compare); schema-invalid and unreadable inputs exit 2 with nothing on stdout.
3. **Asset contract tests (`internal/scaffold/contract_test.go`)** — reporter agent allowlist is exactly `vibe-check analyze *`, `vibe-check report *`, `git rev-parse *` plus deny-all; required sections updated for the new flow (report consumption, violation explanations); frontmatter unchanged (mode, temperature, deny rules).

CI parity gate (from `.github/workflows/ci.yml`): `go build ./...`, `go test -race -count=1 ./...`, `go vet ./...`, `golangci-lint run ./...` (when configured) — all run locally before any task is marked complete.

## Risks / Trade-offs

- [Golden files churn on cosmetic format tweaks] → goldens are the contract; intentional format changes update goldens in the same commit and are reviewed as such.
- [Report exit 1 surprises scripts expecting 0] → new command with a documented contract, matching `analyze`'s existing policy-exit convention; no existing callers.
- [Agent asset rewrite collides with the in-flight revamp change] → sequence after its remaining manual-verification tasks; this change's tasks update the contract tests atomically with the asset.
- [Schema-version doc drift (1.3 emitted vs 1.2 documented)] → `report` validates via `metrics.Validate` and accepts whatever versions it accepts; the AGENTS.md schemaVersion staleness is tracked as a separate fix, out of scope here.
- [Render step silently failing behind `continue-on-error`] → failure is visible in the Actions UI log; the gate step remains authoritative by design.
- [GitHub truncates step summaries beyond 1 MiB, cutting very large PR reports] → accepted limit: the step summary is a convenience view; the full `diff --markdown` output remains in the step log, and the gate verdict is unaffected.
- [Emoji rendering varies across viewers] → closed 12-emoji set already proven by the current agent report; markdown tables do not depend on monospace alignment.
- [The `metrics/report/` subpackage could itself accumulate a second disjoint concern cluster over time] → its creation is what restored `metrics` cohesion; the structural-gate diff watches it like every package once it exists in the base snapshot, and further splits (engine vs render vs model) remain available if a new cluster emerges.

## Migration Plan

All changes are additive or opt-in: `report` is a new subcommand; `diff` default output is unchanged; the workflow change ships in-repo with the gate step untouched; the agent asset reaches downstream consumers only via an explicit `vibe-check init --force`. Rollback is reverting the commit — no data or on-disk format migrations. The example report reference is regenerated once in this change (D11).

## Open Questions

- Whether `report` should also accept `-o/--output` to write markdown to a file (as `analyze` does) — safely deferrable; stdout redirection covers known consumers.
- Resolved during spec review: the `/vibe-check` command asset pair is rewritten in task 5.5 — its mode table and examples now describe the new flow (directory-only invocation, `report`-rendered output), removing the package-pattern example that exits 2.
