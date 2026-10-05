## ADDED Requirements

### Requirement: Duplication detection via AST structural similarity

The system SHALL detect structurally similar Go function and method bodies within a package using AST-based comparison with identifier and literal normalization. Two functions SHALL be considered duplicates when their normalized AST subtrees are deeply equal.

Normalization SHALL replace all identifiers (variable, type, function names) with canonical placeholders. Normalization SHALL replace all basic literals (integer, float, string, rune) with canonical placeholders. Comments SHALL be stripped before comparison.

Only function and method declarations with at least 6 significant lines (excluding blank lines and single-token lines such as solitary braces) SHALL be compared.

Files matching `*.gen.go`, `*.pb.go` SHALL be excluded from duplication detection. Files in directories named `mock`, `generated`, `mocks`, or `testdata` SHALL be excluded.

#### Scenario: Two functions with identical structure but different variable names are detected as duplicates

- **WHEN** a package contains `func foo(a int) { x := a + 1; return x }` and `func bar(b int) { y := b + 1; return y }`
- **THEN** the system reports them as a duplicate pair with a similarity score of 1.0

#### Scenario: Two functions with different control flow are not detected as duplicates

- **WHEN** a package contains `func foo() { if x { a() } }` and `func bar() { for { a() } }`
- **THEN** the system does NOT report a duplication between them

#### Scenario: Functions shorter than 6 significant lines are excluded

- **WHEN** a package contains two identical 3-line functions
- **THEN** the system does NOT count them as duplicated code

#### Scenario: Generated files are excluded

- **WHEN** a package contains a `*.gen.go` file with structural duplicates matching other files
- **THEN** the system excludes the `*.gen.go` file from comparison

#### Scenario: Package with no Go files produces empty result
- **WHEN** a package has no `.go` files or only excluded files
- **THEN** the system returns an empty `Duplications` array with a 0.0% duplication percentage

#### Scenario: AST parse failure in a file is handled gracefully
- **WHEN** `go/packages` successfully loads a package but one or more files fail to parse
- **THEN** those files are skipped and a `Warning` is emitted
- **AND** remaining files in the package are still compared

#### Scenario: Package with only non-function declarations
- **WHEN** a package contains only type declarations, constants, and variables with no functions or methods
- **THEN** the system reports zero duplications

#### Scenario: Package with fewer than two comparable functions
- **WHEN** a package has only one function with 6 or more significant lines
- **THEN** the system reports zero duplications since there is nothing to compare against

### Requirement: Duplication percentage computation

The system SHALL compute a duplication percentage per module as `(duplicatedLines / totalLines) * 100`, where `duplicatedLines` is the count of significant lines covered by at least one duplicate block (each line counted at most once), and `totalLines` is the total number of significant lines in the module's non-generated, non-excluded source files.

#### Scenario: Module with 200 total lines and one 20-line duplicate pair

- **WHEN** a module has 200 total lines and one 20-line block is duplicated once
- **THEN** the duplication percentage is 10.0%

#### Scenario: Module with no duplicates

- **WHEN** a module has no structural duplicates
- **THEN** the duplication percentage is 0.0%

### Requirement: Duplication result in JSON output

The JSON output SHALL include duplication information for each module via a `Duplications` field in `ModuleResult` containing zero or more `Duplication` objects. Each `Duplication` object SHALL contain the module path, two or more block locations (file path, start line, end line), a similarity score from 0.0 to 1.0, and the number of significant lines in each block.

The `Duplication` type SHALL be language-agnostic and defined in the `metrics` package.

#### Scenario: Module has one duplicate pair

- **WHEN** `vibe-check analyze ./...` runs on a project with a duplicated function
- **THEN** the JSON output contains a `duplications` array in the affected module's result
- **AND** each entry contains `blocks` with file paths, `startLine`, `endLine`, and `similarity`

### Requirement: Duplication threshold gate

The system SHALL accept a `--max-duplication` flag on `vibe-check analyze` accepting a float value in [0.0, 100.0] with a default of 5.0. When any module's duplication percentage exceeds this threshold, the command SHALL print a violation message to stderr and exit with code 1.

#### Scenario: Duplication exceeds threshold

- **WHEN** `vibe-check analyze --max-duplication=3.0 ./...` is run on a project where a module has 5.0% duplication
- **THEN** the command exits with code 1
- **AND** a violation message is printed to stderr naming the violating module and its duplication percentage

#### Scenario: Duplication within threshold

- **WHEN** `vibe-check analyze --max-duplication=10.0 ./...` is run on a project where no module exceeds 10.0% duplication
- **THEN** the command exits with code 0
- **AND** no duplication violation is printed

#### Scenario: Invalid flag value

- **WHEN** `vibe-check analyze --max-duplication=-1.0 ./...` is run
- **THEN** the command exits with code 2
- **AND** an error message indicates the valid range [0.0, 100.0]

### Requirement: Schema version bump

The system SHALL increment `SchemaVersionCurrent` from `"1.2"` to `"1.3"` to reflect the addition of the `Duplications` field to `ModuleResult`. The JSON schema at `metrics/modulegraph.schema.json` SHALL be updated accordingly.

#### Scenario: JSON output carries schema version 1.3

- **WHEN** `vibe-check analyze ./...` is run with this change
- **THEN** the output JSON contains `"schemaVersion": "1.3"`