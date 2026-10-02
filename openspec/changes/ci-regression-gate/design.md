## Context

`vibe-check diff <base.json> <pr.json>` already implements the full #8 delta
engine: `metrics.ComputeDelta` (pure, deterministic per-package delta plus
added/removed/new-cycle classification) and `metrics.DecideVerdict` (protected
threshold gates — new cycle, ΔI ≥ 0.15, ΔD ≥ 0.20, ΔLCOM ≥ 2 → `REQUEST_CHANGES`;
sub-gate worsening → `COMMENT`; else `APPROVE`). New-cycle detection already
satisfies #8's `--no-new-circular-deps` acceptance criterion. The verdict and
its reasons are emitted in both human-readable and `--json` output.

The only gap is gate enforcement: `RunDiff` returns `ExitCode: 0` for every
valid-input case (the verdict travels in the payload, not the exit code), and
there is no CI workflow that invokes `diff` as a blocking gate. The `analyze`
command already establishes the `0`/`1`/`2` exit taxonomy (`0` success, `1`
policy failure, `2` tool failure), which `--gate` reuses.

The issue's literal proposal — `analyze --regression-base=origin/main` with
`--max-instability-delta`, `--max-distance-delta`, `--no-new-circular-deps` —
would duplicate the existing delta/verdict engine inside `analyze`, violating
DRY and the zero-waste mandate. This design re-scopes #8 to a `--gate` flag on
`diff` plus CI wiring.

## Goals / Non-Goals

**Goals:**

- Add a `--gate` flag to `vibe-check diff` that encodes a blocking
  `REQUEST_CHANGES` verdict as exit code `1`, keeping `APPROVE`/`COMMENT` at
  exit `0` and tool failures at exit `2`.
- Preserve default `diff` behavior exactly (opt-in gate; valid inputs exit `0`).
- Add a pull_request CI workflow that builds the binary, snapshots the head and
  base, and runs `diff --gate`, failing on `REQUEST_CHANGES`.
- Reuse `ComputeDelta`/`DecideVerdict` and their protected thresholds unchanged.

**Non-Goals:**

- `analyze --regression-base` git-ref orchestration (rejected as duplication).
- Push-to-main self-gating or scheduled drift-tracking jobs (pull_request only).
- A `--no-new-circular-deps` flag on `diff` (already subsumed by `DecideVerdict`
  new-cycle gate).
- New metric computation, schema changes, or adapter-interface changes.
- Documentation gap #36 (`--no-provenance`/provenance for diff/init) — separate.
- Running the CI gate on the `push` event or as a required status check
  configuration (that is repo-admin enablement, out of scope for this change).

## Decisions

### D1: Re-scope to `diff --gate`, not `analyze --regression-base`

**Decision**: Implement gate mode as a `--gate` boolean flag on the existing
`diff` command; wire CI to run `diff --gate`. Do not add `--regression-base` or
delta-threshold flags to `analyze`.

**Rationale**: The delta/verdict engine already lives in `diff`
(`ComputeDelta` + `DecideVerdict`). Reimplementing it in `analyze` would create
two divergent gate paths and violate DRY and the zero-waste mandate. The
`--gate` flag is a minimal, additive surface that exposes the existing verdict
as an exit code.

**Alternatives considered**:
- `analyze --regression-base` (literal issue proposal) — duplicates the engine;
  rejected.
- A new top-level `vibe-check gate` command — a fourth command for a flag-sized
  concern; rejected.

### D2: Gate fails only on REQUEST_CHANGES (COMMENT stays advisory)

**Decision**: Under `--gate`, only a `REQUEST_CHANGES` verdict exits `1`.
`APPROVE` and `COMMENT` exit `0`.

**Rationale**: `COMMENT` is a sub-gate advisory (a non-zero-but-below-threshold
shift). Failing CI on every `COMMENT` would be too strict for day-to-day work
and is inconsistent with the verdict's advisory semantics. `REQUEST_CHANGES` is
the verdict that signals structural degradation past protected thresholds.

**Alternatives considered**:
- Fail on `COMMENT` as well — over-broad; makes the gate unusably strict;
  rejected.

### D3: Exit taxonomy mirrors analyze (0/1/2)

**Decision**: Under `--gate`, valid inputs exit `0` (APPROVE/COMMENT) or `1`
(REQUEST_CHANGES); tool failures (missing/unreadable/schema-invalid input,
looser-than-default override) remain `2`, unchanged.

**Rationale**: `analyze` already documents and tests the `0`/`1`/`2` taxonomy.
Reusing it gives operators a consistent mental model: `1` = policy failure, `2`
= tool failure. `diff`'s existing exit-2 cases are preserved exactly.

**Alternatives considered**:
- Exit `2` for `REQUEST_CHANGES` — conflates a legitimate policy verdict with a
  tool failure; rejected.

### D4: `--gate` is opt-in; default behavior unchanged

**Decision**: `--gate` defaults to `false`. `DiffOptions.Gate` is a new field;
`RunDiff` only changes `ExitCode` when `opts.Gate && verdict ==
metrics.VerdictRequestChanges`.

**Rationale**: The `divisor-entropy` agent and any existing consumers of `diff`
depend on the current "exit 0 for valid inputs" contract (documented in the
`diff-command` spec and README). An opt-in flag avoids a breaking change to that
contract while enabling the CI gate.

**Alternatives considered**:
- Make gate mode the default — breaking change to existing consumers; rejected.

### D5: Separate workflow file `structural-gate.yml`, pull_request only

**Decision**: Add `.github/workflows/structural-gate.yml` triggered on
`pull_request` into `main`, separate from the existing `ci.yml`. It uses
`permissions: contents: read`, a concurrency group, and pinned actions.

**Rationale**: A separate file keeps the structural gate independently
reviewable, runnable, and disable-able without touching build/test/lint. The
`pull_request` trigger (not `push`) matches the decision that the gate compares
PR head vs. base; there is no base to compare on a bare `push` to `main`.

**Alternatives considered**:
- Fold into `ci.yml` — couples gate availability to the main CI job; harder to
  reason about; rejected.
- Trigger on `push` — no base/PR delta; rejected.

### D6: Base snapshot via second checkout of the base SHA

**Decision**: The workflow relies on `actions/checkout`'s default
`pull_request` checkout (the PR merge commit) as the head snapshot, then
performs a second `actions/checkout` of
`${{ github.event.pull_request.base.sha }}` into `./base`, builds the binary
once, and runs `analyze --output head.json .` and
`analyze --output base.json ./base` before `diff --gate base.json head.json`.

**Rationale**: `diff` needs two `ModuleGraph` snapshots produced by the same
binary version. Checking out the base SHA into a sibling directory and analyzing
both with the freshly built binary guarantees a same-version comparison without
installing a released binary.

**Alternatives considered**:
- `diff` against a stored/pinned baseline artifact — adds a release/distribution
  dependency; rejected for a first iteration.
- Use `actions/checkout` with `ref` + `path` for both checkouts — equivalent;
  the two-checkout form is chosen for clarity.

### D7: Fork guard — skip the gate for forked PRs

**Decision**: The workflow SHALL skip the structural gate when the PR head is
from a fork (`github.event.pull_request.head.repo.full_name !=
github.repository`); forked PRs are not analyzed, so no `analyze`/`diff` step
executes for them.

**Rationale**: `vibe-check analyze` drives the Go toolchain to type-check and
compile the target module, which can execute build-time code (cgo, generated
files, custom build steps) from the analyzed repository. Analyzing a forked PR
would therefore execute untrusted code on the CI runner, contradicting the
in-repo `divisor-entropy` agent's constraint that `analyze` "MUST NOT be wired
into CI that analyzes untrusted fork pull requests." Skipping forks keeps the
Constitution V (Security by Default) assessment honest while preserving full
gate coverage for internal-branch PRs (the common development path).

**Alternatives considered**:
- Analyze forks and document parity with `ci.yml` (`go test -race` also
  executes fork code) — technically equivalent exposure under
  `pull_request`'s read-only, secret-less token, but leaves the governance
  contradiction with the `divisor-entropy` agent unaddressed; rejected.

## Risks / Trade-offs

- **[R1] Base-path analysis mechanics** (`analyze ./base/...` against a
  partial/foreign checkout) may surface edge cases (module path resolution,
  partial-build `Warnings`) → Mitigation: the CI gate relies on `diff`'s existing
  partial-build handling (unreliable input downgrades to `COMMENT`, which exits
  `0`), so a flaky checkout degrades to advisory rather than falsely failing;
  validated with a local dry-run during implementation.
- **[R2] Fork/`pull_request` `base.sha` availability** — `pull_request` events
  expose `base.sha` for branch and fork PRs; `base.ref` may be absent on some
  fork events → Mitigation: use `base.sha` (always present on `pull_request`),
  not `base.ref`.
- **[R3] CI gate is advisory until marked required** — the workflow fails the
  check, but repo admins must enable branch protection for it to block merge →
  Mitigation: documented as a follow-up enablement step in tasks.md.
- **[R4] `COMMENT` advisory verdicts do not block CI** — a PR could drift
  through multiple sub-threshold regressions without a single `REQUEST_CHANGES`
  → Mitigation: accepted trade-off of D2; cumulative drift is the entropy
  divisor agent's concern, not this gate's.
- **[R5] Forked PRs skip the gate** — a fork could carry structural regression
  undetected until merged by a maintainer → Mitigation: accepted; forks are the
  minority case, internal branches are fully gated, and a maintainer can re-run
  the gate locally or after opening an internal branch.

## Coverage Strategy

**Unit tests** (`cmd/vibe-check/`): extend the diff CLI tests to cover
`--gate` exit-code behavior — `RunDiff` with `Gate: true` maps
`REQUEST_CHANGES`→`1`, `APPROVE`→`0`, `COMMENT`→`0`; `Gate: false` always `0`;
and the `diffCmd` flag wiring with `--gate` producing `*exitCodeError{code: 1}`
on a REQUEST_CHANGES fixture. Use the testable CLI pattern (`bytes.Buffer`, no
subprocess). Existing exit-2 and default exit-0 tests remain unchanged and
continue to pass, guarding D4.

**CI workflow validation**: the workflow is validated by a local dry-run
(`go build` + two `analyze` snapshots + `diff --gate`) and, post-merge, by the
workflow's own `pull_request` run. The workflow file has no Go unit coverage.

**Coverage target**: ≥ 80% line coverage for the new/modified `diff.go` branch,
consistent with the project's existing `cmd/vibe-check/` tests. No coverage
ratchet is enforced (project-wide follow-up).
