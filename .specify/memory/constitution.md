<!--
  SYNC IMPACT REPORT
  ==================
  Version: 1.0.0 (initial ratification with project-specific
    principles and full development workflow)
  Ratified: 2026-08-28

  Principles:
    - I. Autonomous Collaboration (from org constitution v1.2.0)
    - II. Composability First (from org constitution v1.2.0)
    - III. Observable Quality (from org constitution v1.2.0)
    - IV. Testability (from org constitution v1.2.0)
    - V. Security by Default (from org constitution v1.2.0)
    - VI. Metric Fidelity (project-specific)
    - VII. Language Agnosticism (project-specific)

  Parent constitution alignment:
    ✅ Principles I–V align with org constitution v1.2.0
    ✅ Principles VI–VII are additive; no contradiction with org

  Templates requiring updates:
    ✅ openspec/schemas/unbound-force/templates/* — compatible
-->

# Vibe-Check Constitution

**parent_constitution**: unbound-force org constitution v1.2.0

## Core Principles

### I. Autonomous Collaboration

Agents MUST collaborate through well-defined artifacts — files,
reports, and schemas — rather than runtime coupling or synchronous
interaction.

- Every agent MUST be able to complete its primary function without
  requiring synchronous interaction with another agent. An agent MAY
  consume another agent's artifacts, but MUST NOT block waiting for
  a response.
- Agent outputs MUST be self-describing: each artifact MUST contain
  enough metadata (producer identity, version, timestamp, artifact
  type) for any consumer to interpret it without consulting the
  producing agent.
- Inter-agent communication MUST use well-defined artifact formats.
  Agents MUST NOT invent ad-hoc exchange formats.
- Agents SHOULD publish artifacts to a well-known location within
  the project repository so other agents can discover them without
  explicit coordination.

**Rationale**: A swarm of autonomous agents cannot rely on real-time
negotiation. Artifact-based communication makes collaboration
asynchronous, auditable, and resilient to individual agent
unavailability.

### II. Composability First

Every agent MUST be independently installable and usable without any
other agent being present. Combining agents MUST produce additive
value without introducing mandatory dependencies.

- An agent MUST deliver its core value when deployed alone. No agent
  MAY require another agent as a hard prerequisite for installation
  or primary operation.
- Agents MUST expose well-defined extension points (configuration,
  artifact consumption, convention packs) for integration rather
  than requiring modification of their internals.
- When two or more agents are deployed together, their combination
  MUST produce value greater than the sum of their individual
  capabilities. This additive value MUST NOT come at the cost of
  standalone functionality.
- Agents SHOULD auto-detect the presence of other agents and
  activate enhanced functionality when peers are available, without
  requiring manual configuration.

**Rationale**: Adoption friction kills tools. Composability ensures
each agent earns its place independently, and the swarm becomes
compelling only when its parts are already individually valuable.

### III. Observable Quality

Every agent MUST produce machine-parseable output alongside any
human-readable output. All quality claims MUST be backed by
automated, reproducible evidence.

- Every agent that produces output MUST support at minimum a JSON
  format. Human-readable output (terminal text, Markdown) is
  RECOMMENDED but MUST NOT be the only format available.
- All artifacts MUST include provenance metadata: which agent
  produced the output, which version, when it was produced, and
  against what input (branch, commit, backlog item).
- Quality claims — accuracy rates, coverage percentages, scoring
  thresholds — MUST be backed by automated regression tests or
  benchmarks that can be re-run by any contributor.
- Metrics MUST be comparable across runs. Output formats MUST be
  stable enough that tooling built on an agent's output does not
  break between minor versions.

**Rationale**: A swarm that cannot measure its own performance
cannot improve. Machine-parseable output enables tracking trends,
making data-driven prioritization decisions, and grounding reviews
in evidence rather than opinion.

### IV. Testability

Every component built within this project MUST be testable in
isolation without requiring external services, network access, or
shared mutable state.

- Test contracts MUST verify observable side effects (return values,
  state mutations, I/O operations) rather than implementation
  details.
- Coverage strategy (unit vs. integration vs. e2e, with specific
  targets) MUST be defined in the implementation plan for all new
  code.
- Coverage ratchets MUST be enforced by automated tests; any
  coverage regression MUST be treated as a test failure and block
  the build.
- Missing coverage strategy in a spec or plan is a CRITICAL-severity
  finding and MUST be resolved before implementation begins.

**Rationale**: AI agents generate code rapidly. If that code is not
structurally testable, the resulting system will quickly collapse
under its own unverified complexity. Testability is a first-class
governance concern because untestable code cannot be reliably
verified by any automated mechanism.

### V. Security by Default

Every component built within this project MUST treat security as a
structural property, not a review-time afterthought. Supply chain
integrity, input validation, and least privilege MUST be enforced
by design.

- Dependencies MUST be verified by content hash (SHA256 or
  equivalent) when downloaded outside a package manager's built-in
  verification. CI pipelines MUST pin actions and reusable workflows
  by commit SHA, not mutable tags.
- All external inputs (user input, API payloads, file contents,
  environment variables used as data) MUST be validated and
  sanitized before reaching any security-sensitive operation.
- Components MUST operate with the minimum permissions necessary.
  Secrets MUST be scoped to the narrowest context needed. File
  permissions MUST default to restrictive values (0o644 for files,
  0o755 for executables and directories).
- Before adding an external dependency, the adopter MUST justify
  that the project's existing toolchain cannot cover the same use
  case. Every dependency is attack surface; the default answer is
  "do not add."

**Rationale**: AI agents make adding dependencies and generating
code trivially fast. Without structural security guardrails, the
attack surface of the system grows with each generation cycle.

### VI. Metric Fidelity

Every metric Vibe-Check computes MUST be mathematically faithful to
its published definition. Approximations, heuristics, and sampling
MUST NOT be used unless explicitly documented, justified, and
opt-in.

- Coupling metrics (Ca, Ce, Instability, Abstractness, Distance
  from Main Sequence) MUST implement the formulas as defined by
  Robert C. Martin. Any deviation from the published definition
  MUST be documented in GoDoc and the user-facing output.
- Cohesion metrics MUST define their measurement model (LCOM
  variant or alternative) in the design doc. The chosen model MUST
  be cited and its limitations documented.
- Circular dependency detection MUST be exact (not heuristic). If
  a cycle exists, it MUST be reported. If no cycle exists, it MUST
  NOT be reported. False positives and false negatives in cycle
  detection are bugs.
- All computed values MUST be deterministic: the same input MUST
  produce the same output across runs, platforms, and Go versions
  (within the supported version range). Non-deterministic output
  is a P0 bug.
- Metric values MUST include their valid range and unit in GoDoc
  comments (e.g., "Instability is a float64 in [0.0, 1.0] where
  0.0 is maximally stable").

**Rationale**: Vibe-Check exists to provide trustworthy measurements
that inform architectural decisions. If metrics are inaccurate,
approximate without disclosure, or non-deterministic, they erode
the trust that makes the tool useful. A metric that is wrong is
worse than no metric at all.

### VII. Language Agnosticism

The core analysis engine MUST be decoupled from any single
programming language. Language-specific concerns MUST be isolated
behind adapter interfaces that can be implemented independently.

- The core metric computation engine MUST operate on a
  language-neutral intermediate representation (package graph,
  dependency edges, type classifications). It MUST NOT contain
  Go-specific parsing logic, Python AST traversal, or any other
  language-coupled code.
- Language adapters MUST implement a common interface that
  transforms language-specific source artifacts into the
  intermediate representation. Adding a new language adapter MUST
  NOT require modifications to the core engine or existing
  adapters.
- The Go adapter is the first-class implementation and MUST be
  maintained at feature parity with the core engine. Other language
  adapters MAY lag behind in feature coverage but MUST document
  which metrics they support.
- Adapter discovery SHOULD be automatic (convention-based or
  registry-based). Manual configuration is acceptable as a
  fallback but MUST NOT be the only option.

**Rationale**: The RFC scopes multi-language coupling as P1+ work.
Designing for language agnosticism from the start prevents the Go
implementation from becoming a monolith that must be rewritten when
Python and TypeScript adapters are added. The adapter boundary is
cheaper to establish early than to retrofit.

## Development Workflow

- **Spec-First Development**: All changes that modify production code,
  test code, agent prompts, embedded assets, or CI configuration MUST
  be preceded by a spec workflow (either the Speckit pipeline under
  `specs/` or the OpenSpec pipeline under `openspec/changes/`). The
  spec artifacts (proposal, design, tasks at minimum) MUST exist
  before implementation begins. This ensures every change has a
  planning record, a reviewable intent, and a traceable rationale.
  Exempt from this requirement:
    - Constitution amendments (governed by the Governance section below)
    - Trivial fixes: typo corrections, comment-only changes, and
      single-line formatting fixes that do not alter behavior
    - Emergency hotfixes: critical production bugs where the fix is
      a single well-understood correction (must be retroactively
      documented)
  When in doubt, use a spec. The cost of an unnecessary spec is
  minutes; the cost of an unplanned change is rework, drift, and
  broken CI.
- **Branch Naming**: All work MUST occur on feature branches. Direct
  commits to the main branch are prohibited except for trivial
  documentation fixes. Speckit branches: `NNN-<name>` (e.g.,
  `001-coupling-metrics`). OpenSpec branches: `opsx/<name>` (e.g.,
  `opsx/fix-instability-calc`).
- **Code Review**: Every pull request MUST receive at least one
  approving review before merge.
- **Review Council Gate**: Before submitting a pull request, agents
  MUST run the `/review-council` command and receive an APPROVE
  verdict from all four reviewers (Adversary, Architect, Guard,
  Tester). Any REQUEST CHANGES findings MUST be resolved before
  PR submission. There MUST be minimal to no code changes between
  the council's APPROVE and the PR submission — the council reviews
  the code that will be submitted, not a draft that changes afterward.
- **CI Parity Gate**: Before marking any implementation task complete
  or declaring a PR ready, agents MUST replicate the CI checks
  locally. Read `.github/workflows/` to identify the exact commands
  CI runs, then execute those same commands. Any failure is a
  blocking error — a task is not complete until all CI-equivalent
  checks pass locally. Do not rely on a memorized list of commands;
  always derive them from the workflow files, which are the source
  of truth.
- **Continuous Integration**: The CI pipeline MUST pass (build, lint,
  vet, tests) before a pull request is eligible for merge.
- **Task Completion Bookkeeping**: When a task from `tasks.md` is
  completed during implementation, its checkbox MUST be updated from
  `- [ ]` to `- [x]` immediately. Do not defer this — mark tasks
  complete as they are finished, not in a batch after all work is
  done.
- **Documentation Gate**: Before marking any task complete, agents
  MUST validate whether the change requires documentation updates.
  Check and update as needed: `AGENTS.md`, GoDoc comments, and spec
  artifacts. A task is not complete until its documentation impact
  has been assessed and any necessary updates have been made.
- **Website Documentation Sync**: When a change adds, modifies, or
  removes user-facing behavior (commands, flags, output schemas,
  config fields, installation steps), a GitHub issue MUST be created
  in the `unbound-force/website` repository documenting what changed
  and what website pages need updating. This ensures the public
  documentation stays in sync with the codebase. Exempt: internal
  refactors, test-only changes, spec artifacts.
- **Releases**: Follow semantic versioning (MAJOR.MINOR.PATCH).
  Breaking changes to metric computation behavior, output schemas,
  or adapter interfaces require a MAJOR bump.
- **Commit Messages**: Use conventional commit format
  (`type: description`) to enable automated changelog generation.

## Governance

This constitution extends the Unbound Force org constitution
(v1.2.0). On matters where this document and the org constitution
conflict, the org constitution prevails.

This constitution is the authoritative source for project principles
and development standards. All pull requests and code reviews MUST
verify compliance with these principles.

Amendments to this constitution require:
1. A written proposal documenting the change and its rationale
2. An assessment of impact on existing code and downstream consumers
3. A migration plan if the change affects existing behavior
4. Version increment following semantic versioning:
   - MAJOR: Principle removal or incompatible redefinition
   - MINOR: New principle or materially expanded guidance
   - PATCH: Clarifications, wording fixes, non-semantic changes

Complexity beyond what these principles permit MUST be justified in
the Complexity Tracking section of the implementation plan. The
justification MUST explain why a simpler alternative was rejected.

**Compliance Review**: At each planning phase (spec, plan, tasks),
the Constitution Check gate MUST verify that the proposed work
aligns with all active principles. Constitution violations are
CRITICAL severity and non-negotiable.

**Version**: 1.0.0 | **Ratified**: 2026-08-28
**Parent Constitution**: unbound-force org constitution v1.2.0
