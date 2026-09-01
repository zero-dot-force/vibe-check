---
pack_id: agent-design
language: Any
description: "Structural quality rules for AI coding agents. Covers coupling, cohesion, complexity, naming, file size, duplication, coverage, and test assertions."
version: 1.0.0
---
<!-- scaffolded by uf vdev -->

# Convention Pack: Agent Design

Structural quality rules that agents enforce during code generation and review.
Each rule specifies a measurable threshold, an enforcement tool, and pass/fail
examples. Rules are organized by enforcement category.

## Coupling Rules (vibe-check)

### AD-002 Package Fan-out [MUST]

**Severity**: HIGH

**Rationale**: High efferent coupling (Ce) creates fragile packages that are
sensitive to changes in many dependencies. A package importing 10+ other
packages is a change amplifier — any modification to a dependency risks
breaking the high-fan-out consumer.

**Threshold**: Ce < 10

**Enforcement**: `vibe-check analyze` — inspect the `Ce` field in the JSON
output for each module. Automated threshold enforcement via a `--max-ce` flag
is a forward reference to a planned feature; until implemented, agents enforce
this rule by inspecting vibe-check JSON output or heuristically during review.

**Example**:
- PASS: A package imports 7 other packages (Ce = 7)
- FAIL: A package imports 13 other packages (Ce = 13)

---

### AD-003 Instability Threshold [MUST]

**Severity**: HIGH

**Rationale**: High instability in non-leaf packages propagates breaking changes
downstream. A non-leaf package with I >= 0.7 means most of its coupling is
efferent — it depends on many packages and has few dependents, making it both
volatile and depended-upon. Leaf packages (Ca = 0, no downstream dependents)
are exempt because their instability cannot propagate.

**Threshold**: I < 0.7 for non-leaf packages (Ca > 0). Leaf packages (Ca = 0)
are exempt.

**Enforcement**: `vibe-check analyze --max-instability=0.7`

**Example**:
- PASS: A non-leaf package (Ca = 3) has Instability of 0.55
- PASS: A leaf package (Ca = 0) has Instability of 0.9 (exempt)
- FAIL: A non-leaf package (Ca = 2) has Instability of 0.82

---

### AD-004 No Circular Dependencies [MUST]

**Severity**: CRITICAL

**Rationale**: Circular dependencies between packages prevent independent
compilation and deployment. They create tight coupling that makes it impossible
to change one package without considering all packages in the cycle. Circular
dependencies are a structural defect, not a design trade-off.

**Threshold**: Zero circular dependency cycles

**Enforcement**: `vibe-check analyze --no-circular-deps`

**Example**:
- PASS: All package dependencies form a directed acyclic graph (DAG)
- FAIL: `pkg/a` imports `pkg/b`, which imports `pkg/a` (cycle: a -> b -> a)

---

## Cohesion and Duplication Rules (vibe-check)

### AD-008 DRY — No Duplication [MUST]

**Severity**: MEDIUM

**Rationale**: Duplicated code blocks create maintenance burden — a bug fix in
one copy must be replicated to all copies. Blocks of 6 or more consecutive
duplicated lines indicate an extraction opportunity (helper function, shared
constant, or template).

**Threshold**: No duplicated blocks of 6 or more consecutive lines (maximum
allowed block size is 5 lines; blocks of 6+ lines are violations)

**Enforcement**: `vibe-check analyze --max-duplication=5` — this flag is a
forward reference to a planned feature. Until implemented, agents enforce this
rule heuristically during review by identifying duplicated code blocks.

**Example**:
- PASS: No duplicated blocks exceed 5 consecutive lines
- FAIL: An 8-line block is duplicated across two files

---

### AD-009 Package Cohesion [MUST]

**Severity**: HIGH

**Rationale**: Low cohesion indicates a package is doing too many unrelated
things. LCOM4 (Lack of Cohesion of Methods, variant 4 — Hitz & Montazeri 1995)
counts the number of connected components in a package's method-field graph.
LCOM4 = 1 is maximally cohesive (all methods share fields). LCOM4 > 1
indicates the package can be split into independent units.

**Threshold**: LCOM4 <= 3

**Enforcement**: `vibe-check analyze --max-lcom=3`

**Example**:
- PASS: A package has LCOM4 of 1 (single connected component — all methods
  share fields directly or transitively)
- FAIL: A package has LCOM4 of 5 (five disconnected method groups — the
  package should be split)

---

## Complexity and Coverage Rules (gaze)

### AD-001 Cognitive Complexity [MUST]

**Severity**: HIGH

**Rationale**: High cognitive complexity correlates with defect density and
impedes agent comprehension. Functions with cognitive complexity > 15 are
difficult for both humans and AI agents to reason about correctly. Cognitive
complexity measures the mental effort required to understand control flow —
unlike cyclomatic complexity, it penalizes nested structures and rewards
linear flow.

**Threshold**: Cognitive complexity < 15 per function

**Enforcement**: gaze (cognitive complexity scoring)

**Example**:
- PASS: A function has cognitive complexity score of 12
- FAIL: A function has cognitive complexity score of 18

---

### AD-006 Contract Coverage [MUST]

**Severity**: HIGH

**Rationale**: Exported functions define a package's public contract. Every
exported function must have at least one test exercising its contract to
prevent regressions. This is a binary check (covered or not covered), not a
line-coverage percentage requirement.

**Threshold**: Every exported function has >= 1 test exercising its contract

**Enforcement**: gaze (contract coverage analysis)

**Example**:
- PASS: All exported functions have at least one test exercising their contract
- FAIL: An exported function `ComputeMetrics()` has no test covering its
  contract

---

### AD-010 Behavior-Asserting Tests [MUST]

**Severity**: HIGH

**Rationale**: Tests without meaningful assertions provide false confidence.
A test function that sets up state but never checks results is worse than no
test — it gives the illusion of coverage while catching nothing. Every test
function must contain at least one behavioral assertion (a call to `t.Errorf`,
`t.Fatalf`, `t.Error`, `t.Fatal`, or a comparison check).

**Threshold**: Every test function contains >= 1 behavioral assertion

**Enforcement**: gaze (assertion depth analysis)

**Example**:
- PASS: A test function contains 3 behavioral assertions (`t.Errorf` calls
  checking return values and struct fields)
- FAIL: A test function contains 0 behavioral assertions (only setup code,
  no checks)

---

## Naming and Structure Rules

### AD-005 Naming Conventions [MUST]

**Severity**: MEDIUM

**Rationale**: Consistent naming reduces cognitive load and makes code
discoverable. Go naming conventions are well-established: packages are
lowercase single words, interfaces describe behavior (ending in `-er` for
single-method interfaces), and exported identifiers use PascalCase without
stuttering the package name.

**Threshold**: All exported identifiers follow Go naming conventions

**Enforcement**: `golangci-lint run` with the `revive` linter

**Example**:
- PASS: Package `metrics` exports type `Collector` (no stuttering)
- FAIL: Package `metrics` exports type `MetricsCollector` (stutters package
  name)

---

### AD-007 File Size [MUST]

**Severity**: MEDIUM

**Rationale**: Large files impair navigability and increase merge conflict
probability. Files exceeding 400 lines typically contain multiple concerns
that should be separated into distinct files. Smaller files are easier for
agents to load into context and reason about.

**Threshold**: No source file exceeds 400 lines

**Enforcement**: Review agents — agents count lines when loading files. No
specialized tooling needed; any agent can verify this during review.

**Example**:
- PASS: A source file contains 320 lines
- FAIL: A source file contains 485 lines

---

## Custom Rules

<!-- This section is intentionally empty in the canonical pack. Project-specific custom rules belong in agent-design-custom.md -->
