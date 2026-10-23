# Spec Delta

## Purpose

The vibe-check-reporter agent's post-change behavior: emit the Go-rendered report verbatim as the skeleton, and add LLM value only for violation explanations, remediation suggestions, and trending analysis.

## ADDED Requirements

### Requirement: Report skeleton from the CLI
For summary and detailed modes the agent SHALL produce the report skeleton by running `vibe-check analyze` with `--output <file>` on the project directory and then `vibe-check report --mode <mode>` on that file, emitting the command's markdown output verbatim. The temporary snapshot file SHALL be removed after the report is rendered.

#### Scenario: Summary invocation
- **WHEN** the agent is invoked in summary mode
- **THEN** its output contains, byte-identical, the `vibe-check report --mode summary` markdown for the analyzed snapshot

#### Scenario: Detailed invocation
- **WHEN** the agent is invoked in detailed mode
- **THEN** its output contains the `vibe-check report --mode detailed` markdown, followed by the agent's interpretive sections

#### Scenario: Binary missing
- **WHEN** `vibe-check` is not available on PATH
- **THEN** the agent reports the tooling failure and does not hand-render metric numbers

### Requirement: Violation explanations
When the report command signals violations, the agent SHALL append, for each violation, an explanation of why the threshold failed and a remediation suggestion grounded in the repository's code and AGENTS.md.

#### Scenario: LCOM violation
- **WHEN** the report exits 1 with an LCOM4 split recommendation
- **THEN** the agent explains which package exceeds LCOM4, why, and suggests concrete split candidates referencing actual code in the repository

#### Scenario: Clean report
- **WHEN** the report exits 0
- **THEN** the agent adds no violation explanations and may add only brief interpretation

### Requirement: No in-prompt metric arithmetic
The agent MUST NOT compute metric arithmetic in-prompt; all metric values it emits SHALL come from CLI output. Conformance is verified by inspection of the deployed asset and comparison of agent output against CLI output, not by runtime enforcement.

#### Scenario: Values quoted
- **WHEN** the agent discusses a metric value
- **THEN** the value matches the CLI-emitted report for the same snapshot

### Requirement: Directory-only analysis invocation
The agent SHALL pass a directory path to `vibe-check analyze`, never package patterns, because analyze path validation rejects patterns.

#### Scenario: Invocation shape
- **WHEN** the agent analyzes the repository
- **THEN** it invokes `vibe-check analyze` with a directory argument and `--output <file>`, not a package pattern

### Requirement: Trending mode remains agent-owned
Trending analysis SHALL remain computed by the agent from historical snapshots; the CLI renders no trending mode.

#### Scenario: Trending invocation
- **WHEN** the agent is invoked in trending mode
- **THEN** it renders the trending report itself from prior snapshots, without invoking a trending report mode

### Requirement: Bash allowlist
The agent's bash permissions SHALL allow exactly `vibe-check analyze *`, `vibe-check report *`, and `git rev-parse *`, with all other commands denied.

#### Scenario: Allowlist contents
- **WHEN** the deployed agent asset's frontmatter is inspected
- **THEN** the bash permission block contains exactly those three allow entries and a deny-all catch-all
