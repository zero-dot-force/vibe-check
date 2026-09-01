# AGENTS.md — Vibe-Check

## Project Overview

Vibe-Check is a design quality and architectural metrics toolkit for
Go codebases. It computes package-level coupling metrics (afferent/
efferent coupling, instability, abstractness, distance from main
sequence), cohesion analysis, and circular dependency detection —
providing the full Martin metrics suite, which few OSS tools compute
for Go.

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

## Behavioral Rules

These rules are non-negotiable. Violations are CRITICAL severity.

- **Gatekeeping**: MUST NOT modify quality/governance gates
  (coverage thresholds, CRAP scores, severity definitions,
  CI flags, agent settings, constitution MUST rules, review
  limits, workflow markers). Stop and report instead.
- **Phase boundaries**: MUST NOT cross workflow phase boundaries.
  Spec phases: spec artifacts only. Implement: source code.
  Review: fixes only. Violation = process error, stop immediately.
- **CI parity**: MUST replicate CI checks locally before marking
  tasks complete. Derive commands from `.github/workflows/`.
- **Review council**: MUST run `/uf.review-council` before PR
  submission. Resolve all REQUEST CHANGES. No code changes
  between APPROVE and PR. Exempt: constitution amendments,
  docs-only, emergency hotfixes.
- **Branch protection**: MUST NOT commit directly to `main`.
  All changes via feature branches and PRs.
- **Documentation gate**: Before marking a task complete,
  assess documentation impact: `CHANGELOG.md` for change
  entries, `AGENTS.md` for structural updates (project
  structure, conventions, build commands), `README.md` for
  description changes.
- **Documentation gate**: MUST file a documentation issue
  against the current repo for user-facing changes before
  PR merge. Exempt: internal refactoring, test-only,
  CI-only, spec artifacts.
- **Zero-waste**: No orphaned specs, unused standards, or
  aspirational documents that do not map to actionable work.

### PR Review Commands

| Command | When | Scope |
|---------|------|-------|
| `/uf.review-council` | Pre-PR (local) | 5+ Divisor agents |
| `/uf.review-pr [N]` | Post-PR (GitHub) | Single agent, CI analysis |

## Specification Workflow

All non-trivial changes MUST be preceded by a spec workflow.

| Tier | Tool | When | Artifacts |
|------|------|------|-----------|
| Strategic | Speckit | >= 3 stories, cross-repo | `specs/NNN-*/` |
| Tactical | OpenSpec | < 3 stories, single-repo | `openspec/changes/*/` |

Pipeline: `constitution → specify → clarify → plan → tasks →
analyze → checklist → implement`

**Ordering**: Constitution before specs. Spec before plan. Plan
before tasks. Tasks before implementation. Spec artifacts MUST
be committed/pushed before implementation begins.

**Branches**: Speckit: `NNN-<name>`. OpenSpec: `opsx/<name>`.

**Task bookkeeping**: Mark checkboxes `[x]` immediately on
completion. `[P]` marks parallel-eligible tasks.

**When in doubt**: Start with OpenSpec. Escalate to Speckit if
scope grows beyond 3 stories or crosses repo boundaries.

**What requires a spec**: New features, refactoring that changes
signatures, test additions across multiple functions, agent
changes, CI changes, data model changes.

**Exempt**: Constitution amendments, typo fixes, emergency
hotfixes (retroactively documented).

## Build & Test Commands

```bash
# Build
go build ./...

# Build the CLI binary
go build ./cmd/vibe-check

# Deploy embedded Review Council agent assets into .opencode/agents/ of a project
go run ./cmd/vibe-check init .            # --force to overwrite, --json for machine output

# Analyze a module and write ModuleGraph JSON to a file (default: stdout)
go run ./cmd/vibe-check analyze -o graph.json ./...   # --output is the long form

# Compare two ModuleGraph snapshots (base vs PR): entropy delta + verdict
go run ./cmd/vibe-check diff base.json pr.json        # --json for machine output

# Note: `vibe-check analyze` forces GOTOOLCHAIN=local for its go/packages
# subprocess, so analysis never downloads a toolchain named by a target
# module's go.mod. Trade-off: a trusted module whose go.mod `toolchain`
# directive requires a newer-than-local Go must be built/analyzed manually.

# Test (with race detection)
go test -race -count=1 ./...

# Vet
go vet ./...

# Lint (when golangci-lint is configured)
golangci-lint run ./...
```

## Project Structure

```text
.opencode/          # OpenCode agent configuration, skills, packs
.specify/           # Constitution and governance memory
.uf/                # Unbound Force tooling configuration
cmd/vibe-check/     # CLI entry point (Layer 3)
  main.go           # Binary entry point with ldflags version embedding
  root.go           # Cobra root command with --version flag
  analyze.go        # analyze subcommand with threshold flags (incl --output/-o)
  diff.go           # diff subcommand (base vs PR entropy delta + verdict)
  init.go           # init subcommand (deploys embedded agent assets)
internal/goadapter/ # Go language adapter (Layer 2)
  adapter.go        # Adapter struct implementing metrics.Adapter
  resolve.go        # Package loading via go/packages
  types.go          # Type classification (interfaces = abstract)
  lcom.go           # LCOM4 via connected-component analysis
  cycles.go         # Tarjan's SCC for circular dependency detection
  extensions.go     # go.interfaceWidth and go.interfaceProximity extensions
  doc.go            # Package-level GoDoc
  testdata/         # Test fixtures (coupling, types, lcom, extensions, partial)
internal/scaffold/  # Embedded agent-asset deployment for `vibe-check init`
  doc.go            # Package-level GoDoc
  embed.go          # //go:embed assets/agents/*.md (embedded source of truth)
  scaffold.go       # Symlink-safe asset writer (skip/force; 0o755 dirs, 0o644 files)
  scaffold_test.go  # Writer + embedded-asset contract tests
  assets/agents/    # Embedded Review Council agent assets
    divisor-entropy.md  # Structural-entropy divisor agent (source of truth)
metrics/            # Universal coupling metrics model (Layer 1)
  adapter.go        # Adapter interface and Capability type
  compute.go        # Metric computation functions
  cycle.go          # Cycle type for circular dependency representation
  delta.go          # GraphDelta + ComputeDelta (base vs PR deltas, entropy direction)
  doc.go            # Package-level GoDoc
  external.go       # ExternalAdapter (JSON-RPC subprocess)
  graph.go          # ModuleGraph and ModuleResult types (with Extensions)
  jsonrpc.go        # JSON-RPC 2.0 protocol types
  module.go         # Module type (universal unit of analysis)
  modulegraph.schema.json  # JSON Schema for ModuleGraph validation (v1.1)
  registry.go       # Adapter registry (dependency-injected)
  schema.go         # Embedded JSON schema access
  security.go       # Path validation and environment sanitization
  validate.go       # JSON schema validation (accepts v1.0 and v1.1)
  values.go         # Named metric types (Instability, Abstractness, etc.)
  verdict.go        # Verdict + DecideVerdict (protected entropy gate thresholds)
  warning.go        # Warning type for analysis caveats
  zone.go           # Zone and Status types
  testdata/entropy/ # ModuleGraph diff fixtures + validation README (entropy divisor)
openspec/           # OpenSpec change artifacts (proposals, specs, tasks)
  changes/          # Individual change directories
  schemas/          # Spec validation schemas
  specs/            # Spec templates
```

## Architecture

The architecture follows a three-layer design per the RFC phasing:

- **Layer 1** (`metrics/`): Language-agnostic universal model — Ca, Ce,
  Instability, Abstractness, Distance from Main Sequence, LCOM4,
  circular dependency detection, JSON schema validation, adapter
  interface, security primitives, and the base↔PR entropy delta engine
  (`ComputeDelta`) plus verdict engine (`DecideVerdict`) with protected
  gate thresholds (ΔInstability ≥ 0.15, ΔDistance ≥ 0.20, ΔLCOM ≥ 2, or a
  new circular dependency).
- **Layer 2** (`internal/goadapter/`): Go language adapter implementing
  `metrics.Adapter`. Uses `golang.org/x/tools/go/packages` for
  type-aware dependency resolution, AST-based type classification,
  LCOM4 via connected-component analysis (Hitz & Montazeri 1995),
  Tarjan's SCC for cycle detection, and Go-specific extensions
  (`go.interfaceWidth`, `go.interfaceProximity`).
- **Layer 3** (`cmd/vibe-check/`): CLI entry point using cobra. Provides
  `vibe-check analyze` with threshold flags (`--max-instability`,
  `--max-distance`, `--max-lcom`, `--no-circular-deps`, `--timeout`,
  `--output`/`-o`) and JSON output; `vibe-check diff <base.json> <pr.json>`
  computing the entropy delta and verdict (with tighten-only threshold
  overrides); and `vibe-check init [path]` deploying the embedded Review
  Council agent assets into `.opencode/agents/`.

RFC phasing status:

- **P0**: Core coupling metrics engine (complete — `metrics/` package)
- **P0 effective**: Go adapter and CLI (complete — `internal/goadapter/`
  and `cmd/vibe-check/`; classified as P1 in RFC but required for MVP)
- **P2**: Python adapter, cognitive complexity, branch coverage,
  architectural drift tracking
- **P3**: TS/JS adapter, SBOM integration, mutation testing hooks

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

This repository uses convention packs scaffolded by
unbound-force. Agents MUST read the applicable pack(s)
before writing or reviewing code.

- `.opencode/uf/packs/ci-custom.md`
- `.opencode/uf/packs/ci.md`
- `.opencode/uf/packs/content-custom.md`
- `.opencode/uf/packs/content.md`
- `.opencode/uf/packs/default-custom.md`
- `.opencode/uf/packs/default.md`
- `.opencode/uf/packs/go-custom.md`
- `.opencode/uf/packs/go.md`
- `.opencode/uf/packs/python-custom.md`
- `.opencode/uf/packs/python.md`
- `.opencode/uf/packs/severity.md`
- `.opencode/uf/packs/typescript-custom.md`
- `.opencode/uf/packs/typescript.md`
