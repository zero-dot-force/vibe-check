## ADDED Requirements

### Requirement: AD-001 Cognitive Complexity
The convention pack SHALL define rule AD-001 requiring that no function exceed a
cognitive complexity score of 15. The rule SHALL specify gaze as the enforcement
tool. The rule SHALL include a rationale explaining that high cognitive complexity
correlates with defect density and impedes agent comprehension.

#### Scenario: Function within threshold
- **GIVEN** the agent-design convention pack is loaded
- **WHEN** a function has cognitive complexity score of 12
- **THEN** the rule is satisfied and no violation is reported

#### Scenario: Function exceeding threshold
- **GIVEN** the agent-design convention pack is loaded
- **WHEN** a function has cognitive complexity score of 18
- **THEN** a violation is reported citing AD-001 with severity HIGH

### Requirement: AD-002 Package Fan-out
The convention pack SHALL define rule AD-002 requiring that package efferent
coupling (Ce) not exceed 10. The rule SHALL specify vibe-check as the enforcement
tool. The rule SHALL note that automated threshold enforcement via a `--max-ce`
flag is a forward reference to a planned feature; until implemented, agents
enforce this rule by inspecting vibe-check JSON output or heuristically during
review. The rule SHALL include a rationale explaining that high fan-out creates
fragile packages sensitive to changes in many dependencies.

#### Scenario: Package within Ce threshold
- **GIVEN** the agent-design convention pack is loaded
- **WHEN** a package has Ce of 7
- **THEN** the rule is satisfied and no violation is reported

#### Scenario: Package exceeding Ce threshold
- **GIVEN** the agent-design convention pack is loaded
- **WHEN** a package has Ce of 13
- **THEN** a violation is reported citing AD-002 with severity HIGH

### Requirement: AD-003 Instability Threshold
The convention pack SHALL define rule AD-003 requiring that non-leaf packages
maintain an Instability (I) value below 0.7. A leaf package is one with Ca = 0
(no other packages depend on it); leaf packages are exempt because their
instability cannot propagate breaking changes downstream. The rule SHALL specify
vibe-check as the enforcement tool. The rule SHALL include a rationale explaining
that high instability in non-leaf packages propagates breaking changes downstream.

#### Scenario: Non-leaf package within threshold
- **GIVEN** the agent-design convention pack is loaded
- **WHEN** a non-leaf package (Ca > 0) has Instability of 0.55
- **THEN** the rule is satisfied and no violation is reported

#### Scenario: Non-leaf package exceeding threshold
- **GIVEN** the agent-design convention pack is loaded
- **WHEN** a non-leaf package (Ca > 0) has Instability of 0.82
- **THEN** a violation is reported citing AD-003 with severity HIGH

#### Scenario: Leaf package exemption
- **GIVEN** the agent-design convention pack is loaded
- **WHEN** a leaf package (Ca = 0, no downstream dependents) has Instability of 0.9
- **THEN** no violation is reported because leaf packages are exempt from AD-003

### Requirement: AD-004 No Circular Dependencies
The convention pack SHALL define rule AD-004 prohibiting circular dependencies
between packages. The rule SHALL specify vibe-check as the enforcement tool with
the `--no-circular-deps` flag. The rule SHALL include a rationale explaining that
circular dependencies prevent independent compilation and deployment.

#### Scenario: No circular dependencies
- **GIVEN** the agent-design convention pack is loaded
- **WHEN** vibe-check detects zero circular dependency cycles
- **THEN** the rule is satisfied and no violation is reported

#### Scenario: Circular dependency detected
- **GIVEN** the agent-design convention pack is loaded
- **WHEN** vibe-check detects a cycle (e.g., pkg/a -> pkg/b -> pkg/a)
- **THEN** a CRITICAL violation is reported citing AD-004 with the cycle path

### Requirement: AD-005 Naming Conventions
The convention pack SHALL define rule AD-005 requiring adherence to Go naming
conventions as enforced by golangci-lint with the revive linter. The rule SHALL
cover exported identifiers, package names, and interface naming.

#### Scenario: Compliant naming
- **GIVEN** the agent-design convention pack is loaded
- **WHEN** all exported identifiers follow Go naming conventions and revive reports no naming violations
- **THEN** the rule is satisfied and no violation is reported

#### Scenario: Stuttering package name
- **GIVEN** the agent-design convention pack is loaded
- **WHEN** a package `metrics` exports a type `MetricsCollector` (stuttering)
- **THEN** a violation is reported citing AD-005 with severity MEDIUM

### Requirement: AD-006 Contract Coverage
The convention pack SHALL define rule AD-006 requiring that every exported
function has at least one test exercising its contract (binary: covered or not
covered). This is not a line-coverage percentage requirement. The rule SHALL
specify gaze as the enforcement tool.

#### Scenario: Adequate contract coverage
- **GIVEN** the agent-design convention pack is loaded
- **WHEN** gaze reports that all exported functions have at least one test exercising their contract
- **THEN** the rule is satisfied and no violation is reported

#### Scenario: Missing contract coverage
- **GIVEN** the agent-design convention pack is loaded
- **WHEN** gaze reports an exported function with no test covering its contract
- **THEN** a violation is reported citing AD-006 with severity HIGH

### Requirement: AD-007 File Size
The convention pack SHALL define rule AD-007 requiring that no source file exceed
400 lines. The rule SHALL specify review agents as the enforcement mechanism. The
rule SHALL include a rationale explaining that large files impair navigability and
increase merge conflict probability.

#### Scenario: File within size limit
- **GIVEN** the agent-design convention pack is loaded
- **WHEN** a source file contains 320 lines
- **THEN** the rule is satisfied and no violation is reported

#### Scenario: File exceeding size limit
- **GIVEN** the agent-design convention pack is loaded
- **WHEN** a source file contains 485 lines
- **THEN** a violation is reported citing AD-007 with severity MEDIUM

### Requirement: AD-008 DRY No Duplication
The convention pack SHALL define rule AD-008 prohibiting duplicated blocks of 6
or more consecutive lines. Enforcement: `vibe-check --max-duplication=5` (maximum
allowed block size is 5 lines; blocks of 6+ lines are violations). The rule SHALL
note that this enforcement flag is a forward reference to a planned feature.

#### Scenario: No significant duplication
- **GIVEN** the agent-design convention pack is loaded
- **WHEN** vibe-check reports no duplicated blocks of 6 or more consecutive lines
- **THEN** the rule is satisfied and no violation is reported

#### Scenario: Duplication detected
- **GIVEN** the agent-design convention pack is loaded
- **WHEN** vibe-check reports a duplicated block of 8 lines across two files
- **THEN** a violation is reported citing AD-008 with severity MEDIUM

### Requirement: AD-009 Package Cohesion
The convention pack SHALL define rule AD-009 requiring package cohesion as
measured by LCOM4 (Lack of Cohesion of Methods, variant 4). LCOM4 counts the
number of connected components in a package's method-field graph; LCOM4 = 1 is
maximally cohesive, LCOM4 > 1 indicates the package can be split. The rule SHALL
specify vibe-check with the `--max-lcom=3` flag as the enforcement tool
(packages with LCOM4 > 3 are violations). The rule SHALL include a rationale
explaining that low cohesion indicates a package is doing too many unrelated
things.

#### Scenario: Cohesive package
- **GIVEN** the agent-design convention pack is loaded
- **WHEN** a package has LCOM4 of 1 (single connected component)
- **THEN** the rule is satisfied and no violation is reported

#### Scenario: Low cohesion package
- **GIVEN** the agent-design convention pack is loaded
- **WHEN** a package has LCOM4 of 5 (five disconnected method groups)
- **THEN** a violation is reported citing AD-009 with severity HIGH

### Requirement: AD-010 Behavior-Asserting Tests
The convention pack SHALL define rule AD-010 requiring that every test function
contains at least one behavioral assertion (a call to `t.Errorf`, `t.Fatalf`,
`t.Error`, `t.Fatal`, or a comparison check). The rule SHALL specify gaze as the
enforcement tool. The rule SHALL include a rationale explaining that tests without
meaningful assertions provide false confidence.

#### Scenario: Tests with adequate assertion depth
- **GIVEN** the agent-design convention pack is loaded
- **WHEN** gaze reports a test function contains 3 behavioral assertions
- **THEN** the rule is satisfied and no violation is reported

#### Scenario: Tests with shallow assertions
- **GIVEN** the agent-design convention pack is loaded
- **WHEN** gaze reports a test function with 0 behavioral assertions (only setup, no checks)
- **THEN** a violation is reported citing AD-010 with severity HIGH

### Requirement: Convention pack file structure
The convention pack file SHALL follow the standard convention pack format
used by the Unbound Force ecosystem. Each rule SHALL include the fields: ID,
Name, Severity, Rationale, Threshold, Enforcement (tool and command), and
Example (pass and fail cases).

#### Scenario: Pack file is loadable
- **GIVEN** the agent-design convention pack is loaded
- **WHEN** an agent loads `.opencode/uf/packs/agent-design.md`
- **THEN** the agent can parse all 10 rules (AD-001 through AD-010) with their ID, threshold, and enforcement tool

#### Scenario: Pack follows naming convention
- **GIVEN** the agent-design convention pack is loaded
- **WHEN** the pack file is created
- **THEN** it is located at `.opencode/uf/packs/agent-design.md` and a corresponding `.opencode/uf/packs/agent-design-custom.md` placeholder exists for project-level overrides containing YAML frontmatter with `pack_id: agent-design-custom` and a description of the override mechanism

## MODIFIED Requirements

None

## REMOVED Requirements

None
