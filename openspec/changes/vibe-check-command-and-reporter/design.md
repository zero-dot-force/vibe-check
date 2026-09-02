# Design: /vibe-check Command and vibe-check-reporter Agent

## Context

Vibe-check currently provides CLI commands (`analyze`, `diff`, `init`)
and one embedded agent (`divisor-entropy`). The scaffold system in
`internal/scaffold/` embeds agent assets from `assets/agents/*.md` and
deploys them to `.opencode/agents/` in the target project via
`vibe-check init`.

To add a slash command and companion reporter agent, we need to:

1. Create two new markdown assets (command + agent).
2. Extend the scaffold to embed and deploy a second asset category
   (commands alongside agents).
3. Design the agent to invoke `vibe-check analyze`, interpret the JSON
   output, and optionally use Dewey for historical tracking.

The existing `divisor-entropy.md` agent provides a proven pattern for
agent structure, security constraints, and bash allowlists. The
`/gaze` command in the unbound-force ecosystem provides a proven
pattern for command-to-agent delegation: the command file uses
`agent:` YAML frontmatter to name the delegate agent, the command body
provides usage documentation and examples, and `$ARGUMENTS` is the
mechanism for passing user input to the agent at invocation time.

## Goals / Non-Goals

**Goals:**

- Deploy a `/vibe-check` slash command and `vibe-check-reporter` agent
  to consuming projects via `vibe-check init`.
- Support three modes: summary (traffic-light), detailed (per-package),
  and trending (historical comparison via Dewey).
- Agent interprets raw metric JSON into natural-language guidance.
- Scaffold system supports both `agents/` and `commands/` asset
  directories without breaking existing agent-only consumers.
- Graceful degradation when Dewey is unavailable (trending mode reports
  limitation instead of failing).

**Non-Goals:**

- The agent does NOT compute metrics in-prompt — it delegates to
  `vibe-check analyze` and interprets the JSON output.
- No new Go module dependencies.
- No changes to the `metrics/` or `internal/goadapter/` packages.
- No CI pipeline integration (that is the divisor-entropy agent's domain).
- No web UI or dashboard — output is conversational markdown.

## Decisions

### D1: Command delegates to agent via `agent:` frontmatter

**Decision**: The `/vibe-check` command uses OpenCode's `agent:`
frontmatter field to delegate to the `vibe-check-reporter` agent.

**Rationale**: This follows the established pattern used by `/gaze` →
`gaze-reporter`. The command file defines usage docs and passes
`$ARGUMENTS` to the agent. The agent file contains the actual logic,
tool permissions, and interpretation instructions.

**Alternative considered**: Embedding all logic in the command file.
Rejected because commands lack the `permission:` frontmatter and
behavioral constraints that agents provide.

### D2: Scaffold deploys two asset categories via separate embed directives

**Decision**: Add a second `//go:embed` variable in `embed.go` for
`assets/commands/*.md`. Extend `scaffold.go` `Run()` to iterate both
asset filesystems and deploy agents to `.opencode/agents/` and commands
to `.opencode/commands/` in the target directory.

**Rationale**: Go's `//go:embed` requires compile-time glob patterns.
A single `assets/**/*.md` glob would work but loses the ability to
distinguish asset types by filesystem path. Two separate embed
variables (`AgentAssets`, `CommandAssets`) make the code explicit and
allow each category to have its own deployment target.

**Alternative considered**: Single `//go:embed assets` with runtime
path parsing. Rejected — two explicit embed vars are clearer, and
the pattern naturally extends if future categories are added (e.g.,
`skills/`, `packs/`).

**Refactoring approach**: The current `run()` function uses hardcoded
constants (`assetSourceDir = "assets/agents"`, `targetSubdir =
".opencode/agents"`). To avoid duplicating the deployment loop,
extract the per-category deployment logic into a helper function
(e.g., `deployCategory(assets fs.FS, sourceDir, targetSubdir string,
opts Options) ([]string, []string, []string, error)`) that returns
written/skipped/forced slices. `Run()` calls this helper twice — once
for agents, once for commands — and merges the results. Result entries
use category-prefixed relative paths (e.g., `agents/divisor-entropy.md`,
`commands/vibe-check.md`) to disambiguate entries from different asset
categories.

### D3: Agent uses `vibe-check analyze --output` for snapshot capture

**Decision**: The agent runs `vibe-check analyze --output <tempfile> ./...`
to get the ModuleGraph JSON, then reads and interprets it.

**Rationale**: The `--output` flag already exists (added in issue #2)
and the agent's bash allowlist already permits `vibe-check analyze *`.
This avoids shell redirection (which would require widening the
allowlist).

**Tempfile lifecycle**: The agent MUST create the tempfile in the OS
temporary directory with an unpredictable filename (e.g., using a UUID
or timestamp suffix). The agent MUST clean up the tempfile after
reading the JSON content. If cleanup fails, the agent logs a warning
but does not fail the analysis.

### D4: Trending mode stores/retrieves snapshots via Dewey learnings

**Decision**: The agent uses `dewey_store_learning` to persist
ModuleGraph snapshots tagged with commit SHA and timestamp. Retrieval
uses `dewey_semantic_search` with project+metric context.

**Rationale**: Dewey's learning system provides semantic search,
timestamping, and persistence without requiring a new storage backend.
The `tag` field enables filtering by project. This matches the pattern
used by `divisor-entropy` for prior learnings.

**Alternative considered**: File-based snapshot storage (e.g.,
`.vibe-check/snapshots/`). Rejected — adds filesystem state management
that Dewey already solves, and would require changes to `.gitignore`
in consuming projects.

### D5: Three-mode design with summary as default

**Decision**: The command accepts an optional mode argument:
- `(none)` or `summary` → traffic-light overview
- `detailed` → per-package breakdown
- `trending` → historical comparison

**Rationale**: Most developers want a quick health check (summary).
Detailed mode serves debugging/investigation. Trending mode serves
longitudinal tracking. This mirrors the `/gaze` command's mode pattern.

### D6: Agent bash allowlist is minimal with input validation

**Decision**: The agent's bash permission allowlist permits only:
- `vibe-check analyze *` — to run analysis
- `git rev-parse *` — to get current commit SHA for snapshot tagging

All other bash commands are denied.

**Rationale**: The reporter agent operates on the current working tree
only (unlike divisor-entropy which needs worktrees for base/PR
comparison). It does not need `git worktree`, `git fetch`, or
`vibe-check diff`. Minimal allowlist follows least-privilege.

**Input validation**: The agent MUST validate user-supplied package
patterns against a safe character set (`^[A-Za-z0-9./_-]+$`) before
interpolating them into bash commands. Patterns containing shell
metacharacters, flags not recognized by `vibe-check analyze`, or
empty strings MUST be rejected with a clear error message. This
prevents argument injection through the `*` glob in the allowlist.

## Risks / Trade-offs

### R1: Dewey availability for trending mode
**Risk**: Dewey MCP tools may not be available in all environments.
**Mitigation**: Agent checks for Dewey availability before attempting
trending mode. Falls back to a clear message: "Historical data
requires Dewey MCP tools. Run in summary or detailed mode, or
configure Dewey for trending support."

### R2: Large ModuleGraph JSON in Dewey learnings
**Risk**: Storing full ModuleGraph JSON as a learning may exceed
reasonable content sizes for large codebases.
**Mitigation**: Agent stores a summary snapshot (per-package metric
values, not the full graph with all type details) to keep storage
compact. The summary includes enough data for trend comparison.

### R3: Scaffold backward compatibility
**Risk**: Adding command deployment could break existing `vibe-check init`
consumers who only expect agents.
**Mitigation**: The `Result` struct already uses slices (Written,
Skipped, Forced) that naturally accommodate additional entries. The
behavior is purely additive — agent deployment is unchanged, command
files are added alongside.

### R4: Agent mode parsing
**Risk**: User passes an unrecognized mode argument.
**Mitigation**: Agent validates the mode argument against the known
set and reports available modes if unrecognized. Does not fail
silently.

### R5: Agent subprocess failure modes
**Risk**: `vibe-check analyze` may fail in several ways — binary not
in PATH, timeout, or malformed/truncated JSON output.
**Mitigation**: The agent handles each failure mode explicitly:
- Binary not found: report that vibe-check needs to be installed and
  provide installation guidance.
- Timeout or exit code 2: report the error from `vibe-check analyze`
  and suggest running the CLI manually to diagnose.
- Truncated/malformed JSON: report that analysis output is corrupted
  and suggest re-running.

## Test Strategy

Testing for this change covers three categories:

### 1. Scaffold Go code (unit tests)

New scaffold code (embed directive, `deployCategory` helper, `Run()`
extension) is tested following the existing pattern in
`scaffold_test.go`:
- **Synthetic FS tests** (`fstest.MapFS`): Fresh deploy, skip
  existing, force overwrite, mixed agent+command results, sorted
  output. These tests validate the deployment logic in isolation.
- **Real embedded FS tests**: Contract tests that verify the embedded
  assets contain expected files and content (byte-match against
  source).
- **Coverage target**: Maintain or exceed the existing scaffold
  package coverage level. New code paths (command deployment, category
  disambiguation in Result) MUST be covered.

### 2. Embedded asset contract tests

Each embedded markdown asset gets a contract test verifying:
- **Agent (`vibe-check-reporter.md`)**: Valid YAML frontmatter with
  required fields (`description`, `mode: subagent`, `temperature`,
  `permission` block), provenance marker (`<!-- scaffolded by
  vibe-check -->`), bash allowlist entries (exactly `vibe-check
  analyze *` and `git rev-parse *` plus catch-all deny), and required
  content sections.
- **Command (`vibe-check.md`)**: Valid YAML frontmatter with required
  fields (`description`, `agent: vibe-check-reporter`), mode
  documentation (summary/detailed/trending), and `$ARGUMENTS`
  passthrough instruction.

These tests follow the pattern of the existing
`TestEmbeddedAsset_Contract` for `divisor-entropy.md`.

### 3. Agent behavioral validation

Agent and command are markdown prompt assets — their behavioral
correctness is validated at integration time (manual invocation of
`/vibe-check` in an OpenCode session), not via Go unit tests. The
contract tests above ensure the structural prerequisites for correct
behavior (frontmatter, permissions, content sections) are in place.
