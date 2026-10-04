# Agent Design

This document is a human-readable reference for the 10 structural quality rules (AD-001 through AD-010) that Unbound Force agents enforce during code generation and review. It is intended for contributors and agent maintainers who need to understand the full enforcement model across vibe-check, gaze, golangci-lint, and review agents.

## Rule Reference Table

| ID | Name | Threshold | Severity | Enforcement Tool | Rationale |
|----|------|-----------|----------|-----------------|-----------|
| AD-001 | Cognitive Complexity | < 15 per function | HIGH | gaze | High complexity correlates with defect density and impedes agent comprehension |
| AD-002 | Package Fan-out | Ce < 10 | HIGH | vibe-check (planned) | High efferent coupling creates fragile packages sensitive to many dependencies |
| AD-003 | Instability Threshold | I < 0.7 (non-leaf); leaf packages exempt | HIGH | vibe-check + review agents | High instability in non-leaf packages propagates breaking changes downstream |
| AD-004 | No Circular Dependencies | Zero cycles | CRITICAL | vibe-check | Cycles prevent independent compilation and deployment |
| AD-005 | Naming Conventions | Follow Go conventions | MEDIUM | golangci-lint | Consistent naming reduces cognitive load and makes code discoverable |
| AD-006 | Contract Coverage | Every exported function >= 1 test | HIGH | gaze | Exported functions define public contracts that need test protection |
| AD-007 | File Size | <= 400 lines | MEDIUM | review agents | Large files impair navigability and increase merge conflict probability |
| AD-008 | No Duplication | No duplicated blocks >= 6 lines | MEDIUM | vibe-check (planned) | Duplicated blocks create maintenance burden across copies |
| AD-009 | Package Cohesion | LCOM4 <= 3 | HIGH | vibe-check | Low cohesion indicates a package is doing too many unrelated things |
| AD-010 | Behavior-Asserting Tests | Every test >= 1 assertion | HIGH | gaze | Tests without assertions provide false confidence |

## Rule-by-Rule Detail

### AD-001: Cognitive Complexity

**Severity**: HIGH | **Enforcement**: gaze

**Rationale**: High cognitive complexity correlates with defect density and impedes agent comprehension. Functions with cognitive complexity > 15 are difficult for both humans and AI agents to reason about correctly. Cognitive complexity measures the mental effort required to understand control flow — unlike cyclomatic complexity, it penalizes nested structures and rewards linear flow.

**Tool flag**: gaze (cognitive complexity scoring)

**Examples**:
- PASS: A function has cognitive complexity score of 12
- FAIL: A function has cognitive complexity score of 18

### AD-002: Package Fan-out

**Severity**: HIGH | **Enforcement**: vibe-check (`--max-ce`, planned)

> **Forward reference**: `--max-ce` is a planned vibe-check CLI flag not yet implemented. No tracking issue is filed at this time. Agents currently enforce the Ce < 10 threshold by inspecting vibe-check JSON output or heuristically during review.

**Rationale**: High efferent coupling (Ce) creates fragile packages that are sensitive to changes in many dependencies. A package importing 10+ other packages is a change amplifier — any modification to a dependency risks breaking the high-fan-out consumer.

**Tool flag**: `vibe-check analyze --max-ce=10` (planned)

**Examples**:
- PASS: A package imports 7 other packages (Ce = 7)
- FAIL: A package imports 13 other packages (Ce = 13)

### AD-003: Instability Threshold

**Severity**: HIGH | **Enforcement**: vibe-check + review agents

**Rationale**: High instability in non-leaf packages propagates breaking changes downstream. A non-leaf package with I >= 0.7 means most of its coupling is efferent — it depends on many packages and has few dependents, making it both volatile and depended-upon. Leaf packages (Ca = 0, no downstream dependents) are exempt because their instability cannot propagate.

**Threshold**: I < 0.7 for non-leaf packages (Ca > 0). Leaf packages (Ca = 0) are exempt.

**Enforcement model**:
- `vibe-check analyze --max-instability=0.7` applies the threshold uniformly to all packages
- Review agents enforce the leaf-package exemption by inspecting `Ca` in the JSON output and skipping leaf packages when evaluating this rule

**Examples**:
- PASS: A non-leaf package (Ca = 3) has Instability of 0.55
- PASS: A leaf package (Ca = 0) has Instability of 0.9 (exempt)
- FAIL: A non-leaf package (Ca = 2) has Instability of 0.82

### AD-004: No Circular Dependencies

**Severity**: CRITICAL | **Enforcement**: vibe-check

**Rationale**: Circular dependencies between packages prevent independent compilation and deployment. They create tight coupling that makes it impossible to change one package without considering all packages in the cycle. Circular dependencies are a structural defect, not a design trade-off.

**Tool flag**: `vibe-check analyze --no-circular-deps`

**Examples**:
- PASS: All package dependencies form a directed acyclic graph (DAG)
- FAIL: `pkg/a` imports `pkg/b`, which imports `pkg/a` (cycle: a -> b -> a)

### AD-005: Naming Conventions

**Severity**: MEDIUM | **Enforcement**: golangci-lint

**Rationale**: Consistent naming reduces cognitive load and makes code discoverable. Go naming conventions are well-established: packages are lowercase single words, interfaces describe behavior, and exported identifiers use PascalCase without stuttering the package name.

**Tool flag**: `golangci-lint run` with the `revive` linter

**Examples**:
- PASS: Package `metrics` exports type `Collector` (no stuttering)
- FAIL: Package `metrics` exports type `MetricsCollector` (stutters package name)

### AD-006: Contract Coverage

**Severity**: HIGH | **Enforcement**: gaze

**Rationale**: Exported functions define a package's public contract. Every exported function must have at least one test exercising its contract to prevent regressions. This is a binary check (covered or not covered), not a line-coverage percentage requirement.

**Tool flag**: gaze (contract coverage analysis)

**Examples**:
- PASS: All exported functions have at least one test exercising their contract
- FAIL: An exported function `ComputeMetrics()` has no test covering its contract

### AD-007: File Size

**Severity**: MEDIUM | **Enforcement**: review agents

**Rationale**: Large files impair navigability and increase merge conflict probability. Files exceeding 400 lines typically contain multiple concerns that should be separated into distinct files. Smaller files are easier for agents to load into context and reason about.

**Tool flag**: None — agents count lines when loading files during review

**Examples**:
- PASS: A source file contains 320 lines
- FAIL: A source file contains 485 lines

### AD-008: No Duplication

**Severity**: MEDIUM | **Enforcement**: vibe-check (`--max-duplication`, planned)

> **Forward reference**: `--max-duplication` is a planned vibe-check CLI flag tracked at [zero-dot-force/vibe-check#13](https://github.com/zero-dot-force/vibe-check/issues/13). Until implemented, agents enforce this rule heuristically during review by identifying duplicated code blocks.

**Rationale**: Duplicated code blocks create maintenance burden — a bug fix in one copy must be replicated to all copies. Blocks of 6 or more consecutive duplicated lines indicate an extraction opportunity (helper function, shared constant, or template).

**Tool flag**: `vibe-check analyze --max-duplication=5` (planned)

**Examples**:
- PASS: No duplicated blocks exceed 5 consecutive lines
- FAIL: An 8-line block is duplicated across two files

### AD-009: Package Cohesion

**Severity**: HIGH | **Enforcement**: vibe-check

**Rationale**: Low cohesion indicates a package is doing too many unrelated things. LCOM4 (Lack of Cohesion of Methods, variant 4 — Hitz & Montazeri 1995) counts the number of connected components in a package's method-field graph. See [LCOM4 Semantics](#lcom4-semantics) below for details.

**Tool flag**: `vibe-check analyze --max-lcom=3`

**Examples**:
- PASS: A package has LCOM4 of 1 (single connected component — all methods share fields directly or transitively)
- FAIL: A package has LCOM4 of 5 (five disconnected method groups — the package should be split)

### AD-010: Behavior-Asserting Tests

**Severity**: HIGH | **Enforcement**: gaze

**Rationale**: Tests without meaningful assertions provide false confidence. A test function that sets up state but never checks results is worse than no test — it gives the illusion of coverage while catching nothing. Every test function must contain at least one behavioral assertion.

**Tool flag**: gaze (assertion depth analysis)

**Examples**:
- PASS: A test function contains 3 behavioral assertions (`t.Errorf` calls checking return values and struct fields)
- FAIL: A test function contains 0 behavioral assertions (only setup code, no checks)

## Enforcement Tool Matrix

| Tool | Rules Enforced | Status |
|------|---------------|--------|
| **vibe-check** | AD-003 (Instability), AD-004 (Cycles), AD-009 (LCOM4) | Active |
| **vibe-check** | AD-002 (Fan-out via `--max-ce`), AD-008 (Duplication via `--max-duplication`) | Planned (forward references) |
| **gaze** | AD-001 (Complexity), AD-006 (Coverage), AD-010 (Assertions) | Active |
| **golangci-lint** | AD-005 (Naming, via revive linter) | Active |
| **review agents** | AD-003 (leaf exemption), AD-007 (File Size) | Active |

AD-003 has a split enforcement model: vibe-check enforces the numeric threshold (`--max-instability=0.7`) uniformly, while review agents enforce the leaf-package exemption (Ca = 0) by inspecting the `Ca` value in vibe-check JSON output.

## Override Mechanism

Project-specific threshold overrides use the custom rules file at `.opencode/uf/packs/agent-design-custom.md`. This file is loaded alongside the canonical pack by Cobalt-Crush (during implementation) and all Divisor persona agents (during review).

### CR-NNN Prefix Convention

All custom rules use the `CR-NNN` prefix to distinguish them from canonical AD rules. Rules use `[MUST]`, `[SHOULD]`, or `[MAY]` severity indicators per RFC 2119.

### Scoped Overrides

To override an AD-* threshold for a specific package, add a `CR-NNN` rule scoped to that package with justification. For example:

```markdown
### CR-001 Elevated Ce for utility packages [SHOULD]
**Override**: AD-002 (Ce < 10) → Ce < 15 for `internal/helpers`

**Justification**: This package provides shared formatting utilities that
must import many standard-library packages. Each import is stable (stdlib
APIs are backward-compatible), and the package has no business logic.
```

### Justification Requirement

Every override must include justification explaining why the threshold is being modified for this project. Overrides without justification are rejected during review.

## LCOM4 Semantics

AD-009 uses **LCOM4 integer semantics** (connected-component count per Hitz & Montazeri 1995), not a 0.0–1.0 cohesion float ratio.

- **LCOM4 = 1**: Maximally cohesive — all methods in the package share fields directly or transitively, forming a single connected component.
- **LCOM4 = 2 or 3**: Moderate cohesion — the package has 2 or 3 distinct method groups. Acceptable for packages where the groups serve a common purpose.
- **LCOM4 > 3**: Low cohesion — the package has 4+ disconnected method groups. Strong indicator that the package should be split into smaller, more focused packages.

The value represents the count of connected components in the method-field graph, not a float ratio. This distinction is critical for agents interpreting vibe-check output: a `--max-lcom=3` flag means "at most 3 connected components," not "at most 0.3 on a float scale."

## Authority

This document is a derivative reference. **The canonical and authoritative source** for all rules is the pack file at `.opencode/uf/packs/agent-design.md` (version 1.0.0). If this document conflicts with the pack file, the pack file takes precedence. Always consult the pack file for the most current rule definitions and thresholds.