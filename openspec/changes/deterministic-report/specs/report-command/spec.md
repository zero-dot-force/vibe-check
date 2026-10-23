# Spec Delta

## Purpose

Deterministic markdown rendering of a ModuleGraph JSON snapshot, so CI, agents, and humans share one byte-stable architectural health report with grades and recommendations.

## ADDED Requirements

### Requirement: Render report from a ModuleGraph snapshot
The `vibe-check report <graph.json>` command SHALL read a ModuleGraph JSON file, validate it against the embedded ModuleGraph JSON schema (accepting the schema versions `vibe-check analyze` emits), and render a markdown report to stdout.

#### Scenario: Valid snapshot
- **WHEN** `vibe-check report graph.json` runs against a schema-valid snapshot with no threshold violations
- **THEN** a markdown report is written to stdout and the command exits 0

#### Scenario: Unreadable input
- **WHEN** the input file does not exist or cannot be read
- **THEN** an error is written to stderr, nothing is written to stdout, and the command exits 2

#### Scenario: Schema-invalid input
- **WHEN** the input file is not a schema-valid ModuleGraph
- **THEN** an error is written to stderr, nothing is written to stdout, and the command exits 2

### Requirement: Report modes
The command SHALL offer `--mode summary` (default) and `--mode detailed`. Summary mode SHALL render the report header, metric legend, status summary line, and prioritized recommendations. Detailed mode SHALL additionally render the package metrics table, coupling summary, cohesion summary, circular dependency chains, duplication table, health scorecard, and warnings.

#### Scenario: Default mode
- **WHEN** `vibe-check report graph.json` runs without `--mode`
- **THEN** output contains the header, legend, status summary line, and recommendations, and no package metrics table

#### Scenario: Detailed mode
- **WHEN** `--mode detailed` is passed
- **THEN** output contains the package metrics table with full-word column headers and a zone column, plus the coupling, cohesion, cycles, duplication, scorecard, and warnings sections, in addition to the summary-mode contents

#### Scenario: Invalid mode
- **WHEN** `--mode trending` is passed
- **THEN** the command exits 2 with an error, because trending is agent-owned and not a report mode

### Requirement: JSON output mode
The command SHALL offer `--json`, mutually exclusive with `--mode`, that writes the computed report model — aggregates, per-dimension grades, overall status, and fired recommendations with severity and rule — as a single JSON object to stdout, so human-readable markdown is not the only available format. Recommendation severity serializes as the machine-readable strings `red`, `yellow`, and `green` (not emoji). The exit-code contract applies unchanged under `--json`.

#### Scenario: JSON invocation
- **WHEN** `vibe-check report --json graph.json` runs on a snapshot with violations
- **THEN** a JSON object containing the aggregates, grades, overall status, and fired recommendations is written to stdout and the command exits 1

#### Scenario: Flag mutual exclusion
- **WHEN** both `--json` and `--mode detailed` are passed
- **THEN** the command exits 2 with a usage error and writes no report to stdout

### Requirement: Report header
The report SHALL open with the 🏗️ title, the producer name and version, the package count, and total analyzed lines, and SHALL NOT include a generation timestamp. The producer name and version SHALL come from the snapshot's `provenance` object when present; when the snapshot omits provenance, the header SHALL fall back to the rendering binary's name and embedded version.

#### Scenario: Header contents
- **WHEN** any report is rendered
- **THEN** the first heading contains the report title, the header line contains the producer version and package count, and no timestamp appears anywhere in the output

### Requirement: Status summary line
Summary mode SHALL render one line containing the overall status word ("All Clear" when no violation-severity recommendations fire, otherwise "Needs Attention"), the package count, average instability, average distance, max LCOM4, and cycle count.

#### Scenario: Clean graph
- **WHEN** no red or yellow recommendation rules fire
- **THEN** the summary line begins with "All Clear" and lists package count, average instability, average distance, max LCOM4, and cycle count

#### Scenario: Violations present
- **WHEN** any red or yellow recommendation rule fires
- **THEN** the summary line begins with "Needs Attention"

### Requirement: Health scorecard grading
The report SHALL grade five dimensions on an A–F scale using fixed thresholds: Avg Distance <0.1 A, <0.2 B+, <0.3 B, <0.5 C+, <0.7 C, else F; Max LCOM4 1 A, 2 B+, 3 B, 4 C+, 5 C, ≥6 F; Cycles 0 A, ≥1 F; Duplication 0% A, <3% B+, <5% B, <10% C+, <15% C, ≥15% F; Instability Spread 0 A, 1 B, 2 C+, ≥3 F.

#### Scenario: Distance boundary
- **WHEN** average distance is exactly 0.1
- **THEN** the Avg Distance grade is B+, because A requires strictly less than 0.1

#### Scenario: LCOM boundary
- **WHEN** max LCOM4 is 4
- **THEN** the Max LCOM4 grade is C+

#### Scenario: Any cycle fails
- **WHEN** the snapshot contains one circular dependency
- **THEN** the Cycles grade is F

#### Scenario: Duplication boundary
- **WHEN** overall duplication is exactly 5%
- **THEN** the Duplication grade is C+, because B requires strictly less than 5%

### Requirement: Grade status emoji
Grades SHALL map to status emoji: A and B+ → 🟢, B and C+ → 🟡, C and F → 🔴. The scorecard SHALL render an overall status derived from the worst dimension grade.

#### Scenario: Mixed grades
- **WHEN** dimension grades are A, B+, B, C+, and F
- **THEN** the scorecard shows 🟢, 🟢, 🟡, 🟡, 🔴 respectively, and the overall status is 🔴

### Requirement: Aggregate computations
The report SHALL compute average instability (mean of module instability values), average distance (mean of module distance values), max LCOM4, cycle count, and instability spread (count of modules with instability ≤ 0.001 or ≥ 0.999).

#### Scenario: Spread counting
- **WHEN** modules have instability 0.0, 1.0, and 0.5
- **THEN** the instability spread is 2 extremes

#### Scenario: Empty module set
- **WHEN** the snapshot contains zero modules
- **THEN** aggregates render as zero, no violation-severity recommendations fire, and the command exits 0

### Requirement: Overall duplication percentage
The report SHALL compute overall duplication as deduplicated duplicated lines divided by the sum of all module total lines, times 100, rendered with one decimal place.

#### Scenario: Single duplicated block
- **WHEN** one module has one 16-line duplication block and total lines across all modules are 3797
- **THEN** overall duplication renders as 0.4%

### Requirement: Recommendation rules
The report SHALL generate templated recommendations from a fixed rule table: LCOM4 ≥ 4, distance ≥ 0.5, any cycle, and duplication ≥ 5% produce 🔴 rules; instability ≥ 0.95, instability ≤ 0.05, Ce > 20, and abstractness 0 with more than 10 exported types produce 🟡 rules; duplication strictly between 0% and 5% produces a 🟢 watch rule. When no rule of any severity fires, a single 🟢 all-clear recommendation SHALL be emitted.

#### Scenario: LCOM split fires
- **WHEN** a module has LCOM4 4
- **THEN** a 🔴 recommendation for that module advises splitting it

#### Scenario: Stable package protection
- **WHEN** a module has instability 0.05 or lower
- **THEN** a 🟡 recommendation advises protecting it as a stable foundation

#### Scenario: All clear
- **WHEN** no rule of any severity fires for any module
- **THEN** exactly one recommendation, the 🟢 all-clear, is emitted

#### Scenario: Watch rule only
- **WHEN** only the 🟢 watch rule fires (duplication strictly between 0% and 5%)
- **THEN** the watch rule is emitted and the all-clear is not

### Requirement: Recommendation ordering
Recommendations SHALL be ordered by severity (🔴 before 🟡 before 🟢), then by rule-specific metric extremity, then by package path ascending, and capped at 5 entries.

#### Scenario: Severity precedence
- **WHEN** a 🔴 cycle rule and a 🟡 decouple rule fire
- **THEN** the cycle recommendation renders first

#### Scenario: Deterministic tiebreak
- **WHEN** two modules fire the same rule with equal metric values
- **THEN** the module with the lexicographically smaller path renders first

#### Scenario: Rule cap
- **WHEN** more than 5 recommendations fire
- **THEN** at most 5 are rendered, selected by the ordering above

### Requirement: Byte-stable output
Report output SHALL be byte-identical for identical input files regardless of run time. Modules SHALL be rendered in path order. The `--json` output SHALL also be byte-identical across runs, with stable field order.

#### Scenario: Repeated runs

- **WHEN** `vibe-check report` runs twice on the same file, in markdown and in `--json` mode
- **THEN** stdout is byte-identical between runs
#### Scenario: Input order independence
- **WHEN** the modules array is reordered within the JSON
- **THEN** the rendered output is unchanged

### Requirement: Exit codes
The command SHALL exit 0 when no red or yellow recommendation fires, 1 when any red or yellow recommendation fires, and 2 on tool failure.

#### Scenario: Violations
- **WHEN** the report renders at least one 🔴 or 🟡 recommendation
- **THEN** the command exits 1 while still writing the full report to stdout

#### Scenario: Clean
- **WHEN** only 🟢 recommendations fire
- **THEN** the command exits 0

### Requirement: Fixed formatting vocabulary
The report SHALL use a closed emoji set (🏗️ 📖 📊 🔗 🧩 🔄 📋 🏥 🟢 🟡 🔴 ⚠️) for section markers and severity, and SHALL include a fixed legend defining Instability, Abstractness, Distance, LCOM4, Ca, and Ce that links the metrics guide by absolute URL (https://github.com/zero-dot-force/vibe-check/blob/main/docs/metrics-guide.md).

#### Scenario: Legend present
- **WHEN** any report is rendered
- **THEN** the legend defines all six metric terms and links to the metrics guide

### Requirement: Warnings section
Detailed mode SHALL list analysis warnings carried by the snapshot and packages classified in the Zone of Uselessness.

#### Scenario: Degraded analysis
- **WHEN** the snapshot carries a partial-build warning
- **THEN** the warnings section reproduces that warning

#### Scenario: Zone of Uselessness
- **WHEN** a module is classified in the Zone of Uselessness
- **THEN** the warnings section names that module
