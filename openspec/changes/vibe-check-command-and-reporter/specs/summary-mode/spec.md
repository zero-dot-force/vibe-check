# Spec: Summary Mode

## ADDED Requirements

### Requirement: Default mode is summary

The `/vibe-check` command SHALL default to summary mode when invoked
with no mode argument or with the explicit `summary` argument.

#### Scenario: No arguments

- **GIVEN** the `vibe-check` binary is installed and in PATH
- **WHEN** user invokes `/vibe-check` with no arguments
- **THEN** the agent runs `vibe-check analyze --output <tempfile> ./...`
  and presents results in summary format

#### Scenario: Explicit summary argument

- **GIVEN** the `vibe-check` binary is installed and in PATH
- **WHEN** user invokes `/vibe-check summary`
- **THEN** the agent produces the same output as with no arguments

#### Scenario: Custom package pattern

- **GIVEN** the `vibe-check` binary is installed and in PATH
- **WHEN** user invokes `/vibe-check summary ./internal/...`
- **THEN** the agent validates the package pattern against the safe
  character set (`^[A-Za-z0-9./_-]+$`) and runs analysis on the
  specified package pattern instead of `./...`

### Requirement: Traffic-light health indicator

The summary mode SHALL present an overall health indicator using a
traffic-light metaphor (green/yellow/red) based on the analysis
verdict.

#### Scenario: All metrics within thresholds

- **GIVEN** `vibe-check analyze` has been invoked successfully
- **WHEN** the process exits with code 0 (no threshold violations)
- **THEN** the agent displays a GREEN health indicator with a message
  indicating the module is architecturally healthy

#### Scenario: Threshold violations detected

- **GIVEN** `vibe-check analyze` has been invoked successfully
- **WHEN** the process exits with code 1 (threshold violations)
- **THEN** the agent displays a RED health indicator with a count of
  packages exceeding thresholds

#### Scenario: Analysis error

- **GIVEN** `vibe-check analyze` has been invoked
- **WHEN** the process exits with code 2 (error)
- **THEN** the agent displays the error and does not produce a
  traffic-light indicator

### Requirement: Summary includes top-level metric aggregates

The summary mode SHALL include aggregate instability, distance, and
LCOM4 statistics across all analyzed packages.

#### Scenario: Module with multiple packages

- **GIVEN** `vibe-check analyze` completes with exit code 0 or 1
- **WHEN** the ModuleGraph contains multiple packages
- **THEN** the summary includes the count of packages analyzed, the
  range (min–max) of instability and distance values, and the count of
  packages with LCOM4 above 1

#### Scenario: Circular dependencies present

- **GIVEN** `vibe-check analyze` completes with exit code 0 or 1
- **WHEN** the ModuleGraph contains circular dependencies
- **THEN** the summary includes the number of cycles detected and
  lists the packages involved

### Requirement: Error handling for subprocess failures

The agent SHALL handle subprocess failure modes gracefully and provide
actionable guidance to the user.

#### Scenario: vibe-check binary not found

- **GIVEN** the `vibe-check` binary is not installed or not in PATH
- **WHEN** the agent attempts to run `vibe-check analyze`
- **THEN** the agent reports that vibe-check needs to be installed and
  provides installation guidance (e.g., `go install` command)

#### Scenario: Malformed package pattern

- **GIVEN** the user provides a package pattern containing shell
  metacharacters (e.g., `; rm -rf /` or `$(whoami)`)
- **WHEN** the agent validates the input
- **THEN** the agent rejects the pattern with a clear error message
  listing the valid character set and does NOT pass it to bash

#### Scenario: Unrecognized mode argument

- **GIVEN** the user invokes `/vibe-check` with an argument
- **WHEN** the argument does not match `summary`, `detailed`, or
  `trending` and is not a valid package pattern
- **THEN** the agent reports the unrecognized mode and lists the
  available modes (summary, detailed, trending)
