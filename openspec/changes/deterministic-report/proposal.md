# Proposal

## Why

The pretty architectural health report exists only inside the `vibe-check-reporter` LLM agent (a scaffolded markdown prompt). Roughly 90% of that agent's output contract is mechanical — formatting, emoji zones, A–F health grades, duplication percentages, recommendation ordering — yet it lives in prompt text. Consequences: CI (the structural-gate workflow) renders only raw tabwriter tables instead of the report; agent output varies with the model; and every formatting rule is untestable prose (a prior review found 4 findings of template drift in the manually-maintained example report). This contradicts the project mandate that metric computations produce deterministic results. Moving the deterministic portion into Go lets CI, agents, and humans share one byte-stable report, and reserves the LLM for what it is actually good at: explaining why thresholds fail and suggesting remediations.

## What Changes

- **New `metrics/report/` package (Layer 1, pure)**: grade engine with A–F threshold tables (avg distance, max LCOM4, cycles, duplication %, instability spread), grade-to-emoji mapping, aggregate computations, overall duplication percentage, and a deterministic recommendation rule engine (condition → severity → template, ordered, capped).
- **New `vibe-check report <graph.json>` subcommand**: reads a previously emitted ModuleGraph JSON, validates it, and renders the markdown report to stdout. `--mode summary|detailed` (default summary). `--json` (mutually exclusive with `--mode`) emits the computed report model — aggregates, grades, overall status, fired recommendations — as machine-readable JSON, so markdown is not the only format. Byte-stable output (no timestamps). Exit codes: 0 clean, 1 threshold violations present, 2 tool failure (unreadable/schema-invalid input).
- **`vibe-check diff --markdown` flag**: renders the existing verdict/reasons/per-module deltas/new+resolved cycles as markdown. Mutually exclusive with `--json`; default table output unchanged. Enables CI to publish the diff report.
- **`structural-gate.yml` render step**: a new always-run step appends `vibe-check diff --markdown` output to `$GITHUB_STEP_SUMMARY`, giving every (non-fork) PR the report in the Checks UI. Stays inside the existing fork-guarded job; the gate step itself is untouched.
- **Reporter agent asset rewrite**: the agent's flow becomes `analyze --output` → `vibe-check report` → emit the Go-rendered markdown → (if exit 1) explain each violation and suggest remediations citing code and AGENTS.md. Fixes a bug: the asset instructs passing package patterns (e.g. `./internal/...`) to `analyze`, which rejects them (exit 2, path validation). Bash allowlist gains `vibe-check report *`; scaffold contract tests updated; `.opencode/` synced copy updated.
- **Docs**: README, AGENTS.md (structure + commands), CHANGELOG, package GoDoc; documentation issue filed before PR.

No breaking changes: `analyze` is untouched, `diff` default output is unchanged, `report` is additive, and the agent asset ships to consumers via an explicit `vibe-check init --force`.

## Capabilities

### New Capabilities

- `report-command`: The `vibe-check report` subcommand — deterministic markdown report rendering from a ModuleGraph JSON snapshot, including summary/detailed modes, grade and scorecard computation, recommendation generation, exit-code contract, and byte-stable output.
- `diff-command`: Markdown output for `vibe-check diff` — the `--markdown` flag rendering verdict, reasons, per-module deltas, and cycle changes as markdown, mutually exclusive with `--json`, with the table default unchanged. (Same capability path used by the unarchived `add-divisor-entropy-agent` change; this change adds the markdown output requirements.)
- `structural-gate-report`: The structural-gate CI workflow's report rendering — an always-run step inside the fork-guarded job that appends the markdown diff report to the PR step summary without affecting the gate verdict.
- `vibe-check-reporter-agent`: The `vibe-check-reporter` agent's post-change behavior — consume the Go-rendered report, explain threshold violations with remediation suggestions, own the trending mode, and operate within its bash allowlist. Supersedes the formatting contract of the in-flight `revamped-reporter-agent` capability (whose implementation is complete; remaining tasks are manual verification).

### Modified Capabilities

<!-- None — openspec/specs/ is empty; no archived capabilities to modify. -->

## Impact

- **Code**: new `metrics/report/` package (+ tests); new `cmd/vibe-check/report.go` (+ golden-file tests); `cmd/vibe-check/diff.go` gains `--markdown` rendering; `internal/scaffold/assets/agents/vibe-check-reporter.md` rewritten; `internal/scaffold/contract_test.go` allowlist updated (2 → 3 entries); `.github/workflows/structural-gate.yml` gains one render step.
- **Sync pairs**: `.opencode/agents/vibe-check-reporter.md` must be updated alongside the embedded agent asset, and `.opencode/commands/vibe-check.md` alongside the embedded command asset (both contract-protected) — the command asset's mode table and examples still document the removed package-pattern invocation and are rewritten with the agent asset. `.opencode/references/vibe-check-example-report.md` becomes verifiable/regenerable against the Go renderer, closing a known drift risk.
- **Dependencies**: none added (stdlib only; markdown rendered with `strings.Builder` string building per design D6, no third-party libraries).
- **Coordination**: two unarchived predecessor changes land and archive before this one.
  - `vibe-check-command-and-reporter` (implementation complete): its `summary-mode` and `detailed-mode` capabilities are superseded by this change — package-pattern invocation (`/vibe-check summary ./internal/...`) is replaced by directory-only `vibe-check analyze` (patterns fail path validation with exit 2), agent-side report rendering is replaced by verbatim emission of `vibe-check report` output, and its tempfile-cleanup requirement is carried forward in the rewritten agent asset (snapshot written via `analyze --output`, removed after rendering).
  - `revamp-vibe-check-reporter` (22/28 tasks; remaining are manual verification and a website docs issue): lands first — this change builds on its agent output and supersedes the formatting portions of its `revamped-reporter-agent` capability.
  - Archival sequencing: both predecessors archive before this change; at this change's archive, the reporter-agent delta is reconciled against the then-archived capabilities via MODIFIED/REMOVED so `openspec/specs/` never holds contradictory requirements. Until then, this change expresses its requirements as ADDED against the currently-empty main specs.
- **Out of scope**: trending math (stays agent-side with `mx-f-architecture-trend`), PR comments, package-pattern support in `analyze`, and the stale schemaVersion 1.2/1.3 doc mismatch (separate fix).
