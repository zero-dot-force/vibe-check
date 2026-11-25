## ADDED Requirements

### Requirement: Agent reports SHALL use mandatory emoji section markers

The vibe-check-reporter agent SHALL prefix every major section header with a designated emoji from the closed vocabulary:

| Emoji | Section |
|-------|---------|
| 🏗️ | Report title |
| 📊 | Metrics table |
| 🔗 | Coupling analysis |
| 🧩 | Cohesion analysis |
| 🔄 | Cycle detection |
| 📋 | Duplications |
| 📖 | Legend block |
| 🏥 | Health scorecard |
| 🟢 | Good/healthy severity |
| 🟡 | Moderate/warning severity |
| 🔴 | Critical/danger severity |
| ⚠️ | Warning callout |

No other emojis SHALL appear in report output.

#### Scenario: Summary mode uses emoji title

- **WHEN** the agent produces a summary mode report
- **THEN** the title line begins with `🏗️`

#### Scenario: Detailed mode uses all applicable section emojis

- **WHEN** the agent produces a detailed mode report with coupling, cohesion, and duplication data
- **THEN** the output contains `📊`, `🔗`, `🧩`, `📋`, and `🏥` section headers

#### Scenario: Warning callouts use ⚠️

- **WHEN** the agent emits a warning (e.g., "analysis timed out")
- **THEN** the warning text is prefixed with `⚠️`

### Requirement: Agent reports SHALL use full metric names in tables

Table column headers SHALL use human-readable names: `Instability`, `Abstractness`, `Distance`, `LCOM4`, `Ca`, `Ce`. Single-letter abbreviations (I, A, D) SHALL NOT appear as column headers.

#### Scenario: Detailed mode table uses full names

- **WHEN** the agent produces a detailed mode per-package table
- **THEN** the columns include `Instability`, `Abstractness`, and `Distance` (not `I`, `A`, `D`)

### Requirement: Every report SHALL include a legend block

After the title and metadata, every report SHALL include a `📖` legend block that spells out each column header and links to the metrics guide.

#### Scenario: Legend explains all column headers

- **WHEN** the agent produces any report
- **THEN** the output contains a `📖` line defining Instability, Abstractness, Distance, and LCOM4 in one sentence each, followed by the guide URL

### Requirement: Detailed mode SHALL include a Health Scorecard

The Health Scorecard SHALL grade each architectural dimension (Avg Distance, Max LCOM4, Cycle Count, Duplication, Instability Spread) with a letter grade (A, B+, B, C+, C, F) and severity emoji, using the thresholds defined in the design D2 grade table.

#### Scenario: All-green scorecard for healthy project

- **WHEN** the project has D < 0.1, LCOM4 = 1, 0 cycles, 0% duplication
- **THEN** all scorecard entries show 🟢 with A or B+ grades

#### Scenario: Mixed scorecard for typical project

- **WHEN** the project has D < 0.3, LCOM4 = 4, 0 cycles, 5% duplication
- **THEN** the scorecard shows mixed 🟢 and 🟡 grades with appropriate letter grades

#### Scenario: Red scorecard for problematic project

- **WHEN** the project has cycles ≥ 1 or D ≥ 0.5
- **THEN** at least one scorecard entry shows 🔴 with C- or below

### Requirement: Detailed mode SHALL include Prioritized Recommendations

The agent SHALL produce a numbered list of 1–5 recommendations, each prefixed with a severity emoji and starting with an action verb.

#### Scenario: Recommendations include specific packages and metrics

- **WHEN** a package has LCOM4 ≥ 4
- **THEN** a 🔴 recommendation names that package, cites the LCOM4 value, and suggests a corrective action

#### Scenario: Maximum 5 recommendations

- **WHEN** more than 5 issues are detected
- **THEN** only the top 5 highest-severity issues appear as recommendations

### Requirement: Zones SHALL be visualized with emojis

Zone labels in reports SHALL use emoji prefixes: 🟢 Main Sequence, 🟡 Balanced, 🔴 Zone of Pain, 🔴 Zone of Uselessness.

#### Scenario: Zone column in detailed table uses emojis

- **WHEN** the agent renders a per-package table in detailed mode
- **THEN** the Zone column shows emoji-prefixed labels (e.g., `🟢 Main Sequence`)

### Requirement: Duplication blocks SHALL be surfaced in a dedicated section

When the JSON output contains `duplications` arrays, the agent SHALL render a `📋` section showing file locations, line ranges, and similarity percentages.

#### Scenario: Duplication section appears when data exists

- **WHEN** the JSON output has non-empty `duplications` for any package
- **THEN** the report includes a `📋` section with file paths and similarity percentages

#### Scenario: Duplication section is omitted when no duplicates exist

- **WHEN** the JSON output has empty or null `duplications` for all packages
- **THEN** no `📋` section appears in the report

### Requirement: Report tone SHALL be conversational and data-driven

The agent SHALL NOT include pedagogical explanations of metrics, filler paragraphs, slang, puns on metric names, or excessive exclamation marks. Every sentence SHALL convey data or an actionable observation.

#### Scenario: No metric definitions in body text

- **WHEN** the agent presents Instability values
- **THEN** the body text does not explain what Instability means (that is the legend's job)

### Requirement: Scaffold copy SHALL be synchronized

The file `internal/scaffold/assets/agents/vibe-check-reporter.md` SHALL contain the same content as `.opencode/agents/vibe-check-reporter.md` after this change is implemented.

#### Scenario: Init produces the new agent format

- **WHEN** `vibe-check init` is run in a project after this change
- **THEN** the deployed `.opencode/agents/vibe-check-reporter.md` file uses emoji section markers, full metric names, and the legend block