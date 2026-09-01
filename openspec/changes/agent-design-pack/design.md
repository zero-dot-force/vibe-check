## Context

The Unbound Force ecosystem uses convention packs (`.opencode/uf/packs/*.md`) as
the shared source of truth for coding standards that agents enforce during code
generation and review. Existing packs cover language idioms and CI policy, but no
pack addresses structural quality — coupling, cohesion, complexity, and naming.

The metric engines needed to enforce these rules now exist:
- **Vibe-Check** computes Ca, Ce, Instability, Abstractness, Distance, LCOM4,
  and circular dependency detection for Go packages
- **Gaze** computes CRAP scores, assertion depth, contract coverage, and
  cognitive complexity for Go functions
- **golangci-lint** (with revive) enforces naming conventions

This pack bridges the gap between the metric engines and agent behavior by
defining the thresholds and enforcement mappings.

## Goals / Non-Goals

### Goals
- Define 10 structural quality rules (AD-001 through AD-010) with clear,
  measurable thresholds
- Map each rule to a specific enforcement tool (vibe-check, gaze, golangci-lint,
  or review agent)
- Provide pass/fail examples so agents can self-evaluate without running the tool
- Follow the existing convention pack format so agents discover and apply these
  rules through the standard pack-loading mechanism

### Non-Goals
- Implementing new CLI flags in vibe-check (e.g., `--max-duplication`,
  `--max-ce`) — those are tracked separately; the pack references them as
  forward declarations
- Adding CI pipeline integration — packs are consumed by agents at review time
- Defining language-specific idioms — those belong in `go.md`, `typescript.md`
- Covering runtime or performance metrics — this pack is strictly structural
- Distance from Main Sequence — while vibe-check computes this metric and has a
  `--max-distance` flag, issue #4 did not include it in the rule set. It can be
  added in a future pack revision

## Decisions

### D1: Single file vs. multi-file pack

**Decision**: Single file `.opencode/uf/packs/agent-design.md`

**Rationale**: Convention packs are loaded atomically by agents. A single file
with 10 rules (~200 lines) is well within the readable range. Splitting into
per-rule files would require a pack index mechanism that does not exist. The
AGENTS.md `Convention Packs` section references individual pack files.

**Alternatives considered**:
- One file per rule (AD-001.md, AD-002.md, ...): rejected — no pack index
  mechanism, agents would need to glob and load 10 files
- Grouped files by enforcement tool: rejected — rules should be discovered by
  ID, not by tool

### D2: Rule format

**Decision**: Each rule uses a consistent structure: ID, Name, Severity,
Rationale, Threshold, Enforcement (tool + command), Example (pass/fail).

**Rationale**: This mirrors the format used in `severity.md` for severity
definitions and is what review agents expect when evaluating code against
convention pack rules. The ID prefix `AD-` (Agent Design) follows the pattern
of other rule namespaces.

### D3: Threshold values

**Decision**: Use the thresholds specified in
[GitHub issue #4](https://github.com/zero-dot-force/vibe-check/issues/4):
- Cognitive Complexity < 15/function
- Ce < 10
- Instability < 0.7 for non-leaf packages (Ca = 0 packages exempt)
- LCOM4 ≤ 3 (enforced via `--max-lcom=3`)
- File size < 400 lines
- No duplicated blocks ≥ 6 consecutive lines (`--max-duplication=5`)

**Rationale**: These values are drawn from established software engineering
research (Martin metrics, Hitz & Montazeri) and calibrated against Go ecosystem
norms. They match the thresholds already documented in vibe-check's verdict
engine (`verdict.go`).

### D4: Forward-referencing unimplemented features

**Decision**: Rules may reference vibe-check CLI flags that are not yet
implemented (e.g., `--max-duplication`). These are marked with a note indicating
the feature is planned.

**Rationale**: The convention pack defines the target state. Agents that cannot
run the enforcement tool yet will apply the rule heuristically during review.
This avoids needing to re-release the pack when features ship.

## Risks / Trade-offs

- **[Forward references]** → Rules AD-002 (`--max-ce`), AD-008
  (`--max-duplication`), and AD-009 (LCOM threshold via `--max-lcom` already
  exists but AD-009's threshold of 3 aligns with the existing flag) reference
  vibe-check capabilities that may not yet be fully automated. **Mitigation**:
  Mark these as "planned" enforcement; agents apply heuristically until tooling
  catches up. AD-009 uses the existing `--max-lcom` flag with an integer
  threshold consistent with LCOM4's connected-component semantics.

- **[Threshold rigidity]** → Hard thresholds (e.g., Ce < 10) may be too strict
  or too lenient for some codebases. **Mitigation**: Thresholds are documented
  as defaults; per-project overrides can be specified in project-level
  `agent-design-custom.md` packs (following the existing `*-custom.md` pattern).

- **[Review agent enforcement for AD-007]** → File size (< 400 lines) has no
  automated tool enforcement — it relies on review agents counting lines.
  **Mitigation**: This is a simple check that any agent can perform via file
  read; no specialized tooling needed.
