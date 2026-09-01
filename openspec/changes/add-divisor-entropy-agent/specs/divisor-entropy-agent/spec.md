## ADDED Requirements

### Requirement: Structural Delta Computation

The agent SHALL compute the structural quality delta between the base ref and
the PR ref by running `vibe-check analyze --output <file>` on both refs to produce two
`ModuleGraph` JSON documents, then running `vibe-check diff <base.json> <pr.json>`
to compute the delta in a pure, deterministic Go function (`metrics.ComputeDelta`).
The agent MUST use `vibe-check` — not `gaze` or `goda` — as the metric source,
and MUST use the same binary for both analyses so the delta reflects source changes
rather than tool-version changes. The per-package deltas (Ca, Ce, Instability,
Abstractness, Distance from main sequence, LCOM, and circular dependencies) and the
verdict are Go-computed outputs reported by `vibe-check diff`, not computed in-prompt
by the agent.

#### Scenario: Analyzing base and PR refs

- **WHEN** the agent reviews a PR branch
- **THEN** it runs `vibe-check analyze --output <file>` on the base ref and on the PR ref, passes
  the two resulting `ModuleGraph` JSON files to `vibe-check diff`, and reports
  the per-package deltas for Ca, Ce, Instability, Abstractness, Distance, and LCOM
  sourced from the `vibe-check diff` output

#### Scenario: Base ref derived from merge-base

- **WHEN** the agent determines the base ref
- **THEN** it uses the merge-base of `HEAD` and the default branch (matching the
  council's three-dot `git diff main...HEAD` scope) unless an explicit base ref
  is provided

#### Scenario: Delta threshold correctness guaranteed by diff-command capability

- **WHEN** the agent invokes `vibe-check diff`
- **THEN** the boundary correctness and determinism of all threshold gates is
  guaranteed by the `diff-command` capability's table-driven Go tests
  (see `specs/diff-command/spec.md`); threshold scenarios are not duplicated here

### Requirement: Base Branch Isolation

The agent SHALL analyze the base ref inside an isolated `git worktree` so that
the PR working tree and index are never mutated, and SHALL remove the temporary
worktree after analysis completes, including when analysis fails.

#### Scenario: Base analyzed in an isolated worktree

- **WHEN** the agent analyzes the base ref
- **THEN** it creates a temporary git worktree checked out at the resolved base
  SHA in a unique temp directory outside the repo and runs `vibe-check analyze`
  there without altering the PR checkout

#### Scenario: Worktree removed after analysis

- **WHEN** base analysis finishes or fails
- **THEN** the agent removes the temporary worktree so no residual state remains

### Requirement: New Circular Dependency Detection

The agent SHALL flag any circular dependency that is present in the PR graph but
absent from the base graph, and SHALL return REQUEST CHANGES when a new cycle is
introduced. Pre-existing cycles that are unchanged MUST NOT, on their own, cause
a REQUEST CHANGES verdict.

#### Scenario: PR introduces a new cycle

- **WHEN** the PR graph contains a circular dependency not present in the base
  graph
- **THEN** the agent reports the new cycle and returns REQUEST CHANGES

#### Scenario: Pre-existing cycle is unchanged

- **WHEN** a circular dependency exists identically in both the base and PR
  graphs
- **THEN** the agent does not treat that cycle as a new regression

### Requirement: Deterministic Verdict Thresholds

The agent SHALL render one of APPROVE, REQUEST CHANGES, or COMMENT according to
deterministic thresholds produced by the `vibe-check diff` command (the tested Go
`metrics.DecideVerdict` function) and reported in its JSON output. The agent MUST
return REQUEST CHANGES when a new cycle is introduced, when any existing package's
Instability increases by at least the instability threshold, when any package's
Distance from the main sequence increases by at least the distance threshold, or
when any package's LCOM increases by at least the LCOM threshold. The exact
threshold values and boundary rules (including float rounding to 4 decimal places)
are defined in `specs/diff-command/spec.md` as the single source of truth and are
enforced by the tested Go `metrics.DecideVerdict` function. Added and removed
packages are listed for informational purposes and MUST NOT trigger any threshold
gate. The agent MUST return COMMENT for smaller material shifts, and MUST return
APPROVE when metrics improve or remain stable. The agent reports and explains the
Go-computed verdict rather than computing thresholds in-prompt.

#### Scenario: Metrics improve or remain stable

- **WHEN** no package degrades beyond the materiality bands and no new cycle is
  introduced
- **THEN** the agent returns APPROVE

#### Scenario: Instability increases beyond threshold

- **WHEN** an existing package's Instability increases by at least the instability
  threshold (defined in `specs/diff-command/spec.md`)
- **THEN** the agent returns REQUEST CHANGES citing the affected package

#### Scenario: Cohesion degrades beyond threshold

- **WHEN** a package's LCOM increases by at least the LCOM threshold (defined in
  `specs/diff-command/spec.md`)
- **THEN** the agent returns REQUEST CHANGES citing the cohesion regression

#### Scenario: Minor metric shift

- **WHEN** a package's Instability increases by a non-zero amount below the
  instability threshold and no other threshold is crossed
- **THEN** the agent returns COMMENT flagging the shift for discussion

### Requirement: Auditable Delta Report

The agent SHALL emit an auditable report sourced from the `vibe-check diff` JSON
output, containing the base ref SHA and PR ref SHA compared, a per-package delta
table (Ca, Ce, Instability, Abstractness, Distance, LCOM as base → PR with the
delta), the list of newly introduced cycles, the overall entropy direction
(improving, stable, or degrading), findings in the standard divisor block format,
a domain score from 1 to 10, and a final verdict line.

#### Scenario: Report includes delta table and verdict

- **WHEN** the agent completes its review
- **THEN** the output contains the compared SHAs, a per-package metric delta
  table, any new cycles, the entropy direction, and a final verdict line of
  APPROVE, REQUEST CHANGES, or COMMENT — all sourced from the `vibe-check diff`
  JSON

### Requirement: Minimal Tool Permissions

The agent frontmatter SHALL restrict `bash` to a granular allowlist that denies
all commands by default (via a leading `"*": "deny"` catch-all) and permits only
the exact set of commands needed: `git merge-base *`, `git rev-parse *`,
`git worktree add *`, `git worktree remove *`, `git worktree prune`,
`git fetch origin *`, `git check-ref-format *`, `vibe-check analyze *`, and
`vibe-check diff *`. The agent
frontmatter SHALL deny `edit` and `webfetch`. The agent MUST reject any ref that
does not match the pattern `^[A-Za-z0-9._/-]+$` and MUST additionally validate with
`git check-ref-format` as a complementary check (these two controls are not
equivalent — the regex is the primary metacharacter defense, `check-ref-format` adds
ref-semantic validation). The agent MUST resolve refs with
`git rev-parse --verify --end-of-options "<ref>^{commit}"` before use, never
interpolating untrusted branch strings into shell commands. Any ref passed to
`git fetch origin <ref>` MUST likewise be validated by the `^[A-Za-z0-9._/-]+$`
regex before the fetch; because `git fetch` (unlike `git rev-parse`) accepts no
`--end-of-options` terminator, that regex is the load-bearing pre-fetch guard
(dangerous fetch options require `=` or a space, which the regex already
rejects — defense-in-depth). The compound-command
denial by the OpenCode permission matcher is a defense-in-depth measure and is
assumed but not guaranteed; the ref-sanitization regex and `git rev-parse` resolution
are the primary metacharacter defenses.

#### Scenario: Agent declares least-privilege permissions

- **WHEN** the embedded agent asset is inspected
- **THEN** its frontmatter sets `permission` with `edit: deny`, `webfetch: deny`,
  `mode: subagent`, `temperature: 0.1`, and a granular `bash` block whose catch-all
  `"*"` is `deny` and whose allowlist permits exactly (set-equality) `git merge-base *`,
  `git rev-parse *`, `git worktree add *`, `git worktree remove *`,
  `git worktree prune`, `git fetch origin *`, `git check-ref-format *`,
  `vibe-check analyze *`, and `vibe-check diff *` — no more, no fewer

### Requirement: Review Council Auto-Discovery

The agent file SHALL be named following the `divisor-*.md` convention so the
Review Council discovers and invokes it automatically without any change to the
council command or a central registry.

#### Scenario: Council discovers the entropy divisor

- **WHEN** the Review Council scans `.opencode/agents/` for `divisor-*.md` files
- **THEN** `divisor-entropy` is discovered and delegated as a reviewer in
  parallel with the other divisors

#### Scenario: Embedded filename matches discovery glob

- **WHEN** the embedded agent asset filename is inspected
- **THEN** it matches the `divisor-*.md` glob pattern used by the Review Council
  for auto-discovery

### Requirement: Graceful Degradation

The agent SHALL degrade gracefully when it cannot measure the delta. When the
`vibe-check` binary is not available on `PATH`, the base ref cannot be resolved
or analyzed, `git worktree` creation fails, or either the base or the PR ref
fails to analyze (including partial-build scenarios where `Status != "complete"`
or load-error `Warnings` are present), the agent SHALL report the limitation and
return COMMENT. The agent MUST NOT return a false APPROVE and MUST NOT crash the
review. Partial-build scenarios (where `vibe-check analyze` completes with
`Status: "partial"` and zeroed type metrics) MUST be treated as degraded
measurements, not clean measurements.

#### Scenario: vibe-check binary not on PATH

- **WHEN** the agent cannot locate the `vibe-check` binary
- **THEN** it reports the missing tool and returns COMMENT rather than APPROVE

#### Scenario: Base ref cannot be resolved or analyzed

- **WHEN** the base ref is unavailable (for example a shallow clone) or fails to
  analyze
- **THEN** the agent reports the limitation and returns COMMENT

#### Scenario: PR ref cannot be analyzed

- **WHEN** the PR ref fails to analyze (for example the PR does not build)
- **THEN** the agent reports the limitation and returns COMMENT (never a false
  APPROVE), distinguishing "PR does not build" from a clean measurement

#### Scenario: Partial-build analysis yields degraded verdict

- **WHEN** `vibe-check analyze` completes for either the base or PR ref but
  returns `Status: "partial"` with load-error warnings and zeroed type metrics
- **THEN** the agent treats this as an unreliable measurement, returns COMMENT,
  and clearly distinguishes the partial result from a clean measurement

### Requirement: Trusted-Refs-Only Operating Constraint

The `divisor-entropy` agent SHALL only run on refs the CI context already
trusts and MUST NOT be wired into CI that analyzes untrusted fork pull
requests. This constraint exists because `vibe-check analyze` executes the
target module's build tooling (compilation and cgo) to resolve types, exposing
a residual code-execution surface that `GOTOOLCHAIN=local` does NOT close: that
forcing only prevents a `go.mod` `toolchain` directive from downloading and
executing a different toolchain; it does not prevent cgo or other build-time
code from running on the host. The shipped `divisor-entropy.md` asset MUST carry a
`## Security / Operating Constraints` section documenting this residual surface
and the trusted-refs-only constraint, so downstream adopters cannot deploy the
agent into an untrusted-fork-PR context without an in-asset warning.

#### Scenario: Trusted-refs-only constraint is documented in the shipped asset

- **WHEN** the embedded `divisor-entropy.md` asset is inspected
- **THEN** it contains a `## Security / Operating Constraints` section stating
  that `vibe-check analyze` executes target build tooling and cgo (a residual
  code-execution surface not closed by `GOTOOLCHAIN=local`), that the divisor
  MUST only run on trusted refs, and that it MUST NOT be wired into
  untrusted-fork-PR CI

#### Scenario: Divisor is scoped to trusted refs

- **WHEN** the divisor is deployed into a CI pipeline
- **THEN** it is configured to run only on refs the CI context already trusts,
  and is not wired to analyze untrusted fork pull requests
