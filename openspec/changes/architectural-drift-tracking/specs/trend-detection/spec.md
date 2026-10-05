# Spec: Trend Detection

## ADDED Requirements

### Requirement: Trend detection computes linear regression over time windows

The `mx-f-architecture-trend.md` agent SHALL compute linear regression for each tracked metric (Instability, Distance, LCOM4) per package over 7-day, 30-day, and 90-day windows. Classification thresholds: slope absolute value ≥ 0.01 per snapshot for Instability/Distance, ≥ 0.5 per snapshot for LCOM4.

#### Scenario: Degrading trend detected

- **GIVEN** a package's Instability values are `[0.30, 0.35, 0.39, 0.43, 0.48]` over 5 consecutive daily snapshots
- **WHEN** the agent computes linear regression over a 7-day window
- **THEN** the trend is classified as "degrading" with slope ≥ 0.01 and R² ≥ 0.5

#### Scenario: Improving trend detected

- **GIVEN** a package's Distance values are `[0.50, 0.45, 0.41, 0.38, 0.34]` over 5 consecutive daily snapshots
- **WHEN** the agent computes linear regression over a 7-day window
- **THEN** the trend is classified as "improving" with negative slope and R² ≥ 0.5

#### Scenario: Noisy data classified as stable

- **GIVEN** a package's Instability values are `[0.50, 0.52, 0.47, 0.53, 0.49]` over 5 consecutive daily snapshots
- **WHEN** the agent computes linear regression over a 7-day window
- **THEN** the trend is classified as "stable" because R² < 0.5 (insufficient confidence)

### Requirement: Insufficient snapshots produce no trend classification

The agent SHALL require at minimum 3 snapshots in a time window before classifying a trend. Fewer than 3 snapshots SHALL produce "insufficient data" for that window.

#### Scenario: Two snapshots in 7-day window

- **GIVEN** only 2 snapshots exist for the module within the last 7 days
- **WHEN** the agent computes trends over a 7-day window
- **THEN** the output for the 7-day window is "insufficient data"

#### Scenario: Three snapshots enable classification

- **GIVEN** exactly 3 snapshots exist for the module within the last 7 days
- **WHEN** the agent computes trends with sufficient signal
- **THEN** the 7-day window produces a trend classification (improving/degrading/stable)

### Requirement: Trend output is package-scoped and metric-scoped

Trend classification SHALL be computed independently per package and per metric. A package may show "improving" for Distance while "degrading" for LCOM4 simultaneously.

#### Scenario: Mixed trends per package

- **GIVEN** package `pkg/foo` has improving Distance and degrading LCOM4 across 5 snapshots
- **WHEN** the agent generates the trend report
- **THEN** the output shows both trends independently with their respective classifications and slopes

### Requirement: Projected threshold crossing is included in trend output

When a degrading trend's linear projection crosses the standard threshold (Instability > 0.75, Distance > 0.7, LCOM4 > 5; defaults from the entropy gate thresholds in `metrics/verdict.go`) within the next 30 days, the agent SHALL include the estimated crossing date in the trend output.

#### Scenario: Projected crossing within window

- **GIVEN** a package's Instability is at 0.65 with a slope of 0.02 per day
- **WHEN** the agent projects the trend forward
- **THEN** the output includes a projected threshold crossing at day 5 (when Instability reaches 0.75) and flags it as a drift alert