## ADDED Requirements

### Requirement: Metrics guide exists at docs/metrics-guide.md

The repository SHALL contain a `docs/metrics-guide.md` file that explains each architectural metric in plain language suitable for a developer unfamiliar with Martin design-quality metrics.

The guide SHALL cover every metric emitted by `vibe-check analyze`: Instability (I), Abstractness (A), Distance from Main Sequence (D), Lack of Cohesion of Methods 4 (LCOM4), Afferent Coupling (Ca), Efferent Coupling (Ce), Circular Dependencies, and Code Duplication.

#### Scenario: Guide is discoverable via GitHub URL

- **WHEN** a developer opens `https://github.com/zero-dot-force/vibe-check/blob/main/docs/metrics-guide.md`
- **THEN** the guide renders with all metric sections visible

#### Scenario: Guide covers all emitted metrics

- **WHEN** the guide is compared against the `vibe-check analyze` JSON schema fields
- **THEN** every metric field (instability, abstractness, distance, lcom, cycles, duplications) has a corresponding explanation section

### Requirement: Each metric explanation follows a consistent ELI5 format

For each metric, the guide SHALL include:

1. **What it is**: A single-sentence plain-language definition
2. **Scale**: The range and what the extremes mean
3. **Why care**: One sentence on practical impact to the developer
4. **Emoji ranges**: Severity thresholds with emoji indicators

#### Scenario: Instability section is self-contained

- **WHEN** a developer reads only the Instability section
- **THEN** they understand the 0.0–1.0 scale, that 0.0 means "everyone depends on me" and 1.0 means "I depend on everyone," and what ranges are healthy

#### Scenario: LCOM4 section is self-contained

- **WHEN** a developer reads only the LCOM4 section
- **THEN** they understand the 1–N scale, that 1 means perfectly cohesive, and that ≥4 suggests the package should be split

### Requirement: Guide links to external references

The guide SHALL include at least one link to an external source for the Martin metrics methodology (e.g., the original "Design Principles and Design Patterns" paper or the "OO Design Quality Metrics" paper).

#### Scenario: External references are present

- **WHEN** the guide is rendered
- **THEN** at least one external URL to Martin metrics documentation is present and accessible

### Requirement: Vibe-check reporter agent links to the guide

Every report produced by the vibe-check-reporter agent SHALL include a hyperlink to the metrics guide.

#### Scenario: Summary mode report includes guide link

- **WHEN** the agent produces a summary mode report
- **THEN** the output contains a hyperlink to `https://github.com/zero-dot-force/vibe-check/blob/main/docs/metrics-guide.md`

#### Scenario: Detailed mode report includes guide link

- **WHEN** the agent produces a detailed mode report
- **THEN** the output contains a hyperlink to `https://github.com/zero-dot-force/vibe-check/blob/main/docs/metrics-guide.md`