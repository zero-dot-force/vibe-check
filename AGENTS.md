# AGENTS.md — Vibe-Check

## Project Overview

Vibe-Check is a design quality and architectural metrics toolkit for
Go codebases. It computes package-level coupling metrics (afferent/
efferent coupling, instability, abstractness, distance from main
sequence), cohesion analysis, and circular dependency detection —
providing the Martin metrics suite that no single OSS tool currently
computes for Go.

Vibe-Check also serves as the metrics backbone for the Unbound Force
ecosystem's entropy sentinel and architectural drift tracking
capabilities.

- **Language**: Go (version specified in `go.mod`)
- **Module**: `github.com/zero-dot-force/vibe-check`
- **License**: Apache 2.0
- **RFC**: [Design Quality & Architectural Metrics](https://github.com/orgs/unbound-force/discussions/483)

## Core Mission

Every agent working on this project MUST internalize three strategic
priorities that override tactical convenience:

- **Strategic Architecture**: Never add complexity that does not serve
  the project's long-term architectural goals. Every function, package,
  and dependency must earn its place.
- **Outcome Orientation**: The goal is working, tested, production-ready
  code — not activity. A session that produces one well-tested function
  is more valuable than one that produces ten untested functions.
- **Intent-to-Context**: Before writing any code, understand why the
  change is being made. Read the spec, the design doc, and the
  constitution. If the intent is unclear, ask — do not guess.

## Behavioral Constraints

### Zero-Waste Mandate

Every tool call, file read, and code generation MUST serve the current
task. Agents MUST NOT:

- Generate boilerplate "just in case"
- Create files that are not required by the current spec
- Add dependencies without explicit justification
- Write aspirational TODOs instead of completing the actual work

### Neighborhood Rule

When modifying a file, agents MUST review the surrounding context:

- Read the entire file before editing (not just the target function)
- Check callers and callees of modified functions
- Verify that changes do not break existing tests
- Update documentation that references modified behavior

### Intent Drift Detection

Agents MUST self-monitor for scope creep. Before each tool call, verify:

- "Is this action required by the current task?"
- "Does this change align with the spec and design?"
- "Am I solving the problem stated, or a different problem?"

If the answer to any question is "no," stop and reassess.

### Automated Governance

Agents MUST treat automated checks (linters, tests, CI) as
non-negotiable gates, not suggestions. A passing build is a
prerequisite, not a goal.

## Gatekeeping Value Protection

Agents MUST NOT modify values that serve as quality or governance
gates. Protected values include but are not limited to:

1. **Coverage thresholds** — minimum coverage percentages in CI config
2. **Severity definitions** — CRITICAL/HIGH/MEDIUM/LOW classifications
3. **MUST/SHOULD rule classifications** — RFC 2119 keywords in specs
4. **CI flags** — `-race`, `-count=1`, `-vet=all` and similar
5. **Review iteration limits** — max review cycles before escalation
6. **Agent temperature and tool-access settings** — model parameters
7. **Pinned dependency versions** — in `go.mod` or CI workflows
8. **Metric thresholds** — instability, distance, CRAP score gates

When an implementation cannot meet a gate, the agent MUST report the
failure and stop rather than weakening the gate.

## Workflow Phase Boundaries

### Specification Phase

During specification phases (proposal, spec, design, tasks), agents:

- MUST only write to files within the spec directory
  (`openspec/changes/<name>/` or `specs/NNN-*/`)
- MUST NOT write implementation code
- MUST NOT modify production source files
- MAY read any file in the repository for context

### Implementation Phase

During implementation, agents:

- MUST follow the task list from `tasks.md`
- MUST mark tasks complete as they are finished (`- [ ]` → `- [x]`)
- MUST run CI-equivalent checks before declaring a task complete
- MUST NOT modify spec artifacts unless correcting a factual error

### Review Phase

During review, agents:

- MUST NOT modify code under review (only the author may)
- MUST provide findings with file paths, line numbers, and severity
- MUST cite the specific principle or convention violated

## Technical Guardrails

### CI Parity Gate

Before marking any task complete or declaring a PR ready:

1. Read `.github/workflows/` to identify the exact CI commands
2. Execute those same commands locally
3. Any failure is a blocking error — the task is not complete

Do NOT rely on a memorized list of commands. Always derive them from
the workflow files.

### Project-Specific Guardrails

- Metric computations MUST produce deterministic results for the same
  input. Non-deterministic output is a bug.
- Coupling analysis MUST handle circular dependencies without infinite
  loops or panics.
- All metric values MUST have defined ranges and units documented in
  their GoDoc comments.
- Language adapters MUST implement a common interface; adding a new
  language MUST NOT require changes to the core analysis engine.

## Council Governance Protocol

### Review Council

The review council consists of four reviewers:

| Reviewer  | Domain                                         |
|-----------|-------------------------------------------------|
| Adversary | Security, resilience, error handling, injection |
| Architect | Structure, patterns, conventions, DRY           |
| Guard     | Plan alignment, zero-waste, constitution        |
| Tester    | Test quality, coverage, isolation, assertions   |

### Council as PR Prerequisite

Before submitting any pull request, agents MUST:

1. Run the `/review-council` command
2. Receive an APPROVE verdict from all four reviewers
3. Resolve any REQUEST CHANGES findings before PR submission
4. Ensure minimal-to-no code changes between council approval and PR
   submission — the council reviews what will be submitted

## Spec-First Development

### What Requires a Spec

All changes that modify production code, test code, agent prompts,
embedded assets, or CI configuration MUST be preceded by a spec
workflow. The spec artifacts (proposal, design, tasks at minimum) MUST
exist before implementation begins.

### Exemptions

- Constitution amendments (governed by the Governance section)
- Trivial fixes: typo corrections, comment-only changes, single-line
  formatting fixes that do not alter behavior
- Emergency hotfixes: critical production bugs where the fix is a
  single well-understood correction (must be retroactively documented)

When in doubt, use a spec. The cost of an unnecessary spec is minutes;
the cost of an unplanned change is rework, drift, and broken CI.

### Spec Pipeline

This project supports two spec workflows:

1. **OpenSpec** (`openspec/changes/<name>/`): proposal → spec → design
   → tasks → apply
2. **Speckit** (`specs/NNN-*/`): constitution → specify → clarify →
   plan → tasks → analyze → checklist → implement

### Ordering and Bookkeeping Gates

- Spec artifacts MUST be created in dependency order (proposal before
  design, design before tasks)
- Task checkboxes MUST be updated from `- [ ]` to `- [x]` immediately
  when a task is completed — not in a batch
- Documentation impact MUST be assessed before marking any task
  complete

## Build & Test Commands

<!-- Placeholder — update when go.mod and Makefile are created -->

```bash
# Build
go build ./...

# Test (with race detection)
go test -race -count=1 ./...

# Vet
go vet ./...

# Lint (when golangci-lint is configured)
golangci-lint run ./...
```

## Architecture

<!-- Placeholder — update when package structure is established -->

The planned architecture follows the RFC phasing:

- **P0**: Core coupling metrics engine — Ca, Ce, Instability,
  Abstractness, Distance from Main Sequence, cohesion, circular
  dependency detection
- **P1**: Multi-language adapter interface, metrics command, convention
  pack integration
- **P2**: Python adapter, cognitive complexity, branch coverage,
  architectural drift tracking
- **P3**: TS/JS adapter, SBOM integration, mutation testing hooks

Package layout will be established during the first spec workflow.

## Coding Conventions

- **Formatting**: `gofmt` is the law. All code MUST be formatted with
  `gofmt`. Use `goimports` to manage import grouping.
- **Documentation**: Every exported function, type, method, and
  constant MUST have a GoDoc comment. Package-level doc comments are
  required for every package.
- **Error handling**: Return errors, do not panic. Wrap errors with
  `fmt.Errorf("context: %w", err)` to preserve the error chain.
  Never discard errors silently.
- **Global state**: No global mutable state. Use dependency injection
  for all services, loggers, and configuration.
- **Logging**: Use `charmbracelet/log` for structured logging.
- **CLI**: Use `cobra` for command-line interface if a CLI is needed.
- **Naming**: Follow Go naming conventions. Packages are lowercase
  single words. Interfaces describe behavior (`Analyzer`, `Adapter`).
  Avoid stuttering (`metrics.MetricsCollector` → `metrics.Collector`).
- **Concurrency**: Document goroutine ownership. Use `context.Context`
  for cancellation. Protect shared state with mutexes or channels,
  never both for the same resource.

## Knowledge Retrieval

### Dewey Tool Selection Matrix

| Need                        | Tool                        |
|-----------------------------|-----------------------------|
| Find a specific page        | `dewey_get_page`            |
| Search by keyword           | `dewey_search`              |
| Find similar concepts       | `dewey_semantic_search`     |
| Check page connections      | `dewey_get_links`           |
| Store a learning            | `dewey_store_learning`      |
| Find stored learnings       | `dewey_semantic_search`     |
| Graph overview              | `dewey_graph_overview`      |
| Journal entries by date     | `dewey_journal_range`       |

### Graceful Degradation

When Dewey is unavailable, agents MUST degrade gracefully:

1. **Tier 1** (Dewey available): Use the full tool suite for semantic
   search, knowledge graph queries, and learning storage
2. **Tier 2** (Dewey unavailable, files accessible): Fall back to
   direct file reads of `.specify/memory/`, `openspec/`, and `specs/`
   directories
3. **Tier 3** (minimal): Use in-session context only; document any
   knowledge gaps encountered for later resolution

## Testing Conventions

- **Framework**: Standard library `testing` package only. No testify,
  no gomock, no third-party test frameworks.
- **Assertions**: Use `t.Errorf` and `t.Fatalf` with descriptive
  messages. Format: `got %v, want %v`.
- **Test isolation**: Each test MUST be independent. Use `t.TempDir()`
  for filesystem tests. Use `t.Parallel()` where safe.
- **HTTP tests**: Use `httptest.NewServer` — never call live endpoints.
- **Race detection**: All tests MUST pass with `-race` flag enabled.
- **Table-driven tests**: Prefer table-driven tests for functions with
  multiple input/output combinations.
- **Test naming**: `TestFunctionName_Scenario` (e.g.,
  `TestInstability_ZeroEfferentCoupling`).
- **Fixtures**: Use `testdata/` directories for test fixtures. Fixtures
  are committed to the repository.

## Git & Workflow

- **Commit messages**: Conventional commit format (`type: description`).
  Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`,
  `ci`, `build`.
- **Branching**: All work on feature branches. Speckit: `NNN-<name>`.
  OpenSpec: `opsx/<name>`.
- **Code review**: Every PR requires at least one approving review.
- **Releases**: Semantic versioning. Breaking changes to metric
  computation behavior, output schemas, or adapter interfaces require
  a MAJOR bump.
- **No force pushes**: History rewriting on shared branches is
  prohibited.

## Convention Packs

<!-- Placeholder — update when convention packs are established -->

Convention packs will be stored under `.opencode/uf/packs/` and loaded
by the OpenCode framework. Planned packs:

- `metrics-conventions.md` — metric naming, ranges, and output format
  standards
- `adapter-conventions.md` — language adapter interface contracts
