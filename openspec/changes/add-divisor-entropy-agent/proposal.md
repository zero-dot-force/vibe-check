## Why

The Review Council auto-discovers `divisor-*` agents, but no reviewer guards
*structural* quality across a PR. Nothing today prevents a PR from silently
increasing coupling, introducing circular dependencies, or eroding cohesion.
This change adds a `divisor-entropy` agent that measures the structural delta
between the base and PR branches (via `vibe-check analyze`) to enforce the Boy
Scout Rule — leave the code better than you found it — and ships it to consuming
projects through a new `vibe-check init` command. (Implements RFC Group 2a,
Phase P1 "Enforcement"; depends on the now-complete `vibe-check analyze`.)
This change delivers only the point-in-time base↔PR comparison primitive needed
for P1 enforcement; full P2 longitudinal architectural-drift tracking (comparing
against a historical baseline series, trend analysis, and drift alerting) remains
future work.

## What Changes

- Add an `internal/scaffold/` package that embeds agent assets via `//go:embed`
  and writes them into a target project's `.opencode/agents/` directory.
- Add the embedded asset `internal/scaffold/assets/agents/divisor-entropy.md` —
  a Review Council divisor that computes the base↔PR structural delta by running
  `vibe-check analyze` on both refs (using an isolated `git worktree`) and then
  `vibe-check diff` to compute the metric deltas and verdict, rendering an
  APPROVE / REQUEST CHANGES / COMMENT result backed by those deltas.
- Add a structural delta/verdict engine (`metrics/delta.go`, `metrics/verdict.go`)
  and a `vibe-check diff <base.json> <pr.json>` CLI command that computes exact
  per-package metric deltas, classifies newly introduced cycles, and renders a
  deterministic verdict through documented, table-tested threshold gates — so the
  agent reports a Go-computed verdict rather than performing arithmetic in-prompt.
- Add a `vibe-check init [path]` CLI command that deploys embedded agent assets
  into the target project (idempotent; `--force` to overwrite; `--json` summary),
  with restrictive file permissions and path validation.
- Dogfood the agent into vibe-check's own `.opencode/agents/` by generating it
  with `vibe-check init` (single source of truth: the embedded asset).
- **Notable deviation**: `divisor-entropy` is the second divisor to enable `bash`
  via a granular allowlist (after `divisor-curator`), permitting only
  `vibe-check analyze`, `vibe-check diff`, and a narrow set of ref-resolution and
  `git worktree` subcommands; all other commands are denied. It mirrors the
  `"*": "deny"` + allowlist shape already established by `divisor-curator`. `edit`
  and `webfetch` remain `deny` (least privilege).

## Capabilities

### New Capabilities

- `divisor-entropy-agent`: Review Council divisor that computes the structural
  quality delta between base and PR branches — package coupling (Ca, Ce,
  Instability), newly introduced circular dependencies, and cohesion (LCOM) —
  and issues a verdict (APPROVE / REQUEST CHANGES / COMMENT) justified by metric
  deltas. Specifies discovery (the `divisor-*` naming convention), tooling
  (`vibe-check analyze` on both refs via `git worktree`, then `vibe-check diff`
  for the delta and verdict), agent permissions, and ref-handling safety.
- `diff-command`: CLI command `vibe-check diff <base.json> <pr.json>` and the
  supporting `metrics` delta/verdict engine. Reads two `ModuleGraph` documents,
  computes exact per-package metric deltas, classifies cycles as new /
  pre-existing / resolved, and renders a deterministic verdict (APPROVE /
  COMMENT / REQUEST_CHANGES) via documented threshold gates, emitting both JSON
  and a human-readable table.
- `init-command`: CLI command `vibe-check init [path]` and the `internal/scaffold`
  mechanism that deploys embedded agent assets into a target project's
  `.opencode/agents/` — idempotent, permission-safe, path-validated, with an
  optional JSON summary.

### Modified Capabilities

- `analyze-command`: security hardening — forces `GOTOOLCHAIN=local` in the
  analysis subprocess environment so that an untrusted module's `go.mod`
  `toolchain` directive cannot trigger a toolchain download and execution
  (see `specs/analyze-command/spec.md`, Requirement: Hermetic Toolchain); and
  adds an `--output <file>` (`-o`) flag that writes the `ModuleGraph` JSON to a
  file (stdout remains the default) so the agent can materialize base/PR graphs
  for `vibe-check diff` without shell redirection (Requirement: Output File
  Selection).

## Impact

- **New packages / files**: `internal/scaffold/` (embedded assets + writer);
  `metrics/delta.go` and `metrics/verdict.go` (structural delta + verdict
  engine); plus `cmd/vibe-check/init.go` and `cmd/vibe-check/diff.go` (new
  subcommands in the existing CLI package).
- **New embedded assets**: `internal/scaffold/assets/agents/divisor-entropy.md`.
- **Dependencies**: none new — uses stdlib `embed`, the existing `cobra`
  dependency, and the already-built `vibe-check` binary. The agent invokes
  `vibe-check analyze`, `vibe-check diff`, and `git` at review time; these are
  runtime prerequisites in the consuming environment, not Go build dependencies.
- **Build artifacts**: the `vibe-check` binary gains the `init` and `diff`
  subcommands and embeds the agent asset.
- **APIs**: new exported `scaffold` package surface (`Options`, `Result`, and a
  `Run` writer function); new `metrics` surface (`ComputeDelta`, `GraphDelta`,
  `DecideVerdict`, `Verdict`, `VerdictThresholds`); and `RunInit` / `RunDiff`
  entry points in `cmd/vibe-check`. The `ModuleGraph` schema is unchanged — the
  delta engine consumes existing `analyze` output.
- **CLI**: new `vibe-check init` command with `--force`, `--json`, and an
  optional `[path]` argument (default `.`); new `vibe-check diff` command taking
  two `ModuleGraph` JSON paths with `--json` and threshold-override flags; and a
  new `--output <file>` (`-o`) flag on the existing `vibe-check analyze` command.
- **Distribution**: shipped in the `vibe-check` binary; consumers run
  `vibe-check init` (analogous to `gaze init`).
- **Runtime prerequisites (agent)**: the deployed agent requires the `vibe-check`
  binary on `PATH` and `git` with worktree support in the consuming project.
- **Follow-up obligations** (issues to file or refresh before PR merge):
  - Website/docs: file an issue in `unbound-force/website` documenting the new
    `vibe-check init` and `vibe-check diff` commands and the `divisor-entropy`
    reviewer (user-facing).
  - Provenance: file the follow-up issue (tasks.md task 11.5) to add provenance
    metadata (producer, version, timestamp, compared SHAs) to the `diff --json`
    and `init --json` payloads (Constitution Principles I and III).

## Constitution Alignment

| Principle | Assessment |
|-----------|------------|
| I. Autonomous Collaboration | PARTIAL — the agent consumes `vibe-check analyze` JSON artifacts and emits a review verdict, and it does not block waiting on any other agent; however the verdict, `diff --json`, and `init --json` payloads omit provenance metadata (producer, version, timestamp, compared SHAs), so the artifacts are not yet fully self-describing. Same root cause as Principle III and addressed by the same provenance follow-up (see tasks.md task 11.5). |
| II. Composability First | PASS — `vibe-check init` and the agent are independently usable; the agent is auto-discovered when a Review Council is present but degrades gracefully (reports inability) when the `vibe-check` binary is absent. No hard dependency on a peer agent. |
| III. Observable Quality | PARTIAL — the verdict cites explicit base/PR refs and a metric-delta table, `vibe-check diff --json` emits a machine-readable verdict with reasons, and `init --json` emits a machine-readable summary; however all three payloads (`analyze`, `diff --json`, and `init --json`) omit provenance metadata (producer, version, timestamp, compared SHAs). The `analyze` gap is an existing follow-up from `go-analyze`; the `diff` and `init` payloads introduce the same gap and are tracked by a provenance follow-up issue filed in task 11.5 (adding producer/version/timestamp metadata). |
| IV. Testability | PASS — coverage strategy defined in design.md; the delta/verdict engine is table-tested at threshold boundaries; `scaffold`, `RunInit`, and `RunDiff` are unit-tested with `t.TempDir()` and `bytes.Buffer`; the agent is validated against improvement, degradation, and comment-band fixtures. |
| V. Security by Default | PARTIAL — least privilege is enforced (`edit: deny`, `webfetch: deny`, and a granular `bash` allowlist mirroring `divisor-curator` that denies all commands except `vibe-check analyze`/`diff` and narrow ref-resolution + `git worktree` subcommands); refs are format-validated and resolved with `git rev-parse --verify --end-of-options <ref>^{commit}` before use; path validation uses `metrics.ValidateProjectPath`; and file permissions are restrictive (0o644 files, 0o755 dirs). Residual risk: `vibe-check analyze` executes the analyzed module's build tooling, so analyzing an untrusted PR ref is a code-execution surface — mitigated by requiring `GOTOOLCHAIN=local` and documenting that the divisor must only run on refs the CI context already trusts (see design.md D2/D4 and the threat-model note). |
| VI. Metric Fidelity | PASS — deltas and the verdict are computed by a pure, table-tested Go function (`metrics.ComputeDelta` / `DecideVerdict`) over canonical `vibe-check analyze` output; deltas are exact arithmetic differences, cycle classification is exact, and output ordering is deterministic. No new metric formulas are introduced. |
| VII. Language Agnosticism | PASS — the delta operates on the universal `ModuleGraph`; the agent works for any language `vibe-check` supports, with no language-specific logic. |
