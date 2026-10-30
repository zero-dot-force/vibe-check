# Proposal

## Why

The CI parity gate runs `go test -race -count=1 -coverprofile=coverage.out ./...`
(`.github/workflows/ci.yml`, Test step) but never evaluates the profile against
any threshold — CI goes green at any coverage level, so the per-change "≥80%
coverage" bar that every prior design's Coverage Strategy relies on is enforced
only at task-completion time and nothing prevents regressions once tasks are
checked off. Constitution Article IV (Testability) mandates the opposite:
"Coverage ratchets MUST be enforced by automated tests; any coverage regression
MUST be treated as a test failure and block the build." Found during the
`init-registers-divisor-entropy` review (issue #64); three prior changes
(`go-analyze`, `add-diff-init-provenance`, `ci-regression-gate`) explicitly
deferred ratchet enforcement as a follow-up.

## What Changes

- Introduce a **per-package coverage ratchet** (settled decision, issue #64):
  monotonic per-package statement-coverage floors recorded in a committed
  baseline file `coverage-floors.json` at the repository root.
- Add a deterministic Go gate tool — `cmd/coverage-ratchet` (thin CLI) over a
  new `internal/coverratchet` package — that parses `coverage.out`
  (`golang.org/x/tools/cover`, already a direct dependency), computes
  per-package statement coverage, and fails (exit 1) when any package falls
  below its floor or is missing from the baseline.
- Enforce **tighten-only semantics in code**: the tool's `--update` mode only
  raises existing floors and only adds new packages at or above the project's
  80% bar; no tool path can lower a floor. Loosening a floor requires an
  explicit, human-reviewed edit to the baseline file (gatekeeping rule:
  coverage thresholds are protected values).
- Add a CI step to the `test` job in `.github/workflows/ci.yml`, immediately
  after the Test step, running the same ratchet command used locally
  (CI parity): `go run ./cmd/coverage-ratchet --profile coverage.out
  --floors coverage-floors.json`.
- Commit the initial baseline measured from the current `main`:
  `cmd/vibe-check` 92.7%, `internal/goadapter` 89.3%, `internal/scaffold`
  87.9%, `metrics` 82.9%, `metrics/report` 99.4%.

No breaking changes. No new Go module dependencies. No changes to existing
production packages (`metrics/`, `internal/goadapter/`, `internal/scaffold/`,
`cmd/vibe-check/`) or to the `vibe-check` CLI surface.

## Capabilities

### New Capabilities
- `ci-coverage-ratchet`: Per-package coverage-floor enforcement in CI — the
  committed baseline-file contract, per-package evaluation of the collected
  coverage profile, and tighten-only ratchet semantics (including failure on
  unregistered packages and the 80% entry bar for newly registered packages).

### Modified Capabilities

None. The project has no archived main specs yet (`openspec/specs/` is empty);
prior CI behavior (including the coverage-collecting Test step) was never
specified, so this capability is introduced fresh rather than modified.

## Impact

- **CI**: `.github/workflows/ci.yml` `test` job gains one step after Test.
  The `lint` job and the structural-gate workflows
  (`structural-gate.yml`, `reusable_structural_gate.yml`) are untouched —
  neither runs `go test` or produces a coverage profile.
- **New files**: `coverage-floors.json` (root, committed),
  `cmd/coverage-ratchet/` (entry point + ldflags version embedding),
  `internal/coverratchet/` (profile parsing, aggregation, floor comparison,
  tighten-only update) with unit tests and `testdata/` profile fixtures.
- **Dependencies**: none added — `golang.org/x/tools` (which provides
  `golang.org/x/tools/cover`) is already a direct dependency of
  `internal/goadapter`.
- **Docs**: `CHANGELOG.md` entry; `AGENTS.md` project-structure section gains
  the new `cmd/coverage-ratchet/` and `internal/coverratchet/` entries.
  CI-only change — exempt from the user-facing documentation-issue gate.
- **Process**: from merge onward, a PR that drops any package below its
  recorded floor fails CI until coverage is restored or the floor is raised
  (never lowered) via an explicit reviewed baseline change.
