# Spec: Natural Language Interpretation

## ADDED Requirements

### Requirement: Agent translates metrics into plain-English assessments

The agent SHALL translate raw numeric metric values into natural-language
explanations that a developer unfamiliar with Martin metrics can
understand.

#### Scenario: High instability explanation

- **GIVEN** `vibe-check analyze` has produced a ModuleGraph
- **WHEN** a package has instability > 0.7
- **THEN** the agent explains in plain English that the package depends
  on many other packages but few packages depend on it, making it
  sensitive to upstream changes (e.g., "Package X has high instability
  (0.89) — it depends on many packages but nothing depends on it")

#### Scenario: Low distance explanation

- **GIVEN** `vibe-check analyze` has produced a ModuleGraph
- **WHEN** a package has distance < 0.1
- **THEN** the agent explains that the package sits near the ideal
  balance between stability and abstractness on the main sequence

#### Scenario: High LCOM4 explanation

- **GIVEN** `vibe-check analyze` has produced a ModuleGraph
- **WHEN** a package has LCOM4 > 1
- **THEN** the agent explains that the package has multiple disconnected
  groups of methods/types that do not share state, suggesting it bundles
  unrelated responsibilities

### Requirement: Zone classification with guidance

The agent SHALL explain each package's zone classification with
actionable guidance appropriate to that zone.

#### Scenario: Zone of Pain guidance

- **GIVEN** `vibe-check analyze` has produced a ModuleGraph
- **WHEN** a package is in the Zone of Pain (high stability, low
  abstractness)
- **THEN** the agent explains that the package is concrete and heavily
  depended upon, and suggests introducing interfaces to increase
  abstractness and move toward the main sequence

#### Scenario: Zone of Uselessness guidance

- **GIVEN** `vibe-check analyze` has produced a ModuleGraph
- **WHEN** a package is in the Zone of Uselessness (high instability,
  high abstractness)
- **THEN** the agent explains that the package defines abstractions
  with few concrete dependents and suggests evaluating whether the
  abstractions serve a real need

#### Scenario: Main Sequence guidance

- **GIVEN** `vibe-check analyze` has produced a ModuleGraph
- **WHEN** a package is near the main sequence (distance < 0.1)
- **THEN** the agent confirms the package has a healthy balance and
  does not flag it for action

### Requirement: Circular dependency explanation

The agent SHALL explain circular dependencies in terms a developer
can act on.

#### Scenario: Cycle with two packages

- **GIVEN** `vibe-check analyze` has produced a ModuleGraph with cycles
- **WHEN** a circular dependency involves packages A and B
- **THEN** the agent explains which packages form the cycle, why
  cycles are problematic (compilation order, testability, deployment
  coupling), and suggests strategies to break the cycle (interface
  extraction, dependency inversion, package merging)
