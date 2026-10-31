# Spec Delta

## Purpose

Lets `vibe-check analyze` cover Python and TypeScript codebases by sourcing
structural coupling data from an optional, PATH-discovered ripwire binary,
hybridized with vibe-check's own light language parse, while Go analysis stays
native and releases stay zero-dependency.

## ADDED Requirements

### Requirement: Adapter implements metrics.Adapter for Python and TypeScript

The ripwire adapter SHALL implement the `metrics.Adapter` interface as two
language instances: one returning `"python"` from `Language()` and one returning
`"typescript"`. Both SHALL be registrable in a `metrics.Registry` keyed by their
language identifiers.

#### Scenario: Language identifiers

- **GIVEN** ripwire adapter instances for Python and TypeScript
- **WHEN** `Language()` is called on each
- **THEN** one MUST return `"python"` and the other MUST return `"typescript"`

#### Scenario: Registry registration and retrieval

- **GIVEN** a registry with the Python and TypeScript ripwire adapters registered
- **WHEN** the registry is queried for `"python"` and for `"typescript"`
- **THEN** the respective adapter instances MUST be returned

### Requirement: Per-language capability gating

The adapter SHALL declare, for each language, exactly these universal
capabilities: afferent coupling, efferent coupling, instability, abstractness,
distance, LCOM, and circular dependency detection. Duplication detection SHALL
NOT be declared for either language; its metric values SHALL be zero and the
graph SHALL carry a warning naming duplication as unavailable.

#### Scenario: Capability set per language

- **GIVEN** a ripwire adapter instance for either language
- **WHEN** `Capabilities()` is called
- **THEN** the returned set MUST contain exactly afferent coupling, efferent
  coupling, instability, abstractness, distance, LCOM, and circular dependency
  detection, and MUST NOT contain duplication

#### Scenario: Unsupported metric is zero-valued with warning

- **GIVEN** a successful analysis of a Python or TypeScript project
- **WHEN** the ModuleGraph is produced
- **THEN** duplication fields MUST be zero-valued, graph status MUST be
  `partial`, and `Warnings` MUST contain an entry naming duplication as
  unavailable

### Requirement: Ripwire PATH discovery and absent-binary error

Analysis of Python and TypeScript targets SHALL require a `ripwire` executable
discoverable on PATH. When it is absent, `vibe-check analyze` MUST fail with
exit code 2 and an error message containing install guidance (the ripwire
project location) and the ripwire version vibe-check was validated against. No
ModuleGraph MUST be emitted on this path.

#### Scenario: Absent binary fails with guidance

- **GIVEN** a Python project and no `ripwire` executable on PATH
- **WHEN** `vibe-check analyze <path>` is invoked
- **THEN** the command MUST exit with code 2, print an error naming ripwire and
  where to install it, and write no ModuleGraph to stdout

#### Scenario: Present binary proceeds

- **GIVEN** a Python project and the validated `ripwire` executable on PATH
- **WHEN** `vibe-check analyze <path>` is invoked
- **THEN** analysis MUST proceed via the ripwire adapter

### Requirement: Ripwire version pinning fails closed

Before analyzing, the adapter SHALL determine the discovered ripwire binary's
version and commit and compare them against the pinned, spike-validated surface
(`0.6.9` at commit `bedc0cd`). On any mismatch or unparseable version output,
analysis MUST fail closed with an error naming the validated version and commit;
no ModuleGraph MUST be emitted.

#### Scenario: Matching version proceeds

- **GIVEN** ripwire on PATH reporting version 0.6.9 at commit bedc0cd
- **WHEN** a Python or TypeScript project is analyzed
- **THEN** analysis MUST proceed

#### Scenario: Version mismatch fails closed

- **GIVEN** ripwire on PATH reporting any other version or commit
- **WHEN** a Python or TypeScript project is analyzed
- **THEN** the adapter MUST return an error naming the validated version and
  commit, and no ModuleGraph MUST be produced

#### Scenario: Unparseable version output fails closed

- **GIVEN** ripwire on PATH whose version output cannot be parsed
- **WHEN** a Python or TypeScript project is analyzed
- **THEN** the adapter MUST return an error rather than analyzing with an
  unknown tool version

### Requirement: Provenance records the external source tool

ModuleGraphs produced via ripwire SHALL record the external source tool identity
in the provenance input as optional fields: source tool name (`"ripwire"`),
source tool version, and source tool commit. These fields MUST be omitted for
Go-native graphs. The ModuleGraph JSON schema file, the verdict thresholds, and
the report engine MUST remain unchanged, and ripwire graphs MUST pass
`metrics.Validate`.

#### Scenario: Source tool recorded on ripwire graphs

- **GIVEN** a successful ripwire-backed analysis with provenance enabled
- **WHEN** the ModuleGraph JSON is inspected
- **THEN** the provenance input MUST carry the ripwire name, version, and commit,
  and the graph MUST pass `metrics.Validate`

#### Scenario: Source tool omitted on Go graphs

- **GIVEN** a Go-native analysis with provenance enabled
- **WHEN** the ModuleGraph JSON is inspected
- **THEN** the provenance input MUST NOT carry any source tool fields

#### Scenario: No-provenance flag drops source tool identity

- **GIVEN** a ripwire-backed analysis run with `--no-provenance`
- **WHEN** the ModuleGraph JSON is inspected
- **THEN** no provenance object (and no source tool fields) MUST be present

### Requirement: Analyze detects project language and selects the adapter

`vibe-check analyze` SHALL select the analysis adapter via deterministic marker
precedence: `go.mod` selects the native Go adapter; otherwise Python markers
select the Python ripwire adapter; otherwise TypeScript markers select the
TypeScript ripwire adapter; otherwise a source-file census decides, defaulting
to the Go adapter's existing behavior.

#### Scenario: Go project uses the native adapter

- **GIVEN** a directory containing a `go.mod`
- **WHEN** `vibe-check analyze <path>` is invoked
- **THEN** the native Go adapter MUST be selected and ripwire MUST NOT be
  invoked

#### Scenario: Python project selects the ripwire adapter

- **GIVEN** a directory containing a Python marker (`pyproject.toml`,
  `setup.py`, or `setup.cfg`) and no `go.mod`
- **WHEN** `vibe-check analyze <path>` is invoked
- **THEN** the Python ripwire adapter MUST be selected

#### Scenario: TypeScript project selects the ripwire adapter

- **GIVEN** a directory containing `tsconfig.json`, no `go.mod`, and no Python
  markers
- **WHEN** `vibe-check analyze <path>` is invoked
- **THEN** the TypeScript ripwire adapter MUST be selected

#### Scenario: Markerless census

- **GIVEN** a directory with no `go.mod` or Python/TypeScript markers but with
  source files
- **WHEN** `vibe-check analyze <path>` is invoked
- **THEN** the adapter MUST be selected by majority source-file extension
  (`.py` → Python; `.ts`/`.tsx` → TypeScript), deterministically

#### Scenario: No markers preserve existing behavior

- **GIVEN** a directory with no recognized language markers
- **WHEN** `vibe-check analyze <path>` is invoked
- **THEN** behavior MUST match the pre-change Go-only path (the Go adapter's
  no-packages error)

### Requirement: Go analysis remains native and unchanged

For targets detected as Go, `vibe-check analyze` MUST NOT require or invoke
ripwire and MUST produce ModuleGraph output equivalent to the pre-change
behavior for identical input and provenance settings.

#### Scenario: Go project analyzes without ripwire

- **GIVEN** a valid Go module and no `ripwire` executable on PATH
- **WHEN** `vibe-check analyze <path>` is invoked
- **THEN** the analysis MUST succeed exactly as before this change

#### Scenario: Go output unchanged

- **GIVEN** the same Go module analyzed before and after this change with equal
  provenance settings
- **WHEN** the two ModuleGraph JSON outputs are compared
- **THEN** they MUST be equivalent (byte-identical under `--no-provenance`)

### Requirement: Crawl scoping from the package root

The adapter SHALL scope ripwire's crawl to the detected language-appropriate
package root, never the repository root, because src-layout repo-root crawls
silently yield `ca=ce=0`.

#### Scenario: src-layout Python project yields real coupling

- **GIVEN** a src-layout Python project whose package lives under `src/`
- **WHEN** the project is analyzed
- **THEN** the crawl MUST be rooted so that modules report their real afferent
  and efferent coupling rather than all zeros

#### Scenario: Flat layout project

- **GIVEN** a flat-layout Python or TypeScript project
- **WHEN** the project is analyzed
- **THEN** the crawl root MUST resolve to the package root and analysis MUST
  proceed

### Requirement: Zero-coupling result guard

When a crawl yields no dependency edges although source files exist in scope,
the adapter MUST NOT present an all-zero graph as a complete result; it MUST
either return an error or emit a graph with status `partial` and a warning
describing the empty dependency graph.

#### Scenario: Silent zero graph is not presented as complete

- **GIVEN** a target with source files in scope whose crawl produces no
  dependency edges
- **WHEN** the adapter assembles the ModuleGraph
- **THEN** the result MUST NOT be a status-`complete` all-zero graph

### Requirement: Crawl denylist pre-check warning

The adapter SHALL detect source-bearing directories under the crawl scope whose
names match ripwire's crawl denylist (`build`, `dist`, `out`, `target`,
`vendor`, `third_party`, `captures`), which ripwire hard-prunes, and SHALL emit
a `Warning` listing the pruned directories so their exclusion from metrics is
visible.

#### Scenario: Denylisted source directory warns

- **GIVEN** a project containing source files under a denylist-named directory
  such as `operations/build`
- **WHEN** the project is analyzed
- **THEN** `Warnings` MUST contain an entry with a machine-readable code naming
  the crawl denylist and listing the pruned directory

#### Scenario: No denylisted directories

- **GIVEN** a project with no source files under denylist-named directories
- **WHEN** the project is analyzed
- **THEN** no denylist warning MUST be emitted

### Requirement: Coupling metrics from uncapped ripwire sources

Ca, Ce, and Instability SHALL be sourced from the `--arch` metrics-table
aggregates, with per-file fan-in/out available via `--export=cc.json`, and MUST
NOT be sourced from per-file `<inc>` lists, which are capped at 40 entries per
file. Per-file values SHALL be aggregated into directory modules by
deterministic rules.

#### Scenario: Coupling read from metrics-table aggregates

- **GIVEN** ripwire output containing a metrics table and per-file `<inc>` lists
- **WHEN** coupling values are mapped
- **THEN** Ca/Ce MUST reflect the uncapped table/fan-in values, not the
  40-capped `<inc>` lists

#### Scenario: Deterministic aggregation to directory modules

- **GIVEN** the same ripwire output mapped twice
- **WHEN** per-file values are aggregated into directory modules
- **THEN** the aggregation MUST produce identical module coupling values

### Requirement: Cycle detection over ripwire dependency edges

Circular dependencies SHALL be derived from ripwire's dependency-edge data
(`--deps`, XML parsed adapter-internally) using deterministic strongly-connected
component analysis. Because ripwire's own cycle listing is capped (20 SCCs × 12
members, no total count), cycle membership SHALL be treated as approximate,
corroborated against the dependency edges, and a `Warning` MUST be emitted
noting the approximation.

#### Scenario: Cycle derived from dependency edges

- **GIVEN** ripwire dependency edges forming a cycle between two directory
  modules
- **WHEN** the project is analyzed
- **THEN** the ModuleGraph `Cycles` MUST contain that cycle with canonical,
  deterministically ordered membership

#### Scenario: Cycle approximation warning

- **GIVEN** a project whose analysis includes cycle data
- **WHEN** the ModuleGraph is produced
- **THEN** `Warnings` MUST contain an entry with a machine-readable code noting
  that cycle membership is approximate due to listing caps

### Requirement: TypeScript abstractness from ripwire interface share

For TypeScript, `ExportedTypes` and `AbstractTypes` SHALL be derived from
ripwire's type data as interface share, with interfaces (including nested ones)
counted as abstract. The known biases — `abstract class` undercounting
(conservative) and type aliases inflating the denominator — MUST be surfaced in
a `Warning`.

#### Scenario: Interface share computed

- **GIVEN** a TypeScript project whose ripwire type data reports interfaces
- **WHEN** the project is analyzed
- **THEN** `AbstractTypes` MUST count interfaces including nested ones and
  `ExportedTypes` MUST reflect the type-count denominator

#### Scenario: Known biases warned

- **GIVEN** a TypeScript analysis
- **WHEN** the ModuleGraph is produced
- **THEN** `Warnings` MUST contain an entry documenting the abstract-class
  undercount and type-alias denominator inflation

### Requirement: Python abstractness from vibe-check's own light parse

For Python, `ExportedTypes` and `AbstractTypes` SHALL be computed by
vibe-check's own light parse of class declarations: classes with abstract bases
(`abc.ABC`, `ABCMeta`, `Protocol`) are abstract; classes without
underscore-prefixed names are exported. ripwire's structurally-zero Python
abstractness MUST NOT be used.

#### Scenario: Abstract base class counted as abstract

- **GIVEN** a Python module defining `class Foo(abc.ABC)`
- **WHEN** the project is analyzed
- **THEN** the containing module's `AbstractTypes` MUST include it and
  `ExportedTypes` MUST include it

#### Scenario: Concrete class counted as concrete

- **GIVEN** a Python module defining `class Bar:`
- **WHEN** the project is analyzed
- **THEN** `ExportedTypes` MUST include it and `AbstractTypes` MUST NOT

#### Scenario: Underscore-prefixed class excluded

- **GIVEN** a Python module defining `class _Internal:`
- **WHEN** the project is analyzed
- **THEN** neither `ExportedTypes` nor `AbstractTypes` MUST count it

### Requirement: LCOM4 field usage from vibe-check's own light parse

For both Python and TypeScript, LCOM4 SHALL be computed by vibe-check's own
light parse: methods bound to a class are nodes, edges connect methods that
access at least one common instance field (`self.<attr>` in Python,
`this.<prop>` in TypeScript), and LCOM4 is the number of connected components
per directory module. Module-level functions without a class binding are
excluded.

#### Scenario: Cohesive class

- **GIVEN** a class whose three methods all access the same field
- **WHEN** the module is analyzed
- **THEN** its LCOM MUST be 1

#### Scenario: Split-field class

- **GIVEN** a class with four methods where two access field `x` only and two
  access field `y` only
- **WHEN** the module is analyzed
- **THEN** its LCOM MUST be 2

#### Scenario: No bound methods

- **GIVEN** a module containing only module-level functions and no classes with
  methods
- **WHEN** the module is analyzed
- **THEN** its LCOM MUST be 0

### Requirement: Approximation warnings carry measured error budgets

Every known mapping approximation SHALL emit a `Warning` with a
machine-readable code: name-resolved import edges, name ambiguity rates, the
directory-as-module unit, and the spike-measured error budgets (Python import
graph 1.9% lost and 2.2% spurious edges; TypeScript 0.16% unresolved imports).
Graphs carrying warnings MUST have status `partial`.

#### Scenario: Error-budget warnings present

- **GIVEN** a successful ripwire-backed analysis
- **WHEN** the ModuleGraph is produced
- **THEN** `Warnings` MUST include machine-coded entries for the measured error
  budgets of the analyzed language, and status MUST be `partial`

#### Scenario: Warning codes are machine-readable

- **GIVEN** the warnings emitted by a ripwire-backed analysis
- **WHEN** their `Code` fields are inspected
- **THEN** each MUST be a stable, machine-readable identifier (for example
  `error-budget`, `denylist-pruned`, `cycle-approximation`)

### Requirement: Deterministic output and golden fixtures

Two analyses of the same input MUST produce byte-identical ModuleGraph JSON
modulo the provenance timestamp. Modules MUST be sorted by module path, and
module, cycle, and warning slices MUST be non-nil. Golden fixtures for Python
and TypeScript SHALL be committed under the adapter's testdata, and the full
test suite MUST pass without ripwire installed.

#### Scenario: Two-run byte equality

- **GIVEN** the same Python or TypeScript input analyzed twice with
  `--no-provenance`
- **WHEN** the two ModuleGraph JSON outputs are compared
- **THEN** they MUST be byte-identical

#### Scenario: Modules sorted deterministically

- **GIVEN** a ripwire-backed analysis result
- **WHEN** the `Modules` slice is inspected
- **THEN** it MUST be ordered by module path lexicographically

#### Scenario: Test suite passes without ripwire

- **GIVEN** a development or CI environment with no `ripwire` on PATH
- **WHEN** `go test -race -count=1 ./...` runs
- **THEN** all tests MUST pass, with ripwire-dependent behaviors exercised via
  committed fixture replay

### Requirement: Schema-valid ModuleGraph output

Ripwire-produced ModuleGraphs SHALL pass `metrics.Validate` when serialized to
JSON, use the current schema version constant, and set `Language` to `"python"`
or `"typescript"`.

#### Scenario: Validation passes

- **GIVEN** a successful ripwire-backed analysis
- **WHEN** the ModuleGraph is marshaled to JSON and passed to
  `metrics.Validate`
- **THEN** validation MUST return no error

### Requirement: Subprocess hygiene and data-source constraints

The adapter SHALL invoke ripwire with a sanitized environment
(`metrics.SanitizeEnvironment`), bounded output capture, and context
cancellation that terminates the subprocess. The adapter MUST NOT read
ripwire's internal cache, MUST NOT consume ranked agent-facing answers as
metrics input, and MUST NOT vendor or bundle the ripwire binary.

#### Scenario: Credential environment variables not leaked

- **GIVEN** the process environment contains `GITHUB_TOKEN=secret`
- **WHEN** the adapter spawns ripwire
- **THEN** the subprocess environment MUST NOT contain `GITHUB_TOKEN`

#### Scenario: Context cancellation terminates the subprocess

- **GIVEN** a ripwire-backed analysis in flight
- **WHEN** the context is cancelled
- **THEN** the adapter MUST return an error wrapping the context error and the
  ripwire subprocess MUST NOT be left running

#### Scenario: Analysis does not depend on ripwire's cache

- **GIVEN** an environment where ripwire's internal cache directory is absent
- **WHEN** a Python or TypeScript project is analyzed
- **THEN** analysis MUST succeed using only the documented structural outputs
  (`--arch`, `--deps`, `--export=cc.json`)
