## ADDED Requirements

### Requirement: Rule reference documents all 10 AD rules

The documentation SHALL provide a complete rule reference table covering all 10 agent-design rules (AD-001 through AD-010). Each entry SHALL include the rule ID, descriptive name, threshold value, severity level (CRITICAL/HIGH/MEDIUM), enforcement tool, and rationale for why the rule matters.

#### Scenario: All 10 rules are present

- **WHEN** reviewing the rule reference table
- **THEN** exactly 10 rules (AD-001 through AD-010) are listed with no omissions

#### Scenario: Each rule has complete metadata

- **WHEN** examining any rule entry in the table
- **THEN** the entry contains rule ID, name, threshold, severity level, enforcement tool, and rationale

### Requirement: Enforcement tool mapping is documented

The documentation SHALL include a cross-reference matrix mapping each enforcement tool (vibe-check, gaze, golangci-lint, review agents) to the rules it enforces.

#### Scenario: vibe-check rules are listed

- **WHEN** reviewing the enforcement tool matrix
- **THEN** vibe-check is shown as the enforcement tool for AD-003 (I < 0.7 for non-leaf packages — the numeric threshold; leaf packages with Ca = 0 are exempt, enforced by review agents inspecting Ca in JSON output), AD-004 (no cycles), and AD-009 (LCOM4 <= 3). AD-002 (Ce < 10) and AD-008 (no duplication) are shown as forward references — planned vibe-check features not yet implemented, currently enforced heuristically by agents.

#### Scenario: gaze rules are listed

- **WHEN** reviewing the enforcement tool matrix
- **THEN** gaze is shown as the enforcement tool for AD-001 (cognitive complexity < 15), AD-006 (contract coverage), and AD-010 (behavior-asserting tests)

#### Scenario: Other enforcers are listed

- **WHEN** reviewing the enforcement tool matrix
- **THEN** golangci-lint is shown for AD-005 (naming) and review agents are shown for AD-007 (file size <= 400 lines)

### Requirement: Override mechanism is documented

The documentation SHALL describe how to use `agent-design-custom.md` to override thresholds per-project, including the CR-NNN prefix convention, scoped overrides, and the requirement for justification.

#### Scenario: CR-NNN prefix is documented

- **WHEN** reading the override mechanism section
- **THEN** the CR-NNN naming convention is explained with an example showing how to scope a threshold override to a specific package

#### Scenario: Justification requirement is documented

- **WHEN** reading the override mechanism section
- **THEN** the documentation states that each override must include justification explaining why the threshold is being modified for this project

### Requirement: Forward references are documented

The documentation SHALL note that AD-002 (`--max-ce`) and AD-008 (`--max-duplication`) reference planned vibe-check features that are not yet implemented. The documentation SHALL state that agents enforce these rules heuristically until the tooling ships.

#### Scenario: AD-002 forward reference is noted

- **WHEN** reading AD-002 in the rule reference
- **THEN** a note indicates `--max-ce` is a planned flag and agents currently enforce the Ce < 10 threshold heuristically

#### Scenario: AD-008 forward reference is noted

- **WHEN** reading AD-008 in the rule reference
- **THEN** a note indicates `--max-duplication` is a planned flag and agents currently enforce the duplication threshold heuristically

### Requirement: LCOM4 semantics are documented

The documentation SHALL clarify that AD-009 uses LCOM4 integer semantics (connected-component count, per Hitz & Montazeri 1995), not a 0.0–1.0 cohesion float.

#### Scenario: LCOM4 semantics are explained

- **WHEN** reading AD-009 in the rule reference
- **THEN** the documentation states that LCOM4 = 1 is maximally cohesive, LCOM4 > 1 indicates the package can be split, and the value represents connected-component count, not a float ratio

### Requirement: Per-rule detail sections are provided

The documentation SHALL include a per-rule detail section for each AD rule (AD-001 through AD-010) covering the rule's rationale, compliance examples (PASS and FAIL), and tool-specific enforcement flags.

#### Scenario: Each rule has rationale and examples

- **GIVEN** the agent-design reference documentation exists
- **WHEN** reading any AD rule detail section
- **THEN** the section includes the rule's rationale, a PASS compliance example, and a FAIL compliance example

#### Scenario: Tool-specific flags are documented

- **GIVEN** the agent-design reference documentation exists
- **WHEN** reading any AD rule detail section
- **THEN** the section references the specific CLI flag or tool configuration used to enforce the rule

### Requirement: Introduction and purpose are documented

The documentation SHALL begin with an introduction section explaining the document's purpose and target audience (contributors and agent maintainers).

#### Scenario: Document starts with introduction

- **GIVEN** the agent-design reference documentation exists
- **WHEN** reading the top of the document
- **THEN** an introduction section states the document's purpose and identifies the target audience

### Requirement: Forward reference tracking issues are linked

The documentation SHALL include links to tracking issues for forward-reference features where applicable.

#### Scenario: AD-008 tracking issue is linked

- **GIVEN** the agent-design reference documentation exists
- **WHEN** reading the AD-008 forward reference note
- **THEN** a link to tracking issue `zero-dot-force/vibe-check#13` for `--max-duplication` is provided

#### Scenario: AD-002 tracking issue status is noted

- **GIVEN** the agent-design reference documentation exists
- **WHEN** reading the AD-002 forward reference note
- **THEN** the documentation notes the presence or absence of a tracking issue for `--max-ce`

### Requirement: Pack-file authority is stated

The documentation SHALL state that `.opencode/uf/packs/agent-design.md` is the authoritative source and the documentation is a derivative reference.

#### Scenario: Authority statement is present

- **GIVEN** the agent-design reference documentation exists
- **WHEN** reading the document
- **THEN** a statement identifies the pack file as authoritative and the documentation as a derivative reference