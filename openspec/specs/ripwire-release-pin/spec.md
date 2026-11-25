# ripwire-release-pin Specification

## Purpose
Moves the ripwire adapter's validated pin to the first ripwire release carrying
all four upstream Python import-resolution fixes (redhat-et/ripwire#478), and
adapts the adapter to that release's `--arch` surface so mixed-language trees
analyze with honest floor reporting instead of failing.

`openspec/specs/` holds no archived specs, so refinements of the
`ripwire-adapter` capability are ADDED here and supersede it:

| Requirement here | Supersedes in `ripwire-adapter` | Was |
|------------------|----------------------------------|-----|
| Release version pin fails closed | Ripwire version pinning fails closed | pin `0.6.9` @ `bedc0cd` |
| Denylist warning unions both pruning sources | Crawl denylist pre-check warning | filesystem pre-check only |
| Arch parse requires only consumed attributes | Coupling metrics from uncapped ripwire sources | every row attribute and seven aggregates required |
| Error budgets measured against the pinned release | Approximation warnings carry measured error budgets | spike's 0.6.9 figures (Py 1.9%/2.2%, TS 0.16%) |

## Requirements

### Requirement: Release version pin fails closed

The adapter SHALL validate the discovered ripwire binary against one pinned
(version, commit) pair: the first published ripwire release containing
redhat-et/ripwire#533. Any mismatch — including a development build — or
unparseable version output MUST fail closed with an error naming the pinned
pair, and no ModuleGraph MUST be emitted.

#### Scenario: Matching version proceeds

- **GIVEN** ripwire on PATH reporting the pinned release version at a commit
  extending the pinned commit prefix
- **WHEN** a Python or TypeScript project is analyzed
- **THEN** analysis MUST proceed

#### Scenario: Previous pin is refused

- **GIVEN** ripwire on PATH reporting version `0.6.9` at commit `bedc0cdcf`
- **WHEN** a Python or TypeScript project is analyzed
- **THEN** the adapter MUST return a version-mismatch error naming the new
  validated version and commit, and no ModuleGraph MUST be produced

#### Scenario: Development build is refused

- **GIVEN** ripwire on PATH reporting `0.6.10-dev.65+g1edb6fd6a` with
  `built_from=1edb6fd6a`
- **WHEN** a Python or TypeScript project is analyzed
- **THEN** the adapter MUST return a version-mismatch error, and no ModuleGraph
  MUST be produced

#### Scenario: Unparseable version output fails closed

- **GIVEN** ripwire on PATH whose version output cannot be parsed
- **WHEN** a Python or TypeScript project is analyzed
- **THEN** the adapter MUST return an error rather than analyzing with an
  unknown tool version

### Requirement: Denylist warning unions both pruning sources

The `denylist-pruned` warning SHALL list the sorted, de-duplicated union of the
directories found by the adapter's filesystem pre-check that ripwire did not
index and the package paths
ripwire reports under `<pruned_packages>` in its `--arch` output. When
ripwire's package count exceeds the paths it lists, the warning MUST state the
count.

#### Scenario: Denylisted package reported by both sources

- **GIVEN** a Python project with source under `operations/build/` containing
  an `__init__.py`
- **WHEN** the project is analyzed
- **THEN** `Warnings` MUST contain exactly one `denylist-pruned` entry listing
  `operations/build` once

#### Scenario: Plain pruned directory reported by the pre-check

- **GIVEN** a project with source files under a `dist/` directory that holds no
  `__init__.py`
- **WHEN** the project is analyzed
- **THEN** the `denylist-pruned` warning MUST list that directory even though
  ripwire's `--arch` output does not row it

#### Scenario: Package reported only by ripwire

- **GIVEN** ripwire `--arch` output whose `<pruned_packages>` lists a path the
  filesystem pre-check did not find
- **WHEN** the output is mapped
- **THEN** the `denylist-pruned` warning MUST include that path

#### Scenario: Partial ripwire listing disclosed

- **GIVEN** ripwire `--arch` output whose `<pruned_packages count=>` exceeds
  the number of `<p>` rows it lists
- **WHEN** the output is mapped
- **THEN** the `denylist-pruned` warning MUST state ripwire's total count

#### Scenario: Denylist-named directory indexed by ripwire is not reported

- **GIVEN** ripwire `--arch` output with a module row at or beneath a
  denylist-named, source-bearing directory (a `build/` whose files git tracks)
- **WHEN** the project is analyzed
- **THEN** the `denylist-pruned` warning MUST NOT list that directory, and a
  denylist-named directory nested beneath it that ripwire did not index MUST
  still be listed

#### Scenario: No denylisted directories

- **GIVEN** a project with no source under denylist-named directories and no
  `<pruned_packages>` element in ripwire's output
- **WHEN** the project is analyzed
- **THEN** no `denylist-pruned` warning MUST be emitted

### Requirement: Arch parse requires only consumed attributes

The `--arch` parser SHALL require, fail-closed, only the attributes the adapter
consumes — `path`, `types`, `abstract`, and `ca`/`ce` on rows not marked
`zone="unknown"` — and MUST NOT fail on absent unconsumed attributes or
aggregates. Ca/Ce MUST come from the metrics-table rows, never the capped
`<inc>` lists. Paths with or without a leading `./` MUST map identically.

#### Scenario: Coupling read from metrics-table rows

- **GIVEN** ripwire output containing a metrics table and per-file `<inc>` lists
- **WHEN** coupling values are mapped
- **THEN** Ca/Ce MUST reflect the uncapped table values, not the capped `<inc>`
  lists

#### Scenario: Unconsumed attributes may be absent

- **GIVEN** an `--arch` row carrying `path`, `ca`, `ce`, `types`, and `abstract`
  but no `I`, `D`, or `reachable` attribute, and a `<metrics>` element without
  `propagation_cost`
- **WHEN** the output is parsed
- **THEN** parsing MUST succeed

#### Scenario: Consumed attribute still fails closed

- **GIVEN** an `--arch` row not marked `zone="unknown"` that lacks `ca`
- **WHEN** the output is parsed
- **THEN** parsing MUST fail with an error naming the row and the missing
  attribute

#### Scenario: Both path forms map identically

- **GIVEN** two `--arch` outputs identical except that one prints row paths as
  `./fixturepy/cycle_a` and the other as `fixturepy/cycle_a`
- **WHEN** both are mapped
- **THEN** the resulting module identities MUST be identical

#### Scenario: Deterministic aggregation to directory modules

- **GIVEN** the same ripwire output mapped twice
- **WHEN** rows are mapped into directory modules
- **THEN** the mapping MUST produce identical module coupling values

### Requirement: Error budgets measured against the pinned release

The `error-budget` warning for each language SHALL state figures measured
against the pinned ripwire release by the harness committed with the change
that set the pin, and MUST name the pinned version. Graphs carrying warnings
MUST have status `partial`.

#### Scenario: Error-budget warnings present

- **GIVEN** a successful ripwire-backed analysis
- **WHEN** the ModuleGraph is produced
- **THEN** `Warnings` MUST include an `error-budget` entry for the analyzed
  language naming the pinned ripwire version, and status MUST be `partial`

#### Scenario: Warning codes are machine-readable

- **GIVEN** the warnings emitted by a ripwire-backed analysis
- **WHEN** their `Code` fields are inspected
- **THEN** each MUST be a stable, machine-readable identifier (for example
  `error-budget`, `denylist-pruned`, `coupling-floor`)

### Requirement: Coupling floors for modules ripwire cannot resolve

When ripwire marks an `--arch` row `zone="unknown"`, the adapter SHALL treat
that row's `ca`/`ce` as floors (absent means a floor of zero) and MUST NOT fail
analysis. The graph SHALL carry one `coupling-floor` warning naming the
affected modules and their `no_file_edges=` languages, each list sorted.

#### Scenario: Mixed-language Python tree analyzes

- **GIVEN** a Python project whose `cycle_a` package also contains a `.go` file,
  and ripwire output marking that module `zone="unknown" no_file_edges="go"`
  with no `I` or `D` attribute
- **WHEN** the project is analyzed
- **THEN** a ModuleGraph MUST be produced, it MUST pass `metrics.Validate`, its
  status MUST be `partial`, and `Warnings` MUST contain one `coupling-floor`
  entry naming `fixturepy.cycle_a` and `go`

#### Scenario: Absent coupling on an unknown row

- **GIVEN** an `--arch` row with `zone="unknown"` and no `ca` or `ce` attribute
- **WHEN** the output is parsed and mapped
- **THEN** the module's Ca and Ce MUST be zero and the module MUST be named in
  the `coupling-floor` warning

#### Scenario: No unknown rows

- **GIVEN** ripwire output with no `zone="unknown"` row
- **WHEN** the project is analyzed
- **THEN** no `coupling-floor` warning MUST be emitted

### Requirement: Fixtures exercise the pinned import fixes

The committed Python fixture SHALL contain, each in its own directory module,
import forms covering each upstream fix the pin adopts — the wide
`from pkg import subpkg` form, an aliased `import a.b as c`, and an in-package
absolute import whose name collides with a sibling module — and the committed
goldens SHALL pin the resulting module coupling, so that a pin change altering
their resolution produces a golden diff.

#### Scenario: Wide-form subpackage edge pinned

- **GIVEN** the committed Python fixture and recordings from the pinned release
- **WHEN** the fixture is analyzed through the fake-ripwire harness
- **THEN** the `fixturepy.wide` module, which does `from fixturepy import
  cycle_b`, MUST show an efferent edge to `fixturepy.cycle_b`, matching the
  golden

#### Scenario: Aliased import edge pinned

- **GIVEN** the committed Python fixture and recordings from the pinned release
- **WHEN** the fixture is analyzed
- **THEN** the `fixturepy.alias` module, which does `import
  fixturepy.cycle_a.api as ca`, MUST show an efferent edge to
  `fixturepy.cycle_a`

#### Scenario: Stdlib-name collision draws no edge

- **GIVEN** the committed Python fixture and recordings from the pinned release
- **WHEN** the fixture is analyzed
- **THEN** the `fixturepy.collide` module, whose `import logging` sits beside a
  sibling `logging.py`, MUST have Ce of zero
