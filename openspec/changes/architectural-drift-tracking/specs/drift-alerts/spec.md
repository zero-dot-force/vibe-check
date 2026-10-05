# Spec: Drift Alerts

## ADDED Requirements

### Requirement: Sustained degradation triggers drift alert

The agent SHALL generate a drift alert when a package's metric degrades across N consecutive snapshots (default N=5). Each alert SHALL include the package name, metric, observed values, trend slope, and window.

#### Scenario: Five consecutive degradations trigger alert

- **GIVEN** package `pkg/foo` Instability increases in each of 5 consecutive daily snapshots
- **WHEN** the agent evaluates drift alert thresholds
- **THEN** a drift alert is generated for `pkg/foo` Instability with the 5 values and slope

#### Scenario: Interrupted degradation resets count

- **GIVEN** package `pkg/foo` Instability increases in 3 snapshots, stabilizes for 1, then increases in 2 more
- **WHEN** the agent evaluates drift alert thresholds at N=5
- **THEN** no alert is generated (consecutive count was broken)

### Requirement: Projected threshold crossing generates predictive drift alert

The agent SHALL generate a predictive drift alert when a metric's projection crosses the standard threshold within M days (default M=30), even if the absolute value is currently within bounds.

#### Scenario: Predictive alert for near-term crossing

- **GIVEN** a package's Distance is at 0.55 with a slope of 0.03 per day
- **WHEN** the agent projects forward 30 days
- **THEN** a predictive drift alert is generated estimating the threshold crossing at day 5 (Distance reaches 0.7)

#### Scenario: No alert for distant crossing

- **GIVEN** a package's Distance is at 0.55 with a slope of 0.001 per day
- **WHEN** the agent projects forward 30 days
- **THEN** no predictive alert is generated (crossing is beyond the M-day window)

### Requirement: Drift alert thresholds are independently configurable

The consecutive-degradation count (N) and projection window (M) SHALL be configurable in the agent prompt. Changes to these values SHALL NOT require changes to the Go binary or ModuleGraph schema.

#### Scenario: Agent uses custom N and M values

- **GIVEN** the `mx-f-architecture-trend.md` agent prompt configures N=3 and M=14
- **WHEN** a package degrades across 3 consecutive snapshots
- **THEN** a drift alert is generated (lowered N threshold takes effect)

### Requirement: Drift alerts include actionable context

Each drift alert SHALL include the package name, metric name, current value, trend direction and slope, the affected time window, and a recommendation to investigate recent changes to the package.

#### Scenario: Alert contains required fields

- **GIVEN** a drift alert is generated for `pkg/bar` Distance across 5 snapshots
- **WHEN** the alert is rendered
- **THEN** the output includes package name, metric name, current value, slope, time window, and an investigation recommendation