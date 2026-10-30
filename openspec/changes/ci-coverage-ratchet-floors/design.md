# Design

## Context

`.github/workflows/ci.yml` runs `go test -race -count=1
-coverprofile=coverage.out ./...` in its `test` job (line 37) and stops there:
the profile is written and never evaluated. `coverage.out` is gitignored
(`.gitignore:29`) as generated output. The module currently has five packages
with coverable statements; per-package statement coverage measured on `main`
(2026-10-30, same flags as CI):

| Package              | Measured | Functions | Min/Max function coverage |
|----------------------|----------|-----------|---------------------------|
| `cmd/vibe-check`     | 92.7%    | 31        | 0.0–100.0%                |
| `internal/goadapter` | 89.3%    | 37        | 38.5–100.0%               |
| `internal/scaffold`  | 87.9%    | 10        | 73.7–100.0%               |
| `metrics`            | 82.9%    | 54        | 0.0–100.0%                |
| `metrics/report`     | 99.4%    | 61        | 75.0–100.0%               |

All packages already sit above the project's established 80% bar, so a
ratchet anchored at today's real values costs nothing now and only prevents
downward drift. The `lint` job and the structural-gate workflows run no tests
and produce no profile, so the gate belongs solely in the `test` job.
Constitution Article IV requires coverage ratchets enforced by automation
that blocks the build; AGENTS.md lists coverage thresholds as protected
gatekeeping values.

The enforcement model is settled (issue #64, human decision): **per-package
ratchet** — monotonic per-package floors in a committed baseline, tighten-only.

## Goals / Non-Goals

**Goals:**

- Fail CI when any package's statement coverage drops below its recorded floor.
- Make floor registration explicit: a new package cannot ride ungated.
- Make loosening a floor impossible via tooling — only a reviewed file edit.
- Keep the gate deterministic, locally runnable (CI parity), and dependency-free.

**Non-Goals:**

- Per-function coverage floors (the 80%-per-function bar stays a task-time
  process rule enforced by review, not by CI).
- Raising coverage in any existing package (this change only locks in current
  values; upward movement happens naturally via `--update`).
- Changes to the `vibe-check` product CLI, the `lint` job, or the structural
  gate workflows.
- Coverage trend dashboards or PR annotations.

## Decisions

### D1: Per-package ratchet over the alternatives (settled)

Recorded per the issue-#64 human decision. A repo-wide aggregate floor was
rejected because a single number lets regressions in one package hide behind
slack in another (e.g. `metrics` falling while `metrics/report` absorbs it).
Per-function ratchets were rejected as too granular: function-level coverage
is unstable under routine refactors (renames, inlining, splitting), which
would make the gate noisy and invite loosening pressure. Per-package is the
grain at which the repo already states coverage targets (every prior design's
Coverage Strategy uses package-level percentages) and at which `go test`
itself reports coverage.

### D2: Gate tool is tested Go code, not shell or a marketplace action

New `cmd/coverage-ratchet` (thin `main` + flag wiring, mirroring how
`cmd/vibe-check` stays thin) over a new `internal/coverratchet` package
holding all logic: profile parsing, per-package aggregation, floor
comparison, tighten-only update, report rendering. Rationale:

- Testable per the repo's testing conventions (table-driven, `testdata/`
  fixtures) — a shell/awk pipeline is not, and AGENTS.md treats automated
  checks as non-negotiable gates.
- Deterministic output for identical input (project guardrail: metric
  computations must be deterministic).
- `golang.org/x/tools/cover.ParseProfiles` already ships in the module's
  existing direct dependency (`golang.org/x/tools` v0.49.0, used by
  `internal/goadapter`) — zero new dependencies.

Alternatives rejected: (a) awk over `go tool cover -func` output — per-function
percentiles cannot be aggregated to true statement coverage without statement
counts, and the textual format is not a stable contract; (b) a marketplace
coverage action — an external, SHA-pinned dependency that does not implement
tighten-only semantics, violating the zero-waste mandate; (c) a `vibe-check`
subcommand — would grow the product CLI (architectural metrics for consumers)
with repo-internal CI bookkeeping.

### D3: Baseline file — `coverage-floors.json` at the repository root

Format (committed, human-reviewable, minimal):

```json
{
  "version": 1,
  "floors": {
    "cmd/vibe-check": 92.7,
    "internal/goadapter": 89.3,
    "internal/scaffold": 87.9,
    "metrics": 82.9,
    "metrics/report": 99.4
  }
}
```

Keys are module-relative package paths (the profile stores filenames as
`github.com/zero-dot-force/vibe-check/<pkg>/<file>.go`; the tool strips the
module prefix and the filename). The `version` field allows future format
evolution without a breaking re-read. Serialization is deterministic:
lexicographically sorted keys, two-space indent, trailing newline, so
unchanged updates produce empty diffs. Root location (next to where the
gitignored `coverage.out` lands) rather than `.github/` because CI parity
requires developers to run the same gate locally — the file is not CI-only
configuration. JSON over YAML/TOML because the repo's machine-artifact
convention is JSON with schemas (`metrics/modulegraph.schema.json`).

### D4: Measurement semantics — statement-weighted, truncated to one decimal

Per-package coverage is computed from the profile blocks:
`covered = Σ NumStmt where Count > 0`, `total = Σ NumStmt`, percentage =
`covered/total × 100`. With `-race` the profile is `mode: atomic`, where any
non-zero count means executed. Percentages are truncated (rounded down) to
one decimal for both comparison and stored floors, and measured == floor
passes. Truncation is applied identically on both sides, so a floor recorded
by an update can never exceed what the same run measured — no epsilon, no
float flapping. Alternatives rejected: parsing `go test`'s per-package stdout
(not a machine-stable contract, and interleaved with build output); rounding
to nearest (could store a floor 0.05pp above measured, failing the very run
that produced it).

### D5: Tighten-only update semantics — floors are protected values

`--update` rewrites the baseline with exactly three effects: raise an
existing floor to a higher truncated measured value; add an unregistered
package at its truncated measured value **iff** that value is ≥ 80.0 (the
project's established bar for new code, per prior Coverage Strategies);
leave everything else untouched. There is no force flag, and no invocation
path lowers a floor — if measured < floor, the entry is unchanged and the
regression is reported (exit 1). Loosening a floor requires a direct,
human-reviewed edit of `coverage-floors.json` — this is the AGENTS.md
gatekeeping rule for coverage thresholds made structural: the ratchet's
normal operation cannot weaken the gate. The 80% entry bar for new packages
stems from the same rule set: letting a new package register below the bar
would be a silent gate weakening relative to the process every prior change
followed. Alternative rejected: registering new packages at any measured
value (would formally allow a 5% package to pass CI forever).

### D6: CI wiring — one step in the existing `test` job

Immediately after the Test step:

```yaml
      - name: Coverage ratchet
        run: go run ./cmd/coverage-ratchet --profile coverage.out --floors coverage-floors.json
```

No new actions, no new permissions, no new job. `go run` reuses the warm
module cache from the Test step; a separate job would re-run checkout and
setup-go only to re-generate the same profile. The structural-gate workflows
are untouched (they never produce a profile). The exact command is what
developers run locally (CI parity): generate `coverage.out`, run the gate.

### D7: Reporting contract — table + `--json` with provenance

Default output: one line per package (`measured floor status`), sorted by
package path, then a summary line. `--json` emits the same data as a stable
JSON object plus a provenance envelope (producer `coverage-ratchet`,
version, RFC 3339 UTC `generatedAt`) mirroring the shape of
`cmd/vibe-check`'s provenance object (Constitution III: JSON output and
provenance on produced artifacts). Exit codes follow the project CLI
convention (`cmd/vibe-check/analyze.go:71`): 0 success, 1 policy failure
(floor or registration violation), 2 tool failure (missing/unparseable
profile or baseline). Version is embedded via ldflags with a `dev` default
when built without them (same pattern as `cmd/vibe-check/main.go`), so the
JSON mode is constitution-conformant even under `go run`.

### D8: Self-hosting — the gate gates itself

`cmd/coverage-ratchet` and `internal/coverratchet` are new packages with
coverable statements, so by this change's own completeness rule they must
register in the baseline at ≥ 80% before the CI step goes live. The
implementation order in tasks.md handles this: the tool and its tests land
first, then `--update` generates the baseline including the tool's own
packages, then the workflow step is added. Tests: table-driven unit tests
for aggregation, comparison, truncation, tighten-only update, registration
completeness, and the 80% entry bar, over committed `testdata/` profiles
(hand-written minimal profiles, not generated output); a determinism test
(re-run, byte-identical report); a JSON-shape test (fixed fields, provenance
present, timestamp excluded from stability comparison). Target: ≥ 80%
coverage for both new packages, per project convention.

## Risks / Trade-offs

- [Run-to-run coverage variation from nondeterministic test branches
  (timeouts, parallelism) could flap a borderline package] → Floors are
  recorded at the truncated measured value (up to 0.1pp headroom); if a
  branch is genuinely flaky in coverage terms, the correct fix is
  stabilizing the test, not loosening the floor — the gate makes that
  visible instead of hiding it.
- [Legitimate refactors may drop coverage in hard-to-test error paths] →
  The floor is today's measured value, not an aspiration; the delta a
  refactor must recover is exactly what it dropped. Restoration happens in
  the same PR, which is the behavior the constitution mandates.
- [Bootstrap failure: the new packages measure < 80% at first run] → Tests
  are written alongside the logic (D8 ordering); `--update` refuses to
  register a sub-80 package, so the gap surfaces in task 1, not in CI.
- [Baseline editing by hand can drift (typos, wrong precision)] → The gate
  parses and validates the file every run; a malformed baseline is a tool
  failure (exit 2), and the update mode is the supported edit path.
- [`go run` compiles on every CI run] → Negligible with the warm module
  cache; avoids adding a build step or committing a binary.

## Migration Plan

Single PR, ordered so the repo is never red: (1) land
`internal/coverratchet` + `cmd/coverage-ratchet` with tests; (2) generate
and commit `coverage-floors.json` via `--update` from a clean local run;
(3) add the CI step. Rollback is `git revert` of the PR — the baseline file
is inert data without the workflow step, and no existing job changes
behavior.
