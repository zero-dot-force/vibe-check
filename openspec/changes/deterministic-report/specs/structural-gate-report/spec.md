# Spec Delta

## Purpose

Publishes the structural-entropy diff report as a pull-request step summary in the structural-gate CI workflow, without affecting the gate verdict or expanding the trusted-execution surface.

## ADDED Requirements

### Requirement: Step summary rendering
The workflow SHALL append the `vibe-check diff --markdown` output for the PR's base-vs-head snapshots to `$GITHUB_STEP_SUMMARY`, in a step that runs even when the gate step fails.

#### Scenario: Failing gate
- **WHEN** the gate step exits non-zero
- **THEN** the step summary still contains the markdown diff report

#### Scenario: Passing gate
- **WHEN** the gate step passes
- **THEN** the step summary contains the markdown diff report

#### Scenario: Analysis failure
- **WHEN** the analysis step failed and the base or head snapshot is missing
- **THEN** the render step's non-zero exit is non-blocking, no report is appended to the step summary, and the gate job's failure remains the authoritative conclusion

### Requirement: Fork guard preserved
The render step SHALL only run within the existing fork-guarded job; pull requests whose head comes from a fork skip the entire job, including the render step.

#### Scenario: Fork PR
- **WHEN** a pull request head repository differs from the target repository
- **THEN** no analysis, gate, or render step runs for that pull request

### Requirement: Gate independence
The render step SHALL NOT modify gate thresholds, gate inputs, or the gate step's conclusion; the gate step remains the authoritative check result.

#### Scenario: Render failure
- **WHEN** the render step itself fails
- **THEN** the gate step's result is unchanged and remains the authoritative conclusion for the check
