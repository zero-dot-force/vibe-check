## Context

Vibe-Check ships a `vibe-check analyze` command that emits a deterministic,
schema-valid `metrics.ModuleGraph` (Ca, Ce, Instability, Abstractness, Distance
from main sequence, LCOM4, circular dependencies) as JSON. The Unbound Force
Review Council auto-discovers reviewer agents by scanning `.opencode/agents/`
for files named `divisor-*.md` and delegates to each in parallel; the council's
review scope is the full branch diff (`git diff main...HEAD`).

Today none of the nine existing divisors measures whether a PR *degrades the
structure* of the codebase. This change adds a `divisor-entropy` reviewer that
compares `vibe-check analyze` output for the base branch against the PR branch
and enforces the Boy Scout Rule, plus a `vibe-check init` command that deploys
the agent (and any future embedded agent assets) into a consuming project.

Relevant existing conventions:

- The only `//go:embed` in the repo is `metrics/schema.go`
  (`//go:embed modulegraph.schema.json` → `var schemaJSON []byte`, exposed via a
  defensive-copy accessor). The scaffold package reuses this pattern.
- `cmd/vibe-check/analyze.go` follows the testable-CLI pattern AP-002/AP-003:
  a `RunAnalyze(ctx, opts)` function holds all logic, cobra's `RunE` only wires
  flags and `io.Writer`s. `init` and `diff` mirror this with `RunInit` and `RunDiff`.
- Existing divisors vary in their permission profile. `.opencode/agents/divisor-curator.md`
  is a precedent that already enables `bash` via a granular `"*": "deny"` + allowlist
  pattern (using `temperature: 0.2` and an `ask` tier for issue-creation commands).
  Divisors like `divisor-architect` use `bash: deny` and `temperature: 0.1`.

## Goals / Non-Goals

**Goals:**

- Add an `internal/scaffold/` package that embeds agent assets via `//go:embed`
  and writes them into a target project's `.opencode/agents/` directory with
  safe permissions and path validation.
- Add the embedded asset `internal/scaffold/assets/agents/divisor-entropy.md` — a Review
  Council divisor that computes the base↔PR structural delta by running
  `vibe-check analyze` on both refs, then running `vibe-check diff` to compute the
  metric deltas and verdict deterministically in tested Go code, and renders an
  APPROVE / REQUEST CHANGES / COMMENT verdict backed by an auditable metric-delta table.
- Add a `vibe-check init [path]` CLI command (idempotent, `--force`, `--json`).
- Add a `vibe-check diff <base.json> <pr.json>` CLI command exposing `metrics.ComputeDelta`
  and `metrics.DecideVerdict` — delta and verdict are computed by a pure, deterministic,
  table-tested Go function, not by prompt arithmetic.
- Dogfood the agent into vibe-check's own `.opencode/agents/` by generating it
  with the init command from the embedded single source of truth.
- Define deterministic, documented verdict thresholds implemented and enforced by
  the tested Go `metrics.DecideVerdict` function.

**Non-Goals:**

- A general-purpose plugin/asset-manager for `vibe-check init` beyond agent
  assets. This change scaffolds only `.opencode/agents/` content.
- Breaking changes to existing `metrics` APIs or any change to the `ModuleGraph`
  schema. (This change ADDS additive metrics surface — `ComputeDelta`, `GraphDelta`,
  `DecideVerdict`, `Verdict`, `VerdictThresholds` — but modifies no existing
  signature and leaves the schema unchanged.)
- Provenance metadata in `analyze` output (an existing `go-analyze` follow-up).
- CI enforcement of the entropy verdict as a hard gate in vibe-check's own CI
  (the agent runs inside the Review Council, not the `.github/workflows`).

## Decisions

### D1: Delivery via embedded assets + `vibe-check init`

**Decision**: An `internal/scaffold` package embeds `assets/agents/*.md` via
`//go:embed` and writes them to `<target>/.opencode/agents/`. The new
`vibe-check init [path]` command invokes it. The embedded asset is the single
source of truth for the agent content.

**Rationale**: Single-binary distribution (analogous to `gaze init`); no external
fetch (Constitution V supply chain); reuses the established `metrics/schema.go`
embed pattern; keeps the agent versioned with the binary that runs `analyze`.

**Alternatives considered**:
- Ship a separate template repo consumers clone — more moving parts, drifts from
  the binary version.
- `curl | sh` bootstrap — supply-chain risk, violates Security by Default.
- `go generate` copy — not distributable to downstream (non-Go) consumers.

### D2: Delta and verdict are computed by a pure Go function exposed via `vibe-check diff`

**Decision**: The agent does NOT compute metric arithmetic in-prompt. It runs
`vibe-check analyze` on the base and PR refs to produce two `ModuleGraph` JSON
documents, then runs `vibe-check diff <base.json> <pr.json>`, which computes
per-package deltas, classifies cycles as new/pre-existing/resolved, and applies
the verdict thresholds — all in a pure, deterministic, table-tested Go function
(`metrics.ComputeDelta` + `metrics.DecideVerdict`). The agent reports and explains
the Go-computed verdict.

**Rationale**: Satisfies Constitution III (scoring thresholds backed by re-runnable
automated tests) and VI (deterministic, exact arithmetic and cycle classification);
keeps the threshold gates in protected, tested code rather than fallible prompt
arithmetic; narrows the agent to orchestration + narrative.

**Alternatives considered**: (a) Agent computes deltas in-prompt — rejected:
non-deterministic, untestable, thresholds unprotected. (b) `analyze --baseline`
single-invocation — rejected: couples analyze to git-ref management, whereas a
standalone `diff` over two graphs is simpler and independently testable.

**Note**: This supersedes the earlier plan (agent-computed; `diff` deferred to P2).

**Trust boundary / Threat model**: `vibe-check analyze` loads the target with
`go/packages` type-checking, which executes the module's build tooling (compilation,
cgo, and — absent `GOTOOLCHAIN=local` — a `go.mod` `toolchain` directive can trigger
a toolchain download+execution from GOPROXY). Running the entropy divisor therefore
executes code from the refs it analyzes. Both `vibe-check analyze` invocations
MUST run with `GOTOOLCHAIN=local` in the subprocess environment. This value is
forced by the `vibe-check analyze` binary itself (not by the agent — the agent's
bash allowlist permits only `vibe-check analyze *`, not
`GOTOOLCHAIN=local vibe-check analyze *`, so the env-var cannot be set inline by the
agent). Privileged CI SHOULD also export `GOTOOLCHAIN=local` ambiently as an
additional defense-in-depth layer. The divisor MUST only run on refs the CI context
already trusts (same-repo PRs or already-built branches); it MUST NOT run against
untrusted fork PRs in privileged CI. A prior `go-analyze` review accepted this
code-execution surface as LOW only under the "trusted self-analysis" assumption,
now made explicit.

The shipped `divisor-entropy.md` asset MUST carry a `## Security / Operating
Constraints` section documenting this residual cgo/build-execution surface —
which `GOTOOLCHAIN=local` does NOT close (it closes only the `go.mod`
toolchain-directive download+execution vector) — and the trusted-refs-only /
no-untrusted-fork-PR-CI operating constraint, so downstream adopters see the
warning in the deployed agent.

### D3: Base/PR isolation via `git worktree`

**Decision**: The agent creates a temporary `git worktree` checked out at the
base ref, runs `vibe-check analyze` there, runs `analyze` on the current PR
checkout, then removes the worktree. The base ref is the merge-base of `HEAD` and
the default branch (`git merge-base HEAD origin/main` by default), matching the
council's three-dot `git diff main...HEAD` scope; it is overridable.

**Rationale**: A worktree avoids mutating the working tree or index (no
stash/checkout dance), is safe under concurrent CI, and does not disturb the PR
checkout. Comparing against the merge-base isolates the PR's own contribution
from unrelated base movement.

**Alternatives considered**:
- `git stash` + `checkout base` + analyze + `checkout -` — mutates the working
  tree, unsafe on a dirty tree, not concurrency-safe.
- Fresh `git clone` of the base — slower and more disk than a worktree.

### D4: Agent permissions — a granular `bash` allowlist mirroring `divisor-curator`, everything else `deny`

**Decision**: Frontmatter `mode: subagent`, `temperature: 0.1`, `edit: deny`,
`webfetch: deny`, and a granular `bash` allowlist that denies every command by
default and permits only the specific commands the agent needs:

```yaml
permission:
  edit: deny
  webfetch: deny
  bash:
    "*": "deny"
    "git merge-base *": "allow"
    "git rev-parse *": "allow"
    "git worktree add *": "allow"
    "git worktree remove *": "allow"
    "git worktree prune": "allow"
    "git fetch origin *": "allow"
    "git check-ref-format *": "allow"
    "vibe-check analyze *": "allow"
    "vibe-check diff *": "allow"
```

**Precedence**: OpenCode evaluates `permission.bash` patterns against the parsed
command and applies the LAST matching rule, so the leading `"*": "deny"` sets
deny-by-default and the specific entries grant only the needed commands. A trailing
`*` permits arguments. A metacharacter-bearing or compound command (containing `;`,
`&&`, `||`, `|`, `$(...)`, backticks, or redirection) is denied by the catch-all as
a defense-in-depth measure. OpenCode's last-matching-rule precedence means the
`"*": "deny"` catch-all applies when no more-specific allow entry matches; however,
compound-command denial by the permission matcher is assumed, not guaranteed. The
primary metacharacter defense is the ref-sanitization regex `^[A-Za-z0-9._/-]+$`
combined with `git rev-parse --verify --end-of-options "<ref>^{commit}"`, which is
enforced on the one untrusted interpolated value (the base ref) before it reaches any
command.

**Precedent**: This is the SECOND divisor to enable `bash`, mirroring the
`"*": "deny"` + allowlist shape already established by
`.opencode/agents/divisor-curator.md` (which additionally uses an `ask` tier
for issue-creation commands and `temperature: 0.2`).

**Ref safety (Constitution V)**: The base ref is overridable, so before any ref
reaches a command the agent MUST reject any ref not matching `^[A-Za-z0-9._/-]+$`
(additionally validate with `git check-ref-format` as a complementary check — the regex and `check-ref-format` are not equivalent controls; the regex is the primary metacharacter defense) and resolve it via
`git rev-parse --verify --end-of-options "<ref>^{commit}"`, using only the
resulting SHA thereafter. This makes "resolve refs to SHAs; never interpolate
untrusted branch strings" enforceable rather than self-contradictory.

**Gatekeeping**: Setting a new agent's permissions is a design choice, not
modification of an existing gate; once ratified these permissions and the D5
thresholds are protected gates. `temperature: 0.1` matches `divisor-architect`
(note `divisor-curator` uses `0.2`) — this does not claim to match "the other
divisors" generally.

### D5: Deterministic verdict thresholds

**Decision**: The agent applies explicit rules to the computed deltas. Defaults:

- **REQUEST CHANGES** if ANY of:
  - a new circular dependency is introduced (a cycle present in the PR graph that
    is absent from the base graph) — any new cycle;
  - any existing package's Instability increases by ≥ 0.15 (ΔI ≥ 0.15);
  - any package's Distance from the main sequence increases by ≥ 0.20 (ΔD ≥ 0.20);
  - any package's LCOM increases by ≥ 2 (ΔLCOM ≥ 2, a new disconnected
    responsibility split).
- **COMMENT** if metrics shift within materiality bands (0 < ΔI < 0.15, 0 < ΔD <
  0.20, ΔLCOM == 1) — non-blocking, worth discussion.
- **APPROVE** if metrics improve or stay within noise (no material degradation).

Added and removed packages are reported for informational purposes ONLY and MUST NOT
trigger any threshold gate. The absolute quality of new packages is already enforced
by `vibe-check analyze --max-distance` and related flags; the entropy divisor is
strictly a delta/regression tool. This removes the previous "new package in the zone
of pain/uselessness" gate, which was absent from diff-command/spec.md's verdict list
and violated the zero-waste principle. The `GraphDelta` payload carries an
`EntropyDirection` field: `degrading` when any REQUEST_CHANGES trigger fired;
`improving` when at least one metric delta is negative AND no metric delta is
positive above the 1e-4 rounding noise AND no trigger fired; `stable` otherwise
(no material change, or offsetting improvements and regressions that fire no
gate). Both `improving` and `stable` map to APPROVE, so the field is
reporting-only; diff-command/spec.md documents the identical rule.
All metric deltas include ΔAbstractness alongside ΔCa, ΔCe, ΔInstability, ΔDistance,
and ΔLCOM. Verdict thresholds are policy living in the Layer-1 `metrics/` package,
which is acceptable and documented.

The verdict function uses inclusive `≥` (a regression of exactly the threshold is
already too much) whereas `vibe-check analyze --max-*` flags use strict `>`. Float
deltas are rounded to 4 decimal places before comparison.

These thresholds are implemented and enforced by the tested Go `metrics.DecideVerdict`
function (with boundary tests covering each gate at and just below the threshold),
not by prompt arithmetic.

**Rationale**: Explicit thresholds make the verdict reproducible and defensible
and encode the Boy Scout Rule. New cycles are treated as unambiguous degradation
because `vibe-check` cycle detection is exact (Constitution VI).

**Gatekeeping note**: These thresholds ARE quality gates ("metric thresholds —
instability, distance" are protected values). Once ratified in the spec, they
MUST NOT be weakened without a spec change. Consuming projects MAY tighten them
but SHOULD NOT loosen them without recorded justification.

### D6: `vibe-check init` command semantics

**Decision**:
- `vibe-check init [path]` (default path `.`).
- Creates `<path>/.opencode/agents/` (dir mode `0o755`) when missing; writes each
  embedded asset with file mode `0o644`.
- **Overwrite policy**: by default SKIP files that already exist and report them;
  `--force` overwrites. Skipping is not an error.
- `--json` emits a machine-readable summary
  (`{"written":[...],"skipped":[...],"forced":[...]}`); otherwise a human summary.
- Path validation via `metrics.ValidateProjectPath` rejects traversal/unsafe
  paths.
- Exit codes: `0` success (including skips), `2` failure (invalid path or I/O
  error). No policy exit code is needed for `init`.

**Rationale**: Default-skip protects user customizations (Constitution V least
surprise); `--json` satisfies Observable Quality; exit-code semantics mirror
`analyze` (0 ok / 2 tool failure).

### D7: `internal/scaffold` package layout

**Decision**:
- `internal/scaffold/embed.go` — `//go:embed assets/agents/*.md` →
  `var assetsFS embed.FS`.
- `internal/scaffold/scaffold.go` — `Options{TargetDir string; Force bool}`,
  `Result{Written, Skipped, Forced []string}`, and `Run(opts Options) (*Result, error)`
  (named `Run`, not `Scaffold`, to avoid the package/function stutter and follow the
  AP-002 canonical entry-point name) that walks the embedded FS, computes each
  destination path, honors `Force`, and applies file/dir modes.
- `internal/scaffold/assets/agents/divisor-entropy.md` — the embedded agent.
- `internal/scaffold/doc.go` — package GoDoc.

**Rationale**: `internal/` prevents external import (scaffolding is a CLI
implementation detail). Separating embed, writer, and assets keeps each file
focused, mirroring the `internal/goadapter/` layout.

### D8: Dogfooding into vibe-check's own repository

**Decision**: Generate `.opencode/agents/divisor-entropy.md` in vibe-check itself
by running `vibe-check init .` (from the embedded single source of truth) rather
than hand-authoring a second copy.

**Rationale**: Dogfooding validates the agent and lets vibe-check's own Review
Council run entropy analysis on vibe-check PRs. Generating (not hand-maintaining)
prevents drift between the embedded asset and the deployed copy.

**Follow-on**: A CI check MAY regenerate via `vibe-check init . --force` (or
delete-then-init) and `git diff --exit-code` to prove the committed copy matches
the embedded asset. This approach (using `--force`) is required to detect drift
because a plain `init` skips the existing file (see Open Questions).

The AP-006 provenance marker is a stable, version-less string
(`<!-- scaffolded by vibe-check -->`), so a regenerated asset is byte-identical to
the committed copy and the drift check is deterministic across binary versions —
the marker carries no version to normalize, and `scaffold.Options` has no
`Version` field and performs no version substitution.

### D9: Agent output contract

**Decision**: In Code Review Mode the agent MUST output:
- the base ref (merge-base SHA) and PR ref (HEAD SHA) compared;
- a per-package delta table (Ca, Ce, Instability, Abstractness, Distance, LCOM: base → PR, Δ),
  sourced from the `vibe-check diff` JSON output;
- the list of newly introduced cycles (if any), sourced from the `vibe-check diff`
  JSON output;
- an overall entropy direction (improving / stable / degrading), sourced from the
  `vibe-check diff` JSON output;
- findings in the standard divisor block
  (`### [SEVERITY] Title`, **File**, **Description**, **Recommendation**;
  severity per `.opencode/uf/packs/severity.md`);
- a domain Score (1–10) and a final verdict line
  (`APPROVE` / `REQUEST CHANGES` / `COMMENT`), where the verdict is the Go-computed
  result reported by `vibe-check diff`.

**Rationale**: Matches the other divisors' output format and makes the delta
sourced from Go-computed, auditable output (Constitution III observability, VI fidelity).

### D10: Graceful degradation

**Decision**: If the `vibe-check` binary is not on `PATH`, the base ref cannot be
resolved or analyzed, `git worktree` fails, or EITHER the base OR the PR ref fails
to analyze (e.g., the PR does not build), the agent MUST report the limitation and
return **COMMENT** — never a false APPROVE and never a hard crash. The agent MUST
distinguish "PR does not build" from a clean measurement so reviewers understand
why full delta data is unavailable.

**Rationale**: Constitution II (composability): the reviewer must not block a PR
because of its own tooling gap, and must not silently approve when it could not
measure.

## Coverage Strategy

**Unit tests — `internal/scaffold/` (target ≥ 80% line coverage)**: Using
`t.TempDir()`, verify: assets are written to `.opencode/agents/` with mode
`0o644` and created dirs with mode `0o755`; nested directory creation; default
skip-existing behavior (unchanged file, reported as skipped); `--force`
overwrites; embedded `divisor-entropy.md` is non-empty; `metrics.ValidateProjectPath`
rejects traversal (`../`) targets. `Result` array ordering MUST be asserted
deterministic (stable, sorted).

**Embedded-asset contract test**: Parse the embedded `divisor-entropy.md` and
assert:
- `description` is non-empty;
- frontmatter contains `mode: subagent`, `temperature: 0.1`, `edit: deny`,
  `webfetch: deny`;
- a granular `bash` block whose catch-all `"*"` is `deny`;
- the allowlist is **exactly** (set-equality — failing on any missing OR extra
  entry): `git merge-base *`, `git rev-parse *`, `git worktree add *`,
  `git worktree remove *`, `git worktree prune`, `git fetch origin *`,
  `git check-ref-format *`, `vibe-check analyze *`, and `vibe-check diff *`;
- the AP-006 provenance marker prefix `<!-- scaffolded by vibe-check` is present;
- the embedded filename matches the `divisor-*.md` discovery glob;
- the body contains the required sections (Source Documents, Code Review Mode,
  Output Format, Decision Criteria, Security / Operating Constraints).

The bash allowlist set-equality assertion MUST be implemented using stdlib parsing
only (no YAML dependency). The implementation extracts the frontmatter `bash:` block
region via string/byte scanning and asserts: presence of each of the exact quoted
allow keys, absence of any other allow entry, and `"*": "deny"` (set-equality;
fail on missing OR extra entries). No new Go module dependency is introduced.
This mechanically guards the agent contract.

**Delta+verdict unit tests — `metrics/delta_test.go`, `metrics/verdict_test.go`
(target ≥ 80% line coverage)**: Table-driven tests over schema-valid `ModuleGraph`
fixtures (each validated with `metrics.Validate`) covering:
- ΔInstability at 0.14 (→ COMMENT) vs 0.15 (→ REQUEST_CHANGES);
- ΔDistance at 0.19 (→ COMMENT) vs 0.20 (→ REQUEST_CHANGES);
- ΔLCOM at 1 (→ COMMENT) vs 2 (→ REQUEST_CHANGES);
- a new-cycle positive case (→ REQUEST_CHANGES);
- a pre-existing-cycle negative case (MUST NOT flag as new);
- the COMMENT band (non-zero shifts below all thresholds);
- improve/stable → APPROVE;
- a partial-build case: an input with `Status != "complete"` or a load-error
  `Warnings` entry → COMMENT (never APPROVE) with the unreliable/partial flag set
  and added/removed signal suppressed, plus a both-complete negative case that
  exercises the full verdict path.
Plus a determinism assertion: identical inputs always yield the same verdict and
output ordering (incl. per-package row order).

**CLI tests — `cmd/vibe-check/` (target ≥ 80% line coverage)**: Use the AP-003
pattern — inject `bytes.Buffer` for stdout/stderr into `RunInit` and `RunDiff`.
Test the human summary, the `--json` output with specific per-lifecycle values
(`written`/`skipped`/`forced` per run stage), `--force` vs skip behavior,
and exit codes (0 success, 2 on invalid path). For `RunDiff`, additionally assert
a partial-build input yields COMMENT with the unreliable-measurement flag set in
the `--json` payload. The `init` I/O-failure→exit-2
branch MUST be tested deterministically via an injected filesystem/writer seam
(not guarded with `os.Geteuid()==0`). No subprocess execution.

**Agent behavior validation (acceptance criterion #7)**: Provide fixtures
under `metrics/testdata/` (and/or `internal/scaffold/testdata/`) — an
"improvement" scenario, a "degradation" scenario, a "COMMENT-band" scenario, and a
"partial-build" scenario
(each a pair of pre-computed schema-valid `ModuleGraph` JSON snapshots, each passing
`metrics.Validate`) — and
a documented validation checklist asserting the agent, driven by `vibe-check diff`,
yields APPROVE for the improvement case and REQUEST CHANGES for the degradation
case. The numeric verdict logic itself is covered by the automated table-driven
tests.

**Determinism**: `RunInit` output is stable for identical inputs; `scaffold.Run`
is idempotent under `--force`.

## Operational Notes

- **Single binary for both analyses**: The agent MUST use the *same* `vibe-check`
  binary to analyze both the base worktree and the PR checkout so that the delta
  reflects source changes, not tool-version changes.
- **Worktree location**: The temporary worktree is created in a unique temp
  directory OUTSIDE the repo tree (e.g., via `os.MkdirTemp("", "vibe-check-*")`),
  not inside the repo, to prevent any tool that scans the working tree from
  picking up the worktree as part of the module.
- **Stale worktree recovery**: `git worktree prune` is run best-effort before
  `git worktree add` to recover leaked/stale worktree entries from prior failed
  runs.
- **Bounded analysis**: Both `vibe-check analyze` invocations use a bounded
  `--timeout`; if either times out the agent degrades to COMMENT (per D10) and
  reports the timeout, distinguishing it from a clean measurement.
- **Base ref fetch**: In shallow CI clones the merge-base may be unavailable; the
  agent SHOULD attempt a `git fetch` of the default branch and, if the base still
  cannot be resolved, degrade to COMMENT (per D10). The 9-entry allowlist has no
  `timeout`/`git -c` entry and forbids inline env, so there is no allowlist-level
  bound on the fetch; its duration is bounded only by the ambient runtime/bash-tool
  timeout (fail-safe: a hung fetch is killed → COMMENT).
- **Cleanup**: The temporary worktree MUST be removed (`git worktree remove
  --force`) even when analysis fails.
- **File permissions**: `0o644` for written assets, `0o755` for created
  directories (Constitution V).
- **Error messages**: Init errors MUST name the failed operation, the path, and a
  remediation hint (e.g., "failed to create .opencode/agents: <cause>").

## Risks / Trade-offs

**[R1] Delta/verdict correctness and determinism** →
**Mitigation**: delta and verdict are a pure Go function (`ComputeDelta` /
`DecideVerdict`) with table-driven tests pinning each threshold boundary
(ΔI at 0.14 vs 0.15, ΔD at 0.19 vs 0.20, ΔLCOM at 1 vs 2), new-cycle vs
pre-existing-cycle classification, the COMMENT band, and improve/stable →
APPROVE; determinism is explicitly asserted.

**[R2] Enabling `bash` widens privilege vs divisors that use `bash: deny`** →
**Mitigation**: a granular allowlist (mirroring the `divisor-curator` precedent)
denies all commands except `vibe-check analyze`, `vibe-check diff`, and a narrow
set of ref-resolution and `git worktree` subcommands — no arbitrary shell
(`rm`, `curl`, pipes, redirection) is permitted; `edit` and `webfetch` stay
denied; refs are format-validated and resolved with
`git rev-parse --verify --end-of-options "<ref>^{commit}"` before use; the
worktree is created in a unique temp directory OUTSIDE the repo, removed with
`--force`; metacharacter-bearing or compound commands are denied by the catch-all as a defense-in-depth measure (assumed, not guaranteed; see D4 for the primary metacharacter defense via the ref-sanitization regex and `git rev-parse --verify`).

**[R3] `git worktree` failures (shallow clone, detached HEAD, missing base)** →
**Mitigation**: attempt a fetch of the base (bounded only by the ambient runtime
timeout, not by the allowlist); on failure or timeout degrade to COMMENT rather
than blocking (D10).

**[R4] Tool-version skew confounding the delta** → **Mitigation**: analyze both
refs with the same binary; document that the delta measures source change.

**[R5] Broken base ref cannot serve as a baseline** → **Mitigation**: if base
analysis fails (build errors), degrade to COMMENT and state that a broken base
cannot be a baseline.

**[R6] `init` overwriting user-customized agent files** → **Mitigation**:
default skip-existing; `--force` required to overwrite; skipped files reported.

**[R7] Embedded asset drifting from the dogfooded copy** → **Mitigation**:
optional CI regenerate-and-diff check using `vibe-check init . --force` (or
delete-then-init) so the existing file is overwritten and drift is detectable
(see Open Questions). The provenance marker is version-less, so the
regenerate-and-diff check is deterministic across binary versions.

## Migration Plan

Additive change; no data migration. Deploy: build the `vibe-check` binary → run
`vibe-check init` in the consuming project → the Review Council auto-discovers
`divisor-entropy` on the next `/uf.review-council` run. **Rollback**: delete
`.opencode/agents/divisor-entropy.md` (or `git revert` the init commit); the agent
holds no persistent state.

## Open Questions

- Final threshold defaults (ΔI ≥ 0.15, ΔD ≥ 0.20, ΔLCOM ≥ 2) — proposed here;
  ratify or tune during Review Council. Once ratified they are protected gates.
- Add a deterministic `vibe-check diff` command now, or defer to P2?
  **RESOLVED: now included in this change** (see D2). The `diff` command is part
  of the current scope; the earlier plan (agent-computed, `diff` deferred to P2)
  is superseded.
- Should CI enforce that the dogfooded `.opencode/agents/divisor-entropy.md`
  matches the embedded asset via regenerate-and-diff? (Recommendation: yes; the
  CI check must run `vibe-check init . --force` or delete-then-init to detect
  drift, since a plain `init` skips the existing file.)
- Base-ref default: merge-base(`HEAD`, `origin/main`) vs `origin/main` tip vs
  fully configurable. (Recommendation: merge-base, overridable.)
